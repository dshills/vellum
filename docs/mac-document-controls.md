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
