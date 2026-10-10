"""Record sources, results and artifact hashes without traversing build caches."""
from pathlib import Path
import hashlib,json,os,subprocess
root=Path(__file__).resolve().parent;workspace=root.parent.parent
previous=json.loads((root.parent/'nativeinput/evidence/manifest.json').read_text())
def digest(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def collect(directory,excluded_dirs=(),excluded_files=()):
    result={}
    for path,dirs,files in os.walk(directory):
        dirs[:]=[d for d in dirs if d not in excluded_dirs]
        for filename in sorted(files):
            p=Path(path)/filename;rel=str(p.relative_to(directory))
            if rel in excluded_files:continue
            result[rel]=digest(p)
    return dict(sorted(result.items()))
docs={rel:digest(workspace/rel) for rel in ['FOUNDATION_CONTRACTS_RESULTS.md','PROTOTYPE_PLAN.md','decisions/0006-foundation-contracts.md','research/RECOMMENDATION.md']}
recommendation=(workspace/'research/RECOMMENDATION.md').read_text().replace('(../decisions/','(decisions/').replace('(../FOUNDATION_CHECKS_RESULTS.md)','(FOUNDATION_CHECKS_RESULTS.md)').replace('(../FOUNDATION_CONTRACTS_RESULTS.md)','(FOUNDATION_CONTRACTS_RESULTS.md)')
compiled=[]
for p in sorted((root/'build').glob('*-mailbox.test')):
    compiled.append({'file':p.name,'description':subprocess.check_output(['file',str(p)],text=True).strip(),'sha256':digest(p)})
(root/'evidence/portable-crosscompile.json').write_text(json.dumps({'scope':'Mailbox Windows/Linux amd64 test binaries, CGO disabled; compilation only, no native runtime tests','artifacts':compiled},indent=2)+'\n')
data=dict(recorded_date='2026-10-09',scope='Opt-in exact multiline application text; standalone reserved mailbox; G1 open',pins=previous['pins'],environment=previous['environment'],source_hashes=collect(root,('build','evidence','__pycache__'),('contracts.mod','contracts.sum')),evidence_hashes=collect(root/'evidence',excluded_files=('manifest.json',)),artifact_hashes=collect(root/'build',('upstream',)),document_hashes=docs,installed_document_hashes={**{k:v for k,v in docs.items() if k!='research/RECOMMENDATION.md'},'RECOMMENDATION.md':hashlib.sha256(recommendation.encode()).hexdigest()},boundaries=['Seven local Cogent patches; earlier prototypes unchanged','Mailbox bounds queue-owned payload lengths and slots through Ack, not caller/native/downstream memory','Current native bridge is still unbounded; no native preflight or bounded event-pump integration','Native fixture is a hash-locked derivative of nativeinput changing only multiline constructor','Native probes under U.S. source are not live CJK/VoiceOver checks','No OS preferences changed; no upstream message or release publication','Font-memory policy and broader lexer/render regression failures remain unresolved','Cross-platform compile only; G1 and workbench expansion on hold'])
(root/'evidence/manifest.json').write_text(json.dumps(data,indent=2)+'\n')
