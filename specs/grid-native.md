# Managed grid adapter checkpoint

This specification adds `ui.DataGrid` to the portable Grid contract in grid.md.

Grid accessibility now follows the scoped pending/completed publication contract
in [accessibility.md](accessibility.md#managed-grid-publication). Last-completed
cells remain readable while current layout is pending; their actions are revoked.
Current semantics and measured frames publish together, retaining overlapping
native objects only within the same binding incarnation and parent. Inspection
exposes pending provenance; this does not change `InspectGrid`'s current-layout
requirement or qualify the P6 targets.
It does not replace the Table/Tree extensions, implement editable cells, qualify
P 6 performance targets, or certify Windows/Linux or screen-reader workflows.

`Node.Grid` is required only on the DataGrid leaf. `GridChange(GridEvent)` is a
notification after an accepted owner-held selection or viewport transition;
the model has already changed. Reasons are `GridSelection` and `GridViewport`.
Refused and no-op requests produce no notification. Grid/source state belongs
to the caller and must be mutated on `Application.Do`/PostWait; source Refresh is
explicit. A Grid has one live binding across windows. A same-window node ID
remount retires its old incarnation; failed preparation restores the old binding.
Teardown does not reset or close caller-owned state. Construction rollback also
retires a provisional binding on source callback panic/Goexit and propagates
the abnormal return; a prior remount binding is restored on refused preparation.

Four clipped native panes render frozen/scrolling headers and frozen/scrolling
body cells. Header Y and frozen X stay fixed. The host is the only Tab stop;
pooled cell text and scrollbars cannot take keyboard focus. Arrow and PageUp/
PageDown navigation are supported. Optional row selection adds the exact
modifiers described below. Home/End/other modified variants, editable cells,
cell-range selection, sorting headers and a double-click action are outside this
checkpoint. Tab/ShiftTab and modal Escape keep the existing managed behavior.
A left press captures an opaque current cell reference, binding incarnation,
scale and clipped scene rectangles at MouseDown. MouseUp
cannot reinterpret a recycled slot or a moved/resized scene rectangle. Owner delivery checks lifetime, binding,
modal/disabled scope and generation. Actual native tab selection is checked
against owner-captured tab/page ancestry, so optimistic native tab switches
cannot deliver old-page AX or queued selection intents before controlled Change.
Managed/workbench tab topology is owner-held and nonclosable; the queued check
reads only the tab mutex-protected current page, not native geometry or another
scene lock. Wheel input accumulates fractional logical
pixels independently on each axis, refuses nonfinite input and bounds a single
issued delta. Sliders use finite normalized ranges accounting for frozen/header
panes; they do not create a second independent scroll offset.

`GridConfig.RowSelection` defaults to false, preserving the existing active-cell
cursor and keyboard behavior. Opt-in grids bind portable row membership to the
accepted visible projection: a plain left click replaces membership, exact
Shift-click extends from the portable stable-key anchor, and exact macOS
Command-click toggles the clicked row. The modifier intent is captured at
MouseDown, together with the existing cell, lifetime and scene-rectangle guards;
MouseUp never reinterprets it. Control, Option and combined Command+Shift are
unsupported selection modifiers. CapsLock/NumLock do not change semantic intent.
Plain arrow/Page navigation replaces selection; exact Shift-arrow/Page extends;
Command+A captures All, and plain Escape clears selection in the main scene.
Existing modal Escape handling takes precedence and never clears an underlying
grid. Native command handlers and focus/modal/disabled routing remain in force.
Native input requires valid current Grid state. If a source violates its captured
immutable-index contract, the explicit owner `ClearSelection` recovery API is
separate from keyboard Escape; recovery must precede renewed native delivery.
No failed or no-op request emits `GridChange`, and one accepted intent does not
ask the caller to apply the model mutation again.

Row membership and the active cursor are separate. The accepted row-member flags
paint the normal selection background across materialized frozen and scrolling
cells; the active cell has its own inside border even when toggled out. Render,
Position and accessibility publication consume these flags without querying
selection indices or domain rows. `GridInspection` adds `RowSelection`, detached
`GridSelectionState` metadata and `Cells[].Selected`; `Active` retains cursor
meaning. Selection changes advance portable generation and retire measured cell
refs exactly as other effective transitions do. Inspections refuse a pending
selection projection until its completed native frame.

All/range capture uses the source's immutable `GridRowIndex`; no native widget
enumerates the captured domain. Visible row/cell accessibility selected attributes
use row membership in row mode, while legacy cell selection remains active-cell
based. Native SelectedCells returns only materialized visible cells, not every
captured key. Ordinary AX press replaces with its accepted row and preserves the
existing generation/lifetime/tab guards. The captured selection count can differ
from a later filtered live row count; the inspection publishes the portable
captured scope honestly. Sparse toggle key retention remains bounded by the
portable `MaxSelectionKeys` and text budget; native pools/local geometry limits
are unchanged. Enumeration for a domain action belongs on a worker through the
immutable snapshot, with current domain identity validation before application.

Grid dimensions are integer logical pixels via Core units.Px (DPI/96). Existing
Node Min sizes remain Dp. Shared rounded edges map logical coordinates to scene
pixels. Native policy caps configured materialization at 256 rows,64 columns,
4096 cells and 8 MiB snapshot text before widget allocation. Larger portable budgets
remain valid outside this adapter. Native local viewport and materialized row/header/column
dimensions are at most 65536 logical px and natural relative rectangle edges are
within±262144 logical px; scale must be finite, positive and≤32, producing edges
within±8388608 scene dots. These checks do not cap domain row counts or total
vertical scroll extents. Portable MaxInt dimensions can be refused natively.
An optional `virtual.GridRowMetrics` source supplies indexed variable row
boundaries and pixel lookup. `GridState.VariableHeights` reports that selection;
`GridConfig.RowHeight` remains a positive fixed-height fallback. Its native size
cap is ignored when indexed metrics are active: the actual materialized row
rectangles, including overscan, retain the local 65536 px bound. This does not
permit oversized visible rows or increase record/text budgets. Heights, order
and membership must be immutable within a source revision, and source replacement
preserves the stable row-key anchor with its in-row pixel offset and active key.
Page navigation uses the body-height distance from the current row's actual
top, with at least one row of progress except at the corresponding end; wheel
scroll remains a logical-pixel offset. Native rendering consumes the accepted
bounded rectangle projection and never queries the metric source. Callers may
prepare `RowHeights` before installation or supply a bounded index; the adapter
never builds an entire row-height index on its owner.

Hidden oversized column definitions are checked if/when materialized. Pools
reuse active widgets and destroy surplus slots to release private shaped-text
caches, with no whole-source allocation.

A changed native projection retains overlapping domain identities and assigns
current text, rectangle, visibility, selection and cursor state in accepted
snapshot order. The ordinary full fallback performs one recursive host update
after placement; an admitted local pass updates only changed cells as described
below. Placement does not independently repeat either path. Private editing
synchronization and focus handling always run, while an unchanged field skips
its redundant full Update. Initial mounting, theme changes and other parent
updates may perform additional work.
After a complete successful sync, an unchanged validated model generation reuses
that projection without rereading Cell, moving pooled slots or recursively
updating the host. State/current metadata checks still run. Effective inherited
disabled or help-tooltip changes still update the host; descriptor styles, theme and topology
retain normal reconciliation updates. The private editor is synchronized even
on this reuse path, preserving its separate selection, restore, preedit and focus
lifecycle. An incomplete sync cannot become a reusable projection.
On refresh, scrollbars also retain constant-time canonical value restoration: a native drag
that rounds to the same model pixel or is refused cannot leave a different local
value. Only a differing scrollbar value is assigned and marked for rendering;
an already canonical value does not request another paint.

This relies on immutable source data within an accepted stamp. Unstamped Cell
mutations or callback failures cannot be discovered by reusing a projection;
callers must explicitly Refresh a changed revision or ReplaceSource. Exact
worker-built projections additionally validate their borrowed input stamp when
State is read, including a fixed-height or empty projection. The adapter keeps
completed geometry publication and layout requests; source/projection reuse
does not suppress legitimate native resize, DPI, style or editor repaint work.
A refused sync keeps the existing reconciliation failure and retirement behavior.

Render/Position never read the source/model or invoke user callbacks. Completed
native rendering publishes clipped actual ContentBBox/TotalBBox geometry. A
changed actual content size/scale posts one checked immutable layout request to
the owner, independently of input eligibility. Queue refusal can retry at a
later actual frame. A model capacity refusal preserves state and reports once
per size request, without an invalidation loop; a different valid size can
recover. Pending/current-invalid geometry cannot activate a cell or be inspected.

`InspectGrid(a,model)` is owner-held and returns a detached GridInspection after
an accepted current frame. It exposes source stamp, viewport, full column count,
binding and cell incarnations, projection generation, monotonic completed-frame
count, scale, host bounds and bounded cell/header records. Bounds are actual
clipped scene pixels including modal origin. It refuses `ErrGridLayout` while
model/source/size/scale transitions are ahead of measurement, `ErrGridClosed`
after retirement, or `ErrGridBinding` for a foreign owner. It does not render,
dispatch, mutate, fetch Cell strings or allocate hidden cells.

The retained accessibility tree exposes a table, full row/column counts, bounded
visible rows/cells/column headers, absolute row/column index ranges and selected
cells. IDs encode gridID,rowKey,columnID rather than pool slots. Cell press uses
the accepted opaque reference and existing bridge generation/lifetime guards.
The `grid:` Node.ID prefix is reserved across all managed descriptors for
generated cell, row and header identities, including nested/workbench/modal views.
Validation refuses ordinary descriptors with that prefix before native allocation,
without fetching hidden rows or cells. `GridCellID` may be used as a captured
dynamic input target; it must not be reused as a descriptor ID.
Workbench cell press qualifies its ensuing host focus with the same namespace;
ordinary action focus requests keep their existing behavior.
Cocoa table selectors return only current visible cells; hidden lookups return
nil. The in-process inspection adds detached row/column metadata. Inert native
selector fixtures prove these attributes without creating NSWindow; external
AX clients and human screen-reader operation remain separate runtime gates.

NewManual captures current grid geometry and dynamic cells without calling
callbacks directly. Actual pointer/key/wheel events route through Core scene
focus/modal handling and the grid host. Dynamic cell incarnations retire on
projection changes. A source inspection failure remains an explicit failure,
rather than being mislabeled pending layout. Current inspection uses the
portable State revision/extent verification, including a constant number of
metric boundary reads for variable sources; it does not enumerate rows, perform
pixel search or refetch Cell strings. Desktop GUI, physical input,
external AX and performance acceptance are not proved by these offscreen checks.

## Inline cell drafts

`GridConfig.CellEditing` opts into one private managed `core.TextField` per live
`GridEditState.Session`. The session is independent of pooled visible cell slots.
The field uses the existing Cocoa text-input attachment, FIFO admission, UTF-16
selection/range geometry and composition settlement. No separate document model
or grid-owned source mutation is introduced. `MaxEditBytes` (default 64 KiB,
maximum 1 MiB) bounds the single-line UTF-8 field; newline/NUL and over-budget
native requests are refused before Cocoa mutation/reservation. Legacy widget
refusal restores accepted draft text after the widget event; it does not claim
atomic preservation of legacy private undo/selection.

Plain Return or F2 on the grid host begins the current active cell. In the field,
Return settles composition and earlier admitted text, then invokes the current
`Node.GridCommit(ui.GridEditEvent)` on the owner. This is a commit intent: the
application validates the draft, calls `PrepareEdit`, builds an immutable source
candidate, and calls `CompleteEdit` before replacing its own source pointer.
`GridChange` with reason `GridEditing` reports an already accepted draft/begin/
cancel transition; callers must not apply it again. Native source callbacks stay
on the owner, outside Render/Position.

The visible writable field replaces the grid host's one Tab stop. Tab and focus
loss preserve/synchronize the draft without committing domain data. The first
Escape during preedit cancels composition; otherwise Escape cancels the draft.
An active managed modal retains its existing Escape precedence. Double click,
Tab-to-commit and spreadsheet traversal after commit are outside this checkpoint.
While the field owns focus, text navigation/select-all/private field undo use
ordinary text routing rather than row-selection shortcuts.

Source reorder, filtering, conflict and viewport movement retain the recoverable
portable draft. A hidden/nonready field relinquishes activation and restores the
host Tab stop; it does not steal focus when it reappears. Already admitted
same-session records still finish at their originating FIFO position after a
view-only hide/conflict, followed by native deactivation acknowledgment; fresh
mutations are refused. Owner draft replacement, Cancel/new session, retired
binding, disabled/modal/tab scope and closed lifetime invalidate old callbacks.
A new session always gets a new text client. Native old-client sealing may only
acknowledge unmark/cancel; it cannot admit a new edit.

`SettleGridEdit(ctx, application, originatingWindow, grid, callback)` requires
owner-held admission and an existing exact private binding. One settlement per
window is allowed. Immediate refusal does not invoke callback; an admitted
request completes once on the owner while Application is open, including
cancellation/lifetime errors. Whole-application close suppresses user dispatch.
Do not synchronously wait on Cocoa main or the application owner. Save must use
this barrier before validation/Prepare. A hidden/conflicting draft may settle for
recovery; `PrepareEdit`/application validation still refuses domain commit.

In `NewManual`, settlement enqueues only the originating text-pump record and is
completed by explicit `Step`, including context cancellation. It starts no
worker, main-queue dispatch, frame or implicit clock advance. Current Cocoa
runtime acceptance, external accessibility and physical IME remain separate from
actual offscreen routing/geometry qualification.

`GridInspection.Edit` is the accepted portable draft at the completed frame.
`GridInspection.Editor` (when visible) reports detached private field ID/session,
actual preview text, focus and clipped scene-pixel bounds. The retained inline
field receives the current managed TextField theme (typography and colors),
then applies its mandatory cell geometry overrides without changing the draft
or text-client identity. Its accessible name uses the visible column title and
current display row when materialized. `GridEditorID` is a
reserved generated target for manual/AX observations, never a descriptor ID.
The accessibility tree publishes the inline field as a text field beneath its
table, using the same selection/range geometry bridge and workbench namespace.
Inspection does not render, drain, fetch hidden cells or invoke callbacks.


## Retained viewport layout

The private cell pool matches `(RowKey, ColumnID)` and headers match `ColumnID`.
Departing cell slots first serve every entering slot in their existing frozen
or scrolling pane. Only after all same-pane matches are assigned does an old-order
fallback reuse donors across panes when populations change. Exact domain-key
matches keep precedence, each widget is consumed once, and unused widgets retire
in their old order. Combined row/column scrolling with unchanged pane populations
therefore avoids incidental reparenting and the resulting full-scene layout.
The pool still contains only the current admitted projection; this adds no
offscreen row cache, retained widgets or relaxed geometry/layout guard.
Overlapping identities keep their widgets and shaped text. Departing widgets are
bounded donors for entering identities; surplus widgets are destroyed and their
private caches retired. No off-viewport row cache is introduced. `g.cells` and
headers remain in accepted snapshot order, and each owned pane's child order
matches that order. A private inline field stays last, above cells. Opaque model
refs and completed-frame incarnations keep their existing generation/lifetime
rules independently of widget retention. Source callbacks still materialize the
bounded accepted snapshot on the owner; layout calls no source callbacks.

An admitted fixed-allocation Grid update runs updater/style only for changed
content, selection/cursor/visibility or newly created cells. On the originating
render context, the local pass sizes changed allocations/text and repositions
all pane descendants, including clipped absolute bounds. Integer physical edge
rounding is unchanged. Unchanged text keeps its actual shaped output when its
physical allocation and styling are stable. Complete scene painting follows;
there is no partial scene image, altered painter order or GPU upload change.

Passive grid labels can defer the ordinary updater's allocation shape to the
admitted local sizing pass (owned patch 0055). Value/updater callbacks and rich
parsing still run; old shaped output is cleared before applying new rich styles.
Every updated label must remain passive and belong to the root's sizing set.
Interactive text, wrong/retired/detached roots, missing requests and full-layout
fences refuse without mutation and take the ordinary updater. Sizing clears the
pending state; Render restores allocation shaping if a queued pass is canceled.
This preserves inline editing's immediate shaping and complete scene paint.

Core admits at most 64 live local-layout roots per scene and coalesces actual
initialized tree identities. A full rebuild/restyle/layout overrides previously
queued local work. Capacity or local-layout refusal promotes ordinary full scene
layout. Requests raised during a pass remain for a later frame. Root/scene
retirement clears queued references; mutation requires the originating render
context or exclusive construction, never a foreign window lock under the owner.

Initial construction, changed external descriptor/theme/topology/allocation,
DPI/shaper context, uncertain guards and inline-editor layout changes retain
the full path. Private editing synchronization always runs; an unchanged field
publication skips only its redundant full Update. Creation/session, theme,
accepted text/revision/restore, queued selection, actual visibility/disabled/focus,
parent/scene/rectangle and tooltip changes still request full field layout.
Native preedit/selection/caret APIs keep their existing rendering requests. Core's atomic `GlobalLayoutPending` reports ordinary
requested full layout safely to inspections; it does not authorize foreign
geometry mutation. Grid also fences queued local geometry. Current hit/AX
inspection refuses pending geometry even at unchanged portable generation;
last-completed data remain available to the existing pending/completed AX
transaction. A completed frame republishes that transaction, including a full
layout that leaves its public geometry unchanged. Public inspections remain
fully detached from native scene objects.

This contract does not promise the unchanged 4 ms p95 / 6 ms p99 targets. The
ordinary complete-frame campaign and the broader macOS acceptance workloads
remain authoritative; unchanged full-scene raster/upload is subsequent work.

A private inline field may repaint preedit without requesting layout. When its actual editor text, focus, visibility, session or completed bounds differs from the last Grid publication, owner synchronization requests one bounded host repaint so capture follows the field render. The unchanged completion refresh schedules no further paint or layout. Other controls in the same scene must also quiesce: the native tree compares its complete bounded snapshot (including same-generation pending status), effective disabled state and help before requesting full updates, while still restoring a noncanonical native scrollbar value.

## Reusing intrinsic grid text across rounded cell edges

A passive managed preformatted label can retain completed intrinsic shaping when
its natural width/height remain unchanged and only rounded physical cell edges
change. Core SizeUp starts from the same style-derived zero content size; its
WhiteSpacePre SizeDown does not use the cell allocation. Pane positioning still
assigns every final rounded exterior/content bound and clipping. Full painting
and accepted snapshot order remain unchanged.

The private predicate requires the same live scene and plain owned leaf Text,
no Parts/children, no interactive abilities, non-nil shaped content, zero
Grow/Min/Max and no shadows. It refuses allocation/parent-relative, custom,
unknown or nonfinite unit values across the fixed Style.ToDots field set. It
changes only allocation-rounding dirtiness. Content, selection, cursor,
visibility, parent/new/nil-paint and previously queued dirty work keep their
original sizing path; natural dimension changes and all existing full-layout,
theme, DPI, scene, configuration and editor fences remain effective.

No new cache, widget retention or global state is introduced. Skipped SizeFinal
may leave private UnitContext element dimensions at an older allocation; the
predicate admits only styles whose semantic values do not use those dimensions.
Later normal styling rebuilds its context from current geometry. This does not
promise equality of unused private context numbers, one shaper call per frame,
or arbitrary dimension-dependent/custom-style reuse.

Regressions compare current pixels, complete public geometry/semantics and
visible private intrinsic/glyph-interaction geometry against full scene layout
and the frozen prior sizing rule. They cover fractional DPI, mixed Unicode,
multiline/tab/long/escaped text, cursor changes, real dependent-unit/parts/wrap
fallbacks and size-sensitive lazy text restoration. Intentional missing-unit
and lost-dirty-work faults fail actual glyph/offset assertions; the prior rule
fails useful-reuse assertions. Kernel timings exclude painting and native
frames. Native 4 ms p95 / 6 ms p99 and full Mac gates require separate evidence.
