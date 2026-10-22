"""Existing framework accounting plus completed-staging ledger; diagnostic only."""
from pathlib import Path
import importlib.util
CODE=Path(__file__).resolve().parent
def module(name,path):
    s=importlib.util.spec_from_file_location(name,path);m=importlib.util.module_from_spec(s);s.loader.exec_module(m);return m
staging=module('timing_staging_bytes',CODE/'check.py')
upload=module('timing_existing_upload',CODE.parent/'upload/check.py')
base=upload.base

def validate(doc,match,mode):
    result=upload.validate(doc,match,'on')
    base.need(doc['reuse_mode']==mode,'reuse mode echo')
    ledger=staging.graph(doc['reuse_trace'],mode)
    cpu=doc['native_upload_trace']['records'];leases={v['call']:v for v in ledger['leases']}
    eligible={r['ordinal']+1:r for r in cpu if r['reuse_eligible'] is True}
    base.need(set(leases)==set(eligible),'all eligible calls accounted')
    for call,lease in leases.items():
        r=eligible[call];base.need(r['begin']<=lease['acquire_clock']<=r['end'] and lease['bytes']==r['width']*r['height']*4,'lease CPU span/size')
    receipts={(r['Scope'],r['Tag']):r['Received'] for r in doc['before_close']['Inputs']};samples=doc['before_close']['Samples'];hits=[]
    for index,m in enumerate(doc['measurements']['Samples']):
        if index<128:continue
        f=m['Frame'];key=(m['Scope'],f['Tag']);receipt=receipts[key]
        selected=[r for r in samples if r['Scope']==key[0] and r['Tag']==key[1] and r['Generation']==f['Generation'] and r['Armed']];base.need(len(selected)==1,'measured attribution')
        native=[r for r in cpu if receipt<=r['begin'] and r['end']<=selected[0]['Acquired']]
        base.need(native and all(r['reuse_eligible'] is True for r in native),'measured image eligibility')
        hits.append(sum(doc['reuse_trace']['records'][leases[r['ordinal']+1]['acquire_sequence']]['kind']==3 for r in native))
    base.need(len(hits)==120 and (sum(hits)>0 if mode=='on' else sum(hits)==0),'measured reuse exercised')
    return {**result,'reuse_mode':mode,'staging_lifecycle':ledger,'measured_reuse_hits':sum(hits),
            'presentation_diagnostic':base.validate(doc,match),
            'ordinary_latency_qualified':False,'interpretation':'Same private binary/library, original WC cache fixed, detailed source+CPU+lifecycle observers on. Full eligible ledger closed. Headless byte proof reused; no displayed-pixel, observer-neutrality, ordinary-build equivalence or release qualification.'}
