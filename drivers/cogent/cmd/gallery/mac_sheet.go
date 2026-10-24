package main

import (
	"fmt"
	"github.com/dshills/vellum/ui"
)

type macSheetGallery struct {
	sheet *ui.Sheet
	count int
}

func newMacSheetGallery(dark bool) *macSheetGallery {
	m := &macSheetGallery{}
	m.sheet = &ui.Sheet{ID: "mac-sheet", Title: "Export drawing", ConfirmLabel: "Export", ReturnFocusID: "show-mac-sheet", Children: []ui.Node{{ID: "sheet-explanation", Kind: ui.Label, Text: "Export a copy of this drawing? This demonstration writes no file."}}}
	m.sheet.Style = ui.MacTheme(dark).Kinds[ui.FormGroup]
	m.sheet.Style.Padding = ui.Some(ui.UniformInsets(20))
	m.sheet.Style.MaxSize = ui.Some(ui.Size{Width: 440})
	m.sheet.Cancel = func() { m.sheet.Open = false }
	m.sheet.Confirm = func() { m.count++; m.sheet.Open = false }
	return m
}
func (m *macSheetGallery) view() ui.Node {
	panel, err := m.sheet.View()
	if err != nil {
		return ui.Node{ID: "sheets", Kind: ui.Label, Text: err.Error()}
	}
	return ui.Node{ID: "sheets", Kind: ui.Column, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(20)), Gap: ui.Some(float32(14))}, Children: []ui.Node{{ID: "sheet-heading", Kind: ui.Label, Text: "Window-modal sheet", Style: ui.Style{Typography: ui.Some(ui.Typography{Size: 20, Weight: 600})}}, {ID: "sheet-help", Kind: ui.Label, Text: "Open the sheet. Escape cancels; Return or Export confirms. Closing restores focus to the opener."}, {ID: "show-mac-sheet", Kind: ui.Button, Text: "Export drawing…", Activate: func() { m.sheet.Open = true }}, {ID: "sheet-count", Kind: ui.Label, Text: fmt.Sprintf("Completed exports: %d", m.count)}, panel}}
}
