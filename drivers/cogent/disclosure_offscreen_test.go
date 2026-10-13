//go:build offscreen

package cogent

import (
	"bytes"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
)

func TestDisclosureManualPointerKeyboardGeometryAndFocus(t *testing.T) {
	expanded := true
	changes := 0
	text := "Accepted"
	view := func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "details", Sequence: 1}, Children: []ui.Node{
			{ID: "details", Kind: ui.Disclosure, Text: "Inspector", Checked: expanded, Change: func(e ui.ChangeEvent) { expanded = e.Checked; changes++ }, Children: []ui.Node{
				{ID: "field", Kind: ui.TextField, Text: text, Change: func(e ui.ChangeEvent) { text = e.Text }},
				{ID: "content", Kind: ui.Label, Text: "Retained section content"},
			}},
			{ID: "after", Kind: ui.Button, Text: "After section"},
		}}
	}
	m, a := simpleManual(t, manualConfig(t, 7), view)
	unitSettle(t, m)
	w := m.implementation.find("one")
	r := findRecord(w.state.root, "details")
	heading, host := r.control.disclosure.header, r.control.disclosure.host
	field := findRecord(w.state.root, "field").widget.(*core.TextField)
	before, err := m.Capture(testContext, "one")
	if err != nil {
		t.Fatal(err)
	}
	fullHeight := r.widget.AsWidget().Geom.TotalBBox.Dy()
	afterY := findRecord(w.state.root, "after").widget.AsWidget().Geom.TotalBBox.Min.Y
	send := func(code key.Codes) {
		t.Helper()
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	send(key.CodeLeftArrow)
	if expanded || changes != 1 || field.IsVisible() || host.IsDisplayable() || r.widget.AsWidget().Geom.TotalBBox.Dy() >= fullHeight || findRecord(w.state.root, "after").widget.AsWidget().Geom.TotalBBox.Min.Y >= afterY {
		t.Fatal("collapse did not release visible layout")
	}
	collapsed, err := m.Capture(testContext, "one")
	if err != nil || bytes.Equal(before.Pix, collapsed.Pix) {
		t.Fatal("completed collapsed raster unchanged", err)
	}
	send(key.CodeTab)
	s, err := m.Snapshot(testContext)
	if err != nil || s.Windows[0].FocusedID != "after" {
		t.Fatal("Tab entered collapsed content", s, err)
	}
	// The published heading target resolves a physical hit, independent of focus.
	s, err = m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	target := ManualElement{}
	for _, e := range s.Windows[0].Elements {
		if e.ID == accessibilityDisclosureID("details") {
			target = e
		}
	}
	if target.ID == "" || target.Value != false || target.Bounds != heading.Geom.TotalBBox {
		t.Fatal("heading target/frame/state absent", target)
	}
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: typ, Target: target.ID, TargetIncarnation: target.Incarnation, Button: events.Left}); err != nil {
			t.Fatal(err)
		}
	}
	unitSettle(t, m)
	if !expanded || changes != 2 || !field.IsVisible() || r.control.disclosure.header != heading || findRecord(w.state.root, "field").widget != field {
		t.Fatal("pointer reopen lost children or body geometry")
	}
	send(key.CodeReturnEnter)
	if expanded || changes != 3 {
		t.Fatal("Enter did not toggle heading")
	}
	send(key.CodeRightArrow)
	if !expanded || changes != 4 {
		t.Fatal("Right did not expand")
	}
	send(key.CodeTab)
	s, err = m.Snapshot(testContext)
	if err != nil || s.Windows[0].FocusedID != "field" {
		t.Fatal("expanded Tab did not enter content", s, err)
	}
	if err := a.Do(func() { expanded = false; a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	s, err = m.Snapshot(testContext)
	if err != nil || s.Windows[0].FocusedID != "details" {
		t.Fatal("programmatic collapse retained hidden child focus", s, err)
	}
}

func TestDisclosureManagedGridTreeScopesRetainBindings(t *testing.T) {
	expanded := true
	g, _ := testNativeGrid(t, 50)
	o := nativeOutline(t, 50, true)
	bounded := ui.Style{Grow: ui.Some(float32(0)), MinSize: ui.Some(ui.Size{Width: 400, Height: 140}), MaxSize: ui.Some(ui.Size{Width: 400, Height: 140})}
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "section", Kind: ui.Disclosure, Text: "Data", Checked: expanded, Children: []ui.Node{
			{ID: "outline", Kind: ui.DataTree, Tree: o, Style: bounded},
			{ID: "grid", Kind: ui.DataGrid, Grid: g, Style: bounded},
		}}
	})
	unitSettle(t, m)
	root := m.implementation.find("one").state.root
	grid := findRecord(root, "grid").control.grid
	outline := findRecord(root, "outline").control.outline
	if !grid.scope || !outline.scope {
		t.Fatal("expanded data input scopes absent")
	}
	if err := a.Do(func() { expanded = false; a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if grid.scope || outline.scope {
		t.Fatal("collapsed data input scopes remain active")
	}
	nodes, _ := accessibilitySnapshot(root, nil)
	if len(nodes) != 2 {
		t.Fatal("collapsed data semantics remain reachable", len(nodes))
	}
	if err := a.Do(func() { expanded = true; a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if !grid.scope || !outline.scope || findRecord(root, "grid").control.grid != grid || findRecord(root, "outline").control.outline != outline {
		t.Fatal("restored data scopes lost binding identity")
	}
}
