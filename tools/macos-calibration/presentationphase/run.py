#!/usr/bin/env python3
"""One approved phase-matched pair; no production policy or gate change."""
from pathlib import Path
import argparse,datetime,importlib.util,json,os,signal,subprocess,sys
sys.dont_write_bytecode=True
p=argparse.ArgumentParser();p.add_argument('--artifacts',type=Path,required=True);p.add_argument('--reuse',type=Path,required=True);p.add_argument('--run-authorized',action='store_true');args=p.parse_args();Q=args.artifacts.resolve();X=args.reuse.resolve();CODE=Path(__file__).resolve().parent;R=CODE.parents[2];sys.path.insert(0,str(R/'scripts'))
def module(name,path):
 s=importlib.util.spec_from_file_location(name,path);m=importlib.util.module_from_spec(s);sys.modules[name]=m;s.loader.exec_module(m);return m
a=module('phase_host',X/'macos-release-default-assessment/assessment.py');a.P=Q;a.T=R
v=module('phase_validator',CODE/'check.py');control=module('phase_control_run',CODE/'control_run.py')
a.need(args.run_authorized and not (Q/'started.json').exists() and not (Q/'stop-live').exists(),'authorization/fresh scope')
identity=json.loads((Q/'build-identity.json').read_text());commit=json.loads((Q/'preparation-commit.json').read_text())['head'];head=subprocess.check_output(['git','rev-parse','HEAD'],cwd=R,text=True).strip();a.need(head==commit,'prepared checkpoint drift')
plan=json.loads((X/'macos-release-default-assessment/plan.json').read_text());a.need(a.sdk_vector()==plan['sdk'],'SDK integrity');pins=dict(identity['pins']);pins.update({str(p):a.sha(p) for p in CODE.rglob('*') if p.is_file()});pins.update({str(Q/n):a.sha(Q/n) for n in ['framework','native','match.json','build-identity.json','preparation-commit.json']});pins.update({str(p):a.sha(p) for p in [Path(a.__file__),Path(a.g.__file__),Path(a.c.__file__),a.LEASE,R/'scripts/child_process.py']});plan['pins']=pins;a.need(all(a.sha(p)==h for p,h in pins.items()),'compiled inputs drift')
commands={'native':[str(Q/'native'),str(Q/'match.json')],'framework':[str(Q/'framework'),'timed',str(Q/'match.json'),'on','on']}
a.write(Q/'bound-plan.json',{'pins':pins,'prepared_commit':head,'commands':commands,'order':['native','framework'],'maximum_seconds':100,'per_case_seconds':40,'warmup':128,'measured':120,'phase_fractions':[.125,.375,.625,.875],'anchor_stride':8,'target_lead_periods':3,'maximum_receipt_phase_error_ms':2,'maximum_per_phase_p50_p95_error_difference_ms':1,'gate_changed':False})
start=datetime.datetime.now(datetime.timezone.utc);deadline=start+datetime.timedelta(seconds=100);a.write(Q/'started.json',{'utc':start.isoformat(),'human_authorization':'Yes, test when ready; three quiet minutes for phase-matched native/framework control, current chat','maximum_seconds':100});stopped=[False];old={sig:signal.signal(sig,lambda *_:stopped.__setitem__(0,True)) for sig in [signal.SIGINT,signal.SIGTERM]};result={'outcome':'INCOMPLETE','cases':[]}
try:
 with a.lease_module.LiveLease() as lease:
  def validate(path,kind):
   docs=v.base.load(path);doc=v.base.native_doc(docs) if kind=='native' else docs[0];a.need(kind=='native' or len(docs)==1,'one terminal journal');return v.validate(doc,json.loads((Q/'match.json').read_text()),kind)
  result['pair_validation']=control.execute(Q,commands,lambda out:a.Host(plan,out,deadline,lease,stopped),validate,v.compare,a.write,a.c.passed,result['cases']);result['outcome']='VALID_PHASE_MATCHED_PRESENTATION_CONTROL'
except Exception as error:result.update(outcome='REJECT',error={'type':type(error).__name__,'message':str(error)})
finally:
 (Q/'stop-live').touch();result.update(finished_utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),release_qualified=False,gate_changed=False);a.write(Q/'result.json',result)
 for sig,handler in old.items():signal.signal(sig,handler)
print(json.dumps({k:result[k] for k in ['outcome','error'] if k in result}),flush=True)
