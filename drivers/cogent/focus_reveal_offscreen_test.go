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

func TestWindowFocusRevealCancelsObsoleteIntentBeforePaint(t *testing.T) {
	for _, change := range []string{"disabled", "hidden", "replacement", "empty", "lost-focus", "invalid"} {
		t.Run(change, func(t *testing.T) {
			expanded, disabled, shown, replacement := false, false, true, false
			focus := ui.FocusRequest{ID: "entry", Sequence: 1}
			view := func() ui.Node {
				entry := ui.Node{ID: "entry", Kind: ui.TextField, Text: "draft", Change: func(ui.ChangeEvent) {}}
				if replacement {
					entry = ui.Node{ID: "entry", Kind: ui.Button, Text: "Replacement", Activate: func() {}}
				}
				height := float32(0)
				if expanded {
					height = 500
				}
				return ui.Node{ID: "root", Kind: ui.Scroll, Grow: true, Focus: focus, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(8))}, Children: []ui.Node{
					{ID: "top", Kind: ui.Button, Text: "Top", Activate: func() {}},
					{ID: "before", Kind: ui.Label, Style: ui.Style{MinSize: ui.Some(ui.Size{Height: height})}},
					{ID: "editor", Kind: ui.Disclosure, Text: "Editor", Checked: shown, Disabled: disabled, Children: []ui.Node{entry}},
					{ID: "after", Kind: ui.Label, Text: "After", Style: ui.Style{MinSize: ui.Some(ui.Size{Height: 500})}},
				}}
			}
			m, a := simpleManual(t, manualConfig(t, 962), view)
			unitSettle(t, m)
			w := m.implementation.find("one")
			host := core.AsFrame(findRecord(w.state.root, "root").widget)
			initial := host.Scrolls[math32.Y].Value
			if err := a.Do(func() { expanded = true; focus.Sequence++; _ = a.Invalidate("one") }); err != nil {
				t.Fatal(err)
			}
			if err := m.implementation.refresh(w); err != nil {
				t.Fatal(err)
			}
			pending := w.state.pendingFocusReveal.Load()
			if pending == nil {
				t.Fatal("fixture did not create pending reveal")
			}
			if err := a.Do(func() {
				switch change {
				case "disabled":
					disabled = true
				case "hidden":
					shown = false
				case "replacement":
					replacement = true
				case "empty":
					focus.ID = ""
				case "invalid":
					// Exercise the request validator below without poisoning
					// this runner, whose invalid-descriptor failure is terminal.
				}
				_ = a.Invalidate("one")
			}); err != nil {
				t.Fatal(err)
			}
			if change == "invalid" {
				var validationErr error
				if err := withWindowOwner(a, w.state.window.Ref(), func() {
					w.state.rootMu.Lock()
					defer w.state.rootMu.Unlock()
					validationErr = applyWindowFocus(w.state, ui.FocusRequest{ID: "missing", Sequence: focus.Sequence + 1})
				}); err != nil {
					t.Fatal(err)
				}
				if validationErr == nil {
					t.Fatal("invalid request accepted")
				}
				if w.state.pendingFocusReveal.Load() != nil {
					t.Fatal("failed request retained old intent")
				}
			} else if err := m.implementation.refresh(w); err != nil {
				t.Fatal(err)
			}

			if change == "replacement" {
				next := findRecord(w.state.root, "entry").widget
				if next == pending.target {
					t.Fatal("fixture did not replace native identity")
				}
				// Native focus repair may choose the new same-ID action. It must not
				// inherit the old editor's pending visibility request.
				next.AsWidget().SetFocus()
				if !next.AsWidget().ContainsFocus() {
					t.Fatal("replacement focus fixture failed")
				}
			}
			if change == "lost-focus" {
				findRecord(w.state.root, "top").widget.AsWidget().SetFocus()
			}
			unitSettle(t, m)
			if w.state.pendingFocusReveal.Load() != nil {
				t.Fatal("obsolete intent retained after paint")
			}
			if host.Scrolls[math32.Y].Value != initial {
				t.Fatal("obsolete target scrolled viewport", initial, host.Scrolls[math32.Y].Value)
			}
			if change == "replacement" {
				next := findRecord(w.state.root, "entry").widget.AsWidget()
				if !next.ContainsFocus() || !next.Geom.TotalBBox.Empty() {
					t.Fatal("old intent affected focused same-ID replacement", next.Geom.TotalBBox)
				}
			}
		})
	}
}

func TestWindowFocusRevealLatestRequestWinsBeforePaint(t *testing.T) {
	focus := ui.FocusRequest{ID: "first", Sequence: 1}
	view := func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Scroll, Grow: true, Focus: focus, Children: []ui.Node{
			{ID: "first", Kind: ui.TextField, Text: "First", Change: func(ui.ChangeEvent) {}},
			{ID: "gap", Kind: ui.Label, Text: "Gap", Style: ui.Style{MinSize: ui.Some(ui.Size{Height: 500})}},
			{ID: "last", Kind: ui.TextField, Text: "Last", Change: func(ui.ChangeEvent) {}},
			{ID: "after", Kind: ui.Label, Text: "After", Style: ui.Style{MinSize: ui.Some(ui.Size{Height: 500})}},
		}}
	}
	m, a := simpleManual(t, manualConfig(t, 963), view)
	unitSettle(t, m)
	w := m.implementation.find("one")
	var previous *focusRevealRequest
	for _, id := range []string{"last", "first", "last"} {
		if err := a.Do(func() { focus = ui.FocusRequest{ID: id, Sequence: focus.Sequence + 1}; _ = a.Invalidate("one") }); err != nil {
			t.Fatal(err)
		}
		if err := m.implementation.refresh(w); err != nil {
			t.Fatal(err)
		}
		pending := w.state.pendingFocusReveal.Load()
		if pending == nil || pending == previous || pending.request != focus {
			t.Fatal("request did not replace one pending intent", pending)
		}
		previous = pending
	}
	unitSettle(t, m)
	if w.state.pendingFocusReveal.Load() != nil {
		t.Fatal("latest request remained pending")
	}
	_, last := imageManualElement(t, m, "last")
	host := findRecord(w.state.root, "root").widget.AsWidget()
	field := findRecord(w.state.root, "last").widget.AsWidget()
	raw := math32.RectFromPosSizeMax(field.Geom.Pos.Total, field.Geom.Size.Actual.Total)
	if !last.Focused || raw.Empty() || !raw.In(host.Geom.ContentBBox) {
		t.Fatal("latest focus request not visible", last, raw, host.Geom.ContentBBox)
	}
}

func TestWindowFocusRevealRemainsOnOriginatingWindow(t *testing.T) {
	expanded := false
	focus := ui.FocusRequest{ID: "entry", Sequence: 1}
	view := func() ui.Node {
		height := float32(0)
		if expanded {
			height = 500
		}
		return ui.Node{ID: "root", Kind: ui.Scroll, Grow: true, Focus: focus, Children: []ui.Node{
			{ID: "before", Kind: ui.Label, Style: ui.Style{MinSize: ui.Some(ui.Size{Height: height})}},
			{ID: "entry", Kind: ui.TextField, Text: "Origin", Change: func(ui.ChangeEvent) {}},
			{ID: "after", Kind: ui.Label, Text: "After", Style: ui.Style{MinSize: ui.Some(ui.Size{Height: 500})}},
		}}
	}
	m, a := simpleManual(t, manualConfig(t, 964), view)
	unitSettle(t, m)
	if err := a.Open(app.Window{ID: "two", Title: "Second", Width: 400, Height: 240, View: func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Scroll, Grow: true, Focus: ui.FocusRequest{ID: "entry", Sequence: 1}, Children: []ui.Node{
			{ID: "entry", Kind: ui.TextField, Text: "Second", Change: func(ui.ChangeEvent) {}},
			{ID: "after", Kind: ui.Label, Text: "After", Style: ui.Style{MinSize: ui.Some(ui.Size{Height: 500})}},
		}}
	}}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	origin, other := m.implementation.find("one"), m.implementation.find("two")
	otherHost := core.AsFrame(findRecord(other.state.root, "root").widget)
	otherScroll := otherHost.Scrolls[math32.Y].Value
	if err := a.Do(func() { expanded = true; focus.Sequence++; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	if err := m.implementation.refresh(origin); err != nil {
		t.Fatal(err)
	}
	pending := origin.state.pendingFocusReveal.Load()
	if pending == nil {
		t.Fatal("origin has no pending intent")
	}
	painted := false
	for range 16 {
		step, err := m.Frame(testContext, "two")
		if err != nil {
			t.Fatal(err)
		}
		if origin.state.pendingFocusReveal.Load() != pending {
			t.Fatal("foreign frame consumed originating intent")
		}
		if step.Frames > 0 {
			painted = true
			break
		}
	}
	if !painted {
		t.Fatal("foreign frame bound exhausted")
	}
	if other.state.pendingFocusReveal.Load() != nil || otherHost.Scrolls[math32.Y].Value != otherScroll {
		t.Fatal("originating intent changed foreign scroll state")
	}
	unitSettle(t, m)
	if origin.state.pendingFocusReveal.Load() != nil {
		t.Fatal("origin did not consume its own intent")
	}
	host := findRecord(origin.state.root, "root").widget.AsWidget()
	field := findRecord(origin.state.root, "entry").widget.AsWidget()
	raw := math32.RectFromPosSizeMax(field.Geom.Pos.Total, field.Geom.Size.Actual.Total)
	if raw.Empty() || !raw.In(host.Geom.ContentBBox) {
		t.Fatal("originating focus not revealed", raw, host.Geom.ContentBBox)
	}
	// Identical local IDs still refer to separate retained native fields.
	if !findRecord(other.state.root, "entry").widget.AsWidget().ContainsFocus() || otherHost.Scrolls[math32.Y].Value != otherScroll {
		t.Fatal("origin reveal affected second window")
	}
}
