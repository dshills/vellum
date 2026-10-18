# List box

`ui.ListBox` is a bounded, always-visible single-selection list. It uses the
ordinary public `ui.Node`, `Options`, `SelectedID` and `Change` API. Applications
own accepted selection; the native adapter owns focus and scroll position.

```go
ui.Node{
    ID: "layers", Kind: ui.ListBox,
    Options: []ui.Option{
        {ID: "background", Label: "Background"},
        {ID: "guides", Label: "Guides", Disabled: true},
        {ID: "foreground", Label: "Foreground"},
    },
    SelectedID: model.Layer,
    Change: func(e ui.ChangeEvent) { model.Layer = e.SelectedID },
    Accessibility: ui.Accessibility{Name: "Layers"},
}
```

There are 0..`ui.MaxListBoxOptions` (256) materialized options. IDs are unique,
nonblank UTF-8 without NUL; labels are nonblank literal UTF-8 text. Duplicate
labels are allowed. `SelectedID` is empty or present, including a disabled item
that the application wants to display as selected. Disabled items never become
input targets. The descriptor is a leaf: `Children`, `Orientation`, `Activate`
and menu/tab option metadata are invalid. A nil `Change` makes selection
read-only. Empty/all-disabled lists have no keyboard stop; other lists expose
one managed Tab stop, with focus initially on the enabled selected item or the
first enabled item. Multi-selection and larger data use the separate indexed
data-view APIs.

Up/Down skip disabled options and clamp at the ends. Home/End choose the first
or last enabled option. Page Up/Down move by the completed content viewport
height divided by the focused row's measured full allocation, bounded to
1..256 enabled steps. Before layout this falls back to one step. The target
receives focus and is scrolled into view. Typing a single letter or digit cycles
through labels with that Unicode initial, case-insensitively; there is no timed
multi-character search buffer. Modified shortcuts pass through to the existing
application routing. Tab leaves the list through normal managed focus.

Navigation that moves to another item proposes its stable ID. Space/Return,
pointer activation and accessibility selection explicitly propose the target,
even when it is already accepted. Reaching a navigation boundary adds no
proposal. Every admitted native proposal retains FIFO order and uses the latest
callback; it is not deduplicated against a possibly stale rendered selection.
The application can accept, reject or normalize it. A read-only list still
supports row focus and scrolling. Rendering and programmatic setters never
emit `Change`; transient native checked flags immediately restore the accepted
model, including after queue refusal.

The default native viewport is vertical, at least 160 logical pixels wide and
at most 240 logical pixels high, with a vertical scrollbar when needed. Ordinary
`Style` size, typography and theme overrides apply and can be removed. Labels
are escaped as literal text. Row widgets retain identity by item ID through
reorder/relabel; removed items are retired. List state and queued callbacks are
bounded by option count and the existing application queue.

Queued input retains its originating whole-list visibility/disabled/dialog
lease and row incarnation/eligibility token. An observed removal, hide/show,
disable/enable, modal exclusion or window closure refuses the old input, even
if the same ID returns. Host accessibility `set` also captures an immutable
ordered ID/disabled snapshot so it cannot target a later row incarnation.
Changing accepted selection alone leaves native proposals admissible.
Accessibility retains the bridge's existing semantic-generation revalidation;
admission does not promise callback delivery after a newer semantic snapshot.

Mac semantics use AXList and AXRow, selected children/rows, accepted label value,
row count/index and focus separate from selection. All bounded rows remain in
the semantic tree. Visible rows/children include only positive measured clipped
row frames intersecting the list. Enabled offscreen rows can be focused and
revealed through the existing focus path. Disabled/read-only mutation actions
are refused; namespaced Workbench/modal focus uses the standard managed proxy.
The Cocoa selected-row setter requests one enabled current child through the
same owner selection path. Empty, multiple, foreign and stale row requests
refuse; application code can still clear accepted selection programmatically.
The native selected-rows writability query follows the current mutation policy,
including read-only/disabled transitions without replacing a compatible list.
The transport's setter return code does not promise owner admission or acceptance.

The gallery's **Lists** page demonstrates scrolling, disabled items, rejection,
read-only selection, stable reorder and programmatic clear. Run the gallery
after the usual native preparation with `cd drivers/cogent && go run ./cmd/gallery`.
An independent SDK consumer exercises this public contract. Portable, actual
offscreen input and inert Cocoa selector checks are separate from physical Mac
pointer/key input and human VoiceOver verification. Those live gates and wider
Mac performance/endurance/distribution acceptance remain open.

The gallery also offers `-navigation-native-check`, a bounded native Go FIFO
consumer sharing the ordinary list constructor. Its named 47-input/41-observation
record includes actual list scrolling/pointer input, explicit accepted-row repeat,
rejection/read-only/disabled behavior and Tab. See the
[result](../docs/reports/macos-navigation-catalogue/README.md). Native routing is
separate from physical input, native pixels and human accessibility acceptance.
