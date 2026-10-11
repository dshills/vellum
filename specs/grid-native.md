# Managed grid adapter checkpoint

This specification adds `ui.DataGrid` to the portable Grid contract in grid.md.
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
PageDown navigation are supported. Home/End/modified variants, editable cells,
range selection, sorting headers and a double-click action are outside this
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
