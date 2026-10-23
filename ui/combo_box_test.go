package ui

import "testing"

func TestComboBoxControlledAndCurrentOptions(t *testing.T) {
	c := &ComboBox{ID: "units", Text: "Custom", Options: []Option{{ID: "mm", Label: "Millimeters"}, {ID: "in", Label: "Inches", Disabled: true}}}
	calls := 0
	c.Change = func(e ComboEvent) { calls++ }
	n, err := c.View()
	if err != nil {
		t.Fatal(err)
	}
	field := helperNode(t, n, ComboInputID(c.ID))
	menu := helperNode(t, n, ComboMenuID(c.ID))
	field.Change(ChangeEvent{Text: "Café"})
	menu.Change(ChangeEvent{SelectedID: "mm"})
	menu.Change(ChangeEvent{SelectedID: "in"})
	if calls != 2 || c.Text != "Custom" {
		t.Fatal("controlled refusal", calls, c.Text)
	}
	c.Change = func(e ComboEvent) { c.Text = e.Text }
	c.Options[0].Label = "mm"
	menu.Change(ChangeEvent{SelectedID: "mm"})
	if c.Text != "Custom" {
		t.Fatal("changed option admitted an older rendered callback")
	}
	n, err = c.View()
	if err != nil {
		t.Fatal(err)
	}
	helperNode(t, n, ComboMenuID(c.ID)).Change(ChangeEvent{SelectedID: "mm"})
	if c.Text != "mm" {
		t.Fatal("new option not admitted")
	}
	c.Options = nil
	menu.Change(ChangeEvent{SelectedID: "mm"})
	if c.Text != "mm" {
		t.Fatal("removed suggestion")
	}
	c.Disabled = true
	field.Change(ChangeEvent{Text: "bad"})
	c.Disabled = false
	c.ID = "new"
	field.Change(ChangeEvent{Text: "bad"})
	if c.Text != "mm" {
		t.Fatal("obsolete callback")
	}
}
func TestComboBoxValidation(t *testing.T) {
	for _, c := range []*ComboBox{nil, {}, {ID: "a", Options: []Option{{ID: "a", Label: "A"}, {ID: "a", Label: "B"}}}, {ID: "a", Options: []Option{{ID: "a", Label: "A", Checkable: true}}}} {
		if _, err := c.View(); err == nil {
			t.Fatal("invalid combo accepted")
		}
	}
	c := &ComboBox{ID: "empty"}
	n, err := c.View()
	if err != nil || len(n.Children) != 1 || !n.Disabled {
		t.Fatal(n, err)
	}
}

func TestComboMenuUsesOuterFieldAnchor(t *testing.T) {
	c := &ComboBox{ID: "combo", Options: []Option{{ID: "a", Label: "A"}}, Change: func(ComboEvent) {}}
	n, err := c.View()
	if err != nil {
		t.Fatal(err)
	}
	if helperNode(t, n, ComboMenuID(c.ID)).MenuAnchorID != c.ID {
		t.Fatal("menu anchor is not full field")
	}
	if Validate(Node{ID: "label", Kind: Label, MenuAnchorID: "other"}) == nil {
		t.Fatal("anchor accepted outside menu")
	}
}
