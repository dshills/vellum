# Cursor ownership source checkpoint

The desktop cursor adapter now binds to its requested window and owns native
handles through eviction, invalidation, replacement, close and terminal drain.
Shared raster retention is limited to 64 entries / 8 MiB of pixels, and each
window admits at most 16 native handles / 1 MiB of pixel payload including active
and pending replacements. These are declared retention bounds, not measured
Cocoa allocation. See [the contract](../../../specs/cursor-ownership.md).

The final source checks passed fresh preparation of both pinned dependencies,
16 new regression methods, the existing 390-case original-artwork raster matrix,
selected desktop/base regressions under Go's race detector, color-package race
tests and Go vet. The native adapter tests use a fake backend and never initialize
GLFW or create a native window. Exact receipts and source pins are referenced in
[evidence.json](evidence.json). SDK/native packaging receipts remain separate.

Earlier attempts are retained in the external checkpoint directory: an unquoted
space in the external modfile path, renderer interface compilation errors and a
standalone SVG shaper initialization failure. They are superseded only by the
final passing run; no failed run is relabeled. Internal review added null-wrapper,
queue-publication, main-with-window-lock and overlapping-refusal controls, and
fixed fatal-path theme restoration before the final run.

Official theme setters and cursor rendering now share a read/write transaction;
raw exported global theme/map access still requires caller synchronization.
The known shape-only renderer avoids application factories and system-font scans.
The legacy visibility methods retain status behavior; native Hide/Show and raw
mode interoperation need their own qualification.

No physical-pointer, Retina appearance, OS allocation/thread instrumentation,
idle/endurance or measured display result is claimed here. Native presentation,
M2 methods, M4 lifecycle/endurance, M5 real workflows, M6 distribution/SDK approval
and M7 final audit remain open. Measured 60 Hz responsiveness gates Mac release;
4 ms remains an optimization target. This checkpoint does not complete Mac.

The first Prism run returned two findings discarded by path filtering because
they named prepared Go paths inside the staged patch. Both were inspected. Its
early-worker window-creation concern exposed the legacy pre-queue inline fallback;
window creation now shares explicit native-thread/queue admission with cursors.
Two no-screen controls cover worker startup/stop and direct-main validation
without calling GLFW. Its SetSize concern matches inherited deferred application;
the public method and contract now explicitly require Set to apply a changed size.
Final proof and review supersede the preserved first checkpoint without relabeling it.
