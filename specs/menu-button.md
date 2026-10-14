# Action menu button

`ui.MenuButton` is a leaf that opens a transient flat action menu. `Options`
contains 1..64 actions with unique nonempty UTF-8 IDs (at most 1024 bytes) and
nonblank UTF-8 labels (at most 4096 bytes); NUL is refused. Duplicate labels are
allowed. `SelectedID` remains empty. `Change.SelectedID` requests the enabled
action using its stable ID; the current Change callback runs on the application
owner. `Activate`, `Open`, `Dismiss` and child nodes are refused. A missing Change
handler or entirely disabled options makes the button unavailable to managed
focus and action input. Disabled ancestors also prevent action delivery.

Text labels are literal; an icon-only button needs Accessibility.Name. The anchor
supports bundled Icon, button Appearance, local styles and themes. Private items
use native menu appearance with resolved button theme styles. Descriptor options,
anchor presentation and theme changes close an older menu. Application callbacks
remain current without rebuilding for function identity alone.

Pointer activation, Enter/Space and plain Down open the menu. Native menu arrows
move among enabled actions and Enter/Space chooses. Escape and outside dismissal
close the popup. Tab or Shift-Tab dismisses and resumes managed traversal after
or before the anchor. Each originating main scene has one managed action menu.
A separate window may have its own. Menus remain local to their originating
native window and event/render context.

Owner delivery rechecks menu epoch, lifetime, disabled state and enabled action
ID. Replacing, dismissing or removing the menu revokes obsolete handlers and
queued requests. Queued work only revokes logical state; native scene destruction
runs during originating reconciliation/teardown. Popup scenes forward custom
wakes to the originating driver, so owner work progresses while the menu is
modal. One pending popup retirement bounds retained native references. Refused
owner delivery follows the existing bounded control-error reporting contract.

The Cocoa anchor exposes AXPopUpButton, expanded state and expand/press actions;
an open menu adds collapse/cancel and exposes AXMenuItem children with stable
option DomainIDs. Enabled items support focus/press; disabled items remain
readable without action admission. The open menu scopes out background targets.
Frames come from actual popup item geometry. Dismissal retires item objects;
reopening creates fresh object lifetimes. The generated `menu:` prefix is
reserved by descriptor validation. The manual backend exposes private targets
with per-popup incarnations and rejects old target incarnations after reopening.

This control supplies a flat in-window action list. Nested menus, separators,
checkbox/radio actions and command shortcuts are future API work. Human
VoiceOver, physical input, cross-display popup placement and broader Mac
qualification remain separate from automated Cocoa selector checks.
