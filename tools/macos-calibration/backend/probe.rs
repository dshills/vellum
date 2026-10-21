//! Diagnostic only: fixed, non-evicting Metal submission observations.
use super::*;
use alloc::{format, string::String};
use core::{fmt::Write, sync::atomic::{AtomicBool, AtomicU64, AtomicUsize, Ordering}};
const CAP:usize=4096;
struct Record { kind:AtomicUsize, queue:AtomicUsize, drawable:AtomicUsize, drawable_recorded:AtomicBool, start:AtomicU64, end:AtomicU64, request:AtomicU64, scheduled:AtomicU64, complete:AtomicU64, gpu_start:AtomicU64, gpu_end:AtomicU64, shown:AtomicU64, status:AtomicUsize }
impl Record { const fn new()->Self { Self {kind:AtomicUsize::new(0),queue:AtomicUsize::new(0),drawable:AtomicUsize::new(0),drawable_recorded:AtomicBool::new(false),start:AtomicU64::new(0),end:AtomicU64::new(0),request:AtomicU64::new(0),scheduled:AtomicU64::new(0),complete:AtomicU64::new(0),gpu_start:AtomicU64::new(0),gpu_end:AtomicU64::new(0),shown:AtomicU64::new(0),status:AtomicUsize::new(0)} } }
static ROWS:[Record;CAP]=[const {Record::new()};CAP];
static ENABLED:AtomicBool=AtomicBool::new(false);
static CONFIGURED:AtomicBool=AtomicBool::new(false);
static COUNT:AtomicUsize=AtomicUsize::new(0);
static PENDING:AtomicUsize=AtomicUsize::new(0);
static OVERFLOW:AtomicBool=AtomicBool::new(false);
#[link(name="QuartzCore",kind="framework")]
extern "C" { fn CACurrentMediaTime()->f64; }
fn now()->u64 { unsafe {CACurrentMediaTime()}.to_bits() }
fn reserve(kind:usize)->Option<usize> {
 if !ENABLED.load(Ordering::Relaxed){return None;}
 let i=COUNT.fetch_add(1,Ordering::Relaxed);if i>=CAP {OVERFLOW.store(true,Ordering::Release);return None;}
 ROWS[i].kind.store(kind,Ordering::Relaxed);Some(i)
}
fn observe(raw:&ProtocolObject<dyn MTLCommandBuffer>,i:usize) {
 ROWS[i].queue.store(Retained::as_ptr(&raw.commandQueue()) as usize,Ordering::Relaxed);
 PENDING.fetch_add(2,Ordering::Relaxed);
 let scheduled=block2::RcBlock::new(move |_:NonNull<ProtocolObject<dyn MTLCommandBuffer>>| {ROWS[i].scheduled.store(now(),Ordering::Relaxed);PENDING.fetch_sub(1,Ordering::Release);});
 let complete=block2::RcBlock::new(move |cmd:NonNull<ProtocolObject<dyn MTLCommandBuffer>>| {
  let cmd=unsafe {cmd.as_ref()};let r=&ROWS[i];r.complete.store(now(),Ordering::Relaxed);let gpu_start:f64=unsafe {objc2::msg_send![cmd, GPUStartTime]};let gpu_end:f64=unsafe {objc2::msg_send![cmd, GPUEndTime]};r.gpu_start.store(gpu_start.to_bits(),Ordering::Relaxed);r.gpu_end.store(gpu_end.to_bits(),Ordering::Relaxed);r.status.store(cmd.status().0,Ordering::Relaxed);PENDING.fetch_sub(1,Ordering::Release);
 });
 unsafe {raw.addScheduledHandler(block2::RcBlock::as_ptr(&scheduled));raw.addCompletedHandler(block2::RcBlock::as_ptr(&complete));}
}
fn commit_index(raw:&ProtocolObject<dyn MTLCommandBuffer>,index:Option<usize>) {
 if let Some(i)=index {observe(raw,i);ROWS[i].start.store(now(),Ordering::Relaxed);raw.commit();ROWS[i].end.store(now(),Ordering::Release);} else {raw.commit();}
}
pub(super) fn commit(raw:&ProtocolObject<dyn MTLCommandBuffer>,kind:usize) {commit_index(raw,reserve(kind));}
pub(super) fn present(raw:&ProtocolObject<dyn MTLCommandBuffer>,drawable:&ProtocolObject<dyn MTLDrawable>,transaction:bool) {
 let index=reserve(2);
 if let Some(i)=index {
  ROWS[i].drawable.store(drawable.drawableID(),Ordering::Relaxed);
  ROWS[i].drawable_recorded.store(true,Ordering::Release);
  PENDING.fetch_add(1,Ordering::Relaxed);
  let shown=block2::RcBlock::new(move |d:NonNull<ProtocolObject<dyn MTLDrawable>>| {let drawable=unsafe {d.as_ref()};let time:f64=unsafe {objc2::msg_send![drawable, presentedTime]};ROWS[i].shown.store(time.to_bits(),Ordering::Relaxed);PENDING.fetch_sub(1,Ordering::Release);});
  unsafe {drawable.addPresentedHandler(block2::RcBlock::as_ptr(&shown));}
  ROWS[i].request.store(now(),Ordering::Relaxed);
 }
 if !transaction {raw.presentDrawable(drawable);}
 commit_index(raw,index);
}
/// Configure exactly once, before creating the framework runtime. No hot reset.
pub fn configure(enabled:bool)->bool {if CONFIGURED.swap(true,Ordering::AcqRel){return false;}ENABLED.store(enabled,Ordering::Release);true}
/// Caller collects after runtime shutdown; all storage is atomic and non-evicting.
pub fn snapshot()->String {
 let count=COUNT.load(Ordering::Acquire);let pending=PENDING.load(Ordering::Acquire);
 // Runtime is already closed: no new submissions. An acquire read of zero
 // observes all callback writes before their release decrements. Never walk
 // partially populated rows while the Go caller is waiting for drain.
 if pending!=0 {return format!("{{\"enabled\":{},\"pending\":{},\"overflow\":{},\"count\":{},\"records\":[]}}",ENABLED.load(Ordering::Acquire),pending,OVERFLOW.load(Ordering::Acquire),count);}
 let mut out=format!("{{\"enabled\":{},\"pending\":{},\"overflow\":{},\"count\":{},\"records\":[",ENABLED.load(Ordering::Acquire),pending,OVERFLOW.load(Ordering::Acquire),count);
 for (i,r) in ROWS.iter().enumerate().take(count.min(CAP)) {
  if i>0 {out.push(',');}
  let drawable_recorded=r.drawable_recorded.load(Ordering::Acquire);
  let val=|a:&AtomicU64| f64::from_bits(a.load(Ordering::Acquire));
  write!(out,"{{\"ordinal\":{},\"kind\":{},\"queue\":{},\"drawable_id\":{},\"drawable_recorded\":{},\"commit_start\":{},\"commit_end\":{},\"present_request\":{},\"scheduled_observed\":{},\"complete_observed\":{},\"gpu_start\":{},\"gpu_end\":{},\"shown\":{},\"status\":{}}}",i,r.kind.load(Ordering::Acquire),r.queue.load(Ordering::Acquire),r.drawable.load(Ordering::Acquire),drawable_recorded,val(&r.start),val(&r.end),val(&r.request),val(&r.scheduled),val(&r.complete),val(&r.gpu_start),val(&r.gpu_end),val(&r.shown),r.status.load(Ordering::Acquire)).expect("String formatting");
 }
 out.push_str("]}");out
}
