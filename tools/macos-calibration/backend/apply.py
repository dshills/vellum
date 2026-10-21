#!/usr/bin/env python3
"""Patch only a private cloned backend; refuse unknown input source."""
from pathlib import Path
import hashlib,json,sys
root=Path(sys.argv[1]).resolve();code=Path(__file__).resolve().parent
assert root==Path('/private/tmp/vellum-backend-timing') and (root/'hal').is_dir()
p=root/'hal/src/metal/mod.rs';original=p.read_bytes();assert hashlib.sha256(original).hexdigest()=='16aaa879cf85bf853243bbdaa815b6087cff2ebfe72a2d7e6d08acac4fe906ad'
s=original.decode();assert s.count('cmd_buffer.raw.commit();')==1 and s.count('raw.commit();')==2
s=s.replace('mod time;','mod time;\npub mod probe;').replace('cmd_buffer.raw.commit();','probe::commit(&cmd_buffer.raw, 0);').replace('                raw.commit();','                probe::commit(&raw, 1);')
a='''            if !texture.present_with_transaction {
                command_buffer.presentDrawable(&texture.drawable);
            }

            command_buffer.commit();'''
assert s.count(a)==1;s=s.replace(a,'            probe::present(&command_buffer, &texture.drawable, texture.present_with_transaction);')
p.write_text(s);(p.parent/'probe.rs').write_bytes((code/'probe.rs').read_bytes())
p=root/'source/Cargo.toml';s=p.read_text();assert '[patch.crates-io]' not in s;p.write_text(s+'\n[patch.crates-io]\nwgpu-hal = { path = "../hal" }\n')
p=root/'source/src/lib.rs';s=p.read_text();s+='''
// Private diagnostic ABI. Not installed into the SDK or public framework.
#[no_mangle]
pub extern "C" fn vellumProbeConfigure(enabled: bool) -> bool { hal::metal::probe::configure(enabled) }
#[no_mangle]
pub unsafe extern "C" fn vellumProbeSnapshot(out: *mut u8, capacity: usize) -> usize {
    let text = hal::metal::probe::snapshot();
    if text.len() < capacity && !out.is_null() {
        core::ptr::copy_nonoverlapping(text.as_ptr(), out, text.len());
        *out.add(text.len()) = 0;
    }
    text.len()
}
''';p.write_text(s)
print(json.dumps({'private_root':str(root),'only_backend_patch':True,'SDK_changed':False,'original_hal_SHA256':hashlib.sha256(original).hexdigest(),'patched_hal_SHA256':hashlib.sha256((root/'hal/src/metal/mod.rs').read_bytes()).hexdigest(),'probe_SHA256':hashlib.sha256((code/'probe.rs').read_bytes()).hexdigest()}))
