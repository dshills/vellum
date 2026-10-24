package ui

import (
	"fmt"
	"strings"
	"testing"
)

func pathFixture() *PathControl {
	return &PathControl{ID: "path", Items: []PathItem{{ID: "home", Label: "Home"}, {ID: "docs", Label: "Documents"}, {ID: "project", Label: "Project"}, {ID: "drawing", Label: "Drawing.vellum", Kind: PathFile}}}
}
func pathNode(t *testing.T, n Node, id string) Node {
	t.Helper()
	for _, child := range n.Children {
		if child.ID == id {
			return child
		}
	}
	t.Fatalf("missing path node %s", id)
	return Node{}
}
func TestPathControlControlledMenusAndKinds(t *testing.T) {
	p := pathFixture()
	var events []PathEvent
	p.Navigate = func(e PathEvent) { events = append(events, e) }
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	root := pathNode(t, n, PathSegmentID(p.ID, "home"))
	if root.Kind != Button || root.Icon != SymbolFolder {
		t.Fatal("root folder", root)
	}
	current := pathNode(t, n, PathSegmentID(p.ID, "drawing"))
	if current.Kind != MenuButton || current.Icon != SymbolDocument || current.Accessibility.Help != "Current location" || len(current.Options) != 3 || current.Options[0].ID != "project" {
		t.Fatal("current ancestor menu", current)
	}
	current.Change(ChangeEvent{SelectedID: "drawing"})
	if len(events) != 0 {
		t.Fatal("current location navigated")
	}
	ancestor := pathNode(t, n, PathSegmentID(p.ID, "project"))
	if ancestor.Options[0].ID != "project" || ancestor.Options[1].ID != "docs" {
		t.Fatal("ancestor order", ancestor.Options)
	}
	ancestor.Change(ChangeEvent{SelectedID: "project"})
	root.Activate()
	if len(events) != 2 || events[0] != (PathEvent{ID: p.ID, ItemID: "project"}) || events[1].ItemID != "home" || len(p.Items) != 4 {
		t.Fatal("controlled navigation", events, p.Items)
	}
	p.Items = p.Items[:3]
	current.Change(ChangeEvent{SelectedID: "docs"})
	if len(events) != 2 {
		t.Fatal("stale path navigated")
	}
}
func TestPathControlOverflowAndUnicodeTruncation(t *testing.T) {
	p := &PathControl{ID: "path", MaxVisible: 3, MaxLabelRunes: 8, Navigate: func(PathEvent) {}}
	for i := 0; i < MaxPathItems; i++ {
		p.Items = append(p.Items, PathItem{ID: fmt.Sprint(i), Label: fmt.Sprintf("位置名称%d非常长", i)})
	}
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	// Three visible path segments and one overflow, with three separators.
	if len(n.Children) != 7 {
		t.Fatal("unbounded retained strip", len(n.Children))
	}
	overflow := pathNode(t, n, PathOverflowID(p.ID))
	if len(overflow.Options) != 61 || overflow.Options[0].ID != "61" || overflow.Options[60].ID != "1" {
		t.Fatal("hidden ancestors", overflow.Options)
	}
	for _, index := range []int{0, 62, 63} {
		node := pathNode(t, n, PathSegmentID(p.ID, fmt.Sprint(index)))
		if len([]rune(node.Text)) > 8 || node.Accessibility.Name != p.Items[index].Label {
			t.Fatal("text/AX truncation", node)
		}
	}
	text := "a雪b🌤c風d月e火f"
	got := pathShortLabel(text, 8)
	if len([]rune(got)) != 8 || !strings.HasPrefix(got, "a雪b🌤") || !strings.HasSuffix(got, "e火f") {
		t.Fatal("rune truncation", got)
	}
	calls := 0
	p.Navigate = func(e PathEvent) {
		if e.ItemID != "61" {
			t.Fatal(e)
		}
		calls++
	}
	overflow.Change(ChangeEvent{SelectedID: "63"}) // current is not hidden
	overflow.Change(ChangeEvent{SelectedID: "0"})  // root is not hidden
	overflow.Change(ChangeEvent{SelectedID: "61"})
	if calls != 1 {
		t.Fatal("overflow scope", calls)
	}
	p.Items = append(p.Items, PathItem{ID: "65", Label: "Extra"})
	if _, err = p.View(); err == nil {
		t.Fatal("unbounded path admitted")
	}
}
func TestPathControlStaleDisabledAndLatestCallback(t *testing.T) {
	for _, mutation := range []string{"root", "path", "label", "kind", "disabled", "item-disabled", "callback", "invalid"} {
		t.Run(mutation, func(t *testing.T) {
			p := pathFixture()
			calls := 0
			p.Navigate = func(PathEvent) { calls++ }
			n, _ := p.View()
			action := pathNode(t, n, PathSegmentID(p.ID, "drawing")).Change
			switch mutation {
			case "root":
				p.ID = "replacement"
			case "path":
				p.Items[0], p.Items[1] = p.Items[1], p.Items[0]
			case "label":
				p.Items[1].Label = "Renamed"
			case "kind":
				p.Items[3].Kind = PathFolder
			case "disabled":
				p.Disabled = true
			case "item-disabled":
				p.Items[1].Disabled = true
			case "callback":
				p.Navigate = nil
			case "invalid":
				p.Items[0].Label = " "
			}
			action(ChangeEvent{SelectedID: "docs"})
			if calls != 0 {
				t.Fatal("retired path action delivered")
			}
		})
	}
	p := pathFixture()
	old, newCalls := 0, 0
	p.Navigate = func(PathEvent) { old++ }
	n, _ := p.View()
	p.Navigate = func(PathEvent) { newCalls++ }
	pathNode(t, n, PathSegmentID(p.ID, "drawing")).Change(ChangeEvent{SelectedID: "docs"})
	if old != 0 || newCalls != 1 {
		t.Fatal("latest callback not used", old, newCalls)
	}
	p.Items[1].Disabled = true
	n, _ = p.View()
	action := pathNode(t, n, PathSegmentID(p.ID, "drawing")).Change
	action(ChangeEvent{SelectedID: "docs"})
	p.Items[1].Disabled = false
	action(ChangeEvent{SelectedID: "docs"})
	if newCalls != 1 {
		t.Fatal("disabled rendered option delivered")
	}
}
func TestPathControlReadonlyAndAdmission(t *testing.T) {
	p := pathFixture()
	p.MaxVisible = 3
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	if len(FocusOrder(n)) != 0 {
		t.Fatal("read-only path takes focus", FocusOrder(n))
	}
	if n.Accessibility.Name != "Location" {
		t.Fatal("missing accessible label")
	}
	for _, node := range n.Children {
		if node.Change != nil || node.Activate != nil {
			t.Fatal("read-only action", node)
		}
	}
	p.Items = []PathItem{{ID: "one", Label: "Only file", Kind: PathFile}}
	p.Navigate = func(PathEvent) {}
	n, err = p.View()
	if err != nil || len(FocusOrder(n)) != 0 {
		t.Fatal("single-item path navigates", err, n)
	}
	for _, mutate := range []func(*PathControl){
		func(p *PathControl) { p.ID = "" }, func(p *PathControl) { p.ID = "menu:reserved" },
		func(p *PathControl) { p.Items = nil }, func(p *PathControl) { p.Items[0].ID = "" },
		func(p *PathControl) { p.Items[1].ID = p.Items[0].ID }, func(p *PathControl) { p.Items[0].Kind = PathFile },
		func(p *PathControl) { p.Items[0].Kind = "remote" }, func(p *PathControl) { p.Items[0].Label = " " },
		func(p *PathControl) { p.Items[0].Label = "bad\x00" }, func(p *PathControl) { p.Items[0].ID = "bad\xff" },
		func(p *PathControl) { p.Items[0].Label = strings.Repeat("x", 4097) },
		func(p *PathControl) { p.Items[0].ID = strings.Repeat("x", 1025) },
		func(p *PathControl) { p.MaxVisible = 2 }, func(p *PathControl) { p.MaxVisible = 9 },
		func(p *PathControl) { p.MaxLabelRunes = 7 }, func(p *PathControl) { p.MaxLabelRunes = 129 },
		func(p *PathControl) { p.Style.Gap = Some(float32(-1)) },
	} {
		p := pathFixture()
		mutate(p)
		if _, err := p.View(); err == nil {
			t.Fatal("invalid path admitted", p)
		}
	}
	var nilPath *PathControl
	if _, err := nilPath.View(); err == nil {
		t.Fatal("nil path admitted")
	}
	p = pathFixture()
	n, _ = p.View()
	copyLabel := n.Children[0].Children[1].Text
	p.Items[0].Label = "Mutated"
	if copyLabel != "Home" {
		t.Fatal("view borrowed label")
	}
	if PathSegmentID("path", "a/b") == PathSegmentID("path", "a") {
		t.Fatal("unstable path keys")
	}
	collision := Node{ID: "root", Kind: Column, Children: []Node{n, {ID: PathSegmentID("path", "home"), Kind: Label}}}
	if Validate(collision) == nil {
		t.Fatal("generated identity collision admitted")
	}
}

func TestPathControlReadOnlyOverflowExplainsHiddenPath(t *testing.T) {
	p := &PathControl{ID: "path", MaxVisible: 3, Items: []PathItem{{ID: "r", Label: "Root"}, {ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}, {ID: "c", Label: "Current"}}}
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	overflow := helperNode(t, n, PathOverflowID(p.ID))
	if overflow.Accessibility.Help != "" || overflow.Accessibility.Name != "Hidden ancestors: Alpha" {
		t.Fatal(overflow.Accessibility)
	}
}
