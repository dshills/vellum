"""Apply twelve isolated patches to an owned copy of the pinned module."""
from pathlib import Path
import hashlib,json,os,shutil,subprocess
root=Path(__file__).resolve().parent
info=json.loads(subprocess.check_output(['go','mod','download','-json','cogentcore.org/core@v0.3.46-0.20261009095036-5cc519b24044']))
original=Path(info['Dir']); owned=root/'build/upstream'
if owned.exists(): shutil.rmtree(owned)
shutil.copytree(original,owned)
for directory,dirs,files in os.walk(owned):
    Path(directory).chmod(0o755)
    for filename in files:
        p=Path(directory)/filename;p.chmod(p.stat().st_mode|0o200)
for package in [root.parent/'nativeinput',root.parent/'foundationchecks',root.parent/'foundationcontracts',root]:
    for rel,expected in json.loads((package/'patches/original-hashes.json').read_text()).items():
        p=owned/rel
        actual=hashlib.sha256(p.read_bytes()).hexdigest() if p.exists() else None
        if actual!=expected: raise RuntimeError(f'Patch input mismatch: {package.name}/{rel}')
    for patch in sorted((package/'patches').glob('*.patch')):
        subprocess.run(['patch','--batch','-p1','-i',str(patch)],cwd=owned,check=True)
subprocess.run(['go','mod','edit','-modfile=integration.mod','-replace=cogentcore.org/core=./build/upstream'],cwd=root,check=True)
