package main

import (
	"errors"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/placement"
	"github.com/dshills/vellum/ui"
	"math"
	"slices"
	"testing"
)

func TestOverlapTransformsOrderingGroupsAndHistory(t *testing.T) {
	for _, scale := range []float64{1, 2, 1.5} {
		m := groupedModel(t)
		m.scene.Objects[1].X = 0
		m.scene.Objects[1].Y = 0
		m.scene.Objects[2].X = 0
		m.scene.Objects[2].Y = 0
		g := placement.Geometry{Window: "placement", Origin: canvas.Point{X: 130, Y: 240}, Scale: scale, View: canvas.Viewport{Width: 500, Height: 300, Center: canvas.Point{X: 25, Y: -15}, Zoom: 2}, Clip: canvas.Bounds{Min: canvas.Point{X: 130, Y: 240}, Max: canvas.Point{X: 130 + 500*scale, Y: 240 + 300*scale}}}
		local := g.View.Screen(canvas.Point{X: 10, Y: 10})
		point := canvas.Point{X: g.Origin.X + scale*local.X, Y: g.Origin.Y + scale*local.Y}
		state := m.journal.Snapshot()
		if err := m.beginOverlapAt(g, point); err != nil {
			t.Fatal(err)
		}
		d := m.overlapDraft
		opts, err := m.overlapOptions(d)
		if err != nil || len(opts) != 3 || opts[0].ID != "note-3" || opts[2].ID != "block-1" {
			t.Fatal(opts, err)
		}
		d.choice = "block-1"
		m.prompt = ""
		if err = m.applyOverlap(d); err != nil || !slices.Equal(m.selectedIDs(), []string{"block-1", "note-3"}) || m.scene.Selected != "block-1" || state != m.journal.Snapshot() {
			t.Fatal("group/history", err)
		}
		m.overlapDraft = nil
		if m.beginOverlapAt(g, canvas.Point{}) == nil {
			t.Fatal("outside clip")
		}
	}
}
func TestOverlapHiddenLockedInvalidAndStale(t *testing.T) {
	m := groupedModel(t)
	m.layerInfo = map[string]layerMetadata{"group-1": {Hidden: true}}
	_ = m.beginOverlap(canvas.Point{X: 5, Y: 5})
	if opts, _ := m.overlapOptions(m.overlapDraft); len(opts) != 0 {
		t.Fatal("hidden")
	}
	m.prompt = ""
	m.overlapDraft = nil
	m.layerInfo["group-1"] = layerMetadata{Locked: true}
	_ = m.beginOverlap(canvas.Point{X: 5, Y: 5})
	if opts, _ := m.overlapOptions(m.overlapDraft); len(opts) != 0 {
		t.Fatal("locked")
	}
	d := m.overlapDraft
	d.x = "NaN"
	if _, err := m.overlapOptions(d); err == nil {
		t.Fatal("nan")
	}
	d.x = "0"
	m.prompt = ""
	m.layerInfo = nil
	if !errors.Is(m.applyOverlap(d), history.ErrStale) {
		t.Fatal("stale")
	}
	m.overlapDraft = nil
	_ = m.beginOverlap(canvas.Point{X: math.Inf(1)})
	if m.overlapDraft.message == "" {
		t.Fatal("inf")
	}
}
func TestOverlapDialogCancelAndRetiredCallbacks(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() {
		_ = f.m.beginOverlap(canvas.Point{X: 5, Y: 5})
		d := f.m.overlapDialog(f.a, f.ref)
		state := f.m.journal.Snapshot()
		ids := f.m.selectedIDs()
		geometryDialogControl(t, d, "overlap-cancel").Activate()
		f.m.prompt = "close"
		geometryDialogControl(t, d, "overlap-x").Change(ui.ChangeEvent{Text: "15"})
		geometryDialogControl(t, d, "overlap-select").Activate()
		if f.m.prompt != "close" || f.m.overlapDraft != nil || state != f.m.journal.Snapshot() || !slices.Equal(ids, f.m.selectedIDs()) {
			t.Fatal("retired callback")
		}
	})
}
