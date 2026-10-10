#!/bin/sh
set -eu
cd "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. ./env.sh
action=${1:-build}
if [ "$#" -gt 0 ]; then shift; fi
case "$action" in
  test)
    go test -count=1 -race -v ./semantics ./eventqueue
    exit 0
    ;;
  build|run) ;;
  *) echo "Usage: $0 {build|test|run} [experiment flags]" >&2; exit 2 ;;
esac
[ "$(go env GOOS)/$(go env GOARCH)" = darwin/arm64 ] || {
  echo "This disposable native bridge is tested only on macOS arm64." >&2; exit 2;
}
export CARGO_HOME="${FOUNDATION_CARGO_HOME:-/tmp/vellum-foundation-cargo}"
native_target=${FOUNDATION_RUST_TARGET:-/tmp/vellum-foundation-rust}
cargo build --release --locked --manifest-path native/accesskit-c/Cargo.toml --target-dir "$native_target"
mkdir -p build/native
cp "$native_target/release/libaccesskit.dylib" build/native/
install_name_tool -id @rpath/libaccesskit.dylib build/native/libaccesskit.dylib
go vet ./...
go build -o build/vellum-foundation .
python3 bundle.py
if [ "$action" = run ]; then
  exec build/Vellum-Foundation.app/Contents/MacOS/vellum-foundation "$@"
fi
