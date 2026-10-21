#!/usr/bin/env python3
"""Strict small-fixture calibration accounting; never grants release acceptance."""
import json,math,hashlib
FIELDS=('width','height','scale','pixel_format','drawables','vsync','transactions','timeout')
def need(ok,why):
 if not ok:raise ValueError(why)
def finite(x):return type(x) in (int,float) and math.isfinite(x)
def stats(xs):
 need(xs and all(finite(x) for x in xs),'finite population');v=sorted(xs)
 return {'count':len(v),**{k:v[max(0,math.ceil(len(v)*p)-1)] for k,p in [('p50',.5),('p95',.95),('p99',.99)]},'max':v[-1]}
def validate(doc,match,pilot=False):
 need(doc['kind'] in ['native','framework'] and doc['release_qualified'] is False,'scope')
 need(doc['match']==match,'match drift')
 need(match['scale']==1 and match['drawables']==2 and match['vsync'] is True and match['transactions'] is False and match['timeout'] is False and match['pixel_format'] in [80,81],'native policy')
 need(all(type(match[x]) is int or type(match[x]) is float for x in ['width','height']) and 0<match['width']<=4096 and 0<match['height']<=4096,'extent bound')
 rect=match['rect'];need(len(rect)==4 and all(type(x) is int for x in rect) and 0<=rect[0]<rect[2]<=match['width'] and 0<=rect[1]<rect[3]<=match['height'],'payload rectangle')
 begin,end=doc['begin'],doc['end']
 for state in [begin,end]:
  need(state['active'] is True and state['key'] is True and state['visible'] is True and state['losses']==0 and type(state['pid']) is int and state['pid']>0 and state['front_pid']==state['pid'],'foreground loss')
  need(all(state[k]==match[k] for k in FIELDS),'native endpoint drift')
  need(finite(state['clock']) and state['clock']>0,'endpoint clock')
 need(begin['pid']==end['pid'] and begin['clock']<end['clock'],'endpoint identity/order')
 if doc['kind']=='framework':
  need(doc['error']=='<nil>' and doc['runtime_returned'] is True,'framework return')
  before,after=doc['before_close'],doc['after_close'];need(not before['Failed'] and not before['Overflow'] and before['Pending']==0 and not after['Failed'] and not after['Overflow'] and after['Pending']==0,'recorder failure')
  need(len(before['Scopes'])==len(after['Scopes'])==1 and after['Scopes'][0]['Closed'] is True,'scope retirement')
  samples=before['Samples'];need(after['Samples'][:len(samples)]==samples,'prefix rewritten')
  measurements=doc['measurements'];need(not measurements['Pending'],'pending measured request')
  cohort=measurements['Samples'] or [];need(len(cohort)==(0 if pilot else 248),'cohort count')
  receipts=before['Inputs'] or [];need(len(receipts)==len(cohort),'receipt count')
  keyed={(r['Scope'],r['Tag']):r for r in receipts};need(len(keyed)==len(receipts),'duplicate receipt')
  rows=[];selected=[]
  for i,m in enumerate(cohort,1):
   f=m['Frame'];need(f['Completed'] is True and not f['Failure'] and not f.get('Unavailable',False) and m['AcceptedRevision']==i and f['Generation']>m['ViewRevisionBefore'],'unaccepted model/frame')
   key=(m['Scope'],f['Tag']);need(key in keyed,'unmatched receipt');receipt=keyed[key]['Received']
   hits=[r for r in samples if r['Scope']==key[0] and r['Tag']==key[1] and r['Generation']==f['Generation'] and r['Armed']];need(len(hits)==1,'armed attribution count')
   r=hits[0];need(r['State']=='presented','lost armed presentation');selected.append(r['Sequence']);rows.append({'sequence':i,'phase':1 if i<=128 else 2,'accepted':i,'receipt':receipt,'acquired':r['Acquired'],'shown':r['Presented']})
  need(len(set(selected))==len(selected),'reused acquisition')
  first=rows[0]['receipt'] if rows else end['clock']
  need(all(r['Sequence'] in selected for r in samples if r['Armed'] and r['Acquired']>=first),'unmatched armed composition')
  setup=[r for r in samples if r['Acquired']<first and (r['Scope'],r['Tag']) not in keyed]
  for r in samples:
   need(r['State']=='presented' and r['Presented']>0 or r in setup and r['Tag']==0 and not r['Armed'] and r['Generation']==0 and r['State']=='skipped' and r['Presented']==0,'non-startup feedback loss')
  ready=[r['Presented'] for r in setup if r['State']=='presented' and r['Presented']>0]
  setup_losses=sum(r['State']=='skipped' for r in setup)
 else:
  need(doc['runtime_returned'] is True,'native return');allrows=doc['records'];setup=[r for r in allrows if r['phase']==0];rows=[r for r in allrows if r['phase']!=0];need(allrows==setup+rows,'phase ordering')
  need(1<=len(setup)<=64 and all(r['sequence']==r['accepted']==0 and all(finite(r[k]) for k in ['receipt','acquired','shown']) and begin['clock']<=r['receipt']<=r['acquired']<=end['clock'] and (r['shown']==0 or r['acquired']<=r['shown']<=end['clock']) for r in setup),'native setup accounting')
  ready=[r['shown'] for r in setup if r['shown']>0];setup_losses=sum(r['shown']==0 for r in setup)
 need(ready,'positive startup readiness')
 if pilot:return {'outcome':'VALID_SETUP_ONLY','setup_losses':setup_losses,'release_qualified':False}
 need(len(rows)==248 and len(doc['lateness_ns'])==248 and all(finite(x) and 0<=x<=20_000_000 for x in doc['lateness_ns']),'request schedule')
 need(min(ready)<=rows[0]['receipt'],'cohort before readiness')
 for i,r in enumerate(rows,1):
  need(r['sequence']==r['accepted']==i and r['phase']==(1 if i<=128 else 2),'state identity')
  need(all(finite(r[k]) for k in ['receipt','acquired','shown']) and begin['clock']<=r['receipt']<=r['acquired']<r['shown']<=end['clock'],'clock/order')
  if i>1:need(rows[i-2]['shown']<r['receipt'],'overlapping/replaced request')
 need(all(r['shown']-r['receipt']<.120 for r in rows),'response deadline')
 measured=rows[128:];return {'outcome':'VALID_MATCHED_SMALL_FIXTURE_CALIBRATION','warmup':128,'measured':120,'setup_losses':setup_losses,'receipt_to_present_ms':stats([(r['shown']-r['receipt'])*1000 for r in measured]),'receipt_to_acquire_ms':stats([(r['acquired']-r['receipt'])*1000 for r in measured]),'acquire_to_present_ms':stats([(r['shown']-r['acquired'])*1000 for r in measured]),'release_qualified':False,'physical_input_verified':False}
def load(path):
 need(path.stat().st_size<=4<<20,'journal byte bound');return [json.loads(line,parse_constant=lambda v:(_ for _ in ()).throw(ValueError(v))) for line in path.read_text().splitlines() if line.strip()]
def native_doc(lines):
 need(len(lines)==2 and lines[1]=={'runtime_returned':True},'native terminal record');return {**lines[0],**lines[1]}
