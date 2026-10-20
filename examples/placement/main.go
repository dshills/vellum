// Placement is an independent consumer of Vellum's public Go APIs.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

func run() {
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
	var windowRef app.WindowRef
	if windowRef, err = a.OpenWindow(app.Window{ID: "placement", Title: "Vellum palette to canvas", Width: 1100, Height: 760, View: func() ui.Node { return view(a, windowRef, m) }, CloseRequest: func() bool { return m.requestClose(a) }}); err != nil {
		fail(err)
	}
	if err := m.bindFiles(a, windowRef, nil); err != nil {
		fail(err)
	}
	if err := m.bindClipboard(a, windowRef); err != nil {
		fail(err)
	}
	if err := a.Run(context.Background(), cogent.Driver{Commands: m.commands}); err != nil {
		fail(err)
	}
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
func view(a *app.Application, ref app.WindowRef, m *model) ui.Node {
	update := func() { _ = a.Invalidate("placement") }
	busy := m.fileBusy()
	if m.canvasEditor == nil {
		m.canvasEditor = &canvasEditor{a: a, ref: ref, model: m}
	}
	navigatingDisabled := busy || m.move != nil
	filename := "Untitled.vellum"
	if m.files != nil && m.files.State().Path != "" {
		filename = filepath.Base(m.files.State().Path)
	}
	if m.journal.Snapshot().Dirty {
		filename += " • Unsaved changes"
	}
	invoke := func(id string) func() {
		return func() {
			life := ref.Context()
			if life.Err() != nil {
				return
			}
			err := m.commands.Invoke(life, id)
			if err != nil {
				m.status = err.Error()
			}
			update()
		}
	}
	tools := []ui.Node{{ID: "tools-heading", Kind: ui.Label, Text: "Components"}}
	for _, id := range []string{"block", "connector", "note"} {
		source := m.palette.Source(id)
		tools = append(tools, ui.Node{ID: "tool-" + id, Kind: ui.Button, Text: source.Label(), Disabled: navigatingDisabled, PlacementSource: source, Accessibility: ui.Accessibility{Name: source.Label(), Help: "Drag onto the canvas, or press Return to place at the current canvas center."}, Activate: func() {
			life := ref.Context()
			if life.Err() != nil {
				return
			}
			err := source.PlaceKeyboard(life, m.target, m.scene.View.Center)
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
	selection := m.selectedIDs()
	selectedSet := make(map[string]bool, len(selection))
	for _, id := range selection {
		selectedSet[id] = true
	}
	objects := make([]ui.CanvasObject, 0, len(m.scene.Objects))
	for _, object := range m.scene.Objects {
		kind, _, _ := strings.Cut(object.ID, "-")
		label := map[string]string{"block": "Block", "connector": "Connector", "note": "Note"}[kind]
		if selectedSet[object.ID] && len(selection) > 1 {
			label = "Selected " + label
		}
		if group := m.groups[object.ID]; group != "" {
			label = group + " · " + label
		}
		objects = append(objects, ui.CanvasObject{ID: object.ID, Name: label + " " + object.ID, Help: fmt.Sprintf("Select this component; size %.0f by %.0f.", object.Width, object.Height)})
	}
	dimensions := "Select a component"
	if i := m.selectedIndex(); i >= 0 {
		o := m.scene.Objects[i]
		dimensions = fmt.Sprintf("%s: %.0f × %.0f", o.ID, o.Width, o.Height)
	}
	if len(selection) > 1 {
		dimensions = fmt.Sprintf("%d selected • drag group corners to resize", len(selection))
	}
	selectedGroups := make(map[string]bool)
	for _, id := range selection {
		if group := m.groups[id]; group != "" {
			selectedGroups[group] = true
		}
	}
	if len(selectedGroups) > 0 {
		dimensions = fmt.Sprintf("%d selected • %d groups • corners resize", len(selection), len(selectedGroups))
	}
	help := "Group: Cmd+G • Ungroup: Cmd+Shift+G • Shift-drag selects • plain empty drag pans • corners resize selection"
	arrangementTools := []ui.Node{{ID: "arrange-label", Kind: ui.Label, Text: "Arrange:"}}
	for _, action := range arrangementActions {
		arrangementTools = append(arrangementTools, ui.Node{ID: action.id, Kind: ui.Button, Text: action.label, Disabled: m.commands.Enabled(action.id) != nil, Activate: invoke(action.id)})
	}
	root := ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{
		{ID: "file-tools", Kind: ui.Toolbar, Children: []ui.Node{
			{ID: "open", Kind: ui.Button, Text: "Open…", Disabled: m.files == nil || m.commands.Enabled("open") != nil, Activate: invoke("open")},
			{ID: "save", Kind: ui.Button, Text: "Save", Disabled: m.files == nil || m.commands.Enabled("save") != nil, Activate: invoke("save")},
			{ID: "save-as", Kind: ui.Button, Text: "Save As…", Disabled: m.files == nil || m.commands.Enabled("save-as") != nil, Activate: invoke("save-as")},
			{ID: "filename", Kind: ui.Label, Text: filename},
		}},
		{ID: "toolbar", Kind: ui.Toolbar, Children: []ui.Node{
			{ID: "undo", Kind: ui.Button, Text: "Undo", Disabled: m.commands.Enabled("undo") != nil, Activate: invoke("undo")},
			{ID: "redo", Kind: ui.Button, Text: "Redo", Disabled: m.commands.Enabled("redo") != nil, Activate: invoke("redo")},
			{ID: "fit", Kind: ui.Button, Text: "Fit objects", Disabled: navigatingDisabled, Activate: func() { m.scene.Fit(); update() }},
			{ID: "reset", Kind: ui.Button, Text: "Reset view", Disabled: navigatingDisabled, Activate: func() { m.scene.View.Center = canvas.Point{}; m.scene.View.Zoom = 1; update() }},
			{ID: "snap", Kind: ui.Checkbox, Text: "Snap to grid", Disabled: navigatingDisabled, Checked: m.snap, Change: func(e ui.ChangeEvent) { m.snap = e.Checked; update() }},
			{ID: "lock", Kind: ui.Checkbox, Text: "Reject all drops", Checked: m.locked, Change: func(e ui.ChangeEvent) { m.locked = e.Checked; update() }},
			{ID: "reject-notes", Kind: ui.Checkbox, Text: "Reject notes", Checked: m.rejectNotes, Change: func(e ui.ChangeEvent) { m.rejectNotes = e.Checked; update() }},
		}},
		{ID: "edit-tools", Kind: ui.Toolbar, Children: []ui.Node{
			{ID: "copy", Kind: ui.Button, Text: "Copy", Disabled: m.commands.Enabled("copy") != nil, Activate: invoke("copy")},
			{ID: "paste", Kind: ui.Button, Text: "Paste", Disabled: m.commands.Enabled("paste") != nil, Activate: invoke("paste")},
			{ID: "duplicate", Kind: ui.Button, Text: "Duplicate", Disabled: m.commands.Enabled("duplicate") != nil, Activate: invoke("duplicate")},
			{ID: "move-label", Kind: ui.Label, Text: "Move selected:"},
			{ID: "move-left", Kind: ui.Button, Text: "Left", Disabled: m.commands.Enabled("move-left") != nil, Activate: invoke("move-left")},
			{ID: "move-right", Kind: ui.Button, Text: "Right", Disabled: m.commands.Enabled("move-right") != nil, Activate: invoke("move-right")},
			{ID: "move-up", Kind: ui.Button, Text: "Up", Disabled: m.commands.Enabled("move-up") != nil, Activate: invoke("move-up")},
			{ID: "move-down", Kind: ui.Button, Text: "Down", Disabled: m.commands.Enabled("move-down") != nil, Activate: invoke("move-down")},
			{ID: "delete", Kind: ui.Button, Text: "Delete selected", Disabled: m.commands.Enabled("delete") != nil, Activate: invoke("delete")},
		}},
		{ID: "resize-tools", Kind: ui.Toolbar, Children: []ui.Node{
			{ID: "group", Kind: ui.Button, Text: "Group", Disabled: m.commands.Enabled("group") != nil, Activate: invoke("group")},
			{ID: "ungroup", Kind: ui.Button, Text: "Ungroup", Disabled: m.commands.Enabled("ungroup") != nil, Activate: invoke("ungroup")},
			{ID: "width-less", Kind: ui.Button, Text: "Width −", Disabled: m.commands.Enabled("width-less") != nil, Activate: invoke("width-less")},
			{ID: "width-more", Kind: ui.Button, Text: "Width +", Disabled: m.commands.Enabled("width-more") != nil, Activate: invoke("width-more")},
			{ID: "height-less", Kind: ui.Button, Text: "Height −", Disabled: m.commands.Enabled("height-less") != nil, Activate: invoke("height-less")},
			{ID: "height-more", Kind: ui.Button, Text: "Height +", Disabled: m.commands.Enabled("height-more") != nil, Activate: invoke("height-more")},
			{ID: "dimensions", Kind: ui.Label, Text: dimensions},
		}},
		{ID: "arrange-tools", Kind: ui.Toolbar, Children: arrangementTools},
		{ID: "help", Kind: ui.Label, Text: help},
		{ID: "workspace", Kind: ui.Row, Grow: true, Children: []ui.Node{
			{ID: "palette", Kind: ui.Column, MinWidth: 145, Children: tools},
			{ID: "drawing", Kind: ui.Canvas2D, Disabled: busy, Canvas: m.scene, CanvasObjects: objects, CanvasGrid: grid, PlacementTarget: m.target, Grow: true, Accessibility: ui.Accessibility{Name: "Component canvas", Help: "Shift-click toggles a component; Shift-drag empty canvas adds intersected components using a marquee; click empty space to clear selection; drag a selected component to move the selection; Option+Left/Right/Home/End selects one; drag selection corner handles to resize the set; plain empty drag or arrow keys pan; scroll zooms."}, Change: func(e ui.ChangeEvent) {
				m.managedSelection(e.SelectedID)
				m.status = fmt.Sprintf("Selected: %s • zoom %.0f%%", e.SelectedID, e.Number*100)
				update()
			}},
			{ID: "canvas-edit-events", Kind: ui.Extension, Custom: m.canvasEditor},
		}},
		{ID: "status", Kind: ui.StatusBar, Children: []ui.Node{{ID: "message", Kind: ui.Label, Text: fmt.Sprintf("%s • %d objects • %d selected • Undo %d / Redo %d", m.status, len(m.scene.Objects), len(selection), m.journal.Snapshot().UndoCount, m.journal.Snapshot().RedoCount)}}},
	}}
	root.Children = append(root.Children, m.filePrompt(a, ref))
	return root
}
