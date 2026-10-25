package main

import (
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"image/color"
)

func (m *macComponentsGallery) gradientView() ui.Node {
	if m.gradient == nil {
		g := &ui.GradientWell{ID: "preview/gradient", Label: "Drawing gradient", Value: ui.GradientValue{Stops: []ui.GradientStop{{ID: "red", Position: 0, Color: color.NRGBA{R: 240, G: 70, B: 80, A: 255}}, {ID: "blue", Position: .5, Color: color.NRGBA{R: 50, G: 130, B: 245, A: 255}}, {ID: "green", Position: 1, Color: color.NRGBA{R: 35, G: 195, B: 130, A: 150}}}, Angle: 0}}
		g.Change = func(e ui.GradientEdit) {
			next, err := e.Apply(g.Value)
			if err != nil {
				m.status = err.Error()
				return
			}
			g.Value = next
			m.count++
			m.status = "Gradient updated"
		}
		g.NativeColor = func(request ui.GradientColorRequest) {
			a, w := m.essentials.application, m.essentials.window
			if a == nil {
				m.status = "Native Colors unavailable offscreen"
				return
			}
			ticket, err := a.ReserveWindowCompletion(w, nil)
			if err != nil {
				m.status = err.Error()
				return
			}
			g.NativeColorBusy = true
			go func() {
				service, err := a.ColorPanels()
				value := request.Color
				if err == nil {
					value, err = service.ChooseColor(w.Context(), w, app.ColorPanelOptions{Title: "Gradient stop color", Color: request.Color, ShowsAlpha: true})
				}
				_ = ticket.Publish(func() {
					g.NativeColorBusy = false
					if err == nil {
						request.Accept(value)
					} else {
						m.status = fmt.Sprint("Gradient color: ", err)
					}
					_ = a.Invalidate(w.ID())
				})
			}()
		}
		m.gradient = g
	}
	m.gradient.PopoverStyle = ui.Style{Background: ui.Some(m.theme().Palette.Value.Popup)}
	n, e := m.gradient.View()
	if e != nil {
		return ui.Node{ID: "preview/gradient-error", Kind: ui.Label, Text: e.Error()}
	}
	return ui.Node{ID: "preview/gradient-page", Kind: ui.Column, Style: ui.Style{GrowX: ui.Some(float32(1))}, Children: []ui.Node{{ID: "preview/gradient-help", Kind: ui.Label, Text: "Open the gradient well. Drag its colored stops, edit position/angle, add/remove stops, and adjust RGBA. Up/Down chooses a stop; Left/Right changes its position. Native Colors closes the inline editor and applies to that same stop."}, {ID: "preview/gradient-small", Kind: ui.Checkbox, Text: "Small track", Checked: m.gradient.Small, Change: func(e ui.ChangeEvent) { m.gradient.Small = e.Checked }}, n}}
}
