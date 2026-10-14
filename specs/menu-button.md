# Action menu button

`ui.MenuButton` is a leaf that opens a transient flat action menu. `Options`
contains 1..64 entries with unique nonempty UTF-8 IDs (at most 1024 bytes) and
nonblank action labels (at most 4096 UTF-8 bytes); NUL is refused. Duplicate labels are
allowed. `SelectedID` remains empty. `Change.SelectedID` requests the enabled
action using its stable ID; check actions also propose Change.Checked. The current Change callback runs on the application
owner. `Activate`, `Open`, `Dismiss` and child nodes are refused. A missing Change
handler or no enabled action options makes the button unavailable to managed
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

`Option.Checkable` adds a controlled boolean action. `Checked` is its accepted
state; ordinary actions require Checked=false. Choosing a check action closes
the popup and submits the inverse accepted state in Change.Checked, without
optimistic mutation. The callback may accept, reject or normalize it. Reopening
uses the current descriptor and displays a checkmark only for Checked=true.
Descriptor check-state changes revoke the old menu epoch and queued requests.

`Option.Separator` adds a passive native line. It has its own stable ID, empty
Label and false Disabled/Checkable/Checked fields. It counts toward the 64-entry
bound. Separators have no focus, action, accessible child or manual input target.
Keyboard navigation skips them and disabled actions. A menu containing only
separators/disabled actions remains readable but is unavailable to input/focus.
The separator/check flags are refused on all other kinds' Options.

Check actions publish AXMenuItem with an accepted boolean accessibilityValue
and Checked/Unchecked accessibilityValueDescription. Ordinary menu actions have
no boolean value/description. Inert production Cocoa selector tests verify that
mapping without native attachments/windows; human VoiceOver remains separate.
See [Apple's accessibility value contract](https://developer.apple.com/documentation/appkit/nsaccessibilityprotocol).

This control supplies a flat in-window action list. Nested menus,
radio action groups and command shortcuts are future API work. Human
VoiceOver, physical input, cross-display popup placement and broader Mac
qualification remain separate from automated Cocoa selector checks.
