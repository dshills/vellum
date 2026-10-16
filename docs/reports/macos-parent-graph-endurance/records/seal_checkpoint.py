"""Bind completed scoped evidence, the actual provider review and its narrow fix."""
from pathlib import Path
import datetime,hashlib,json,shutil,subprocess,tarfile
p=Path(__file__).resolve().parent
r=Path('/Users/dshills/Development/projects/Vellum')
o=p/'checkpoint-overlay';d=o/'docs/reports/macos-parent-graph-endurance/records'
sha=lambda q:hashlib.sha256(q.read_bytes()).hexdigest()
read=lambda q:json.loads(q.read_text())
write=lambda q,v:q.write_text(json.dumps(v,indent=2)+'\n')
review=read(p/'checkpoint-source-review-final.json');before=review['sources']
execution=read(p/'prism-execution.json');provider=read(p/'prism.json')
assert execution['exit_code']==0 and execution['error'] is None
assert all(execution[k]for k in ['sources_stable','staged_sources_stable','staged_diff_stable'])
assert execution['sources']==execution['staged_sources']==before
assert execution['review_sha256']==sha(p/'checkpoint-source-review-final.json')
assert execution['producer_sha256']==sha(p/'prism_run.py')
for channel in ['json','stdout','stderr']:assert execution[channel+'_sha256']==sha(p/('prism.'+channel))
assert execution['staged_diff_sha256']==sha(p/'prism-staged.diff')
assert sha(p/'prism-staged.diff')==hashlib.sha256(subprocess.check_output(['git','diff','--cached','--binary'],cwd=r,timeout=15)).hexdigest()
coverage=provider['coverage']
assert coverage['complete'] and coverage['reviewer']==[{'provider':'anthropic','model':'claude-sonnet-5-5'}]
assert coverage['files']==4 and coverage['chunks']==coverage['llmCalls']==1
assert not any([coverage['excluded'],coverage['skipped'],coverage['truncatedBytes']])
assert [f['id']for f in provider['findings']]==['bc0f6ba65055fc6c']
assert [f['finding']['id']for f in provider['discarded']]==['5d8bac8e24391b60'] and not provider['suppressed']
assert not(p/'provider-resolution.json').exists()
for n,h in before.items():
 assert sha(r/n)==sha(o/n)==h
 q=p/'provider-reviewed-sources'/n;q.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(o/n,q)
history=p/'history/checkpoint-before-provider-fix';history.mkdir(parents=True,exist_ok=False)
shutil.copyfile(d/'qualification.json',history/'qualification.json')
with tarfile.open(d/'evidence.tar.xz','r:xz')as archive:
 members={m.name:m for m in archive.getmembers()}
 index=read(d/'archive-index.json')
 assert set(members)==set(index['members'])
 assert len(members)==119
 for n,meta in index['members'].items():
  m=members[n];data=archive.extractfile(m).read()
  assert m.isfile() and len(data)==meta['bytes'] and hashlib.sha256(data).hexdigest()==meta['sha256']
names=['evidence/run/completion-evidence-review.json','evidence/run/history/controller-v1/controls.json','evidence/run/history/controller-v2/controls-v2.json','evidence/scheduling-predecessor/soak-overnight.json']
assert all(n in members for n in names)
name='docs/reports/macos-parent-graph-endurance/README.md'
q=o/name;text=q.read_text();needle='findings. The earlier controller-fault and failed-soak records remain preserved.\n'
assert text.count(needle)==1
replacement=('findings. Its record is the archive-index entry\n'
 '`evidence/run/completion-evidence-review.json`. The preserved controller-fault\n'
 'records are `evidence/run/history/controller-v1/controls.json` and\n'
 '`evidence/run/history/controller-v2/controls-v2.json`; the failed-soak receipt is\n'
 '`evidence/scheduling-predecessor/soak-overnight.json`. Final source and provider\n'
 'review records are bound separately by `records/qualification.json`.\n')
q.write_text(text.replace(needle,replacement))
after={n:sha(o/n)for n in before}
resolution={'schema':1,'result':'PROVIDER_FINDINGS_ADDRESSED_WITH_NARROW_TRACEABILITY_FIX',
 'prism_report_sha256':sha(p/'prism.json'),'prism_execution_sha256':sha(p/'prism-execution.json'),
 'observed_finding_counts':{k:len(provider[k])for k in ['findings','discarded','suppressed']},
 'all_findings_triaged':True,'unresolved_blocking_findings':[],
 'source_changed':True,'provider_reviewed_sources':before,'final_sources':after,
 'final_fix_provider_rerun':False,'final_fix_requires_internal_review':True,
 'dispositions':[
  {'id':'bc0f6ba65055fc6c','channel':'findings','disposition':'FIXED_EXACT_ARCHIVE_ENTRY_REFERENCES','evidence':names,'changed_file':name},
  {'id':'5d8bac8e24391b60','channel':'discarded','disposition':'ADDRESSED_BY_COMMITTING_BOUND_RECORDS_IN_SAME_CHECKPOINT','evidence':'Archive, index, completed qualification, actual provider records and internal reviews are included in the final Git file vector. The tool discard is not used as a reason to ignore the concern.'}],
 'scope':'No runtime or raw evidence changed. One provider review covers the pre-fix four-source vector; the additive traceability fix and final package require a separate internal sealed-stage review.'}
write(p/'provider-resolution.json',resolution)
selected=['checkpoint-source-review.json','checkpoint-source-review-final.json','checkpoint-report-sources.json','report-correction.json','stage-checkpoint.json','prism_run.py','prism-staged.diff','prism.json','prism.stdout','prism.stderr','prism-execution.json','provider-resolution.json','correct_report.py','stage_checkpoint.py','seal_checkpoint.py']
for n in selected:assert(p/n).is_file()and not(p/n).is_symlink();shutil.copyfile(p/n,d/n)
for n,h in before.items():
 q=d/'provider-reviewed-sources'/n;q.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(p/'provider-reviewed-sources'/n,q)
records={q.relative_to(d).as_posix():{'bytes':q.stat().st_size,'sha256':sha(q)}for q in sorted(d.rglob('*'))if q.is_file() and q.name!='qualification.json'}
qualification=read(d/'qualification.json')
qualification.update(result='PASS_SCOPED_LIFECYCLE_EVIDENCE_CHECKPOINT',sealed_utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),checkpoint_sources=after,
 report_source_review_pending=False,provider_review_pending=False,checkpoint_complete=True,
 final_sealed_stage_internal_review_required=True,
 internal_source_review_sha256=sha(p/'checkpoint-source-review-final.json'),
 provider_review={'execution_sha256':sha(p/'prism-execution.json'),'report_sha256':sha(p/'prism.json'),'resolution_sha256':sha(p/'provider-resolution.json'),'coverage':coverage,'counts':resolution['observed_finding_counts'],'reviewed_sources':before,'final_traceability_fix_separately_internally_reviewed':True},
 records=records,seal_producer_sha256=sha(Path(__file__)))
write(d/'qualification.json',qualification)
final={q.relative_to(o).as_posix():sha(q)for q in sorted(o.rglob('*'))if q.is_file()}
write(p/'sealed-checkpoint-files.json',{'schema':1,'result':'PREPARED_FOR_FINAL_INTERNAL_SEALED_STAGE_REVIEW','files':final,'final_sources':after,'provider_reviewed_sources':before,'producer_sha256':sha(Path(__file__)),'archive_members':119,'runtime_or_raw_evidence_changed':False})
print(json.dumps({'result':qualification['result'],'files':len(final),'provider_findings':resolution['observed_finding_counts'],'traceability_fix_files':1,'final_internal_review_pending':True}))
