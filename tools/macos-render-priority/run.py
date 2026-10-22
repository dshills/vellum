#!/usr/bin/env python3
"""One private priority candidate through the existing representative-case runner."""
from pathlib import Path
import argparse,datetime,importlib.util,json,signal,subprocess,sys
sys.dont_write_bytecode=True
p=argparse.ArgumentParser();p.add_argument('--artifacts',type=Path,required=True);p.add_argument('--reuse',type=Path,required=True);p.add_argument('--run-authorized',action='store_true');args=p.parse_args();Q=args.artifacts.resolve();X=args.reuse.resolve();R=Path(__file__).resolve().parents[2];sys.path.insert(0,str(R/'scripts'))
def module(name,path):
 s=importlib.util.spec_from_file_location(name,path);m=importlib.util.module_from_spec(s);sys.modules[name]=m;s.loader.exec_module(m);return m
a=module('priority_runtime',X/'macos-release-default-assessment/assessment.py');a.P=Q;a.T=R
plan=json.loads((Q/'plan.json').read_text());commit=json.loads((Q/'preparation-commit.json').read_text());a.need(args.run_authorized and not any((Q/n).exists() for n in ['started.json','stop-live','result.json']),'authorization/fresh priority scope');head=subprocess.check_output(['git','rev-parse','HEAD'],cwd=R,text=True).strip();a.need(head==commit['head'] and commit['review_complete'] is True and not subprocess.check_output(['git','status','--porcelain=v1'],cwd=R),'reviewed clean priority checkpoint');a.need(a.sdk_vector()==plan['sdk'] and all(a.sha(p)==h for p,h in plan['pins'].items()),'priority SDK/source/native/binary drift')
plan['pins'].update({str(Q/n):a.sha(Q/n) for n in ['plan.json','preparation-commit.json']});a.write(Q/'bound-plan.json',plan);start=datetime.datetime.now(datetime.timezone.utc);session_end=datetime.datetime.fromisoformat(plan['session_deadline_utc']);a.need(start<session_end,'bounded candidate session expired');deadline=min(start+datetime.timedelta(seconds=150),session_end);a.write(Q/'started.json',{'utc':start.isoformat(),'maximum_seconds':150,'per_native_seconds':60,'human_authorization':commit['human_authorization']});stopped=[False];previous={sig:signal.signal(sig,lambda *_:stopped.__setitem__(0,True)) for sig in [signal.SIGINT,signal.SIGTERM]};result=None
class Host(a.Host):
 def native(self,case,journal,out):return self.command(a.arguments(case,plan['binaries']['grid']['path'],journal),a.SDK/'examples/loadedgrid',out,60)
 def validate(self,case,path):
  journal=a.validator.load(path,a.sha(path));a.need(journal['Config']==a.expected(case),'priority workload drift');v=a.validator.validate(journal,head,plan['binaries']['grid']['sha256']);a.write(path.parent/'validation.json',v);self.last_validated=v
  a.need(v['measured_input_latency_ms']['p95']<=33.3,'priority candidate fails unchanged33.3ms latency gate')
  if case['paced']:a.need(v['paced_metrics']['cadence_threshold_met'] and v['paced_metrics']['input_latency_threshold_met'],'priority candidate fails unchanged paced gate')
  return v
try:
 with a.lease_module.LiveLease() as lease:result=a.c.execute_cases(plan,Host(plan,Q,deadline,lease,stopped),Q)
finally:
 for sig,handler in previous.items():signal.signal(sig,handler)
 (Q/'stop-live').touch()
print(json.dumps(result),flush=True);sys.exit(0 if result and result['result']=='PASS_SCOPED_NATIVE_FEEDBACK_PREFLIGHT' else 1)
