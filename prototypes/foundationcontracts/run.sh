#!/bin/sh
set -eu
cd "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. ./env.sh
cp go.mod contracts.mod
cp go.sum contracts.sum
python3 prepare.py
go vet -modfile=contracts.mod ./...
go test -race -count=1 -timeout 30s -modfile=contracts.mod ./mailbox
(cd build/upstream && go test -race -count=1 -timeout 30s -run '^(TestContracts|TestFoundation|TestVellum)' ./text/lines ./text/textcore ./text/shaped/shapers/shapedgt)
go build -tags offscreen -modfile=contracts.mod -o build/contractcheck ./cmd/contractcheck
