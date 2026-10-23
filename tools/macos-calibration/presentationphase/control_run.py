"""Fixed native/framework order with fail-first, retained partial evidence."""
def execute(root,commands,host_factory,validate,compare,save,passed,cases):
 for kind in ['native','framework']:
  out=root/(kind+'-case');out.mkdir();host=host_factory(out);case={'kind':kind};cases.append(case)
  case['geometry_before']=host.geometry();case['configuration_before']=host.probes(out/'before')
  case['native']=host.command(commands[kind],root,out/'native',40)
  case['configuration_after']=host.probes(out/'after');case['geometry_after']=host.geometry();host.finish()
  if not passed(case['native']) or case['native']['logs']['stderr']['bytes']:raise ValueError('native command failure')
  if case['geometry_before']!=case['geometry_after']:raise ValueError('geometry drift')
  a,b=case['configuration_before'],case['configuration_after']
  if set(a)!=set(b) or any(a[k]['logs']['stdout']['sha256']!=b[k]['logs']['stdout']['sha256'] for k in a):raise ValueError('configuration drift')
  case['validation']=validate(out/'native/stdout',kind);save(out/'validation.json',case['validation'])
  if host.stopping():raise ValueError('stop/deadline after validation')
 a,b=cases
 if a['geometry_before']!=b['geometry_before'] or set(a['configuration_before'])!=set(b['configuration_before']) or any(a['configuration_before'][k]['logs']['stdout']['sha256']!=b['configuration_before'][k]['logs']['stdout']['sha256'] for k in a['configuration_before']):raise ValueError('cross-case configuration drift')
 result=compare(a['validation'],b['validation']);save(root/'pair-validation.json',result);return result
