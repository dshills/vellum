"""Seal an owned development-only copy and record distribution checks."""
from pathlib import Path
import json
import shutil
import subprocess

root = Path(__file__).resolve().parent
source = root.parent / 'nativeinput/build/Vellum-NativeInput.app'
owned = root / 'build/development/Vellum-NativeInput.app'
if owned.exists():
    shutil.rmtree(owned)
owned.parent.mkdir(parents=True, exist_ok=True)
shutil.copytree(source, owned)
results = []
def run(args):
    p = subprocess.run(args, capture_output=True, text=True)
    results.append(dict(command=args, exit_code=p.returncode, stdout=p.stdout, stderr=p.stderr))
    return p.returncode
for library in sorted((owned / 'Contents').rglob('*.dylib')):
    if run(['/usr/bin/codesign', '--force', '--sign', '-', str(library)]):
        raise RuntimeError('development library signing failed')
if run(['/usr/bin/codesign', '--force', '--sign', '-', str(owned)]):
    raise RuntimeError('development bundle sealing failed')
if run(['/usr/bin/codesign', '--verify', '--deep', '--strict', '--verbose=2', str(owned)]):
    raise RuntimeError('development seal verification failed')
run(['/usr/bin/xcrun', 'stapler', 'validate', str(owned)])
run(['/usr/sbin/spctl', '--assess', '--type', 'execute', '--verbose=2', str(owned)])
(root / 'evidence/package-development.json').write_text(json.dumps(dict(scope='Ad-hoc development copy; no Developer ID, notarization submission or release trust established', results=results), indent=2)+'\n')
