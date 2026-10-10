#!/bin/sh
set -eu
cd "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. ./env.sh
(cd build/upstream && go test -race -count=1 -timeout 30s -run '^(TestFoundation|TestVellum)' -v ./text/lines ./text/textcore ./text/shaped/shapers/shapedgt)
go test -race -count=1 -modfile=checks.mod example.com/vellum/nativeinput/semantics example.com/vellum/nativeinput/eventqueue example.com/vellum/nativeinput/textbridge example.com/vellum/nativeinput/inputbridge
