"""Real POSIX descendant cleanup and deterministic Windows termination seam."""
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import tempfile
import time
import unittest
from unittest import mock
import child_process as child

class ChildChecks(unittest.TestCase):
    def test_success_capture_and_failed_status(self):
        self.assertEqual(child.check_child_output([sys.executable, "-c", "print('ok')"], timeout=5, text=True), "ok\n")
        with self.assertRaises(subprocess.CalledProcessError) as failure:
            child.run_child([sys.executable, "-c", "import sys; print('failure'); sys.exit(7)"], timeout=5, capture_output=True, text=True, check=True)
        self.assertEqual(failure.exception.returncode, 7)
        self.assertEqual(failure.exception.output, "failure\n")
        with self.assertRaisesRegex(ValueError, "argv without a shell"):
            child.run_child("echo unsafe", timeout=1, shell=True)

    @unittest.skipIf(os.name == "nt", "actual process-group acceptance requires a POSIX host")
    def test_timeout_kills_real_grandchild_holding_capture_pipes(self):
        with tempfile.TemporaryDirectory() as temporary:
            pidfile = Path(temporary) / "grandchild.pid"
            grandchild = "import time; time.sleep(30)"
            parent = "import subprocess,sys,time; from pathlib import Path; p=subprocess.Popen([sys.executable,'-c',sys.argv[2]]); Path(sys.argv[1]).write_text(str(p.pid)); time.sleep(30)"
            actual_popen = subprocess.Popen
            ready_at = []
            def start_ready(*arguments, **options):
                process = actual_popen(*arguments, **options)
                startup_deadline = time.monotonic() + 15
                try:
                    while True:
                        if pidfile.is_file() and pidfile.read_text().isdigit():
                            ready_at.append(time.monotonic())
                            return process
                        if process.poll() is not None:
                            raise AssertionError("fixture parent exited before descendant readiness")
                        if time.monotonic() >= startup_deadline:
                            raise AssertionError("fixture descendant startup exceeded its separate deadline")
                        time.sleep(0.02)
                except BaseException:
                    child.terminate_owned_group(process)
                    process.communicate(timeout=child.CLEANUP_TIMEOUT_SECONDS)
                    raise
            # Only the fixture's startup is synchronized. Popen, communicate,
            # timeout and process-group termination still use real native APIs.
            with mock.patch.object(child.subprocess, "Popen", side_effect=start_ready):
                with self.assertRaises(subprocess.TimeoutExpired):
                    child.run_child([sys.executable, "-c", parent, str(pidfile), grandchild], timeout=0.5, capture_output=True, text=True)
            self.assertEqual(len(ready_at), 1)
            self.assertLess(time.monotonic() - ready_at[0], child.CLEANUP_TIMEOUT_SECONDS + 10)
            pid = int(pidfile.read_text())
            # A zombie is terminated and awaiting reaping, not a live descendant.
            state = subprocess.run(["ps", "-p", str(pid), "-o", "stat="], capture_output=True, text=True, timeout=2).stdout.strip()
            self.assertTrue(not state or state.startswith("Z"), state)

    def test_windows_tree_termination_is_fixed_argv_and_failure_is_not_success(self):
        process = mock.Mock(pid=1234)
        with mock.patch.object(child.os, "name", "nt"), mock.patch.object(child.subprocess, "run", return_value=subprocess.CompletedProcess([], 0)) as execute:
            child.terminate_owned_group(process)
            self.assertEqual(execute.call_args.args[0], ["taskkill", "/PID", "1234", "/T", "/F"])
            self.assertEqual(execute.call_args.kwargs["timeout"], child.CLEANUP_TIMEOUT_SECONDS)
            self.assertNotIn("shell", execute.call_args.kwargs)
        with mock.patch.object(child.os, "name", "nt"), mock.patch.object(child.subprocess, "run", return_value=subprocess.CompletedProcess([], 1)):
            with self.assertRaises(child.ChildCleanupError): child.terminate_owned_group(process)

    def test_windows_unknown_tree_cleanup_preserves_original_timeout_cause(self):
        original = subprocess.TimeoutExpired(["owned-child"], 1)
        for cleanup_result in (("output", "diagnostic"), subprocess.TimeoutExpired(["owned-child"], child.CLEANUP_TIMEOUT_SECONDS)):
            with self.subTest(pipes_remain_open=isinstance(cleanup_result, subprocess.TimeoutExpired)):
                process = mock.Mock(pid=1234, stdin=None, stdout=None, stderr=None)
                process.poll.return_value = 0  # Parent exited; descendants remain unknown.
                process.communicate.side_effect = [original, cleanup_result]
                with mock.patch.object(child.os, "name", "nt"), mock.patch.object(child.subprocess, "CREATE_NEW_PROCESS_GROUP", 512, create=True), mock.patch.object(child.subprocess, "Popen", return_value=process), mock.patch.object(child.subprocess, "run", return_value=subprocess.CompletedProcess([], 128)):
                    with self.assertRaises(child.ChildCleanupError) as failure:
                        child.run_child(["owned-child"], timeout=1, capture_output=True, text=True)
                self.assertIs(failure.exception.__cause__, original)
                self.assertIn("cleanup", str(failure.exception))
                process.kill.assert_not_called()
                self.assertEqual(process.communicate.call_args.kwargs["timeout"], child.CLEANUP_TIMEOUT_SECONDS)

if __name__ == "__main__": unittest.main()
