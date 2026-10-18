"""Notice payload identities, bounds and sampled copy validation; no GUI or signing."""
import hashlib
import json
from pathlib import Path
import tempfile
import unittest
from unittest import mock

import app_notices as notices


class NoticeChecks(unittest.TestCase):
    def fixture(self, root):
        root.mkdir()
        (root/"NOTICES/component").mkdir(parents=True)
        body = b"Original declaration\n\xc2\xa9 Example\n"
        (root/"NOTICES/component/LICENSE.txt").write_bytes(body)
        value = {"schema": 1, "target": "darwin", "artifact_sha256": "a"*64,
                 "distribution_audit": "pending", "applicability": "review-required",
                 "scope": "Caller-selected fixture; completeness unverified",
                 "files": {"NOTICES/component/LICENSE.txt": {"bytes": len(body), "sha256": hashlib.sha256(body).hexdigest()}},
                 "components": [{"id": "example@1", "notice_files": ["NOTICES/component/LICENSE.txt"]}]}
        self.write(root, value)
        return value

    def write(self, root, value):
        (root/notices.INVENTORY).write_text(json.dumps(value, indent=2)+"\n")

    def test_copies_original_bytes_and_relocates_with_identical_record(self):
        with tempfile.TemporaryDirectory() as temporary:
            root=Path(temporary);source=root/"source";value=self.fixture(source)
            record=notices.copy(source,root/"copied","a"*64)
            self.assertEqual(record["components"],1);self.assertEqual(record["files"],1)
            self.assertEqual(record["artifact_sha256"],"a"*64)
            self.assertEqual(record["distribution_audit"],"pending")
            self.assertEqual(record["trust"],"caller-selected-not-attested")
            self.assertEqual(notices.load(root/"copied","a"*64)[0],record)
            for path in [notices.INVENTORY,*value["files"]]:self.assertEqual((source/path).read_bytes(),(root/"copied"/path).read_bytes())

    def test_wrong_binary_corrupt_file_and_changed_inventory_refuse(self):
        with tempfile.TemporaryDirectory() as temporary:
            root=Path(temporary)/"source";value=self.fixture(root)
            with self.assertRaisesRegex(ValueError,"executable binding"):notices.load(root,"b"*64)
            record,_=notices.load(root,"a"*64)
            (root/"NOTICES/component/LICENSE.txt").write_bytes(b"x"*value["files"]["NOTICES/component/LICENSE.txt"]["bytes"])
            with self.assertRaisesRegex(ValueError,"checksum"):notices.load(root,"a"*64)
            # Valid but different metadata must not replace the recorded snapshot.
            body=(root/"NOTICES/component/LICENSE.txt").read_bytes();value["files"]["NOTICES/component/LICENSE.txt"]["sha256"]=hashlib.sha256(body).hexdigest();self.write(root,value)
            with self.assertRaisesRegex(ValueError,"changed"):notices.verify(root,"a"*64,record)

    def test_traversal_and_portable_path_aliases_refuse(self):
        for name in ["NOTICES/../escape","/NOTICES/file","NOTICES//file","NOTICES/./file","NOTICES/a\\b","NOTICES/.hidden","NOTICES/a.","NOTICES/é"]:
            with self.subTest(name=name),self.assertRaisesRegex(ValueError,"path"):notices.notice_name(name)
        with tempfile.TemporaryDirectory() as temporary:
            root=Path(temporary)/"source";value=self.fixture(root);name="NOTICES/component/license.TXT";value["files"][name]=next(iter(value["files"].values()));value["components"][0]["notice_files"].append(name);self.write(root,value)
            with self.assertRaisesRegex(ValueError,"case-colliding"):notices.load(root,"a"*64)

    def test_case_aliases_of_directory_prefixes_refuse_before_payload_read(self):
        with tempfile.TemporaryDirectory() as temporary:
            root=Path(temporary)/"source";value=self.fixture(root)
            value["files"]["NOTICES/Component/NOTICE"]=next(iter(value["files"].values()))
            value["components"][0]["notice_files"].append("NOTICES/Component/NOTICE");self.write(root,value)
            with mock.patch.object(notices,"read_regular",wraps=notices.read_regular) as read:
                with self.assertRaisesRegex(ValueError,"case-colliding"):notices.load(root,"a"*64)
                self.assertEqual(read.call_count,1)

    def test_symlinks_root_parent_leaf_and_dangling_refuse(self):
        for kind in ["root","parent","leaf","dangling"]:
            with self.subTest(kind=kind),tempfile.TemporaryDirectory() as temporary:
                base=Path(temporary);root=base/"source";self.fixture(root)
                if kind=="root":(base/"alias").symlink_to(root,target_is_directory=True);root=base/"alias"
                elif kind=="parent":
                    (root/"NOTICES/component").rename(base/"other");(root/"NOTICES/component").symlink_to(base/"other",target_is_directory=True)
                else:
                    path=root/"NOTICES/component/LICENSE.txt";path.rename(base/"other");path.symlink_to(base/("absent" if kind=="dangling" else "other"))
                with self.assertRaisesRegex(ValueError,"symlinks"):notices.load(root,"a"*64)

    def test_exact_coverage_and_references_refuse(self):
        for kind in ["extra-file","extra-directory","missing-file","unreferenced","duplicate-reference","unknown-reference","duplicate-component"]:
            with self.subTest(kind=kind),tempfile.TemporaryDirectory() as temporary:
                root=Path(temporary)/"source";value=self.fixture(root)
                if kind=="extra-file":(root/"extra").write_text("unlisted")
                elif kind=="extra-directory":(root/"extra").mkdir()
                elif kind=="missing-file":(root/"NOTICES/component/LICENSE.txt").unlink()
                elif kind=="unreferenced":value["files"]["NOTICES/unused"]=next(iter(value["files"].values()))
                elif kind=="duplicate-reference":value["components"][0]["notice_files"]*=2
                elif kind=="unknown-reference":value["components"][0]["notice_files"]=["NOTICES/unknown"]
                else:value["components"]*=2
                self.write(root,value)
                with self.assertRaises(ValueError):notices.load(root,"a"*64)

    def test_schema_approval_and_strict_json_refuse(self):
        for key,value in [("schema",True),("target","linux"),("distribution_audit","approved"),("applicability","approved"),("scope",""),("files",[]),("components",[]),("extra","ignored")]:
            with self.subTest(key=key),tempfile.TemporaryDirectory() as temporary:
                root=Path(temporary)/"source";inventory=self.fixture(root);inventory[key]=value;self.write(root,inventory)
                with self.assertRaises(ValueError):notices.load(root,"a"*64)
        for body in [b'{"schema":1,"schema":1}',b'{"schema":NaN}',b'{"schema":1e999}',b'[]']:
            with self.subTest(body=body),tempfile.TemporaryDirectory() as temporary:
                root=Path(temporary)/"source";self.fixture(root);(root/notices.INVENTORY).write_bytes(body)
                with self.assertRaises(ValueError):notices.load(root,"a"*64)

    def test_file_and_aggregate_budgets_admit_before_payload_read(self):
        for key,bound in [("MAX_FILES",0),("MAX_COMPONENTS",0),("MAX_FILE_BYTES",1),("MAX_TOTAL_BYTES",1),("MAX_MANIFEST_BYTES",1)]:
            with self.subTest(key=key),tempfile.TemporaryDirectory() as temporary:
                root=Path(temporary)/"source";self.fixture(root)
                actual=notices.read_regular
                with mock.patch.object(notices,key,bound),mock.patch.object(notices,"read_regular",wraps=actual) as read:
                    with self.assertRaises(ValueError):notices.load(root,"a"*64)
                    self.assertEqual(read.call_count,1)
        with tempfile.TemporaryDirectory() as temporary:
            root=Path(temporary)/"source";value=self.fixture(root);value["files"]["NOTICES/component/LICENSE.txt"]["bytes"]=True;self.write(root,value)
            with self.assertRaisesRegex(ValueError,"identity"):notices.load(root,"a"*64)

    def test_copy_uses_validated_bytes_if_caller_source_changes(self):
        with tempfile.TemporaryDirectory() as temporary:
            root=Path(temporary);source=root/"source";self.fixture(source);original=(source/"NOTICES/component/LICENSE.txt").read_bytes();actual=notices.load
            def mutate_after_admission(path,binary):
                result=actual(path,binary)
                if path==source:(source/"NOTICES/component/LICENSE.txt").write_bytes(b"later unadmitted bytes")
                return result
            with mock.patch.object(notices,"load",side_effect=mutate_after_admission):notices.copy(source,root/"copy","a"*64)
            self.assertEqual((root/"copy/NOTICES/component/LICENSE.txt").read_bytes(),original)


if __name__=="__main__":unittest.main()
