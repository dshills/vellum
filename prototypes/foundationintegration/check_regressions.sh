#!/bin/sh
set -eu
cd "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. ./env.sh
mkdir -p build
if ! mkdir build/regression-run.lock 2>/dev/null; then
 echo "A regression run already owns this checkout; TeX output files require serial runs." >&2
 if [ -n "${VELLUM_OWNED_GO_CACHE:-}" ]; then rm -rf "$VELLUM_OWNED_GO_CACHE"; fi
 exit 1
fi
cleanup(){ rmdir build/regression-run.lock; if [ -n "${VELLUM_OWNED_GO_CACHE:-}" ]; then rm -rf "$VELLUM_OWNED_GO_CACHE"; fi; }
trap cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM
export VELLUM_FONT_CACHE="${VELLUM_FONT_CACHE:-$PWD/../fontresources/build/font-cache}"
out=${1:-build/regression-checks}
mkdir -p "$out"
cp integration.mod build/regression.mod
cp integration.sum build/regression.sum
go test -race -tags offscreen -modfile=integration.mod cogentcore.org/core/tree cogentcore.org/core/text/parse -count=1 -timeout=60s > "$out/tree-parse-race.txt" 2>&1
go test -tags offscreen -modfile=integration.mod cogentcore.org/core/text/lines cogentcore.org/core/text/textcore cogentcore.org/core/core -count=1 -timeout=90s > "$out/editor-core.txt" 2>&1
# The separate TeX image corpus invokes hundreds of TeX/DVI conversions. It is
# deliberately given its own larger time budget; a timeout is never a pass.
go test -mod=mod -tags offscreen -modfile=build/regression.mod cogentcore.org/core/text/... -count=1 -timeout=5m > "$out/text-all.txt" 2>&1
