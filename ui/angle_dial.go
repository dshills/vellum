package ui

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
	"image/color"
	"math"
)

type AngleEvent struct {
	ID      string
	Degrees float32
}

// AngleDialInput is a detached immutable description of the generated dial.
// Applications configure AngleDial; the native adapter never reads its owner.
type AngleDialInput struct {
	source              *AngleDial
	ID                  string
	Degrees, Step, Size float32
	Set                 func(float32)
	Adjust              func(int)
}

func (d *AngleDialInput) SameSource(other *AngleDialInput) bool {
	if d == nil || other == nil {
		return d == other
	}
	return d.source != nil && d.source == other.source && d.ID == other.ID
}

// AngleDial is a controlled circular degree editor. Zero points upward; angles
// increase clockwise and wrap at 360. Step must divide 360, between 0.1 and 180 (default 1 degree).
// Retain one instance on one owner. Live proposals do not own application history.
type AngleDial struct {
	ID, Label           string
	Degrees, Step, Size float32
	Disabled            bool
	IconImage           *canvas.TiledImage
	Change              func(AngleEvent)
	Style               Style
}

func AngleDialButtonID(id string) string { return id + "/angle/dial" }
func AngleDialFieldID(id string) string  { return id + "/angle/degrees" }
func angleStep(step float32) float32 {
	if step == 0 {
		return 1
	}
	return step
}
func NormalizeAngle(value, step float32) (float32, bool) {
	step = angleStep(step)
	if !finite(value) || !finite(step) || step < .1 || step > 180 || math.Abs(360/float64(step)-math.Round(360/float64(step))) > 0.0001 {
		return 0, false
	}
	v := math.Mod(float64(value), 360)
	if v < 0 {
		v += 360
	}
	v = math.Mod(math.Round(v/float64(step))*float64(step), 360)
	if v == 0 {
		return 0, true
	}
	return float32(v), true
}
func (d *AngleDial) View() (Node, error) {
	if d == nil || !validTokenText(d.ID) || !validTokenText(d.Label) || !finite(d.Degrees) || d.Degrees < 0 || d.Degrees >= 360 {
		return Node{}, fmt.Errorf("ui: angle dial needs ID/label and degrees in [0,360)")
	}
	step := angleStep(d.Step)
	if _, ok := NormalizeAngle(d.Degrees, step); !ok {
		return Node{}, fmt.Errorf("ui: angle step must divide 360")
	}
	size := d.Size
	if size == 0 {
		size = 96
	}
	if !finite(size) || size < 64 || size > 160 {
		return Node{}, fmt.Errorf("ui: angle dial size must be 64..160")
	}
	id := d.ID
	live := func() bool { return d.ID == id && !d.Disabled && d.Change != nil }
	set := func(value float32) {
		if !live() {
			return
		}
		if normalized, ok := NormalizeAngle(value, d.Step); ok {
			d.Change(AngleEvent{ID: id, Degrees: normalized})
		}
	}
	input := &AngleDialInput{source: d, ID: id, Degrees: d.Degrees, Step: step, Size: size, Set: set, Adjust: func(delta int) {
		if delta >= -360 && delta <= 360 {
			set(d.Degrees + float32(delta)*angleStep(d.Step))
		}
	}}
	header := Node{ID: id + "/angle/header", Kind: Row, Children: []Node{{ID: id + "/angle/label", Kind: Label, Text: d.Label}}}
	if d.IconImage != nil {
		header.Children = append([]Node{{ID: id + "/angle/icon", Kind: Image, Image: d.IconImage, Text: d.Label, Fit: ImageContain, Style: Style{MinSize: Some(Size{Width: 24, Height: 24}), MaxSize: Some(Size{Width: 24, Height: 24})}}}, header.Children...)
	}
	n := Node{ID: id, Kind: Column, Disabled: !live(), Style: Style{GrowX: Some(float32(0)), Gap: Some(float32(8))}.Overlay(d.Style), Children: []Node{header, {ID: id + "/angle/controls", Kind: Row, Children: []Node{
		{ID: AngleDialButtonID(id), Kind: Button, Number: d.Degrees, AngleDial: input, Appearance: ButtonTool, Accessibility: Accessibility{Name: d.Label, Help: "Circular angle. Drag clockwise; arrows adjust, Shift adjusts ten steps; Home resets; End selects the last step."}, Activate: func() { input.Adjust(1) }, Style: Style{Padding: Some(Insets{}), Border: Some(Border{}), Background: Some(color.NRGBA{}), CornerRadius: Some(float32(8)), GrowX: Some(float32(0))}},
		{ID: id + "/angle/entry", Kind: Row, Children: []Node{
			{ID: AngleDialFieldID(id), Kind: NumberField, Number: d.Degrees, Numeric: Numeric{Min: 0, Max: 360, Step: step}, NumberButtons: NumberButtonsStepper, Accessibility: Accessibility{Name: d.Label + " in degrees"}, Change: func(e ChangeEvent) { set(e.Number) }},
			{ID: id + "/angle/unit", Kind: Label, Text: "°"},
		}},
	}}}}
	if err := Validate(n); err != nil {
		return Node{}, err
	}
	return n, nil
}
