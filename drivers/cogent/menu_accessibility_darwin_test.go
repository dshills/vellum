//go:build darwin && cgo

package cogent

import (
	"encoding/json"
	"github.com/dshills/vellum/ui"
	"slices"
	"testing"
)

func TestMenuButtonCocoaPayloadIncludesDisabledActionsArray(t *testing.T) {
	r := &record{node: ui.Node{ID: "actions", Kind: ui.MenuButton}}
	r.control = &controlState{menu: &menuState{options: []ui.Option{{ID: "blocked", Label: "Blocked", Disabled: true}, {ID: "run", Label: "Run"}}, popup: &menuPopup{}}}
	r.control.menu.popup.options = slices.Clone(r.control.menu.options)
	nodes, _ := menuAccessibility(r, "", nil)
	semantic, err := accessibilitySemantic(nodes)
	if err != nil || !nativeAccessibilitySnapshotValid(accessibilityPayload(1, semantic)) {
		t.Fatal("popup Cocoa payload rejected", err, semantic)
	}
	if len(nodes) != 2 {
		t.Fatal("expected current popup options", nodes)
	}
	nodes[0].Actions = nil
	semantic, err = accessibilitySemantic(nodes)
	if err != nil || nativeAccessibilitySnapshotValid(accessibilityPayload(1, semantic)) {
		t.Fatal("negative null-action control unexpectedly passed", err)
	}
}

func TestMenuCheckActionCocoaInertSelectors(t *testing.T) {
	r := &record{node: ui.Node{ID: "view", Kind: ui.MenuButton}}
	r.control = &controlState{menu: &menuState{options: []ui.Option{
		{ID: "checked", Label: "Checked", Checkable: true, Checked: true},
		{ID: "divider", Separator: true},
		{ID: "unchecked", Label: "Unchecked", Checkable: true},
		{ID: "disabled", Label: "Disabled", Checkable: true, Checked: true, Disabled: true},
		{ID: "plain", Label: "Plain"},
	}, popup: &menuPopup{}}}
	r.control.menu.popup.options = slices.Clone(r.control.menu.options)
	nodes, _ := menuAccessibility(r, "", nil)
	semantic, err := accessibilitySemantic(nodes)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Elements []struct {
			Role         string
			Value        any
			Description  string
			Enabled      bool
			PressAllowed bool `json:"press_allowed"`
		}
	}
	raw := nativeAccessibilityMenuFixture(accessibilityPayload(1, semantic))
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatal(err, raw)
	}
	if len(got.Elements) != 4 {
		t.Fatal("separator in Cocoa action list", raw)
	}
	for i, expected := range []struct {
		value       any
		description string
		enabled     bool
	}{{true, "Checked", true}, {false, "Unchecked", true}, {true, "Checked", false}, {nil, "", true}} {
		e := got.Elements[i]
		if e.Role != "AXMenuItem" || e.Value != expected.value || e.Description != expected.description || e.Enabled != expected.enabled || e.PressAllowed != expected.enabled {
			t.Fatal("Cocoa check selector mismatch", i, raw)
		}
	}
}
