#!/bin/sh
export GOWORK=off
VELLUM_OWNED_GO_CACHE=""
if [ -z "${P0_CACHE_DIR:-}" ]; then
 P0_CACHE_DIR=$(mktemp -d "${TMPDIR:-/tmp}/vellum-cache.XXXXXXXX") || return 1
 export P0_CACHE_DIR
 VELLUM_OWNED_GO_CACHE="$P0_CACHE_DIR"
fi
export GOCACHE="$P0_CACHE_DIR/build"
export GOMODCACHE="$P0_CACHE_DIR/mod"
export GOPATH="$P0_CACHE_DIR/gopath"
export CGO_LDFLAGS_ALLOW='-Wl,-rpath,@executable_path/native'
