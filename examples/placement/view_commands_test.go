package main

import (
	"context"
	"github.com/dshills/vellum/canvas"
	"math"
	"slices"
	"testing"
)

func TestViewCommandsPreserveWorldCenterSelectionAndHistory(t *testing.T) {
	m := groupedModel(t)
	m.scene.View = canvas.Viewport{Width: 600, Height: 350, Zoom: 2, Center: canvas.Point{X: 123.25, Y: -81.5}}
	before := slices.Clone(m.scene.Objects)
	selection := m.selectedIDs()
	state := m.journal.Snapshot()
	center := m.scene.View.Center
	for _, test := range []struct {
		id   string
		zoom float64
	}{{"zoom-in", 2.5}, {"zoom-out", 2}, {"zoom-100", 1}} {
		if err := m.commands.Invoke(context.Background(), test.id); err != nil {
			t.Fatal(err)
		}
		if m.scene.View.Zoom != test.zoom || m.scene.View.Center != center {
			t.Fatal("center drift", m.scene.View)
		}
	}
	m.scene.View.Zoom = 1.73
	_ = m.viewCommand("zoom-100")
	if m.scene.View.Zoom != 1 {
		t.Fatal("inexact 100%")
	}
	m.scene.View.Width = 0
	invalid := m.scene.View
	if m.viewCommand("fit-selection") == nil || m.scene.View != invalid {
		t.Fatal("invalid viewport admitted")
	}
	m.scene.View.Width = 600
	m.scene.View.Zoom = 100
	_ = m.viewCommand("zoom-in")
	if m.scene.View.Zoom != 100 {
		t.Fatal("upper zoom clamp")
	}
	m.scene.View.Zoom = .01
	_ = m.viewCommand("zoom-out")
	if m.scene.View.Zoom != .01 {
		t.Fatal("lower zoom clamp")
	}
	m.setSelection([]string{"block-1"}, "block-1")
	selected := m.selectedIDs()
	if err := m.viewCommand("fit-selection"); err != nil {
		t.Fatal(err)
	}
	for _, o := range before {
		if slices.Contains(selected, o.ID) {
			p, q := m.scene.View.Screen(canvas.Point{X: o.X, Y: o.Y}), m.scene.View.Screen(canvas.Point{X: o.X + o.Width, Y: o.Y + o.Height})
			if p.X < 0 || p.Y < 0 || q.X > 600 || q.Y > 350 {
				t.Fatal("selected group not fitted", p, q)
			}
		}
	}
	_ = m.viewCommand("reset")
	if m.scene.View.Zoom != 1 || m.scene.View.Center != (canvas.Point{}) {
		t.Fatal("reset")
	}
	if !slices.Equal(before, m.scene.Objects) || state != m.journal.Snapshot() {
		t.Fatal("view changed geometry/history")
	}
	m.setSelection(selection, selection[0])
	m.prompt = "geometry"
	v := m.scene.View
	if m.viewCommand("zoom-in") == nil || v != m.scene.View {
		t.Fatal("busy view changed")
	}
	m.prompt = ""
	m.setSelection(nil, "")
	v = m.scene.View
	if m.viewCommand("fit-selection") == nil || v != m.scene.View {
		t.Fatal("empty fit changed")
	}
	m.scene.View.Zoom = math.NaN()
	if m.viewCommand("zoom-in") == nil {
		t.Fatal("invalid view accepted")
	}
}
func TestViewMenuAndZoomLabel(t *testing.T) {
	f := fileModel(t, false)
	f.owner(t, func() {
		f.m.scene.View.Zoom = 1.25
		root := view(f.a, f.ref, f.m)
		menu := geometryDialogControl(t, root, "view-menu")
		if len(menu.Options) != 5 || menu.Options[4].ID != "saved-views" {
			t.Fatal("view menu")
		}
		if geometryDialogControl(t, root, "zoom-percent").Text != "125%" {
			t.Fatal("missing zoom readout")
		}
	})
}
