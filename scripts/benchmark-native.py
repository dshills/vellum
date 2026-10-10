#!/usr/bin/env python3
"""Measure the separately built hello process; JSON lines are host observations."""
import argparse
import json
import os
from pathlib import Path
import platform
import selectors
import subprocess
import time


def cpu_seconds(value):
    days, clock = (value.split("-", 1) if "-" in value else ("0", value))
    fields = list(map(float, clock.split(":")))
    result = 0.0
    for field in fields:
        result = result * 60 + field
    return int(days) * 86400 + result


def linux_stat_sample(value, ticks, page_size):
    # Fields are documented at https://www.kernel.org/doc/html/latest/filesystems/proc.html
    # Split after comm, which can contain spaces or closing parentheses.
    fields = value.rsplit(")", 1)[1].split()
    if len(fields) < 22 or ticks <= 0 or page_size <= 0:
        raise ValueError("Invalid Linux process accounting data")
    return {"rss_bytes": max(0, int(fields[21])) * page_size,
            "cpu_seconds": (int(fields[11]) + int(fields[12])) / ticks,
            "cpu_time_quantum_seconds": 1 / ticks, "cpu_sampling_source": "proc-stat"}


def sample(pid):
    if platform.system() == "Linux":
        return linux_stat_sample(Path(f"/proc/{pid}/stat").read_text(),
                                 os.sysconf("SC_CLK_TCK"), os.sysconf("SC_PAGE_SIZE"))
    output = subprocess.check_output(["ps", "-o", "rss=", "-o", "time=", "-p", str(pid)], text=True)
    rss, cpu = output.split()
    quantum = 10 ** -len(cpu.rsplit(".", 1)[1]) if "." in cpu else 1
    return {"rss_bytes": int(rss) * 1024, "cpu_seconds": cpu_seconds(cpu),
            "cpu_time_quantum_seconds": quantum, "cpu_sampling_source": "ps"}


def read_ready(process, selector, deadline):
    # Read only OS-ready bytes: buffered text readline can prefetch additional
    # lines behind the selector, or wait indefinitely for a partial line's end.
    pending = bytearray()
    while time.monotonic() < deadline:
        remaining = deadline - time.monotonic()
        if remaining <= 0:
            break
        if selector.select(timeout=min(.2, remaining)):
            chunk = os.read(process.stdout.fileno(), 65536)
            if not chunk:
                break
            pending.extend(chunk)
            while True:
                newline = pending.find(b"\n")
                if newline < 0:
                    break
                line = bytes(pending[:newline])
                del pending[:newline + 1]
                try:
                    candidate = json.loads(line)
                except (json.JSONDecodeError, UnicodeDecodeError):
                    continue
                if isinstance(candidate, dict) and candidate.get("event") == "vellum.native.ready":
                    return candidate
        elif process.poll() is not None:
            break
    return None


def run(binary, idle, samples):
    import tempfile
    with tempfile.TemporaryFile(mode="w+") as errors:
        started = time.monotonic()
        process = subprocess.Popen([str(binary), "-benchmark", f"-idle={idle}s", f"-samples={samples}"],
                                   stdout=subprocess.PIPE, stderr=errors, bufsize=0)
        selector = selectors.DefaultSelector()
        selector.register(process.stdout, selectors.EVENT_READ)
        try:
            ready = read_ready(process, selector, started + 40)
            if ready is None:
                raise RuntimeError("Native example did not report a verified GPU frame")
            report_received = time.monotonic()
            # Let render/readback notifications settle before quiet CPU sampling.
            time.sleep(min(.5, idle * .1))
            first = sample(process.pid)
            idle_start = time.monotonic()
            # Leave a margin before the example's timed shutdown.
            time.sleep(idle * .6)
            last = sample(process.pid)
            idle_elapsed = time.monotonic() - idle_start
            ready.update({
                "host": platform.platform(), "machine": platform.machine(),
                "startup_process_to_report_ns": int((report_received-started)*1e9),
                "idle_sample_seconds": idle_elapsed,
                "cpu_sampling_source": first.get("cpu_sampling_source", "unknown"),
                "cpu_time_quantum_seconds": first.get("cpu_time_quantum_seconds"),
                "idle_cpu_seconds": last["cpu_seconds"]-first["cpu_seconds"],
                "idle_cpu_percent_one_core": 100*(last["cpu_seconds"]-first["cpu_seconds"])/idle_elapsed,
                "rss_bytes_before_idle": first["rss_bytes"], "rss_bytes_after_idle": last["rss_bytes"],
            })
            # The app starts its idle shutdown timer after reporting readiness.
            # Long accepted intervals still need their unsampled remainder, plus
            # a fixed cleanup margin; a fixed ten seconds kills a 60s run early.
            remaining_idle = max(0, idle - (time.monotonic() - report_received))
            process.communicate(timeout=remaining_idle + 10)
            if process.returncode:
                raise RuntimeError(f"Native example exited {process.returncode}")
            return ready
        finally:
            selector.close()
            if process.poll() is None:
                process.kill()
                process.communicate()
            errors.seek(0)
            diagnostics = errors.read()
            if process.returncode:
                import sys
                print(diagnostics, file=sys.stderr)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--runs", type=int, default=3)
    parser.add_argument("--idle", type=float, default=3)
    parser.add_argument("--samples", type=int, default=10)
    parser.add_argument("--binary", type=Path, default=Path(__file__).resolve().parents[1]/"examples/hello/build/vellum-hello")
    args = parser.parse_args()
    if args.runs < 1 or args.idle < 1 or not 1 <= args.samples <= 1000:
        parser.error("positive runs, idle >= 1 second and 1..1000 samples required")
    for index in range(args.runs):
        measured = run(args.binary.resolve(), args.idle, args.samples)
        measured["run"] = index + 1
        print(json.dumps(measured), flush=True)


if __name__ == "__main__":
    main()
