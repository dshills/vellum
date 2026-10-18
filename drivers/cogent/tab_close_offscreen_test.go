//go:build offscreen

package cogent

import (
	"bytes"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"flag"
	"github.com/dshills/vellum/ui"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

var tabCloseCapture = flag.String("vellum-tab-close-capture", "", "optional exclusive absolute offscreen PNG output")

func TestTabCloseActualPointerKeyboardAndGeometry(t *testing.T) {
	calls, selections := 0, 0
	accept := false
	removed := false
	view := func() ui.Node {
		n := closeTabsFixture(func(e ui.TabCloseEvent) {
			calls++
			if e.PageID != "a" {
				t.Error("wrong routed page", e)
			}
			if accept {
				removed = true
			}
		})
		n.Change = func(ui.ChangeEvent) { selections++ }
		if removed {
			n.SelectedID = "b"
			n.Options = n.Options[1:]
			n.Children = n.Children[1:]
		}
		n.Focus = ui.FocusRequest{ID: "tabs", Sequence: 1}
		return n
	}
	m, _ := simpleManual(t, manualConfig(t, 37), view)
	unitSettle(t, m)
	before, close := imageManualElement(t, m, tabCloseID("tabs", "a"))
	if close.Bounds.Empty() || close.Incarnation == 0 || !close.Enabled || !close.Visible || before.FocusedID != "tabs" {
		t.Fatal("missing close geometry/focus", close, before.FocusedID)
	}
	first, err := m.Capture(testContext, "one")
	if err != nil {
		t.Fatal(err)
	}
	if *tabCloseCapture != "" {
		if !filepath.IsAbs(*tabCloseCapture) {
			t.Fatal("capture must be absolute")
		}
		f, e := os.OpenFile(*tabCloseCapture, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if e != nil {
			t.Fatal(e)
		}
		if e = png.Encode(f, first); e != nil {
			f.Close()
			t.Fatal(e)
		}
		if e = f.Close(); e != nil {
			t.Fatal(e)
		}
	}
	root := m.implementation.find("one").state.root
	page := root.control.tabs.pages["a"]
	binding := root.control.tabs.bindings["a"]
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if err = m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: before.Incarnation, Target: close.ID, TargetIncarnation: close.Incarnation, Type: typ, Button: events.Left}); err != nil {
			t.Fatal(err)
		}
	}
	unitSettle(t, m)
	if calls != 1 || selections != 0 || removed || root.control.tabs.pages["a"] != page {
		t.Fatal("actual pointer close selected or destroyed rejected page", calls, selections)
	}
	nodes, handlers := accessibilitySnapshot(root, nil)
	found := false
	for _, n := range nodes {
		if n.ID == close.ID {
			found = true
			if n.Role != "button" || n.Name != "Close First" || n.Parent != accessibilityOptionID("tabs", "a") || len(n.Actions) != 1 {
				t.Fatal("close semantics", n)
			}
		}
	}
	if !found || handlers[close.ID] == nil {
		t.Fatal("close semantic target missing")
	}
	bounds := accessibilityBounds(root)
	box := bounds[close.ID]
	if int(box[0]) != close.Bounds.Min.X || int(box[1]) != close.Bounds.Min.Y || int(box[2]) != close.Bounds.Dx() {
		t.Fatal("close semantic/manual bounds differ")
	}
	// Reject a real keyboard request while the focused tab, rather than an editor,
	// owns Delete/Backspace. Modified chords remain unconsumed by this handler.
	if err = m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: before.Incarnation, Type: events.KeyChord, Code: key.CodeBackspace}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if calls != 2 || selections != 0 || removed {
		t.Fatal("keyboard close did not remain controlled", calls, selections)
	}
	accept = true
	if !handlers[close.ID](accessibilityAction{Kind: "press"}) {
		t.Fatal("semantic press refused")
	}
	unitSettle(t, m)
	if calls != 3 || selections != 0 || !removed || root.control.tabs.pages["a"] != nil {
		t.Fatal("semantic accepted close missing", calls, selections)
	}
	last, err := m.Capture(testContext, "one")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(first.Pix, last.Pix) {
		t.Fatal("accepted close produced no completed pixel change")
	}
	if err = m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: before.Incarnation, Target: close.ID, TargetIncarnation: close.Incarnation, Type: events.MouseDown, Button: events.Left}); err != ErrManualStale {
		t.Fatal("retired close target admitted", err)
	}
	if binding.button.CloseRequest == nil {
		t.Fatal("retirement reenabled unmanaged deletion seam")
	}
}

func TestTabCloseRejectedPointerRetainsHeaderFocus(t *testing.T) {
	for _, closed := range []string{"a", "b"} {
		t.Run(closed, func(t *testing.T) {
			var proposals []string
			selections := 0
			view := func() ui.Node {
				n := closeTabsFixture(func(e ui.TabCloseEvent) { proposals = append(proposals, e.PageID) })
				n.Change = func(ui.ChangeEvent) { selections++ }
				n.Focus = ui.FocusRequest{ID: "tabs", Sequence: 1}
				return n
			}
			m, _ := simpleManual(t, manualConfig(t, 7), view)
			unitSettle(t, m)
			before, close := imageManualElement(t, m, tabCloseID("tabs", closed))
			root := m.implementation.find("one").state.root
			page := root.control.tabs.pages[closed]
			for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
				if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: before.Incarnation, Target: close.ID, TargetIncarnation: close.Incarnation, Type: typ, Button: events.Left}); err != nil {
					t.Fatal(err)
				}
			}
			unitSettle(t, m)
			snap, _ := imageManualElement(t, m, tabCloseID("tabs", closed))
			if len(proposals) != 1 || proposals[0] != closed || selections != 0 || root.node.SelectedID != "a" || root.control.tabs.pages[closed] != page {
				t.Fatal("rejected pointer close changed accepted state", proposals, selections)
			}
			if snap.FocusedID != "tabs" || !root.control.tabs.buttons[closed].ContainsFocus() {
				t.Fatal("rejected close lost actual header focus", closed, snap.FocusedID)
			}
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: before.Incarnation, Type: events.KeyChord, Code: key.CodeDelete}); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			if len(proposals) != 2 || proposals[1] != closed || selections != 0 || root.node.SelectedID != "a" || root.control.tabs.pages[closed] != page {
				t.Fatal("Delete did not use the retained header focus", proposals, selections)
			}
		})
	}
}
