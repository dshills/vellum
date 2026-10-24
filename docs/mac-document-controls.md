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

## Native Colors panel

Get `Application.ColorPanels()` and call `ChooseColor` on a worker with an exact
live parent WindowRef and ColorPanelOptions. The actual AppKit NSColorPanel offers
its normal pickers, optional opacity, and a small accessory with Use Color/Cancel.
Apply returns an sRGB straight-alpha `color.NRGBA`; extended-gamut values clamp to
sRGB bytes. Cancel/Escape/close returns ErrCanceled. Preview is optional, runs
serially on that calling worker, and coalesces to one latest value. Publish through
window-scoped posts, guarded by your session epoch; commit only the returned value
through a reserved completion. The gallery demonstrates that separation.

One native document panel is admitted per process. A Colors panel previously
created outside this service is refused with ErrServiceBusy because AppKit exposes
no target/action getters to restore a foreign owner safely. Do not mix direct
AppKit ownership of the shared panel with this service. Parent closure, context
cancellation and shutdown revoke the session, remove observers/child attachment,
hide the panel and restore its prior appearance/accessory state. Hidden/minimized
parents refuse without activating another Space. Other platforms return unavailable.
Offscreen tests compile the native path and cover request/reply parsing, exact
parent revocation and bounded preview delivery without creating any native panel.
Actual AppKit interaction is checked separately after the final package.
