#!/usr/bin/env python3
"""Validate a serial native feedback journal. Does not qualify 60 Hz cadence."""
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

def validate(journal,source,binary_sha256,go_version='go1.25.6'):
 obj(journal,['Schema','Result','Window','RunError','CutoffMethod','BinarySHA256','Config','Build','ContinuousWork','WorkerRetired','RuntimeReturned','AcquisitionCutoff','ReceiptCutoff','BeforeClose','AfterClose'])
 need(journal['Schema']=='vellum-presentation-capture-v1'and journal['Result']=='CAPTURED_DIAGNOSTIC'and journal['RunError']=='','failed or unsupported capture')
 need(journal['Window']=='perf'and boolean(journal['ContinuousWork'])is False and boolean(journal['WorkerRetired'])is True and boolean(journal['RuntimeReturned'])is True,'method/worker scope differs')
 need(journal['CutoffMethod']=='acquisition-sequence prefix at completed CPU cohort','cutoff method differs')
 config=obj(journal['Config'],['Method','Samples','Warmup','DataSize','Variable','Deep','Novel','GPURequested'])
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
 prefix=before[:cutoff];need(all(s['State']=='presented'for s in prefix),'lost/missing prefix feedback')
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
 times=sorted(sample['Presented']for sample in prefix)
 intervals=[(right-left)*1000 for left,right in zip(times,times[1:])]
 return {'result':'VALID_SERIAL_NATIVE_FEEDBACK_JOURNAL','source':source,'source_binding':'matching clean VCS stamp plus external build/launch binding'if vcs_present else'external build/launch binding required; VCS stamp absent','binary_sha256':binary_sha256,'method':config['Method'],'config':config,'latencies':latencies,'measured_input_latency_ms':stats([s['milliseconds']for s in latencies if s['role']=='measured']),'prefix_presentations':len(prefix),'prefix_intervals_ms':intervals,'equal_presentation_times':sum(v==0 for v in intervals),'tail_records':len(after)-cutoff,'before_state_counts':before_counts,'after_state_counts':after_counts,'cadence_qualified':False,'physical_latency_qualified':False,'actual_gpu_route':'not established by this journal','provenance':'JSON validation requires external binary/native-library/workload execution binding; synthetic data is not native evidence'}

def main():
 parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('journal',type=Path);parser.add_argument('--sha256',required=True);parser.add_argument('--source',required=True);parser.add_argument('--binary-sha256',required=True);parser.add_argument('--go-version',default='go1.25.6');args=parser.parse_args()
 try:result=validate(load(args.journal,args.sha256),args.source,args.binary_sha256,args.go_version)
 except(ValueError,OSError,UnicodeError,RecursionError,OverflowError)as error:print(json.dumps({'result':'REFUSED_JOURNAL','error':str(error)}));return 1
 print(json.dumps(result,allow_nan=False));return 0
if __name__=='__main__':sys.exit(main())
