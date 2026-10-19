package calculator

import (
	"fmt"
	"github.com/dshills/vellum/ui"
	"image/color"
)

var background = color.NRGBA{R: 36, G: 37, B: 39, A: 255}
var text = color.NRGBA{R: 235, G: 235, B: 237, A: 255}

func Theme() ui.Theme {
	return ui.Theme{Default: ui.Style{Background: ui.Some(background), Typography: ui.Some(ui.Typography{Family: "Helvetica Neue", Size: 14}), Foreground: ui.Some(text)}, Kinds: map[ui.Kind]ui.Style{ui.Column: {Gap: ui.Some(float32(0))}, ui.Row: {Gap: ui.Some(float32(0))}}}
}
func fixed(w, h float32) ui.Style {
	return ui.Style{MinSize: ui.Some(ui.Size{Width: w, Height: h}), MaxSize: ui.Some(ui.Size{Width: w, Height: h}), Grow: ui.Some(float32(0)), Padding: ui.Some(ui.UniformInsets(0))}
}

func (m *Model) View() ui.Node {
	resultSize := float32(32)
	if len(m.Display()) > 12 {
		resultSize = 26
	}
	expr := ui.Node{ID: "expression", Kind: ui.Label, Text: m.Expression(), Style: fixed(210, 24)}
	expr.Style.TextAlign = ui.Some(ui.AlignEnd)
	expr.Style.Foreground = ui.Some(color.NRGBA{R: 160, G: 160, B: 164, A: 255})
	expr.Style.Typography = ui.Some(ui.Typography{Size: 21, Weight: 500})
	result := ui.Node{ID: "result", Kind: ui.Label, Text: m.Display(), Style: fixed(210, 44), Accessibility: ui.Accessibility{Name: "Result"}}
	result.Style.TextAlign = ui.Some(ui.AlignEnd)
	result.Style.Typography = ui.Some(ui.Typography{Size: resultSize, Weight: 500})

	keys := [][]string{{"delete", "AC", "%", "÷"}, {"7", "8", "9", "×"}, {"4", "5", "6", "−"}, {"1", "2", "3", "+"}, {"±", "0", ".", "="}}
	pad := ui.Node{ID: "keypad", Kind: ui.Column, Style: ui.Style{Gap: ui.Some(float32(6))}}
	for i, row := range keys {
		r := ui.Node{ID: fmt.Sprintf("key-row-%d", i), Kind: ui.Row, Style: ui.Style{Gap: ui.Some(float32(6))}}
		for j, label := range row {
			key := label
			n := ui.Node{ID: fmt.Sprintf("key-%d-%d", i, j), Kind: ui.Button, Appearance: ui.ButtonTool, Text: label, Style: fixed(46, 46), Accessibility: ui.Accessibility{Name: label}, Activate: func() { m.Press(key) }}
			n.Style.Typography = ui.Some(ui.Typography{Size: 24, Weight: 500})
			n.Style.CornerRadius = ui.Some(float32(24))
			n.Style.Border = ui.Some(ui.Border{})
			n.Style.Background = ui.Some(color.NRGBA{R: 58, G: 59, B: 61, A: 255})
			if label == "delete" {
				n.Text = ""
				n.Icon = ui.SymbolBackspace
				n.Accessibility.Name = "Delete last digit"
			}
			if label == "AC" {
				n.Accessibility.Name = "All clear"
			}
			r.Children = append(r.Children, n)
		}
		pad.Children = append(pad.Children, r)
	}
	main := ui.Node{ID: "calculator", Kind: ui.Column, Style: ui.Style{Padding: ui.Some(ui.Insets{Top: 62, Right: 10, Bottom: 10, Left: 10}), Background: ui.Some(background), GrowX: ui.Some(float32(0)), GrowY: ui.Some(float32(1))}, Children: []ui.Node{

		{ID: "display", Kind: ui.Column, Children: []ui.Node{expr, result}},
		{ID: "keypad-gap", Kind: ui.Column, Style: fixed(1, 4)}, pad,
	}}
	root := ui.Node{ID: "root", Kind: ui.Row, Grow: true, Style: ui.Style{Background: ui.Some(background), Gap: ui.Some(float32(0))}, Children: []ui.Node{main}}
	if m.HistoryOpen {
		h := ui.Node{ID: "history-list", Kind: ui.Column, Children: []ui.Node{{ID: "history-help", Kind: ui.Label, Text: "Choose a result to recall"}}}
		for i := len(m.history) - 1; i >= 0; i-- {
			index := i
			c := m.history[i]
			h.Children = append(h.Children, ui.Node{ID: fmt.Sprintf("history-%d", i), Kind: ui.Button, Appearance: ui.ButtonTool, Text: c.Expression + " = " + c.Result, Activate: func() { m.UseHistory(index); m.HistoryOpen = false }})
		}
		h.Children = append(h.Children, ui.Node{ID: "close-history", Kind: ui.Button, Text: "Done", Activate: func() { m.HistoryOpen = false }})
		root.Children = append(root.Children, ui.Node{ID: "history-dialog", Kind: ui.Dialog, Open: true, Text: "Calculation History", Dismiss: func() { m.HistoryOpen = false }, Children: []ui.Node{h}})
	}
	return root
}
