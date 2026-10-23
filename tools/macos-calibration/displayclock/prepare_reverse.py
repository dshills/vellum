#!/usr/bin/env python3
"""Fresh reversed scope using the exact previously tested binary and inputs."""
from pathlib import Path
import argparse,json,hashlib,os,shutil
BINARY='4ea7d61cb26c52405b8222a1fd1ea1c2069bb8f4b32f8932bb6cbb91b9ce3c1d'
NATIVE='10e39953bda1045da3bd9c315dcac0add646fa86e74627b3c896723d31b86a35'
def prepare(prior,q):
 prior,q=prior.resolve(),q.resolve();sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
 if q.exists():raise ValueError('refuse existing destination')
 old=json.loads((prior/'build-identity.json').read_text())
 if old['binary_sha256']!=BINARY or old['native_archive_sha256']!=NATIVE or sha(prior/'framework')!=BINARY or not os.access(prior/'framework',os.X_OK):raise ValueError('retained executable/native identity')
 if not all(sha(Path(p))==v for p,v in old['pins'].items()):raise ValueError('retained compiled-input drift')
 q.mkdir()
 for name in ['framework','match.json','framework.mod','framework.sum','overlay.json']:shutil.copy2(prior/name,q/name)
 shutil.copytree(prior/'consumer',q/'consumer')
 if sha(q/'framework')!=BINARY or not os.access(q/'framework',os.X_OK):raise ValueError('copied executable identity/mode')
 pins={**old['pins'],**{str(p):sha(p) for p in (q/'consumer').rglob('*') if p.is_file()}}
 (q/'build-identity.json').write_text(json.dumps({**old,'pins':pins,'experiment':'reversed-order-observer-control','retained_binary_origin':str(prior/'framework'),'rebuilt':False},indent=2)+'\n');return q
if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--prior',type=Path,required=True);p.add_argument('--artifacts',type=Path,required=True);a=p.parse_args();print(prepare(a.prior,a.artifacts))
