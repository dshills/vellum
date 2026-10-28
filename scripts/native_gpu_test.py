"""Private native SDK selection, tamper refusal and actual relocated module graph."""
import copy
from contextlib import nullcontext
import hashlib
import importlib.util
import io
import json
import os
from pathlib import Path
import shutil
import struct
import subprocess
import sys
import tarfile
import tempfile
import unittest
from unittest import mock

import native_gpu as gpu
import native_linkers as linkers


def load(name, path):
    spec = importlib.util.spec_from_file_location(name, path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


SCRIPTS = Path(__file__).resolve().parent
release = load("gpu_release_test", SCRIPTS / "release-sdk.py")
checker = load("gpu_sdk_test", SCRIPTS / "check-sdk.py")


def ar(payload=None, name="one.o/"):
    if payload is None:
        payload = struct.pack("<IIIIIIII", 0xFEEDFACF, 0x0100000C, 0, 1, 0, 0, 0, 0)
    header = (name.ljust(16) + "0".ljust(12) + "0".ljust(6) + "0".ljust(6) +
              "644".ljust(8) + str(len(payload)).ljust(10) + "`\n").encode()
    return b"!<arch>\n" + header + payload + (b"\n" if len(payload) % 2 else b"")


def inputs(base, body=None):
    library = base / "candidate.a"
    library.write_bytes(ar() if body is None else body)
    checksum = gpu.digest(library.read_bytes())
    provenance = base / "provenance.json"
    provenance.write_text(json.dumps(dict(schema=1, target="darwin-arm64", artifact_sha256=checksum,
                                        declaration="unit fixture, not source attestation")))
    return dict(macos_gpu_library=library, macos_gpu_sha256=checksum, macos_gpu_provenance=provenance)


def source(base):
    root = base / "source checkout"
    root.mkdir()
    files = {}

    def write(name, text):
        p = root / name
        p.parent.mkdir(parents=True, exist_ok=True)
        p.write_text(text)
        files[name] = p

    write("release/version.json", json.dumps(dict(version="v0.1.0-alpha.1", module="github.com/dshills/vellum")))
    write("go.mod", "module github.com/dshills/vellum\ngo 1.25.6\n")
    write("drivers/cogent/go.mod", "module github.com/dshills/vellum/drivers/cogent\ngo 1.25.6\n" + "".join("require " + module + " v0.0.0\n" for module in linkers.DIRECTORIES))
    write("drivers/cogent/go.sum", "")
    for module, name in [("cogentcore.org/core", "upstream"), ("github.com/go-text/typesetting", "typesetting")]:
        write("drivers/cogent/build/" + name + "/go.mod", "module " + module + "\ngo 1.25.6\n")
    write("examples/hello/go.mod", '''module example.com/hello
go 1.25.6
require (
 github.com/dshills/vellum v0.0.0
 github.com/dshills/vellum/drivers/cogent v0.0.0
 cogentcore.org/core v0.0.0
 github.com/go-text/typesetting v0.0.0
)
''')
    example_module = files["examples/hello/go.mod"]
    example_module.write_text(example_module.read_text() + "".join("require " + module + " v0.0.0\n" for module in linkers.DIRECTORIES))
    for name, text in [("main.go", "package main\nfunc main() {}\n"), ("model.go", "package main\n"),
                       ("model_test.go", "package main\n"), ("go.sum", ""), ("LICENSE", "starter notice\n")]:
        write("examples/hello/" + name, text)
    for name in ["main.go", "model.go", "model_test.go"]:
        write("scripts/templates/app/" + name + ".txt", "package main\n")
    for name in ["new-app.py", "native_gpu.py", "native_archive.py", "native_linkers.py"]:
        write("scripts/" + name, (SCRIPTS / name).read_text())
    return root, files


def package(root, files, path, **kwargs):
    with mock.patch.object(release, "source_files", return_value=sorted(files.items())), \
         mock.patch.object(release, "source_state", return_value=dict(clean=True, commit="a" * 40)), \
         mock.patch.object(release, "license_inventory", return_value=({"modules": []}, {})):
        return release.package(root, path, **kwargs)


def notice_bundle(root, files, checksum):
    # Larger than the old provenance bound: bundled Rust COPYRIGHT is >1MiB.
    name = gpu.COMPONENT_DIRECTORY + "/NOTICES/crate/LICENSE"
    content = b"declared license text\n" * 60000
    path = root / name
    path.parent.mkdir(parents=True)
    path.write_bytes(content)
    files[name] = path
    inventory = {"schema": 1, "artifact_sha256": checksum, "distribution_audit": "pending",
                 "components": [{"id": "crate-1", "notice_files": ["NOTICES/crate/LICENSE"]}],
                 "files": {"NOTICES/crate/LICENSE": gpu.digest(content)}}
    path = root / gpu.COMPONENT_INVENTORY
    path.write_text(json.dumps(inventory))
    files[gpu.COMPONENT_INVENTORY] = path
    return inventory


class NativeGPUChecks(unittest.TestCase):
    def test_component_notices_ship_and_verify_after_relocation(self):
        with tempfile.TemporaryDirectory() as tmp:
            base = Path(tmp)
            root, files = source(base)
            options = inputs(base)
            notice_bundle(root, files, options["macos_gpu_sha256"])
            result = package(root, files, base / "sdk.tar.gz", **options)
            verified = checker.verify_archive(base / "sdk.tar.gz", result["sha256"])
            note = verified["manifest"]["native_gpu"]["component_notices"]
            self.assertEqual(note["components"], 1)
            self.assertEqual(note["distribution_audit"], "pending")
            sdk = checker.extract_verified(base / "sdk.tar.gz", base / "relocated", verified)
            shutil.rmtree(root)
            self.assertEqual(gpu.sdk_candidate(sdk)["component_notices"], note)
            # Consumer validation cannot rely on producer-side checks or a
            # hash-consistent but semantically inconsistent manifest alone.
            with mock.patch.object(gpu, "MAX_COMPONENT_TOTAL", 100), mock.patch.object(gpu, "read_regular", wraps=gpu.read_regular) as reads:
                with self.assertRaisesRegex(ValueError, "bound"):
                    gpu.sdk_candidate(sdk)
                self.assertFalse(any(str(call.args[0]).startswith(str(sdk / gpu.COMPONENT_DIRECTORY / "NOTICES")) for call in reads.call_args_list), "already exhausted inventory budget still read a notice")
            manifest_path = sdk / "SDK_MANIFEST.json"
            manifest = json.loads(manifest_path.read_text())
            manifest["native_gpu"]["component_notices"]["components"] = 2
            manifest_path.write_text(json.dumps(manifest))
            with self.assertRaisesRegex(ValueError, "does not match"):
                gpu.sdk_candidate(sdk)
            manifest["native_gpu"]["component_notices"]["components"] = 1
            notice_inventory = sdk / gpu.COMPONENT_INVENTORY
            original_inventory = notice_inventory.read_bytes()
            inventory_data = json.loads(original_inventory)
            inventory_data["artifact_sha256"] = "b" * 64
            notice_inventory.write_text(json.dumps(inventory_data))
            changed_hash = gpu.digest(notice_inventory.read_bytes())
            manifest["files"][gpu.COMPONENT_INVENTORY] = changed_hash
            manifest["native_gpu"]["component_notices"]["files"][gpu.COMPONENT_INVENTORY] = changed_hash
            manifest_path.write_text(json.dumps(manifest))
            with self.assertRaisesRegex(ValueError, "does not match"):
                gpu.sdk_candidate(sdk)
            notice_inventory.write_bytes(original_inventory)
            original_hash = gpu.digest(original_inventory)
            manifest["files"][gpu.COMPONENT_INVENTORY] = original_hash
            manifest["native_gpu"]["component_notices"]["files"][gpu.COMPONENT_INVENTORY] = original_hash
            manifest_path.write_text(json.dumps(manifest))
            with tarfile.open(base / "sdk.tar.gz") as archive:
                inventory = json.loads(archive.extractfile("vellum-v0.1.0-alpha.1/THIRD_PARTY_INVENTORY.json").read())
            self.assertEqual(inventory["native_gpu_candidate"]["component_notices"], note)
            name = next(n for n in note["files"] if n != gpu.COMPONENT_INVENTORY)
            (sdk / name).write_bytes(b"tampered")
            with self.assertRaisesRegex(ValueError, "checksum mismatch"):
                gpu.sdk_candidate(sdk)

    def test_known_component_archive_requires_its_inventory(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            with self.assertRaisesRegex(ValueError, "requires"):
                gpu.component_notices(root, gpu.KNOWN_COMPONENT_ARCHIVE)
            self.assertIsNone(gpu.component_notices(root, "a" * 64))
            files = {}
            notice_bundle(root, files, "a" * 64)
            with self.assertRaisesRegex(ValueError, "mismatched"):
                gpu.component_notices(root, gpu.KNOWN_COMPONENT_ARCHIVE)

    def test_component_notices_refuse_missing_unreferenced_and_escaping_inputs(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            files = {}
            original = notice_bundle(root, files, "a" * 64)
            self.assertIsNone(gpu.component_notices(root, "b" * 64))
            self.assertIsNotNone(gpu.component_notices(root, "a" * 64))
            inventory_path = root / gpu.COMPONENT_INVENTORY
            for name, change in [
                ("traversal", lambda d: d["files"].update({"NOTICES/../elsewhere": "a" * 64})),
                ("absolute", lambda d: d["files"].update({"/outside": "a" * 64})),
                ("backslash", lambda d: d["files"].update({"NOTICES/x\\escape": "a" * 64})),
                ("missing reference", lambda d: d["components"][0]["notice_files"].append("NOTICES/missing")),
                ("unreferenced", lambda d: d["components"][0].update(notice_files=[])),
                ("duplicate identity", lambda d: d["components"].append(d["components"][0])),
                ("false approval", lambda d: d.update(distribution_audit="approved")),
                ("missing component", lambda d: d.update(components=[])),
                ("checksum", lambda d: d["files"].update({"NOTICES/crate/LICENSE": "b" * 64})),
                ("count bound", lambda d: d["files"].update({f"NOTICES/extra-{i}": "a" * 64 for i in range(512)})),
            ]:
                with self.subTest(name=name):
                    changed = copy.deepcopy(original)
                    change(changed)
                    inventory_path.write_text(json.dumps(changed))
                    with self.assertRaises(ValueError):
                        gpu.component_notices(root, "a" * 64)
            inventory_path.write_text(json.dumps(original))
            notice = root / gpu.COMPONENT_DIRECTORY / "NOTICES/crate/LICENSE"
            notice.unlink()
            notice.symlink_to(root / "missing")
            with self.assertRaisesRegex(ValueError, "symlink"):
                gpu.component_notices(root, "a" * 64)
            notice.unlink()
            notice.parent.rmdir()
            notice.parent.symlink_to(root / "missing-directory")
            with self.assertRaisesRegex(ValueError, "symlink"):
                gpu.component_notices(root, "a" * 64)

    def test_component_notices_cannot_be_omitted_or_changed_during_packaging(self):
        with tempfile.TemporaryDirectory() as tmp:
            base = Path(tmp)
            root, files = source(base)
            options = inputs(base)
            notice_bundle(root, files, options["macos_gpu_sha256"])
            name = gpu.COMPONENT_DIRECTORY + "/NOTICES/crate/LICENSE"
            for case in ["omitted", "changed", "extra"]:
                with self.subTest(case=case):
                    selected = dict(files)
                    if case == "omitted":
                        del selected[name]
                    elif case == "changed":
                        selected[name].write_bytes(b"changed")
                    else:
                        extra = root / gpu.COMPONENT_DIRECTORY / "NOTICES/extra"
                        extra.write_bytes(b"extra")
                        selected[gpu.COMPONENT_DIRECTORY + "/NOTICES/extra"] = extra
                    with self.assertRaises(ValueError):
                        package(root, selected, base / (case + ".tar.gz"), **options)
                    if case == "changed":
                        selected[name].write_bytes(b"declared license text\n" * 60000)

    def test_nonfinite_declarations_include_ordinary_exponent_overflow(self):
        for literal in ["NaN", "Infinity", "-Infinity", "1e999", "-1e999"]:
            with self.subTest(literal=literal), self.assertRaisesRegex(ValueError, "nonfinite"):
                gpu.strict_json('{"opaque_declaration":[{"value":' + literal + '}]}')
        self.assertEqual(gpu.strict_json('{"opaque_declaration":1e3}')["opaque_declaration"], 1000)

    def test_sdk_mode_refuses_unlisted_capsule_and_dangling_manifest(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            capsule = root / gpu.DIRECTORY
            capsule.mkdir(parents=True)
            self.assertIsNone(gpu.sdk_candidate(root))  # ordinary source checkout
            manifest = root / "SDK_MANIFEST.json"
            manifest.write_text(json.dumps(dict(files={}, private_build=True)))
            with self.assertRaisesRegex(ValueError, "no manifest record"):
                gpu.sdk_candidate(root)
            capsule.rmdir()
            capsule.symlink_to(root / "missing")
            with self.assertRaisesRegex(ValueError, "no manifest record"):
                gpu.sdk_candidate(root)
            capsule.unlink()
            manifest.unlink()
            manifest.symlink_to(root / "missing-manifest")
            with self.assertRaisesRegex(ValueError, "symlink"):
                gpu.sdk_candidate(root)

    def test_candidate_check_refuses_wrong_host_before_running_archive_code(self):
        verified = {"manifest": {"native_gpu": {"target": "darwin-arm64"}}}
        with mock.patch.object(checker, "verify_archive", return_value=verified), \
             mock.patch.object(checker.sys, "platform", "linux"), \
             mock.patch.object(checker, "run_child") as child, \
             self.assertRaisesRegex(ValueError, "ARM64 Mac host"):
            checker.check(Path("fixture"), "a" * 64)
        child.assert_not_called()

    def test_public_native_requires_both_release_approval_and_exact_notices(self):
        with tempfile.TemporaryDirectory() as tmp:
            base = Path(tmp)
            root, files = source(base)
            options = inputs(base)
            output = base / "not-created/sdk.tar.gz"
            for public in [False, True]:
                with mock.patch.object(release, "license_inventory") as inventory, self.assertRaisesRegex(ValueError, "all three"):
                    release.package(root, output, public=public, macos_gpu_library=options["macos_gpu_library"])
                inventory.assert_not_called()
                self.assertFalse(output.parent.exists())
            metadata = json.loads((root / "release/version.json").read_text())
            metadata.update(namespace_status="selected", license_status="selected", third_party_distribution_audit="pending")
            (root / "release/version.json").write_text(json.dumps(metadata))
            (root / "LICENSE").write_text("framework license fixture")
            files["LICENSE"] = root / "LICENSE"
            with self.assertRaisesRegex(ValueError, "audit is incomplete"):
                package(root, files, output, public=True, **options)
            self.assertFalse(output.parent.exists())
            metadata["third_party_distribution_audit"] = "approved"
            (root / "release/version.json").write_text(json.dumps(metadata))
            with mock.patch.object(release, "hidden_index_paths", return_value=[]):
                for notice_hash in [None, "b" * 64]:
                    if notice_hash:
                        notice_bundle(root, files, notice_hash)
                    with self.assertRaisesRegex(ValueError, "notices matching the exact archive"):
                        package(root, files, output, public=True, **options)
                    self.assertFalse(output.parent.exists())
                # Rebind the fixture inventory to the selected archive, leaving
                # its declaration audit pending, just as in the real source.
                inventory_path = root / gpu.COMPONENT_INVENTORY
                inventory = json.loads(inventory_path.read_text())
                inventory["artifact_sha256"] = options["macos_gpu_sha256"]
                inventory_path.write_text(json.dumps(inventory))
                metadata["third_party_distribution_audit"] = "pending"
                (root / "release/version.json").write_text(json.dumps(metadata))
                with self.assertRaisesRegex(ValueError, "audit is incomplete"):
                    package(root, files, output, public=True, **options)
                self.assertFalse(output.parent.exists())
                metadata["third_party_distribution_audit"] = "approved"
                (root / "release/version.json").write_text(json.dumps(metadata))
                with mock.patch.object(release, "public_dependencies", return_value=nullcontext(None)):
                    result = package(root, files, output, public=True, **options)
            verified = checker.verify_archive(output, result["sha256"])
            manifest = verified["manifest"]
            self.assertFalse(manifest["private_build"])
            self.assertEqual(manifest["native_gpu"]["distribution_audit"], "approved")
            self.assertEqual(manifest["native_gpu"]["component_notices"]["distribution_audit"], "pending")
            sdk = checker.extract_verified(output, base / "relocated", verified)
            self.assertIsNotNone(gpu.sdk_candidate(sdk))
            # Rehash both manifest maps after substituting a different archive
            # identity in the copied inventory: archive verification must refuse.
            with tarfile.open(output) as archive:
                bodies = {m.name: archive.extractfile(m).read() for m in archive.getmembers()}
            prefix = verified["prefix"]
            name = prefix + "/" + gpu.COMPONENT_INVENTORY
            changed = json.loads(bodies[name])
            changed["artifact_sha256"] = "b" * 64
            bodies[name] = json.dumps(changed).encode()
            forged = copy.deepcopy(manifest)
            checksum = gpu.digest(bodies[name])
            forged["files"][gpu.COMPONENT_INVENTORY] = checksum
            forged["native_gpu"]["component_notices"]["files"][gpu.COMPONENT_INVENTORY] = checksum
            bodies[prefix + "/SDK_MANIFEST.json"] = json.dumps(forged).encode()
            invalid = base / "forged.tar.gz"
            with tarfile.open(invalid, "w:gz") as archive:
                for name, body in bodies.items():
                    member = tarfile.TarInfo(name)
                    member.size = len(body)
                    archive.addfile(member, io.BytesIO(body))
            with self.assertRaisesRegex(ValueError, "does not match selected archive"):
                checker.verify_archive(invalid)
            for alteration in ["release", "record", "notices"]:
                bad = copy.deepcopy(manifest)
                if alteration == "release":
                    bad["release"]["third_party_distribution_audit"] = "pending"
                elif alteration == "record":
                    bad["native_gpu"]["distribution_audit"] = "pending"
                else:
                    del bad["native_gpu"]["component_notices"]
                (sdk / "SDK_MANIFEST.json").write_text(json.dumps(bad))
                with self.assertRaises(ValueError):
                    gpu.sdk_candidate(sdk)
                with self.assertRaises(ValueError):
                    gpu.validate_record(bad["native_gpu"], bad["files"], False, bad["release"])

    def test_exact_bytes_hash_and_declared_provenance_are_separate(self):
        with tempfile.TemporaryDirectory() as tmp:
            base = Path(tmp)
            options = inputs(base)
            record, files = gpu.candidate(*options.values())
            self.assertEqual(record["arm64_object_members"], 1)
            self.assertEqual(record["trust"], gpu.TRUST)
            self.assertEqual(files[gpu.DIRECTORY + "/" + gpu.PROVENANCE], options["macos_gpu_provenance"].read_bytes())
            with self.assertRaisesRegex(ValueError, "SHA-256 mismatch"):
                gpu.candidate(options["macos_gpu_library"], "0" * 64, options["macos_gpu_provenance"])
            with self.assertRaisesRegex(ValueError, "exceeds"):
                with mock.patch.object(gpu, "MAX_LIBRARY_BYTES", 1):
                    gpu.candidate(*options.values())
            options["macos_gpu_provenance"].write_text('{"schema":1,"schema":1}')
            with self.assertRaisesRegex(ValueError, "duplicate"):
                gpu.candidate(*options.values())

    def test_wrong_fat_nonobject_and_index_only_payloads_refuse(self):
        bodies = [b"\xca\xfe\xba\xbe" + b"\0" * 28,
                  struct.pack("<IIIIIIII", 0xFEEDFACF, 0x01000007, 0, 1, 0, 0, 0, 0),
                  struct.pack("<IIIIIIII", 0xFEEDFACF, 0x0100000C, 0, 6, 0, 0, 0, 0)]
        for body in bodies:
            with self.subTest(body=body), tempfile.TemporaryDirectory() as tmp:
                with self.assertRaisesRegex(ValueError, "thin ARM64"):
                    gpu.candidate(*inputs(Path(tmp), ar(body)).values())
        with tempfile.TemporaryDirectory() as tmp:
            with self.assertRaisesRegex(ValueError, "nonempty"):
                gpu.candidate(*inputs(Path(tmp), ar(b"index", "__.SYMDEF/")).values())

    def test_source_and_portable_modules_are_not_changed_by_overlay(self):
        for name in ["go.mod", "examples/automation/go.mod", "examples/sdk-contract/go.mod"]:
            self.assertIsNone(gpu.overlay_modfile(name, b"module example.com/portable\n"))
        data = b"module example.com/native\nrequire github.com/dshills/vellum/drivers/cogent v0.0.0\n"
        output = gpu.overlay_modfile("examples/hello/go.mod", data)
        self.assertIn(b"=> ../../drivers/cogent/build/libs-darwin", output)
        self.assertNotEqual(output, data)
        with self.assertRaisesRegex(ValueError, "already has"):
            gpu.overlay_modfile("examples/hello/go.mod", output)

    def test_private_archive_relocates_and_actual_consumer_graph_requires_candidate(self):
        with tempfile.TemporaryDirectory(prefix="vellum GPU SDK spaces ") as tmp:
            base = Path(tmp)
            root, files = source(base)
            before = {name: path.read_bytes() for name, path in files.items()}
            options = inputs(base)
            first = package(root, files, base / "first.tar.gz", **options)
            second = package(root, files, base / "second.tar.gz", **options)
            self.assertEqual(first["sha256"], second["sha256"])
            self.assertEqual(before, {name: path.read_bytes() for name, path in files.items()})
            verified = checker.verify_archive(base / "first.tar.gz", first["sha256"])
            record = verified["manifest"]["native_gpu"]
            self.assertEqual(record["go_mod_overlays"]["drivers/cogent/go.mod"]["source_sha256"], gpu.digest(before["drivers/cogent/go.mod"]))
            sdk = checker.extract_verified(base / "first.tar.gz", base / "relocated SDK", verified)
            shutil.rmtree(root)
            env = dict(os.environ, GOWORK="off", GOENV="off", GOFLAGS="-mod=readonly", GOTOOLCHAIN="local", GOPROXY="off", GOSUMDB="off")
            consumer = base / "independent consumer"
            made = subprocess.run([sys.executable, str(sdk / "scripts/new-app.py"), str(consumer)], env=env, capture_output=True, timeout=30)
            self.assertEqual(made.returncode, 0, made.stderr.decode())
            graph = subprocess.run(["go", "list", "-m", "-json", "all"], cwd=consumer, env=env, capture_output=True, timeout=30)
            self.assertEqual(graph.returncode, 0, graph.stderr.decode())
            modules = checker.decode_objects(graph.stdout.decode())
            origins = checker.check_origin(modules, sdk, gpu=record)
            self.assertEqual(origins[gpu.MODULE], gpu.DIRECTORY)
            with self.assertRaisesRegex(ValueError, "missing explicit"):
                checker.check_origin([m for m in modules if m["Path"] != gpu.MODULE], sdk, gpu=record)
            changed = copy.deepcopy(modules)
            next(m for m in changed if m["Path"] == gpu.MODULE)["Replace"]["Dir"] = str(sdk / "wrong")
            with self.assertRaisesRegex(ValueError, "wrong directory"):
                checker.check_origin(changed, sdk, gpu=record)
            self.assertIsNotNone(gpu.sdk_candidate(sdk))
            (sdk / gpu.DIRECTORY / "arm64/libwgpu_native.dylib").write_bytes(b"not allowed")
            with self.assertRaisesRegex(ValueError, "extra or missing"):
                gpu.sdk_candidate(sdk)
            (sdk / gpu.DIRECTORY / "arm64/libwgpu_native.dylib").unlink()
            (sdk / gpu.DIRECTORY / "unexpected-directory").mkdir()
            with self.assertRaisesRegex(ValueError, "extra or missing"):
                gpu.sdk_candidate(sdk)
            (sdk / gpu.DIRECTORY / "unexpected-directory").rmdir()
            (sdk / gpu.DIRECTORY / gpu.LIBRARY).write_bytes(b"changed")
            with self.assertRaisesRegex(ValueError, "checksum mismatch"):
                gpu.sdk_candidate(sdk)

    def test_manifest_record_cannot_drop_capsule_or_claim_public_approval(self):
        with tempfile.TemporaryDirectory() as tmp:
            base = Path(tmp)
            record, generated = gpu.candidate(*inputs(base).values())
            generated["drivers/cogent/go.mod"] = b"module example.com/driver\n"
            files = {name: gpu.digest(data) for name, data in generated.items()}
            record["go_mod_overlays"] = {"drivers/cogent/go.mod": {"source_sha256": "a" * 64, "archived_sha256": files["drivers/cogent/go.mod"]}}
            gpu.validate_record(record, files, True)
            for bad, private in [(None, True), (record, False), (dict(record, distribution_audit="approved"), True)]:
                with self.assertRaises(ValueError):
                    gpu.validate_record(bad, files, private)
            missing = dict(files)
            del missing[gpu.DIRECTORY + "/" + gpu.PROVENANCE]
            with self.assertRaisesRegex(ValueError, "coverage"):
                gpu.validate_record(record, missing, True)


if __name__ == "__main__":
    unittest.main()
