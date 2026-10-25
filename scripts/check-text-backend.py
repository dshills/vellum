#!/usr/bin/env python3
"""Run exact-text/rendering regressions with owned dependencies and offscreen DPI.

Copy module metadata so preparation outputs remain immutable. The direct upstream
module does not inherit the driver's typesetting replacement. These fixtures need
an initialized offscreen display; separate native consumers test actual AppKit.
"""
import os
from pathlib import Path
import shutil
import sys
import tempfile

from child_process import run_child
from sdk_environment import child_environment

ROOT = Path(__file__).resolve().parents[1]
TEXT_PACKAGES = ("./events", "./core", "./text/textcore", "./text/lines")
# Cold compilation includes cgo and large font/text test fixtures. Keep it
# separate from the unchanged ninety-second execution deadline.
TEXT_TESTS = "TestVellumComposition|TestVellumNativeCommit|TestVellumDeferred|TestVellumEventBoundary|TestVellumField|TestVellumEditingSnapshot|TestVellumEditor|TestVellumNativeEditor|TestContractsExact"


def main():
    upstream = ROOT / "drivers/cogent/build/upstream"
    typesetting = ROOT / "drivers/cogent/build/typesetting"
    if not all((p / "go.mod").is_file() for p in (upstream, typesetting)):
        raise ValueError("Run drivers/cogent/prepare.sh before text backend checks")
    with tempfile.TemporaryDirectory(prefix="vellum text backend ") as temporary:
        stage = Path(temporary)
        for name in ("go.mod", "go.sum"):
            shutil.copy2(upstream / name, stage / name)
        modfile = "-modfile=" + str(stage / "go.mod")
        env = child_environment()
        # Installed fonts provide the fixture's Hebrew/emoji fallback corpus.
        # This cache is owned build data and does not change system preferences.
        cache = os.environ.get("VELLUM_FONT_CACHE")
        if not cache:
            owned = os.environ.get("VELLUM_CACHE_DIR")
            cache = str(Path(owned) / "font-cache") if owned else str(stage / "font-cache")
        env["VELLUM_FONT_CACHE"] = cache
        run_child(["go", "mod", "edit", modfile,
                   "-replace=github.com/go-text/typesetting=" + str(typesetting)],
                  cwd=upstream, env=env, timeout=60, check=True)
        for index, package in enumerate(TEXT_PACKAGES):
            binary = stage / ("text-backend-" + str(index) + (".exe" if sys.platform == "win32" else ""))
            run_child(["go", "test", modfile, "-c", "-race", "-tags", "offscreen",
                       "-o", str(binary), package],
                      cwd=upstream, env=env, timeout=300, check=True)
            if not binary.is_file():
                raise ValueError("required text test binary missing: " + package)
            run_child([str(binary), "-test.run=" + TEXT_TESTS,
                       "-test.count=1", "-test.timeout=90s"],
                      cwd=upstream / package.removeprefix("./"),
                      env=env, timeout=120, check=True)


if __name__ == "__main__":
    main()
