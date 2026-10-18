"""Inventory identities, target inputs and refusal behavior without a desktop."""
import hashlib
import importlib.util
import json
from pathlib import Path
import tempfile
import unittest
from unittest import mock

spec = importlib.util.spec_from_file_location("distribution_inventory", Path(__file__).with_name("distribution-inventory.py"))
inventory = importlib.util.module_from_spec(spec)
spec.loader.exec_module(inventory)


class InventoryChecks(unittest.TestCase):
    def fixture(self, base):
        root = base / "SDK with spaces"
        driver = root / "drivers/cogent"
        native = base / "external native module"
        toolchain = base / "Go installation/libexec"
        for directory in (root, driver / "native", native / "arm64", toolchain): directory.mkdir(parents=True, exist_ok=True)
        (root / "LICENSE").write_bytes(b"framework MIT declaration")
        (driver / "LICENSE").write_bytes(b"driver MIT declaration")
        (driver / "native/dependencies.json").write_text('{"dependencies": []}\n')
        (driver / "font.ttf").write_bytes(b"fake font bytes\x00")
        (driver / "bridge.go").write_bytes(b"package bridge\n")
        (native / "arm64/libfake.a").write_bytes(b"!<arch>\nfixture\x00archive")
        (native / "LICENSE").write_bytes(b"native declaration candidate")
        (toolchain.parent / "LICENSE").write_bytes(b"installed Go declaration candidate")
        (toolchain / "src/runtime").mkdir(parents=True)
        (toolchain / "src/runtime/runtime.go").write_bytes(b"package runtime\n")
        graph = [{"Path": "example.com/framework", "Replace": {"Dir": str(root)}}, {"Path": "example.com/driver", "Main": True, "Dir": "/retired/temporary/graph"}, {"Path": "example.com/native", "Dir": str(native), "Version": "v1.0.0", "Sum": "h1:fixture"}]
        packages = [{"ImportPath": "runtime", "Standard": True, "Dir": str(toolchain / "src/runtime"), "GoFiles": ["runtime.go"]}, {"ImportPath": "example.com/driver", "Module": {"Path": "example.com/driver"}, "Dir": str(driver), "EmbedFiles": ["font.ttf"], "CgoFiles": ["bridge.go"], "CgoLDFLAGS": ["-L", str(native / "arm64"), "-lfake", "-framework", "Cocoa"]}]
        notices = {"schema": 1, "modules": [{"module": "example.com/native", "license_files": [{"file": "LICENSE", "sha256": hashlib.sha256((native / "LICENSE").read_bytes()).hexdigest()}], "module_license_files": ["LICENSE"]}]}
        sources = [(name, root / name) for name in ["LICENSE", "drivers/cogent/LICENSE", "drivers/cogent/bridge.go", "drivers/cogent/font.ttf", "drivers/cogent/native/dependencies.json"]]
        configuration = {"GOOS": "darwin", "GOARCH": "arm64", "CGO_ENABLED": "1", "GOVERSION": "go-fixture", "GOROOT": str(toolchain)}
        return root, driver, native, graph, packages, notices, sources, configuration

    def collect_fixture(self, fixture, package_change=None, source_change=None, **observation):
        root, driver, native, graph, packages, notices, sources, configuration = fixture
        calls = []
        def output(argv, **kwargs):
            calls.append((argv, kwargs))
            if argv[:2] == ["go", "env"]: return json.dumps(configuration)
            if package_change and len(calls) == 3: return "\n".join(json.dumps(p) for p in package_change)
            return "\n".join(json.dumps(p) for p in packages)
        source_results = [iter(sources), iter(source_change if source_change is not None else sources)]
        with mock.patch.object(inventory, "check_child_output", side_effect=output), mock.patch.object(inventory.release, "source_state", return_value={"commit": "fixture", "clean": True}), mock.patch.object(inventory.release, "verified_module_graph", return_value=graph), mock.patch.object(inventory.release, "license_inventory", return_value=(notices, {})), mock.patch.object(inventory.release, "source_files", side_effect=source_results):
            result = inventory.collect(root, **observation)
        self.assertEqual([call[0][:2] for call in calls], [["go", "env"], ["go", "list"], ["go", "list"]])
        return result

    def test_exact_input_identity_private_status_and_relocation(self):
        results = []
        for _ in range(2):
            with tempfile.TemporaryDirectory() as temporary:
                fixture = self.fixture(Path(temporary))
                result = self.collect_fixture(fixture)
                self.assertEqual(result["distribution_audit"], "pending")
                self.assertFalse(result["runtime_executed"])
                self.assertNotIn("GOROOT", result["build_configuration"])
                self.assertEqual(result["toolchain_declaration_candidates"][0]["source_root"], "GOROOT-parent")
                font = result["embedded_assets"][0]
                self.assertEqual(font["file"], "font.ttf")
                self.assertEqual(font["sha256"], hashlib.sha256(b"fake font bytes\x00").hexdigest())
                self.assertEqual(font["nearby_notice_candidates"], ["LICENSE"])
                self.assertEqual(font["rights_status"], "review-required")
                library = result["native_link_candidates"][0]
                self.assertEqual(library["searches"], ["module:example.com/native/arm64"])
                self.assertEqual(library["file_candidates"][0]["sha256"], hashlib.sha256(b"!<arch>\nfixture\x00archive").hexdigest())
                self.assertEqual(library["link_selection"], "unverified")
                self.assertEqual(result["packages"][0]["framework_references"], ["Cocoa"])
                self.assertEqual(result["packages"][0]["framework_resolution"], "unverified")
                self.assertNotIn(temporary, json.dumps(result))
                results.append(result)
        self.assertEqual(results[0], results[1])

    def test_standard_sources_declarations_and_relocation_are_bound(self):
        reports = []
        for _ in range(2):
            with tempfile.TemporaryDirectory() as temporary:
                fixture = self.fixture(Path(temporary)); toolchain = Path(fixture[7]["GOROOT"])
                runtime = toolchain / "src/runtime"
                (toolchain / "PATENTS").write_bytes(b"installed patent declaration")
                (runtime / "NOTICE").write_bytes(b"local declaration candidate")
                (runtime / "asm.s").write_bytes(b"assembly fixture")
                (runtime / "object.syso").write_bytes(b"object fixture")
                (runtime / "embedded.bin").write_bytes(b"embedded fixture")
                package = fixture[4][0]
                package.update(SFiles=["asm.s"], SysoFiles=["object.syso"], EmbedFiles=["embedded.bin"])
                report = self.collect_fixture(fixture)
                standard = report["go_standard_library"]
                self.assertEqual(standard["rights_status"], "review-required")
                self.assertEqual([r["file"] for r in standard["files"]], ["src/runtime/asm.s", "src/runtime/embedded.bin", "src/runtime/object.syso", "src/runtime/runtime.go"])
                self.assertEqual([r["file"] for r in standard["declaration_candidates"]], ["PATENTS", "src/runtime/NOTICE"])
                self.assertEqual(standard["packages"][0]["source_fields"]["SysoFiles"], ["src/runtime/object.syso"])
                self.assertEqual(standard["packages"][0]["nearby_declaration_candidates"], ["PATENTS", "src/runtime/NOTICE"])
                self.assertEqual(standard["files"][-1]["sha256"], hashlib.sha256(b"package runtime\n").hexdigest())
                self.assertIn({"source_root": "GOROOT", "file": "PATENTS", "bytes": 28, "sha256": hashlib.sha256(b"installed patent declaration").hexdigest()}, report["toolchain_declaration_candidates"])
                self.assertNotIn(temporary, json.dumps(report))
                reports.append(report)
        self.assertEqual(reports[0], reports[1])

    def test_goroot_alias_resolves_without_accepting_internal_source_links(self):
        with tempfile.TemporaryDirectory() as temporary:
            fixture = self.fixture(Path(temporary)); toolchain = Path(fixture[7]["GOROOT"])
            alias = Path(temporary) / "Go alias"; alias.symlink_to(toolchain, target_is_directory=True)
            original = inventory.standard_library_inventory([fixture[4][0]], toolchain)
            package = {**fixture[4][0], "Dir": str(alias / "src/runtime")}
            self.assertEqual(original, inventory.standard_library_inventory([package], toolchain.resolve()))

    def test_standard_source_and_metadata_drift_refuse(self):
        with tempfile.TemporaryDirectory() as temporary:
            fixture = self.fixture(Path(temporary)); original = inventory.file_identity
            reads = 0
            def changing(base, name):
                nonlocal reads
                result = original(base, name)
                if name == "src/runtime/runtime.go":
                    reads += 1
                    if reads == 2: result["sha256"] = "changed"
                return result
            with mock.patch.object(inventory, "file_identity", side_effect=changing):
                with self.assertRaisesRegex(ValueError, "input changed"):
                    self.collect_fixture(fixture)
            packages = [dict(p) for p in fixture[4]]
            packages[0]["GoFiles"] = ["different.go"]
            with self.assertRaisesRegex(ValueError, "build metadata changed"):
                self.collect_fixture(fixture, package_change=packages)

    def test_new_standard_declaration_and_changed_notice_refuse(self):
        for create in [False, True]:
            with self.subTest(create=create), tempfile.TemporaryDirectory() as temporary:
                fixture = self.fixture(Path(temporary)); original = inventory.standard_library_inventory
                toolchain = Path(fixture[7]["GOROOT"])
                notice = toolchain / "src/runtime/NOTICE"
                if not create: notice.write_bytes(b"first notice")
                calls = 0
                def change(packages, root):
                    nonlocal calls
                    calls += 1
                    if calls == 2: notice.write_bytes(b"later notice")
                    return original(packages, root)
                with mock.patch.object(inventory, "standard_library_inventory", side_effect=change):
                    with self.assertRaisesRegex(ValueError, "standard-library inputs or declarations changed"):
                        self.collect_fixture(fixture)

    def test_standard_paths_markers_duplicates_and_budgets_refuse(self):
        with tempfile.TemporaryDirectory() as temporary:
            fixture = self.fixture(Path(temporary)); toolchain = Path(fixture[7]["GOROOT"])
            package = fixture[4][0]; method = inventory.standard_library_inventory
            cases = [{"Dir": str(toolchain / "src/../outside")}, {"Dir": str(toolchain.parent)}, {"GoFiles": ["../LICENSE"]}, {"GoFiles": ["a//b"]}, {"GoFiles": ["/LICENSE"]}, {"GoFiles": ["runtime.go", "runtime.go"]}, {"GoFiles": "runtime.go"}, {"Incomplete": True}, {"Standard": "true"}]
            for mutation in cases:
                with self.subTest(mutation=mutation), self.assertRaises(ValueError):
                    method([{**package, **mutation}], toolchain)
            with self.assertRaises(ValueError): method([package, package], toolchain)
            (toolchain / "src/alias").symlink_to(toolchain / "src/runtime", target_is_directory=True)
            with self.assertRaisesRegex(ValueError, "symlinks"):
                method([{**package, "Dir": str(toolchain / "src/alias")}], toolchain)
            (toolchain / "src/loop").symlink_to(toolchain, target_is_directory=True)
            with self.assertRaisesRegex(ValueError, "symlinks"):
                method([{**package, "Dir": str(toolchain / "src/loop/src/runtime")}], toolchain)
            (toolchain / "src/runtime/alias.go").symlink_to(toolchain / "src/runtime/runtime.go")
            with self.assertRaisesRegex(ValueError, "symlinks"):
                method([{**package, "GoFiles": ["alias.go"]}], toolchain)
            with mock.patch.object(inventory, "MAX_FILES", 1), mock.patch.object(inventory, "file_identity", wraps=inventory.file_identity) as identity:
                with self.assertRaisesRegex(ValueError, "file-count bound"):
                    method([{**package, "SFiles": ["unhashed.s"]}], toolchain)
                identity.assert_not_called()
            (toolchain / "PATENTS").write_bytes(b"oversize declaration")
            with mock.patch.object(inventory, "MAX_FILE_BYTES", 4):
                with self.assertRaisesRegex(ValueError, "file bound"):
                    method([{**package, "GoFiles": []}], toolchain)

    def test_missing_native_declaration_or_file_is_an_unresolved_candidate(self):
        with tempfile.TemporaryDirectory() as temporary:
            fixture = self.fixture(Path(temporary)); native = fixture[2]
            roots = {"example.com/driver": fixture[1], "example.com/native": native}
            result = inventory.package_inventory(fixture[4], roots, {})
            self.assertEqual(result["native_link_candidates"][0]["file_candidates"][0]["nearby_notice_candidates"], [])
            (native / "arm64/libfake.a").unlink()
            result = inventory.package_inventory(fixture[4], roots, {})
            self.assertEqual(result["native_link_candidates"][0]["file_candidates"], [])
            self.assertEqual(result["native_link_candidates"][0]["embedded_native_component_provenance"], "unreviewed")
            self.assertEqual(result["unresolved_native_inputs"][0]["reason"], "no-module-file-candidate")

    def test_direct_inputs_and_unsupported_or_external_flags_are_explicit(self):
        with tempfile.TemporaryDirectory() as temporary:
            fixture = self.fixture(Path(temporary)); driver, native = fixture[1:3]
            roots = {"example.com/driver": driver, "example.com/native": native}
            package = dict(fixture[4][-1])
            package["CgoLDFLAGS"] = [str(native / "arm64/libfake.a"), "/outside/libother.dylib", "-Wl,-framework,Custom", "-Wl,-L/outside", "-F", str(native / "Frameworks"), "-F/outside", "-L/outside", "-lother", "-framework", "Custom"]
            result = inventory.package_inventory([package], roots, {})
            self.assertEqual(result["native_link_candidates"][0]["file_candidates"][0]["sha256"], hashlib.sha256(b"!<arch>\nfixture\x00archive").hexdigest())
            self.assertEqual(result["native_link_candidates"][1]["file_candidates"], [])
            reasons = [item["reason"] for item in result["unresolved_native_inputs"]]
            self.assertEqual(reasons.count("unparsed-native-linker-option"), 2)
            self.assertEqual(reasons.count("framework-search-path-not-resolved"), 2)
            self.assertIn("direct-native-file-missing-or-outside-verified-modules", reasons)
            self.assertIn("search-path-outside-verified-modules", reasons)
            self.assertIn("no-module-file-candidate", reasons)
            self.assertEqual(result["packages"][0]["framework_references"], ["Custom"])
            self.assertEqual(result["packages"][0]["framework_resolution"], "unverified")

    def test_source_set_and_build_input_drift_refuse(self):
        with tempfile.TemporaryDirectory() as temporary:
            fixture = self.fixture(Path(temporary))
            with self.assertRaisesRegex(ValueError, "source set changed"):
                self.collect_fixture(fixture, source_change=fixture[6][:-1])
            packages = [dict(p) for p in fixture[4]]
            packages[-1]["Imports"] = ["new/import"]
            with self.assertRaisesRegex(ValueError, "build metadata changed"):
                self.collect_fixture(fixture, package_change=packages)

    def test_source_bytes_drift_refuses_without_publishing(self):
        with tempfile.TemporaryDirectory() as temporary:
            root, driver, native, graph, packages, notices, sources, configuration = self.fixture(Path(temporary))
            original = inventory.file_identity
            count = 0
            def identity(base, name):
                nonlocal count
                result = original(base, name)
                if name == "font.ttf":
                    count += 1
                    if count == 2: result["sha256"] = "changed"
                return result
            with mock.patch.object(inventory, "file_identity", side_effect=identity):
                with self.assertRaisesRegex(ValueError, "input changed"):
                    self.collect_fixture((root, driver, native, graph, packages, notices, sources, configuration))

    def test_sdk_count_refuses_before_hashing_surplus_file(self):
        with tempfile.TemporaryDirectory() as temporary:
            fixture = list(self.fixture(Path(temporary)))
            fixture[4] = [fixture[4][0]]
            with mock.patch.object(inventory, "MAX_FILES", 1), mock.patch.object(inventory, "file_identity", wraps=inventory.file_identity) as identity:
                with self.assertRaisesRegex(ValueError, "SDK inventory exceeds file-count bound"):
                    self.collect_fixture(tuple(fixture))
                self.assertFalse(any(name == "drivers/cogent/LICENSE" for (base, name), _ in identity.call_args_list))

    def test_paths_symlinks_missing_graph_identity_and_bounds_refuse(self):
        with tempfile.TemporaryDirectory() as temporary:
            fixture = self.fixture(Path(temporary)); driver = fixture[1]
            for name in ("../LICENSE", "/LICENSE", "a//b", ""):
                with self.assertRaises(ValueError): inventory.file_identity(driver, name)
            (driver / "alias.ttf").symlink_to(driver / "font.ttf")
            with self.assertRaisesRegex(ValueError, "symlinks"): inventory.file_identity(driver, "alias.ttf")
            with mock.patch.object(inventory, "MAX_FILE_BYTES", 2):
                with self.assertRaisesRegex(ValueError, "file bound"): inventory.file_identity(driver, "font.ttf")
            roots = {"example.com/driver": driver, "example.com/native": fixture[2]}
            for mutation in ({"Incomplete": True}, {"Module": None}, {"Dir": temporary}, {"CgoLDFLAGS": ["-L"]}):
                package = {**fixture[4][-1], **mutation}
                with self.assertRaises(ValueError): inventory.package_inventory([package], roots, {})
            with mock.patch.object(inventory, "MAX_FILES", 1), mock.patch.object(inventory, "file_identity", wraps=inventory.file_identity) as identity:
                with self.assertRaisesRegex(ValueError, "file-count bound"): inventory.package_inventory(fixture[4], roots, {})
                identity.assert_not_called()
            with mock.patch.object(inventory, "MAX_FILES", 2):
                result = inventory.package_inventory(fixture[4], roots, {})
                self.assertEqual(len(result["embedded_assets"]) + len(result["packages"][0]["cgo_source_files"]), 2)

    def test_exclusive_output_and_no_inventory_side_effect_before_refusal(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary) / "source"; root.mkdir()
            existing = Path(temporary) / "inventory.json"; existing.write_bytes(b"keep")
            with mock.patch.object(inventory, "collect") as collect:
                for output in (existing, root / "inside.json"):
                    with self.assertRaises(ValueError): inventory.write_inventory(root, output)
                dangling = Path(temporary) / "dangling.json"; dangling.symlink_to("absent")
                with self.assertRaises(ValueError): inventory.write_inventory(root, dangling)
                collect.assert_not_called()
            output = Path(temporary) / "new.json"
            record = {"sdk_files": [], "embedded_assets": [], "native_link_candidates": [], "distribution_audit": "pending"}
            with mock.patch.object(inventory, "collect", return_value=record): result = inventory.write_inventory(root, output)
            self.assertEqual(json.loads(output.read_bytes()), record)
            self.assertEqual(result["sha256"], hashlib.sha256(output.read_bytes()).hexdigest())
            self.assertFalse(result["runtime_executed"])
            self.assertEqual(list(Path(temporary).glob(".vellum-inventory-*")), [])

    def test_notice_metadata_streams_and_refuses_oversize_without_read_bytes(self):
        with tempfile.TemporaryDirectory() as temporary:
            fixture = self.fixture(Path(temporary)); root, native = fixture[0], fixture[2]
            (root / "release").mkdir()
            (root / "release/version.json").write_text('{"module":"example.com/framework"}')
            # This isolated module record exercises the real notice collector.
            graph = [fixture[3][-1]]
            with mock.patch.object(Path, "read_bytes", side_effect=AssertionError("unbounded read forbidden")):
                result, copies = inventory.release.license_inventory(root, modules=graph, copy_notices=False, max_notice_bytes=100)
                self.assertFalse(result["notice_contents_included"])
                self.assertEqual(copies, {})
                notice = result["modules"][0]["license_files"][0]
                self.assertEqual(notice["sha256"], hashlib.sha256(b"native declaration candidate").hexdigest())
                self.assertEqual(notice["bytes"], len(b"native declaration candidate"))
                self.assertNotIn("notice", notice)
                with self.assertRaisesRegex(ValueError, "notice exceeds inventory file bound"):
                    inventory.release.license_inventory(root, modules=graph, copy_notices=False, max_notice_bytes=2)

    def test_link_member_observation_stays_unverified_and_reports_unknown_archive(self):
        from native_archive_test import member, link_map
        with tempfile.TemporaryDirectory() as temporary:
            fixture=self.fixture(Path(temporary));native=fixture[2];archive=native/'arm64/libfake.a'
            archive.write_bytes(b'!<arch>\n'+member('used.o',b'used object')+member('unused.o',b'unused object',True))
            binary=Path(temporary)/'app';binary.write_bytes(b'\xcf\xfa\xed\xfe\x0c\x00\x00\x01'+b'fixture bytes')
            mapping=Path(temporary)/'app.map';mapping.write_text(link_map(archive,['used.o']).replace('# Sections:', '[2] /unknown/libother.a(other.o)\n# Sections:'))
            result=self.collect_fixture(fixture,binary=binary,link_map=mapping)
            observation=result['native_link_observation'];record=observation['archives'][0]
            self.assertEqual(observation['artifact_pair_binding'],'caller-supplied-unverified')
            self.assertEqual(record['archive_member_count'],2)
            self.assertEqual([m['name']for m in record['selected_members']],['used.o'])
            self.assertEqual(record['selected_members'][0]['sha256'],hashlib.sha256(b'used object').hexdigest())
            self.assertEqual(observation['unresolved_archives'],[{'file':'/unknown/libother.a','reason':'not-a-verified-module-file-candidate'}])
            mapping.write_text(link_map(archive,['used.o']).replace(str(archive),'libfake.a'))
            relative=self.collect_fixture(fixture,binary=binary,link_map=mapping)['native_link_observation']
            self.assertEqual(relative['archives'],[])
            self.assertEqual(relative['unresolved_archives'],[{'file':'libfake.a','reason':'relative-linker-directory-unbound'}])
            self.assertEqual(result['distribution_audit'],'pending');self.assertFalse(result['runtime_executed'])
            self.assertEqual(result['native_link_candidates'][0]['link_selection'],'unverified')

    def test_incomplete_pair_mismatched_architecture_and_candidate_drift_refuse(self):
        from native_archive_test import member, link_map
        with tempfile.TemporaryDirectory() as temporary:
            fixture=self.fixture(Path(temporary));root,driver,native=fixture[:3];archive=native/'arm64/libfake.a'
            archive.write_bytes(b'!<arch>\n'+member('used.o',b'object'))
            binary=Path(temporary)/'app';binary.write_bytes(b'\xcf\xfa\xed\xfe\x0c\x00\x00\x01')
            mapping=Path(temporary)/'app.map';mapping.write_text(link_map(archive,['used.o']))
            with mock.patch.object(inventory.release,'source_state')as source:
                with self.assertRaisesRegex(ValueError,'supplied together'):inventory.collect(root,binary=binary)
                source.assert_not_called()
            roots={'example.com/driver':driver,'example.com/native':native};builds=inventory.package_inventory(fixture[4],roots,{})
            mapping.write_text(link_map(archive,['used.o']).replace('arm64','x86_64'))
            with self.assertRaisesRegex(ValueError,'architecture differ'):inventory.native_link_observation(binary,mapping,roots,builds)
            binary.write_bytes(b'\xcf\xfa\xed\xfe\x07\x00\x00\x01')
            with self.assertRaisesRegex(ValueError,'selected Go build architecture'):
                self.collect_fixture(fixture,binary=binary,link_map=mapping)
            binary.write_bytes(b'\xcf\xfa\xed\xfe\x0c\x00\x00\x01')
            mapping.write_text(link_map(archive,['used.o']))
            archive.write_bytes(b'!<arch>\n'+member('used.o',b'changed'))
            with self.assertRaisesRegex(ValueError,'differs from collected candidate'):inventory.native_link_observation(binary,mapping,roots,builds)


if __name__ == "__main__":
    unittest.main()
