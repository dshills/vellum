# Available UI elements

Vellum describes retained controls with `ui.Node`. Each node has a stable ID;
application-owned values are accepted state, and callbacks request changes on the
application owner. The portable API has no native dependencies. The Mac adapter
renders these elements and publishes their supported accessibility semantics.

| Element | Public kind | Main state |
| --- | --- | --- |
| Label | `ui.Label` | `Text` |
| Action button | `ui.Button` | `Text`, `Activate` |
| Checkbox | `ui.Checkbox` | `Checked`, `Change` |
| On/off switch | `ui.Switch` | `Checked`, `Change` |
| Toggle button | `ui.ToggleButton` | `Checked`, `Change` |
| Selection menu | `ui.Selector` | `Options`, `SelectedID`, `Change` |
| Radio choices | `ui.RadioGroup` | `Options`, `SelectedID`, `Change`, `Orientation` |
| Segmented choices | `ui.SegmentedControl` | `Options`, `SelectedID`, `Change`, `Orientation` |
| Slider | `ui.Slider` | `Numeric`, `Number`, `Change` |
| Numeric field | `ui.NumberField` | `Numeric`, `Number`, `Change` |
| Single-line field | `ui.TextField` | `Text`, `Change` |
| Document editor | `ui.MultilineText` | `Editor`, `SoftWrap`, `TabInserts` |
| Progress bar | `ui.ProgressBar` | `Number` fraction, `Orientation` |
| Decorative separator | `ui.Separator` | `Orientation` |
| Managed data grid | `ui.DataGrid` | `Grid`, `GridChange`, `GridCommit` |
| Managed outline | `ui.DataTree` | `Tree`, `TreeChange`, `TreeDisclosure` |
| Tabs | `ui.Tabs` | `Options`, `SelectedID`, page children |
| Scroll container | `ui.Scroll` | Children |
| Modal dialog | `ui.Dialog` | `Open`, `Dismiss`, children |
| Toolbar | `ui.Toolbar` | Children, `Orientation`, scrolling overflow |
| Status bar | `ui.StatusBar` | `Text` status, children |
| Labeled form group | `ui.FormGroup` | `Text` heading, children |
| Adjustable split pane | `ui.SplitPane` | Two children, `Number` fraction, `Numeric`, `Orientation`, `Change` |
| Horizontal/vertical layout | `ui.Row`, `ui.Column` | Children, `Style` |
| Custom driver extension | `ui.Extension` | `Custom` |

All elements support stable identity, disabled ancestry, local styles and themes.
Focusable controls support managed focus requests. Interactive controls use their
specified callbacks; progress bars and separators are passive leaves and reject
Change/Activate callbacks. Separators do not enter the accessible hierarchy.

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
