#!/usr/bin/env python3
"""Explicitly authorized single-attempt private headless completed-staging byte check; reuses existing owned execution."""
from pathlib import Path
import argparse,importlib.util,sys,json,datetime,signal,os
sys.dont_write_bytecode=True
parser=argparse.ArgumentParser();parser.add_argument('--artifacts',type=Path,required=True);parser.add_argument('--reuse',type=Path,required=True);parser.add_argument('--run-authorized',action='store_true');args=parser.parse_args()
P=args.artifacts.resolve();CODE=Path(__file__).resolve().parent;REUSE=args.reuse.resolve()
def module(n,p):
 s=importlib.util.spec_from_file_location(n,p);m=importlib.util.module_from_spec(s);sys.modules[n]=m;s.loader.exec_module(m);return m
a=module('existing_calibration_execution',REUSE/'macos-release-default-assessment/assessment.py');a.P=P
v=module('matched_calibration_classifier',CODE/'check.py')
a.need(args.run_authorized,'explicit continued-testing authorization required');a.need(not (P/'started.json').exists() and not (P/'stop-live').exists(),'sealed calibration')
plan=json.loads((REUSE/'macos-release-default-assessment/plan.json').read_text());a.need(a.sdk_vector()==plan['sdk'],'SDK drift')
sources=[p for p in CODE.parent.rglob('*') if p.is_file() and p.suffix in ['.py','.go','.m','.h','.mod','.sum','.md','.rs']];sources +=[CODE.parents[2]/'specs/macos-calibration.md',P/'bytecheck',P/'build-identity.json',P/'source-audit.json',Path('/private/tmp/vellum-staging-reuse/libs-darwin/arm64/libwgpu_native.a'),CODE.parents[2]/'scripts/benchmark-grid.py',Path(a.__file__),Path(a.g.__file__),a.LEASE]
plan['pins']={str(p):a.sha(p) for p in sources};sources.extend(Path(p) for p in json.loads((P/'build-identity.json').read_text())['private_source_pins']);plan['pairs']=[['off','on']]
a.need(os.access(P/'bytecheck',os.X_OK),'binary executable mode')
a.write(P/'bound-plan.json',{'source':a.SOURCE,'pins':plan['pins'],'pairs':plan['pairs'],'CPU_writes_per_mode':24,'GPU_readbacks_per_mode':22,'maximum_seconds':90,'per_case_seconds':30,'no_gate_change':True,'no_release_qualification':True})
start=datetime.datetime.now(datetime.timezone.utc);deadline=start+datetime.timedelta(seconds=90);a.write(P/'started.json',{'started_utc':start.isoformat(),'maximum_seconds':90,'human_authorization':'user signed out and explicitly authorized continued testing; see chat','features_frozen':True});stopped=[False];old={sig:signal.signal(sig,lambda *_:stopped.__setitem__(0,True)) for sig in [signal.SIGINT,signal.SIGTERM]};results=[];outcome='INCOMPLETE';run_error=None
def case(kind,root):
 root.mkdir();host=a.Host(plan,root,deadline,lease,stopped);r={'kind':kind}
 try:
  r['geometry_before']=host.geometry();r['configuration_before']=host.probes(root/'before')
  r['native']=host.command([str(P/'bytecheck'),kind],P,root/'native',30)
  r['configuration_after']=host.probes(root/'after');r['geometry_after']=host.geometry();host.finish()
  a.need(a.c.passed(r['native']) and r['native']['logs']['stderr']['bytes']==0,'native process failed/warned;retain raw scope')
  a.need(r['geometry_before']==r['geometry_after'],'persistent geometry changed')
  a.need(all(r['configuration_before'][k]['logs']['stdout']['sha256']==r['configuration_after'][k]['logs']['stdout']['sha256'] for k in r['configuration_before']),'configuration changed')
  lines=v.base.load(root/'native/stdout');a.need(len(lines)==1,'standalone terminal document');doc=lines[0]
  r['validation']=v.validate(doc,kind);a.write(root/'validation.json',r['validation']);r['outcome']='VALID_PRIVATE_STAGING_BYTE_CASE'
 except Exception as e:r['outcome']='REJECT';r['error']={'type':type(e).__name__,'message':str(e)}
 a.write(root/'result.json',r);(root/'sealed').touch();print(json.dumps({k:r[k] for k in ['kind','outcome','error'] if k in r}),flush=True);return r
try:
 with a.lease_module.LiveLease() as lease:
  for pair,order in enumerate(plan['pairs'],1):
   for kind in order:
    r=case(kind,P/('pair'+str(pair)+'-'+kind));r['pair']=pair;results.append(r)
    if r['outcome']=='REJECT':break
   if results and results[-1]['outcome']=='REJECT':break
  if len(results)==2 and all(r['outcome']=='VALID_PRIVATE_STAGING_BYTE_CASE' for r in results):outcome='VALID_PRIVATE_STAGING_BYTE_PAIR'
  if results and results[-1]['outcome']=='REJECT':outcome='BLOCKED_BY_RETAINED_INVALID_CASE'
  if stopped[0] or datetime.datetime.now(datetime.timezone.utc)>=deadline:outcome='STOPPED_OR_DEADLINE'
except Exception as error:
 outcome='ORCHESTRATION_ERROR';run_error={'type':type(error).__name__,'message':str(error)}
finally:
 try:
  (P/'stop-live').touch();a.write(P/'result.json',{'outcome':outcome,'error':run_error,'cases':results,'finished_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'release_qualified':False,'gate_changed':False})
 finally:
  for sig,handler in old.items():signal.signal(sig,handler)
