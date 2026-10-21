#!/usr/bin/env python3
"""Existing nested source-cost validation on the matched fixture; no gate pass."""
from pathlib import Path
import importlib.util,sys
CODE=Path(__file__).resolve().parent
sys.path.insert(0,str(CODE.parents[2]/'scripts'))
def module(n,p):
 s=importlib.util.spec_from_file_location(n,p);m=importlib.util.module_from_spec(s);s.loader.exec_module(m);return m
base=module('source_parent',CODE.parent/'check.py');cost=module('existing_source_cost_validator',CODE.parents[2]/'scripts/benchmark-grid.py')
def validate(doc,match):
 baseline=base.validate(doc,match);base.need(doc.get('diagnostic_mode')=='source-cost-only','source diagnostic marker')
 allframes=doc['measurements']['Samples'];values=[]
 for m in allframes:
  f=m['Frame'];d=f.get('SourceDetails');cost.validate_source_details(d,f['RasterUploadNS'])
  base.need(any(r['Kind']==1 and r['RasterAvailable'] and r['CopyAvailable'] for r in d['Records']),'CPU scene source required')
  base.need(all(not r['SceneHost']['Available'] for r in d['Records']),'experimental scene route forbidden')
 for m in allframes[128:]:
  f=m['Frame'];rows=f['SourceDetails']['Records'];values.append({k:sum(r[k] for r in rows) for k in ['DrawNS','RasterNS','CopyNS','SetFromGoImageNS','SetFromGoImageAttempts','ImageCacheHits','LogicalUploadRGBABytes']})
 totals={k:sum(r[k] for r in values) for k in values[0]};base.need(totals['DrawNS']>0,'positive source work')
 return {'outcome':'VALID_CPU_SOURCE_DIAGNOSTIC','warmup':128,'measured':120,'release_qualified':False,'physical_input_verified':False,'source_phase_ms':{k:base.stats([r[k]/1e6 for r in values]) for k in ['DrawNS','RasterNS','CopyNS','SetFromGoImageNS']},'totals':totals,'nested_shares_of_source_draw':{k:totals[k]/totals['DrawNS'] for k in ['RasterNS','CopyNS','SetFromGoImageNS']},'presentation_feedback_accounting_valid':True,'latency_gate_qualified':False,'diagnostic_only':'Additional source clocks/counters and private compiler overlays; these timings are not ordinary latency qualification or successful GPU transfer evidence.'}
