package main

import (
	"bytes"
	"encoding/xml"
	"github.com/dshills/vellum/canvas"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type parsedSVG struct {
	XMLName xml.Name
	ViewBox string `xml:"viewBox,attr"`
	Rects   []struct {
		ID      string `xml:"id,attr"`
		X       string `xml:"x,attr"`
		Y       string `xml:"y,attr"`
		Width   string `xml:"width,attr"`
		Height  string `xml:"height,attr"`
		Fill    string `xml:"fill,attr"`
		Opacity string `xml:"fill-opacity,attr"`
		Title   string `xml:"title"`
	} `xml:"rect"`
}

func TestSVGExportWorldGeometryOrderVisibilityLockedAndEscaping(t *testing.T) {
	m := groupedModel(t)
	m.layerInfo = map[string]layerMetadata{"block-1": {Name: "Housing <&\"test\">"}, "connector-2": {Locked: true}, "note-3": {Hidden: true}}
	// A hidden member hides its whole flat group. Unlocking does not affect export.
	image, err := m.captureSVG()
	if err != nil {
		t.Fatal(err)
	}
	if len(image) != 1 || image[0].Rect.ID != "connector-2" {
		t.Fatal("hidden/locked export", image)
	}
	m.layerInfo["note-3"] = layerMetadata{}
	image, err = m.captureSVG()
	if err != nil {
		t.Fatal(err)
	}
	data, err := encodeSVG(image)
	if err != nil {
		t.Fatal(err)
	}
	var out parsedSVG
	if err = xml.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if out.XMLName.Space != "http://www.w3.org/2000/svg" || len(out.Rects) != 3 || out.Rects[0].ID != "block-1" || out.Rects[1].ID != "connector-2" || out.Rects[2].ID != "note-3" {
		t.Fatal("SVG paint order/namespace", out)
	}
	o := m.scene.Objects[0]
	if out.Rects[0].Title != m.layerLabel(o.ID) || !strings.Contains(string(data), "&lt;") || out.Rects[0].Fill != "#3075d1" {
		t.Fatal("title/color", out.Rects[0])
	}
	m.scene.View.Zoom = 3.75
	m.scene.View.Center = canvas.Point{X: 123, Y: -98}
	m.gridHidden = false
	m.gridSpacing = 100
	m.setSelection(nil, "")
	cameraImage, err := m.captureSVG()
	if err != nil {
		t.Fatal(err)
	}
	again, err := encodeSVG(cameraImage)
	if err != nil || !bytes.Equal(data, again) {
		t.Fatal("camera/selection/grid changed SVG", err)
	}
	image[0].Rect.X = 999
	if m.scene.Objects[0] != o {
		t.Fatal("export snapshot borrowed model")
	}
}
func TestSVGDetachedSavePreservesDrawingAndHistory(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{X: -33, Y: 44})
	f.place(t, "note", canvas.Point{X: 200, Y: 100})
	f.path = filepath.Join(t.TempDir(), "drawing.svg")
	f.owner(t, func() {
		if err := f.m.exportSVG(f.a, f.ref, nil); err != nil {
			t.Fatal(err)
		}
	})
	waitOutput(t, f)
	data, err := os.ReadFile(f.path)
	if err != nil {
		t.Fatal(err)
	}
	var svg parsedSVG
	if err = xml.Unmarshal(data, &svg); err != nil || len(svg.Rects) != 2 || svg.Rects[0].X != "-40" || svg.Rects[0].Y != "40" {
		t.Fatal("exported world geometry", err, svg)
	}
	f.owner(t, func() {
		if f.m.files.State().Path != "" || !f.m.journal.Snapshot().Dirty || f.m.journal.Snapshot().UndoCount != 2 {
			t.Fatal("export changed active document")
		}
	})
}
func TestSVGRejectsEmptyInvalidAndHonorsBounds(t *testing.T) {
	m, err := newModel()
	if err != nil {
		t.Fatal(err)
	}
	defer m.palette.Close()
	if _, err = m.captureSVG(); err == nil {
		t.Fatal("empty SVG admitted")
	}
	if _, err = encodeSVG(nil); err == nil {
		t.Fatal("empty encode")
	}
	bad := []svgComponent{{Rect: canvas.Rect{ID: "a", Width: 0, Height: 20}}}
	if _, err = encodeSVG(bad); err == nil {
		t.Fatal("invalid geometry")
	}
	item := svgComponent{Rect: canvas.Rect{ID: "a", Width: 8, Height: 8, Color: color.RGBA{R: 64, A: 128}}, Name: "transparent"}
	data, err := encodeSVG([]svgComponent{item})
	if err != nil {
		t.Fatal(err)
	}
	var out parsedSVG
	_ = xml.Unmarshal(data, &out)
	if out.Rects[0].Fill != "#7f0000" || out.Rects[0].Opacity == "1" {
		t.Fatal("premultiplied color export", out)
	}
	if _, err = encodeSVG([]svgComponent{item, item}); err == nil {
		t.Fatal("duplicate IDs")
	}
}
