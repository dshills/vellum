#!/usr/bin/env python3
"""Validate serial/paced native feedback. Execution provenance remains external."""
import argparse,hashlib,json,math,re,sys,os,stat
from pathlib import Path
MAX_BYTES=64<<20
U64=(1<<64)-1

def need(ok,message):
 if not ok:raise ValueError(message)
def obj(v,keys,optional=()):
 need(type(v)is dict and set(keys)<=set(v)<=set(keys)|set(optional),'object fields differ: '+str(keys))
 return v
def integer(v,lo=0,hi=U64):
 need(type(v)is int and lo<=v<=hi,'invalid bounded integer');return v
def number(v,positive=False):
 need(type(v)in[int,float]and 0<=v<=1e12 and math.isfinite(v)and(v>0 if positive else v>=0),'invalid native clock');return v
def boolean(v):need(type(v)is bool,'invalid boolean');return v
def text(v,limit=4096):need(type(v)is str and len(v.encode('utf-8'))<=limit and '\0'not in v,'invalid text');return v
def array(v,limit):need(type(v)is list and len(v)<=limit,'invalid bounded array');return v
def pairs(items):
 out={}
 for key,value in items:
  need(key not in out,'duplicate JSON key');out[key]=value
 return out

def load(path,expected_hash):
 path=Path(path)
 flags=os.O_RDONLY|os.O_NONBLOCK|getattr(os,'O_NOFOLLOW',0)
 with os.fdopen(os.open(path,flags),'rb')as file:
  info=os.fstat(file.fileno());need(stat.S_ISREG(info.st_mode)and info.st_size<=MAX_BYTES,'regular bounded journal file required')
  raw=file.read(MAX_BYTES+1)
 need(len(raw)<=MAX_BYTES,'journal byte budget exceeded')
 need(re.fullmatch('[0-9a-f]{64}',expected_hash)is not None and hashlib.sha256(raw).hexdigest()==expected_hash,'journal hash differs')
 return json.loads(raw.decode('utf-8'),object_pairs_hook=pairs,parse_constant=lambda _:(_ for _ in()).throw(ValueError('nonfinite JSON constant')))

FRAME_KEYS=['Tag','Window','Generation','QueuedNS','InputNS','OwnerNS','ReconcileNS','LayoutPaintNS','GatherNS','RasterUploadNS','EncodeSubmitNS','AcquireNS','PresentCallNS','TotalUIWorkNS','Completed','Failure']
SAMPLE_KEYS=['Sequence','Scope','Tag','Generation','Composition','Armed','DrawableID','Acquired','Presented','State']
STATES=['pending','presented','skipped','unavailable','invalid','retired']

def point(value,closed):
 obj(value,['Frames','Presentation']);frames=value['Frames'];snapshot=value['Presentation']
 obj(frames,['Samples','Dropped','Pending','Failure']);integer(frames['Dropped']);integer(frames['Pending']);text(frames['Failure'])
 need(frames['Dropped']==frames['Pending']==0 and frames['Failure']=='','CPU recorder not terminal')
 cpu={}
 for frame in array(frames['Samples'],65536):
  obj(frame,FRAME_KEYS,['Unavailable','SourceDetails']);tag=integer(frame['Tag'],1);integer(frame['Generation'],1)
  need(tag not in cpu and frame['Window']=='perf'and boolean(frame['Completed'])and frame['Failure']=='','refused/duplicate CPU frame')
  need(frame.get('Unavailable',False)is False and frame.get('SourceDetails')is None,'other/refused diagnostic mode')
  for name in FRAME_KEYS:
   if name.endswith('NS'):integer(frame[name],hi=(1<<63)-1)
  need(frame['TotalUIWorkNS']==sum(frame[name]for name in ['InputNS','OwnerNS','ReconcileNS','LayoutPaintNS','GatherNS','RasterUploadNS','EncodeSubmitNS']),'CPU work accounting differs')
  cpu[tag]=frame
 obj(snapshot,['Scopes','Inputs','Samples','Pending','Skipped','Unavailable','Invalid','Retired','Overflow','Failed','Failure'])
 need(boolean(snapshot['Failed'])is False and boolean(snapshot['Overflow'])is False and text(snapshot['Failure'])=='','native recorder failed')
 scopes=array(snapshot['Scopes'],64);need(len(scopes)==1,'one native window incarnation required')
 scope=obj(scopes[0],['ID','Window','Closed']);need(integer(scope['ID'],1)==1 and scope['Window']=='perf'and boolean(scope['Closed'])is closed,'scope/window retirement differs')
 inputs=array(snapshot['Inputs'],65536);samples=array(snapshot['Samples'],65536)
 need(len(inputs)+len(samples)<=65536,'combined record budget exceeded')
 receipts={};last_receipt=0
 for receipt in inputs:
  obj(receipt,['Scope','Tag','Received']);need(integer(receipt['Scope'],1)==1,'foreign input scope')
  tag=integer(receipt['Tag'],1);need(tag not in receipts,'duplicate input tag');at=number(receipt['Received'],True);need(at>last_receipt,'receipt clock/order regressed');last_receipt=at;receipts[tag]=receipt
 counts={state:0 for state in STATES};last_acquired=0;last_composition=0;stamps={}
 for index,sample in enumerate(samples,1):
  obj(sample,SAMPLE_KEYS);need(integer(sample['Sequence'],1)==index and integer(sample['Scope'],1)==1,'acquisition identity differs')
  tag=integer(sample['Tag']);generation=integer(sample['Generation']);composition=integer(sample['Composition'],1);integer(sample['DrawableID']);armed=boolean(sample['Armed'])
  need(composition>=last_composition,'composition ordinal regressed');last_composition=composition
  stamp=(tag,generation,armed);need(composition not in stamps or stamps[composition]==stamp,'composition identity changed');stamps[composition]=stamp
  acquired=number(sample['Acquired'],True);presented=number(sample['Presented']);need(acquired>=last_acquired,'acquisition clock regressed');last_acquired=acquired
  need(sample['State']in STATES,'unknown sample state');counts[sample['State']]+=1
  if armed:need(tag>0 and generation>0,'armed metadata missing')
  if tag==0:need(generation==0 and not armed,'untagged sample has owner identity')
  else:need(tag in receipts and receipts[tag]['Received']<=acquired,'sample has no earlier receipt')
  if sample['State']=='presented':need(presented>=acquired and presented>0,'presentation clock regressed')
  else:need(presented==0,'unpresented sample claims display time')
 for key,state in [('Pending','pending'),('Skipped','skipped'),('Unavailable','unavailable'),('Invalid','invalid'),('Retired','retired')]:need(integer(snapshot[key],hi=65536)==counts[state],'native state count differs')
 if closed:need(snapshot['Pending']==0,'close left pending handlers')
 else:need(snapshot['Retired']==0,'live scope has retired records')
 return cpu,receipts,samples,counts

def percentile(values,p):
 need(bool(values),'no percentile observations');return sorted(values)[math.ceil(len(values)*p)-1]
def stats(values):return {'count':len(values),'p50':percentile(values,.5),'p95':percentile(values,.95),'p99':percentile(values,.99),'max':max(values)}

def pace_view(value,tree):
 keys=['ScrollX','ScrollY','Width','Height','Bounds','Scale','Visible','Rows','Columns','Retained','Materialized','LayoutSHA256']
 obj(value,keys)
 for key in ['ScrollX','ScrollY']:integer(value[key],hi=10_000_000)
 for key in ['Width','Height']:integer(value[key],1,65536)
 bounds=array(value['Bounds'],4);need(len(bounds)==4,'paced bounds count')
 for v in bounds:integer(v,-1_000_000,1_000_000)
 need(bounds[0]<bounds[2]and bounds[1]<bounds[3],'paced bounds empty')
 need(number(value['Scale'],True)<=16,'paced scale differs')
 cap=500 if tree else 4096
 for key in ['Visible','Rows','Columns','Retained','Materialized']:integer(value[key],hi=cap)
 need(value['Visible']<=value['Materialized'],'paced visible/materialized count')
 if tree:need(value['Visible']>=64 and value['Rows']==value['Visible']and value['Columns']==1 and value['ScrollX']==0,'paced tree workload differs')
 else:need(value['Visible']>=400 and value['Rows']>=50 and value['Columns']>=8,'paced grid workload differs')
 need(type(value['LayoutSHA256'])is str and re.fullmatch('[0-9a-f]{64}',value['LayoutSHA256'])is not None,'paced layout hash differs')
 return value

def pace_observation(value,tree):
 obj(value,['Tag','Generation','Frame','View'])
 for key in ['Tag','Generation','Frame']:integer(value[key],1)
 pace_view(value['View'],tree);return value

def pace_delta(index,tree):
 if tree:return(0,12 if index%32<16 else-12)
 phase=index%64;dy=12 if phase<32 else-12
 dx=96 if phase in[7,15,23,31]else-96 if phase in[39,47,55,63]else 0
 return dx,dy

def native_rounding_ns(value):
 return math.ceil(math.ulp(value)*1e9+math.ulp(value*1e9))+2

def validate_pacing(value,config,cpu,receipts,armed,ordered,measured):
 obj(value,['Hz','Clock','Method','PlannedEndNS','Started','Completed','HoldStartNS','HoldEndNS','Failure','Initial','Slots'])
 count=config['Samples'];tree=config['Method']=='routed-tree-scroll'
 need(count>=600 and config['Paced']is True and not config['Novel'],'paced sample/config bound')
 need(integer(value['Hz'],60,60)==60 and value['Clock']=='Go monotonic relative nanoseconds; separate from native host clock'and value['Method']=='absolute rational deadlines; serial CPU acknowledgement; retained backlog','paced clock/method differs')
 need(boolean(value['Started'])and boolean(value['Completed'])and text(value['Failure'])=='','paced cohort did not complete')
 offset=lambda index:index*1_000_000_000//60
 need(integer(value['PlannedEndNS'])==offset(count),'paced planned end changed')
 slots=array(value['Slots'],2048);need(len(slots)==count,'paced slots omitted')
 tags=[tag for tag in ordered if tag in measured]
 initial=pace_observation(value['Initial'],tree)
 initial_tag=config['Warmup']+2 if tree else(1<<32)+config['Warmup']
 expected_initial_y=(config['DataSize']*3//4+64)*9 if tree and config['Deep']else 0
 need(initial['Tag']==initial_tag and initial['Generation']==cpu[initial_tag]['Generation']and initial['View']['ScrollX']==0 and initial['View']['ScrollY']==expected_initial_y,'paced initial/warmup baseline differs')
 previous=initial;previous_end=0;lower_origin=None;upper_origin=None;last_precision=0
 for index,slot in enumerate(slots):
  obj(slot,['Index','Tag','ScheduledNS','State','StartNS','DispatchNS','EnqueuedNS','AcknowledgedNS','InspectedNS','EndNS','Observation'])
  need(integer(slot['Index'])==index and integer(slot['Tag'],1)==tags[index]and integer(slot['ScheduledNS'])==offset(index)and slot['State']=='completed','paced planned slot identity differs')
  times=[integer(slot[key],hi=(1<<63)-1)for key in ['StartNS','DispatchNS','EnqueuedNS','AcknowledgedNS','InspectedNS','EndNS']]
  need(times==sorted(times)and times[0]>=offset(index)and times[0]>=previous_end,'paced timeline regressed/early/concurrent');previous_end=times[-1]
  observed=pace_observation(slot['Observation'],tree);view=observed['View'];before=previous['View'];dx,dy=pace_delta(index,tree)
  need(observed['Tag']==slot['Tag']and observed['Generation']==cpu[slot['Tag']]['Generation']==previous['Generation']+1 and observed['Frame']>previous['Frame'],'paced CPU/generation/frame identity differs')
  need(view['ScrollX']==before['ScrollX']+dx and view['ScrollY']==before['ScrollY']+dy,'paced viewport clamped/no-op')
  need(all(view[key]==before[key]for key in ['Width','Height','Bounds','Scale']),'paced geometry changed')
  # The native receipt occurs inside the measured Go dispatch/return bracket.
  # Unknown clock origins are aligned by an intersection, never by arbitrary
  # subtraction. Retain explicit rounding uncertainty for native float seconds.
  at=receipts[slot['Tag']]['Received'];native_ns=round(at*1e9);precision=native_rounding_ns(at);last_precision=max(last_precision,precision)
  lo=native_ns-slot['EnqueuedNS']-precision;hi=native_ns-slot['DispatchNS']+precision
  lower_origin=lo if lower_origin is None else max(lower_origin,lo)
  upper_origin=hi if upper_origin is None else min(upper_origin,hi)
  previous=observed
 need(lower_origin<=upper_origin,'paced Go/native receipt brackets inconsistent')
 hold_start=integer(value['HoldStartNS'],hi=(1<<63)-1);hold_end=integer(value['HoldEndNS'],hi=(1<<63)-1)
 need(hold_start>=previous_end and hold_end>=max(hold_start,value['PlannedEndNS']),'paced immutable end/hold differs')
 presentation_times=[armed[tag]['Presented']for tag in tags]
 intervals=[(right-left)*1000 for left,right in zip(presentation_times,presentation_times[1:])]
 anomalies=[{'left_tag':tags[index],'right_tag':tags[index+1],'milliseconds':gap}for index,gap in enumerate(intervals)if gap<=0]
 limit=1100/60;long=[{'left_tag':tags[index],'right_tag':tags[index+1],'milliseconds':gap}for index,gap in enumerate(intervals)if gap>limit]
 stalls=[row for row in long if row['milliseconds']>2000/60]
 passing=sum(0<gap<=limit for gap in intervals);denominator=count-1
 lateness=[slot['StartNS']-slot['ScheduledNS']for slot in slots]
 misses=[{'index':index,'overrun_ns':slot['EndNS']-offset(index+1)}for index,slot in enumerate(slots)if slot['EndNS']>offset(index+1)]
 debt=[max(0,slot['EndNS']-offset(index+1))for index,slot in enumerate(slots)]
 worst_consecutive=run=0
 for v in debt:
  run=run+1 if v>0 else 0;worst_consecutive=max(worst_consecutive,run)
 first_native_ns=round(presentation_times[0]*1e9);last_native_ns=round(presentation_times[-1]*1e9)
 first_precision=native_rounding_ns(presentation_times[0]);last_endpoint_precision=native_rounding_ns(presentation_times[-1])
 return {'result':'VALID_PACED_NATIVE_FEEDBACK_JOURNAL','paced_metrics':{'planned_states':count,'ordered_distinct_final_times':not anomalies,'distinct_final_presentation_times':len(set(presentation_times)),'interval_population':'one final armed state per scheduled tag, in request order; no warmup anchor; N-1 intervals','interval_count':denominator,'intervals_ms':intervals,'interval_stats_ms':stats(intervals),'within_1_1_period':passing,'within_fraction':passing/denominator,'over_1_1_period':long,'over_two_periods':stalls,'collapsed_or_reordered_states':anomalies,'cadence_threshold_met':not anomalies and passing*100>=99*denominator,'planned_duration_ns':value['PlannedEndNS'],'actual_hold_end_ns':hold_end,'final_work_overrun_ns':max(0,slots[-1]['EndNS']-value['PlannedEndNS']),'start_lateness_ns':stats(lateness),'iteration_ns':stats([s['EndNS']-s['StartNS']for s in slots]),'dispatch_to_enqueue_return_ns':stats([s['EnqueuedNS']-s['DispatchNS']for s in slots]),'enqueue_return_to_ack_observation_ns':stats([s['AcknowledgedNS']-s['EnqueuedNS']for s in slots]),'ack_to_inspection_ns':stats([s['InspectedNS']-s['AcknowledgedNS']for s in slots]),'iteration_deadline_misses':misses,'worst_schedule_debt_ns':max(debt),'worst_consecutive_deadline_misses':worst_consecutive,'clock_correlation_model':'conditional constant origin offset and common nanosecond rate over cohort; receipt call brackets, not independent clock calibration','native_origin_bracket_ns':[lower_origin,upper_origin],'clock_origin_uncertainty_ns':upper_origin-lower_origin,'native_receipt_rounding_bound_ns':last_precision,'native_presentation_rounding_bound_ns':max(first_precision,last_endpoint_precision),'first_scheduled_to_presented_ms_bounds':[(first_native_ns-first_precision-upper_origin)/1e6,(first_native_ns+first_precision-lower_origin)/1e6],'final_presentation_after_planned_end_ms_bounds':[(last_native_ns-last_endpoint_precision-upper_origin-value['PlannedEndNS'])/1e6,(last_native_ns+last_endpoint_precision-lower_origin-value['PlannedEndNS'])/1e6],'held_tail':'idle; no extra changing interval; callback drain/teardown separate','release_qualified':False}}

def validate(journal,source,binary_sha256,go_version='go1.25.6'):
 paced=type(journal)is dict and journal.get('Schema')=='vellum-presentation-capture-v2'
 obj(journal,['Schema','Result','Window','RunError','CutoffMethod','BinarySHA256','Config','Build','ContinuousWork','WorkerRetired','RuntimeReturned','AcquisitionCutoff','ReceiptCutoff','BeforeClose','AfterClose']+(['Pacing']if paced else[]),['PrefixObservations'])
 need(journal['Schema']in['vellum-presentation-capture-v1','vellum-presentation-capture-v2']and journal['Result']=='CAPTURED_DIAGNOSTIC'and journal['RunError']=='','failed or unsupported capture')
 need(journal['Window']=='perf'and boolean(journal['ContinuousWork'])is paced and boolean(journal['WorkerRetired'])is True and boolean(journal['RuntimeReturned'])is True,'method/worker scope differs')
 need(journal['CutoffMethod']=='acquisition-sequence prefix at completed CPU cohort','cutoff method differs')
 config=obj(journal['Config'],['Method','Samples','Warmup','DataSize','Variable','Deep','Novel','GPURequested']+(['Paced']if paced else[]))
 samples=integer(config['Samples'],120,2048);warmup=integer(config['Warmup'],32,2048);size=integer(config['DataSize'],1,1_000_000)
 for key in ['Variable','Deep','Novel','GPURequested']:boolean(config[key])
 if config['Method']=='routed-grid-scroll':
  need(size in[10000,1000000]and warmup%64==0 and not config['Deep'],'grid configuration differs')
  measured=set(range(1,samples+1));ordered=list(range((1<<32)+1,(1<<32)+warmup+1))+list(range(1,samples+1));expected=set(ordered);setup=set()
 elif config['Method']=='routed-tree-scroll':
  need(size in[10000,100000]and warmup%32==0 and not config['Variable']and not config['Novel'],'tree configuration differs')
  measured=set(range(warmup+3,warmup+samples+3));setup={1,2};ordered=list(range(1,warmup+samples+3));expected=set(ordered)
 else:raise ValueError('unsupported input method')
 build=journal['Build'];need(type(build)is dict and len(build)<=64,'invalid build metadata')
 for key,value in build.items():text(key,256);text(value)
 need(re.fullmatch('[0-9a-f]{40}',source)is not None,'invalid external source identity')
 need(re.fullmatch('[0-9a-f]{64}',binary_sha256)is not None and journal['BinarySHA256']==binary_sha256,'executable identity differs')
 vcs_present=any(key in build for key in ['vcs','vcs.revision','vcs.modified'])
 if vcs_present:need(build.get('vcs')=='git'and build.get('vcs.revision')==source and build.get('vcs.modified')=='false','clean source identity differs')
 need(build.get('GOOS')=='darwin'and build.get('GOARCH')=='arm64'and build.get('CGO_ENABLED')=='1'and 'offscreen'not in re.split(r'[,\s]+',build.get('-tags','')),'native platform/build mode differs')
 need(build.get('go_version')==go_version and build.get('-race')!='true'and '-N'not in build.get('-gcflags','')and '-l'not in build.get('-gcflags',''),'optimized compiler identity differs')
 cpu,receipts,before,before_counts=point(journal['BeforeClose'],False)
 after_cpu,after_receipts,after,after_counts=point(journal['AfterClose'],True)
 cutoff=integer(journal['AcquisitionCutoff'],1,65536);receipt_cutoff=integer(journal['ReceiptCutoff'],1,65536)
 need(cutoff<=len(before)<=len(after),'acquisition prefix was truncated')
 need(receipt_cutoff==len(receipts)and set(cpu)==set(receipts)==expected,'missing/extra warmup or measured input')
 need(list(receipts)==ordered,'warmup/setup/measured input order differs')
 need(cpu==after_cpu and receipts==after_receipts,'CPU/input cohort changed after seal')
 for old,new in zip(before,after):
  identity=[key for key in SAMPLE_KEYS if key not in['State','DrawableID','Presented']]
  need(all(old[key]==new[key]for key in identity),'acquisition metadata changed after seal')
  if old['State']!='pending':need(old==new,'terminal feedback changed after seal')
 prefix=before[:cutoff]
 first_receipt=next(iter(receipts.values()))['Received']
 startup=[s for s in prefix if s['State']=='skipped'and not s['Armed']and s['Tag']==s['Generation']==0 and s['Scope']==1 and s['Presented']==0 and s['Acquired']<first_receipt]
 presented_prefix=[s for s in prefix if s['State']=='presented']
 need(len(startup)+len(presented_prefix)==len(prefix),'lost/missing prefix feedback')
 if 'PrefixObservations'in journal:
  accounting=obj(journal['PrefixObservations'],['Records','Presented','StartupSkipped'])
  need(integer(accounting['Records'],hi=65536)==len(prefix)and integer(accounting['Presented'],hi=65536)==len(presented_prefix)and integer(accounting['StartupSkipped'],hi=65536)==len(startup),'prefix observation counts differ')
 else:need(not startup,'startup skips require explicit prefix accounting')
 need(all(sample['Tag']==0 for sample in after[cutoff:]),'tagged input escaped declared cohort')
 need(prefix==after[:cutoff],'sealed prefix changed after close')
 armed={}
 for sample in prefix:
  if sample['Armed']:
   tag=sample['Tag'];need(tag not in armed,'more than one final armed drawable')
   need(cpu[tag]['Generation']==sample['Generation'],'final owner generation differs');armed[tag]=sample
 need(set(armed)==expected,'missing armed actual presentation')
 latencies=[]
 for tag in sorted(expected):
  sample=armed[tag];latency=(sample['Presented']-receipts[tag]['Received'])*1000
  need(math.isfinite(latency)and latency>=0,'invalid latency')
  latencies.append({'tag':tag,'generation':cpu[tag]['Generation'],'sequence':sample['Sequence'],'role':'measured'if tag in measured else'setup'if tag in setup else'warmup','milliseconds':latency})
 times=sorted(sample['Presented']for sample in presented_prefix)
 intervals=[(right-left)*1000 for left,right in zip(times,times[1:])]
 result={'result':'VALID_SERIAL_NATIVE_FEEDBACK_JOURNAL','source':source,'source_binding':'matching clean VCS stamp plus external build/launch binding'if vcs_present else'external build/launch binding required; VCS stamp absent','binary_sha256':binary_sha256,'method':config['Method'],'config':config,'latencies':latencies,'measured_input_latency_ms':stats([s['milliseconds']for s in latencies if s['role']=='measured']),'prefix_records':len(prefix),'prefix_presentations':len(presented_prefix),'startup_skipped':len(startup),'prefix_intervals_ms':intervals,'equal_presentation_times':sum(v==0 for v in intervals),'tail_records':len(after)-cutoff,'before_state_counts':before_counts,'after_state_counts':after_counts,'cadence_qualified':False,'physical_latency_qualified':False,'actual_gpu_route':'not established by this journal','provenance':'JSON validation requires external binary/native-library/workload execution binding; synthetic data is not native evidence'}

 if paced:
  result.update(validate_pacing(journal['Pacing'],config,cpu,receipts,armed,ordered,measured))
  result['paced_metrics']['input_latency_threshold_met']=result['measured_input_latency_ms']['p95']<=33.3
 return result

def main():
 parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('journal',type=Path);parser.add_argument('--sha256',required=True);parser.add_argument('--source',required=True);parser.add_argument('--binary-sha256',required=True);parser.add_argument('--go-version',default='go1.25.6');args=parser.parse_args()
 try:result=validate(load(args.journal,args.sha256),args.source,args.binary_sha256,args.go_version)
 except(ValueError,OSError,UnicodeError,RecursionError,OverflowError)as error:print(json.dumps({'result':'REFUSED_JOURNAL','error':str(error)}));return 1
 print(json.dumps(result,allow_nan=False));return 0
if __name__=='__main__':sys.exit(main())
