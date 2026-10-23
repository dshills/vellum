#!/usr/bin/env python3
"""Assemble a private physical consumer; never edit/rebuild native library or SDK."""
from pathlib import Path
import argparse,json,hashlib,shutil

def prepare(repo,prior,q):
 repo,prior,q=[p.resolve() for p in [repo,prior,q]];code=repo/'tools/macos-calibration/displayclock';source=repo/'tools/macos-calibration/upload/framework';identity=json.loads((prior/'build-identity.json').read_text());sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
 if q.exists():raise ValueError('refuse existing artifacts')
 if not all(sha(Path(k))==v for k,v in identity['pins'].items()):raise ValueError('reviewed joined backend input drift')
 q.mkdir();consumer=q/'consumer';consumer.mkdir()
 for p in source.iterdir():
  if p.is_file():shutil.copy2(p,consumer/p.name)
 for name in ['fixture_test.go','backend_probe_darwin.go']:shutil.copy2(prior/name,consumer/name)
 main=(repo/'tools/macos-calibration/joinedbackend/testdata/native_main_darwin.go').read_text()
 main=main.replace('len(os.Args) != 4','len(os.Args) != 5').replace('pilot|timed MATCH.json on','pilot|timed MATCH.json on off|on')
 needle='	if !rasterx.VellumRasterConfigure(desktop.PresentationHostTime) {'
 assert main.count(needle)==1
 main=main.replace(needle,'	if os.Args[4] != "off" && os.Args[4] != "on" { return errors.New("display clock off|on") }\n	if err := displayConfigure(os.Args[4] == "on"); err != nil { return err }\n'+needle)
 needle='			out["end"] = observe(2)';assert main.count(needle)==1
 main=main.replace(needle,'			clock, clockErr := displayFinish()\n			out["display_clock_mode"] = os.Args[4]\n			out["display_clock_trace"] = clock\n			failure = errors.Join(failure, clockErr)\n'+needle)
 needle='		var rect [4]int';assert main.count(needle)==1;main=main.replace(needle,'		if err := displayStart(); err != nil { return err }\n'+needle)
 needle='		return nil\n	}\n	go func()';assert main.count(needle)==1;main=main.replace(needle,'		time.Sleep(50 * time.Millisecond) // identical held tail, outside cohort\n'+needle)
 (consumer/'main_darwin.go').write_text(main)
 for p in (code/'testdata').iterdir():
  if p.name.startswith('display_clock'):shutil.copy2(p,consumer/p.name)
 p=consumer/'observe_darwin.m';p.write_text('//go:build darwin && !offscreen\n#include '+json.dumps(str(repo/'tools/macos-calibration/observe.m'))+'\n')
 p=consumer/'observe_darwin.go';p.write_text(p.read_text().replace('#cgo CFLAGS: -I../..','#cgo CFLAGS: -I'+str(repo/'tools/macos-calibration')+' -mmacosx-version-min=14.0'))
 for name in ['framework.mod','framework.sum','core.mod','core.sum','desktop_session.go','desktop_measurement.go','frame_trace.go','match.json']:shutil.copy2(prior/name,q/name)
 overlay=json.loads((prior/'overlay.json').read_text());overlay['Replace']={k:str(q/Path(v).name) for k,v in overlay['Replace'].items() if k.endswith('/desktop_session.go') or k.endswith('/desktop_measurement.go') or k.endswith('/core/frame_trace.go')};(q/'overlay.json').write_text(json.dumps(overlay,indent=2)+'\n')
 (q/'reused-build-identity.json').write_text(json.dumps(identity,indent=2)+'\n');return q
if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--prior',type=Path,required=True);p.add_argument('--artifacts',type=Path,required=True);a=p.parse_args();print(prepare(Path(__file__).resolve().parents[3],a.prior,a.artifacts))
