"""Create an owned module copy with the four previous and two local foundation patches."""
from pathlib import Path
import hashlib
import json
import os
import shutil
import subprocess

root = Path(__file__).resolve().parent
previous = root.parent / "nativeinput"
pin = "cogentcore.org/core@v0.3.46-0.20261009095036-5cc519b24044"
info = json.loads(subprocess.check_output(["go", "mod", "download", "-json", pin]))
original = Path(info["Dir"])
for rel, expected in json.loads((previous / "patches/original-hashes.json").read_text()).items():
    path = original / rel
    actual = hashlib.sha256(path.read_bytes()).hexdigest() if path.exists() else None
    if actual != expected:
        raise RuntimeError(f"Pinned input mismatch: {rel}")
destination = root / "build/upstream"
if destination.exists():
    destination.chmod(0o755)
    for path in destination.rglob("*"):
        if path.is_dir():
            path.chmod(0o755)
    shutil.rmtree(destination)
shutil.copytree(original, destination)
destination.chmod(0o755)
for path in destination.rglob("*"):
    if path.is_dir():
        path.chmod(0o755)
    elif path.is_file():
        path.chmod(path.stat().st_mode | 0o200)
for patch in sorted((previous / "patches").glob("*.patch")):
    subprocess.run(["patch", "--batch", "-p1", "-i", str(patch)], cwd=destination,
                   check=True, env=dict(os.environ, TMPDIR="/private/tmp"))
for rel, expected in json.loads((root / "patches/original-hashes.json").read_text()).items():
    path = destination / rel
    actual = hashlib.sha256(path.read_bytes()).hexdigest() if path.exists() else None
    if actual != expected:
        raise RuntimeError(f"Foundation patch input mismatch: {rel}")
for patch in sorted((root / "patches").glob("*.patch")):
    subprocess.run(["patch", "--batch", "-p1", "-i", str(patch)], cwd=destination,
                   check=True, env=dict(os.environ, TMPDIR="/private/tmp"))
subprocess.run(["go", "mod", "edit", "-modfile=checks.mod",
                "-replace=cogentcore.org/core=./build/upstream"], cwd=root, check=True)
