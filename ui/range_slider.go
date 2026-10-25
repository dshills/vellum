package ui

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
	"math"
)

type RangeValue struct{ Lower, Upper float32 }
type RangeHandle string

const (
	RangeLower RangeHandle = "lower"
	RangeUpper RangeHandle = "upper"
)

type RangeEvent struct {
	ID     string
	Handle RangeHandle
	Value  float32
}

func (v RangeValue) Validate(bounds Numeric) error {
	if err := bounds.Validate(v.Lower); err != nil {
		return err
	}
	if err := bounds.Validate(v.Upper); err != nil {
		return err
	}
	if v.Lower > v.Upper {
		return fmt.Errorf("ui: range lower exceeds upper")
	}
	return nil
}

// Apply merges one endpoint into current accepted state, snapping relative to
// Min and clamping at the other endpoint. Equal endpoints are supported.
func (e RangeEvent) Apply(current RangeValue, bounds Numeric) (RangeValue, error) {
	if err := current.Validate(bounds); err != nil {
		return current, err
	}
	if !finite(e.Value) {
		return current, fmt.Errorf("ui: nonfinite range input")
	}
	value := math.Max(float64(bounds.Min), math.Min(float64(bounds.Max), float64(e.Value)))
	value = float64(bounds.Min) + math.Round((value-float64(bounds.Min))/float64(bounds.Step))*float64(bounds.Step)
	value = math.Max(float64(bounds.Min), math.Min(float64(bounds.Max), value))
	switch e.Handle {
	case RangeLower:
		current.Lower = min(float32(value), current.Upper)
	case RangeUpper:
		current.Upper = max(float32(value), current.Lower)
	default:
		return current, fmt.Errorf("ui: unknown range handle")
	}
	return current, nil
}

// RangeSlider is controlled. Retain it on one owner. Pointer/keyboard/AX and
// numeric fields propose RangeEvent; applications own history and acceptance.
type RangeSlider struct {
	ID, Label       string
	Value           RangeValue
	Numeric         Numeric
	Width           float32
	TickCount       int
	Small, Disabled bool
	IconImage       *canvas.TiledImage
	Change          func(RangeEvent)
	Style           Style
	active          RangeHandle
	identity        string
	epoch           uint64
	wasEnabled      bool
}

func (*RangeSlider) scalarTrackOrigin()                 {}
func RangeTrackID(id string) string                     { return id + "/range/track" }
func RangeFieldID(id string, handle RangeHandle) string { return id + "/range/" + string(handle) }
func (r *RangeSlider) validate() error {
	if r == nil || !validTokenText(r.ID) || !validTokenText(r.Label) || r.TickCount < 0 || r.TickCount == 1 || r.TickCount > 21 {
		return fmt.Errorf("ui: range requires ID/label and 0 or 2..21 ticks")
	}
	if r.Width != 0 && (!finite(r.Width) || r.Width < 200 || r.Width > 600) {
		return fmt.Errorf("ui: range track width must be 200..600")
	}
	return r.Value.Validate(r.Numeric)
}
func (r *RangeSlider) View() (Node, error) {
	if err := r.validate(); err != nil {
		return Node{}, err
	}
	enabled := !r.Disabled && r.Change != nil
	if r.identity != r.ID || r.wasEnabled != enabled {
		r.epoch++
		r.identity = r.ID
		r.wasEnabled = enabled
	}
	if r.active != RangeLower && r.active != RangeUpper {
		r.active = RangeLower
	}
	id, epoch := r.ID, r.epoch
	live := func() bool {
		return r.ID == id && r.epoch == epoch && !r.Disabled && r.Change != nil && r.validate() == nil
	}
	propose := func(handle RangeHandle, value float32) {
		if !live() {
			return
		}
		event := RangeEvent{ID: id, Handle: handle, Value: value}
		next, e := event.Apply(r.Value, r.Numeric)
		if e != nil || next == r.Value {
			return
		}
		if handle == RangeLower {
			event.Value = next.Lower
		} else {
			event.Value = next.Upper
		}
		r.Change(event)
	}
	width, height, fieldHeight := r.Width, float32(40), float32(30)
	if r.Small {
		height = 32
		fieldHeight = 24
	}
	if width == 0 {
		width = 320
		if r.Small {
			width = 240
		}
	}
	fraction := func(value float32) float32 {
		return float32((float64(value) - float64(r.Numeric.Min)) / (float64(r.Numeric.Max) - float64(r.Numeric.Min)))
	}
	track := &TrackInput{source: r, epoch: epoch, ID: RangeTrackID(id), ActiveID: string(r.active), Min: r.Numeric.Min, Max: r.Numeric.Max, Step: r.Numeric.Step, Width: width, Height: height, Ticks: r.TickCount, Handles: []TrackHandle{{ID: string(RangeLower), Label: "Lower endpoint", Position: fraction(r.Value.Lower)}, {ID: string(RangeUpper), Label: "Upper endpoint", Position: fraction(r.Value.Upper)}}}
	track.Current = func() string { return string(r.active) }
	track.Select = func(target string) {
		if live() && (target == string(RangeLower) || target == string(RangeUpper)) {
			r.active = RangeHandle(target)
		}
	}
	track.Set = func(target string, pos float32) {
		if live() && finite(pos) {
			value := float32(float64(r.Numeric.Min) + float64(max(0, min(1, pos)))*(float64(r.Numeric.Max)-float64(r.Numeric.Min)))
			propose(RangeHandle(target), value)
		}
	}
	track.Adjust = func(target string, steps int) {
		if !live() || steps < -100 || steps > 100 {
			return
		}
		value := r.Value.Lower
		if RangeHandle(target) == RangeUpper {
			value = r.Value.Upper
		}
		v := float64(value) + float64(steps)*float64(r.Numeric.Step)
		v = math.Max(float64(r.Numeric.Min), math.Min(float64(r.Numeric.Max), v))
		propose(RangeHandle(target), float32(v))
	}
	header := Node{ID: id + "/range/header", Kind: Row, Children: []Node{{ID: id + "/range/label", Kind: Label, Text: r.Label}}}
	if r.IconImage != nil {
		header.Children = append([]Node{{ID: id + "/range/art", Kind: Image, Image: r.IconImage, Fit: ImageContain, Text: r.Label, Style: Style{MinSize: Some(Size{Width: 24, Height: 24}), MaxSize: Some(Size{Width: 24, Height: 24})}}}, header.Children...)
	}
	fields := Node{ID: id + "/range/fields", Kind: Row}
	for _, h := range []RangeHandle{RangeLower, RangeUpper} {
		handle := h
		label, value := "Lower", r.Value.Lower
		if handle == RangeUpper {
			label, value = "Upper", r.Value.Upper
		}
		fieldID := RangeFieldID(id, handle)
		fields.Children = append(fields.Children, Node{ID: fieldID + "/field", Kind: Column, Children: []Node{{ID: fieldID + "/label", Kind: Label, Text: label}, {ID: fieldID, Kind: NumberField, Number: value, Numeric: r.Numeric, NumberButtons: NumberButtonsStepper, Accessibility: Accessibility{Name: r.Label + " " + label}, Style: Style{MinSize: Some(Size{Width: 100, Height: fieldHeight}), MaxSize: Some(Size{Width: 120}), GrowX: Some(float32(0))}, Disabled: !enabled, Change: func(e ChangeEvent) { propose(handle, e.Number) }}}})
	}
	n := Node{ID: id, Kind: Column, Style: Style{GrowX: Some(float32(0)), Gap: Some(float32(6))}.Overlay(r.Style), Children: []Node{header, trackNode(track, r.Label, !enabled), fields}}
	return n, Validate(n)
}
