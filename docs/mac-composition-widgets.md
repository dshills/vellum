# Cascading action menus

Create a stable `ui.NewMenuBranch([]ui.Option{...})` and put its pointer in an
option's `Branch`. Retain that branch across Views; replacement retires an open
chain. `Items()` returns a copy. IDs are unique across the whole tree, limited to
64 items and four popup levels. Leaves can have immutable full-color `IconImage`,
checks, disabled state and a `Shortcut` caption. Captions do not install bindings.
Use the application's command registry for actual shortcuts.

Branches open with the pointer, Return or Right; Left closes the submenu.
Escape/Tab dismiss the chain. Only enabled leaves request `Change`, once, using
the current callback and check state. Disabled ancestors, removed branches and
retired popups revoke actions. The opened hierarchy has accessibility actions
and actual popup bounds. These remain managed popups, not an AppKit menu bar.

The Mac Widgets **Components → Cascading menus** page demonstrates export
actions with full-color artwork, a disabled leaf and a controlled check action.
Focused offscreen tests cover real key routing, exactly-once dispatch, Left,
Escape and stale accessibility handlers; physical pointer checks are recorded
separately in the packaged delivery.

## Split buttons

Retain a `ui.SplitButton` and call `View()`. Set `Primary`, `Options` and `Change`
independently; `Small` selects the compact size. `IconImage` preserves full-color
artwork. The disclosure anchors to the combined control and accepts cascading
options. Tab reaches each segment; Return activates the focused segment and
Down opens its menu. Disabled or removed options revoke old action closures.
The Components → Split buttons page demonstrates regular and compact variants.

## Inspector panel

Retain `ui.Inspector`, pass `InspectorSection` and `InspectorProperty` values and
call `View()`. Editors keep their own IDs, values and callbacks. Accept disclosure
proposals by updating `Expanded`; sections retain child controls when collapsed.
`Mixed` adds a hint beside the editor without replacing its accepted value.
Set `LabelWidth` (48–320dp) to align rows. Header artwork uses immutable pixels.
At most 32 sections/128 properties are accepted. No document-model logic is built
in. The Components → Inspector panel page shows geometry and appearance sections.

## Column browser

Retain `ui.ColumnBrowser` with a bounded `BrowserItem` hierarchy (256 items, four
levels). Accept `BrowserEvent.Path` in `Change`. Disabled ancestry and stale
source descriptions reject callbacks. `Thumbnail` retains full-color immutable
pixels; `Icon` is the symbol alternative. `Preview` is application-owned content.
Columns reuse multi-list keyboard routing but propose exactly one path target.
Up/Down/Home/End/typeahead navigate rows; Tab traverses columns and dividers.
This version does not add Finder's Left/Right column shortcuts or filesystem I/O.

Nested split dividers resize columns. Accept `BrowserResizeEvent.Fraction` into
`Fractions[event.ColumnID]`; bound the map to the visible columns (the example
resets it when the path changes). Horizontal scrolling preserves usable minimum
column widths. Components → Column browser demonstrates hierarchy and previews.
