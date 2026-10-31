#!/usr/bin/env python3
"""Verify the selected local suite, restore its exact archive if needed, then open the hub."""
import argparse
import json
from pathlib import Path
import subprocess
import sys
import tempfile
sys.dont_write_bytecode = True
from bundle_tools import HERE, digest, verify_suite, publish_receipt


def selected_suite(receipt):
    archive = Path(receipt['archive'])
    suite = Path(receipt['suite'])
    if not archive.is_absolute() or not suite.is_absolute():
        raise ValueError('selected archive and suite paths must be absolute')
    if digest(archive) != receipt['archive_sha256']:
        raise ValueError('selected archive checksum differs')
    if not suite.is_dir():
        restored = Path(tempfile.mkdtemp(prefix='vellum-demo-restored-', dir='/private/tmp'))
        subprocess.run(['ditto', '-x', '-k', str(archive), str(restored)], check=True)
        suite = restored / 'Vellum Demo Suite'
    manifest = verify_suite(suite, receipt['manifest_sha256'])
    if manifest['version'] != receipt['version'] or manifest['source_commit'] != receipt['source_commit']:
        raise ValueError('suite identity differs from selected delivery')
    return suite


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--verify-only', action='store_true')
    args = parser.parse_args()
    try:
        receipt = json.loads((HERE / 'build/latest-suite.json').read_text())
        suite = selected_suite(receipt)
        if str(suite) != receipt['suite']:
            receipt['suite'] = str(suite)
            publish_receipt(receipt)
        if args.verify_only:
            print('Verified ' + receipt['version'] + ' at ' + receipt['source_commit'])
        else:
            subprocess.run(['open', '-n', str(suite / 'Vellum Demo Suite.app')], check=True)
    except (OSError, ValueError, KeyError, TypeError, subprocess.CalledProcessError) as error:
        parser.exit(1, str(error) + '\nBuild with ./build.sh before launching.\n')


if __name__ == '__main__':
    main()
