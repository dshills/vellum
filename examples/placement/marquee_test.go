package main

import (
	"slices"
	"testing"

	"github.com/dshills/vellum/canvas"
)

func TestMarqueeTransformedDirectionsAndSelectionOnlyCommit(t *testing.T) {
	for _, direction := range []struct{ start, end canvas.Point }{
		{canvas.Point{X: -20, Y: -20}, canvas.Point{X: 340, Y: 80}},
		{canvas.Point{X: 340, Y: 80}, canvas.Point{X: -20, Y: -20}},
		{canvas.Point{X: 340, Y: -20}, canvas.Point{X: -20, Y: 80}},
		{canvas.Point{X: -20, Y: 80}, canvas.Point{X: 340, Y: -20}},
	} {
		m := editModel(t)
		m.setSelection(nil, "")
		_ = m.journal.MarkSaved()
		before := slices.Clone(m.scene.Objects)
		g := groupGeometry()
		start, end := selectionPoint(g, direction.start), selectionPoint(g, direction.end)
		s := m.selectionToggle(g, start)
		if s == nil || !s.marquee {
			t.Fatal("empty Shift press did not capture marquee")
		}
		view, next := m.scene.View, m.next
		m.previewMove(s, canvas.Point{X: start.X + 3*g.Scale, Y: start.Y})
		if s.dragging {
			t.Fatal("marquee started below threshold")
		}
		for range 10 {
			m.previewMove(s, end)
		}
		if !s.dragging || s.marqueeBounds != marqueeRect(direction.start, direction.end) || len(m.selectedIDs()) != 0 || !slices.Equal(m.scene.Objects, before) {
			t.Fatal("preview mapping or transient selection", s.marqueeBounds)
		}
		if err := m.finishMove(s, end); err != nil {
			t.Fatal(err)
		}
		_ = m.finishMove(s, end)
		if !slices.Equal(m.selectedIDs(), []string{"block-1", "connector-2"}) || m.scene.Selected != "connector-2" || m.move != nil || m.scene.View != view || m.next != next || !slices.Equal(m.scene.Objects, before) || m.journal.Snapshot().UndoCount != 3 || m.journal.Snapshot().Dirty {
			t.Fatal("marquee changed document/camera/history or selection", m.selectedIDs())
		}
	}
}

func TestMarqueeIntersectionAdditiveOrderingAndNoop(t *testing.T) {
	objects := []canvas.Rect{{ID: "bottom", X: 0, Y: 0, Width: 80, Height: 50}, {ID: "overlap", X: 10, Y: 10, Width: 20, Height: 20}, {ID: "edge", X: 100, Y: 0, Width: 20, Height: 20}, {ID: "prior", X: 500, Y: 500, Width: 20, Height: 20}}
	got := marqueeIDs(objects, []string{"prior", "missing", "prior"}, canvas.Rect{X: 79, Y: 0, Width: 21, Height: 1})
	if !slices.Equal(got, []string{"bottom", "prior"}) {
		t.Fatal("partial/edge-only intersection or paint order", got)
	}
	got = marqueeIDs(objects, nil, canvas.Rect{X: 15, Y: 15, Width: 1, Height: 1})
	if !slices.Equal(got, []string{"bottom", "overlap"}) {
		t.Fatal("overlapping objects must both select", got)
	}
	if len(marqueeIDs(objects, nil, canvas.Rect{Width: 100, Height: 0})) != 0 {
		t.Fatal("zero-area marquee selected objects")
	}
	m := editModel(t)
	m.setSelection([]string{"note-3"}, "note-3")
	g := groupGeometry()
	p := selectionPoint(g, canvas.Point{X: -20, Y: -20})
	status := m.status
	s := m.selectionToggle(g, p)
	_ = m.finishMove(s, p)
	if m.status != status || !slices.Equal(m.selectedIDs(), []string{"note-3"}) {
		t.Fatal("empty Shift click cleared selection")
	}
	s = m.selectionToggle(g, p)
	_ = m.finishMove(s, selectionPoint(g, canvas.Point{X: 340, Y: 80}))
	if !slices.Equal(m.selectedIDs(), []string{"block-1", "connector-2", "note-3"}) || m.scene.Selected != "note-3" {
		t.Fatal("additive marquee lost prior primary", m.selectedIDs())
	}
	// Unmodified empty drag still falls through to existing canvas pan routing.
	if m.beginMove(g, p) != nil {
		t.Fatal("plain empty input captured by editing")
	}
	// The resulting selection uses the existing one-step group edit history.
	before := slices.Clone(m.scene.Objects)
	if err := m.deleteSelected(); err != nil {
		t.Fatal(err)
	}
	if err := m.journal.Undo(); err != nil || !slices.Equal(m.scene.Objects, before) || len(m.selectedIDs()) != 3 {
		t.Fatal("delete/Undo after marquee", err)
	}
}

func TestMarqueeCancelNoopAndLifecycle(t *testing.T) {
	for _, how := range []string{"click", "empty", "escape", "capture-loss", "outside", "camera", "scale", "window", "semantic", "close"} {
		t.Run(how, func(t *testing.T) {
			m := editModel(t)
			m.setSelection([]string{"note-3"}, "note-3")
			_ = m.journal.MarkSaved()
			before := slices.Clone(m.scene.Objects)
			g := groupGeometry()
			p := selectionPoint(g, canvas.Point{X: -20, Y: -20})
			s := m.selectionToggle(g, p)
			end := selectionPoint(g, canvas.Point{X: 340, Y: 80})
			if how != "click" {
				m.previewMove(s, end)
			}
			if _, err := drawingCodec(m).Capture(); err == nil {
				t.Fatal("saved active marquee")
			}
			switch how {
			case "click":
				_ = m.finishMove(s, p)
			case "empty":
				_ = m.finishMove(s, selectionPoint(g, canvas.Point{X: -100, Y: -100}))
			case "outside":
				_ = m.finishMove(s, canvas.Point{X: -1, Y: -1})
			case "camera":
				g.View.Center.X++
				m.updateMoveGeometry(s, g)
			case "scale":
				g.Scale++
				m.updateMoveGeometry(s, g)
			case "window":
				g.Window = "other"
				m.updateMoveGeometry(s, g)
			case "semantic":
				m.managedSelection("note-3")
			case "close":
				if !m.requestClose(nil) {
					t.Fatal("clean marquee close blocked")
				}
			default:
				m.cancelMove(s)
			}
			_ = m.finishMove(s, end)
			if m.move != nil || !slices.Equal(m.selectedIDs(), []string{"note-3"}) || !slices.Equal(m.scene.Objects, before) || m.journal.Snapshot().UndoCount != 3 || m.journal.Snapshot().Dirty {
				t.Fatal("cancel/no-op changed selection/document/history")
			}
		})
	}
}
