//! Private CPU upload observer. No resource handles, GPU timestamps or eviction.
use core::sync::atomic::{AtomicBool,AtomicU64,AtomicUsize,Ordering};
use alloc::{format,string::String};
use core::fmt::Write;
const CAP:usize=4096;
struct Record {begin:AtomicU64,end:AtomicU64,create:AtomicU64,map:AtomicU64,copy:AtomicU64,flush:AtomicU64,encode:AtomicU64,bytes:AtomicU64,width:AtomicU64,height:AtomicU64,pitch:AtomicU64,stages:AtomicUsize,status:AtomicUsize,complete:AtomicBool}
impl Record {const fn new()->Self {Self {begin:AtomicU64::new(0),end:AtomicU64::new(0),create:AtomicU64::new(0),map:AtomicU64::new(0),copy:AtomicU64::new(0),flush:AtomicU64::new(0),encode:AtomicU64::new(0),bytes:AtomicU64::new(0),width:AtomicU64::new(0),height:AtomicU64::new(0),pitch:AtomicU64::new(0),stages:AtomicUsize::new(0),status:AtomicUsize::new(0),complete:AtomicBool::new(false)}}}
static ROWS:[Record;CAP]=[const {Record::new()};CAP];
static COUNT:AtomicUsize=AtomicUsize::new(0);
static ACTIVE:AtomicUsize=AtomicUsize::new(0);
static OVERFLOW:AtomicBool=AtomicBool::new(false);
static INVALID_STAGE:AtomicBool=AtomicBool::new(false);
static CONFIGURED:AtomicBool=AtomicBool::new(false);
static ENABLED:AtomicBool=AtomicBool::new(false);
#[link(name="QuartzCore",kind="framework")]
extern "C" {fn CACurrentMediaTime()->f64;}
fn now()->f64 {unsafe {CACurrentMediaTime()}}
pub struct Probe {index:Option<usize>,success:bool}
impl Probe {
 pub fn begin(bytes:usize,width:u32,height:u32)->Self {
  if !ENABLED.load(Ordering::Relaxed){return Self {index:None,success:false};}
  let i=COUNT.fetch_add(1,Ordering::Relaxed);if i>=CAP {OVERFLOW.store(true,Ordering::Release);return Self {index:None,success:false};}
  ACTIVE.fetch_add(1,Ordering::Relaxed);let r=&ROWS[i];r.begin.store(now().to_bits(),Ordering::Relaxed);r.bytes.store(bytes as u64,Ordering::Relaxed);r.width.store(width as u64,Ordering::Relaxed);r.height.store(height as u64,Ordering::Relaxed);
  Self {index:Some(i),success:false}
 }
 pub fn clock(&self)->f64 {if self.index.is_some(){now()}else{0.0}}
 fn duration(&self,start:f64)->u64 {((now()-start)*1e9).to_bits()}
 pub fn create_done(&self,start:f64){if let Some(i)=self.index {ROWS[i].create.store(self.duration(start),Ordering::Relaxed);if ROWS[i].stages.fetch_or(1,Ordering::Relaxed)&1!=0 {INVALID_STAGE.store(true,Ordering::Release);}}}
 pub fn map_done(&self,start:f64){if let Some(i)=self.index {ROWS[i].map.store(self.duration(start),Ordering::Relaxed);if ROWS[i].stages.fetch_or(2,Ordering::Relaxed)&2!=0 {INVALID_STAGE.store(true,Ordering::Release);}}}
 pub fn copy_done(&self,start:f64){if let Some(i)=self.index {ROWS[i].copy.store(self.duration(start),Ordering::Relaxed);if ROWS[i].stages.fetch_or(4,Ordering::Relaxed)&4!=0 {INVALID_STAGE.store(true,Ordering::Release);}}}
 pub fn flush_done(&self,start:f64){if let Some(i)=self.index {ROWS[i].flush.store(self.duration(start),Ordering::Relaxed);if ROWS[i].stages.fetch_or(8,Ordering::Relaxed)&8!=0 {INVALID_STAGE.store(true,Ordering::Release);}}}
 pub fn encode_done(&self,start:f64){if let Some(i)=self.index {ROWS[i].encode.store(self.duration(start),Ordering::Relaxed);if ROWS[i].stages.fetch_or(16,Ordering::Relaxed)&16!=0 {INVALID_STAGE.store(true,Ordering::Release);}}}
 pub fn pitch(&self,pitch:u32){if let Some(i)=self.index {ROWS[i].pitch.store(pitch as u64,Ordering::Relaxed);}}
 pub fn success(&mut self){self.success=true;}
}
impl Drop for Probe {fn drop(&mut self){if let Some(i)=self.index {let r=&ROWS[i];r.end.store(now().to_bits(),Ordering::Relaxed);r.status.store(if self.success {1}else{2},Ordering::Relaxed);r.complete.store(true,Ordering::Release);ACTIVE.fetch_sub(1,Ordering::Release);}}}
pub fn configure(enabled:bool)->bool {if CONFIGURED.swap(true,Ordering::AcqRel){return false;}ENABLED.store(enabled,Ordering::Release);true}
pub fn snapshot()->String {
 let count=COUNT.load(Ordering::Acquire);let active=ACTIVE.load(Ordering::Acquire);
 let mut out=format!("{{\"enabled\":{},\"pending\":{},\"overflow\":{},\"count\":{},\"invalid_stage\":{},\"records\":[",ENABLED.load(Ordering::Acquire),active,OVERFLOW.load(Ordering::Acquire),count,INVALID_STAGE.load(Ordering::Acquire));
 if active==0 {for (i,r) in ROWS.iter().enumerate().take(count.min(CAP)) {
  let complete=r.complete.load(Ordering::Acquire);if i>0 {out.push(',');}
  write!(out,"{{\"ordinal\":{},\"complete\":{},\"status\":{},\"stages\":{},\"begin\":{},\"end\":{},\"create_ns\":{},\"map_ns\":{},\"copy_ns\":{},\"flush_ns\":{},\"encode_ns\":{},\"bytes\":{},\"width\":{},\"height\":{},\"stage_pitch\":{}}}",i,complete,r.status.load(Ordering::Acquire),r.stages.load(Ordering::Acquire),f64::from_bits(r.begin.load(Ordering::Acquire)),f64::from_bits(r.end.load(Ordering::Acquire)),f64::from_bits(r.create.load(Ordering::Acquire)),f64::from_bits(r.map.load(Ordering::Acquire)),f64::from_bits(r.copy.load(Ordering::Acquire)),f64::from_bits(r.flush.load(Ordering::Acquire)),f64::from_bits(r.encode.load(Ordering::Acquire)),r.bytes.load(Ordering::Acquire),r.width.load(Ordering::Acquire),r.height.load(Ordering::Acquire),r.pitch.load(Ordering::Acquire)).expect("String formatting");
 }}out.push_str("]}");out
}
