# Integrating the Mac input widgets

ColorWell, TokenField, FontPicker, PathControl and SecureField are retained Go
components in `github.com/dshills/vellum/ui`. Keep each instance in your application
model and call its `View()` for every application view. Each returns `(ui.Node,
error)`; handle errors before composing the node into your window. Do not create
new component instances on every render: color-editor sessions and secure reveal
state belong to the retained instance.

Callbacks propose edits on the application owner. Accept an edit by updating the
component's exported value; ignore it to reject. The native driver's normal edit
path refreshes the window after delivery. For changes from workers or other
external sources, use your application's owner queue and invalidate the window.
Do not mutate or render the same component concurrently across owners/windows.

The [independent public-API contract](../examples/sdk-contract/mac_widgets_test.go)
compiles all five in one descriptor tree, checks their exported focus targets,
and accepts color, tag, font, path and protected-text proposals. It has no native
driver dependency. The [Mac widget demo](mac-widget-demo.md) covers running the
native gallery; descriptor tests do not establish native input or visual results.

## Theme and layout

Create `theme := ui.MacTheme(false)` for light or `ui.MacTheme(true)` for dark, and
supply it as `cogent.Driver{Theme: theme}`. Apply `theme.Kinds[ui.TextField]` to
SecureField.Style, TokenField.InputStyle and FontPicker.FieldStyle. Apply
`theme.Kinds[ui.FormGroup]` to ColorWell.PopoverStyle. Composite internals own their
text/accessory padding; do not add a second field padding layer to their children.

Place the returned nodes in FormGroup or Column containers with suitable padding
and gaps. Give each composite a distinct stable ID and reserve its generated child
namespace. Use exported ID helpers for focus requests instead of constructing
internal child names. Put a root `FocusRequest` on your window's actual root, not
on a nested component. The gallery's return-to-entry policy after token acceptance, rejection or removal
is an application policy, not an implicit TokenField side effect. Its root focus
request advances to the next field at capacity and is canceled when inputs are
disabled; request only enabled targets after applying your model edit.

## Accepting values

The snippets below show owner-model initialization. Retain these values; render
with `node, err := component.View()`. Integrate error handling into your window
view as in the [gallery](../drivers/cogent/cmd/gallery/input_helpers.go).

A color edit identifies the changed channel. Apply it to the latest accepted
color so independent queued channel edits compose:

```go
well := &ui.ColorWell{ID: "ink", Label: "Ink",
    Color: color.NRGBA{R: 52, G: 120, B: 246, A: 255},
    PopoverStyle: theme.Kinds[ui.FormGroup]}
well.Change = func(e ui.ColorEdit) { well.Color = e.Apply(well.Color) }
```

`Recent` is an optional caller-owned list of up to 16 accepted colors. Vellum does
not append proposals to it. The editor is a same-window managed popover, not the
shared system color panel. `ui.ColorWellButtonID("ink")` is its focus target.

Token additions need application-assigned identities. Use a monotonic counter or
a domain identity; never use the current list length, which can collide after a
removal. Keep accepted labels unique. Removal uses TokenID, not list position:

```go
nextToken := 1
tags := &ui.TokenField{ID: "tags", Label: "Tags",
    InputStyle: theme.Kinds[ui.TextField]}
tags.Change = func(e ui.TokenEvent) {
    if e.Remove {
        kept := make([]ui.Token, 0, len(tags.Tokens))
        for _, token := range tags.Tokens {
            if token.ID != e.TokenID { kept = append(kept, token) }
        }
        tags.Tokens = kept
        return
    }
    tags.Tokens = append(tags.Tokens, ui.Token{
        ID: fmt.Sprintf("tag-%d", nextToken), Text: e.Text})
    nextToken++
    tags.Draft = ""
}
```

The field bounds tokens to 32 and suggestions to 64; labels are at most 256 UTF-8
bytes. Return or a suggestion produces one proposal. Blank/duplicate/invalid text
calls Invalid without changing the accepted tokens. `ui.TokenInputID("tags")` and
`ui.TokenRemoveID("tags", tokenID)` are the entry/removal focus targets.

Font catalogs contain installed family names with stable family/typeface IDs.
The selected IDs must exist in the supplied catalog:

```go
font := &ui.FontPicker{ID: "font", Label: "Drawing font",
    Value: ui.FontSelection{FamilyID: "sans", TypefaceID: "regular", Size: 14},
    Families: []ui.FontFamily{{ID: "sans", Name: "Helvetica Neue",
        Typefaces: []ui.FontTypeface{{ID: "regular", Name: "Regular", Weight: 400}}}},
    FieldStyle: theme.Kinds[ui.TextField]}
font.Change = func(e ui.FontSelectionEvent) { font.Value = e.Apply(font.Value) }
```

Size is 6–144 logical pixels. Typed names must match the catalog. The driver can
substitute an unavailable font; Vellum does not enumerate or install fonts.
`FontFamilyID`/`FontTypefaceID` identify combo composites: use
`ui.ComboInputID(ui.FontFamilyID("font"))` to focus the family text entry.
`ui.FontSizeID("font")` addresses the size stepper.

Path navigation changes your virtual domain path; rendering grants no filesystem
access. Accept an ancestor by replacing Items with your domain's new path:

```go
path := &ui.PathControl{ID: "location", Label: "Location", Items: []ui.PathItem{
    {ID: "home", Label: "Home"},
    {ID: "drawing", Label: "Untitled.vellum", Kind: ui.PathFile},
}}
path.Navigate = func(e ui.PathEvent) {
    for i, item := range path.Items {
        if item.ID == e.ItemID {
            path.Items = append([]ui.PathItem(nil), path.Items[:i+1]...)
            return
        }
    }
}
```

The current item is not a navigation destination. IDs remain stable when labels
change. Use `ui.PathSegmentID("location", itemID)` for a visible segment and
`ui.PathOverflowID("location")` for omitted ancestors. Overflow is bounded by the
explicit MaxVisible setting rather than automatically adapting to window width.

Use only synthetic text when trying protected entry:

```go
secure := &ui.SecureField{ID: "password", Label: "Synthetic password",
    Text: "synthetic", AllowReveal: true, Style: theme.Kinds[ui.TextField]}
secure.Change = func(e ui.ChangeEvent) { secure.Text = e.Text }
```

Return commits entry through Change. `ui.SecureInputID("password")` targets the
editor. Reveal changes visible glyphs; accessibility/native surrounding-text
queries remain protected. Revealed entry permits ordinary copying. The value is
an ordinary Go string, with no memory-erasure or OS secure-event-input guarantee.

## Limits

Appearance is custom rendered with explicit light/dark colors. Automatic OS theme,
accent, high-contrast and locale discovery are not implemented. Panels stay in
one native window. Font/color pickers do not own shared system panels, path
controls do not access files, and token chips do not support inline editing or
drag rearrangement. Chip abbreviation uses a fixed 24-rune budget rather than
viewport-responsive truncation. A chip wider than its toolbar viewport can be
clipped; a reproduced example uses the 400px software window at DPI320 (about
200 logical pixels) with long mixed-script labels. Allow enough space for the
chosen typography. The matched-scale regression uses a 320-logical-pixel field,
32 long labels, both themes and DPI160/320; it verifies last-chip reveal and
exactly-once keyboard removal. This does not qualify physical display scaling.
See the [full control reference](controls.md) and
[demo limitations](mac-widget-demo.md#current-limits).
