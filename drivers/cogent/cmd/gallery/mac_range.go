package main

import (
	"fmt"
	"github.com/dshills/vellum/ui"
)

type macRangeGallery struct {
	ranges           []*ui.RangeSlider
	reject, disabled bool
}

func (m *macComponentsGallery) rangeView() ui.Node {
	if m.ranges == nil {
		p := &macRangeGallery{ranges: []*ui.RangeSlider{{ID: "preview/range/regular", Label: "Distance range", Numeric: ui.Numeric{Min: 0, Max: 100, Step: 1}, Value: ui.RangeValue{Lower: 25, Upper: 75}, TickCount: 5, IconImage: m.artwork[ui.SymbolDocument]}, {ID: "preview/range/small", Label: "Temperature range", Numeric: ui.Numeric{Min: -20, Max: 40, Step: .5}, Value: ui.RangeValue{Lower: 0, Upper: 25}, TickCount: 7, Small: true, IconImage: m.artwork[ui.SymbolFolder]}}}
		for _, r := range p.ranges {
			target := r
			r.Change = func(e ui.RangeEvent) {
				if p.reject {
					m.status = "Range edit rejected"
					return
				}
				next, err := e.Apply(target.Value, target.Numeric)
				if err != nil {
					m.status = err.Error()
					return
				}
				target.Value = next
				m.count++
				m.status = fmt.Sprintf("%s: %.1f to %.1f", target.Label, next.Lower, next.Upper)
			}
		}
		m.ranges = p
	}
	p := m.ranges
	children := []ui.Node{{ID: "preview/range-help", Kind: ui.Label, Text: "Two handles share one Mac-style track. Drag either endpoint; they clamp at each other. Up/Down chooses a handle, Left/Right adjusts, Shift ten steps. Numeric fields use the same validation."}, {ID: "preview/range-options", Kind: ui.Row, Children: []ui.Node{{ID: "preview/range-disabled", Kind: ui.Checkbox, Text: "Disable ranges", Checked: p.disabled, Change: func(e ui.ChangeEvent) { p.disabled = e.Checked }}, {ID: "preview/range-reject", Kind: ui.Checkbox, Text: "Reject changes", Checked: p.reject, Change: func(e ui.ChangeEvent) { p.reject = e.Checked }}}}}
	for _, r := range p.ranges {
		r.Disabled = p.disabled
		n, e := r.View()
		if e != nil {
			return ui.Node{ID: "preview/range-error", Kind: ui.Label, Text: e.Error()}
		}
		children = append(children, n)
	}
	return ui.Node{ID: "preview/range-page", Kind: ui.Column, Children: children}
}
