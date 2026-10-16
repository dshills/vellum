"""Apply only the finalization explicitly permitted by the sealed-stage review."""
from pathlib import Path
import datetime,hashlib,json,shutil
p=Path(__file__).resolve().parent;o=p/'checkpoint-overlay'
d=o/'docs/reports/macos-parent-graph-endurance/records'
sha=lambda q:hashlib.sha256(q.read_bytes()).hexdigest()
read=lambda q:json.loads(q.read_text())
review=read(p/'sealed-stage-review.json');pending=read(p/'sealed-checkpoint-files.json')
assert review['commit_ready'] and not review['unresolved_findings']
assert review['sources']==pending['files']
assert all(sha(o/n)==h for n,h in review['sources'].items())
assert review['accepted_finalization']['source_runtime_archive_or_provider_changes_permitted']is False
assert not(p/'finalized-checkpoint-files.json').exists()
for n in ['sealed-stage-review.json','sealed-stage-review.md','sealed-checkpoint-files.json','finalize_checkpoint.py']:shutil.copyfile(p/n,d/n)
q=d/'qualification.json';before=read(q)
assert before['checkpoint_complete']is False
v=json.loads(json.dumps(before))
v.update(result='PASS_SCOPED_LIFECYCLE_EVIDENCE_CHECKPOINT',checkpoint_complete=True,final_sealed_stage_internal_review_required=False,final_internal_review_sha256=sha(p/'sealed-stage-review.json'))
v['provider_review']['final_traceability_fix_internal_review_required']=False
v['provider_review']['final_traceability_fix_internally_reviewed']=True
v['records']={f.relative_to(d).as_posix():{'bytes':f.stat().st_size,'sha256':sha(f)}for f in sorted(d.rglob('*'))if f.is_file() and f.name!='qualification.json'}
q.write_text(json.dumps(v,indent=2)+'\n')
for n,h in review['sources'].items():
 if n!='docs/reports/macos-parent-graph-endurance/records/qualification.json':assert sha(o/n)==h
assert all(v[k]is False for k in ['full_mac_qualified','full_overnight_qualified','ordinary_frame_budget_qualified','professional_memory_qualified','zero_cocoa_windows_qualified','physical_input_qualified'])
files={f.relative_to(o).as_posix():sha(f)for f in sorted(o.rglob('*'))if f.is_file()}
receipt={'schema':1,'result':'FINALIZED_INTERNALLY_APPROVED_SCOPED_LIFECYCLE_CHECKPOINT','utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'files':files,'final_qualification_sha256':sha(q),'pending_vector_sha256':sha(p/'sealed-checkpoint-files.json'),'final_internal_review_sha256':sha(p/'sealed-stage-review.json'),'producer_sha256':sha(Path(__file__)),'runtime_or_report_source_changes':False,'provider_invocations':1}
with(p/'finalized-checkpoint-files.json').open('x')as f:f.write(json.dumps(receipt,indent=2)+'\n')
print(json.dumps({'result':receipt['result'],'files':len(files),'full_mac_qualified':False}))
