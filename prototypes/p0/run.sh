#!/bin/sh
set -eu
p0_root=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
. "$p0_root/env.sh"
p0_action=${1:-}
if [ "$#" -gt 0 ]; then shift; fi
case "$p0_action" in
  gio|cogent)
    cd "$p0_root/$p0_action"
    exec go run . "$@"
    ;;
  gpu)
    cd "$p0_root/cogent"
    exec go run . -gpu-probe "$@"
    ;;
  test)
	    mkdir -p "$p0_root/build"
    cd "$p0_root/model"
    go test -race -v ./...
    go vet ./...
    for p0_candidate in gio cogent; do
      cd "$p0_root/$p0_candidate"
      go vet ./...
      go build -o "$p0_root/build/vellum-p0-$p0_candidate" .
    done
    ;;
  build)
    mkdir -p "$p0_root/build"
    for p0_candidate in gio cogent; do
      cd "$p0_root/$p0_candidate"
      go build -o "$p0_root/build/vellum-p0-$p0_candidate" .
    done
    ;;
  *)
    echo 'Usage: ./run.sh {gio|cogent|gpu|test|build} [Go program flags]' >&2
    exit 2
    ;;
esac
