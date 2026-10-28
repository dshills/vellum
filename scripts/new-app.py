#!/usr/bin/env python3
"""Create a standalone module with explicit local native replacements."""
import argparse
import hashlib
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


def copy_document_template(stage, module):
    """Copy the tested app recipe into the caller's module with private imports."""
    mapping = {
        "github.com/dshills/vellum/examples/macapps/notes": module + "/internal/notebook",
        "github.com/dshills/vellum/examples/macapps/artwork": module + "/internal/artwork",
        "com.dshills.vellum.notes.toolbar.v1": "vellum.document." + hashlib.sha256(module.encode()).hexdigest()[:16] + ".toolbar.v1",
        "Vellum Notes": "Vellum Document",
    }
    for source, relative in [(ROOT / "examples/macapps/notes", "internal/notebook"),
                             (ROOT / "examples/macapps/artwork", "internal/artwork"),
                             (ROOT / "examples/macapps/cmd/notes", ".")]:
        destination = stage / relative
        destination.mkdir(parents=True, exist_ok=True)
        paths = sorted(source.glob("*.go")) if relative != "." else [source / "toolbar.go"]
        if not paths or not all(path.is_file() for path in paths):
            raise ValueError("document template source missing: " + str(source))
        for path in paths:
            text = path.read_text(encoding="utf-8")
            for old, new in mapping.items():
                text = text.replace(old, new)
            (destination / path.name).write_text(text, encoding="utf-8")
    main = (ROOT / "scripts/templates/document/main.go.txt").read_text(encoding="utf-8")
    (stage / "main.go").write_text(main.replace("__VELLUM_MODULE__", module), encoding="utf-8")
    (stage / "README.md").write_text("# Document starter\n\nRun go test -race -tags offscreen ./... and go run . -data-dir /absolute/private/app-directory.\n\nThis complete app recipe includes Save/Open/Save As, dirty decisions, recovery, prose editing, full-color native toolbar and Undo/Redo. Its application-owned sources live in internal/notebook. The explicit private directory prevents using another app's notebook. Do not reuse one directory across different applications. Recovery and file workers retire before the repository lock is released. This is a larger reference starter; the default template remains minimal.\n")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("destination", type=Path)
    parser.add_argument("--module", default="example.com/vellum-app", help="your application's Go module path")
    parser.add_argument("--template", choices=["minimal", "document"], default="minimal")
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
        if args.template == "document":
            copy_document_template(stage, args.module)
        else:
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
