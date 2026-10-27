package ui

import (
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/viewport"
	"image"
	"image/color"
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
	nodeByID(n, "t/later").Activate()
	if calls != 1 {
		t.Fatal(calls)
	}
	p.Items[0].Disabled = true
	nodeByID(n, "t/later").Activate()
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

func TestCompactTimelineKeepsNamesFallbackMarkerAndAdmission(t *testing.T) {
	pixels := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	pixels.SetNRGBA(0, 0, color.NRGBA{R: 230, A: 255})
	art, _ := canvas.NewTiledImage(pixels)
	var got []FeatureTimelineEvent
	p := &FeatureTimeline{ID: "compact", Label: "Feature history", Items: []FeatureItem{{ID: "a", Label: "Create A", IconImage: art}, {ID: "b", Label: "Create B"}}, SelectedID: "a", Marker: 1, Change: func(e FeatureTimelineEvent) { got = append(got, e) }}
	n, e := p.View()
	if e != nil {
		t.Fatal(e)
	}
	a := nodeByID(n, "compact/feature/"+ribbonPart("a"))
	b := nodeByID(n, "compact/feature/"+ribbonPart("b"))
	if a.Text != "" || a.IconImage != art || a.Accessibility.Name != "Create A" || b.Text != "Create B" || b.Accessibility.Name != "Create B · paused" {
		t.Fatal("lost names/art/fallback")
	}
	a.Activate()
	nodeByID(n, "compact/later").Activate()
	nodeByID(n, "compact/back").Activate()
	if len(got) != 3 || got[0].Action != "select" || got[1].Action != "move" || got[1].Index != 1 || got[2].Action != "marker" || got[2].Index != 0 {
		t.Fatal(got)
	}
	p.SelectedID = "b"
	paused, e := p.View()
	if e != nil {
		t.Fatal(e)
	}
	selectedPaused := nodeByID(paused, "compact/feature/"+ribbonPart("b"))
	if selectedPaused.Style.Background.Value != (color.NRGBA{R: 237, G: 240, B: 244, A: 255}) || selectedPaused.Style.Border.Value.Width != 1 {
		t.Fatal("paused selection lost fill/outline")
	}
	p.Items[0].Disabled = true
	a.Activate()
	if len(got) != 3 {
		t.Fatal("disabled feature accepted")
	}
	if nodeByID(n, "compact/marker") == nil {
		t.Fatal("marker missing")
	}
	p.ID = "retired"
	nodeByID(n, "compact/back").Activate()
	if len(got) != 3 {
		t.Fatal("retired callback")
	}
}
