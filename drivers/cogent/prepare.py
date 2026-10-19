"""Prepare driver-owned, pinned native modules without reading prototypes."""
from pathlib import Path, PurePosixPath
from contextlib import contextmanager
import re
import stat
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


# Text/binary asset overlays are admitted before a stage mutates its owned copy.
MAX_STAGE_ASSETS = 256
MAX_ASSET_BYTES = 64 << 10
MAX_PINNED_ASSET_BYTES = 512 << 10
MAX_STAGE_ASSET_BYTES = 4 << 20
MAX_PATCH_BYTES = 8 << 20
MAX_STAGE_PATCH_BYTES = 32 << 20


def _relative_asset_name(name):
    if (type(name) is not str or not name or len(name) > 4096 or "\\" in name
            or any(ord(c) < 32 or ord(c) == 127 for c in name)
            or len(name.split("/")) > 32
            or any(part in {"", ".", ".."} for part in name.split("/"))
            or PurePosixPath(name).is_absolute()):
        raise RuntimeError("Invalid stage asset path")
    return name


def _asset_path(base, name):
    current = base.resolve(strict=True)
    parts = _relative_asset_name(name).split("/")
    for index, part in enumerate(parts):
        current = current / part
        mode = current.lstat().st_mode
        if stat.S_ISLNK(mode) or not (stat.S_ISREG(mode) if index == len(parts) - 1 else stat.S_ISDIR(mode)):
            raise RuntimeError("Stage asset must be a regular file without symlinks: " + name)
    return current


def _asset_identity(value):
    return (value.st_dev, value.st_ino, value.st_size, value.st_mtime_ns, value.st_ctime_ns)


def _read_stage_file(path, bound):
    flags = os.O_RDONLY | getattr(os, "O_NOFOLLOW", 0) | getattr(os, "O_NONBLOCK", 0)
    descriptor = os.open(path, flags)
    try:
        # This owner closes even when constructing the stream itself fails.
        with os.fdopen(descriptor, "rb", closefd=False) as stream:
            before = os.fstat(stream.fileno())
            if not stat.S_ISREG(before.st_mode) or before.st_size > bound:
                raise RuntimeError("Stage file exceeds byte bound or is nonregular")
            remaining, body = before.st_size, bytearray()
            while remaining:
                chunk = stream.read(min(65536, remaining))
                if not chunk:
                    raise RuntimeError("Stage file shortened during admission")
                body.extend(chunk)
                remaining -= len(chunk)
            if stream.read(1):
                raise RuntimeError("Stage file grew beyond admitted size")
            if (_asset_identity(os.fstat(stream.fileno())) != _asset_identity(before)
                    or _asset_identity(os.stat(path, follow_symlinks=False)) != _asset_identity(before)):
                raise RuntimeError("Stage file changed during admission")
    finally:
        os.close(descriptor)
    return bytes(body), before


def _sha256_value(value):
    return type(value) is str and re.fullmatch(r"[0-9a-f]{64}", value) is not None


def _stage_patch_snapshots(stage):
    entries = stage["patches"]
    if type(entries) is not list or len(entries) > MAX_STAGE_ASSETS:
        raise RuntimeError("Stage patch count bound exceeded")
    patches, total = [], 0
    for entry in entries:
        if type(entry) is not dict or set(entry) != {"file", "sha256"} or not _sha256_value(entry["sha256"]):
            raise RuntimeError("Invalid stage patch record")
        path = _asset_path(ROOT, entry["file"])
        body, _ = _read_stage_file(path, min(MAX_PATCH_BYTES, MAX_STAGE_PATCH_BYTES - total))
        total += len(body)
        if hashlib.sha256(body).hexdigest() != entry["sha256"]:
            raise RuntimeError("Patch checksum mismatch: " + entry["file"])
        patches.append(body)
    return patches


def _asset_stage_patch_targets(body):
    """Asset stages accept canonical unified diffs only, including tab timestamps.

    Parsing complete hunks prevents header-like payload lines or another diff
    format from hiding a patch/overlay collision from the preflight check.
    Legacy stages without asset overlays retain their existing patch format.
    """
    lines, offset, targets = body.splitlines(), 0, set()

    def header(line, prefix):
        if not line.startswith(prefix):
            raise RuntimeError("Asset stage requires canonical unified patch headers")
        name = line[len(prefix):].split(b"\t", 1)[0]
        if name == b"/dev/null":
            return None
        side = b"a/" if prefix == b"--- " else b"b/"
        if not name.startswith(side) or any(c in b" \t\r\n" for c in name):
            raise RuntimeError("Asset stage requires canonical unified patch paths")
        try:
            return _relative_asset_name(name[2:].decode("utf-8")).casefold()
        except UnicodeDecodeError as error:
            raise RuntimeError("Invalid asset stage patch path encoding") from error

    if not lines:
        raise RuntimeError("Asset stage patch is empty")
    while offset < len(lines):
        old = header(lines[offset], b"--- ")
        offset += 1
        if offset == len(lines):
            raise RuntimeError("Asset stage patch lacks new file header")
        new = header(lines[offset], b"+++ ")
        offset += 1
        if old is None and new is None:
            raise RuntimeError("Asset stage patch lacks target")
        targets.update(n for n in (old, new) if n is not None)
        hunks = 0
        while offset < len(lines) and lines[offset].startswith(b"@@ "):
            match = re.fullmatch(rb"@@ -[0-9]+(?:,([0-9]+))? \+[0-9]+(?:,([0-9]+))? @@.*", lines[offset])
            if match is None:
                raise RuntimeError("Invalid asset stage unified hunk")
            remaining = [int(n) if n is not None else 1 for n in match.groups()]
            offset += 1
            hunks += 1
            while any(remaining):
                if offset == len(lines) or lines[offset][:1] not in (b" ", b"-", b"+"):
                    raise RuntimeError("Incomplete asset stage unified hunk")
                marker = lines[offset][:1]
                remaining[0] -= marker != b"+"
                remaining[1] -= marker != b"-"
                if min(remaining) < 0:
                    raise RuntimeError("Asset stage unified hunk count mismatch")
                offset += 1
                if offset < len(lines) and lines[offset] == b"\\ No newline at end of file":
                    offset += 1
        if not hunks:
            raise RuntimeError("Asset stage patch lacks unified hunks")
    return targets


@contextmanager
def stage_assets(stage, destination, inputs, outputs, patches):
    entries = stage.get("assets", [])
    if type(entries) is not list or len(entries) > MAX_STAGE_ASSETS:
        raise RuntimeError("Stage asset count bound exceeded")
    byte_limit = stage.get("asset_byte_limit", MAX_ASSET_BYTES)
    if type(byte_limit) is not int or not 1 <= byte_limit <= MAX_PINNED_ASSET_BYTES:
        raise RuntimeError("Invalid stage asset byte limit")
    snapshots, descriptors = [], []
    try:
        sources, targets = set(), set()
        for entry in entries:
            if type(entry) is not dict or set(entry) != {"source", "target", "sha256"} or not _sha256_value(entry["sha256"]):
                raise RuntimeError("Invalid stage asset record")
            source = _relative_asset_name(entry["source"])
            target = _relative_asset_name(entry["target"])
            folded_source, folded_target = source.casefold(), target.casefold()
            if (any(folded_source == n or folded_source.startswith(n + "/") or n.startswith(folded_source + "/") for n in sources)
                    or any(folded_target == n or folded_target.startswith(n + "/") or n.startswith(folded_target + "/") for n in targets)):
                raise RuntimeError("Duplicate or overlapping stage asset path")
            if not _sha256_value(inputs.get(target)) or outputs.get(target) != entry["sha256"]:
                raise RuntimeError("Stage asset requires exact old input and new output pins")
            sources.add(folded_source); targets.add(folded_target)
        patch_targets = set().union(*(_asset_stage_patch_targets(body) for body in patches)) if entries else set()
        if patch_targets.intersection(targets):
            raise RuntimeError("Stage patches and asset overlays cannot replace the same path")
        total, source_ids, target_ids = 0, set(), set()
        for entry in entries:
            source = _asset_path(ROOT, entry["source"])
            target = _asset_path(destination, entry["target"])
            data, source_stat = _read_stage_file(source, min(byte_limit, MAX_STAGE_ASSET_BYTES - total))
            total += len(data)
            original, before = _read_stage_file(target, byte_limit)
            if hashlib.sha256(data).hexdigest() != entry["sha256"] or hashlib.sha256(original).hexdigest() != inputs[entry["target"]]:
                raise RuntimeError("Stage asset input or source checksum mismatch")
            source_id, target_id = (source_stat.st_dev, source_stat.st_ino), (before.st_dev, before.st_ino)
            if source_id in source_ids or target_id in target_ids:
                raise RuntimeError("Stage asset file aliases are not supported")
            source_ids.add(source_id); target_ids.add(target_id)
            snapshots.append({"target": entry["target"], "path": target, "before": before, "data": data, "byte_limit": byte_limit})
        if source_ids.intersection(target_ids):
            raise RuntimeError("Stage asset source and destination files overlap")
        # Open all writers without truncation before any stage patch/write.
        for item in snapshots:
            fd = os.open(item["path"], os.O_WRONLY | getattr(os, "O_NOFOLLOW", 0) | getattr(os, "O_NONBLOCK", 0))
            descriptors.append(fd)
            if _asset_identity(os.fstat(fd)) != _asset_identity(item["before"]):
                raise RuntimeError("Stage asset target changed before mutation")
            item["fd"] = fd
        yield snapshots
    finally:
        for fd in descriptors:
            os.close(fd)


def _write_stage_assets(snapshots):
    # Check every held target before truncating the first one. Writes still use
    # the dependency's owned staging/installation boundary, not atomic multi-file IO.
    for item in snapshots:
        if (_asset_identity(os.fstat(item["fd"])) != _asset_identity(item["before"])
                or _asset_identity(os.stat(item["path"], follow_symlinks=False)) != _asset_identity(item["before"])):
            raise RuntimeError("Stage asset target drift before overlay")
    for item in snapshots:
        os.lseek(item["fd"], 0, os.SEEK_SET)
        os.ftruncate(item["fd"], 0)
        view = memoryview(item["data"])
        while view:
            written = os.write(item["fd"], view)
            if written <= 0:
                raise OSError("Stage asset write made no progress")
            view = view[written:]
        body, current = _read_stage_file(item["path"], item["byte_limit"])
        if (body != item["data"] or (current.st_dev, current.st_ino) != (item["before"].st_dev, item["before"].st_ino)):
            raise RuntimeError("Stage asset output mismatch")


def apply_stage(stage, destination):
    inputs = json.loads((ROOT / stage["inputs"]).read_text())
    outputs = json.loads((ROOT / stage["outputs"]).read_text()) if "outputs" in stage else {}
    patches = _stage_patch_snapshots(stage)
    with stage_assets(stage, destination, inputs, outputs, patches) as assets:
        asset_names = {item["target"] for item in assets}
        verify_inputs(destination, {n: v for n, v in inputs.items() if n not in asset_names})
        for body in patches:
            subprocess.run(["patch", "--batch", "-p1"], input=body, cwd=destination, check=True)
        _write_stage_assets(assets)
        verify_inputs(destination, {n: v for n, v in outputs.items() if n not in asset_names})


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
        apply_stage(stage, destination)


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
