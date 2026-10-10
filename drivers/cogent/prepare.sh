#!/bin/sh
set -eu
VELLUM_REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
. "$VELLUM_REPO_ROOT/prototypes/foundationintegration/env.sh"
cd "$VELLUM_REPO_ROOT/prototypes/fontresources"
python3 prepare.py
cd "$VELLUM_REPO_ROOT/prototypes/foundationintegration"
cp go.mod integration.mod
cp go.sum integration.sum
python3 prepare.py
