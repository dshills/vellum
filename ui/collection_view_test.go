package ui

import (
	"reflect"
	"testing"
)

func collectionFixture() *CollectionView {
	return &CollectionView{ID: "files", Label: "Documents", Items: []CollectionItem{{ID: "a", Label: "Alpha", Icon: SymbolFolder}, {ID: "b", Label: "Beta"}, {ID: "c", Label: "Charlie", Disabled: true}, {ID: "d", Label: "Delta"}, {ID: "e", Label: "Echo"}}, SelectedIDs: []string{"a"}}
}
func TestCollectionControlledSnapshotsAndCallbacks(t *testing.T) {
	f := collectionFixture()
	var events []ListSelectionEvent
	var opened []CollectionEvent
	f.Change = func(e ListSelectionEvent) { events = append(events, e); f.SelectedIDs = e.Apply(f.SelectedIDs) }
	f.Activate = func(e CollectionEvent) { opened = append(opened, e) }
	n, err := f.View()
	if err != nil {
		t.Fatal(err)
	}
	f.Items[0].Label = "Edited"
	f.SelectedIDs[0] = "b"
	if n.Options[0].Label != "Alpha" || n.Collection.Items[0].Label != "Alpha" || n.SelectedIDs[0] != "a" {
		t.Fatal("view aliases model slices")
	}
	n.SelectionChange(ListSelectionEvent{TargetID: "a", IDs: []string{"a"}, Toggle: true})
	n.SelectionChange(ListSelectionEvent{TargetID: "d", IDs: []string{"d"}, Toggle: true})
	if !reflect.DeepEqual(f.SelectedIDs, []string{"b", "a", "d"}) || len(events) != 2 {
		t.Fatal(f.SelectedIDs, events)
	}
	n.Collection.Activate(CollectionEvent{ItemID: "d"})
	if len(opened) != 1 || opened[0] != (CollectionEvent{ID: "files", ItemID: "d"}) {
		t.Fatal(opened)
	}
	f.Disabled = true
	n.SelectionChange(ListSelectionEvent{TargetID: "a", IDs: []string{"a"}})
	n.Collection.Activate(CollectionEvent{ItemID: "d"})
	if len(events) != 2 || len(opened) != 1 {
		t.Fatal("disabled snapshot acted")
	}
	f.Disabled = false
	f.Items = f.Items[:2]
	f.SelectedIDs = nil
	n.Collection.Activate(CollectionEvent{ItemID: "d"})
	n.SelectionChange(ListSelectionEvent{TargetID: "d", IDs: []string{"d"}})
	if len(events) != 2 || len(opened) != 1 {
		t.Fatal("removed item acted")
	}
}
func TestCollectionValidation(t *testing.T) {
	for name, edit := range map[string]func(*CollectionView){"duplicate": func(f *CollectionView) { f.Items[1].ID = "a" }, "selection": func(f *CollectionView) { f.SelectedIDs = []string{"absent"} }, "width": func(f *CollectionView) { f.ItemWidth = 79 }, "line": func(f *CollectionView) { f.Items[0].Label = "a\nb" }, "icon": func(f *CollectionView) { f.Items[0].Icon = "unknown" }} {
		t.Run(name, func(t *testing.T) {
			f := collectionFixture()
			edit(f)
			if _, err := f.View(); err == nil {
				t.Fatal("accepted invalid collection")
			}
		})
	}
	f := collectionFixture()
	n, err := f.View()
	if err != nil {
		t.Fatal(err)
	}
	n.Options[0].Label = "mismatch"
	if ValidateCollectionNode(n) == nil {
		t.Fatal("accepted mismatched native snapshot")
	}
}
func TestCollectionGridNavigation(t *testing.T) {
	items := collectionFixture().Items
	for _, tc := range []struct {
		from         string
		cols, dx, dy int
		want         string
	}{{"a", 2, 1, 0, "b"}, {"a", 2, 0, 1, "e"}, {"d", 2, 0, -1, "b"}, {"d", 3, 0, -1, "a"}, {"e", 2, 0, 1, "e"}, {"a", 2, -1, 0, "a"}, {"missing", 2, 1, 0, "a"}, {"b", 3, 0, 1, "e"}, {"d", 3, 0, 1, "d"}, {"e", 3, 0, 1, "e"}} {
		got, ok := MoveCollectionItem(items, tc.from, tc.cols, tc.dx, tc.dy)
		if !ok || got != tc.want {
			t.Fatalf("%+v got %s %v", tc, got, ok)
		}
	}
	if _, ok := MoveCollectionItem(items, "a", 0, 0, 1); ok {
		t.Fatal("accepted invalid columns")
	}
	if _, ok := MoveCollectionItem(items, "a", 2, 1, 1); ok {
		t.Fatal("accepted diagonal")
	}
}

func TestCollectionVerticalDisabledColumnDoesNotDrift(t *testing.T) {
	items := []CollectionItem{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}, {ID: "e", Disabled: true}, {ID: "f"}, {ID: "g"}, {ID: "h"}}
	if got, ok := MoveCollectionItem(items, "b", 3, 0, 1); !ok || got != "h" {
		t.Fatal("Down drifted from second column", got, ok)
	}
	if got, ok := MoveCollectionItem(items, "h", 3, 0, -1); !ok || got != "b" {
		t.Fatal("Up drifted from second column", got, ok)
	}
	items[7].Disabled = true
	if got, ok := MoveCollectionItem(items, "b", 3, 0, 1); !ok || got != "b" {
		t.Fatal("all disabled column should retain focus", got, ok)
	}
	if got, ok := MoveCollectionItem(items, "f", 3, 0, 1); !ok || got != "f" {
		t.Fatal("incomplete final row changed column", got, ok)
	}
}

func TestCollectionSnapshotSourceIdentity(t *testing.T) {
	f := collectionFixture()
	a, err := f.View()
	if err != nil {
		t.Fatal(err)
	}
	f.Items[0].Label = "Renamed"
	b, err := f.View()
	if err != nil {
		t.Fatal(err)
	}
	if !a.Collection.SameSource(b.Collection) {
		t.Fatal("same retained owner lost identity on rename")
	}
	replacement := collectionFixture()
	c, err := replacement.View()
	if err != nil {
		t.Fatal(err)
	}
	if a.Collection.SameSource(c.Collection) || a.Collection.SameSource(nil) {
		t.Fatal("replacement inherited source identity")
	}
	var absent *CollectionLayout
	if !absent.SameSource(nil) {
		t.Fatal("ordinary list comparison")
	}
}
