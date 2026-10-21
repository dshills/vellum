package main

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/placement"
)

func TestResizeDragTransformedHistoryAndCancel(t *testing.T) {
	m := editModel(t)
	m.scene.Selected = "block-1"
	m.snap = false
	g := placement.Geometry{Window: "placement", Origin: canvas.Point{X: 35, Y: -20}, Clip: canvas.Bounds{Min: canvas.Point{X: 40, Y: 0}, Max: canvas.Point{X: 1600, Y: 1200}}, Scale: 2, View: canvas.Viewport{Width: 800, Height: 600, Zoom: 1.5, Center: canvas.Point{X: 30, Y: -20}}}
	startLocal := g.View.Screen(canvas.Point{X: 80, Y: 50})
	start := canvas.Point{X: g.Origin.X + startLocal.X*g.Scale, Y: g.Origin.Y + startLocal.Y*g.Scale}
	before := slices.Clone(m.scene.Objects)
	count, next := m.journal.Snapshot().UndoCount, m.next
	s := m.beginMove(g, start)
	if s == nil || !s.resize {
		t.Fatal("resize not captured")
	}
	for n := 1; n <= 20; n++ {
		m.previewMove(s, canvas.Point{X: start.X + float64(n*6), Y: start.Y + float64(n*3)})
	}
	if m.scene.Objects[0].Width != 120 || m.scene.Objects[0].Height != 70 || m.scene.Objects[0].X != before[0].X || m.scene.Objects[0].Y != before[0].Y || m.journal.Snapshot().UndoCount != count {
		t.Fatal("resize preview changed anchor/history or used wrong map")
	}
	if _, err := drawingCodec(m).Capture(); err == nil {
		t.Fatal("saved transient size")
	}
	end := canvas.Point{X: start.X + 120, Y: start.Y + 60}
	if err := m.finishMove(s, end); err != nil {
		t.Fatal(err)
	}
	resized := slices.Clone(m.scene.Objects)
	if m.move != nil || m.journal.Snapshot().UndoCount != count+1 || m.next != next {
		t.Fatal("resize not single history entry")
	}
	_ = m.finishMove(s, end)
	if m.journal.Snapshot().UndoCount != count+1 {
		t.Fatal("duplicate release")
	}
	if err := m.journal.Undo(); err != nil || !slices.Equal(m.scene.Objects, before) {
		t.Fatal("resize undo", err)
	}
	if err := m.journal.Redo(); err != nil || !slices.Equal(m.scene.Objects, resized) {
		t.Fatal("resize redo identity/order", err)
	}
	_ = m.journal.Undo()
	for _, cancel := range []string{"click", "escape", "outside", "camera"} {
		t.Run(cancel, func(t *testing.T) {
			s := m.beginMove(g, start)
			if s == nil {
				t.Fatal("begin")
			}
			if cancel != "click" {
				m.previewMove(s, end)
			}
			switch cancel {
			case "click":
				_ = m.finishMove(s, start)
			case "escape":
				m.cancelMove(s)
			case "outside":
				_ = m.finishMove(s, canvas.Point{X: 1, Y: 1})
			case "camera":
				changed := g
				changed.View.Zoom = 2
				m.updateMoveGeometry(s, changed)
			}
			if m.move != nil || !slices.Equal(m.scene.Objects, before) || m.journal.Snapshot().UndoCount != count {
				t.Fatal("canceled resize changed document")
			}
		})
	}
}

func TestResizeLayoutGrabSnappingAndBounds(t *testing.T) {
	m := editModel(t)
	m.scene.Selected = "block-1"
	m.snap = false
	g := placement.Geometry{Window: "placement", Clip: canvas.Bounds{Max: canvas.Point{X: 800, Y: 600}}, Scale: 1, View: canvas.Viewport{Width: 800, Height: 600, Zoom: 1}}
	start := canvas.Point{X: 480, Y: 350}
	s := m.beginMove(g, start)
	g.Origin.Y = 18
	g.View.Height = 582
	g.Clip.Min.Y = 18
	if !m.updateMoveGeometry(s, g) {
		t.Fatal("layout canceled")
	}
	if err := m.finishMove(s, canvas.Point{X: 520, Y: 370}); err != nil || m.scene.Objects[0].Width != 120 || m.scene.Objects[0].Height != 61 {
		t.Fatal("layout broke resize offset", err, m.scene.Objects[0])
	}
	m.scene.Selected = "connector-2"
	height := m.scene.Objects[1].Height
	m.snap = true
	if err := m.commands.Invoke(context.Background(), "width-more"); err != nil || m.scene.Objects[1].Width != 140 || m.scene.Objects[1].Height != height {
		t.Fatal("width button changed untouched height", err)
	}
	m.snap = false
	if err := m.commands.Invoke(context.Background(), "height-more"); err != nil || m.scene.Objects[1].Height != height+1 {
		t.Fatal("fine sizing", err)
	}
	if err := m.resizeSelected(-1e6, -1e6); err != nil || m.scene.Objects[1].Width != 8 || m.scene.Objects[1].Height != 8 {
		t.Fatal("minimum clamp", err)
	}
	n := m.journal.Snapshot().UndoCount
	if err := m.resizeSelected(-1, -1); err != nil || m.journal.Snapshot().UndoCount != n {
		t.Fatal("minimum noop added history", err)
	}
	if err := m.resizeSelected(1e6, 1e6); err != nil || m.scene.Objects[1].Width != 2000 || m.scene.Objects[1].Height != 2000 {
		t.Fatal("maximum clamp", err)
	}
	n = m.journal.Snapshot().UndoCount
	if err := m.resizeSelected(math.NaN(), 1); err == nil || m.journal.Snapshot().UndoCount != n {
		t.Fatal("invalid delta accepted")
	}
	m.prompt = "open"
	if m.resizeSelected(1, 0) == nil {
		t.Fatal("resize during prompt")
	}
}

func TestResizedDrawingFileRoundTripAndLegacy(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	var saved drawingFile
	f.owner(t, func() {
		f.m.snap = false
		if err := f.m.resizeSelected(37, 21); err != nil {
			t.Fatal(err)
		}
		var err error
		saved, err = drawingCodec(f.m).Capture()
		if err != nil || saved.Version != 2 {
			t.Fatal(err)
		}
		if err = f.m.fileAction(f.a, f.ref, "save", false); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	data, err := os.ReadFile(f.path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := drawingCodec(f.m).Decode(data)
	if err != nil || !reflect.DeepEqual(got, saved) || got.Objects[0].Width != 117 || got.Objects[0].Height != 71 {
		t.Fatal("size lost on disk", err, got)
	}
	f.owner(t, func() {
		_ = f.m.resizeSelected(10, 10)
		if err := f.m.files.Open(f.ref.Context(), document.OpenOptions{DiscardChanges: true}); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	f.owner(t, func() {
		if f.m.scene.Objects[0].Width != 117 || f.m.scene.Objects[0].Height != 71 || f.m.journal.Snapshot().UndoCount != 0 {
			t.Fatal("open size/history")
		}
	})
	legacy := []byte(`{"format":"vellum-placement","version":1,"objects":[{"id":"connector-1","kind":"connector","x":20,"y":40}],"center":{"X":0,"Y":0},"zoom":1,"snap":true,"next_id":1}`)
	old, err := drawingCodec(f.m).Decode(legacy)
	if err != nil {
		t.Fatal("legacy decode", err)
	}
	scene, err := old.scene()
	if err != nil || scene.Objects[0].Width != 120 || scene.Objects[0].Height != 16 {
		t.Fatal("legacy defaults", err)
	}
}

func TestDrawingRejectsInvalidDimensions(t *testing.T) {
	m := editModel(t)
	good, err := drawingCodec(m).Capture()
	if err != nil {
		t.Fatal(err)
	}
	for _, badSize := range []float64{0, -1, 7, 2001, math.NaN(), math.Inf(1)} {
		bad := good
		bad.Objects = slices.Clone(good.Objects)
		bad.Objects[0].Width = badSize
		if _, err := drawingCodec(m).Encode(bad); err == nil {
			t.Fatal("encoded invalid width", badSize)
		}
		if data, err := json.Marshal(bad); err == nil {
			if _, err := drawingCodec(m).Decode(data); err == nil {
				t.Fatal("decoded invalid width", badSize)
			}
		}
	}
	for _, version := range []int{0, 1, 6} {
		bad := good
		bad.Version = version
		if _, err := bad.scene(); err == nil {
			t.Fatal("invalid version or v1 with dimensions", version)
		}
	}
}
