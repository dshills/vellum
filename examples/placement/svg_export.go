package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/document"
	"image/color"
	"math"
	"path/filepath"
	"strconv"
)

type svgComponent struct {
	Rect canvas.Rect
	Name string
}

func (m *model) svgAvailable() error {
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	if len(m.visibleObjects()) == 0 {
		return errors.New("no visible components to export")
	}
	return nil
}
func (m *model) captureSVG() ([]svgComponent, error) {
	if err := m.svgAvailable(); err != nil {
		return nil, err
	}
	objects := m.visibleObjects()
	out := make([]svgComponent, 0, len(objects))
	for _, o := range objects {
		out = append(out, svgComponent{Rect: o, Name: m.layerLabel(o.ID)})
	}
	return out, nil
}

// Geometry-only SVG: original world coordinates/paint order/colors and escaped
// names. Camera, grid, editor chrome, selection and resize handles are omitted.
func encodeSVG(objects []svgComponent) ([]byte, error) {
	if len(objects) == 0 || len(objects) > 512 {
		return nil, errors.New("SVG requires 1..512 visible components")
	}
	left, top, right, bottom := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	seen := map[string]bool{}
	for _, v := range objects {
		o := v.Rect
		if o.ID == "" || seen[o.ID] || !validPosition(o.X, o.Y) || !validSize(o.Width, o.Height) {
			return nil, errors.New("invalid SVG component")
		}
		seen[o.ID] = true
		left = math.Min(left, o.X)
		top = math.Min(top, o.Y)
		right = math.Max(right, o.X+o.Width)
		bottom = math.Max(bottom, o.Y+o.Height)
	}
	number := func(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }
	attr := func(name, value string) xml.Attr { return xml.Attr{Name: xml.Name{Local: name}, Value: value} }
	root := xml.StartElement{Name: xml.Name{Local: "svg"}, Attr: []xml.Attr{attr("xmlns", "http://www.w3.org/2000/svg"), attr("version", "1.1"), attr("viewBox", number(left)+" "+number(top)+" "+number(right-left)+" "+number(bottom-top)), attr("width", number(right-left)), attr("height", number(bottom-top))}}
	var b bytes.Buffer
	b.WriteString(xml.Header)
	e := xml.NewEncoder(&b)
	if err := e.EncodeToken(root); err != nil {
		return nil, err
	}
	for _, v := range objects {
		o := v.Rect
		c := color.NRGBAModel.Convert(o.Color).(color.NRGBA)
		rect := xml.StartElement{Name: xml.Name{Local: "rect"}, Attr: []xml.Attr{attr("id", o.ID), attr("x", number(o.X)), attr("y", number(o.Y)), attr("width", number(o.Width)), attr("height", number(o.Height)), attr("fill", fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)), attr("fill-opacity", number(float64(c.A)/255))}}
		if err := e.EncodeToken(rect); err != nil {
			return nil, err
		}
		if err := e.EncodeElement(v.Name, xml.StartElement{Name: xml.Name{Local: "title"}}); err != nil {
			return nil, err
		}
		if err := e.EncodeToken(rect.End()); err != nil {
			return nil, err
		}
	}
	if err := e.EncodeToken(root.End()); err != nil {
		return nil, err
	}
	if err := e.Flush(); err != nil {
		return nil, err
	}
	b.WriteByte('\n')
	if b.Len() > drawingByteLimit {
		return nil, document.ErrTooLarge
	}
	return b.Bytes(), nil
}
func (m *model) exportSVG(a *app.Application, ref app.WindowRef, store document.Store) error {
	image, err := m.captureSVG()
	if err != nil {
		return err
	}
	codec := document.Codec[[]svgComponent]{Capture: func() ([]svgComponent, error) { return image, nil }, Encode: encodeSVG, Decode: func([]byte) ([]svgComponent, error) { return nil, app.ErrUnavailable }, Install: func([]svgComponent) error { return app.ErrUnavailable }}
	name := "Drawing"
	if m.files != nil && m.files.State().Path != "" {
		name = drawingSuggestion(filepath.Base(m.files.State().Path))
	}
	return startDetachedSave(m, a, ref, detachedSaveConfig[[]svgComponent]{title: "Export SVG", filename: name, extension: "svg", success: "Exported SVG", codec: codec, store: store})
}
