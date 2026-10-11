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
