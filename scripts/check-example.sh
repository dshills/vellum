#!/bin/sh
set -eu
VELLUM_REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
. "$VELLUM_REPO_ROOT/drivers/cogent/env.sh"
cd "$VELLUM_REPO_ROOT/examples/hello"
go test -race ./...
go vet ./...
mkdir -p build
go build -o build/vellum-hello .
