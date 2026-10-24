//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestMacSheetModalActionsFocusAndCleanup(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(map[bool]string{false: "light", true: "dark"}[dark], func(t *testing.T) {
			cfg := manualConfig(t, 982)
			bg := cfg.Theme.Default.Background
			cfg.Theme = ui.MacTheme(dark)
			cfg.Theme.Default.Background = bg
			cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
			s := &ui.Sheet{ID: "sheet", Title: "Export", ReturnFocusID: "opener", Style: cfg.Theme.Kinds[ui.FormGroup]}
			confirms, cancels := 0, 0
			s.Cancel = func() { cancels++; s.Open = false }
			s.Confirm = func() { confirms++; s.Open = false }
			m, a := simpleManual(t, cfg, func() ui.Node {
				panel, err := s.View()
				if err != nil {
					panic(err)
				}
				return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "opener", Kind: ui.Button, Text: "Export…", Activate: func() { s.Open = true }}, panel}}
			})
			unitSettle(t, m)
			w, opener := imageManualElement(t, m, "opener")
			click := func(target ManualElement) {
				t.Helper()
				for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
					if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: typ, Target: target.ID, TargetIncarnation: target.Incarnation, Button: events.Left}); err != nil {
						t.Fatal(err)
					}
				}
				unitSettle(t, m)
			}
			click(opener)
			root := m.implementation.find("one").state.root
			r := findRecord(root, "sheet")
			d := r.control.dialog
			if !d.active || d.body == nil {
				t.Fatal("sheet did not open")
			}
			stage := d.body.Scene.Stage
			if !stage.Modal || !stage.Scrim || stage.Resizable || stage.DisplayTitle || stage.NewWindow || stage.Scene.Parts != nil {
				t.Fatal("incorrect sheet presentation")
			}
			if len(ui.FocusOrder(ui.FocusScope(m.implementation.find("one").state.root.node))) != 2 {
				t.Fatal("modal scope leaked parent")
			}
			oldScope := d.tabScope.Load()
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeEscape}); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			if s.Open || cancels != 1 || confirms != 0 || d.active || !oldScope.closed.Load() || !findRecord(root, "opener").widget.AsWidget().ContainsFocus() {
				t.Fatal("Escape did not cancel/restore", s.Open, cancels, confirms)
			}
			click(opener)
			_, done := imageManualElement(t, m, ui.SheetConfirmID("sheet"))
			click(done)
			if s.Open || confirms != 1 || cancels != 1 {
				t.Fatal("confirmation duplicated", confirms, cancels)
			}
			s.Open = true
			a.Invalidate("one")
			unitSettle(t, m)
			if err := m.Close(testContext); err != nil {
				t.Fatal(err)
			}
			if d.active || d.body != nil {
				t.Fatal("window closure retained modal")
			}
		})
	}
}
