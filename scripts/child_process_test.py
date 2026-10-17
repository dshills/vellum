"""Real POSIX descendant cleanup and deterministic Windows termination seam."""
import json
import io
import hashlib
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
    def assert_retired(self, pid):
        observation = subprocess.run(["ps", "-p", str(pid), "-o", "stat="], capture_output=True, text=True, timeout=2)
        self.assertEqual(observation.stderr, "", "process observation returned diagnostics")
        state = observation.stdout.strip()
        if observation.returncode == 1:
            self.assertEqual(state, "", "absent-PID status included process data")
        else:
            self.assertEqual(observation.returncode, 0, "process observation failed")
            self.assertTrue(state.startswith("Z") and len(state.splitlines()) == 1, state)

    def test_process_observation_refuses_errors_and_live_descendants(self):
        for status, state, diagnostic in [(2, "", ""), (1, "", "permission denied"), (1, "S", ""), (0, "", ""), (0, "S", "")]:
            with self.subTest(status=status, state=state, diagnostic=diagnostic), mock.patch.object(subprocess, "run", return_value=subprocess.CompletedProcess([], status, state, diagnostic)):
                with self.assertRaises(AssertionError): self.assert_retired(123)

    @unittest.skipIf(os.name == "nt", "bounded pipe logs require POSIX selectors")
    def test_bounded_binary_logs_preserve_both_channels(self):
        stdout, stderr = io.BytesIO(), io.BytesIO()
        command = [sys.executable, "-c", "import sys;sys.stdout.buffer.write(b'a'*65537);sys.stderr.buffer.write(b'z'*4097)"]
        result = child.run_child(command, timeout=5, stdout=stdout, stderr=stderr, log_limit_bytes=65537)
        self.assertEqual(result.returncode, 0)
        self.assertIsNone(result.stdout)
        self.assertEqual(stdout.getvalue(), b'a'*65537)
        self.assertEqual(stderr.getvalue(), b'z'*4097)
        self.assertEqual(result.log_summaries['stdout'],{'bytes':65537,'sha256':hashlib.sha256(b'a'*65537).hexdigest()})

    @unittest.skipIf(os.name == "nt", "bounded pipe logs require POSIX selectors")
    def test_log_limit_kills_owned_group_and_keeps_only_prefix(self):
        with tempfile.TemporaryDirectory() as temporary:
            pidfile = Path(temporary)/"child.pid"
            command = "import subprocess,sys,time;from pathlib import Path;p=subprocess.Popen([sys.executable,'-c','import time;time.sleep(30)']);Path(sys.argv[1]).write_text(str(p.pid));sys.stdout.buffer.write(b'x'*65536);sys.stdout.flush();time.sleep(30)"
            stdout, stderr = io.BytesIO(), io.BytesIO()
            with self.assertRaises(child.ChildOutputLimitError):
                child.run_child([sys.executable,"-c",command,str(pidfile)], timeout=5, stdout=stdout, stderr=stderr, log_limit_bytes=23)
            self.assertEqual(stdout.getvalue(),b'x'*23)
            pid = int(pidfile.read_text())
            self.assert_retired(pid)

    @unittest.skipIf(os.name == "nt", "bounded pipe logs require POSIX selectors")
    def test_stderr_limit_timeout_and_sink_failure(self):
        for command,limit,error in [("import sys;sys.stderr.buffer.write(b'e'*38)",37,child.ChildOutputLimitError),("import time;time.sleep(30)",100,subprocess.TimeoutExpired)]:
            stdout,stderr=io.BytesIO(),io.BytesIO()
            with self.assertRaises(error):
                child.run_child([sys.executable,"-c",command],timeout=5 if error is child.ChildOutputLimitError else .2,stdout=stdout,stderr=stderr,log_limit_bytes=limit)
            self.assertLessEqual(len(stderr.getvalue()),limit)
        sink=mock.Mock();sink.write.side_effect=OSError("broken log sink")
        with self.assertRaisesRegex(OSError,"broken log sink"):
            child.run_child([sys.executable,"-c","print('out')"],timeout=5,stdout=sink,stderr=io.BytesIO(),log_limit_bytes=100)

    def test_invalid_bounded_log_options_refuse_before_start(self):
        for options in [{"log_limit_bytes":0},{"log_limit_bytes":True},{"log_limit_bytes":1,"text":True},{"log_limit_bytes":1,"capture_output":True},{"log_limit_bytes":1,"input":b'x'},{"log_limit_bytes":1,"stderr":None}]:
            with mock.patch.object(child.subprocess,"Popen") as launch:
                with self.assertRaises(ValueError):
                    child.run_child(["unused"],timeout=1,**{"stdout":io.BytesIO(),"stderr":io.BytesIO(),**options})
                launch.assert_not_called()

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
            self.assert_retired(pid)

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

SIGNAL_GRANDCHILD = r"""
import os, signal, sys, time
from pathlib import Path
signal.signal(signal.SIGINT, signal.SIG_IGN)
signal.signal(signal.SIGTERM, signal.SIG_IGN)
Path(sys.argv[1]).write_text(str(os.getpid()))
time.sleep(45)
"""

SIGNAL_CHILD = r"""
import json, os, signal, subprocess, sys, time
from pathlib import Path
root = Path(sys.argv[1])
signal.signal(signal.SIGINT, signal.SIG_IGN)
signal.signal(signal.SIGTERM, signal.SIG_IGN)
p = subprocess.Popen([sys.executable, '-c', sys.argv[2], str(root/'grandchild.ready')])
deadline = time.monotonic()+5
while not (root/'grandchild.ready').exists():
    if time.monotonic()>deadline: raise TimeoutError('grandchild startup')
    time.sleep(.01)
(root/'owned.json').write_text(json.dumps(dict(child=os.getpid(), group=os.getpgrp(), grandchild=p.pid)))
time.sleep(45)
"""

SIGNAL_WORKER = r"""
import io, json, os, signal, subprocess, sys, time
from pathlib import Path
sys.path.insert(0, sys.argv[1])
import child_process as child
root, stage, mode = Path(sys.argv[2]), sys.argv[3], sys.argv[4]
signum = int(sys.argv[5])
before = {n:signal.getsignal(n) for n in (signal.SIGINT, signal.SIGTERM)}
actual_popen, actual_retire = subprocess.Popen, child.terminate_owned_group
actual_stream_logs = child.stream_logs
def wait_file(path):
    deadline=time.monotonic()+5
    while not path.exists():
        if time.monotonic()>deadline: raise TimeoutError('fixture boundary')
        time.sleep(.01)
def starting(*arguments, **options):
    process=actual_popen(*arguments, **options)
    wait_file(root/'owned.json')
    if stage=='creation':
        (root/'boundary').touch()
        # Actual signal before the Popen wrapper has returned its owned handle.
        os.kill(os.getpid(), signum)
        (root/'creation-returned').touch()
    elif stage=='running':
        actual_communicate=process.communicate
        def communicate(*a,**kw):
            (root/'boundary').touch()
            return actual_communicate(*a,**kw)
        process.communicate=communicate
    return process
def retiring(process):
    if stage=='cleanup':
        (root/'boundary').touch()
        wait_file(root/'release')
    return actual_retire(process)
def streaming(*a, **kw):
    if stage=='running':(root/'boundary').touch()
    return actual_stream_logs(*a, **kw)
child.subprocess.Popen=starting
child.terminate_owned_group=retiring
child.stream_logs=streaming
try:
    args=[sys.executable,'-c',sys.argv[6],str(root),sys.argv[7]]
    options=dict(capture_output=True, text=True) if mode=='capture' else dict(stdout=io.BytesIO(), stderr=io.BytesIO(),log_limit_bytes=1024)
    child.run_child(args, timeout=.15 if stage=='cleanup' else 10, **options)
except BaseException as error:
    result=dict(exception=type(error).__name__,handlers_restored=all(signal.getsignal(n)==h for n,h in before.items()))
    if isinstance(error,SystemExit): code=error.code
    elif isinstance(error,KeyboardInterrupt):code=130
    else:code=1
    result['exit_code']=code
    (root/'outcome.json').write_text(json.dumps(result))
    raise SystemExit(code)
raise SystemExit('fixture unexpectedly completed')
"""


@unittest.skipIf(os.name == "nt", "real SIGINT/SIGTERM acceptance is POSIX-only")
class ChildSignalChecks(unittest.TestCase):
    assert_retired = ChildChecks.assert_retired
    def run_signal_case(self, stage, signum, mode="capture"):
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory)
            process=subprocess.Popen([sys.executable,'-c',SIGNAL_WORKER,str(Path(child.__file__).parent),str(root),stage,mode,str(signum),SIGNAL_CHILD,SIGNAL_GRANDCHILD],stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True,start_new_session=True)
            owned=None
            try:
                deadline=time.monotonic()+10
                while not (root/'boundary').exists():
                    self.assertIsNone(process.poll(),'fixture exited before boundary')
                    self.assertLess(time.monotonic(),deadline,'fixture startup deadline')
                    time.sleep(.01)
                owned=json.loads((root/'owned.json').read_text())
                self.assertEqual(owned['group'],owned['child'],'fixture must own a separate session')
                if stage!='creation':
                    os.kill(process.pid,signum)
                    if stage=='cleanup':
                        # Both default cancellation kinds arrive while group
                        # retirement is deliberately held, before release.
                        time.sleep(.03)
                        os.kill(process.pid,signal.SIGINT if signum==signal.SIGTERM else signal.SIGTERM)
                        time.sleep(.03)
                        (root/'release').touch()
                stdout,stderr=process.communicate(timeout=10)
                for pid in (owned['child'],owned['grandchild']):
                    deadline=time.monotonic()+3
                    while True:
                        try:self.assert_retired(pid);break
                        except AssertionError:
                            if time.monotonic()>deadline:raise
                            time.sleep(.02)
                expected=1 if stage=='cleanup' else 130 if signum==signal.SIGINT else 143
                self.assertEqual(process.returncode,expected,(stage,mode,stdout,stderr))
                outcome=json.loads((root/'outcome.json').read_text())
                self.assertTrue(outcome['handlers_restored'])
                self.assertEqual(outcome['exception'],'TimeoutExpired' if stage=='cleanup' else 'KeyboardInterrupt' if signum==signal.SIGINT else 'ChildCancelled')
                if stage=='creation':self.assertTrue((root/'creation-returned').exists(),'Popen returned before cancellation unwound')
            finally:
                # Also retire real fixtures on a failing old-source control.
                if owned is None and (root/'owned.json').exists():owned=json.loads((root/'owned.json').read_text())
                if owned is not None:
                    self.assertEqual(owned['group'],owned['child'])
                    try:os.killpg(owned['group'],signal.SIGKILL)
                    except ProcessLookupError:pass
                try:os.killpg(process.pid,signal.SIGKILL)
                except ProcessLookupError:pass
                process.communicate(timeout=5)

    def test_cancel_during_creation(self):
        for signum in (signal.SIGTERM,signal.SIGINT):
            for mode in ("capture","logs"):
                with self.subTest(signal=signum,mode=mode):self.run_signal_case('creation',signum,mode)

    def test_cancel_during_communication(self):
        for signum in (signal.SIGTERM,signal.SIGINT):
            for mode in ("capture","logs"):
                with self.subTest(signal=signum,mode=mode):self.run_signal_case('running',signum,mode)

    def test_repeated_signals_do_not_interrupt_cleanup(self):
        for signum in (signal.SIGTERM,signal.SIGINT):
            for mode in ("capture","logs"):
                with self.subTest(signal=signum,mode=mode):self.run_signal_case('cleanup',signum,mode)

    def test_signal_policy_restores_after_success_and_spawn_error(self):
        handlers={n:signal.getsignal(n) for n in (signal.SIGINT,signal.SIGTERM)}
        child.run_child([sys.executable,'-c','pass'],timeout=5)
        self.assertEqual({n:signal.getsignal(n) for n in handlers},handlers)
        with self.assertRaises(FileNotFoundError):child.run_child(['/nonexistent/vellum-command'],timeout=5)
        self.assertEqual({n:signal.getsignal(n) for n in handlers},handlers)

    def test_custom_and_ignored_signal_handlers_remain_caller_owned(self):
        previous=signal.getsignal(signal.SIGTERM)
        for handler in (lambda signum,frame:None,signal.SIG_IGN):
            try:
                signal.signal(signal.SIGTERM,handler)
                def check_policy(*a,**kw):
                    self.assertEqual(signal.getsignal(signal.SIGTERM),handler)
                    raise FileNotFoundError('controlled spawn refusal')
                with mock.patch.object(child.subprocess,'Popen',side_effect=check_policy),self.assertRaises(FileNotFoundError):child.run_child(['unused'],timeout=5)
                self.assertEqual(signal.getsignal(signal.SIGTERM),handler)
            finally:signal.signal(signal.SIGTERM,previous)

    def test_sigint_during_restoration_does_not_strand_sigterm_handler(self):
        before={n:signal.getsignal(n) for n in (signal.SIGINT,signal.SIGTERM)}
        actual_signal=signal.signal
        injected=[]
        def restore_then_interrupt(signum,handler):
            previous=actual_signal(signum,handler)
            if signum==signal.SIGINT and handler==before[signal.SIGINT] and not injected:
                injected.append(True)
                os.kill(os.getpid(),signal.SIGINT)
            return previous
        try:
            with mock.patch.object(child.signal,'signal',side_effect=restore_then_interrupt):
                with self.assertRaises(KeyboardInterrupt):child.run_child([sys.executable,'-c','pass'],timeout=5)
            self.assertEqual(injected,[True])
            self.assertEqual({n:signal.getsignal(n) for n in before},before)
        finally:
            for signum,handler in before.items():actual_signal(signum,handler)

    def test_worker_thread_does_not_install_process_signal_handlers(self):
        import threading
        outcome=[]
        def run():
            try:outcome.append(child.run_child([sys.executable,'-c',"print('worker')"],timeout=5,capture_output=True,text=True).stdout)
            except BaseException as error:outcome.append(error)
        worker=threading.Thread(target=run);worker.start();worker.join(timeout=10)
        self.assertFalse(worker.is_alive());self.assertEqual(outcome,['worker\n'])


if __name__ == "__main__": unittest.main()
