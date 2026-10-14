package ui

import (
	"errors"
	"image/color"
	"math"
	"strings"
	"testing"
)

func colorNode(n Node, id string) Node {
	if n.ID == id {
		return n
	}
	for _, child := range n.Children {
		if found := colorNode(child, id); found.ID != "" {
			return found
		}
	}
	return Node{}
}
func TestColorHexFormsAndStraightAlphaRoundTrip(t *testing.T) {
	cases := map[string]color.NRGBA{"#123": {R: 17, G: 34, B: 51, A: 255}, "#1234": {R: 17, G: 34, B: 51, A: 68}, " #aBcDeF ": {R: 171, G: 205, B: 239, A: 255}, "#F001AB00": {R: 240, G: 1, B: 171, A: 0}}
	for text, want := range cases {
		got, err := ParseColorHex(text)
		if err != nil || got != want {
			t.Fatal(text, got, want, err)
		}
	}
	for a := range 256 {
		for _, rgb := range []color.NRGBA{{}, {R: 255, G: 129, B: 7}, {R: 3, G: 250, B: 230}} {
			rgb.A = uint8(a)
			got, err := ParseColorHex(ColorHexString(rgb))
			if err != nil || got != rgb {
				t.Fatal("straight alpha or byte precision lost", rgb, got, err)
			}
		}
	}
	for _, bad := range []string{"", "red", "123456", "#12", "#12345", "#1234567", "#GG112233", "#🙂🙂", "#FFFFFFFFF", strings.Repeat(" ", 33) + "#FFF"} {
		if _, err := ParseColorHex(bad); !errors.Is(err, ErrColorHex) {
			t.Fatal("malformed/budget color accepted", bad, err)
		}
	}
}
func TestColorPickerQueuedChannelEditsUseCurrentAcceptedValue(t *testing.T) {
	p := &ColorPicker{ID: "ink", Label: "Ink", Color: color.NRGBA{R: 1, G: 2, B: 3, A: 0}}
	events := []ColorEdit{}
	p.Change = func(e ColorEdit) { events = append(events, e); p.Color = e.Apply(p.Color) }
	view := p.View()
	if err := Validate(view); err != nil {
		t.Fatal(err)
	}
	// Both callbacks belong to the same old view; preserve independently accepted
	// channels and RGB under zero alpha when the owner drains them before redraw.
	colorNode(view, ColorFieldID(p.ID, ColorRed)).Change(ChangeEvent{Number: 37.6})
	colorNode(view, ColorFieldID(p.ID, ColorGreen)).Change(ChangeEvent{Number: 129})
	if p.Color != (color.NRGBA{R: 38, G: 129, B: 3, A: 0}) || len(events) != 2 {
		t.Fatal("queued edits overwrote accepted channel or hidden RGB", p.Color, events)
	}
	colorNode(view, ColorFieldID(p.ID, ColorHex)).Change(ChangeEvent{Text: "#4567"})
	if p.Color != (color.NRGBA{R: 68, G: 85, B: 102, A: 119}) {
		t.Fatal("whole-color edit not accepted", p.Color)
	}
	if e := (ColorEdit{Channel: "unknown", Byte: 255}); e.Apply(p.Color) != p.Color {
		t.Fatal("unknown edit mutated accepted color")
	}
}
func TestColorPickerInvalidReadonlyFocusAndCollision(t *testing.T) {
	calls, invalid := 0, 0
	p := &ColorPicker{ID: "ink", Color: color.NRGBA{R: 255, A: 255}, Change: func(ColorEdit) { calls++ }, Invalid: func(error) { invalid++ }}
	n := p.View()
	if order := FocusOrder(n); len(order) != 5 || order[0] != ColorFieldID("ink", ColorRed) || order[4] != ColorFieldID("ink", ColorHex) {
		t.Fatal("managed five-field order", order)
	}
	colorNode(n, ColorFieldID("ink", ColorHex)).Change(ChangeEvent{Text: "invalid"})
	for _, number := range []float32{-1, 256, float32(math.NaN()), float32(math.Inf(1))} {
		colorNode(n, ColorFieldID("ink", ColorAlpha)).Change(ChangeEvent{Number: number})
	}
	if calls != 0 || invalid != 5 {
		t.Fatal("invalid input changed model or lost diagnostics", calls, invalid)
	}
	p.Disabled = true
	colorNode(n, ColorFieldID("ink", ColorRed)).Change(ChangeEvent{Number: 25})
	if calls != 0 || len(FocusOrder(p.View())) != 0 {
		t.Fatal("disabled picker admitted stale callback/focus")
	}
	p.Disabled = false
	p.Change = nil
	if !p.View().Disabled || len(FocusOrder(p.View())) != 0 {
		t.Fatal("readonly picker retained editing")
	}
	collision := Node{ID: "root", Kind: Column, Children: []Node{p.View(), {ID: ColorFieldID("ink", ColorRed), Kind: Label}}}
	if Validate(collision) == nil {
		t.Fatal("caller collision in component namespace accepted")
	}
}
func TestColorPickerPreviewImmutabilityAndReuse(t *testing.T) {
	p := &ColorPicker{ID: "ink", Color: color.NRGBA{R: 255, A: 255}, Change: func(ColorEdit) {}}
	first := colorNode(p.View(), "ink/color/preview").Image
	if first.Size().X != 32 || first.TileCount() != 1 || first.NRGBAAt(0, 0) != (color.NRGBA{R: 255, A: 255}) {
		t.Fatal("opaque swatch", first)
	}
	if next := colorNode(p.View(), "ink/color/preview").Image; next != first {
		t.Fatal("unchanged view replaces immutable source")
	}
	p.Color = color.NRGBA{G: 255, A: 128}
	second := colorNode(p.View(), "ink/color/preview").Image
	if second == first || first.NRGBAAt(0, 0) != (color.NRGBA{R: 255, A: 255}) || second.NRGBAAt(0, 0) != (color.NRGBA{R: 120, G: 248, B: 120, A: 255}) {
		t.Fatal("checkerboard alpha or old source mutated", first.NRGBAAt(0, 0), second.NRGBAAt(0, 0))
	}
	p.Color.A = 0
	transparent := colorNode(p.View(), "ink/color/preview").Image
	if transparent.NRGBAAt(0, 0) != (color.NRGBA{R: 240, G: 240, B: 240, A: 255}) || transparent.NRGBAAt(9, 0) != (color.NRGBA{R: 176, G: 176, B: 176, A: 255}) {
		t.Fatal("zero-alpha checkerboard", transparent.NRGBAAt(0, 0), transparent.NRGBAAt(9, 0))
	}
	p.Color.R = 77
	if next := colorNode(p.View(), "ink/color/preview").Image; next != transparent {
		t.Fatal("invisible hidden-RGB edit replaced identical pixels")
	}
	if p.Color.R != 77 {
		t.Fatal("preview altered hidden RGB")
	}
}
