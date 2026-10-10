#!/bin/sh
set -eu
VELLUM_PLATFORM_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$VELLUM_PLATFORM_ROOT"
. ./drivers/cogent/env.sh
VELLUM_PLATFORM_MODE=build
VELLUM_PLATFORM_CLIPBOARD=false
for argument in "$@"; do
 case "$argument" in
 --runtime) VELLUM_PLATFORM_MODE=runtime ;;
 --qualification) VELLUM_PLATFORM_MODE=qualification ;;
 --clipboard-write) VELLUM_PLATFORM_CLIPBOARD=true ;;
 *) echo "usage: check-platform.sh [--runtime|--qualification] [--clipboard-write]" >&2; exit 2 ;;
 esac
done
if [ ! -f drivers/cogent/build/upstream/go.mod ]; then
 echo 'Run drivers/cogent/prepare.sh before native platform checks.' >&2
 exit 1
fi
go test -race ./...
go vet ./...
cd drivers/cogent
go test -race ./internal/platformio ./cmd/platform-smoke
go test ./...
go vet ./...
go build -o build/vellum-platform-smoke ./cmd/platform-smoke
case "$VELLUM_PLATFORM_MODE" in
 runtime) build/vellum-platform-smoke -clipboard-write="$VELLUM_PLATFORM_CLIPBOARD" ;;
 qualification) build/vellum-platform-smoke -require-complete -clipboard-write="$VELLUM_PLATFORM_CLIPBOARD" ;;
 *) echo 'PASS build/unit checks only; target runtime and qualification unverified.' ;;
esac
