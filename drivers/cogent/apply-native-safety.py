#!/usr/bin/env python3
"""Apply the cumulative multiwindow safety patch to a prepared owned module."""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess

PATCH_ROOT = Path(__file__).resolve().parent / "native/patches/multiwindow"


def hashes_match(root, expected):
    for relative, digest in expected.items():
        path = root / relative
        actual = hashlib.sha256(path.read_bytes()).hexdigest() if path.exists() else None
        if actual != digest:
            return False
    return True


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("owned_module", type=Path)
    args = parser.parse_args()
    root = args.owned_module.resolve()
    metadata = json.loads((PATCH_ROOT / "metadata.json").read_text())
    patch = PATCH_ROOT / metadata["patch"]
    if hashlib.sha256(patch.read_bytes()).hexdigest() != metadata["sha256"]:
        raise RuntimeError("Native safety patch checksum mismatch")
    outputs = json.loads((PATCH_ROOT / "output-hashes.json").read_text())
    if hashes_match(root, outputs):
        print("Native multiwindow safety patch already verified")
        return
    inputs = json.loads((PATCH_ROOT / "original-hashes.json").read_text())
    if not hashes_match(root, inputs):
        raise RuntimeError("Native safety patch requires the exact state after patches 0001–0015")
    # Check every hunk before modifying the owned module. If application still
    # fails (for example, disk exhaustion), rerun prepare.sh to regenerate it.
    subprocess.run(["patch", "--batch", "--dry-run", "-p1", "-i", str(patch)], cwd=root, check=True)
    subprocess.run(["patch", "--batch", "-p1", "-i", str(patch)], cwd=root, check=True)
    if not hashes_match(root, outputs):
        raise RuntimeError("Native safety patch output checksum mismatch")
    print("Applied and verified native multiwindow safety patch 0016")


if __name__ == "__main__":
    main()
