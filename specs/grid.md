# Fixed-height two-dimensional grid draft

This independent portable checkpoint adds a `virtual.Grid` without changing
`Table`, `Tree`, or their source interfaces. It qualifies bounded projection,
geometry and identity behavior. It does not qualify native rendering, frame
performance, accessibility, editable cells, variable heights, all-row selection,
tree indexing, or completion of P6.

## Domain source and ownership

`GridSource` has `Identity() string`, `Revision() uint64`, `RowCount() int`,
`RowKey(index) (string,error)`, `IndexOf(key) (index,found,error)` and
`Cell(rowKey,columnID) (string,error)`. Identity is a nonempty UTF-8 incarnation
identifier of at most 1,024 bytes. A row key cannot be reused for a different
entity within one incarnation. Revision is nonzero and increases on changes
to content, membership or order within that incarnation. A new incarnation may
restart its revision. Reusing an incarnation for different domain entities is a
source contract violation.

The caller serializes a Grid, its source mutation, and all calls on one owner.
Independent Grids may run on independent owners; there is no package-global
state. This is not a concurrent mutable-source API. Metadata must be O(1), RowKey
must be direct indexed access, and IndexOf must use a prepared O(1) or O(log N)
index, not a row scan. Calls must return promptly and must not do IO or call back
into Grid. Whole-source validation/index creation belongs to source construction
or an immutable worker snapshot, not to an owner Grid transition.

Within an accepted revision RowKey and IndexOf are inverse and keys are unique,
nonempty valid UTF-8 of at most 1,024 bytes. Grid checks that invariant only for
accessed rows; it intentionally does not enumerate hidden keys. Cells are plain
UTF-8 strings addressed by stable row key and column ID. Grid preserves source
errors wrapped with `ErrGridSource`, rather than silently substituting cells.
The source owns data and index memory; Grid retains neither a million-row copy
nor a million-entry projection array. On successful retention Grid clones source
incarnation, column IDs/titles, active keys and anchor keys, so a small substring
does not pin a large caller backing allocation. A source may itself retain its
domain memory; that borrowed source memory is outside the Grid payload budget.

Public methods refuse changed metadata with `ErrGridStale` until `Refresh`
acknowledges a new revision. `Refresh` with unchanged metadata is a no-op.
`ReplaceSource` explicitly installs a borrowed indexed snapshot and always
retires references, even when its stamp is unchanged. Within the same
incarnation revision cannot decrease, and membership cannot change at an
unchanged revision. The active key survives order changes if the new index still
contains it. Preserving the top-row key and fractional row offset is best effort:
normal extent clamping may place that key farther down the viewport near the
source end or when the viewport becomes taller. Deleted/filtered-out active keys clear active
state; a new incarnation clears active state and scroll offsets. Refusals do not
partially install Grid state. Source mutation itself is not rolled back.

## API and geometry

`NewGrid(source, []GridColumn, GridConfig, GridViewport)` returns a pointer or an
error. The Grid cannot be copied by value; nil, zero and copied Grid instances
return `ErrGridInvalid`. Column definitions are copied once, have unique stable
UTF-8 IDs and titles, positive integer logical-pixel widths, and at most 4,096
definitions. The summed widths and total row extent plus header must fit `int`.

`GridConfig` specifies fixed `RowHeight`, optional `HeaderHeight`, count of
`FrozenColumns`, row/column overscan, and capacities. `GridViewport` specifies
positive Width/Height and nonnegative ScrollX/ScrollY. Grid clamps scrolling
to content extents. The body excludes header height; nonfrozen columns exclude
the leading frozen-column region. Header Y is always zero. Frozen columns retain
their X while other columns pan. There are no frozen body rows.

`SetViewport` and `Scroll` validate the prospective pool before committing. A
viewport that exceeds a capacity returns `ErrGridCapacity`, preserving the
previous viewport and references. It never truncates visible cells or pretends
an unrenderable resize was accepted. A renderer must surface/refuse that case or
provide a larger checked budget before constructing its Grid.

`Snapshot` returns detached `GridSnapshot{State,Headers,Cells}`. It clones each
materialized row key once and shares that detached key across the row's cells and
opaque references. It clones each cell string after accepting its payload length
and UTF-8. Snapshot headers/column IDs and state identities share already detached
Grid-owned immutable strings. Transient source metadata comparisons are not
cloned because they are not retained. Opaque cell references retain only a
nonzero-sized identity token, never a Grid/owner/source backpointer, so retaining
a detached snapshot does not itself keep the borrowed domain model alive.
Header and cell
records include natural Rect, region-clipped Clip, stable column/row identity,
absolute row/column indices, and frozen/active flags. `State.Columns` reports all
column definitions, independently of virtualized Headers; `State.Stamp.Rows`
reports the full accepted row count. Cell order is row-major;
column order follows source column definitions. Overscan records can have empty
Clip and must not be painted or hit tested as visible cells. HeaderHeight zero
produces empty header clips. If a narrow viewport is consumed by frozen columns
or header, that pane has no visible scrolling cells. Navigation may retain a
valid but clipped active identity in that case; it does not claim it is visible.

## Bounds and costs

Capacity defaults are MaxRows 256, MaxColumns 64, MaxCells 4,096, MaxTextBytes
8 MiB. Accepted maxima are respectively 4,096, 4,096, 1,048,576, and 64 MiB.
Negative capacities are invalid; zero selects defaults. Overscan counts must be
nonnegative and are clipped to source extents, not silently reduced to fit a
pool. MaxRows/MaxColumns include overscan and frozen records; MaxCells bounds
body records. Headers are separately bounded by MaxColumns.

MaxTextBytes bounds retained column-definition payload during construction and,
separately, each snapshot's unique row keys, header IDs/titles, and cell strings.
It is not a bound on source allocation before returning a string, entire encoded
snapshot size, allocator overhead, or total UI heap. Accepted strings are detached
from caller backing allocations, including empty substrings. Snapshot metadata/record memory is proportional
to the configured row/column/cell capacities. Payload length is checked before
UTF-8 scanning. Metadata is checked again after source reads; revision drift
returns an error and a zero snapshot, never partially published mixed cells.

Construction costs O(column definitions) plus one anchor lookup. Viewport
geometry uses a bounded column prefix index and binary search, then visits only
frozen/visible/overscan columns. Snapshot calls RowKey and IndexOf once per
materialized row, and Cell once per materialized cell. Refresh and navigation
perform a bounded number of indexed active/anchor/key lookups independent of
total row count. No Grid operation scans or copies the million-row source.

## Active cells and renderer identity

`JumpToKey(key,columnID)` activates the indexed cell and reveals it where pane
space permits. Oversized cells align their leading edge and are clipped.
`Move(rowDelta,columnDelta)` clamps to row/column extents; with no active cell
it starts at the viewport's first row/first column before applying the deltas.
`Page(-1|1)` moves max(1, floor(bodyHeight/RowHeight)) rows. Empty sources and
unknown keys/columns return `ErrGridRange`. Navigation does not materialize cell
strings. Single active cell is the entire selection scope of this draft.

Each materialized cell carries opaque `GridCellRef`. The renderer must capture
this ref when arming a press and pass it to `Activate` when delivering the action.
It must not reinterpret a pooled slot using the new viewport. Foreign refs,
retired generations, source changes, and refs with empty current clips are
refused. Effective viewport/active/source transitions retire prior refs; no-ops
and failed operations do not. Generation exhaustion is refused atomically.
Native hit testing, event-origin lifetime, focus, disabled/modal routing, and
native widget reuse still require an independent adapter qualification.

## Portable acceptance and next seams

Tests compare a brute geometry oracle against 800 seeded geometries, overscan
visibility against 300 seeded cases, overflow/refusal/source-failure boundaries,
indexed revision/anchor retention, detached snapshots and stale cell refs.
Backing-pointer regressions verify short and empty substrings from MiB caller
allocations do not remain in accepted state or snapshots; this is not a GC heap
measurement. Parallel tests use independent owner-confined Grids, not concurrent mutation of
one Grid.

A procedural indexed source exposes both 10,000 and 1,000,000 rows × 20 columns.
The same aligned 800×1,224 viewport has 50×8 = 400 visible cells and 520 total
records including overscan. Each snapshot makes exactly 52 RowKey, 52 IndexOf,
and 520 Cell calls in both datasets. A far jump makes two RowKey and two IndexOf
calls and zero Cell calls. This is a callback/geometry workload proof, not a
stored-dataset memory test, allocation benchmark or native performance result.

The native adapter can own its bounded measured-cell/header pools and validate
refs through this model. Variable-height indices, selection universes, editing
draft conflict/settlement, background projection index preparation, native
accessibility ranges, and full million-row native acceptance remain explicit
later work. Existing Table/Tree callers need no migration for this checkpoint.
