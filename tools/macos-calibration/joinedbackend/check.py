"""Join existing complete CPU and backend validators on the same actual drawable."""
from pathlib import Path
import importlib.util, math
CODE=Path(__file__).resolve().parent

def module(name,path):
 s=importlib.util.spec_from_file_location(name,path);m=importlib.util.module_from_spec(s);s.loader.exec_module(m);return m
joined=module('joined_backend_cpu',CODE.parent/'joined/check.py')
backend=module('joined_backend_gpu',CODE.parent/'backend/check.py')
base=joined.base
POST=('acquire_to_first_submit','first_submit_to_first_gpu','render_gpu_span','last_render_gpu_to_shown')

def validate(doc,match):
 cpu=joined.validate(doc,match);native=backend.validate(doc,match,'on');need=base.need
 need(doc['joined_backend_mode']=='same-input-cpu-submit-gpu-shown-v1','joined backend marker')
 trace=doc['backend_trace']['records'];presents=[r for r in trace if r['kind']==2]
 need(trace[-1]['kind']==2,'unassigned trailing native submission')
 rows=[];used=set()
 for source in cpu['joined_rows']:
  row=dict(source);hits=[p for p in presents if p['drawable_id']==row['drawable_id']]
  need(len(hits)==1,'one backend drawable')
  present=hits[0];index=present['ordinal'];previous=max((p['ordinal'] for p in presents if p['ordinal']<index),default=-1)
  render=[r for r in trace[previous+1:index] if r['kind']==0]
  need(render and index not in used,'unique render/present attribution');used.add(index)
  clock=row['raw_host_clocks'];first=render[0]['commit_start'];gpu_start=min(r['gpu_start'] for r in render);gpu_end=max(r['gpu_end'] for r in render)
  need(clock['acquired']<=first<=gpu_start<=gpu_end<=clock['shown'],'joined backend GPU partition order')
  need(all(clock['acquired']<=r['commit_start']<=r['commit_end']<=present['present_request'] for r in render),'render commits inside drawable interval')
  need(clock['acquired']<=present['present_request']<=present['commit_start']<=present['commit_end']<=clock['shown'],'present submission bounds')
  need(all(r['gpu_end']<=r['complete_observed'] for r in render),'render completion observation precedes GPU end')
  need(present['gpu_start']<=present['gpu_end'],'present GPU order')
  if present['gpu_start'] or present['gpu_end']:
   need(present['commit_start']<=present['gpu_start']<=present['gpu_end']<=present['complete_observed'] and present['gpu_end']<=clock['shown'],'present GPU bounds')
  boundaries=[clock['acquired'],first,gpu_start,gpu_end,clock['shown']]
  parts={k:(boundaries[i+1]-boundaries[i])*1e9 for i,k in enumerate(POST)}
  need(abs(sum(parts.values())-row['intervals_ns']['acquire_to_shown'])<=joined.EPS_NS,'post-acquire partition closure')
  row.update(backend_render_ordinals=[r['ordinal'] for r in render],backend_present_ordinal=index,backend_commands=[dict(r) for r in trace[previous+1:index+1]],post_acquire_ns=parts,present_request_minus_render_gpu_end_ns=(present['present_request']-gpu_end)*1e9,present_request_to_shown_ns=(clock['shown']-present['present_request'])*1e9)
  rows.append(row)
 need(len(used)==248,'all backend cohort joins')
 first_receipt=min(r['raw_host_clocks']['receipt'] for r in rows)
 startup=[p for p in presents if p['ordinal'] not in used]
 need(len(startup)<=128 and presents[:len(startup)]==startup,'noncohort submissions outside startup')
 attributed={i for row in rows for i in range(max((p['ordinal'] for p in presents if p['ordinal']<row['backend_present_ordinal']),default=-1)+1,row['backend_present_ordinal'])}
 startup_commands=[]
 for p in startup:
  previous=max((v['ordinal'] for v in presents if v['ordinal']<p['ordinal']),default=-1)
  commands=trace[previous+1:p['ordinal']]
  need(p['present_request']<first_receipt and (p['shown']==0 or p['shown']<first_receipt) and all(c['commit_end']<first_receipt for c in commands),'late unmatched startup submission')
  startup_commands.extend(c['ordinal'] for c in commands);attributed.update(c['ordinal'] for c in commands)
 need(attributed=={r['ordinal'] for r in trace if r['kind']!=2},'unassigned native commands')
 measured=rows[128:];p95=sorted(measured,key=lambda r:(r['receipt_to_shown_ns'],r['accepted_revision']))[math.ceil(len(measured)*.95)-1]
 return {'outcome':'VALID_JOINED_CPU_BACKEND_TIMING_DIAGNOSTIC','joined_rows':rows,'p95_rank_ticket':p95,'post_acquire_ms':{k:base.stats([r['post_acquire_ns'][k]/1e6 for r in measured]) for k in POST},'receipt_to_shown_ms':cpu['receipt_to_shown_ms'],'over_33_3_ms_observations':cpu['over_33_3_ms_observations'],'cpu_validation':cpu,'backend_validation':native,'retained_startup_backend_present_ordinals':[p['ordinal'] for p in startup],'retained_startup_backend_command_ordinals':startup_commands,'release_qualified':False,'gate_changed':False,'interpretation':'One private observer build, no calibrated overhead or ordinary-build equivalence. Same-input partitions, GPU span includes gaps; CPU submit/present may overlap GPU. Final-render-GPU-end-to-shown is not exclusively compositor/display cost. No physical input or displayed-byte proof.'}
