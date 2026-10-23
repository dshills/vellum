"""Materialize native diagnostic changes; avoid reliance on cgo C-source overlays."""
from pathlib import Path
import argparse,hashlib,importlib.util,json,shutil

def prepare(repo,prior,out):
 repo,prior,out=[p.resolve() for p in [repo,prior,out]];code=repo/'tools/macos-calibration/presentationphase';s=importlib.util.spec_from_file_location('original_phase_recipe',code/'prepare.py');m=importlib.util.module_from_spec(s);s.loader.exec_module(m)
 m.prepare(repo,prior,out)
 original=Path('/private/tmp/vellum-joined-backend-core');physical=out/'private-core'
 if any(p.is_symlink() for p in original.rglob('*')):raise ValueError('physical Core copy requires regular source tree')
 sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest();before={p.relative_to(original).as_posix():sha(p) for p in original.rglob('*') if p.is_file()};shutil.copytree(original,physical)
 actual={p.relative_to(physical).as_posix():sha(p) for p in physical.rglob('*') if p.is_file()}
 if actual!=before:raise ValueError('private Core copy drift')
 for name in ['presentation_darwin.m','presentation_darwin.go']:shutil.copyfile(out/name,physical/'system/driver/desktop'/name)
 overlay=json.loads((out/'overlay.json').read_text())
 frame=str(original/'core/frame_trace.go');shutil.copyfile(overlay['Replace'][frame],physical/'core/frame_trace.go')
 overlay['Replace']={k:v for k,v in overlay['Replace'].items() if not k.startswith(str(original)+'/')};(out/'overlay.json').write_text(json.dumps(overlay,indent=2)+'\n')
 p=out/'framework.mod';text=p.read_text();old='"'+str(original)+'"';new='"'+str(physical)+'"'
 if text.count(old)!=1:raise ValueError('expected original Core module replacement')
 p.write_text(text.replace(old,new))
 observer=(repo/'tools/macos-calibration/observe.m').read_text();needle='@"clock":@(cal_clock()),'
 if observer.count(needle)!=1:raise ValueError('observer source changed')
 observer=observer.replace(needle,'@"layer_class":layer?NSStringFromClass(layer.class):@"nil",@"view_layer_class":w.contentView.layer?NSStringFromClass(w.contentView.layer.class):@"nil",'+needle)
 (out/'consumer/observe_darwin.m').write_text('//go:build darwin && !offscreen\n'+observer)
 if before!={p.relative_to(original).as_posix():sha(p) for p in original.rglob('*') if p.is_file()}:raise ValueError('original Core mutated')
 modified={p.relative_to(physical).as_posix():sha(p) for p in physical.rglob('*') if p.is_file()};changed=sorted(k for k in before if before[k]!=modified[k]);allowed={'system/driver/desktop/presentation_darwin.m','system/driver/desktop/presentation_darwin.go','core/frame_trace.go'}
 if not set(changed)<=allowed or not {'system/driver/desktop/presentation_darwin.m','system/driver/desktop/presentation_darwin.go'}<=set(changed):raise ValueError('unexpected Core source changes')
 (out/'physical-core-identity.json').write_text(json.dumps({'original':str(original),'physical':str(physical),'original_files':before,'physical_files':modified,'changed':changed,'original_unchanged':True},indent=2)+'\n')
 return out

if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--prior',type=Path,required=True);p.add_argument('--artifacts',type=Path,required=True);a=p.parse_args();print(prepare(Path(__file__).resolve().parents[3],a.prior,a.artifacts))
