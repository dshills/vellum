#!/bin/sh
set -eu
VELLUM_REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
. "$VELLUM_REPO_ROOT/drivers/cogent/env.sh"
for VELLUM_EXAMPLE_NAME in hello drawing viewport3d imaging editor grid placement; do
cd "$VELLUM_REPO_ROOT/examples/$VELLUM_EXAMPLE_NAME"
go test -race ./...
go vet ./...
mkdir -p build
go build -o "build/vellum-$VELLUM_EXAMPLE_NAME" .
done

cd "$VELLUM_REPO_ROOT/examples/replay"
go test -race ./...
go vet ./...
mkdir -p build
go build -o build/vellum-workspace ./cmd/workspace

# Standard-library-only semantic client tests do not start a native window.
cd "$VELLUM_REPO_ROOT/examples/automation"
go test -race ./...
go vet ./...
mkdir -p build
go build -o build/vellum-automation .

# Independently consumable native macOS-style application examples.
cd "$VELLUM_REPO_ROOT/examples/macapps"
go test -race ./...
go vet ./...
mkdir -p build
for VELLUM_APP_NAME in calculator notes weather reminders; do
 go build -o "build/vellum-$VELLUM_APP_NAME" "./cmd/$VELLUM_APP_NAME"
done
