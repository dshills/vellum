"""Existing segmented reuse timing plus exact per-texture Metal cache selection."""
from pathlib import Path
import importlib.util
s=importlib.util.spec_from_file_location('texture_cache_timing_parent',Path(__file__).resolve().parent.parent/'stagingreuse/timing_segment_check.py');parent=importlib.util.module_from_spec(s);s.loader.exec_module(parent)
base=parent.base

def validate(doc,match,mode):
    result=parent.validate(doc,match,'on')
    base.need(mode in ('writecombined','default') and doc['cache_mode']==mode,'cache mode echo')
    row=doc['native_upload_trace']['cache_policy']
    base.need(type(row) is list and len(row)==6 and all(type(x) is int and 0<=x<=4096 for x in row),'cache scalar bounds')
    configured,eligible,wc,default,mismatch,unconfigured=row
    fresh=result['staging_lifecycle']['fresh']
    base.need(configured==(1 if mode=='writecombined' else 2) and eligible==fresh>0 and mismatch==unconfigured==0 and
              ((wc==fresh and default==0) if mode=='writecombined' else (default==fresh and wc==0)),'exact texture cache/fresh ledger')
    return {**result,'cache_mode':mode,'cache_policy':row,'interpretation':'One ordered private texture-cache comparison, reuse fixed on. Actual marked cache/storage counts equal fresh resource IDs. Existing closed lifetime, source/CPU/presentation/segments checks unchanged; headless byte proofs separate. No ordinary equivalence, noise bound, GPU pixels, policy adoption or release qualification.'}
