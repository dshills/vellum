from pathlib import Path
import datetime,hashlib,json,shutil
p=Path(__file__).resolve().parent;overlay=p/'checkpoint-overlay'
sha=lambda q:hashlib.sha256(q.read_bytes()).hexdigest()
review=json.loads((p/'checkpoint-source-review.json').read_text())
assert review['result']=='NEEDS_REPORT_CORRECTIONS'
history=p/'history/checkpoint-report-v1';history.mkdir(parents=True,exist_ok=False)
shutil.copytree(overlay,history/'checkpoint-overlay')
for name in ['checkpoint-source-review.json','checkpoint-source-review.md','checkpoint-report-sources.json','draft_report.py','archive_completed.py']:shutil.copyfile(p/name,history/name)
name='docs/reports/macos-parent-graph-endurance/README.md';path=overlay/name
before=path.read_text();before_sha=sha(path)
replacements=[
 ('List/canvas/tree native identities, generation and scope are checked\nagainst owner publication by the bound fixture.',
  'ListBox checks require ordinary native host/row identities, accepted values\nand clipping. Tree/canvas checks additionally require current generation and\nscope against owner publication by the bound fixture.'),
 ('Clipboard access and preference changes are disabled\nand reported false by the fixture; global OS state is outside this observation.',
  'The fixture requests no general clipboard or OS preference changes and\nexplicitly disables window geometry persistence. Its reported false clipboard/\npreference flags do not independently observe global OS state.')]
after=before
for old,new in replacements:assert after.count(old)==1;after=after.replace(old,new)
path.write_text(after)
sources=json.loads((p/'checkpoint-report-sources.json').read_text());sources[name]=sha(path)
(p/'checkpoint-report-sources.json').write_text(json.dumps(sources,indent=2)+'\n')
qualification=overlay/'docs/reports/macos-parent-graph-endurance/records/qualification.json'
q=json.loads(qualification.read_text());assert q['checkpoint_complete']is False and q['report_source_review_pending']is True and q['provider_review_pending']is True
q['checkpoint_sources']=sources;qualification.write_text(json.dumps(q,indent=2)+'\n')
assert sha(overlay/'docs/reports/macos-parent-graph-endurance/records/evidence.tar.xz')=='9b95d7585db56859d6bc22c96217ac40a3b85f2e8304814286a4743a32b6286c'
receipt={'created_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'result':'CORRECTED_TWO_REPORT_SCOPE_OVERCLAIMS','source':name,'before_sha256':before_sha,'after_sha256':sha(path),'prior_review_sha256':sha(history/'checkpoint-source-review.json'),'replacements':[{'before':old,'after':new}for old,new in replacements],'archive_unchanged':True,'native_evidence_or_runtime_source_changed':False,'qualification_pending':True}
with(p/'report-correction.json').open('x')as out:out.write(json.dumps(receipt,indent=2)+'\n')
print(json.dumps({k:receipt[k]for k in ['result','before_sha256','after_sha256','archive_unchanged','native_evidence_or_runtime_source_changed']}))
