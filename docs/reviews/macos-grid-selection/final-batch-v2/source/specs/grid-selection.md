# Captured grid row selection

`GridConfig.RowSelection` opts into row membership; false preserves the existing
active-cell behavior and does not call `SelectionSnapshot`. Enabled sources must
implement `GridSelectionSource.SelectionSnapshot() (GridRowIndex,error)`.
`GridRowIndex` supplies the same identity, revision, count, indexed RowKey and
inverse IndexOf methods already required by GridSource, without Cell access.

The snapshot must exactly describe the current source stamp and order. It must
remain immutable forever and permit concurrent worker reads. Capture is bounded,
nonblocking and performs no IO or full-source scan. Prepare loaded indexes before
installation, normally on workers. Returning a mutable owner source or retaining
its mutable diagnostic counters does not meet the contract. Procedural snapshots
may retain only detached incarnation and scalar projection parameters. Loaded
snapshots may borrow immutable domain/index storage; that O(N) storage and its
lifetime are explicit source costs outside Grid's payload budget.

## Membership and identity

Grid retains one half-open interval in a captured index plus a bounded immutable
map of stable-key boolean exceptions. Replace captures one row. Extend captures
the current interval between the stable anchor and target, inclusive. A hidden
anchor starts a new range at the target; old ordinals are never reused. With no
selection anchor, Extend uses the active cursor when still present. Toggle
changes the target membership and active cursor while retaining the earlier
range anchor. The first Toggle starts a single-row selection and anchor.
`SelectAll` captures the current whole view with the active cursor as anchor,
or the viewport's first row when no cursor exists. It creates no object per row.

Same-incarnation sorting, filtering and height revisions preserve captured
identities and anchor, independent of the new row ordinals. Toggle can add a key
from a later view that was absent from the captured index. A new incarnation
clears membership and anchor. Hidden or removed keys remain captured: a view
index cannot distinguish domain deletion from filtering. Domain actions must
validate current live identities before applying worker results.

`GridSelectionState` contains a monotonic Revision, CapturedRows, AnchorKey,
base Scope stamp and Overrides count. CapturedRows counts the captured identities
after exceptions, not the current visible or filtered row count. Scope describes
the base captured order; additions can come from later same-incarnation views.
`GridState.Selection` and `GridCell.Selected` expose accepted membership while
`Active` remains a separate cell cursor, including when that row is toggled out.

`SelectToKey`, `SelectCell`, `MoveSelection` and `PageSelection` accept
`GridSelectReplace`, `GridSelectExtend` or `GridSelectToggle`. Legacy JumpToKey,
Activate, Move and Page use Replace in enabled mode. ClearSelection clears
membership without changing the cursor or viewport. Effective selection/cursor/
viewport transitions advance Grid generation once; semantic selection changes
also advance Selection.Revision. Source replacement always retires references.
No-op and failed operations preserve generations, selection and geometry.
Both clocks refuse exhaustion atomically.

## Bounds and validation

MaxSelectionKeys defaults to 4,096, accepts 1 through 65,536, and bounds sparse
exceptions. Zero selects the default. Grid separately charges captured incarnation,
anchor and unique exception-key payload against MaxTextBytes, excluding borrowed
index memory and allocator overhead. Retained strings are cloned; no small key
substring pins a caller backing allocation through Grid's own metadata/map.
Toggle copies at most this bounded map. Count arithmetic is checked for overflow.

Capture checks indexed range endpoints and validates membership for each admitted
visible/overscan row before committing. When its scope matches the current view,
the captured row index must agree with that row's current index, not merely contain
the key. Later revisions may legitimately contain new keys outside the capture.
Capacity refusal precedes this bounded validation. Snapshot evaluates membership
once per materialized row, independently of the number of columns. No selection
owner transition enumerates all captured rows. Hidden spans remain a source
contract responsibility, rather than a whole-index certification claim.

Source/capture/query errors retain their cause through ErrGridSource. Metadata is
checked after reads, including a final live-source stamp after captured-index
checks. Refusals publish no partial model or geometry. Mutation of a borrowed
source itself cannot be rolled back. ClearSelection deliberately reads no source
or captured index, allowing explicit owner recovery after an invalid capture;
stale live geometry still requires Refresh. Native input requires valid current
state and cannot perform that recovery until the owner clears it explicitly.

## Detached worker use

`Selection()` returns a copyable immutable GridSelection in O(1), sharing only its
private immutable map and immutable source index. It has no Grid/application
backpointer. Future owner transitions never mutate its map. State() reads detached
metadata; Contains(key) performs checked indexed membership reads. These are
safe concurrently with owner transitions provided the source honors immutability.

`VisitContext(ctx, callback)` is an explicit worker operation costing O(interval
length + exceptions), even if most interval keys are excluded. It delivers detached
keys in captured row order, followed by sparse additions in lexical key order,
without duplicates. Cancellation is checked between callbacks and before return;
callbacks cannot be forcibly cancelled. Errors or cancellation can follow partial
callback delivery. Accumulate privately and publish only after a nil result.
Do not enumerate a million-row capture on the UI owner or apply a stale worker
result without domain revision/identity validation.

The [managed adapter](grid-native.md) specifies Mac modifiers, stale gesture
retirement, focus/modal precedence, measured selection and bounded visible AX
attributes. This checkpoint does not add cell editing, loaded worker projection
integration, hidden-cell AX navigation or complete P6 performance acceptance.
