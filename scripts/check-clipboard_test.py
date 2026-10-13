"""Clipboard harness failures without native calls or user clipboard access."""
import contextlib
import importlib.util
import io
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest import mock

spec = importlib.util.spec_from_file_location("check_clipboard", Path(__file__).with_name("check-clipboard.py"))
clipboard = importlib.util.module_from_spec(spec)
spec.loader.exec_module(clipboard)


class ClipboardHarnessChecks(unittest.TestCase):
    def test_interrupt_records_receipt_and_stops_before_later_build_or_clipboard_run(self):
        for error in (KeyboardInterrupt(), SystemExit(7)):
            with self.subTest(error=type(error).__name__), tempfile.TemporaryDirectory() as temporary:
                output = Path(temporary) / "output"
                args = ["check-clipboard.py", "--self-test", "--output", str(output)]
                with mock.patch.object(sys, "argv", args), mock.patch.object(sys, "platform", "darwin"), mock.patch.object(clipboard, "child_environment", return_value={}), mock.patch.object(clipboard, "run_child", side_effect=error) as child:
                    with self.assertRaises(type(error)):
                        clipboard.main()
                self.assertEqual(child.call_count, 1)
                receipt = json.loads((output / "build-guardian-native.json").read_text())
                self.assertEqual(receipt["error"], {"type": type(error).__name__})
                self.assertFalse((output / "guardian-run.json").exists())

    def test_ordinary_process_error_retains_bounded_receipt(self):
        with tempfile.TemporaryDirectory() as temporary:
            with mock.patch.object(clipboard, "run_child", side_effect=OSError("dummy private detail")):
                receipt = clipboard.process(["dummy"], Path(temporary), "process", {}, 1)
            self.assertEqual(receipt["error"], {"type": "OSError"})
            self.assertNotIn("dummy private detail", json.dumps(receipt))

    def prebuilt(self, temporary):
        root = Path(temporary)
        guardian, binary = root / "guardian", root / "fixture"
        guardian.write_bytes(b"dummy guardian")
        binary.write_bytes(b"dummy fixture")
        output = root / "output"
        args = ["check-clipboard.py", "--live", "--output", str(output), "--guardian", str(guardian), "--binary", str(binary)]
        return output, args

    def test_partial_or_nonobject_summary_produces_failure_and_source_receipt(self):
        for summary in ("{\"result\":", "[]", "null"):
            with self.subTest(summary=summary), tempfile.TemporaryDirectory() as temporary:
                output, args = self.prebuilt(temporary)
                def child(arguments, **kwargs):
                    job = Path(arguments[2])
                    (job / "summary.json").write_text(summary)
                    return subprocess.CompletedProcess(arguments, 0)
                with mock.patch.object(sys, "argv", args), mock.patch.object(sys, "platform", "darwin"), mock.patch.object(clipboard, "child_environment", return_value={}), mock.patch.object(clipboard, "run_child", side_effect=child), contextlib.redirect_stdout(io.StringIO()):
                    self.assertEqual(clipboard.main(), 1)
                result = json.loads((output / "result.json").read_text())
                self.assertEqual(result["result"], "FAILED_OR_REFUSED")
                self.assertTrue(result["source_stable"])
                self.assertFalse(result["protected_backup_retained"])
                if summary != "null":
                    self.assertEqual(result["summary"], {"result": "UNREADABLE_SUMMARY"})

    def test_interrupted_guardian_attempts_recovery_then_propagates_cancellation(self):
        with tempfile.TemporaryDirectory() as temporary:
            output, args = self.prebuilt(temporary)
            calls = []
            def child(arguments, **kwargs):
                calls.append(arguments[1])
                job = Path(arguments[2])
                backup = job / "original.plist"
                if arguments[1] == "--run":
                    backup.write_bytes(b"dummy backup only")
                    raise KeyboardInterrupt()
                self.assertEqual(arguments[1], "--recover")
                backup.unlink()
                return subprocess.CompletedProcess(arguments, 0)
            with mock.patch.object(sys, "argv", args), mock.patch.object(sys, "platform", "darwin"), mock.patch.object(clipboard, "child_environment", return_value={}), mock.patch.object(clipboard, "run_child", side_effect=child):
                with self.assertRaises(KeyboardInterrupt):
                    clipboard.main()
            self.assertEqual(calls, ["--run", "--recover"])
            receipt = json.loads((output / "guardian-run.json").read_text())
            self.assertEqual(receipt["error"], {"type": "KeyboardInterrupt"})
            recovery = json.loads((output / "guardian-recovery.json").read_text())
            self.assertEqual(recovery["exit_code"], 0)
            self.assertFalse((output / "private/job/original.plist").exists())


if __name__ == "__main__":
    unittest.main()
