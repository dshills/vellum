#!/bin/sh
set -eu
cd "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. ./env.sh
cp go.mod checks.mod
cp go.sum checks.sum
python3 prepare.py
go vet -modfile=checks.mod ./...
mkdir -p build/native
if [ ! -f ../nativeinput/build/native/libaccesskit.dylib ]; then
  echo "Build the nativeinput prototype first to provide the pinned AccessKit library." >&2
  exit 1
fi
cp ../nativeinput/build/native/libaccesskit.dylib build/native/
for name in textcheck fontcheck queuecheck; do
  go build -tags offscreen -modfile=checks.mod -o "build/$name" "./cmd/$name"
done
go build -modfile=checks.mod -o build/lifecyclecheck ./cmd/lifecyclecheck
