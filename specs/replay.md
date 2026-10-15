# Explicit managed UI replay

`replay` defines bounded value records and a synchronous runner for deterministic
application tests. The Cogent manual adapter routes pointer and keyboard input
through the real retained controls, focus handlers and modal hit testing. The
software renderer supplies completed frame images. Portable protocol tests and
actual widget acceptance are separate evidence.

## Ownership and execution

The caller supplies an initialized `execution.ManualClock` and
`execution.Manual` executor to `cogent.NewManual`. The managed adapter borrows
them. Its windows share that explicit clock and render in application registration
order. This checkpoint permits one active managed manual application per process,
with multiple windows. The caller keeps Cogent global palette, settings, keymap
and highlighting configuration fixed for the application lifetime. Offscreen
bootstrap reads its temporary data directory; the explicit Vellum theme overlays
pinned bootstrap defaults rather than resetting every Cogent global. The lower-level Core adapter's independent per-window
clock configuration is a separate capability.

Manual construction requires the `offscreen` build tag. Desktop construction
returns `ErrManualUnsupported`; it never switches a global driver at runtime.
There is no window event loop, hidden sleep, first frame, worker selection or
automatic advance. `ManualConfig.Executor` reports borrowed readiness; callers
must inject the same executor into task/document managers when creating them.
Construction does not replace executors in existing managers. Other caller
goroutines are outside this accounting. `Input` enqueues checked native input; `Step` bounds the sum
of originating events and owner callbacks. Frames are a subset of event dispatch.
`RequestFrame` only requests paint; `Frame` is the documented request-and-dispatch
convenience operation. The replay backend uses enqueue-only requests so dispatch
remains within the runner's step budget.

Each method has a single dispatch owner and refuses concurrent/reentrant use.
Callbacks may use Application APIs but may not reenter the manual adapter.
Snapshots are inspections, not dispatch. Closure retires domain window lifetimes,
controlled bindings, originating events and native windows; caller-owned clocks
and executors remain caller-owned. Abnormal callback cleanup follows the scoped
native dispatch contract; broad arbitrary destructor recovery is not promised.

## Fonts, identity and routing

Manual configuration supplies checked, detached font bytes with required SHA-256
identities. Each scene uses that shaper, explicit DPI and captured theme. Ambient
fonts, clipboard content, input preferences and OS services are not replay inputs.
Native clipboard/menu/chooser/drop services, context-menu shortcuts and custom
extensions refuse until separately qualified. Mapped screenshot, zoom, refresh,
clipboard, menu, find and window-switch shortcuts refuse before event admission;
these manager routes can write files, change renderer settings or leave the
qualified control topology.

A record manifest fixes its version, fixture/theme/font digests, layout/renderer
versions, UTC start time, seed, window incarnations, viewport and DPI. Operations
have contiguous sequence numbers and nondecreasing elapsed timestamps. Only
`Advance` moves time. Worker selection records an admitted worker ID. Backends
check current window/control incarnations and optional model revisions before
routing. Physical input enters the actual event source and focus/hit-testing
path; it never invokes a descriptor's callback directly. Explicit command input
is a separate route through the fixture's owner and command registry.

The example backend records physical coordinates. Optional target identities
guard those coordinates against stale controls; pointer-down/up coordinates must
fall inside the measured target, and target-specific keys require actual focus.
An unrendered target refuses until an explicit frame establishes current bounds.
Disabled controls and modal-hidden targets refuse. Raw-coordinate tests separately
prove that native routing blocks disabled/background actions.

## Bounds and observations

Record validation precedes backend calls and rejects invalid UTF-8, NULs,
non-finite/out-of-viewport coordinates, unknown or irrelevant variant fields,
invalid identities, timestamps, and operation/payload/encoded-byte limits.
Recorder retention detaches caller data; a copied or uninitialized recorder refuses.

The runner validates and detaches the whole record before execution. `MaxSteps`
charges one unit per operation plus actual dispatch units reported by `Step`.
Explicit clock/worker operations have their separate execution subsystem bounds;
the runner cannot preempt arbitrary callbacks. Partial failure reports the applied
prefix and last successfully inspected residual readiness, without rollback.

`MaxCaptures` bounds observation count. `MaxCaptureBytes` checks raw observation
JSON before compaction. `MaxCheckpointBytes` bounds the sum of compact JSON
encodings of retained checkpoints, including their labels and identities. Stored
JSON uses tight detached backing slices. This budget excludes Result's bounded
metadata, array separators and outer encoding wrapper.

Capture never requests a frame or advances time. It refuses queued/dirty or
unrendered output, returns detached semantic/model/layout data, and identifies
an already-completed detached software image. Native selection snapshots expose
UTF-16 range and directional focus. Actual images and measured geometry are
distinct from a model's declared semantics.

`RunUntilIdle` drains only currently ready work. Its `QuiescentNow` requires no
ready, running, awaiting, due-timer or blocked work. Future-only timers do not
prevent current quiescence and are reported with the next deadline. Unrendered
layout, selected workers, reserved results and blocked originating nodes remain
visible. Reaching the dispatch budget or reporting unexplained ready work without
progress fails; neither becomes a successful idle observation.

## Owned widget timing

Completion debounce, popup/snackbar expiry and cursor paint deadlines use the
originating scene clock and bounded event queue. Arming a timer never dispatches
it. Explicit clock advancement makes its event eligible; an explicit step handles
the widget work. Cancellation retains a queued reservation until dispatch or
withdrawal, and closure retires originating reservations.

Text buffers borrow a checked manual-execution lease for one live view. Shared
leases, conflicting views or ambient work already running refuse adoption.
Markup executes synchronously only while dispatching the originating manual
work; edit/event timestamps use that clock. Manual output readers and filesystem
autosave refuse before launching hidden workers. Desktop timing retains its
existing real-clock behavior and repeat-click update bookkeeping. Timed repeat
callbacks recheck the captured widget/scene under their render ownership.
Unsupported frame-request window extensions refuse before dispatch. Invalid
output buffers retain their explicit programmer-error failure; manual reader
admission still refuses before launching work.

## Acceptance and scope

The independent `examples/replay` consumer freezes one input record and replays
it at batches 1, 7 and 64. Its checkpoints compare exact model state, actual
focus/selection, measured geometry/semantics and software image identities.
Per-batch transport event/frame counts are reported separately and may differ.
The workload opens a fixture, edits a property through keyboard input, sorts and
selects rows, admits/cancels a worker, delivers its reserved result and undoes
the edit. Separate fixtures cover modal routing, disabled input, stale lifetimes,
timer scheduling, configuration refusal and teardown.

Software replay does not qualify Cocoa IME, external accessibility, VoiceOver,
GPU presentation, physical device loss, cross-display scale or performance.
The separate [local semantic endpoint](automation.md) reuses owner/domain actions;
it is not physical input replay. Broader external automation workflows,
large-grid acceptance, output-reader timers and
filesystem autosave remain separate work.

### Stored managed-grid replay

The driver offscreen regression `TestStoredGridRoutedReplayAcrossBatches` adds
two stored datasets: 10,000 and 1,000,000 rows, each retaining twenty `uint32`
fields. This is separate from the four-row independent replay consumer and its
bundled record. A pilot measures one sixteen-event physical input trace per
dataset; fresh instances replay that unchanged trace at dispatch batches 1, 7
and 64. Coordinates are measured once by the pilot, not resolved again during
replay. The source bytes and fixed sequence define this test; it does not yet
provide an independently archived, versioned large-workspace replay record.

The trace selects a cell, extends a two-row selection, scrolls both axes,
reverses an immutable source index, selects a visible cell at the new deep
offset, opens its inline editor with F2, types `42`, commits with Enter, and
uses actual Undo/Redo buttons. The source replacement reverses an already
ordered index in constant time; it is not a background sort/filter workload.
An application-owned journal restores the original cell bytes and reapplies
the edit with increasing cell versions. One inspector window shares the domain.

Ten completed checkpoints compare source revision, active cell, captured
selection, draft/editor state, viewport, measured cell/header geometry, actual
control semantics/focus and both software-image digests. Explicit assertions
check `42`, original bytes, cell versions, source revisions and history counts
independently of pilot equality. A row/column formula also verifies every
measured cell's key and value against the requested normal/reversed ordering;
the undo baseline is calculated from the record identity, not copied from the
pilot. Image identities include bounds and stride. Process-global native incarnation tokens and
transport frame/generation counters are excluded from cross-instance equality;
their lifetime checks are separate. A captured scrolled-out cell refuses stale
input without changing a completed checkpoint. Closing the grid window refuses
old window input and preserves the surviving inspector image.

Dispatch has an explicit step bound, checks each call's event/owner charge and
requests pending frames explicitly. No sleep, real clock advance, desktop
session, native GPU or OS service is used. This regression covers its stored
grid/editor/history/window composition; the P5/P6 workspace with canvas,
worker-built views, cancellation, a versioned large replay record and the full
P9/P11 acceptance requirements remain separate.

The Manual scene size is the requested physical viewport, including its native
scene padding/border. The body grows into the allocated content area; requested
window dimensions are not reapplied as a body-content minimum. This keeps grown
roots and narrow far-side split panes inside the completed viewport. Explicit
application/theme content minimums retain their normal layout meaning. This is
a software-driver layout correction, separate from live window-size acceptance.
