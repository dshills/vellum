#!/usr/bin/env python3
"""Verify and launch the frozen private Documents candidate for daily evaluation."""
import argparse
import json
import os
from pathlib import Path
import stat
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
DATA = Path.home() / 'Library/Application Support/Vellum Internal Evaluation/Documents'
CANDIDATE = 'build24-b9212489afc9'


def private_directory(path, create=False):
    """Refuse links and unsafe existing evaluation directories; never repair them."""
    for ancestor in reversed((path, *path.parents)):
        if ancestor.is_symlink():
            raise ValueError(f'symlink in data path: {ancestor}')
    if path.exists():
        info = path.stat()
        if not stat.S_ISDIR(info.st_mode) or info.st_uid != os.getuid() or info.st_mode & 0o077:
            raise ValueError(f'data directory must be owned by you with mode 0700: {path}')
    elif create:
        path.mkdir(mode=0o700, parents=True)
        private_directory(path)


def version(value):
    return tuple(int(part) for part in value.split('.'))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check-only', action='store_true', help='verify without creating data or launching')
    args = parser.parse_args()
    if sys.platform != 'darwin':
        parser.error('macOS is required')
    try:
        index = json.loads((ROOT / 'release/private-macos-candidate.json').read_text())
        if index['id'] != CANDIDATE:
            raise ValueError('daily evaluation candidate changed; requalify before updating this launcher')
        host = subprocess.check_output(['/usr/bin/sw_vers', '-productVersion'], text=True).strip()
        if version(host) < version(index['minimum_macos']):
            raise ValueError('host is older than the candidate minimum macOS')
        subprocess.run([sys.executable, '-B', str(ROOT / 'scripts/private-candidate.py'), '--verify-local'], check=True)
        bundle = Path.home() / index['installation'] / 'apps/Vellum Documents.app'
        executable = str(bundle / 'Contents/MacOS/Vellum Documents')
        processes = subprocess.check_output(['/bin/ps', '-axo', 'command='], text=True).splitlines()
        if any(line.strip() == executable or line.strip().startswith(executable + ' ') for line in processes):
            raise ValueError('Documents candidate is already running; quit it normally before using this launcher')
        private_directory(DATA)
        if args.check_only:
            print(f'Ready: {CANDIDATE}; data: {DATA} (not created or opened)')
            return
        private_directory(DATA, create=True)
        subprocess.run(['/usr/bin/open', '-n', str(bundle), '--args', '-data-dir', str(DATA)], check=True)
        print(f'Launch requested: {CANDIDATE}; data: {DATA}')
    except (OSError, ValueError, KeyError, subprocess.CalledProcessError) as error:
        parser.exit(1, str(error) + '\n')


if __name__ == '__main__':
    main()
