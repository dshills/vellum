"""One requested Anthropic review of the internally approved staged checkpoint."""
from pathlib import Path
import datetime, hashlib, json, subprocess, sys, time
R = Path('/Users/dshills/Development/projects/Vellum')
P = Path(__file__).resolve().parent
sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
review = json.loads((P/'checkpoint-source-review-final.json').read_text())
assert review['prism_ready'] is True and not review['unresolved_findings']
assert all(sha(R/n) == h for n,h in review['sources'].items())
assert not any((P/('prism.'+s)).exists() for s in ['json','stdout','stderr'])
staged = subprocess.check_output(['git','diff','--cached','--name-only','-z'], cwd=R, timeout=10)
assert {n for n in staged.decode().split('\0') if n} == set(review['sources'])
def index_sources():
    return {name: hashlib.sha256(subprocess.check_output(['git','show',':'+name],
        cwd=R,timeout=10)).hexdigest() for name in review['sources']}
staged_sources=index_sources()
assert staged_sources==review['sources']
diff = subprocess.check_output(['git','diff','--cached','--binary'], cwd=R, timeout=10)
with (P/'prism-staged.diff').open('xb') as out:
    out.write(diff)
argv = ['/Users/dshills/Development/Go/bin/prism','review','staged','--provider','anthropic','--model','claude-sonnet-5-5',
        '--baseline','none','--min-severity','low','--fail-on','medium','--no-auto-exclude',
        '--chunk-bytes','48000','--max-diff-bytes','250000','--format','json','--out',str(P/'prism.json')]
sys.path.insert(0, str(R/'scripts'))
from child_process import run_child
record = {'argv':argv, 'cwd':str(R), 'producer_sha256':sha(Path(__file__)),
          'review_sha256':sha(P/'checkpoint-source-review-final.json'), 'sources':review['sources'],
          'staged_sources':staged_sources,
          'staged_diff_sha256':hashlib.sha256(diff).hexdigest(),
          'started_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),
          'exit_code':None, 'error':None}
start = time.monotonic()
try:
    with (P/'prism.stdout').open('xb') as out, (P/'prism.stderr').open('xb') as err:
        record['exit_code'] = run_child(argv, cwd=R, stdout=out, stderr=err, timeout=600).returncode
except BaseException as error:
    record['error'] = {'type':type(error).__name__, 'message':str(error)}
finally:
    record.update(elapsed_seconds=time.monotonic()-start,
                  finished_utc=datetime.datetime.now(datetime.timezone.utc).isoformat())
    try:
        record['sources_stable']=all(sha(R/n)==h for n,h in review['sources'].items())
    except BaseException as error:
        record.update(sources_stable=False,source_observation_error={'type':type(error).__name__,'message':str(error)})
    try:
        record['staged_sources_stable']=index_sources()==staged_sources
        after=subprocess.check_output(['git','diff','--cached','--binary'],cwd=R,timeout=10)
        record['staged_diff_stable']=hashlib.sha256(after).hexdigest()==record['staged_diff_sha256']
    except BaseException as error:
        record.update(staged_sources_stable=False,staged_diff_stable=False,
            staged_observation_error={'type':type(error).__name__,'message':str(error)})
    for suffix in ['json','stdout','stderr']:
        path=P/('prism.'+suffix)
        record[suffix+'_sha256']=sha(path) if path.exists() else None
    with (P/'prism-execution.json').open('x') as out:
        out.write(json.dumps(record,indent=2)+'\n')
    print(json.dumps({k:record[k] for k in ['exit_code','error','elapsed_seconds','sources_stable']}))
assert record['exit_code'] in [0,1] and record['error'] is None and record['sources_stable']
assert record['staged_sources_stable'] and record['staged_diff_stable']
