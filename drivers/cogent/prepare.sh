#!/bin/sh
set -eu
VELLUM_DRIVER_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
. "$VELLUM_DRIVER_ROOT/env.sh"
cd "$VELLUM_DRIVER_ROOT"
python3 prepare.py
