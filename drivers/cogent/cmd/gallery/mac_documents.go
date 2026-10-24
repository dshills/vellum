package main

import (
	"fmt"
	"github.com/dshills/vellum/ui"
	"time"
)

type macDocumentGallery struct {
	parent  *macComponentsGallery
	list    *ui.DocumentList
	outline *ui.OutlineView
}

func (m *macComponentsGallery) documentListView() ui.Node {
	if m.documents == nil {
		m.documents = &macDocumentGallery{parent: m}
	}
	e := m.documents
	if e.list == nil {
		e.list = &ui.DocumentList{ID: "documents/files", Label: "Drawing documents", Columns: ui.DefaultDocumentColumns(), Sort: ui.DocumentSort{ColumnID: "name"}, Items: []ui.DocumentItem{
			{ID: "projects", Name: "Projects", Kind: "Folder", Folder: true, Image: m.artwork[ui.SymbolFolder]},
			{ID: "floor", Name: "Floor plan.vellum", Kind: "Drawing", SizeBytes: 18432, Modified: time.Date(2026, 10, 24, 9, 0, 0, 0, time.Local), Image: m.artwork[ui.SymbolDocument]},
			{ID: "details", Name: "Details.vellum", Kind: "Drawing", SizeBytes: 8192, Modified: time.Date(2026, 10, 23, 12, 0, 0, 0, time.Local), Image: m.artwork[ui.SymbolSave]},
			{ID: "archive", Name: "Archive (unavailable)", Kind: "Folder", Folder: true, Disabled: true, Image: m.artwork[ui.SymbolFolder]},
		}}
		e.list.Change = func(event ui.ListSelectionEvent) {
			e.list.SelectedIDs = event.Apply(e.list.SelectedIDs)
			m.status = fmt.Sprint("Document selection: ", e.list.SelectedIDs)
		}
		e.list.SortChange = func(event ui.DocumentSortEvent) { e.list.Sort = event.Sort }
		e.list.Resize = func(event ui.DocumentResizeEvent) {
			for i := range e.list.Columns {
				if e.list.Columns[i].ID == event.ColumnID {
					e.list.Columns[i].Width = event.Width
				}
			}
		}
		e.list.Activate = func(event ui.DocumentEvent) { m.count++; m.status = "Open document: " + event.ItemID }
	}
	n, err := e.list.View()
	if err != nil {
		return ui.Node{ID: "documents/list-error", Kind: ui.Label, Text: err.Error()}
	}
	return ui.Node{ID: "documents/list-page", Kind: ui.Column, Style: ui.Style{GrowX: ui.Some(float32(1))}, Children: []ui.Node{{ID: "documents/list-help", Kind: ui.Label, Text: "Finder-style document metadata. Click a header to sort. Drag its divider or focus the divider and use Left/Right. Command/Shift select; Return opens the sample only."}, n}}
}

func (m *macComponentsGallery) outlineView() ui.Node {
	if m.documents == nil {
		m.documents = &macDocumentGallery{parent: m}
	}
	e := m.documents
	if e.outline == nil {
		e.outline = &ui.OutlineView{ID: "documents/outline", Label: "Project outline", ExpandedIDs: []string{"project", "drawings"}, Items: []ui.OutlineItem{
			{ID: "project", Label: "My project", Badge: "4", Image: m.artwork[ui.SymbolFolder]},
			{ID: "drawings", Parent: "project", Label: "Drawings", Image: m.artwork[ui.SymbolFolder]},
			{ID: "floor", Parent: "drawings", Label: "Floor plan", Image: m.artwork[ui.SymbolDocument]},
			{ID: "detail", Parent: "drawings", Label: "Details", Image: m.artwork[ui.SymbolSave]},
			{ID: "components", Parent: "project", Label: "Components", Image: m.artwork[ui.SymbolFolder]},
			{ID: "block", Parent: "components", Label: "Block", Image: m.artwork[ui.SymbolDocument]},
			{ID: "archive", Parent: "project", Label: "Archive (unavailable)", Disabled: true, Image: m.artwork[ui.SymbolFolder]},
		}}
		e.outline.Change = func(event ui.ListSelectionEvent) {
			e.outline.SelectedIDs = event.Apply(e.outline.SelectedIDs)
			m.status = fmt.Sprint("Outline selection: ", e.outline.SelectedIDs)
		}
		e.outline.Disclosure = func(event ui.OutlineDisclosureEvent) { e.outline.ExpandedIDs = event.Apply(e.outline.ExpandedIDs) }
		e.outline.Activate = func(event ui.OutlineEvent) { m.count++; m.status = "Open outline item: " + event.ItemID }
	}
	n, err := e.outline.View()
	if err != nil {
		return ui.Node{ID: "documents/outline-error", Kind: ui.Label, Text: err.Error()}
	}
	return ui.Node{ID: "documents/outline-page", Kind: ui.Column, Style: ui.Style{GrowX: ui.Some(float32(1))}, Children: []ui.Node{{ID: "documents/outline-help", Kind: ui.Label, Text: "Expand folders with their triangles or Right. Left collapses or moves to the parent. Command/Shift selects; Return opens a leaf. Hidden selections survive collapse."}, n}}
}
