package ui

import (
	"fmt"
	"image/color"
	"math"
)

// TrackInput is detached generated metadata for GradientWell and RangeSlider.
// Do not construct/mutate it. Each handle's fraction is independent of display
// scale; the native adapter routes proposals to current owner-held state.
type trackOrigin interface{ scalarTrackOrigin() }
type TrackHandle struct {
	ID, Label string
	Position  float32
	Color     color.NRGBA
}
type TrackInput struct {
	Current                       func() string
	source                        trackOrigin
	epoch                         uint64
	ID, ActiveID                  string
	Handles                       []TrackHandle
	Min, Max, Step, Width, Height float32
	Gradient                      bool
	Set                           func(string, float32)
	Adjust                        func(string, int)
	Select                        func(string)
}

func (t *TrackInput) SameSource(other *TrackInput) bool {
	if t == nil || other == nil {
		return t == other
	}
	if t.source == nil || t.source != other.source || t.epoch != other.epoch || t.ID != other.ID || len(t.Handles) != len(other.Handles) {
		return false
	}
	for _, h := range t.Handles {
		found := false
		for _, o := range other.Handles {
			found = found || o.ID == h.ID
		}
		if !found {
			return false
		}
	}
	return true
}
func (t *TrackInput) ActiveValue() float32 {
	for _, h := range t.Handles {
		if h.ID == t.ActiveID {
			return t.Min + h.Position*(t.Max-t.Min)
		}
	}
	return t.Min
}
func (t *TrackInput) ColorAt(position float32) color.NRGBA {
	if len(t.Handles) == 0 {
		return color.NRGBA{}
	}
	left := t.Handles[0]
	for _, right := range t.Handles[1:] {
		if position < right.Position {
			if position <= left.Position {
				return left.Color
			}
			f := (position - left.Position) / (right.Position - left.Position)
			mix := func(a, b uint8) uint8 { return uint8(math.Round(float64(float32(a) + (float32(b)-float32(a))*f))) }
			return color.NRGBA{R: mix(left.Color.R, right.Color.R), G: mix(left.Color.G, right.Color.G), B: mix(left.Color.B, right.Color.B), A: mix(left.Color.A, right.Color.A)}
		}
		left = right
	}
	return left.Color
}
func validateTrack(n Node) error {
	t := n.Track
	if t == nil {
		return nil
	}
	if n.Kind != Button || n.ID != t.ID || n.AngleDial != nil || n.ColumnResize != nil || n.PlacementSource != nil || n.ShortcutRecorder != nil || t.source == nil || t.Set == nil || t.Adjust == nil || t.Select == nil || t.Current == nil || len(t.Handles) < 2 || len(t.Handles) > 16 || t.Width < 200 || t.Width > 600 || t.Height < 24 || t.Height > 64 || !finite(t.Width) || !finite(t.Height) || !finite(t.Min) || !finite(t.Max) || !finite(t.Step) || t.Min >= t.Max || t.Step <= 0 || t.Step > t.Max-t.Min || !finite(t.ActiveValue()) || n.Number != t.ActiveValue() {
		return fmt.Errorf("ui: invalid generated scalar track")
	}
	found := false
	seen := map[string]bool{}
	last := float32(-1)
	for _, h := range t.Handles {
		if !validTokenText(h.ID) || !validTokenText(h.Label) || seen[h.ID] || !finite(h.Position) || h.Position < 0 || h.Position > 1 || h.Position < last {
			return fmt.Errorf("ui: invalid track handle")
		}
		seen[h.ID] = true
		found = found || h.ID == t.ActiveID
		last = h.Position
	}
	if !found {
		return fmt.Errorf("ui: track active handle absent")
	}
	return nil
}
func trackNode(t *TrackInput, label string, disabled bool) Node {
	return Node{ID: t.ID, Kind: Button, Track: t, Number: t.ActiveValue(), Disabled: disabled, Appearance: ButtonTool, Style: Style{Padding: Some(Insets{}), Border: Some(Border{}), GrowX: Some(float32(0))}, Accessibility: Accessibility{Name: label, Help: "Left/Right adjusts the active handle; Up/Down selects a handle; Shift adjusts ten steps; Home/End selects its limit."}, Activate: func() { t.Adjust(t.Current(), 1) }}
}
