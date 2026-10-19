"""Meaningful failure/ordering checks; no downloads or native display needed."""
import importlib.util
import contextlib
import io
import json
from pathlib import Path
import tempfile
import unittest

ROOT = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location("prepare", ROOT / "prepare.py")
prepare = importlib.util.module_from_spec(spec)
spec.loader.exec_module(prepare)


class PreparationChecks(unittest.TestCase):
    def previous_installations(self, root):
        (root / "native").mkdir()
        dependencies = [{"name": "upstream"}, {"name": "typesetting"}]
        (root / "native/dependencies.json").write_text(json.dumps({"dependencies": dependencies}))
        for dependency in dependencies:
            directory = root / "build" / dependency["name"]
            (directory / "local-data").mkdir(parents=True)
            (directory / "go.mod").write_text("previous " + dependency["name"])
            (directory / "local-data/notes.txt").write_text("retain original local data")
        return dependencies

    def staged_installation(self, dependency, destination):
        destination.mkdir()
        (destination / "go.mod").write_text("replacement " + dependency["name"])

    def readonly_partial_copy(self, destination):
        directory = destination / "readonly/nested"
        directory.mkdir(parents=True)
        source = directory / "partial.go"
        source.write_text("partial immutable module source")
        source.chmod(0o400)
        directory.chmod(0o400)
        directory.parent.chmod(0o500)
        destination.chmod(0o500)

    def test_readonly_partial_copy_cleanup_preserves_preparation_error(self):
        from unittest import mock
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            dependencies = self.previous_installations(root)
            original_error = RuntimeError("copy interrupted before chmod completed")

            def failed_copy(dependency, destination):
                self.readonly_partial_copy(destination)
                raise original_error

            with mock.patch.object(prepare, "ROOT", root), mock.patch.object(prepare, "prepare_dependency", failed_copy):
                with self.assertRaises(RuntimeError) as failed:
                    prepare.main()
            self.assertIs(failed.exception, original_error)
            self.assertEqual(list((root / "build").glob("prepare-*")), [])
            self.assertFalse((root / "build/.prepare-lock").exists())
            for dependency in dependencies:
                directory = root / "build" / dependency["name"]
                self.assertEqual((directory / "go.mod").read_text(), "previous " + dependency["name"])

    def test_cleanup_failure_warns_and_keeps_original_error_and_inspection_path(self):
        from unittest import mock
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            self.previous_installations(root)
            original_error = RuntimeError("original checksum failure")
            remove = prepare.shutil.rmtree

            def failed_copy(dependency, destination):
                self.readonly_partial_copy(destination)
                raise original_error

            def failed_cleanup(path, *args, **kwargs):
                if Path(path).name.startswith("prepare-"):
                    raise OSError("simulated cleanup denial")
                return remove(path, *args, **kwargs)

            warnings = io.StringIO()
            with mock.patch.object(prepare, "ROOT", root), mock.patch.object(prepare, "prepare_dependency", failed_copy), mock.patch.object(prepare.shutil, "rmtree", failed_cleanup), contextlib.redirect_stderr(warnings):
                with self.assertRaises(RuntimeError) as failed:
                    prepare.main()
            self.assertIs(failed.exception, original_error)
            retained, = (root / "build").glob("prepare-*")
            self.assertIn(str(retained), warnings.getvalue())
            self.assertIn("simulated cleanup denial", warnings.getvalue())
            self.assertEqual((retained / "upstream/readonly/nested/partial.go").read_text(), "partial immutable module source")
            self.assertFalse((root / "build/.prepare-lock").exists())

    def test_cleanup_failure_after_success_keeps_successful_installations(self):
        from unittest import mock
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            dependencies = self.previous_installations(root)
            remove = prepare.shutil.rmtree

            def failed_cleanup(path, *args, **kwargs):
                if Path(path).name.startswith("prepare-"):
                    raise OSError("simulated cleanup denial")
                return remove(path, *args, **kwargs)

            warnings, output = io.StringIO(), io.StringIO()
            with mock.patch.object(prepare, "ROOT", root), mock.patch.object(prepare, "prepare_dependency", self.staged_installation), mock.patch.object(prepare.shutil, "rmtree", failed_cleanup), contextlib.redirect_stderr(warnings), contextlib.redirect_stdout(output):
                prepare.main()
            retained, = (root / "build").glob("prepare-*")
            self.assertIn(str(retained), warnings.getvalue())
            self.assertIn("Prepared driver-owned", output.getvalue())
            for dependency in dependencies:
                name = dependency["name"]
                self.assertEqual((root / "build" / name / "go.mod").read_text(), "replacement " + name)
                self.assertEqual((retained / (name + ".previous") / "local-data/notes.txt").read_text(), "retain original local data")
            self.assertFalse((root / "build/.prepare-lock").exists())

    def test_lock_release_failure_does_not_mask_original_error(self):
        from unittest import mock
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            self.previous_installations(root)
            original_error = RuntimeError("original preparation failure")
            remove = Path.rmdir

            def failed_release(path):
                if path == root / "build/.prepare-lock":
                    raise OSError("simulated lock release denial")
                return remove(path)

            warnings = io.StringIO()
            with mock.patch.object(prepare, "ROOT", root), mock.patch.object(prepare, "prepare_dependency", side_effect=original_error), mock.patch.object(Path, "rmdir", failed_release), contextlib.redirect_stderr(warnings):
                with self.assertRaises(RuntimeError) as failed:
                    prepare.main()
            self.assertIs(failed.exception, original_error)
            lock = root / "build/.prepare-lock"
            self.assertTrue(lock.exists())
            self.assertIn(str(lock), warnings.getvalue())
            self.assertIn("before manual removal", warnings.getvalue())
            self.assertEqual(list((root / "build").glob("prepare-*")), [])

    def test_missing_and_modified_inputs_fail_closed(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            prepare.verify_inputs(root, {"new.go": None})
            (root / "new.go").write_text("unexpected source")
            with self.assertRaisesRegex(RuntimeError, "Patch input mismatch"):
                prepare.verify_inputs(root, {"new.go": None})
            with self.assertRaisesRegex(RuntimeError, "Patch input mismatch"):
                prepare.verify_inputs(root, {"absent.go": "expected"})

    def test_second_dependency_failure_preserves_existing_installations(self):
        from unittest import mock
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            (root / "native").mkdir()
            dependencies = [{"name": "upstream"}, {"name": "typesetting"}]
            (root / "native/dependencies.json").write_text(json.dumps({"dependencies": dependencies}))
            for dependency in dependencies:
                directory = root / "build" / dependency["name"]
                directory.mkdir(parents=True)
                (directory / "go.mod").write_text("previous installation")

            def staged(dependency, destination):
                destination.mkdir()
                (destination / "go.mod").write_text("replacement")
                if dependency["name"] == "typesetting":
                    raise RuntimeError("simulated checksum failure")

            with mock.patch.object(prepare, "ROOT", root), mock.patch.object(prepare, "prepare_dependency", staged):
                with self.assertRaisesRegex(RuntimeError, "checksum failure"):
                    prepare.main()
            for dependency in dependencies:
                self.assertEqual((root / "build" / dependency["name"] / "go.mod").read_text(), "previous installation")
            self.assertFalse((root / "build/.prepare-lock").exists())

    def test_partial_installation_failure_restores_both_prior_directories(self):
        from unittest import mock
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            dependencies = self.previous_installations(root)
            rename = Path.rename

            def install_failure(source, destination):
                if source.parent.name.startswith("prepare-") and source.name == "typesetting":
                    raise OSError("simulated second installation failure")
                return rename(source, destination)

            with mock.patch.object(prepare, "ROOT", root), mock.patch.object(prepare, "prepare_dependency", self.staged_installation), mock.patch.object(Path, "rename", install_failure):
                with self.assertRaisesRegex(OSError, "second installation failure"):
                    prepare.main()
            for dependency in dependencies:
                directory = root / "build" / dependency["name"]
                self.assertEqual((directory / "go.mod").read_text(), "previous " + dependency["name"])
                self.assertEqual((directory / "local-data/notes.txt").read_text(), "retain original local data")
            self.assertFalse((root / "build/.prepare-lock").exists())
            self.assertEqual(list((root / "build").glob("prepare-*")), [])

    def test_failed_restore_preserves_recoverable_original_backup(self):
        from unittest import mock
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            self.previous_installations(root)
            rename = Path.rename

            def restore_failure(source, destination):
                if source.parent.name.startswith("prepare-") and source.name == "typesetting":
                    raise OSError("simulated second installation failure")
                if source.name == "upstream.previous":
                    raise OSError("simulated restore failure")
                return rename(source, destination)

            with mock.patch.object(prepare, "ROOT", root), mock.patch.object(prepare, "prepare_dependency", self.staged_installation), mock.patch.object(Path, "rename", restore_failure), mock.patch.object(prepare, "cleanup_staging_directory") as cleanup:
                with self.assertRaisesRegex(RuntimeError, "rollback incomplete") as failed:
                    prepare.main()
                cleanup.assert_not_called()
            recovery, = (root / "build").glob("prepare-*")
            backup = recovery / "upstream.previous"
            self.assertIn(str(recovery), str(failed.exception))
            self.assertIn(str(backup), str(failed.exception))
            self.assertEqual((backup / "go.mod").read_text(), "previous upstream")
            self.assertEqual((backup / "local-data/notes.txt").read_text(), "retain original local data")
            self.assertEqual((root / "build/typesetting/go.mod").read_text(), "previous typesetting")
            self.assertEqual((root / "build/typesetting/local-data/notes.txt").read_text(), "retain original local data")
            self.assertFalse((root / "build/upstream").exists())
            self.assertFalse((root / "build/.prepare-lock").exists())

    def test_failed_replacement_removal_keeps_original_backup(self):
        from unittest import mock
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            self.previous_installations(root)
            rename, remove = Path.rename, prepare.shutil.rmtree

            def install_failure(source, destination):
                if source.parent.name.startswith("prepare-") and source.name == "typesetting":
                    raise OSError("simulated second installation failure")
                return rename(source, destination)

            def removal_failure(path, *args, **kwargs):
                if Path(path) == root / "build/upstream":
                    raise OSError("simulated removal failure")
                return remove(path, *args, **kwargs)

            with mock.patch.object(prepare, "ROOT", root), mock.patch.object(prepare, "prepare_dependency", self.staged_installation), mock.patch.object(Path, "rename", install_failure), mock.patch.object(prepare.shutil, "rmtree", removal_failure):
                with self.assertRaisesRegex(RuntimeError, "rollback incomplete") as failed:
                    prepare.main()
            recovery, = (root / "build").glob("prepare-*")
            backup = recovery / "upstream.previous"
            self.assertIn(str(backup), str(failed.exception))
            self.assertEqual((backup / "go.mod").read_text(), "previous upstream")
            self.assertEqual((backup / "local-data/notes.txt").read_text(), "retain original local data")
            self.assertEqual((root / "build/upstream/go.mod").read_text(), "replacement upstream")
            self.assertEqual((root / "build/typesetting/go.mod").read_text(), "previous typesetting")
            self.assertFalse((root / "build/.prepare-lock").exists())

    def test_interruption_after_backup_rename_preserves_unjournaled_original(self):
        from unittest import mock
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            self.previous_installations(root)
            rename = Path.rename

            def interrupted_rename(source, destination):
                result = rename(source, destination)
                if source == root / "build/upstream":
                    raise KeyboardInterrupt("interrupted after original moved")
                return result

            with mock.patch.object(prepare, "ROOT", root), mock.patch.object(prepare, "prepare_dependency", self.staged_installation), mock.patch.object(Path, "rename", interrupted_rename):
                with self.assertRaisesRegex(RuntimeError, "rollback incomplete") as failed:
                    prepare.main()
            recovery, = (root / "build").glob("prepare-*")
            backup = recovery / "upstream.previous"
            self.assertIn(str(backup), str(failed.exception))
            self.assertEqual((backup / "go.mod").read_text(), "previous upstream")
            self.assertEqual((backup / "local-data/notes.txt").read_text(), "retain original local data")
            self.assertEqual((root / "build/typesetting/go.mod").read_text(), "previous typesetting")
            self.assertFalse((root / "build/.prepare-lock").exists())

    def test_manifest_pin_and_patch_order(self):
        manifest = json.loads((ROOT / "native/dependencies.json").read_text())
        core, font = manifest["dependencies"]
        self.assertEqual(core["module"], "cogentcore.org/core")
        numbers = [int(Path(patch["file"]).name[:4]) for stage in core["stages"] for patch in stage["patches"]]
        self.assertEqual(numbers, list(range(1, 90)))
        font_numbers = [int(Path(patch["file"]).name[:4]) for stage in font["stages"] for patch in stage["patches"]]
        self.assertEqual(font_numbers, list(range(1, 6)))
        import hashlib
        for dependency in manifest["dependencies"]:
            self.assertTrue(dependency["sum"].startswith("h1:"))
            for stage in dependency["stages"]:
                for asset in stage.get("assets", []):
                    self.assertEqual(hashlib.sha256((ROOT / asset["source"]).read_bytes()).hexdigest(), asset["sha256"])
                for patch in stage["patches"]:
                    self.assertEqual(hashlib.sha256((ROOT / patch["file"]).read_bytes()).hexdigest(), patch["sha256"])


class AssetOverlayChecks(unittest.TestCase):
    def fixture(self, root):
        import hashlib
        owned, destination = root / "owned", root / "destination"
        owned.mkdir(); destination.mkdir()
        inputs, outputs, assets = {}, {}, []
        for name in ("a.svg", "b.svg"):
            old, new = ("old-" + name).encode(), ("new-" + name).encode()
            (destination / name).write_bytes(old); (owned / name).write_bytes(new)
            inputs[name] = hashlib.sha256(old).hexdigest()
            outputs[name] = hashlib.sha256(new).hexdigest()
            assets.append({"source": name, "target": name, "sha256": outputs[name]})
        inputs["version.txt"] = hashlib.sha256(b"original").hexdigest()
        outputs["version.txt"] = inputs["version.txt"]
        (destination / "version.txt").write_bytes(b"original")
        stage = {"inputs": "inputs.json", "outputs": "outputs.json", "patches": [], "assets": assets}
        self.save(owned, inputs, outputs)
        return owned, destination, stage, inputs, outputs

    def save(self, owned, inputs, outputs):
        (owned / "inputs.json").write_text(json.dumps(inputs))
        (owned / "outputs.json").write_text(json.dumps(outputs))

    def unchanged(self, destination):
        for name in ("a.svg", "b.svg"):
            self.assertEqual((destination / name).read_bytes(), ("old-" + name).encode())
        self.assertEqual((destination / "version.txt").read_bytes(), b"original")

    def test_exact_asset_replacement_leaves_sources_and_no_backups(self):
        from unittest import mock
        with tempfile.TemporaryDirectory() as tmp:
            owned, dest, stage, inputs, outputs = self.fixture(Path(tmp))
            with mock.patch.object(prepare, "ROOT", owned): prepare.apply_stage(stage, dest)
            for name in ("a.svg", "b.svg"):
                self.assertEqual((dest / name).read_bytes(), (owned / name).read_bytes())
            self.assertEqual(sorted(p.name for p in dest.iterdir()), ["a.svg", "b.svg", "version.txt"])

    def test_late_source_or_target_mismatch_never_patches_or_writes_first_target(self):
        from unittest import mock
        for changed in ("source", "target"):
            with self.subTest(changed=changed), tempfile.TemporaryDirectory() as tmp:
                owned, dest, stage, inputs, outputs = self.fixture(Path(tmp))
                path = (owned if changed == "source" else dest) / "b.svg"
                path.write_bytes(b"unrecognized data")
                with mock.patch.object(prepare, "ROOT", owned), mock.patch.object(prepare.subprocess, "run") as run:
                    with self.assertRaisesRegex(RuntimeError, "checksum mismatch"): prepare.apply_stage(stage, dest)
                run.assert_not_called()
                self.assertEqual((dest / "a.svg").read_bytes(), b"old-a.svg")
                self.assertEqual((dest / "version.txt").read_bytes(), b"original")
                if changed == "target": self.assertEqual((dest / "b.svg").read_bytes(), b"unrecognized data")

    def test_every_target_requires_original_and_result_pins(self):
        from unittest import mock
        for records in ("input", "output", "wrong_output", "missing_outputs"):
            with self.subTest(records=records), tempfile.TemporaryDirectory() as tmp:
                owned, dest, stage, inputs, outputs = self.fixture(Path(tmp))
                if records == "input": inputs["b.svg"] = None
                elif records == "output": outputs.pop("b.svg")
                elif records == "wrong_output": outputs["b.svg"] = "0" * 64
                else: stage.pop("outputs")
                self.save(owned, inputs, outputs)
                with mock.patch.object(prepare, "ROOT", owned):
                    with self.assertRaisesRegex(RuntimeError, "exact old input and new output pins"): prepare.apply_stage(stage, dest)
                self.unchanged(dest)

    def test_noncanonical_asset_paths_refuse_before_payload_reads(self):
        from unittest import mock
        for field in ("source", "target"):
            for value in ("../outside", "/outside", "a\\b", "", "./a.svg", "a//b", "a/../b", "a\nb"):
                with self.subTest(field=field, value=value), tempfile.TemporaryDirectory() as tmp:
                    owned, dest, stage, inputs, outputs = self.fixture(Path(tmp)); stage["assets"][1][field] = value
                    with mock.patch.object(prepare, "ROOT", owned), mock.patch.object(prepare, "_read_stage_file") as read:
                        with self.assertRaisesRegex(RuntimeError, "Invalid stage asset path"): prepare.apply_stage(stage, dest)
                    read.assert_not_called(); self.unchanged(dest)

    def test_symlinked_and_special_sources_or_targets_refuse(self):
        import os
        from unittest import mock
        for field in ("source", "target"):
            for kind in ("file-link", "parent-link", "fifo"):
                if kind == "fifo" and not hasattr(os, "mkfifo"): continue
                with self.subTest(field=field, kind=kind), tempfile.TemporaryDirectory() as tmp:
                    owned, dest, stage, inputs, outputs = self.fixture(Path(tmp)); base = owned if field == "source" else dest
                    path = base / "b.svg"; path.rename(base / "original-b.svg")
                    if kind == "file-link": path.symlink_to(base / "original-b.svg")
                    elif kind == "fifo": os.mkfifo(path)
                    else:
                        (base / "real").mkdir(); (base / "real/b.svg").write_bytes((base / "original-b.svg").read_bytes()); (base / "link").symlink_to(base / "real", target_is_directory=True)
                        stage["assets"][1][field] = "link/b.svg"
                        if field == "target": inputs["link/b.svg"] = inputs["b.svg"]; outputs["link/b.svg"] = outputs["b.svg"]
                    self.save(owned, inputs, outputs)
                    with mock.patch.object(prepare, "ROOT", owned):
                        with self.assertRaisesRegex(RuntimeError, "regular file without symlinks"): prepare.apply_stage(stage, dest)
                    self.assertEqual((dest / "a.svg").read_bytes(), b"old-a.svg")
                    self.assertEqual((dest / "version.txt").read_bytes(), b"original")

    def test_asset_count_member_and_aggregate_limits_precede_mutation(self):
        from unittest import mock
        for limit in ("count", "member", "aggregate"):
            with self.subTest(limit=limit), tempfile.TemporaryDirectory() as tmp:
                owned, dest, stage, inputs, outputs = self.fixture(Path(tmp))
                name, value = {"count": ("MAX_STAGE_ASSETS", 1), "member": ("MAX_ASSET_BYTES", 1), "aggregate": ("MAX_STAGE_ASSET_BYTES", 10)}[limit]
                with mock.patch.object(prepare, "ROOT", owned), mock.patch.object(prepare, name, value):
                    with self.assertRaisesRegex(RuntimeError, "bound exceeded|exceeds byte bound"): prepare.apply_stage(stage, dest)
                self.unchanged(dest)

    def test_duplicate_alias_and_patch_overlap_reject_before_writes(self):
        import os, hashlib
        from unittest import mock
        for kind in ("duplicate", "case-alias", "source-link", "target-link", "cross-link", "patch"):
            with self.subTest(kind=kind), tempfile.TemporaryDirectory() as tmp:
                owned, dest, stage, inputs, outputs = self.fixture(Path(tmp))
                if kind == "duplicate": stage["assets"].append(dict(stage["assets"][0]))
                elif kind == "case-alias": stage["assets"][1]["target"] = "A.SVG"
                elif kind == "source-link":
                    (owned / "b.svg").unlink(); os.link(owned / "a.svg", owned / "b.svg"); outputs["b.svg"] = outputs["a.svg"]; stage["assets"][1]["sha256"] = outputs["a.svg"]
                elif kind == "target-link":
                    (dest / "b.svg").unlink(); os.link(dest / "a.svg", dest / "b.svg"); inputs["b.svg"] = inputs["a.svg"]
                elif kind == "cross-link":
                    (dest / "a.svg").unlink(); os.link(owned / "a.svg", dest / "a.svg"); inputs["a.svg"] = outputs["a.svg"]
                else:
                    body = b"--- a/a.svg\n+++ b/a.svg\n@@ -1 +1 @@\n-old-a.svg\n+different\n"; (owned / "overlap.patch").write_bytes(body)
                    stage["patches"] = [{"file": "overlap.patch", "sha256": hashlib.sha256(body).hexdigest()}]
                self.save(owned, inputs, outputs)
                before = {p.name: p.read_bytes() for p in dest.iterdir()}
                with mock.patch.object(prepare, "ROOT", owned), mock.patch.object(prepare.subprocess, "run") as run:
                    with self.assertRaisesRegex(RuntimeError, "Duplicate|aliases|overlap|same path"): prepare.apply_stage(stage, dest)
                run.assert_not_called(); self.assertEqual(before, {p.name: p.read_bytes() for p in dest.iterdir()})

    def test_target_drift_after_admission_refuses_before_first_write(self):
        from unittest import mock
        with tempfile.TemporaryDirectory() as tmp:
            owned, dest, stage, inputs, outputs = self.fixture(Path(tmp))
            with mock.patch.object(prepare, "ROOT", owned), prepare.stage_assets(stage, dest, inputs, outputs, []) as snapshots:
                (dest / "b.svg").write_bytes(b"changed after admission")
                with self.assertRaisesRegex(RuntimeError, "drift before overlay"): prepare._write_stage_assets(snapshots)
            self.assertEqual((dest / "a.svg").read_bytes(), b"old-a.svg")
            self.assertEqual((dest / "b.svg").read_bytes(), b"changed after admission")

    def test_noncanonical_patch_formats_and_timestamped_overlap_refuse_before_execution(self):
        import hashlib
        from unittest import mock
        cases = [
            b"--- old/a.svg\n+++ new/a.svg\n@@ -1 +1 @@\n-old\n+new\n",
            b"--- a/a.svg\t2026-10-19 10:00:00\n+++ b/a.svg\t2026-10-19 10:01:00\n@@ -1 +1 @@\n-old\n+new\n",
            b"--- a/a.svg\n+++ b/A.SVG\n@@ -1 +1 @@\n-old\n+new\n",
            b"*** old/a.svg\n--- new/a.svg\n***************\n*** 1 ****\n-old\n--- 1 ----\n+new\n",
            b"--- a/version.txt\n+++ b/version.txt\n@@ -1 +1 @@\n-original\n+updated\n1c1\n< old\n---\n> new\n",
            b"--- a/../a.svg\n+++ b/../a.svg\n@@ -1 +1 @@\n-old\n+new\n",
            b"--- a/a.svg\n+++ b/a.svg\n@@ -1,2 +1 @@\n-old\n+new\n",
        ]
        for body in cases:
            with self.subTest(body=body), tempfile.TemporaryDirectory() as tmp:
                owned, dest, stage, inputs, outputs = self.fixture(Path(tmp))
                (owned / "overlap.patch").write_bytes(body)
                stage["patches"] = [{"file": "overlap.patch", "sha256": hashlib.sha256(body).hexdigest()}]
                with mock.patch.object(prepare, "ROOT", owned), mock.patch.object(prepare.subprocess, "run") as run:
                    with self.assertRaises(RuntimeError): prepare.apply_stage(stage, dest)
                run.assert_not_called(); self.unchanged(dest)

    def test_canonical_patch_parser_counts_payload_and_created_deleted_files(self):
        body = (b"--- a/version.txt\told timestamp\n+++ b/version.txt\tnew timestamp\n"
                b"@@ -1,2 +1,2 @@ heading\n--- old header-shaped payload\n+++ new header-shaped payload\n context\n"
                b"--- /dev/null\n+++ b/created.txt\n@@ -0,0 +1 @@\n+new\n\\ No newline at end of file\n"
                b"--- a/deleted.txt\n+++ /dev/null\n@@ -1 +0,0 @@\n-old\n")
        self.assertEqual(prepare._asset_stage_patch_targets(body), {"version.txt", "created.txt", "deleted.txt"})

    def test_admitted_source_and_patch_snapshots_are_used(self):
        import hashlib
        from unittest import mock
        with tempfile.TemporaryDirectory() as tmp:
            owned, dest, stage, inputs, outputs = self.fixture(Path(tmp))
            body = b"--- a/version.txt\n+++ b/version.txt\n@@ -1 +1 @@\n-original\n+updated\n"; (owned / "change.patch").write_bytes(body)
            stage["patches"] = [{"file": "change.patch", "sha256": hashlib.sha256(body).hexdigest()}]; outputs["version.txt"] = hashlib.sha256(b"updated").hexdigest(); self.save(owned, inputs, outputs)
            def patch(argv, **kwargs):
                self.assertEqual(kwargs["input"], body)
                (owned / "a.svg").write_bytes(b"unadmitted new source"); (owned / "change.patch").write_bytes(b"unadmitted patch")
                (dest / "version.txt").write_bytes(b"updated")
            with mock.patch.object(prepare, "ROOT", owned), mock.patch.object(prepare.subprocess, "run", side_effect=patch): prepare.apply_stage(stage, dest)
            self.assertEqual((dest / "a.svg").read_bytes(), b"new-a.svg")
            self.assertEqual((dest / "version.txt").read_bytes(), b"updated")

    def test_partial_write_failure_propagates_without_creating_backups(self):
        import os
        from unittest import mock
        with tempfile.TemporaryDirectory() as tmp:
            owned, dest, stage, inputs, outputs = self.fixture(Path(tmp)); write = os.write; calls = []
            def failed(fd, data):
                calls.append(fd)
                if len(calls) > 1: raise OSError("simulated write failure")
                return write(fd, data)
            with mock.patch.object(prepare, "ROOT", owned), mock.patch.object(prepare.os, "write", side_effect=failed):
                with self.assertRaisesRegex(OSError, "simulated write failure"): prepare.apply_stage(stage, dest)
            self.assertEqual((owned / "a.svg").read_bytes(), b"new-a.svg")
            self.assertEqual(sorted(p.name for p in dest.iterdir()), ["a.svg", "b.svg", "version.txt"])
            for fd in calls:
                with self.assertRaises(OSError): os.fstat(fd)

    def test_bounded_reader_rejects_growth_and_short_reads(self):
        from unittest import mock
        original = prepare.os.fdopen
        for growing in (False, True):
            with self.subTest(growing=growing), tempfile.TemporaryDirectory() as tmp:
                path = Path(tmp) / "source"; path.write_bytes(b"ab"); seen = []
                class Reader:
                    def __init__(self, stream): self.stream = stream; self.requests = []
                    def __enter__(self): return self
                    def __exit__(self, *args): self.stream.close()
                    def fileno(self): return self.stream.fileno()
                    def read(self, n):
                        self.requests.append(n)
                        if len(self.requests) > 2: raise AssertionError("unbounded growing read")
                        return b"x" * n if growing else b""
                def opened(fd, *args):
                    stream = Reader(original(fd, "rb")); seen.append(stream); return stream
                with mock.patch.object(prepare.os, "fdopen", side_effect=opened):
                    with self.assertRaisesRegex(RuntimeError, "grew|shortened"): prepare._read_stage_file(path, 10)
                self.assertLessEqual(sum(seen[0].requests), 3); self.assertTrue(seen[0].stream.closed)

    def test_all_patch_digests_are_preflighted_before_stage_mutation(self):
        import hashlib
        from unittest import mock
        with tempfile.TemporaryDirectory() as tmp:
            owned, dest, stage, inputs, outputs = self.fixture(Path(tmp))
            for name in ("first.patch", "last.patch"):
                body = b"fixture patch"; (owned / name).write_bytes(body)
                stage["patches"].append({"file": name, "sha256": hashlib.sha256(body).hexdigest()})
            (owned / "last.patch").write_bytes(b"changed last patch")
            with mock.patch.object(prepare, "ROOT", owned), mock.patch.object(prepare.subprocess, "run") as run:
                with self.assertRaisesRegex(RuntimeError, "Patch checksum mismatch"): prepare.apply_stage(stage, dest)
            run.assert_not_called(); self.unchanged(dest)


if __name__ == "__main__":
    unittest.main()
