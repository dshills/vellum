#!/usr/bin/env python3
"""Read-only external AX qualification of the editor's dedicated Mac fixture.

Requires existing accessibility trust and a built examples/editor binary. Never
prompts for, enables, or changes OS accessibility/input preferences. Swift is only
needed for this optional acceptance tool, not Vellum consumers.
"""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import time

from child_process import run_child, terminate_owned_group
from sdk_environment import child_environment

ROOT = Path(__file__).resolve().parents[1]


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def report_json(output):
    reports = [json.loads(line) for line in output.splitlines() if line.startswith("{")]
    if len(reports) != 1:
        raise ValueError("Expected one inspector JSON result")
    return reports[0]


def check(binary, output):
    if sys.platform != "darwin":
        raise ValueError("External editor AX qualification requires macOS")
    binary = Path(binary).resolve(strict=True)
    output = Path(output).absolute()
    output.mkdir(parents=True, exist_ok=False)
    source = ROOT / "scripts/check-editor-accessibility.swift"
    env = child_environment(runtime=True)
    binary_hash = digest(binary)
    source_hash = digest(source)
    runner_hash = digest(Path(__file__))
    with tempfile.TemporaryDirectory(prefix="Vellum external editor AX ") as temporary:
        inspector = Path(temporary) / "inspector"
        with (output / "inspector-build.log").open("w") as log:
            run_child(["xcrun", "swiftc", str(source), "-o", str(inspector)],
                      env=env, stdout=log, stderr=subprocess.STDOUT, timeout=60, check=True)
        trusted = run_child([str(inspector), "--trust-only"], env=env,
                            capture_output=True, text=True, timeout=5)
        (output / "trust.log").write_text(trusted.stdout + trusted.stderr)
        if trusted.returncode != 0 or not report_json(trusted.stdout).get("accessibility_trusted"):
            raise ValueError("Inspector lacks existing accessibility trust; no prompt or settings change attempted")
        native_log = output / "native.log"
        with native_log.open("w") as log:
            native_started = time.monotonic()
            process = subprocess.Popen([str(binary), "-ax-fixture"], cwd=temporary,
                                       env=env, stdout=log, stderr=subprocess.STDOUT,
                                       start_new_session=True)
            try:
                record = None
                ready_deadline = time.monotonic() + 12
                while time.monotonic() < ready_deadline:
                    for line in native_log.read_text().splitlines():
                        if line.startswith("EDITOR_AX_FIXTURE "):
                            record = json.loads(line.removeprefix("EDITOR_AX_FIXTURE "))
                    if record is not None:
                        break
                    if process.poll() is not None:
                        raise ValueError("Native fixture exited before its ready record")
                    time.sleep(0.05)
                if not isinstance(record, dict) or record.get("pid") != process.pid:
                    raise ValueError("Expected the owned native fixture's PID/ready record")
                lifetime = record.get("lifetime_seconds")
                if lifetime != 60:
                    raise ValueError("Expected the dedicated 60s fixture lifetime")
                text = record.get("text")
                if not isinstance(text, str) or len(text.encode("utf-8")) > 4 << 20 or "\0" in text:
                    raise ValueError("Invalid bounded UTF-8 fixture text")
                expected = output / "expected.txt"
                expected.write_bytes(text.encode("utf-8"))
                inspector_exit = None
                inspector_error = None
                inspector_timed_out = False
                try:
                    inspected = run_child([str(inspector), str(process.pid), str(expected)],
                                          env=env, capture_output=True, text=True, timeout=35)
                    inspector_exit = inspected.returncode
                    (output / "inspector.log").write_text(inspected.stdout + inspected.stderr)
                    acceptance = report_json(inspected.stdout)
                except (ValueError, OSError, subprocess.SubprocessError) as error:
                    inspector_error = str(error)
                    inspector_timed_out = isinstance(error, subprocess.TimeoutExpired)
                    # Timeout output can be bytes even with text=True. Keep the
                    # inspector's partial diagnostics before waiting for native close.
                    if isinstance(error, subprocess.SubprocessError):
                        def decoded(value):
                            return value.decode("utf-8", errors="replace") if isinstance(value, bytes) else value or ""
                        (output / "inspector.log").write_text(decoded(getattr(error, "output", None)) + decoded(getattr(error, "stderr", None)))
                    elif not (output / "inspector.log").exists():
                        (output / "inspector.log").write_text(inspector_error + "\n")
                    acceptance = {"result": "failure", "error": inspector_error}
                # Keep the application running until its normal 60s fixture close,
                # even after an inspector failure, to record native teardown status.
                remaining = native_started + lifetime + 15 - time.monotonic()
                native_timed_out = False
                try:
                    if remaining <= 0:
                        raise subprocess.TimeoutExpired([str(binary), "-ax-fixture"], lifetime + 15)
                    process.wait(timeout=remaining)
                except subprocess.TimeoutExpired:
                    native_timed_out = True
                    terminate_owned_group(process)
                    process.wait(timeout=10)
                elapsed = time.monotonic() - native_started
                result = {"event": "vellum.editor.external-ax.acceptance", "result": "pass",
                          "native_command": [str(binary), "-ax-fixture"],
                          "native_binary_sha256": binary_hash,
                          "native_exit_code": process.returncode,
                          "native_elapsed_seconds": elapsed,
                          "announced_lifetime_seconds": lifetime,
                          "timer_lifetime_reached": elapsed >= lifetime - 0.5,
                          "inspector_exit_code": inspector_exit,
                          "inspector_error": inspector_error,
                          "inspector_timed_out": inspector_timed_out,
                          "native_timed_out": native_timed_out,
                          "inspector_source_sha256": source_hash,
                          "inspector_binary_sha256": digest(inspector),
                          "runner_source_sha256": runner_hash,
                          "fixture": record, "external_ax": acceptance,
                          "settings_changed": False}
                if inspector_exit != 0 or acceptance.get("result") != "pass" or process.returncode != 0 or native_timed_out or not result["timer_lifetime_reached"]:
                    result["result"] = "failure"
                if digest(source) != source_hash or digest(Path(__file__)) != runner_hash:
                    result["result"] = "failure"
                    result["acceptance_source_changed_during_run"] = True
                if digest(binary) != binary_hash:
                    result["result"] = "failure"
                    result["binary_changed_during_run"] = True
                (output / "acceptance.json").write_text(json.dumps(result, indent=2) + "\n")
                if result["result"] != "pass":
                    raise ValueError("External AX/native fixture failed; see acceptance.json and logs")
                return result
            finally:
                if process.poll() is None:
                    terminate_owned_group(process)
                    process.wait(timeout=10)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", type=Path, required=True, help="built examples/editor executable")
    parser.add_argument("--output", type=Path, required=True, help="new directory for immutable run evidence")
    args = parser.parse_args()
    try:
        result = check(args.binary, args.output)
    except (ValueError, OSError, subprocess.SubprocessError) as error:
        print(str(error), file=sys.stderr)
        return 1
    print(json.dumps({"result": result["result"], "output": str(args.output.absolute()),
                      "native_exit_code": result["native_exit_code"],
                      "inspector_exit_code": result["inspector_exit_code"]}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
