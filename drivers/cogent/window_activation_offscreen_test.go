//go:build offscreen

package cogent

import (
	"image/color"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/styles"
	"github.com/dshills/vellum/ui"
)

func TestMacSourceListNativeWindowActivation(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(map[bool]string{false: "light", true: "dark"}[dark], func(t *testing.T) {
			cfg := manualConfig(t, 960)
			bg := cfg.Theme.Default.Background
			cfg.Theme = ui.MacTheme(dark)
			cfg.Theme.Default.Background = bg
			cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
			calls := 0
			id := "list"
			seq := uint64(1)
			view := func() ui.Node {
				return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: id, Sequence: seq}, Children: []ui.Node{{ID: id, Kind: ui.ListBox, Multiple: true, SelectedIDs: []string{"a", "b"}, Options: []ui.Option{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}}, SelectionChange: func(ui.ListSelectionEvent) { calls++ }}}}
			}
			m, a := simpleManual(t, cfg, view)
			unitSettle(t, m)
			managed := m.implementation.find("one")
			w := managed.state
			bindWindowOriginCustom(w, nil, func(*nativeWindow, events.Event) {})
			v := &windowActivation{}
			t.Cleanup(v.close)
			if err := bindWindowActivationState(w, v); err != nil {
				t.Fatal(err)
			}
			pump := func() {
				t.Helper()
				if _, err := managed.core.Step(testContext, 128); err != nil {
					t.Fatal(err)
				}
			}
			assertRows := func(active bool) {
				t.Helper()
				list := findRecord(w.root, id)
				if !list.widget.AsWidget().ContainsFocus() {
					t.Fatal("activation changed retained keyboard focus")
				}
				for _, row := range list.control.choice.buttons {
					if !active && row.Styles.Border.Width.Top.Value != 0 {
						t.Fatal("inactive source row kept an active focus outline")
					}
					got := color.NRGBAModel.Convert(row.Styles.Background.At(0, 0)).(color.NRGBA)
					if got != sourceSelectionBackground(active) {
						t.Fatal("row activation mismatch", got, active)
					}
				}
				if calls != 0 {
					t.Fatal("appearance emitted selection edits")
				}
			}
			assertRows(true)
			passes := 0
			w.host.FinalStyler(func(*styles.Style) { passes++ })
			// Fill the model mailbox: this appearance signal must neither drain nor use it.
			for range 4 {
				if err := a.Post(func() {}); err != nil {
					t.Fatal(err)
				}
			}
			v.observe(false)
			if passes != 0 {
				t.Fatal("callback directly styled native tree")
			}
			assertRows(true)
			pump()
			assertRows(false)
			if a.Pending().Ordinary != 4 {
				t.Fatal("activation drained the application queue")
			}
			before := passes
			v.observe(false)
			pump()
			if passes != before {
				t.Fatal("duplicate activation restyled")
			}
			v.observe(true)
			v.observe(false)
			pump()
			if passes != before {
				t.Fatal("coalesced unchanged state restyled")
			}
			if err := a.Drain(); err != nil {
				t.Fatal(err)
			}
			// New controls in an inactive window inherit the same native-window state.
			if err := a.Do(func() { id = "replacement"; seq++; _ = a.Invalidate("one") }); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			assertRows(false)
			v.observe(true)
			pump()
			assertRows(true)
			foreign := &windowActivation{}
			t.Cleanup(foreign.close)
			if err := foreign.bind(w.body.SystemEvents()); err != nil {
				t.Fatal(err)
			}
			before = passes
			foreign.observe(false)
			pump()
			assertRows(true)
			if passes != before {
				t.Fatal("foreign state restyled this window")
			}
			v.observe(false)
			v.close()
			pump()
			assertRows(true)
			if passes != before || v.pending != nil || v.source != nil {
				t.Fatal("closed activation retained native work")
			}
		})
	}
}
