#!/bin/sh
set -eu
cd "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. ./env.sh
action=${1:-build}
if [ "$#" -gt 0 ]; then shift; fi
case "$action" in build|test|run) ;; *) echo "Usage: $0 {build|test|run} [flags]" >&2; exit 2;; esac
cp go.mod remediation.mod
cp go.sum remediation.sum
python3 prepare.py
if [ "$action" = test ]; then
  go test -modfile=remediation.mod -count=1 -race -v ./semantics ./eventqueue ./textbridge
  (cd build/upstream && go test -count=1 -race -v ./text/textpos)
  (cd build/upstream && go test -count=1 -race -run '^TestVellum' -v ./system/driver/base)
  (cd build/upstream && go test -tags offscreen -count=1 -race -run '^TestVellum' -v ./core ./text/textcore)
  exit 0
fi
[ "$(go env GOOS)/$(go env GOARCH)" = darwin/arm64 ] || {
  echo "Native bridge verification supports macOS arm64 only." >&2; exit 2;
}
export CARGO_HOME="${FOUNDATION_CARGO_HOME:-/tmp/vellum-foundation-cargo}"
native_target=${FOUNDATION_RUST_TARGET:-/tmp/vellum-foundation-rust}
cargo build --release --locked --manifest-path native/accesskit-c/Cargo.toml --target-dir "$native_target"
mkdir -p build/native
cp "$native_target/release/libaccesskit.dylib" build/native/
install_name_tool -id @rpath/libaccesskit.dylib build/native/libaccesskit.dylib
go vet -modfile=remediation.mod ./...
go build -modfile=remediation.mod -o build/vellum-remediation .
python3 bundle.py
if [ "$action" = run ]; then
  exec build/Vellum-Remediation.app/Contents/MacOS/vellum-remediation "$@"
fi
