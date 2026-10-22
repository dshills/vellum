"""Strict phase/pending-state joins over existing native/framework validators."""
from pathlib import Path
import bisect,importlib.util,math
CODE=Path(__file__).resolve().parent
s=importlib.util.spec_from_file_location('phase_display_parent',CODE.parent/'displayclock/check.py');parent=importlib.util.module_from_spec(s);s.loader.exec_module(parent);base=parent.base
need=base.need

def clock(doc):
 t=doc['display_clock_trace'];need(all(t[k] is True for k in ['configured','enabled','started','stopped']) and all(t[k] is False for k in ['invalid','overflow']),'display clock terminal')
 ticks=t['records'];need(type(t['count']) is int and t['count']==len(ticks) and 2<=len(ticks)<=8192,'tick capacity')
 need(type(t['window']) is int and t['window']>0 and type(t['display']) is int and t['display']>0,'clock identity')
 need(doc['begin']['clock']<=t['start']<t['stop']<=doc['end']['clock'],'clock lifetime')
 for i,row in enumerate(ticks):
  need(type(row['ordinal']) is int and row['ordinal']==i and all(base.finite(row[k]) for k in ['timestamp','target','duration','observed']),'tick scalar')
  need(t['start']<=row['observed']<=t['stop'] and 0<row['timestamp']<row['target'] and abs(row['duration']-1/60)<.0001,'tick timing')
  if i:need(ticks[i-1]['observed']<=row['observed'] and ticks[i-1]['timestamp']<row['timestamp'] and abs(row['timestamp']-ticks[i-1]['timestamp']-1/60)<.0001,'tick gap/order')
 return ticks

def replay(trace):
 need(trace['configured'] is True and trace['invalid'] is False and trace['overflow'] is False and trace['pending']==trace['acquiring']==0,'pending terminal')
 attempts=trace['attempts'];events=trace['events'];need(len(attempts)<=512 and len(events)<=4096,'pending capacity')
 for i,a in enumerate(attempts):
  need(all(type(a[k]) is int and a[k]>=0 for k in ['attempt','scope','tag','generation','drawable','state']) and a['attempt']==i+1 and a['state'] in [3,4],'attempt identity/terminal')
 states={};seen={};receipts=[];observed=[];snapshots=[]
 def census():return [i for i in sorted(states) if states[i]==2]
 for i,e in enumerate(events):
  need(type(e['ordinal']) is int and e['ordinal']==i and type(e['kind']) is int and e['kind'] in range(1,6) and type(e['request']) is int and 0<=e['request']<=248,'event identity')
  need(type(e['attempt']) is int and type(e['tag']) is int and e['tag']>=0 and all(base.finite(e[k]) for k in ['at','observed','shown']) and 0<e['at']<=e['observed'] and e['shown']>=0,'event clock/scalar')
  if i:need(events[i-1]['observed']<=e['observed'],'event observation order')
  need(all(type(e[k]) is list and all(type(v) is int and v>0 for v in e[k]) for k in ['pending_before','pending_after']),'pending scalar identities')
  need(e['pending_before']==census() and type(e['acquiring_before']) is int and e['acquiring_before']==sum(v==1 for v in states.values()),'pending before replay')
  a=e['attempt'];kind=e['kind']
  if kind==1:
   need(a==0 and e['request']==len(receipts)+1,'one ordered receipt per request');receipts.append(e)
  elif kind==2:
   need(a==len(states)+1 and a<=len(attempts),'new attempt identity');states[a]=1;seen[a]={2:e}
  else:
   need(a in states and a in seen and kind not in seen[a],'unique attempt transition');need(e['tag']==attempts[a-1]['tag'],'attempt tag')
   if kind==3:need(states[a]==1,'acquired transition');states[a]=2
   elif kind==4:need(states[a]==1,'unavailable transition');states[a]=4
   else:need(states[a]==2 and (not e['shown'] or e['shown']<=e['observed']),'callback transition');states[a]=3
   seen[a][kind]=e
  need(e['pending_after']==census() and len(e['pending_after'])<=8 and type(e['acquiring_after']) is int and e['acquiring_after']==sum(v==1 for v in states.values()),'pending after replay')
  observed.append(e['observed']);snapshots.append(census())
 need(len(states)==len(attempts) and all(states[a['attempt']]==a['state'] for a in attempts),'all attempt terminal states')
 return receipts,seen,observed,snapshots

def validate(doc,match,kind):
 need(kind in ['native','framework'] and doc['kind']==kind,'implementation')
 checked=parent.validate(doc,match,'on') if kind=='framework' else base.validate(doc,match)
 ticks=clock(doc);trace=doc['pending_trace'];receipts,seen,observed,snapshots=replay(trace);requests=trace['requests']
 need(len(requests)==len(receipts)==248,'complete phase requests/receipts')
 if kind=='framework':
  source=checked['parent_validation']['joined_rows'];need(len(doc['phase_bridges'])==248,'Go target bridge count')
  for i,b in enumerate(doc['phase_bridges']):need(b['sequence']==i+1 and b['target']==requests[i]['target'] and type(b['go_bracket_ns']) is int and 0<=b['go_bracket_ns']<=100_000 and base.finite(b['host']),'Go clock-read bridge')
 else:
  source=[]
  for index,r in enumerate(doc['records']):
   need(all(base.finite(r[k]) for k in ['present_request','commit_start','commit_end','scheduled_observed','complete_observed','gpu_start','gpu_end']) and r['status']==4,'native GPU terminal')
   need(r['acquired']<=r['present_request']<=r['commit_start']<=r['commit_end']<=r['complete_observed'] and r['commit_start']<=r['gpu_start']<=r['gpu_end']<=r['complete_observed'] and (r['shown']==0 or r['gpu_end']<=r['shown']),'native GPU timing')
   if r['phase']:
    source.append({'accepted_revision':r['accepted'],'opaque_tag':index,'generation':r['accepted'],'drawable_id':r['drawable_id'],'phase':'warmup' if r['phase']==1 else 'measured','raw_host_clocks':{k:r[k] for k in ['receipt','acquired','shown']},'native_gpu':r})
 rows=[];used=set();stamps=[r['timestamp'] for r in ticks]
 for i,(r,req,receipt) in enumerate(zip(source,requests,receipts)):
  need(all(type(req[k]) is int for k in ['sequence','phase','anchor']),'phase scalar identities')
  need(req['sequence']==i+1 and req['phase']==i%4 and type(req['anchor']) is int and req['anchor']==requests[0]['anchor']+8*i and 0<=req['anchor']<len(ticks),'phase sequence/anchor')
  anchor=ticks[req['anchor']];target=anchor['timestamp']+(3+(req['phase']+.5)/4)*anchor['duration'];need(req['target']==target and anchor['observed']<target,'phase target')
  c=r['raw_host_clocks'];error_ms=(c['receipt']-target)*1000
  need(0<=error_ms<=2 and receipt['at']==c['receipt'] and receipt['tag']==r['opaque_tag'],'actual receipt phase match/join')
  hits=[a for a,v in seen.items() if 3 in v and v[3]['at']==c['acquired'] and trace['attempts'][a-1]['drawable']==r['drawable_id']]
  need(len(hits)==1 and hits[0] not in used,'unique acquired drawable join');a=hits[0];used.add(a);attempt=trace['attempts'][a-1];events=seen[a]
  need(attempt['tag']==r['opaque_tag'] and attempt['generation']==r['generation'] and 5 in events and events[5]['shown']==c['shown'] and c['receipt']<=events[2]['at']<=c['acquired']<=c['shown'],'cohort lifecycle attribution')
  if kind=='framework':commands=r['backend_commands'];request=commands[-1]['present_request'];gpu=commands[-1]['gpu_end'];render_end=max(v['gpu_end'] for v in commands if v['kind']==0)
  else:request=r['native_gpu']['present_request'];gpu=r['native_gpu']['gpu_end'];render_end=gpu
  phases={}
  for label,at in [('receipt',c['receipt']),('present_request',request),('final_gpu_end',gpu)]:
   j=bisect.bisect_right(observed,at)-1;need(j>=0,'pending observation coverage');phases[label]={'basis':'last recorded transition observed at or before event','attempts':snapshots[j],'other_attempts':[v for v in snapshots[j] if v!=a]}
  lo=bisect.bisect_right(stamps,gpu);hi=bisect.bisect_right(stamps,c['shown']);need(0<lo<len(ticks) and 0<hi<len(ticks),'reported clock coverage')
  rows.append({'sequence':i+1,'phase':req['phase'],'cohort':r['phase'],'attempt':a,'tag':r['opaque_tag'],'drawable':r['drawable_id'],'target':target,'receipt':c['receipt'],'receipt_phase_error_ms':error_ms,'receipt_to_shown_ms':(c['shown']-c['receipt'])*1000,'receipt_to_gpu_end_ms':(gpu-c['receipt'])*1000,'gpu_end_to_shown_ms':(c['shown']-gpu)*1000,'render_gpu_end_to_shown_ms':(c['shown']-render_end)*1000,'pending_at_receipt_hook':receipt['pending_before'],'acquiring_at_receipt_hook':receipt['acquiring_before'],'receipt_hook_observed':receipt['observed'],'receipt_hook_delay_ms':(receipt['observed']-receipt['at'])*1000,'acquire_begin':events[2],'acquired':events[3],'presented':events[5],'pending_observations':phases,'reported_ticks_gpu_to_shown':hi-lo})
 first=rows[0]['receipt'];need(all(a in used or (v[2]['at']<first and (3 not in v or v[3]['at']<first)) for a,v in seen.items()),'unmatched acquisition after cohort start')
 measured=rows[128:];need(len(measured)==120,'measured count')
 return {'outcome':'VALID_PHASE_PENDING_DIAGNOSTIC','kind':kind,'rows':rows,'parent_validation':checked,'receipt_to_shown_ms':base.stats([r['receipt_to_shown_ms'] for r in measured]),'phase_buckets':{str(phase):{'count':30,'receipt_phase_error_ms':base.stats([r['receipt_phase_error_ms'] for r in measured if r['phase']==phase]),'receipt_to_shown_ms':base.stats([r['receipt_to_shown_ms'] for r in measured if r['phase']==phase])} for phase in range(4)},'p95_rank_ticket':sorted(measured,key=lambda r:(r['receipt_to_shown_ms'],r['sequence']))[math.ceil(len(measured)*.95)-1],'release_qualified':False,'gate_changed':False,'interpretation':'Pending means recorded acquisitions awaiting callback, not Metal pool availability, retained-object lifetime or compositor queue depth. Source observer overhead and distinct rendering algorithms uncalibrated; no pure framework cost, physical/latch/optical or ordinary qualification claim.'}

def compare(native,framework):
 need(native['kind']=='native' and framework['kind']=='framework','pair modes')
 for phase in range(4):
  a=native['phase_buckets'][str(phase)];b=framework['phase_buckets'][str(phase)]
  need(a['count']==b['count']==30 and abs(a['receipt_phase_error_ms']['p50']-b['receipt_phase_error_ms']['p50'])<=1 and abs(a['receipt_phase_error_ms']['p95']-b['receipt_phase_error_ms']['p95'])<=1,'actual per-phase receipt mismatch')
 return {'outcome':'VALID_PHASE_MATCHED_PAIR','difference_of_run_p95_framework_minus_native_ms':framework['receipt_to_shown_ms']['p95']-native['receipt_to_shown_ms']['p95'],'release_qualified':False,'gate_changed':False}
