"""Record bounded core-module race/vet checks against unchanged source inputs."""
from pathlib import Path
import hashlib
import json
import subprocess
import sys
import time

root = Path(__file__).resolve().parents[3]
sys.dont_write_bytecode = True
sys.path.insert(0, str(root))
from scripts.child_process import run_child

out = Path(sys.argv[1]).resolve() if len(sys.argv) == 2 else root / 'docs/reports/macos-grid-selection'
out.mkdir(parents=True, exist_ok=True)
logs = out / 'logs'
logs.mkdir(exist_ok=True)
path = out / 'core-qualification.json'
if path.exists() or any((logs / (name + '.log')).exists() for name in ['portable-race3', 'core-race', 'core-vet']):
    raise SystemExit('Refusing to overwrite prior evidence; supply a fresh output directory')
with (out / '.core-qualification-started').open('x') as marker:
    marker.write('Exclusive evidence directory; never reuse after interruption.\n')
sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
names = subprocess.check_output(['git', 'ls-files', '--cached', '--others', '--exclude-standard'], cwd=root, text=True).splitlines()
paths = sorted(p for p in names if (p.endswith('.go') and not p.startswith(('drivers/', 'examples/', 'prototypes/', 'docs/'))) or p in ['go.mod', 'go.sum', 'drivers/cogent/env.sh'])
inputs = lambda: {p: sha(root / p) for p in paths}
before = inputs()
proof = {'base': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root, text=True).strip(),
         'scope': 'Selected core-module Go files, module manifests and native environment wrapper; no native GUI or dependency source closure claim.',
         'source_inputs': before, 'checks': []}
prefix = 'export VELLUM_CACHE_DIR=/tmp/vellum-p0-cache\n. drivers/cogent/env.sh\n'
for name, command in [('portable-race3', 'go test -race -count=3 -v ./virtual'),
                      ('core-race', 'go test -race -count=1 ./...'),
                      ('core-vet', 'go vet ./...')]:
    log = logs / (name + '.log')
    start = time.monotonic()
    with log.open('w') as stream:
        argv = ['/bin/zsh', '-c', prefix + command]
        failure = None
        try:
            result = run_child(argv, cwd=root, stdout=stream, stderr=subprocess.STDOUT, timeout=300)
            code = result.returncode
        except Exception as error:
            code = 124 if isinstance(error, subprocess.TimeoutExpired) else 125
            failure = {'type': type(error).__name__, 'message': str(error)}
    check = {'name': name, 'argv': argv, 'exit_code': code, 'failure': failure,
             'elapsed_seconds': round(time.monotonic() - start, 3),
             'log': str(log), 'log_sha256': sha(log)}
    proof['checks'].append(check)
    print(json.dumps(check), flush=True)
    if code:
        break
proof['unchanged'] = before == inputs()
proof['result'] = 'PASS' if proof['unchanged'] and len(proof['checks']) == 3 and all(c['exit_code'] == 0 for c in proof['checks']) else 'FAIL'
path.write_text(json.dumps(proof, indent=2) + '\n')
print(proof['result'], sha(path), flush=True)
raise SystemExit(0 if proof['result'] == 'PASS' else 1)
