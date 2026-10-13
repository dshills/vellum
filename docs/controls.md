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
A toggle button holds one boolean value; it does not imply a radio or segmented
single-selection group.

The independent gallery in `drivers/cogent/cmd/gallery` demonstrates the control
set. After driver dependency preparation, run `go run ./cmd/gallery` from
`drivers/cogent`. This opens a desktop window. Software-frame tests and native
object fixtures cover controlled behavior and the new Mac role/value mappings;
human VoiceOver and physical input checks remain part of Mac qualification.
See the [controls contract](../specs/controls.md) for detailed event, focus,
editing, accessibility and lifetime behavior.
