//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"image/color"
	"strings"
	"testing"
)

func TestTokenFieldManagedCommitAndKeyboardRemoval(t *testing.T) {
	p := &ui.TokenField{ID: "tags", Tokens: []ui.Token{{ID: "a", Text: "Design"}}}
	calls := 0
	p.Change = func(e ui.TokenEvent) {
		calls++
		if e.Remove {
			p.Tokens = nil
		} else {
			p.Tokens = append(p.Tokens, ui.Token{ID: "b", Text: e.Text})
		}
	}
	focus := ui.FocusRequest{ID: ui.TokenInputID(p.ID), Sequence: 1}
	view := func() ui.Node {
		n, err := p.View()
		if err != nil {
			panic(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Focus: focus, Children: []ui.Node{n}}
	}
	m, a := simpleManual(t, manualConfig(t, 922), view)
	unitSettle(t, m)
	root := m.implementation.find("one").state.root
	field := findRecord(root, ui.TokenInputID(p.ID)).widget.(*core.TextField)
	field.SetText("Review").SendChange()
	a.Drain()
	unitSettle(t, m)
	if calls != 1 || len(p.Tokens) != 2 || p.Tokens[1].Text != "Review" {
		t.Fatal(calls, p.Tokens)
	}
	if err := a.Do(func() { focus = ui.FocusRequest{ID: ui.TokenRemoveID(p.ID, "a"), Sequence: 2}; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	w, _ := imageManualElement(t, m, ui.TokenRemoveID(p.ID, "a"))
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeReturnEnter}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if calls != 2 || len(p.Tokens) != 0 {
		t.Fatal("keyboard chip action", calls, p.Tokens)
	}
}

func TestTokenFieldManagedQueuedSuggestionRevocation(t *testing.T) {
	p := &ui.TokenField{ID: "tags", Suggestions: []ui.Option{{ID: "work", Label: "Work"}}}
	calls := 0
	p.Change = func(ui.TokenEvent) { calls++ }
	view := func() ui.Node {
		n, err := p.View()
		if err != nil {
			panic(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{n}}
	}
	m, a := simpleManual(t, manualConfig(t, 931), view)
	unitSettle(t, m)
	root := m.implementation.find("one").state.root
	menu := findRecord(root, ui.ComboMenuID("tags/tokens/entry"))
	// Capture the actual mounted descriptor's owner callback, then change the
	// application catalog in the same owner transaction before reconciliation.
	if err := a.Do(func() {
		p.Suggestions = nil
		menu.node.Change(ui.ChangeEvent{ID: menu.node.ID, SelectedID: "work"})
		_ = a.Invalidate("one")
	}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if calls != 0 {
		t.Fatal("revoked suggestion added a token")
	}
}

func TestTokenFieldLongChipKeepsRemovalVisible(t *testing.T) {
	p := &ui.TokenField{ID: "tags", Tokens: []ui.Token{{ID: "long", Text: strings.Repeat("Wide", 60)}}}
	calls := 0
	p.Change = func(e ui.TokenEvent) {
		if e.Remove {
			calls++
			p.Tokens = nil
		}
	}
	view := func() ui.Node {
		n, err := p.View()
		if err != nil {
			panic(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: ui.TokenRemoveID(p.ID, "long"), Sequence: 1}, Children: []ui.Node{n}}
	}
	cfg := manualConfig(t, 948)
	cfg.Theme = ui.MacTheme(false)
	cfg.Theme.Default.Background = ui.Some(color.NRGBA{R: 246, G: 246, B: 248, A: 255})
	cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
	m, _ := simpleManual(t, cfg, view)
	unitSettle(t, m)
	root := m.implementation.find("one").state.root
	chip := findRecord(root, ui.TokenRemoveID("tags", "long"))
	toolbar := findRecord(root, "tags/tokens/chips").widget.AsWidget()
	if chip.widget.AsWidget().Geom.Size.Actual.Total.X > float32(toolbar.Geom.ContentBBox.Dx())+1 || chip.widget.AsWidget().Geom.TotalBBox.Empty() || !chip.widget.AsWidget().Geom.TotalBBox.In(toolbar.Geom.ContentBBox) {
		t.Fatal("long chip clips its removal affordance", chip.widget.AsWidget().Geom.Size.Actual.Total, chip.widget.AsWidget().Geom.TotalBBox, toolbar.Geom.ContentBBox)
	}
	if chip.node.Accessibility.Name != "Remove "+p.Tokens[0].Text {
		t.Fatal("long chip lost full accessible name")
	}
	w, _ := imageManualElement(t, m, chip.node.ID)
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeReturnEnter}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if calls != 1 || len(p.Tokens) != 0 {
		t.Fatal("long chip removal failed", calls, len(p.Tokens))
	}
}
