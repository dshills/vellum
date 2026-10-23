# Mac desktop widgets — personal-use expansion

Owner approved the next five widgets: editable combo box, context menu, anchored
popover, multi-selection source list and date/time picker. This supersedes the
older internal-app API freeze in CONTRIBUTING.md for these five capabilities only.
Rendering, pacing and released alpha.3 artifacts remain unchanged. Each checkpoint
gets focused tests, Prism Anthropic claude-sonnet-5-5 review and a commit.

## 1. Editable combo box

ui.ComboBox composes the retained TextField and managed MenuButton; text uses the
existing native editing bridge. Typed values and suggestions propose ComboEvent
through one owner callback, never auto-add options. Stable option IDs are separate
from labels. Current disabled/removed options and changed component IDs refuse
old proposals. Changed labels also revoke callbacks from the older rendered option. At most 64 suggestions, with no background work. Empty options
leave a text-only field. Mac appearance uses one compact outline around the row.
Tab reaches the suggestions affordance; Down/Return opens the ordinary managed
menu, arrows choose and Escape dismisses. Down in the text entry itself does not
open suggestions in this initial composite. No live filtering or completion.

All controls are custom-rendered, not AppKit wrappers. Fixed MacTheme colors and
Helvetica Neue are retained; automatic system appearance/accent tracking and
pixel-identical AppKit claims are outside scope.

## 2. Context menus

Node.ContextMenu supports Button, Canvas2D and ListBox, including list rows. Flat
actions, checkmarks and separators reuse MenuButton contracts and bounds. Menus
open on right/Control-click or Shift-F10; the host also exposes AX show-menu.
Context actions do not activate the host. Pointer menus anchor at the event
position; keyboard menus anchor at the host. Escape/Tab/outside release dismiss
using the existing modal popup lifecycle. Removal, disable, option changes and
window retirement revoke old actions. Nil Change leaves normal host use enabled.
No submenus or shortcut-label metadata in this increment. Text fields retain
their native editing context menus.

## 3. Anchored popover

ui.Popover returns a Dialog descriptor with PopoverAnchor. Reuses retained
children, native editing, focus scope, Escape and outside dismissal. Its compact
modal surface has no title, scrim, resize affordance or appended Close bar. It
anchors below a visible enabled same-window managed Button, clamped inside the
window, and returns focus there after closing. Anchor removal/hiding/disabling
closes and proposes Dismiss. Dismiss must set Open false. A small directional
marker leads content; automatic edge switching, detachable panels and cross-window
anchors are not supported. At cramped edges the surface clamps rather than
reversing the marker. Existing Dialog lifecycle owns parking/cleanup.
