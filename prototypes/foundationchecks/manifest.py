"""Record immutable inputs, stage-scoped evidence and local artifacts."""
from pathlib import Path
import hashlib
import json

root = Path(__file__).resolve().parent
workspace = root.parent.parent
baseline = json.loads((root.parent / 'nativeinput/evidence/manifest.json').read_text())
def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()
def collect(directory, predicate):
    return {str(p.relative_to(directory)): digest(p) for p in sorted(directory.rglob('*')) if p.is_file() and predicate(p.relative_to(directory))}
sources = collect(root, lambda p: p.parts[0] not in {'build','evidence','__pycache__'} and str(p) not in {'checks.mod','checks.sum'})
evidence = collect(root/'evidence', lambda p: str(p) != 'manifest.json')
artifacts = collect(root/'build', lambda p: p.parts[0] not in {'upstream','font-cache','Vellum-FoundationChecks.app'})
# Generated source is recorded as an artifact; it is reproducibly derived from a
# hash-locked baseline fixture, not an additional source of truth.
documents = {rel:digest(workspace/rel) for rel in ['FOUNDATION_CHECKS_RESULTS.md','PROTOTYPE_PLAN.md','decisions/0005-foundation-checks.md','research/RECOMMENDATION.md']}
data = dict(recorded_date='2026-10-09',scope='Local remaining-foundation investigation; G1 open; no workbench expansion',pins=baseline['pins'],environment=baseline['environment'],source_hashes=sources,evidence_hashes=evidence,artifact_hashes=artifacts,document_hashes=documents,
 stages={
 'four_nativeinput_patches':['text-baseline','font-default','font-system-only','relocated','package-validation'],
 'six_patches':['text-patched','font-patched','targeted-tests','all-tests','minimized'],
 'pooled_native_lifetime':['adapter-pooled-1000','window-pooled-1000','window-pooled-5000'],
 'excluded_diagnostics':['diagnostic-font-discovery-failed*','diagnostic-minimized-dispatch-failed*','adapter-1000 (unpooled)','window-1000 (unpooled)','notarization.txt (superseded by absolute-path structured validation)'],
 'development_bundle':'Owned copy of original nativeinput fixture, library signed and bundle sealed ad-hoc; does not contain patches 0005/0006',
 'portable_crosscompile':'CGO disabled, Windows/Linux amd64 compile only; no native runtime evidence'},
 installed_document_hashes={**{k:v for k,v in documents.items() if k != 'research/RECOMMENDATION.md'}, 'RECOMMENDATION.md':hashlib.sha256((workspace/'research/RECOMMENDATION.md').read_text().replace('(../decisions/', '(decisions/').replace('(../FOUNDATION_CHECKS_RESULTS.md)', '(FOUNDATION_CHECKS_RESULTS.md)').encode()).hexdigest()},
 limitations=['No callable desktop automation this session; no new live candidate-panel/VoiceOver workflow evidence','No Windows/Linux native test host','Font coverage IDs are not visual typography correctness','No GPU/native allocation attribution; pooled hidden GLFW lifetime is not Cogent GPU lifecycle','Expanded upstream package suite fails baseline lexer markup and graphical suite stalls; focused race checks only','No bounded transport/loaded-font budget/exact-export implementation','Ad-hoc bundle is not Developer ID/notarized trusted distribution','No clean-machine test, occluded idle or cross-monitor DPI evidence','No settings changed, upstream messages or release publication'])
(root/'evidence/manifest.json').write_text(json.dumps(data,indent=2)+'\n')
