#!/bin/sh
# Source this from a prototype module to isolate disposable build caches.
export GOWORK=off
export GOCACHE="${P0_CACHE_DIR:-/tmp/vellum-p0-cache}/build"
export GOMODCACHE="${P0_CACHE_DIR:-/tmp/vellum-p0-cache}/mod"
export GOPATH="${P0_CACHE_DIR:-/tmp/vellum-p0-cache}/gopath"
