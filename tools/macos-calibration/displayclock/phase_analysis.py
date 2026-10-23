#!/usr/bin/env python3
"""Offscreen analysis of immutable display-clock journals; never a release gate."""
from pathlib import Path
import argparse
import bisect
import hashlib
import importlib.util
import json
import math
import tarfile

PAIRS = (
    ('macos-display-scheduling', ('off', 'on'),
     'bdabf97d8833d2f235096522345b74cad6e34af100a68ccc494080aeac8f590f',
     '02155f923cbcf9ee3e53848a683f03a971b04406d87092b614c99dc8b4d70552'),
    ('macos-display-reversed', ('on', 'off'),
     '937bfee0d2de09e9a0782e6de1a256d3e34e0cbb067f9a724aded509fea84895',
     'cef6cc5e9bb242c96f2f85bb0f1d0a19b6c83bd4c51b4b3f064a4c1f397391d2'),
)
VALIDATORS = ('tools/macos-calibration/check.py',
              'tools/macos-calibration/sourcecost/check.py',
              'tools/macos-calibration/upload/check.py',
              'tools/macos-calibration/joined/check.py',
              'tools/macos-calibration/backend/check.py',
              'tools/macos-calibration/joinedbackend/check.py',
              'tools/macos-calibration/displayclock/check.py',
              'scripts/benchmark-grid.py')
IDENTITY = ('accepted_revision', 'scope', 'opaque_tag', 'generation', 'drawable_id')
PARTS = ('receipt_to_acquire', 'acquire_to_first_submit',
         'first_submit_to_first_gpu', 'render_gpu_span', 'render_gpu_end_to_shown')


def need(ok, message):
    if not ok:
        raise ValueError(message)


def finite(value):
    return type(value) in (int, float) and math.isfinite(value)


def sha(data):
    return hashlib.sha256(data).hexdigest()


def parse(data):
    def reject(value):
        raise ValueError('nonfinite JSON: ' + value)
    return json.loads(data, parse_constant=reject)


def stats(values):
    need(values and all(finite(v) for v in values), 'finite nonempty population')
    values = sorted(values)
    return {'count': len(values), **{
        name: values[math.ceil(len(values) * fraction) - 1]
        for name, fraction in [('p50', .5), ('p95', .95), ('p99', .99)]
    }, 'max': values[-1]}


def read_archive(directory, kind, expected, names):
    """Read exact members in memory, without extraction or executable loading."""
    archive = directory / (kind + '-evidence.tar.xz')
    need(archive.stat().st_size <= 64 << 20, 'archive byte bound')
    digest = sha(archive.read_bytes())
    index_bytes = (directory / (kind + '-evidence-index.json')).read_bytes()
    need(len(index_bytes) <= 4 << 20, 'index byte bound')
    index = parse(index_bytes)
    need(digest == expected == index['archive_sha256'], 'pinned archive identity')
    output = {}
    with tarfile.open(archive, 'r:xz') as tar:
        members = tar.getmembers()
        member_names = [m.name for m in members]
        need(len(members) <= 4096 and len(set(member_names)) == len(members),
             'archive member count/uniqueness')
        need(set(member_names) == set(index['files']), 'archive index coverage')
        for name in names:
            member = tar.getmember(name)
            need(member.isfile() and member.size <= 16 << 20, 'member type/byte bound')
            data = tar.extractfile(member).read()
            record = index['files'][name]
            need(len(data) == record['bytes'] and sha(data) == record['sha256'],
                 'indexed member identity')
            output[name] = data
    proof = {'archive': str(archive), 'archive_sha256': digest,
             'index_sha256': sha(index_bytes),
             'read_members': {n: {'bytes': len(b), 'sha256': sha(b)}
                              for n, b in output.items()}}
    return output, proof


class ClockIndex:
    """Independent right-inclusive preceding brackets for reports and delivery."""
    def __init__(self, ticks):
        need(len(ticks) >= 2, 'callback coverage')
        self.ticks = ticks
        for i, tick in enumerate(ticks):
            need(tick['ordinal'] == i and all(finite(tick[k]) for k in
                 ['timestamp', 'observed']), 'callback identity/clocks')
            if i:
                need(ticks[i-1]['timestamp'] < tick['timestamp'] and
                     ticks[i-1]['observed'] <= tick['observed'], 'callback order')
        self.reported = [t['timestamp'] for t in ticks]
        self.delivered = [t['observed'] for t in ticks]

    def join(self, event):
        need(finite(event), 'finite event')
        ri = bisect.bisect_right(self.reported, event)
        di = bisect.bisect_right(self.delivered, event)
        need(0 < ri < len(self.ticks) and 0 < di < len(self.ticks),
             'reported and delivered coverage around event')
        preceding = self.ticks[ri-1]
        nearest = min((ri-1, ri), key=lambda i: abs(self.reported[i]-event))
        return {
            'preceding_reported_ordinal': ri-1, 'next_reported_ordinal': ri,
            'nearest_reported_ordinal': nearest,
            'event_minus_nearest_reported_ms': (event-self.reported[nearest])*1000,
            'nearest_reported_delivery_minus_event_ms':
                (self.delivered[nearest]-event)*1000,
            'preceding_delivered_ordinal': di-1, 'next_delivered_ordinal': di,
            'event_minus_preceding_reported_ms': (event-self.reported[ri-1])*1000,
            'next_reported_minus_event_ms': (self.reported[ri]-event)*1000,
            'event_minus_preceding_delivered_ms': (event-self.delivered[di-1])*1000,
            'next_delivered_minus_event_ms': (self.delivered[di]-event)*1000,
            'preceding_reported_delivery_minus_event_ms':
                (preceding['observed']-event)*1000,
            'preceding_reported_already_delivered': preceding['observed'] <= event,
        }


def timeline(row):
    clocks = dict(row['raw_host_clocks'])
    commands = row['backend_commands']
    renders = [c for c in commands if c['kind'] == 0]
    need(renders and commands[-1]['kind'] == 2, 'render and final present required')
    present = commands[-1]
    clocks.update(first_submit=renders[0]['commit_start'],
                  first_gpu=min(c['gpu_start'] for c in renders),
                  render_gpu_end=max(c['gpu_end'] for c in renders),
                  present_request=present['present_request'],
                  present_commit=present['commit_start'],
                  present_gpu_end=present['gpu_end'])
    boundaries = [clocks[k] for k in ['receipt', 'acquired', 'first_submit',
                                     'first_gpu', 'render_gpu_end', 'shown']]
    need(all(finite(v) for v in clocks.values()) and
         all(a <= b for a, b in zip(boundaries, boundaries[1:])), 'partition order')
    parts = {k: (b-a)*1000 for k, a, b in zip(PARTS, boundaries, boundaries[1:])}
    total = row['receipt_to_shown_ns']/1e6
    need(abs(sum(parts.values())-total) <= .001, 'same-input partition closure')
    need(clocks['receipt'] <= clocks['present_gpu_end'] <= clocks['shown'],
         'present GPU split order')
    split = {'receipt_to_present_gpu_end':
             (clocks['present_gpu_end']-clocks['receipt'])*1000,
             'present_gpu_end_to_shown':
             (clocks['shown']-clocks['present_gpu_end'])*1000}
    need(abs(sum(split.values())-total) <= .001, 'present GPU split closure')
    return clocks, parts, split


def analyze_case(checked, doc, mode):
    parent = checked['parent_validation']
    rows = parent['joined_rows']
    need(len(rows) == 248 and [r['phase'] for r in rows] ==
         ['warmup']*128 + ['measured']*120, 'cohort identity')
    clock_index = ClockIndex(doc['display_clock_trace']['records']) if mode == 'on' else None
    derived = []
    for row in rows:
        clocks, partition, split = timeline(row)
        item = {k: row[k] for k in (*IDENTITY, 'phase')}
        item.update(receipt_to_shown_ms=row['receipt_to_shown_ns']/1e6,
                    partition_ms=partition, present_gpu_split_ms=split,
                    receipt_lateness_go_ms=row['receipt_lateness_go_ns']/1e6,
                    receipt_clock_bracket_go_ms=row['receipt_clock_bracket_go_ns']/1e6)
        if clock_index:
            item['clock_phases'] = {k: clock_index.join(v) for k, v in clocks.items()}
            gpu_phase = item['clock_phases']['present_gpu_end']
            shown_phase = item['clock_phases']['shown']
            item['reported_timestamps_gpu_to_shown'] = (
                shown_phase['preceding_reported_ordinal'] -
                gpu_phase['preceding_reported_ordinal'])
            item['callbacks_delivered_gpu_to_shown'] = (
                shown_phase['preceding_delivered_ordinal'] -
                gpu_phase['preceding_delivered_ordinal'])
        derived.append(item)
    measured = derived[128:]
    ticket = parent['p95_rank_ticket']
    ranked = sorted(measured, key=lambda v: (v['receipt_to_shown_ms'],
                                             v['accepted_revision']))[math.ceil(len(measured)*.95)-1]
    need(all(ranked[k] == ticket[k] for k in IDENTITY), 'actual p95 input identity')
    summary = {
        'mode': mode, 'joined_inputs': 248, 'warmup': 128, 'measured': 120,
        'receipt_to_shown_ms': stats([r['receipt_to_shown_ms'] for r in measured]),
        'partition_ms': {k: stats([r['partition_ms'][k] for r in measured]) for k in PARTS},
        'present_gpu_split_ms': {k: stats([r['present_gpu_split_ms'][k] for r in measured])
                                 for k in measured[0]['present_gpu_split_ms']},
        'receipt_lateness_go_ms': stats([r['receipt_lateness_go_ms'] for r in measured]),
        'receipt_clock_bracket_go_ms': stats([r['receipt_clock_bracket_go_ms'] for r in measured]),
        'p95_rank_ticket': ranked,
        'p95_ticket_post_present_gpu_fraction_pct':
            ranked['present_gpu_split_ms']['present_gpu_end_to_shown'] /
            ranked['receipt_to_shown_ms'] * 100,
        'display_clock_phase': None,
    }
    need(summary['receipt_to_shown_ms'] == parent['receipt_to_shown_ms'],
         'parent receipt distribution')
    if clock_index:
        summary['clock_event_counts'] = {k: {str(value): sum(r[k] == value for r in measured)
                for value in sorted({r[k] for r in measured})}
                for k in ['reported_timestamps_gpu_to_shown',
                          'callbacks_delivered_gpu_to_shown']}
        summary['display_clock_phase'] = {}
        for event in measured[0]['clock_phases']:
            values = [r['clock_phases'][event] for r in measured]
            summary['display_clock_phase'][event] = {
                'preceding_reported_callback_not_yet_delivered': sum(
                    not v['preceding_reported_already_delivered'] for v in values),
                **{k: stats([v[k] for v in values]) for k in values[0] if k.endswith('_ms')},
            }
    return summary, derived


def classifier_identity(plan, root):
    proof = {}
    for relative in VALIDATORS:
        hits = [digest for name, digest in plan['pins'].items()
                if name.endswith('/'+relative)]
        digest = sha((root/relative).read_bytes())
        need(hits == [digest], 'unchanged classifier: '+relative)
        proof[relative] = digest
    return proof


def analyze(reports, validator):
    pairs = []
    all_rows = {}
    inputs = []
    for name, order, result_sha, prep_sha in PAIRS:
        directory = reports/name
        raw_names = ['result/artifacts/result.json', 'result/artifacts/bound-plan.json']
        for mode in order:
            raw_names.extend('result/artifacts/'+mode+'/'+suffix for suffix in
                             ['native/stdout', 'validation.json'])
        data, proof = read_archive(directory, 'result', result_sha, raw_names)
        prep, prep_proof = read_archive(directory, 'preparation', prep_sha,
                                       ['preparation/artifacts/match.json'])
        plan = parse(data['result/artifacts/bound-plan.json'])
        proof['unchanged_classifiers'] = classifier_identity(
            plan, Path(__file__).resolve().parents[3])
        inputs.extend([proof, prep_proof])
        match = parse(prep['preparation/artifacts/match.json'])
        result = parse(data['result/artifacts/result.json'])
        need(result['outcome'] in ['VALID_DISPLAY_CLOCK_OFF_ON_CONTROL',
                                  'VALID_REVERSED_DISPLAY_CLOCK_CONTROL'] and
             [c['mode'] for c in result['cases']] == list(order), 'completed scope/order')
        cases = []
        for case in result['cases']:
            mode = case['mode']
            raw = data['result/artifacts/'+mode+'/native/stdout']
            need(len(raw) <= 4 << 20 and sha(raw) ==
                 case['native']['logs']['stdout']['sha256'], 'raw journal identity/bound')
            documents = [parse(line) for line in raw.splitlines() if line.strip()]
            need(len(documents) == 1, 'one terminal journal')
            checked = validator.validate(documents[0], match, mode)
            stored = parse(data['result/artifacts/'+mode+'/validation.json'])
            need(checked == stored == case['validation'], 'complete unchanged reclassification')
            summary, rows = analyze_case(checked, documents[0], mode)
            cases.append(summary)
            all_rows[name+'/'+mode] = rows
        pairs.append({'pair': name, 'order': list(order), 'cases': cases})
    return {'scope': 'Existing immutable journals only; offscreen, no native execution',
            'pairs': pairs, 'input_archives': inputs,
            'release_qualified': False, 'gate_changed': False,
            'production_change': False,
            'interpretation': 'Each run separate. On cases join independent reported and actual-delivery clocks; off cases have no display-clock inference. Same-input partitions close; their marginal quantiles must not be added. Go lateness is not converted to native target time. Reported events are not latch/scanout evidence; no calibrated overhead, causal or ordinary-equivalence claim.'}, all_rows


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--reports', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    code = Path(__file__).resolve().parent
    spec = importlib.util.spec_from_file_location('phase_existing_validator', code/'check.py')
    validator = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(validator)
    summary, rows = analyze(args.reports, validator)
    args.output.mkdir(parents=True, exist_ok=False)
    for name, value in [('summary.json', summary), ('rows.json', rows)]:
        (args.output/name).write_text(json.dumps(value, indent=2, allow_nan=False)+'\n')
    print(json.dumps({'outcome': 'VALID_OFFSCREEN_PHASE_ANALYSIS', 'cases': 4,
                      'joined_inputs': sum(map(len, rows.values())),
                      'measured_inputs': 480, 'native_executions': 0}))


if __name__ == '__main__':
    main()
