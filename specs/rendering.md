# Rendering and resource failure contract

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
