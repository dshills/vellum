"""Independent raw-container controls; no font installation or native UI.

Copyright (c) 2026 Davin Hills. SPDX-License-Identifier: MIT.
"""
from pathlib import Path
import argparse
import importlib.util
import json
import os
import struct
import subprocess
import sys
import tempfile
import unittest

EXPECTED = {1: "Vellum Invalid Variation Fixture", 3: "Vellum Invalid Variation Fixture; revision 1",
            4: "Vellum Invalid Variation Fixture", 6: "VellumInvalidVariationFixture-Regular"}


def total(raw):
    return sum(int.from_bytes(raw[i:i + 4].ljust(4, b"\0"), "big") for i in range(0, len(raw), 4)) & 0xffffffff


def parse(raw, checked=False):
    if raw[:4] != b"\0\1\0\0" or len(raw) < 12:
        raise ValueError("header")
    count, search, selector, shift = struct.unpack_from(">4H", raw, 4)
    if count != 28 or 12 + 16 * count > len(raw):
        raise ValueError("directory count")
    if checked and (search, selector, shift) != (256, 4, 192):
        raise ValueError("search fields")
    tables, spans, entries = {}, [], {}
    for i in range(count):
        tag, declared, offset, length = struct.unpack_from(">4sIII", raw, 12 + 16 * i)
        if tag in tables or offset < 12 + 16 * count or offset + length > len(raw):
            raise ValueError("table bounds")
        data = raw[offset:offset + length]
        if checked:
            if offset % 4 or raw[offset + length:(offset + length + 3) // 4 * 4] != bytes((-length) % 4):
                raise ValueError("alignment or padding")
            check_data = data[:8] + bytes(4) + data[12:] if tag == b"head" else data
            if total(check_data) != declared:
                raise ValueError("table checksum")
        tables[tag] = data
        entries[tag] = (12 + 16 * i, offset, length)
        if length:
            spans.append((offset, offset + length))
    spans.sort()
    if any(left[1] > right[0] for left, right in zip(spans, spans[1:])):
        raise ValueError("table overlap")
    if checked and (len(raw) % 4 or total(raw) != 0xb1b0afba or list(tables) != sorted(tables)):
        raise ValueError("container checksum/order")
    name = tables[b"name"]
    version, count, strings = struct.unpack_from(">HHH", name)
    if version != 0 or count != 26 or strings < 6 + count * 12:
        raise ValueError("name directory")
    records = []
    for i in range(count):
        p, e, language, identifier, length, offset = struct.unpack_from(">6H", name, 6 + i * 12)
        if strings + offset + length > len(name):
            raise ValueError("name bounds")
        value = name[strings + offset:strings + offset + length]
        records.append(((p, e, language, identifier), value))
    return tables, records, entries


def verify(original, candidate):
    before, old_names, _ = parse(original)
    after, names, _ = parse(candidate, checked=True)
    if before.keys() != after.keys():
        raise ValueError("changed table set")
    for tag in before:
        a, b = before[tag], after[tag]
        if tag == b"name":
            continue
        if tag == b"head":
            a, b = a[:8] + bytes(4) + a[12:], b[:8] + bytes(4) + b[12:]
        if a != b:
            raise ValueError("changed non-name payload " + repr(tag))
    changed = set()
    for (old_key, old_value), (key, value) in zip(old_names, names):
        if old_key != key:
            raise ValueError("changed record identity")
        if key[3] in EXPECTED:
            if key[:3] != (3, 1, 1033) or value.decode("utf-16-be") != EXPECTED[key[3]]:
                raise ValueError("primary name")
            changed.add(key[3])
        elif old_value != value:
            raise ValueError("changed declaration/style name")
    if changed != set(EXPECTED):
        raise ValueError("primary name set")
    return after


def repaired(raw):
    """Make tampered data checksum-valid to exercise payload/name guards."""
    raw = bytearray(raw)
    _, _, entries = parse(bytes(raw))
    _, head, _ = entries[b"head"]
    raw[head + 8:head + 12] = bytes(4)
    for tag, (directory, offset, length) in entries.items():
        value = total(bytes(raw[offset:offset + length]))
        struct.pack_into(">I", raw, directory + 4, value)
    struct.pack_into(">I", raw, head + 8, (0xb1b0afba - total(raw)) & 0xffffffff)
    return bytes(raw)


class FixtureControls(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.source = SOURCE.read_bytes()
        cls.asset = (ROOT / "VellumInvalidVariationFixture.ttf").read_bytes()
        spec = importlib.util.spec_from_file_location("fixture_generator", ROOT / "generate.py")
        cls.generator = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(cls.generator)

    def test_raw_descriptor_closed_on_wrapper_failure_directory_rejection_and_success(self):
        import errno
        from unittest import mock
        original_open = os.open
        with tempfile.TemporaryDirectory() as name:
            for case in ("wrapper-failure", "directory", "bound-rejection", "success"):
                opened = []
                def capture(*args, **kwargs):
                    descriptor = original_open(*args, **kwargs)
                    opened.append(descriptor)
                    return descriptor
                with self.subTest(case=case), mock.patch.object(self.generator.os, "open", side_effect=capture):
                    if case == "wrapper-failure":
                        with mock.patch.object(self.generator.os, "fdopen", side_effect=OSError("injected stream construction failure")):
                            with self.assertRaisesRegex(OSError, "injected stream construction"):
                                self.generator.read_regular(SOURCE, len(self.source))
                    elif case == "directory":
                        with self.assertRaises((IsADirectoryError, ValueError)):
                            self.generator.read_regular(Path(name), 65536)
                    elif case == "bound-rejection":
                        with self.assertRaisesRegex(ValueError, "bounded regular"):
                            self.generator.read_regular(SOURCE, 1)
                    else:
                        self.assertEqual(self.generator.read_regular(SOURCE, len(self.source)), self.source)
                self.assertEqual(len(opened), 1)
                with self.assertRaises(OSError) as error:
                    os.fstat(opened[0])
                self.assertEqual(error.exception.errno, errno.EBADF)

    def test_metadata_tables_and_malformed_payloads_preserved(self):
        tables = verify(self.source, self.asset)
        self.assertEqual(tables[b"glyf"], b"")
        self.assertEqual(tables[b"loca"], bytes(770))
        self.assertEqual(len(tables[b"gvar"]), 26142)
        self.assertLessEqual(len(EXPECTED[6]), 63)

    def test_deterministic_asset_and_manifest(self):
        first, manifest = self.generator.rename_fixture(self.source)
        second, again = self.generator.rename_fixture(self.source)
        self.assertEqual(first, second)
        self.assertEqual(first, self.asset)
        self.assertEqual(manifest, again)
        recorded = json.loads((ROOT / "manifest.json").read_text())
        self.assertEqual(json.loads(json.dumps(manifest)), recorded)
        self.assertEqual(recorded["unchanged_other_name_records"], 22)

    def test_changed_or_truncated_input_refused(self):
        changed = bytearray(self.source)
        changed[100] ^= 1
        for value in [bytes(changed), self.source[:-1], self.source + b"x", b""]:
            with self.assertRaisesRegex(ValueError, "exact pinned"):
                self.generator.rename_fixture(value)

    def test_checksum_valid_program_tamper_refused(self):
        _, _, entries = parse(self.asset)
        changed = bytearray(self.asset)
        changed[entries[b"gvar"][1] + 50] ^= 1
        changed = repaired(changed)
        parse(changed, checked=True)
        with self.assertRaisesRegex(ValueError, "non-name payload"):
            verify(self.source, changed)

    def test_checksum_valid_primary_name_tamper_refused(self):
        _, _, entries = parse(self.asset)
        changed = bytearray(self.asset)
        offset = entries[b"name"][1]
        _, count, start = struct.unpack_from(">HHH", changed, offset)
        for i in range(count):
            _, _, _, identifier, _, at = struct.unpack_from(">6H", changed, offset + 6 + i * 12)
            if identifier == 6:
                changed[offset + start + at + 1] = ord("X")
                break
        changed = repaired(changed)
        parse(changed, checked=True)
        with self.assertRaisesRegex(ValueError, "primary name"):
            verify(self.source, changed)

    def test_directory_overlap_refused(self):
        _, _, entries = parse(self.asset)
        changed = bytearray(self.asset)
        struct.pack_into(">I", changed, entries[b"name"][0] + 8, entries[b"GPOS"][1])
        with self.assertRaisesRegex(ValueError, "overlap"):
            parse(changed)

    def test_cli_existing_directory_and_bad_input_leave_output_untouched(self):
        with tempfile.TemporaryDirectory() as name:
            root = Path(name)
            output = root / "existing"
            output.mkdir()
            sentinel = output / "sentinel"
            sentinel.write_text("keep")
            command = [sys.executable, "-B", str(ROOT / "generate.py"), "--source", str(SOURCE), "--output", str(output)]
            result = subprocess.run(command, capture_output=True, timeout=30)
            self.assertNotEqual(result.returncode, 0)
            self.assertEqual(list(output.iterdir()), [sentinel])
            self.assertEqual(sentinel.read_text(), "keep")
            invalid = root / "bad"
            invalid.write_bytes(b"bad")
            command[-3] = str(invalid)
            command[-1] = str(root / "new")
            result = subprocess.run(command, capture_output=True, timeout=30)
            self.assertNotEqual(result.returncode, 0)
            self.assertFalse((root / "new").exists())

    def test_cli_refuses_symlink_directory_and_fifo_without_creating_output(self):
        with tempfile.TemporaryDirectory() as name:
            root = Path(name)
            link = root / "symlink"
            link.symlink_to(SOURCE)
            fifo = root / "fifo"
            os.mkfifo(fifo)
            for source in [link, root, fifo]:
                with self.subTest(source=source.name):
                    output = root / "new"
                    result = subprocess.run([sys.executable, "-B", str(ROOT / "generate.py"),
                                             "--source", str(source), "--output", str(output)],
                                            capture_output=True, timeout=3)
                    self.assertNotEqual(result.returncode, 0)
                    self.assertFalse(output.exists())


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--source", required=True, type=Path)
    parser.add_argument("--asset-root", type=Path, default=Path(__file__).resolve().parent)
    args = parser.parse_args()
    SOURCE, ROOT = args.source.resolve(), args.asset_root.resolve()
    unittest.main(argv=[sys.argv[0]], verbosity=2)
