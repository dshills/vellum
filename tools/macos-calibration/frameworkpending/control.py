"""One measured framework cohort through the existing bounded Host."""
def execute(root,binary,match,host,validate,save,passed,result):
 if not match.is_file():raise ValueError('locked match missing')
 result['geometry_before']=host.geometry();result['configuration_before']=host.probes(root/'before')
 result['native']=host.command([str(binary),'timed',str(match),'on','on'],root,root/'native',40)
 # Host.finish asserts integrity. Host.command's executor and outer lease own cleanup.
 result['configuration_after']=host.probes(root/'after');result['geometry_after']=host.geometry();host.finish()
 if not passed(result['native']) or result['native']['logs']['stderr']['bytes']:raise ValueError('framework command failure')
 if result['geometry_before']!=result['geometry_after']:raise ValueError('framework geometry drift')
 a,b=result['configuration_before'],result['configuration_after']
 if set(a)!=set(b) or any(a[k]['logs']['stdout']['sha256']!=b[k]['logs']['stdout']['sha256'] for k in a):raise ValueError('framework configuration drift')
 if host.stopping():raise ValueError('framework stop/deadline')
 result['validation']=validate(root/'native/stdout',match);save(root/'validation.json',result['validation'])
 return result['validation']
