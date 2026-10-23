"""One private render-priority candidate; shipping Core/SDK remain immutable."""
from pathlib import Path
import hashlib,json,shutil

def prepare(core,out):
    core,out=Path(core).resolve(),Path(out).resolve()
    if out.exists():raise ValueError('fresh candidate output required')
    sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
    before={p.relative_to(core).as_posix():sha(p) for p in core.rglob('*') if p.is_file()}
    if any(p.is_symlink() for p in core.rglob('*')):raise ValueError('regular source required')
    shutil.copytree(core,out)
    if before!={p.relative_to(out).as_posix():sha(p) for p in out.rglob('*') if p.is_file()}:raise ValueError('copy drift')
    p=out/'core/display_backpressure.go';text=p.read_text();old='func (w *renderWindow) renderDisplayAsyncOwned(cp composer.Composer, trace *FrameTrace, tag, credit uint64) {\n\tdefer finishDisplayFrame(w.SystemWindow, credit)\n'
    if text.count(old)!=1:raise ValueError('render owner seam drift')
    p.write_text(text.replace(old,old+'\tpriority := beginRenderPriority()\n\tdefer priority.end()\n'))
    for name in ['frame_priority_darwin.go','frame_priority_other.go']:
        shutil.copyfile(Path(__file__).with_name('testdata')/name,out/'core'/name)
    if before!={p.relative_to(core).as_posix():sha(p) for p in core.rglob('*') if p.is_file()}:raise ValueError('original Core changed')
    after={p.relative_to(out).as_posix():sha(p) for p in out.rglob('*') if p.is_file()};changed=[n for n in before if before[n]!=after[n]];added=sorted(set(after)-set(before))
    if changed!=['core/display_backpressure.go'] or added!=['core/frame_priority_darwin.go','core/frame_priority_other.go']:raise ValueError('unexpected candidate delta')
    return {'original':str(core),'candidate':str(out),'original_files':before,'candidate_files':after,'changed':changed,'added':added,'shipping_unchanged':True}
