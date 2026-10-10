#!/bin/sh
set -eu
VELLUM_REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
. "$VELLUM_REPO_ROOT/drivers/cogent/env.sh"
cd "$VELLUM_REPO_ROOT"
go test ./... -run '^$' -bench . -benchmem -count 3
cd "$VELLUM_REPO_ROOT/examples/hello"
go test . -run '^$' -bench . -benchmem -count 3
# Native measurements launch visible windows. Opt in with --native.
if [ "${1:-}" = '--native' ]; then
 "$VELLUM_REPO_ROOT/scripts/check-example.sh"
 python3 "$VELLUM_REPO_ROOT/scripts/benchmark-native.py"
fi
