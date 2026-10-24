//go:build offscreen

package cogent

import (
	"fmt"
	"slices"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
)

func nativeCollectionFixture() *ui.CollectionView {
	f := &ui.CollectionView{ID: "files", Label: "Files", SelectedIDs: []string{"0"}}
	for i := range 12 {
		f.Items = append(f.Items, ui.CollectionItem{ID: fmt.Sprint(i), Label: fmt.Sprintf("Document %d", i), Icon: ui.SymbolDocument, Disabled: i == 5})
	}
	return f
}
func TestMacCollectionResponsiveColumnsAndRoutedKeyboard(t *testing.T) {
	f := nativeCollectionFixture()
	calls, opens := 0, 0
	reject := false
	width := float32(380)
	f.Change = func(e ui.ListSelectionEvent) {
		calls++
		if !reject {
			f.SelectedIDs = e.Apply(f.SelectedIDs)
		}
	}
	f.Activate = func(e ui.CollectionEvent) {
		opens++
		if e.ItemID != "2" {
			t.Errorf("activation ID %s", e.ItemID)
		}
	}
	m, a := simpleManual(t, manualConfig(t, 63), func() ui.Node {
		n, err := f.View()
		if err != nil {
			t.Fatal(err)
		}
		n.Style = ui.Style{MaxSize: ui.Some(ui.Size{Width: width, Height: 210})}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: "files", Sequence: 1}, Children: []ui.Node{n}}
	})
	unitSettle(t, m)
	w := m.implementation.find("one")
	r := findRecord(w.state.root, "files")
	if r.widget.AsWidget().Styles.Max.Y.Value != 210 {
		t.Fatal("explicit maximum height lost", r.widget.AsWidget().Styles.Max.Y)
	}
	if got := collectionColumns(r); got != 2 {
		t.Fatalf("wide layout columns %d", got)
	}
	send := func(code key.Codes, mods key.Modifiers) {
		t.Helper()
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code, Modifiers: mods}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	send(key.CodeDownArrow, 0)
	if !slices.Equal(f.SelectedIDs, []string{"2"}) {
		t.Fatal("down did not use measured row stride", f.SelectedIDs)
	}
	send(key.CodeReturnEnter, 0)
	if opens != 1 {
		t.Fatal("Return activation", opens)
	}
	send(key.CodeRightArrow, 1<<key.Shift|1<<key.CapsLock|1<<key.NumLock)
	if !slices.Equal(f.SelectedIDs, []string{"2", "3"}) {
		t.Fatal("Shift range", f.SelectedIDs)
	}
	reject = true
	send(key.CodeLeftArrow, 0)
	if !slices.Equal(f.SelectedIDs, []string{"2", "3"}) || calls != 3 {
		t.Fatal("rejected selection changed model", f.SelectedIDs, calls)
	}
	width = 250
	a.Invalidate("one")
	unitSettle(t, m)
	if got := collectionColumns(r); got != 1 {
		t.Fatalf("narrow layout columns %d", got)
	}
	if len(r.widget.AsTree().Children) != 12 {
		t.Fatal("resize changed identities")
	}
}
func TestMacCollectionActivationRetirementAndThumbnailPlan(t *testing.T) {
	f := nativeCollectionFixture()
	f.Items[1].Thumbnail = managedImageSource(t)
	f.Change = func(ui.ListSelectionEvent) {}
	opens := 0
	f.Activate = func(ui.CollectionEvent) { opens++ }
	m, a := simpleManual(t, manualConfig(t, 67), func() ui.Node {
		n, err := f.View()
		if err != nil {
			t.Fatal(err)
		}
		return n
	})
	unitSettle(t, m)
	r := m.implementation.find("one").state.root
	b := r.control.choice.buttons["1"]
	if b.ChildByName("collection-thumbnail", 0) == nil || b.ChildByName("collection-label", 0) == nil || b.ChildByName("stack", 0) != nil {
		t.Fatal("collection thumbnail plan retained toggle chrome")
	}
	oldThumbnail := b.ChildByName("collection-thumbnail", 0).(*managedImage)
	replacement := managedImageSource(t)
	f.Items[1].Thumbnail = replacement
	a.Invalidate("one")
	unitSettle(t, m)
	if b.ChildByName("collection-thumbnail", 0) != oldThumbnail || oldThumbnail.source != replacement {
		t.Fatal("thumbnail-only replacement did not refresh retained image source")
	}
	f.Items[1].Thumbnail = nil
	a.Invalidate("one")
	unitSettle(t, m)
	if b.ChildByName("collection-icon", 0) == nil || b.ChildByName("collection-thumbnail", 0) != nil {
		t.Fatal("thumbnail to icon replacement kept stale plan")
	}
	queueCollectionActivation(r, "1", b)
	f.Items = slices.Delete(f.Items, 1, 2)
	a.Invalidate("one")
	// Reconcile removal before draining queued activation to retire catalog lease.
	if err := m.implementation.refresh(m.implementation.find("one")); err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if opens != 0 {
		t.Fatal("retired activation delivered")
	}
	f.Disabled = true
	a.Invalidate("one")
	unitSettle(t, m)
	queueCollectionActivation(r, "2", r.control.choice.buttons["2"])
	a.Drain()
	if opens != 0 {
		t.Fatal("disabled activation delivered")
	}
}

func TestMacCollectionSameIDPresentationTransitions(t *testing.T) {
	f := nativeCollectionFixture()
	f.Change = func(ui.ListSelectionEvent) {}
	collection := false
	m, a := simpleManual(t, manualConfig(t, 71), func() ui.Node {
		n, err := f.View()
		if err != nil {
			t.Fatal(err)
		}
		if !collection {
			n.Collection = nil
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{n}}
	})
	unitSettle(t, m)
	r := findRecord(m.implementation.find("one").state.root, "files")
	b := r.control.choice.buttons["0"]
	if r.widget.AsWidget().Styles.Wrap || b.ChildByName("stack", 0) == nil {
		t.Fatal("fixture normal list")
	}
	collection = true
	a.Invalidate("one")
	unitSettle(t, m)
	if !r.widget.AsWidget().Styles.Wrap || collectionColumns(r) < 2 || b.ChildByName("collection-icon", 0) == nil || b.ChildByName("stack", 0) != nil {
		t.Fatal("same-ID collection did not replace host/row presentation")
	}
	collection = false
	a.Invalidate("one")
	unitSettle(t, m)
	if r.widget.AsWidget().Styles.Wrap || b.ChildByName("stack", 0) == nil || b.ChildByName("collection-icon", 0) != nil {
		t.Fatal("same-ID ordinary list retained collection presentation")
	}
}

func TestMacCollectionRoutedCommandClicksAndDoubleActivation(t *testing.T) {
	f := nativeCollectionFixture()
	f.Items = f.Items[:4]
	f.SelectedIDs = nil
	f.Change = func(e ui.ListSelectionEvent) { f.SelectedIDs = e.Apply(f.SelectedIDs) }
	opens := 0
	f.Activate = func(e ui.CollectionEvent) {
		opens++
		if e.ItemID != "1" {
			t.Errorf("wrong opened item: %s", e.ItemID)
		}
	}
	m, _ := simpleManual(t, manualConfig(t, 73), func() ui.Node {
		n, err := f.View()
		if err != nil {
			t.Fatal(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{n}}
	})
	unitSettle(t, m)
	w := m.implementation.find("one")
	r := findRecord(w.state.root, "files")
	click := func(id string, mods key.Modifiers) {
		t.Helper()
		box := r.control.choice.buttons[id].Geom.TotalBBox
		if box.Empty() || !box.In(r.widget.AsWidget().Geom.ContentBBox) {
			t.Fatalf("fixture clipped tile %s %v", id, box)
		}
		for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: typ, Position: box.Min.Add(box.Size().Div(2)), Button: events.Left, Modifiers: mods}); err != nil {
				t.Fatal(err)
			}
		}
		unitSettle(t, m)
	}
	click("0", 0)
	click("1", 1<<key.Meta)
	if !slices.Equal(f.SelectedIDs, []string{"0", "1"}) {
		t.Fatal("Command-click lost controlled membership", f.SelectedIDs)
	}
	click("0", 1<<key.Meta)
	if !slices.Equal(f.SelectedIDs, []string{"1"}) {
		t.Fatal("Command-click did not toggle", f.SelectedIDs)
	}
	click("1", 0)
	click("1", 0)
	if opens != 1 {
		t.Fatal("routed double click activation count", opens)
	}
}

func TestMacCollectionReplacementRetiresSelectionActivationAndAX(t *testing.T) {
	for _, mode := range []string{"new-owner", "presentation-aba"} {
		t.Run(mode, func(t *testing.T) {
			current := nativeCollectionFixture()
			selectionCalls, activationCalls := 0, 0
			install := func(f *ui.CollectionView) {
				f.Change = func(ui.ListSelectionEvent) { selectionCalls++ }
				f.Activate = func(ui.CollectionEvent) { activationCalls++ }
			}
			install(current)
			showCollection := true
			m, a := simpleManual(t, manualConfig(t, 79), func() ui.Node {
				n, err := current.View()
				if err != nil {
					t.Fatal(err)
				}
				if !showCollection {
					n.Collection = nil
				}
				return n
			})
			unitSettle(t, m)
			w := m.implementation.find("one")
			r := w.state.root
			b := r.control.choice.buttons["1"]
			_, handlers := accessibilitySnapshot(r, nil)
			oldRow := handlers[accessibilityOptionID("files", "1")]
			oldHost := handlers[accessibilityManagedID("files")]
			if oldRow == nil || oldHost == nil {
				t.Fatal("fixture missing accessible actions")
			}
			queueCollectionActivation(r, "1", b)
			queueListSelection(r, "1", b, 0, "0")
			refresh := func() {
				t.Helper()
				a.Invalidate("one")
				if err := m.implementation.refresh(w); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "new-owner" {
				current = nativeCollectionFixture()
				install(current)
				refresh()
			} else {
				showCollection = false
				refresh()
				showCollection = true
				refresh()
			}
			a.Drain()
			if selectionCalls != 0 || activationCalls != 0 || oldRow(accessibilityAction{Kind: "press"}) || oldHost(accessibilityAction{Kind: "set", Value: "1"}) {
				t.Fatal("replacement inherited old input", selectionCalls, activationCalls)
			}
			queueCollectionActivation(r, "1", r.control.choice.buttons["1"])
			queueListSelection(r, "1", r.control.choice.buttons["1"], 0, "0")
			a.Drain()
			if selectionCalls != 1 || activationCalls != 1 {
				t.Fatal("fresh replacement input refused", selectionCalls, activationCalls)
			}
		})
	}
}
