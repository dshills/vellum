//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"image/color"
	"testing"
	"time"
)

func TestFieldRegainsActualFocusOnMultiClickAfterReenable(t *testing.T) {
	for _, kind := range []ui.Kind{ui.NumberField, ui.TextField} {
		for _, clicks := range []int{1, 2} {
			for _, separated := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/prior-clicks=%d/separated=%t", kind, clicks, separated), func(t *testing.T) {
					enabled := true
					number := float32(99)
					calls := 0
					text := "99"
					if kind == ui.TextField {
						text = "alpha beta"
					}
					originalText := text
					a, err := app.New(app.Config{})
					if err != nil {
						t.Fatal(err)
					}
					defer a.Close()
					view := func() ui.Node {
						return ui.Node{ID: "control-catalogue", Kind: ui.Column, Grow: true, Style: ui.Style{Gap: ui.Some(float32(12))}, Children: []ui.Node{
							{ID: "heading", Kind: ui.Label, Text: "Selector and numeric field acceptance"},
							{ID: "fields", Kind: ui.FormGroup, Text: "Ordinary controls", Disabled: !enabled, Children: []ui.Node{
								{ID: "selection", Kind: ui.Selector, SelectedID: "basic", Options: []ui.Option{{ID: "basic", Label: "Basic"}, {ID: "advanced", Label: "Advanced"}}, Change: func(ui.ChangeEvent) {}},
								{ID: "number", Kind: kind, Text: text, Number: number, Numeric: ui.Numeric{Min: 0, Max: 100, Step: 1}, Change: func(e ui.ChangeEvent) { number = e.Number; text = e.Text; calls++ }},
							}},
							{ID: "outside", Kind: ui.Button, Text: "Outside action"},
						}}
					}
					if err = a.Open(app.Window{ID: "one", Title: "Control prefix", Width: 729, Height: 658, View: view}); err != nil {
						t.Fatal(err)
					}
					cfg := manualConfig(t, 7)
					cfg.LogicalDPI = 162
					cfg.Theme = ui.Theme{Default: ui.Style{Background: ui.Some(color.NRGBA{R: 248, G: 248, B: 248, A: 255}), Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: 16}), Foreground: ui.Some(color.NRGBA{R: 32, G: 36, B: 42, A: 255})}, Kinds: map[ui.Kind]ui.Style{ui.Column: {Gap: ui.Some(float32(8))}}}
					m, err := NewManual(testContext, a, cfg)
					if err != nil {
						t.Fatal(err)
					}
					defer m.Close(testContext)
					unitSettle(t, m)
					w := m.implementation.find("one")
					click := func() {
						t.Helper()
						if separated {
							if _, err := cfg.Clock.Advance(core.TimingSettings.DoubleClickInterval + time.Millisecond); err != nil {
								t.Fatal(err)
							}
						}
						b := findRecord(w.state.root, "number").widget.AsWidget().Geom.TotalBBox
						for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
							if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: typ, Button: events.Left, Position: b.Min.Add(b.Size().Div(2))}); err != nil {
								t.Fatal(err)
							}
							unitSettle(t, m)
						}
					}
					keyUp := func() {
						t.Helper()
						if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: key.CodeUpArrow}); err != nil {
							t.Fatal(err)
						}
						unitSettle(t, m)
					}
					for range clicks {
						click()
					}
					if !findRecord(w.state.root, "number").widget.AsWidget().ContainsFocus() {
						t.Fatal("initial number not focused")
					}
					if err = a.Do(func() { enabled = false; _ = a.Invalidate("one") }); err != nil {
						t.Fatal(err)
					}
					unitSettle(t, m)
					keyUp()
					if calls != 0 || number != 99 {
						t.Fatal("disabled key mutated model")
					}
					if err = a.Do(func() { enabled = true; _ = a.Invalidate("one") }); err != nil {
						t.Fatal(err)
					}
					unitSettle(t, m)
					click()
					if !findRecord(w.state.root, "number").widget.AsWidget().ContainsFocus() {
						t.Fatal("number lost actual focus after disabled prefix", number, calls)
					}
					if calls != 0 || number != 99 || text != originalText {
						t.Fatal("refocus changed model", number, text, calls)
					}
					if !separated {
						var field *core.TextField
						switch widget := findRecord(w.state.root, "number").widget.(type) {
						case *core.Spinner:
							field = &widget.TextField
						case *core.TextField:
							field = widget
						}
						state := field.EditingState()
						lo, hi := min(state.Anchor, state.Focus), max(state.Anchor, state.Focus)
						if lo == hi || clicks == 2 && (lo != 0 || hi != len([]rune(originalText))) || kind == ui.TextField && clicks == 1 && hi-lo >= len([]rune(originalText)) {
							t.Fatal("ordinary multi-click selection lost", clicks, state)
						}
					}
					if kind == ui.NumberField {
						keyUp()
						if number != 100 || calls != 1 {
							t.Fatal("step after disabled prefix", number, calls)
						}
					} else {
						inputs := []ManualInput{{Type: events.KeyChord, Code: key.CodeUnknown, Rune: 'x'}, {Type: events.KeyChord, Code: key.CodeReturnEnter}}
						expectedText := "x"
						if separated {
							inputs = append([]ManualInput{{Type: events.KeyChord, Code: key.CodeEnd}}, inputs...)
							expectedText = originalText + "x"
						} else if clicks == 1 {
							expectedText = "alpha x"
						}
						for _, in := range inputs {
							in.Window = "one"
							in.WindowIncarnation = w.incarnation
							if err := m.Input(testContext, in); err != nil {
								t.Fatal(err)
							}
							unitSettle(t, m)
						}
						if text != expectedText || calls != 1 {
							t.Fatal("text commit after disabled prefix", text, calls)
						}
					}
				})
			}
		}
	}
}
