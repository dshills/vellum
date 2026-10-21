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
	defer func() { m.cancelLayerDrag(m.layerDrag) }()
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
	var recoveryErr error
	var recentErr error
	if err := a.Do(func() {
		m.recent, recentErr = newRecentDrawings(a, windowRef, recentConfig{})
		if recentErr != nil {
			m.status = "Recent drawings unavailable: " + recentErr.Error()
		}
		m.recovery, recoveryErr = newPlacementRecovery(a, windowRef, m, recoveryConfig{})
		if recoveryErr != nil {
			m.status = "Recovery unavailable: " + recoveryErr.Error()
		}
	}); err != nil {
		fail(err)
	}
	defer m.recovery.Stop(false)
	defer m.recent.Stop()
	runErr := a.Run(context.Background(), cogent.Driver{Commands: m.commands})
	if err := m.recent.Stop(); err != nil {
		fmt.Fprintln(os.Stderr, "Recent drawings:", err)
	}
	if err := m.recovery.Stop(runErr == nil && m.removeRecoveryOnClose); err != nil {
		fmt.Fprintln(os.Stderr, "Recovery cleanup:", err)
	}
	if runErr != nil {
		fail(runErr)
	}
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
func view(a *app.Application, ref app.WindowRef, m *model) ui.Node {
	update := func() { _ = a.Invalidate("placement") }
	// Capture gates execution while preserving native layout/presentation.
	busy := (m.fileBusy() && m.layerDrag == nil) || m.recovery.pending()
	if m.canvasEditor == nil {
		m.canvasEditor = &canvasEditor{a: a, ref: ref, model: m}
	}
	if m.layerEditor == nil {
		m.layerEditor = &layerEditor{a: a, ref: ref, model: m}
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
	for _, object := range m.visibleObjects() {
		if m.layerLocked(object.ID) {
			continue
		}
		kind, _, _ := strings.Cut(object.ID, "-")
		label := map[string]string{"block": "Block", "connector": "Connector", "note": "Note"}[kind]
		if selectedSet[object.ID] && len(selection) > 1 {
			label = "Selected " + label
		}
		if group := m.groups[object.ID]; group != "" {
			label = group + " · " + label
		}
		objects = append(objects, ui.CanvasObject{ID: object.ID, Name: label + " " + m.layerLabel(object.ID), Help: fmt.Sprintf("Select this component; size %.0f by %.0f.", object.Width, object.Height)})
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
		arrangementTools = append(arrangementTools, ui.Node{ID: action.id, Kind: ui.Button, Text: action.label, Disabled: m.commandDisabled(action.id), Activate: invoke(action.id)})
	}
	display := m.canvasForView()
	root := ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{
		{ID: "file-tools", Kind: ui.Toolbar, Children: []ui.Node{
			{ID: "open", Kind: ui.Button, Text: "Open…", Disabled: m.files == nil || m.commandDisabled("open"), Activate: invoke("open")},
			{ID: "save", Kind: ui.Button, Text: "Save", Disabled: m.files == nil || m.commandDisabled("save"), Activate: invoke("save")},
			{ID: "save-as", Kind: ui.Button, Text: "Save As…", Disabled: m.files == nil || m.commandDisabled("save-as"), Activate: invoke("save-as")},
			m.recentMenu(a, ref, update),
			{ID: "filename", Kind: ui.Label, Text: filename},
		}},
		{ID: "toolbar", Kind: ui.Toolbar, Children: []ui.Node{
			{ID: "undo", Kind: ui.Button, Text: "Undo", Disabled: m.commandDisabled("undo"), Activate: invoke("undo")},
			{ID: "redo", Kind: ui.Button, Text: "Redo", Disabled: m.commandDisabled("redo"), Activate: invoke("redo")},
			{ID: "fit", Kind: ui.Button, Text: "Fit objects", Disabled: navigatingDisabled, Activate: func() {
				if m.fileBusy() || m.move != nil {
					return
				}
				m.fitVisible()
				update()
			}},
			{ID: "reset", Kind: ui.Button, Text: "Reset view", Disabled: navigatingDisabled, Activate: func() {
				if m.fileBusy() || m.move != nil {
					return
				}
				m.scene.View.Center = canvas.Point{}
				m.scene.View.Zoom = 1
				update()
			}},
			{ID: "snap", Kind: ui.Checkbox, Text: "Snap to grid", Disabled: navigatingDisabled, Checked: m.snap, Change: func(e ui.ChangeEvent) {
				if m.fileBusy() || m.move != nil {
					return
				}
				m.snap = e.Checked
				update()
			}},
			{ID: "lock", Kind: ui.Checkbox, Text: "Reject all drops", Checked: m.locked, Change: func(e ui.ChangeEvent) { m.locked = e.Checked; update() }},
			{ID: "reject-notes", Kind: ui.Checkbox, Text: "Reject notes", Checked: m.rejectNotes, Change: func(e ui.ChangeEvent) { m.rejectNotes = e.Checked; update() }},
		}},
		{ID: "edit-tools", Kind: ui.Toolbar, Children: []ui.Node{
			{ID: "cut", Kind: ui.Button, Text: "Cut", Disabled: m.commandDisabled("cut"), Activate: invoke("cut")},
			{ID: "copy", Kind: ui.Button, Text: "Copy", Disabled: m.commandDisabled("copy"), Activate: invoke("copy")},
			{ID: "paste", Kind: ui.Button, Text: "Paste", Disabled: m.commandDisabled("paste"), Activate: invoke("paste")},
			{ID: "duplicate", Kind: ui.Button, Text: "Duplicate", Disabled: m.commandDisabled("duplicate"), Activate: invoke("duplicate")},
			{ID: "move-label", Kind: ui.Label, Text: "Move selected:"},
			{ID: "move-left", Kind: ui.Button, Text: "Left", Disabled: m.commandDisabled("move-left"), Activate: invoke("move-left")},
			{ID: "move-right", Kind: ui.Button, Text: "Right", Disabled: m.commandDisabled("move-right"), Activate: invoke("move-right")},
			{ID: "move-up", Kind: ui.Button, Text: "Up", Disabled: m.commandDisabled("move-up"), Activate: invoke("move-up")},
			{ID: "move-down", Kind: ui.Button, Text: "Down", Disabled: m.commandDisabled("move-down"), Activate: invoke("move-down")},
			{ID: "delete", Kind: ui.Button, Text: "Delete selected", Disabled: m.commandDisabled("delete"), Activate: invoke("delete")},
		}},
		{ID: "stack-tools", Kind: ui.Toolbar, Children: []ui.Node{
			{ID: "group", Kind: ui.Button, Text: "Group", Disabled: m.commandDisabled("group"), Activate: invoke("group")},
			{ID: "ungroup", Kind: ui.Button, Text: "Ungroup", Disabled: m.commandDisabled("ungroup"), Activate: invoke("ungroup")},
			{ID: "bring-front", Kind: ui.Button, Text: "Bring to front", Disabled: m.commandDisabled("bring-front"), Activate: invoke("bring-front")},
			{ID: "send-back", Kind: ui.Button, Text: "Send to back", Disabled: m.commandDisabled("send-back"), Activate: invoke("send-back")},
			{ID: "bring-forward", Kind: ui.Button, Text: "Bring forward", Disabled: m.commandDisabled("bring-forward"), Activate: invoke("bring-forward")},
			{ID: "send-backward", Kind: ui.Button, Text: "Send backward", Disabled: m.commandDisabled("send-backward"), Activate: invoke("send-backward")},
		}},
		{ID: "resize-tools", Kind: ui.Toolbar, Children: []ui.Node{
			{ID: "width-less", Kind: ui.Button, Text: "Width −", Disabled: m.commandDisabled("width-less"), Activate: invoke("width-less")},
			{ID: "width-more", Kind: ui.Button, Text: "Width +", Disabled: m.commandDisabled("width-more"), Activate: invoke("width-more")},
			{ID: "height-less", Kind: ui.Button, Text: "Height −", Disabled: m.commandDisabled("height-less"), Activate: invoke("height-less")},
			{ID: "height-more", Kind: ui.Button, Text: "Height +", Disabled: m.commandDisabled("height-more"), Activate: invoke("height-more")},
			{ID: "dimensions", Kind: ui.Label, Text: dimensions},
		}},
		{ID: "arrange-tools", Kind: ui.Toolbar, Children: arrangementTools},
		{ID: "help", Kind: ui.Label, Text: help},
		{ID: "workspace", Kind: ui.Row, Grow: true, Children: []ui.Node{
			{ID: "palette", Kind: ui.Column, MinWidth: 145, Children: tools},
			{ID: "drawing", Kind: ui.Canvas2D, Disabled: busy, Canvas: display, CanvasObjects: objects, CanvasGrid: grid, PlacementTarget: m.target, Grow: true, Accessibility: ui.Accessibility{Name: "Component canvas", Help: "Shift-click toggles a component; Shift-drag empty canvas adds intersected components using a marquee; click empty space to clear selection; drag a selected component to move the selection; Option+Left/Right/Home/End selects one; drag selection corner handles to resize the set; plain empty drag or arrow keys pan; scroll zooms."}, Change: func(e ui.ChangeEvent) {
				m.scene.View = display.View
				m.managedSelection(e.SelectedID)
				m.status = fmt.Sprintf("Selected: %s • zoom %.0f%%", e.SelectedID, e.Number*100)
				update()
			}},
			m.layersPanel(ref, update, invoke),
			{ID: "canvas-edit-events", Kind: ui.Extension, Custom: m.canvasEditor},
			{ID: "layers-drag-events", Kind: ui.Extension, Custom: m.layerEditor},
		}},
		{ID: "status", Kind: ui.StatusBar, Children: []ui.Node{{ID: "message", Kind: ui.Label, Text: fmt.Sprintf("%s • %d objects • %d selected • Undo %d / Redo %d", m.status, len(m.scene.Objects), len(selection), m.journal.Snapshot().UndoCount, m.journal.Snapshot().RedoCount)}}},
	}}
	root.Children = append(root.Children, m.filePrompt(a, ref), m.renameDialog(a, ref), m.recovery.promptNode())
	return root
}
