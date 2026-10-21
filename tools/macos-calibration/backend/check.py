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
  if r['kind']==2:need(0<r['present_request']<=r['commit_start'] and (r['shown']==0 or r['commit_start']<=r['shown']),'present request order')
  else:need(r['present_request']==r['shown']==0,'nonpresent role')
  if r['kind']==0:need(r['commit_start']<=r['gpu_start']<=r['gpu_end'] and r['gpu_start']>0,'render execution timing')
 samples=doc['before_close']['Samples'];present=[r for r in rows if r['kind']==2];positive=[r for r in present if r['shown']>0];shown=[r for r in samples if r['State']=='presented' and r['Presented']>0]
 need(len(present)==len(samples) and sum(r['shown']==0 for r in present)==sum(r['State']=='skipped' for r in samples),'all presentation feedback accounting')
 joins={}
 for f in shown:
  hits=[r for r in positive if abs(r['shown']-f['Presented'])<=EPS];need(len(hits)==1,'ambiguous/missing backend presentation');need(hits[0]['ordinal'] not in joins,'reused backend presentation');joins[hits[0]['ordinal']]=f
 need(len(joins)==len(positive),'unmatched backend presentation')
 receipts={(r['Scope'],r['Tag']):r['Received'] for r in doc['before_close']['Inputs']};phases=[]
 for p in present:
  f=joins.get(p['ordinal'])
  if f is None or not f['Armed']:continue
  previous=max((r['ordinal'] for r in present if r['ordinal']<p['ordinal']),default=-1)
  render=[r for r in rows[previous+1:p['ordinal']] if r['kind']==0];need(render,'no attributed render')
  receipt=receipts[(f['Scope'],f['Tag'])];first=render[0];gpu_start=min(r['gpu_start'] for r in render);gpu_end=max(r['gpu_end'] for r in render)
  need(receipt<=first['commit_start'] and gpu_end<=f['Presented'],'cohort GPU boundary')
  if f['Tag']>128:phases.append({'receipt_to_submit':first['commit_start']-receipt,'submit_to_gpu_start':gpu_start-first['commit_start'],'render_gpu_span':gpu_end-gpu_start,'gpu_end_to_present_request':p['present_request']-gpu_end,'gpu_end_to_shown':f['Presented']-gpu_end,'render_buffers':len(render)})
 need(len(phases)==120,'measured trace join count')
 return {**result,'observer_mode':mode,'trace_records':len(rows),'positive_feedback_joins':len(joins),'phase_ms':{k:base.stats([p[k]*1000 for p in phases]) for k in phases[0] if k!='render_buffers'},'render_buffers_per_measured_frame':base.stats([p['render_buffers'] for p in phases]),'release_qualified':False}
