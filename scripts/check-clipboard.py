#!/usr/bin/env python3
"""Opt-in Mac clipboard acceptance with an independent preservation guardian.

Original raw pasteboard bytes live only in protected per-job original.plist. They
are never copied into outputs, hashes, logs, reviews or archives. After verified
restoration the guardian deletes that backup. Uncertain ownership retains it.
"""
import argparse
import datetime
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import time

from child_process import run_child
from sdk_environment import child_environment

ROOT = Path(__file__).resolve().parents[1]
PUBLIC_JOB_FILES = ("config.json", "state.json", "child.pid", "expected.json", "summary.json", "child.stdout", "child.stderr")

def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def process(arguments, directory, name, env, timeout):
    start = time.monotonic()
    receipt = {"argv": list(arguments), "cwd": str(ROOT), "exit_code": None, "error": None}
    with (directory / (name + ".stdout")).open("xb") as stdout, (directory / (name + ".stderr")).open("xb") as stderr:
        try:
            receipt["exit_code"] = run_child(arguments, cwd=ROOT, env=env, stdout=stdout, stderr=stderr, timeout=timeout).returncode
        except BaseException as error:
            # These errors contain only argv/process information, never backup data.
            receipt["error"] = {"type": type(error).__name__}
            if not isinstance(error, Exception):
                raise
        finally:
            receipt["elapsed_seconds"] = time.monotonic() - start
            for channel in ("stdout", "stderr"):
                receipt[channel + "_sha256"] = digest(directory / (name + "." + channel))
            (directory / (name + ".json")).write_text(json.dumps(receipt, indent=2) + "\n")
    return receipt

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--live", action="store_true", help="explicitly authorize general pasteboard writes while the computer is available")
    parser.add_argument("--self-test", action="store_true", help="private named pasteboard only; no general clipboard access")
    parser.add_argument("--case", choices=("normal", "exit-after-owned", "kill-after-owned"), default="normal")
    parser.add_argument("--guardian", type=Path, help="use an already built guardian")
    parser.add_argument("--binary", type=Path, help="use an already built Go fixture")
    args = parser.parse_args()
    if sys.platform != "darwin" or args.live == args.self_test:
        parser.error("choose exactly one of --live or --self-test on macOS")
    output = args.output.resolve()
    output.mkdir(mode=0o700, parents=True, exist_ok=False)
    private = output / "private"
    private.mkdir(mode=0o700)
    job = private / "job"
    job.mkdir(mode=0o700)
    env = child_environment(runtime=True)
    guardian = args.guardian.resolve() if args.guardian else output / "guardian"
    binary = args.binary.resolve() if args.binary else output / "clipboard-smoke"
    selected = ["scripts/check-clipboard.py", "scripts/check-clipboard.swift", "scripts/check-clipboard-native.h", "scripts/check-clipboard-native.m", "drivers/cogent/native_clipboard_checks.h", "drivers/cogent/native_clipboard_checks_darwin.go", "drivers/cogent/native_clipboard_checks_darwin.m", "drivers/cogent/clipboard_darwin.m", "drivers/cogent/cmd/clipboard-smoke/main.go"]
    before = {name: digest(ROOT / name) for name in selected}
    builds = []
    if args.guardian is None:
        shim = output / "guardian-native.o"
        builds.append(process(["xcrun", "clang", "-fobjc-arc", "-c", str(ROOT / "scripts/check-clipboard-native.m"), "-o", str(shim)], output, "build-guardian-native", env, 120))
        if builds[-1]["exit_code"] == 0 and builds[-1]["error"] is None:
            builds.append(process(["xcrun", "swiftc", "-module-cache-path", str(output / "swift-cache"), "-import-objc-header", str(ROOT / "scripts/check-clipboard-native.h"), str(shim), str(ROOT / "scripts/check-clipboard.swift"), "-o", str(guardian)], output, "build-guardian", env, 120))
    if args.binary is None:
        builds.append(process(["go", "build", "-C", str(ROOT / "drivers/cogent"), "-race", "-tags=vellum_native_checks", "-o", str(binary), "./cmd/clipboard-smoke"], output, "build-fixture", env, 120))
    result = {"started_utc": datetime.datetime.now(datetime.timezone.utc).isoformat(), "live": args.live, "case": args.case, "sources_before": before, "builds": builds, "original_content_logged": False, "backup_content_hashed": False, "result": "NOT_RUN"}
    if any(item["exit_code"] != 0 or item["error"] for item in builds):
        result["result"] = "FAILED_BUILD"
    elif not guardian.is_file() or not binary.is_file():
        result["result"] = "MISSING_BINARY"
    else:
        result["guardian_sha256"] = digest(guardian)
        result["fixture_sha256"] = digest(binary)
        if args.self_test:
            native = private / "native"
            native.mkdir(mode=0o700)
            with (native / "state.json").open("xb") as file:
                os.fchmod(file.fileno(), 0o600)
                file.write(b"{}\n")
                file.flush()
                os.fsync(file.fileno())
            safety = process([str(binary), "-safety-state", str(native / "state.json")], output, "native-safety", env, 30)
            result["native_guard_safety"] = safety
        arguments = [str(guardian), "--self-test", str(job)] if args.self_test else [str(guardian), "--run", str(job), str(binary), args.case]
        backup = job / "original.plist"
        # The bounded process helper terminates its owned group before returning
        # on timeout/interruption. Recovery independently checks the durable child
        # PID has disappeared and obtains the global guardian lock before writing.
        try:
            receipt = process(arguments, output, "guardian-run", env, 180)
            result["process"] = receipt
        finally:
            if args.live and backup.exists():
                result["recovery"] = process([str(guardian), "--recover", str(job)], output, "guardian-recovery", env, 60)
        # Only these fixture-owned files may leave the protected job directory.
        # Never enumerate/read/hash/copy original.plist or arbitrary private files.
        for name in PUBLIC_JOB_FILES:
            source = job / name
            if source.is_file():
                (output / ("observed-" + name)).write_bytes(source.read_bytes())
        summary_path = job / "summary.json"
        try:
            summary = json.loads(summary_path.read_text()) if summary_path.exists() else None
            if summary is not None and not isinstance(summary, dict):
                raise ValueError("invalid summary shape")
        except (OSError, ValueError):
            summary = {"result": "UNREADABLE_SUMMARY"}
        result["summary"] = summary
        result["protected_backup_retained"] = backup.exists()
        native_safe = not args.self_test or result["native_guard_safety"]["exit_code"] == 0 and result["native_guard_safety"]["error"] is None
        if native_safe and receipt["exit_code"] == 0 and receipt["error"] is None and not backup.exists() and summary and summary.get("result") in ("PASS_GENERAL_CLIPBOARD_RESTORED", "PASS_PRIVATE_GUARDIAN_SAFETY"):
            result["result"] = "PASS"
        else:
            result["result"] = "FAILED_OR_REFUSED"
    result["sources_after"] = {name: digest(ROOT / name) for name in selected}
    result["source_stable"] = result["sources_before"] == result["sources_after"]
    if not result["source_stable"]:
        result["result"] = "SOURCE_CHANGED"
    (output / "result.json").write_text(json.dumps(result, indent=2) + "\n")
    print(json.dumps({"result": result["result"], "live": args.live, "case": args.case, "protected_backup_retained": result.get("protected_backup_retained"), "summary": result.get("summary"), "source_stable": result["source_stable"]}))
    return 0 if result["result"] == "PASS" else 1

if __name__ == "__main__":
    sys.exit(main())
