//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
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
