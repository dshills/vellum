# Private completed staging prototype: preparation

A texture-only completed-staging prototype is implemented and builds in an
isolated native library. **It has not executed GPU work or passed native byte
correctness yet. No shipping library or framework default changes.**

The [protocol](../../../tools/macos-calibration/stagingreuse/PLAN.md) requires one
bounded headless off/on pair before any timing work: 24 CPU texture writes and
22 full GPU byte readbacks per mode, explicit pending-lease separation, original
cache policy, row/offset/size cases, eviction, a second device, R8 exclusion and
unsent cleanup. Thirty seconds/case, ninety overall, stop first invalid, no retry.

Five tests pass for the actual private no_std-adapted policy; nine methods pass
for the lifecycle/protocol verifier, including synthetic documents explicitly
marked as non-GPU evidence. The earlier two compiler rejection controls are reused.
Native Rust and Go builds pass; a fresh recipe clone reproduces every source
modification byte for byte, with unchanged manifests/lock/HAL/cache policy. An
initial Vec/prelude compile error and a pre-execution format-narrowing revision
remain retained. No native correctness campaign has run.

The existing completed-submission path retires encoders before explicit cache
admission. Generic destruction never caches; failed/unsent work retains original
destruction. Raw idle handles hold no Arc<Device>, eviction/destruction happens
outside the cache mutex, and Device drains before backend/fence teardown. Scope
is the pinned M4 Pro/Metal host and allowlisted single-layer 2D RGBA/BGRA8 uploads.
Public mapped staging, write_buffer and other formats stay fresh. Idle limits are
eight entries/32 MiB per device, not total GPU memory. Counters prove only their
source-bound API lifetime, not physical allocation balance, injected GPU faults,
physical overlap, displayed pixels or presentation latency.

No compiler overlay, SDK byte, user data, Placement app, renderer/pacing setting
or release gate changes. Features frozen, 33.3 ms gate unchanged, M3–M7 open.

Complete 15-file Prism Anthropic Claude Sonnet 5.5 review and a focused four-file
follow-up are recorded. Ownership transfer and snapshot stability were hardened;
one scalar concurrency test passes in addition to policy5/protocol9. Remaining
findings are source-backed dispositions for the declared per-host, pre-device,
normal-execution protocol. It does not qualify concurrent configuration or poisoned/
unwinding native state.
