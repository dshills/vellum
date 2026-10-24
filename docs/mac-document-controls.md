# Mac document controls

## Document list

Retain `ui.DocumentList` with up to 256 `DocumentItem` values. Metadata and
immutable original-color images come from the application; no filesystem scan,
preview decoding or file launch happens inside this widget. Columns are name,
kind, modified date and byte size. Use `DefaultDocumentColumns()` for independent
mutable widths (60..600 points). Sort proposals update Sort; folder rows sort
first and text comparison is case-insensitive, stable and not locale-aware.
`Resize` proposes widths through pointer dragging or Left/Right on header grips
(Shift adjusts ten points). The shared horizontal scroller keeps header/row
columns together. Vertical scrolling keeps the header visible.

Selection uses the existing controlled source-list Command/Shift/Command-A
routes; accept with `event.Apply(list.SelectedIDs)`. Return/double-click requests
Activate separately. Disabled/removed items and stale component sources cannot
act. Size is shown in bytes, dates in English. Context menus and automatic
filesystem operations are outside this component. Components → Document list
demonstrates all operations using in-memory sample metadata.

## Mac outline

`ui.OutlineView` composes the existing bounded `virtual.Tree` index with managed
source-list selection. Provide <=256 flat parent-linked `OutlineItem` values,
depth <=16, stable IDs, original-color images, optional badges and disabled items.
Larger indexed datasets retain the existing worker-built `DataTree` route.
Accept Disclosure with `event.Apply(ExpandedIDs)` and selection with
`event.Apply(SelectedIDs)`. Collapsed descendants retain application selection;
only visible selected rows are published to the native list. Disabled ancestors
disable their descendants. Right opens/enters, Left closes/returns to the parent,
Up/Down moves rows and Return/double-click opens a leaf or toggles a branch.
Disclosure also has accessible expand/collapse actions. Optional ContextMenu
reuses existing source-list contextual actions. There is no item dragging,
filesystem traversal, lazy child loading or built-in persistence.
