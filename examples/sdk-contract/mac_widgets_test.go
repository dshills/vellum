package sdkcontract_test

import (
	"image/color"
	"slices"
	"testing"

	"github.com/dshills/vellum/ui"
)

// This module has no driver dependency and cannot access ui's private state.
// It verifies integration of the five composites through exported descriptors;
// native routing and visual appearance are checked separately by the driver.
func TestPublicMacWidgetCompositionAndAcceptedValues(t *testing.T) {
	theme := ui.MacTheme(false)
	secure := &ui.SecureField{ID: "password", Label: "Synthetic password", Text: "synthetic", AllowReveal: true, Style: theme.Kinds[ui.TextField]}
	path := &ui.PathControl{ID: "location", Items: []ui.PathItem{{ID: "home", Label: "Home"}, {ID: "drawing", Label: "Drawing", Kind: ui.PathFile}}}
	font := &ui.FontPicker{ID: "font", Value: ui.FontSelection{FamilyID: "sans", TypefaceID: "regular", Size: 14}, Families: []ui.FontFamily{{ID: "sans", Name: "Helvetica Neue", Typefaces: []ui.FontTypeface{{ID: "regular", Name: "Regular", Weight: 400}}}}, FieldStyle: theme.Kinds[ui.TextField]}
	tags := &ui.TokenField{ID: "tags", Tokens: []ui.Token{{ID: "design", Text: "Design"}}, InputStyle: theme.Kinds[ui.TextField]}
	well := &ui.ColorWell{ID: "color", Color: color.NRGBA{R: 52, G: 120, B: 246, A: 255}, PopoverStyle: theme.Kinds[ui.FormGroup]}
	requests := map[string]int{}
	secure.Change = func(e ui.ChangeEvent) { requests["secure"]++; secure.Text = e.Text }
	path.Navigate = func(e ui.PathEvent) {
		requests["path"]++
		for i, item := range path.Items {
			if item.ID == e.ItemID {
				path.Items = append([]ui.PathItem(nil), path.Items[:i+1]...)
				return
			}
		}
	}
	font.Change = func(e ui.FontSelectionEvent) { requests["font"]++; font.Value = e.Apply(font.Value) }
	tags.Change = func(e ui.TokenEvent) {
		requests["tags"]++
		if e.Remove {
			tags.Tokens = slices.DeleteFunc(tags.Tokens, func(v ui.Token) bool { return v.ID == e.TokenID })
		} else {
			tags.Tokens = append(tags.Tokens, ui.Token{ID: "review", Text: e.Text})
			tags.Draft = ""
		}
	}
	well.Change = func(e ui.ColorEdit) { requests["color"]++; well.Color = e.Apply(well.Color) }
	view := func() ui.Node {
		root := ui.Node{ID: "fields", Kind: ui.Column}
		for _, render := range []func() (ui.Node, error){secure.View, path.View, font.View, tags.View, well.View} {
			n, err := render()
			if err != nil {
				t.Fatal(err)
			}
			root.Children = append(root.Children, n)
		}
		if err := ui.Validate(root); err != nil {
			t.Fatal(err)
		}
		return root
	}
	root := view()
	for _, id := range []string{ui.SecureInputID("password"), ui.PathSegmentID("location", "home"), ui.ComboInputID(ui.FontFamilyID("font")), ui.FontSizeID("font"), ui.TokenRemoveID("tags", "design"), ui.TokenInputID("tags"), ui.ColorWellButtonID("color")} {
		if !slices.Contains(ui.FocusOrder(root), id) {
			t.Fatal("composed public focus target missing", id)
		}
	}
	publicMacWidgetNode(t, root, ui.SecureInputID("password")).Change(ui.ChangeEvent{Text: "new synthetic"})
	publicMacWidgetNode(t, root, ui.PathSegmentID("location", "home")).Activate()
	publicMacWidgetNode(t, root, ui.FontSizeID("font")).Change(ui.ChangeEvent{Number: 18})
	publicMacWidgetNode(t, root, ui.TokenInputID("tags")).Change(ui.ChangeEvent{Text: "Review"})
	publicMacWidgetNode(t, root, ui.ColorWellButtonID("color")).Activate()
	root = view()
	publicMacWidgetNode(t, root, ui.ColorFieldID(ui.ColorWellPickerID("color"), ui.ColorRed)).Change(ui.ChangeEvent{Number: 200})
	root = view()
	if secure.Text != "new synthetic" || len(path.Items) != 1 || font.Value.Size != 18 || len(tags.Tokens) != 2 || tags.Tokens[1].Text != "Review" || well.Color != (color.NRGBA{R: 200, G: 120, B: 246, A: 255}) {
		t.Fatal("public owner acceptance failed")
	}
	for _, kind := range []string{"secure", "path", "font", "tags", "color"} {
		if requests[kind] != 1 {
			t.Fatal("proposal count", kind, requests[kind])
		}
	}
	field := publicMacWidgetNode(t, root, ui.SecureInputID("password"))
	if !field.Secure || field.Reveal {
		t.Fatal("composed secure field lost masked entry contract")
	}
	publicMacWidgetNode(t, root, ui.TokenRemoveID("tags", "review")).Activate()
	if len(tags.Tokens) != 1 || tags.Tokens[0].ID != "design" || requests["tags"] != 2 {
		t.Fatal("public token removal failed")
	}
	view()
}

func publicMacWidgetNode(t *testing.T, n ui.Node, id string) ui.Node {
	t.Helper()
	var find func(ui.Node) (ui.Node, bool)
	find = func(n ui.Node) (ui.Node, bool) {
		if n.ID == id {
			return n, true
		}
		for _, child := range n.Children {
			if found, ok := find(child); ok {
				return found, true
			}
		}
		return ui.Node{}, false
	}
	if found, ok := find(n); ok {
		return found
	}
	t.Fatalf("public descriptor %q missing", id)
	return ui.Node{}
}
