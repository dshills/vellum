#!/usr/bin/env python3
"""Native submission topology control; preserves the parent strict classifier."""
from pathlib import Path
import importlib.util
s=importlib.util.spec_from_file_location('base_calibration',Path(__file__).resolve().parents[1]/'check.py')
base=importlib.util.module_from_spec(s);s.loader.exec_module(base)
CLOCKS=('render_commit_start','render_commit_end','present_request','present_commit_start','present_commit_end','render_scheduled_observed','render_complete_observed','gpu_start','gpu_end','present_scheduled_observed','present_complete_observed')
def validate(doc,match,mode):
 base.need(doc.get('kind')=='native' and mode in ('combined','split') and doc.get('mode')==mode,'submission mode')
 result=base.validate(doc,match)
 for r in doc['records']:
  t=r['timing'];base.need(all(base.finite(t[k]) and t[k]>=0 for k in CLOCKS),'finite timing')
  base.need(t['render_status']==t['present_status']==4,'command buffer completion')
  base.need(r['acquired']<=t['present_request']<=t['present_commit_start']<=t['present_commit_end']<=doc['end']['clock'],'present CPU timing')
  base.need(r['acquired']<=t['render_commit_start']<=t['render_commit_end']<=doc['end']['clock'],'render CPU timing')
  for role in ('render','present'):
   base.need(t[role+'_commit_start']<=t[role+'_scheduled_observed']<=doc['end']['clock'] and t[role+'_commit_start']<=t[role+'_complete_observed']<=doc['end']['clock'],'callback observation bounds')
  if mode=='combined':
   base.need(t['present_request']<=t['render_commit_start'] and all(t['render_'+k]==t['present_'+k] for k in ('commit_start','commit_end','scheduled_observed','complete_observed','status')),'combined timing identity')
  else:base.need(t['render_commit_end']<=t['present_request'],'split submission order')
  if r['phase']!=0:
   base.need(t['render_commit_start']<=t['gpu_start']<=t['gpu_end']<=r['shown'] and t['gpu_start']>0,'actual render GPU timing')
 measured=[r for r in doc['records'] if r['phase']==2]
 phases={
  'receipt_to_render_commit':lambda r:r['timing']['render_commit_start']-r['receipt'],
  'render_commit_to_gpu_start':lambda r:r['timing']['gpu_start']-r['timing']['render_commit_start'],
  'gpu_duration':lambda r:r['timing']['gpu_end']-r['timing']['gpu_start'],
  'gpu_end_to_present':lambda r:r['shown']-r['timing']['gpu_end'],
  'present_request_to_present':lambda r:r['shown']-r['timing']['present_request']}
 return {**result,'mode':mode,'phase_ms':{k:base.stats([f(r)*1000 for r in measured]) for k,f in phases.items()},'interpretation':'CPU callback observation clocks are not GPU clocks; quantiles are not additive; no release qualification'}
