//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestMacSidebarAdornmentsSelectionAndCollapse(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(map[bool]string{false: "light", true: "dark"}[dark], func(t *testing.T) {
			cfg := manualConfig(t, 980)
			bg := cfg.Theme.Default.Background
			cfg.Theme = ui.MacTheme(dark)
			cfg.Theme.Default.Background = bg
			cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
			s := &ui.Sidebar{ID: "nav", Dark: dark, SelectedID: "a", Sections: []ui.SidebarSection{{ID: "g", Label: "Library", Items: []ui.Option{{ID: "a", Label: "All", Icon: ui.SymbolFolder, Badge: "12"}, {ID: "b", Label: "Recent", Icon: ui.SymbolDocument}}}}}
			calls := 0
			s.Navigate = func(e ui.SidebarEvent) { calls++; s.SelectedID = e.ItemID }
			s.Disclosure = func(e ui.SidebarDisclosureEvent) { s.Sections[0].Collapsed = !e.Expanded }
			m, a := simpleManual(t, cfg, func() ui.Node {
				n, err := s.View()
				if err != nil {
					panic(err)
				}
				n.Focus = ui.FocusRequest{ID: ui.SidebarListID("nav", "g"), Sequence: 1}
				return n
			})
			unitSettle(t, m)
			if _, err := m.Capture(testContext, "one"); err != nil {
				t.Fatal(err)
			}
			w := m.implementation.find("one")
			r := findRecord(w.state.root, ui.SidebarListID("nav", "g"))
			button := r.control.choice.buttons["a"]
			if button.IconOn != nativeSymbol(ui.SymbolFolder) || button.IconOn != button.IconOff || button.Styles.IconSize.X.Value != 16 {
				t.Fatal("sidebar glyph missing")
			}
			badge, ok := button.ChildByName("list-badge", 0).(*core.Text)
			if !ok || badge.Text != "12" {
				t.Fatal("badge missing")
			}
			if badge.Geom.TotalBBox.Min.X <= button.ChildByName("text", 0).AsTree().This.(core.Widget).AsWidget().Geom.TotalBBox.Min.X {
				t.Fatal("badge not trailing", badge.Geom.TotalBBox)
			}
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: key.CodeDownArrow}); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			if s.SelectedID != "b" || calls != 1 {
				t.Fatal("navigation", s.SelectedID, calls)
			}
			s.Sections[0].Items[0].Badge = ""
			s.Sections[0].Items[0].Icon = ""
			a.Invalidate("one")
			unitSettle(t, m)
			if button.ChildByName("list-badge", 0) != nil || button.Styles.IconSize.X.Value != 0 {
				t.Fatal("adornments not removed")
			}
			s.Sections[0].Collapsed = true
			a.Invalidate("one")
			unitSettle(t, m)
			queueListBox(r, "a", button)
			a.Drain()
			if calls != 1 {
				t.Fatal("collapsed navigation")
			}
		})
	}
}
