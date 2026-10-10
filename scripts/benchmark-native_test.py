"""Process/timing/pipe checks; no display, process launch or sleeps."""
import importlib.util
import io
import json
import os
from pathlib import Path
import subprocess
import time
import types
import unittest
from unittest import mock

ROOT = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location("benchmark_native", ROOT / "benchmark-native.py")
benchmark = importlib.util.module_from_spec(spec)
spec.loader.exec_module(benchmark)


class FakeClock:
    def __init__(self):
        self.now = 0.0

    def monotonic(self):
        return self.now

    def sleep(self, seconds):
        self.now += seconds


class FakeSelector:
    def register(self, stream, events):
        pass

    def select(self, timeout):
        return [True]

    def close(self):
        pass


class TimedExample:
    def __init__(self, clock, idle, stalled=False):
        self.clock = clock
        self.shutdown_at = idle
        self.stalled = stalled
        self.stdout = io.BytesIO((json.dumps({"event": "vellum.native.ready"}) + "\n").encode())
        self.stdout.fileno = lambda: 1234
        self.pid = 123
        self.returncode = None
        self.killed = False
        self.waits = []

    def poll(self):
        return self.returncode

    def communicate(self, timeout=None):
        if self.returncode is not None:
            return "", ""
        self.waits.append(timeout)
        remaining = self.shutdown_at - self.clock.now
        if self.stalled or remaining > timeout:
            self.clock.sleep(timeout)
            raise subprocess.TimeoutExpired("fake example", timeout)
        self.clock.sleep(max(0, remaining))
        self.returncode = 0
        return "", ""

    def kill(self):
        self.killed = True
        self.returncode = -9


class NativeBenchmarkTimingChecks(unittest.TestCase):
    def invoke(self, idle, stalled=False):
        clock = FakeClock()
        process = TimedExample(clock, idle, stalled)
        self.process, self.clock = process, clock
        readings = [{"rss_bytes": 1024, "cpu_seconds": 1}, {"rss_bytes": 2048, "cpu_seconds": 1.05}]
        with (
            mock.patch.object(benchmark.subprocess, "Popen", return_value=process) as launched,
            mock.patch.object(benchmark.selectors, "DefaultSelector", FakeSelector),
            mock.patch.object(benchmark.time, "monotonic", clock.monotonic),
            mock.patch.object(benchmark.time, "sleep", clock.sleep),
            mock.patch.object(benchmark.os, "read", side_effect=lambda fd, size: process.stdout.read(size)),
            mock.patch.object(benchmark, "sample", side_effect=readings),
            mock.patch.object(benchmark.platform, "platform", return_value="fake OS"),
            mock.patch.object(benchmark.platform, "machine", return_value="fake arch"),
        ):
            report = benchmark.run(Path("fake-vellum-hello"), idle, 10)
        self.assertEqual(launched.call_args.kwargs["bufsize"], 0)
        self.assertFalse(launched.call_args.kwargs.get("text", False))
        return report

    def test_sixty_second_idle_finishes_without_premature_kill(self):
        report = self.invoke(60)
        self.assertEqual(report["idle_sample_seconds"], 36)
        self.assertEqual(self.process.waits, [33.5])
        self.assertEqual(self.clock.now, 60)
        self.assertEqual(self.process.returncode, 0)
        self.assertFalse(self.process.killed)

    def test_default_idle_still_leaves_fixed_cleanup_margin(self):
        self.invoke(3)
        self.assertAlmostEqual(self.process.waits[0], 10.9)
        self.assertEqual(self.clock.now, 3)
        self.assertFalse(self.process.killed)

    def test_stalled_shutdown_is_killed_after_remaining_idle_and_margin(self):
        with self.assertRaises(subprocess.TimeoutExpired):
            self.invoke(60, stalled=True)
        self.assertEqual(self.process.waits, [33.5])
        self.assertEqual(self.clock.now, 70)
        self.assertTrue(self.process.killed)


class ReadinessPipeChecks(unittest.TestCase):
    def read_pipe(self, content, read_limit=None, timeout=.1):
        read_fd, write_fd = os.pipe()
        stdout = os.fdopen(read_fd, "rb", buffering=0)
        selector = benchmark.selectors.DefaultSelector()
        selector.register(stdout, benchmark.selectors.EVENT_READ)
        process = types.SimpleNamespace(stdout=stdout, poll=lambda: None)
        try:
            os.write(write_fd, content)
            if read_limit is None:
                return benchmark.read_ready(process, selector, time.monotonic() + timeout)
            raw_read = os.read
            with mock.patch.object(benchmark.os, "read", side_effect=lambda fd, size: raw_read(fd, min(size, read_limit))) as read:
                result = benchmark.read_ready(process, selector, time.monotonic() + timeout)
            self.assertGreater(read.call_count, 1)
            return result
        finally:
            selector.close()
            stdout.close()
            os.close(write_fd)

    def test_log_burst_and_ready_record_in_one_pipe_write(self):
        ready = {"event": "vellum.native.ready", "width": 640}
        result = self.read_pipe(b"initializing GPU\n\xff\n[]\nother log\n" + json.dumps(ready).encode() + b"\n")
        self.assertEqual(result, ready)

    def test_ready_record_split_across_raw_read_chunks(self):
        ready = {"event": "vellum.native.ready", "width": 640}
        result = self.read_pipe(b"initializing\n" + json.dumps(ready).encode() + b"\n", read_limit=7)
        self.assertEqual(result, ready)

    def test_partial_line_with_open_writer_respects_deadline(self):
        started = time.monotonic()
        self.assertIsNone(self.read_pipe(b"incomplete log without newline", timeout=.05))
        self.assertLess(time.monotonic() - started, .5)


class ProcessAccountingChecks(unittest.TestCase):
    def linux_stat(self):
        fields = ["S"] + ["0"] * 21
        fields[11], fields[12], fields[21] = "120", "35", "64"
        return "123 (example ) with spaces) " + " ".join(fields)

    def test_linux_ticks_and_rss_with_parentheses_in_process_name(self):
        result = benchmark.linux_stat_sample(self.linux_stat(), 100, 4096)
        self.assertEqual(result["cpu_seconds"], 1.55)
        self.assertEqual(result["rss_bytes"], 64 * 4096)
        self.assertEqual(result["cpu_time_quantum_seconds"], .01)

    def test_linux_sample_uses_tick_accounting_instead_of_coarse_ps(self):
        with mock.patch.object(benchmark.platform, "system", return_value="Linux"), mock.patch.object(Path, "read_text", return_value=self.linux_stat()) as read, mock.patch.object(benchmark.os, "sysconf", side_effect=[100, 4096]), mock.patch.object(benchmark.subprocess, "check_output") as ps:
            result = benchmark.sample(123)
        self.assertEqual(read.call_count, 1)
        ps.assert_not_called()
        self.assertEqual(result["cpu_seconds"], 1.55)
        self.assertEqual(result["cpu_sampling_source"], "proc-stat")

    def test_ps_preserves_subsecond_time_and_reports_observed_quantum(self):
        with mock.patch.object(benchmark.platform, "system", return_value="Darwin"), mock.patch.object(benchmark.subprocess, "check_output", return_value="256 01:02.34\n"):
            result = benchmark.sample(123)
        self.assertEqual(result["rss_bytes"], 256 * 1024)
        self.assertEqual(result["cpu_seconds"], 62.34)
        self.assertEqual(result["cpu_time_quantum_seconds"], .01)

    def test_truncated_linux_accounting_and_invalid_clock_are_rejected(self):
        with self.assertRaises(ValueError): benchmark.linux_stat_sample("1 (name) S 0", 100, 4096)
        with self.assertRaises(ValueError): benchmark.linux_stat_sample(self.linux_stat(), 0, 4096)


if __name__ == "__main__":
    unittest.main()
