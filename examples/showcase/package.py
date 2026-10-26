#!/usr/bin/env python3
"""Build three private, ad-hoc-signed showcase apps into one NEW local directory."""
import argparse
import hashlib
import json
from pathlib import Path
import platform
import sys

sys.dont_write_bytecode = True
MODULE = Path(__file__).resolve().parent
ROOT = MODULE.parents[1]
sys.path.insert(0, str(ROOT / "scripts"))
from child_process import run_child
from sdk_environment import child_environment


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True, help="new directory; prefer local storage outside synchronized Documents")
    args = parser.parse_args()
    if platform.system() != "Darwin":
        parser.error("macOS is required")
    output = args.output.absolute()
    output.mkdir()  # exclusive; never overwrite a prior package
    binaries = output / "binaries"
    binaries.mkdir()
    env = child_environment()
    def run(argv):
        run_child(argv, cwd=MODULE, env=env, timeout=600, check=True)
    run(["go", "run", "./cmd/artwork", "-output", str(output / "icons")])
    records = []
    for name in ("forma", "luma", "pulse"):
        binary = binaries / name
        bundle = output / (name.title() + ".app")
        run(["go", "build", "-o", str(binary), "./cmd/" + name])
        run([sys.executable, "-B", str(ROOT / "scripts/package-app.py"), "--binary", str(binary),
             "--output", str(bundle), "--name", name.title(), "--identifier", "com.dshills.vellum.showcase." + name,
             "--version", "1.0.0", "--build-version", "1", "--icon", str(output / "icons" / (name + ".png")), "--sign-identity", "-"])
        run(["codesign", "--verify", "--deep", "--strict", str(bundle)])
        archive = output / (name.title() + ".zip")
        run(["ditto", "-c", "-k", "--keepParent", str(bundle), str(archive)])
        records.append({"name": name.title(), "archive": archive.name, "sha256": hashlib.sha256(archive.read_bytes()).hexdigest()})
    (output / "SHA256SUMS").write_text("".join(r["sha256"] + "  " + r["archive"] + "\n" for r in records))
    source_files = [*sorted((MODULE / "demo").glob("*.go")), *sorted((MODULE / "cmd").rglob("*.go")), MODULE / "go.mod", MODULE / "go.sum", Path(__file__)]
    (output / "manifest.json").write_text(json.dumps({"private": True, "notarized": False, "apps": records,
        "source_sha256": {str(p.relative_to(MODULE)): hashlib.sha256(p.read_bytes()).hexdigest() for p in source_files}}, indent=2) + "\n")
    print("Private showcase apps: " + str(output))


if __name__ == "__main__":
    main()
