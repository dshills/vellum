# Available UI elements

Vellum describes retained controls with `ui.Node`. Each node has a stable ID;
application-owned values are accepted state, and callbacks request changes on the
application owner. The portable API has no native dependencies. The Mac adapter
renders these elements and publishes their supported accessibility semantics.

| Element | Public kind | Main state |
| --- | --- | --- |
| Label | `ui.Label` | `Text` |
| Action button | `ui.Button` | `Text`, `Activate` |
| Bundled icon | `ui.Icon` | `Icon`, optional alternate `Text`/accessible name |
| Image preview | `ui.Image` | Immutable `Image`, `Fit`, alternate `Text`/accessible name |
| Checkbox | `ui.Checkbox` | `Checked`, `Change` |
| On/off switch | `ui.Switch` | `Checked`, `Change` |
| Toggle button | `ui.ToggleButton` | `Checked`, `Change` |
| Action menu | `ui.MenuButton` | `Options`, `Change.SelectedID` action requests |
| Selection menu | `ui.Selector` | `Options`, `SelectedID`, `Change` |
| Radio choices | `ui.RadioGroup` | `Options`, `SelectedID`, `Change`, `Orientation` |
| Segmented choices | `ui.SegmentedControl` | `Options`, `SelectedID`, `Change`, `Orientation` |
| Slider | `ui.Slider` | `Numeric`, `Number`, `Change` |
| Inline color picker | `ui.ColorPicker.View()` | Straight-alpha `Color`, `Change(ColorEdit)`, `Invalid` |
| Numeric field | `ui.NumberField` | `Numeric`, `Number`, `Change` |
| Single-line field | `ui.TextField` | `Text`, `Change` |
| Document editor | `ui.MultilineText` | `Editor`, `SoftWrap`, `TabInserts` |
| Progress bar | `ui.ProgressBar` | `Number` fraction, `Orientation` |
| Decorative separator | `ui.Separator` | `Orientation` |
| Managed data grid | `ui.DataGrid` | `Grid`, `GridChange`, `GridCommit` |
| Managed outline | `ui.DataTree` | `Tree`, `TreeChange`, `TreeDisclosure` |
| Tabs / closable document tabs | `ui.Tabs` | `Options.Closable`, `SelectedID`, `TabClose`, page children |
| Scroll container | `ui.Scroll` | Children |
| Modal dialog | `ui.Dialog` | `Open`, `Dismiss`, children |
| Toolbar | `ui.Toolbar` | Children, `Orientation`, scrolling overflow |
| Status bar | `ui.StatusBar` | `Text` status, children |
| Labeled form group | `ui.FormGroup` | `Text` heading, children |
| Collapsible section | `ui.Disclosure` | `Text` heading, `Checked` expansion, `Change`, children |
| Adjustable split pane | `ui.SplitPane` | Two children, `Number` fraction, `Numeric`, `Orientation`, `Change` |
| Horizontal/vertical layout | `ui.Row`, `ui.Column` | Children, `Style` |
| Custom driver extension | `ui.Extension` | `Custom` |

All elements support stable identity, disabled ancestry, local styles and themes.
Focusable controls support managed focus requests. Interactive controls use their
specified callbacks; progress bars and separators are passive leaves and reject
Change/Activate callbacks. Separators do not enter the accessible hierarchy.

Use `ui.Disclosure` for collapsible inspector or settings sections. Keep its child
descriptors present in both states; the adapter retains their native controls
and private editing state while collapsed. `Checked` controls expansion and the
Change callback accepts or rejects the requested boolean. A collapsed section
keeps its heading in focus order and hides child focus/accessibility targets.

```go
ui.Node{ID: "details", Kind: ui.Disclosure, Text: "Details", Checked: expanded,
    Change: func(e ui.ChangeEvent) { expanded = e.Checked },
    Children: []ui.Node{{ID: "name", Kind: ui.TextField, Text: name,
        Change: func(e ui.ChangeEvent) { name = e.Text }}},
}
```

```go
func view() ui.Node {
    return ui.Node{ID: "settings", Kind: ui.Column, Children: []ui.Node{
        {ID: "enabled", Kind: ui.Switch, Text: "Enable snapping",
            Checked: enabled, Change: func(e ui.ChangeEvent) { enabled = e.Checked }},
        {ID: "pinned", Kind: ui.ToggleButton, Text: "Pin panel",
            Checked: pinned, Change: func(e ui.ChangeEvent) { pinned = e.Checked }},
        {ID: "divider", Kind: ui.Separator},
        {ID: "loading", Kind: ui.ProgressBar, Number: progress,
            Accessibility: ui.Accessibility{Name: "Loading progress"}},
    }}
}
```

The example uses application-owned `enabled`, `pinned` and `progress` values.
Progress must be finite and within 0..1. Progress bars and separators default to
horizontal; set `Orientation: ui.Vertical` for a vertical element. Explicit size
and grow styles override native defaults; removing them restores those defaults.
A toggle button holds one boolean value. Use `ui.RadioGroup` or
`ui.SegmentedControl` for a mutually exclusive choice. These retain 1..64 options
with distinct nonempty IDs and one selected ID. Duplicate labels are allowed.
Disabled selections can be displayed, but users cannot choose disabled options.
Radio groups default to vertical, segmented controls to horizontal; Orientation
can change either. Tab visits each group once; arrows wrap through enabled options,
Home/End jump to first/last, and Enter/Space request the focused choice. Selection
cannot be cleared by clicking its currently selected option.

The independent gallery in `drivers/cogent/cmd/gallery` demonstrates the control
set. After driver dependency preparation, run `go run ./cmd/gallery` from
`drivers/cogent`. This opens a desktop window. Software-frame tests and native
object fixtures cover controlled behavior and the new Mac role/value mappings;
human VoiceOver and physical input checks remain part of Mac qualification.
See the [controls contract](../specs/controls.md) for detailed event, focus,
editing, accessibility and lifetime behavior.

Toolbars retain their child controls when a narrow viewport requires scrolling.
They support horizontal/default or vertical Orientation. Tab visits child controls;
unconsumed arrows wrap through enabled children, and Home/End focus the first/last.
Text-field editing keeps precedence. A FormGroup gives related controls a visible
Text heading and semantic group. StatusBar Text displays and exposes the current
status; use Accessibility.Name for a stable region name. These containers do not
add a focus stop or own input callback.

```go
ui.Node{ID: "preferences", Kind: ui.Column, Children: []ui.Node{
    {ID: "tools", Kind: ui.Toolbar, Accessibility: ui.Accessibility{Name: "Document tools"},
        Children: []ui.Node{{ID: "save", Kind: ui.Button, Text: "Save", Activate: save}}},
    {ID: "details", Kind: ui.FormGroup, Text: "Details", Children: []ui.Node{
        {ID: "name", Kind: ui.TextField, Text: name,
            Accessibility: ui.Accessibility{Name: "Name"},
            Change: func(e ui.ChangeEvent) { name = e.Text }},
    }},
    {ID: "status", Kind: ui.StatusBar, Text: status,
        Accessibility: ui.Accessibility{Name: "Application status"}},
}}
```

The example uses application-owned name/status and a save callback. Gap, alignment
and justification style the managed form/status content as well as the outer
container; padding and dimensions apply to the outer container. Empty headings
or status messages hide the private caption without removing child controls.

SplitPane places exactly two retained children side by side, or top/bottom with
`Orientation: ui.Vertical`. `Number` is the first pane's accepted space fraction.
Omitted `Numeric` defaults to 5..95% with a 1% keyboard step; explicit bounds must
stay strictly inside 0..1 to prevent complete pane collapse. Dragging the divider or
using its main-axis arrows/Home/End requests a `ChangeEvent.Number`. Tab moves
through the first pane, divider, then second pane. Each pane scrolls independently.

```go
ui.Node{ID: "workspace", Kind: ui.SplitPane, Number: divider, Grow: true,
    Accessibility: ui.Accessibility{Name: "Inspector width"},
    Change: func(e ui.ChangeEvent) { divider = e.Number },
    Children: []ui.Node{inspector, content},
}
```

The application owns divider and the two child nodes. Without Change the divider
is readable and focusable, while its position remains fixed. Nested SplitPane
nodes support more than two regions. The gallery has a Split panes page.

`ui.Icon` shows one of the 23 bundled vector symbols returned by `ui.Symbols()`.
It is decorative when unlabeled, or a passive accessible image when Text or
Accessibility.Name supplies alternate text. It adds no Tab stop. The default
size is 24 device-independent pixels; MinSize/theme styles can override it.
Raw SVG, paths and unknown names are refused.

Buttons and toggle buttons accept Icon beside their Text. Icon-only actions
require a nonblank Accessibility.Name; Accessibility.Help supplies the tooltip.
Button Appearance may be ButtonPrimary (filled), ButtonSecondary (outlined),
ButtonQuiet (text emphasis) or ButtonTool (compact toolbar action). Omission
preserves the existing filled appearance. Styles remain application-local.
Removing a toggle's Icon restores its ordinary checked-state indicator.

```go
ui.Node{ID: "search", Kind: ui.Button, Icon: ui.SymbolSearch,
    Appearance: ui.ButtonTool, Activate: search,
    Accessibility: ui.Accessibility{Name: "Search", Help: "Search the document"}}
```

The application supplies search. The gallery's Icons and actions page shows the
built-in catalog and action appearances. Presentation changes keep each control's
identity, callback and accepted state; icons add no private accessible action.

`ui.Image` displays an immutable `*canvas.TiledImage` shared with other windows or
GPU canvases. Fit is ImageContain/default, ImageCover or ImageStretch. Contain and
cover preserve aspect, center the source, and respectively letterbox or crop.
Styles determine the box; fitting does not change it. The default preferred size
uses source pixels as device-independent pixels, capped at 256 on the longest
edge. A single minimum dimension derives its partner from source aspect.
An unlabeled preview is decorative; Text/Accessibility.Name adds alternate text
and a passive image role. It adds no Tab stop or action.

```go
ui.Node{ID: "preview", Kind: ui.Image, Image: picture, Fit: ui.ImageContain,
    Text: "Document preview", Style: ui.Style{
        MinSize: ui.Some(ui.Size{Width: 320, Height: 180}),
    }}
```

The application owns picture. Construct it with canvas.NewTiledImage from NRGBA
pixels, or ui.DecodeImage from a caller-owned PNG/JPEG reader on a worker.
DecodeImage bounds encoded data to 32 MiB and retains the existing source limits;
reader cancellation and temporary decode memory remain caller concerns. Source
replacement/patches preserve control identity and do not mutate earlier versions.
The gallery's Images page demonstrates fitting, transparency and source changes.

Use `ui.MenuButton` for a flat list of transient actions, controlled check actions
and passive separators. Keep SelectedID empty;
the callback receives the stable action ID in ChangeEvent.SelectedID. Duplicate
labels are allowed and disabled items remain readable. Enter/Space, pointer or
Down opens; Escape dismisses; Tab/Shift-Tab resumes managed traversal.

```go
ui.Node{ID: "actions", Kind: ui.MenuButton, Text: "Document actions",
    Options: []ui.Option{{ID: "export", Label: "Export"},
        {ID: "archive", Label: "Archive", Disabled: !canArchive}},
    Change: func(e ui.ChangeEvent) { requestAction(e.SelectedID) },
}
```

The application supplies canArchive and requestAction. This menu has no selected
value; the gallery's Icons and actions page demonstrates it. See the
[MenuButton contract](../specs/menu-button.md) and
[Mac checkpoint](reports/macos-menu-button/README.md) for tested scope.

`ui.ColorPicker` supplies inline RGBA sliders, a hex field and a checker preview.
Retain it in your application model and render its View on the owner:

```go
picker := &ui.ColorPicker{ID: "ink", Label: "Ink", Color: color.NRGBA{R: 32, G: 96, B: 224, A: 255}}
picker.Change = func(edit ui.ColorEdit) {
    picker.Color = edit.Apply(picker.Color)
    picker.ValidationMessage = ""
}
picker.Invalid = func(err error) { picker.ValidationMessage = err.Error() }
// Return picker.View() from an app.Window View, or include it among children.
```

The example imports image/color and github.com/dshills/vellum/ui. Apply edits to
the current accepted value; this preserves other queued channel changes and RGB
at zero alpha. You may reject or normalize proposals in Change. Nil Change makes
the component read-only. Reserve its ID + "/color/" namespace. Focus the hex
field with ui.ColorFieldID("ink", ui.ColorHex) and a normal FocusRequest. See the
[contract](../specs/color-picker.md) for bounded input and cache behavior. The
controls gallery includes a Colors page with reset, disable and focus actions.


Checkable menu actions receive a proposed boolean along with their stable ID.
Accept it in the application model; reopening displays the accepted checkmark:

```go
ui.Node{ID: "view-options", Kind: ui.MenuButton, Text: "View", Options: []ui.Option{
    {ID: "grid", Label: "Show grid", Checkable: true, Checked: showGrid},
    {ID: "divider", Separator: true},
    {ID: "reset", Label: "Reset view"},
}, Change: func(e ui.ChangeEvent) {
    if e.SelectedID == "grid" { showGrid = e.Checked }
    if e.SelectedID == "reset" { resetView() }
}}
```

The application owns showGrid and resetView. Separator entries require a unique
ID and no other fields. They count toward the 64-entry menu limit but add no
keyboard/accessibility/input target. Checkable/Checked/Separator fields apply only
to MenuButton Options. The gallery's Icons and actions page includes View options.
The [check-action checkpoint](reports/macos-menu-checks/README.md) records the
software routing, visual and inert Cocoa checks; physical input and human
VoiceOver remain separate acceptance work.

Use closable Tabs for document workspaces. Set Closable on individual options and
handle TabClose with the Tabs node ID and stable PageID. The handler accepts by
removing the matching option/page and selecting a remaining page; leave state
unchanged to reject. Closing the last page must replace/remove the Tabs node.
Native pages and drafts remain intact until the accepted view removes them.
Delete/Backspace on a focused tab and its visible close button request the same
owner decision. Application command shortcuts remain available. The gallery's
Document tabs page demonstrates rejection, reopening and the empty last-page state.
See [the contract](../specs/tab-close.md) for lifetime and accessibility details.

Canvas2D optionally exposes existing scene objects through `CanvasObjects` stable
IDs and names. Option+Left/Right/Home/End selects visible named objects in paint
order, with one host focus stop. See the [contract](../specs/canvas-accessibility.md).


## Mac input helpers

The post-preview widget branch adds three reusable components. They compose the
existing managed primitives and work through the public Go API:

| Component | Purpose |
| --- | --- |
| ui.SearchBox | Query input, placeholder, search symbol and Clear |
| ui.FormField | A label, help, required indication and validation text for a control |
| ui.InlineNotice | Inline feedback with an optional action |
| ui.MacTheme(dark) | Application-local compact light/dark defaults |

Retain mutable components on the application owner:

~~~go
search := &ui.SearchBox{ID: "find", Label: "Search objects"}
search.Change = func(e ui.SearchEvent) { search.Text = e.Text }
node, err := search.View() // include node in your window's view; handle err
// Request focus with ui.SearchInputID("find").
~~~

Clear requests empty Text through the same handler. Leave Text unchanged to reject
a proposal. Nil Change disables the input. The ordinary single-line field commit
behavior is retained; there is no new per-keystroke query stream or search worker.
Use InputStyle and ClearStyle for the private input/action presentation; Style
applies to the outer row. Placeholder defaults to Label, while Accessibility.Name
continues to supply the input's name independently of the hint.

~~~go
field, err := (ui.FormField{
    ID: "name-field", Label: "Drawing name", Required: true,
    Help: "Use a unique name.", ValidationMessage: nameError,
    Content: ui.Node{ID: "name", Kind: ui.TextField, Text: name,
        Change: func(e ui.ChangeEvent) { name = e.Text }},
}).View()
~~~

The application supplies name/nameError and performs validation. Content must be
a focusable leaf; its stable ID and callbacks are retained. Required is descriptive.
An explicit accessible name is preserved; label/help/error text otherwise supplies
the field's name and description. Error labels add no extra focus stop.

~~~go
notice := &ui.InlineNotice{
    ID: "save-message", Tone: ui.NoticeWarning, Text: "The drawing has unsaved changes.",
    ActionLabel: "Save", Activate: save,
}
message, err := notice.View()
~~~

The application supplies save. Omit both action fields for passive feedback.
Information/success/warning/error tones choose a symbol and accessible description.
These are ordinary semantic groups, not modal alerts or live-region announcements.

To try the focused gallery after dependency preparation:

~~~sh
cd drivers/cogent
go run ./cmd/gallery -mac-widgets
go run ./cmd/gallery -mac-widgets -mac-dark
~~~

The regular gallery also has an Input helpers page. Applications opt into the
same appearance with cogent.Driver{Theme: ui.MacTheme(false)} (true for dark).
It uses installed Helvetica Neue and fixed light/dark colors. This is Mac-style
custom rendering with existing native input/window integration, not AppKit
NSControls or automatic accent/high-contrast tracking. See the
[post-preview contract](../specs/mac-widget-expansion.md) for behavior and limits.

For a compact Mac-style up/down affordance, set NumberButtons to
ui.NumberButtonsStepper on a NumberField. The upper/lower arrow halves and normal
arrow keys use the same value, step and clamped bounds; the field remains directly
editable. Omitting NumberButtons preserves the existing inline minus/plus style.

### Editable combo box

`ui.ComboBox` provides controlled free text plus up to 64 predefined suggestions.
Use `ComboInputID(root)` for field focus. `ComboEvent.OptionID` is empty for typing
and identifies a stable choice for menu selection. Retain one component per owner;
update `Text` in `Change` to accept. Apply `MacTheme(dark).Kinds[ui.TextField]` to
its outer `Style`, with zero padding, for a single compact field outline. Tab to
the suggestions button and press Down/Return to choose; Escape dismisses.

### Context menus

Set `Node.ContextMenu` on Button, Canvas2D or ListBox with `Options` and `Change`.
The event identifies the host and stable action; checked actions propose `Checked`.
Right/Control-click and Shift-F10 open it; Escape closes. Existing MenuButton
check/separator rules apply. This does not replace text-editing context menus.

### Anchored popover

`ui.Popover{ID, AnchorID, Open, Dismiss, Children}.View()` supplies a compact
managed panel below a same-window Button. Set `Open=false` in `Dismiss`. Apply a
Mac FormGroup style for an inset neutral surface. Escape/outside release closes;
focus returns to the anchor. Children retain keyed identity. Missing/hidden anchors
close the panel. Edge clamping is supported; detachable panels are not.

### Multiple-selection source list

Use `ListBox` with `Multiple:true`, `SelectedIDs` and `SelectionChange`; accept via
`selected = e.Apply(selected)`. Single-selection APIs remain unchanged. Command
click toggles, Shift selects a range and Command-Shift adds a range. Shift with
Up/Down/Home/End extends by keyboard. Apply to your current model so queued edits
compose; ignore the event to reject. Keep accepted IDs present in Options.

### Date/time picker

Retain `ui.DateTimePicker` with a `time.Time` Value, optional inclusive Min/Max,
`Mode` and `Change`. Accept by updating Value; Invalid reports refused text.
`DateTimeInputID(root)` targets the field. Date modes offer an anchored calendar;
month navigation does not change Value until a day is selected. A caller-supplied
Go `Layout` controls formatting. Default ISO/24-hour formatting and English
Sunday-first calendar are explicit; system-locale discovery is not implemented.

Combo suggestions align below the full field's left edge and match its width.
`MenuButton.MenuAnchorID` optionally names a same-window visible managed anchor;
without it, action menus open below their own button. Popup options are full-width
menu rows on one surface, not separately outlined push buttons.

### Color well

`ui.ColorWell.View()` returns a compact swatch and same-window anchored editor.
Retain the component; accept `ColorEdit` with `well.Color = edit.Apply(well.Color)`.
Optional `Recent` contains at most 16 caller-owned colors; rejected edits never
mutate it. RGBA/opacity and hex editing reuse ColorPicker validation. Set
`PopoverStyle` from MacTheme's FormGroup style. `ColorWellButtonID` is the focus
target. This is a managed panel, not the shared system color panel.

### Token field

`ui.TokenField.View()` composes removable chips and a controlled entry with
suggestions. Supply at most 32 `Token{ID, Text}` values and at most 64 suggestions.
`TokenEvent` requests one addition or removal; the application assigns new IDs,
accepts by updating Tokens and clears Draft on acceptance. Enter one token with
Return, or choose a suggestion. Focus a chip with Tab and press Return/Space to
remove it. Duplicate labels, blank/invalid or >256-byte tokens are refused.
No in-place chip editing, comma-triggered commit or chip dragging is provided.

### Font picker

`ui.FontPicker.View()` offers family, typeface, compact size stepper and a live
plain-text preview. Supply a bounded catalog of `FontFamily`/`FontTypeface`, using
installed font names. Accept with `picker.Value = event.Apply(picker.Value)`. Sizes
are 6–144 logical pixels. Set FieldStyle from MacTheme TextField for light/dark
inputs. Typed names must match a catalog entry; unavailable installed fonts may
substitute. No automatic enumeration, installation or system font panel.

### Path control

`ui.PathControl.View()` extends breadcrumb navigation with folder/document icons,
ancestor menus, bounded visible segments and central label truncation. Supply
`PathItem` values with stable IDs; accept `PathEvent` by updating Items. At most 64
items; MaxVisible defaults to four, configurable 3–8, plus an overflow menu. Root,
parent and current location remain visible. Full labels remain in accessibility
and menu text. Truncation is explicit rather than window-width responsive. No
filesystem access, URL resolution, permission grants or drag-and-drop is implied.

### Secure text field

`ui.SecureField.View()` composes a native-backed masked TextField and optional
Show/Hide button (`AllowReveal`). Nil Change disables entry and reveal. Accept Text through Change; `SecureInputID` is
the managed focus target. `Node.Secure` and `Node.Reveal` are available directly
on TextField. Reveal requires Secure and only changes visible glyphs: AX values,
selection/geometry text and native surrounding-text/diagnostic queries remain
protected, with the macOS secure text subrole. Masked core fields disallow copy
and cut; explicit reveal permits ordinary text copying. Disabling remasks the
composite. Values remain ordinary Go strings; no secret store, memory erasure or
OS secure-event-input guarantee is provided.
