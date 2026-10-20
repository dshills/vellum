package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/document"
)

const drawingByteLimit = 256 << 10

// Version 1 stores the example's fixed component types, not toolkit internals.
// Viewport dimensions belong to the receiving window and are never restored.
type drawingFile struct {
	Format   string          `json:"format"`
	Version  int             `json:"version"`
	Objects  []drawingObject `json:"objects"`
	Center   canvas.Point    `json:"center"`
	Zoom     float64         `json:"zoom"`
	Selected string          `json:"selected,omitempty"`
	Snap     bool            `json:"snap"`
	NextID   uint64          `json:"next_id"`
}
type drawingObject struct {
	ID   string  `json:"id"`
	Kind string  `json:"kind"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}

func (f drawingFile) scene() (*canvas.Scene, error) {
	finite := func(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
	if f.Format != "vellum-placement" || f.Version != 1 || len(f.Objects) > 512 || !finite(f.Zoom) || f.Zoom < .01 || f.Zoom > 100 || !finite(f.Center.X) || !finite(f.Center.Y) || math.Abs(f.Center.X) > 1e12 || math.Abs(f.Center.Y) > 1e12 {
		return nil, errors.New("invalid or unsupported Vellum drawing")
	}
	objects := make([]canvas.Rect, 0, len(f.Objects))
	for _, o := range f.Objects {
		c, ok := componentByKind(o.Kind)
		kind, suffix, split := strings.Cut(o.ID, "-")
		n, err := strconv.ParseUint(suffix, 10, 64)
		if !ok || !split || kind != o.Kind || err != nil || n == 0 || n > f.NextID || strconv.FormatUint(n, 10) != suffix || !finite(o.X) || !finite(o.Y) || math.Abs(o.X) > 10000 || math.Abs(o.Y) > 10000 {
			return nil, fmt.Errorf("invalid drawing component %q", o.ID)
		}
		objects = append(objects, canvas.Rect{ID: o.ID, X: o.X, Y: o.Y, Width: c.Width, Height: c.Height, Color: c.Color})
	}
	scene, err := canvas.New(objects)
	if err != nil {
		return nil, err
	}
	if err := scene.Select(f.Selected); err != nil {
		return nil, err
	}
	scene.View.Center, scene.View.Zoom = f.Center, f.Zoom
	return scene, nil
}

func drawingCodec(m *model) document.Codec[drawingFile] {
	return document.Codec[drawingFile]{
		Capture: func() (drawingFile, error) {
			if m.palette.State().Active || m.pending != nil {
				return drawingFile{}, errors.New("finish or cancel the drag before saving")
			}
			f := drawingFile{Format: "vellum-placement", Version: 1, Objects: make([]drawingObject, 0, len(m.scene.Objects)), Center: m.scene.View.Center, Zoom: m.scene.View.Zoom, Selected: m.scene.Selected, Snap: m.snap, NextID: m.next}
			for _, o := range m.scene.Objects {
				kind, _, _ := strings.Cut(o.ID, "-")
				f.Objects = append(f.Objects, drawingObject{ID: o.ID, Kind: kind, X: o.X, Y: o.Y})
			}
			_, err := f.scene()
			return f, err
		},
		Encode: func(f drawingFile) ([]byte, error) {
			if _, err := f.scene(); err != nil {
				return nil, err
			}
			data, err := json.MarshalIndent(f, "", "  ")
			if err != nil {
				return nil, err
			}
			if len(data)+1 > drawingByteLimit {
				return nil, document.ErrTooLarge
			}
			return append(data, '\n'), nil
		},
		Decode: func(data []byte) (drawingFile, error) {
			var f drawingFile
			if len(data) > drawingByteLimit {
				return f, document.ErrTooLarge
			}
			d := json.NewDecoder(bytes.NewReader(data))
			d.DisallowUnknownFields()
			if err := d.Decode(&f); err != nil {
				return f, err
			}
			if err := d.Decode(new(any)); err != io.EOF {
				return f, errors.New("drawing contains trailing data")
			}
			_, err := f.scene()
			return f, err
		},
		Install: func(f drawingFile) error {
			scene, err := f.scene()
			if err != nil {
				return err
			}
			// Preserve the bound Scene identity and current window dimensions.
			m.scene.Objects, m.scene.Selected = scene.Objects, scene.Selected
			m.scene.View.Center, m.scene.View.Zoom = scene.View.Center, scene.View.Zoom
			m.next, m.snap = f.NextID, f.Snap
			return nil
		},
	}
}
