#!/bin/sh
set -eu
VELLUM_REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$VELLUM_REPO_ROOT"
. ./prototypes/foundationintegration/env.sh
if [ ! -f prototypes/foundationintegration/build/upstream/go.mod ] || [ ! -f prototypes/fontresources/build/typesetting/go.mod ]; then
 echo 'Run drivers/cogent/prepare.sh before native checks.' >&2
 exit 1
fi
go test -race ./...
go vet ./...
cd drivers/cogent
go test ./...
go vet ./...
mkdir -p build
go build -o build/vellum-workbench ./cmd/workbench
