# Framework contract

This describes the implemented experimental framework. It preserves the API and
limits documented with the five implementation phases; unresolved foundation
checks still prevent a production-ready claim. See [implementation status](../docs/implementation.md).

## Application ownership and delivery

`app.Application.Post` provides bounded worker delivery with explicit refusal.
Application state changes in `Do`, posted callbacks, or UI actions; views run
under the same serialized owner. Calls to `Do` and `Drain` are not reentrant.
Views only describe UI; they do not start workers or mutate domain state.

Queue capacity bounds pending callbacks, plus the one callback currently executing.
Driver wake functions are nonblocking notifications and never enter the UI owner.
Native wakes use Cogent's lossless custom-event deque; they are coalesced while
one wake is outstanding. A panic closes the application and propagates.

## Commands and tasks

Commands live in an owner-thread registry. Availability is rechecked on invocation
from menus, toolbars, shortcuts and the palette. `tasks.Start[T]` accepts a
context, an ordinary Go worker and a typed completion callback. It reserves
completion mailbox credit before executor admission and refuses immediately if
that reservation is full. Worker count and undelivered results share one bound;
task credit returns only after both worker termination and owner terminal
delivery. Progress is coalesced. Workers cannot mutate live UI state.

## Document edits

The bounded [history contract](history.md) supplies sparse application-owned
Undo/Redo and revision-aware dirty/save state. Native command shortcut scopes
preserve focused text editing before document fallback. The workbench's component
changes use the shared journal and invalidate stale worker results on replay.
The portable [document session](document.md) captures detached snapshots, performs
file IO/decode on a worker and commits only a matching revision on the owner.
Window.CloseRequest permits asynchronous unsaved-change prompts; confirmed
CloseWindow forces teardown. The workbench persists its component document.

The public [editor model](editor.md) binds exact UTF-8 document bytes and directional
scalar selection to `ui.MultilineText`. One application-owned journal records
sparse committed edits; native preedit stays transient until composition commits.
The Mac adapter uses the originating native event FIFO and asynchronous settlement
before persistence or replay. The independent `examples/editor` application gives
each of its two windows an independent model/session and preserves CRLF and final
empty lines through Save/Open. Native protocol fixtures and external IME/assistive
technology acceptance remain distinct evidence.

## Workbench layouts

The `workbench` package owns a validated split/tab tree. Move, split and restore
are atomic model operations; malformed or incompatible layouts leave the current
workspace intact. The example provides dock/split/reset commands and resizable
native panes. `-layout /path/layout.json` enables explicit Save/Restore; startup
does not overwrite preferences. Pointer drag supports docking/reordering inside a window. Explicit tab menus
float panels into independent windows and redock them into another group/window.
Session v2 persists the complete window layout, bounds, title and focus, while
Workspace v1 remains supported. Closing a floating window transfers its panels to
a surviving window. Panel factories share application data and create fresh native
extensions per mount. See [workbench interaction](workbench-interaction.md). Dock containers may be recreated for topology changes; compatible panel contents
retain their native identities. See [keyed reconciliation](keyed-reconciliation.md).

## Virtualized data

The example generates one million stored component records and 100,101 tree
nodes, with 20 table-row and 24 tree-row controls. Wheel scrolling and paging
materialize the current viewport; stable domain IDs connect selection to the
inspector. Table sorting runs against an immutable snapshot in a bounded task;
revision checks discard stale results. Tree expansion rebuilds a lightweight visible-node index; subsequent viewport
materialization takes O(viewport) work even at deep offsets.
Domain records, tree indexes and optional sort indexes retain O(N) memory. The
viewport bound is not a claim that all application memory is independent of N.

`-tree-rows` controls the outline's explicit subset of the component dataset
(default 100,000); its label shows the subset and full table count. Versioned
sources expose `Revision()` so same-length edits invalidate active projections.
Providers without revisions explicitly reset projections after edits. Source
growth preserves the viewport offset, clamping only when necessary.

Tree expansion rebuilds a lightweight O(N) visible-node index; scrolling materializes O(viewport) records. Worker projections carry source revision metadata and require an owner-captured table token when installed. Background sources must be immutable snapshots. The Cancel command also cancels sorting.

The separate `virtual.Grid` projects an owner-confined indexed source in two axes,
with fixed or indexed variable row heights, fixed headers, frozen leading columns, bounded overscan and stable
row-key/column-ID active cells. It validates accessed indices instead of copying
or scanning the full row source. Snapshot rectangles and clips are portable model
geometry. The [managed adapter](grid-native.md) separately owns bounded native
cell pools, clipped measured geometry, routed input and viewport accessibility.
Source revisions preserve the active key and top-row anchor where still present,
subject to viewport extent clamping;
retired opaque cell refs refuse activation after a viewport/source transition.
`BuildGridProjection` prepares immutable loaded sort/filter and optional height
indexes on workers; opaque owner requests refuse superseded results and preserve
current drafts and selection. View revisions are separate from the captured full
domain revision. See [worker-built views](grid-projection.md) and the independent
stored-data consumer in `examples/loadedgrid`. See the [grid contract](grid.md). This adds an indexed foundation for the P6
million-row workload. Optional [row selection](grid-selection.md) captures
immutable range/All membership, with bounded key toggles and detached worker
enumeration. The active cell remains a separate cursor. Optional [cell editing](grid-editing.md)
adds one versioned draft with validation owned by the application and atomic
immutable-source admission. Native performance and loaded-data/tree acceptance
remain open. Existing Table/Tree interfaces remain supported.

## GPU canvas

`canvas.Scene` owns explicit rectangle objects, viewport transforms and picking.
The Cogent `GPUCanvas` adapter uses its public `xyzcore.Scene` direct renderer
on the shared WebGPU device, with one reusable mesh and retained instances.
Wheel zoom anchors the pointer; dragging pans; clicking selects a component in
the inspector; Fit/Zoom commands share the command registry. Native scene
destruction releases GPU resources. That API remains a bounded rectangle
renderer, with a 10,000-object limit. [Retained drawing](drawing.md) separately
adds immutable arbitrary indexed 2D meshes, strokes, waveform reduction and a
rectangle spatial index through the shared-device `MeshCanvas` adapter.
[Retained images](images.md) add immutable RGBA tiles, changed-tile uploads and
nearest/linear filtering in that adapter, including image-only canvases.
[Advanced GPU views](advanced-gpu.md) expose a versioned shared-device context,
ordered preparation and custom shader/draw callbacks, retained targets and optional
depth. An independent 3D viewport exercises camera-only uploads and native depth
pixels. An independent imaging consumer adds compute/float-target preparation and
bounded asynchronous captures. Broader performance qualification and automatic
physical-device recreation remain future work.
Demand painting, live resource rebuilding and actionable device-loss shutdown are
implemented; see [rendering](rendering.md). Runtime verification
covers macOS; the integration is not a production performance certification.

## Layout and appearance

See [layout and theming](layout-theme.md) for the optional value API and native
reset semantics. The driver snapshots its theme for each run; workbench panel
trees inherit that theme through the optional ThemeExtension adapter contract.

## Native semantics, services and release scope

The macOS driver publishes managed controls and active workbench content through
NSAccessibility, including actions, post-layout geometry and modal focus scopes.
Virtual views publish bounded row snapshots and explicit paging semantics. The
[accessibility contract](accessibility.md) distinguishes native selector probes
from external assistive-technology acceptance and target input/AX gaps.

External file-drop callbacks deliver bounded, privately copied file references
on the application owner. The Mac adapter handles native file URL payloads and
modal/active target routing. See [file drops](file-drop.md) for coordinates,
limits and the separate physical Finder acceptance gate.

The optional [clipboard extension](clipboard.md) adds bounded Mac PNG/HTML/custom
MIME representations and capability reporting. Its plain-text methods share the
same cancellation-aware bridge. Real clipboard acceptance remains deferred.

The optional [tray service](tray.md) adds a retained Mac status item with bounded
PNG icons, accessible naming and shared command validation. Native object/menu
interaction and real shutdown acceptance remain deferred; offscreen fixtures
and the patched desktop cleanup lifecycle establish their tested source behavior.

The optional [notification service](notifications.md) adds explicit Mac permission
queries/requests, immediate local delivery and targeted ID withdrawal in app
bundles. Bounded shared dispatch and offscreen content/settings fixtures remain
separate from real permission, presentation and packaged runtime acceptance.

Windows/Linux service adapters implement clipboard, chooser transports and
window-local menu widgets. Actual target runtime, native IME/accessibility and
native Wayland acceptance remain open; [platform support](platform-support.md)
reports those gates without inferring qualification from fixtures/crosscompiles.

The independently relocated private SDK uses an explicit alpha version, complete
archive integrity checks and public consumer scenarios. [Release artifacts](release.md)
document packaging, the selected `github.com/dshills/vellum` namespace and MIT
license, and remaining distribution audit/qualification gates for public release.

## Explicit execution foundation

[Execution](execution.md) adds selected/FIFO manual workers and bounded real/manual
timers, plus separate reserved completion admission. Tasks and document sessions
accept an injected Executor; owner batches retain an initial arrival cutoff.
Worker and owner delivery termination remain independently visible.

The internally qualified [managed replay checkpoint](replay.md) supplies explicit
fonts/theme and routes actual pointer/key input through retained controls, focus
and modal hit testing on the offscreen backend. Its fixed 51-operation record
and nine captures match exact model state, measured semantics/geometry, focus,
UTF-16 selection and completed software image identities at batches 1, 7 and 64.
Input admission, frame requests, bounded dispatch, selected workers and clock
advancement remain separate operations; originating widget timers do not dispatch
implicitly. Readiness distinguishes due/running/blocked work from future timers.
One active managed application per process may own multiple windows on its shared
clock. The caller keeps Cogent global configuration fixed and injects that same
borrowed executor into task/document managers. See the
[checkpoint report](../docs/reports/macos-replay/README.md).

The concrete built-in Workbench now participates in bounded managed replay.
Complete hosted-panel preflight, scoped target IDs, retained drafts, actual dock
tab/split gestures and independent session windows have offscreen regressions.
Tab clicks request owner model activation, and queued gestures lose eligibility
permanently when their originating visibility/modal scope is revoked. See the
[workbench checkpoint](../docs/reports/macos-managed-workbench/README.md). Full
professional workspace records and native acceptance remain open.

The [local semantic automation](automation.md) endpoint now provides opt-in Mac
transport for explicitly registered typed domain queries/actions. Full external
professional automation workflows remain unqualified.
The complete P9 professional workloads, large-data replay,
native services/IME/accessibility, GPU presentation and sustained
performance acceptance remain open.
