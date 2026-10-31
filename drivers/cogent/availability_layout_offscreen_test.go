//go:build offscreen

package cogent

import (
	"fmt"
	"image"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/styles/states"
	"github.com/dshills/vellum/ui"
)

func TestAvailabilityKeepsControlLayout(t *testing.T) {
	for _, dpi := range []float32{96, 160, 162, 192} {
		for _, mac := range []bool{false, true} {
			t.Run(fmt.Sprintf("dpi%.0f/mac%t", dpi, mac), func(t *testing.T) {
				cfg := manualConfig(t, 7)
				cfg.LogicalDPI = dpi
				var theme *ui.Theme
				if mac {
					bg := cfg.Theme.Default.Background
					cfg.Theme = ui.MacTheme(false)
					cfg.Theme.Default.Background = bg
					cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
				} else {
					theme = &ui.Theme{}
				}
				disabled := false
				calls := 0
				view := func() ui.Node {
					return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Theme: theme, Children: []ui.Node{
						{ID: "actions", Kind: ui.Row, Children: []ui.Node{
							{ID: "button", Kind: ui.Button, Text: "Action", Disabled: disabled, Activate: func() { calls++ }},
							{ID: "menu", Kind: ui.MenuButton, Text: "Menu", Disabled: disabled, Options: []ui.Option{{ID: "action", Label: "Action"}}, Change: func(ui.ChangeEvent) { calls++ }},
						}},
						{ID: "check", Kind: ui.Checkbox, Text: "Check", Disabled: disabled, Change: func(ui.ChangeEvent) { calls++ }},
						{ID: "field", Kind: ui.TextField, Text: "Text", Disabled: disabled, Change: func(ui.ChangeEvent) { calls++ }},
						{ID: "sibling", Kind: ui.Label, Text: "Unrelated content"},
					}}
				}
				m, a := simpleManual(t, cfg, view)
				unitSettle(t, m)
				bounds := func() map[string]image.Rectangle {
					s, err := m.Snapshot(testContext)
					if err != nil {
						t.Fatal(err)
					}
					out := map[string]image.Rectangle{}
					for _, e := range s.Windows[0].Elements {
						out[e.ID] = e.Bounds
					}
					return out
				}
				before := bounds()
				for _, value := range []bool{true, false, true, false} {
					if err := a.Post(func() { disabled = value; _ = a.Invalidate("one") }); err != nil {
						t.Fatal(err)
					}
					unitSettle(t, m)
					after := bounds()
					for _, id := range []string{"button", "menu", "check", "field"} {
						native := findRecord(m.implementation.find("one").state.root, id).widget.AsWidget()
						if native.StateIs(states.Disabled) != value {
							t.Fatalf("%s lost disabled state", id)
						}
						if value && native.Styles.Opacity >= 1 {
							t.Fatalf("%s lost disabled appearance", id)
						}
					}
					for _, id := range []string{"button", "menu", "check", "field", "sibling"} {
						if after[id] != before[id] {
							t.Errorf("disabled=%t changed %s: before=%v after=%v", value, id, before[id], after[id])
						}
					}
					win, button := imageManualElement(t, m, "button")
					for _, kind := range []events.Types{events.MouseDown, events.MouseUp} {
						if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: win.Incarnation, Type: kind, Button: events.Left, Position: button.Bounds.Min.Add(button.Bounds.Size().Div(2))}); err != nil {
							t.Fatal(err)
						}
					}
					unitSettle(t, m)
				}
				if calls != 2 {
					t.Fatalf("disabled controls admitted action or enabled controls refused: calls=%d", calls)
				}
			})
		}
	}
}
