"""Bound SDK commands and clean their owned groups on timeout/interruption.

Normal completion retains ordinary subprocess semantics; this is not a sandbox.
"""
import os
import signal
import subprocess

CLEANUP_TIMEOUT_SECONDS = 5

class ChildCleanupError(OSError):
    """The command failed and descendant cleanup could not be confirmed."""


def terminate_owned_group(process):
    if os.name == "nt":
        # Fixed argv only; the PID belongs to the Popen instance we created.
        completed = subprocess.run(["taskkill", "/PID", str(process.pid), "/T", "/F"],
                                   stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL,
                                   timeout=CLEANUP_TIMEOUT_SECONDS, check=False)
        if completed.returncode != 0:
            raise ChildCleanupError("Windows taskkill could not confirm owned process-tree termination")
    else:
        try:
            os.killpg(process.pid, signal.SIGKILL)
        except ProcessLookupError:
            pass  # The entire owned group has already exited.


def run_child(arguments, *, timeout, check=False, capture_output=False, input=None, **options):
    if timeout <= 0 or options.get("shell"):
        raise ValueError("SDK children require a positive deadline and argv without a shell")
    if capture_output:
        if "stdout" in options or "stderr" in options:
            raise ValueError("capture_output conflicts with explicit stdout/stderr")
        options["stdout"] = options["stderr"] = subprocess.PIPE
    if input is not None:
        if "stdin" in options:
            raise ValueError("input conflicts with explicit stdin")
        options["stdin"] = subprocess.PIPE
    if os.name == "nt":
        options["creationflags"] = options.get("creationflags", 0) | subprocess.CREATE_NEW_PROCESS_GROUP
    else:
        options["start_new_session"] = True
    process = subprocess.Popen(arguments, **options)
    try:
        try:
            stdout, stderr = process.communicate(input=input, timeout=timeout)
        except BaseException as original:
            cleanup_error = None
            try:
                terminate_owned_group(process)
            except Exception as error:
                cleanup_error = error
                if process.poll() is None:
                    process.kill()
            try:
                stdout, stderr = process.communicate(timeout=CLEANUP_TIMEOUT_SECONDS)
            except subprocess.TimeoutExpired as error:
                raise ChildCleanupError("SDK child pipes remained open after cleanup attempt" + (": " + str(cleanup_error) if cleanup_error is not None else "")) from original
            if cleanup_error is not None:
                raise ChildCleanupError("SDK descendant cleanup failed: " + str(cleanup_error)) from original
            if isinstance(original, subprocess.TimeoutExpired):
                original.output, original.stderr = stdout, stderr
            raise
        completed = subprocess.CompletedProcess(arguments, process.returncode, stdout, stderr)
        if check and completed.returncode:
            raise subprocess.CalledProcessError(completed.returncode, arguments, output=stdout, stderr=stderr)
        return completed
    finally:
        for stream in (process.stdin, process.stdout, process.stderr):
            if stream is not None:
                stream.close()
        if process.poll() is None:
            process.kill()
            try:
                process.wait(timeout=CLEANUP_TIMEOUT_SECONDS)
            except subprocess.TimeoutExpired as error:
                raise ChildCleanupError("SDK direct child remained alive after termination") from error


def check_child_output(arguments, *, timeout, **options):
    if "stdout" in options:
        raise ValueError("check_child_output owns stdout capture")
    return run_child(arguments, timeout=timeout, check=True, stdout=subprocess.PIPE, **options).stdout
