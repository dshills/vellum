# Standard controls contract

The public `ui.Node` vocabulary now includes checkbox, selector, slider, numeric
field, single-line text field, tabs, scroll container and modal dialog. Cogent
adapts these to its existing native controls. The portable module remains free
of native dependencies.

## Controlled values and events

`Checked`, `Number`, `Text` and `SelectedID` describe accepted model values.
`Change func(ui.ChangeEvent)` receives the relevant typed payload plus stable
node ID and kind. A callback changes application-owned model state; the next
view describes the accepted value. Programmatic setters do not emit changes.
`Activate` continues to serve buttons. No callback is bound to an obsolete view
closure: retained controls read the current callback when owner delivery runs.

Native changes, tab choices and dialog dismissal enqueue bounded application
owner work. Values are captured when the native event occurs; mounted state,
disabled state, available options and current callbacks are rechecked on drain.
Removed widgets do not deliver queued stale edits. Queued callbacks use owner
snapshots for disabled/retired state; they only update model state and request
restoration. Native value restoration runs in the originating window's next
refresh, even when another window drains the shared application queue. Refused value/tab delivery atomically requests restoration during the next
originating native refresh under the UI owner. That refresh restores the
controlled native value/selection and reports the error through `Driver.OnError` when
provided, also storing it in native property `vellum.control-error`. Error
observers must not reenter application ownership. Changes emitted synchronously
by native focus transitions therefore cannot nest `Application.Do`.

`Disabled` blocks model changes, including descendants of disabled managed
containers. Native styles reflect that disabled state. Selector/tabs options use
stable IDs independent of displayed labels; duplicate/empty IDs and missing
selected IDs are rejected. Native selector menus omit disabled choices; a
model-selected disabled choice can still be displayed. Tabs display disabled
buttons but reject their actions. Selector empty selection uses its `Text` as a
placeholder; tabs require one selected page and one option per child page ID,
in the same order.

## Numeric and text editing

`Numeric{Min, Max, Step}` supplies finite ordered float32 bounds, a positive
keyboard step no larger than the range, and a finite accepted `Number` inside
the bounds. Range spans that overflow native float32 arithmetic are rejected.
Finite native events are clamped to bounds; nonfinite values are refused.
Keyboard increments follow Step; pointer/typed values are not rounded to step
multiples. Precision/display rounding remains the native backend's behavior.

Text fields use the existing native editing/input foundation. Unrelated view
updates preserve uncommitted text, composition, selection and numeric-field
editing. A changed model text/value replaces pending edits. After a committed Change
callback, the next redraw reapplies the accepted model value even if it stayed
unchanged, so rejected/normalized commits cannot leave stale native text/value. Adapter reads
use the read-only native editing snapshot, avoiding the native `Text()` getter's
implicit commit and change event. Display labels, option labels, tab labels and
help tooltips are escaped before entering Cogent's markup-capable text widgets;
editable text remains literal text.

## Containers, focus and semantic metadata

Tabs retain native page hosts by stable ID. Reordering or removing pages preserves
compatible managed widgets; moving a widget from a removed page happens before
its old native host is destroyed. The reconciler never replaces Cogent Tabs'
internal tab/frame structure. Scroll containers use native overflow and retain
their native state.

A root `FocusRequest{ID, Sequence}` requests keyboard focus by managed node ID.
Changing Sequence repeats a request for the same ID. Unchanged requests are
consumed once and do not steal focus on later redraws, even if the old target is
subsequently hidden/disabled/removed. A new request must name an enabled focusable
node in the visible tree (selected tab and open dialogs). Before the native scene
is shown the driver installs start focus; afterward it sets native focus. Child
nodes cannot carry independent focus requests. A Tabs request targets its selected
enabled native tab button, or the first enabled option when the selected option
is disabled. A Tabs control with no enabled options cannot receive focus.

`Accessibility{Name, Help}` carries a semantic label independently of visible
text. The native adapter stores typed `cogent.Semantics` under
`cogent.SemanticsProperty`, including stable ID, kind/role, name, help, disabled
state and controlled value. Help also appears as a native tooltip. Platform
screen-reader publication and acceptance remain explicit foundation gates; this
property seam does not claim an operating-system accessibility bridge.

## Modal lifetime

`Dialog` is a descriptor with `Open`, title in `Text`, `Children` and `Dismiss`.
A hidden managed placeholder owns one stable child host. Opening moves that host
into a native modal Body/Stage. Closing parks it before Cogent destroys the old
scene, preserving retained content across reopening. Programmatic close/removal
suppresses Dismiss; native Escape/Close posts Dismiss once. If the owner mailbox refuses dismissal,
the modal remains closed, reports `Driver.OnError`, and preserves content; it
cannot restore a native scene already closed by the user. The application can
acknowledge `Open=false` and then request `Open=true` to reopen. There is no
unbounded retry queue. The application should
set `Open=false` in Dismiss. Until that acknowledgment, the driver does not reopen
a user-dismissed dialog on each refresh. Removing a dialog closes its external
stage before removing the placeholder. Active modal scenes forward native custom
wake events so worker delivery and programmatic close continue to function.

## Executable acceptance and tests

Run `go run ./cmd/gallery` from `drivers/cogent` after dependency preparation.
It demonstrates every control, shared slider/numeric state, a checkbox disabling
other controls, named focus requests, local theme values, tabs and a native modal.
`-smoke -exit-after=4s` opens and closes a modal through posted owner callbacks;
window-render diagnostics confirm the modal reached the native stage pipeline.

Portable tests cover finite constraints, option/page identity and focus visibility.
Native tests cover every change payload, latest callbacks, disabled ancestors and
options, mailbox refusal, removed queued callbacks, NaN rejection, text editing
across unrelated redraws, tab reorder/reparent, modal scene cleanup preserving
content, custom wake forwarding and repeated focus sequence requests. The gallery
model test checks the complete controlled workflow. Native visual/input behavior
still requires platform acceptance beyond these adapter checks.
