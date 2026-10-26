//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestChecklistSpaceAndDisabledRouting(t *testing.T) {
	calls := 0
	var c *ui.Checklist
	c = &ui.Checklist{ID: "tasks", Items: []ui.ChecklistItem{{ID: "a", Text: "Synthetic task"}}, Change: func(e ui.ChecklistEvent) { calls++; c.Items[0].Completed = e.Checked }}
	m, a := simpleManual(t, manualConfig(t, 1801), func() ui.Node {
		n, e := c.View()
		if e != nil {
			t.Fatal(e)
		}
		n.Focus = ui.FocusRequest{ID: ui.ChecklistCheckID(c.ID, "a"), Sequence: 1}
		return n
	})
	unitSettle(t, m)
	w, _ := imageManualElement(t, m, ui.ChecklistCheckID(c.ID, "a"))
	send := func() {
		t.Helper()
		if e := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeSpacebar, Rune: ' '}); e != nil {
			t.Fatal(e)
		}
		unitSettle(t, m)
	}
	send()
	if calls != 1 || !c.Items[0].Completed {
		t.Fatal(calls)
	}
	a.Do(func() { c.Disabled = true; a.Invalidate("one") })
	unitSettle(t, m)
	send()
	if calls != 1 {
		t.Fatal("disabled checklist changed")
	}
}

func TestChecklistRenameCommitsAfterDraftEditing(t *testing.T) {
	calls, invalid := 0, 0
	var c *ui.Checklist
	c = &ui.Checklist{ID: "edit", Editable: true, Items: []ui.ChecklistItem{{ID: "a", Text: "Synthetic"}}, Change: func(e ui.ChecklistEvent) {
		if e.Action == ui.ChecklistRename {
			calls++
			c.Items[0].Text = e.Text
		}
	}, Invalid: func(error) { invalid++ }}
	m, _ := simpleManual(t, manualConfig(t, 1802), func() ui.Node {
		n, e := c.View()
		if e != nil {
			t.Fatal(e)
		}
		n.Focus = ui.FocusRequest{ID: ui.ChecklistNameID(c.ID, "a"), Sequence: 1}
		return n
	})
	unitSettle(t, m)
	w, _ := imageManualElement(t, m, ui.ChecklistNameID(c.ID, "a"))
	send := func(code key.Codes, char rune, mods key.Modifiers) {
		t.Helper()
		if e := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: code, Rune: char, Modifiers: mods}); e != nil {
			t.Fatal(e)
		}
		unitSettle(t, m)
	}
	send(key.CodeA, 0, 1<<key.Control)
	send(key.CodeBackspace, 0, 0)
	if calls != 0 || invalid != 0 {
		t.Fatal("draft clearing prematurely proposed rename", calls, invalid)
	}
	send(key.CodeN, 'n', 0)
	send(key.CodeReturnEnter, 0, 0)
	if calls != 1 || invalid != 0 || c.Items[0].Text != "n" {
		t.Fatal("rename commit", calls, invalid, c.Items[0].Text)
	}
}
