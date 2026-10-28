#!/usr/bin/env python3
"""Report the exact maintained dependency patch burden; --check refuses drift."""
import argparse
import hashlib
import json
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def inventory(driver):
    manifest = driver / 'native/dependencies.json'
    records = []
    for dep in json.loads(manifest.read_text())['dependencies']:
        targets, identities = set(), []
        for stage in dep['stages']:
            for patch in stage['patches']:
                path = (driver / patch['file']).resolve()
                if not path.is_relative_to(driver.resolve()):
                    raise ValueError('patch outside driver')
                data = path.read_bytes()
                digest = hashlib.sha256(data).hexdigest()
                if digest != patch['sha256']:
                    raise ValueError('patch digest mismatch: ' + patch['file'])
                identities.append([patch['file'], digest])
                # Git parses complete hunks, additions/deletions and path prefixes;
                # --numstat is read-only and never applies the patch.
                parsed = subprocess.run(['git', 'apply', '--numstat', '-z'], input=data,
                                        capture_output=True, check=True, timeout=10)
                for record in parsed.stdout.split(b'\0'):
                    if record:
                        targets.add(record.split(b'\t', 2)[2].decode())
        records.append(dict(name=dep['name'], module=dep['module'], version=dep['version'],
                            stages=len(dep['stages']), patches=len(identities),
                            affected_paths=len(targets),
                            patch_set_sha256=hashlib.sha256(json.dumps(identities, separators=(',', ':')).encode()).hexdigest()))
    return dict(schema=1, manifest_sha256=hashlib.sha256(manifest.read_bytes()).hexdigest(), dependencies=records)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    result = inventory(ROOT / 'drivers/cogent')
    if args.check:
        expected = json.loads((ROOT / 'release/native-patch-inventory.json').read_text())
        if result != expected:
            raise SystemExit('Native patch inventory changed. Review ownership/upgrade impact and regenerate release/native-patch-inventory.json.')
        print('Native patch inventory matches reviewed maintenance record.')
    else:
        print(json.dumps(result, indent=2))


if __name__ == '__main__':
    main()
