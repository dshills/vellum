#!/usr/bin/env python3
"""Validate committed multiline adapter evidence without rerunning native UI."""
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
EVIDENCE = Path(__file__).resolve().parent / 'evidence/20261010-editor-native'


def require(condition, message):
    if not condition:
        raise SystemExit(message)


def read(name):
    return json.loads((EVIDENCE / name).read_text())


def verify_geometry(queries, count):
    require(len(queries) == count, 'geometry query coverage')
    for query in queries:
        actual = query['native']['screen_rect']
        expected = query['expected_screen_rect']
        require(len(actual) == len(expected) == 4, 'invalid geometry rectangle')
        require(all(abs(a-b) <= 0.05 for a, b in zip(actual, expected)), 'native geometry mismatch')
        require(query['native']['rect_range'][0] >= 0, 'missing native range')


def verify():
    manifest = read('manifest.json')
    for name, expected in manifest['files_sha256'].items():
        require(hashlib.sha256((ROOT / name).read_bytes()).hexdigest() == expected,
                'hash mismatch: ' + name)
    rendered = read('rendered-canonical.json')
    queries = rendered['rendered_geometry_queries']
    require(rendered['native_client_id'] == 40, 'editor geometry client')
    verify_geometry(queries, 63)
    for name, count in [('editor-input-final.json', 13), ('editor-input-canonical.json', 10)]:
        report = read(name)
        updates = report['input_updates']
        require(report['native_input_ack'] == count and len(updates) == count, 'editor acknowledgement')
        require(report['rejected_actions'] == 0 and report['native_client_id'] == 40, 'editor routing')
        require([u['Sequence'] for u in updates] == list(range(1, count+1)), 'input sequence')
        require(all(u['ClientID'] == 40 for u in updates), 'origin identity')
        require(all(a['Text'] == b['BeforeText'] for a, b in zip(updates, updates[1:])), 'shadow chain')
        queue = report['transport']['Queue']
        require(queue['Closed'] and queue['Outstanding'] == queue['PayloadBytes'] == queue['InFlight'] == 0, 'queue drain')
    saturation = read('saturation-final.json')
    require(0 < saturation['fill_admitted'] <= 256 and
            saturation['transport']['Queue']['Outstanding'] == 256, 'saturation capacity')
    require(len(saturation['refused_states']) == 5, 'missing refusal state comparisons')
    before = {k: v for k, v in saturation['before'].items() if k != 'step'}
    require(all({k: v for k, v in state.items() if k != 'step'} == before
                for state in saturation['refused_states']), 'refusal changed native state')
    lifetime = read('lifecycle-final.json')
    require(lifetime['cycles'] == lifetime['attachments'] == lifetime['detaches'] == 1000, 'native lifetime')
    require(lifetime['accepted'] == lifetime['applied_after_close'] == 11000, 'close drain')
    contracts = lifetime['publication_contracts']['native_clients']
    for key, value in contracts.items():
        if isinstance(value, bool):
            require(value, 'native client contract: ' + key)
    verify_geometry(read('field-regression.json')['rendered_geometry_queries'], 31)
    for name in ['focused-final.txt', 'focused-canonical.txt', 'race.txt', 'core-regressions.txt']:
        output = (EVIDENCE / name).read_text()
        require('ok ' in output or 'ok\t' in output, 'no passed tests: ' + name)
        require('FAIL' not in output, 'test failure: ' + name)
    require('FAIL' in (EVIDENCE / 'editor-packages.txt').read_text(), 'broader failure evidence missing')
    for name in manifest['final_prism_reports']:
        review = json.loads((ROOT / name).read_text())
        require(review['coverage']['complete'] and not review['findings'], 'incomplete/actionable Prism review: ' + name)
    print('PASS: hashes, native geometry/input/refusal/lifetime, focused tests and final Prism coverage; G1 remains open')


if __name__ == '__main__':
    verify()
