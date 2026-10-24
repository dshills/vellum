//go:build offscreen

package cogent

import (
	"bytes"
	"fmt"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/ui"
)

func TestChromeManualToolbarOverflowAndStatusUpdates(t *testing.T) {
	status := "Ready"
	actions := 0
	view := func() ui.Node {
		tools := []ui.Node{}
		for i := range 8 {
			i := i
			tools = append(tools, ui.Node{ID: fmt.Sprintf("tool-%d", i), Kind: ui.Button, Text: fmt.Sprintf("Tool %d", i), MinWidth: 100, Disabled: i == 1, Activate: func() { status = fmt.Sprintf("Tool %d activated", i); actions++ }})
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: "tool-0", Sequence: 1}, Children: []ui.Node{
			{ID: "tools", Kind: ui.Toolbar, Accessibility: ui.Accessibility{Name: "Document tools"}, Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 260})}, Children: tools},
			{ID: "form", Kind: ui.FormGroup, Text: "<Project>", Children: []ui.Node{{ID: "field", Kind: ui.TextField, Text: "Name"}}},
			{ID: "status", Kind: ui.StatusBar, Text: status, Accessibility: ui.Accessibility{Name: "Application status"}, Children: []ui.Node{{ID: "reset", Kind: ui.Button, Text: "Reset", Activate: func() { status = "Ready" }}}},
		}}
	}
	m, _ := simpleManual(t, manualConfig(t, 7), view)
	unitSettle(t, m)
	w := m.implementation.find("one")
	toolbar := findRecord(w.state.root, "tools").control.chrome.host
	if toolbar.Scrolls[math32.X] == nil {
		t.Fatal("narrow toolbar has no overflow scrolling")
	}
	before, err := m.Capture(testContext, "one")
	if err != nil {
		t.Fatal(err)
	}
	send := func(code key.Codes) {
		t.Helper()
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	focus := func() string {
		t.Helper()
		s, err := m.Snapshot(testContext)
		if err != nil {
			t.Fatal(err)
		}
		return s.Windows[0].FocusedID
	}
	send(key.CodeEnd)
	if focus() != "tool-7" || toolbar.Scrolls[math32.X].Value <= 0 {
		t.Fatal("End failed to reveal overflowing last tool", focus(), toolbar.Scrolls[math32.X].Value)
	}
	box := findRecord(w.state.root, "tool-7").widget.AsWidget().Geom.TotalBBox
	if box.Empty() || !box.In(toolbar.Geom.TotalBBox) {
		t.Fatal("focused tool clipped outside toolbar", box, toolbar.Geom.TotalBBox)
	}
	send(key.CodeReturnEnter)
	if status != "Tool 7 activated" || actions != 1 {
		t.Fatal("overflowing focused tool activation failed", status, actions)
	}
	statusRecord := findRecord(w.state.root, "status")
	if statusRecord.control.chrome.caption.Text != status {
		t.Fatal("status caption stale")
	}
	nodes, _ := accessibilitySnapshot(w.state.root, nil)
	found := false
	for _, n := range nodes {
		if n.DomainID == "status" {
			found = true
			if n.Value != status || n.Focusable || len(n.Actions) != 0 {
				t.Fatal("status value/action metadata", n)
			}
		}
	}
	if !found {
		t.Fatal("status semantics missing")
	}
	after, err := m.Capture(testContext, "one")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(before.Pix, after.Pix) {
		t.Fatal("toolbar/status produced no completed raster change")
	}
	send(key.CodeHome)
	if focus() != "tool-0" || toolbar.Scrolls[math32.X].Value != 0 {
		t.Fatal("Home failed to reveal first tool", focus(), toolbar.Scrolls[math32.X].Value)
	}
	send(key.CodeRightArrow)
	if focus() != "tool-2" {
		t.Fatal("toolbar arrows did not skip disabled tool", focus())
	}
	send(key.CodeEnd)
	send(key.CodeTab)
	if focus() != "field" {
		t.Fatal("toolbar/form added private managed focus stops", focus())
	}
	form := findRecord(w.state.root, "form")
	if form.control.chrome.caption.Text != "&lt;Project&gt;" {
		t.Fatal("form title entered markup")
	}
}

func TestChromeManualVerticalToolbarScrollAndDisabledNavigation(t *testing.T) {
	view := func() ui.Node {
		tools := []ui.Node{}
		for i := range 7 {
			tools = append(tools, ui.Node{ID: fmt.Sprintf("tool-%d", i), Kind: ui.Button, Text: fmt.Sprintf("Tool %d", i), Disabled: i == 1})
		}
		return ui.Node{ID: "root", Kind: ui.Row, Grow: true, Focus: ui.FocusRequest{ID: "tool-0", Sequence: 1}, Children: []ui.Node{{ID: "tools", Kind: ui.Toolbar, Orientation: ui.Vertical, Style: ui.Style{MaxSize: ui.Some(ui.Size{Height: 120})}, Children: tools}}}
	}
	m, _ := simpleManual(t, manualConfig(t, 7), view)
	unitSettle(t, m)
	w := m.implementation.find("one")
	toolbar := findRecord(w.state.root, "tools").control.chrome.host
	if toolbar.Scrolls[math32.Y] == nil {
		t.Fatal("vertical toolbar has no main-axis scrollbar")
	}
	send := func(code key.Codes) {
		t.Helper()
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	focus := func() string {
		t.Helper()
		s, err := m.Snapshot(testContext)
		if err != nil {
			t.Fatal(err)
		}
		return s.Windows[0].FocusedID
	}
	send(key.CodeEnd)
	box := findRecord(w.state.root, "tool-6").widget.AsWidget().Geom.TotalBBox
	if focus() != "tool-6" || toolbar.Scrolls[math32.Y].Value <= 0 || box.Empty() || !box.In(toolbar.Geom.TotalBBox) {
		t.Fatal("vertical End failed to reveal tool", focus(), box, toolbar.Geom.TotalBBox)
	}
	send(key.CodeHome)
	send(key.CodeDownArrow)
	if focus() != "tool-2" {
		t.Fatal("vertical toolbar did not skip disabled tool", focus())
	}
	send(key.CodeUpArrow)
	if focus() != "tool-0" {
		t.Fatal("vertical previous enabled tool wrong", focus())
	}
	send(key.CodeUpArrow)
	if focus() != "tool-6" {
		t.Fatal("vertical toolbar did not wrap enabled tools", focus())
	}
}

func TestChromeManualToolbarLeavesTextEditingHomeEndToField(t *testing.T) {
	m, _ := simpleManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "field", Sequence: 1}, Children: []ui.Node{{ID: "tools", Kind: ui.Toolbar, Children: []ui.Node{{ID: "first", Kind: ui.Button, Text: "First"}, {ID: "field", Kind: ui.TextField, Text: "alpha"}, {ID: "last", Kind: ui.Button, Text: "Last"}}}}}
	})
	unitSettle(t, m)
	w := m.implementation.find("one")
	field := findRecord(w.state.root, "field").widget.(*core.TextField)
	for _, test := range []struct {
		code key.Codes
		want int
	}{{key.CodeHome, 0}, {key.CodeEnd, 5}} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: test.code}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
		snapshot, err := m.Snapshot(testContext)
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Windows[0].FocusedID != "field" || field.EditingState().Focus != test.want {
			t.Fatal("toolbar stole text editing chord", snapshot.Windows[0].FocusedID, field.EditingState())
		}
	}
}

func TestChromeManualNestedToolbarNavigationSkipsDisabled(t *testing.T) {
	m, _ := simpleManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "a", Sequence: 1}, Children: []ui.Node{{ID: "tools", Kind: ui.Toolbar, Children: []ui.Node{{ID: "nested", Kind: ui.Row, Children: []ui.Node{{ID: "a", Kind: ui.Button, Text: "A"}, {ID: "b", Kind: ui.Button, Text: "B", Disabled: true}, {ID: "c", Kind: ui.Button, Text: "C"}}}, {ID: "d", Kind: ui.Button, Text: "D"}}}}}
	})
	unitSettle(t, m)
	w := m.implementation.find("one")
	for _, want := range []string{"c", "d", "a"} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: key.CodeRightArrow}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
		s, err := m.Snapshot(testContext)
		if err != nil {
			t.Fatal(err)
		}
		if s.Windows[0].FocusedID != want {
			t.Fatal("nested layout bypassed enabled toolbar order", s.Windows[0].FocusedID, want)
		}
	}
}
func TestChromeManualCaptionAppearsAndReappears(t *testing.T) {
	for _, kind := range []ui.Kind{ui.StatusBar, ui.FormGroup} {
		t.Run(string(kind), func(t *testing.T) {
			text := ""
			m, a := simpleManual(t, manualConfig(t, 7), func() ui.Node {
				return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "group", Kind: kind, Text: text, Children: []ui.Node{{ID: "child", Kind: ui.Button, Text: "Child"}}}}}
			})
			unitSettle(t, m)
			w := m.implementation.find("one")
			caption := findRecord(w.state.root, "group").control.chrome.caption
			blank, err := m.Capture(testContext, "one")
			if err != nil {
				t.Fatal(err)
			}
			for _, value := range []string{"First caption", "", "Restored caption"} {
				text = value
				if err := a.Invalidate("one"); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
				if value != "" && (caption.Geom.TotalBBox.Empty() || !caption.IsDisplayable()) {
					t.Fatal("caption stayed hidden after nonempty text", value, caption.Geom.TotalBBox)
				}
				if value == "" && caption.IsDisplayable() {
					t.Fatal("empty caption visible")
				}
			}
			after, err := m.Capture(testContext, "one")
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Equal(blank.Pix, after.Pix) {
				t.Fatal("restored caption produced no completed pixels")
			}
		})
	}
}

func TestChromeManualNearestToolbarScopeAndReparentClearing(t *testing.T) {
	moved := false
	innerRow := ui.Node{ID: "row", Kind: ui.Row, Children: []ui.Node{{ID: "a", Kind: ui.Button, Text: "A"}, {ID: "b", Kind: ui.Button, Text: "B"}}}
	view := func() ui.Node {
		children := []ui.Node{innerRow, {ID: "stay", Kind: ui.Button, Text: "Stay"}}
		outside := []ui.Node{}
		if moved {
			children = children[1:]
			outside = append(outside, innerRow)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "a", Sequence: 1}, Children: []ui.Node{{ID: "tools", Kind: ui.Toolbar, Children: children}, {ID: "outside", Kind: ui.Column, Children: outside}}}
	}
	m, a := simpleManual(t, manualConfig(t, 7), view)
	unitSettle(t, m)
	w := m.implementation.find("one")
	row := findRecord(w.state.root, "row")
	if row.widget.AsTree().Property(toolbarScopeProperty) != findRecord(w.state.root, "tools") {
		t.Fatal("nested row missing toolbar scope")
	}
	moved = true
	if err := a.Invalidate("one"); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if findRecord(w.state.root, "row") != row || row.widget.AsTree().Property(toolbarScopeProperty) != nil {
		t.Fatal("reparent retained obsolete toolbar scope")
	}
	findRecord(w.state.root, "a").widget.AsWidget().SetFocus()
	event := events.NewKey(events.KeyChord, 0, key.CodeHome, 0)
	row.widget.AsWidget().HandleEvent(event)
	if event.IsHandled() || !findRecord(w.state.root, "a").widget.AsWidget().ContainsFocus() {
		t.Fatal("moved-out row retained toolbar Home behavior")
	}
}
func TestChromeManualNestedToolbarKeepsNearestScope(t *testing.T) {
	m, _ := simpleManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "inner-a", Sequence: 1}, Children: []ui.Node{{ID: "outer", Kind: ui.Toolbar, Children: []ui.Node{{ID: "before", Kind: ui.Button, Text: "Before"}, {ID: "inner", Kind: ui.Toolbar, Children: []ui.Node{{ID: "inner-a", Kind: ui.Button, Text: "A"}, {ID: "inner-b", Kind: ui.Button, Text: "B"}}}, {ID: "after", Kind: ui.Button, Text: "After"}}}}}
	})
	unitSettle(t, m)
	w := m.implementation.find("one")
	for _, want := range []string{"inner-b", "inner-a"} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: key.CodeRightArrow}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
		s, err := m.Snapshot(testContext)
		if err != nil {
			t.Fatal(err)
		}
		if s.Windows[0].FocusedID != want {
			t.Fatal("nested toolbar escaped nearest scope", s.Windows[0].FocusedID, want)
		}
	}
}

func TestChromeManualAllDisabledToolsDoNotNavigateToDisabledSibling(t *testing.T) {
	for _, mode := range []string{"children", "toolbar", "ancestor"} {
		t.Run(mode, func(t *testing.T) {
			disabled := false
			m, a := simpleManual(t, manualConfig(t, 7), func() ui.Node {
				return ui.Node{ID: "root", Kind: ui.Column, Disabled: disabled && mode == "ancestor", Focus: ui.FocusRequest{ID: "a", Sequence: 1}, Children: []ui.Node{{ID: "tools", Kind: ui.Toolbar, Disabled: disabled && mode == "toolbar", Children: []ui.Node{{ID: "a", Kind: ui.Button, Text: "A", Disabled: disabled && mode == "children"}, {ID: "b", Kind: ui.Button, Text: "B", Disabled: disabled && mode == "children"}}}}}
			})
			unitSettle(t, m)
			w := m.implementation.find("one")
			disabled = true
			if err := a.Invalidate("one"); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: key.CodeRightArrow}); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			if findRecord(w.state.root, "b").widget.AsWidget().ContainsFocus() {
				t.Fatal("all-disabled toolbar fell through to disabled sibling")
			}
		})
	}
}

func TestChromeManualOversizedVerticalFocusVisible(t *testing.T) {
	calls := 0
	view := func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "small", Sequence: 1}, Children: []ui.Node{
			{ID: "tools", Kind: ui.Toolbar, Orientation: ui.Vertical, Style: ui.Style{MaxSize: ui.Some(ui.Size{Height: 100})}, Children: []ui.Node{
				{ID: "small", Kind: ui.Button, Text: "First", Activate: func() {}},
				{ID: "middle1", Kind: ui.Button, Text: "Second", Activate: func() {}},
				{ID: "middle2", Kind: ui.Button, Text: "Third", Activate: func() {}},
				{ID: "tall", Kind: ui.Button, Text: "Large action", Style: ui.Style{MinSize: ui.Some(ui.Size{Height: 180})}, Activate: func() { calls++ }},
			}},
		}}
	}
	m, _ := simpleManual(t, manualConfig(t, 959), view)
	unitSettle(t, m)
	w, _ := imageManualElement(t, m, "small")
	send := func(code key.Codes) {
		t.Helper()
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: code}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	send(key.CodeEnd)
	_, large := imageManualElement(t, m, "tall")
	host := findRecord(m.implementation.find("one").state.root, "tools").control.chrome.host
	r := findRecord(m.implementation.find("one").state.root, "tall").widget.AsWidget()
	raw := math32.RectFromPosSizeMax(r.Geom.Pos.Total, r.Geom.Size.Actual.Total)
	if !large.Focused || large.Bounds.Empty() || raw.Dy() <= host.Geom.ContentBBox.Dy() || raw.Min.Y < host.Geom.ContentBBox.Min.Y || raw.Min.Y >= host.Geom.ContentBBox.Max.Y {
		t.Fatal("oversized vertical target invisible", large, raw, host.Geom.ContentBBox)
	}
	send(key.CodeReturnEnter)
	if calls != 1 {
		t.Fatal("focused oversized action did not run exactly once", calls)
	}
	send(key.CodeHome)
	_, small := imageManualElement(t, m, "small")
	if !small.Focused || small.Bounds.Empty() {
		t.Fatal("Home lost ordinary target", small)
	}
}
