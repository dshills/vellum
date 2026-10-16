//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestManualWindowRefReplacesSameNameAndRetiresOldInput(t *testing.T) {
	a, _ := app.New(app.Config{QueueCapacity: 32})
	defer a.Close()
	a.Open(scopeDeclaration("survivor"))
	calls := []int{0, 0}
	decl := func(index int) app.Window {
		spec := scopeDeclaration("same")
		spec.View = func() ui.Node {
			return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "button", Kind: ui.Button, Text: "Activate", Activate: func() { calls[index]++ }}, {ID: "field", Kind: ui.TextField, Text: "Original"}}}
		}
		return spec
	}
	old, err := a.OpenWindow(decl(0))
	if err != nil {
		t.Fatal(err)
	}
	m, err := NewManual(testContext, a, manualConfig(t, 7))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := m.Close(testContext); err != nil {
			t.Error(err)
		}
	}()
	unitSettle(t, m)
	oldNative := m.implementation.find("same")
	oldText := oldNative.state.input.Load()
	var button *record
	for _, r := range oldNative.state.root.children {
		if r.node.ID == "button" {
			button = r
		}
	}
	if button == nil {
		t.Fatal("button missing")
	}
	if err = a.PostWindow(old, func() { calls[0]++ }); err != nil {
		t.Fatal(err)
	}
	if err = a.CloseWindow(old); err != nil {
		t.Fatal(err)
	}
	replacement, err := a.OpenWindow(decl(1))
	if err != nil {
		t.Fatal(err)
	}
	// Delivery already in the native listener, after domain closure but before
	// explicit topology reconciliation, must not target either opening.
	button.widget.AsWidget().Send(events.Click)
	if lookupTextWindow("same") != nil || oldText.acceptsNativeInput() {
		t.Fatal("old text attachment exposed as current")
	}
	stale := ManualInput{Window: "same", WindowIncarnation: oldNative.incarnation, Type: events.MouseDown, Button: events.Left}
	if err = m.Input(testContext, stale); !errors.Is(err, ErrManualStale) {
		t.Fatal("old native incarnation admitted", err)
	}
	unitSettle(t, m)
	next := m.implementation.find("same")
	if next == nil || next == oldNative || next.state.window.Ref() != replacement || next.incarnation == oldNative.incarnation {
		t.Fatal("native name replacement did not occur")
	}
	if !oldNative.state.closing.Load() || oldNative.state.root != nil || oldText.acceptsNativeInput() || lookupTextWindow("same") != next.state.input.Load() {
		t.Fatal("old native/text state was retained or displaced replacement")
	}
	if oldNative.state.input.Load().closed.Load() != true || replacement.Context().Err() != nil || calls[0] != 0 || calls[1] != 0 {
		t.Fatal("retirement affected replacement", calls)
	}
	window, el := manualDockElement(t, m, "same", "button")
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if err = m.Input(testContext, ManualInput{Window: window.ID, WindowIncarnation: window.Incarnation, Type: typ, Button: events.Left, Position: el.Bounds.Min.Add(el.Bounds.Size().Div(2))}); err != nil {
			t.Fatal(err)
		}
	}
	unitSettle(t, m)
	if calls[0] != 0 || calls[1] != 1 {
		t.Fatal("replacement routing did not commit exactly once", calls)
	}
	if err = a.CloseWindow(old); !errors.Is(err, app.ErrWindowClosed) {
		t.Fatal("retired close accepted", err)
	}
	if replacement.Context().Err() != nil {
		t.Fatal("old close retired replacement")
	}
}
