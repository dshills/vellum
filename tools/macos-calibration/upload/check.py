#!/usr/bin/env python3
"""Private CPU upload intervals, not GPU execution or an ordinary latency pass."""
from pathlib import Path
import importlib.util
CODE=Path(__file__).resolve().parent
s=importlib.util.spec_from_file_location('upload_source_cost',CODE.parent/'sourcecost/check.py');source=importlib.util.module_from_spec(s);s.loader.exec_module(source);base=source.base
PHASES=('create_ns','map_ns','copy_ns','flush_ns','encode_ns')

def validate(doc,match,mode):
 result=source.validate(doc,match);need=base.need
 need(mode in ('off','on') and doc['observer_mode']==mode,'observer mode')
 trace=doc['native_upload_trace'];need(trace['enabled'] is (mode=='on') and trace['pending']==0 and trace['overflow'] is False and trace['invalid_stage'] is False,'trace terminal/overflow/duplicate stage')
 rows=trace['records'];need(type(trace['count']) is int and trace['count']==len(rows)<=4096,'bounded trace count')
 if mode=='off':need(not rows,'off observer recorded');return {**result,'observer_mode':mode}
 for i,r in enumerate(rows):
  need(r['ordinal']==i and r['complete'] is True and r['status']==1,'CPU function outcome/incomplete slot')
  need(all(base.finite(r[k]) for k in ('begin','end',*PHASES)) and 0<r['begin']<r['end'],'CPU clocks')
  need(all(type(r[k]) is int and r[k]>=0 for k in ('bytes','width','height','stage_pitch','stages')),'integer payload')
  need(all(r[k]>=0 for k in PHASES) and sum(r[k] for k in PHASES)<=(r['end']-r['begin'])*1e9+1,'disjoint CPU stage bounds')
  need(r['stages']==(31 if r['width'] and r['height'] else 0),'recorded stage availability')
 receipts={(r['Scope'],r['Tag']):r['Received'] for r in doc['before_close']['Inputs']};samples=doc['before_close']['Samples'];used=set();measured=[]
 for index,m in enumerate(doc['measurements']['Samples']):
  key=(m['Scope'],m['Frame']['Tag']);receipt=receipts[key];f=m['Frame'];hits=[r for r in samples if r['Scope']==key[0] and r['Tag']==key[1] and r['Generation']==f['Generation'] and r['Armed']];need(len(hits)==1,'frame attribution')
  acquired=hits[0]['Acquired'];details=f['SourceDetails']['Records'];attempts=sum(r['SetFromGoImageAttempts'] for r in details);logical=sum(r['LogicalUploadRGBABytes'] for r in details);api=sum(r['SetFromGoImageNS'] for r in details)
  uploads=[r for r in rows if receipt<=r['begin'] and r['end']<=acquired]
  need(attempts>0 and len(uploads)==attempts and sum(r['bytes'] for r in uploads)==logical,'source API/native upload accounting')
  need(all(r['width']==match['width'] and r['height']==match['height'] and r['stage_pitch']==r['width']*4 and r['bytes']==4*r['width']*r['height'] for r in uploads),'fixture RGBA input/pitch')
  need(all(r['ordinal'] not in used for r in uploads),'reused CPU upload');used.update(r['ordinal'] for r in uploads)
  total=sum((r['end']-r['begin'])*1e9 for r in uploads);need(total<=api+1000,'native intervals exceed enclosing Go image API')
  if index>=128:
   v={k:sum(r[k] for r in uploads) for k in PHASES};v.update(native_total_ns=total,other_native_ns=total-sum(v.values()),Go_API_outside_native_ns=api-total,uploads=len(uploads));measured.append(v)
 first=min(receipts.values());prefix=[r for r in rows if r['ordinal'] not in used];need(len(prefix)<=128 and all(r['end']<=first for r in prefix),'unmatched upload after cohort admission')
 need(len(measured)==120,'measured join count')
 return {**result,'observer_mode':mode,'native_trace_records':len(rows),'retained_startup_uploads':len(prefix),'joined_cohort_uploads':len(used),'CPU_phase_ms':{k:base.stats([v[k]/1e6 for v in measured]) for k in measured[0] if k!='uploads'},'uploads_per_measured_frame':base.stats([v['uploads'] for v in measured]),'interpretation':'CPU function success is not GPU upload success or completion. Signed residuals retain clock rounding. No phase quantile addition, observer-neutrality or shipping-equivalence proof.'}
