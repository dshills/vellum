"""Keep the sealed package pending until the final internal review actually occurs."""
from pathlib import Path
import hashlib,json,shutil
p=Path(__file__).resolve().parent;o=p/'checkpoint-overlay'
d=o/'docs/reports/macos-parent-graph-endurance/records'
sha=lambda q:hashlib.sha256(q.read_bytes()).hexdigest()
q=d/'qualification.json';v=json.loads(q.read_text())
history=p/'history/seal-before-final-review';history.mkdir(parents=True,exist_ok=False)
shutil.copyfile(q,history/'qualification.json')
v.update(result='SEALED_SCOPED_LIFECYCLE_EVIDENCE_PENDING_FINAL_INTERNAL_REVIEW',checkpoint_complete=False)
v['provider_review'].pop('final_traceability_fix_separately_internally_reviewed')
v['provider_review'].update(final_traceability_fix_provider_reviewed=False,final_traceability_fix_internal_review_required=True)
shutil.copyfile(Path(__file__),d/Path(__file__).name)
v['records'][Path(__file__).name]={'bytes':Path(__file__).stat().st_size,'sha256':sha(Path(__file__))}
q.write_text(json.dumps(v,indent=2)+'\n')
f=p/'sealed-checkpoint-files.json';m=json.loads(f.read_text())
shutil.copyfile(f,history/f.name)
m['files']={a.relative_to(o).as_posix():sha(a)for a in sorted(o.rglob('*'))if a.is_file()}
m['pending_review_correction_producer_sha256']=sha(Path(__file__))
f.write_text(json.dumps(m,indent=2)+'\n')
print(json.dumps({'result':v['result'],'files':len(m['files']),'checkpoint_complete':False}))
