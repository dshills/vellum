//go:build offscreen

package cogent

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles/units"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/workbench"
)

func TestListBoxInPlaceReorderUpdatesMeasuredRows(t *testing.T) {
	options := []ui.Option{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}, {ID: "c", Label: "Charlie"}}
	selected := "a"
	m, a := simpleManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "list", Kind: ui.ListBox, SelectedID: selected, Options: options, Change: func(e ui.ChangeEvent) { selected = e.SelectedID }}
	})
	unitSettle(t, m)
	w := m.implementation.find("one")
	r := w.state.root
	oldA, oldB := r.control.choice.buttons["a"], r.control.choice.buttons["b"]
	before, err := m.Capture(testContext, "one")
	if err != nil {
		t.Fatal(err)
	}
	if oldA.Geom.TotalBBox.Min.Y >= oldB.Geom.TotalBBox.Min.Y {
		t.Fatal("fixture initial order")
	}
	options[0], options[1] = options[1], options[0]
	a.Invalidate("one")
	unitSettle(t, m)
	if oldB.Geom.TotalBBox.Min.Y >= oldA.Geom.TotalBBox.Min.Y || r.widget.AsTree().Child(0) != oldB {
		t.Fatal("in-place option reorder retained stale row geometry", oldA.Geom.TotalBBox, oldB.Geom.TotalBBox)
	}
	after, err := m.Capture(testContext, "one")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(before.Pix, after.Pix) {
		t.Fatal("reordered selected row did not change pixels")
	}
	box := oldB.Geom.TotalBBox
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: typ, Position: box.Min.Add(box.Size().Div(2)), Button: events.Left}); err != nil {
			t.Fatal(err)
		}
	}
	unitSettle(t, m)
	if selected != "b" {
		t.Fatal("pointer used stale reorder geometry", selected)
	}
}

func TestListBoxActualWorkbenchModalFocusAndClosure(t *testing.T) {
	a, err := app.New(app.Config{QueueCapacity: 16})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	model, err := workbench.New(workbench.Layout{Version: workbench.Version, Root: &workbench.Group{ID: "dock", Tabs: []string{"pane"}, Active: "pane"}}, []string{"pane"})
	if err != nil {
		t.Fatal(err)
	}
	open := true
	calls := 0
	dock := &Workbench{Owner: a, Window: "one", Workspace: model, Panels: []Panel{{ID: "pane", Title: "Pane", View: func() ui.Node {
		return ui.Node{ID: "panel", Kind: ui.Column, Children: []ui.Node{{ID: "modal", Kind: ui.Dialog, Open: open, Text: "Items", Dismiss: func() { open = false }, Children: []ui.Node{listBoxDescriptor(func(ui.ChangeEvent) { calls++ })}}}}
	}}}}
	if err := a.Open(app.Window{ID: "one", Title: "Lists", Width: 800, Height: 560, View: func() ui.Node { return ui.Node{ID: "workbench", Kind: ui.Extension, Custom: dock, Grow: true} }}); err != nil {
		t.Fatal(err)
	}
	m, err := NewManual(testContext, a, manualConfig(t, 7))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close(testContext)
	unitSettle(t, m)
	w := m.implementation.find("one")
	r := findRecord(dock.records["pane"], "list")
	b := &accessibilityBridge{}
	_, handlers := accessibilitySnapshot(w.state.root, b)
	id := accessibilityWorkbenchNamespace("workbench", "pane") + accessibilityOptionID("list", "b")
	if handlers[id] == nil || !handlers[id](accessibilityAction{Kind: "focus"}) {
		t.Fatal("namespaced current modal list focus refused")
	}
	pending := b.pendingFocus.Swap(nil)
	if pending == nil || *pending != id || !accessibilityApplyFocus(w.state.root, id) {
		t.Fatal("namespaced focus did not resolve current row")
	}
	if !r.control.choice.buttons["b"].ContainsFocus() {
		t.Fatal("row did not get actual native focus")
	}
	queueListBox(r, "b", r.control.choice.buttons["b"])
	scope := r.control.choice.listScope.Load()
	if !scope.allowed() || len(scope.scopes) == 0 || scope.scopes[0] == nil {
		t.Fatal("fixture has no native modal lease")
	}
	// Native dismissal retires this atomic lease before the owner's next view.
	scope.scopes[0].closed.Store(true)
	a.Drain()
	if calls != 0 || handlers[id](accessibilityAction{Kind: "press"}) {
		t.Fatal("native modal closure admitted queued list mutation")
	}
	open = false
	a.Invalidate("one")
	unitSettle(t, m)
}

func TestListBoxNarrowLargeFontAndStyleRemoval(t *testing.T) {
	for _, dpi := range []float32{96, 120, 160, 240} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			config := manualConfig(t, 7)
			config.LogicalDPI = dpi
			config.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 28})
			styled := true
			m, a := simpleManual(t, config, func() ui.Node {
				n := listBoxDescriptor(func(ui.ChangeEvent) {})
				n.Options[0].Label = "A long <literal> background label"
				if styled {
					n.Style = ui.Style{MaxSize: ui.Some(ui.Size{Width: 180, Height: 100})}
				}
				return n
			})
			unitSettle(t, m)
			w := m.implementation.find("one")
			r := w.state.root
			if r.widget.AsWidget().Geom.ContentBBox.Empty() {
				t.Fatal("narrow large-font list has empty viewport")
			}
			if r.control.choice.buttons["a"].Text != "A long &lt;literal&gt; background label" {
				t.Fatal("label interpreted as markup")
			}
			styled = false
			a.Invalidate("one")
			unitSettle(t, m)
			s := r.widget.AsWidget().Styles
			if s.Max.Y.Value != 240 || s.Max.Y.Unit != units.UnitDp || s.Min.X.Value != 160 {
				t.Fatal("removing style did not restore native defaults", s.Min, s.Max)
			}
		})
	}
}

func TestListBoxHeldActualKeysPreserveFIFO(t *testing.T) {
	for _, batch := range []int{1, 7, 64} {
		for _, accept := range []bool{false, true} {
			t.Run(fmt.Sprintf("batch%d/accept%v", batch, accept), func(t *testing.T) {
				selected := "a"
				var proposals []string
				m, a := simpleManual(t, manualConfig(t, batch), func() ui.Node {
					n := listBoxDescriptor(func(e ui.ChangeEvent) {
						proposals = append(proposals, e.SelectedID)
						if accept {
							selected = e.SelectedID
						}
					})
					n.SelectedID = selected
					return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "list", Sequence: 1}, Children: []ui.Node{n}}
				})
				unitSettle(t, m)
				w := m.implementation.find("one")
				for _, code := range []key.Codes{key.CodeDownArrow, key.CodeUpArrow} {
					if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code}); err != nil {
						t.Fatal(err)
					}
				}
				// Route actual queued key events while withholding owner dispatch. This
				// differs from directly invoking the private choice helper.
				m.implementation.ownerBudget = 0
				for range 64 {
					if a.Pending().Ordinary == 2 {
						break
					}
					if _, err := w.core.Step(testContext, 1); err != nil {
						t.Fatal(err)
					}
				}
				if a.Pending().Ordinary != 2 || len(proposals) != 0 {
					t.Fatal("did not hold both actual input proposals", a.Pending(), proposals)
				}
				for a.Pending().Ordinary > 0 {
					if _, err := a.DrainN(batch); err != nil {
						t.Fatal(err)
					}
					if err := m.implementation.refresh(w); err != nil {
						t.Fatal(err)
					}
				}
				unitSettle(t, m)
				if !reflect.DeepEqual(proposals, []string{"b", "a"}) || selected != "a" {
					t.Fatal("held input lost or reordered", selected, proposals)
				}
			})
		}
	}
}

func TestListBoxActualScrollingNavigationAndPixels(t *testing.T) {
	for _, dpi := range []float32{96, 120, 160, 240} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			config := manualConfig(t, 7)
			config.LogicalDPI = dpi
			options := make([]ui.Option, 32)
			for i := range options {
				options[i] = ui.Option{ID: fmt.Sprint(i), Label: fmt.Sprintf("Item %02d <literal>", i), Disabled: i == 1}
			}
			options[24].Label = "Zebra"
			options[25].Label = "Zinc"
			selected := "0"
			calls := 0
			readOnly := false
			m, a := gridManual(t, config, func() ui.Node {
				var change func(ui.ChangeEvent)
				if !readOnly {
					change = func(e ui.ChangeEvent) { selected = e.SelectedID; calls++ }
				}
				return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "list", Sequence: 1}, Children: []ui.Node{{ID: "list", Kind: ui.ListBox, SelectedID: selected, Options: options, Style: ui.Style{MaxSize: ui.Some(ui.Size{Height: 170})}, Change: change}, {ID: "after", Kind: ui.Button, Text: "After"}}}
			})
			unitSettle(t, m)
			w := m.implementation.find("one")
			r := findRecord(w.state.root, "list")
			if r.widget.AsWidget().Geom.ContentBBox.Empty() || !r.control.choice.buttons["31"].Geom.TotalBBox.Empty() {
				t.Fatal("list has no bounded scroll viewport", r.widget.AsWidget().Geom.ContentBBox)
			}
			before, _ := m.Capture(testContext, "one")
			send := func(code key.Codes, runeValue rune) {
				t.Helper()
				if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code, Rune: runeValue}); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
			}
			send(key.CodeDownArrow, 0)
			if selected != "2" {
				t.Fatal("disabled row not skipped", selected)
			}
			for _, code := range []key.Codes{key.CodePageDown, key.CodeEnd} {
				old := selected
				send(code, 0)
				if selected == old || r.control.choice.buttons[selected].Geom.TotalBBox.Empty() {
					t.Fatal("navigation failed to reveal target", code, selected)
				}
			}
			oldCalls := calls
			send(key.CodeDownArrow, 0)
			if selected != "31" || calls != oldCalls {
				t.Fatal("end boundary wrapped or proposed change", selected, calls)
			}
			for _, focusID := range []string{accessibilityOptionID("list", "31"), accessibilityManagedID("list")} {
				box := r.widget.AsWidget().Geom.ContentBBox
				if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.Scroll, Position: box.Min.Add(box.Size().Div(2)), Delta: math32.Vec2(0, -10000)}); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
				if !r.control.choice.buttons["31"].Geom.TotalBBox.Empty() || !r.control.choice.buttons["31"].ContainsFocus() {
					t.Fatal("wheel-away did not hide the focused row")
				}
				if !accessibilityApplyFocus(w.state.root, focusID) {
					t.Fatal("current focused row semantic focus refused")
				}
				unitSettle(t, m)
				if r.control.choice.buttons["31"].Geom.TotalBBox.Empty() {
					t.Fatal("same-focus semantic request did not reveal row", focusID)
				}
			}
			bounds := accessibilityBounds(w.state.root)
			if bounds[accessibilityOptionID("list", "0")][3] != 0 || bounds[accessibilityOptionID("list", "31")][3] <= 0 {
				t.Fatal("AX row geometry is not clipped", bounds)
			}
			send(key.CodePageUp, 0)
			if selected == "31" {
				t.Fatal("PageUp did not move")
			}
			send(key.CodeHome, 0)
			if selected != "0" || r.control.choice.buttons["0"].Geom.TotalBBox.Empty() {
				t.Fatal("Home did not reveal start")
			}
			send(key.CodeZ, 'z')
			if selected != "24" {
				t.Fatal("initial letter navigation", selected)
			}
			send(key.CodeZ, 'z')
			if selected != "25" {
				t.Fatal("initial letter did not cycle", selected)
			}
			box := r.control.choice.buttons["24"].Geom.TotalBBox
			if box.Empty() {
				t.Fatal("neighbor row not visible")
			}
			for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
				if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: typ, Position: box.Min.Add(box.Size().Div(2)), Button: events.Left}); err != nil {
					t.Fatal(err)
				}
			}
			unitSettle(t, m)
			if selected != "24" {
				t.Fatal("measured pointer did not select row", selected)
			}
			after, _ := m.Capture(testContext, "one")
			if bytes.Equal(before.Pix, after.Pix) {
				t.Fatal("list input produced no raster change")
			}
			readOnly = true
			a.Invalidate("one")
			unitSettle(t, m)
			oldCalls = calls
			send(key.CodeEnd, 0)
			if selected != "24" || calls != oldCalls || !r.control.choice.buttons["31"].ContainsFocus() {
				t.Fatal("read-only focus/selection separated incorrectly")
			}
			send(key.CodeTab, 0)
			s, _ := m.Snapshot(testContext)
			if s.Windows[0].FocusedID != "after" {
				t.Fatal("list added more than one managed Tab stop", s.Windows[0].FocusedID)
			}
		})
	}
}

func TestListBoxObservedScopeABA(t *testing.T) {
	for _, kind := range []string{"tab", "disclosure", "modal"} {
		t.Run(kind, func(t *testing.T) {
			blocked := false
			calls := 0
			m, a := simpleManual(t, manualConfig(t, 7), func() ui.Node {
				list := listBoxDescriptor(func(ui.ChangeEvent) { calls++ })
				switch kind {
				case "tab":
					selected := "shown"
					if blocked {
						selected = "hidden"
					}
					return ui.Node{ID: "tabs", Kind: ui.Tabs, SelectedID: selected, Options: []ui.Option{{ID: "shown", Label: "Shown"}, {ID: "hidden", Label: "Hidden"}}, Children: []ui.Node{{ID: "shown", Kind: ui.Column, Children: []ui.Node{list}}, {ID: "hidden", Kind: ui.Label, Text: "Hidden"}}}
				case "disclosure":
					return ui.Node{ID: "section", Kind: ui.Disclosure, Text: "Items", Checked: !blocked, Children: []ui.Node{list}}
				default:
					return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{list, {ID: "modal", Kind: ui.Dialog, Text: "Block", Open: blocked, Children: []ui.Node{{ID: "message", Kind: ui.Label, Text: "Modal"}}}}}
				}
			})
			unitSettle(t, m)
			w := m.implementation.find("one")
			r := findRecord(w.state.root, "list")
			queueListBox(r, "b", r.control.choice.buttons["b"])
			for _, next := range []bool{true, false} {
				blocked = next
				a.Invalidate("one")
				if err := m.implementation.refresh(w); err != nil {
					t.Fatal(err)
				}
			}
			a.Drain()
			unitSettle(t, m)
			if calls != 0 {
				t.Fatal("old input survived observed visibility ABA", calls)
			}
			queueListBox(r, "b", r.control.choice.buttons["b"])
			a.Drain()
			unitSettle(t, m)
			if calls != 1 {
				t.Fatal("newly visible list input refused")
			}
		})
	}
}
