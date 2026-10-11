"""Run the qualified consumer binary in actual Cocoa selection smoke modes."""
from pathlib import Path
import datetime
import hashlib
import json
import os
import subprocess
import sys
import time

sys.dont_write_bytecode = True
root = Path(__file__).resolve().parents[3]
out = Path(sys.argv[1]).resolve() if len(sys.argv) >= 2 else root / 'docs/reports/macos-grid-selection'
out.mkdir(parents=True, exist_ok=True)
(out / 'logs').mkdir(exist_ok=True)
path = out / 'desktop-qualification.json'
if path.exists() or any((out / 'logs' / (name + '.log')).exists() for name in ['fixed-selection', 'variable-selection']):
    raise SystemExit('Refusing to overwrite prior evidence; supply a fresh output directory')
with (out / '.desktop-qualification-started').open('x') as marker:
    marker.write('Exclusive evidence directory; never reuse after interruption.\n')
sys.path.insert(0, str(root))
from scripts.child_process import run_child

sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
producer = Path(sys.argv[2]).resolve() if len(sys.argv) == 3 else Path('/private/tmp/vellum-grid-selection-example/qualified-v3/qualification.json')
qualification = json.loads(producer.read_text())
binary = Path(qualification['binary']['path'])
expected = {(str(Path('examples/grid') / p) if group in ['owned', 'preserved_example'] else p): h
            for group, entries in qualification['before'].items() for p, h in entries.items()}
sources = lambda: {p: sha(root / p) for p in sorted(expected)}

def prepared():
    result = []
    for name in ['upstream', 'typesetting']:
        directory = (root / 'drivers/cogent/build' / name).resolve()
        values = {str(p.relative_to(directory)): sha(p) for p in sorted(directory.rglob('*'))
                  if p.is_file() and not p.is_symlink() and '.git' not in p.parts and '__pycache__' not in p.parts}
        result.append({'resolved_root': str(directory), 'files': len(values),
                       'sorted_relative_path_hashes_sha256': hashlib.sha256(json.dumps(values, sort_keys=True, separators=(',', ':')).encode()).hexdigest()})
    return result

before = sources()
dependencies = prepared()
if qualification['result'] != 'PASS' or sha(binary) != qualification['binary']['sha256'] or before != expected:
    raise SystemExit('Desktop qualification refused: producer/source/binary mismatch')
proof = {'created_utc': datetime.datetime.now(datetime.timezone.utc).isoformat(),
         'scope': 'Actual Cocoa completed frames, explicit owner selection/projection commands and independent window lifetimes; no physical input, external AX, compositor teardown or performance claim.',
         'producer_sha256': sha(producer), 'binary': qualification['binary'],
         'source_inputs': before, 'prepared_before': dependencies, 'checks': []}
for name, args in [('fixed-selection', []), ('variable-selection', ['-variable-rows'])]:
    stages = ['ready', 'far-both-axes', 'immutable-projection']
    if args:
        stages.append('variable-pixel-page')
    stages += ['captured-row-selection', 'pass']
    argv = [str(binary), '-smoke', '-row-selection'] + args
    log = out / 'logs' / (name + '.log')
    start = time.monotonic()
    with log.open('w') as stream:
        failure = None
        try:
            process = run_child(argv, cwd=root / 'examples/grid', env=dict(os.environ, PYTHONDONTWRITEBYTECODE='1'),
                                stdout=stream, stderr=subprocess.STDOUT, timeout=60, text=True)
            code = process.returncode
        except Exception as error:
            code = 124 if isinstance(error, subprocess.TimeoutExpired) else 125
            failure = {'type': type(error).__name__, 'message': str(error)}
    records = []
    for line in log.read_text().splitlines():
        try:
            item = json.loads(line)
        except json.JSONDecodeError:
            continue
        if isinstance(item, dict) and 'stage' in item:
            records.append(item)
    observed = [r['stage'] for r in records]
    check = {'name': name, 'argv': argv, 'exit_code': code, 'failure': failure,
             'elapsed_seconds': round(time.monotonic() - start, 3),
             'log': str(log), 'log_sha256': sha(log),
             'expected_stages': stages, 'stages': observed, 'records': records}
    proof['checks'].append(check)
    print(json.dumps({k: check[k] for k in ['name', 'exit_code', 'stages']}), flush=True)
    if code or stages != observed:
        break
proof['source_unchanged'] = sources() == before
proof['prepared_after'] = prepared()
proof['binary_unchanged'] = sha(binary) == qualification['binary']['sha256']
proof['result'] = 'PASS' if len(proof['checks']) == 2 and all(c['exit_code'] == 0 and c['stages'] == c['expected_stages'] for c in proof['checks']) and proof['source_unchanged'] and proof['binary_unchanged'] and dependencies == proof['prepared_after'] else 'FAIL'
path.write_text(json.dumps(proof, indent=2) + '\n')
print(proof['result'], sha(path), flush=True)
raise SystemExit(0 if proof['result'] == 'PASS' else 1)
