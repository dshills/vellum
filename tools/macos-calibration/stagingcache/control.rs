//! Private diagnostic only. Configure before creating a device; no retained handles.
use core::sync::atomic::{AtomicU64, Ordering};
static MODE: AtomicU64 = AtomicU64::new(0);
static ELIGIBLE: AtomicU64 = AtomicU64::new(0);
static WC: AtomicU64 = AtomicU64::new(0);
static DEFAULT: AtomicU64 = AtomicU64::new(0);
static MISMATCH: AtomicU64 = AtomicU64::new(0);
static UNCONFIGURED: AtomicU64 = AtomicU64::new(0);

pub fn configure(default_cache: bool) -> bool {
    // Caller must be quiescent, before the first device/resource, not just this load.
    ELIGIBLE.load(Ordering::Acquire) == 0
        && MODE.compare_exchange(0, if default_cache { 2 } else { 1 }, Ordering::AcqRel, Ordering::Acquire).is_ok()
}
pub fn default_cache() -> bool { MODE.load(Ordering::Acquire) == 2 }
pub fn record(cpu_cache_mode: u64, storage_mode: u64) {
    let mode = MODE.load(Ordering::Acquire);
    ELIGIBLE.fetch_add(1, Ordering::Relaxed);
    if mode == 0 { UNCONFIGURED.fetch_add(1, Ordering::Relaxed); }
    match cpu_cache_mode {
        0 => { DEFAULT.fetch_add(1, Ordering::Relaxed); }
        1 => { WC.fetch_add(1, Ordering::Relaxed); }
        _ => { MISMATCH.fetch_add(1, Ordering::Relaxed); return; }
    }
    if storage_mode != 0 || mode == 0 || cpu_cache_mode != if mode == 2 { 0 } else { 1 } {
        MISMATCH.fetch_add(1, Ordering::Relaxed);
    }
}
pub fn snapshot() -> [u64; 6] {
    [&MODE, &ELIGIBLE, &WC, &DEFAULT, &MISMATCH, &UNCONFIGURED]
        .map(|v| v.load(Ordering::Acquire))
}

#[cfg(test)]
mod tests {
    use super::*;
    // Run each in a separate process, because production configuration is one-shot.
    #[test]
    fn writecombined() {
        assert!(!default_cache()); assert!(configure(false)); assert!(!configure(true));
        record(1, 0); assert_eq!(snapshot(), [1, 1, 1, 0, 0, 0]);
        record(0, 0); record(1, 1); record(99, 0);
        assert_eq!(snapshot(), [1, 4, 2, 1, 3, 0]);
    }
    #[test]
    fn default_mode() {
        assert!(configure(true)); assert!(default_cache()); assert!(!configure(false));
        record(0, 0); assert_eq!(snapshot(), [2, 1, 0, 1, 0, 0]);
        record(1, 0); assert_eq!(snapshot(), [2, 2, 1, 1, 1, 0]);
    }
    #[test]
    fn refuse_late_configuration() {
        record(1, 0); assert!(!configure(true)); assert_eq!(snapshot(), [0, 1, 1, 0, 1, 1]);
    }
}
