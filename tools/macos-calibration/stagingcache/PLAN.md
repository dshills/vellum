# Texture-only cached completed staging: byte correctness first

The original-WC reuse pair (1f37976) reduces CPU copy p95 2.953→0.813 ms but still
misses33.3ms presentation (38.273→34.969ms). Earlier broad transient-buffer cache
control also misses the gate. This fresh private candidate narrows cache selection
to the already byte-proved texture allowlist, preserves proven completed-only
reuse/limits, and must pass its own GPU byte proof before any framework timing.

Clone the exact pinned reuse prototype, original locked HAL29.0.3 registry archive,
Cargo cache and target into a new private root. Verify all original archive members
and its lock checksum; no network, shipping source/SDK/library/app mutations.
Introduce an explicit private HAL memory-flag bit only for tracked Metal D2
one-layer RGBA8/BGRA8 uploads, both pool modes. Metal consumes only the exact
TRANSIENT|marker and MAP_WRITE|COPY_SRC descriptor; configured default cache changes
that allocation only. Malformed marked descriptors record a mismatch/refuse.
R8, write-buffer staging, mapped staging and all other descriptors remain original.
The private marker is not a public SDK/API or cross-backend policy.

Configure before any device/resource. Reuse the immutable one-shot scalar cache
module; record actual Metal cpuCacheMode/storageMode only for marked allocations.
Counters must exactly equal eligible fresh resource IDs from the original closed
lifetime ledger (pool hits retain the same immutable raw buffer cache policy).
Pool ownership, completion fence, caps, fallback and teardown are byte-identical
to the verified prior prototype. Config remains quiescent, not concurrency safe.

One headless pool-off/on pair with default caching in both, original24-write matrix,
22 full GPU readbacks each including padded/BGRA/queued/large/oversize/R8/device/
unsent cases. Ninety seconds overall/30 per case, exclusive lease/owned groups,
source/archive/binary/log/SDK8321/configuration/geometry pins. Stop/seal first
invalid; no unchanged retry. No window/surface or physical pointer tests here.

Two new wrapper methods reuse the original synthetic matrix and reject eight
cache-scope/format/readback failures. Reuse original policy/lifetime/byte protocol
and three one-shot cache Rust tests, run separately. Private native/Go builds,
source-diff audit and internal/Prism Anthropic Claude Sonnet5.5 review precede
checkpoint/actual GPU run. Go overlay changes only the private probe and module
library path; actual application and frozen SDK remain untouched.

A valid result permits preparing one further cache-mode comparison with reuse
fixed enabled. It does not adopt this candidate or close any release item. Features
frozen, experimental pacing/rendering disabled,33.3ms gate unchanged, M3–M7 open.
No whole-GPU allocation balance, device fault/unwind or other hardware claims.

The unchanged CGo probe lives under testdata, excluded from normal Go package
discovery. The private compiler overlay supplies it only to the headless fixture.
