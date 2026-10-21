# Private framework transient-upload cache control

The serial offscreen control found fresh-copy p95 0.642–0.684 ms with shared
write-combined storage and 0.471–0.507 ms with shared default-cache storage. Its
subsequent copies were faster, but synchronization, working set and source data
differ from the framework. It does not explain the framework's 2.612 ms CPU-copy
p95 or justify a pool. Compare cache modes inside the actual private framework
before choosing any shipping policy. This is a new controlled boundary, not an
unchanged repetition of earlier latency failures.

Clone the completed private upload-observer source/core/cache/target trees;
preserve those originals and the frozen SDK. Clone exact locked wgpu-hal 29.0.3
from its registry source into a private path. Keep every package name/version and
all other lock entries; only that HAL package switches from registry to path.
Add an immutable, one-time pre-device cache-mode switch and six bounded scalar
counters. No pointer, buffer, device, image or window reference enters storage.

In Metal create_buffer, only the exact transient MAP_WRITE|COPY_SRC staging
descriptor participates. Both modes retain shared storage and original hazard
tracking, size, label, validation, creation/map/copy/flush/encode order, pending
write ownership and GPU-completion destruction. WC preserves original cache mode;
default changes only CPUCacheMode. Other resource descriptors retain original
selection. Eligible staging for write_buffer also participates, so this is not a
texture-only change. Record actual created Metal CPU cache/storage properties;
refuse wrong mode/storage, unconfigured creation, unknown policy, mixed modes or
counter bounds. Every texture staging record must be covered; extra eligible
write-buffer staging is retained in counts. No buffer pool or lifetime change.

Both modes run the same private archive and Go binary with the existing CPU
upload recorder enabled and the same three detailed-source compiler overlays.
View/routing/pixel-model bytes remain identical to the preceding fixture. Native
archive hash forces a Go relink; preserve compiler/package/overlay/source/binary
identities. Export diagnostic-only configuration and scalar snapshot ABI. Configure
before creating the session/device; snapshot after runtime shutdown and CPU trace
drain. No public framework API, SDK byte or app packaging change.

Run three pairs: WC/default, default/WC, WC/default. Each case keeps the existing
128 warmup and 120 measured requests at 120 ms, native actual-presentation and
foreground/state/accounting checks, final armed-frame/source-attempt/native-write
joins, CPU stage masks and full startup/tail retention. One attempt each, 40 seconds
per native case and 240 seconds overall including probes. Stop/seal at the first
invalid result, focus/config/pin drift, signal or deadline. Existing exclusive
lease, owned watchdog/process retirement, configuration/persistent geometry and
8,321-file immutable SDK checks remain. Human has explicitly signed out and
authorized continued testing until return.

Use focused scalar policy/schema/refusal tests and actual Rust/native/Go builds;
reuse parent upload/source accounting tests and exact view/routing proof. Run each
one-shot Rust configuration test in its own process. Internal review plus Prism
Anthropic Claude Sonnet 5.5 precede source commit and native execution. Separately
validate, review and commit results. No broad campaign or unchanged retry.

Report each run's source/CPU upload and actual presentation distributions. A mode
comparison does not establish zero observer overhead or ordinary-library/binary
equivalence. Standalone GPU byte readback supports the two Metal cache options,
but this framework journal does not independently verify physical displayed
pixels or full GPU bytes. Counter totals and process retirement are not allocation
balance or asynchronous pooling proof. No result automatically adopts the cache
policy or closes the shipping gate. Features frozen, experimental rendering/pacing
disabled, 33.3 ms receipt-to-actual-presentation gate unchanged and M3–M7 open.
