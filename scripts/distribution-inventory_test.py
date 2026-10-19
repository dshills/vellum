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

    def collect_fixture(self, fixture, package_change=None, source_change=None, source_hook=None, **observation):
        root, driver, native, graph, packages, notices, sources, configuration = fixture
        calls = []
        def output(argv, **kwargs):
            calls.append((argv, kwargs))
            if argv[:2] == ["go", "env"]: return json.dumps(configuration)
            if source_hook and len(calls) == 3: source_hook()
            if package_change and len(calls) == 3: return "\n".join(json.dumps(p) for p in package_change)
            return "\n".join(json.dumps(p) for p in packages)
        source_results = [iter(sources), iter(source_change if source_change is not None else sources)]
        with mock.patch.object(inventory.release, "git_output", return_value=str(root)), mock.patch.object(inventory, "check_child_output", side_effect=output), mock.patch.object(inventory.release, "source_state", return_value={"commit": "fixture", "clean": True}), mock.patch.object(inventory.release, "verified_module_graph", return_value=graph), mock.patch.object(inventory.release, "license_inventory", return_value=(notices, {})), mock.patch.object(inventory.release, "source_files", side_effect=source_results):
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




class SDKSourceChecks(unittest.TestCase):
    def fixture(self, base):
        root = base / 'SDK payload with spaces'; root.mkdir()
        self.write_manifest(root, {'go.mod': b'module example.com/framework\ngo 1.25.6\n', 'api.go': b'package framework\n'})
        return root

    def write_manifest(self, root, bodies):
        for name, body in bodies.items():
            p = root / name;p.parent.mkdir(parents=True, exist_ok=True);p.write_bytes(body)
        manifest = {'schema': 1, 'release': {'module': 'example.com/framework', 'version': 'v0.1.0-alpha.2'},
                    'source': {'commit': 'a' * 40, 'clean': False}, 'private_build': True,
                    'files': {name: hashlib.sha256(body).hexdigest() for name, body in bodies.items()}}
        (root / 'SDK_MANIFEST.json').write_text(json.dumps(manifest))
        return manifest

    def test_exact_relocated_payload_and_manifest_identity_without_git(self):
        results = []
        for _ in range(2):
            with tempfile.TemporaryDirectory() as temporary:
                root = self.fixture(Path(temporary))
                with mock.patch.object(inventory.release, 'git_output', side_effect=AssertionError('SDK must not query Git')):
                    state, files = inventory.source_snapshot(root)
                self.assertEqual(state['kind'], 'sdk-payload')
                self.assertEqual(state['declared_source'], {'commit': 'a' * 40, 'clean': False})
                self.assertEqual(state['authentication'], 'not independently verified')
                self.assertEqual(state['sdk_manifest_sha256'], hashlib.sha256((root / 'SDK_MANIFEST.json').read_bytes()).hexdigest())
                self.assertEqual([n for n, _ in files], ['SDK_MANIFEST.json', 'api.go', 'go.mod'])
                self.assertNotIn(str(root), json.dumps(state));results.append(state)
        self.assertEqual(results[0], results[1])

    def test_git_checkout_fallback_keeps_existing_source_identity_and_rejects_ancestor(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary).resolve()
            with mock.patch.object(inventory.release, 'git_output', return_value=str(root)), mock.patch.object(inventory.release, 'source_state', return_value={'commit': 'fixture', 'clean': True}), mock.patch.object(inventory.release, 'source_files', return_value=[('api.go', root / 'api.go')]):
                self.assertEqual(inventory.source_snapshot(root), ({'commit': 'fixture', 'clean': True}, [('api.go', root / 'api.go')]))
            with mock.patch.object(inventory.release, 'git_output', return_value=str(root.parent)), mock.patch.object(inventory.release, 'source_state') as state:
                with self.assertRaisesRegex(ValueError, 'Git checkout root'):
                    inventory.source_snapshot(root)
                state.assert_not_called()

    def test_malformed_sdk_metadata_never_falls_back_to_git(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = self.fixture(Path(temporary)); path = root / 'SDK_MANIFEST.json'; original = json.loads(path.read_text())
            variants = [b'{"schema":1,"schema":1}', b'{"schema":NaN}', b'[]']
            for key, value in [('schema', True), ('release', []), ('source', {'commit': 'a' * 40, 'clean': 1}), ('files', []), ('source', {'commit': 'pretend', 'clean': True})]:
                data = dict(original);data[key] = value;variants.append(json.dumps(data).encode())
            for raw in variants:
                with self.subTest(raw=raw), mock.patch.object(inventory.release, 'git_output', side_effect=AssertionError('no fallback')):
                    path.write_bytes(raw)
                    with self.assertRaises(ValueError):inventory.source_snapshot(root)

    def test_invalid_paths_checksums_module_and_manifest_membership_refuse(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = self.fixture(Path(temporary));path = root / 'SDK_MANIFEST.json';original = json.loads(path.read_text())
            for name in ['../escape', '/absolute', 'a//b', 'a/./b', 'a/../b', 'a\\b', 'bad\x00name']:
                data = json.loads(json.dumps(original));data['files'][name] = 'a' * 64;path.write_text(json.dumps(data))
                with self.subTest(name=name), self.assertRaisesRegex(ValueError, 'path/checksum'):inventory.source_snapshot(root)
            for mutate in [lambda d:d['files'].update({'api.go': 'z' * 64}), lambda d:d['files'].update({'SDK_MANIFEST.json': 'a' * 64}), lambda d:d['files'].pop('go.mod'), lambda d:d['release'].update(module='example.com/another')]:
                data = json.loads(json.dumps(original));mutate(data);path.write_text(json.dumps(data))
                with self.assertRaises(ValueError):inventory.source_snapshot(root)

    def test_changed_missing_unknown_and_symlink_members_refuse(self):
        import os
        with tempfile.TemporaryDirectory() as temporary:
            root = self.fixture(Path(temporary));source = root / 'api.go';original = source.read_bytes()
            source.write_bytes(b'other source')
            with self.assertRaisesRegex(ValueError, 'checksum mismatch'):inventory.source_snapshot(root)
            source.unlink()
            with self.assertRaisesRegex(ValueError, 'coverage mismatch'):inventory.source_snapshot(root)
            source.write_bytes(original)
            for name, is_dir in [('extra.go', False), ('__pycache__', True)]:
                p = root / name;p.mkdir() if is_dir else p.write_bytes(b'extra')
                with self.assertRaisesRegex(ValueError, 'coverage mismatch'):inventory.source_snapshot(root)
                p.rmdir() if is_dir else p.unlink()
            source.unlink();source.symlink_to(root / 'go.mod')
            with self.assertRaisesRegex(ValueError, 'symlinks'):inventory.source_snapshot(root)
            source.unlink();os.mkfifo(source)
            with self.assertRaisesRegex(ValueError, 'coverage mismatch'):inventory.source_snapshot(root)

    def test_manifest_special_symlink_and_oversized_inputs_refuse(self):
        import os
        with tempfile.TemporaryDirectory() as temporary:
            root = self.fixture(Path(temporary));path = root / 'SDK_MANIFEST.json';path.unlink()
            path.symlink_to(root / 'missing')
            with self.assertRaises(ValueError):inventory.source_snapshot(root)
            path.unlink();os.mkfifo(path)
            with self.assertRaises(ValueError):inventory.source_snapshot(root)
            path.unlink();path.mkdir()
            with self.assertRaises((ValueError, IsADirectoryError)):inventory.source_snapshot(root)
            path.rmdir()
            with path.open('wb') as stream:stream.truncate(inventory.MAX_SDK_MANIFEST_BYTES + 1)
            with self.assertRaisesRegex(ValueError, 'bound'):inventory.source_snapshot(root)

    def test_count_and_member_bounds_precede_surplus_hashing(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = self.fixture(Path(temporary))
            with mock.patch.object(inventory, 'MAX_FILES', 2), mock.patch.object(inventory, 'regular_stream') as reader:
                with self.assertRaisesRegex(ValueError, 'file-count bound'):inventory.source_snapshot(root)
                reader.assert_not_called()
            with (root / 'api.go').open('wb') as stream:stream.truncate(inventory.MAX_SDK_MEMBER_BYTES + 1)
            with self.assertRaisesRegex(ValueError, 'bound'):inventory.source_snapshot(root)

    def test_aggregate_bound_refuses_before_surplus_payload_read(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = self.fixture(Path(temporary));seen = [];original = inventory.regular_stream
            class Reader:
                def __init__(self, stream, name):self.stream=stream;self.name=name
                def __enter__(self):return self
                def __exit__(self, *args):self.stream.close()
                def read(self, size):seen.append(self.name);return self.stream.read(size)
                def fileno(self):return self.stream.fileno()
            def observed(path, bound):
                stream, before = original(path, bound)
                return Reader(stream, path.name), before
            total = sum(p.stat().st_size for p in root.iterdir())
            with mock.patch.object(inventory, 'MAX_SDK_TOTAL_BYTES', total-1), mock.patch.object(inventory, 'regular_stream', side_effect=observed):
                with self.assertRaisesRegex(ValueError, 'aggregate byte bound'):inventory.source_snapshot(root)
            self.assertIn('api.go', seen);self.assertNotIn('go.mod', seen)

    def test_growing_or_shortened_stream_refuses_with_admitted_read_bound(self):
        for mode in ['grow', 'short']:
            with self.subTest(mode=mode), tempfile.TemporaryDirectory() as temporary:
                root=self.fixture(Path(temporary));original=inventory.regular_stream;requests=[]
                class Reader:
                    def __init__(self, stream):self.stream=stream
                    def __enter__(self):return self
                    def __exit__(self, *args):self.stream.close()
                    def fileno(self):return self.stream.fileno()
                    def read(self, size):
                        requests.append(size)
                        if len(requests)>2:raise AssertionError('reader did not stop at admitted size')
                        if mode=='short':return b''
                        return self.stream.read(size) or b'x'*size
                def changed(path, bound):
                    stream,before=original(path,bound)
                    return (Reader(stream) if path.name=='api.go' else stream),before
                with mock.patch.object(inventory,'regular_stream',side_effect=changed):
                    with self.assertRaisesRegex(ValueError,'admitted byte bound|shortened'):
                        inventory.source_snapshot(root)
                self.assertLessEqual(sum(requests),(root/'api.go').stat().st_size+1)

    def test_include_untracked_refused_for_sdk_payload(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = self.fixture(Path(temporary))
            with self.assertRaisesRegex(ValueError, 'include_untracked'):inventory.source_snapshot(root, include_untracked=True)

    def test_valid_redeclaration_changes_observed_manifest_identity_not_original_provenance(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = self.fixture(Path(temporary));before, _ = inventory.source_snapshot(root)
            manifest = json.loads((root / 'SDK_MANIFEST.json').read_text());body=b'package framework\n// new\n';(root / 'api.go').write_bytes(body);manifest['files']['api.go']=hashlib.sha256(body).hexdigest();(root / 'SDK_MANIFEST.json').write_text(json.dumps(manifest));after, _ = inventory.source_snapshot(root)
            self.assertEqual(before['declared_source'], after['declared_source'])
            self.assertNotEqual(before['sdk_manifest_sha256'], after['sdk_manifest_sha256'])
            self.assertEqual(after['authentication'], 'not independently verified')

    def test_full_collector_consumes_sdk_without_git_and_binds_manifest(self):
        with tempfile.TemporaryDirectory() as temporary:
            helper = InventoryChecks();fixture = helper.fixture(Path(temporary));root = fixture[0]
            bodies = {p.relative_to(root).as_posix():p.read_bytes() for p in root.rglob('*') if p.is_file()};bodies['go.mod']=b'module example.com/framework\ngo 1.25.6\n';self.write_manifest(root, bodies)
            result = helper.collect_fixture(fixture)
            self.assertEqual(result['source']['kind'], 'sdk-payload')
            self.assertIn('SDK_MANIFEST.json', {r['file'] for r in result['sdk_files']})
            self.assertEqual(result['distribution_audit'], 'pending');self.assertFalse(result['runtime_executed'])

    def test_full_collector_refuses_consistent_sdk_rewrite_after_package_listing(self):
        with tempfile.TemporaryDirectory() as temporary:
            helper = InventoryChecks();fixture = helper.fixture(Path(temporary));root = fixture[0]
            bodies = {p.relative_to(root).as_posix():p.read_bytes() for p in root.rglob('*') if p.is_file()};bodies['go.mod']=b'module example.com/framework\ngo 1.25.6\n';self.write_manifest(root, bodies)
            def mutate():
                body=b'package bridge\n// late\n';(root / 'drivers/cogent/bridge.go').write_bytes(body);p=root/'SDK_MANIFEST.json';data=json.loads(p.read_text());data['files']['drivers/cogent/bridge.go']=hashlib.sha256(body).hexdigest();p.write_text(json.dumps(data))
            with self.assertRaisesRegex(ValueError, 'source state changed'):
                helper.collect_fixture(fixture, source_hook=mutate)

    def test_normal_cli_help_keeps_distributed_source_free_of_bytecode(self):
        import shutil, sys
        from child_process import run_child
        with tempfile.TemporaryDirectory(prefix='Vellum SDK ordinary Python ') as temporary:
            root=Path(temporary);scripts=root/'scripts';scripts.mkdir()
            for source in Path(__file__).parent.glob('*.py'):shutil.copyfile(source, scripts/source.name)
            env=inventory.child_environment();env.pop('PYTHONDONTWRITEBYTECODE', None)
            for script in ['new-app.py', 'build-macos.py', 'distribution-inventory.py', 'package-app.py', 'check-sdk.py', 'release-sdk.py']:
                with self.subTest(script=script):
                    result=run_child([sys.executable,str(scripts/script),'--help'],cwd=root,env=env,capture_output=True,text=True,timeout=30)
                    self.assertEqual(result.returncode,0,result.stderr)
                    self.assertEqual(list(root.rglob('__pycache__')),[])


class OwnedAssetDeclarationChecks(unittest.TestCase):
    def fixture(self, base):
        helper = InventoryChecks()
        fixture = helper.fixture(base)
        root, driver = fixture[:2]
        symbols = driver / "native/assets/symbols"
        sibling = driver / "native/assets/unselected"
        symbols.mkdir(parents=True)
        sibling.mkdir()
        (symbols / "rain.svg").write_bytes(b'<svg xmlns="http://www.w3.org/2000/svg"/>\n')
        (symbols / "LICENSE_APACHE.txt").write_bytes(b"original Apache declaration\n")
        (symbols / "NOTICE.md").write_bytes(b"original source attribution\n")
        (symbols / "SOURCE.json").write_bytes(b'{"scope":"source receipt, not a license"}\n')
        (sibling / "NOTICE.md").write_bytes(b"unselected sibling declaration\n")
        fixture[4][1]["EmbedFiles"] = ["native/assets/symbols/rain.svg"]
        fixture[6].extend((p.relative_to(root).as_posix(), p) for p in symbols.iterdir())
        return helper, fixture

    def test_collect_selected_owned_svg_original_ancestor_declarations_and_relocation(self):
        reports = []
        for _ in range(2):
            with tempfile.TemporaryDirectory() as temporary:
                helper, fixture = self.fixture(Path(temporary))
                result = helper.collect_fixture(fixture)
                asset = result["embedded_assets"][0]
                self.assertEqual(asset["file"], "native/assets/symbols/rain.svg")
                self.assertEqual(asset["sha256"], hashlib.sha256(b'<svg xmlns="http://www.w3.org/2000/svg"/>\n').hexdigest())
                self.assertEqual(asset["kind"], "embedded-asset")
                self.assertEqual(asset["nearby_notice_candidates"], [
                    "LICENSE", "native/assets/symbols/LICENSE_APACHE.txt", "native/assets/symbols/NOTICE.md"])
                self.assertEqual(asset["rights_status"], "review-required")
                declarations = {(r["module"], r["file"]): r for r in result["framework_declarations"]}
                for name in asset["nearby_notice_candidates"]:
                    record = declarations["example.com/driver", name]
                    self.assertEqual(record["sha256"], hashlib.sha256((fixture[1] / name).read_bytes()).hexdigest())
                self.assertNotIn(("example.com/driver", "native/assets/symbols/SOURCE.json"), declarations)
                self.assertNotIn(("example.com/driver", "native/assets/unselected/NOTICE.md"), declarations)
                self.assertEqual(result["distribution_audit"], "pending")
                self.assertFalse(result["runtime_executed"])
                reports.append(result)
        self.assertEqual(reports[0], reports[1])

    def test_collect_owned_root_license_remains_required(self):
        with tempfile.TemporaryDirectory() as temporary:
            helper, fixture = self.fixture(Path(temporary))
            (fixture[1] / "LICENSE").unlink()
            with self.assertRaisesRegex(ValueError, "regular file"):
                helper.collect_fixture(fixture)

    def test_collect_refuses_selected_asset_ancestor_and_notice_symlinks(self):
        for mode in ["ancestor", "notice"]:
            with self.subTest(mode=mode), tempfile.TemporaryDirectory() as temporary:
                base = Path(temporary)
                helper, fixture = self.fixture(base)
                symbols = fixture[1] / "native/assets/symbols"
                if mode == "ancestor":
                    destination = base / "redirected symbols"
                    symbols.rename(destination)
                    symbols.symlink_to(destination, target_is_directory=True)
                else:
                    notice = symbols / "NOTICE.md"
                    target = base / "other original declaration"
                    notice.rename(target)
                    notice.symlink_to(target)
                with self.assertRaisesRegex(ValueError, "symlink"):
                    helper.collect_fixture(fixture)

    def test_collect_refuses_nonregular_adjacent_notice_without_opening_it(self):
        import os
        with tempfile.TemporaryDirectory() as temporary:
            helper, fixture = self.fixture(Path(temporary))
            notice = fixture[1] / "native/assets/symbols/NOTICE.md"
            notice.unlink()
            os.mkfifo(notice)
            with self.assertRaisesRegex(ValueError, "regular file"):
                helper.collect_fixture(fixture)

    def test_collect_refuses_owned_notice_drift_after_discovery(self):
        with tempfile.TemporaryDirectory() as temporary:
            helper, fixture = self.fixture(Path(temporary))
            real_identity = inventory.file_identity
            changed = []
            def mutate(base, name):
                record = real_identity(base, name)
                if name == "native/assets/symbols/NOTICE.md" and not changed:
                    (base / name).write_bytes(b"changed after the observed declaration\n")
                    changed.append(name)
                return record
            with mock.patch.object(inventory, "file_identity", side_effect=mutate):
                with self.assertRaisesRegex(ValueError, "inventory input changed"):
                    helper.collect_fixture(fixture)
            self.assertEqual(changed, ["native/assets/symbols/NOTICE.md"])

    def test_collect_refuses_adjacent_notice_added_after_discovery(self):
        with tempfile.TemporaryDirectory() as temporary:
            helper, fixture = self.fixture(Path(temporary))
            original = inventory.owned_asset_declarations
            calls = []
            def discover(*args):
                result = original(*args)
                if not calls:
                    (fixture[1] / "native/assets/symbols/OFL.txt").write_bytes(b"late declaration")
                calls.append(result)
                return result
            with mock.patch.object(inventory, "owned_asset_declarations", side_effect=discover):
                with self.assertRaisesRegex(ValueError, "owned asset declarations changed"):
                    helper.collect_fixture(fixture)
            self.assertEqual(len(calls), 2)

    def test_collect_owned_notice_bytes_and_directory_entries_remain_bounded(self):
        for mode in ["bytes", "entries", "ancestors"]:
            with self.subTest(mode=mode), tempfile.TemporaryDirectory() as temporary:
                helper, fixture = self.fixture(Path(temporary))
                symbols = fixture[1] / "native/assets/symbols"
                if mode == "bytes":
                    (symbols / "NOTICE.md").write_bytes(b"x" * 65)
                    setting, limit, reason = "MAX_FILE_BYTES", 64, "file bound"
                elif mode == "entries":
                    for i in range(25):
                        (symbols / ("unrelated-" + str(i))).write_bytes(b"not a declaration")
                    setting, limit, reason = "MAX_FILES", 20, "notice search exceeds file-count bound"
                else:
                    before = symbols / "rain.svg"
                    after = symbols.joinpath(*(["nested"] * 12), "rain.svg")
                    after.parent.mkdir(parents=True)
                    before.rename(after)
                    fixture[4][1]["EmbedFiles"] = [after.relative_to(fixture[1]).as_posix()]
                    fixture[6][:] = [(after.relative_to(fixture[0]).as_posix(), after) if path == before else (name, path)
                                     for name, path in fixture[6]]
                    setting, limit, reason = "MAX_FILES", 10, "notice search exceeds ancestor-count bound"
                with mock.patch.object(inventory, setting, limit):
                    with self.assertRaisesRegex(ValueError, reason):
                        helper.collect_fixture(fixture)


if __name__ == "__main__":
    unittest.main()
