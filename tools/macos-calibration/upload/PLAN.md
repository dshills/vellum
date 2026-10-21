# Private CPU upload-stage observer

The preceding source-cost result identifies image API work as76.734% of summed
scene Draw CPU time, p95 2.6495ms. It does not distinguish native allocation,
mapping, copying, flushing or encoding. The exact locked source audit shows a
fresh staging buffer per nonempty write, direct borrowed Go data and Metal pitch
alignment4. Do not assume allocation is the cause, remove validation, add a pool,
or change cache/resource/renderer/pacing policy before measurement.

Instrument only a private clone of wgpu-core29.0.3 and wgpu-native's private ABI.
Keep validation, buffer descriptors, create/map/copy/flush/unmap/encode order,
original errors, pending-write ownership and GPU completion lifetime unchanged.
The ordinary StagingBuffer::new delegates to the same implementation with no probe;
Queue::write_texture uses the observer. No buffer or device handle enters trace
storage. RAII records CPU function return (success/error), never GPU completion.

Fixed4096 atomic slots never evict. Track begin/end host clocks, data length,
texture dimensions, staging pitch and disjoint create/map/copy/flush/encode CPU
intervals. Missing-stage mask or duplicate-stage flag refuses; zero observed time
is distinct from an unobserved stage. Preserve negative/nonfinite clock faults
as invalid data, not clamped measurements. Configure exactly once before runtime;
snapshot after shutdown with no new uploads, active-call drain and2MiB JSON limit.
The recorder allocates no per-upload heap object or resource reference.

Use the existing detailed-source fixture and three private Go compiler-overlay
locations from29f6cf4, leaving retained SDK files untouched. Reuse byte-identical
view/routing/pixel and existing scalar validation. Only the private native archive
and observer adapter differ. Build identities and archive-derived Go link identity
prevent old binary reuse. No new public API or shipping default change.

Run one diagnostic pair off/on using the same private binary/archive. Recording
off emits no trace; both modes retain detailed Go clocks. This does not prove zero
observer overhead, native binary equivalence or ordinary latency.128warmup plus
120measured accepted requests per process at120ms; locked match/configuration,
foreground checks, complete existing accounting, source/SDK/binary/log pins and
owned executor/exclusive lease remain unchanged.90seconds global,40seconds per
process, one attempt each, stopfirstinvalid/error/focus/config/deadline/accounting
failure, seal all data. No unchanged retry or broad campaign.

For on, correlate every cohort texture write between the actual receipt and its
final armed acquisition. Require count/logical bytes equal all image API attempts
in that ticket's detailed prefix+final records, matching fixtureRGBA extent/pitch,
unique attribution and native CPU span within enclosing Go image API (1us clock
serialization allowance). Every remaining upload must finish before first cohort
receipt, with at most128 retained startup records; no unknown post-admission write
is dropped. All rows must complete successfully with finite ordered clocks and
recorded nonempty stages. Disjoint stage sums fit total span within1ns conversion
allowance; signed residuals are preserved rather than clamped. Report final120
per-ticket summed phases, call counts and all raw startup/error records. Quantiles
are not additive. CPU success does not establish successful/physical GPU transfer,
GPU resource deallocation or latency acceptance.

Use focused wrapper/refusal controls, existing source/parent tests and view/routing
reuse, actual native/Go build and identity/overlay verification. Internal review
and Prism AnthropicClaudeSonnet5.5 precede source commit/native run; separately
validate/review/commit results. Keep features frozen, experimental renderer/pacing
options off,33.3ms gate unchanged and M3–M7 open.
