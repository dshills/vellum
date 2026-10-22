#!/usr/bin/env python3
"""One approved unchanged-binary startup check; no timing cohort or native reference."""
from pathlib import Path
import argparse,datetime,importlib.util,json,signal,subprocess,sys
sys.dont_write_bytecode=True
p=argparse.ArgumentParser();p.add_argument('--artifacts',type=Path,required=True);p.add_argument('--reuse',type=Path,required=True);p.add_argument('--run-authorized',action='store_true');args=p.parse_args();Q=args.artifacts.resolve();X=args.reuse.resolve();CODE=Path(__file__).resolve().parent;R=CODE.parents[2];sys.path.insert(0,str(R/'scripts'))
def module(name,path):
 s=importlib.util.spec_from_file_location(name,path);m=importlib.util.module_from_spec(s);sys.modules[name]=m;s.loader.exec_module(m);return m
a=module('startup_host',X/'macos-release-default-assessment/assessment.py');a.P=Q;a.T=R
v=module('startup_validator',CODE/'check.py');c=module('startup_control',CODE/'control.py');plan=json.loads((Q/'plan.json').read_text());commit=json.loads((Q/'preparation-commit.json').read_text())
a.need(args.run_authorized and not (Q/'started.json').exists() and not (Q/'stop-live').exists(),'startup authorization/fresh scope')
head=subprocess.check_output(['git','rev-parse','HEAD'],cwd=R,text=True).strip();dirty=subprocess.check_output(['git','status','--porcelain=v1'],cwd=R)
a.need(head==commit['head'] and not dirty and commit['review_complete'] is True,'startup prepared clean checkpoint drift')
a.need(a.sdk_vector()==plan['sdk'] and all(a.sha(p)==h for p,h in plan['pins'].items()),'startup SDK/compiled pins')
pins=dict(plan['pins']);pins.update({str(p):a.sha(p) for p in CODE.rglob('*') if p.is_file()});pins[str(Q/'preparation-commit.json')]=a.sha(Q/'preparation-commit.json');plan['pins']=pins;a.write(Q/'bound-plan.json',plan)
start=datetime.datetime.now(datetime.timezone.utc);deadline=start+datetime.timedelta(seconds=45);a.write(Q/'started.json',{'utc':start.isoformat(),'maximum_seconds':45,'per_process_seconds':20,'human_authorization':commit['human_authorization']});stopped=[False];old={sig:signal.signal(sig,lambda *_:stopped.__setitem__(0,True)) for sig in [signal.SIGINT,signal.SIGTERM]};result={'outcome':'INCOMPLETE','release_qualified':False,'gate_changed':False}
try:
 with a.lease_module.LiveLease() as lease:
  def validate(path,match):
   docs=v.base.load(path);a.need(len(docs)==1,'one startup journal');actual=json.loads(match.read_text());a.need(actual==plan['expected_match'],'startup fixture match drift');return v.validate(docs[0],actual)
  result['startup']=c.execute(Q,Path(plan['binary']),a.Host(plan,Q,deadline,lease,stopped),validate,a.write,a.c.passed,result);result['outcome']='VALID_CORRECTED_FRAMEWORK_STARTUP_ONLY'
except Exception as error:result.update(outcome='REJECT',error={'type':type(error).__name__,'message':str(error)})
finally:
 for sig,handler in old.items():signal.signal(sig,handler)
 (Q/'stop-live').touch();result.update(finished_utc=datetime.datetime.now(datetime.timezone.utc).isoformat());a.write(Q/'result.json',result)
print(json.dumps({k:result[k] for k in ['outcome','error'] if k in result}),flush=True)
sys.exit(0 if result["outcome"]=="VALID_CORRECTED_FRAMEWORK_STARTUP_ONLY" else 1)
