//! Private scalar lifecycle journal. No resource or device references are stored here.
use core::sync::atomic::{AtomicBool, AtomicU64, AtomicUsize, Ordering};
use alloc::{format, string::String};
use core::fmt::Write;
const CAP: usize = 4096;
pub const DEVICE: u64 = 1;
pub const FRESH: u64 = 2;
pub const TAKE: u64 = 3;
pub const SUBMIT: u64 = 4;
pub const RETIRE: u64 = 5;
pub const CACHE: u64 = 6;
pub const DESTROY: u64 = 7;
pub const DRAIN: u64 = 8;
pub const DEVICE_END: u64 = 9;
pub const EVICT: u64 = 10;
static MODE: AtomicU64 = AtomicU64::new(0);
static NEXT_OWNER: AtomicU64 = AtomicU64::new(1);
static NEXT_RESOURCE: AtomicU64 = AtomicU64::new(1);
static COUNT: AtomicUsize = AtomicUsize::new(0);
static ACTIVE: AtomicUsize = AtomicUsize::new(0);
static DEVICES: AtomicUsize = AtomicUsize::new(0);
static OVERFLOW: AtomicBool = AtomicBool::new(false);
static INVALID: AtomicBool = AtomicBool::new(false);
struct Row { fields: [AtomicU64; 10], complete: AtomicBool }
impl Row { const fn new() -> Self { Self { fields: [const {AtomicU64::new(0)}; 10], complete: AtomicBool::new(false) } } }
static ROWS: [Row; CAP] = [const {Row::new()}; CAP];
#[link(name="QuartzCore", kind="framework")]
extern "C" { fn CACurrentMediaTime() -> f64; }
pub fn enabled() -> bool { MODE.load(Ordering::Acquire) == 2 }
pub fn invalid() { INVALID.store(true, Ordering::Release); }
pub fn configure(enabled: bool) -> bool {
    NEXT_OWNER.load(Ordering::Acquire) == 1
        && MODE.compare_exchange(0, if enabled {2} else {1}, Ordering::AcqRel, Ordering::Acquire).is_ok()
}
pub fn record(kind: u64, owner: u64, id: u64, bytes: u64, submit: u64, fence: u64,
              call: u64, cached_count: usize, cached_bytes: u64) {
    if MODE.load(Ordering::Acquire) == 0 { invalid(); }
    ACTIVE.fetch_add(1, Ordering::Relaxed);
    let i = COUNT.fetch_add(1, Ordering::Relaxed);
    if i >= CAP { OVERFLOW.store(true, Ordering::Release); ACTIVE.fetch_sub(1, Ordering::Release); return; }
    let fields = [kind, owner, id, bytes, submit, fence, call, cached_count as u64,
                  cached_bytes, unsafe {CACurrentMediaTime()}.to_bits()];
    for (target, value) in ROWS[i].fields.iter().zip(fields) { target.store(value, Ordering::Relaxed); }
    ROWS[i].complete.store(true, Ordering::Release);
    ACTIVE.fetch_sub(1, Ordering::Release);
}
pub fn owner_created() -> u64 {
    let id = NEXT_OWNER.fetch_add(1, Ordering::Relaxed);
    DEVICES.fetch_add(1, Ordering::Relaxed);
    record(DEVICE, id, 0, 0, 0, 0, 0, 0, 0); id
}
pub fn owner_ended(owner: u64) {
    record(DEVICE_END, owner, 0, 0, 0, 0, 0, 0, 0);
    DEVICES.fetch_sub(1, Ordering::Release);
}
pub fn fresh(owner: u64, bytes: u64, call: u64) -> u64 {
    let id = NEXT_RESOURCE.fetch_add(1, Ordering::Relaxed);
    record(FRESH, owner, id, bytes, 0, 0, call, 0, 0); id
}
pub fn snapshot() -> String {
    let active_before = ACTIVE.load(Ordering::Acquire);
    let count = COUNT.load(Ordering::Acquire);
    let devices = DEVICES.load(Ordering::Acquire);
    let mut partial = false;
    let mut records = String::new();
    if active_before == 0 {
        for (i, row) in ROWS.iter().take(count.min(CAP)).enumerate() {
            if !row.complete.load(Ordering::Acquire) { partial = true; break; }
            let f = row.fields.each_ref().map(|v| v.load(Ordering::Acquire));
            if i > 0 { records.push(','); }
            write!(records, "{{\"sequence\":{},\"complete\":true,\"kind\":{},\"owner\":{},\"id\":{},\"bytes\":{},\"submit\":{},\"fence\":{},\"call\":{},\"cached_count\":{},\"cached_bytes\":{},\"clock\":{}}}",
                i,f[0],f[1],f[2],f[3],f[4],f[5],f[6],f[7],f[8],f64::from_bits(f[9])).expect("String formatting");
        }
    }
    let active_after = ACTIVE.load(Ordering::Acquire);
    let unstable = partial || active_before != 0 || active_after != 0
        || COUNT.load(Ordering::Acquire) != count || DEVICES.load(Ordering::Acquire) != devices;
    if unstable { records.clear(); }
    format!("{{\"mode\":{},\"count\":{},\"pending\":{},\"unstable\":{},\"overflow\":{},\"invalid\":{},\"active_devices\":{},\"records\":[{}]}}",
        MODE.load(Ordering::Acquire),count,if unstable {active_after.max(1)} else {0},unstable,
        OVERFLOW.load(Ordering::Acquire),INVALID.load(Ordering::Acquire),devices,records)
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::sync::{Arc, Barrier};
    #[test]
    fn incomplete_and_concurrent_records_never_publish_as_stable_partial_data() {
        assert!(configure(false)); let owner = owner_created();
        // Deliberately hold an unpublished slot: scalar recorder control, no GPU resource.
        ACTIVE.fetch_add(1, Ordering::Relaxed);
        let slot = COUNT.fetch_add(1, Ordering::Relaxed);
        assert!(snapshot().contains("\"unstable\":true"));
        ACTIVE.fetch_sub(1, Ordering::Release);
        assert!(snapshot().contains("\"unstable\":true")); // Missing publication is also refused.
        let fields = [FRESH,owner,999,24,0,0,0,0,0,unsafe {CACurrentMediaTime()}.to_bits()];
        for (target,value) in ROWS[slot].fields.iter().zip(fields) { target.store(value,Ordering::Relaxed); }
        ROWS[slot].complete.store(true, Ordering::Release);
        let barrier = Arc::new(Barrier::new(5));
        let threads = (0..4).map(|worker| {
            let barrier = barrier.clone();
            std::thread::spawn(move || { barrier.wait(); for i in 0..100 {
                record(FRESH, owner, 1000+worker*100+i, 24, 0, 0, 0, 0, 0);
                std::thread::yield_now();
            } })
        }).collect::<std::vec::Vec<_>>();
        barrier.wait();
        for _ in 0..64 {
            let text = snapshot();
            if text.contains("\"unstable\":false") {
                assert!(text.contains("\"pending\":0"));
                let count = text.split("\"count\":").nth(1).unwrap().split(',').next().unwrap().parse::<usize>().unwrap();
                assert_eq!(text.matches("\"complete\":true").count(), count);
            }
        }
        for thread in threads { thread.join().unwrap(); }
        owner_ended(owner);
        let text = snapshot();
        assert!(text.contains("\"unstable\":false"));
        assert!(text.contains("\"pending\":0"));
        assert!(text.contains("\"count\":403"));
        assert!(text.contains("\"active_devices\":0"));
    }
}
