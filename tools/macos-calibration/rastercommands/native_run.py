#!/usr/bin/env python3
"""One authorized structural capture; no release or latency qualification."""
from pathlib import Path
import argparse,sys,importlib.util,json,datetime,signal,os
sys.dont_write_bytecode=True
p=argparse.ArgumentParser();p.add_argument('--artifacts',type=Path,required=True);p.add_argument('--reuse',type=Path,required=True);p.add_argument('--run-authorized',action='store_true');args=p.parse_args();Q=args.artifacts.resolve();X=args.reuse.resolve();CODE=Path(__file__).resolve().parent;R=CODE.parents[2]
sys.path.insert(0,str(R/'scripts'))
def module(n,path):
 s=importlib.util.spec_from_file_location(n,path);m=importlib.util.module_from_spec(s);sys.modules[n]=m;s.loader.exec_module(m);return m
a=module('raster_structure_host',X/'macos-release-default-assessment/assessment.py');a.P=Q;a.T=R
v=module('raster_structure_validator',CODE/'native_check.py')
a.need(args.run_authorized,'quiet desktop authorization required');a.need(not (Q/'started.json').exists() and not (Q/'stop-live').exists(),'sealed attempt')
plan=json.loads((X/'macos-release-default-assessment/plan.json').read_text());a.need(a.sdk_vector()==plan['sdk'],'SDK integrity')
sources=[p for p in CODE.rglob('*') if p.is_file()]+[Q/n for n in ['framework-v2','match.json','framework.mod','framework.sum','overlay.json','build-identity.json']]+[R/'scripts/child_process.py',R/'scripts/sdk_environment.py',Path(a.__file__),Path(a.g.__file__),Path(a.c.__file__),a.LEASE,Path(v.base.__file__)]
sources.extend(Path(p) for p in json.loads((Q/'build-identity.json').read_text())['pins']);plan['pins']={str(p):a.sha(p) for p in sources};a.need(os.access(Q/'framework-v2',os.X_OK),'executable mode')
a.write(Q/'bound-plan.json',{'pins':plan['pins'],'transitions':3,'period_ms':120,'request_lateness_limit_ms':20,'response_deadline_ms':120,'maximum_seconds':30,'per_case_seconds':20,'original_native_library':True,'no_latency_qualification':True,'no_gate_change':True})
start=datetime.datetime.now(datetime.timezone.utc);deadline=start+datetime.timedelta(seconds=30);a.write(Q/'started.json',{'utc':start.isoformat(),'human_authorization':'Yes, test when ready; up to two quiet minutes, current chat','maximum_seconds':30});stopped=[False];old={sig:signal.signal(sig,lambda *_:stopped.__setitem__(0,True)) for sig in [signal.SIGINT,signal.SIGTERM]};result={'outcome':'INCOMPLETE'}
try:
 with a.lease_module.LiveLease() as lease:
  root=Q/'capture';root.mkdir();host=a.Host(plan,root,deadline,lease,stopped)
  result['geometry_before']=host.geometry();result['configuration_before']=host.probes(root/'before')
  result['native']=host.command([str(Q/'framework-v2'),'timed',str(Q/'match.json'),'capture'],Q,root/'native',20)
  result['configuration_after']=host.probes(root/'after');result['geometry_after']=host.geometry();host.finish()
  a.need(a.c.passed(result['native']) and result['native']['logs']['stderr']['bytes']==0,'native process failed/warned; retain scope')
  a.need(result['geometry_before']==result['geometry_after'],'geometry drift')
  a.need(all(result['configuration_before'][k]['logs']['stdout']['sha256']==result['configuration_after'][k]['logs']['stdout']['sha256'] for k in result['configuration_before']),'configuration drift')
  docs=v.base.load(root/'native/stdout');a.need(len(docs)==1,'terminal document');result['validation']=v.validate(docs[0],json.loads((Q/'match.json').read_text()));a.write(Q/'validation.json',result['validation'])
  a.need(not host.stopping(),'stop/deadline');result['outcome']='VALID_NATIVE_RASTER_STRUCTURE'
except Exception as e:result.update(outcome='REJECT',error={'type':type(e).__name__,'message':str(e)})
finally:
 (Q/'stop-live').touch();result.update(finished_utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),release_qualified=False,gate_changed=False);a.write(Q/'result.json',result)
 for sig,handler in old.items():signal.signal(sig,handler)
print(json.dumps({k:result[k] for k in ['outcome','error'] if k in result}),flush=True)
