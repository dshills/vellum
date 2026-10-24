# Standard controls contract

The public `ui.Node` vocabulary now includes checkbox, switch, toggle button,
progress bar, circular progress ring, semicircular progress gauge, separator, action menu button, radio group, segmented control, selector, list box, slider, numeric
field, single-line text field, multiline document editor, managed data grid,
tabs, scroll container, toolbar, status bar, labeled form group and modal dialog. Cogent
adapts these to its existing native controls. The portable module remains free
of native dependencies.

The [list box](list-box.md) adds bounded always-visible single selection with
scrolling, clamped keyboard/page navigation and distinct Mac list/row semantics.

Managed node IDs beginning with `grid:` are reserved for generated grid rows,
cells and headers. Descriptor validation rejects that prefix before mounting controls,
so ordinary controls cannot alias a generated input or accessibility target.

## Controlled values and events

`Checked`, `Number`, `Text` and `SelectedID` describe accepted model values.
`Change func(ui.ChangeEvent)` receives the relevant typed payload plus stable
node ID and kind. A callback changes application-owned model state; the next
view describes the accepted value. Programmatic setters do not emit changes.
`Activate` continues to serve buttons. No callback is bound to an obsolete view
closure: retained controls read the current callback when owner delivery runs.

`DataGrid` binds `Node.Grid` to the portable indexed model. Its `GridChange`
callback reports an already accepted owner-held selection or viewport change;
the caller does not apply it again. See the [managed grid contract](grid-native.md)
for native bounds, input, focus and measured geometry.

Standard value changes, tab choices and dialog dismissal enqueue bounded
application owner work. Managed native text input uses a separately bounded
originating-window FIFO and legacy editor event boundaries before owner changes. Values are captured when the native event occurs; mounted state,
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
containers. Native styles reflect that disabled state. Reconciliation refreshes
retained descendants when effective disabled state changes, even when their own
descriptors are unchanged. Re-enabling an ancestor preserves a child's own
Disabled state. Text/numeric multi-clicks restore actual focus when eligible
before ordinary word/all selection; they do not propose a value just to focus. Selector/tabs options use
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
editing. Active native preedit also survives the backend Maker refresh while
the controlled value remains uncommitted. A changed model text/value replaces pending edits. After a committed Change
callback, the next redraw reapplies the accepted model value even if it stayed
unchanged, so rejected/normalized commits cannot leave stale native text/value. Adapter reads
use the read-only native editing snapshot, avoiding the native `Text()` getter's
implicit commit and change event. Display labels, option labels, tab labels and
help tooltips are escaped before entering Cogent's markup-capable text widgets;
editable text remains literal text.

A managed Mac text field finishes a focus-loss `Change` only after the outgoing
native client acknowledges all already-admitted input and becomes inactive.
Native unmark may add another FIFO record; the adapter waits for its owner
acknowledgment before finishing the field or activating the next native client.
This handoff is asynchronous and does not block the UI owner. A refocused field
retains its fresh directional selection when a late, same-visible native unmark
seals the outgoing composition. Retired fields suppress completion callbacks;
fields without an active native client retain the backend's immediate focus-loss
behavior. This contract covers admitted native field input through handoff;
programmatic focus requests do not impose a global order on legacy key events.

`cogent.SettleTextField(ctx, application, windowID, nodeID, callback)` admits an
asynchronous barrier for an ordinary managed text field on the application
owner. A dependent action, such as applying a filter, reads its controlled model
value inside the successful callback. The originating-window FIFO processes
already admitted input, seals composition and delivers the final `Change` before
that callback. It retires older queued change snapshots for that field, so they
cannot overwrite the settled value. Repeated settlement without a value change
does not repeat `Change`, including before the next redraw. Reconciliation still
applies the application's accepted, rejected or normalized controlled value.

The barrier shares the existing one-pending-settlement capacity per window with
document and inline-grid settlement. Missing, ambiguous, disabled, retired,
nonordinary or callback-free field bindings are refused; an immediate refusal
has no callback.
Cancellation and binding/window retirement terminate admitted requests while the
application remains open. Application shutdown suppresses outstanding callbacks.
The callback and `Change` run on the owner without nested owner draining. Manual
drivers require explicit `Step`; settlement starts no worker or ambient frame.
This is an ordering contract for admitted native input, rather than a guarantee
about future input or input from a different window.

Native text-cursor sprite ownership is local to the main scene for both manual
and desktop execution. Each stage's sprite lock guards its own ownership map, so
concurrent window closure does not write a shared last-widget pointer. This does
not impose a global order on window teardown or native event delivery.

The Mac native adapter refuses a single-line `TextField.Text` above 4 MiB before
mounting; the portable descriptor has no such platform bound. Cocoa input
preflights that native budget before mutation. A legacy field edit above the
budget is refused as a whole pending value and restored to the accepted
controlled value without `Change`; pending field edits and private editing state
are reset. The document editor's stronger model/history atomic refusal contract
is specified separately below.

`MultilineText` binds `Node.Editor` to the portable document model. It updates
that model on the application owner and reports committed bytes through `Change`.
Preedit remains outside committed history. `TabInserts` chooses literal tab input;
otherwise Tab participates in managed focus traversal. `SoftWrap` changes visual
wrapping without changing document bytes. Editor input uses its
originating native FIFO and explicit persistence settlement; see the
[document editor contract](editor.md) for bounds, position units, history and
lifetime rules.

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

For accepted main-scene requests, the Mac adapter keeps one pending visibility
intent per native window. After layout and before child painting it reveals the
still-focused native target using existing scrolling. This includes a repeated
request for a retained field whose position changed during reconciliation. It
does not change focus or editing selection, replay on later manual scrolling,
or re-use an intent for a replaced/hidden/disabled target. Empty/failed requests
and native retirement clear it. Dialog-scene focus retains its existing policy.
This uses normal paint invalidation, without new frame scheduling or renderer
options.

Focused tab headers handle unmodified Left/Right with enabled-option wrap and
Home/End with enabled endpoints; Space/Enter proposes the focused option.
Keyboard movement changes native header focus and queues one stable-ID proposal
through the latest owner callback. Rejected selection stays accepted as before.
Read-only headers can move focus without proposing selection. Activating an
already accepted focused option is a no-op; existing explicit pointer repeats
remain proposals. Modified/previously handled chords keep command priority.
Queued movement retains absolute input-time intent even before a redraw.
Activation compares with the latest live pending intent, so focus-only traversal
back to a still-rendered selected header cannot lose an explicit activation.
One immutable intent per strip survives owner delivery until authoritative
originating reconciliation. Admission refusal restores only the exact previous
intent; rejection, lifecycle changes and reconciliation retire it. Input-time
read-only authority cannot revive when a callback is later enabled.
If the actually focused header becomes ineligible, originating reconciliation
moves focus to the selected/first enabled header without proposing selection.
If no header is eligible, traversal skips the complete strip subtree (including
Parts and child descriptors awaiting sync), focusing the next enabled control or
clearing focus when there is none. Existing valid focus is preserved.
Hiding/disabling/removing a strip, dialog scope closure and nil/non-nil Change
transitions revoke old selection capabilities. Owner callbacks never traverse
originating native widget trees or change their focus from another window.


`Accessibility{Name, Help}` carries a semantic label independently of visible
text. The native adapter stores typed `cogent.Semantics` under
`cogent.SemanticsProperty`, including stable ID, kind/role, name, help, disabled
state and controlled value. Help also appears as a native tooltip. Platform
screen-reader acceptance remains an explicit foundation gate. The macOS bridge
publishes these semantics through NSAccessibility; see the
[accessibility contract](accessibility.md) for its verified scope.

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

The [Mac control acceptance map](control-acceptance.md) names scenarios for all
34 public node kinds and the Breadcrumb/ColorPicker composites. The focused
Selector/NumberField keyboard regression checks disabled-option skipping, numeric
step, single owner delivery, rejected native-value restoration and disabled
refusal through the offscreen event route. Physical desktop acceptance remains
a separate gate.

## Expanded desktop elements

`Disclosure` is a labeled collapsible container for inspector/settings sections.
Its nonblank `Text` is the heading; `Checked` is accepted expansion state and
`Change` requests `ChangeEvent.Checked` with the section's ID/kind. `Activate`
and Orientation are refused. Empty content is allowed. Programmatic synchronization
does not issue Change. The private heading/body and keyed children are retained
across collapse, expansion and label updates; removal still retires them.

The heading is one managed focus target. Unmodified Left collapses and Right
expands; Enter/Space and pointer activation toggle. Read-only sections with no
Change callback remain focusable. Collapsed descendants do not join Tab order,
modal discovery, grid/tree input scopes or native accessibility. Queued child
actions are revoked when their accepted ancestor collapses. A programmatic
collapse of focused content moves focus to its enabled heading. Disabled ancestry
refuses heading and child input. Queued heading requests use the current callback
and are revoked after a later accepted state, disable or retirement.
Manual inspection also exposes its separately framed heading target. Its
`disclosure:` namespace is reserved so domain nodes cannot alias routed targets.

The native accessibility hierarchy is an AXGroup with a separately framed
AXDisclosureTriangle heading and, when expanded, its child hierarchy. Heading
identity survives expansion changes. It exposes Expanded plus focus/press/expand/
collapse actions as allowed by controlled state. Collapsed child semantic objects
are retired and may acquire new accessibility identities when restored; native
widget identity and private editing state remain retained. Human VoiceOver and
physical Mac input remain separate from deterministic routing and Cocoa selectors.

`Switch` and `ToggleButton` are controlled boolean leaves. `Checked` is the
accepted value and `Change` receives the requested Checked value with stable ID
and kind. Switch displays an on/off switch; ToggleButton displays a checkable
button. They share bounded owner delivery, latest callbacks, ancestor-disabled
refusal, accepted-value restoration, retained identity and focus traversal with
Checkbox. Rejected native boolean changes restyle their accepted value even if
the descriptor stays unchanged, preventing stale selected toggle backgrounds.
Programmatic synchronization emits no Change.

`ProgressBar` displays a finite accepted `Number` fraction in 0..1 without a
keyboard step or input callbacks. `Separator` is a decorative line. Both accept
`Orientation` Horizontal/Vertical; empty defaults to horizontal. Kinds outside these display, choice and toolbar controls refuse Orientation. Explicit managed size/grow styles override native final
defaults; removing those styles restores the native dimensions and growth. They are leaves and do not join managed focus order.
Progress exposes its fraction and 0/1 bounds with the Mac progress-indicator role;
Separator is omitted from the accessible hierarchy. Switch and ToggleButton
publish checkable Mac roles and boolean press/set actions. Human VoiceOver and
physical keyboard/pointer acceptance remain separate from offscreen/native-object
fixtures. No indeterminate animation is implied by these display elements.

The gallery demonstrates all four additions using ordinary descriptors and
accepted application state. Further vocabulary work prioritizes richer reusable
command/button presentation, with keyboard, disabled, lifetime and accessibility
contracts.

## Radio and segmented choices

`RadioGroup` and `SegmentedControl` are controlled leaves with 1..64 Options,
stable nonempty unique option IDs, and one required SelectedID from those options.
Displayed labels can repeat. An accepted disabled choice remains displayed; it
cannot be selected by input. A disabled group/ancestor refuses changes. Empty,
missing-selected and oversized descriptors are refused before retained mutation.
`ui.NextChoiceOption` wraps enabled IDs for ±1; missing current IDs choose the
first/last enabled value. Invalid/disabled/all-disabled descriptors have no target.

Radio groups default vertical and segmented controls horizontal; Orientation can
set either explicitly. Option widgets retain identity by domain ID through label
updates/reordering. Removal retires the private binding; readding the same ID
creates a new binding. Native input captures option ID and binding identity, then
rechecks current enabled state and callback on owner delivery. Retired queued
requests cannot select a new same-ID binding. Programmatic synchronization emits
no Change. Native clicks restore the accepted exclusive selection while their
owner request is pending. Choosing the accepted selected option cannot clear it.

The group is one managed Tab stop, preferring its selected enabled option and
otherwise the first enabled option. Arrow keys wrap through enabled options;
Home/End select the first/last; Enter/Space request the focused option. Modified
chords remain available to enclosing shortcuts. Actual focused options stay
within the group's managed traversal identity. All-disabled groups are skipped.
Keyboard navigation requests Change through bounded originating-window owner
admission. Refusal restores controlled values and reports the standard OnError
and control-error property. No unbounded retry queue is added.

The Mac semantic group has a radio-group role and readable selected label;
option records have radio-button roles, stable generated IDs, selected boolean,
measured frames and per-option focus/press actions. Set uses domain IDs, without
label fallback. Semantic focus admission validates the live option binding; its
pending ID is an intent to focus the current visible, enabled same-ID target at
application time, including a replacement binding and workbench/modal namespaces.
If the option disappears or its group becomes hidden/disabled before application,
the request is refused. This differs from queued Change, which retains binding
identity and refuses replacements. Manual captures use full semantic identities
so group-local option IDs cannot shadow unrelated managed controls. Tests of native
objects and completed software frames remain separate from physical OS input,
human VoiceOver, graphics-memory and sustained-performance qualification.

## Toolbars, status bars and form groups

`Toolbar`, `StatusBar` and `FormGroup` are retained containers, with ordinary
managed Children and disabled ancestry. They add no managed focus stop or own
Activate/Change callback; callbacks belong on interactive children. Child IDs,
editing drafts, current callbacks and semantic identity survive label/status
updates and reordering. Moving children out of a removed container happens
before its private host is destroyed.

Toolbar defaults horizontal; Orientation Vertical selects a vertical toolbar.
Only its main axis scrolls when constrained. The cross axis retains natural
control size subject to explicit style limits. Overflow uses the same retained
controls in a native scrolling region, preserving private editing state. Keyboard
focus reveals an offscreen tool. Tab follows each child's ordinary managed stop;
unconsumed main-axis arrows wrap through enabled stops, and Home/End focus the
first/last enabled stop. Child text and choice editing consume their keys before
toolbar navigation. Modified chords remain available to enclosing shortcuts.
Toolbar publishes the Mac toolbar role and ordinary child semantics. Nested
managed layouts follow the same enabled traversal; a nested Toolbar has its own
nearest scope. Moving a layout out removes the old navigation scope. Modal
content is a separate input scope and extensions own their private input trees.
When all stops or the toolbar/ancestor are disabled, recognized traversal keys
are consumed without changing focus. Ordinary trees without a toolbar/old scope
skip native binding traversal.

FormGroup lays out a visible escaped Text heading over a retained column of
managed children. Its semantic group uses Accessibility.Name when supplied and
Text otherwise. StatusBar lays out an escaped Text status message beside a row of
managed children, exposing that literal Text as its readable semantic Value.
Use Accessibility.Name for a stable status-region name. These private captions
are omitted as duplicate semantic nodes and do not enter focus traversal. An
empty Text hides the caption and a later nonempty Text restores it. StatusBar
and FormGroup publish Mac group roles, with the children's own supported actions.
These kinds do not imply live-region announcements or individual field-label
associations; give fields their own Accessibility.Name.

Managed styles override outer defaults. Gap, Align and Justify also apply to the
managed content host in status/form containers; removing these properties restores
native defaults. Typography/colors inherit through the private structure. No
native child plan can delete/rebuild managed content during an overflow/layout
pass. The gallery demonstrates all three containers. Actual Manual tests cover
horizontal/vertical overflow, keyboard reveal, disabled navigation, text-key
precedence and completed status pixels. Inert Cocoa objects check roles. Physical
Mac input, human VoiceOver and native memory/performance remain separate gates.

## Adjustable split panes

`SplitPane` is a controlled container with exactly two stable child nodes. Empty
or Horizontal Orientation places them side by side; Vertical places them top/bottom.
Number is the first pane's accepted fraction of available space after the divider.
Pixel rounding reserves at least one pixel for each positive pane when space permits,
without changing its accepted fraction.
An omitted Numeric resolves to Min .05, Max .95 and Step .01. Explicit bounds and
Number must be finite, 0 < Min < Max < 1, with a positive step no larger than the
range. Both panes stay mounted without implicit collapse. Drawing a pixel for
each pane requires space for the divider and gap plus two pixels on the main axis;
explicit content minimums can still overflow or clip a viewport. Collapsing/showing
regions belongs in application layout state. Nested split panes form larger resizable layouts.

The native divider requests Change.Number from pointer positions clamped to the
current range; main-axis arrows use Step and Home/End select Min/Max. Programmatic
reconciliation emits no callback. Native resize does not optimistically mutate
accepted pane proportions. On the bounded owner queue, the latest callback and
bounds are used. Removal, disabled ancestry, pane identity/order changes or an
orientation change retire an earlier queued resize. A refused request leaves the
accepted ratio unchanged. A missing Change makes resizing read-only while leaving
the divider a readable focus target. Native digit collapse/equalization shortcuts
are suppressed because they would bypass owner acceptance; application commands
still run through the ordinary command router.

Private pane hosts retain their scroll state and isolate native layout styling
from managed child styles. Reorder/reparent preserves compatible child native
identity and uncommitted editor state. Each pane scrolls independently when its
content does not fit. Explicit outer styles override defaults and removing them
restores native defaults. Focus request to SplitPane's ID targets its divider;
Tab visits first-pane targets, divider, then second-pane targets within the usual
active modal/selected-tab scope. `split:` is a reserved descriptor ID prefix for
qualified private divider targets.

Accessibility exposes an AXSplitGroup with child content and a named AXSplitter
in visual order. The divider has current value, configured bounds, current focus,
set/increment/decrement when writable, and exact rendered geometry. Its native
orientation describes the divider: vertical for side-by-side panes, horizontal
for top/bottom panes. Actions obey current disabled/lifetime guards and use the
managed focus proxy. Inert Cocoa and actual Manual software tests remain separate
from human VoiceOver, physical pointer/key routing, native attachment and Mac
performance/endurance qualification.
## Built-in symbols and action presentation

`ui.Icon` is a passive retained symbol leaf. `Node.Icon` selects one of the
portable `ui.Symbol` names listed by `ui.Symbols()`: add, remove, check, close,
search, settings, save, open, undo, redo, copy, cut, paste, info, warning, delete,
left/right/up/down, menu, expand and collapse. Unknown names and raw SVG/path
strings are refused. No application filesystem or network lookup occurs.
The Mac adapter supplies its bundled vector artwork. This is a finite built-in
set; custom image/assets remain a separate API and distribution-rights gate.

An Icon requires a symbol and rejects Activate/Change callbacks and children.
It never enters Tab order. Accessibility.Name, or Text as its fallback alternate
label, exposes a passive image role; an unlabeled symbol is decorative and omitted
from the accessible hierarchy. Icon Text is alternate text, not visible text.
Default dimensions are 24 device-independent pixels; explicit size/style and
application-local theme override them. A changed symbol keeps the widget identity
and replaces the rendered artwork. Removal releases the native widget/cached art.

Button and ToggleButton optionally use Icon beside their visible Text. An
icon-only action/toggle requires a nonblank Accessibility.Name. Symbol changes or
removal retain control identity, accepted toggle state and owner callback binding.
Omitting Icon restores the ToggleButton's ordinary checked-state indicator.
Other kinds refuse Icon. A standalone icon has no action/value; an action's icon
does not add a separate accessible child or focus stop.

`Node.Appearance` belongs only on Button. Empty preserves the existing filled
button; ButtonPrimary is filled, ButtonSecondary is outlined, ButtonQuiet uses
text emphasis, and ButtonTool uses compact toolbar-action treatment. Appearance
is presentation: activation, enabled state and keyboard/semantic routing retain
their contracts. App-local theme and explicit styles remain authoritative.
Owner delivery revalidates the current callback and disabled ancestry; removal
retires queued button activations. These software/native-object checks do not
qualify physical Mac input, human VoiceOver or native memory/performance.

## Managed image control

`ui.Image` is a passive leaf whose Image is a nonempty immutable
`*canvas.TiledImage`. It reuses the retained-image ownership/patch contract and
can share one source between windows. It rejects children, Activate/Change, nil
or zero sources. Other kinds refuse Image/Fit. Rendering performs no file/network
lookup or decoding; replacement is an owner-accepted source pointer change.

`Fit` defaults to ImageContain (preserve aspect and center all source pixels).
ImageCover preserves aspect, fills the box and crops centrally; ImageStretch
fills both axes independently. Fit changes painting, not the layout box. Style
minimum/maximum and Grow determine that box. Without explicit dimensions the
preferred size uses source pixels as device-independent pixels, with the longest
edge capped at 256; one nonzero minimum derives the other axis from source aspect.
The preferred size respects maximums. Removing styles restores these defaults.
Padding is outside the image's content. Painting clips to visible content, retains
the full unscrolled source mapping, honors opacity/ancestor masks, and blends
straight-alpha pixels through the scene painter. It does not allocate a scaled
per-widget bitmap or promise direct GPU/partial-texture upload.

Text supplies alternate text; Accessibility.Name overrides it. A nonblank name
exposes a passive Mac image role; an unlabeled image is decorative. It adds no
Tab stop or semantic action. Semantic geometry is the visible widget rectangle.
Source, fit and style updates retain widget identity. Removal drops the widget's
source reference; the immutable caller-owned image remains usable.

`ui.DecodeImage(io.Reader)` accepts PNG/JPEG on a caller-owned reader. It reads at
most 32 MiB plus one refusal byte, verifies bounded dimensions before decoding,
then normalizes/copies pixels into TiledImage. It never opens a path or URL.
Encoded byte bounds and the existing 8192-axis/256-MiB source bounds are separate
from decoder/conversion temporary memory and elapsed reader time. Large decoding
belongs on a cancellable application worker; the caller owns reader cancellation.
Unknown formats, malformed images, oversized input/dimensions and reader errors
are refused. Human Mac input/display/accessibility and native memory/performance
acceptance remain separate from software rendering and inert Cocoa role checks.

Action menu buttons have transient actions rather than controlled selection. See
the [MenuButton contract](menu-button.md) for input, focus, Cocoa semantics and
originating-window popup retirement.

The [inline color picker](color-picker.md) composes existing Slider, TextField,
Image and FormGroup nodes into an owner-held RGBA editor.


MenuButton Options also support controlled `Checkable`/`Checked` actions and
passive `Separator` entries. Change.Checked proposes the inverse accepted check
state. Ordinary actions only identify their action, and separators have no input
or focus. Other kinds refuse these fields. Full lifetime, bounds and semantics
are specified by the [action-menu contract](menu-button.md).

## Closable document tabs

Tabs options may set Closable and supply Node.TabClose. Closing proposes a stable
PageID on the owner without native selection/deletion. The application accepts by
removing that option/page and selecting a remaining page, or removes the Tabs node
when closing the last page. Rejected requests retain pages and drafts. Pointer
close and Delete/Backspace on a focused tab share bounded guarded admission;
modified command chords retain application routing. See [the close contract](tab-close.md)
for binding, modal lifetime, private target geometry and scope limits.


## Progress shapes

`ui.ProgressRing` and `ui.ProgressGauge` render controlled fractions from zero to
one as a circle or semicircle. `Number` is the accepted fraction; optional `Text`
is literal text inside the shape. These passive leaves reject input callbacks,
children, orientation, nonfinite values and values outside the interval. They
remain outside focus order and expose a read-only native progress indicator with
minimum zero, maximum one and the accepted value. Accessibility.Name supplies
a stable descriptive label independently of the changing visible text.

The normal style options override and restore native size/typography defaults.
Default ring size is 128×128 logical units; gauge size is 112×64. The controls
start no animation or timer. Their native meter type is selected before initial
styling, keyed updates retain identity, and kind changes replace the old shape.
Manual completed-frame tests cover fractions 0, 0.25, 0.75 and 1 at DPI
96/120/144/192, with pixel changes, nonempty geometry and no ongoing timers.
Offscreen/native-selector fixtures remain separate from physical input, human
screen-reader and the complete current-source Mac qualification.

## Breadcrumb navigation

The owner-held `ui.Breadcrumb` component adds bounded controlled hierarchy/location
navigation over ordinary managed nodes. Ancestors are quiet buttons and the current
location is a passive label. Item/root identity, queued request guards, read-only
behavior and overflow/focus/accessibility follow [the breadcrumb contract](breadcrumb.md).

## Selector click focus

A pointer click on an enabled focusable Selector takes actual quiet keyboard
focus before opening its menu. The menu owns its own item focus; selection or
cancellation leaves keyboard navigation with the originating Selector. Disabled,
read-only or nonfocusable native controls do not acquire focus through this hook.
Taking focus does not itself propose a value or consume the click. Model rejection
continues to restore the controlled selection, including the next real menu.
The native gallery check and actual software FIFO regression separately verify
pointer selection, actual ContainsFocus/public focus, Return reopen and Escape.

An accepted NumberField Enter completes editing and follows normal focus
traversal. A later numeric key step requires actual focus to be routed back to
the field. The native check's remaining focus guard reflects that consumer
sequence requirement, not a request to keep the Spinner focused artificially.

Post-preview reusable input helpers and Mac appearance are specified in
[mac-widget-expansion.md](mac-widget-expansion.md). They compose existing controls;
TextField additionally accepts a literal Placeholder hint.
