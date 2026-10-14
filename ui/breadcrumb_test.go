package ui

import (
	"fmt"
	"strings"
	"testing"
)

func breadcrumbFixture() *Breadcrumb {
	return &Breadcrumb{ID: "path", Items: []BreadcrumbItem{{ID: "root", Label: "Home"}, {ID: "a/b", Label: "Folder"}, {ID: "file", Label: "Current"}}}
}
func crumb(t *testing.T, n Node, id string) Node {
	t.Helper()
	for _, c := range n.Children {
		if c.ID == id {
			return c
		}
	}
	t.Fatal("item missing", id)
	return Node{}
}
func TestBreadcrumbControlledPathAndLiteralLabels(t *testing.T) {
	b := breadcrumbFixture()
	b.Items[1].Label = "<b>Folder & stuff</b>"
	var events []BreadcrumbEvent
	b.Navigate = func(e BreadcrumbEvent) { events = append(events, e) }
	n, err := b.View()
	if err != nil {
		t.Fatal(err)
	}
	current := crumb(t, n, BreadcrumbItemID(b.ID, "file"))
	if current.Kind != Label || current.Activate != nil || IsFocusable(current.Kind) || current.Accessibility.Help != "Current location" {
		t.Fatal("current location must be passive", current)
	}
	ancestor := crumb(t, n, BreadcrumbItemID(b.ID, "a/b"))
	if ancestor.Kind != Button || ancestor.Text != b.Items[1].Label {
		t.Fatal("literal ancestor", ancestor)
	}
	ancestor.Activate()
	if len(events) != 1 || events[0] != (BreadcrumbEvent{ID: "path", ItemID: "a/b"}) || len(b.Items) != 3 {
		t.Fatal("proposal mutated accepted path", events, b.Items)
	}
	b.Items = b.Items[:2]
	ancestor.Activate()
	if len(events) != 1 {
		t.Fatal("stale path delivered")
	}
	n, err = b.View()
	if err != nil || crumb(t, n, BreadcrumbItemID(b.ID, "a/b")).Kind != Label {
		t.Fatal("accepted location not current", err)
	}
}
func TestBreadcrumbQueuedGuardsAndLatestCallback(t *testing.T) {
	for _, mutation := range []string{"disabled", "item-disabled", "callback-removed", "root-replaced", "path-reordered", "invalid-path"} {
		t.Run(mutation, func(t *testing.T) {
			b := breadcrumbFixture()
			calls := 0
			b.Navigate = func(BreadcrumbEvent) { calls++ }
			n, _ := b.View()
			activate := crumb(t, n, BreadcrumbItemID(b.ID, "root")).Activate
			switch mutation {
			case "disabled":
				b.Disabled = true
			case "item-disabled":
				b.Items[0].Disabled = true
			case "callback-removed":
				b.Navigate = nil
			case "root-replaced":
				b.ID = "other"
			case "path-reordered":
				b.Items[0], b.Items[1] = b.Items[1], b.Items[0]
			case "invalid-path":
				b.Items[1].Label = " "
			}
			activate()
			if calls != 0 {
				t.Fatal("retired request delivered")
			}
		})
	}
	b := breadcrumbFixture()
	old, newCalls := 0, 0
	b.Navigate = func(BreadcrumbEvent) { old++ }
	n, _ := b.View()
	b.Navigate = func(BreadcrumbEvent) { newCalls++ }
	b.Items[0].Label = "Renamed"
	crumb(t, n, BreadcrumbItemID(b.ID, "root")).Activate()
	if old != 0 || newCalls != 1 {
		t.Fatal("latest callback/label identity", old, newCalls)
	}
}
func TestBreadcrumbItemLimitAccepts64Refuses65ValidItems(t *testing.T) {
	b := &Breadcrumb{ID: "path", Navigate: func(BreadcrumbEvent) {}}
	for i := range MaxBreadcrumbItems {
		b.Items = append(b.Items, BreadcrumbItem{ID: fmt.Sprintf("item-%d", i), Label: fmt.Sprintf("Location %d", i)})
	}
	n, err := b.View()
	if err != nil || len(n.Children) != 2*MaxBreadcrumbItems-1 {
		t.Fatal("64 valid locations refused", err, len(n.Children))
	}
	b.Items = append(b.Items, BreadcrumbItem{ID: "item-64", Label: "Location 64"})
	if _, err := b.View(); err == nil {
		t.Fatal("65 valid locations admitted")
	}
}

func TestBreadcrumbAdmissionDetachedIDsAndReadonly(t *testing.T) {
	b := breadcrumbFixture()
	n, err := b.View()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range n.Children {
		if IsFocusable(c.Kind) || c.Activate != nil {
			t.Fatal("readonly path has action", c)
		}
	}
	b.Items[0].Label = "Changed"
	if crumb(t, n, BreadcrumbItemID(b.ID, "root")).Text != "Home" {
		t.Fatal("borrowed item label")
	}
	for _, id := range []string{"a/b", "a", "b/", "雪", "x?y", "separator/root"} {
		if strings.HasSuffix(BreadcrumbItemID("path", id), "/") {
			t.Fatal("ambiguous key")
		}
	}
	seen := map[string]bool{}
	for _, id := range []string{"a/b", "a", "b/", "雪", "x?y", "separator/root"} {
		key := BreadcrumbItemID("path", id)
		if seen[key] {
			t.Fatal("ID collision")
		}
		seen[key] = true
	}
	for _, mutate := range []func(*Breadcrumb){func(b *Breadcrumb) { b.ID = "" }, func(b *Breadcrumb) { b.ID = "grid:reserved" }, func(b *Breadcrumb) { b.Items = nil }, func(b *Breadcrumb) { b.Items = make([]BreadcrumbItem, MaxBreadcrumbItems+1) }, func(b *Breadcrumb) { b.Items[0].ID = "" }, func(b *Breadcrumb) { b.Items[1].ID = b.Items[0].ID }, func(b *Breadcrumb) { b.Items[0].Label = " " }, func(b *Breadcrumb) { b.Style.Gap = Some(float32(-1)) }} {
		b := breadcrumbFixture()
		mutate(b)
		if _, err := b.View(); err == nil {
			t.Fatal("invalid path admitted", b)
		}
	}
	var nilPath *Breadcrumb
	if _, err := nilPath.View(); err == nil {
		t.Fatal("nil path admitted")
	}
	b = breadcrumbFixture()
	b.Navigate = func(BreadcrumbEvent) {}
	n, _ = b.View()
	root := Node{ID: "outer", Kind: Column, Children: []Node{n, {ID: BreadcrumbItemID(b.ID, "root"), Kind: Label}}}
	if Validate(root) == nil {
		t.Fatal("generated ID collision admitted")
	}
}
