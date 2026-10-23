package ui

import (
	"image/color"
	"testing"
)

func TestColorWellControlledEditingAndOpacity(t *testing.T) {
	p := &ColorWell{ID: "ink", Color: color.NRGBA{R: 1, G: 2, B: 3, A: 255}}
	calls, invalid := 0, 0
	p.Change = func(e ColorEdit) { calls++; p.Color = e.Apply(p.Color) }
	p.Invalid = func(error) { invalid++ }
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	if n.Children[1].Open {
		t.Fatal("editor unexpectedly open")
	}
	colorNode(n, ColorWellButtonID(p.ID)).Activate()
	n, err = p.View()
	if err != nil || !n.Children[1].Open {
		t.Fatal("open", err)
	}
	red := colorNode(n, ColorFieldID(ColorWellPickerID(p.ID), ColorRed))
	alpha := colorNode(n, ColorFieldID(ColorWellPickerID(p.ID), ColorAlpha))
	red.Change(ChangeEvent{Number: 11})
	alpha.Change(ChangeEvent{Number: 0})
	if calls != 2 || p.Color != (color.NRGBA{R: 11, G: 2, B: 3, A: 0}) {
		t.Fatal("queued edits", calls, p.Color)
	}
	colorNode(n, ColorFieldID(ColorWellPickerID(p.ID), ColorHex)).Change(ChangeEvent{Text: "#badHEX"})
	if invalid != 1 || calls != 2 {
		t.Fatal("invalid hex", invalid, calls)
	}
	// A rejected proposal neither overwrites Color nor appends recent history.
	p.Change = func(ColorEdit) { calls++ }
	alpha.Change(ChangeEvent{Number: 100})
	if calls != 3 || p.Color.A != 0 || len(p.Recent) != 0 {
		t.Fatal("rejection changed accepted state")
	}
}

func TestColorWellSessionsDisabledRecentAndIdentity(t *testing.T) {
	red, blue := color.NRGBA{R: 255, A: 255}, color.NRGBA{B: 255, A: 255}
	calls := 0
	p := &ColorWell{ID: "ink", Recent: []color.NRGBA{red, red, blue}, Change: func(ColorEdit) { calls++ }}
	n, _ := p.View()
	colorNode(n, ColorWellButtonID(p.ID)).Activate()
	n, _ = p.View()
	oldRecent := colorNode(n, "ink/color-well/recent/"+ColorHexString(red))
	oldEdit := colorNode(n, ColorFieldID(ColorWellPickerID(p.ID), ColorRed))
	oldDismiss := n.Children[1].Dismiss
	p.Recent = []color.NRGBA{blue}
	oldRecent.Activate()
	if calls != 0 {
		t.Fatal("removed recent color admitted")
	}
	p.Disabled = true
	oldEdit.Change(ChangeEvent{Number: 9})
	if calls != 0 {
		t.Fatal("disabled stale edit admitted")
	}
	oldDismiss()
	p.Disabled = false
	n, _ = p.View()
	if n.Children[1].Open {
		t.Fatal("disabled-dismissed editor reopened")
	}
	colorNode(n, ColorWellButtonID(p.ID)).Activate()
	n, _ = p.View()
	oldDismiss()
	oldEdit.Change(ChangeEvent{Number: 9})
	oldRecent.Activate()
	if calls != 0 || !p.open {
		t.Fatal("old session changed reopened panel")
	}
	currentEdit := colorNode(n, ColorFieldID(ColorWellPickerID(p.ID), ColorRed))
	currentEdit.Change(ChangeEvent{Number: 9})
	if calls != 1 {
		t.Fatal("current session did not edit")
	}
	p.ID = "new-ink"
	currentEdit.Change(ChangeEvent{Number: 10})
	if calls != 1 {
		t.Fatal("old identity admitted")
	}
}

func TestColorWellBoundsFocusAndPreviewReuse(t *testing.T) {
	for _, p := range []*ColorWell{nil, {}, {ID: "ink", Recent: make([]color.NRGBA, 17)}} {
		if _, err := p.View(); err == nil {
			t.Fatal("invalid component accepted")
		}
	}
	p := &ColorWell{ID: "ink", Color: color.NRGBA{G: 128, A: 128}, Change: func(ColorEdit) {}}
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	if got := FocusOrder(n); len(got) != 1 || got[0] != ColorWellButtonID(p.ID) {
		t.Fatal("closed focus", got)
	}
	previewID := ColorWellPickerID(p.ID) + "/color/preview"
	first := colorNode(n, previewID).Image
	n, _ = p.View()
	if colorNode(n, previewID).Image != first {
		t.Fatal("unchanged immutable preview replaced")
	}
	p.Color = color.NRGBA{R: 255, A: 255}
	n, _ = p.View()
	if colorNode(n, previewID).Image == first || first.NRGBAAt(0, 0) == (color.NRGBA{R: 255, A: 255}) {
		t.Fatal("old preview changed")
	}
	p.Change = nil
	n, _ = p.View()
	if len(FocusOrder(n)) != 0 || !n.Children[0].Disabled {
		t.Fatal("read-only retained activation/focus")
	}
	collision := Node{ID: "root", Kind: Column, Children: []Node{n, {ID: ColorWellButtonID(p.ID), Kind: Label}}}
	if Validate(collision) == nil {
		t.Fatal("component namespace collision accepted")
	}
}
