//go:build darwin && cgo

package cogent

import (
	"encoding/json"
	"strings"
	"testing"

	"cogentcore.org/core/core"
	"github.com/dshills/vellum/ui"
)

func TestBreadcrumbNativeRolesAndOwnerRetirementWithoutWindow(t *testing.T) {
	a, b := newAXTest(t, 8)
	calls := 0
	path := &ui.Breadcrumb{ID: "path", Label: "Location", Items: []ui.BreadcrumbItem{{ID: "home", Label: "Home"}, {ID: "folder", Label: "Folder"}, {ID: "current", Label: "Current"}}, Navigate: func(ui.BreadcrumbEvent) { calls++ }}
	n, err := path.View()
	if err != nil {
		t.Fatal(err)
	}
	host := core.NewFrame(core.NewBody("inert breadcrumb"))
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { disposeControlTree(r) }()
	nodes, handlers := accessibilitySnapshot(r, b)
	b.update(nodes, handlers)
	payload, err := json.Marshal(struct {
		Generation int                 `json:"generation"`
		Nodes      []accessibilityNode `json:"nodes"`
	}{1, nodes})
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		id, role string
		action   bool
	}{{"home", "AXButton", true}, {"folder", "AXButton", true}, {"current", "AXStaticText", false}} {
		id := accessibilityManagedID(ui.BreadcrumbItemID(path.ID, test.id))
		var observed struct {
			Role string `json:"role"`
		}
		if err = json.Unmarshal([]byte(nativeAccessibilityGridFixture(string(payload), id, 0, 0)), &observed); err != nil || observed.Role != test.role {
			t.Fatal("actual Cocoa role", test, observed, err)
		}
		found := false
		for _, node := range nodes {
			if node.ID == id {
				found = true
				if node.Focusable != test.action || containsAccessibilityAction(node.Actions, "press") != test.action {
					t.Fatal("path action semantics", node)
				}
			}
		}
		if !found {
			t.Fatal("managed path element absent", id)
		}
	}
	for _, node := range nodes {
		if node.Role == string(ui.Icon) || strings.Contains(node.DomainID, "/breadcrumb/separator/") {
			t.Fatal("decorative separator announced", node)
		}
	}
	id := accessibilityManagedID(ui.BreadcrumbItemID(path.ID, "home"))
	if !accessibilityDeliver(b.token, b.generation.Load(), id, accessibilityAction{Kind: "press"}) {
		t.Fatal("owner admission failed")
	}
	path.Items[0], path.Items[1] = path.Items[1], path.Items[0] // Same-sized path changes before admitted owner work drains.
	if err = a.Drain(); err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatal("queued old path navigated replacement")
	}
	n, err = path.View()
	if err != nil {
		t.Fatal(err)
	}
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	nodes, handlers = accessibilitySnapshot(r, b)
	b.update(nodes, handlers)
	if !accessibilityDeliver(b.token, b.generation.Load(), id, accessibilityAction{Kind: "press"}) {
		t.Fatal("new path admission failed")
	}
	if err = a.Drain(); err != nil || calls != 1 {
		t.Fatal("current path navigation", calls, err)
	}
	path.Disabled = true
	n, err = path.View()
	if err != nil {
		t.Fatal(err)
	}
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	nodes, handlers = accessibilitySnapshot(r, b)
	b.update(nodes, handlers)
	if accessibilityDeliver(b.token, b.generation.Load(), id, accessibilityAction{Kind: "press"}) {
		t.Fatal("disabled path admitted")
	}
}
