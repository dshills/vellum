package main

import (
	"errors"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

func gridUnit(value float64) float64 {
	switch value {
	case 5, 10, 20, 50, 100:
		return value
	}
	return 20
}
func (m *model) spacing() float64 { return gridUnit(m.gridSpacing) }
func (m *model) setGridSpacing(id string) error {
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	values := map[string]float64{"grid-5": 5, "grid-10": 10, "grid-20": 20, "grid-50": 50, "grid-100": 100}
	value, ok := values[id]
	if !ok {
		return errors.New("unknown grid spacing")
	}
	m.gridSpacing = value
	m.status = fmt.Sprintf("Grid spacing %.0f world units", value)
	return nil
}
func (m *model) gridControls(ref app.WindowRef, update func()) []ui.Node {
	disabled := m.controlDisabled("interaction", m.selectionAvailable() != nil)
	options := []ui.Option{}
	for _, n := range []int{5, 10, 20, 50, 100} {
		options = append(options, ui.Option{ID: fmt.Sprintf("grid-%d", n), Label: fmt.Sprintf("%d world units", n)})
	}
	return []ui.Node{
		{ID: "grid-visible", Kind: ui.Checkbox, Text: "Grid", Checked: !m.gridHidden, Disabled: disabled, Change: func(e ui.ChangeEvent) {
			if ref.Context().Err() != nil || m.selectionAvailable() != nil {
				return
			}
			m.gridHidden = !e.Checked
			update()
		}},
		{ID: "grid-spacing", Kind: ui.MenuButton, Text: fmt.Sprintf("Grid: %.0f", m.spacing()), Options: options, Disabled: disabled, Change: func(e ui.ChangeEvent) {
			if ref.Context().Err() != nil {
				return
			}
			if err := m.setGridSpacing(e.SelectedID); err != nil {
				m.status = err.Error()
			}
			update()
		}},
	}
}
