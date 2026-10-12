# Rendering and resource failure contract

## Optional source diagnostics

`NewFrameMetricsWithSourceDetails` explicitly enables per-source native work
observations with recorder capacity 1..4096. Ordinary `NewFrameMetrics` keeps
its existing 1..65536 bound and no source clocks/counter collection. Detailed
composition preserves source order and each actual raster/Copy/End call. The
whole draw batch retains originating renderContext and native-window ownership.
Separate drawer counter locks never cross source callbacks; deferred cleanup
clears diagnostic state through panic and Goexit.

One ticket may include intermediate compositions before its final armed frame.
At most 32 scalar records span that complete ticket, identified by one-based
Composition and zero-based Source ordinals. Numeric Kind values distinguish
other/scene/sprites/scrim/insets as 0/1/2/3/4. No source/image/widget/user-text
references are retained. Snapshot returns deeply detached records. Missing
details, record/count/byte overflow or unsupported measured composition refuses
diagnostic completion without omitting source rendering or submission.

DrawNS encloses the actual source. Known paint sources expose lazy RasterNS
and CopyNS; scrim/insets expose CopyNS. Unknown breakdowns have explicit false
availability. Existing-and-unchanged Go images count cache hits. The backend
SetFromGoImage API is timed only when called, with attempts and checked logical
4×full-input-width×height RGBA bytes. Uniform Fill and resident textures need no
Go-image attempts. Unknown drawers have unavailable image counters.

These counters do not claim successful queue writes, physical transfer bytes,
GPU execution, or presentation. Upload failures remain unavailable because the
existing Value API logs errors without returning them; this checkpoint does not
change that policy. Raster/Copy/API durations are nested subsets of DrawNS and
RasterUploadNS, never additions to TotalUIWorkNS. Diagnostic overhead affects
elapsed time; detailed samples cannot qualify ordinary latency targets.

Native windows default to requested painting instead of refresh-rate ticks.
`Driver.ContinuousRendering` explicitly restores continuous ticks for applications
that require them. Model invalidation and native input/resize events request
frames; active cursor/sprite timers use the backend's delayed paint requests.
GPUCanvas.RequestFrameAfter schedules one frame for an application-owned animation.
It is called on the originating window owner; it does not invalidate a model.
Static work must not keep an unconditional refresh ticker running.

GPUCanvas.RebuildResources rebuilds mesh/material resources from the unchanged
scene model on a healthy live device. Native layout updates resize the offscreen
render target; viewport/camera size is derived from post-layout pixel dimensions.
Snapshots are explicit synchronized GPU readbacks and never run during ordinary
composition. The API returns ErrFrameNotReady before resource construction.

The owned native patch supplies device-loss subscriptions shared by borrowed
Device copies. Expected device destruction is ignored. Unexpected loss is retained
once, observers run without the health lock, and unregister cannot recall already
copied callbacks. Driver callbacks only trigger failure/shutdown and never enter
application ownership from the foreign GPU thread. Window closing disables its
subscription before native resource destruction. Snapshots/resource rebuild refuse
a lost device with ErrDeviceLost. Run returns this actionable error and cancels
application work; applications retain responsibility for durable document state.
Automatic physical device recreation/process restart is not yet implemented.

## Qualification

Target idle CPU is below 1% of one core for a static unfocused fixture. Measure
quiet intervals with the SDK harness and retain host/workload/source metadata;
short accounting samples alone cannot certify this target. A 60 Hz frame budget
is 16.67 ms, but synchronous readback is not ordinary presentation latency.
No p99 frame/input or global memory acceptance is inferred from readback medians.

Native checks must verify idle paint wake counts, invalidation after idle, explicit
resource rebuild, resizing, and window shutdown. Backend race tests cover shared
loss state, concurrent subscription/cancel and reentrant observer registration.
Physical GPU failure, changing displays/DPI, occlusion and minimize/restore need
separate runtime evidence; the loss-state tests do not simulate a hardware reset.
Production memory/font budgets and representative sustained workloads remain open.

## Composition ownership

Async composition owns the native renderContext through lazy raster rendering
and the complete GPU draw batch, preventing resize/destruction from replacing
images or releasing texture views mid-batch. Inline resize composition uses the
context already owned by its caller. Canvas resource mutation/rebuild/readback/
destruction also owns the native window mutex. Direct sources gathered before a
resize only submit their current live texture; stale or destroyed sources yield.
The canvas adapter initializes its own resource updater so the inherited XYZ
updater cannot resize textures outside that ownership boundary.

Occluded and timed-out surfaces yield immediately, without polling or main-thread
calls while holding render locks. Outdated/lost surface configuration gets one
reconfiguration and one retry on a healthy device. This repairs a recoverable
surface state, distinct from physical device recreation. A missing render pass is
accepted as no presentation for that attempt; exposure/input/resize requests the
next frame. The acquired surface texture and its view are retained through submission
and presentation, then both references are released. Synthetic surface results test bounded retry/yield,
and actual race-instrumented resize/minimize/restore verifies the integration.

## Cached resize replay and borrowed texture lifetime

A GPU drawer can replay a previous composition only when all texture bindings
are owned by that drawer. Direct-render textures passed through UseTexture are
borrowed from their canvas owner; resize, explicit Snapshot readback and canvas
retirement may replace or release those views between complete draw batches.
Native window locking serializes each batch with mutation but cannot extend a
view's lifetime across batches.

The drawer exposes optional CanRedraw admission and rejects borrowed batches
before entering the GPU replay path. ComposerDrawer propagates that admission.
The native resize shortcut falls through to fresh gather and composition under
the renderContext when replay is refused, even if the model is unchanged. This
uses current direct-source lifetime checks rather than reusing cached bindings;
a later raster-only batch can use the cached resize shortcut again. This policy
may do more composition work during canvas resize; it does not add a ticker.

Snapshot's caller holds the Application UI owner through Do/PostWait; this is
application model ownership, not a requirement to run on the source window's
native event thread. Snapshot uses the GPU allocation and viewport last published
by the completed canvas Render, and applies the current accepted Model center/zoom
through Sync. It must not read raw widget geometry: another native window's event
thread may drain the application mailbox while the source window lays out. The
native GPU mutex protects textures during readback, not that separate layout.

Snapshot does not flush an outstanding native resize. A caller needing a newly
requested size waits for Render to update Model.View; then image dimensions match
that completed owner viewport. An immediate owner-held center/zoom edit is
reflected in Snapshot without waiting for another native Render callback.

## Empty sprite composition

Fresh composition omits the sprite source when every tier has no nonnil active
sprite with a Draw callback. The predicate and modification reset run under the
existing sprite lock before allocating, clearing or rasterizing sprite resources.
Replacing the composition source/operation list restores underlying scenes when
the last drawable sprite is removed or inactivated.

Active Draw callbacks still run on every composition in First/Normal/Final order,
including an unmodified layer and an active cursor's blink-off callback. Active
event-only sprites retain existing input and delayed-paint eligibility; HasActive
skips nil entries without requiring a Draw callback. This guard introduces no
new wake, deadline or event policy.

Previously allocated sprite resources remain retained for reuse. Empty resize
does not replace them; reactivation uses the existing size-correct allocation
path. Early cached resize replay and borrowed-texture admission are unchanged.
Software copy observations prove composition requests and pixels, not physical
GPU upload counts, immediate texture reclamation or display latency. Partial
uploads and process-wide wakeup reductions require separate evidence.

## Owned raster scanner row ranges

Native raster renderers and the private glyph-mask cache use the pinned scanner's
explicit row-tracking constructor. Its accumulation buffers belong exclusively
to that renderer; callers never directly supply Cell or CellIndex. General
NewScanner retains the original complete-row Draw/Clear behavior, including
externally supplied cell lists.

Tracked scanners retain the bounding half-open interval of in-range rows where
scanner methods allocated cells, independently of the current clip. Draw first
flushes the pending final cell, then intersects traversal with that interval.
Repeated drawing and subsequent clip changes keep the accumulated state. Clear
resets every row in the interval, including clipped-out rows, and all pending
accumulation. SetBounds initializes all current row heads even after capacity
reuse, shrink/regrowth or allocation, before resuming local clears.

Winding, coverage arithmetic, span order, transforms, glyphs, clipping and
composition are unchanged. This bounds per-path row-index work without retaining
another image, changing widget invalidation, or making a partial GPU upload.
Differential ordered-span and rendered-pixel checks establish software parity;
small-path microbenchmarks do not establish a native frame-latency improvement.

## Adjacent uniform rectangle replay

The raster backend may omit a Draw/Src item immediately covered by a Draw/Over
item. Both must have identical nonempty rectangles and source positions, nil
masks, and exact image.Uniform sources with built-in value color types; the
second color must have full alpha. Eligibility is evaluated at replay so a
borrowed uniform changed after recording is respected. Captured items and
painter/gradient recording remain unchanged. Other image implementations and
operation boundaries use the original replay path. Dirty-buffer differential
tests and actual StandardBox replay establish pixel and observable-call parity,
including clipping, fractional scales and border/shadow removal. This is CPU
rectangle write elision and does not establish GPU upload or native memory gains.

Qualification and ordinary latency results are recorded in
[the covered-fill report](../docs/reports/macos-covered-fills/README.md).

## Ordered scroll surfaces

An optional painter-group hint retains the full original command stream. Raster
reuse occurs at that group's original position, with an opaque rectangular pixel
cache captured before later siblings. Disjoint exposed strips replay the current
paint under its existing clip and order. The main source image stays current for
software snapshots and receives the ordinary complete GPU copy/upload. CPU
memmove does not admit strip-only uploads to an unchanged GPU texture.

Per-renderer retention is limited to 32 MiB of owned RGBA pixels, 8 MiB of detached
paint descriptors and 64 groups. The descriptor account includes operation slots,
fixed path data, mask headers and tightly copied alpha pixels. Map/renderer and
Go allocator metadata are separate; this is not a process/native/GPU memory
limit. A replacement plan temporarily coexists with retained plans and operation
builder/trim allocations. It retains no widgets, domain models, shaped runs, font
faces, strings, render commands or borrowed masks. Resize, bounded LRU eviction
and clock exhaustion retire all associated pixels and descriptors.

Opted-in managed Grid panes emit checked hints after actual layout/painting, preserving
frozen-header, scrolling-header, frozen-body and scrolling-body order. The raster
owner captures actual text output through the existing renderer: exact cached
mask bytes or post-scale/post-position fixed outline control points, effective
clips, built-in value colors and selection/background geometry. Capture preserves
SVG/bitmap/outline precedence, refuses SVG/bitmap/stroked/decorated/vertical/math
text and stops at its byte/primitive limit. Unsupported operations, uncertain
opaque coverage, source clipping and large jumps use ordinary complete replay.
A disabled fill-and-stroke standard-box path produces no pixels and is ignored.

For each overlap, every affecting operation in the old plan translated by the
checked integer delta must equal the current ordered operation stream clipped to
that overlap. Whole intersecting masks/outlines may be compared conservatively;
uncertainty repaints. The accepted current plan drives both complete and exposed
strip replay, avoiding a second font/color read between checking and drawing.
Model revision, selection/style subsets, font pointer identities and hashes do
not establish equality. Fractional physical phase/clip differences trigger full
repair even when rounded logical scroll offsets suggest an integer translation.
Other callers can explicitly use the original scalar epoch promise without this
checked paint protocol; the native Grid always requests VerifyPaint.

Filled quadratic/cubic curves flatten in coordinates relative to their integer
control hull and restore the origin and terminal endpoint exactly. This removes
absolute-position floating-weight jitter under integer translation. Some previous
antialias edge pixels intentionally change; logical paths, shaping and interaction
geometry do not. Exact emitted-edge and new full-replay pixel comparisons test
the new rule. The unchanged 4/6 ms frame targets and full Mac acceptance remain open.

The owned pane buffer uses toroidal storage. Checked scroll deltas change a
bounded logical-to-storage offset; overlap restore splits at most once in each
axis. Only disjoint exposed strips are written back before later siblings paint.
Full repair resets the mapping and captures the complete pane. CapturedPixels
counts owned CPU cache writes, not GPU uploads. This avoids a second full-pane
copy after every reuse. Capture shares identical copied glyph masks within one
plan, checking current bytes again before sharing. A changed borrowed mask gets
a new owned copy; the temporary lookup is cleared before retaining the plan.
Detached replay preserves covered-clear elision: an owned uniform Src immediately
covered by an equal-rectangle opaque uniform Over is replaced by that Over.
Admission has already checked each original uniform's finite source coverage.

Managed main-window Grid reuse is opt-in through Driver.ExperimentalGridScrollReuse
or ManualConfig.ExperimentalGridScrollReuse. Both default to false and are copied
at adapter startup; additional modal scenes retain ordinary replay. Current
matched stored-table experiments regress latency even though pixels agree, so
normal managed rendering emits no cache hints and retains no pane cache. The
experimental mode remains available for further renderer work; this is not a
recommended performance setting or completed latency acceptance.

## Ordered GPU scene candidate

`Driver.GPUSceneRendering` and `ManualConfig.GPUSceneRendering` opt main scenes
into a shared-device scene adapter. The candidate remains disabled by default
pending representative native eligibility, latency and memory qualification.
The software drawer exercises the same routing with CPU fallback and creates no
GPU device. Popup/modal/sprite sources retain their existing independent policy.

The adapter executes at the scene source's existing composition ordinal, before
later siblings, scrims, popups and sprites. Ordinary and measured source drawing
share the same adapter dispatch; renderers are no longer assumed to be rasterx.
Size checks use Renderer.Size rather than exposing an image. Scene destruction
optionally retires its renderer after stopping owned timers.

A supplied paint batch may update only a dirty subtree. The compiler preflights
all commands, built-in source/colors, destination aliases, aggregate path/text/
selection work and finite transformed coordinates before capture. It uses a
private scanner/Dasher and the reference text placement/mask cache, producing
owned ordered uniform/image/mask and exact 16-bit coverage operations. Unknown
callbacks, self-blits (including destination subimages), unsafe encoded fonts,
CFF/variable outlines, and unsupported operations refuse the complete batch.
Static TrueType composite traversal/point growth is bounded before allocation;
the normal font API retains its existing behavior. Pre-rasterized RGBA icons and
supported ordinary outline text remain accepted. A failed capture changes no CPU
pixels, scanner clip or winding, and invokes no unknown source/color hooks.

Packets have zero-origin targets, at most 8192 pixels per axis and 16 million
pixels, 65,536 operations, 32 MiB of detached retained payload, and two million
ordered tile references. Encoding preflights exact aligned sizes and allocates
exact output capacities. Coverage keeps the scanner's full 16-bit alpha and span
order. A first accepted full-target Src or opaque uniform operation establishes
an independent reset; command count alone never establishes a complete frame.

Span construction allocates row headers lazily for a geometric staging band,
bounded by the target height. The band can move/grow for arbitrary ascending,
descending or gapped arrival while preserving order within each row. Header
capacity and each accepted span are charged before allocation. Empty/clipped
streams allocate no row headers. Sealing refunds spare headers and emits the
unchanged exact relative-row format, without exposing staging growth capacity.
The shader, encoded packet arithmetic and retained payload limits remain fixed.

Operation construction uses private blocks of 64 records. Accepted operation
count and logical payload bytes are checked before a block is allocated. At
most 1024 blocks can exist per builder; only the last block has unused slots,
at most 63. Seal copies painter order once into an exact-capacity operation
slice and retires all construction blocks, borrowed lookups and stream state
on success or refusal. First-operation reset detection uses accepted count.
An already refused stream allocates no compact rows. Operation admission after
span-header refund preserves the preceding retained payload boundary.

Tile encoding reuses the count scratch array as write positions only after each
original count is recorded in the output row table. Encoded results own their
buffers, and the operation/row/reference/data byte format remains unchanged.
Block metadata, spare slots, construction/flattening overlap and encoder scratch
are separate from `Packet.Bytes`; the packet limit is not a process-memory cap.

Compute composites packed premultiplied RGBA bytes with the reference Go/scan
u32 blend and truncation rules, then copies the storage buffer to a sampled
RGBA8UnormSrgb texture. The final source uses the existing drawer composition
pipeline; no floating-point blend or separate logical device substitutes for the
reference byte math. An empty batch reuses only a current live source. Checked
allocation/uploads/copy/finish errors and device loss take CPU fallback. The
pinned API's BeginComputePass, dispatch and Submit are not error-returning or
completion measurements; successful submission is not proof of presentation.

A distinct CPU base plus at most 256 sealed packets/32 MiB of packet payload
reconstructs accepted partial work without GPU readback. An independent reset
retires old history; journal limits materialize it into the CPU base. Unsupported
paint reconstructs prior work before the normal CPU render. A recovered partial
GPU batch seeds that current base once. GPU readiness and CPU materialization are
separate. Explicit Image returns the usual mutable image: after exposure, accepted
packets update the retained pointer immediately, later partial GPU packets seed
possible caller mutations, and empty batches use a current CPU copy. This
conservative compatibility mode lasts until resize; it gives up deferred replay
and full-image-upload elision. Concurrent unsynchronized image mutation is no
more supported than with the original raster renderer.

Drawer-owned leases serialize GPU preparation and checked source append on the
originating device. Capacity is 64 sources, 128 MiB of source-requested GPU
allocations each and 256 MiB aggregate. Retained plus transient/in-flight replaced
input buffers remain charged until completion flags are observed; growth uses
charged geometric capacity and at most 16 pending retired buffers. Retirement
drains queue references before returning credit. These limits exclude legacy
drawer/surface allocations, pipeline/driver metadata, CPU base/packet scratch and
opaque GPU-driver caching; they are not a total native/process memory limit.
Callbacks add no source-owned UI/model/font references and never enter
Scene/application ownership; borrowed device health remains owner-managed. Copied, stale, reentrant or over-budget leases/contexts refuse.
Close during a pending composition or another resource scope requests retirement;
registry ownership and credit remain until a safe End/scope drain or drawer
teardown. A request arriving after selection waits for a later drain. Cleanup is
idempotent and preserves callback unwind behavior. Borrowed batches still refuse
cached Redraw. A later owned Go-image batch can replay normally; recycling a
borrowed texture slot must allocate independent owned storage before uploading.

GPU source diagnostics expose outer host Draw time and explicit unavailable
CPU Raster/Copy breakdowns. CPU fallback timings measure the current ordinary
render/copy calls; earlier journal reconstruction is also in outer Draw. No
physical upload bytes, GPU completion, presentation, whole-scene eligibility or
latency acceptance follows from these routing counts or bounded native readbacks.

## Renderer-owned static outline reuse

The Rasterx renderer and its private scene/scroll captures may reuse immutable
font-unit outlines keyed by Font identity and full GID. Entries are admitted only
from successful bounded static TrueType retrieval; variable/CFF and encoded
bitmap/SVG resources retain existing routes. Ordinary public font outputs remain
independently mutable. CPU admission refusal falls back to ordinary retrieval;
cache capacity alone does not reject a successful GPU outline.

Every cached bounded hit consumes the opaque exact successful retrieval cost in
the caller's aggregate outline budget, including composite point/traversal work.
Later per-glyph, segment, finite-coordinate and extent-product checks still run.
Cache warmth must not change successful budget admission or revive a failed batch.

Each renderer retains at most 2MiB segment capacity, 1024 entries, 8 Font references,
and 64KiB segment capacity per entry. These bounds exclude metadata, transient
retrieval, Font/backing storage and existing glyph-mask caches. Resize retains
size-independent geometry; replacement Font identity gets a distinct entry.
CPU/GPU Release drops new outline references while preserving image snapshots.
Faces remain generally unsafe for concurrent mutation; Font/backing resource
immutability and the renderer's existing owner serialization remain required.
Matched original geometry/pixels, budget/refusal and native measurements must
qualify this optimization. It does not itself establish performance or memory
acceptance or enable the GPU scene path by default.
