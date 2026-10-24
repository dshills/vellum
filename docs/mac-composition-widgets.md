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
