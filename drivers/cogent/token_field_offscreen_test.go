//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"fmt"
	"github.com/dshills/vellum/app"
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

// Exercise a real overflowing chip toolbar at the public token bound. The
// consumer owns post-removal focus; the component only proposes one edit.
func TestTokenFieldFullToolbarKeyboardRemoval(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, dpi := range []float32{160, 320} {
			t.Run(fmt.Sprintf("dark-%v/dpi-%.0f", dark, dpi), func(t *testing.T) {
				p := &ui.TokenField{ID: "tags"}
				for i := range ui.MaxTokens {
					p.Tokens = append(p.Tokens, ui.Token{ID: fmt.Sprintf("token-%d", i), Text: fmt.Sprintf("%d %s", i, strings.Repeat("Segment😀/", 12))})
				}
				first, last := ui.TokenRemoveID(p.ID, p.Tokens[0].ID), ui.TokenRemoveID(p.ID, p.Tokens[len(p.Tokens)-1].ID)
				fullLast := p.Tokens[len(p.Tokens)-1]
				focus := ui.FocusRequest{ID: first, Sequence: 1}
				var edits []ui.TokenEvent
				p.Change = func(e ui.TokenEvent) {
					edits = append(edits, e)
					if !e.Remove {
						return
					}
					for i, tok := range p.Tokens {
						if tok.ID == e.TokenID {
							p.Tokens = append(p.Tokens[:i], p.Tokens[i+1:]...)
							focus = ui.FocusRequest{ID: ui.TokenInputID(p.ID), Sequence: focus.Sequence + 1}
							return
						}
					}
				}
				view := func() ui.Node {
					n, err := p.View()
					if err != nil {
						panic(err)
					}
					return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: focus, Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 320})}, Children: []ui.Node{n}}
				}
				cfg := manualConfig(t, 958)
				cfg.Theme = ui.MacTheme(dark)
				cfg.Theme.Default.Background = ui.Some(color.NRGBA{R: 246, G: 246, B: 248, A: 255})
				if dark {
					cfg.Theme.Default.Background = ui.Some(color.NRGBA{R: 29, G: 29, B: 31, A: 255})
				}
				cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
				cfg.LogicalDPI = dpi
				// Scale the software window together with DPI, retaining a 320dp
				// content region. simpleManual's fixed 400px window is narrower
				// in logical units at DPI320 and is not a matched-scale fixture.
				a, err := app.New(app.Config{QueueCapacity: 16})
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(a.Close)
				scale := dpi / 160
				if err = a.Open(app.Window{ID: "one", Title: "Full tag toolbar", Width: 400 * scale, Height: 240 * scale, View: view}); err != nil {
					t.Fatal(err)
				}
				m, err := NewManual(testContext, a, cfg)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := m.Close(testContext); err != nil {
						t.Error(err)
					}
				})
				unitSettle(t, m)
				w, entry := imageManualElement(t, m, ui.TokenInputID(p.ID))
				if entry.Enabled {
					t.Fatal("full token field admitted entry")
				}
				toolbar := findRecord(m.implementation.find("one").state.root, "tags/tokens/chips").control.chrome.host
				if toolbar.Scrolls[math32.X] == nil {
					t.Fatal("full chip toolbar has no overflow")
				}
				send := func(code key.Codes) {
					t.Helper()
					if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: code}); err != nil {
						t.Fatal(err)
					}
					unitSettle(t, m)
				}
				send(key.CodeEnd)
				_, chip := imageManualElement(t, m, last)
				lastRecord := findRecord(m.implementation.find("one").state.root, last)
				box := lastRecord.widget.AsWidget().Geom.TotalBBox
				if !chip.Focused || toolbar.Scrolls[math32.X].Value <= 0 || box.Empty() || !box.In(toolbar.Geom.ContentBBox) {
					t.Fatal("last chip not fully revealed", chip, box, toolbar.Geom.ContentBBox)
				}
				if lastRecord.node.Accessibility.Name != "Remove "+fullLast.Text {
					t.Fatal("full label lost")
				}
				send(key.CodeReturnEnter)
				if len(edits) != 1 || !edits[0].Remove || edits[0].TokenID != fullLast.ID || edits[0].Text != fullLast.Text || len(p.Tokens) != ui.MaxTokens-1 {
					t.Fatal("keyboard removal did not deliver one full-value edit", edits, len(p.Tokens))
				}
				_, entry = imageManualElement(t, m, ui.TokenInputID(p.ID))
				if !entry.Enabled || !entry.Focused {
					t.Fatal("consumer focus did not reach re-enabled entry", entry)
				}
				// A settling frame must not redeliver the removal after its target vanished.
				unitSettle(t, m)
				if len(edits) != 1 {
					t.Fatal("removal redelivered", edits)
				}
			})
		}
	}
}
