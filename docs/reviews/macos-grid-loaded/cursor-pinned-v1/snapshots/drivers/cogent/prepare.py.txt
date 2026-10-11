"""Prepare driver-owned, pinned native modules without reading prototypes."""
from pathlib import Path
import hashlib
import json
import os
import shutil
import sys
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parent


def cleanup_staging_directory(staging):
    """Best-effort cleanup of owned partial copies, never recovery backups."""
    try:
        # copytree preserves immutable module-cache permissions. A copy/chmod
        # failure can leave descendants read-only; enable traversal/removal
        # before descending, and never follow symlinks outside owned staging.
        staging.chmod(staging.stat().st_mode | 0o700)
        for directory, directories, files in os.walk(staging):
            for name in directories:
                path = Path(directory) / name
                if not path.is_symlink():
                    path.chmod(path.stat().st_mode | 0o700)
            for name in files:
                path = Path(directory) / name
                if not path.is_symlink():
                    path.chmod(path.stat().st_mode | 0o600)
        shutil.rmtree(staging)
    except OSError as error:
        # A cleanup problem must not replace the original preparation error or
        # turn a successfully installed dependency into a reported failure.
        print(f"Warning: native staging cleanup failed; inspect retained path {staging}: {error}", file=sys.stderr)


def verify_inputs(owned, hashes):
    for relative, expected in hashes.items():
        path = owned / relative
        actual = hashlib.sha256(path.read_bytes()).hexdigest() if path.exists() else None
        if actual != expected:
            raise RuntimeError(f"Patch input mismatch: {relative}")


def prepare_dependency(dependency, destination):
    output = subprocess.check_output([
        "go", "mod", "download", "-json",
        f"{dependency['module']}@{dependency['version']}",
    ], cwd=ROOT)
    downloaded = json.loads(output)
    if downloaded.get("Error") or downloaded.get("Sum") != dependency["sum"]:
        raise RuntimeError(f"Module checksum mismatch: {dependency['module']}")
    shutil.copytree(downloaded["Dir"], destination)
    for directory, _, files in os.walk(destination):
        Path(directory).chmod(0o755)
        for filename in files:
            path = Path(directory) / filename
            path.chmod(path.stat().st_mode | 0o200)
    for stage in dependency["stages"]:
        verify_inputs(destination, json.loads((ROOT / stage["inputs"]).read_text()))
        for entry in stage["patches"]:
            patch = ROOT / entry["file"]
            if hashlib.sha256(patch.read_bytes()).hexdigest() != entry["sha256"]:
                raise RuntimeError(f"Patch checksum mismatch: {entry['file']}")
            subprocess.run(["patch", "--batch", "-p1", "-i", str(patch)],
                           cwd=destination, check=True)
        if "outputs" in stage:
            verify_inputs(destination, json.loads((ROOT / stage["outputs"]).read_text()))


def main():
    build = ROOT / "build"
    build.mkdir(exist_ok=True)
    lock = build / ".prepare-lock"
    try:
        lock.mkdir()
    except FileExistsError:
        raise RuntimeError("Native preparation already running; inspect build/.prepare-lock")
    staging = None
    cleanup_staging = True
    try:
        manifest = json.loads((ROOT / "native/dependencies.json").read_text())
        staging = Path(tempfile.mkdtemp(prefix="prepare-", dir=build))
        dependencies = manifest["dependencies"]
        for dependency in dependencies:
            prepare_dependency(dependency, staging / dependency["name"])
        # Replace only after every module and patch has passed validation.
        installed = []
        backups = []
        try:
            for dependency in dependencies:
                name = dependency["name"]
                current = build / name
                backup = staging / (name + ".previous")
                if current.exists():
                    current.rename(backup)
                    backups.append((current, backup))
                (staging / name).rename(current)
                installed.append(current)
        except BaseException as installation_error:
            # Once old installations move here, cleanup is safe only after every
            # backup has returned. A failed removal/restore must retain recovery
            # data even if another rollback step also fails.
            cleanup_staging = False
            rollback_errors = []
            for current in reversed(installed):
                try:
                    shutil.rmtree(current)
                except BaseException as error:
                    rollback_errors.append(f"remove {current}: {error}")
            for current, backup in reversed(backups):
                try:
                    if current.exists() or current.is_symlink():
                        raise RuntimeError("replacement still exists; original backup retained")
                    backup.rename(current)
                except BaseException as error:
                    rollback_errors.append(f"restore {backup} to {current}: {error}")
            # A signal may arrive immediately after rename succeeds but before
            # its journal entry is appended. Inspect all expected backup paths
            # before allowing staging cleanup, including unjournaled originals.
            retained = []
            for dependency in dependencies:
                backup = staging / (dependency["name"] + ".previous")
                if backup.exists() or backup.is_symlink():
                    retained.append(str(backup))
            if retained and not rollback_errors:
                rollback_errors.append("original backup remains after interrupted restoration")
            if rollback_errors:
                raise RuntimeError(
                    f"Native installation failed: {installation_error}; rollback incomplete. "
                    f"Recovery staging retained at {staging}; original backups: "
                    f"{', '.join(retained)}. {'; '.join(rollback_errors)}"
                ) from installation_error
            cleanup_staging = True
            raise
        print("Prepared driver-owned upstream and typesetting dependencies")
    finally:
        try:
            if staging is not None and cleanup_staging:
                cleanup_staging_directory(staging)
        finally:
            try:
                lock.rmdir()
            except OSError as error:
                print(f"Warning: native preparation lock release failed; inspect {lock} before manual removal: {error}", file=sys.stderr)


if __name__ == "__main__":
    main()
