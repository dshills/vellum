#!/usr/bin/env python3
"""One authorized joined diagnostic. Observations cannot grant release acceptance."""
from pathlib import Path
import argparse,sys,importlib.util,json,datetime,signal,os,subprocess
sys.dont_write_bytecode=True
p=argparse.ArgumentParser();p.add_argument('--artifacts',type=Path,required=True);p.add_argument('--reuse',type=Path,required=True);p.add_argument('--run-authorized',action='store_true');args=p.parse_args();Q=args.artifacts.resolve();X=args.reuse.resolve();CODE=Path(__file__).resolve().parent;R=CODE.parents[2]
sys.path.insert(0,str(R/'scripts'))
def module(n,path):
 s=importlib.util.spec_from_file_location(n,path);m=importlib.util.module_from_spec(s);sys.modules[n]=m;s.loader.exec_module(m);return m
a=module('joined_host',X/'macos-release-default-assessment/assessment.py');a.P=Q;a.T=R
v=module('joined_validator',CODE/'check.py')
a.need(args.run_authorized,'current quiet desktop authorization required');a.need(not (Q/'started.json').exists() and not (Q/'stop-live').exists(),'sealed attempt')
identity=json.loads((Q/'build-identity.json').read_text());a.need(os.access(Q/'framework',os.X_OK) and a.sha(Q/'framework')==identity['binary_sha256'],'executable identity')
commit=json.loads((Q/'preparation-commit.json').read_text())['head'];head=subprocess.run(['git','rev-parse','HEAD'],cwd=R,text=True,capture_output=True,check=True).stdout.strip();a.need(head==commit,'prepared source checkpoint changed')
plan=json.loads((X/'macos-release-default-assessment/plan.json').read_text());a.need(a.sdk_vector()==plan['sdk'],'SDK integrity')
sources=[p for p in CODE.rglob('*') if p.is_file()]+[Q/n for n in ['framework','match.json','framework.mod','framework.sum','overlay.json','build-identity.json','preparation-commit.json']]+[R/'scripts/child_process.py',R/'scripts/sdk_environment.py',R/'scripts/benchmark-grid.py',Path(a.__file__),Path(a.g.__file__),Path(a.c.__file__),a.LEASE]
sources.extend(CODE.parent/n for n in ['check.py','sourcecost/check.py','upload/check.py','joined/check.py','backend/check.py','joinedbackend/check.py']);sources.extend(Path(p) for p in identity['pins']);plan['pins']={str(p):a.sha(p) for p in sources}
a.write(Q/'bound-plan.json',{'pins':plan['pins'],'prepared_commit':head,'requests':248,'warmup':128,'measured':120,'period_ms':120,'maximum_seconds':100,'per_case_seconds':40,'no_gate_change':True,'no_release_qualification':True,'no_pool_or_cache_policy_change':True,'order':['off','on']})
start=datetime.datetime.now(datetime.timezone.utc);deadline=start+datetime.timedelta(seconds=100);a.write(Q/'started.json',{'utc':start.isoformat(),'human_authorization':'Yes, test when ready; up to three quiet minutes for one display-clock off/on control, current chat','maximum_seconds':100});stopped=[False];old={sig:signal.signal(sig,lambda *_:stopped.__setitem__(0,True)) for sig in [signal.SIGINT,signal.SIGTERM]};result={'outcome':'INCOMPLETE'}
try:
 with a.lease_module.LiveLease() as lease:
  result['cases']=[]
  for mode in ['off','on']:
   root=Q/mode;root.mkdir();host=a.Host(plan,root,deadline,lease,stopped);case={'mode':mode};result['cases'].append(case)
   case['geometry_before']=host.geometry();case['configuration_before']=host.probes(root/'before')
   case['native']=host.command([str(Q/'framework'),'timed',str(Q/'match.json'),'on',mode],Q/'consumer',root/'native',40)
   case['configuration_after']=host.probes(root/'after');case['geometry_after']=host.geometry();host.finish()
   a.need(a.c.passed(case['native']) and case['native']['logs']['stderr']['bytes']==0,'native process failed/warned; retain attempt')
   a.need(case['geometry_before']==case['geometry_after'],'geometry drift');a.need(all(case['configuration_before'][k]['logs']['stdout']['sha256']==case['configuration_after'][k]['logs']['stdout']['sha256'] for k in case['configuration_before']),'configuration drift')
   docs=v.base.load(root/'native/stdout');a.need(len(docs)==1,'terminal document');case['validation']=v.validate(docs[0],json.loads((Q/'match.json').read_text()),mode);a.write(root/'validation.json',case['validation']);a.need(not host.stopping(),'stop/deadline')
  a.need(all(result['cases'][0]['configuration_before'][k]['logs']['stdout']['sha256']==result['cases'][1]['configuration_before'][k]['logs']['stdout']['sha256'] for k in result['cases'][0]['configuration_before']),'cross-case configuration drift')
  a.need(result['cases'][0]['geometry_before']==result['cases'][1]['geometry_before'],'cross-case geometry drift');result['outcome']='VALID_DISPLAY_CLOCK_OFF_ON_CONTROL'

except Exception as e:result.update(outcome='REJECT',error={'type':type(e).__name__,'message':str(e)})
finally:
 (Q/'stop-live').touch();result.update(finished_utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),release_qualified=False,gate_changed=False);a.write(Q/'result.json',result)
 for sig,handler in old.items():signal.signal(sig,handler)
print(json.dumps({k:result[k] for k in ['outcome','error'] if k in result}),flush=True)
