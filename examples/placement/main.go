// Placement is an independent consumer of Vellum's public Go APIs.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

func main() {
	m, err := newModel()
	if err != nil {
		fail(err)
	}
	defer m.palette.Close()
	a, err := app.New(app.Config{})
	if err != nil {
		fail(err)
	}
	defer a.Close()
	if err := a.Open(app.Window{ID: "placement", Title: "Vellum palette to canvas", Width: 1100, Height: 760, View: func() ui.Node { return view(a, m) }}); err != nil {
		fail(err)
	}
	if err := a.Run(context.Background(), cogent.Driver{Commands: m.commands}); err != nil {
		fail(err)
	}
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
func view(a *app.Application, m *model) ui.Node {
	update := func() { _ = a.Invalidate("placement") }
	invoke := func(id string) func() {
		return func() {
			life, err := a.WindowContext("placement")
			if err == nil {
				err = m.commands.Invoke(life, id)
			}
			if err != nil {
				m.status = err.Error()
			}
			update()
		}
	}
	tools := []ui.Node{{ID: "tools-heading", Kind: ui.Label, Text: "Components"}}
	for _, id := range []string{"block", "connector", "note"} {
		source := m.palette.Source(id)
		tools = append(tools, ui.Node{ID: "tool-" + id, Kind: ui.Button, Text: source.Label(), PlacementSource: source, Accessibility: ui.Accessibility{Name: source.Label(), Help: "Drag onto the canvas, or press Return to place at the current canvas center."}, Activate: func() {
			life, err := a.WindowContext("placement")
			if err == nil {
				err = source.PlaceKeyboard(life, m.target, m.scene.View.Center)
			}
			if err != nil {
				m.status = err.Error()
			}
			update()
		}})
	}
	grid := float32(0)
	if m.snap {
		grid = 20
	}
	return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{
		{ID: "toolbar", Kind: ui.Toolbar, Children: []ui.Node{
			{ID: "undo", Kind: ui.Button, Text: "Undo", Disabled: m.commands.Enabled("undo") != nil, Activate: invoke("undo")},
			{ID: "redo", Kind: ui.Button, Text: "Redo", Disabled: m.commands.Enabled("redo") != nil, Activate: invoke("redo")},
			{ID: "fit", Kind: ui.Button, Text: "Fit objects", Activate: func() { m.scene.Fit(); update() }},
			{ID: "reset", Kind: ui.Button, Text: "Reset view", Activate: func() { m.scene.View.Center = canvas.Point{}; m.scene.View.Zoom = 1; update() }},
			{ID: "snap", Kind: ui.Checkbox, Text: "Snap to grid", Checked: m.snap, Change: func(e ui.ChangeEvent) { m.snap = e.Checked; update() }},
			{ID: "lock", Kind: ui.Checkbox, Text: "Reject all drops", Checked: m.locked, Change: func(e ui.ChangeEvent) { m.locked = e.Checked; update() }},
			{ID: "reject-notes", Kind: ui.Checkbox, Text: "Reject notes", Checked: m.rejectNotes, Change: func(e ui.ChangeEvent) { m.rejectNotes = e.Checked; update() }},
		}},
		{ID: "help", Kind: ui.Label, Text: "Drag to place • drag the canvas or use arrow keys to pan • wheel to zoom • Escape cancels • Return on a tool places at center"},
		{ID: "workspace", Kind: ui.Row, Grow: true, Children: []ui.Node{
			{ID: "palette", Kind: ui.Column, MinWidth: 145, Children: tools},
			{ID: "drawing", Kind: ui.Canvas2D, Canvas: m.scene, CanvasGrid: grid, PlacementTarget: m.target, Grow: true, Accessibility: ui.Accessibility{Name: "Component canvas", Help: "Click to select; drag or use arrow keys to pan; scroll to zoom."}, Change: func(e ui.ChangeEvent) {
				m.status = fmt.Sprintf("Selected: %s • zoom %.0f%%", e.SelectedID, e.Number*100)
				update()
			}},
		}},
		{ID: "status", Kind: ui.StatusBar, Children: []ui.Node{{ID: "message", Kind: ui.Label, Text: fmt.Sprintf("%s • %d objects • Undo %d / Redo %d", m.status, len(m.scene.Objects), m.journal.Snapshot().UndoCount, m.journal.Snapshot().RedoCount)}}},
	}}
}
