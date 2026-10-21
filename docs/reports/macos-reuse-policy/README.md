# Completed staging reuse: policy and source audit

This checkpoint supplies a tested ownership policy and a concrete private-prototype
plan. **At this policy checkpoint, no native pool was implemented or connected to the framework.**
The preceding cache-mode control still misses the 33.3 ms gate.

The exact locked wgpu-core 29.0.3 source has an explicit completed-submission path:
Device reads the backend fence, Queue passes the observed index to LifetimeTracker,
and triage retires encoders/resources only through that index. Generic staging
Drop also handles failure/unsent cleanup and cannot be treated as completion proof.
Public create_staging_buffer exposes a mapped pointer, so proposed reuse must
remain internal to texture uploads. Cached raw handles must avoid an Arc<Device>
cycle and be drained through backend destruction before Device teardown.

The [design](../../../tools/macos-calibration/reusepolicy/PLAN.md) proposes exact-size
per-device idle storage bounded to eight entries/32 MiB, with explicit completion
admission, original failure/destruction paths and a scalar lifecycle journal.
Those bounds would limit idle cache storage; they do not bound in-flight GPU
allocation. No cache-mode, pacing, renderer or public API change is included.

The standalone [policy](../../../tools/macos-calibration/reusepolicy/policy.rs) passes
five Rust test methods: completion/unique checkout, exact sizes/out-of-order
completion/payload preservation, entry/byte caps/FIFO/oversize rejection, arithmetic
extremes and single-owner drain/failure cleanup. Two negative compiler controls
reject direct Submitted→Pool admission and mutable Submitted resource access.
An integer comparison does not authenticate a real GPU fence; these types also do
not wait on the GPU, prevent dropping a value, synchronize native threads or prove
safe native destruction. Box<DynBuffer> integration still needs source, event,
readback, lock-order, error and teardown verification. The destruction escape
hatch must retain existing native lifetime rules.

Evidence binds the exact Core registry archive checksum, relevant original and
observer-source hashes, completion/destruction call sites, test source/compiler/
binary and negative diagnostics. The pure Rust tests use ordinary CLI execution;
no owned native-GPU campaign or allocation-balance proof is claimed. Existing
headless NoDisplayGPU/Texture/ReadData APIs can support the proposed native byte
checks; those future checks have not run for a pool.

Next implement/review only the private prototype, then require bounded headless
WebGPU byte/ownership proof before any timing pair. Keep the shipped library, SDK
and Placement app untouched. Features remain frozen, experimental rendering/pacing
is disabled, the 33.3 ms gate is unchanged and M3–M7 stay open.

Internal review and complete six-file Prism Anthropic Claude Sonnet 5.5 review
are recorded. A proposed empty-eviction concern was disposed by tracing the private
field invariants and all mutation paths; no unchecked empty removal is reachable
through this module API. The disposition and tests remain in the evidence archive.

A later [private native prototype preparation](../macos-staging-reuse/README.md)
builds from this policy; it requires its own native byte/lifetime proof.
