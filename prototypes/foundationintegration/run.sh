#!/bin/sh
set -eu
cd "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. ./env.sh
cleanup(){ if [ -n "${private_build:-}" ];then rm -rf "$private_build";fi; if [ -n "${VELLUM_OWNED_GO_CACHE:-}" ];then rm -rf "$VELLUM_OWNED_GO_CACHE";fi; }
trap cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM
cp go.mod integration.mod
cp go.sum integration.sum
python3 prepare.py
(cd ../fontresources && cp go.mod font.mod && cp go.sum font.sum && python3 prepare.py)
go mod edit -modfile=integration.mod -replace=github.com/go-text/typesetting=../fontresources/build/typesetting
# Build the pinned AccessKit library using a private default Cargo home/target.
# Explicit overrides refer only to trusted directories selected by the caller.
private_build=$(mktemp -d "${TMPDIR:-/tmp}/vellum-native.XXXXXXXX")
export CARGO_HOME="${VELLUM_CARGO_HOME:-$private_build/cargo}"
target_dir="${VELLUM_RUST_TARGET:-$private_build/target}"
cargo build --release --locked --manifest-path ../nativeinput/native/accesskit-c/Cargo.toml --target-dir "$target_dir"
mkdir -p build/native
cp "$target_dir/release/libaccesskit.dylib" build/native/
go vet -modfile=integration.mod ./...
go test -race -count=1 -timeout 30s -modfile=integration.mod ./transport ./inputbridge ./semantics
go build -modfile=integration.mod -o build/integration-fixture .
go build -modfile=integration.mod -o build/lifecyclecheck ./cmd/lifecyclecheck
