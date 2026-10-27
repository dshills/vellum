package main

import (
	"github.com/dshills/vellum/quantity"
	"github.com/dshills/vellum/ui"
	"math"
	"testing"
)

func TestDocumentCommandsAndHistory(t *testing.T) {
	m := newWorkbench()
	count := len(m.state.Objects)
	if e := m.add(); e != nil {
		t.Fatal(e)
	}
	id := m.state.Selected
	if len(m.state.Objects) != count+1 {
		t.Fatal("add")
	}
	if e := m.journal.Undo(); e != nil {
		t.Fatal(e)
	}
	if len(m.state.Objects) != count {
		t.Fatal("undo")
	}
	if e := m.journal.Redo(); e != nil || m.state.Selected != id {
		t.Fatal("redo identity", e)
	}
	v, e := quantity.Length("1 in", "mm", nil)
	if e != nil {
		t.Fatal(e)
	}
	if e = m.resize(v); e != nil {
		t.Fatal(e)
	}
	o := m.selected()
	if math.Abs((o.Bounds.Max.X-o.Bounds.Min.X)*20-25.4) > 1e-9 {
		t.Fatal("width")
	}
	if e = m.visibility(id, false); e != nil {
		t.Fatal(e)
	}
	if m.scene.Objects[len(m.scene.Objects)-1].Visible {
		t.Fatal("visibility")
	}
	if e = m.journal.Undo(); e != nil {
		t.Fatal(e)
	}
	if !m.selected().Visible {
		t.Fatal("visibility undo")
	}
	if e = m.deleteSelected(); e != nil {
		t.Fatal(e)
	}
	if e = m.journal.Undo(); e != nil || m.state.Selected != id {
		t.Fatal("delete undo", e)
	}
}
func TestTimelineAndDescriptions(t *testing.T) {
	m := newWorkbench()
	for _, tab := range []string{"solid", "inspect", "body"} {
		m.tab = tab
		if e := ui.Validate(m.View()); e != nil {
			t.Fatal(tab, e)
		}
	}
	first := m.state.Features[0]
	m.featureChange(ui.FeatureTimelineEvent{Action: "move", ItemID: first.ID, Index: 2})
	if m.state.Features[2].ID != first.ID {
		t.Fatal("reorder")
	}
	m.result(m.journal.Undo())
	if m.state.Features[0] != first {
		t.Fatal("undo reorder")
	}
	m.featureChange(ui.FeatureTimelineEvent{Action: "marker", Index: 0})
	for _, o := range m.scene.Objects {
		if o.Visible {
			t.Fatal("marker did not hide body")
		}
	}
	m.result(m.journal.Undo())
	if !m.scene.Objects[0].Visible {
		t.Fatal("marker undo")
	}
	m.marking.Open = true
	if e := ui.Validate(m.View()); e != nil {
		t.Fatal(e)
	}
}

func TestUnfeaturedBodyKeepsExplicitVisibility(t *testing.T) {
	m := newWorkbench()
	id := m.state.Objects[0].ID
	features := m.state.Features[:0]
	for _, f := range m.state.Features {
		if f.Object != id {
			features = append(features, f)
		}
	}
	m.state.Features = features
	m.state.Marker = 0
	m.refresh()
	if !m.scene.Objects[0].Visible {
		t.Fatal("unfeatured object hidden")
	}
	if m.scene.Objects[1].Visible {
		t.Fatal("known feature ignored marker")
	}
}

func TestContextTabFallsBackAfterDeletion(t *testing.T) {
	m := newWorkbench()
	m.tab = "body"
	if e := m.deleteSelected(); e != nil {
		t.Fatal(e)
	}
	if m.tab != "solid" {
		t.Fatal("inactive context remained selected")
	}
	if e := ui.Validate(m.View()); e != nil {
		t.Fatal(e)
	}
}
