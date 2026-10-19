#!/usr/bin/env python3
"""Build a private Mac workbench and bind its binary, linker map and input inventory."""
import argparse
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import platform
import shutil
import subprocess
import sys
import tempfile

if __name__ == "__main__":
    sys.dont_write_bytecode = True

from build_output import BuildOutput, json_bytes
from child_process import run_child
from sdk_environment import child_environment

ROOT = Path(__file__).resolve().parents[1]
spec = importlib.util.spec_from_file_location("vellum_build_inventory", ROOT / "scripts/distribution-inventory.py")
inventory = importlib.util.module_from_spec(spec)
spec.loader.exec_module(inventory)
BUILD_TIMEOUT_SECONDS = 600
BINARY = "vellum-workbench"
LINK_MAP = BINARY + ".map"
RECEIPT = "build-receipt.json"


def error_detail(error):
    detail = {"type": type(error).__name__, "message": str(error)}
    if getattr(error, "build_cleanup_error", None) is not None:
        detail["cleanup_error"] = error.build_cleanup_error
    return detail


def go_identity(env):
    executable = shutil.which("go", path=env.get("PATH", ""))
    if executable is None:
        raise ValueError("Go executable is unavailable")
    executable = Path(executable).resolve(strict=True)
    return executable, inventory.file_identity(executable.parent, executable.name)


def build(root, output, include_untracked=False):
    """Observe one compiler invocation; never execute the produced application.

    Before/after checks detect sampled drift, not transient changes reverted
    during compilation. This is neither a hermetic build nor an attestation.
    Failures retain diagnostics without reporting successful completion.
    """
    if platform.system() != "Darwin":
        raise ValueError("owned Mac builds require a macOS host")
    root = root.resolve()
    output = BuildOutput(root, output, inventory.MAX_FILE_BYTES)
    command = None
    try:
        env = child_environment()
        producer = inventory.file_identity(ROOT, "scripts/build-macos.py")
        go, tool = go_identity(env)
        before = inventory.collect(root, include_untracked=include_untracked)
        if before["build_configuration"]["GOARCH"] not in {"arm64", "amd64"}:
            raise ValueError("unsupported Mac build architecture")
        output.require()
        output.write_json("inputs-before.json", before)
        # Generated paths avoid shell/Go-linker quoting ambiguity even when the
        # user-selected output directory contains spaces or punctuation.
        with tempfile.TemporaryDirectory(prefix="vellum-macos-build-", dir="/private/tmp") as temporary:
            stage = Path(temporary)
            binary, link_map = stage / BINARY, stage / LINK_MAP
            if any(path.exists() or path.is_symlink() for path in (binary, link_map)):
                raise ValueError("compiler output paths were not initially absent")
            argv = [str(go), "build", "-ldflags=-linkmode=external -extldflags=-Wl,-map," + str(link_map), "-o", str(binary), "./cmd/workbench"]
            command = {"argv": argv, "cwd": str(root / "drivers/cogent"), "timeout_seconds": BUILD_TIMEOUT_SECONDS, "outputs_initially_absent": True, "exit_code": None}
            output.require()
            with output.create("build.stdout") as stdout, output.create("build.stderr") as stderr:
                result = run_child(argv, cwd=root / "drivers/cogent", env=env, stdout=stdout, stderr=stderr, timeout=BUILD_TIMEOUT_SECONDS, log_limit_bytes=4 * 1024 * 1024)
            command["exit_code"] = result.returncode
            if result.returncode:
                raise subprocess.CalledProcessError(result.returncode, argv)
            log_identities = {"build." + channel: {"file": "build." + channel, **identity} for channel, identity in result.log_summaries.items()}
            for name, identity in log_identities.items():
                if output.identity(name) != identity:
                    raise ValueError("compiler log differs from streamed bytes")
            emitted = {name: inventory.file_identity(stage, name) for name in (BINARY, LINK_MAP)}
            after = inventory.collect(root, include_untracked=include_untracked, binary=binary, link_map=link_map)
            observation = after["native_link_observation"]
            if {key: value for key, value in after.items() if key != "native_link_observation"} != before:
                raise ValueError("build inputs changed across compiler invocation")
            if observation["binary"] != emitted[BINARY] or observation["link_map"] != emitted[LINK_MAP]:
                raise ValueError("compiler outputs changed during inventory collection")
            for name, identity in emitted.items():
                if inventory.file_identity(stage, name) != identity:
                    raise ValueError("compiler output changed before publication")
            if go_identity(env) != (go, tool) or child_environment() != env or inventory.file_identity(ROOT, "scripts/build-macos.py") != producer:
                raise ValueError("Go executable, environment or producer changed during compilation")
            binding = "producer-observed-same-build"
            output.require()
            output.write_json("inventory.json", after)
            for name in (BINARY, LINK_MAP):
                output.require()
                output.copy(stage / name, name, mode=0o755 if name == BINARY else 0o644)
            artifacts = {name: output.identity(name) for name in (BINARY, LINK_MAP, "inventory.json", "inputs-before.json", "build.stdout", "build.stderr")}
            expected = {**emitted, **log_identities}
            for name, value in [("inputs-before.json", before), ("inventory.json", after)]:
                data = json_bytes(value)
                expected[name] = {"file": name, "bytes": len(data), "sha256": hashlib.sha256(data).hexdigest()}
            if artifacts != expected:
                raise ValueError("published compiler artifact, JSON or log differs")
            receipt = {
                "schema": 1,
                "event": "vellum.macos.build_complete",
                "private_build": True,
                "runtime_executed": False,
                "distribution_audit": "pending",
                "signing_identity_requested": None,
                "notarized": False,
                "published": False,
                "source": before["source"],
                "include_untracked": include_untracked,
                "package_pattern": "./cmd/workbench",
                "build_configuration": before["build_configuration"],
                "go_executable": {"path": str(go), **tool},
                "producer": producer,
                "environment_keys": sorted(env),
                "environment_sha256": hashlib.sha256(json_bytes(env)).hexdigest(),
                "command": command,
                "artifacts": artifacts,
                "artifact_pair_binding": binding,
                "scope": "One observed private compiler invocation and sampled unchanged distribution inputs. Not independent build provenance, complete Rust/C closure, native rights approval or Mac runtime qualification.",
            }
            for name, identity in artifacts.items():
                if output.identity(name) != identity:
                    raise ValueError("published build input changed before completion")
            output.require()
            if set(os.listdir(output.fd)) != set(artifacts):
                raise ValueError("unexpected build output entry before completion")
            completed = {"event": receipt["event"], "output": str(output.path), "receipt_sha256": hashlib.sha256(json_bytes(receipt)).hexdigest(), "binary_sha256": emitted[BINARY]["sha256"], "artifact_pair_binding": receipt["artifact_pair_binding"], "runtime_executed": False, "distribution_audit": "pending"}
            output.write_json(RECEIPT, receipt)  # last publication
            identity = output.identity(RECEIPT)
            if identity["sha256"] != completed["receipt_sha256"] or identity["bytes"] != len(json_bytes(receipt)):
                raise ValueError("published build receipt differs")
            return completed
    except BaseException as error:
        # Keep original failures, including timeout/cleanup errors. Never write
        # diagnostic state into a replacement directory or overwrite a file.
        phase = "receipt-retirement"
        try:
            output.remove_created(RECEIPT)
            phase = "failure-diagnostics"
            output.require()
            output.write_json("build-failed.json", {"event": "vellum.macos.build_failed", "error": error_detail(error), "command": command, "runtime_executed": False})
        except Exception as cleanup:
            error.build_cleanup_error = {"phase": phase, "type": type(cleanup).__name__, "message": str(cleanup), "output_completion_unconfirmed": True, "owned_receipt_retirement_confirmed": phase != "receipt-retirement"}
            if hasattr(error, "add_note"):
                error.add_note("Build cleanup failed; output completion is unconfirmed: " + json.dumps(error.build_cleanup_error))
        raise
    finally:
        output.close()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", required=True, type=Path, help="new private output directory outside the checkout")
    parser.add_argument("--include-untracked", action="store_true", help="record nonignored whitelisted private precommit inputs")
    args = parser.parse_args()
    try:
        report = build(ROOT, args.output, args.include_untracked)
    except Exception as error:
        print(json.dumps({"event": "vellum.macos.build_failed", "error": error_detail(error), "runtime_executed": False}), file=sys.stderr)
        return 1
    print(json.dumps(report, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
