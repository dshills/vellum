#!/usr/bin/env python3
"""Private framework cache control; never grants a shipping latency pass."""
from pathlib import Path
import importlib.util
CODE = Path(__file__).resolve().parent
s = importlib.util.spec_from_file_location('cache_control_upload', CODE.parent/'upload/check.py')
upload = importlib.util.module_from_spec(s)
s.loader.exec_module(upload)
base = upload.base

def policy(doc, mode):
    base.need(mode in ('writecombined', 'default') and doc['cache_mode'] == mode, 'cache mode')
    row = doc['cache_policy']
    base.need(type(row) is list and len(row) == 6 and all(type(x) is int and 0 <= x <= 16384 for x in row), 'policy scalar bounds')
    configured, eligible, wc, default, mismatch, unconfigured = row
    base.need(configured == (1 if mode == 'writecombined' else 2)
              and mismatch == unconfigured == 0 and 0 < eligible == wc+default
              and (default == 0 if mode == 'writecombined' else wc == 0), 'observed cache policy')
    # Every nonempty texture write uses this exact staging descriptor. Extra
    # eligible write-buffer staging is counted, not silently treated as texture.
    base.need(eligible >= doc['native_upload_trace']['count'], 'missing eligible texture staging')
    return row

def validate(doc, match, mode):
    observed = policy(doc, mode)
    result = upload.validate(doc, match, 'on')
    return {**result, 'cache_mode': mode, 'cache_policy': observed,
            'interpretation': 'Same private binary with source/GPU-upload timing observers enabled in both modes. Exact transient staging cache mode changes; ownership and completion unchanged. Extra write-buffer staging participates. No ordinary binary equivalence, observer neutrality, pixel correctness or shipping latency acceptance.'}
