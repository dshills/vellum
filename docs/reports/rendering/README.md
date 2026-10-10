# Rendering hardening acceptance

The race-instrumented macOS native rendering smoke completed twice, including
normal process shutdown. Both checks reported zero paint wakeups across a one
second quiet interval, an owner update after idle, live resource rebuilding,
resize and minimize/restore. The checks do not establish a sustained CPU budget,
frame/input p99, physical GPU-reset recovery or cross-display DPI acceptance.

The live checks exposed and drove fixes for asynchronous raster composition,
canvas resource ownership and acquired surface-texture lifetime. Patch 0017 is
owned by this driver, with input/output hashes and synthetic regressions for
composition ownership, stale direct sources, bounded surface retries and shared
device-loss subscriptions. Loss notification fails the application with an
explicit ErrDeviceLost; automatic physical-device recreation remains open.

No input sources, VoiceOver or system preferences were changed.

## Platform integration: cached borrowed-view replay

The two-window GPU platform smoke exposed a native abort in cached resize redraw:
a bind group used a view released after Snapshot/resize/canvas retirement.
Patch 0019 adds a separate lifetime admission boundary without modifying the
committed rendering or accessibility patches. Drawer batches containing borrowed
textures cannot replay; the window requests fresh owner-held composition instead,
including when no model changed. Raster-only cached redraw remains available.

The deterministic regression failed before the fix by proving that a retired
borrowed batch entered the GPU path. After the fix, race tests cover rejection
before backend access, retained batch admission after End, independent raster
batch recovery, composer propagation/legacy compatibility and presentation of a
fresh replacement source. Logs: [before](redraw-regression-before.txt) and
[after](redraw-regression-after.txt). This unit evidence does not itself establish
native presentation or resize acceptance; the final two-window GPU smoke records
that integration separately.

The expanded live integration also exposed a cross-window layout/readback race:
Snapshot read widget ContentBBox while its source window laid out. Snapshot now
uses the viewport captured by the owner during canvas Render and applies it
through Sync, avoiding foreign reads of layout geometry. The native GPU mutex
continues to protect readback textures; it does not claim ownership of layout.

The final race-built macOS two-window platform smoke passed twice after both
fixes, with six snapshot/resize/readback cycles per run, survivor readback after
first-window teardown and normal final shutdown. Logs: [run 1](redraw-native-smoke-1.txt)
and [run 2](redraw-native-smoke-2.txt). Both records retain their qualification gaps;
these checks do not certify full OS support or native presentation timing. The
preceding run's [Snapshot layout race](redraw-snapshot-race-before.txt) remains
failed evidence. [Fresh adapter/transport/command tests](redraw-platform-tests.txt)
and standard driver/platform vet pass. Upstream core's standard vet still reports
[existing unkeyed literal and unreachable benchmark diagnostics](redraw-upstream-vet-existing.txt);
only its composites/unreachable checks were disabled for the additional core vet
pass. GPUDraw/composer standard vet passed without exclusions.

The Snapshot contract is checked explicitly by the platform fixture: every
readback must have image dimensions equal to the completed owner viewport, and
center/zoom are edited immediately before readback in one owner callback. The
new probe position must be background in the baseline image and contain the
rectangle after Snapshot, proving current camera application through Sync.
Synthetic probe tests reject stale dimensions and stale camera projections.
Window creation readiness uses the actual canvas widget Show handoff before
owner-posted GPU readback work. Shutdown waits on a captured native lifetime's
atomic IsClosed publication rather than polling mutable names.

The final two native records above include the race-build metadata, exact runtime
source hashes and exit status 0 for the dimensions/current-camera assertions.
No native layout callback can apply the edited camera between the model mutation
and Snapshot because both happen in the same Application-owner callback.

The final current-source native records include the diagnostic error-joining
follow-up and bounded two-axis resize acceptance. Requested native pixel geometry
must be observed and every resized canvas axis must have a completed allocation
change. Ignored/clamped requests and already-requested sizes fail qualification
instead of counting as transitions. All emitted source hashes match the current
files and platform source-hashes.json; the [follow-up record](../platform-support/resize-followup.txt)
contains test/build and consistency-check results.

The final native records also wait for actual canvas Show publication in the first
window before its initial readback, matching the second window's barrier. Both
current-source race runs pass with this startup topology ordering in place.
