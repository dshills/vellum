package sdkcontract_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/dshills/vellum/ui"
)

// This independently compiled module imports no driver or ui internals. Native
// interaction and pixels require separate driver/gallery checks.
func TestPublicNextMacWidgetCompositionAndControlledProposals(t *testing.T) {
	requests := map[string]int{}
	reject := false
	sidebar := &ui.Sidebar{ID: "sdk-sidebar", Label: "Library", SelectedID: "all", Sections: []ui.SidebarSection{{ID: "favorites", Label: "Favorites", Items: []ui.Option{{ID: "all", Label: "All drawings", Icon: ui.SymbolFolder}, {ID: "recent", Label: "Recent drawings", Icon: ui.SymbolDocument}}}}}
	sidebar.Navigate = func(e ui.SidebarEvent) {
		if e.ID != sidebar.ID {
			t.Errorf("sidebar event identity %q", e.ID)
		}
		requests["sidebar"]++
		if !reject {
			sidebar.SelectedID = e.ItemID
		}
	}
	sidebar.Disclosure = func(e ui.SidebarDisclosureEvent) {
		if !reject {
			sidebar.Sections[0].Collapsed = !e.Expanded
		}
	}
	collection := &ui.CollectionView{ID: "sdk-collection", Label: "Drawings", Items: []ui.CollectionItem{{ID: "plan", Label: "Plan", Icon: ui.SymbolDocument}, {ID: "elevation", Label: "Elevation", Icon: ui.SymbolDocument}}, SelectedIDs: []string{"plan"}}
	collection.Change = func(e ui.ListSelectionEvent) {
		if e.ID != collection.ID {
			t.Errorf("collection event identity %q", e.ID)
		}
		requests["collection"]++
		if !reject {
			collection.SelectedIDs = e.Apply(collection.SelectedIDs)
		}
	}
	collection.Activate = func(e ui.CollectionEvent) {
		if e.ID != collection.ID || e.ItemID != "elevation" {
			t.Errorf("activation identity %+v", e)
		}
		requests["activate"]++
	}
	recorder := &ui.ShortcutRecorder{ID: "sdk-shortcut", Label: "Export shortcut", Value: ui.Shortcut{Key: "E", Modifiers: ui.ShortcutCommand}}
	recorder.Change = func(e ui.ShortcutEvent) {
		if e.ID != recorder.ID {
			t.Errorf("shortcut event identity %q", e.ID)
		}
		requests["shortcut"]++
		recorder.Value = e.Value
	}
	recorder.Validate = func(ui.Shortcut) error {
		if reject {
			return errors.New("reserved by application")
		}
		return nil
	}
	recorder.Invalid = func(error) { requests["invalid-shortcut"]++ }
	level := &ui.LevelIndicator{ID: "sdk-rating", Label: "Drawing rating", Mode: ui.LevelRating, Max: 5, Value: 2, AllowClear: true}
	level.Change = func(e ui.LevelEvent) {
		if e.ID != level.ID {
			t.Errorf("level event identity %q", e.ID)
		}
		requests["level"]++
		if !reject {
			level.Value, level.Unrated = e.Value, e.Unrated
		}
	}
	capacity := &ui.LevelIndicator{ID: "sdk-capacity", Label: "Storage", Mode: ui.LevelContinuous, Min: 0, Max: 100, Value: 75, Warning: ui.Some(float32(75)), Critical: ui.Some(float32(90))}
	sheet := &ui.Sheet{ID: "sdk-sheet", Title: "Export drawing", ReturnFocusID: "sdk-open-sheet", Children: []ui.Node{{ID: "sdk-export-name", Kind: ui.Label, Text: "Export the accepted drawing"}}}
	sheet.Confirm = func() {
		requests["sheet"]++
		if !reject {
			sheet.Open = false
		}
	}
	sheet.Cancel = func() { requests["cancel-sheet"]++; sheet.Open = false }
	view := func() ui.Node {
		t.Helper()
		root := ui.Node{ID: "sdk-next-widgets", Kind: ui.Column, Children: []ui.Node{{ID: "sdk-open-sheet", Kind: ui.Button, Text: "Export", Activate: func() { sheet.Open = true }}}}
		for _, render := range []func() (ui.Node, error){sidebar.View, collection.View, recorder.View, level.View, capacity.View, sheet.View} {
			n, err := render()
			if err != nil {
				t.Fatal(err)
			}
			root.Children = append(root.Children, n)
		}
		if err := ui.Validate(root); err != nil {
			t.Fatal("public widget composition", err)
		}
		return root
	}
	root := view()
	list := publicMacWidgetNode(t, root, ui.SidebarListID(sidebar.ID, "favorites"))
	list.SelectionChange(ui.ListSelectionEvent{TargetID: "recent", IDs: []string{"recent"}})
	grid := publicMacWidgetNode(t, root, collection.ID)
	grid.SelectionChange(ui.ListSelectionEvent{TargetID: "elevation", IDs: []string{"elevation"}})
	if grid.Collection == nil || grid.Collection.Activate == nil {
		t.Fatal("public collection activation metadata missing")
	}
	grid.Collection.Activate(ui.CollectionEvent{ItemID: "elevation"})
	publicMacWidgetNode(t, root, recorder.ID).Activate()
	if !recorder.Recording() {
		t.Fatal("public recorder did not begin")
	}
	acceptedShortcut := ui.Shortcut{Key: "E", Modifiers: ui.ShortcutCommand | ui.ShortcutShift}
	if err := recorder.Record(recorder.Session(), acceptedShortcut); err != nil {
		t.Fatal(err)
	}
	publicMacWidgetNode(t, root, ui.LevelStarID(level.ID, 4)).Activate()
	publicMacWidgetNode(t, root, "sdk-open-sheet").Activate()
	root = view()
	publicMacWidgetNode(t, root, ui.SheetConfirmID(sheet.ID)).Activate()
	if sidebar.SelectedID != "recent" || !reflect.DeepEqual(collection.SelectedIDs, []string{"elevation"}) || recorder.Value != acceptedShortcut || level.Value != 4 || sheet.Open {
		t.Fatal("public proposal acceptance failed")
	}
	for _, name := range []string{"sidebar", "collection", "activate", "shortcut", "level", "sheet"} {
		if requests[name] != 1 {
			t.Fatal("accepted proposal count", name, requests[name])
		}
	}

	// Keeping accepted state rejects an application proposal. Shortcut conflicts
	// use the recorder's exported validation path instead of changing its value.
	reject = true
	sheet.Open = true
	root = view()
	list = publicMacWidgetNode(t, root, ui.SidebarListID(sidebar.ID, "favorites"))
	list.SelectionChange(ui.ListSelectionEvent{TargetID: "all", IDs: []string{"all"}})
	grid = publicMacWidgetNode(t, root, collection.ID)
	grid.SelectionChange(ui.ListSelectionEvent{TargetID: "plan", IDs: []string{"plan"}})
	publicMacWidgetNode(t, root, ui.LevelStarID(level.ID, 1)).Activate()
	publicMacWidgetNode(t, root, ui.SheetConfirmID(sheet.ID)).Activate()
	publicMacWidgetNode(t, root, recorder.ID).Activate()
	if err := recorder.Record(recorder.Session(), ui.Shortcut{Key: "Q", Modifiers: ui.ShortcutCommand}); err == nil {
		t.Fatal("application shortcut conflict accepted")
	}
	if sidebar.SelectedID != "recent" || !reflect.DeepEqual(collection.SelectedIDs, []string{"elevation"}) || level.Value != 4 || recorder.Value != acceptedShortcut || !sheet.Open {
		t.Fatal("rejected proposal mutated accepted state")
	}
	if requests["shortcut"] != 1 || requests["invalid-shortcut"] != 1 {
		t.Fatal("shortcut validation bypassed callback boundary", requests)
	}

	// Previously returned descriptions must honor current disabled owner state.
	before := make(map[string]int, len(requests))
	for name, n := range requests {
		before[name] = n
	}
	sidebar.Disabled, collection.Disabled, recorder.Disabled, level.Disabled, sheet.Disabled = true, true, true, true, true
	list.SelectionChange(ui.ListSelectionEvent{TargetID: "all", IDs: []string{"all"}})
	grid.SelectionChange(ui.ListSelectionEvent{TargetID: "plan", IDs: []string{"plan"}})
	grid.Collection.Activate(ui.CollectionEvent{ItemID: "elevation"})
	publicMacWidgetNode(t, root, ui.LevelStarID(level.ID, 1)).Activate()
	publicMacWidgetNode(t, root, ui.SheetConfirmID(sheet.ID)).Activate()
	publicMacWidgetNode(t, root, recorder.ID).Activate()
	if recorder.Begin() {
		t.Fatal("disabled recorder began a session")
	}
	if err := recorder.Record(recorder.Session(), acceptedShortcut); err != nil {
		t.Fatal("disabled recorder returned error", err)
	}
	if !reflect.DeepEqual(requests, before) {
		t.Fatal("disabled description delivered a callback", before, requests)
	}
	view()
}
