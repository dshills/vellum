#!/usr/bin/env python3
"""First corrected-framework phase/pending cohort; no native pair or release waiver."""
from pathlib import Path
import argparse,datetime,importlib.util,json,signal,subprocess,sys
sys.dont_write_bytecode=True
p=argparse.ArgumentParser();p.add_argument('--artifacts',type=Path,required=True);p.add_argument('--reuse',type=Path,required=True);p.add_argument('--run-authorized',action='store_true');args=p.parse_args();Q=args.artifacts.resolve();X=args.reuse.resolve();CODE=Path(__file__).resolve().parent;R=CODE.parents[2];sys.path.insert(0,str(R/'scripts'))
def module(name,path):
 s=importlib.util.spec_from_file_location(name,path);m=importlib.util.module_from_spec(s);sys.modules[name]=m;s.loader.exec_module(m);return m
a=module('pending_host',X/'macos-release-default-assessment/assessment.py');a.P=Q;a.T=R
v=module('pending_validator',CODE.parent/'presentationphase/check.py');c=module('pending_control',CODE/'control.py');plan=json.loads((Q/'plan.json').read_text());commit=json.loads((Q/'preparation-commit.json').read_text())
a.need(args.run_authorized and not (Q/'started.json').exists() and not (Q/'stop-live').exists(),'pending authorization/fresh scope')
head=subprocess.check_output(['git','rev-parse','HEAD'],cwd=R,text=True).strip();dirty=subprocess.check_output(['git','status','--porcelain=v1'],cwd=R)
a.need(head==commit['head'] and not dirty and commit['review_complete'] is True,'pending prepared clean checkpoint drift')
a.need(a.sdk_vector()==plan['sdk'] and all(a.sha(p)==h for p,h in plan['pins'].items()),'pending SDK/compiled pins')
pins=dict(plan['pins']);pins.update({str(p):a.sha(p) for p in CODE.rglob('*') if p.is_file()});pins[str(Q/'preparation-commit.json')]=a.sha(Q/'preparation-commit.json');pins[str(Q/'plan.json')]=a.sha(Q/'plan.json');plan['pins']=pins;a.write(Q/'bound-plan.json',plan)
start=datetime.datetime.now(datetime.timezone.utc);deadline=start+datetime.timedelta(seconds=60);a.write(Q/'started.json',{'utc':start.isoformat(),'maximum_seconds':60,'per_process_seconds':40,'human_authorization':commit['human_authorization']});stopped=[False];old={sig:signal.signal(sig,lambda *_:stopped.__setitem__(0,True)) for sig in [signal.SIGINT,signal.SIGTERM]};result={'outcome':'INCOMPLETE','release_qualified':False,'gate_changed':False,'native_comparison':False}
try:
 with a.lease_module.LiveLease() as lease:
  def validate(path,match):
   docs=v.base.load(path);a.need(len(docs)==1,'one framework journal');actual=json.loads(match.read_text());a.need(actual==plan['expected_match'],'framework locked match drift');return v.validate(docs[0],actual,'framework')
  result['cohort']=c.execute(Q,Path(plan['binary']),Q/'match.json',a.Host(plan,Q,deadline,lease,stopped),validate,a.write,a.c.passed,result);result['outcome']='VALID_CORRECTED_FRAMEWORK_PENDING_COHORT'
except Exception as error:result.update(outcome='REJECT',error={'type':type(error).__name__,'message':str(error)})
finally:
 for sig,handler in old.items():signal.signal(sig,handler)
 (Q/'stop-live').touch();result.update(finished_utc=datetime.datetime.now(datetime.timezone.utc).isoformat());a.write(Q/'result.json',result)
print(json.dumps({k:result[k] for k in ['outcome','error'] if k in result}),flush=True)
sys.exit(0 if result['outcome']=='VALID_CORRECTED_FRAMEWORK_PENDING_COHORT' else 1)
