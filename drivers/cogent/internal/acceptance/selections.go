package acceptance

import (
	"image/color"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
)

const SelectionListID = "component-list"
const SelectionCanvasID = "component-canvas"
const SelectionLeftID = "selection-left"
const SelectionRightID = "selection-right"

func newSelectionScene() (*canvas.Scene, error) {
	return canvas.New([]canvas.Rect{
		{ID: SelectionLeftID, X: -80, Y: -25, Width: 50, Height: 50, Color: color.RGBA{R: 200, G: 70, B: 30, A: 255}},
		{ID: SelectionRightID, X: 30, Y: -25, Width: 50, Height: 50, Color: color.RGBA{R: 30, G: 100, B: 200, A: 255}},
	})
}

func (w *Workspace) selectionView() ui.Node {
	return ui.Node{ID: "selections", Kind: ui.Row, Grow: true, Children: []ui.Node{
		{ID: SelectionListID, Kind: ui.ListBox, SelectedID: w.listSelection, Options: []ui.Option{
			{ID: "a", Label: "Rejected first"}, {ID: "b", Label: "Accepted initial"}, {ID: "locked", Label: "Locked", Disabled: true},
			{ID: "c", Label: "Resistor"}, {ID: "d", Label: "Capacitor"}, {ID: "e", Label: "Connector"}, {ID: "f", Label: "Label"}, {ID: "last", Label: "Last component"},
		}, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 160, Height: 120}), MaxSize: ui.Some(ui.Size{Width: 220, Height: 160})}, Change: func(e ui.ChangeEvent) {
			w.changes[e.ID]++
			if e.SelectedID != "a" {
				w.listSelection = e.SelectedID
			}
		}},
		{ID: SelectionCanvasID, Kind: ui.Canvas2D, Grow: true, Canvas: w.Drawing, CanvasObjects: []ui.CanvasObject{{ID: SelectionLeftID, Name: "Left object"}, {ID: SelectionRightID, Name: "Right object"}},
			Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 220, Height: 160})}, Accessibility: ui.Accessibility{Name: "Selectable components"},
			Change: func(e ui.ChangeEvent) { w.changes[e.ID]++ }},
	}}
}
