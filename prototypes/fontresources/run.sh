#!/bin/sh
set -eu
cd "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. ./env.sh
cleanup(){ if [ -n "${private_build:-}" ];then rm -rf "$private_build";fi; if [ -n "${VELLUM_OWNED_GO_CACHE:-}" ];then rm -rf "$VELLUM_OWNED_GO_CACHE";fi; }
trap cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM
# The owned Cogent copy can be prepared without compiling the native fixture.
(cd ../foundationintegration && cp go.mod integration.mod && cp go.sum integration.sum && python3 prepare.py)
cp go.mod font.mod
cp go.sum font.sum
python3 prepare.py
go vet -modfile=font.mod ./...
go test -race -count=1 -timeout 30s -modfile=font.mod -run 'TestDeferredSbix|TestStableDeferred|TestDeferredGlyf' github.com/go-text/typesetting/font github.com/go-text/typesetting/fontscan
go build -modfile=font.mod -tags offscreen -o build/fontcheck ./cmd/fontcheck
mkdir -p build/font-cache
./build/fontcheck -nogui -font-cache=build/font-cache -repeat=3 -profile=evidence/font-lazy-outline.heap.pprof > evidence/font-lazy-outline.json 2> evidence/font-lazy-outline.log
cp font.mod eager.mod
cp font.sum eager.sum
go mod edit -modfile=eager.mod -dropreplace=github.com/go-text/typesetting
go build -modfile=eager.mod -tags offscreen -o build/fontcheck-eager ./cmd/fontcheck
./build/fontcheck-eager -nogui -font-cache=build/font-cache -repeat=1 > evidence/font-eager-outline.json 2> evidence/font-eager-outline.log
python3 verify.py
