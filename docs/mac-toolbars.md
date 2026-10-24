# Mac-style content toolbars

Open the **Toolbars** page in the private Mac widget gallery. The light and dark
packages contain the same examples. These recipes use existing public `ui.Toolbar`,
`ui.Button`, `ui.Separator`, and `ui.Style` APIs; no new widget API is required.

The catalogue compares three deliberate, application-local action sizes:

| Size | Minimum/control height | Label size | Icon-only width |
| --- | --- | --- | --- |
| Small | 26 pt | 12 pt | 26 pt |
| Regular | 32 pt | 13 pt | 32 pt |
| Large | 40 pt | 16 pt | 40 pt |

Choose Icons, Text, or Icons and text. **Full-color icons** is enabled initially;
turn it off to compare the bundled monochrome symbols. Original 96 px sample
artwork is shared across the three sizes and both appearances. Explicit icon-only accessible names and
hover help remain present in every mode; labels do not become the identity.
The mixed bar combines a labeled New action, text-only Open, and icon actions.
Separators divide file and history groups. Delete demonstrates disabled appearance
and skipped keyboard navigation. Sample callbacks only update a visible status.

Narrow preview caps horizontal toolbars at 280 pt, preserving the same controls
in a scrolling viewport. The vertical icon bar has a 154 pt viewport. Tab follows
ordinary control order, unconsumed orientation arrows move within the toolbar,
and Home/End reveal the first/last enabled action. Text editors keep their own
editing commands. Lock-key flags preserve ordinary navigation, including the
NumLock bit supplied on macOS. Disable actions applies to all sample toolbars but leaves the
catalogue display controls available.

Full-color action artwork is available through the public `IconImage` field on
Button and MenuButton. Supply immutable pixels with `ui.DecodeImage` (PNG/JPEG)
or `canvas.NewTiledImage`; keep that source and reuse it between View calls.
Leave `Icon` empty. The pixels retain their colors, aspect ratio and transparency,
scale into the font-relative leading icon slot, and dim with a disabled action.
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
ToggleButton and popup menu *items* do not support IconImage in this increment.

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
provide title-bar merging, vibrancy, customization palettes, draggable items,
icon-above-label layouts, or overflow menus. Explicit light/dark palettes and size
metrics do not follow automatic NSControlSize or OS appearance changes. Artwork
uses original full-color sample pixels or Vellum’s bundled vector symbols;
neither extracts macOS app icons or uses dynamic SF Symbols.

The focused public-Manual regression checks light/dark relative size geometry,
retained action identities across display modes, exactly-once activation,
disabled actions, and keyboard revelation of narrow overflow. Native keyboard,
accessibility actions and screenshots are recorded separately in the delivery
receipt; neither is a claim of human VoiceOver or physical-pointer qualification.
