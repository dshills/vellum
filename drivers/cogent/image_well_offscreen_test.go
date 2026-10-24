//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestImageWellNativeClearAndRetainedPreview(t *testing.T) {
	source := fullColorButtonSource(t, false)
	calls := 0
	well := &ui.ImageWell{ID: "well", Label: "Artwork", Image: source, Request: func(e ui.ImageWellEvent) bool { calls++; return true }}
	view := func() ui.Node {
		n, err := well.View()
		if err != nil {
			t.Fatal(err)
		}
		// Manual deliberately refuses external OS-drop transport; it is covered by FileDrop fixtures.
		n.Children[1].FileDrop = nil
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: "well/image-well/clear", Sequence: 1}, Children: []ui.Node{n}}
	}
	m, _ := simpleManual(t, manualConfig(t, 825), view)
	unitSettle(t, m)
	w := m.implementation.find("one")
	preview := findRecord(w.state.root, "well/image-well/preview").widget.(*managedImage)
	if preview.source != source {
		t.Fatal("image pixels changed")
	}
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: key.CodeReturnEnter}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if calls != 1 || well.Image != source || findRecord(w.state.root, "well/image-well/preview").widget != preview {
		t.Fatal("rejected clear changed preview or dispatched twice")
	}
}
