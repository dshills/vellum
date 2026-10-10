#!/usr/bin/env python3
"""Compile both service adapters against pinned Cogent on the current host.

This validates shared API/ownership code, not Windows/Linux cgo or target runtime.
"""
from pathlib import Path
import os
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]
DRIVER = ROOT / "drivers/cogent"


def without_build_constraint(source):
    first, rest = source.split("\n", 1)
    if not first.startswith("//go:build "):
        raise RuntimeError("fixture input is missing an explicit build constraint")
    return rest


def main():
    with tempfile.TemporaryDirectory(prefix="vellum-platform-fixture-") as directory:
        root = Path(directory)
        (root / "go.mod").write_bytes((DRIVER / "go.mod").read_bytes())
        # go mod edit parses both single-line and grouped directives and quotes
        # every native absolute path, including spaces and Windows drive letters.
        subprocess.run([
            "go", "mod", "edit", "-module=example.com/vellum/cogent/platformfixture",
            "-require=example.com/vellum/cogent@v0.0.0",
            f"-replace=example.com/vellum={ROOT}",
            f"-replace=cogentcore.org/core={DRIVER / 'build/upstream'}",
            f"-replace=github.com/go-text/typesetting={DRIVER / 'build/typesetting'}",
            f"-replace=example.com/vellum/cogent={DRIVER}",
        ], cwd=root, check=True)
        (root / "go.sum").write_bytes((DRIVER / "go.sum").read_bytes())
        for name in ("services_desktop.go", "services_clipboard_desktop_test.go", "services_menus_desktop.go", "services_menus_desktop_test.go"):
            (root / name).write_text(without_build_constraint((DRIVER / name).read_text(encoding="utf-8")), encoding="utf-8")
        (root / "services.go").write_bytes((DRIVER / "services.go").read_bytes())
        for target in ("linux", "windows"):
            # Rename to remove Go's filename GOOS selection, as this is explicitly
            # a current-host API fixture rather than a target-platform build.
            (root / "choosefile_adapter.go").write_text(without_build_constraint((DRIVER / f"services_{target}.go").read_text(encoding="utf-8")), encoding="utf-8")
            subprocess.run(["go", "test", "-race", "./..."], cwd=root, check=True)
            print(f"PASS {target} service current-host API fixture; target runtime unverified", flush=True)


if __name__ == "__main__":
    main()
