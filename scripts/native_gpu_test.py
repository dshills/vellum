"""Private native SDK selection, tamper refusal and actual relocated module graph."""
import copy
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
    write("drivers/cogent/go.mod", "module github.com/dshills/vellum/drivers/cogent\ngo 1.25.6\nrequire " + gpu.MODULE + " v0.0.0\n")
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
    for name, text in [("main.go", "package main\nfunc main() {}\n"), ("model.go", "package main\n"),
                       ("model_test.go", "package main\n"), ("go.sum", ""), ("LICENSE", "starter notice\n")]:
        write("examples/hello/" + name, text)
    for name in ["new-app.py", "native_gpu.py", "native_archive.py"]:
        write("scripts/" + name, (SCRIPTS / name).read_text())
    return root, files


def package(root, files, path, **kwargs):
    with mock.patch.object(release, "source_files", return_value=sorted(files.items())), \
         mock.patch.object(release, "source_state", return_value=dict(clean=True, commit="a" * 40)), \
         mock.patch.object(release, "license_inventory", return_value=({"modules": []}, {})):
        return release.package(root, path, **kwargs)


class NativeGPUChecks(unittest.TestCase):
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

    def test_all_or_none_and_public_refusal_precede_inventory_or_output(self):
        with tempfile.TemporaryDirectory() as tmp:
            base = Path(tmp)
            options = inputs(base)
            for public, kwargs in [(True, options), (False, dict(macos_gpu_library=options["macos_gpu_library"]))]:
                with mock.patch.object(release, "license_inventory") as inventory, self.assertRaisesRegex(ValueError, "private SDK"):
                    release.package(base, base / "not-created/sdk.tar.gz", public=public, **kwargs)
                inventory.assert_not_called()
                self.assertFalse((base / "not-created").exists())

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
