package ui

import (
	"slices"
	"testing"
)

func TestDocumentListSortSelectionAndRevocation(t *testing.T) {
	d := &DocumentList{ID: "files", Label: "Files", Columns: DefaultDocumentColumns(), Sort: DocumentSort{ColumnID: "name"}, Items: []DocumentItem{{ID: "b", Name: "Beta", Kind: "Drawing", SizeBytes: 20}, {ID: "a", Name: "Alpha", Kind: "Drawing", SizeBytes: 10}, {ID: "f", Name: "Folder", Folder: true}}}
	d.SortChange = func(e DocumentSortEvent) { d.Sort = e.Sort }
	d.Resize = func(e DocumentResizeEvent) {
		for i := range d.Columns {
			if d.Columns[i].ID == e.ColumnID {
				d.Columns[i].Width = e.Width
			}
		}
	}
	d.Change = func(e ListSelectionEvent) { d.SelectedIDs = e.Apply(d.SelectedIDs) }
	calls := 0
	d.Activate = func(DocumentEvent) { calls++ }
	n, err := d.View()
	if err != nil {
		t.Fatal(err)
	}
	list := pickerNode(n, DocumentListID(d.ID))
	if got := []string{list.Options[0].ID, list.Options[1].ID, list.Options[2].ID}; !slices.Equal(got, []string{"f", "a", "b"}) {
		t.Fatal(got)
	}
	if d.Items[0].ID != "b" {
		t.Fatal("sorted caller slice")
	}
	list.SelectionChange(ListSelectionEvent{TargetID: "a", IDs: []string{"a"}})
	list.SelectionChange(ListSelectionEvent{TargetID: "b", IDs: []string{"b"}, Toggle: true})
	if !slices.Equal(d.SelectedIDs, []string{"a", "b"}) {
		t.Fatal(d.SelectedIDs)
	}
	pickerNode(n, DocumentHeaderID(d.ID, "name")).Activate()
	if !d.Sort.Descending {
		t.Fatal("sort not toggled")
	}
	handle := pickerNode(n, DocumentHeaderID(d.ID, "name")+"/resize").ColumnResize
	handle.Change(999)
	if d.Columns[0].Width != 600 {
		t.Fatal("resize not clamped")
	}
	d.Items = d.Items[:1]
	d.SelectedIDs = nil
	list.DocumentList.Activate(DocumentEvent{ItemID: "a"})
	if calls != 0 {
		t.Fatal("removed item activated")
	}
	d.Disabled = true
	handle.Change(120)
	if d.Columns[0].Width != 600 {
		t.Fatal("disabled resize acted")
	}
}
func TestDocumentListRejectsMalformedMetadata(t *testing.T) {
	d := &DocumentList{ID: "d", Label: "Documents", Items: []DocumentItem{{ID: "a", Name: "A"}}}
	for _, cols := range [][]DocumentColumn{{{ID: "name", Title: "Name", Width: 0}}, {{ID: "oops", Title: "Oops", Width: 100}}, {{ID: "name", Title: "Name", Width: 100}, {ID: "name", Title: "Again", Width: 100}}} {
		d.Columns = cols
		if _, err := d.View(); err == nil {
			t.Fatal("invalid columns admitted")
		}
	}
	d.Columns = nil
	d.SelectedIDs = []string{"missing"}
	if _, err := d.View(); err == nil {
		t.Fatal("absent selection")
	}
}

func TestDocumentListDetachedMetadataValidation(t *testing.T) {
	d := &DocumentList{ID: "files", Label: "Files", Items: []DocumentItem{{ID: "a", Name: "A"}}}
	n, _ := d.View()
	list := pickerNode(n, DocumentListID(d.ID))
	list.DocumentList.Columns[0].Width = -1
	if ValidateControl(list) == nil {
		t.Fatal("mutated width admitted")
	}
	n, _ = d.View()
	list = pickerNode(n, DocumentListID(d.ID))
	list.DocumentList.Rows[0].Cells[0] = "bad\x00text"
	if ValidateControl(list) == nil {
		t.Fatal("invalid cell admitted")
	}
}
