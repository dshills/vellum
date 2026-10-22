# Texture-only cached completed staging: preparation

The original-WC pool diagnostic (1f37976) reduces CPU copy p95 to0.813ms but still
misses33.3ms presentation at34.969ms. This new private experiment tests cache policy
only on the exact allowlisted texture allocations, preserving the byte-proved
pool/completion/limits/teardown. It is not adopted by the framework or app.

The [protocol](../../../tools/macos-calibration/stagingcache/PLAN.md) requires a new
headless pool-off/on full GPU byte pair, default caching both modes, before timing.
An explicit private HAL descriptor bit distinguishes allowlisted Metal texture
staging; write-buffer/mapped staging and R8 retain the original policy. Actual
Metal cache/shared-storage counters must equal fresh IDs in the closed eligible
ledger. A malformed marker refuses. Configure only before devices/resources.

All original HAL29.0.3 crate members/checksum and reuse-prototype preimages were
verified. Only the private marker/cache bridge and HAL path lock source change;
package versions/dependencies, pool/lifetime modules and original44-readback matrix
remain equal. The [native source diff](private-native-diff.patch) is reviewable.
Initial preparation expected an absent Cargo checksum metadata file and stopped
before private writes; corrected preparation verifies the locked crate directly.
No GPU attempt was retried or run before review.

Offline native and Go overlay builds pass. Two new wrapper tests reject eight
cache-scope/format/readback failures; the three unchanged one-shot Rust cache tests
pass in separate processes. The original Go byte fixture body remains unchanged;
only its private CGo probe and module library selection are overlaid. Source/SDK/
archive/binary/log/configuration/geometry pins, owned lease,90-second bound and
stop-first-invalid remain required. No GPU correctness/timing result claimed yet.

Features frozen, experimental rendering/pacing disabled, SDK/app/shipping library
unchanged,33.3ms gate unchanged, M3–M7 open. No ordinary-build equivalence, whole
GPU allocation balance, physical overlap, device fault/unwind or other hardware
qualification follows from this private control.

Internal layout review moved the unchanged probe under testdata before commit,
so ordinary Go discovery does not build a private CGo ABI fixture. The private
overlay and focused Go rebuild cover that correction; native sources unchanged.

## Full GPU byte result

Both pool modes pass in 4.654seconds (<90), with18owned groups retired.
The original24-write matrix passes22full GPU readbacks/mode:44readbacks and
97,912,700bytes total. Source/native/binary/log, SDK8321,
configuration and persistent geometry pins hold; eligible/device ledgers close.

Actual Metal default-cache/shared-storage counts equal exactly22fresh texture
resources off and16on, with zero wrong-cache, unconfigured or malformed records.
The enabled mode has6checkouts,5evictions,8peak idle entries/7,692,572bytes.
Queued distinct leases, oversize bypass, separate device, R8exclusion, padded
source/BGRA and unsent destruction retain the original strict checks. No claim
of physical GPU overlap, whole allocation balance, fault/unwind or UI latency.

This proves bytes/lifetime and actual cache scope for this private pair. It
permits preparing one bounded framework cache comparison with reuse enabled in
both modes; no shipping adoption, ordinary-build equivalence, performance pass
or release gate change follows. Features/defaults remain frozen, M3–M7 open.
