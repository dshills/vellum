package workspace

import (
	"fmt"
	"github.com/dshills/vellum/ui"
)

func (f *Fixture) button(id, label string) ui.Node {
	return ui.Node{ID: id, Kind: ui.Button, Text: label, Disabled: f.baseAvailable(id) != nil, Activate: func() { f.uiAction(id, "") }}
}
func (f *Fixture) View() ui.Node {
	return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{
		{ID: "toolbar", Kind: ui.Toolbar, Children: []ui.Node{f.button("open", "Open project"), f.button("sort", "Sort descending"), f.button("filter", "Toggle table filter"), f.button("tree-filter", "Filter outline"), f.button("cancel", "Cancel worker"), f.button("undo", "Undo"), f.button("redo", "Redo"), f.button("modal", "Modal"), f.button("close-main", "Close main")}},
		{ID: "dock", Kind: ui.Extension, Custom: f.Dock, Grow: true},
		{ID: "status", Kind: ui.StatusBar, Children: []ui.Node{{ID: "message", Kind: ui.Label, Text: f.status}}},
		{ID: "modal-dialog", Kind: ui.Dialog, Open: f.modal, Text: "Workspace modal", Dismiss: func() { f.uiAction("close-modal", "") }, Children: []ui.Node{{ID: "modal-message", Kind: ui.Label, Text: "Background input must remain blocked."}, f.button("close-modal", "Close modal")}},
	}}
}
func (f *Fixture) panel(id string) ui.Node {
	switch id {
	case "tree":
		return ui.Node{ID: "outline", Kind: ui.DataTree, Tree: f.Tree, Grow: true, Accessibility: ui.Accessibility{Name: "Stored project hierarchy"}, TreeChange: func(ui.TreeEvent) { f.touch() }, TreeDisclosure: func(e ui.TreeDisclosureEvent) {
			if err := f.changeTree(f.life, e.Change); err != nil {
				f.report(err)
			}
		}}
	case "grid":
		return ui.Node{ID: "table", Kind: ui.DataGrid, Grid: f.Grid, Grow: true, Accessibility: ui.Accessibility{Name: "Stored numeric records"}, GridChange: func(ui.GridEvent) { f.touch() }}
	case "notes":
		return ui.Node{ID: "editor", Kind: ui.MultilineText, Editor: f.Editor, Grow: true, SoftWrap: true, Accessibility: ui.Accessibility{Name: "Project notes"}}
	case "properties":
		return ui.Node{ID: "properties-root", Kind: ui.Column, Grow: true, Children: []ui.Node{{ID: "name-label", Kind: ui.Label, Text: "Project name"}, {ID: "name", Kind: ui.TextField, Text: f.name, Disabled: !f.loaded, Accessibility: ui.Accessibility{Name: "Project name"}, Change: func(e ui.ChangeEvent) { f.uiAction("name", e.Text) }}, {ID: "summary", Kind: ui.Label, Text: fmt.Sprintf("%d stored rows × %d fields", len(f.data.rows), Columns)}}}
	case "canvas":
		return ui.Node{ID: "canvas-root", Kind: ui.Column, Grow: true, Children: []ui.Node{{ID: "tool", Kind: ui.Button, Text: "Block", Disabled: !f.loaded, PlacementSource: f.Palette.Source("block"), Activate: func() {
			if e := f.Palette.Source("block").PlaceKeyboard(f.life, f.Target, f.Canvas.View.Center); e != nil {
				f.report(e)
			}
			f.touch()
		}}, {ID: "drawing", Kind: ui.Canvas2D, Canvas: f.Canvas, CanvasGrid: 20, PlacementTarget: f.Target, Grow: true, Accessibility: ui.Accessibility{Name: "Project canvas"}, Change: func(ui.ChangeEvent) { f.touch() }}}}
	}
	return ui.Node{ID: "unknown", Kind: ui.Label, Text: "Unknown panel"}
}
func (f *Fixture) Inspector() ui.Node {
	grid, _ := f.Grid.State()
	tree, _ := f.Tree.State()
	h := f.Journal.Snapshot()
	return ui.Node{ID: "inspector-root", Kind: ui.Column, Grow: true, Children: []ui.Node{{ID: "project", Kind: ui.Label, Text: f.name}, {ID: "rows", Kind: ui.Label, Text: fmt.Sprintf("Rows %d/%d", grid.Stamp.Rows, len(f.data.rows))}, {ID: "selection", Kind: ui.Label, Text: "Cell " + grid.Active.RowKey + " / " + grid.Active.ColumnID}, {ID: "tree", Kind: ui.Label, Text: "Tree " + tree.Selected}, {ID: "objects", Kind: ui.Label, Text: fmt.Sprintf("Objects %d; undo %d redo %d", len(f.Canvas.Objects), h.UndoCount, h.RedoCount)}, {ID: "worker", Kind: ui.Label, Text: f.status}}}
}
