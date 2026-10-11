#!/usr/bin/env python3
"""Run the opt-in loadedgrid release benchmark; retain every native raw record.

Only --run launches windows. Statistics and record validation are independently
CPU-testable. Measurements are process/Go/UI-scope observations, not physical
presentation, native allocation attribution or a complete P6 qualification.
"""
import argparse
import hashlib
import json
import math
import os
from pathlib import Path
import platform
import selectors
import statistics
import subprocess
import sys
import time

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))
from child_process import ChildCleanupError, terminate_owned_group
from sdk_environment import child_environment

METRICS = ("ui_work_ns", "input_ns", "owner_ns", "reconcile_ns", "layout_paint_ns", "gather_ns", "raster_upload_ns", "encode_submit_ns", "queued_ns", "acquire_ns", "present_call_ns")
NATIVE_PHASES = {"ui_work_ns": "TotalUIWorkNS", "input_ns": "InputNS", "owner_ns": "OwnerNS", "reconcile_ns": "ReconcileNS", "layout_paint_ns": "LayoutPaintNS", "gather_ns": "GatherNS", "raster_upload_ns": "RasterUploadNS", "encode_submit_ns": "EncodeSubmitNS", "queued_ns": "QueuedNS", "acquire_ns": "AcquireNS", "present_call_ns": "PresentCallNS"}
WORK_PHASES = ("input_ns", "owner_ns", "reconcile_ns", "layout_paint_ns", "gather_ns", "raster_upload_ns", "encode_submit_ns")
MAX_LINE = 1 << 20
MAX_OUTPUT = 64 << 20


def distribution(values):
    """Nearest-rank percentiles; never sum independent phase percentiles."""
    if not values or any(isinstance(v, bool) or not isinstance(v, (int, float)) or not math.isfinite(v) or v < 0 for v in values):
        raise ValueError("finite nonnegative samples required")
    ordered = sorted(values)
    def percentile(p):
        return ordered[max(0, math.ceil(p * len(ordered)) - 1)]
    mean = statistics.mean(ordered)
    sd = statistics.stdev(ordered) if len(ordered) > 1 else 0.0
    return {"count": len(ordered), "p50": percentile(.50), "p95": percentile(.95), "p99": percentile(.99),
            "max": ordered[-1], "min": ordered[0], "mean": mean, "sample_stddev": sd,
            "coefficient_of_variation": sd / mean if mean else 0.0}


def validate_records(records, rows, variable, samples, novel, *, allow_profile=False):
    """Require exact case, current completed geometry and one sample per tag."""
    if len(records) != samples + 3 or [r.get("event") for r in records[:2]] != ["grid.perf.setup", "grid.perf.ready"] or records[-1].get("event") != "grid.perf.end":
        raise ValueError("missing, repeated or unordered benchmark stages")
    setup, ready, end = records[0], records[1], records[-1]
    for r in (setup, ready, end):
        if r.get("rows") != rows or r.get("variable") is not variable or r.get("novel") is not novel:
            raise ValueError("benchmark case identity mismatch")
    profiled = setup.get("cpu_profile", False)
    if not isinstance(profiled, bool) or profiled and not allow_profile:
        raise ValueError("diagnostic CPU profile cannot qualify warm timing")
    if any(r.get("cpu_profile", False) is not profiled for r in (ready, end)):
        raise ValueError("CPU profiling case identity mismatch")
    if setup.get("release") is not True or setup.get("stored_value_bytes") != rows * 20 * 8 or setup.get("seed") != 7:
        raise ValueError("release stored-data fixture required")
    if end.get("samples") != samples or end.get("dropped") != 0 or end.get("pending") != 0 or end.get("callbacks") != samples:
        raise ValueError("sample/callback/recorder completion mismatch")
    result = []
    last_generation = last_frame = 0
    for i, record in enumerate(records[2:-1]):
        if record.get("event") != "grid.perf.sample" or record.get("index") != i or record.get("tag") != i + 1:
            raise ValueError("sample ordering/tag mismatch")
        geometry = record.get("geometry", {})
        counts = [geometry.get(k) for k in ("cells", "rows", "columns")]
        if any(isinstance(n, bool) or not isinstance(n, int) for n in counts) or counts[0] < 400 or counts[1] < 50 or counts[2] < 8:
            raise ValueError("insufficient actually clipped rendered cells")
        pool = [geometry.get(k) for k in ("materialized_cells", "retained_cells", "retained_headers")]
        if any(isinstance(n, bool) or not isinstance(n, int) or n < 0 for n in pool) or pool[0] < counts[0] or pool[1] < pool[0] or pool[1] > 4096 or pool[2] > 64:
            raise ValueError("invalid materialized/retained native pool observation")
        digest = geometry.get("layout_sha256", "")
        if len(digest) != 64 or any(c not in "0123456789abcdef" for c in digest):
            raise ValueError("invalid measured layout identity")
        identities = [record.get(k) for k in ("generation", "frame")]
        if any(isinstance(n, bool) or not isinstance(n, int) or n <= 0 for n in identities):
            raise ValueError("missing completed native generation/frame")
        if record["generation"] <= last_generation or record["frame"] <= last_frame:
            raise ValueError("repeated or backward native generation/frame")
        last_generation, last_frame = record["generation"], record["frame"]
        native = record.get("native_sample", {})
        if native.get("Tag") != i+1 or native.get("Generation") != record["generation"] or native.get("Window") != "perf" or native.get("Completed") is not True or native.get("Failure"):
            raise ValueError("raw native ticket/frame mismatch")
        if record.get("callbacks") != 1:
            raise ValueError("scroll failed or callback executed more than once")
        metrics = record.get("metrics", {})
        if any(k not in metrics for k in METRICS):
            raise ValueError("incomplete actual timing phases")
        for name in METRICS:
            value = metrics[name]
            if isinstance(value, bool) or not isinstance(value, int) or value < 0 or value != native.get(NATIVE_PHASES[name]):
                raise ValueError("reported timing differs from raw native sample")
        if metrics["ui_work_ns"] != sum(metrics[k] for k in WORK_PHASES):
            raise ValueError("same-ticket total does not match measured work phases")
        result.append(record)
    return result


def summarize(runs):
    result = {}
    for name in METRICS:
        individual = [distribution([s["metrics"][name] for s in run]) for run in runs]
        result[name] = {"pooled": distribution([s["metrics"][name] for run in runs for s in run]),
                        "per_run": individual, "run_p95_variation": distribution([r["p95"] for r in individual])}
    return result


def compare_geometry(small, large):
    if len(small) != len(large):
        raise ValueError("paired rowcount sample cardinality mismatch")
    for a, b in zip(small, large):
        ga, gb = a["geometry"], b["geometry"]
        fields = ("cells", "rows", "columns", "scale", "layout_sha256")
        if any(ga[k] != gb[k] for k in fields):
            raise ValueError("10k/million rendered workload differs")


def process_sample(pid):
    # Darwin ps reports current RSS and aggregate CPU time with a finite quantum.
    completed = subprocess.run(["ps", "-o", "rss=", "-o", "time=", "-p", str(pid)], capture_output=True, text=True, timeout=2, check=True)
    rss, cpu = completed.stdout.split()
    days, clock = (cpu.split("-", 1) if "-" in cpu else ("0", cpu))
    seconds = 0.0
    for field in clock.split(":"):
        seconds = seconds * 60 + float(field)
    quantum = 10 ** -len(cpu.rsplit(".", 1)[1]) if "." in cpu else 1
    return {"rss_bytes": int(rss) * 1024, "cpu_seconds": int(days) * 86400 + seconds,
            "cpu_quantum_seconds": quantum, "source": "ps; RSS is process-wide"}


def run_case(binary, directory, rows, variable, samples, warmup, novel, timeout, *, cpu_profile=None):
    directory.mkdir(parents=True, exist_ok=False)
    argv = [str(binary), "-perf", f"-rows={rows}", f"-perf-samples={samples}", f"-perf-warmup={warmup}"]
    if cpu_profile is not None:
        argv.append(f"-perf-cpu-profile={cpu_profile}")
    if variable: argv.append("-variable-rows")
    if novel: argv.append("-perf-novel")
    process_options = {"start_new_session": True} if os.name != "nt" else {"creationflags": subprocess.CREATE_NEW_PROCESS_GROUP}
    records, observations = [], []
    started = time.monotonic()
    pending = bytearray()
    total = 0
    returncode = None
    failure = None
    process = selector = None
    with (directory / "native.log").open("xb") as raw:
        try:
            process = subprocess.Popen(argv, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                                       env=child_environment(runtime=True), bufsize=0, **process_options)
            selector = selectors.DefaultSelector()
            selector.register(process.stdout, selectors.EVENT_READ)
            while True:
                remaining = timeout - (time.monotonic() - started)
                if remaining <= 0: raise subprocess.TimeoutExpired(argv, timeout)
                if selector.select(timeout=min(.2, remaining)):
                    chunk = os.read(process.stdout.fileno(), 65536)
                    if not chunk: break
                    raw.write(chunk)
                    total += len(chunk)
                    if total > MAX_OUTPUT: raise ValueError("native benchmark output budget exceeded")
                    pending.extend(chunk)
                    while b"\n" in pending:
                        line, _, remainder = pending.partition(b"\n")
                        pending = bytearray(remainder)
                        if len(line) > MAX_LINE: raise ValueError("native benchmark line budget exceeded")
                        try: record = json.loads(line)
                        except (ValueError, UnicodeDecodeError): continue
                        if isinstance(record, dict) and str(record.get("event", "")).startswith("grid.perf."):
                            records.append(record)
                            # Pipe parsing can lag the child: even a setup-stage
                            # observer could overlap measured native frames.
                            # End is emitted after every measured sample, so no
                            # external observer is spawned before completion.
                            if record["event"] == "grid.perf.end":
                                try: observation = process_sample(process.pid)
                                except (OSError, subprocess.SubprocessError, ValueError) as error:
                                    observation = {"unavailable": type(error).__name__}
                                observations.append({"event": record["event"], "elapsed_seconds": time.monotonic() - started, **observation})
                    if len(pending) > MAX_LINE: raise ValueError("native benchmark unterminated line budget exceeded")
                elif process.poll() is not None:
                    # Continue draining the pipe until EOF, including buffered final stage.
                    continue
            returncode = process.wait(timeout=max(.01, timeout - (time.monotonic() - started)))
            if pending: raise ValueError("incomplete benchmark record at EOF")
            if returncode != 0: raise RuntimeError(f"native benchmark exited {returncode}")
            selected = validate_records(records, rows, variable, samples, novel, allow_profile=cpu_profile is not None)
            if records[0].get("cpu_profile", False) is not (cpu_profile is not None):
                raise ValueError("requested CPU profile case mismatch")
        except BaseException as error:
            failure = f"{type(error).__name__}: {error}"
            raise
        finally:
            cleanup_error = None
            if selector is not None:
                try:
                    selector.close()
                except Exception as error:
                    cleanup_error = error
            # Setup failures also retire any child already launched. Descendants
            # can own the pipe after direct-child exit, so failure always cleans
            # the owned group rather than relying on the direct child's state.
            if process is not None:
                if failure or cleanup_error or process.poll() is None:
                    try:
                        terminate_owned_group(process)
                    except Exception as error:
                        cleanup_error = cleanup_error or error
                        if process.poll() is None: process.kill()
                    try:
                        process.wait(timeout=5)
                    except Exception as error:
                        cleanup_error = cleanup_error or error
                if process.stdout is not None:
                    try:
                        process.stdout.close()
                    except Exception as error:
                        cleanup_error = cleanup_error or error
                returncode = process.returncode
            raw.flush()
            if cleanup_error:
                failure = (failure + "; " if failure else "") + f"cleanup failed: {type(cleanup_error).__name__}"
            outcome = {"argv": argv, "exit_code": returncode, "elapsed_seconds": time.monotonic() - started,
                       "records": records, "process_observations": observations,
                       "process_observation_scope": {"stages": ["grid.perf.end"],
                           "baseline_rss": "unavailable", "warm_rss": "unavailable", "cpu_delta": "unavailable",
                           "note": "External ps runs only after the completed end record; no external process observer runs during measured native frames."},
                       "result": "FAILED" if failure else "PASS", "error": failure,
                       "log_sha256": hashlib.sha256((directory / "native.log").read_bytes()).hexdigest()}
            (directory / "result.json").write_text(json.dumps(outcome, indent=2) + "\n")
            if cleanup_error:
                raise ChildCleanupError("benchmark owned process cleanup failed") from cleanup_error
    return selected, outcome


def metadata_command(argv, timeout=10):
    try:
        result = subprocess.run(argv, capture_output=True, text=True, timeout=timeout, check=False)
        return {"argv": argv, "exit_code": result.returncode, "stdout": result.stdout, "stderr": result.stderr}
    except (OSError, subprocess.SubprocessError) as error:
        return {"argv": argv, "unavailable": type(error).__name__}


def selected_inputs():
    # Compact selected source identity, not a compiler, module-cache or SDK
    # closure. The integration qualifier binds complete prepared dependencies.
    root = HERE.parent
    candidates = list((root / "examples/loadedgrid").glob("*.go"))
    candidates += list((root / "drivers/cogent").glob("frame_metrics*.go"))
    candidates += list((root / "drivers/cogent").glob("grid*.go"))
    candidates += [root / path for path in ("go.mod", "go.sum", "examples/loadedgrid/go.mod", "examples/loadedgrid/go.sum",
                    "drivers/cogent/driver.go", "drivers/cogent/native/dependencies.json", "scripts/benchmark-grid.py")]
    return {str(path.relative_to(root)): hashlib.sha256(path.read_bytes()).hexdigest() for path in sorted(set(candidates)) if path.is_file()}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--run", action="store_true", help="explicitly launch visible Mac benchmark processes")
    parser.add_argument("--binary", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True, help="new directory; raw results are never overwritten")
    parser.add_argument("--runs", type=int, default=5)
    parser.add_argument("--samples", type=int, default=120)
    parser.add_argument("--warmup", type=int, default=128)
    parser.add_argument("--novel", action="store_true", help="separate newly visited row strings; not a cold-font claim")
    parser.add_argument("--timeout", type=float, default=180)
    args = parser.parse_args()
    if not args.run: parser.error("--run is required; this tool launches visible native windows")
    if platform.system() != "Darwin": parser.error("this checkpoint measures the actual Mac backend")
    if not 5 <= args.runs <= 20 or not 120 <= args.samples <= 2048 or not 64 <= args.warmup <= 2048 or args.warmup % 64 or not math.isfinite(args.timeout) or not 10 <= args.timeout <= 600:
        parser.error("runs5..20, samples120..2048, warmup64..2048 in multiples64, timeout10..600 required")
    binary = args.binary.resolve(strict=True)
    args.output.mkdir(parents=True, exist_ok=False)
    source_before = selected_inputs()
    machine = {"cpu": metadata_command(["sysctl", "-n", "machdep.cpu.brand_string"]),
               "ram": metadata_command(["sysctl", "-n", "hw.memsize"]),
               "displays_gpu": metadata_command(["system_profiler", "SPDisplaysDataType", "-json"], 15),
               "power_configuration": metadata_command(["pmset", "-g", "custom"]),
               "binary_build": metadata_command(["go", "version", "-m", str(binary)])}
    (args.output / "machine.json").write_text(json.dumps(machine, indent=2) + "\n")
    report = {"schema": "vellum-grid-performance-v1", "binary": str(binary),
              "binary_sha256": hashlib.sha256(binary.read_bytes()).hexdigest(),
              "host": platform.platform(), "machine": platform.machine(), "machine_metadata": "machine.json",
              "selected_inputs_before": source_before, "input_scope": "selected sources, not complete compiler/dependency closure", "runs": args.runs,
              "samples_per_run": args.samples, "warmup": args.warmup, "novel": args.novel, "cases": {},
              "scope": "release actual widget routing/frame elapsed work; end-only process-wide RSS/CPU and staged Go observations; baseline/warm process RSS, process CPU deltas, native/GPU allocations and compositor presentation unavailable"}
    try:
        for variable in (False, True):
            results = {}
            for rows in (10_000, 1_000_000):
                case = f"{'variable' if variable else 'fixed'}-{rows}"
                all_runs, raw = [], []
                for index in range(args.runs):
                    measured, outcome = run_case(binary, args.output / case / f"run-{index+1}", rows, variable, args.samples, args.warmup, args.novel, args.timeout)
                    all_runs.append(measured); raw.append(outcome)
                    if hashlib.sha256(binary.read_bytes()).hexdigest() != report["binary_sha256"]:
                        raise ValueError("binary changed during measurement")
                results[rows] = all_runs
                report["cases"][case] = {"timings_ns": summarize(all_runs), "runs": raw,
                    "go_process_alloc_bytes": distribution([s["go_process_alloc_bytes"] for run in all_runs for s in run]),
                    "go_process_mallocs": distribution([s["go_process_mallocs"] for run in all_runs for s in run]),
                    "source_callbacks": {k: distribution([s["source_callbacks"][k] for run in all_runs for s in run]) for k in ("Keys", "Inverse", "Cells")}}
            for small, large in zip(results[10_000], results[1_000_000]): compare_geometry(small, large)
        source_after = selected_inputs()
        report["selected_inputs_after"] = source_after
        if source_before != source_after: raise ValueError("selected source changed during measurement")
        report["result"] = "MEASURED"
        # Targets are fixed provisional P6 thresholds, not changed by this tool.
        report["warm_targets_ns"] = {"p95": 4_000_000, "p99": 6_000_000}
        report["warm_target_met"] = (not args.novel and all(all(r["p95"] <= 4_000_000 and r["p99"] <= 6_000_000 for r in v["timings_ns"]["ui_work_ns"]["per_run"]) for v in report["cases"].values()))
        report["rowcount_p95_ratios"] = {}
        for heights in ("fixed", "variable"):
            small = report["cases"][f"{heights}-10000"]["timings_ns"]["ui_work_ns"]["pooled"]["p95"]
            large = report["cases"][f"{heights}-1000000"]["timings_ns"]["ui_work_ns"]["pooled"]["p95"]
            report["rowcount_p95_ratios"][heights] = large / small if small else None
    except BaseException as error:
        report["result"] = "FAILED"
        report["error"] = f"{type(error).__name__}: {error}"
        raise
    finally:
        (args.output / "summary.json").write_text(json.dumps(report, indent=2) + "\n")
    print(json.dumps({"summary": str(args.output / "summary.json"), "result": report["result"], "warm_target_met": report["warm_target_met"]}))


if __name__ == "__main__": main()
