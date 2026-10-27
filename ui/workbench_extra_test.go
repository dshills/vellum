package ui

import (
	"github.com/dshills/vellum/viewport"
	"testing"
)

func TestQuantityAdmissionAndStaleCallbacks(t *testing.T) {
	calls := 0
	errors := 0
	p := &QuantityField{ID: "q", Label: "Width", Unit: "mm", Text: "2 in", Min: 1, Max: 100, Commit: func(v float64) {
		if v != 50.8 {
			t.Fatal(v)
		}
		calls++
	}, Invalid: func(error) { errors++ }}
	n, e := p.View()
	if e != nil {
		t.Fatal(e)
	}
	apply := n.Children[1].Children[1].Activate
	apply()
	p.Text = "1/0"
	apply()
	p.Disabled = true
	p.Text = "2 in"
	apply()
	p.Disabled = false
	p.ID = "changed"
	apply()
	if calls != 1 || errors != 1 {
		t.Fatal(calls, errors)
	}
}
func TestTimelineAndMarkingControlled(t *testing.T) {
	calls := 0
	p := &FeatureTimeline{ID: "t", Label: "Features", Items: []FeatureItem{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}}, Marker: 2, SelectedID: "a", Change: func(e FeatureTimelineEvent) { calls++ }}
	n, e := p.View()
	if e != nil {
		t.Fatal(e)
	}
	n.Children[0].Children[4].Activate()
	if calls != 1 {
		t.Fatal(calls)
	}
	p.Items[0].Disabled = true
	n.Children[0].Children[4].Activate()
	if calls != 1 {
		t.Fatal("stale disabled action")
	}
	p.Marker = 3
	if _, e = p.View(); e == nil {
		t.Fatal("marker bounds")
	}
	m := &MarkingMenu{ID: "m", Label: "Actions", Open: true, Actions: []MarkingAction{{ID: "a", Label: "A"}}, Choose: func(string) { calls++ }}
	node, e := m.View()
	if e != nil {
		t.Fatal(e)
	}
	node.Children[0].Children[1].Activate()
	m.Open = false
	m.ChooseDirection(0)
	if calls != 2 {
		t.Fatal(calls)
	}
	v := &ViewCube{ID: "v", Label: "View", Change: func(viewport.View) { calls++ }}
	cube, e := v.View()
	if e != nil {
		t.Fatal(e)
	}
	cube.Children[1].Activate()
	v.Disabled = true
	cube.Children[1].Activate()
	if calls != 3 {
		t.Fatal(calls)
	}
}
