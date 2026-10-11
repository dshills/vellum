# Owned native dependency inputs

`dependencies.json` pins upstream module versions, Go module checksums, ordered
patch stages, and individual patch SHA-256 digests. `prepare.py` verifies each
stage's input-file hashes before applying that stage. Upstream module source and
its original licenses are retained intact in ignored `../build/` copies.

These patch bytes are copied without modification from the research foundation:
Cogent stages `nativeinput`, `foundationchecks`, `foundationcontracts`, then
`foundationintegration`; go-text stages deferred sbix then deferred static outlines.
The manifest records their original locations for provenance. Historical prototype
sources and evidence stay in the repository, but preparation never reads them.
No AccessKit prototype C/Rust bridge is included in the framework driver.

The copied patch set carries forward experimental fixes; a published fork/release,
upstream disposition and full third-party distribution/license audit remain gates
before distributing framework binaries or a public SDK release.

Patch 0016 fixes native multiwindow ownership races discovered by live race
instrumented handoff checks: composer publication/withdrawal and lock rechecking,
queued-close loop termination after teardown, protected fast resize readback, and
scene title routing to each scene's own window. Its stage verifies original hashes
after patches 0001–0015 and expected final source hashes, including regression tests.

Patch 0016 also gives new windows independent sizing contexts, snapshots shared
screen metadata, and synchronizes startup main-queue publication. The manifest
and regression list identify the complete cumulative safety change.

Patch 0017 adds shared device-loss notification, bounds recoverable surface
reconfiguration, and holds composition/resource ownership through GPU submission.
Surface textures remain retained through presentation. Its input/output hashes
and backend regression tests make these adapter-owned fixes reproducible.

Patch 0020 delivers window shortcut first/final phases even when a Scene has no
focused widget. Focused widgets retain their existing precedence. The owned
regression verifies document fallback, consumed window commands and unregistered
chords, with stage input/output hashes for independent SDK preparation.

Patch 0021 makes main-loop stop a closed, idempotent signal that also works
before queue publication. It releases startup/unaccepted delivery waiters while
preserving synchronous output for accepted callbacks. Both loops acknowledge
refused deliveries; native window creation is one admitted callback and returns
an error on refusal. Six owned race regressions cover these transitions.

Patch 0022 acknowledges admitted callback failure before replaying the original
panic, stops further admissions and avoids a crash-dialog wait on a terminal main
loop. A process-wide desktop barrier serializes native empty posts and GLFW
termination; all five input-event posts use it. Its race tests cover queued/inline
panic, nil panic, incomplete Goexit, pending callers and event lifecycle ordering.
The barrier does not guard every GLFW operation or standalone GPU helper lifetimes.
See [dispatch contract](../../../specs/native-dispatch.md) for supported outcomes.


Patch 0026 supplies `Scene.EventBoundary` after normal dispatch of each accepted
event, including consumed widget events. The managed editor publishes legacy
text/selection changes at that originating boundary before a following Undo or
native input action. Already-handled events skipped on entry have no boundary.
Callbacks run under the originating render-context lock and must not retain the
event or reenter dispatch. Input/output hashes and consumed/unhandled event
regressions make the seam reproducible in independently prepared SDKs.

Patch 0027 preserves deferred FIFO nodes when an ordinary event of the same type
arrives. Ordinary `Send` compression must not replace a publication wake or a
reserved input action; deferred merging remains governed by `Ready`/`Merge` at
dequeue. Normal event compression and priority window events retain their
existing behavior. Backend and managed-adapter regressions verify startup,
subsequent publication, reservation commit and cancellation, with pinned
input/output hashes for independent SDK preparation.

Patch 0028 preserves an active native TextField composition across Maker refresh.
Preedit deliberately has not marked the controlled value edited; refresh must
also check the composition lifetime before replacing the private edit buffer.
Repeated Update, same-visible commit, one native undo entry and cancellation
regressions verify the seam with pinned input/output hashes.

Patch 0029 exposes a managed TextField focus-loss deferral seam. The adapter can
retain the controlled Change boundary while already admitted native records are
still queued, quiesce the outgoing native client, and finish exactly once. Normal
fields retain immediate focus-loss behavior. Refocusing before finalization
preserves the fresh directional selection and resumes the caret; other focus
listeners still receive the event. The core regressions cover default, declined,
deferred, refocused and read-only behavior. This seam preserves accepted native
text through a focus handoff; it does not redefine all programmatic focus and
legacy key routing as one FIFO. Input/output hashes make the change reproducible.

Patch 0033 gives every main scene its own named text-cursor ownership map,
protected by that stage's sprite lock. Desktop window closure can run
concurrently; different stage locks cannot protect the old package-global cursor
owner pointers. `TextCursor` retains its legacy pointer argument for source
compatibility but ignores it. Independent cursor names and manual scene clocks
retain their behavior. Pinned input/output hashes and concurrent-scene regression
tests accompany the fix; the loaded-data Cocoa report records the triggering
race and the subsequent acceptance scope.

Patch 0034 adds optional tagged native work recording. Protected deferred scroll
nodes unwrap before real event dispatch. A ticket includes intermediate native
update/composition work and the final armed owner generation; cancellation and
abnormal return retain exact input/owner/frame terminal ownership. Surface
acquisition and presentation calls are separate from configuration, view creation
and cleanup costs. Default composition order remains unchanged. See
[the patch contract](patches/frametrace/README.md) and
[the measurement specification](../../../specs/grid-performance.md).

Patch 0035 omits the scene-sized sprite source before allocation/clear when no
active sprite has a Draw callback. Fresh composition replaces the previous
source/operation list. Active callbacks retain their tier order and run on every
pass, including blink-off; event-only sprites retain their Active-only deadline
criterion. HasActive skips nil entries. Existing resources remain retained and
reactivation follows the existing size-correct path. See
[the patch contract](patches/emptysprites/README.md) and
[the rendering specification](../../../specs/rendering.md).

Patch 0036 adds explicit per-source diagnostic composition. Ordinary measured
composition retains its path. Detailed tickets retain at most 32 scalar source
records across intermediate/final compositions; nested raster/copy/image-API
intervals never add to UI-work totals. GPU counters report cache hits and API
attempts with checked logical RGBA input bytes, while upload success/failure
remains unavailable. See [the patch contract](patches/sourcecost/README.md) and
[the measurement specification](../../../specs/grid-performance.md).

Stage 37 enables bounding-row scans for exclusively owned native raster and
private glyph-mask scanners. General scanner construction retains full-row and
direct-cell behavior. Exact span/lifecycle differentials and independent
baseline oracles accompany the pinned patch. See
[row tracking](patches/rowtracking/README.md).
