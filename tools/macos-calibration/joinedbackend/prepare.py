#!/usr/bin/env python3
"""Compose reviewed observers in fresh private copies; shipping inputs are immutable."""
from pathlib import Path
import argparse,hashlib,importlib.util,json,shutil

def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def vector(root):return {p.relative_to(root).as_posix():sha(p) for p in root.rglob('*') if p.is_file() and '.git' not in p.relative_to(root).parts}
def prepare(sdk,core,native,q,reuse):
 code=Path(__file__).resolve().parent;repo=code.parents[2]
 sdk,core,native,q,reuse=[p.resolve() for p in [sdk,core,native,q,reuse]]
 if any(p.exists() for p in [core,native,q]):raise ValueError('refuse existing destination')
 upload_identity=json.loads((repo/'docs/reports/macos-upload-probe/build-identity.json').read_text())
 backend_identity=json.loads((repo/'docs/reports/macos-backend-timing/build-identity-v2.json').read_text())
 upload=Path(upload_identity['root']);backend=Path(backend_identity['private_root'])
 if not all(sha(Path(p))==h for p,h in upload_identity['private_source_pins'].items()):raise ValueError('reviewed upload source drift')
 if sha(backend/'hal/src/metal/mod.rs')!=backend_identity['hal_patched_SHA256'] or sha(backend/'hal/src/metal/probe.rs')!=backend_identity['probe_SHA256']:raise ValueError('reviewed backend source drift')
 if sha(backend/'hal/src/metal/probe.rs')!=sha(code.parent/'backend/probe.rs'):raise ValueError('backend observer no longer matches reviewed code')
 s=importlib.util.spec_from_file_location('cpu_prepare',code.parent/'joined/prepare.py');m=importlib.util.module_from_spec(s);s.loader.exec_module(m)
 m.prepare(sdk,core,q,reuse/'macos-upload-probe')
 native.mkdir()
 for name,source in [('source',upload/'source'),('core',upload/'core'),('hal',backend/'hal')]:shutil.copytree(source,native/name,ignore=shutil.ignore_patterns('.git','target','__pycache__'))
 original={'source':vector(native/'source'),'core':vector(native/'core'),'hal':vector(native/'hal')}
 p=native/'source/Cargo.toml';p.write_text(p.read_text()+'wgpu-hal = { path = "../hal" }\n')
 p=native/'source/src/lib.rs';s=p.read_text();b=(backend/'source/src/lib.rs').read_text();needle='// Private diagnostic ABI. Not installed into the SDK or public framework.'
 if b.count(needle)!=1:raise ValueError('backend ABI boundary')
 p.write_text(s+'\n'+b[b.index(needle):])
 shutil.copytree(upload/'libs-darwin',native/'libs-darwin',ignore=shutil.ignore_patterns('*.a','*.dylib','PROVENANCE.json'))
 for name in ['native_main_darwin.go','backend_probe_darwin.go']:shutil.copy2(code/'testdata'/name,q/name)
 p=q/'framework.mod';s=p.read_text().replace(json.dumps(str(upload/'libs-darwin')),json.dumps(str(native/'libs-darwin')));p.write_text(s)
 p=q/'overlay.json';d=json.loads(p.read_text());f=repo/'tools/macos-calibration/upload/framework';d['Replace'][str(f/'backend_probe_darwin.go')]=str(q/'backend_probe_darwin.go');p.write_text(json.dumps(d,indent=2)+'\n')
 (q/'composition.json').write_text(json.dumps({'reviewed_upload_identity':upload_identity,'reviewed_backend_identity':backend_identity,'copied_before_composition':original,'private_root':str(native),'SDK_changed':False,'policy_changed':False},indent=2)+'\n')
 return q
if __name__=='__main__':
 p=argparse.ArgumentParser()
 for n in ['sdk','core','native','artifacts','reuse']:p.add_argument('--'+n,type=Path,required=True)
 a=p.parse_args();print(prepare(a.sdk,a.core,a.native,a.artifacts,a.reuse))
