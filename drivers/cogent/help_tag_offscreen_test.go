//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestHelpTagKeyboardEscapeAndRetirement(t *testing.T) {
	enabled, show := true, true
	m, a := simpleManual(t, manualConfig(t, 1701), func() ui.Node {
		n, _ := ui.WithTooltip(ui.Node{ID: "hint", Kind: ui.Button, Text: "Save", Disabled: !enabled, Activate: func() {}}, "Save <plain> & safely")
		children := []ui.Node{{ID: "after", Kind: ui.Button, Text: "After", Activate: func() {}}}
		if show {
			children = append([]ui.Node{n}, children...)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: "hint", Sequence: 1}, Children: children}
	})
	unitSettle(t, m)
	r := findRecord(m.implementation.find("one").state.root, "hint")
	w, _ := imageManualElement(t, m, "hint")
	press := func(c key.Codes) {
		t.Helper()
		if e := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: c}); e != nil {
			t.Fatal(e)
		}
		unitSettle(t, m)
	}
	press(key.CodeF1)
	if r.helpTag == nil {
		t.Fatal("focused F1 did not show help")
	}
	press(key.CodeEscape)
	if r.helpTag != nil {
		t.Fatal("Escape left help open")
	}
	press(key.CodeF1)
	press(key.CodeTab)
	if r.helpTag != nil {
		t.Fatal("focus departure left help open")
	}
	if !showHelpTag(r) {
		t.Fatal("could not reopen before disable")
	}
	if e := a.Do(func() { enabled = false; a.Invalidate("one") }); e != nil {
		t.Fatal(e)
	}
	unitSettle(t, m)
	if showHelpTag(r) {
		t.Fatal("disabled target showed help")
	}
	if e := a.Do(func() { enabled = true; a.Invalidate("one") }); e != nil {
		t.Fatal(e)
	}
	unitSettle(t, m)
	if !showHelpTag(r) {
		t.Fatal("could not reopen before removal")
	}
	if e := a.Do(func() { show = false; a.Invalidate("one") }); e != nil {
		t.Fatal(e)
	}
	unitSettle(t, m)
	if r.helpTag != nil {
		t.Fatal("removed target retained popup")
	}
}
