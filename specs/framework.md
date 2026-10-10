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
context, an ordinary Go worker and a typed completion callback. Worker count and
undelivered results share one bound. Progress is coalesced; completion waits for
mailbox credit or lifetime cancellation. Workers cannot mutate live UI state.

## Document edits

The bounded [history contract](history.md) supplies sparse application-owned
Undo/Redo and revision-aware dirty/save state. Native command shortcut scopes
preserve focused text editing before document fallback. The workbench's component
changes use the shared journal and invalidate stale worker results on replay.

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

## GPU canvas

`canvas.Scene` owns explicit rectangle objects, viewport transforms and picking.
The Cogent `GPUCanvas` adapter uses its public `xyzcore.Scene` direct renderer
on the shared WebGPU device, with one reusable mesh and retained instances.
Wheel zoom anchors the pointer; dragging pans; clicking selects a component in
the inspector; Fit/Zoom commands share the command registry. Native scene
destruction releases GPU resources. The current canvas is a bounded rectangle
renderer, with a 10,000-object limit; arbitrary shaders, image tiles, spatial
indexes/LOD and automatic physical-device recreation remain future work.
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

Windows/Linux service adapters implement clipboard, chooser transports and
window-local menu widgets. Actual target runtime, native IME/accessibility and
native Wayland acceptance remain open; [platform support](platform-support.md)
reports those gates without inferring qualification from fixtures/crosscompiles.

The independently relocated private SDK uses an explicit alpha version, complete
archive integrity checks and public consumer scenarios. [Release artifacts](release.md)
document packaging, the selected `github.com/dshills/vellum` namespace and MIT
license, and remaining distribution audit/qualification gates for public release.
