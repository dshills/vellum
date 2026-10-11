package main

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/dshills/vellum/canvas"
)

type drawingModel struct {
	mode                   string
	waveform               *canvas.Waveform
	schematic              *canvas.RectIndex
	fixture                *canvas.Mesh
	imageLayer             *canvas.ImageLayer
	query                  canvas.QueryStats
	columns                int
	sampleStart, sampleEnd int
}

func newDrawingModel() (*drawingModel, error) {
	samples := make([]float32, 1_000_000)
	for i := range samples {
		samples[i] = float32(.85*math.Sin(float64(i)*.077) + .1*math.Sin(float64(i)*.003))
	}
	samples[123456] = 1
	waveform, err := canvas.NewWaveform(samples)
	if err != nil {
		return nil, err
	}
	objects := make([]canvas.Rect, 100_000)
	for i := range objects {
		objects[i] = canvas.Rect{ID: fmt.Sprint(i), X: float64(i%400) * 20, Y: float64(i/400) * 20, Width: 12, Height: 8, Color: color.RGBA{67, 197, 166, 255}}
	}
	schematic, err := canvas.NewRectIndex(objects)
	if err != nil {
		return nil, err
	}
	fixture, err := fixtureMesh()
	if err != nil {
		return nil, err
	}
	source := image.NewNRGBA(image.Rect(0, 0, 4096, 4096))
	for i := 0; i < len(source.Pix); i += 4 {
		copy(source.Pix[i:i+4], []byte{40, 100, 160, 255})
	}
	tiled, err := canvas.NewTiledImage(source)
	if err != nil {
		return nil, err
	}
	layer := &canvas.ImageLayer{Image: tiled, Bounds: canvas.Bounds{Min: canvas.Point{X: -256, Y: -256}, Max: canvas.Point{X: 256, Y: 256}}}
	return &drawingModel{mode: "mesh", waveform: waveform, schematic: schematic, fixture: fixture, imageLayer: layer}, nil
}
func fixtureMesh() (*canvas.Mesh, error) {
	v := []canvas.Vertex{{Position: canvas.Point{X: -160, Y: -100}, Color: color.NRGBA{255, 0, 0, 255}}, {Position: canvas.Point{X: 0, Y: -100}, Color: color.NRGBA{0, 255, 0, 255}}, {Position: canvas.Point{X: -80, Y: 80}, Color: color.NRGBA{0, 0, 255, 255}}}
	ix := []uint32{0, 1, 2}
	for _, q := range []struct {
		x, y, w, h float64
		c          color.NRGBA
	}{{40, -80, 120, 120, color.NRGBA{0, 0, 255, 255}}, {80, -40, 120, 120, color.NRGBA{255, 0, 0, 128}}} {
		j := uint32(len(v))
		v = append(v, canvas.Vertex{Position: canvas.Point{X: q.x, Y: q.y}, Color: q.c}, canvas.Vertex{Position: canvas.Point{X: q.x + q.w, Y: q.y}, Color: q.c}, canvas.Vertex{Position: canvas.Point{X: q.x + q.w, Y: q.y + q.h}, Color: q.c}, canvas.Vertex{Position: canvas.Point{X: q.x, Y: q.y + q.h}, Color: q.c})
		ix = append(ix, j, j+1, j+2, j, j+2, j+3)
	}
	stroke, err := canvas.StrokePolyline([]canvas.Point{{X: -180, Y: 110}, {X: -100, Y: 140}, {X: -20, Y: 110}, {X: 20, Y: 160}}, 6, color.NRGBA{255, 190, 69, 255})
	if err != nil {
		return nil, err
	}
	offset := uint32(len(v))
	v = append(v, stroke.Vertices()...)
	for _, i := range stroke.Indices() {
		ix = append(ix, offset+i)
	}
	return canvas.NewMesh(v, ix)
}
func (m *drawingModel) prepare(v canvas.Viewport) (*canvas.Mesh, error) {
	switch m.mode {
	case "mesh", "image":
		return m.fixture, nil
	case "schematic":
		mesh, stats, err := m.schematic.Mesh(v)
		m.query = stats
		return mesh, err
	case "waveform":
		bounds, err := v.Bounds()
		if err != nil {
			return nil, err
		}
		left, right := max(-500, bounds.Min.X), min(500, bounds.Max.X)
		if left >= right {
			m.columns = 0
			m.sampleStart = 0
			m.sampleEnd = 0
			return canvas.NewMesh(nil, nil)
		}
		samples := m.waveform.Samples()
		start := max(0, int(math.Floor((left+500)*float64(samples)/1000)))
		end := min(samples, int(math.Ceil((right+500)*float64(samples)/1000)))
		m.sampleStart = start
		m.sampleEnd = end
		m.columns = max(1, min(int(math.Ceil((right-left)*v.Zoom)), canvas.MaxEnvelopeColumns, end-start))
		return m.waveform.EnvelopeMesh(start, end, m.columns, canvas.Bounds{Min: canvas.Point{X: -500 + float64(start)*1000/float64(samples), Y: -120}, Max: canvas.Point{X: -500 + float64(end)*1000/float64(samples), Y: 120}}, color.NRGBA{67, 197, 166, 255})

	}
	return nil, fmt.Errorf("unknown drawing mode")
}
