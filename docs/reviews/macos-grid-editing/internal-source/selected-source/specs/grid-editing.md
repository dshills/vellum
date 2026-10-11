# Versioned grid cell drafts

The optional single-cell editing model supports a recoverable bounded draft and
atomic immutable-source admission. The managed Mac adapter uses one native text
field in the edited cell. This checkpoint does not complete P6 performance,
loaded-data workers, tree acceptance, external AX or human IME acceptance.

## Source and ownership

Set `GridConfig.CellEditing` and supply a `GridEditSource`, which embeds
`GridSource` and adds `EditCell(rowKey,columnID) (GridEditValue,error)`.
`GridEditValue` contains `Text`, positive `Version`, and `Editable`. Reads are
indexed, bounded, nonblocking and owner-confined, without IO or Grid reentry.
Text agrees exactly with `Cell`. Source values remain immutable within a source
revision. Cell versions increase on every value or editability change, including
A→B→A; unrelated edits, sorting, filtering and geometry changes preserve them.
Stable keys cannot be reused within an incarnation. Providers must retain that
version history across temporary filtering or removal. Grid validates accessed
cells, rather than certifying the entire source.

`MaxEditBytes` defaults to 64 KiB and accepts 1 through 1 MiB. Each accessed edit
value and each retained draft string stays within that bound. This separate
payload bound does not cover domain allocations, all detached caller snapshots,
allocator/native renderer overhead or the entire UI heap. Accepted retained
strings are cloned, avoiding large substring backing allocations. Drafts accept
valid UTF-8 single-line text, including exact combining sequences; NUL, CR and LF
are refused. Ordinary grid display strings may remain multiline.

## Draft and conflict policy

`BeginEdit(ref)` verifies a current opaque cell reference. `BeginActiveEdit()`
resolves the active identity on the owner for commands. Both refuse a second
draft, a noneditable cell, malformed source or oversized baseline. One draft
retains stable row/column identity, original text, base cell version and current
source status. `GridEditState.Session` increases on each successful begin and
never wraps; native clients use it to distinguish cancel/restart of the same
cell. `Revision` increases on effective draft text changes, including draft ABA.

`SetEditText` changes the draft, never the source. `Edit()` returns a detached
state; `GridState.Edit` reports the last admitted state. Failed/no-op operations
preserve clocks and refs. Effective begin/text/cancel transitions retire pointer
refs. Generation, session and draft revision exhaustion refuse atomically.

Source replacement updates status without destroying the draft. Sorting,
filtering and row-height changes preserve it. A hidden or removed key becomes
`unavailable` until it returns; saving that target is refused. A higher edited
cell version becomes `conflict`, even if text returns to its original value.
Permission loss becomes `read-only`. A changed source identity becomes
`incarnation`; no draft can commit into that different domain. Original and draft
text remain recoverable. Cancel/restart is the explicit resolution policy; there
is no implicit overwrite, merge or force-commit operation.

`Edit`, `SetEditText` and `CancelEdit` validate the Grid identity without requiring
current source metadata, preserving recovery even before a changed source is
acknowledged. Their status reflects the last accepted source. Other Grid reads,
prepare and completion continue to refuse unacknowledged metadata drift.

## Validation and atomic completion

The application validates the draft before `PrepareEdit`. Validation errors stay
application state, preserving the draft and native editor. Grid does not invoke
validation or domain mutation callbacks. `PrepareEdit()` verifies that the current
edited cell still has the original text and version and returns an opaque
`GridEditRequest`. Its `State()` and `Stamp()` are detached values. A request has
no Grid, owner or source backpointer. Changed draft revision, cancel/restart or
accepted source replacement retires it; prepare again after those transitions.
Viewport/cursor changes alone do not change the prepared edit proposal.

The application builds an immutable candidate from its current domain view
without changing the live domain. `CompleteEdit(request,candidate)` requires the
same incarnation, a higher source revision, the edited row present, exact draft
text and a higher cell version. It validates candidate geometry, row/selection
admission and capacities, then rechecks the live request/cell and candidate value
and source stamps before publishing the source and closing the draft in one
generation. Any refusal leaves the Grid source, draft and generation unchanged.
The application installs its domain/source pointer only after success. Grid
cannot roll back a caller that mutates its live domain prematurely.

`ErrGridEditBusy`, `ErrGridEditReadOnly` and `ErrGridEditConflict` distinguish
session, permission and conflict refusal; existing stale/source/capacity/range/
exhaustion errors preserve their usual meaning. No-draft prepare returns range.

The Mac input and settlement contract is in [grid-native.md](grid-native.md).
Return intent invokes `Node.GridCommit` only after originating native input has
settled. Toolbar actions use `SettleGridEdit` before validation/prepare. Focus
loss and Tab preserve drafts; saving remains explicit.
