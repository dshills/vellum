package main

import (
	"maps"
	"slices"
	"testing"

	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/canvas"
)

func testLayerLayout(m *model, scale float64) layerLayout {
	g := layerLayout{clip: canvas.Bounds{Min: canvas.Point{X: 100 * scale, Y: 10 * scale}, Max: canvas.Point{X: 400 * scale, Y: 1000 * scale}}, scale: scale}
	for i, r := range m.layerRows() {
		y := float64(10+i*40) * scale
		g.rows = append(g.rows, layerBox{r.id, canvas.Bounds{Min: canvas.Point{X: 110 * scale, Y: y}, Max: canvas.Point{X: 390 * scale, Y: y + 40*scale}}})
	}
	return g
}
func layerPoint(g layerLayout, id string, front bool) canvas.Point {
	for _, r := range g.rows {
		if r.id == id {
			fraction := .75
			if front {
				fraction = .25
			}
			return canvas.Point{X: (r.bounds.Min.X + r.bounds.Max.X) / 2, Y: r.bounds.Min.Y + (r.bounds.Max.Y-r.bounds.Min.Y)*fraction}
		}
	}
	panic("fixture row missing")
}
func TestLayerInsertionBoundariesAndStableOrder(t *testing.T) {
	before := []canvas.Rect{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}, {ID: "e"}}
	groups := map[string]string{"a": "g1", "c": "g1", "d": "g2", "e": "g2"}
	for _, tc := range []struct {
		ids    []string
		target string
		front  bool
		want   []string
	}{
		{[]string{"a", "c"}, "b", true, []string{"b", "a", "c", "d", "e"}},
		{[]string{"a", "c"}, "b", false, []string{"a", "c", "b", "d", "e"}},
		{[]string{"b"}, "g1", true, []string{"a", "c", "b", "d", "e"}},
		{[]string{"b"}, "g1", false, []string{"b", "a", "c", "d", "e"}},
		{[]string{"a", "c"}, "g2", true, []string{"b", "d", "e", "a", "c"}},
		{[]string{"a", "c"}, "g2", false, []string{"b", "a", "c", "d", "e"}},
		{[]string{"b"}, "a", true, []string{"a", "b", "c", "d", "e"}},
	} {
		after, err := reorderedLayers(before, groups, tc.ids, tc.target, tc.front)
		if err != nil || !slices.Equal(objectIDs(after), tc.want) {
			t.Fatal("insertion", tc, objectIDs(after), err)
		}
	}
	for _, tc := range []struct {
		ids    []string
		target string
	}{{nil, "a"}, {[]string{"missing"}, "a"}, {[]string{"a"}, ""}, {[]string{"a"}, "missing"}, {[]string{"a", "c"}, "g1"}, {[]string{"a"}, "a"}} {
		if _, err := reorderedLayers(before, groups, tc.ids, tc.target, true); err == nil {
			t.Fatal("invalid insertion accepted", tc)
		}
	}
	if !slices.Equal(objectIDs(before), []string{"a", "b", "c", "d", "e"}) || groups["a"] != "g1" {
		t.Fatal("input mutated")
	}
}
func TestLayerDragGroupCommitHistoryExactlyOnceAndScale(t *testing.T) {
	for _, scale := range []float64{1, 2} {
		m := groupedModel(t)
		m.scene.View.Zoom = 2.75
		m.scene.View.Center = canvas.Point{X: 73, Y: -82}
		before, groups, view, next, ng := slices.Clone(m.scene.Objects), maps.Clone(m.groups), m.scene.View, m.next, m.nextGroup
		count := m.journal.Snapshot().UndoCount
		_ = m.journal.MarkSaved()
		g := testLayerLayout(m, scale)
		start := layerPoint(g, "group-1", true)
		s := m.beginLayerDrag("group-1", start, g)
		if s == nil || len(s.ids) != 2 {
			t.Fatal("group not captured")
		}
		p := layerPoint(g, "connector-2", true)
		if !m.updateLayerDrag(s, p, g) || !s.valid || !s.dragging || s.target != "connector-2" || !s.front || m.journal.Snapshot().UndoCount != count || !slices.Equal(m.scene.Objects, before) {
			t.Fatal("preview committed or wrong target")
		}
		if err := m.finishLayerDrag(s, p, g); err != nil {
			t.Fatal(err)
		}
		after := slices.Clone(m.scene.Objects)
		if !slices.Equal(objectIDs(after), []string{"connector-2", "block-1", "note-3"}) || m.layerDrag != nil || m.journal.Snapshot().UndoCount != count+1 || !m.journal.Snapshot().Dirty || !maps.Equal(groups, m.groups) || m.scene.View != view || m.next != next || m.nextGroup != ng {
			t.Fatal("commit changed unrelated state")
		}
		_ = m.finishLayerDrag(s, p, g)
		if m.journal.Snapshot().UndoCount != count+1 {
			t.Fatal("duplicate release committed")
		}
		if err := m.journal.Undo(); err != nil || !slices.Equal(m.scene.Objects, before) || !maps.Equal(m.groups, groups) || m.journal.Snapshot().Dirty {
			t.Fatal("drag Undo", err)
		}
		if err := m.journal.Redo(); err != nil || !slices.Equal(m.scene.Objects, after) {
			t.Fatal("drag Redo", err)
		}
	}
}
func TestLayerDragUnselectedSourceSelectionAndNoop(t *testing.T) {
	m := groupedModel(t)
	g := testLayerLayout(m, 1)
	prior := m.selectedIDs()
	primary := m.scene.Selected
	s := m.beginLayerDrag("connector-2", layerPoint(g, "connector-2", true), g)
	if s == nil || !slices.Equal(s.ids, []string{"connector-2"}) || !slices.Equal(m.selectedIDs(), prior) {
		t.Fatal("press selected unselected source early")
	}
	count := m.journal.Snapshot().UndoCount
	_ = m.finishLayerDrag(s, layerPoint(g, "group-1", true), g)
	if m.scene.Selected != "connector-2" || len(m.selectedIDs()) != 1 || m.scene.Objects[2].ID != "connector-2" || m.journal.Snapshot().UndoCount != count+1 {
		t.Fatal("unselected row drop")
	}
	_ = m.journal.Undo()
	if !slices.Equal(m.selectedIDs(), prior) || m.scene.Selected != primary {
		t.Fatal("Undo lost original selection")
	}
	// A component already immediately above the target remains a true no-op.
	g = testLayerLayout(m, 1)
	s = m.beginLayerDrag("connector-2", layerPoint(g, "connector-2", true), g)
	_ = m.finishLayerDrag(s, layerPoint(g, "block-1", true), g)
	if m.layerDrag != nil || m.journal.Snapshot().UndoCount != count || m.journal.Snapshot().RedoCount != 1 || !slices.Equal(m.selectedIDs(), prior) {
		t.Fatal("no-op branched history or selected source")
	}
}
func TestLayerDragClicksThresholdCancelOutsideAndSelf(t *testing.T) {
	m := editModel(t)
	m.layerAdd = true
	_ = m.selectLayer("connector-2")
	g := testLayerLayout(m, 2)
	p := layerPoint(g, "connector-2", true)
	s := m.beginLayerDrag("connector-2", p, g)
	if s == nil {
		t.Fatal("begin")
	}
	count := m.journal.Snapshot().UndoCount
	p.X += 5 * 2
	_ = m.finishLayerDrag(s, p, g)
	if !slices.Equal(m.selectedIDs(), []string{"note-3"}) || m.journal.Snapshot().UndoCount != count {
		t.Fatal("short click must toggle once")
	}
	for _, mode := range []string{"cancel", "outside", "self"} {
		g = testLayerLayout(m, 1)
		before, ids := slices.Clone(m.scene.Objects), m.selectedIDs()
		s = m.beginLayerDrag("connector-2", layerPoint(g, "connector-2", true), g)
		p = layerPoint(g, "block-1", true)
		_ = m.updateLayerDrag(s, p, g)
		switch mode {
		case "cancel":
			m.cancelLayerDrag(s)
			_ = m.finishLayerDrag(s, p, g)
		case "outside":
			_ = m.finishLayerDrag(s, canvas.Point{}, g)
		case "self":
			p = layerPoint(g, "connector-2", false)
			_ = m.finishLayerDrag(s, p, g)
		}
		if m.layerDrag != nil || !slices.Equal(m.scene.Objects, before) || !slices.Equal(m.selectedIDs(), ids) || m.journal.Snapshot().UndoCount != count {
			t.Fatal("canceled drag changed document", mode)
		}
	}
}
func TestLayerDragStaleGeometryAndBusyLifecycle(t *testing.T) {
	for _, mode := range []string{"layout", "scale", "clip", "objects", "groups", "selection", "close"} {
		m := groupedModel(t)
		_ = m.journal.MarkSaved()
		g := testLayerLayout(m, 1)
		s := m.beginLayerDrag("connector-2", layerPoint(g, "connector-2", true), g)
		if s == nil {
			t.Fatal("begin")
		}
		if m.commands.Enabled("undo") == nil || m.commands.Enabled("bring-forward") == nil || m.selectLayer("block-1") == nil {
			t.Fatal("busy guard")
		}
		if _, err := drawingCodec(m).Capture(); err == nil {
			t.Fatal("save during layer drag")
		}
		switch mode {
		case "layout":
			g.rows = slices.Clone(g.rows)
			g.rows[0].bounds.Max.Y++
		case "scale":
			g.scale = 2
		case "clip":
			g.clip.Max.Y--
		case "objects":
			m.scene.Objects = m.scene.Objects[:2]
		case "groups":
			m.groups = nil
		case "selection":
			m.setSelection(nil, "")
		case "close":
			if !m.requestClose(nil) {
				t.Fatal("clean close refused")
			}
		}
		p := layerPoint(g, "group-1", true)
		_ = m.finishLayerDrag(s, p, g)
		if m.layerDrag != nil || m.journal.Snapshot().Dirty || m.journal.Snapshot().UndoCount != 4 {
			t.Fatal("stale or closed drag committed", mode)
		}
	}
}
func TestLayerLayoutClipHalvesAndSelectedSet(t *testing.T) {
	m := groupedModel(t)
	m.setSelection([]string{"block-1", "connector-2"}, "connector-2")
	g := testLayerLayout(m, 1)
	s := m.beginLayerDrag("group-1", layerPoint(g, "group-1", true), g)
	if s == nil || len(s.ids) != 3 || s.primary != "connector-2" {
		t.Fatal("selected set not captured")
	}
	m.cancelLayerDrag(s)
	p := layerPoint(g, "connector-2", true)
	id, front, _ := g.rowAt(p)
	if id != "connector-2" || !front {
		t.Fatal("upper half")
	}
	p = layerPoint(g, "connector-2", false)
	id, front, _ = g.rowAt(p)
	if id != "connector-2" || front {
		t.Fatal("lower half")
	}
	g.clip.Max.Y = p.Y - 1
	if id, _, _ := g.rowAt(p); id != "" {
		t.Fatal("unclipped row target")
	}
}

func TestLayerGroupMarkerPreviewUsesBackmostMember(t *testing.T) {
	for _, visible := range []bool{true, false} {
		m := groupedModel(t)
		g := testLayerLayout(m, 1)
		if !visible {
			g.clip.Max.Y = 150
		}
		s := m.beginLayerDrag("connector-2", layerPoint(g, "connector-2", true), g)
		p := layerPoint(g, "group-1", false)
		_ = m.updateLayerDrag(s, p, g)
		if visible {
			if !s.valid || s.line != 170 {
				t.Fatal("group preview must mark last member boundary", s.line)
			}
		} else if s.valid {
			t.Fatal("offscreen group boundary accepted")
		}
		m.cancelLayerDrag(s)
	}
}
func TestLayerMountDestroyTransfersCancellationWithoutOwnerReentry(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.place(t, "note", canvas.Point{X: 200})
	f.owner(t, func() {
		g := testLayerLayout(f.m, 1)
		s := f.m.beginLayerDrag("block-1", layerPoint(g, "block-1", true), g)
		if s == nil {
			t.Fatal("begin")
		}
		ticket, err := f.a.ReserveWindowCompletion(f.ref, nil)
		if err != nil {
			t.Fatal(err)
		}
		w := tree.New[layerMount]()
		w.editor = &layerEditor{a: f.a, ref: f.ref, model: f.m}
		w.active = s
		w.ticket = ticket
		w.Destroy() // This call already holds the non-reentrant app owner.
		if w.ticket != nil || w.active != nil || w.sprite != nil || w.owner != nil {
			t.Fatal("native references retained")
		}
	})
	if err := f.a.Drain(); err != nil {
		t.Fatal(err)
	}
	f.owner(t, func() {
		if f.m.layerDrag != nil {
			t.Fatal("destroy canceled completion before dispatch")
		}
	})
}

func TestManagedCanvasSelectionCancelsLayerDragImmediately(t *testing.T) {
	m := groupedModel(t)
	g := testLayerLayout(m, 1)
	s := m.beginLayerDrag("connector-2", layerPoint(g, "connector-2", true), g)
	if s == nil {
		t.Fatal("begin")
	}
	count := m.journal.Snapshot().UndoCount
	m.scene.Selected = "block-1" // Managed canvas selects exact ID before its callback.
	m.managedSelection("block-1")
	if m.layerDrag != nil || m.scene.Selected != "block-1" || len(m.selectedIDs()) != 2 || m.journal.Snapshot().UndoCount != count {
		t.Fatal("AX selection left a busy layer gesture")
	}
	_ = m.finishLayerDrag(s, layerPoint(g, "group-1", true), g)
	if m.journal.Snapshot().UndoCount != count {
		t.Fatal("release after semantic cancellation committed")
	}
}

func TestLayerCapturePreservesPresentationButGuardsExecution(t *testing.T) {
	m := groupedModel(t)
	g := testLayerLayout(m, 1)
	before := make(map[string]bool)
	for _, c := range m.commands.Search("") {
		before[c.ID] = m.commandDisabled(c.ID)
	}
	s := m.beginLayerDrag("connector-2", layerPoint(g, "connector-2", true), g)
	if s == nil {
		t.Fatal("begin")
	}
	for id, want := range before {
		if m.commandDisabled(id) != want {
			t.Fatal("capture restyled command", id)
		}
	}
	if m.commands.Enabled("undo") == nil || m.selectLayer("block-1") == nil {
		t.Fatal("stable presentation bypassed semantic guard")
	}
	m.cancelLayerDrag(s)
	if m.commandDisabled("undo") {
		t.Fatal("availability did not recover")
	}
}
