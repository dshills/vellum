# Stored-table performance measurement

This workload measures the P6 warm table-scroll target on the local macOS
backend. It does not establish the complete P6 table/tree gate, P11 endurance,
physical input-to-photon latency or professional application performance.
The unchanged provisional warm target is total UI work p95 ≤4 ms and p99 ≤6 ms.
A failed target remains failed; neither this runner nor its consumer adjusts it.

## Fixture and operation

`examples/loadedgrid -perf` constructs one native window with twenty stored
numeric columns, two frozen leading columns and a fixed header. Four separate
cases compare 10,000 and 1,000,000 actual stored rows at the same native viewport,
with fixed or stable-key indexed variable heights. Every record contains twenty
`int64` values generated with seed 7. Generation and worker-built initial
projection precede native scrolling and have separately reported durations.
The borrowed domain payload and projection order/inverse/prefix are excluded
from UI-retention interpretation and retain their own logical byte counts.

The intended content viewport is 768×624 logical grid pixels, with 96-pixel
columns, a 24-pixel header and fixed 12-pixel rows. The variable pattern is
9/11/13/15 pixels per four stable row keys, also averaging 12. The host's public
Style size uses device-independent pixels (Dp, 1/160in), rather than grid Px
(1/96in): 1299×1060 Dp includes the 12-pixel scrollbars. Small numeric typography
fits these compact cells. This is a numeric-table fixture, not the text-editor
or 1,000-control baseline. Window dimensions, actual scale, measured viewport
and clipped cell rectangles are retained.

Performance mode disables the native window-geometry saver before constructing
the fixture and keeps it disabled through native teardown. This process-local
switch prevents saved window geometry from affecting the requested fixture and
prevents measurement windows from persisting geometry. Setup records report the
actual switch as `geometry_persistence_disabled_in_fixture`; the previous switch
value is restored on return, including setup failures. Ordinary interactive
mode retains its normal geometry behavior.

Current records declare `window_geometry_guard: 1` on setup, ready and end.
Setup records the requested 1500×1250 Dp initial client size. After warm-up and
after measured work, the worker observes native client pixel and window-manager
sizes, render geometry, selected screen pixel/manager extent, screen/window
logical DPI and DPR. Queries wait for native main-queue publication, execute on
main without the application owner, and copy geometry under the native lock.
These endpoint reads and their JSON output occur outside recorded frame work.
They must agree; grid clips must lie inside the observed render area and grid
scale must agree with window logical DPI / 96. A mismatch refuses measurement.
Endpoints do not prove absence of a transient resize during the interval.
Historical records without this guard retain their historical observation scope.
The measurement worker is joined through application-close completion before a
successful runtime return. A queue that never publishes/stops still relies on
the fixture context and outer process watchdog; queued dispatch alone adds no
cancellation guarantee.

At least 400 cells, 50 row identities and eight column identities must have
nonempty measured clips after each completed native frame. Empty overscan
records and pool capacity do not count. A display/layout that cannot satisfy
this workload is refused instead of reducing it. Paired 10k/million cases must
have identical actual cell cardinalities, scale and normalized clip hashes at
each trajectory phase.

A 64-event zero-sum trajectory scrolls vertically by ±12 logical pixels and
horizontally by ±96 pixels at four positions in each direction. At least 64
unmeasured events, in complete cycles, warm this bounded working set; the
default is 128. Each measured event is an initialized MouseScroll admitted to
the originating native event deque and routed through the actual grid handler,
owner callback and native reconciliation. There is one traced event outstanding
per window. Recorder attribution requires one active native window;
multiwindow/integrated frame attribution remains open. The consumer requires one GridChange notification and the matching
completed tag/generation before the next event. It does not directly mutate the
viewport during measured scrolling or call descriptor callbacks.

New records declare `warmup_guard: 1`. The initial viewport must start at zero
offset, and the complete warm trajectory must return to the same viewport,
bounds and scale with exactly one notification per admitted warm event. Extra
transitions, offset drift and layout changes refuse the run before measurements;
the consumer never resets offsets to hide drift. Ready records retain start/end
viewports and generations plus the observed warm callback count. Stage and sample
guard/mode identities must agree. Historical records lacking this guard remain
historical observations; adding this check does not qualify an earlier refusal.

`-perf-novel` is a separate workload: after warm-up it jumps, unmeasured, into an
untouched row range and scrolls forward to newly visited numeric strings.
Fonts/device state are warm and numeric strings can coincide. It is not a
cold-font, shader-cache or guaranteed cache-miss qualification, and its samples
are not pooled into the warm P6 target.

## Timing boundaries

The opt-in bounded `FrameMetrics` recorder reports raw same-ticket durations:
input processing, owner transition, reconciliation, combined native layout/paint,
composition gathering, raster/upload and encoding/submission. `TotalUIWorkNS`
sums measured work regions after the routed event through its matching armed composition, including priority frames that run before owner reconciliation. These are wall
durations of measured work, not kernel/thread CPU accounting. Queue delay,
surface acquire wait and present-call duration are retained separately and
excluded from the total. The acquire observation includes surface mutex wait
and native acquisition attempts; reconfiguration/status/view work is charged
to encoding/submission. Only the native Present call is excluded there; texture
release/cleanup is charged to encoding/submission. The initial render-context
lock wait is part of queue delay, and asynchronous render/window/drawer lock
waits are not charged as UI work. No independent phase percentiles are added together.
The JSON summary key is `ui_work_ns`; it is the raw `TotalUIWorkNS`, not an
OS-thread CPU estimate. GPU execution, compositor presentation and photons are unavailable.

A current `InspectGrid` publication must match the completed recorder generation.
Inspection, JSON output, recorder polling and memory snapshots happen outside
the recorded work regions. No synchronous screenshot/readback is used. Dropped,
failed, missing, duplicate or pending recorder tickets invalidate the run.

`FrameMetrics.Poll(tag)` collects through the same synchronized path as
`Snapshot` and returns one detached `FrameMetricsPoll` with `Sample`, `Found`,
recorder-wide `Dropped` and `Pending`, and configuration `Failure`. Zero tags
and nil, zero or copied recorders are refused. Unknown and still-pending tags
return `Found=false`; a collected refused ticket returns `Found=true` with its
sample failure. Cancellation does not bypass the dispatch/owner terminal.
Collected ordinary records need no history-sized observer allocation; optional
source details are deeply copied. Poll retains every record and tag reservation,
including failed records, so duplicate admission and capacity limits remain.
The loaded-grid waiting loop polls its requested tag and retains a full final
snapshot to validate all history, dropped records and pending work.

Polling remains outside recorded UI work but can change process allocation and
GC pressure. Comparisons spanning observer implementations cannot isolate a
shipping-runtime improvement. The
[frame polling checkpoint](../docs/reports/macos-frame-poll/README.md) preserves
a matched native campaign and the still-failed unchanged warm targets.

## Memory and callback observations

Each sample distinguishes actually clipped visible cells, materialized
projection records and retained native cell/header pool counts. Native pool
counts are source-free observations, not byte attribution.

Each sample retains source RowKey/IndexOf/Cell counter deltas, exactly one grid
notification, process-wide Go TotalAlloc/Mallocs/GC deltas and measured geometry.
The process Go allocation interval includes observation helpers and other Go
runtime/native-adapter goroutines; it is not a precise framework-only allocation
counter. Domain generation and projection building are absent from these warm
intervals. Retained Go HeapAlloc/Inuse/Sys and goroutines are reported before data,
after data/index plus an explicit pre-mount GC, after warm-up plus an explicit GC,
and after measurement plus an explicit GC. These GC calls are outside measured
frames. The mount/warm delta includes Go caches and measurement state; it cannot
be attributed solely to retained controls. Native/GPU allocation attribution is
unknown. Host `ps` runs only after the completed end record, outside measured
native frames. Its end-only RSS and aggregate CPU sample are process-wide
observations, with the finite CPU accounting quantum retained. Baseline/warm
process RSS and process CPU deltas are unavailable; the native post-warm Go heap
snapshot remains. Setup/ready records still validate the workload, but never
spawn an external process observer. No subtraction of logical data bytes from
RSS pretends to measure private UI bytes.

## Reproducibility and results

Build an optimized release executable without `-race` or disabled optimization,
then explicitly opt into windows:

```sh
export VELLUM_CACHE_DIR=/tmp/vellum-p0-cache
. drivers/cogent/env.sh
(cd examples/loadedgrid && go build -o /private/tmp/vellum-loadedgrid-perf .)
python3 scripts/benchmark-grid.py --run --binary /private/tmp/vellum-loadedgrid-perf \
  --output /private/tmp/vellum-grid-performance-run --runs 5 --samples 120 --warmup 128
```

The output directory must be new. The runner preserves all raw stdout/stderr,
per-run JSON records, binary SHA, compact selected source hashes and build/host
metadata. The integration qualification separately binds prepared dependency
bytes; selected inputs are not a complete compiler/module-cache/SDK closure.
CPU/GPU/display/power metadata collection is read-only and may be unavailable.
It changes no power mode, scale, input source, permissions or OS preferences.

At least five release processes per case are required. Raw samples accompany
nearest-rank p50/p95/p99/max, counts, means, sample standard deviations,
coefficients of variation and the variation of per-run p95. Results retain every
outlier. A measurement report must distinguish observed target pass/failure,
workload fidelity and missing GPU/presentation/native-memory/endurance evidence.
These scripts establish a repeatable workload and recording protocol; actual
Mac measurements and optimization evidence are separate from CPU tests.

The runner's `--gpu-scene` selects the opt-in scene renderer through the consumer's
`-gpu-scene` switch. Setup, ready, every sample and end must declare the requested
boolean `gpu_scene_rendering` mode. Missing historical fields mean false; a GPU
campaign therefore requires explicit true on every record and cannot reuse older
unmarked samples. Ordinary target thresholds remain unchanged. The mode selects
the renderer, whose unsupported batches may fall back to CPU; it does not prove
that every batch was submitted on the GPU. Source-detail eligibility checks remain
separate diagnostic processes. A matched comparison uses the same optimized
binary, viewport and guarded trajectory, varying only this renderer option and
alternating process order. CPU/GPU measurements are separate populations.

The [single host-update comparison](../docs/reports/macos-grid-update/README.md)
removes proven duplicate pooled cell/header updater/style passes. Three
interleaved optimized fixed-10k processes per build preserve the 400-cell
geometry and reduce pooled elapsed UI-work p95 from 19.55 to 17.07 ms and p99
from 20.41 to 18.00 ms. The unchanged 4/6 ms targets still fail. This scoped
exploration neither meets the five-process requirement nor exercises the other
three cases, so P6 remains open. It does not change raster/upload invalidation
or prove a reduction in process-wide scheduler wakeups.

The [unchanged-refresh comparison](../docs/reports/macos-grid-refresh/README.md)
reuses validated grid projections and canonicalizes retained parent identity.
Three interleaved fixed-10k processes per build observe pooled elapsed UI-work
p95 17.233→14.624 ms (15.14%) and p99 18.954→15.566 ms with identical
400-cell geometry. The retained baseline maximum is 71.056 ms; its cause is
unclassified. Reconciliation p95 decreases 6.066→3.343 ms, while raster/upload
p95 is 8.748→9.026 ms. Stage percentiles are not additive. Original-overlay
regressions fail, final native/portable race checks pass, and eight software
frames match the previous checkpoint's repeated pixels. This remains an
exploratory comparison: both unchanged targets fail and the full P6 campaign
remains open. It makes no sprite/partial-upload or process-wide wakeup claim.

The [empty-sprite checkpoint](../docs/reports/macos-empty-sprites/README.md)
omits a scene-sized transparent layer when no active sprite can draw. Active
cursor and drag callbacks retain their every-composition behavior. Three
interleaved fixed-10k processes per build observe pooled UI-work p95
14.178→13.090 ms and p99 14.891→13.294 ms with identical 400-cell geometry;
raster/upload p95 is 8.907→7.406 ms. Other recorded stages increase, and their
percentiles are not additive. Both unchanged 4/6 ms targets still fail. Software
pixel/copy proofs do not measure native GPU upload counts, partial uploads or
process-wide wakeups. Full P6 remains open.

## Diagnostic source costs

`-perf-source-details` requires `-perf` and selects the separate detailed frame
recorder. Setup/ready/sample/end agree on `source_details`; ordinary historical
records missing that flag mean false. Warmup+samples+2 must fit the 4096-ticket
diagnostic capacity. Each successful Native sample exposes at most 32 scalar
source records across its complete ticket. Availability, ordinals, nested
intervals, full-input logical bytes and aggregate counters are validated.
Source details are rejected by ordinary warm-budget validation even when CPU
profile records are separately allowed. The explicit CLI mode always reports
DIAGNOSTIC with warm_target_met false.

All nested costs follow [the rendering contract](rendering.md). Image API time
and attempted logical RGBA input bytes cannot identify confirmed upload success,
physical transfer bytes or GPU completion. Unknown/custom sources may copy
multiple differently sized images; their aggregate bytes must not be inferred
from a fabricated common extent. Detail elapsed timings include diagnostic
overhead and remain separate from ordinary release comparisons and the full
five-process P6 qualification.

```sh
python3 scripts/benchmark-grid.py --run --source-details \
  --binary /private/tmp/vellum-loadedgrid-perf \
  --output /private/tmp/vellum-grid-source-costs --runs 5 --samples 120 --warmup 128
```

## GPU scene host diagnostics

The same explicit source-details recorder exposes `FrameSourceSample.SceneHost`
as a detached `FrameSceneHostSample`. `Available` identifies an observation from
the GPU scene renderer, including an early CPU fallback. Unavailable records
contain zero durations/counters and false flags. Attempt flags distinguish a
stage that ran from one that never started; an attempted interval can be zero
at the clock's resolution. The existing 32-record shared ticket cap and
4,096-ticket diagnostic capacity remain. Records retain no application, scene,
image, font or GPU-handle references.

| Interval | Measured host work | Parent |
| --- | --- | --- |
| `CaptureNS` | `ScenePacket` capture | source `DrawNS` |
| `EncodeNS` | `EncodeGPU` serialization | source `DrawNS` |
| `LeaseNS` | synchronous scene-lease draw admission | source `DrawNS` |
| `ResourcesNS` | resource preparation callback, including initialization, seeding and rendering | `LeaseNS` |
| `InputNS` | input-buffer retirement, admission, allocation and checked queue writes | `ResourcesNS` |

Capture, Encode and Lease are disjoint portions of source Draw. Resources and
Input are nested subsets. CPU fallback raster/copy intervals are disjoint from
the attempted GPU portions; other source work can remain outside these
subintervals. These observations neither add to the seven-region
`TotalUIWorkNS` sum nor measure OS-thread CPU, GPU execution or presentation.
Diagnostic clocks and storage affect the samples, which cannot qualify the
ordinary 4 ms p95 / 6 ms p99 targets or an optimization gain.

`InputWriteAttempts` and `InputWriteErrors` count actual checked input-write
calls and their returned errors. `LogicalInputBytes` sums the supplied slice
lengths, including a refused call, for Ops, Rows, Refs, Data and the 16-byte
configuration buffer. It excludes retained buffer capacity, seed-image writes,
physical traffic and confirmation of transfer. An accepted changed source has
five input-write attempts and one `QueueSubmitAttempts`; this submit counter
records a call without claiming completion. A partial refusal can retain fewer
writes without an input-write error when admission/allocation failed instead.
These checked input errors are separate from legacy image API errors, whose
`FrameSourceDetails.UploadFailuresAvailable` remains false.

`DrawAccepted` records successful scene-lease admission. `Reused` identifies an
accepted unchanged source with no capture, encoding, input writes or new compute
submission. Captured batches report `ImageChanged=true`; an early fallback can
report changed commands before capture starts. `Fallback` preserves attempted
GPU stages together with the CPU
raster/copy observation. A post-submit lease refusal can therefore report one
submit attempt, `DrawAccepted=false` and `Fallback=true`. Counter snapshots and
JSON copies remain independent; ticket accumulation refuses unrepresentable
unsigned counter sums.

The fresh diagnostic validator requires the complete scalar `SceneHost` schema,
type/bounds checks, stage ordering, nested/disjoint intervals, reuse/fallback
consistency and checked attempt/byte sums. A CPU-renderer case refuses available
GPU host observations. Archived earlier diagnostic records retain their recorded
validator and source association; they are not relabeled as this richer schema.
Regression requirements include actual checked-write refusal after a prior
successful write, exact slice-length accounting with retained capacity headroom,
post-submit lease refusal, unconditional owned-resource restoration, recovery
pixels, and composer-to-ticket nesting without duplicate drawing or UI work.
Full native preparation, Go tests and admitted GPU execution must establish those
requirements before the new observations support a measurement report.

## Diagnostic CPU profiling

`-perf-cpu-profile /private/tmp/new-grid-profile.pprof` is an optional process-wide
Go CPU diagnostic. The path must be new; the writer has a 64 MiB byte cap and
reports truncation/write failure. Profiling begins after warm-up and stops after
measured samples, before the final GC. It includes measurement observers and
concurrent runtime/adapter work; native C/GPU/compositor costs are not reliably
attributed by Go samples. The flagged process is never a warm-budget result:
setup/ready/end records expose `cpu_profile`, and ordinary five-run validation
refuses profiled records. The benchmark CLI never enables this flag. Profiled
exploration has its own raw records, source/binary/profile hashes and scope;
its timing samples are not pooled with the unprofiled qualification runs.

## Renderer row-range comparison

The [row-range checkpoint](../docs/reports/macos-raster-rows/README.md) removes
surface-height-dependent index work for small paths in owned raster scanners.
Five optimized processes per build across all four ordinary table cases retain
4,800 samples and matching paired geometry. Native UI-work differences are small
and mixed; the variable-million case rises slightly and every run still fails
the unchanged targets. Software span/mask/pixel evidence and microbenchmarks are
separate from frame timings. This qualifies the comparison's table protocol,
while P6 tree latency, broader memory, cold caches and sustained work remain open.

## Diagnostic grid input provenance

`-perf-input-details=/new/path.json` requires `-perf` and a new exclusive file.
It cannot combine with CPU profiling or source-detail timing. The separate
`NewFrameMetricsWithGridInputDetails` recorder admits at most 4096 frame tickets
and preallocates `3*tickets+256` scalar input records. Grid/window identifiers
are at most4096 bytes. No record is overwritten; identifier/capacity refusal
is explicit and failed diagnostic evidence is still saved. Ordinary
`NewFrameMetrics` allocates no input ledger and adds no input clocks.

The recorder binds before native grid mount. Admissions copy requested logical
deltas, encoded physical deltas, completed scale, viewport and generation.
Actual wheel dispatch copies the current event tag, handler scale, physical
deltas, bounded integer steps and signed remainder before/after arithmetic.
Owner transitions reuse their existing state reads to report origin, input
sequence, generation, viewport and changed/unchanged/refused result. Scrollbar
and layout origins remain distinct. Records contain no event/model/widget,
callback, cell values or native pointer. Native wheel/style reads stay at the
originating scene boundary; owner observers never read those mutable fields.

The independent ledger mutex is released before posting, application callbacks
or trace calls. Admission records append after releasing the metrics mutex;
native delivery can append first. Tags and ingress sequences establish links;
append order is not wall-clock time or native FIFO ordering. Tag zero means
untraced input, not a claim about desktop/user origin. A snapshot does not prove
FIFO exhaustion; last-delivered remainder is not an independent native barrier.

After native Run retires, the exclusive file saves the detached ledger, including
failed warmup runs. Setup/ready/sample/end identify this diagnostic mode.
Ordinary validation refuses input-detail samples, even if profile or source
diagnostics are separately allowed. Explicit diagnostic validation and elapsed
observations never qualify the unchanged warm latency targets.

## Covered-fill comparison

Stage 38 measures adjacent uniform rectangle write elision separately from
complete native frame latency. A same-work pure replay microbenchmark compares
the original dispatch and optimized backend. The native comparison uses fresh
ordinary baseline and candidate processes with identical cases, closed guarded
warmups, 120 measured samples and five repetitions per case/build. Every raw
observation is retained; actual geometry must match across builds and between
10,000 and 1,000,000 row cases. Instrumented source/CPU diagnostics remain
separate. Targets stay p95 <= 4 ms and p99 <= 6 ms; a CPU microbenchmark gain
cannot qualify those targets. See [results](../docs/reports/macos-covered-fills/README.md).

## Native accessibility publication comparison

Compare the complete native accessibility work reduction against the preceding
committed release build, with all source changes included and no diagnostics in
either measured process. The native benchmark retains 400 clipped cells, atomic
scoped geometry publication and current action guards. Omitting the legacy
geometry that Cocoa already discards does not reduce the semantic workload.
Separate Cocoa acceptance verifies live values, frames, stable identities,
replacement, actions and closure. Payload ownership fixtures do not qualify
process/native/GPU memory or establish an ordinary frame improvement.

Matched baseline/candidate timing uses all four stored-table cases, five fresh
interleaved processes per case/build, 120 samples and 128 guarded warmups per
process. Preserve every raw record and require exact paired geometry; CPU
profiling remains separately labelled diagnostic evidence. Existing 4 ms p95 /
6 ms p99 targets and the full P6/Mac acceptance scope remain unchanged.

## Differential allocation diagnostics

`-perf-alloc-profile=/new/prefix` requires `-perf` and a separate process from
CPU profiling, source timing and input-ledger diagnostics. It reserves new
exclusive `prefix.before.pprof` and `prefix.after.pprof` files with mode0600
and a64MiB cap per snapshot. Refusal never overwrites existing files; failed
captures preserve their owned empty/partial outputs and fail the run.

The baseline captures the cumulative Go `allocs` profile after guarded warmup;
the final snapshot follows the measured events and observer work, before the
end record. Each capture forces two GCs outside frame timing because Go memory
profiles can lag two GC cycles. The process's `runtime.MemProfileRate` stays
unchanged and is recorded in setup. Use `go tool pprof -base before.pprof
-alloc_space after.pprof` (or `-alloc_objects`) to subtract startup/warmup.

Setup, ready, every sample and end explicitly identify `alloc_profile`.
Ordinary validation refuses these records, independently of CPU/source/input
permissions. Only explicit allocation-diagnostic validation admits them, and
`warm_target_met(..., alloc_profile=True)` always returns false. The ordinary
benchmark CLI never enables this diagnostic. Preserve exact source/binary,
snapshot hashes and raw records for attribution. These are sampled process-wide
Go allocations, including observers, concurrent adapter work and profile
encoding; they cannot establish framework-only, native/GPU or retained-memory
cost, or satisfy P6 latency targets.

## Accessibility capacity and semantic string comparison

The [service/allocation checkpoint](../docs/reports/macos-service-performance/README.md) compares exact completed-grid capacity reservation and owned semantic string encoding against the preceding committed build. Five interleaved process pairs per stored case retain all 4,800 samples with exact paired geometry. Process-wide allocation is lower, while frame timings are mixed and all unchanged warm targets remain failed. Separate sampled Go allocation profiles remain diagnostic; cache ownership, visible semantics and action guards are unchanged.

## Benchmark process retirement

On Mac/POSIX the runner retires its owned process group after every launched
case, including a successful direct-child exit and pipe EOF. Descendants can
close inherited output and outlive the direct child. Retirement runs after
measurement; it does not alter frame timing regions or the acceptance workload.
Cleanup failure rejects the case, preserves the direct child's status and raw
records, and records the exception type and message. Windows retains its existing
failure/live-child tree-termination policy.

The [retirement regression](../docs/reports/macos-benchmark-retirement/README.md)
uses real CPU-only processes with a successful producer and a surviving
same-group descendant. It verifies termination, permitting an orphan zombie
awaiting OS reaping. This does not qualify native resource inventories, current
frame performance, app-wide process absence or any broader Mac gate.

## Native input byte fixture

The input-attempt diagnostic fixture uses three ordered paint operations so an
encoded input requires capacity beyond its payload length. A CPU encoding
regression checks that precondition through the same fixture used by the native
GPU test; the native test separately requires actual retained headroom and exact
success/refusal/recovery byte counts and reference pixels. The original single-
operation fixture failed its headroom precondition in the admitted native run;
that failure remains recorded. This fixture correction changes no production
renderer, latency target, compiled-component or distribution qualification.

## Scene encoding tile-bin counts

Owned stage0075 changes CPU-side scene encoding work inside the GPU-requested
source path. Original tile-reference/encoded-byte admission still charges every
operation, including covered clears later omitted from compact output. Dense
batches (retained references > twice the tile count) use rectangle differences
and an in-place prefix pass; sparse batches keep direct counting. The existing
count array becomes fill positions afterward. No new cache or retained backing
is introduced. Encoded ABI bytes, painter order, detached ownership and packet
replay/reset semantics must match the frozen pre-change encoder exactly.

Exact admission/branch boundaries, mixed operations, frozen bytes, a wrong-corner
control and full actual-driver software pixel comparisons precede acceptance.
Supported software DPI96/160/192 does not qualify native DPI162. Synthetic
encoder microbenchmarks are separate from ordinary native frame targets; all
historical failures and full P4/P6/P11/Mac qualification remain open.

## Ready accessibility duplicates

The [ready-snapshot checkpoint](../docs/reports/macos-ready-snapshot/README.md)
removes repeated JSON encoding when owned ready semantics exactly match the
existing validated baseline. Its offscreen duplicate microbenchmark includes
fresh snapshot/handler allocation. It does not measure a full native frame or
change any workload, timing boundary or target above. Current ordinary failures
remain open until the candidate receives the required native comparisons.

## Rejected grid accessibility construction experiment

The direct owner-snapshot construction benchmark compares against a frozen prior
constructor and frame stamper on the same 400-cell completed measurement. It
includes node construction, handler merge and stamping, but excludes JSON/native
publication and full refresh. Time and allocation improvements in that benchmark
do not qualify a native frame budget. Native comparisons must retain ordinary
runtime settings, matched clip geometry and the existing p95/p99 thresholds.
In the experiment, the prior implementation served as the geometry fallback
when the completed measurement changed during construction.

The [recorded experiment](../docs/reports/macos-grid-ax-construction/README.md)
reduced isolated construction time/allocation but was slower in all five ordinary
native p95 comparisons. The runtime candidate was reverted. Its archived diff
retains the benchmark, parity tests and frozen prior implementation; none of this
qualifies or changes the current runtime's frame result.

## Pane-aware managed grid painting

A full managed pane repaint may omit a passive cell/header's StandardBox when
that box would only replace the pane's freshly painted opaque uniform background
with the same color. Text shaping, cell widget identity, per-cell clipping,
selection/cursor borders, measurement and accessibility retain their ordinary
paths. This removes redundant draw construction before raster capture; it does
not relax scene operation, byte, work or tile-reference limits.

Admission requires a live ready binding without an editor; a normally painted
pane whose actual integer write rectangle matches its clipped bounding box;
zero box margins, borders, radii and shadows; unit opacity and no state layer;
and known opaque uniform colors. Every active ancestor context must have an
identity cumulative transform, integer rectangular bounds, no rounded corners,
mask or clip path. Only normal managed Text leaves in their expected pane order
participate. A linear row/column interval proof checks disjoint cell clips,
including preceding selected/cursor painters. Foreign, overlapping or reordered
content refuses the entire pane proof. Decorated/differently colored cells keep
their original box paint. Backgrounds are derived from current style and the
already rendered parent's current background, not a stale pane cache.

Temporary RenderBox changes apply only within the complete pane frame
traversal and restore on normal or abnormal exit. An individual Text repaint
therefore retains its normal background clearing. The existing experimental
verified-scroll protocol keeps its original painter. A private reference switch
exists for original-painter regression evidence; it is not a public configuration.

Original CPU pixels and captured packet replay must match across scrolling,
frozen panes, light/dark backgrounds, selection, disabled state, fresh background
changes and isolated text repainting. The matrix includes 162 DPI and requires
actual command reduction there. Native timing and actual GPU-route evidence
remain separate; this optimization alone does not qualify the unchanged 4/6 ms
frame targets or memory/endurance requirements.

## Passive text batches and immutable bitmap resources

Stage0079 groups contiguous eligible passive text leaves within the same full
managed pane repaint. The driver first applies the background proof above.
Custom-frame traversal then preserves visibility, dirty consumption, background
updates, current shaped layouts, painter order, parts and scrollbars. Pending
shaping, selected/highlighted or decorated cells, children/parts and debug views
use ordinary widget rendering. Other renderers decline the explicit capability;
rasterx and the GPU wrapper support it, including ordinary CPU fallback.

This is a transient command representation. Public PaintText layouts remain
mutable and are read each frame. No cross-frame layout identity or translated
text-packet cache is introduced. Capture charges the original logical context,
text and pop commands and retains full text/outline preflight and original glyph
placement arithmetic. Both bitmap and small-outline text use the batch; the
existing small-outline span cache and shader remain unchanged.

Only immutable zero-origin masks returned by the private bitmap glyph cache may
use a sealed snapshot bank. Each renderer retains at most 1,024 entries and
8 MiB of charged original mask backing, detached snapshot backing and entry
structs. Queue/map/allocator metadata and other caches are separate. Every packet
charges each immutable snapshot once and owns its bytes through ordinary Go
references after bank eviction, resize or release. No accessor exposes mutable
snapshot pixels. Arbitrary public GlyphMask/image inputs retain copy-and-compare
ownership. No snapshot captures fonts, layouts, applications or GPU handles.

The [checkpoint](../docs/reports/macos-text-batch/README.md) records exact original
CPU/packet pixels at five DPI values, actual GPU-wrapper batching at 162 DPI,
mutable layout visibility, source detachment, clipping, legacy encoding, logical
admission, refusal/fallback and cache retirement. Five matched fixed 10k/400-cell
native pairs reduce median run-p95 from 5.168 to 4.951 ms and observed process Go
allocation by about 20%. Every candidate p95 still misses 4 ms. Diagnostic GPU
captures and native compositor pixel tests remain separate from ordinary frame,
presentation, physical pointer, memory and full Mac qualification.

## Rejected packed GPU upload experiment

The [packed-upload record](../docs/reports/macos-packed-upload/README.md) preserves
an implemented and reviewed one-write transport candidate. Exact GPU pixels,
ownership, padding and failure/recovery checks pass, but three of five ordinary
paired p95s are higher (median 5.189→5.265 ms). Its small diagnostic input-phase
reduction does not justify adoption. The candidate, decoder and preparation stage
are reverted; the active renderer retains five checked input writes and stage0079.
No frame, allocation, memory or Mac acceptance requirement changes. Current
profiles point to synchronous accessibility snapshot/serialization/publication
and full capture as the remaining investigation, without shifting work outside
the measured frame or weakening native transaction/fallback semantics.

## Builtin accessibility semantic encoding

The [serializer checkpoint](../docs/reports/macos-ax-semantic/README.md) preserves
the fixed private JSON schema in owned strings, with whole-snapshot standard
encoder fallback for custom/unsupported/nonfinite values. No cache authority or
native publication changes. Full driver race, default/legacy JSON byte oracles,
vet and release build pass. Five matched ordinary pairs improve median run-p95
4.875→4.788 ms; four of five improve. Observed process allocation is about 10%
lower, including fixture/runtime work. Every p95 still exceeds 4 ms; no table
or broader Mac gate closes. Separate diagnostics confirm 120 GPU captures per
build, all accepted, no fallback, unchanged logical input. These are not physical
pointer or presentation observations. Stage0079 prepared dependencies are unchanged.

## Full current matrix and bounded next experiment

The [bd672a6 matrix](../docs/reports/macos-grid-matrix/README.md) completes all
eight fixed/variable 10k/1M CPU/GPU cases, five 120-event runs, 128 warm each. All 40
p95s fail 4 ms; GPU case medians 4.71–4.82 ms and CPU 5.61–5.77 ms. Preserve the GPU
1M p99 failure 8.617 ms. Every paired 10k/1M visible geometry check passes. Four
separate 120-event GPU diagnostics show actual accepted captures with no fallback;
those timings are excluded from ordinary acceptance. Current profiles are
diagnostic and retain the native sample's redacted-path check failure plus a
separate owned-PID/Mach-O-UUID assessment.

Next implement M3-A's owned segmented ready snapshot under the early-fallback,
immutable per-grid storage and action-generation contracts in macos-delivery.md.
Require positive fast-route evidence and full old-path equivalence. Adopt only
a >=0.4 ms median run-p95 reduction in five matched pairs with >=4/5 improvements
and no correctness/retention/tail regression. Otherwise reject the experiment;
a microbenchmark gain alone is insufficient. These are experiment-adoption
criteria, not changes to the original 4 ms/6 ms targets or the full Mac finish line.

## Managed body-cell layout checkpoint

The managed grid can retain current-projection body-cell layout values instead
of one native Text widget per cell. Stable row/column identities retain unchanged
shapes, while headers, host input/accessibility, panes, scrollbars and editing
keep their native ownership. Only supported passive styles and renderers admit;
unsupported or invalidated configurations restore the original widget route.
`GridInspection.RetainedCells` counts current cell representations, including
value layouts. It must not be used as proof of native widget cardinality.

The [cell-surface checkpoint](../docs/reports/macos-grid-cell-surface/README.md)
records a five-pair prototype improvement in both CPU and GPU modes, separate
positive native widget/layout counters, corrected source integration, original
pixel/geometry/routing comparisons and the actual final qualification scope.
The prototype median-run p95 improves 5.929751→4.811373 ms CPU and
5.022791→4.249374 ms GPU, with 5/5 improved pairs each. All twenty run-p95s still
fail the unchanged 4 ms threshold. Whole-process allocation p50 rises about
15.21% CPU/10.17% GPU despite lower post-GC terminal Go heap; no allocation,
native/GPU resource or full M3 gate closes from this comparison. Corrected
integration timings must retain their separate source/binary identity.

The value route owns no hidden widget substitutes, offscreen-row cache, font
identity cache or paint-fragment cache. Late invalidation withdraws completed
geometry/action admission and requests owner reconstruction; failed/incomplete
values are not painted as completed cells. A late refusal may omit body values
for that invalidated paint before the ordinary frame is rebuilt. The original
route remains the independent pixel/geometry/style oracle. Explicit legacy
widget test helpers do not change normal NewManual routing.

## Current stage0080 eight-case matrix

The [current-grid matrix](../docs/reports/macos-grid-current/README.md) records
all eight CPU/GPU fixed/variable 10k/1M cases on the exact current `a2b4143` Go
implementation over the retained private SDK graphics library: five ordinary
120-event processes per case, 128 warm events each and 4,800 raw samples. All
twenty ordinary paired 10k/1M geometry arrays match. All forty p99s meet 6 ms;
seven p95s meet 4 ms and thirty-three fail. Only fixed-height million-row GPU
meets both targets in all five runs (median run-p95 3.840415 ms). This closes
the missing current table measurement, not the full M3 performance gate.

Separate eight-case final-composition diagnostics prove the actual value route.
The initial intermediate-reuse validation failure remains FAILED/sealed, with
six saved raw runs revalidated under a reviewed final-scene freshness rule;
only the remaining 46 cases execute in a new scope. No diagnostic timing is
pooled into ordinary results. Allocation captures identify the per-cell paint
style copy as a substantial new allocation site. A per-render style ownership
fix is proposed; no cache, runtime fix or new allocation/latency pass is adopted
by this result checkpoint. Existing targets and broader Mac gates remain.


## Render-local paint ownership experiment: rejected

The [paint-style experiment](../docs/reports/macos-grid-paint/README.md) replaces
per-cell style copies with four fresh copies per pane render. Compiler escape
output, retained paint/ScenePacket lifetime and independent pixel/geometry tests
at five DPIs, driver-root/example race checks, vet and builds pass. Separate
native CPU/GPU diagnostics confirm the value route and fresh final pipeline.

Five ordinary fixed-10k pairs per mode use the current `81b85e39…` baseline and
an isolated candidate, with exact geometry and 120 measured/128 warm events per
process. Median run-p95 improves CPU 4.748167→4.690332 ms and GPU
4.254502→4.104204 ms. CPU improves 3/5 pairs, GPU 5/5; CPU median p99 worsens
5.162665→5.243124 ms. Both miss the frozen 0.4 ms gain; all twenty ordinary p95s
exceed 4 ms, all p99s meet 6 ms. Ordinary median per-run allocation p50 falls
4,160,840→3,457,944 B CPU and 5,749,032→5,043,288 B GPU. Separate sampled process
profiles show the intended paint allocation reduction, without qualifying a
framework-only budget or latency pass.

Reject the private candidate, preserve all evidence and stop style follow-up
tuning. Exactly 26 native processes complete (2 diagnostics +20 ordinary
+4 allocation captures); the reviewed conditional 36-case wider matrix remains
unexecuted. No source/prepared/SDK runtime change is adopted. The preceding
`a2b4143` full matrix remains the current production result; all unchanged M3
and broader Mac gates remain open. Next complete native M2 integration, then
choose a materially different measured M3 implementation.
