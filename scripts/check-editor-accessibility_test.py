"""External AX failure evidence and owned-fixture teardown without OS access."""
import importlib.util
import json
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest import mock

spec = importlib.util.spec_from_file_location("editor_ax", Path(__file__).with_name("check-editor-accessibility.py"))
ax = importlib.util.module_from_spec(spec)
spec.loader.exec_module(ax)


class ExternalAXEvidence(unittest.TestCase):
    def run_fixture(self, inspector_output='{"result":"pass"}\n', inspector_exit=0,
                    inspector_timeout=False, native_timeout=False):
        temporary = tempfile.TemporaryDirectory(prefix="editor AX failure test ")
        self.addCleanup(temporary.cleanup)
        base = Path(temporary.name)
        binary = base / "editor"
        binary.write_bytes(b"owned fake editor")
        output = base / "evidence"
        clock = [0.0]
        processes = []

        class Fixture:
            pid = 4242
            returncode = None
            waits = 0

            def __init__(self, arguments, **options):
                self.arguments = arguments
                self.options = options
                processes.append(self)
                ready = {"pid": self.pid, "lifetime_seconds": 60,
                         "text": "A😀e\u0301\r\n\r\n"}
                options["stdout"].write("EDITOR_AX_FIXTURE " + json.dumps(ready) + "\n")
                options["stdout"].flush()

            def poll(self):
                return self.returncode

            def wait(self, timeout):
                self.waits += 1
                if native_timeout and self.waits == 1:
                    clock[0] = 75
                    raise subprocess.TimeoutExpired(self.arguments, timeout)
                clock[0] = 76 if native_timeout else 61
                if self.returncode is None:
                    self.returncode = 0
                return self.returncode

        def child(arguments, **options):
            if arguments[0] == "xcrun":
                Path(arguments[-1]).write_bytes(b"fake inspector")
                return subprocess.CompletedProcess(arguments, 0)
            if arguments[-1] == "--trust-only":
                return subprocess.CompletedProcess(arguments, 0, '{"accessibility_trusted":true}\n', "")
            clock[0] = 35 if inspector_timeout else 1
            if inspector_timeout:
                raise subprocess.TimeoutExpired(arguments, 35, output=b"partial inspector stdout\n", stderr=b"partial inspector stderr\n")
            return subprocess.CompletedProcess(arguments, inspector_exit, inspector_output, "inspector stderr\n")

        def terminate(process):
            self.assertIs(process, processes[0])
            process.returncode = -9

        with mock.patch.object(ax.sys, "platform", "darwin"), \
             mock.patch.object(ax, "run_child", side_effect=child), \
             mock.patch.object(ax.subprocess, "Popen", Fixture), \
             mock.patch.object(ax.time, "monotonic", side_effect=lambda: clock[0]), \
             mock.patch.object(ax, "terminate_owned_group", side_effect=terminate) as cleanup:
            if inspector_exit or inspector_timeout or native_timeout or inspector_output != '{"result":"pass"}\n':
                with self.assertRaisesRegex(ValueError, "External AX/native fixture failed"):
                    ax.check(binary, output)
            else:
                ax.check(binary, output)
            self.assertEqual(cleanup.call_count, int(native_timeout))
        result = json.loads((output / "acceptance.json").read_text())
        self.assertFalse(result["settings_changed"])
        self.assertEqual(processes[0].options["start_new_session"], True)
        self.assertEqual((output / "expected.txt").read_bytes(), "A😀e\u0301\r\n\r\n".encode())
        return result, output, processes[0]

    def test_success_waits_for_normal_native_timer(self):
        result, _, process = self.run_fixture()
        self.assertEqual(result["result"], "pass")
        self.assertEqual(result["native_exit_code"], 0)
        self.assertEqual(result["native_elapsed_seconds"], 61)
        self.assertTrue(result["timer_lifetime_reached"])
        self.assertEqual(process.waits, 1)

    def test_missing_json_records_failure_and_normal_native_teardown(self):
        result, output, process = self.run_fixture(inspector_output="inspector crashed\n")
        self.assertEqual(result["result"], "failure")
        self.assertEqual(result["native_exit_code"], 0)
        self.assertEqual(process.waits, 1)
        self.assertIn("inspector crashed", (output / "inspector.log").read_text())

    def test_malformed_json_keeps_original_diagnostics(self):
        result, output, _ = self.run_fixture(inspector_output="{malformed result\n")
        self.assertEqual(result["result"], "failure")
        self.assertIsNotNone(result["inspector_error"])
        self.assertIn("{malformed result", (output / "inspector.log").read_text())

    def test_multiple_results_fail_closed(self):
        result, _, _ = self.run_fixture(inspector_output='{"result":"pass"}\n{"result":"pass"}\n')
        self.assertEqual(result["result"], "failure")
        self.assertIn("one inspector JSON", result["inspector_error"])

    def test_nonzero_inspector_cannot_claim_pass(self):
        result, _, _ = self.run_fixture(inspector_exit=2)
        self.assertEqual(result["result"], "failure")
        self.assertEqual(result["inspector_exit_code"], 2)
        self.assertEqual(result["native_exit_code"], 0)

    def test_inspector_timeout_keeps_partial_output_and_normal_native_close(self):
        result, output, process = self.run_fixture(inspector_timeout=True)
        self.assertEqual(result["result"], "failure")
        self.assertTrue(result["inspector_timed_out"])
        self.assertIsNone(result["inspector_exit_code"])
        self.assertEqual(result["native_exit_code"], 0)
        self.assertEqual(process.waits, 1)
        self.assertEqual((output / "inspector.log").read_text(), "partial inspector stdout\npartial inspector stderr\n")

    def test_native_timeout_records_owned_termination(self):
        result, _, process = self.run_fixture(native_timeout=True)
        self.assertEqual(result["result"], "failure")
        self.assertTrue(result["native_timed_out"])
        self.assertEqual(result["native_exit_code"], -9)
        self.assertEqual(process.waits, 2)


if __name__ == "__main__":
    unittest.main()
