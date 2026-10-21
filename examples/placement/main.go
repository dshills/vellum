// Placement is an independent consumer of Vellum's public Go APIs.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dshills/vellum/app"
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
	tools = append(tools,
		ui.Node{ID: "lock", Kind: ui.Checkbox, Text: "Reject all drops", Checked: m.locked, Change: func(e ui.ChangeEvent) { m.locked = e.Checked; update() }},
		ui.Node{ID: "reject-notes", Kind: ui.Checkbox, Text: "Reject notes", Checked: m.rejectNotes, Change: func(e ui.ChangeEvent) { m.rejectNotes = e.Checked; update() }},
	)
	grid := float32(0)
	if !m.gridHidden {
		grid = float32(m.spacing())
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
	selectionAction := "adds"
	if m.selectionSubtract {
		selectionAction = "subtracts"
	}
	help := "Shift-drag empty canvas " + selectionAction + " selection • empty drag pans • scroll zooms • Shift+corner locks ratio"
	display := m.canvasForView()
	root := ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{
		{ID: "file-tools", Kind: ui.Toolbar, Children: []ui.Node{
			{ID: "new", Kind: ui.Button, Text: "New", Disabled: m.files == nil || m.commandDisabled("new"), Activate: invoke("new")},
			{ID: "open", Kind: ui.Button, Text: "Open…", Disabled: m.files == nil || m.commandDisabled("open"), Activate: invoke("open")},
			{ID: "save", Kind: ui.Button, Text: "Save", Disabled: m.files == nil || m.commandDisabled("save"), Activate: invoke("save")},
			{ID: "save-as", Kind: ui.Button, Text: "Save As…", Disabled: m.files == nil || m.commandDisabled("save-as"), Activate: invoke("save-as")},
			m.recentMenu(a, ref, update),
			{ID: "filename", Kind: ui.Label, Text: filename, Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 320})}},
		}},
		{ID: "toolbar", Kind: ui.Toolbar, Children: []ui.Node{
			{ID: "undo", Kind: ui.Button, Text: m.historyCaption("Undo"), Disabled: m.commandDisabled("undo"), Activate: invoke("undo")},
			{ID: "redo", Kind: ui.Button, Text: m.historyCaption("Redo"), Disabled: m.commandDisabled("redo"), Activate: invoke("redo")},
			{ID: "fit", Kind: ui.Button, Text: "Fit", Disabled: m.commandDisabled("fit"), Activate: invoke("fit")},
			{ID: "reset", Kind: ui.Button, Text: "Reset", Disabled: m.commandDisabled("reset"), Activate: invoke("reset")},
			m.actionMenu("view-menu", "Zoom", []ui.Option{{ID: "zoom-in", Label: "Zoom In"}, {ID: "zoom-out", Label: "Zoom Out"}, {ID: "zoom-100", Label: "100%"}, {ID: "fit-selection", Label: "Fit Selection"}}, invoke),
			{ID: "zoom-percent", Kind: ui.Label, Text: fmt.Sprintf("%.0f%%", m.scene.View.Zoom*100), Accessibility: ui.Accessibility{Name: "Canvas zoom"}},
			{ID: "snap", Kind: ui.Checkbox, Text: "Snap", Disabled: navigatingDisabled, Checked: m.snap, Change: func(e ui.ChangeEvent) {
				if m.fileBusy() || m.move != nil {
					return
				}
				m.snap = e.Checked
				update()
			}},
		}},
		m.compactActions(invoke),
		{ID: "geometry-tools", Kind: ui.Row, Children: []ui.Node{
			{ID: "geometry-open", Kind: ui.Button, Text: "Geometry…", Style: compactControlStyle(), Disabled: m.controlDisabled("geometry", m.editAvailable() != nil), Activate: func() {
				if ref.Context().Err() != nil {
					return
				}
				if err := m.beginGeometry(); err != nil {
					m.status = err.Error()
				}
				update()
			}},
			{ID: "aspect-lock", Kind: ui.Checkbox, Text: "Lock aspect ratio", Checked: m.aspectLocked, Disabled: m.controlDisabled("interaction", m.clipboardAvailable() != nil), Change: func(e ui.ChangeEvent) {
				if ref.Context().Err() != nil || m.clipboardAvailable() != nil {
					return
				}
				m.aspectLocked = e.Checked
				update()
			}},
			{ID: "selection-subtract", Kind: ui.Checkbox, Text: "Subtract selection", Checked: m.selectionSubtract, Disabled: m.controlDisabled("interaction", m.clipboardAvailable() != nil), Change: func(e ui.ChangeEvent) {
				if ref.Context().Err() != nil || m.clipboardAvailable() != nil {
					return
				}
				m.selectionSubtract = e.Checked
				update()
			}},
		}},
		{ID: "help", Kind: ui.Label, Text: dimensions + " • " + help, Style: compactControlStyle()},
		{ID: "workspace", Kind: ui.Row, Grow: true, Children: []ui.Node{
			{ID: "palette", Kind: ui.Column, MinWidth: 145, Children: tools},
			{ID: "drawing", Kind: ui.Canvas2D, Disabled: busy, Canvas: display, CanvasObjects: objects, CanvasGrid: grid, PlacementTarget: m.target, Grow: true, Accessibility: ui.Accessibility{Name: "Component canvas", Help: "Subtract selection changes Shift-click and Shift-marquee from add/toggle to remove; Shift-drag starts on empty canvas; click empty space to clear selection; drag a selected component to move the selection; Option+Left/Right/Home/End selects one; drag selection corner handles to resize the set; plain empty drag or arrow keys pan; scroll zooms."}, Change: func(e ui.ChangeEvent) {
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
	root.Children[1].Children = append(root.Children[1].Children, m.gridControls(ref, update)...)
	root.Children = append(root.Children, m.filePrompt(a, ref), m.renameDialog(a, ref), m.recovery.promptNode(), m.geometryDialog(a, ref))
	for i := range root.Children {
		if root.Children[i].Kind == ui.Toolbar {
			for j := range root.Children[i].Children {
				n := &root.Children[i].Children[j]
				n.Style = compactControlStyle().Overlay(n.Style)
			}
		}
	}
	return root
}
