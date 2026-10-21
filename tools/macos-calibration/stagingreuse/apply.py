#!/usr/bin/env python3
"""Private prototype only. Exact pinned upload-observer preimages; never patch SDK."""
from pathlib import Path
import json, hashlib
root = Path('/private/tmp/vellum-staging-reuse')
code = Path(__file__).resolve().parent
pins = json.loads(Path('/Users/dshills/Documents/ChatGPT/Go Native UI/artifacts/macos-upload-probe/build-identity.json').read_text())['private_source_pins']
for old, digest in pins.items():
    p = root/Path(old).relative_to('/private/tmp/vellum-upload-probe')
    assert hashlib.sha256(p.read_bytes()).hexdigest() == digest, str(p)
def replace(s, old, new):
    assert s.count(old) == 1, (old, s.count(old))
    return s.replace(old, new)

p = root/'core/src/resource.rs';s = p.read_text()
start = s.index('pub struct StagingBuffer {');end = s.index('pub type TextureDescriptor', start);body = s[start:end]
body = replace(body, 'Self::new_with_upload_probe(device, size, None)', 'Self::new_with_upload_probe(device, size, None, false)')
body = replace(body, 'probe: Option<&crate::upload_probe::Probe>) -> Result<Self, DeviceError>', 'probe: Option<&crate::upload_probe::Probe>, allow_reuse: bool) -> Result<Self, DeviceError>')
body = replace(body, '    is_coherent: bool,', '    is_coherent: bool,\n    reuse_id: u64,')
old = '''        let raw = unsafe { device.raw().create_buffer(&stage_desc) }
            .map_err(|e| device.handle_hal_error(e))?;'''
new = '''        let tracked = allow_reuse && probe.is_some() && device.adapter.backend() == wgt::Backend::Metal;
        let owner = device.texture_staging_owner;
        let call = probe.and_then(|p| p.ordinal()).map_or(0, |i| i as u64 + 1);
        let cached = if tracked && crate::staging_reuse::enabled() {
            let mut pool = device.texture_staging_pool.lock().expect("texture staging cache");
            let hit = pool.take_exact(size);
            if let Some(ref upload) = hit {
                crate::staging_reuse::record(crate::staging_reuse::TAKE, owner,
                    upload.id(), upload.bytes(), 0, 0, call, pool.len(), pool.bytes());
            }
            hit
        } else { None };
        let (raw, reuse_id) = if let Some(upload) = cached {
            let id = upload.id(); (upload.into_resource(), id)
        } else {
            let raw = unsafe { device.raw().create_buffer(&stage_desc) }
                .map_err(|e| device.handle_hal_error(e))?;
            let id = if tracked { crate::staging_reuse::fresh(owner, size.get(), call) } else { 0 };
            (raw, id)
        };'''
body = replace(body, old, new)
body = replace(body, '''        let mapping = unsafe { device.raw().map_buffer(raw.as_ref(), 0..size.get()) }
            .map_err(|e| device.handle_hal_error(e))?;''', '''        let mapping = match unsafe { device.raw().map_buffer(raw.as_ref(), 0..size.get()) } {
            Ok(mapping) => mapping,
            Err(error) => {
                // No GPU can use a fresh or completed/check-out buffer here.
                unsafe { device.raw().destroy_buffer(raw) };
                if reuse_id != 0 {
                    crate::staging_reuse::record(crate::staging_reuse::DESTROY, owner,
                        reuse_id, size.get(), 0, 0, 0, 0, 0);
                }
                return Err(device.handle_hal_error(error));
            }
        };''')
body = replace(body, '            is_coherent: mapping.is_coherent,', '            is_coherent: mapping.is_coherent,\n            reuse_id,')
body = replace(body, '            raw, device, size, ..', '            raw, device, size, reuse_id, ..')
body = replace(body, '''            raw: ManuallyDrop::new(raw),
            device,
            size,''', '''            raw: ManuallyDrop::new(raw),
            device,
            size,
            reuse_id,
            submitted: 0,
            raw_taken: false,''')
body = replace(body, '''pub struct FlushedStagingBuffer {
    raw: ManuallyDrop<Box<dyn hal::DynBuffer>>,''', '''pub struct FlushedStagingBuffer {
    raw: ManuallyDrop<Box<dyn hal::DynBuffer>>,
    reuse_id: u64,
    submitted: u64,
    raw_taken: bool,''')
needle = 'impl FlushedStagingBuffer {\n'
body = replace(body, needle, needle+'''
    pub(crate) fn mark_reuse_submission(&mut self, index: u64) {
        if self.reuse_id == 0 { return; }
        if self.submitted != 0 || index == 0 { crate::staging_reuse::invalid(); return; }
        self.submitted = index;
        crate::staging_reuse::record(crate::staging_reuse::SUBMIT,
            self.device.texture_staging_owner, self.reuse_id, self.size.get(), index, 0, 0, 0, 0);
    }
    // Called only from drained, completed submissions, after retiring encoder.
    pub(crate) fn recycle_completed(mut self, index: u64, fence: u64) {
        if self.reuse_id == 0 { return; }
        if self.submitted != index || index == 0 || fence < index {
            crate::staging_reuse::invalid(); return;
        }
        if !self.device.is_valid() { return; }
        let owner = self.device.texture_staging_owner;
        crate::staging_reuse::record(crate::staging_reuse::RETIRE, owner,
            self.reuse_id, self.size.get(), index, fence, 0, 0, 0);
        if !crate::staging_reuse::enabled() || self.size.get() > 32 * 1024 * 1024 { return; }
        let identity = NonZeroU64::new(self.reuse_id).expect("resource id already checked");
        let submission = NonZeroU64::new(index).expect("submission already checked");
        // SAFETY: sole raw owner, GPU completed above, no later use of self.raw.
        let raw = unsafe { ManuallyDrop::take(&mut self.raw) };
        self.raw_taken = true; // No fallible/panicking operation between take and flag.
        let upload = crate::staging_reuse_policy::Upload::new(identity, self.size, raw);
        let completed = match upload.submit(submission).complete(fence) {
            Ok(completed) => completed,
            Err(pending) => {
                // Defensive invariant failure: original completed-resource destruction.
                unsafe { self.device.raw().destroy_buffer(pending.into_resource()) };
                crate::staging_reuse::record(crate::staging_reuse::DESTROY, owner,
                    self.reuse_id, self.size.get(), index, 0, 0, 0, 0);
                crate::staging_reuse::invalid(); return;
            }
        };
        let evicted = {
            let mut pool = self.device.texture_staging_pool.lock().expect("texture staging cache");
            let evicted = pool.insert(completed);
            for old in &evicted {
                // Metadata access only; ownership remains in evicted until backend destruction.
                crate::staging_reuse::record(crate::staging_reuse::EVICT, owner,
                    old.id(), old.bytes(), 0, 0, 0, 0, 0);
            }
            crate::staging_reuse::record(crate::staging_reuse::CACHE, owner,
                self.reuse_id, self.size.get(), index, fence, 0, pool.len(), pool.bytes());
            evicted
        };
        for old in evicted {
            let upload = old.into_upload(); let id = upload.id(); let bytes = upload.bytes();
            unsafe { self.device.raw().destroy_buffer(upload.into_resource()) };
            crate::staging_reuse::record(crate::staging_reuse::DESTROY, owner, id, bytes, 0, 0, 0, 0, 0);
        }
    }
''')
body = replace(body, '''        resource_log!("Destroy raw StagingBuffer");''', '''        if self.raw_taken { return; }
        resource_log!("Destroy raw StagingBuffer");''')
body = replace(body, '''        unsafe { self.device.raw().destroy_buffer(raw) };''', '''        unsafe { self.device.raw().destroy_buffer(raw) };
        if self.reuse_id != 0 {
            crate::staging_reuse::record(crate::staging_reuse::DESTROY,
                self.device.texture_staging_owner, self.reuse_id, self.size.get(), self.submitted, 0, 0, 0, 0);
        }''')
p.write_text(s[:start]+body+s[end:])

p = root/'core/src/device/resource.rs';s=p.read_text()
s=replace(s,'pub struct Device {','''pub struct Device {
    pub(crate) texture_staging_pool: std::sync::Mutex<crate::staging_reuse_policy::Pool<Box<dyn hal::DynBuffer>>>,
    pub(crate) texture_staging_owner: u64,''')
s=replace(s,'            raw: raw_device,','''            raw: raw_device,
            texture_staging_pool: std::sync::Mutex::new(crate::staging_reuse_policy::Pool::new(8, 32 * 1024 * 1024)),
            texture_staging_owner: crate::staging_reuse::owner_created(),''')
s=replace(s,'        resource_log!("Drop {}", self.error_ident());','''        resource_log!("Drop {}", self.error_ident());
        let idle = self.texture_staging_pool.get_mut().expect("texture staging cache").drain();
        crate::staging_reuse::record(crate::staging_reuse::DRAIN,
            self.texture_staging_owner, 0, 0, 0, 0, 0, 0, 0);
        for old in idle {
            let upload = old.into_upload(); let id = upload.id(); let bytes = upload.bytes();
            unsafe { self.raw.destroy_buffer(upload.into_resource()) };
            crate::staging_reuse::record(crate::staging_reuse::DESTROY,
                self.texture_staging_owner, id, bytes, 0, 0, 0, 0, 0);
        }
''')
s=replace(s,'            self.raw.destroy_fence(fence);\n        }',
    '            self.raw.destroy_fence(fence);\n        }\n        crate::staging_reuse::owner_ended(self.texture_staging_owner);')
p.write_text(s)
p=root/'core/src/device/life.rs';s=p.read_text()
s=replace(s,'pub fn track_submission(&mut self, index: SubmissionIndex, encoders: Vec<EncoderInFlight>) {','''pub fn track_submission(&mut self, index: SubmissionIndex, mut encoders: Vec<EncoderInFlight>) {
        for encoder in &mut encoders {
            for resource in &mut encoder.temp_resources {
                if let TempResource::StagingBuffer(staging) = resource { staging.mark_reuse_submission(index); }
            }
        }''')
s=replace(s,'''            for encoder in a.encoders {''','''            for mut encoder in a.encoders {''')
s=replace(s,'''                drop(encoder);''','''                let temporary = core::mem::take(&mut encoder.temp_resources);
                drop(encoder);
                for resource in temporary {
                    match resource {
                        TempResource::StagingBuffer(staging) => staging.recycle_completed(a.index, last_done),
                        other => drop(other),
                    }
                }''')
p.write_text(s)
p=root/'core/src/device/queue.rs';s=p.read_text();a=s.index('    pub fn write_texture(');b=s.index('    #[cfg(webgl)]\n    pub fn copy_external_image_to_texture',a);body=s[a:b]
body=replace(body,'        let mut pending_writes = self.pending_writes.lock();','''        let texture_reuse = self.device.adapter.backend() == wgt::Backend::Metal
            && dst.desc.dimension == wgt::TextureDimension::D2 && size.depth_or_array_layers == 1
            && matches!(dst.desc.format, wgt::TextureFormat::Rgba8Unorm | wgt::TextureFormat::Rgba8UnormSrgb
                | wgt::TextureFormat::Bgra8Unorm | wgt::TextureFormat::Bgra8UnormSrgb);
        probe.reuse_eligibility(texture_reuse);
        let mut pending_writes = self.pending_writes.lock();''')
old='StagingBuffer::new_with_upload_probe(&self.device, stage_size, Some(&probe))?';assert body.count(old)==2;body=body.replace(old,'StagingBuffer::new_with_upload_probe(&self.device, stage_size, Some(&probe), texture_reuse)?');p.write_text(s[:a]+body+s[b:])
p=root/'core/src/upload_probe.rs';s=p.read_text();s=replace(s,'impl Probe {','impl Probe {\n pub(crate) fn ordinal(&self) -> Option<usize> { self.index }');p.write_text(s)
p=root/'core/src/upload_probe.rs';s=p.read_text()
s=replace(s,'struct Record {begin:', 'struct Record {reuse_eligible:AtomicBool,begin:')
s=replace(s,'Self {begin:AtomicU64::new(0)', 'Self {reuse_eligible:AtomicBool::new(false),begin:AtomicU64::new(0)')
s=replace(s,'impl Probe {','impl Probe {\n pub(crate) fn reuse_eligibility(&self, eligible:bool) { if let Some(i)=self.index { ROWS[i].reuse_eligible.store(eligible,Ordering::Relaxed); } }')
s=replace(s,'\\\"stage_pitch\\\":{}}}', '\\\"stage_pitch\\\":{},\\\"reuse_eligible\\\":{}}}')
s=replace(s,'r.pitch.load(Ordering::Acquire)).expect(', 'r.pitch.load(Ordering::Acquire),r.reuse_eligible.load(Ordering::Acquire)).expect(')
p.write_text(s)
p=root/'core/src/lib.rs';p.write_text(p.read_text()+'\npub mod staging_reuse;\npub mod staging_reuse_policy;\n')
(root/'core/src/staging_reuse.rs').write_bytes((code/'trace.rs').read_bytes())
(root/'core/src/staging_reuse_policy.rs').write_bytes((code.parent/'reusepolicy/policy.rs').read_bytes()+b'''
extern crate alloc;
use alloc::{vec, vec::Vec};

// Private integration metadata access; no mutable or raw resource exposure.
impl<T> Completed<T> {
    pub fn id(&self) -> u64 { self.upload.id() }
    pub fn bytes(&self) -> u64 { self.upload.bytes() }
}
''')
p=root/'source/src/lib.rs';p.write_text(p.read_text()+'''
#[no_mangle]
pub extern "C" fn vellumReuseConfigure(enabled: bool) -> bool { wgc::staging_reuse::configure(enabled) }
#[no_mangle]
pub unsafe extern "C" fn vellumReuseSnapshot(out: *mut u8, capacity: usize) -> usize {
    let text = wgc::staging_reuse::snapshot();
    if !out.is_null() && text.len() < capacity {
        core::ptr::copy_nonoverlapping(text.as_ptr(), out, text.len()); *out.add(text.len()) = 0;
    }
    text.len()
}
''')
print('Private texture-only completed-staging prototype applied; not built, reviewed or executed yet.')
