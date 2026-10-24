package ui

import (
	"reflect"
	"testing"
)

func customizationCatalogue() Node {
	return Node{ID: "tools", Kind: Toolbar, Children: []Node{
		{ID: "a", Kind: Button, Icon: SymbolAdd, Accessibility: Accessibility{Name: "New"}, Activate: func() {}},
		{ID: "rule", Kind: Separator},
		{ID: "b", Kind: Button, Text: "Open", Activate: func() {}},
		{ID: "c", Kind: Button, Text: "Delete", Disabled: true, Activate: func() {}},
	}}
}
func customizationIDs(n Node) []string {
	ids := []string{}
	for _, c := range n.Children {
		ids = append(ids, c.ID)
	}
	return ids
}
func customizationNode(t *testing.T, n Node, id string) Node {
	t.Helper()
	if n.ID == id {
		return n
	}
	for _, child := range n.Children {
		if found, ok := customizationFind(child, id); ok {
			return found
		}
	}
	t.Fatal("missing customization node", id)
	return Node{}
}
func customizationFind(n Node, id string) (Node, bool) {
	if n.ID == id {
		return n, true
	}
	for _, c := range n.Children {
		if found, ok := customizationFind(c, id); ok {
			return found, true
		}
	}
	return Node{}, false
}

func TestToolbarLayoutOrderVisibilityAndNormalization(t *testing.T) {
	source := customizationCatalogue()
	calls := 0
	source.Children[2].Activate = func() { calls++ }
	n, err := (ToolbarLayout{Order: []string{"b", "rule", "c", "a"}, Hidden: []string{"c"}}).Apply(source)
	if err != nil {
		t.Fatal(err)
	}
	if got := customizationIDs(n); !reflect.DeepEqual(got, []string{"b", "rule", "a"}) {
		t.Fatal(got)
	}
	n.Children[0].Activate()
	if calls != 1 || n.Children[2].Icon != SymbolAdd {
		t.Fatal("lost callback or artwork")
	}
	if !reflect.DeepEqual(customizationIDs(source), []string{"a", "rule", "b", "c"}) {
		t.Fatal("mutated source")
	}
	n, err = (ToolbarLayout{Hidden: []string{"a", "b", "c"}}).Apply(source)
	if err != nil || len(n.Children) != 0 {
		t.Fatal("all-hidden toolbar not empty", n, err)
	}
	n, err = (ToolbarLayout{Order: []string{"c"}}).Apply(source)
	if err != nil || !reflect.DeepEqual(customizationIDs(n), []string{"c", "a", "rule", "b"}) {
		t.Fatal("omitted catalogue items not appended")
	}
	for _, layout := range []ToolbarLayout{{Order: []string{"missing"}}, {Hidden: []string{"missing"}}, {Order: []string{"a", "a"}}, {Hidden: []string{"a", "a"}}} {
		if _, e := layout.Apply(source); e == nil {
			t.Fatal("invalid configuration accepted", layout)
		}
	}
	if _, e := (ToolbarLayout{}).Apply(WithToolbarOverflow(source)); e == nil {
		t.Fatal("overflow applied before customization")
	}
	source.Orientation = Vertical
	if _, e := (ToolbarLayout{}).Apply(source); e == nil {
		t.Fatal("vertical catalogue accepted")
	}
}

func TestToolbarCustomizerControlledEditsSessionsAndCurrentOrder(t *testing.T) {
	changes, dismissals := 0, 0
	c := &ToolbarCustomizer{ID: "customize", Toolbar: customizationCatalogue(), Open: true, Dismiss: func() { dismissals++ }}
	c.Change = func(value ToolbarLayout) { changes++; c.Value = value }
	view := func() Node {
		t.Helper()
		n, e := c.View()
		if e != nil {
			t.Fatal(e)
		}
		return n
	}
	n := view()
	earlier := customizationNode(t, n, ToolbarCustomizationEarlierID(c.ID, "c"))
	earlier.Activate()
	earlier.Activate()
	if !reflect.DeepEqual(c.Value.Order, []string{"a", "c", "rule", "b"}) || changes != 2 {
		t.Fatal("queued moves did not merge latest accepted layout", c.Value, changes)
	}
	customizationNode(t, n, ToolbarCustomizationVisibilityID(c.ID, "a")).Change(ChangeEvent{Checked: false})
	customizationNode(t, n, ToolbarCustomizationVisibilityID(c.ID, "b")).Change(ChangeEvent{Checked: false})
	if !reflect.DeepEqual(c.Value.Hidden, []string{"a", "b"}) {
		t.Fatal("queued visibility edits lost accepted settings", c.Value)
	}
	before := changes
	customizationNode(t, n, ToolbarCustomizationVisibilityID(c.ID, "a")).Change(ChangeEvent{Checked: false})
	if changes != before {
		t.Fatal("duplicate visibility request emitted")
	}
	customizationNode(t, n, ToolbarCustomizationDoneID(c.ID)).Activate()
	earlier.Activate()
	if changes != before || dismissals != 1 {
		t.Fatal("closed-session callback admitted")
	}
	c.Open = false
	view()
	c.Open = true
	n = view()
	earlier.Activate()
	if changes != before {
		t.Fatal("previous session revived")
	}
	customizationNode(t, n, ToolbarCustomizationRestoreID(c.ID)).Activate()
	if !reflect.DeepEqual(c.Value.Order, []string{"a", "rule", "b", "c"}) || len(c.Value.Hidden) != 0 {
		t.Fatal("restore did not reset order and visibility", c.Value)
	}
	c.Disabled = true
	customizationNode(t, n, ToolbarCustomizationVisibilityID(c.ID, "a")).Change(ChangeEvent{Checked: false})
	if changes != before+1 {
		t.Fatal("disabled customization edited")
	}
	c.Disabled = false
	c.Open = false
	view()
	c.Open = true
	n = view()
	stale := customizationNode(t, n, ToolbarCustomizationLaterID(c.ID, "a"))
	c.Toolbar.Children = c.Toolbar.Children[1:]
	c.Value = ToolbarLayout{}
	stale.Activate()
	if changes != before+1 {
		t.Fatal("removed action stale callback edited")
	}
}

func TestToolbarCustomizerRejectedChangeAndInvalidReplacement(t *testing.T) {
	c := &ToolbarCustomizer{ID: "customize", Toolbar: customizationCatalogue(), Open: true, Dismiss: func() {}, Change: func(ToolbarLayout) {}}
	n, e := c.View()
	if e != nil {
		t.Fatal(e)
	}
	customizationNode(t, n, ToolbarCustomizationVisibilityID(c.ID, "a")).Change(ChangeEvent{Checked: false})
	n, e = c.View()
	if e != nil || !customizationNode(t, n, ToolbarCustomizationVisibilityID(c.ID, "a")).Checked {
		t.Fatal("rejected setting became accepted")
	}
	old := customizationNode(t, n, ToolbarCustomizationLaterID(c.ID, "a"))
	changes := 0
	c.Change = func(ToolbarLayout) { changes++ }
	c.Value.Hidden = []string{"unknown"}
	if _, e = c.View(); e == nil {
		t.Fatal("invalid replacement accepted")
	}
	c.Value = ToolbarLayout{}
	c.View()
	old.Activate()
	if changes != 0 {
		t.Fatal("invalid replacement did not revoke callbacks")
	}
}
