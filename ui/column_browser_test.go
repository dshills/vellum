package ui

import (
	"reflect"
	"testing"
)

func TestColumnBrowserPathsAndStaleSources(t *testing.T) {
	b := &ColumnBrowser{ID: "browser", Items: []BrowserItem{{ID: "folder", Label: "Folder", Children: []BrowserItem{{ID: "leaf", Label: "Drawing"}}}}, Fractions: map[string]float32{}}
	calls := 0
	b.Change = func(e BrowserEvent) { calls++; b.Path = e.Path }
	n, err := b.View()
	if err != nil {
		t.Fatal(err)
	}
	old := n.Children[0].SelectionChange
	old(ListSelectionEvent{TargetID: "folder"})
	if !reflect.DeepEqual(b.Path, []string{"folder"}) {
		t.Fatal(b.Path)
	}
	n, err = b.View()
	if err != nil {
		t.Fatal(err)
	}
	split := n.Children[0]
	if split.Kind != SplitPane {
		t.Fatal(split)
	}
	second := split.Children[1]
	second.SelectionChange(ListSelectionEvent{TargetID: "leaf"})
	if calls != 2 || len(b.Path) != 2 {
		t.Fatal(b.Path)
	}
	b.Items[0].Disabled = true
	old(ListSelectionEvent{TargetID: "folder"})
	if calls != 2 {
		t.Fatal("disabled ancestry acted")
	}
	if _, err = b.View(); err == nil {
		t.Fatal("disabled path accepted")
	}
	b.Items[0].Disabled = false
	b.Path = nil
	b.Items[0].Label = "Renamed"
	old(ListSelectionEvent{TargetID: "folder"})
	if calls != 2 {
		t.Fatal("stale source acted")
	}
	b.Items[0].Children[0].ID = "folder"
	if _, err = b.View(); err == nil {
		t.Fatal("duplicate ID accepted")
	}
}

func TestColumnBrowserRejectsInvalidFractions(t *testing.T) {
	b := &ColumnBrowser{ID: "b", Fractions: map[string]float32{"column": 2}}
	if _, err := b.View(); err == nil {
		t.Fatal("invalid fraction accepted")
	}
	b.Fractions = map[string]float32{}
	for i := 0; i < 6; i++ {
		b.Fractions[string(rune('a'+i))] = .5
	}
	if _, err := b.View(); err == nil {
		t.Fatal("unbounded fraction catalog accepted")
	}
}

func TestColumnBrowserDeepestLeafSelectionWithPreview(t *testing.T) {
	b := &ColumnBrowser{ID: "deep", Path: []string{"a", "b", "c", "d"}, Items: []BrowserItem{{ID: "a", Label: "A", Children: []BrowserItem{{ID: "b", Label: "B", Children: []BrowserItem{{ID: "c", Label: "C", Children: []BrowserItem{{ID: "d", Label: "D"}}}}}}}}, Fractions: map[string]float32{}}
	preview := Node{ID: "preview", Kind: Label, Text: "D preview"}
	b.Preview = &preview
	for i := 0; i < 4; i++ {
		b.Fractions[BrowserColumnID(b.ID, b.Path[:i])] = .5
	}
	n, err := b.View()
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	var walk func(Node)
	walk = func(n Node) {
		if n.Kind == ListBox {
			found++
			if len(n.SelectedIDs) != 1 {
				t.Fatal("missing selected row", n.ID, n.SelectedIDs)
			}
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(n)
	if found != 4 {
		t.Fatal("deepest column omitted", found)
	}
}
