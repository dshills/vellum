package ui

import (
	"strings"
	"testing"
)

func TestMenuButtonValidationFocusAndActionContract(t *testing.T) {
	n := Node{ID: "actions", Kind: MenuButton, Text: "Actions", Options: []Option{{ID: "a", Label: "Same"}, {ID: "b", Label: "Same"}}, Change: func(ChangeEvent) {}}
	if err := Validate(n); err != nil {
		t.Fatal(err)
	}
	if !MenuEnabled(n) {
		t.Fatal("enabled menu lost")
	}
	for _, mutate := range []func(*Node){func(n *Node) { n.SelectedID = "a" }, func(n *Node) { n.Open = true }, func(n *Node) { n.Activate = func() {} }, func(n *Node) { n.Options = nil }, func(n *Node) { n.Options[1].ID = "a" }, func(n *Node) { n.Options[0].Label = "" }, func(n *Node) { n.Options[0].ID = "x\x00" }, func(n *Node) { n.Options[0].Label = strings.Repeat("a", 4097) }, func(n *Node) { n.ID = "menu:reserved" }} {
		bad := n
		bad.Options = append([]Option(nil), n.Options...)
		mutate(&bad)
		if Validate(bad) == nil {
			t.Fatal("invalid menu accepted", bad)
		}
	}
	inactive := n
	inactive.Change = nil
	if _, ok := FocusTarget(inactive, inactive.ID); ok {
		t.Fatal("handlerless menu focus")
	}
	inactive = n
	inactive.Options = []Option{{ID: "a", Label: "Disabled", Disabled: true}}
	if MenuEnabled(inactive) {
		t.Fatal("disabled actions enabled")
	}
	if _, ok := FocusTarget(inactive, inactive.ID); ok {
		t.Fatal("empty focus group")
	}
}

func TestMenuCheckActionsAndSeparatorsStayScoped(t *testing.T) {
	n := Node{ID: "view", Kind: MenuButton, Text: "View", Options: []Option{
		{ID: "enabled", Label: "Enabled", Checkable: true, Checked: true},
		{ID: "divider", Separator: true},
		{ID: "reset", Label: "Reset"},
	}, Change: func(ChangeEvent) {}}
	if err := Validate(n); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*Node){
		func(n *Node) { n.Options[0].Checkable = false },
		func(n *Node) { n.Options[1].Label = "Separator" },
		func(n *Node) { n.Options[1].Checked = true },
		func(n *Node) { n.Options[1].Checkable = true },
		func(n *Node) { n.Options[1].Disabled = true },
		func(n *Node) { n.Options[1].ID = "enabled" },
		func(n *Node) { n.Options[1].ID = "" },
		func(n *Node) { n.Options[1].ID = "bad\x00" },
	} {
		bad := n
		bad.Options = append([]Option(nil), n.Options...)
		mutate(&bad)
		if Validate(bad) == nil {
			t.Fatal("invalid check/separator metadata accepted", bad.Options)
		}
	}
	for _, kind := range []Kind{Selector, Tabs, RadioGroup, SegmentedControl, Label} {
		for _, flag := range []Option{{Separator: true}, {Checkable: true}, {Checked: true}} {
			bad := Node{ID: "other", Kind: kind, Options: []Option{{ID: "a", Label: "A", Separator: flag.Separator, Checkable: flag.Checkable, Checked: flag.Checked}}, SelectedID: "a"}
			if ValidateControl(bad) == nil {
				t.Fatal("menu metadata leaked to other kind", kind, flag)
			}
		}
	}
	for _, options := range [][]Option{
		{{ID: "divider", Separator: true}},
		{{ID: "divider", Separator: true}, {ID: "unavailable", Label: "Unavailable", Checkable: true, Checked: true, Disabled: true}},
	} {
		passive := n
		passive.Options = options
		if err := Validate(passive); err != nil {
			t.Fatal(err)
		}
		if MenuEnabled(passive) {
			t.Fatal("separator counted as enabled action")
		}
		if _, ok := FocusTarget(passive, passive.ID); ok {
			t.Fatal("passive menu entered focus order")
		}
	}
}
