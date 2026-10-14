# Controlled breadcrumb navigation

`ui.Breadcrumb` is an owner-held reusable path component. `View() (ui.Node, error)`
returns ordinary managed toolbar, quiet-button, label and decorative-icon nodes.
It requires a nonempty root ID, 1..64 uniquely keyed items, nonblank literal labels
and a valid style. Item IDs are stable domain identity; display labels and indexes
never identify a location. The final item is the current location and is a passive
label with “Current location” help. Ancestors request `BreadcrumbEvent{ID, ItemID}`.
The application accepts by replacing the path or rejects by leaving it unchanged.

Nil Navigate produces only read-only location labels. Disabled component/items
refuse navigation. Queued callbacks use the latest owner callback and check the
current root, complete path identity, current validity and target enabled state;
a changed path cannot apply an old request to new content. Descriptions and captured
ID slices are detached from the caller's Items slice. A component instance is used
by one application owner and is never mutated/rendered concurrently.

`BreadcrumbItemID(root, item)` uses base64url domain IDs under the owned
`root/breadcrumb/` namespace. Ordinary Node validation detects surrounding caller
collisions. Compatible ancestor buttons retain their managed identity on redraw;
an ancestor becoming current changes to a label and retires the prior button.
Ancestor buttons use ordinary Tab/Shift-Tab and Enter/Space activation, originating
window/owner lifetime checks and accessible press actions. The current label and
separator artwork add no focus stop. Decorative icons are omitted from accessibility.
Toolbar overflow and focus revelation handle long paths when the parent allocates
a width (for example a growing root column or an explicit size), as for ordinary
toolbars. A content-sized parent cannot infer the width of scrollable content.
Automatic ellipsis,
filesystem lookup and background navigation are application concerns.

The gallery must demonstrate accepted/rejected navigation, reset and disabling.
Tests must exercise actual managed pointer/key input, rejected identity retention,
accepted current-label retirement, literal labels, disabled/read-only/stale refusal,
independent namespace collisions, completed frames at four DPI settings and inert
Cocoa role/owner-queue behavior. These checks are separate from physical input,
human VoiceOver, sustained P6/P11, native/GPU memory and full Mac qualification.
