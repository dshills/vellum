#!/bin/sh
# Source this file from any working directory before native Go commands.
export GOWORK=off
if [ -n "${VELLUM_CACHE_DIR:-}" ]; then
 P0_CACHE_DIR="$VELLUM_CACHE_DIR"
fi
if [ -z "${P0_CACHE_DIR:-}" ]; then
 P0_CACHE_DIR=$(mktemp -d "${TMPDIR:-/tmp}/vellum-cache.XXXXXXXX") || return 1
 export P0_CACHE_DIR
fi
export VELLUM_CACHE_DIR="$P0_CACHE_DIR"
export P0_CACHE_DIR
export GOCACHE="$P0_CACHE_DIR/build"
export GOMODCACHE="$P0_CACHE_DIR/mod"
export GOPATH="$P0_CACHE_DIR/gopath"
export CGO_LDFLAGS_ALLOW='-Wl,-rpath,@executable_path/native'
