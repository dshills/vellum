"""Use existing Host for one unchanged framework pilot and fail-first checks."""
def execute(root,binary,host,validate,save,passed,result):
 result['geometry_before']=host.geometry();result['configuration_before']=host.probes(root/'before')
 match=root/'pilot-match.json'
 if match.exists():raise ValueError('fresh pilot match required')
 result['native']=host.command([str(binary),'pilot',str(match),'on','on'],root,root/'native',20)
 # finish asserts source/SDK integrity; it is not process or lease cleanup.
 # Host.command reaps through its owned executor, and run.py owns the lease context.
 result['configuration_after']=host.probes(root/'after');result['geometry_after']=host.geometry();host.finish()
 if not passed(result['native']) or result['native']['logs']['stderr']['bytes']:raise ValueError('startup command failure')
 if result['geometry_before']!=result['geometry_after']:raise ValueError('startup geometry drift')
 a,b=result['configuration_before'],result['configuration_after']
 if set(a)!=set(b) or any(a[k]['logs']['stdout']['sha256']!=b[k]['logs']['stdout']['sha256'] for k in a):raise ValueError('startup configuration drift')
 if host.stopping():raise ValueError('startup stop/deadline')
 result['validation']=validate(root/'native/stdout',match);save(root/'validation.json',result['validation'])
 return result['validation']
