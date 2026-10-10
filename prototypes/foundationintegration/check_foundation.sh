#!/bin/sh
# Requires the owned dependencies prepared by run.sh. Exit 1 means G1 is open;
# exit 2 (or a build/test error) means this run did not produce valid evidence.
set -eu
cd "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. ./env.sh
cleanup(){ if [ -n "${VELLUM_OWNED_GO_CACHE:-}" ]; then rm -rf "$VELLUM_OWNED_GO_CACHE"; fi; }
trap cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM
out=${1:-build/acceptance}
mkdir -p "$out"
go vet -modfile=integration.mod ./... > "$out/vet.txt" 2>&1
go test -race -count=1 -timeout 30s -modfile=integration.mod ./transport ./inputbridge ./semantics > "$out/race-tests.txt" 2>&1
go build -tags offscreen -modfile=integration.mod -o build/foundationaudit ./cmd/foundationaudit
go build -modfile=integration.mod -o build/lifecyclecheck ./cmd/lifecyclecheck
set +e
./build/foundationaudit -nogui -font-cache=../fontresources/build/font-cache > "$out/acceptance-audit.json" 2> "$out/acceptance-audit.log"
status=$?
set -e
printf '%s\n' "$status" > "$out/acceptance-audit.exit"
if [ "$status" -gt 1 ]; then cat "$out/acceptance-audit.log" >&2; exit "$status"; fi
./build/lifecyclecheck -cycles=1000 > "$out/native-queries-lifetime.json" 2> "$out/native-queries-lifetime.log"
python3 verify_acceptance.py "$out"
printf '%s\n' 'Evidence validated. G1 remains open; see the reported contracts and unverified checks.'
exit 1
