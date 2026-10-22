"""Existing full-byte/lifetime protocol plus exact texture-only cache accounting."""
from pathlib import Path
import importlib.util
s=importlib.util.spec_from_file_location('cached_staging_parent',Path(__file__).resolve().parent.parent/'stagingreuse/check.py');parent=importlib.util.module_from_spec(s);s.loader.exec_module(parent)
base=parent.base

def policy(doc,ledger):
    row=doc['native_upload_trace']['cache_policy']
    base.need(type(row) is list and len(row)==6 and all(type(x) is int and 0<=x<=4096 for x in row),'cache scalar bounds')
    configured,eligible,wc,default,mismatch,unconfigured=row
    base.need(configured==2 and eligible==default==ledger['fresh']>0 and wc==mismatch==unconfigured==0,'exact texture-only default cache allocations')
    return row

def validate(doc,mode):
    result=parent.validate(doc,mode)
    return {**result,'cache_policy':policy(doc,result['lifecycle']),
            'interpretation':'Full original byte/lifetime checks plus actual Metal cache/shared-storage counts bijective with eligible fresh allocations. Excluded texture/write-buffer/mapped staging descriptors preserve original policy. No UI latency, whole allocation/fault/hardware or shipping qualification.'}
