package main

import (
	"context"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
	"os"
	"path/filepath"
)

type macEssentialsGallery struct {
	parent      *macComponentsGallery
	application *app.Application
	window      app.WindowRef
	well        *ui.ImageWell
	picker      *ui.IconPicker
}

func newMacEssentialsGallery(parent *macComponentsGallery) *macEssentialsGallery {
	e := &macEssentialsGallery{parent: parent}
	e.well = &ui.ImageWell{ID: "essentials/well", Label: "Drawing artwork", ChooseIcon: parent.artwork[ui.SymbolFolder], PlaceholderImage: parent.artwork[ui.SymbolDocument], Request: e.imageRequest}
	return e
}
func (e *macEssentialsGallery) imageRequest(event ui.ImageWellEvent) bool {
	if e.well.Busy {
		return false
	}
	if event.Action == ui.ImageWellClear {
		e.well.Image = nil
		e.well.Status = "Image cleared"
		return true
	}
	if e.application == nil || e.window.Context().Err() != nil {
		e.well.Status = "Native file services are unavailable in this offscreen preview. Use the sample."
		return false
	}
	ticket, err := e.application.ReserveWindowCompletion(e.window, nil)
	if err != nil {
		e.well.Status = err.Error()
		return false
	}
	e.well.Busy = true
	go func() {
		path := event.Path
		var result *canvas.TiledImage
		var failure error
		if event.Action == ui.ImageWellChoose {
			service, err := e.application.Services()
			failure = err
			if err == nil {
				path, failure = service.ChooseFile(e.window.Context(), app.FileOptions{Title: "Choose drawing artwork", Extensions: []string{"png", "jpg", "jpeg"}})
			}
		}
		if failure == nil {
			result, failure = loadWellImage(path)
		}
		// A refused publication means this captured window lifetime is retired; never mutate UI from this worker.
		_ = ticket.Publish(func() {
			defer func() { _ = e.application.Invalidate(e.window.ID()) }()
			e.well.Busy = false
			if failure != nil {
				if failure == app.ErrCanceled || failure == context.Canceled {
					e.well.Status = "Image selection canceled"
				} else {
					e.well.Status = fmt.Sprintf("Image rejected: %.512s", failure)
				}
				return
			}
			e.well.Image = result
			e.well.Status = "Loaded " + filepath.Base(path)
			e.parent.count++
			_ = e.application.Invalidate(e.window.ID())
		})
	}()
	return true
}
func (e *macEssentialsGallery) imageWellView() ui.Node {
	n, err := e.well.View()
	if err != nil {
		return ui.Node{ID: "essentials/error", Kind: ui.Label, Text: err.Error()}
	}
	return ui.Node{ID: "essentials/image-page", Kind: ui.Column, Style: ui.Style{GrowX: ui.Some(float32(1))}, Children: []ui.Node{n, {ID: "essentials/sample-image", Kind: ui.Button, Text: "Use full-color sample", IconImage: e.parent.artwork[ui.SymbolAdd], Activate: func() {
		if !e.well.Busy {
			e.well.Image = e.parent.artwork[ui.SymbolFolder]
			e.well.Status = "Full-color sample accepted"
			e.parent.count++
		}
	}}, {ID: "essentials/image-note", Kind: ui.Label, Text: fmt.Sprint("Images decode on a bounded worker; files are never changed.")}}}
}

func loadWellImage(path string) (*canvas.TiledImage, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > ui.MaxEncodedImageBytes {
		return nil, fmt.Errorf("image must be a regular PNG/JPEG of at most 32MiB")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err = file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("image file changed or is not regular")
	}
	return ui.DecodeImage(file)
}

func (e *macEssentialsGallery) iconPickerView() ui.Node {
	if e.picker == nil {
		e.picker = &ui.IconPicker{ID: "essentials/picker", Label: "Choose component artwork", SelectedID: "folder", Change: func(event ui.IconPickerEvent) {
			e.picker.SelectedID = event.ItemID
			e.parent.count++
			e.parent.status = "Icon selected: " + event.ItemID
		}}
		for _, item := range []struct {
			id, label, category string
			symbol              ui.Symbol
		}{{"folder", "Folder", "Documents", ui.SymbolFolder}, {"drawing", "Drawing", "Documents", ui.SymbolDocument}, {"new", "New document", "Documents", ui.SymbolAdd}, {"save", "Save", "Actions", ui.SymbolSave}, {"undo", "Undo", "Actions", ui.SymbolUndo}, {"redo", "Redo", "Actions", ui.SymbolRedo}, {"delete", "Delete", "Actions", ui.SymbolDelete}} {
			e.picker.Items = append(e.picker.Items, ui.IconPickerItem{ID: item.id, Label: item.label, Category: item.category, Image: e.parent.artwork[item.symbol]})
		}
	}
	n, err := e.picker.View()
	if err != nil {
		return ui.Node{ID: "essentials/picker-error", Kind: ui.Label, Text: err.Error()}
	}
	return ui.Node{ID: "essentials/picker-page", Kind: ui.Column, Style: ui.Style{GrowX: ui.Some(float32(1))}, Children: []ui.Node{{ID: "essentials/picker-help", Kind: ui.Label, Text: "Search or filter full-color artwork. Arrows highlight; Return or Choose accepts. Cancel keeps the original selection."}, n}}
}
