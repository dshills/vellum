# Advanced GPU view

`drivers/cogent.GPUView` is an experimental public native extension for custom
shaders and 3D rendering. It borrows the window's existing WebGPU device and
queue, renders into a retained offscreen target, and lets the normal compositor
combine that target with standard UI. It never creates its own device or presents
an independent surface. The bounded rectangle `GPUCanvas` and portable
`MeshCanvas` mesh/image API remain supported.

## Ownership and callbacks

Fields and methods belong to `Owner.Do`/posted owner callbacks. `Owner`, `Window`,
`Program` and `Depth` are fixed while mounted. Each `*GPUProgram` mounts in only
one live widget; create a separate instance for each independent window. Callback
functions are captured at mount; replacing public function fields affects only a
later mount. Retirement is one-shot, including stale destruction after remount.

Advance `Revision` whenever program/model state changes. The framework compares
revision, full content pixel size, content-local visible clip, current dp scale
and background. An unchanged owner redraw reuses the completed texture without
calling preparation/drawing or reading GPU pixels.

`Prepare(GPUFrame)` receives a borrowed encoder before the main render pass. It
may encode ordered preparation/offscreen/compute work. `Draw(GPUFrame, pass)`
receives the framework-owned main pass and a nil encoder. Neither callback may
nest `Do`/`Drain`, invoke adapter/window/UI methods, start concurrent GPU work,
end/release/finish transient handles, independently submit/present, or destroy
borrowed device/queue objects. `Input` runs on the owner after lifetime guards;
it may update model state and invalidate. It returns true only for consumed
events; unhandled keys and scrolling propagate to normal UI handling. Mouse-down
focus is consumed when the mounted view actually takes focus. `OnError` runs on the owner after
native and resource locks release and may query `Stats`.

`Release` frees program-owned pipelines, layouts, bind groups and texture views;
it must tolerate partial initialization, including a failed first preparation.
Framework helper buffers/textures are released afterwards. Framework resource
cleanup runs even if `Release` panics; this does not certify application-wide
panic recovery. A program's raw allocations and borrowed-handle discipline are
an advanced caller contract, not an isolation boundary against arbitrary code.

## Versioned context and targets

`GPUContextVersion == 1`. `GPUFrame` exposes the window's device/queue, logical
limits, target metadata, callback-scoped helper and preparation encoder. `Size`
is the complete physical content allocation; `Clip` is its visible content-local
pixel rectangle. `DpScale` comes from the widget's current unit context. Output
is linear premultiplied RGBA stored in a four-sample RGBA8 sRGB target. An opaque
background or transparent black is supported. Optional depth uses a checked,
matching-size four-sample Depth32Float attachment with RenderAttachment usage.

`ResourceGeneration` changes on explicit rebuild, callback failure or a changed
device. `TargetGeneration` also changes on size replacement. Resize retains
program resources while recreating resolve/MSAA/depth/readback targets. A program
must recreate its own generation-dependent objects when its resource generation
changes. Metadata does not promise a generation value is nonzero.

## Resource helpers and accounting

`GPUResources.Buffer` and `Texture` maintain named framework-owned resources.
Same-name identical descriptors reuse handles; descriptor replacement checks the
transient old-plus-new byte budget, allocates successfully, then retires the old
handle. Helpers reject invalid names, buffer sizes, mapped-at-creation buffers,
unsupported texture descriptors and device-limit violations before allocation.
Caller-owned bind groups/views referencing a replaced/dropped handle must first
be retired. `Drop` explicitly retires named helper resources. Helper handles may
be held by a program for that resource generation and must not be released by it.

Texture helpers support one mip, one layer, 2D RGBA8/BGRA8 unorm/sRGB,
RGBA16Float and Depth32Float, with one or four samples. Unsupported compressed,
array/mip resources require raw APIs. Four-sample textures require render usage
and reject copy/storage usage. Byte arithmetic checks overflow. `WriteBuffer`
checks CopyDst usage, four-byte alignment and subtraction-based bounds, and
counts successful helper uploads. Each Prepare and Draw gets a distinct helper
lease; a saved Prepare helper is rejected even during Draw in the same frame.

`MaxGPUBytes` defaults to 512 MiB. The estimate covers resolve + MSAA (20 bytes
per pixel), optional depth (16 more), retained synchronous readback staging, active asynchronous staging and helper
resource bytes. It excludes pipeline/view/bind-group/driver overhead, raw allocations,
foreign writes and process-wide memory. The estimate counts currently owned
helper handles; retired objects retained by in-flight queue work or native
references are excluded. It is not a hard physical-memory or driver high-water
limit. Admission checks the owned old-plus-new replacement peak; physical resource
retirement may complete later. Statistics count helper uploads and
callback attempts and submitted frames; they are not GPU timings or all foreign calls.
Budget refusal before mutation preserves the previous completed frame. A callback
or checked encoding failure withdraws readiness, releases partial program/helper
resources and requires preparation again; failed work never advances the cache.

`Snapshot` synchronously copies the last completed target and returns straight
NRGBA, with linear-space unpremultiplication. It does not flush pending updates.
Readback has device/budget admission and is absent from normal composition.

`SnapshotAsync(ctx, callback)` admits one capture per `GPUView` through callback
return or canceled worker cleanup. Admission, like all view methods, belongs to
`Owner.Do`. Nil arguments and pre-canceled contexts fail immediately. A busy
view returns `ErrSnapshotBusy`; remount does not evade that reservation.
`MaxSnapshotBytes` defaults to 64 MiB, bounding padded staging and one tightly
packed CPU image separately. Delivered images retained by the application are
outside that bound. Staging is included in every `MaxGPUBytes` admission path,
including synchronous readback and target/helper replacement.

The worker calls nonblocking `Poll(false)` under the originating native window
teardown lock and resource mutex, independently of paint, visibility and widget
geometry. Successful mapping copies padded rows directly into one owned image,
then unmaps/releases staging before linear alpha decoding on the worker. No
mapped native memory reaches decoding or delivery. The bounded row copy retains
those locks so cancellation/teardown cannot unmap its borrowed range. It can delay
paint/resize while copying; `MaxSnapshotBytes` permits a lower application limit.
This checkpoint establishes byte bounds, not a copy-latency guarantee.
A spontaneous map callback
only notifies a bounded, never-closed channel. Active requests alone poll; there
is no idle repaint or polling loop after completion.

Delivery waits for bounded owner mailbox credit without holding GPU/native locks.
The callback receives `GPUSnapshot` with detached `Image`, capture-time `Frame`,
`Revision`, `ResourceGeneration` and `TargetGeneration`. Pending changes are not
flushed. Normal resize/revision changes preserve the captured old frame; rebuild,
render failure, caller/window cancellation, retirement, native closure or a changed
mount/resource generation quietly drop it. Canceled queued closures release their
image even if the owner never drains again. Cancellation retains the reservation
until a decoder or running callback relinquishes image ownership. A callback may
query `Stats`, but must not nest `Do`/`Drain`; recursive capture remains busy until
that callback returns. Error callbacks count as completed deliveries. Async stats
report pending reservation, polls, admissions, deliveries, cancellations and owned
GPU staging bytes. They are diagnostics, not timings.

A failed native-window lock forbids access to the released window device. The
worker can still abort/unmap/release its independently owned staging buffer, which
retains its private native device reference. This fallback does not depend on
later widget destruction following forced native closure.

`RebuildResources` explicitly retires resources and requests another frame.
`RequestFrameAfter` requests one paint; it does not create a permanent ticker.
Closed, retired or remounted adapters reject stale input, sources and resource
queries. A physical device loss follows the existing fail-stop driver policy.

## Evidence and remaining qualification

The independent [3D consumer](../examples/viewport3d/README.md) supplies a real
24-vertex/12-triangle cube, camera matrix, checked raw shader/pipeline, UI buttons,
drag/scroll/arrows and two independent windows. Native depth acceptance submits
a near red face before an overlapping far blue face and verifies red center
pixels. It checks retained chrome redraws, camera-only uploads, controlled
preparation/drawing fault repair, budgets, rebuild, three resizes, empty-frame
clearing, expired helpers and closure. Offscreen snapshots do not prove presented
clipping, MSAA edge quality, input/presentation latency or sustained performance.

Reproducible backend patch25 synchronizes render-window registry snapshots and
existing-member focus promotion, and publishes newly created windows after their
main stage is initialized. Native focus/geometry calls run after the registry mutex
is released. Detached snapshots protect collection membership; they do not extend
native lifetimes or synchronize all mutable scene state. Concurrent registry and
actual focus/close-hook race regressions accompany native teardown acceptance.

The fixture also exposed concurrent window removal during the backend's cached
quit-loop indexing. Patch 23 snapshots callbacks and window handles under the app
mutex, invokes callbacks unlocked, skips already removed windows, and checks the
final list under lock. A veto or window created during closing prevents a clean
quit. Deterministic concurrent-remove, veto/new-window and reentrant callback
checks accompany the real native shutdown acceptance.

The independent [imaging consumer](../examples/imaging/README.md) encodes ordered
compute, an RGBA16Float preparation render pass and straight-linear tone mapping
into the main sRGB pass. Quantitative HDR/alpha and asynchronous lifecycle evidence
is recorded separately from general compute/float feature support. Advanced GPU
timestamps, physical device loss, cross-display geometry and broader P4/Mac
performance qualification remain open. See [macOS completion](macos-completion.md).

## Isolated native smoke geometry

The independent example's `--smoke` mode disables backend window-geometry loading
and saving before creating the application. The process-local setting remains
active through native teardown and restores its prior value on normal return.
Interactive mode retains the normal persistence behavior. The smoke result emits
its actual `geometry_persistence_disabled_in_fixture` flag, so external validation
can reject a run that did not isolate fixture geometry. Geometry isolation does
not establish physical input, compositor presentation, native/GPU memory, sustained
performance or complete Mac acceptance.

The 3D smoke establishes a paired 100 ms quiet interval for both initial targets,
then captures both baselines together before its first chrome mutation. It refuses
later frame/resource drift with the original strict retention checks. This avoids
using a provisional first layout as the retention baseline; the quiet interval is
an operational fixture barrier, not a guarantee that all OS work has drained.
Bounded owner-thread observations retain independent-window state for diagnosis.
No frame, budget, depth, pixel, fault-repair, resize or closure assertion is waived.

The compute/HDR imaging example applies the same process-only geometry isolation
in `run(true)` before constructing either fixture window and retains it through
normal application/native teardown. Its `vellum.imaging.smoke` event reports the
actual isolation flag without changing the existing native compute, capture,
minimize, cancellation, resource-isolation or closure assertions. Interactive
mode retains normal persistence; broader GPU memory and Mac qualification remain
separate from this fixture policy.

Every post-launch smoke return closes the application and joins the smoke worker
before restoring the process-local geometry setting. Worker completion follows
its context cleanup and application-close callback; receiving its result alone
is not a completion barrier. The outer watchdog remains active during the join.

## Typed input extension

GPUView.InputEvent optionally captures a typed callback at mount and replaces
legacy Input routing for that mount. GPUEvent includes kind, button/modifiers,
full content-local pixel position/size, dp scale, visible-clip membership and both
scroll axes. Primary pointer gestures use same-window capture and swallow release;
Escape/focus/retirement cancel. Callbacks run only on a live application/window.
Secondary/middle events remain targeted and do not promise capture outside the
view. Legacy Input and all rendering/presentation behavior remain unchanged.
See the [workbench track](fusion-workbench.md) for scope and camera helpers.

Typed cancellation is posted to the originating owner to avoid nested ownership
when focus/popup transitions happen inside a command. GPUEvent.Sequence identifies
the primary gesture; viewport.Gesture ignores cancellation from an older sequence.
Widget/window retirement suppresses outstanding callbacks and revokes capture.
