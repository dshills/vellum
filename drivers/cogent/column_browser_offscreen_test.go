//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestColumnBrowserNativeSelectionArtworkAndResize(t *testing.T) {
	source := fullColorButtonSource(t, false)
	b := &ui.ColumnBrowser{ID: "browser", Items: []ui.BrowserItem{{ID: "folder", Label: "Folder", Thumbnail: source, Children: []ui.BrowserItem{{ID: "leaf", Label: "Drawing", Thumbnail: source}}}}, Fractions: map[string]float32{}}
	b.Change = func(e ui.BrowserEvent) { b.Path = e.Path }
	b.Resize = func(e ui.BrowserResizeEvent) { b.Fractions[e.ColumnID] = e.Fraction }
	view := func() ui.Node {
		n, err := b.View()
		if err != nil {
			t.Fatal(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: ui.BrowserColumnID("browser", nil), Sequence: 1}, Children: []ui.Node{n}}
	}
	m, _ := simpleManual(t, manualConfig(t, 821), view)
	unitSettle(t, m)
	w := m.implementation.find("one")
	row := findRecord(w.state.root, ui.BrowserColumnID("browser", nil)).control.choice.buttons["folder"]
	image := row.ChildByName("list-thumbnail").(*managedImage)
	if image.source != source {
		t.Fatal("full-color source lost")
	}
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: key.CodeReturnEnter}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if len(b.Path) != 1 || findRecord(w.state.root, ui.BrowserColumnID("browser", []string{"folder"})) == nil {
		t.Fatal("keyboard did not reveal column")
	}
	// Tab reaches the divider; Right increases its accepted fraction.
	for _, code := range []key.Codes{key.CodeTab, key.CodeRightArrow} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	if b.Fractions[ui.BrowserColumnID("browser", nil)] <= .5 {
		t.Fatal("keyboard resize not accepted", b.Fractions)
	}
	if findRecord(w.state.root, ui.BrowserColumnID("browser", nil)).control.choice.buttons["folder"] != row {
		t.Fatal("parent row recreated")
	}
}
