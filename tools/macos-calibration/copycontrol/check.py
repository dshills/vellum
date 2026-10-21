#!/usr/bin/env python3
"""Offscreen copy control only; never classifies UI responsiveness."""
from pathlib import Path
import importlib.util
s = importlib.util.spec_from_file_location('copycontrol_base', Path(__file__).resolve().parent.parent/'check.py')
base = importlib.util.module_from_spec(s)
s.loader.exec_module(base)

def validate(doc, mode):
    need = base.need
    need(mode in ('writecombined', 'default') and doc['mode'] == mode, 'cache mode')
    need(doc['kind'] == 'offscreen-metal-copy-control' and doc['release_qualified'] is False
         and doc['runtime_returned'] is True and not doc.get('error'), 'terminal/scope')
    need(doc['comparator_negative_control'] is True and doc['unified_memory'] is True
         and type(doc['device']) is str and bool(doc['device']), 'device/comparator')
    need((doc['width'], doc['height'], doc['pitch'], doc['bytes'], doc['pixel_format'])
         == (1519, 1266, 6076, 7692216, 71), 'extent/format')
    need(doc['warmup'] == 128 and doc['measured'] == 120, 'cohort')
    need(base.finite(doc['begin']) and base.finite(doc['end']) and 0 < doc['begin'] < doc['end'], 'endpoint clocks')
    rows = doc['records']
    need(len(rows) == 248, 'row count')
    previous = doc['begin']
    for i, row in enumerate(rows, 1):
        need(type(row['sequence']) is int and row['sequence'] == i
             and type(row['phase']) is int and row['phase'] == (1 if i <= 128 else 2), 'identity/phase')
        need(type(row['storage_mode']) is int and row['storage_mode'] == 0
             and type(row['cpu_cache_mode']) is int and row['cpu_cache_mode'] == (1 if mode == 'writecombined' else 0), 'resource policy')
        need(base.finite(row['begin']) and base.finite(row['end'])
             and previous <= row['begin'] < row['end'] <= doc['end'], 'iteration clocks')
        previous = row['end']
        intervals = [row[k] for k in ('create_ns', 'map_ns', 'first_copy_ns', 'subsequent_copy_ns')]
        for key in ('first_gpu', 'subsequent_gpu'):
            gpu = row[key]
            need(gpu['completed'] is True and type(gpu['verified_bytes']) is int
                 and gpu['verified_bytes'] == doc['bytes'], 'GPU completion/full byte verification')
            intervals.extend(gpu[k] for k in ('submit_wait_ns', 'verify_ns'))
        need(all(base.finite(x) and x >= 0 for x in intervals)
             and sum(intervals) <= (row['end']-row['begin'])*1e9+1, 'disjoint CPU intervals')
    selected = rows[128:]
    return {'outcome': 'VALID_OFFSCREEN_COPY_CONTROL', 'mode': mode, 'warmup': 128, 'measured': 120,
            'GPU_verified_transfers': 496, 'GPU_verified_bytes': doc['bytes']*496,
            'CPU_ms': {k: base.stats([r[k]/1e6 for r in selected])
                       for k in ('create_ns', 'map_ns', 'first_copy_ns', 'subsequent_copy_ns')},
            'paired_subsequent_minus_first_ms': base.stats([(r['subsequent_copy_ns']-r['first_copy_ns'])/1e6 for r in selected]),
            'release_qualified': False, 'interpretation': 'Standalone serial Metal copies with complete GPU readback. No shipping equivalence, asynchronous pooling safety or UI latency acceptance.'}
