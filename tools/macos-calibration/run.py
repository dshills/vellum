#!/usr/bin/env python3
"""Explicitly authorized single-attempt calibration; reuses existing owned execution."""
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
sources=[p for p in CODE.rglob('*') if p.is_file() and p.suffix in ['.py','.go','.m','.h','.mod','.sum','.md']];sources +=[CODE.parent.parent/'specs/macos-calibration.md',P/'native',P/'framework',Path(a.__file__),Path(a.g.__file__),a.LEASE]
plan['pins']={str(p):a.sha(p) for p in sources};plan['pairs']=[['native','framework'],['framework','native'],['native','framework'],['framework','native'],['native','framework']]
a.write(P/'bound-plan.json',{'source':a.SOURCE,'pins':plan['pins'],'pairs':plan['pairs'],'warmup':128,'measured':120,'period_ms':120,'maximum_seconds':480,'per_case_seconds':45,'no_gate_change':True,'no_release_qualification':True})
start=datetime.datetime.now(datetime.timezone.utc);deadline=start+datetime.timedelta(seconds=480);a.write(P/'started.json',{'started_utc':start.isoformat(),'maximum_seconds':480,'human_authorization':'quiet desktop window explicitly approved;see chat','features_frozen':True});stopped=[False];old={sig:signal.signal(sig,lambda *_:stopped.__setitem__(0,True)) for sig in [signal.SIGINT,signal.SIGTERM]};results=[];outcome='INCOMPLETE'
def case(kind,root,pilot=False):
 root.mkdir();host=a.Host(plan,root,deadline,lease,stopped);r={'kind':kind,'pilot':pilot}
 try:
  r['geometry_before']=host.geometry();r['configuration_before']=host.probes(root/'before')
  command=[str(P/kind),'pilot' if pilot else 'timed',str(P/'match.json')] if kind=='framework' else [str(P/'native'),str(P/'match.json')]
  r['native']=host.command(command,P,root/'native',15 if pilot else 45)
  r['configuration_after']=host.probes(root/'after');r['geometry_after']=host.geometry();host.finish()
  a.need(a.c.passed(r['native']) and r['native']['logs']['stderr']['bytes']==0,'native process failed/warned;retain raw scope')
  a.need(r['geometry_before']==r['geometry_after'],'persistent geometry changed')
  a.need(all(r['configuration_before'][k]['logs']['stdout']['sha256']==r['configuration_after'][k]['logs']['stdout']['sha256'] for k in r['configuration_before']),'configuration changed')
  lines=v.load(root/'native/stdout');doc=v.native_doc(lines) if kind=='native' else lines[0];a.need(kind=='native' or len(lines)==1,'unexpected framework output')
  match=json.loads((P/'match.json').read_text());r['validation']=v.validate(doc,match,pilot);a.write(root/'validation.json',r['validation']);r['outcome']='VALID_CALIBRATION_SETUP' if pilot else 'VALID_CALIBRATION_CASE'
 except Exception as e:r['outcome']='REJECT';r['error']={'type':type(e).__name__,'message':str(e)}
 a.write(root/'result.json',r);(root/'sealed').touch();print(json.dumps({k:r[k] for k in ['kind','pilot','outcome','error'] if k in r}),flush=True);return r
try:
 with a.lease_module.LiveLease() as lease:
  pilot=case('framework',P/'pilot',True);results.append(pilot)
  if pilot['outcome']!='REJECT':
   plan['pins'][str(P/'match.json')]=a.sha(P/'match.json');a.write(P/'matched-plan.json',{'match':json.loads((P/'match.json').read_text()),'match_sha256':a.sha(P/'match.json'),'initial_plan_sha256':a.sha(P/'bound-plan.json')})
   for pair,order in enumerate(plan['pairs'],1):
    for kind in order:
     r=case(kind,P/('pair'+str(pair)+'-'+kind));r['pair']=pair;results.append(r)
     if r['outcome']=='REJECT':break
    if results[-1]['outcome']=='REJECT':break
   if len(results)==11 and results[-1]['outcome']!='REJECT':outcome='FIVE_VALID_MATCHED_SMALL_FIXTURE_PAIRS'
  if results[-1]['outcome']=='REJECT':outcome='BLOCKED_BY_RETAINED_INVALID_CASE'
finally:
 (P/'stop-live').touch();a.write(P/'result.json',{'outcome':outcome,'cases':results,'finished_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'release_qualified':False,'gate_changed':False});[signal.signal(sig,h) for sig,h in old.items()]
