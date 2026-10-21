#!/usr/bin/env python3
"""Apply cache-mode diagnostic only to a fresh clone of the pinned upload probe."""
from pathlib import Path
import hashlib, json
root = Path('/private/tmp/vellum-cache-control')
code = Path(__file__).resolve().parent
prior = json.loads(Path('/Users/dshills/Documents/ChatGPT/Go Native UI/artifacts/macos-upload-probe/build-identity.json').read_text())
for old, digest in prior['private_source_pins'].items():
    p = root / Path(old).relative_to('/private/tmp/vellum-upload-probe')
    assert hashlib.sha256(p.read_bytes()).hexdigest() == digest, str(p)
p = root/'hal/src/metal/device.rs'
s = p.read_text()
old = '        options.set(MTLResourceOptions::CPUCacheModeWriteCombined, map_write);'
assert s.count(old) == 1
s = s.replace(old, '''        // Only exact transient upload staging buffers participate. All other
        // resources preserve their original descriptor and cache selection.
        let controlled = desc.usage == (wgt::BufferUses::MAP_WRITE | wgt::BufferUses::COPY_SRC)
            && desc.memory_flags == crate::MemoryFlags::TRANSIENT;
        options.set(MTLResourceOptions::CPUCacheModeWriteCombined,
            map_write && !(controlled && super::vellum_cache_control::default_cache()));''')
old = '            self.counters.buffers.add(1);'
assert s.count(old) == 1  # Exact indentation selects only creation, not add_raw_buffer.
s = s.replace(old, '''            if controlled {
                super::vellum_cache_control::record(raw.cpuCacheMode().0 as u64, raw.storageMode().0 as u64);
            }
''' + old, 1)
p.write_text(s)
p = root/'hal/src/metal/mod.rs';s = p.read_text();assert 'vellum_cache_control' not in s
p.write_text(s+'\n/// Private cache-policy control, not part of the SDK.\npub mod vellum_cache_control;\n')
(root/'hal/src/metal/vellum_cache_control.rs').write_bytes((code/'control.rs').read_bytes())
p = root/'source/Cargo.toml';s = p.read_text();assert '[patch.crates-io]' in s
p.write_text(s+'wgpu-hal = { path = "../hal" }\n')
p = root/'source/src/lib.rs';s = p.read_text();assert 'vellumCacheConfigure' not in s
p.write_text(s+'''
// Private diagnostic ABI: immutable mode, scalar counters, no retained resource.
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
''')
print('Private exact transient upload cache control applied; all other descriptors and ownership unchanged.')
