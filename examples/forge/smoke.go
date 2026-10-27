package main

import (
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/viewport"
	"image"
	"time"
)

// This smoke joins actual GPU readiness/readback with owner-routed synthetic
// input. It does not claim physical mouse, native event delivery or accessibility.
func smokeCheck(a *app.Application, ref app.WindowRef, m *workbench) error {
	deadline := time.Now().Add(40 * time.Second)
	ready := false
	for time.Now().Before(deadline) {
		time.Sleep(150 * time.Millisecond)
		var st cogent.GPUViewStats
		var err error
		if e := a.Do(func() { st, err = m.surface.Stats() }); e != nil {
			return e
		}
		if err == nil && st.Ready && st.Frames > 0 {
			ready = true
			break
		}
	}
	if !ready {
		return fmt.Errorf("Forge scene did not become ready")
	}
	var result error
	if e := a.Do(func() {
		st, e := m.surface.Stats()
		if e != nil {
			result = e
			return
		}
		im, e := m.surface.Snapshot()
		if e != nil {
			result = e
			return
		}
		if im.Bounds().Dx() < 100 || im.Bounds().Dy() < 100 {
			result = fmt.Errorf("small readback")
			return
		}
		colors := map[[4]uint8]bool{}
		for y := 0; y < im.Bounds().Dy(); y += 11 {
			for x := 0; x < im.Bounds().Dx(); x += 11 {
				c := im.NRGBAAt(x, y)
				colors[[4]uint8{c.R, c.G, c.B, c.A}] = true
			}
		}
		if len(colors) < 10 {
			result = fmt.Errorf("GPU readback appears blank: %d colors", len(colors))
			return
		}
		initial := m.scene.Camera
		size := st.Size
		p := image.Pt(size.X/2, size.Y/2)
		down := cogent.GPUEvent{Kind: cogent.GPUPointerDown, Button: cogent.GPUButtonPrimary, Sequence: 100, Position: p, Size: size, Scale: 1, Inside: true}
		m.surface.HandleInput(down)
		move := down
		move.Kind = cogent.GPUPointerMove
		move.Position.X += 45
		m.surface.HandleInput(move)
		up := move
		up.Kind = cogent.GPUPointerUp
		m.surface.HandleInput(up)
		if m.scene.Camera == initial {
			result = fmt.Errorf("orbit failed")
			return
		}
		before := m.scene.Camera
		m.surface.HandleInput(cogent.GPUEvent{Kind: cogent.GPUScroll, ScrollY: 40, Inside: true, Size: size, Scale: 1})
		if m.scene.Camera.Distance == before.Distance {
			result = fmt.Errorf("zoom failed")
			return
		}
		before = m.scene.Camera
		down.Sequence++
		down.Modifiers = cogent.GPUShift
		m.surface.HandleInput(down)
		move = down
		move.Kind = cogent.GPUPointerMove
		move.Position.Y += 40
		m.surface.HandleInput(move)
		up = move
		up.Kind = cogent.GPUPointerUp
		m.surface.HandleInput(up)
		if m.scene.Camera.Target == before.Target {
			result = fmt.Errorf("pan failed")
			return
		}
		m.orient(viewport.Front)
		m.fit()
		count := len(m.state.Objects)
		if e := m.add(); e != nil {
			result = e
			return
		}
		id := m.state.Selected
		if len(m.state.Objects) != count+1 {
			result = fmt.Errorf("add failed")
			return
		}
		if e := m.journal.Undo(); e != nil {
			result = e
			return
		}
		if e := m.journal.Redo(); e != nil {
			result = e
			return
		}
		if m.state.Selected != id {
			result = fmt.Errorf("redo changed identity")
			return
		}
		if e := m.resize(25.4); e != nil {
			result = e
			return
		}
		if e := m.visibility(id, false); e != nil {
			result = e
			return
		}
		if e := m.journal.Undo(); e != nil {
			result = e
			return
		}
		m.changed()
		fmt.Printf("Forge smoke: real GPU %dx%d, %d sampled colors; synthetic orbit/pan/zoom and command Undo/Redo passed\n", size.X, size.Y, len(colors))
	}); e != nil {
		return e
	}
	return result
}
