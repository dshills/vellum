"""Release gates and deterministic artifact construction without a native display."""
import importlib.util
from contextlib import nullcontext
import errno
import os
import subprocess
import gzip
import io
import json
from pathlib import Path
import tarfile
import tempfile
import unittest

import artifact_io
from unittest import mock

spec = importlib.util.spec_from_file_location("release_sdk", Path(__file__).with_name("release-sdk.py"))
release = importlib.util.module_from_spec(spec)
spec.loader.exec_module(release)

class ReleaseChecks(unittest.TestCase):
    def test_public_release_fails_closed_without_owner_choices(self):
        with tempfile.TemporaryDirectory() as tmp:
            with self.assertRaisesRegex(ValueError,"namespace.*license.*audit"):
                with mock.patch.object(release,"source_state",return_value={"clean":False}):
                    release.release_gates(Path(tmp), {"module":"example.com/vellum"}, True)
            release.release_gates(Path(tmp), {}, False)

    def test_git_provenance_ignores_ambient_repository_index_and_config_redirection(self):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp)/"source";root.mkdir()
            env=release.child_environment()
            subprocess.run(["git","init","--quiet"],cwd=root,env=env,check=True)
            source=root/"api.go";source.write_text("package api\n")
            subprocess.run(["git","add","api.go"],cwd=root,env=env,check=True)
            subprocess.run(["git","-c","user.name=Fixture","-c","user.email=fixture@example.invalid","commit","--quiet","-m","fixture"],cwd=root,env=env,check=True)
            expected=release.source_state(root)
            self.assertTrue(expected["clean"])
            dependencies={}
            for name in ("upstream","typesetting"):
                path=Path(tmp)/name;path.mkdir();(path/"go.mod").write_text("module fixture\n");dependencies[name]=path
            hostile={"GIT_DIR":str(Path(tmp)/"missing.git"),"GIT_WORK_TREE":str(Path(tmp)/"elsewhere"),"GIT_INDEX_FILE":str(Path(tmp)/"wrong-index"),"GIT_CONFIG_COUNT":"1","GIT_CONFIG_KEY_0":"core.worktree","GIT_CONFIG_VALUE_0":str(Path(tmp)/"wrong-tree")}
            with mock.patch.dict(release.os.environ,hostile):
                self.assertEqual(release.source_state(root),expected)
                self.assertIn("api.go",dict(release.source_files(root,dependencies=dependencies)))
                source.write_text("package changed\n")
                self.assertFalse(release.source_state(root)["clean"])
            with mock.patch.object(release,"check_child_output",return_value=b"") as command:
                release.git_output(root,["status"])
                self.assertEqual(command.call_args.kwargs["timeout"],release.COMMAND_TIMEOUT_SECONDS)
                self.assertFalse(any(name.startswith("GIT_") for name in command.call_args.kwargs["env"]))

    def test_concatenated_module_records(self):
        self.assertEqual(release.decode_objects('{"Path":"one"}\n{"Path":"two"}\n'), [{"Path":"one"},{"Path":"two"}])

    def test_archive_is_deterministic_and_refuses_overwrite(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)/"source";root.mkdir()
            (root/"release").mkdir()
            (root/"release/version.json").write_text(json.dumps({"version":"v0.1.0-alpha.1"}))
            source = root/"api.go"
            source.write_text("package api\n")
            with mock.patch.object(release,"source_files",return_value=[("api.go",source)]), mock.patch.object(release,"license_inventory",return_value=({"modules":[]},{})), mock.patch.object(release,"source_state",return_value={"clean":True,"commit":"fixture"}):
                first=release.package(root,Path(tmp)/"one.tar.gz")
                second=release.package(root,Path(tmp)/"two.tar.gz")
                self.assertEqual(first["sha256"],second["sha256"])
                if os.name != "nt":self.assertEqual((Path(tmp)/"one.tar.gz").stat().st_mode & 0o777,0o644)
                with self.assertRaisesRegex(ValueError,"exists"):
                    release.package(root,Path(tmp)/"one.tar.gz")
            with tarfile.open(Path(tmp)/"one.tar.gz") as archive:
                member=archive.getmember("vellum-v0.1.0-alpha.1/api.go")
                self.assertEqual(member.uid,0)
                self.assertEqual(member.mtime,0)
                self.assertTrue(archive.getmember("vellum-v0.1.0-alpha.1/PRIVATE_BUILD.txt"))

    def test_version_cannot_escape_archive_prefix(self):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp)/"source";root.mkdir();(root/"release").mkdir()
            (root/"release/version.json").write_text('{"version":"v../../escape"}')
            with self.assertRaisesRegex(ValueError,"version"):
                release.package(root,Path(tmp)/"output")


    def test_generated_records_are_all_checksummed(self):
        import hashlib
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp)/"source";root.mkdir();(root/"release").mkdir();(root/"release/version.json").write_text('{"version":"v0.1.0-alpha.1"}')
            source=root/"api.go";source.write_text("package api\n")
            notices={"THIRD_PARTY_NOTICES/dependency/LICENSE":b"notice"}
            with mock.patch.object(release,"source_files",return_value=[("api.go",source)]),mock.patch.object(release,"license_inventory",return_value=({"modules":[]},notices)),mock.patch.object(release,"source_state",return_value={"clean":False,"commit":"fixture"}):
                release.package(root,Path(tmp)/"sdk.tar.gz")
            with tarfile.open(Path(tmp)/"sdk.tar.gz") as archive:
                prefix="vellum-v0.1.0-alpha.1/"
                manifest=json.loads(archive.extractfile(prefix+"SDK_MANIFEST.json").read())
                files={member.name[len(prefix):] for member in archive if member.name!=prefix+"SDK_MANIFEST.json"}
                self.assertEqual(files,set(manifest["files"]))
                self.assertFalse(manifest["source"]["clean"])
                for name,expected in manifest["files"].items():
                    self.assertEqual(hashlib.sha256(archive.extractfile(prefix+name).read()).hexdigest(),expected)

    def test_public_release_refuses_dirty_source(self):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp);(root/"LICENSE").write_text("owner license");(root/"go.mod").write_text("module github.com/owner/vellum\n")
            metadata={"module":"github.com/owner/vellum","namespace_status":"selected","license_status":"selected","third_party_distribution_audit":"approved"}
            with mock.patch.object(release,"source_state",return_value={"clean":False}):
                with self.assertRaisesRegex(ValueError,"committed and clean"):release.release_gates(root,metadata,True)
            with mock.patch.object(release,"source_state",return_value={"clean":True}),mock.patch.object(release,"hidden_index_paths",return_value=[]):release.release_gates(root,metadata,True)

    def test_public_source_change_prevents_publication_and_cleans_temporary(self):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp)/"source";root.mkdir();(root/"release").mkdir()
            metadata={"version":"v0.1.0-alpha.1","module":"github.com/owner/vellum","namespace_status":"selected","license_status":"selected","third_party_distribution_audit":"approved"}
            (root/"release/version.json").write_text(json.dumps(metadata))
            (root/"LICENSE").write_text("fixture license")
            (root/"go.mod").write_text("module github.com/owner/vellum\n")
            source=root/"api.go";source.write_text("package api\n")
            states=[{"clean":True,"commit":"before"},{"clean":True,"commit":"before"},{"clean":True,"commit":"after"}]
            with mock.patch.object(release,"hidden_index_paths",return_value=[]),mock.patch.object(release,"source_state",side_effect=states),mock.patch.object(release,"source_files",return_value=[("api.go",source)]),mock.patch.object(release,"license_inventory",return_value=({"modules":[]},{})),mock.patch.object(release,"public_dependencies",return_value=nullcontext(None)):
                with self.assertRaisesRegex(ValueError,"changed during packaging"):
                    release.package(root,Path(tmp)/"sdk.tar.gz",public=True)
            self.assertFalse((Path(tmp)/"sdk.tar.gz").exists())
            self.assertEqual(list(Path(tmp).glob(".vellum-sdk-*")),[])

    def test_source_whitelist_covers_root_api_and_refuses_links(self):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp);(root/"vellum.go").write_text("package vellum")
            for dependency in ("upstream","typesetting"):
                directory=root/"drivers/cogent/build"/dependency;directory.mkdir(parents=True);(directory/"go.mod").write_text("module source")
            with mock.patch.object(release,"check_child_output",return_value=b"vellum.go\0"):
                self.assertIn("vellum.go",[name for name,_ in release.source_files(root)])
            (root/"link").symlink_to(root/"drivers",target_is_directory=True)
            with self.assertRaisesRegex(ValueError,"symlinks"):release.require_regular_source(root,root/"link/cogent/build/upstream/go.mod")

    def test_legitimate_nested_build_package_is_not_excluded(self):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp);source=root/"examples/build/api.go";source.parent.mkdir(parents=True);source.write_text("package build")
            dependencies={}
            for name in ("upstream","typesetting"):
                path=root/"fresh"/name;path.mkdir(parents=True);(path/"go.mod").write_text("module fresh");dependencies[name]=path
            with mock.patch.object(release,"check_child_output",return_value=b"examples/build/api.go\0drivers/cogent/build/generated.log\0"):
                files=dict(release.source_files(root,dependencies=dependencies))
            self.assertIn("examples/build/api.go",files)
            self.assertNotIn("drivers/cogent/build/generated.log",files)
            self.assertEqual(files["drivers/cogent/build/upstream/go.mod"],dependencies["upstream"]/"go.mod")

    def public_dependency_fixture(self,root):
        driver=root/"drivers/cogent";(driver/"native").mkdir(parents=True)
        (root/"go.mod").write_text("module example.com/fixture\ngo 1.26.1\n")
        entries=[{"name":"upstream","module":"cogentcore.org/core","version":"v0.3.0","sum":"h1:core"},{"name":"typesetting","module":"github.com/go-text/typesetting","version":"v0.3.0","sum":"h1:text"}]
        (driver/"native/dependencies.json").write_text(json.dumps({"dependencies":entries}))
        (driver/"prepare.py").write_text("from pathlib import Path\ndef prepare_dependency(entry,destination):\n destination.mkdir()\n (destination/'go.mod').write_text('module '+entry['module'])\n (Path(__file__).parent/'prepare-called').write_text('yes')\n")
        return driver,entries

    def test_public_dependencies_are_fresh_and_verified_before_and_after_copy(self):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp);driver,entries=self.public_dependency_fixture(root)
            existing=driver/"build/upstream";existing.mkdir(parents=True);(existing/"go.mod").write_text("modified ignored tree")
            def download(args,**kwargs):
                entry=next(item for item in entries if args[-1]==item["module"]+"@"+item["version"])
                return json.dumps({"Sum":entry["sum"]})
            original_run=subprocess.run
            def run(arguments,**kwargs):
                if arguments[0]==release.sys.executable:return original_run(arguments,**kwargs)
            with mock.patch.object(release,"check_child_output",side_effect=download),mock.patch.object(release,"run_child",side_effect=run) as commands:
                with release.public_dependencies(root) as paths:
                    self.assertNotEqual(paths["upstream"],existing)
                    self.assertEqual((paths["upstream"]/"go.mod").read_text(),"module cogentcore.org/core")
                    verification=[call for call in commands.call_args_list if call.args[0]==["go","mod","verify"]]
                    bootstrap=[call for call in commands.call_args_list if call.args[0][0]==release.sys.executable]
                    self.assertEqual(len(verification),2);self.assertEqual(len(bootstrap),2)
                    for call in commands.call_args_list:
                        self.assertEqual(call.kwargs["env"]["GOENV"],"off")
                        self.assertEqual(call.kwargs["env"]["GOFLAGS"],"-mod=readonly")
                        self.assertEqual(call.kwargs["env"]["GONOSUMDB"],"")
                    fresh=paths["upstream"]
            self.assertFalse(fresh.exists())
            self.assertEqual((existing/"go.mod").read_text(),"modified ignored tree")

    def test_public_cached_module_verification_failure_stops_regeneration(self):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp);driver,entries=self.public_dependency_fixture(root)
            outputs=[json.dumps({"Sum":entry["sum"]}) for entry in entries]
            with mock.patch.object(release,"check_child_output",side_effect=outputs),mock.patch.object(release,"run_child",side_effect=subprocess.CalledProcessError(1,["go","mod","verify"])):
                with self.assertRaises(subprocess.CalledProcessError):
                    with release.public_dependencies(root):self.fail("invalid module cache was accepted")
            self.assertFalse((driver/"prepare-called").exists())

    def test_exclusive_publication_falls_back_and_preserves_existing_outputs(self):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp);source=root/"source";source.write_bytes(b"complete artifact");source.chmod(0o600);output=root/"output"
            actual_copy=artifact_io.shutil.copyfileobj
            def inspect_mode(original,destination):
                if os.name != "nt":self.assertEqual(os.fstat(destination.fileno()).st_mode & 0o777,0o600)
                return actual_copy(original,destination)
            with mock.patch.object(artifact_io.os,"link",side_effect=OSError(errno.ENOTSUP,"no hardlinks")),mock.patch.object(artifact_io.shutil,"copyfileobj",side_effect=inspect_mode):
                artifact_io.publish_exclusive(source,output)
                self.assertEqual(output.read_bytes(),b"complete artifact")
                if os.name != "nt":self.assertEqual(output.stat().st_mode & 0o777,0o600)
                with self.assertRaises(FileExistsError):artifact_io.publish_exclusive(source,output)
                self.assertEqual(output.read_bytes(),b"complete artifact")
                link=root/"link";link.symlink_to(root/"missing")
                with self.assertRaises(FileExistsError):artifact_io.publish_exclusive(source,link)
                self.assertTrue(link.is_symlink())

    def test_failed_publication_copy_removes_only_owned_partial_output(self):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp);source=root/"source";source.write_bytes(b"artifact");output=root/"output"
            def partial(original,destination):destination.write(b"partial");raise OSError("copy failed")
            with mock.patch.object(artifact_io.os,"link",side_effect=OSError(errno.EPERM,"no hardlinks")),mock.patch.object(artifact_io.shutil,"copyfileobj",side_effect=partial):
                with self.assertRaisesRegex(OSError,"copy failed"):artifact_io.publish_exclusive(source,output)
            self.assertFalse(output.exists());self.assertEqual(source.read_bytes(),b"artifact")

    def test_only_conventional_text_declarations_count_as_notices(self):
        for name in ("LICENSE", "LICENSE.md", "LICENCE.txt", "COPYING.LESSER", "NOTICE-MIT", "third-party/LICENSE-APACHE"):
            self.assertTrue(release.declared_notice(Path(name)),name)
        for name in ("license.go", "license_test.go", "license.py", "NOTICE.json", "testdata/LICENSE", "fixtures/COPYING", "examples/NOTICE", "licensechecker.txt"):
            self.assertFalse(release.declared_notice(Path(name)),name)
        self.assertTrue(release.module_license(Path("LICENSE.md")))
        self.assertTrue(release.module_license(Path("licenses/LICENSE-MIT")))
        self.assertFalse(release.module_license(Path("NOTICE")))
        self.assertFalse(release.module_license(Path("vendor/dependency/LICENSE")))

    def test_notice_inventory_requires_module_own_license(self):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp);(root/"release").mkdir();(root/"release/version.json").write_text('{"module":"example.com/vellum"}')
            dependency=root/"dependency";dependency.mkdir()
            for name in ("license.go", "testdata/LICENSE", "third-party/LICENSE-MIT"):
                path=dependency/name;path.parent.mkdir(parents=True,exist_ok=True);path.write_text("fixture")
            modules=json.dumps({"Path":"example.com/dependency","Dir":str(dependency)})
            with mock.patch.object(release,"check_child_output",return_value=modules):
                inventory,notices=release.license_inventory(root)
                record=inventory["modules"][0]
                self.assertEqual([item["file"] for item in record["license_files"]],["third-party/LICENSE-MIT"])
                self.assertEqual(record["module_license_files"],[])
                self.assertEqual(record["status"],"manual-review-required")
                (dependency/"LICENSE.md").write_text("module declaration")
                inventory,notices=release.license_inventory(root)
                self.assertEqual(inventory["modules"][0]["module_license_files"],["LICENSE.md"])
                self.assertEqual(len(notices),2)

    def test_locally_replaced_native_notice_records_keep_pinned_provenance(self):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp);(root/"release").mkdir();(root/"release/version.json").write_text('{"module":"example.com/vellum"}')
            native=root/"drivers/cogent/native";native.mkdir(parents=True)
            (native/"dependencies.json").write_text(json.dumps({"dependencies":[{"module":"cogentcore.org/core","version":"v0.3.0","sum":"h1:pinned"}]}))
            source=root/"core";source.mkdir();(source/"LICENSE").write_text("notice")
            module={"Path":"cogentcore.org/core","Dir":str(source),"Replace":{"Path":str(source),"Dir":str(source)}}
            with mock.patch.object(release,"check_child_output",return_value=json.dumps(module)):
                inventory,_=release.license_inventory(root)
            self.assertEqual(inventory["modules"][0]["version"],"v0.3.0")
            self.assertEqual(inventory["modules"][0]["sum"],"h1:pinned")

    def test_public_graph_checks_use_controlled_go_environment(self):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp);(root/"release").mkdir();(root/"release/version.json").write_text('{"module":"example.com/vellum"}')
            driver=root/"drivers/cogent";driver.mkdir(parents=True);(driver/"go.mod").write_text("module fixture");(driver/"go.sum").write_text("")
            with mock.patch.dict(release.os.environ,{"GOFLAGS":"-modfile=outside","GOENV":"outside","AWS_SECRET_ACCESS_KEY":"fixture"}),mock.patch.object(release,"run_child") as run,mock.patch.object(release,"check_child_output",return_value="") as listed:
                release.license_inventory(root,{"upstream":root/"core","typesetting":root/"text"})
            self.assertEqual(run.call_args.args[0],["go","mod","verify"])
            for environment in [listed.call_args.kwargs["env"],*[call.kwargs["env"] for call in run.call_args_list]]:
                self.assertEqual(environment["GOFLAGS"],"-mod=readonly")
                self.assertEqual(environment["GOENV"],"off")
                self.assertEqual(environment["GOTOOLCHAIN"],"local")
                self.assertEqual(environment["GOPROXY"],"https://proxy.golang.org,direct")
                self.assertEqual(environment["GOSUMDB"],"sum.golang.org")
                self.assertNotIn("AWS_SECRET_ACCESS_KEY",environment)

    def test_public_packaging_rejects_only_nested_third_party_notices(self):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp)/"source";root.mkdir();(root/"release").mkdir();(root/"release/version.json").write_text('{"version":"v0.1.0-alpha.1"}')
            nested={"modules":[{"license_files":[{"file":"vendor/LICENSE"}],"module_license_files":[]}]}
            with mock.patch.object(release,"release_gates"),mock.patch.object(release,"source_state",return_value={"clean":True}),mock.patch.object(release,"public_dependencies",return_value=nullcontext(None)),mock.patch.object(release,"license_inventory",return_value=(nested,{})):
                with self.assertRaisesRegex(ValueError,"module license declaration"):release.package(root,Path(tmp)/"sdk.tar.gz",public=True)
            self.assertFalse((Path(tmp)/"sdk.tar.gz").exists())
            self.assertEqual(list(Path(tmp).glob(".vellum-sdk-*")),[])

    def test_output_inside_checkout_and_symlinked_ancestor_are_rejected_before_source_reads(self):
        with tempfile.TemporaryDirectory() as tmp:
            parent=Path(tmp);root=parent/"source";root.mkdir()
            alias=parent/"alias";alias.symlink_to(root,target_is_directory=True)
            for output in (root/"scripts/sdk.tar.gz",alias/"sdk.tar.gz"):
                with self.subTest(output=output),mock.patch.object(release,"source_state") as state,mock.patch.object(release,"source_files") as files:
                    with self.assertRaisesRegex(ValueError,"outside the source checkout"):release.package(root,output)
                    state.assert_not_called();files.assert_not_called()
                    self.assertFalse(output.exists())
            self.assertEqual(list(root.iterdir()),[])

    def test_public_gate_rejects_changes_hidden_by_real_git_index_flags(self):
        for flag in ("assume-unchanged","skip-worktree"):
            with self.subTest(flag=flag),tempfile.TemporaryDirectory() as tmp:
                root=Path(tmp);env=release.child_environment()
                subprocess.run(["git","init","--quiet"],cwd=root,env=env,check=True)
                (root/"LICENSE").write_text("fixture license")
                (root/"go.mod").write_text("module github.com/owner/vellum\n")
                source=root/"api.go";source.write_text("package api\n")
                subprocess.run(["git","add","."],cwd=root,env=env,check=True)
                subprocess.run(["git","-c","user.name=Fixture","-c","user.email=fixture@example.invalid","commit","--quiet","-m","fixture"],cwd=root,env=env,check=True)
                subprocess.run(["git","update-index","--"+flag,"api.go"],cwd=root,env=env,check=True)
                source.write_text("package hidden_change\n")
                self.assertTrue(release.source_state(root)["clean"])
                self.assertEqual(release.hidden_index_paths(root),["api.go"])
                metadata={"module":"github.com/owner/vellum","namespace_status":"selected","license_status":"selected","third_party_distribution_audit":"approved"}
                with self.assertRaisesRegex(ValueError,"assume-unchanged or skip-worktree"):release.release_gates(root,metadata,True)

    def test_untracked_private_source_requires_opt_in_and_manifest_records_shipped_names(self):
        with tempfile.TemporaryDirectory() as tmp:
            parent=Path(tmp);root=parent/"source";root.mkdir();env=release.child_environment()
            subprocess.run(["git","init","--quiet"],cwd=root,env=env,check=True)
            (root/"release").mkdir();(root/"release/version.json").write_text('{"version":"v0.1.0-alpha.1"}')
            tracked=root/"api.go";tracked.write_text("package api\n")
            subprocess.run(["git","add","."],cwd=root,env=env,check=True)
            subprocess.run(["git","-c","user.name=Fixture","-c","user.email=fixture@example.invalid","commit","--quiet","-m","fixture"],cwd=root,env=env,check=True)
            untracked=root/"new.go";untracked.write_text("package api\n")
            def files(directory,include_untracked=False,dependencies=None):
                yield "api.go",tracked
                if include_untracked:yield "new.go",untracked
            with mock.patch.object(release,"source_files",side_effect=files),mock.patch.object(release,"license_inventory",return_value=({"modules":[]},{})):
                release.package(root,parent/"tracked.tar.gz")
                release.package(root,parent/"opted-in.tar.gz",include_untracked=True)
                with self.assertRaisesRegex(ValueError,"public SDK cannot include untracked"):release.package(root,parent/"public.tar.gz",public=True,include_untracked=True)
            for name,expected in [("tracked.tar.gz",[]),("opted-in.tar.gz",["new.go"])]:
                with tarfile.open(parent/name) as archive:
                    prefix="vellum-v0.1.0-alpha.1/";manifest=json.load(archive.extractfile(prefix+"SDK_MANIFEST.json"))
                    self.assertEqual(manifest["untracked_files"],expected)
                    self.assertEqual("new.go" in manifest["files"],bool(expected))
                    self.assertFalse(manifest["source"]["clean"])

    def test_dangling_output_symlink_is_refused(self):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp)/"source";root.mkdir();(root/"release").mkdir();(root/"release/version.json").write_text('{"version":"v0.1.0-alpha.1"}')
            output=Path(tmp)/"sdk.tar.gz";output.symlink_to(Path(tmp)/"missing")
            with mock.patch.object(release,"source_state",return_value={"clean":True}):
                with self.assertRaisesRegex(ValueError,"exists"):release.package(root,output)
            self.assertTrue(output.is_symlink());self.assertFalse((Path(tmp)/"missing").exists())

if __name__=="__main__":unittest.main()
