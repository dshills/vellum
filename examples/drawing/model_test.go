package main

import (
	"github.com/dshills/vellum/canvas"
	"testing"
)

func TestIndependentConsumerWorkloads(t *testing.T) {
	m, err := newDrawingModel()
	if err != nil {
		t.Fatal(err)
	}
	view := canvas.Viewport{Width: 800, Height: 600, Zoom: 1, Center: canvas.Point{X: 400, Y: 300}}
	m.mode = "schematic"
	mesh, err := m.prepare(view)
	if err != nil {
		t.Fatal(err)
	}
	_, indices := mesh.Counts()
	if m.query.TotalObjects != 100000 || m.query.VisibleObjects > 2000 || indices != 6*m.query.VisibleObjects {
		t.Fatal(m.query, indices)
	}
	m.mode = "waveform"
	view.Center = canvas.Point{}
	mesh, err = m.prepare(view)
	if err != nil {
		t.Fatal(err)
	}
	if m.waveform.Samples() != 1000000 || m.columns != 800 {
		t.Fatal("workload source/resolution")
	}
	if v, ix := mesh.Counts(); v != 3200 || ix != 4800 {
		t.Fatal(v, ix)
	}
	view.Zoom = 4
	view.Center.X = 80
	_, err = m.prepare(view)
	if err != nil || m.sampleStart != 480000 || m.sampleEnd != 680000 || m.columns != 800 {
		t.Fatal("visible sample interval", m.sampleStart, m.sampleEnd, m.columns, err)
	}
	m.mode = "mesh"
	mesh, err = m.prepare(view)
	if err != nil || mesh != m.fixture {
		t.Fatal("retained fixture identity", err)
	}
}
