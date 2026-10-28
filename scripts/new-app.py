#!/usr/bin/env python3
"""Create a standalone module with explicit local native replacements."""
import argparse
from pathlib import Path
import shutil
import shlex
import subprocess
import tempfile
import sys

if __name__ == "__main__":
    sys.dont_write_bytecode = True

import native_gpu
import native_linkers

ROOT = Path(__file__).resolve().parents[1]


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("destination", type=Path)
    parser.add_argument("--module", default="example.com/vellum-app", help="your application's Go module path")
    args = parser.parse_args()
    destination = args.destination.resolve()
    if destination.exists():
        parser.error("destination already exists; choose a new application directory")
    for dependency in [ROOT / "drivers/cogent/build/upstream", ROOT / "drivers/cogent/build/typesetting"]:
        if not (dependency / "go.mod").exists():
            parser.error("run drivers/cogent/prepare.sh first")
    gpu = native_gpu.sdk_candidate(ROOT)
    destination.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="vellum-new-app-", dir=destination.parent) as temporary:
        stage = Path(temporary) / "application"
        stage.mkdir()
        for name in ["go.mod", "go.sum", "LICENSE"]:
            shutil.copy2(ROOT / "examples/hello" / name, stage / name)
        for name in ["main.go", "model.go", "model_test.go"]:
            shutil.copy2(ROOT / "scripts/templates/app" / (name + ".txt"), stage / name)
        command = ["go", "mod", "edit", f"-module={args.module}",
                        f"-replace=github.com/dshills/vellum={ROOT}",
                        f"-replace=github.com/dshills/vellum/drivers/cogent={ROOT / 'drivers/cogent'}",
                        f"-replace=cogentcore.org/core={ROOT / 'drivers/cogent/build/upstream'}",
                        f"-replace=github.com/go-text/typesetting={ROOT / 'drivers/cogent/build/typesetting'}"]
        if gpu is not None:
            command.extend(f"-replace={module}={ROOT / relative}" for module, relative in native_linkers.replacements(gpu).items())
        subprocess.run(command, cwd=stage, check=True)
        stage.rename(destination)
    print(f"Created {destination}; source {shlex.quote(str(ROOT / 'drivers/cogent/env.sh'))} then run go test ./... and go run .")


if __name__ == "__main__":
    main()
