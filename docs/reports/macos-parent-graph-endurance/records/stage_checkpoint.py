"""Stage only the internally approved four lifecycle evidence documents."""
from pathlib import Path
import datetime,hashlib,json,subprocess
p=Path(__file__).resolve().parent
r=Path('/Users/dshills/Development/projects/Vellum')
sha=lambda q:hashlib.sha256(q.read_bytes()).hexdigest()
git=lambda *a:subprocess.check_output(['git',*a],cwd=r,timeout=15)
review=json.loads((p/'checkpoint-source-review-final.json').read_text())
assert review['result']=='APPROVED_SOURCE_AND_SCOPED_EVIDENCE' and review['prism_ready'] and not review['unresolved_findings']
assert git('rev-parse','HEAD').decode().strip()=='6330b2d42050b46c2019e558074e087ed281e56a'
assert not git('status','--porcelain','-z')
assert not (p/'stage-checkpoint.json').exists()
for name,h in review['sources'].items():
 q=p/'checkpoint-overlay'/name
 assert q.is_file() and not q.is_symlink() and sha(q)==h
for name,h in review['sources'].items():
 q=r/name;q.parent.mkdir(parents=True,exist_ok=True)
 q.write_bytes((p/'checkpoint-overlay'/name).read_bytes())
 assert sha(q)==h
subprocess.run(['git','add','--',*review['sources']],cwd=r,timeout=15,check=True)
assert set(git('diff','--cached','--name-only','-z').decode().strip('\0').split('\0'))==set(review['sources'])
for name,h in review['sources'].items():assert hashlib.sha256(git('show',':'+name)).hexdigest()==h
record={'result':'STAGED_EXACT_INTERNALLY_REVIEWED_LIFECYCLE_SOURCES','utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'sources':review['sources'],'review_sha256':sha(p/'checkpoint-source-review-final.json'),'producer_sha256':sha(Path(__file__))}
with(p/'stage-checkpoint.json').open('x')as f:f.write(json.dumps(record,indent=2)+'\n')
print(json.dumps({'result':record['result'],'files':len(review['sources'])}))
