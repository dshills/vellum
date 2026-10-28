package main

import (
	"fmt"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/viewport"
	"math"
	"slices"
	"strconv"
	"strings"
)

type handleDrag struct {
	sequence      uint64
	axis          int
	origin        viewport.Vec3
	start         float64
	before, after document
	mode          string
	moved         bool
}

func axisVector(i int) viewport.Vec3 {
	v := viewport.Vec3{}
	switch i {
	case 0:
		v.X = 1
	case 1:
		v.Y = 1
	default:
		v.Z = 1
	}
	return v
}
func component(v viewport.Vec3, i int) float64 {
	switch i {
	case 0:
		return v.X
	case 1:
		return v.Y
	default:
		return v.Z
	}
}
func (m *workbench) addHandles() {
	if m.handleMode == "" || m.selectedIndex() < 0 {
		return
	}
	o := m.selected()
	if m.drag != nil {
		for _, v := range m.drag.after.Objects {
			if v.ID == o.ID {
				o = v
			}
		}
	}
	center := o.Bounds.Min.Add(o.Bounds.Max).Mul(.5)
	for i, c := range []struct{ r, g, b uint8 }{{220, 75, 75}, {70, 170, 105}, {65, 125, 225}} {
		pos := center
		switch i {
		case 0:
			pos.X = o.Bounds.Max.X + .18
		case 1:
			pos.Y = o.Bounds.Max.Y + .18
		case 2:
			pos.Z = o.Bounds.Max.Z + .18
		}
		half := viewport.Vec3{X: .065, Y: .065, Z: .065}
		m.scene.Objects = append(m.scene.Objects, viewport.Object{ID: fmt.Sprintf("forge-handle-%d", i), Name: []string{"X handle", "Y handle", "Z handle"}[i], Bounds: viewport.Box{Min: pos.Sub(half), Max: pos.Add(half)}, Color: rgb(c.r, c.g, c.b), Visible: true, Overlay: true, ScreenSize: 14})
	}
}

// Closest point between the pointer ray and a world axis. Near-parallel views
// are refused rather than producing an unstable jump; use numeric entry there.
func axisAt(c viewport.Camera, e cogent.GPUEvent, origin viewport.Vec3, axis int) (float64, bool) {
	r, err := c.RayAt(float64(e.Position.X), float64(e.Position.Y), float64(e.Size.X), float64(e.Size.Y))
	if err != nil {
		return 0, false
	}
	a := axisVector(axis)
	d := a.Dot(r.Direction)
	den := 1 - d*d
	if den < .015 {
		return 0, false
	}
	w := r.Origin.Sub(origin)
	v := (a.Dot(w) - d*r.Direction.Dot(w)) / den
	return v, !math.IsNaN(v) && !math.IsInf(v, 0)
}
func (m *workbench) transform(before document, axis int, delta float64, mode string) (document, error) {
	s := clone(before)
	if m.selectedIndex() < 0 || axis < 0 || axis > 2 || (mode != "move" && mode != "resize") {
		return s, fmt.Errorf("select bodies and a transform mode")
	}
	if math.IsNaN(delta) || math.IsInf(delta, 0) || math.Abs(delta) > 100 {
		return s, fmt.Errorf("transform outside range")
	}
	ids := m.selectionSet()
	for i := range s.Objects {
		if !ids[s.Objects[i].ID] {
			continue
		}
		o := &s.Objects[i]
		v := axisVector(axis).Mul(delta)
		if mode == "resize" {
			o.Bounds.Max = o.Bounds.Max.Add(v)
		} else {
			o.Bounds.Min = o.Bounds.Min.Add(v)
			o.Bounds.Max = o.Bounds.Max.Add(v)
		}
		if o.Bounds.Validate() != nil {
			return s, fmt.Errorf("size must stay positive")
		}
	}
	return s, nil
}
func (m *workbench) handleInput(e cogent.GPUEvent) bool {
	d := m.drag
	if d == nil {
		if e.Kind != cogent.GPUPointerDown || e.Button != cogent.GPUButtonPrimary || !e.Inside || m.prompt != "" {
			return false
		}
		id, err := m.scene.PickScaled(float64(e.Position.X), float64(e.Position.Y), float64(e.Size.X), float64(e.Size.Y), max(1, e.Scale))
		if err != nil || !strings.HasPrefix(id, "forge-handle-") {
			return false
		}
		i, err := strconv.Atoi(strings.TrimPrefix(id, "forge-handle-"))
		if err != nil || i < 0 || i > 2 {
			return false
		}
		o := m.selected()
		origin := o.Bounds.Min.Add(o.Bounds.Max).Mul(.5)
		start, ok := axisAt(m.scene.Camera, e, origin, i)
		if !ok {
			m.status = "Axis points toward camera; use numeric entry"
			m.changed()
			return true
		}
		m.drag = &handleDrag{sequence: e.Sequence, axis: i, origin: origin, start: start, before: clone(m.state), after: clone(m.state), mode: m.handleMode}
		return true
	}
	if e.Kind == cogent.GPUKey && (strings.EqualFold(e.Key, "Escape") || strings.EqualFold(e.Key, "Esc")) {
		m.drag = nil
		m.status = "Handle edit canceled"
		m.refresh()
		return true
	}
	if e.Sequence != d.sequence {
		return true
	}
	switch e.Kind {
	case cogent.GPUCancel:
		m.drag = nil
		m.status = "Handle edit canceled"
		m.refresh()
	case cogent.GPUPointerMove:
		at, ok := axisAt(m.scene.Camera, e, d.origin, d.axis)
		if !ok {
			return true
		}
		delta := at - d.start
		if m.snap {
			delta = math.Round(delta/.05) * .05
		}
		s, err := m.transform(d.before, d.axis, delta, d.mode)
		if err != nil {
			m.status = err.Error()
			return true
		}
		d.after = s
		d.moved = math.Abs(delta) > 1e-5
		m.scene.Objects = displayObjects(s)
		m.addHandles()
		_ = m.scene.Touch()
		m.status = fmt.Sprintf("%s %s · %.2f mm", d.mode, []string{"X", "Y", "Z"}[d.axis], delta*20)
		m.changed()
	case cogent.GPUPointerUp:
		m.drag = nil
		if e.Inside && d.moved {
			m.result(m.apply("Handle "+d.mode, d.after))
		} else {
			m.refresh()
		}
	case cogent.GPUScroll:
		return true
	}
	return true
}
func (m *workbench) handleControls() ui.Node {
	mode := func(v string) {
		if m.drag != nil {
			m.drag = nil
		}
		m.handleMode = v
		m.refresh()
	}
	controls := column("handles", row("handle-modes", button("handles-off", "Off", func() { mode("") }), button("handles-move", "Move", func() { mode("move") }), button("handles-resize", "Resize", func() { mode("resize") })), ui.Node{ID: "handle-snap", Kind: ui.Checkbox, Text: "Snap to 1 mm", Checked: m.snap, Change: func(e ui.ChangeEvent) { m.snap = e.Checked; m.changed() }}, text("handle-mode", "Handles: "+m.handleMode, 11, 400, rgb(117, 126, 141)))
	if m.handleMode == "" {
		return controls
	}
	for i, label := range []string{"X", "Y", "Z"} {
		i := i
		q := ui.QuantityField{ID: "axis-" + label, Label: label + " delta", Text: m.handleValues[i], Unit: "mm", Min: -2000, Max: 2000, Disabled: m.handleMode == "" || m.selectedIndex() < 0, Change: func(s string) { m.handleValues[i] = s; m.changed() }, Commit: func(v float64) {
			s, e := m.transform(m.state, i, v/20, m.handleMode)
			if e == nil {
				e = m.apply("Numeric "+m.handleMode, s)
			}
			m.result(e)
		}}
		controls.Children = append(controls.Children, must(q.View()))
	}
	return controls
}

func displayObjects(s document) []viewport.Object {
	objects := slices.Clone(s.Objects)
	visible := map[string]bool{}
	for i, f := range s.Features {
		visible[f.Object] = i < s.Marker
	}
	for i := range objects {
		if v, ok := visible[objects[i].ID]; ok {
			objects[i].Visible = objects[i].Visible && v
		}
	}
	return objects
}
