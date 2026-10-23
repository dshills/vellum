"""Fixed reversed order with complete per-case accounting and no retries."""
from pathlib import Path
ORDER=('on','off')
def need(ok,why):
 if not ok:raise ValueError(why)
def execute(root,binary,cwd,match,host_factory,validate,save,passed,cases):
 for mode in ORDER:
  out=root/mode;out.mkdir();host=host_factory(out)
  case={'mode':mode};cases.append(case)
  case['geometry_before']=host.geometry();case['configuration_before']=host.probes(out/'before')
  case['native']=host.command([str(binary),'timed',str(match),'on',mode],cwd,out/'native',40)
  case['configuration_after']=host.probes(out/'after');case['geometry_after']=host.geometry();host.finish()
  need(passed(case['native']) and case['native']['logs']['stderr']['bytes']==0,'native process failed/warned; retain attempt')
  need(case['geometry_before']==case['geometry_after'],'geometry drift')
  before,after=case['configuration_before'],case['configuration_after']
  need(before.keys()==after.keys() and all(before[k]['logs']['stdout']['sha256']==after[k]['logs']['stdout']['sha256'] for k in before),'configuration drift')
  case['validation']=validate(out/'native/stdout',mode);save(out/'validation.json',case['validation']);need(not host.stopping(),'stop/deadline')
 first,last=cases
 need(first['configuration_before'].keys()==last['configuration_before'].keys() and all(first['configuration_before'][k]['logs']['stdout']['sha256']==last['configuration_before'][k]['logs']['stdout']['sha256'] for k in first['configuration_before']),'cross-case configuration drift')
 need(first['geometry_before']==last['geometry_before'],'cross-case geometry drift')
