#!/bin/sh
set -eu
cd "$(dirname "$0")"
export GOWORK=off
export CGO_LDFLAGS_ALLOW='-Wl,-rpath,@executable_path/native'
exec go run . "$@"
