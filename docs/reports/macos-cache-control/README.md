# Private framework upload-cache control result

All three predeclared pairs are valid. **Default caching consistently reduces CPU
copy cost inside this private framework process, but does not meet the unchanged
33.3 ms presentation gate.** No shipping cache policy or buffer pool is adopted.

| Pair | Mode | CPU copy p95 | Go image API p95 | Receipt→actual presentation p95 |
|---|---|---:|---:|---:|
| 1 | writecombined | 2.577 ms | 2.844 ms | 38.421 ms |
| 1 | default | 1.631 ms | 1.869 ms | 36.084 ms |
| 2 | default | 1.755 ms | 2.011 ms | 35.993 ms |
| 2 | writecombined | 2.436 ms | 2.677 ms | 36.370 ms |
| 3 | writecombined | 2.433 ms | 2.660 ms | 37.010 ms |
| 3 | default | 1.793 ms | 2.035 ms | 36.242 ms |

Pair 2 reverses mode order. Both modes use the same private library and Go binary,
with detailed Go source and native CPU observers enabled. Every value describes
its own 120 measured requests. Differences between run percentiles are not paired
sample deltas, additive phase quantiles or calibrated observer overhead. Default
runs still show 35.993–36.242 ms p95 presentation; all six diagnostic cases exceed
the existing 33.3 ms number. This does not qualify an ordinary build or demonstrate
physical input/displayed-pixel correctness.

Every case observes 517 eligible buffers: all use requested cache mode and shared
storage, with zero mismatch/unconfigured counts. Each retains 254 CPU texture-write
records, including six startup and 248 uniquely joined cohort writes. The extra
263 eligible buffers are other transient staging, including write_buffer; this
experiment changes their cache mode too and is not a texture-only comparison.
Every texture write is accounted against its source-image attempts/bytes, receipt
and final armed acquisition. Startup/tail rows and all warmups remain retained.

The control filters the exact transient MAP_WRITE|COPY_SRC descriptor. Storage,
hazard tracking, size, label, validation/errors, creation/map/copy/flush/encode order,
pending-write ownership and GPU-completion destruction remain unchanged. Only CPU
cache mode differs. Mode is configured once before creating the session/device;
fixed scalar counters retain no resource references and snapshot after runtime
shutdown/CPU trace drain. Both modes' actual Metal cache/storage properties pass.
Numeric enums are independently verified against locked objc2-metal 0.3.2 bindings
and the Apple SDK (default cache/shared storage 0, write-combined 1).

The campaign runs 21 October 2026, 22:40:09.998520–22:43:27.349976 UTC,
197.351456 seconds within 240. All 54 owned native/probe process groups retire.
Source/native/binary/overlay/log pins, before/after configuration, persistent
geometry and immutable SDK 8,321-file inventory pass. No original source tree,
frozen SDK, Placement app or shipping library is modified.

Preparation uses a private clone of the prior upload observer and the exact locked
wgpu-hal 29.0.3 archive: all 84 original archive members/checksum verified, only two
HAL files changed plus the scalar module. Only HAL's lock entry changes from
registry to private path; package names/versions and all other entries are equal.
The original Core observer and exact view/routing/pixel model are reused, together
with the three detailed-source overlays and archive-derived forced Go link ID.
Three Rust configuration tests run in independent processes; three new Python
methods cover both modes, ten invalid counter rows and parent upload failure.
Actual Rust/native/Go builds and full 18-file Prism Anthropic Claude Sonnet 5.5
preparation review precede source checkpoint 9acc3ab and execution. An initial
apply-script assertion about indentation failed before any private source write;
it was corrected and retained. No native case was retried.

The [protocol](../../../tools/macos-calibration/cachecontrol/PLAN.md),
[build identity](build-identity.json), [exact summary](summary.json) and indexed
archives retain raw journals, configuration, failures and review dispositions.
The existing strict base classifier separately reproduces the reported presentation
distributions; the CPU-source classifier supplies accounting, not a gate waiver.
Standalone full GPU readbacks for the two Metal options are separate evidence;
this framework journal does not independently verify actual GPU bytes or pixels.
Process retirement and scalar counts do not prove native allocation balance.

Next inspect a bounded staging-reuse design that returns buffers only after the
existing GPU-completion retirement boundary. The standalone first/subsequent-copy
result motivates that investigation; it does not prove asynchronous pooling safe.
No pool or ownership change is implemented or selected by this checkpoint.
Features stay frozen, experimental rendering/pacing remains disabled, the 33.3 ms
gate remains unchanged and M3–M7 remain open.
