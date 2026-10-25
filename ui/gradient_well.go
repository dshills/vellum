package ui

import (
	"encoding/base64"
	"fmt"
	"github.com/dshills/vellum/canvas"
	"image"
	"image/color"
	"math"
	"reflect"
	"slices"
)

type GradientStop struct {
	ID       string
	Position float32
	Color    color.NRGBA
}
type GradientValue struct {
	Stops []GradientStop
	Angle float32
}

func (v GradientValue) Validate() error {
	if len(v.Stops) < 2 || len(v.Stops) > 16 || !finite(v.Angle) || v.Angle < 0 || v.Angle >= 360 {
		return fmt.Errorf("ui: gradient requires 2..16 stops and angle [0,360)")
	}
	seen := map[string]bool{}
	last := float32(-1)
	for _, s := range v.Stops {
		if !validTokenText(s.ID) || seen[s.ID] || !finite(s.Position) || s.Position < last || s.Position < 0 || s.Position > 1 {
			return fmt.Errorf("ui: invalid gradient stop")
		}
		seen[s.ID] = true
		last = s.Position
	}
	return nil
}

type GradientEditKind uint8

const (
	GradientMove GradientEditKind = iota
	GradientColor
	GradientAdd
	GradientRemove
	GradientAngle
)

type GradientEdit struct {
	Kind            GradientEditKind
	StopID          string
	Position, Angle float32
	Color           ColorEdit
	Stop            GradientStop
}

// Apply merges one typed edit into current accepted state, preserving other stops.
func (e GradientEdit) Apply(current GradientValue) (GradientValue, error) {
	if err := current.Validate(); err != nil {
		return current, err
	}
	v := GradientValue{Stops: slices.Clone(current.Stops), Angle: current.Angle}
	index := -1
	for i, s := range v.Stops {
		if s.ID == e.StopID {
			index = i
		}
	}
	switch e.Kind {
	case GradientMove:
		if index < 0 || !finite(e.Position) {
			return current, fmt.Errorf("ui: absent/nonfinite gradient stop")
		}
		v.Stops[index].Position = max(0, min(1, e.Position))
	case GradientColor:
		if index < 0 {
			return current, fmt.Errorf("ui: absent gradient stop")
		}
		v.Stops[index].Color = e.Color.Apply(v.Stops[index].Color)
	case GradientAdd:
		if len(v.Stops) >= 16 || !validTokenText(e.Stop.ID) || !finite(e.Stop.Position) || e.Stop.Position < 0 || e.Stop.Position > 1 {
			return current, fmt.Errorf("ui: gradient stop budget/identity")
		}
		v.Stops = append(v.Stops, e.Stop)
	case GradientRemove:
		if index < 0 || len(v.Stops) <= 2 {
			return current, fmt.Errorf("ui: gradient needs two stops")
		}
		v.Stops = append(v.Stops[:index], v.Stops[index+1:]...)
	case GradientAngle:
		angle, ok := NormalizeAngle(e.Angle, 1)
		if !ok {
			return current, fmt.Errorf("ui: invalid gradient angle")
		}
		v.Angle = angle
	default:
		return current, fmt.Errorf("ui: invalid gradient edit")
	}
	slices.SortStableFunc(v.Stops, func(a, b GradientStop) int {
		if a.Position < b.Position {
			return -1
		}
		if a.Position > b.Position {
			return 1
		}
		return 0
	})
	if err := v.Validate(); err != nil {
		return current, err
	}
	return v, nil
}
func gradientHandles(v GradientValue) []TrackHandle {
	h := make([]TrackHandle, len(v.Stops))
	for i, s := range v.Stops {
		h[i] = TrackHandle{ID: s.ID, Label: fmt.Sprintf("Color stop %d", i+1), Position: s.Position, Color: s.Color}
	}
	return h
}
func GradientImage(v GradientValue, width, height int) (*canvas.TiledImage, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	if width < 1 || width > 1024 || height < 1 || height > 512 {
		return nil, fmt.Errorf("ui: gradient raster budget")
	}
	im := image.NewNRGBA(image.Rect(0, 0, width, height))
	t := TrackInput{Handles: gradientHandles(v)}
	angle := float64(v.Angle) * math.Pi / 180
	dx, dy := math.Cos(angle), math.Sin(angle)
	span := math.Abs(dx) + math.Abs(dy)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pos := .5 + ((float64(x)/float64(max(1, width-1))-.5)*dx+(float64(y)/float64(max(1, height-1))-.5)*dy)/span
			c := t.ColorAt(float32(pos))
			base := uint16(238)
			if (x/12+y/12)%2 != 0 {
				base = 198
			}
			blend := func(v uint8) uint8 { return uint8((uint16(v)*uint16(c.A) + base*(255-uint16(c.A)) + 127) / 255) }
			im.SetNRGBA(x, y, color.NRGBA{R: blend(c.R), G: blend(c.G), B: blend(c.B), A: 255})
		}
	}
	return canvas.NewTiledImage(im)
}

// GradientColorRequest is emitted on the owner. Native file/panel services run
// on workers; publish Accept back to the same owner. Stale sessions refuse it.
type GradientColorRequest struct {
	StopID string
	Color  color.NRGBA
	Accept func(color.NRGBA)
}
type GradientWell struct {
	ID, Label                        string
	Value                            GradientValue
	Disabled, Small, NativeColorBusy bool
	Change                           func(GradientEdit)
	NativeColor                      func(GradientColorRequest)
	Style, PopoverStyle              Style
	active, identity                 string
	open                             bool
	session, nextStop                uint64
	wasEnabled                       bool
	picker                           ColorPicker
	cacheValue                       GradientValue
	preview, swatch                  *canvas.TiledImage
}

func (*GradientWell) scalarTrackOrigin()    {}
func GradientWellButtonID(id string) string { return id + "/gradient/button" }
func GradientTrackID(id string) string      { return id + "/gradient/track" }
func (p *GradientWell) close()              { p.open = false; p.session++ }
func (p *GradientWell) View() (result Node, err error) {
	if p == nil {
		return Node{}, fmt.Errorf("ui: gradient well required")
	}
	defer func() {
		if err != nil && p.open {
			p.close()
		}
	}()
	if !validTokenText(p.ID) || !validTokenText(p.Label) {
		return Node{}, fmt.Errorf("ui: gradient ID/label required")
	}
	if err = p.Value.Validate(); err != nil {
		return Node{}, err
	}
	enabled := !p.Disabled && p.Change != nil
	if p.identity != p.ID || p.wasEnabled != enabled {
		p.close()
		p.identity = p.ID
		p.wasEnabled = enabled
	}
	id := p.ID
	if !enabled && p.open {
		p.close()
	}
	session := p.session
	live := func() bool { return p.ID == id && !p.Disabled && p.Change != nil && p.Value.Validate() == nil }
	editing := func() bool { return live() && p.open && p.session == session }
	index := func(id string) int {
		for i, s := range p.Value.Stops {
			if s.ID == id {
				return i
			}
		}
		return -1
	}
	if index(p.active) < 0 {
		p.active = p.Value.Stops[0].ID
	}
	active := p.active
	emit := func(e GradientEdit) {
		if !editing() {
			return
		}
		if _, e2 := e.Apply(p.Value); e2 == nil {
			p.Change(e)
		}
	}
	if p.preview == nil || !reflect.DeepEqual(p.cacheValue, p.Value) {
		p.preview, err = GradientImage(p.Value, 640, 128)
		if err != nil {
			return Node{}, err
		}
		p.swatch, err = GradientImage(p.Value, 48, 24)
		if err != nil {
			return Node{}, err
		}
		p.cacheValue = GradientValue{Stops: slices.Clone(p.Value.Stops), Angle: p.Value.Angle}
	}
	height := float32(40)
	if p.Small {
		height = 32
	}
	track := &TrackInput{source: p, epoch: session, ID: GradientTrackID(id), Handles: gradientHandles(p.Value), ActiveID: active, Min: 0, Max: 100, Step: 1, Width: 320, Height: height, Gradient: true}
	track.Current = func() string { return p.active }
	track.Select = func(target string) {
		if editing() && index(target) >= 0 {
			p.active = target
		}
	}
	track.Set = func(target string, pos float32) {
		i := index(target)
		if i >= 0 && finite(pos) {
			pos = float32(math.Round(float64(max(0, min(1, pos)))*100)) / 100
			if pos != p.Value.Stops[i].Position {
				emit(GradientEdit{Kind: GradientMove, StopID: target, Position: pos})
			}
		}
	}
	track.Adjust = func(target string, steps int) {
		i := index(target)
		if i >= 0 && steps >= -100 && steps <= 100 {
			track.Set(target, p.Value.Stops[i].Position+float32(steps)*.01)
		}
	}
	stop := p.Value.Stops[index(active)]
	p.picker = ColorPicker{ID: id + "/gradient/color/" + base64.RawURLEncoding.EncodeToString([]byte(active)), Label: "Selected stop", Color: stop.Color, Disabled: !enabled, Change: func(e ColorEdit) {
		if index(active) >= 0 {
			emit(GradientEdit{Kind: GradientColor, StopID: active, Color: e})
		}
	}}
	snapshotPicker := p.picker
	picker := snapshotPicker.View()
	p.picker.preview, p.picker.previewColor, p.picker.previewValid = snapshotPicker.preview, snapshotPicker.previewColor, snapshotPicker.previewValid
	controls := Node{ID: id + "/gradient/positions", Kind: Row, Children: []Node{{ID: id + "/gradient/position/" + base64.RawURLEncoding.EncodeToString([]byte(active)), Kind: NumberField, Number: stop.Position * 100, Numeric: Numeric{Min: 0, Max: 100, Step: 1}, NumberButtons: NumberButtonsStepper, Accessibility: Accessibility{Name: "Selected stop position percent"}, Change: func(e ChangeEvent) {
		if editing() {
			track.Set(active, e.Number/100)
		}
	}}, {ID: id + "/gradient/angle", Kind: NumberField, Number: p.Value.Angle, Numeric: Numeric{Min: 0, Max: 360, Step: 1}, NumberButtons: NumberButtonsStepper, Accessibility: Accessibility{Name: "Gradient angle degrees"}, Change: func(e ChangeEvent) { emit(GradientEdit{Kind: GradientAngle, Angle: e.Number}) }}}}
	for i := range controls.Children {
		label := "Position (%)"
		if i == 1 {
			label = "Angle (°)"
		}
		field := controls.Children[i]
		controls.Children[i] = Node{ID: field.ID + "/field", Kind: Column, Children: []Node{{ID: field.ID + "/label", Kind: Label, Text: label}, field}}
	}
	buttons := Node{ID: id + "/gradient/actions", Kind: Row, Children: []Node{{ID: id + "/gradient/add", Kind: Button, Text: "Add stop", Disabled: len(p.Value.Stops) >= 16, Activate: func() {
		if !editing() {
			return
		}
		left, right := float32(0), float32(1)
		gap := float32(-1)
		for i := 0; i < len(p.Value.Stops)-1; i++ {
			a, b := p.Value.Stops[i].Position, p.Value.Stops[i+1].Position
			if b-a > gap {
				gap = b - a
				left, right = a, b
			}
		}
		pos := (left + right) / 2
		var name string
		for {
			p.nextStop++
			name = fmt.Sprintf("stop-%d", p.nextStop)
			if index(name) < 0 {
				break
			}
		}
		emit(GradientEdit{Kind: GradientAdd, Stop: GradientStop{ID: name, Position: pos, Color: track.ColorAt(pos)}})
	}}, {ID: id + "/gradient/remove", Kind: Button, Text: "Remove stop", Disabled: len(p.Value.Stops) <= 2, Activate: func() { emit(GradientEdit{Kind: GradientRemove, StopID: active}) }}, {ID: id + "/gradient/done", Kind: Button, Text: "Done", Activate: func() {
		if editing() {
			p.close()
		}
	}}}}
	content := []Node{trackNode(track, p.Label+" stops", !enabled), controls, picker, buttons}
	if p.NativeColor != nil {
		content = append(content, Node{ID: id + "/gradient/native", Kind: Button, Text: "Show macOS Colors…", IconImage: p.swatch, Disabled: p.NativeColorBusy, Activate: func() {
			if !editing() || p.NativeColor == nil || p.NativeColorBusy || index(active) < 0 {
				return
			}
			c := p.Value.Stops[index(active)].Color
			p.close()
			epoch := p.session
			p.NativeColor(GradientColorRequest{StopID: active, Color: c, Accept: func(c color.NRGBA) {
				if live() && p.session == epoch && index(active) >= 0 {
					p.Change(GradientEdit{Kind: GradientColor, StopID: active, Color: ColorEdit{Channel: ColorHex, RGBA: c}})
				}
			}})
		}})
	}

	panel, err := (Popover{ID: id + "/gradient/popup", AnchorID: GradientWellButtonID(id), Label: p.Label, Open: p.open, Disabled: !enabled, Style: p.PopoverStyle, Children: []Node{{ID: id + "/gradient/scroll", Kind: Scroll, Style: Style{MinSize: Some(Size{Width: 350, Height: 250}), MaxSize: Some(Size{Height: 380})}, Children: content}}, Dismiss: func() {
		if p.ID == id && p.session == session {
			p.close()
		}
	}}).View()
	if err != nil {
		return Node{}, err
	}
	root := Node{ID: id, Kind: Column, Style: Style{GrowX: Some(float32(0)), Gap: Some(float32(8))}.Overlay(p.Style), Children: []Node{{ID: id + "/gradient/preview", Kind: Image, Image: p.preview, Fit: ImageContain, Text: p.Label + " preview", Style: Style{MinSize: Some(Size{Width: 320, Height: 64}), MaxSize: Some(Size{Width: 320, Height: 64})}}, {ID: GradientWellButtonID(id), Kind: Button, Text: p.Label + "…", IconImage: p.swatch, Disabled: !enabled, Activate: func() {
		if live() && p.session == session {
			p.open = !p.open
			p.session++
		}
	}}, panel}}
	return root, Validate(root)
}
