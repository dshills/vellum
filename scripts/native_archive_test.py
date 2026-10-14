import hashlib
import os
from pathlib import Path
import tempfile
import unittest
from unittest import mock

import native_archive as native


def member(name, data, extended=False):
    encoded = name.encode()
    if extended:
        encoded += b"\x00" * ((-len(encoded)) % 4)
        label, body = ("#1/" + str(len(encoded))).encode(), encoded + data
    else:
        label, body = (name + "/").encode(), data
    header = label.ljust(16) + b"0".ljust(12) + b"0".ljust(6) + b"0".ljust(6) + b"644".ljust(8) + str(len(body)).encode().ljust(10) + b"`\n"
    return header + body + (b"\n" if len(body) % 2 else b"")


def link_map(archive, names):
    return "# Path: /temporary/a.out\n# Arch: arm64\n# Object files:\n[ 0] linker synthesized\n" + "".join(f"[{i+1}] {archive}({name})\n" for i,name in enumerate(names)) + "# Sections:\n# Symbols:\nnot interpreted as an object\n"


class NativeArchiveTests(unittest.TestCase):
    def test_exact_payloads_and_selected_subset_without_extraction(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp); archive = root / "lib test.a"; mapping = root / "build.map"
            data = b"!<arch>\n" + member("short.o",b"odd") + member("long.rust.object-name.rcgu.o",b"\x00object bytes\xff",True)
            archive.write_bytes(data); mapping.write_text(link_map(archive,["long.rust.object-name.rcgu.o"]))
            a = native.archive_members(archive); m = native.linker_objects(mapping); selected = native.selected_archive_members(m,archive,a)
            self.assertEqual(a['sha256'],hashlib.sha256(data).hexdigest())
            self.assertEqual([(v['name'],v['bytes'],v['sha256'])for v in a['members']], [('short.o',3,hashlib.sha256(b'odd').hexdigest()),('long.rust.object-name.rcgu.o',14,hashlib.sha256(b'\x00object bytes\xff').hexdigest())])
            self.assertEqual(selected,[{'linker_object_index':1,**a['members'][1]}]);self.assertEqual(len(list(root.iterdir())),2)

    def test_missing_or_ambiguous_selected_names_refuse(self):
        with tempfile.TemporaryDirectory() as temp:
            a=Path(temp)/'lib.a';a.write_bytes(b'!<arch>\n'+member('one.o',b'a')+member('one.o',b'b'))
            records=native.archive_members(a)
            for name in ['missing.o','one.o']:
                with self.assertRaisesRegex(ValueError,'absent or ambiguous'):
                    native.selected_archive_members({'objects':[str(a)+'('+name+')']},a,records)
            self.assertEqual(native.selected_archive_members({'objects':['/elsewhere/lib.a(one.o)']},a,records),[])
            previous = Path.cwd()
            try:
                os.chdir(a.parent)
                # Even a collector cwd that contains the named archive cannot
                # supply the unrecorded linker cwd.
                self.assertEqual(native.selected_archive_members({'objects':['lib.a(one.o)']},a,records),[])
            finally:
                os.chdir(previous)

    def test_archive_corruption_and_unsupported_names_refuse(self):
        with tempfile.TemporaryDirectory() as temp:
            f=Path(temp)/'bad.a';valid=b'!<arch>\n'+member('one.o',b'odd')
            corrupt=[b'!<thin>\n',valid[:-1],valid[:-1]+b'X',valid[:8]+b'bad',valid[:56]+b'x'+valid[57:],b'!<arch>\n'+member('../escape.o',b'x'),b'!<arch>\n'+member('bad\x00name.o',b'x',True),b'!<arch>\n'+member('/123',b'x')]
            for data in corrupt:
                f.write_bytes(data)
                with self.assertRaises((ValueError,UnicodeError),msg=repr(data[:70])):native.archive_members(f)
            # Use the literal GNU table header: suffix normalization must not
            # turn it into the distinct '/' symbol-table name.
            f.write_bytes(b'!<arch>\n' + member('one.o', b'x').replace(b'one.o/          ', b'//              ', 1))
            with self.assertRaisesRegex(ValueError, 'GNU archive name table'):
                native.archive_members(f)

    def test_byte_member_and_name_bounds_refuse(self):
        with tempfile.TemporaryDirectory() as temp:
            f=Path(temp)/'lib.a';f.write_bytes(b'!<arch>\n'+member('one.o',b'x',True))
            for name,bound in [('MAX_ARCHIVE_BYTES',7),('MAX_MEMBERS',0),('MAX_NAME_BYTES',1)]:
                with mock.patch.object(native,name,bound),self.assertRaises(ValueError):native.archive_members(f)

    def test_malformed_map_missing_sections_duplicate_indices_and_limits_refuse(self):
        with tempfile.TemporaryDirectory() as temp:
            f=Path(temp)/'build.map';valid=link_map('/verified/lib.a',['one.o'])
            variants=[valid.replace('# Sections:','# NotSections:'),valid.replace('[1]','[0]'),valid.replace('# Arch: arm64','# Arch: unknown'),valid.replace('[1]','bad'),valid.replace('# Object files:','# Object files:\n# Object files:')]
            for data in variants:
                f.write_text(data)
                with self.assertRaises(ValueError):native.linker_objects(f)
            f.write_text(valid)
            for name,bound in [('MAX_MAP_BYTES',2),('MAX_OBJECTS',1),('MAX_NAME_BYTES',4)]:
                with mock.patch.object(native,name,bound),self.assertRaises(ValueError):native.linker_objects(f)
            f.write_bytes(valid.encode()+b'x'*65537)
            with self.assertRaisesRegex(ValueError,'line exceeds bound'):native.linker_objects(f)

    def test_symlinks_nonregular_and_changed_streams_refuse(self):
        with tempfile.TemporaryDirectory() as temp:
            f=Path(temp)/'lib.a';f.write_bytes(b'!<arch>\n'+member('one.o',b'x'));alias=Path(temp)/'alias';alias.symlink_to(f)
            with self.assertRaisesRegex(ValueError,'symlink'):native.archive_members(alias)
            with self.assertRaises((ValueError,IsADirectoryError)):native.archive_members(Path(temp))
            if hasattr(os, 'mkfifo'):
                fifo=Path(temp)/'no-writer.map';os.mkfifo(fifo)
                with self.assertRaisesRegex(ValueError, 'nonregular'):
                    native.linker_objects(fifo)
            actual=native.unchanged
            def change(stream,before):
                with f.open('ab')as out:out.write(b'x')
                actual(stream,before)
            with mock.patch.object(native,'unchanged',side_effect=change),self.assertRaisesRegex(ValueError,'changed'):native.archive_members(f)


if __name__=='__main__':unittest.main()
