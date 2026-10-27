package ui

import (
	"github.com/dshills/vellum/canvas"
	"image"
	"image/color"
	"testing"
)

func nodeByID(n Node, id string) *Node {
	if n.ID == id {
		return &n
	}
	for _, c := range n.Children {
		if f := nodeByID(c, id); f != nil {
			return f
		}
	}
	return nil
}
func TestRibbonGroupedMenusContextAndCurrentAvailability(t *testing.T) {
	pixels := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	pixels.SetNRGBA(0, 0, color.NRGBA{255, 30, 10, 255})
	art, _ := canvas.NewTiledImage(pixels)
	var events []RibbonEvent
	r := &Ribbon{ID: "r", Label: "Tools", SelectedTab: "solid", Action: func(e RibbonEvent) { events = append(events, e) }, Tabs: []RibbonTab{{ID: "solid", Label: "Solid", Groups: []RibbonGroup{{ID: "create", Label: "Create", Actions: []RibbonAction{{ID: "box", Label: "Box", IconImage: art, Menu: []Option{{ID: "cube", Label: "Cube", IconImage: art}}}}}}}, {ID: "sketch", Label: "Sketch", Contextual: true, Groups: []RibbonGroup{{ID: "draw", Label: "Draw", Actions: []RibbonAction{{ID: "line", Label: "Line"}}}}}}}
	n, e := r.View()
	if e != nil {
		t.Fatal(e)
	}
	a := nodeByID(n, RibbonActionID("r", "solid", "box"))
	if a == nil || a.IconImage != art {
		t.Fatal("artwork")
	}
	a.Activate()
	if len(events) != 1 || events[0].ActionID != "box" {
		t.Fatal(events)
	}
	r.Disabled = true
	a.Activate()
	r.Disabled = false
	r.SelectedTab = "sketch"
	a.Activate()
	if len(events) != 1 {
		t.Fatal("stale context action")
	}
	r.SelectedTab = "solid"
	n2, e := r.View()
	if e != nil {
		t.Fatal(e)
	}
	bar := n.Children[0]
	bar2 := n2.Children[0]
	if bar.Options != nil || !bar.ToolbarGroups || bar.Children[len(bar.Children)-1].Options[0].Branch != bar2.Children[len(bar2.Children)-1].Options[0].Branch {
		t.Fatal("submenu identity drift")
	}
	r.Tabs[0].Groups[0].Actions[0].Menu = nil
	r.dispatch("solid", "box", "cube")
	if len(events) != 1 {
		t.Fatal("removed option ran")
	}
}
func TestModelBrowserIndependentActionsCollapseAndRetirement(t *testing.T) {
	var got []ModelBrowserEvent
	b := &ModelBrowser{ID: "b", Label: "Browser", Items: []ModelBrowserItem{{ID: "root", Label: "Assembly", Visible: true}, {ID: "part", Parent: "root", Label: "Part", Visible: true, Activatable: true}}, ExpandedIDs: []string{"root"}, Change: func(e ModelBrowserEvent) { got = append(got, e) }}
	n, e := b.View()
	if e != nil {
		t.Fatal(e)
	}
	visibility := nodeByID(n, ModelBrowserRowID("b", "part")+"/visibility")
	selectButton := nodeByID(n, ModelBrowserRowID("b", "part")+"/select")
	visibility.Change(ChangeEvent{Checked: false})
	if len(got) != 1 || got[0].Kind != ModelBrowserVisibility {
		t.Fatal(got)
	}
	selectButton.Activate()
	if got[1].Kind != ModelBrowserSelect {
		t.Fatal(got)
	}
	b.ExpandedIDs = nil
	selectButton.Activate()
	if len(got) != 2 {
		t.Fatal("hidden descendant acted")
	}
	n, e = b.View()
	if e != nil || nodeByID(n, ModelBrowserRowID("b", "part")+"/select") != nil {
		t.Fatal("collapse")
	}
	b.Items = []ModelBrowserItem{{ID: "x", Parent: "y", Label: "X"}, {ID: "y", Parent: "x", Label: "Y"}}
	if _, e = b.View(); e == nil {
		t.Fatal("cycle accepted")
	}
}

func TestModelBrowserExactAncestryBoundary(t *testing.T) {
	b := &ModelBrowser{ID: "depth", Label: "Depth"}
	parent := ""
	for i := 0; i < 17; i++ {
		id := string(rune('a' + i))
		b.Items = append(b.Items, ModelBrowserItem{ID: id, Label: id, Parent: parent})
		parent = id
	}
	if _, e := b.View(); e != nil {
		t.Fatal("16 ancestors must fit", e)
	}
	b.Items = append(b.Items, ModelBrowserItem{ID: "too-deep", Label: "Too deep", Parent: parent})
	if _, e := b.View(); e == nil {
		t.Fatal("17 ancestors accepted")
	}
}
