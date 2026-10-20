package main

import (
	"math"
	"slices"
	"testing"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/placement"
)

func groupGeometry() placement.Geometry {
	g := cornerGeometry()
	g.View.Width, g.View.Center.X = 1200, 245
	g.Clip.Max.X = 3000
	return g
}

func TestGroupCornersScaleAboutOppositeAnchorAndHistory(t *testing.T) {
	for _, corner := range []resizeCorner{northWest, northEast, southEast, southWest} {
		t.Run(string(rune('0'+corner)), func(t *testing.T) {
			m := editModel(t)
			m.snap = false
			ids := []string{"block-1", "note-3"}
			m.setSelection(ids, "note-3")
			before := slices.Clone(m.scene.Objects)
			bounds, _ := selectionBounds(before, ids)
			g := groupGeometry()
			start := cornerHandles(bounds, g)[corner-1].center
			s := m.beginMove(g, start)
			if s == nil || !s.resize || s.corner != corner || len(s.selection) != 2 {
				t.Fatal("group handle failed to capture", s)
			}
			// Clicking an empty bounding-box corner preserves the full group.
			_ = m.finishMove(s, start)
			if !slices.Equal(m.selectedIDs(), ids) || m.journal.Snapshot().UndoCount != 3 {
				t.Fatal("corner click changed selection/history")
			}
			s = m.beginMove(g, start)
			dx, dy := bounds.Width*.25, bounds.Height*.5
			ax, ay := bounds.X, bounds.Y
			if corner == northWest || corner == southWest {
				dx = -dx
				ax += bounds.Width
			}
			if corner == northWest || corner == northEast {
				dy = -dy
				ay += bounds.Height
			}
			end := canvas.Point{X: start.X + dx*g.Scale*g.View.Zoom, Y: start.Y + dy*g.Scale*g.View.Zoom}
			for range 8 {
				m.previewMove(s, end)
			}
			if m.journal.Snapshot().UndoCount != 3 {
				t.Fatal("preview journal entry")
			}
			if err := m.finishMove(s, end); err != nil {
				t.Fatal(err)
			}
			after := slices.Clone(m.scene.Objects)
			for i, o := range before {
				if i == 1 {
					if after[i] != o {
						t.Fatal("unselected object changed")
					}
					continue
				}
				want := o
				want.X, want.Y = ax+(o.X-ax)*1.25, ay+(o.Y-ay)*1.5
				want.Width, want.Height = o.Width*1.25, o.Height*1.5
				if after[i] != want {
					t.Fatal("group affine transform/anchor", after[i], want)
				}
			}
			_ = m.finishMove(s, end)
			if m.journal.Snapshot().UndoCount != 4 || m.next != 3 || m.move != nil || !slices.Equal(m.selectedIDs(), ids) || m.scene.Selected != "note-3" {
				t.Fatal("not exactly one edit with original identities/selection")
			}
			if err := m.journal.Undo(); err != nil || !slices.Equal(m.scene.Objects, before) || !slices.Equal(m.selectedIDs(), ids) {
				t.Fatal("Undo", err)
			}
			if err := m.journal.Redo(); err != nil || !slices.Equal(m.scene.Objects, after) || !slices.Equal(m.selectedIDs(), ids) {
				t.Fatal("Redo", err)
			}
		})
	}
}

func TestGroupResizeSharedSnapAndStrictestMemberLimits(t *testing.T) {
	before := []canvas.Rect{{ID: "a", X: 10, Y: 20, Width: 80, Height: 50}, {ID: "b", X: 210, Y: 70, Width: 40, Height: 100}, {ID: "untouched", X: -500, Width: 20, Height: 20}}
	ids := []string{"a", "b"}
	after, err := resizedSelection(before, ids, southEast, 23, 0, true)
	if err != nil || math.Abs(after[0].Width-80*260.0/240) > 1e-10 || math.Abs(after[1].X-(10+200*260.0/240)) > 1e-10 || after[0].Height != before[0].Height || after[1].Y != before[1].Y || after[2] != before[2] {
		t.Fatal("shared bbox snap/unchanged axis", after, err)
	}
	for _, corner := range []resizeCorner{northWest, northEast, southEast, southWest} {
		dx, dy := -100000.0, -100000.0
		if corner == northWest || corner == southWest {
			dx = -dx
		}
		if corner == northWest || corner == northEast {
			dy = -dy
		}
		after, err = resizedSelection(before, ids, corner, dx, dy, false)
		if err != nil || after[1].Width != 8 || after[0].Width != 16 || after[0].Height != 8 || after[1].Height != 16 {
			t.Fatal("minimum shared scale/crossing", corner, after, err)
		}
	}
	after, err = resizedSelection(before, ids, southEast, 100000, 100000, false)
	if err != nil || after[0].Width != 2000 || after[1].Width != 1000 || after[1].Height != 2000 || after[0].Height != 1000 {
		t.Fatal("maximum shared scale", after, err)
	}
	for _, delta := range []float64{math.NaN(), math.Inf(1)} {
		if _, err := resizedSelection(before, ids, southEast, delta, 0, false); err == nil {
			t.Fatal("nonfinite delta accepted")
		}
	}
	// A member origin outside bounds rejects the entire detached result.
	before[1].X = 9990
	if after, err := resizedSelection(before, ids, southEast, 200, 10, false); err == nil || after != nil {
		t.Fatal("partial invalid group result")
	}
}

func TestGroupResizeCancelClipAndLifecycle(t *testing.T) {
	for _, cancel := range []string{"escape", "outside", "camera", "scale", "semantic", "close"} {
		t.Run(cancel, func(t *testing.T) {
			m := editModel(t)
			m.snap = false
			ids := []string{"block-1", "note-3"}
			m.setSelection(ids, "note-3")
			g := groupGeometry()
			before := slices.Clone(m.scene.Objects)
			bounds, _ := selectionBounds(before, ids)
			start := cornerHandles(bounds, g)[2].center
			clipped := g
			clipped.Clip.Max.X = start.X - 1
			if handleAt(bounds, clipped, canvas.Point{X: start.X - 2, Y: start.Y}) != noCorner {
				t.Fatal("hidden handle hittable")
			}
			s := m.beginMove(g, start)
			end := canvas.Point{X: start.X + 90, Y: start.Y + 60}
			m.previewMove(s, end)
			if _, err := drawingCodec(m).Capture(); err == nil {
				t.Fatal("saved transient preview")
			}
			switch cancel {
			case "outside":
				_ = m.finishMove(s, canvas.Point{X: -1, Y: -1})
			case "camera":
				g.View.Center.X++
				m.updateMoveGeometry(s, g)
			case "scale":
				g.Scale++
				m.updateMoveGeometry(s, g)
			case "semantic":
				m.managedSelection("connector-2")
			case "close":
				_ = m.journal.MarkSaved()
				m.requestClose(nil)
			default:
				m.cancelMove(s)
			}
			_ = m.finishMove(s, end)
			if m.move != nil || !slices.Equal(m.scene.Objects, before) || m.journal.Snapshot().UndoCount != 3 {
				t.Fatal("cancel/stale release committed")
			}
		})
	}
}

func TestGroupResizeFileRoundTrip(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.place(t, "note", canvas.Point{X: 200, Y: 100})
	var saved []canvas.Rect
	f.owner(t, func() {
		f.m.setSelection([]string{"block-1", "note-2"}, "note-2")
		after, err := resizedSelection(f.m.scene.Objects, f.m.selectedIDs(), northWest, -45, -35, false)
		if err != nil {
			t.Fatal(err)
		}
		if err = f.m.applyObjects("Resize selection", after, f.m.scene.Selected); err != nil {
			t.Fatal(err)
		}
		saved = slices.Clone(f.m.scene.Objects)
		if err = f.m.fileAction(f.a, f.ref, "save", false); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	f.owner(t, func() {
		if err := f.m.deleteSelected(); err != nil {
			t.Fatal(err)
		}
		if err := f.m.files.Open(f.ref.Context(), document.OpenOptions{DiscardChanges: true}); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	f.owner(t, func() {
		if !slices.Equal(f.m.scene.Objects, saved) || !slices.Equal(f.m.selectedIDs(), []string{"note-2"}) || f.m.journal.Snapshot().UndoCount != 0 {
			t.Fatal("resized group roundtrip")
		}
	})
}

func TestGroupResizeFloatingPointLimitsKeepSharedScale(t *testing.T) {
	for _, tc := range []struct{ size, delta float64 }{{49, -100000}, {15, 100000}} {
		// 49*(8/49) rounds below 8; 15*(2000/15) rounds above 2000.
		before := []canvas.Rect{{ID: "a", Width: tc.size, Height: tc.size}, {ID: "b", X: 10, Width: tc.size, Height: tc.size}, {ID: "untouched", X: -50, Width: 20, Height: 20}}
		after, err := resizedSelection(before, []string{"a", "b"}, southEast, tc.delta, tc.delta, false)
		limit := minComponentSize
		if tc.delta > 0 {
			limit = maxComponentSize
		}
		scale := math.Nextafter(limit/tc.size, 1)
		if err != nil || !validSize(after[0].Width, after[0].Height) || after[0].Width != tc.size*scale || after[1].Width != tc.size*scale || after[0].Height != tc.size*scale || after[1].X != 10*scale || after[0].X != 0 || after[0].Y != 0 || after[2] != before[2] {
			t.Fatal("floating size limit lost shared scale", tc, after, err)
		}
	}
}
