#!/usr/bin/env python3
"""Prepare an isolated offscreen fixture; no shipping source/native mutation."""
from pathlib import Path
import argparse,shutil,json
p=argparse.ArgumentParser();p.add_argument('--sdk',type=Path,required=True);p.add_argument('--core',type=Path,required=True);p.add_argument('--artifacts',type=Path,required=True);args=p.parse_args()
code=Path(__file__).resolve().parent;repo=code.parents[2];f=repo/'tools/macos-calibration/stagingreuse/framework';sdk=args.sdk.resolve();core=args.core.resolve();q=args.artifacts.resolve()
if core.exists() or q.exists():raise SystemExit('refuse to replace existing Core/artifacts')
source=sdk/'drivers/cogent/build/upstream'
if not (source/'paint/renderers/rasterx/renderer.go').is_file():raise SystemExit('prepared SDK Core required')
shutil.copytree(source,core);q.mkdir(parents=True)
s=(f/'fixture.go').read_text();needle='Width: 1500, Height: 1250';assert s.count(needle)==1;(q/'fixture.go').write_text(s.replace(needle,'Width: 1519, Height: 1266'))
shutil.copy2(code/'testdata/replay_test.go',q/'replay_test.go')
mod=(f/'go.mod').read_text().split('\nreplace (',1)[0]+'\nreplace (\n'
for name,path in [('github.com/dshills/vellum',sdk),('github.com/dshills/vellum/drivers/cogent',sdk/'drivers/cogent'),('cogentcore.org/core',core),('github.com/go-text/typesetting',sdk/'drivers/cogent/build/typesetting'),('github.com/oliverbestmann/webgpu/libs-darwin',sdk/'drivers/cogent/build/libs-darwin')]:mod+=name+' => '+json.dumps(str(path))+'\n'
(q/'framework.mod').write_text(mod+')\n');shutil.copy2(f/'go.sum',q/'framework.sum')
(q/'overlay.json').write_text(json.dumps({'Replace':{str(f/'fixture.go'):str(q/'fixture.go'),str(f/'fixture_test.go'):str(q/'replay_test.go')}},indent=2)+'\n')
print(q)
