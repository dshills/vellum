#!/usr/bin/env python3
"""Texture-only private cache experiment; apply once to a fresh reuse-proof clone."""
from pathlib import Path
import argparse,json,hashlib
p=argparse.ArgumentParser();p.add_argument('--root',type=Path,required=True);p.add_argument('--preimages',type=Path,required=True);a=p.parse_args();root=a.root.resolve();code=Path(__file__).resolve().parent
assert root==Path('/private/tmp/vellum-staging-cache'),'private owned root only'
pre=json.loads(a.preimages.read_text())
for group in ['reuse_private','HAL_original_members']:
 for name,h in pre[group].items():assert hashlib.sha256(Path(name).read_bytes()).hexdigest()==h,name
def change(path,old,new):
 s=path.read_text();assert s.count(old)==1,(str(path),old);path.write_text(s.replace(old,new))
# Explicit private descriptor bit, used only by the already-allowlisted Metal path.
change(root/'hal/src/lib.rs','        const PREFER_COHERENT = 1 << 1;','        const PREFER_COHERENT = 1 << 1;\n        const VELLUM_TEXTURE_UPLOAD = 1 << 31;')
change(root/'core/src/resource.rs','        let stage_desc = hal::BufferDescriptor {','        let mut stage_desc = hal::BufferDescriptor {')
change(root/'core/src/resource.rs','        let owner = device.texture_staging_owner;\n        let call = probe.and_then',
       '        if tracked { stage_desc.memory_flags.insert(hal::MemoryFlags::VELLUM_TEXTURE_UPLOAD); }\n        let owner = device.texture_staging_owner;\n        let call = probe.and_then')
change(root/'hal/src/metal/device.rs','        options.set(MTLResourceOptions::CPUCacheModeWriteCombined, map_write);',"""        let marked = desc.memory_flags.contains(crate::MemoryFlags::VELLUM_TEXTURE_UPLOAD);
        let controlled = marked
            && desc.usage == (wgt::BufferUses::MAP_WRITE | wgt::BufferUses::COPY_SRC)
            && desc.memory_flags == (crate::MemoryFlags::TRANSIENT | crate::MemoryFlags::VELLUM_TEXTURE_UPLOAD);
        options.set(MTLResourceOptions::CPUCacheModeWriteCombined,
            map_write && !(controlled && super::vellum_cache_control::default_cache()));""")
change(root/'hal/src/metal/device.rs','            self.counters.buffers.add(1);',"""            if marked {
                // A malformed marker must be observable as a refusal, never silently broaden scope.
                super::vellum_cache_control::record(
                    if controlled { raw.cpuCacheMode().0 as u64 } else { u64::MAX },
                    raw.storageMode().0 as u64);
            }
            self.counters.buffers.add(1);""")
with (root/'hal/src/metal/mod.rs').open('a') as f:f.write('\n/// Private texture-upload cache diagnostic.\npub mod vellum_cache_control;\n')
(root/'hal/src/metal/vellum_cache_control.rs').write_bytes((code/'control.rs').read_bytes())
with (root/'source/Cargo.toml').open('a') as f:f.write('wgpu-hal = { path = "../hal" }\n')
with (root/'source/src/lib.rs').open('a') as f:f.write("""
// Private one-shot cache configuration before device creation; no handles.
#[no_mangle]
pub extern "C" fn vellumCacheConfigure(default_cache: bool) -> bool {
    hal::metal::vellum_cache_control::configure(default_cache)
}
#[no_mangle]
pub unsafe extern "C" fn vellumCacheSnapshot(out: *mut u64, capacity: usize) -> usize {
    let values = hal::metal::vellum_cache_control::snapshot();
    if !out.is_null() && capacity >= values.len() {
        core::ptr::copy_nonoverlapping(values.as_ptr(), out, values.len());
    }
    values.len()
}
""")
print('Private exact-texture marker/cache patch applied; original pool/lifetime/other descriptors unchanged.')
