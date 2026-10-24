package main

import (
	"bytes"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"image/png"
)

// A native-only experiment, kept separate from portable toolbar descriptions.
// Reuse the gallery's original immutable artwork and ordinary owner model.
func (m *macWidgetsGallery) titlebarPrototype(ref app.WindowRef) (*cogent.TitlebarToolbarPrototype, error) {
	if m.toolbars.artworkError != nil {
		return nil, m.toolbars.artworkError
	}
	result := &cogent.TitlebarToolbarPrototype{Window: ref}
	for i, action := range []struct {
		name   string
		symbol ui.Symbol
	}{{"New", ui.SymbolAdd}, {"Open", ui.SymbolFolder}, {"Save", ui.SymbolSave}, {"Undo", ui.SymbolUndo}, {"Redo", ui.SymbolRedo}, {"Delete", ui.SymbolDelete}} {
		if i == 3 {
			result.Items = append(result.Items, cogent.TitlebarToolbarItem{ID: "history-space", Kind: ui.ToolbarSpace})
		}
		if i == 5 {
			result.Items = append(result.Items, cogent.TitlebarToolbarItem{ID: "trailing-space", Kind: ui.ToolbarFlexibleSpace})
		}
		var image bytes.Buffer
		if err := png.Encode(&image, m.toolbars.artwork[action.symbol]); err != nil {
			return nil, err
		}
		name := action.name
		result.Items = append(result.Items, cogent.TitlebarToolbarItem{ID: name, Label: name, Kind: ui.Button, PNG: image.Bytes(), Disabled: name == "Delete", Activate: func() {
			if name == "Delete" {
				return
			}
			m.toolbars.count++
			m.toolbars.last = "Native titlebar: " + name
		}})
	}
	return result, nil
}
