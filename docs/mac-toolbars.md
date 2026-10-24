# Mac-style content toolbars

Open the **Toolbars** page in the private Mac widget gallery. The light and dark
packages contain the same examples. These recipes use existing public `ui.Toolbar`,
`ui.Button`, `ui.Separator`, and `ui.Style` APIs. `Button.IconPlacement` selects
inline or icon-above-caption content without adding another action.

The catalogue compares three deliberate, application-local action sizes:

| Size | Inline height | Icon-above height | Label size | Icon-only width |
| --- | --- | --- | --- | --- |
| Small | 26 pt | 52 pt | 12 pt | 26 pt |
| Regular | 32 pt | 60 pt | 13 pt | 32 pt |
| Large | 40 pt | 72 pt | 16 pt | 40 pt |

Choose Icons, Text, Icons and text, or **Icon above text**. **Full-color icons** is enabled initially;
turn it off to compare the bundled monochrome symbols. Original 96 px sample
artwork is shared across the three sizes and both appearances. Explicit icon-only accessible names and
hover help remain present in every mode; labels do not become the identity.
The mixed bar combines a labeled New action, text-only Open, and icon actions.
Separators divide file and history groups. Delete demonstrates disabled appearance
and skipped keyboard navigation. Sample callbacks only update a visible status.

Narrow preview caps horizontal toolbars at 280 pt. **Overflow menu** starts enabled:
a trailing double chevron opens the actions that do not fit, with their labels
and original artwork. Turn it off to compare retained scrolling overflow. The vertical icon bar has a 154 pt viewport. Tab follows
ordinary control order, unconsumed orientation arrows move within the toolbar,
and Home/End reveal the first/last enabled action. Text editors keep their own
editing commands. Lock-key flags preserve ordinary navigation, including the
NumLock bit supplied on macOS. Disable actions applies to all sample toolbars but leaves the
catalogue display controls available.

Full-color action artwork is available through the public `IconImage` field on
Button and MenuButton. Supply immutable pixels with `ui.DecodeImage` (PNG/JPEG)
or `canvas.NewTiledImage`; keep that source and reuse it between View calls.
Leave `Icon` empty. The pixels retain their colors, aspect ratio and transparency,
scale into the font-relative icon slot, and dim with a disabled action.
The image adds no separate accessible action or focus stop. `Icon` and `IconImage`
are mutually exclusive; image-only actions require an accessible name.

```go
artwork, err := ui.DecodeImage(bytes.NewReader(pngBytes))
// Handle err, then retain artwork in application state.
button := ui.Node{ID: "open", Kind: ui.Button, IconImage: artwork,
    Appearance: ui.ButtonTool, Activate: openDrawing,
    Accessibility: ui.Accessibility{Name: "Open", Help: "Open a drawing"}}
```

Here `bytes` is the standard library package; the application supplies PNG bytes
and `openDrawing`. No filesystem path or SVG string is loaded by the native button.
MenuButton options also accept `IconImage` (or a bundled `Icon`), alongside their
required label. Checked actions keep their check marker beside supplied artwork.
ToggleButton does not support IconImage in this increment.

Set `IconPlacement: ui.IconAboveText` on an ordinary Button with artwork and a
nonblank caption. Full-color `IconImage` and monochrome `Icon` both work. Artwork
is centered above the caption with a 4 pt default gap and a 2 em icon slot; local
Gap/Padding/Typography styles can override spacing and caption size. Give a fixed
control enough height for both parts, or omit the fixed height for intrinsic sizing.
The gallery uses 64/72/84 pt widths and 52/60/72 pt heights, with short labels.
Scroll the page to see the lower examples when the taller option is selected.
The mixed and vertical comparisons retain their own per-item layouts.

```go
ui.Node{ID: "save", Kind: ui.Button, Text: "Save", IconImage: artwork,
    IconPlacement: ui.IconAboveText, Appearance: ui.ButtonTool, Activate: save,
    Accessibility: ui.Accessibility{Name: "Save drawing"},
    Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 72, Height: 60}),
        Padding: ui.Some(ui.UniformInsets(5)), Gap: ui.Some(float32(4)),
        Typography: ui.Some(ui.Typography{Size: 13})}}
```

Empty placement or `ui.IconLeading` restores the usual inline content. Changing
placement keeps the same native action, caption, keyboard focus, and callback;
artwork and caption add no private accessibility targets. MenuButton and
ToggleButton keep their inline presentation in this increment.

Use `ui.WithToolbarOverflow(toolbar)` after composing a fresh horizontal toolbar.
It adds the stable `ui.ToolbarOverflowID(toolbar.ID)` More actions button. The
adapter measures actual controls and available content width on each layout,
including display scaling, item gaps and padding. All actions remain visible when
they fit; otherwise it reserves the chevron and keeps a fitting prefix. Hidden
separators are normalized so the menu has no leading, trailing or doubled rules.
The menu always shows captions, including names for icon-only actions, and
preserves disabled state, order, full-color pixels and the original callback.

```go
tools := ui.WithToolbarOverflow(ui.Node{ID: "drawing-tools", Kind: ui.Toolbar,
    Accessibility: ui.Accessibility{Name: "Drawing tools"},
    Children: []ui.Node{{ID: "drawing-save", Kind: ui.Button, IconImage: artwork,
        Text: "Save", IconPlacement: ui.IconAboveText, Activate: save,
        Appearance: ui.ButtonTool}},
})
```

The menu supports 1..64 horizontal Buttons, Separators or toolbar spaces. Spaces never become menu entries. Do not reuse the reserved
chevron child ID; call the helper once, after per-item presentation changes.
Editors, nested toolbars, placement sources, context-menu hosts and vertical bars
keep ordinary scrolling. Allow at least the chevron's intrinsic width plus bar
padding; an arbitrarily smaller viewport can clip any control. Changing display
mode or toolbar width revokes a menu whose contents change. Removal, disabling,
Escape, outside click and window shutdown use managed-menu lifetime handling.
Return/Space or Down opens More actions; the popup uses native menu keyboard
navigation. Tab and toolbar arrows skip hidden controls, with End reaching More.
Widening restores the same native controls. Overflow is a custom-rendered
same-window menu, with no NSToolbar title-bar integration or macOS system palette.

A regular monochrome action, independently consumable by an application:

```go
ui.Node{ID: "document-tools", Kind: ui.Toolbar,
    Accessibility: ui.Accessibility{Name: "Document tools"},
    Style: ui.Style{Gap: ui.Some(float32(4)),
        Padding: ui.Some(ui.UniformInsets(5)),
        Background: ui.MacTheme(dark).Kinds[ui.FormGroup].Background},
    Children: []ui.Node{{ID: "save", Kind: ui.Button, Icon: ui.SymbolSave,
        Appearance: ui.ButtonTool, Activate: save,
        Accessibility: ui.Accessibility{Name: "Save", Help: "Save the drawing"},
        Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 32, Height: 32}),
            MaxSize: ui.Some(ui.Size{Width: 32, Height: 32}),
            Padding: ui.Some(ui.UniformInsets(3)),
            Border: ui.Some(ui.Border{}),
            Background: ui.Some(color.NRGBA{}),
            Typography: ui.Some(ui.Typography{Size: 13})},
    }},
}
```

Here `save` and `dark` belong to the application, and `color` is `image/color`.
Set Text and omit Icon for a text action; set both for an inline icon and label.
Remove the fixed maximum width for intrinsic-width text. Apply the same size
recipe to every action in a bar. The generated vector icon follows the action's
font-relative sizing. Local styles leave hover and keyboard-focus feedback intact.

These are custom-rendered **content** toolbars. They do not install NSToolbar,
provide title-bar merging, vibrancy, the macOS system customization palette, direct dragging on the toolbar,
or cross-window overflow menus. Explicit light/dark palettes and size
metrics do not follow automatic NSControlSize or OS appearance changes. Artwork
uses original full-color sample pixels or Vellum’s bundled vector symbols;
neither extracts macOS app icons or uses dynamic SF Symbols.

The focused public-Manual regression checks light/dark relative size geometry,
stacked artwork/caption alignment without clipping, inline restoration,
retained action identities across display modes, exactly-once activation,
disabled actions, and keyboard revelation of narrow overflow. Native keyboard,
accessibility actions and screenshots are recorded separately in the delivery
receipt; neither is a claim of human VoiceOver or physical-pointer qualification.


## Reordering and hiding actions

Choose **Customize…** beside Small, Regular, Large or Mixed. Each toolbar has its
own settings. The window-modal content sheet shows full-color artwork and labeled
visibility checkboxes. Drag the leading grip to a new position; a blue insertion
line shows the proposed order after a six-point movement threshold. Hidden actions
remain in the panel, retain their artwork, and can be dragged normally. Up/down
buttons move an item one place earlier/later,
including while hidden. These buttons and checkboxes support keyboard focus and
named accessibility actions. Disabled toolbar commands can still be configured;
editing the layout never runs their command. Restore Defaults resets both order
and visibility. Done and Escape close the sheet and restore focus to its opener.
Changes apply immediately; Escape keeps them. The packaged gallery remembers order and hidden actions across launches in
`~/Library/Application Support/Vellum/MacWidgets/toolbars-light.json` (or
`toolbars-dark.json`). Restore Defaults is saved too. A visible error explains a
failed load/save; current edits stay usable. Retired IDs are pruned when the
catalogue changes. Invalid or unsupported files are preserved without overwriting them; saving
stays disabled for that session. Repair/remove the file and relaunch to resume saving. These files belong to the demo; portable components still perform no I/O.

Public applications use `ui.ToolbarLayout` and a retained `ui.ToolbarCustomizer`:

```go
// Keep catalogue complete, with stable action IDs, original artwork/callbacks.
customizer := &ui.ToolbarCustomizer{ID: "customize-tools", Toolbar: catalogue,
    ReturnFocusID: "show-customizer"}
customizer.Change = func(next ui.ToolbarLayout) { customizer.Value = next }
customizer.Dismiss = func() { customizer.Open = false }
// On the owner, opening sets customizer.Open = true. In each View:
configured, err := customizer.Value.Apply(catalogue)
// Handle err. Add overflow only if at least one visible item remains.
if len(configured.Children) > 0 {
    configured = ui.WithToolbarOverflow(configured)
}
customizer.Toolbar = catalogue
panel, err := customizer.View()
// Handle err; return configured and panel alongside an external opener.
```

Retain the customizer on one application owner. Accept a Change proposal by
assigning Value immediately; the component resolves queued edits against that
latest accepted configuration. Reject by leaving Value unchanged. Keep View
updated on close/disable/configuration changes so old sessions are revoked.
The driver also guards removed/closed native controls. The generated control ID
helpers support independent consumers and automation. The complete catalogue
supports 1..64 ordinary horizontal Buttons, Separators or toolbar spaces, with at least one action;
editors, nested controls, placement sources and action-level context-menu hosts are refused in the customizable catalogue. The Toolbar itself can own its context menu.
Apply customization before adding overflow, so hidden actions never reappear in
More actions. Reordering preserves retained visible control identity; hiding
removes an action, and showing it mounts a new native control with the same public
ID, artwork and application callback. Leading/trailing/adjacent separators are
normalized. Hiding every action is valid: keep the opener outside the toolbar.

Order is a partial ID list, with omitted items appended in catalogue order.
Hidden is also a list of IDs. Unknown/duplicate IDs produce an error rather than
silently changing a different action. The value has JSON fields `order`/`hidden`;
applications may store it and handle catalogue migrations themselves. The gallery
persists these values in its application-owned preference files. Customization
uses an explicit content sheet with drag grips and keyboard-accessible arrow
controls. Cross-window customization and NSToolbar's system palette are not supplied.


Customization dragging is confined to the open panel in the same native window.
Release inside the list commits at most one controlled proposal. Release outside,
Escape, capture loss, disabled/removed source, changed order, panel closure and
window shutdown cancel without editing the layout or clicking the release target.
A grip click does not reorder. During a drag, wheel/trackpad input scrolls this
list only; moving near its top/bottom edge scrolls in bounded steps. Holding still
at an edge does not run a background autoscroll timer. Scroll clipped rows use
full row midpoints, current scroll geometry and the originating sheet/display
coordinate mapping. The retained panel owns at most 64 handles and one native
capture/preview, with no drag timers or file I/O. The arrow buttons remain the
keyboard/accessibility alternative through the same controlled Change path.

`ToolbarCustomizationDragID(panelID, itemID)` identifies a generated grip.
`Node.ToolbarReorder` and its `ToolbarReorderHandle` are adapter capabilities
emitted by `ToolbarCustomizer.View`; applications should consume the customizer
rather than construct drag descriptors or call their native drop hooks directly.
Automated routing covers drag commits, hidden membership, full-color row artwork,
scrolling, scale, release consumption and teardown. Packaged native-pointer checks
are recorded separately from deterministic routing evidence.

## Fixed and flexible spaces

The Mixed toolbar demonstrates `ui.ToolbarSpace` (16 pt fixed gap) and
`ui.ToolbarFlexibleSpace` (8 pt minimum, expands to consume spare width). Supply
these leaf kinds with stable IDs among horizontal toolbar actions. Fixed-space
width can be styled with matching MinSize/MaxSize widths. Customize lists them as
Space and Flexible Space; reorder or hide them just like actions. Their settings
persist in the gallery. They add no focus/AX target and never become menu entries.
Flexible space contracts as the toolbar narrows; actions still use the original
overflow menu, artwork and callbacks. Hiding every action leaves an empty bar.

## Toolbar context menus

Right-click or Control-click a horizontal toolbar, action or gap for display
choices and **Customize Toolbar…**. Shift-F10 on an action provides keyboard
access. Each bar can use Icons Only, Text Only, Icons and Text, or Icon Above Text;
the menu checks its current mode. Mixed Presentation restores the mixed recipe.
These per-toolbar choices persist across launches. The display buttons above the
examples reset overrides and apply their common mode. Layout order/hidden settings
are preserved when switching display. A child's explicit context menu takes
precedence over the containing toolbar. Applications attach `ui.ContextMenu` to
the public Toolbar node; existing menu options, callbacks and lifecycle apply.

The customization grip now uses six dots and a grab cursor. The blue insertion
marker eases to its new slot over 100 ms; it fades in without moving the rows or
changing drop validation. Completion/cancellation removes the scene animation and
preview immediately. Focus outlines follow the current enabled control and modal
scope, including after repeated opening/closing; changing focus does not resize
controls. Animation uses the existing native scene clock, with no drag timers.
