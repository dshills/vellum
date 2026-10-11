from pathlib import Path
import json,hashlib,subprocess,time,datetime,os,sys
sys.dont_write_bytecode=True
root=Path('/private/tmp/vellum-grid-variable-rows');out=Path('/private/tmp/vellum-grid-variable-native-final');out.mkdir()
sys.path.insert(0,str(root));from scripts.child_process import run_child
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
owned=['drivers/cogent/grid.go','drivers/cogent/grid_variable_rows_test.go','specs/grid-native.md']
portable=['virtual/grid.go','virtual/grid_geometry.go','virtual/grid_rows.go']
metadata=root/'drivers/cogent/native/dependencies.json'
prepared=[root/'drivers/cogent/build/upstream',root/'drivers/cogent/build/typesetting']
def tree_identity(directory):
 directory=directory.resolve();items={}
 for p in sorted(directory.rglob('*')):
  if p.is_file()and not p.is_symlink()and '.git'not in p.parts and '__pycache__'not in p.parts:
   items[str(p.relative_to(directory))]=sha(p)
 return {'resolved_root':str(directory),'files':len(items),'sorted_relative_path_hashes_sha256':hashlib.sha256(json.dumps(items,sort_keys=True,separators=(',',':')).encode()).hexdigest()}
def native_sources():
 values={}
 for p in sorted(root.rglob('*')):
  if not p.is_file()or p.is_symlink()or '.git'in p.parts or '__pycache__'in p.parts:continue
  rel=p.relative_to(root)
  if rel.parts[0]in ('examples','docs','specs','prototypes')or 'build'in rel.parts:continue
  if p.suffix in('.go','.m','.h','.c','.rs','.toml','.mod','.sum','.ttf','.otf')or p.name in('go.mod','go.sum'):
   values[str(rel)]=sha(p)
 return {'files':len(values),'sorted_relative_path_hashes_sha256':hashlib.sha256(json.dumps(values,sort_keys=True,separators=(',',':')).encode()).hexdigest()}
env=dict(os.environ,GOENV='off',GOWORK='off',GOTOOLCHAIN='local',GOFLAGS='',GOMODCACHE='/tmp/vellum-p0-cache/mod',GOPATH='/tmp/vellum-p0-cache/gopath',GOCACHE='/private/tmp/vellum-grid-variable-gocache',PYTHONDONTWRITEBYTECODE='1')
source_hashes={p:sha(root/p)for p in owned+portable}
proof={'schema':1,'scope':'native indexed variable-row acceptance; actual Core offscreen geometry/pointer/wheel/Page, stable anchor/ref retirement and native local/pool budget refusals. No desktop GUI/external AX/performance qualification.','base_commit':'1aea5eb','created_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'source_root':str(root),'owned_sources':{p:source_hashes[p]for p in owned},'portable_runtime':{p:source_hashes[p]for p in portable},'prepared_dependencies_metadata':{'path':str(metadata),'sha256':sha(metadata)},'prepared_before':[tree_identity(p)for p in prepared],'native_runtime_and_test_tree_before':native_sources(),'verifier':{'path':str(Path(__file__)),'sha256':sha(Path(__file__))},'go_version':run_child(['go','version'],cwd=root,env=env,timeout=20,capture_output=True,text=True,check=True).stdout.strip(),'preliminary_runs':{str(p):{'sha256':sha(p),'scope':'exploratory before root portable metric consistency corrections; not final acceptance'}for p in [Path('/private/tmp/vellum-grid-variable-native-initial.log'),Path('/private/tmp/vellum-grid-variable-native-corrected-1.log')]},'checks':[]}
checks=[('variable-native-race3',['go','test','-tags=offscreen','-race','-count=3','-run','TestManagedGridVariable','.']),('fixed-native-regressions',['go','test','-tags=offscreen','-race','-count=1','-run','TestManagedGridActualGeometryAndPools|TestManagedGridNativeBudgetRefusedBeforeAllocation|TestReviewGridEmptyHugeViewportRefusedBeforeMount|TestReviewGridFractionalGlobalEdges','.']),('native-offscreen-vet',['go','vet','-tags=offscreen','./...'])]
for name,argv in checks:
 start=time.monotonic();log=out/(name+'.log');print('START',name,flush=True)
 with log.open('x')as f:r=run_child(argv,cwd=root/'drivers/cogent',env=env,timeout=600,stdout=f,stderr=subprocess.STDOUT,text=True)
 proof['checks'].append({'name':name,'argv':argv,'cwd':str(root/'drivers/cogent'),'exit_code':r.returncode,'elapsed_seconds':time.monotonic()-start,'log':str(log),'log_sha256':sha(log)});print('DONE',name,r.returncode,flush=True)
 if r.returncode:break
proof['source_hashes_after']={p:sha(root/p)for p in source_hashes}
proof['changed_sources']={p:{'before':h,'after':proof['source_hashes_after'][p]}for p,h in source_hashes.items()if h!=proof['source_hashes_after'][p]}
proof['prepared_after']=[tree_identity(p)for p in prepared]
proof['native_runtime_and_test_tree_after']=native_sources()
proof['result']='PASS'if len(proof['checks'])==len(checks)and all(c['exit_code']==0 for c in proof['checks'])and not proof['changed_sources']and proof['prepared_before']==proof['prepared_after']and proof['native_runtime_and_test_tree_before']==proof['native_runtime_and_test_tree_after']else 'FAIL'
p=out/'qualification.json';p.write_text(json.dumps(proof,indent=2)+'\n');print(proof['result'],sha(p),flush=True)
