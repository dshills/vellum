from pathlib import Path
import datetime, hashlib, json, os, subprocess, sys, time
sys.dont_write_bytecode = True
root = Path('/private/tmp/vellum-grid-variable-rows')
out = Path('/private/tmp/vellum-grid-variable-desktop'); out.mkdir(exist_ok=True)
sys.path.insert(0, str(root))
from scripts.child_process import run_child
sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
producer = Path('/private/tmp/vellum-variable-grid-example/qualified-v1/qualification.json')
qualification = json.loads(producer.read_text())
binary = Path(qualification['binary']['path'])
expected_inputs = {(str(Path('examples/grid')/p) if name in ['owned','preserved_example'] else p):h for name,group in qualification['before'].items() for p,h in group.items()}
source_paths = sorted(expected_inputs)
def sources(): return {p:sha(root/p) if (root/p).is_file() else None for p in source_paths}
def prepared():
    result=[]
    for name in ['upstream','typesetting']:
        directory=(root/'drivers/cogent/build'/name).resolve()
        values={str(p.relative_to(directory)):sha(p) for p in sorted(directory.rglob('*')) if p.is_file() and not p.is_symlink() and '.git' not in p.parts and '__pycache__' not in p.parts}
        result.append({'resolved_root':str(directory),'files':len(values),'sorted_relative_path_hashes_sha256':hashlib.sha256(json.dumps(values,sort_keys=True,separators=(',',':')).encode()).hexdigest()})
    return result
before=sources(); dependencies=prepared()
assert qualification['result']=='PASS' and sha(binary)==qualification['binary']['sha256']
assert before == expected_inputs
proof={'created_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'scope':'Actual Cocoa completed-frame geometry, explicit owner commands and independent window lifetimes; no physical input, compositor teardown or performance qualification','producer_sha256':sha(producer),'binary':qualification['binary'],'source_inputs':before,'prepared_before':dependencies,'checks':[]}
for name,args,stages in [('fixed',[],['initial-measured','far-jump-frozen','indexed-replacement','pass']),('variable',['-variable-rows'],['initial-measured','far-jump-frozen','indexed-replacement','variable-pixel-page','pass'])]:
    argv=[str(binary),'-smoke']+args; log=out/(name+'.log'); start=time.monotonic()
    with log.open('w') as stream:
        proc=run_child(argv,cwd=root/'examples/grid',env=dict(os.environ,PYTHONDONTWRITEBYTECODE='1'),stdout=stream,stderr=subprocess.STDOUT,timeout=60,text=True)
    records=[]
    for line in log.read_text().splitlines():
        try: item=json.loads(line)
        except json.JSONDecodeError: continue
        if isinstance(item,dict) and 'stage' in item:records.append(item)
    observed=[item['stage'] for item in records]
    proof['checks'].append({'name':name,'argv':argv,'exit_code':proc.returncode,'seconds':round(time.monotonic()-start,3),'log':str(log),'log_sha256':sha(log),'stages':observed,'records':records})
    print(json.dumps({'name':name,'exit_code':proc.returncode,'stages':observed}),flush=True)
    if proc.returncode:break
proof['source_unchanged']=sources()==before; proof['prepared_after']=prepared();proof['binary_unchanged']=sha(binary)==qualification['binary']['sha256']
proof['result']='PASS' if len(proof['checks'])==2 and all(c['exit_code']==0 and c['stages'] and c['stages'][-1]=='pass' for c in proof['checks']) and proof['source_unchanged'] and proof['binary_unchanged'] and dependencies==proof['prepared_after'] else 'FAIL'
path=out/'qualification.json';path.write_text(json.dumps(proof,indent=2)+'\n')
print(proof['result'],sha(path),flush=True)
raise SystemExit(0 if proof['result']=='PASS' else 1)
