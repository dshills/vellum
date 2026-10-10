"""Apply the deferred-bitmap patch to an owned pinned go-text copy."""
from pathlib import Path
import hashlib,json,os,shutil,subprocess
r=Path(__file__).resolve().parent
info=json.loads(subprocess.check_output(['go','mod','download','-json','github.com/go-text/typesetting@v0.3.5-0.20260418130854-c41d02a44bec']))
o=r/'build/typesetting'
if o.exists():shutil.rmtree(o)
shutil.copytree(info['Dir'],o)
for directory,dirs,files in os.walk(o):
 Path(directory).chmod(0o755)
 for filename in files:
  p=Path(directory)/filename;p.chmod(p.stat().st_mode|0o200)
for patch in sorted((r/'patches').glob('*.patch')):
 hashes='original-hashes.json' if patch.name.startswith('0001') else '0002-hashes.json'
 for rel,expected in json.loads((r/'patches'/hashes).read_text()).items():
  p=o/rel;actual=hashlib.sha256(p.read_bytes()).hexdigest() if p.exists() else None
  if actual!=expected:raise RuntimeError(f'Patch input mismatch: {patch.name}/{rel}')
 subprocess.run(['patch','--batch','-p1','-i',str(patch)],cwd=o,check=True)
subprocess.run(['go','mod','edit','-modfile=font.mod','-replace=cogentcore.org/core=../foundationintegration/build/upstream','-replace=github.com/go-text/typesetting=./build/typesetting'],cwd=r,check=True)
