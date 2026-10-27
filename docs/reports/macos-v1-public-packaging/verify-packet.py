#!/usr/bin/env python3
"""Check this historical packet, optionally against the retained candidate files."""
import argparse
import hashlib
import json
import re
import sys
from pathlib import Path
import tarfile


def require(condition, message):
    if not condition:
        raise ValueError(message)


def digest(path):
    h = hashlib.sha256()
    with path.open('rb') as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b''):
            h.update(block)
    return h.hexdigest()


def checksums(path):
    values = {}
    for line in path.read_text().splitlines():
        require('  ' in line, 'malformed checksum line')
        value, name = line.split('  ', 1)
        require(len(value) == 64 and all(c in '0123456789abcdef' for c in value), 'invalid SHA256')
        require(name not in values and Path(name).name == name, 'duplicate or unsafe archive name')
        values[name] = value
    return values


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--packet', type=Path, default=Path(__file__).resolve().parent)
    parser.add_argument('--assets', type=Path, help='optional directory containing all seven retained archives')
    args = parser.parse_args()
    root = args.packet
    production = json.loads((root / 'package-result.json').read_text())
    native = json.loads((root / 'postlaunch-verification.json').read_text())
    all_hashes = checksums(root / 'SHA256SUMS')
    recommended = checksums(root / 'RECOMMENDED_SHA256SUMS')
    expected = {production['sdk']['archive']: production['sdk']['sha256']}
    readme = (root / 'README.md').read_text()
    table = {}
    app_names = {a['name'] for a in production['apps']}
    for line in readme.splitlines():
        columns = [part.strip() for part in line.strip().strip('|').split('|')]
        if len(columns) == 4 and columns[0] in app_names:
            require(columns[0] not in table, 'duplicate README app row')
            table[columns[0]] = tuple(columns[1:])
    expected_table = {}
    for app in production['apps']:
        require(app['archive'] not in expected, 'duplicate production archive')
        expected[app['archive']] = app['archive_sha256']
        outcome = 'Held: dark contrast' if app['id'] == 'shell' else 'Pass'
        expected_table[app['name']] = (str(app['notice_files']), outcome, '`' + app['archive_sha256'] + '`')
    require(len(expected) == 7 and expected == all_hashes, 'production/checksum mismatch')
    require(table == expected_table, 'README app row/digest mismatch')
    sdk_line = re.search(r'^Final SDK SHA256: `([0-9a-f]{64})`\.$', readme, re.MULTILINE)
    require(sdk_line is not None and sdk_line.group(1) == production['sdk']['sha256'], 'README SDK digest mismatch')
    require({a['file']: a['sha256'] for a in native['archives']} == expected, 'postlaunch/checksum mismatch')
    require(len(native['archives']) == 7, 'duplicate postlaunch archive')
    expected_recommended = {k: v for k, v in expected.items() if k != 'Vellum Mac Shell.zip'}
    require(len(recommended) == 6 and recommended == expected_recommended, 'recommended set includes held/missing asset')
    require(len(native['apps']) == 6 and len({a['id'] for a in native['apps']}) == 6, 'duplicate/missing native app')
    require([a['id'] for a in native['apps'] if not a['recommended_for_publication']] == ['shell'], 'held fixture disposition changed')
    require(all(a['visual_result'] == ('held-dark-contrast' if a['id'] == 'shell' else 'pass') for a in native['apps']), 'visual outcome mismatch')
    require(not production['uploaded'] and not production['publication_authorized'] and not native['uploaded'] and not native['publication_authorized'], 'publication status changed')
    index = json.loads((root / 'evidence-index.json').read_text())
    require(digest(root / 'evidence.tar.gz') == index['archive_sha256'], 'evidence archive mismatch')
    indexed = {m['path']: m for m in index['members']}
    require(len(indexed) == len(index['members']), 'duplicate evidence member')
    seen = set()
    with tarfile.open(root / 'evidence.tar.gz', 'r:gz') as archive:
        for member in archive:
            require(member.isfile() and member.name in indexed and member.name not in seen, 'unlisted/duplicate/nonregular evidence')
            require(not member.name.startswith('/') and '..' not in Path(member.name).parts, 'unsafe evidence path')
            record = indexed[member.name]
            require(member.size == record['bytes'] and 0 <= member.size <= 16 * 1024 * 1024, 'evidence member size mismatch/bound')
            stream = archive.extractfile(member)
            require(stream is not None, 'unreadable evidence member')
            require(hashlib.sha256(stream.read(16 * 1024 * 1024 + 1)).hexdigest() == record['sha256'], 'evidence member checksum mismatch')
            seen.add(member.name)
    require(seen == set(indexed), 'missing evidence member')
    for app in native['apps']:
        for capture in app['evidence']:
            record = indexed[capture['file']]
            require(capture['sha256'] == record['sha256'] and capture['bytes'] == record['bytes'], 'native evidence mismatch')
    if args.assets:
        for name, expected_hash in expected.items():
            require((args.assets / name).is_file(), 'missing asset archive: ' + name)
            require(digest(args.assets / name) == expected_hash, 'actual archive mismatch: ' + name)
    print(json.dumps({'passed': True, 'consistent_archive_hashes': len(expected), 'recommended_assets': len(recommended), 'held_apps': ['shell'], 'evidence_members': len(indexed), 'actual_asset_bytes_checked': args.assets is not None}))


if __name__ == '__main__':
    try:
        main()
    except (OSError, ValueError, KeyError, TypeError, tarfile.TarError) as error:
        print('Packet validation failed: ' + str(error), file=sys.stderr)
        sys.exit(1)
