package main

import "github.com/dshills/vellum/ui"

type macLevelGallery struct {
	continuous, discrete, rating ui.LevelIndicator
	rejected, disabled           bool
	last                         string
}

func newMacLevelGallery(dark bool) *macLevelGallery {
	m := &macLevelGallery{}
	m.continuous = ui.LevelIndicator{ID: "mac-level-storage", Label: "Storage capacity", Mode: ui.LevelContinuous, Min: 0, Max: 100, Value: 62, Warning: ui.Some(float32(75)), Critical: ui.Some(float32(90)), Dark: dark}
	m.discrete = ui.LevelIndicator{ID: "mac-level-signal", Label: "Signal strength", Mode: ui.LevelDiscrete, Min: 0, Max: 100, Value: 62, Segments: 12, Warning: ui.Some(float32(75)), Critical: ui.Some(float32(90)), Dark: dark}
	m.rating = ui.LevelIndicator{ID: "mac-level-rating", Label: "Drawing rating", Mode: ui.LevelRating, Min: 0, Max: 5, Value: 3, AllowClear: true, Dark: dark}
	m.rating.Change = func(e ui.LevelEvent) {
		m.last = "Rating proposal rejected"
		if !m.rejected {
			m.rating.Value, m.rating.Unrated = e.Value, e.Unrated
			m.last = "Rating accepted"
		}
	}
	return m
}
func (m *macLevelGallery) view() ui.Node {
	m.continuous.Disabled, m.discrete.Disabled, m.rating.Disabled = m.disabled, m.disabled, m.disabled
	nodes := []ui.Node{{ID: "mac-level-title", Kind: ui.Label, Text: "Capacity and rating", Style: ui.Style{Typography: ui.Some(ui.Typography{Size: 20, Weight: 600})}}, {ID: "mac-level-help", Kind: ui.Label, Text: "Capacity changes color at 75 (warning) and 90 (critical). Tab to a star, then press Return or Space. Zero and Unrated are distinct."}}
	for _, p := range []*ui.LevelIndicator{&m.continuous, &m.discrete, &m.rating} {
		n, err := p.View()
		if err != nil {
			n = ui.Node{ID: p.ID + "-error", Kind: ui.Label, Text: err.Error()}
		}
		nodes = append(nodes, n)
	}
	nodes = append(nodes, ui.Node{ID: "mac-level-capacity", Kind: ui.Slider, Number: m.continuous.Value, Numeric: ui.Numeric{Min: 0, Max: 100, Step: 1}, Disabled: m.disabled, Accessibility: ui.Accessibility{Name: "Sample capacity, zero to one hundred"}, Change: func(e ui.ChangeEvent) {
		if !m.disabled {
			if v, ok := ui.ClampNumeric(e.Number, ui.Numeric{Min: 0, Max: 100, Step: 1}); ok {
				m.continuous.Value, m.discrete.Value = v, v
			}
		}
	}}, ui.Node{ID: "mac-level-reject", Kind: ui.Checkbox, Text: "Reject rating changes", Checked: m.rejected, Change: func(e ui.ChangeEvent) { m.rejected = e.Checked }}, ui.Node{ID: "mac-level-disable", Kind: ui.Checkbox, Text: "Disable level controls", Checked: m.disabled, Change: func(e ui.ChangeEvent) { m.disabled = e.Checked }}, ui.Node{ID: "mac-level-status", Kind: ui.Label, Text: m.last})
	return ui.Node{ID: "levels", Kind: ui.Column, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(20)), Gap: ui.Some(float32(18))}, Children: nodes}
}
