"""Bind actual final approval and mark the scoped code checkpoint complete."""
from pathlib import Path
import datetime,hashlib,json,shutil
p=Path(__file__).resolve().parent;o=p/'integration';d=o/'docs/reports/window-context-retirement/records'
sha=lambda q:hashlib.sha256(q.read_bytes()).hexdigest()
review=json.loads((p/'final-source-review.json').read_text());m=json.loads((p/'final-checkpoint-files.json').read_text())
assert review['commit_ready']and not review['unresolved_findings']
assert review['sources']==m['files']and all(sha(o/n)==h for n,h in m['files'].items())
assert not(p/'sealed-checkpoint-files.json').exists()
for name in ['final-source-review.json','final-source-review.md','final-checkpoint-files.json','finalize_checkpoint.py']:shutil.copyfile(p/name,d/name)
q=d/'qualification.json';v=json.loads(q.read_text());assert v['checkpoint_complete']is False
v.update(result='PASS_SCOPED_OFFSCREEN_WINDOW_RETIREMENT_CHECKPOINT',checkpoint_complete=True,internal_report_review_pending=False,final_internal_review_required=False,final_internal_review_sha256=sha(p/'final-source-review.json'),sealed_utc=datetime.datetime.now(datetime.timezone.utc).isoformat())
v['provider_review']['final_delta_internally_reviewed']=True
v['records']={f.relative_to(d).as_posix():{'bytes':f.stat().st_size,'sha256':sha(f)}for f in sorted(d.rglob('*'))if f.is_file()and f.name!='qualification.json'}
assert v['full_mac_qualified']is False and v['native_ui_qualified']is False
q.write_text(json.dumps(v,indent=2)+'\n')
for n,h in m['files'].items():
 if n!='docs/reports/window-context-retirement/records/qualification.json':assert sha(o/n)==h
files={f.relative_to(o).as_posix():sha(f)for f in sorted(o.rglob('*'))if f.is_file()}
v={'result':'SEALED_APPROVED_WINDOW_RETIREMENT_CHECKPOINT','files':files,'pending_vector_sha256':sha(p/'final-checkpoint-files.json'),'final_review_sha256':sha(p/'final-source-review.json'),'producer_sha256':sha(Path(__file__)),'final_qualification_sha256':sha(q),'native_or_gpu_executed':False,'full_mac_qualified':False}
with(p/'sealed-checkpoint-files.json').open('x')as f:f.write(json.dumps(v,indent=2)+'\n')
print(json.dumps({'result':v['result'],'files':len(files),'native_or_gpu_executed':False}))
