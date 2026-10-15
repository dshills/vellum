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
        self.assertEqual(numbers, list(range(1, 68)))
        self.assertEqual(len(font["stages"]), 4)
        import hashlib
        for dependency in manifest["dependencies"]:
            self.assertTrue(dependency["sum"].startswith("h1:"))
            for stage in dependency["stages"]:
                for patch in stage["patches"]:
                    self.assertEqual(hashlib.sha256((ROOT / patch["file"]).read_bytes()).hexdigest(), patch["sha256"])


if __name__ == "__main__":
    unittest.main()
