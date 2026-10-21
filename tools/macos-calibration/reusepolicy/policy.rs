//! Diagnostic ownership policy, not connected to any native upload path.
//! T must be a uniquely owned resource (the proposed integration uses Box<DynBuffer>).
//! These types do not wait on the GPU or prevent a caller from dropping a value.
//! The native tracker must retain Submitted until its real completion/destruction boundary.
use std::num::NonZeroU64;

#[derive(Debug)]
pub struct Upload<T> {
    id: NonZeroU64,
    bytes: NonZeroU64,
    resource: T,
}
#[derive(Debug)]
pub struct Submitted<T> {
    upload: Upload<T>,
    index: NonZeroU64,
}
#[derive(Debug)]
pub struct Completed<T> {
    upload: Upload<T>,
}

impl<T> Upload<T> {
    pub fn new(id: NonZeroU64, bytes: NonZeroU64, resource: T) -> Self {
        Self { id, bytes, resource }
    }
    pub fn id(&self) -> u64 { self.id.get() }
    pub fn bytes(&self) -> u64 { self.bytes.get() }
    pub fn resource(&self) -> &T { &self.resource }
    pub fn resource_mut(&mut self) -> &mut T { &mut self.resource }
    pub fn into_resource(self) -> T { self.resource }
    pub fn submit(self, index: NonZeroU64) -> Submitted<T> {
        Submitted { upload: self, index }
    }
}
impl<T> Submitted<T> {
    pub fn id(&self) -> u64 { self.upload.id() }
    pub fn index(&self) -> u64 { self.index.get() }
    /// Only the native completion tracker may supply this observed fence value.
    /// Comparing integers cannot authenticate a GPU fence on its own.
    pub fn complete(self, observed_fence: u64) -> Result<Completed<T>, Self> {
        if observed_fence < self.index.get() { return Err(self); }
        Ok(Completed { upload: self.upload })
    }
    /// Failure/destruction escape hatch. Must retain the existing native
    /// destruction rules; this is never an admission path into Pool.
    pub fn into_resource(self) -> T { self.upload.resource }
}
impl<T> Completed<T> {
    pub fn into_upload(self) -> Upload<T> { self.upload }
}

/// Stores only completed resources. It never owns pending/in-flight resources.
/// The owner must explicitly destroy evictions/drained native handles through
/// their original backend API, outside the cache mutex, before destroying Device.
#[derive(Debug)]
pub struct Pool<T> {
    entries: Vec<Completed<T>>,
    bytes: u64,
    max_entries: usize,
    max_bytes: u64,
}
impl<T> Pool<T> {
    pub fn new(max_entries: usize, max_bytes: u64) -> Self {
        Self { entries: Vec::new(), bytes: 0, max_entries, max_bytes }
    }
    pub fn len(&self) -> usize { self.entries.len() }
    pub fn is_empty(&self) -> bool { self.entries.is_empty() }
    pub fn bytes(&self) -> u64 { self.bytes }
    pub fn take_exact(&mut self, bytes: NonZeroU64) -> Option<Upload<T>> {
        let index = self.entries.iter().position(|entry| entry.upload.bytes == bytes)?;
        let entry = self.entries.remove(index);
        self.bytes -= entry.upload.bytes();
        Some(entry.into_upload())
    }
    /// FIFO eviction; oversized/disabled entries are returned without retention.
    /// Returned resources must follow the original backend destruction path.
    #[must_use]
    pub fn insert(&mut self, entry: Completed<T>) -> Vec<Completed<T>> {
        let bytes = entry.upload.bytes();
        if self.max_entries == 0 || bytes > self.max_bytes { return vec![entry]; }
        let mut evicted = Vec::new();
        // Subtraction is safe after the oversized check. Avoid overflow in a
        // hypothetical bytes+new_bytes sum, even for u64::MAX test capacities.
        while self.entries.len() >= self.max_entries || self.bytes > self.max_bytes - bytes {
            let old = self.entries.remove(0);
            self.bytes -= old.upload.bytes();
            evicted.push(old);
        }
        self.bytes += bytes;
        self.entries.push(entry);
        evicted
    }
    #[must_use]
    pub fn drain(&mut self) -> Vec<Completed<T>> {
        self.bytes = 0;
        std::mem::take(&mut self.entries)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::{cell::Cell, rc::Rc};
    fn nz(n: u64) -> NonZeroU64 { NonZeroU64::new(n).unwrap() }
    fn upload(id: u64, bytes: u64) -> Upload<Vec<u8>> {
        Upload::new(nz(id), nz(bytes), vec![id as u8; bytes as usize])
    }
    fn completed(id: u64, bytes: u64) -> Completed<Vec<u8>> {
        upload(id, bytes).submit(nz(7)).complete(7).unwrap()
    }
    #[test]
    fn no_checkout_before_completion_and_no_duplicate_lease() {
        let mut pool = Pool::new(8, 32);
        let pending = upload(1, 8).submit(nz(5));
        assert!(pool.take_exact(nz(8)).is_none());
        let pending = pending.complete(4).unwrap_err();
        assert_eq!((pending.id(), pending.index()), (1, 5));
        assert!(pool.insert(pending.complete(5).unwrap()).is_empty());
        let mut leased = pool.take_exact(nz(8)).unwrap();
        assert_eq!(leased.resource(), &[1; 8]);
        leased.resource_mut()[0] = 9;
        assert!(pool.take_exact(nz(8)).is_none());
        let pending = leased.submit(nz(9));
        let pending = pending.complete(8).unwrap_err();
        assert!(pool.is_empty());
        assert!(pool.insert(pending.complete(9).unwrap()).is_empty());
        assert_eq!(pool.take_exact(nz(8)).unwrap().resource()[0], 9);
    }
    #[test]
    fn exact_size_out_of_order_completion_and_payload_preservation() {
        let mut pool = Pool::new(8, 32);
        let earlier = upload(1, 8).submit(nz(10));
        let later = upload(2, 12).submit(nz(11));
        assert!(pool.insert(later.complete(11).unwrap()).is_empty());
        assert!(pool.take_exact(nz(8)).is_none());
        assert!(pool.take_exact(nz(9)).is_none());
        assert!(pool.insert(earlier.complete(11).unwrap()).is_empty());
        assert_eq!(pool.take_exact(nz(12)).unwrap().into_resource(), vec![2; 12]);
        assert_eq!(pool.take_exact(nz(8)).unwrap().id(), 1);
        assert_eq!((pool.len(), pool.bytes()), (0, 0));
    }
    #[test]
    fn entry_and_byte_caps_fifo_evictions_and_oversized_rejection() {
        let mut pool = Pool::new(2, 16);
        assert!(pool.insert(completed(1, 8)).is_empty());
        assert!(pool.insert(completed(2, 8)).is_empty());
        let evicted = pool.insert(completed(3, 12));
        assert_eq!(evicted.into_iter().map(|e| e.into_upload().id()).collect::<Vec<_>>(), vec![1, 2]);
        assert_eq!((pool.len(), pool.bytes()), (1, 12));
        assert_eq!(pool.insert(completed(4, 17))[0].upload.id(), 4);
        assert_eq!((pool.len(), pool.bytes()), (1, 12));
        let mut zero = Pool::new(0, 32);
        assert_eq!(zero.insert(completed(5, 8)).len(), 1);
        assert!(zero.is_empty());
        let mut entry_bound = Pool::new(2, 100);
        assert!(entry_bound.insert(completed(1, 8)).is_empty());
        assert!(entry_bound.insert(completed(2, 8)).is_empty());
        assert_eq!(entry_bound.insert(completed(3, 8))[0].upload.id(), 1);
        assert_eq!((entry_bound.len(), entry_bound.bytes()), (2, 16));
    }
    #[test]
    fn arithmetic_extremes_do_not_overflow() {
        let mut pool = Pool::new(2, u64::MAX);
        let entry = |id, bytes| Upload::new(nz(id), nz(bytes), ()).submit(nz(1)).complete(1).unwrap();
        assert!(pool.insert(entry(1, u64::MAX)).is_empty());
        assert_eq!(pool.insert(entry(2, 1)).len(), 1);
        assert_eq!(pool.bytes(), 1);
    }
    #[derive(Debug)]
    struct Resource(Rc<Cell<usize>>);
    impl Drop for Resource { fn drop(&mut self) { self.0.set(self.0.get()+1); } }
    #[test]
    fn drain_failure_and_eviction_preserve_single_ownership() {
        let drops = Rc::new(Cell::new(0));
        let make = |id| Upload::new(nz(id), nz(8), Resource(drops.clone()));
        let mut pool = Pool::new(1, 8);
        assert!(pool.insert(make(1).submit(nz(1)).complete(1).unwrap()).is_empty());
        let evicted = pool.insert(make(2).submit(nz(2)).complete(2).unwrap());
        assert_eq!(drops.get(), 0); drop(evicted); assert_eq!(drops.get(), 1);
        let drained = pool.drain(); assert_eq!((pool.len(), pool.bytes()), (0, 0));
        drop(drained); assert_eq!(drops.get(), 2);
        drop(make(3)); assert_eq!(drops.get(), 3); // CPU-side cancellation.
        drop(make(4).submit(nz(4)).into_resource()); assert_eq!(drops.get(), 4);
    }
}
