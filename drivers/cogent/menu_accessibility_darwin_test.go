//go:build darwin && cgo

package cogent

import (
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestMenuButtonCocoaPayloadIncludesDisabledActionsArray(t *testing.T) {
	r := &record{node: ui.Node{ID: "actions", Kind: ui.MenuButton}}
	r.control = &controlState{menu: &menuState{options: []ui.Option{{ID: "blocked", Label: "Blocked", Disabled: true}, {ID: "run", Label: "Run"}}, popup: &menuPopup{}}}
	nodes, _ := menuAccessibility(r, "", nil)
	semantic, err := accessibilitySemantic(nodes)
	if err != nil || !nativeAccessibilitySnapshotValid(accessibilityPayload(1, semantic)) {
		t.Fatal("popup Cocoa payload rejected", err, semantic)
	}
	nodes[0].Actions = nil
	semantic, err = accessibilitySemantic(nodes)
	if err != nil || nativeAccessibilitySnapshotValid(accessibilityPayload(1, semantic)) {
		t.Fatal("negative null-action control unexpectedly passed", err)
	}
}
