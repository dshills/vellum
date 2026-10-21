#!/usr/bin/env python3
"""Explicitly authorized single-attempt one-shot source cost diagnostic; reuses existing owned execution."""
from pathlib import Path
import argparse,importlib.util,sys,json,datetime,signal
sys.dont_write_bytecode=True
parser=argparse.ArgumentParser();parser.add_argument('--artifacts',type=Path,required=True);parser.add_argument('--reuse',type=Path,required=True);parser.add_argument('--run-authorized',action='store_true');args=parser.parse_args()
P=args.artifacts.resolve();CODE=Path(__file__).resolve().parent;REUSE=args.reuse.resolve()
def module(n,p):
 s=importlib.util.spec_from_file_location(n,p);m=importlib.util.module_from_spec(s);sys.modules[n]=m;s.loader.exec_module(m);return m
a=module('existing_calibration_execution',REUSE/'macos-release-default-assessment/assessment.py');a.P=P
v=module('matched_calibration_classifier',CODE/'check.py')
a.need(args.run_authorized,'explicit quiet-window authorization required');a.need(not (P/'started.json').exists() and not (P/'stop-live').exists(),'sealed calibration')
plan=json.loads((REUSE/'macos-release-default-assessment/plan.json').read_text());a.need(a.sdk_vector()==plan['sdk'],'SDK drift')
sources=[p for p in CODE.parent.rglob('*') if p.is_file() and p.suffix in ['.py','.go','.m','.h','.mod','.sum','.md']];sources +=[CODE.parents[2]/'specs/macos-calibration.md',P/'framework',P/'match.json',P/'build-identity.json',P/'overlay.json',P/'overlay-identity.json',P/'desktop_session.go',P/'desktop_measurement.go',P/'frame_trace.go',CODE.parents[2]/'scripts/benchmark-grid.py',Path(a.__file__),Path(a.g.__file__),a.LEASE]
plan['pins']={str(p):a.sha(p) for p in sources};plan['pairs']=[['sourcecost']]
a.write(P/'bound-plan.json',{'source':a.SOURCE,'pins':plan['pins'],'pairs':plan['pairs'],'warmup':128,'measured':120,'period_ms':120,'maximum_seconds':60,'per_case_seconds':40,'no_gate_change':True,'no_release_qualification':True})
start=datetime.datetime.now(datetime.timezone.utc);deadline=start+datetime.timedelta(seconds=60);a.write(P/'started.json',{'started_utc':start.isoformat(),'maximum_seconds':60,'human_authorization':'user signed out and explicitly authorized continued testing; see chat','features_frozen':True});stopped=[False];old={sig:signal.signal(sig,lambda *_:stopped.__setitem__(0,True)) for sig in [signal.SIGINT,signal.SIGTERM]};results=[];outcome='INCOMPLETE'
def case(kind,root):
 root.mkdir();host=a.Host(plan,root,deadline,lease,stopped);r={'kind':kind}
 try:
  r['geometry_before']=host.geometry();r['configuration_before']=host.probes(root/'before')
  r['native']=host.command([str(P/'framework'),'timed',str(P/'match.json')],P,root/'native',40)
  r['configuration_after']=host.probes(root/'after');r['geometry_after']=host.geometry();host.finish()
  a.need(a.c.passed(r['native']) and r['native']['logs']['stderr']['bytes']==0,'native process failed/warned;retain raw scope')
  a.need(r['geometry_before']==r['geometry_after'],'persistent geometry changed')
  a.need(all(r['configuration_before'][k]['logs']['stdout']['sha256']==r['configuration_after'][k]['logs']['stdout']['sha256'] for k in r['configuration_before']),'configuration changed')
  lines=v.base.load(root/'native/stdout');a.need(len(lines)==1,'framework terminal document');doc=lines[0]
  match=json.loads((P/'match.json').read_text());r['validation']=v.validate(doc,match);a.write(root/'validation.json',r['validation']);r['outcome']='VALID_SOURCE_COST_CASE'
 except Exception as e:r['outcome']='REJECT';r['error']={'type':type(e).__name__,'message':str(e)}
 a.write(root/'result.json',r);(root/'sealed').touch();print(json.dumps({k:r[k] for k in ['kind','outcome','error'] if k in r}),flush=True);return r
try:
 with a.lease_module.LiveLease() as lease:
  for pair,order in enumerate(plan['pairs'],1):
   for kind in order:
    r=case(kind,P/('pair'+str(pair)+'-'+kind));r['pair']=pair;results.append(r)
    if r['outcome']=='REJECT':break
   if results[-1]['outcome']=='REJECT':break
  if len(results)==1 and results[-1]['outcome']!='REJECT':outcome='VALID_CPU_SOURCE_COST_DIAGNOSTIC'
  if results[-1]['outcome']=='REJECT':outcome='BLOCKED_BY_RETAINED_INVALID_CASE'

finally:
 (P/'stop-live').touch();a.write(P/'result.json',{'outcome':outcome,'cases':results,'finished_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'release_qualified':False,'gate_changed':False});[signal.signal(sig,h) for sig,h in old.items()]
