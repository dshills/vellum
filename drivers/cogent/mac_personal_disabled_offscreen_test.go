//go:build offscreen

package cogent

import (
	"errors"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
)

// Native pointer verification is separate. These are actual mounted Manual
// focus, event-admission and retained-control transitions, without a desktop.
func TestMacPersonalWidgetsDisableReenableRouting(t *testing.T) {
	for _, kind := range []string{"secure", "path", "font", "tokens", "color"} {
		t.Run(kind, func(t *testing.T) {
			calls := 0
			var render func() (ui.Node, error)
			var disable func(bool)
			var target string
			var secureValue func() string
			switch kind {
			case "secure":
				p := &ui.SecureField{ID: "secure", Text: "synthetic", AllowReveal: true}
				p.Change = func(e ui.ChangeEvent) { calls++; p.Text = e.Text }
				secureValue = func() string { return p.Text }
				render, disable, target = p.View, func(v bool) { p.Disabled = v }, ui.SecureInputID(p.ID)
			case "path":
				p := &ui.PathControl{ID: "path", Items: []ui.PathItem{{ID: "home", Label: "Home"}, {ID: "file", Label: "Draft.vellum", Kind: ui.PathFile}}, Navigate: func(ui.PathEvent) { calls++ }}
				render, disable, target = p.View, func(v bool) { p.Disabled = v }, ui.PathSegmentID(p.ID, "home")
			case "font":
				p := &ui.FontPicker{ID: "font", Value: ui.FontSelection{FamilyID: "noto", TypefaceID: "regular", Size: 16}, Families: []ui.FontFamily{{ID: "noto", Name: "Noto Sans", Typefaces: []ui.FontTypeface{{ID: "regular", Name: "Regular", Weight: 400}}}}}
				p.Change = func(e ui.FontSelectionEvent) { calls++; p.Value = e.Apply(p.Value) }
				render, disable, target = p.View, func(v bool) { p.Disabled = v }, ui.FontSizeID(p.ID)
			case "tokens":
				p := &ui.TokenField{ID: "tags", Tokens: []ui.Token{{ID: "design", Text: "Design"}}, Change: func(ui.TokenEvent) { calls++ }}
				render, disable, target = p.View, func(v bool) { p.Disabled = v }, ui.TokenRemoveID(p.ID, "design")
			case "color":
				p := &ui.ColorWell{ID: "well", Change: func(ui.ColorEdit) { calls++ }}
				render, disable, target = p.View, func(v bool) { p.Disabled = v }, ui.ColorWellButtonID(p.ID)
			}
			focus := ui.FocusRequest{ID: target, Sequence: 1}
			view := func() ui.Node {
				n, err := render()
				if err != nil {
					panic(err)
				}
				return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: focus, Children: []ui.Node{n, {ID: "after", Kind: ui.Button, Text: "After", Activate: func() {}}}}
			}
			m, a := simpleManual(t, manualConfig(t, 949), view)
			unitSettle(t, m)
			original := findRecord(m.implementation.find("one").state.root, target).widget
			if err := a.Do(func() { disable(true); _ = a.Invalidate("one") }); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			w, element := imageManualElement(t, m, target)
			if element.Enabled {
				t.Fatal("disabled component remains enabled")
			}
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Target: element.ID, TargetIncarnation: element.Incarnation, Type: events.MouseDown, Button: events.Left}); !errors.Is(err, ErrManualDisabled) {
				t.Fatal("disabled target admitted pointer", err)
			}
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeReturnEnter}); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			if calls != 0 {
				t.Fatal("disabled keyboard action delivered", calls)
			}
			if err := a.Do(func() { disable(false); focus.Sequence++; _ = a.Invalidate("one") }); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			w, element = imageManualElement(t, m, target)
			if !element.Enabled || !element.Focused || findRecord(m.implementation.find("one").state.root, target).widget != original {
				t.Fatal("reenable lost retained control or focus", element)
			}
			send := func(code key.Codes, r rune) {
				t.Helper()
				if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: code, Rune: r}); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
			}
			expectedSecure := ""
			switch kind {
			case "secure":
				before := original.(*core.TextField).EditingState()
				start, end := before.Anchor, before.Focus
				if start > end {
					start, end = end, start
				}
				// The fixture is ASCII; its UTF-16 selection offsets are byte offsets.
				expectedSecure = before.Text[:start] + "x" + before.Text[end:]
				send(key.CodeX, 'x')
				if calls != 0 || secureValue() != "synthetic" {
					t.Fatal("keystroke committed before Return")
				}
				send(key.CodeReturnEnter, 0)
			case "font":
				send(key.CodeUpArrow, 0)
			default:
				send(key.CodeReturnEnter, 0)
			}
			if kind == "color" {
				panel := findRecord(m.implementation.find("one").state.root, "well/color-well/popup")
				if panel == nil || panel.control == nil || panel.control.dialog == nil || !panel.node.Open || !panel.control.dialog.active {
					t.Fatal("reenabled color well did not open")
				}
				send(key.CodeEscape, 0)
				panel = findRecord(m.implementation.find("one").state.root, "well/color-well/popup")
				if panel == nil || panel.control == nil || panel.control.dialog == nil {
					t.Fatal("color panel record not retained")
				}
				if panel.node.Open || panel.control.dialog.active || calls != 0 {
					t.Fatal("reenabled color cancel created edit or retained panel")
				}
			} else if calls != 1 {
				t.Fatal("reenabled control did not propose exactly one edit", calls)
			}
			if secureValue != nil && secureValue() != expectedSecure {
				t.Fatal("secure typed value not accepted", secureValue())
			}
		})
	}
}
