#!/usr/bin/env python3
"""Instrument only the exact locked private wgpu-core clone, retaining allocation semantics."""
from pathlib import Path
import hashlib,json
root=Path('/private/tmp/vellum-upload-probe');code=Path(__file__).resolve().parent
q=root/'core/src/device/queue.rs';r=root/'core/src/resource.rs';audit=Path('/Users/dshills/Documents/ChatGPT/Go Native UI/artifacts/macos-upload-allocation-plan/source-audit.json');pins=json.loads(audit.read_text())['members_equal_registry_archive']
for p in [q,r]:expected=next(h for name,h in pins.items() if name.endswith('/'+str(p.relative_to(root/'core'))));assert hashlib.sha256(p.read_bytes()).hexdigest()==expected
s=q.read_text();start=s.index('    pub fn write_texture(');end=s.index('    #[cfg(webgl)]\n    pub fn copy_external_image_to_texture',start);body=s[start:end]
needle='        profiling::scope!("Queue::write_texture");';assert body.count(needle)==1;body=body.replace(needle,needle+'\n        let mut probe = crate::upload_probe::Probe::begin(data.len(), size.width, size.height);')
assert body.count('StagingBuffer::new(&self.device, stage_size)?')==2;body=body.replace('StagingBuffer::new(&self.device, stage_size)?','StagingBuffer::new_with_upload_probe(&self.device, stage_size, Some(&probe))?')
body=body.replace('            staging_buffer.write(&data[data_layout.offset as usize..]);','            let copy_start = probe.clock();\n            staging_buffer.write(&data[data_layout.offset as usize..]);\n            probe.copy_done(copy_start);')
body=body.replace('            for layer in 0..size.depth_or_array_layers {','            let copy_start = probe.clock();\n            for layer in 0..size.depth_or_array_layers {');needle='            staging_buffer\n        };';assert body.count(needle)==1;body=body.replace(needle,'            probe.copy_done(copy_start);\n'+needle)
body=body.replace('        let staging_buffer = staging_buffer.flush();','        probe.pitch(stage_bytes_per_row);\n        let flush_start = probe.clock();\n        let staging_buffer = staging_buffer.flush();\n        probe.flush_done(flush_start);\n        let encode_start = probe.clock();')
body=body.replace('            return Ok(());','            probe.success();\n            return Ok(());');needle='        Ok(())\n    }';assert body.count(needle)==1;body=body.replace(needle,'        probe.encode_done(encode_start);\n        probe.success();\n'+needle);q.write_text(s[:start]+body+s[end:])
s=r.read_text();needle='    pub(crate) fn new(device: &Arc<Device>, size: wgt::BufferSize) -> Result<Self, DeviceError> {\n';assert s.count(needle)==1;s=s.replace(needle,needle+'        Self::new_with_upload_probe(device, size, None)\n    }\n\n    pub(crate) fn new_with_upload_probe(device: &Arc<Device>, size: wgt::BufferSize, probe: Option<&crate::upload_probe::Probe>) -> Result<Self, DeviceError> {\n');s=s.replace('        let raw = unsafe { device.raw().create_buffer(&stage_desc) }','        let create_start = probe.map_or(0.0, |p| p.clock());\n        let raw = unsafe { device.raw().create_buffer(&stage_desc) }');needle='        let mapping = unsafe { device.raw().map_buffer(raw.as_ref(), 0..size.get()) }';assert s.count(needle)==1;s=s.replace(needle,'        if let Some(p) = probe { p.create_done(create_start); }\n        let map_start = probe.map_or(0.0, |p| p.clock());\n'+needle);needle='        let staging_buffer = StagingBuffer {';assert s.count(needle)==1;s=s.replace(needle,'        if let Some(p) = probe { p.map_done(map_start); }\n\n'+needle);r.write_text(s)
p=root/'core/src/lib.rs';s=p.read_text();assert 'pub mod upload_probe;' not in s;p.write_text(s+'\n/// Private diagnostic module; never installed in the SDK.\npub mod upload_probe;\n');(root/'core/src/upload_probe.rs').write_bytes((code/'probe.rs').read_bytes())
p=root/'source/Cargo.toml';s=p.read_text();assert '[patch.crates-io]' not in s;p.write_text(s+'\n[patch.crates-io]\nwgpu-core = { path = "../core" }\n')
p=root/'source/src/lib.rs';s=p.read_text();p.write_text(s+'''
// Private CPU upload observer, not a shipping API.
#[no_mangle]
pub extern "C" fn vellumUploadConfigure(enabled: bool) -> bool { wgc::upload_probe::configure(enabled) }
#[no_mangle]
pub unsafe extern "C" fn vellumUploadSnapshot(out: *mut u8, capacity: usize) -> usize {
    let text = wgc::upload_probe::snapshot();
    if text.len() < capacity && !out.is_null() {
        core::ptr::copy_nonoverlapping(text.as_ptr(), out, text.len());
        *out.add(text.len()) = 0;
    }
    text.len()
}
''')
print('Private upload probe applied; staging allocation/copy/flush/retention order preserved; original SDK/build inputs untouched.')
