//go:build offscreen

package cogent

import (
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

func TestWindowFocusRevealPreservesEditingAndManualScroll(t *testing.T) {
	expanded := false
	focus := ui.FocusRequest{ID: "entry", Sequence: 1}
	view := func() ui.Node {
		height := float32(100)
		if expanded {
			height = 300
		}
		return ui.Node{ID: "root", Kind: ui.Scroll, Grow: true, Focus: focus, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(16))}, Children: []ui.Node{
			{ID: "before", Kind: ui.Label, Text: "Content above the editor", Style: ui.Style{MinSize: ui.Some(ui.Size{Height: height})}},
			{ID: "entry", Kind: ui.TextField, Text: "draft", Change: func(ui.ChangeEvent) {}},
			{ID: "after", Kind: ui.Label, Text: "Content below the editor", Style: ui.Style{MinSize: ui.Some(ui.Size{Height: 400})}},
		}}
	}
	m, a := simpleManual(t, manualConfig(t, 960), view)
	unitSettle(t, m)
	w := m.implementation.find("one")
	field := findRecord(w.state.root, "entry").widget.(*core.TextField)
	if !field.SetEditingSelection(1, 4) {
		t.Fatal("fixture selection")
	}
	focusEvents := 0
	field.OnFocus(func(events.Event) { focusEvents++ })
	field.OnFocusLost(func(events.Event) { focusEvents++ })
	if err := a.Do(func() { expanded = true; focus.Sequence++; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	host := findRecord(w.state.root, "root").widget.AsWidget()
	assertVisible := func() {
		t.Helper()
		raw := math32.RectFromPosSizeMax(field.Geom.Pos.Total, field.Geom.Size.Actual.Total)
		if raw.Empty() || !raw.In(host.Geom.ContentBBox) {
			t.Fatal("requested field clipped after layout", raw, host.Geom.ContentBBox)
		}
		edit := field.EditingState()
		if edit.Text != "draft" || edit.Anchor != 1 || edit.Focus != 4 || focusEvents != 0 {
			t.Fatal("reveal changed editing or focus", edit, focusEvents)
		}
		if w.state.pendingFocusReveal.Load() != nil {
			t.Fatal("reveal intent retained after paint")
		}
	}
	assertVisible()
	box := host.Geom.ContentBBox
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.Scroll, Position: box.Min.Add(box.Size().Div(2)), Delta: math32.Vec2(0, -10000)}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if !field.Geom.TotalBBox.Empty() || !field.ContainsFocus() {
		t.Fatal("manual wheel did not move away from retained focus")
	}
	if err := a.Invalidate("one"); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if !field.Geom.TotalBBox.Empty() {
		t.Fatal("persistent request replayed after manual scrolling")
	}
	// A genuinely new request must work even without descriptor/layout changes.
	if err := a.Do(func() { focus.Sequence++; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	assertVisible()
}

func TestWindowFocusRevealCancelsRemovedTargetAndClose(t *testing.T) {
	shown := true
	focus := ui.FocusRequest{ID: "entry", Sequence: 1}
	view := func() ui.Node {
		children := []ui.Node{{ID: "other", Kind: ui.Button, Text: "Other", Activate: func() {}}}
		if shown {
			children = append(children, ui.Node{ID: "entry", Kind: ui.TextField, Text: "draft", Change: func(ui.ChangeEvent) {}})
		}
		return ui.Node{ID: "root", Kind: ui.Column, Focus: focus, Children: children}
	}
	m, a := simpleManual(t, manualConfig(t, 961), view)
	unitSettle(t, m)
	// Keep the application alive so public Step can perform origin retirement;
	// closing the last application window requires explicit Manual.Close.
	if err := a.Open(app.Window{ID: "survivor", Title: "Survivor", Width: 200, Height: 100, View: func() ui.Node { return ui.Node{ID: "survivor-root", Kind: ui.Label, Text: "Survivor"} }}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	w := m.implementation.find("one")
	refresh := func() {
		t.Helper()
		if err := m.implementation.refresh(w); err != nil {
			t.Fatal(err)
		}
	}
	if err := a.Do(func() { focus.Sequence++; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	refresh()
	if w.state.pendingFocusReveal.Load() == nil {
		t.Fatal("focus request did not defer reveal")
	}
	if err := a.Do(func() { shown = false; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	refresh()
	unitSettle(t, m)
	if w.state.pendingFocusReveal.Load() != nil {
		t.Fatal("removed target retained reveal reference")
	}
	if err := a.Do(func() { focus = ui.FocusRequest{ID: "other", Sequence: focus.Sequence + 1}; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	refresh()
	if w.state.pendingFocusReveal.Load() == nil {
		t.Fatal("close fixture has no pending reveal")
	}
	if err := a.CloseWindow(w.state.window.Ref()); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Step(testContext, 7); err != nil {
		t.Fatal(err)
	}
	if w.state.pendingFocusReveal.Load() != nil {
		t.Fatal("window close retained reveal reference")
	}
}
