//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestToolbarOverflowResponsiveMenuColorRoutingAndRetirement(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(fmt.Sprint(dark), func(t *testing.T) {
			source := fullColorButtonSource(t, false)
			width := float32(230)
			disabled, removed := false, false
			lastDisabled := false
			calls := 0
			cfg := manualConfig(t, 7)
			cfg.Theme = ui.MacTheme(dark)
			cfg.Theme.Default.Background = cfg.Theme.Kinds[ui.FormGroup].Background
			cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
			a, err := app.New(app.Config{})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(a.Close)
			view := func() ui.Node {
				items := []ui.Node{}
				for i := range 6 {
					i := i
					items = append(items, ui.Node{ID: fmt.Sprint("action-", i), Kind: ui.Button, Text: fmt.Sprint("Action ", i), IconImage: source, Disabled: i == 4 || i == 5 && lastDisabled, Activate: func() { calls++ }, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 95, Height: 40}), MaxSize: ui.Some(ui.Size{Width: 95, Height: 40}), GrowX: ui.Some(float32(0))}})
				}
				bar := ui.WithToolbarOverflow(ui.Node{ID: "tools", Kind: ui.Toolbar, Disabled: disabled, Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: width}), Gap: ui.Some(float32(4))}, Children: items})
				children := []ui.Node{bar, {ID: "after", Kind: ui.Button, Text: "After", Activate: func() {}}}
				if removed {
					children = children[1:]
				}
				return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: "action-0", Sequence: 1}, Children: children}
			}
			if err := a.Open(app.Window{ID: "one", Title: "Overflow", Width: 800, Height: 600, View: view}); err != nil {
				t.Fatal(err)
			}
			m, err := NewManual(testContext, a, cfg)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := m.Close(testContext); err != nil {
					t.Error(err)
				}
			})
			unitSettle(t, m)
			w := m.implementation.find("one")
			bar := findRecord(w.state.root, "tools")
			more := findRecord(w.state.root, ui.ToolbarOverflowID("tools"))
			if toolbarOverflowHidden(more) || !toolbarOverflowHidden(bar.children[5]) {
				t.Fatal("missing responsive overflow")
			}
			original := bar.children[5].widget
			for _, c := range bar.children {
				if !toolbarOverflowHidden(c) && !c.widget.AsWidget().Geom.TotalBBox.In(bar.widget.AsWidget().Geom.TotalBBox) {
					t.Fatal("visible toolbar control clipped", c.node.ID, c.widget.AsWidget().Geom.TotalBBox, bar.widget.AsWidget().Geom.TotalBBox)
				}
			}
			_, handlers := accessibilitySnapshot(w.state.root, nil)
			if handlers[accessibilityManagedID("action-5")] != nil {
				t.Fatal("hidden action still exposed")
			}
			keypress := func(code key.Codes) {
				t.Helper()
				if e := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code, Modifiers: 1 << key.NumLock}); e != nil {
					t.Fatal(e)
				}
				unitSettle(t, m)
			}
			keypress(key.CodeEnd)
			if !more.widget.AsWidget().ContainsFocus() {
				t.Fatal("End did not reach More", accessibilityFocusedID(w.state.root))
			}
			keypress(key.CodeDownArrow)
			if !menuOpen(more) {
				t.Fatal("Down did not open overflow")
			}
			p := more.control.menu.popup
			item := p.items["action-5"]
			if item == nil || item.Text != "Action 5" {
				t.Fatal("lost label")
			}
			picture, ok := item.ChildByName(buttonImageName).(*managedImage)
			if label := item.ChildByName("label"); label != nil && picture != nil && label.(interface{ AsWidget() *core.WidgetBase }).AsWidget().Geom.TotalBBox.Min.X <= picture.Geom.TotalBBox.Max.X {
				t.Fatal("menu artwork touches label")
			}
			if !ok || picture.source != source {
				t.Fatal("lost full-color source")
			}
			frame, e := m.Capture(testContext, "one")
			if e != nil {
				t.Fatal(e)
			}
			box := picture.Geom.ContentBBox.Add(p.scene.SceneGeom.Pos)
			y := (box.Min.Y + box.Max.Y) / 2
			left := frame.RGBAAt(box.Min.X+box.Dx()/6, y)
			right := frame.RGBAAt(box.Max.X-1-box.Dx()/6, y)
			if left.R < 180 || right.B < 180 || left.B > 60 || right.R > 60 {
				t.Fatal("popup recolored artwork", left, right, box)
			}
			_, handlers = accessibilitySnapshot(w.state.root, nil)
			stale := handlers[menuActionID(more.node.ID, "action-5")]
			if handlers[menuActionID(more.node.ID, "action-4")] != nil {
				t.Fatal("disabled menu action exposed")
			}
			snap, e := m.Snapshot(testContext)
			if e != nil {
				t.Fatal(e)
			}
			var target ManualElement
			for _, el := range snap.Windows[0].Elements {
				if el.ID == menuActionID(more.node.ID, "action-5") {
					target = el
				}
			}
			for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
				if e := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: typ, Target: target.ID, TargetIncarnation: target.Incarnation, Button: events.Left}); e != nil {
					t.Fatal(e)
				}
			}
			unitSettle(t, m)
			if calls != 1 || menuOpen(more) {
				t.Fatal("action not exactly once", calls)
			}
			if stale(accessibilityAction{Kind: "press"}) {
				t.Fatal("stale menu admitted")
			}
			refresh := func() {
				t.Helper()
				if e := a.Invalidate("one"); e != nil {
					t.Fatal(e)
				}
				unitSettle(t, m)
			}
			lastDisabled = true
			width = 700
			refresh()
			if !toolbarOverflowHidden(bar.children[5]) || toolbarOverflowHidden(bar.children[4]) {
				t.Fatal("disabled-suffix setup did not fit prefix", bar.widget.AsWidget().Geom.Size, bar.widget.(*overflowToolbarFrame).widths)
			}
			if menuAvailable(more) {
				t.Fatal("disabled-only suffix enabled menu")
			}
			_, disabledHandlers := accessibilitySnapshot(w.state.root, nil)
			if disabledHandlers[accessibilityManagedID(more.node.ID)] != nil {
				t.Fatal("disabled-only overflow exposed an action")
			}
			keypress(key.CodeHome)
			keypress(key.CodeEnd)
			if accessibilityFocusedID(w.state.root) != accessibilityManagedID("action-3") {
				t.Fatal("End reached disabled-only menu", accessibilityFocusedID(w.state.root))
			}
			lastDisabled = false
			width = 1000
			refresh()
			if !toolbarOverflowHidden(more) || toolbarOverflowHidden(bar.children[5]) || bar.children[5].widget != original {
				t.Fatal("widening lost identity/visibility")
			}
			width = 60
			refresh()
			if toolbarOverflowHidden(more) {
				t.Fatal("tiny toolbar lost menu")
			}
			if !more.widget.AsWidget().Geom.TotalBBox.In(bar.widget.AsWidget().Geom.TotalBBox) {
				t.Fatal("tiny toolbar clipped menu")
			}
			if !applyAccessibilityAction(more, accessibilityAction{Kind: "press"}) {
				t.Fatal("cannot reopen")
			}
			unitSettle(t, m)
			if !menuOpen(more) {
				t.Fatal("reopen")
			}
			oldPopup := more.control.menu.popup
			width = 350
			refresh()
			if menuOpen(more) || !oldPopup.dead.Load() {
				t.Fatal("resize did not revoke popup")
			}
			disabled = true
			refresh()
			if applyAccessibilityAction(more, accessibilityAction{Kind: "press"}) {
				t.Fatal("disabled overflow opened")
			}
			disabled = false
			refresh()
			applyAccessibilityAction(more, accessibilityAction{Kind: "press"})
			unitSettle(t, m)
			if !menuOpen(more) {
				t.Fatal("enabled overflow didn't return")
			}
			picture = more.control.menu.popup.items["action-5"].ChildByName(buttonImageName).(*managedImage)
			removed = true
			refresh()
			if !more.control.retired || menuOpen(more) || picture.source != nil {
				t.Fatal("removal retained popup/artwork")
			}

		})
	}
}

func TestMenuArtworkCheckedMarkersUseFreshPopup(t *testing.T) {
	source := fullColorButtonSource(t, false)
	checked := true
	m, a := simpleManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "menu", Kind: ui.MenuButton, Text: "Actions", Options: []ui.Option{
			{ID: "color", Label: "Color", IconImage: source, Checkable: true, Checked: checked},
			{ID: "symbol", Label: "Symbol", Icon: ui.SymbolFolder, Checkable: true, Checked: checked},
		}, Change: func(ui.ChangeEvent) {}}}}
	})
	unitSettle(t, m)
	r := findRecord(m.implementation.find("one").state.root, "menu")
	open := func() *menuPopup {
		t.Helper()
		if !applyAccessibilityAction(r, accessibilityAction{Kind: "press"}) {
			t.Fatal("menu refused")
		}
		unitSettle(t, m)
		return r.control.menu.popup
	}
	first := open()
	for _, id := range []string{"color", "symbol"} {
		if first.items[id].Indicator != nativeSymbol(ui.SymbolCheck) {
			t.Fatal("missing check alongside artwork", id)
		}
	}
	checked = false
	if e := a.Invalidate("one"); e != nil {
		t.Fatal(e)
	}
	unitSettle(t, m)
	second := open()
	for _, id := range []string{"color", "symbol"} {
		if second.items[id] == first.items[id] || second.items[id].Indicator != "" {
			t.Fatal("popup reused/stale check", id)
		}
	}
	if second.items["color"].ChildByName(buttonImageName).(*managedImage).source != source {
		t.Fatal("unchecked artwork lost")
	}
}

func TestToolbarOverflowCannotOpenBeforeFirstLayout(t *testing.T) {
	a, _ := newAXTest(t, 8)
	host := core.NewFrame(core.NewBody("overflow before layout"))
	n := ui.WithToolbarOverflow(ui.Node{ID: "tools", Kind: ui.Toolbar, Children: []ui.Node{{ID: "action", Kind: ui.Button, Text: "Action", Activate: func() {}}}})
	r, e := reconcile(host, nil, n, a, "ax")
	if e != nil {
		t.Fatal(e)
	}
	more := findRecord(r, ui.ToolbarOverflowID("tools"))
	if more.widget.AsWidget().Property(toolbarOverflowOwnerProperty) != r || len(menuDescriptor(more).Options) != 0 || menuAvailable(more) {
		t.Fatal("pre-layout menu exposed unfiltered actions")
	}
	disposeControlTree(r)
	if more.widget.AsWidget().Property(toolbarOverflowOwnerProperty) != nil {
		t.Fatal("disposed overflow retained owner")
	}
}
