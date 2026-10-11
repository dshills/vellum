"""Record affected unchanged native text/grid compatibility checks, without GUI."""
from pathlib import Path
import hashlib,json,subprocess,sys,time
ROOT=Path(__file__).resolve().parents[3]
sys.dont_write_bytecode=True
sys.path.insert(0,str(ROOT))
from scripts.child_process import run_child
OUT=Path(sys.argv[1]).resolve(); OUT.mkdir(parents=True,exist_ok=True)
with (OUT/'.started').open('x') as f:f.write('Exclusive evidence directory; never reuse.\n')
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def entry(p):
    if not p.exists():return {'absent':True}
    if not p.is_file():return {'non_regular':True}
    return sha(p)
def inputs():
    paths=subprocess.check_output(['git','ls-files','--cached','--others','--exclude-standard'],cwd=ROOT,text=True).splitlines()
    selected=sorted(set(p for p in paths if (p.endswith('.go') and not p.startswith(('examples/','prototypes/','docs/'))) or p in ['go.mod','go.sum','drivers/cogent/go.mod','drivers/cogent/go.sum','drivers/cogent/env.sh','drivers/cogent/prepare.sh','drivers/cogent/prepare.py','drivers/cogent/prepare_test.py','drivers/cogent/native/dependencies.json','scripts/child_process.py','docs/reports/macos-grid-loaded/qualify-native.py'] or p.startswith('drivers/cogent/native/patches/')))
    return {p:entry(ROOT/p) for p in selected}
def prepared():
    result={}
    for name in ['upstream','typesetting']:
        root=ROOT/'drivers/cogent/build'/name
        files={str(p.relative_to(root)):sha(p) for p in sorted(root.rglob('*')) if p.is_file() and '.git' not in p.relative_to(root).parts}
        result[name]={'files':len(files),'sha256':hashlib.sha256(json.dumps(files,sort_keys=True,separators=(',',':')).encode()).hexdigest()}
    return result
before=inputs();trees=prepared();proof={'scope':'Pinned preparation/manifest tests, upstream scene cursor race20, affected ordinary text settlement, managed text/editor/inline-grid/control compatibility race3 and both vet modes. Selected owned source/manifests/patches/wrappers and prepared-tree fingerprint before/after; no GUI, complete dependency/toolchain closure or performance claim.','source_root':str(ROOT),'inputs_before':before,'prepared_before':trees,'checks':[]}
commands=[('prepare-tests','python3 prepare_test.py'),('upstream-cursor-race','cd ./build/upstream\ngo test -race -count=20 -v -run "^TestVellumReviewTextCursor" ./core'),('text-grid-race', 'go test -tags offscreen -race -count=3 -v -run "^Test(ReviewFieldSettlement|ManagedTextField|ManagedEditor|ManagedGridInline|Control|RejectedCommittedText|FieldAccessibility|ManualCaptureKeepsLiveField)" .'),('offscreen-vet','go vet -tags offscreen .'),('desktop-vet','go vet .')]
for name,command in commands:
    log=OUT/(name+'.log');argv=['/bin/zsh','-c','export VELLUM_CACHE_DIR=/tmp/vellum-p0-cache\n. '+str(ROOT/'drivers/cogent/env.sh')+'\n'+command]
    start=time.monotonic();failure=None
    with log.open('xb') as stream:
        try:code=run_child(argv,cwd=ROOT/'drivers/cogent',stdout=stream,stderr=subprocess.STDOUT,timeout=600).returncode
        except Exception as e:code=124 if isinstance(e,subprocess.TimeoutExpired) else 125;failure={'type':type(e).__name__,'message':str(e)}
    required={}
    if name=='upstream-cursor-race':
        required={test:20 for test in ['TestVellumReviewTextCursorConcurrentScenesIgnoreLegacyOwner','TestVellumReviewTextCursorSceneOwnerAndPositionRetention','TestVellumReviewTextCursorSingleSceneSwitchAndNames','TestVellumReviewTextCursorManualClockRetained','TestVellumReviewTextCursorMainSpriteLockOwnsCursorMap']}
    if name=='text-grid-race':
        required={test:3 for test in ['TestReviewFieldSettlementOrdinaryKeyBeforeCallback','TestReviewFieldSettlementNoDuplicateForUnchangedSecondBarrier']}
    lines=log.read_text(errors='replace').splitlines()
    observed={test:lines.count('=== RUN   '+test) for test in required}
    covered=required==observed
    c={'name':name,'argv':argv,'exit_code':code,'failure':failure,'required_test_runs':required,'observed_test_runs':observed,'coverage_qualified':covered,'elapsed_seconds':round(time.monotonic()-start,3),'log':str(log),'log_sha256':sha(log)};proof['checks'].append(c);print(json.dumps(c),flush=True)
    if code or not covered:break
proof['inputs_after']=inputs();proof['prepared_after']=prepared();proof['unchanged']=before==proof['inputs_after'] and trees==proof['prepared_after'];proof['result']='PASS' if proof['unchanged'] and len(proof['checks'])==len(commands) and all(c['exit_code']==0 and c['coverage_qualified'] for c in proof['checks']) else 'FAIL'
p=OUT/'qualification.json'
with p.open('x') as f:f.write(json.dumps(proof,indent=2)+'\n')
print(proof['result'],sha(p),flush=True);sys.exit(0 if proof['result']=='PASS' else 1)
