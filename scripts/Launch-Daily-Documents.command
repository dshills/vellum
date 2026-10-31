#!/bin/zsh
set -eu
script_dir="${0:A:h}"
cd "$script_dir"
exec python3 -B "$script_dir/daily-documents.py" "$@"
