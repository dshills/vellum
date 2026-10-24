package ui

import (
	"math"
	"testing"
)

func ruleTestNode(n Node, id string) Node {
	if n.ID == id {
		return n
	}
	for _, c := range n.Children {
		if found := ruleTestNode(c, id); found.ID != "" {
			return found
		}
	}
	return Node{}
}
func ruleTestEditor() *RuleEditor {
	return &RuleEditor{ID: "filter", Fields: []RuleField{{ID: "name", Label: "Name", Kind: RuleText}, {ID: "size", Label: "Size", Kind: RuleNumber}, {ID: "visible", Label: "Visible", Kind: RuleBoolean}}, Root: Rule{ID: "root", Group: &RuleGroup{Match: RuleAll, Rules: []Rule{{ID: "r1", FieldID: "name", Operator: RuleContains, Value: RuleValue{Text: "Drawing"}}}}}}
}
func TestRuleEditorTypedProposalsAndHeldValues(t *testing.T) {
	e := ruleTestEditor()
	calls := 0
	e.Change = func(c RuleEditEvent) { calls++; e.Root = c.Root }
	n, err := e.View()
	if err != nil {
		t.Fatal(err)
	}
	value := ruleTestNode(n, RuleControlID(e.ID, "r1", "value"))
	value.Change(ChangeEvent{Text: "A"})
	value.Change(ChangeEvent{Text: "AB"})
	if calls != 2 || e.Root.Group.Rules[0].Value.Text != "AB" {
		t.Fatal("held absolute edits dropped")
	}
	ruleTestNode(n, RuleControlID(e.ID, "r1", "field")).Change(ChangeEvent{SelectedID: "size"})
	value.Change(ChangeEvent{Text: "stale"})
	if calls != 3 || e.Root.Group.Rules[0].Value.Text != "" {
		t.Fatal("typed transition admitted stale text")
	}
	n, err = e.View()
	if err != nil {
		t.Fatal(err)
	}
	number := ruleTestNode(n, RuleControlID(e.ID, "r1", "value"))
	number.Change(ChangeEvent{Number: float32(math.NaN())})
	if calls != 3 {
		t.Fatal("nonfinite proposal accepted")
	}
	number.Change(ChangeEvent{Number: 25})
	if calls != 4 {
		t.Fatal("number rejected")
	}
	remove := ruleTestNode(n, RuleControlID(e.ID, "r1", "remove"))
	remove.Activate()
	number.Change(ChangeEvent{Number: 99})
	if calls != 5 || len(e.Root.Group.Rules) != 0 {
		t.Fatal("removed rule acted")
	}
}
func TestRuleEditorGroupsBoundsAndDetachedRejection(t *testing.T) {
	e := ruleTestEditor()
	n, err := e.View()
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	e.Change = func(c RuleEditEvent) { calls++; c.Root.Group.Rules[0].Value.Text = "Only candidate" }
	ruleTestNode(n, RuleControlID(e.ID, "root", "add-group")).Activate()
	if calls != 1 || len(e.Root.Group.Rules) != 1 || e.Root.Group.Rules[0].Value.Text != "Drawing" {
		t.Fatal("rejected proposal mutated source")
	}
	e.Root.Group.Rules[0].Operator = RuleGreater
	if e.Validate() == nil {
		t.Fatal("incompatible text operator")
	}
	e = ruleTestEditor()
	e.Root.Group.Rules = append(e.Root.Group.Rules, e.Root.Group.Rules[0])
	if e.Validate() == nil {
		t.Fatal("duplicate rules")
	}
	e = ruleTestEditor()
	e.Root.Group.Rules = []Rule{e.Root}
	if e.Validate() == nil {
		t.Fatal("cyclic/deep rules")
	}
}

func TestRuleEditorNeverReusesRemovedProposalIDs(t *testing.T) {
	e := ruleTestEditor()
	calls := 0
	e.Change = func(c RuleEditEvent) { calls++; e.Root = c.Root }
	n, _ := e.View()
	add := ruleTestNode(n, RuleControlID(e.ID, "root", "add-rule"))
	add.Activate()
	n, _ = e.View()
	staleRemove := ruleTestNode(n, RuleControlID(e.ID, "rule-1", "remove"))
	staleRemove.Activate()
	add.Activate()
	if len(e.Root.Group.Rules) != 2 || e.Root.Group.Rules[1].ID != "rule-2" {
		t.Fatal("ID reused", e.Root)
	}
	staleRemove.Activate()
	if calls != 3 || len(e.Root.Group.Rules) != 2 {
		t.Fatal("stale removal hit replacement")
	}
	e.nextID = ^uint64(0)
	add.Activate()
	if calls != 3 {
		t.Fatal("exhausted identity admitted")
	}
}

func TestRuleEditorFieldRemovalRevokesAdd(t *testing.T) {
	e := ruleTestEditor()
	calls := 0
	e.Change = func(RuleEditEvent) { calls++ }
	n, _ := e.View()
	add := ruleTestNode(n, RuleControlID(e.ID, "root", "add-rule"))
	e.Fields = nil
	add.Activate()
	if calls != 0 {
		t.Fatal("removed fields admitted addition")
	}
}
