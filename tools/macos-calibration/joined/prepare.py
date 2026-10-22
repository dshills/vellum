#!/usr/bin/env python3
"""Prepare private same-input timing instrumentation; never edits the source SDK."""
from pathlib import Path
import argparse,hashlib,json,shutil

def prepare(sdk,core,q,upload):
    code=Path(__file__).resolve().parent;repo=code.parents[2];f=repo/'tools/macos-calibration/upload/framework'
    sdk=sdk.resolve();core=core.resolve();q=q.resolve();upload=upload.resolve()
    if core.exists() or q.exists():raise ValueError('refuse to overwrite Core/artifacts')
    identity=json.loads((repo/'docs/reports/macos-upload-probe/build-identity.json').read_text())
    sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
    root=Path(identity['root']);lib=root/'libs-darwin/arm64/libwgpu_native.a'
    if sha(lib)!=identity['archive_sha256'] or not all(sha(Path(k))==v for k,v in identity['private_source_pins'].items()):raise ValueError('reviewed native observer source/archive drift')
    for name in ['desktop_session.go','desktop_measurement.go','frame_trace.go']:
        expected=identity['overlay_files_sha256']['artifacts/macos-upload-probe/'+name]
        if sha(upload/name)!=expected:raise ValueError('reviewed Go source-detail overlay drift: '+name)
    shutil.copytree(sdk/'drivers/cogent/build/upstream',core);q.mkdir(parents=True)
    renderer=core/'paint/renderers/rasterx/renderer.go';s=renderer.read_text();needle='func (rs *Renderer) Render(r render.Render) render.Renderer {'
    if s.count(needle)!=1:raise ValueError('unexpected renderer entry')
    s=s.replace(needle,needle+'\n\ttoken := vellumRasterBegin(rs.image)\n\treturned := false\n\tdefer func() { vellumRasterEnd(token, returned) }()')
    needle='\trs.scrollStats.DescriptorBytes = rs.scrollPlanBytes\n\treturn rs'
    if s.count(needle)!=1:raise ValueError('unexpected renderer return')
    renderer.write_text(s.replace(needle,'\trs.scrollStats.DescriptorBytes = rs.scrollPlanBytes\n\treturned = true\n\treturn rs'))
    for name in ['raster_clock.go','raster_clock_test.go']:shutil.copy2(code/'testdata'/name,core/'paint/renderers/rasterx'/('vellum_'+name))
    for name in ['native_main_darwin.go','fixture_test.go']:shutil.copy2(code/'testdata'/name,q/name)
    for name in ['desktop_session.go','desktop_measurement.go','frame_trace.go']:shutil.copy2(upload/name,q/name)
    mod=(f/'go.mod').read_text().split('\nreplace (',1)[0]+'\nreplace (\n'
    for name,path in [('github.com/dshills/vellum',sdk),('github.com/dshills/vellum/drivers/cogent',sdk/'drivers/cogent'),('cogentcore.org/core',core),('github.com/go-text/typesetting',sdk/'drivers/cogent/build/typesetting'),('github.com/oliverbestmann/webgpu/libs-darwin',root/'libs-darwin')]:mod+=name+' => '+json.dumps(str(path))+'\n'
    (q/'framework.mod').write_text(mod+')\n');shutil.copy2(f/'go.sum',q/'framework.sum')
    mod=(core/'go.mod').read_text();needle='replace github.com/go-text/typesetting => ../typesetting'
    if mod.count(needle)!=1:raise ValueError('unexpected typesetting replacement')
    (q/'core.mod').write_text(mod.replace(needle,'replace github.com/go-text/typesetting => '+json.dumps(str(sdk/'drivers/cogent/build/typesetting'))));shutil.copy2(core/'go.sum',q/'core.sum')
    (q/'overlay.json').write_text(json.dumps({'Replace':{str(sdk/'drivers/cogent/desktop_session.go'):str(q/'desktop_session.go'),str(sdk/'drivers/cogent/desktop_measurement.go'):str(q/'desktop_measurement.go'),str(core/'core/frame_trace.go'):str(q/'frame_trace.go'),str(f/'main_darwin.go'):str(q/'native_main_darwin.go'),str(f/'fixture_test.go'):str(q/'fixture_test.go')}},indent=2)+'\n')
    return q
if __name__=='__main__':
    p=argparse.ArgumentParser();p.add_argument('--sdk',type=Path,required=True);p.add_argument('--core',type=Path,required=True);p.add_argument('--artifacts',type=Path,required=True);p.add_argument('--reviewed-upload-overlays',type=Path,required=True);args=p.parse_args();print(prepare(args.sdk,args.core,args.artifacts,args.reviewed_upload_overlays))
