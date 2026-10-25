package main

import (
	"bytes"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/examples/macapps/artwork"
	"github.com/dshills/vellum/examples/macapps/notes"
	"github.com/dshills/vellum/ui"
	"image/png"
)

func nativeNotesToolbar(ref app.WindowRef, m *notes.Model) (*cogent.TitlebarToolbar, error) {
	toolbar := &cogent.TitlebarToolbar{Window: ref, Commands: m.Commands(), CustomizationID: "com.dshills.vellum.notes.toolbar.v1"}
	for _, action := range []struct {
		id   string
		icon ui.Symbol
	}{{"new", ui.SymbolAdd}, {"open", ui.SymbolFolder}, {"save", ui.SymbolSave}, {"undo", ui.SymbolUndo}, {"redo", ui.SymbolRedo}, {"delete", ui.SymbolDelete}, {"pin", ui.SymbolDocument}} {
		icon, err := artwork.ToolbarIcon(action.icon)
		if err != nil {
			return nil, err
		}
		var pngData bytes.Buffer
		if err := png.Encode(&pngData, icon); err != nil {
			return nil, err
		}
		if action.id == "undo" {
			toolbar.Items = append(toolbar.Items, cogent.TitlebarToolbarItem{ID: "space", Kind: ui.ToolbarFlexibleSpace})
		}
		toolbar.Items = append(toolbar.Items, cogent.TitlebarToolbarItem{ID: action.id, Command: "notes." + action.id, Kind: ui.Button, PNG: pngData.Bytes()})
	}
	return toolbar, nil
}
func nativeNotesMenus() []app.MenuGroup {
	return []app.MenuGroup{
		{Label: "File", Commands: []string{"notes.new", "notes.open", "notes.save", "notes.save-as"}},
		{Label: "Edit", Commands: []string{"notes.undo", "notes.redo", "notes.delete"}},
		{Label: "Note", Commands: []string{"notes.pin"}},
	}
}
