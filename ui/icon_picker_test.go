package ui

import "testing"

func pickerNode(n Node, id string) Node {
	if n.ID == id {
		return n
	}
	for _, c := range n.Children {
		if got := pickerNode(c, id); got.ID != "" {
			return got
		}
	}
	return Node{}
}
func TestIconPickerSearchAcceptanceAndSessionRevocation(t *testing.T) {
	p := &IconPicker{ID: "picker", Label: "Choose artwork", Items: []IconPickerItem{{ID: "a", Label: "Folder", Category: "Documents", Icon: SymbolFolder}, {ID: "b", Label: "Save", Category: "Actions", Icon: SymbolSave}}}
	calls := 0
	reject := true
	p.Change = func(e IconPickerEvent) {
		calls++
		if !reject {
			p.SelectedID = e.ItemID
		}
	}
	n, _ := p.View()
	pickerNode(n, IconPickerButtonID(p.ID)).Activate()
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	old := pickerNode(n, IconPickerCollectionID(p.ID))
	old.Collection.Activate(CollectionEvent{ItemID: "a"})
	if calls != 1 || !p.Open || len(p.recent) != 0 {
		t.Fatal("rejected choice committed")
	}
	pickerNode(n, SearchInputID("picker/icon-picker/search")).Change(ChangeEvent{Text: "save"})
	old.Collection.Activate(CollectionEvent{ItemID: "a"})
	if calls != 1 {
		t.Fatal("stale query selection admitted")
	}
	n, _ = p.View()
	grid := pickerNode(n, IconPickerCollectionID(p.ID))
	if len(grid.Options) != 1 || grid.Options[0].ID != "b" {
		t.Fatal("filter failed")
	}
	reject = false
	grid.Collection.Activate(CollectionEvent{ItemID: "b"})
	if calls != 2 || p.Open || p.SelectedID != "b" || len(p.recent) != 1 {
		t.Fatal("accepted choice not committed")
	}
	grid.Collection.Activate(CollectionEvent{ItemID: "b"})
	if calls != 2 {
		t.Fatal("closed session acted")
	}
}

func TestIconPickerExternalReopenAndRecentPruning(t *testing.T) {
	p := &IconPicker{ID: "p", Label: "Picker", Items: []IconPickerItem{{ID: "a", Label: "A"}}, Open: true, Change: func(IconPickerEvent) {}}
	p.View()
	p.query = "old"
	p.Open = false
	p.View()
	p.Open = true
	p.View()
	if p.query != "" {
		t.Fatal("external reopen kept query")
	}
	p.recent = []string{"gone"}
	p.View()
	if len(p.recent) != 0 {
		t.Fatal("removed recent retained")
	}
}
