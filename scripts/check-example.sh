#!/bin/sh
set -eu
VELLUM_REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
. "$VELLUM_REPO_ROOT/drivers/cogent/env.sh"
for VELLUM_EXAMPLE_NAME in hello drawing viewport3d; do
cd "$VELLUM_REPO_ROOT/examples/$VELLUM_EXAMPLE_NAME"
go test -race ./...
go vet ./...
mkdir -p build
go build -o "build/vellum-$VELLUM_EXAMPLE_NAME" .
done
