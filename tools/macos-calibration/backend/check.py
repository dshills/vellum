#!/usr/bin/env python3
"""Strict correlation of private backend timings to public framework feedback."""
from pathlib import Path
import importlib.util
s=importlib.util.spec_from_file_location('backend_parent',Path(__file__).resolve().parents[1]/'check.py');base=importlib.util.module_from_spec(s);s.loader.exec_module(base)
CLOCKS=('commit_start','commit_end','present_request','scheduled_observed','complete_observed','gpu_start','gpu_end','shown')
EPS=1e-6

def validate(doc,match,mode):
 result=base.validate(doc,match);need=base.need
 need(doc['kind']=='framework' and mode in ('on','off') and doc['observer_mode']==mode,'observer mode')
 trace=doc['backend_trace'];need(trace['enabled'] is (mode=='on') and trace['pending']==0 and trace['overflow'] is False,'trace drain/overflow')
 rows=trace['records'];need(type(trace['count']) is int and trace['count']==len(rows)<=4096,'bounded trace count')
 if mode=='off':
  need(not rows,'disabled observer recorded');return {**result,'observer_mode':mode}
 need(rows and len({r['queue'] for r in rows})==1 and type(rows[0]['queue']) is int and rows[0]['queue']>0,'single queue identity')
 for i,r in enumerate(rows):
  need(r['ordinal']==i and type(r['kind']) is int and r['kind'] in (0,1,2) and r['status']==4,'trace identity/completion')
  need(all(base.finite(r[k]) and r[k]>=0 for k in CLOCKS),'finite trace clock')
  need(0<r['commit_start']<=r['commit_end'] and r['commit_start']<=r['scheduled_observed'] and r['commit_start']<=r['complete_observed'],'commit/callback bounds')
  if i:need(rows[i-1]['commit_end']<=r['commit_start'],'reordered commit reservation')
  need(r['drawable_recorded'] is (r['kind']==2),'explicit drawable recording')
  if r['kind']==2:need(0<r['present_request']<=r['commit_start'] and (r['shown']==0 or r['commit_start']<=r['shown']),'present request order')
  else:need(r['present_request']==r['shown']==0,'nonpresent role')
  if r['kind']==0:need(r['commit_start']<=r['gpu_start']<=r['gpu_end'] and r['gpu_start']>0,'render execution timing')
 samples=doc['after_close']['Samples'];present=[r for r in rows if r['kind']==2]
 need(present and samples,'presentation population')
 ids=[r['drawable_id'] for r in present];public_ids=[f['DrawableID'] for f in samples]
 need(all(type(i) is int and i>=0 for i in ids+public_ids) and ids==sorted(set(ids)) and public_ids==sorted(set(public_ids)),'drawable identities')
 public={f['DrawableID']:f for f in samples};first=samples[0];first_receipt=min(r['Received'] for r in doc['before_close']['Inputs'])
 prefix=[r for r in present if r['drawable_id']<first['DrawableID']]
 need(len(prefix)<=64 and all(r['present_request']<first['Acquired'] and (r['shown']==0 or r['shown']<first_receipt) for r in prefix),'bounded pre-admission startup prefix')
 need(present[:len(prefix)]==prefix and len(present)==len(samples)+len(prefix),'all presentation feedback accounting')
 need(present[len(prefix)]['drawable_id']==first['DrawableID'],'first admitted drawable')
 joins={}
 for r in present[len(prefix):]:
  need(r['drawable_id'] in public,'unknown admitted drawable');f=public[r['drawable_id']]
  need(abs(r['shown']-f['Presented'])<=EPS and (r['shown']>0 and f['State']=='presented' or r['shown']==0 and f['State']=='skipped'),'drawable feedback mismatch')
  joins[r['ordinal']]=f
 need(len(joins)==len(samples),'missing admitted drawable')
 measured_keys={(m['Scope'],m['Frame']['Tag'],m['Frame']['Generation']) for m in doc['measurements']['Samples'][128:]}
 receipts={(r['Scope'],r['Tag']):r['Received'] for r in doc['before_close']['Inputs']};phases=[]
 for p in present:
  f=joins.get(p['ordinal'])
  if f is None or not f['Armed']:continue
  previous=max((r['ordinal'] for r in present if r['ordinal']<p['ordinal']),default=-1)
  render=[r for r in rows[previous+1:p['ordinal']] if r['kind']==0];need(render,'no attributed render')
  receipt=receipts[(f['Scope'],f['Tag'])];first=render[0];gpu_start=min(r['gpu_start'] for r in render);gpu_end=max(r['gpu_end'] for r in render)
  need(receipt<=first['commit_start'] and gpu_end<=f['Presented'],'cohort GPU boundary')
  if (f['Scope'],f['Tag'],f['Generation']) in measured_keys:phases.append({'receipt_to_submit':first['commit_start']-receipt,'submit_to_gpu_start':gpu_start-first['commit_start'],'render_gpu_span':gpu_end-gpu_start,'gpu_end_to_present_request':p['present_request']-gpu_end,'gpu_end_to_shown':f['Presented']-gpu_end,'render_buffers':len(render)})
 need(len(phases)==120,'measured trace join count')
 return {**result,'observer_mode':mode,'trace_records':len(rows),'admitted_feedback_joins':len(joins),'retained_pre_admission_prefix':{'count':len(prefix),'zero_feedback':sum(r['shown']==0 for r in prefix),'positive_feedback':sum(r['shown']>0 for r in prefix),'drawable_ids':[r['drawable_id'] for r in prefix]},'phase_ms':{k:base.stats([p[k]*1000 for p in phases]) for k in phases[0] if k!='render_buffers'},'render_buffers_per_measured_frame':base.stats([p['render_buffers'] for p in phases]),'release_qualified':False}
