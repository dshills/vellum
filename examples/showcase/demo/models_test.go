package demo

import (
	"context"
	"math"
	"reflect"
	"slices"
	"testing"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
)

func TestFormaHistoryRestoresIdentityAndCompleteEdits(t *testing.T) {
	m, e := NewForma()
	if e != nil {
		t.Fatal(e)
	}
	initial := slices.Clone(m.state.Shapes)
	if e = m.Add("square"); e != nil {
		t.Fatal(e)
	}
	id := m.state.Selected
	if m.journal.Snapshot().UndoCount != 1 {
		t.Fatal("placement must be one command")
	}
	if e = m.Undo(); e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(m.state.Shapes, initial) {
		t.Fatal("undo failed")
	}
	if e = m.Redo(); e != nil {
		t.Fatal(e)
	}
	if m.state.Selected != id || m.selected().Kind != "square" {
		t.Fatal("redo changed object identity")
	}
	before := m.selected()
	if e = m.ChangeSelected(rgb(50, 170, 180), 144, 16); e != nil {
		t.Fatal(e)
	}
	if m.selected().X != before.X+16 || m.selected().Size != 144 {
		t.Fatal("edit missing")
	}
	if e = m.Undo(); e != nil || m.selected() != before {
		t.Fatal("edit undo failed", e)
	}
	state := m.journal.Snapshot()
	if e = m.ChangeSelected(blue, math.NaN(), 0); e == nil || m.journal.Snapshot() != state {
		t.Fatal("invalid edit entered history")
	}
	if e = m.Commands().Invoke(context.Background(), "forma.redo"); e != nil {
		t.Fatal(e)
	}
	if m.selected().Size != 144 {
		t.Fatal("registry differs from button path")
	}
	if e = ui.Validate(m.View()); e != nil {
		t.Fatal(e)
	}
}
func TestFormaBoundedCanvasAndFit(t *testing.T) {
	m, _ := NewForma()
	for len(m.state.Shapes) < 48 {
		if e := m.Add("circle"); e != nil {
			t.Fatal(e)
		}
	}
	if e := m.Add("circle"); e == nil {
		t.Fatal("unbounded shapes")
	}
	m.Fit()
	if m.view.Zoom <= 0 || m.view.Zoom > 1.1 {
		t.Fatal("fit")
	}
	s := m.selected()
	m.pick(canvas.Point{X: s.X, Y: s.Y})
	if m.state.Selected != s.ID {
		t.Fatal("hit test selection")
	}
	mesh, e := m.prepare(m.view)
	if e != nil {
		t.Fatal(e)
	}
	v, ix := mesh.Counts()
	if v == 0 || ix == 0 {
		t.Fatal("empty scene")
	}
}
func TestLumaSelectionSearchFavoritesAndImmutableLooks(t *testing.T) {
	m, e := NewLuma()
	if e != nil {
		t.Fatal(e)
	}
	if len(m.visible()) != 6 {
		t.Fatal("fixture count")
	}
	_ = m.View()
	m.collection.Change(ui.ListSelectionEvent{TargetID: "art-1", IDs: []string{"art-1"}, Add: true})
	if len(m.selection) != 2 || m.selected != "art-1" {
		t.Fatal("filmstrip multi-selection")
	}
	if e = m.SetSearch("canyon"); e != nil {
		t.Fatal(e)
	}
	if m.selected != "art-0" || !reflect.DeepEqual(m.selection, []string{"art-0"}) {
		t.Fatal("filter retained absent selection")
	}
	original := landscapeAssets()[0][0]
	if e = m.SetLook(2); e != nil {
		t.Fatal(e)
	}
	mono := landscapeAssets()[0][2].NRGBAAt(300, 260)
	if mono.R != mono.G || mono.G != mono.B || landscapeAssets()[0][0] != original {
		t.Fatal("look changed original")
	}
	if e = m.SetSearch("no match at all"); e != nil || m.selected != "" || len(m.selection) != 0 {
		t.Fatal("empty filter")
	}
	if e = ui.Validate(m.View()); e != nil {
		t.Fatal(e)
	}
	_ = m.SetSearch("")
	_ = m.SetFilter("favorites")
	id := m.selected
	if e = m.ToggleFavorite(); e != nil {
		t.Fatal(e)
	}
	if m.selected == id {
		t.Fatal("removed favorite remained in filtered view")
	}
	if e = m.SetLook(9); e == nil {
		t.Fatal("unknown look admitted")
	}
	if e = m.SetSearch("\xff"); e == nil {
		t.Fatal("invalid search")
	}
}
func TestPulseSourceIsImmutableAndIndexed(t *testing.T) {
	m, e := NewPulse()
	if e != nil {
		t.Fatal(e)
	}
	old := m.source
	if old.RowCount() != 1024 {
		t.Fatal("fixture count")
	}
	for i := range old.RowCount() {
		key, e := old.RowKey(i)
		if e != nil {
			t.Fatal(e)
		}
		index, ok, e := old.IndexOf(key)
		if e != nil || !ok || index != i {
			t.Fatal("inverse index")
		}
	}
	before, _ := old.Cell("svc-0", "latency")
	if e = m.NextSample(); e != nil {
		t.Fatal(e)
	}
	after, _ := old.Cell("svc-0", "latency")
	if before != after || m.source == old || m.source.Revision() <= old.Revision() {
		t.Fatal("source mutated or revision unchanged")
	}
	if e = m.SetFilter("degraded"); e != nil {
		t.Fatal(e)
	}
	for _, r := range m.source.rows {
		if r.Status != "Degraded" {
			t.Fatal("filter ignored")
		}
	}
	if e = m.SetSearch("no such service"); e != nil {
		t.Fatal(e)
	}
	if m.source.RowCount() != 0 || m.selected != "" {
		t.Fatal("empty result state")
	}
	if e = ui.Validate(m.View()); e != nil {
		t.Fatal(e)
	}
	if e = m.SetPeriod("invalid"); e == nil {
		t.Fatal("bad period")
	}
}
func TestPulseChartRangeChangesAndGridStaysBounded(t *testing.T) {
	m, _ := NewPulse()
	a := m.latency(14, false)
	_ = m.SetPeriod("1h")
	if m.latency(14, false) == a {
		t.Fatal("range changed no chart data")
	}
	s, e := m.grid.Snapshot()
	if e != nil {
		t.Fatal(e)
	}
	if len(s.Cells) > 1024 || s.State.Stamp.Rows != 1024 {
		t.Fatal("projection unbounded or source lost")
	}
	mesh, e := m.prepareChart(m.view)
	if e != nil {
		t.Fatal(e)
	}
	v, ix := mesh.Counts()
	if v < 100 || ix < 100 {
		t.Fatal("chart missing geometry")
	}
	if e = ui.Validate(m.View()); e != nil {
		t.Fatal(e)
	}
}
