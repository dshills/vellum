//go:build offscreen

package cogent

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"math"
	"testing"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/ui"
)

func sameSplitDraft(got, want core.TextFieldEditingState) bool {
	// Resizing changes the caret rectangle/visual layout, not draft or selection.
	return got.Text == want.Text && got.Cursor == want.Cursor && got.Anchor == want.Anchor && got.Focus == want.Focus
}

func TestSplitManualKeyboardPointerGeometryAndRetainedDraft(t *testing.T) {
	for _, orientation := range []ui.Orientation{ui.Horizontal, ui.Vertical} {
		t.Run(string(orientation), func(t *testing.T) {
			fraction := float32(0.5)
			calls := 0
			config := manualConfig(t, 7)
			view := func() ui.Node {
				n := splitFixture()
				n.Number, n.Orientation, n.Grow = fraction, orientation, true
				n.Change = func(e ui.ChangeEvent) { fraction = e.Number; calls++ }
				return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: "split", Sequence: 1}, Children: []ui.Node{n}}
			}
			m, a := simpleManual(t, config, view)
			unitSettle(t, m)
			w := m.implementation.find("one")
			r := findRecord(w.state.root, "split")
			c := r.control.split
			field := findRecord(w.state.root, "left").widget.(*core.TextField)
			if !field.ApplyNativeEdit("Retained draft", 14, 14, -1, -1, "insert") {
				t.Fatal("draft refused")
			}
			draft := field.EditingState()
			unitSettle(t, m) // The direct draft edit invalidates the explicit frame.
			before, err := m.Capture(testContext, "one")
			if err != nil {
				t.Fatal(err)
			}
			send := func(code key.Codes, modifiers key.Modifiers) {
				t.Helper()
				if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code, Modifiers: modifiers}); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
			}
			increase, decrease := key.CodeRightArrow, key.CodeLeftArrow
			axis := math32.X
			if orientation == ui.Vertical {
				increase, decrease, axis = key.CodeDownArrow, key.CodeUpArrow, math32.Y
			}
			send(increase, 0)
			if math.Abs(float64(fraction-0.51)) > 1e-6 || calls != 1 || findRecord(w.state.root, "left").widget != field || !sameSplitDraft(field.EditingState(), draft) {
				t.Fatal("keyboard resize lost accepted value/draft", fraction, calls, field.EditingState())
			}
			send(decrease, 0)
			send(key.CodeHome, 0)
			if fraction != 0.05 {
				t.Fatal("Home did not use lower bound", fraction)
			}
			send(key.CodeEnd, 0)
			if fraction != 0.95 {
				t.Fatal("End did not use upper bound", fraction)
			}
			fraction = 0.5
			if err := a.Invalidate("one"); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			if c.handle.Geom.TotalBBox.Empty() || c.panes[0].Geom.TotalBBox.Empty() || c.panes[1].Geom.TotalBBox.Empty() {
				t.Fatal("split pane/handle geometry missing")
			}
			snapshot, err := m.Snapshot(testContext)
			if err != nil {
				t.Fatal(err)
			}
			id := accessibilitySplitHandleID("split")
			var target ManualElement
			for _, e := range snapshot.Windows[0].Elements {
				if e.ID == id {
					target = e
				}
			}
			if target.ID != id || target.Bounds != c.handle.Geom.TotalBBox {
				t.Fatal("measured private divider target missing", target)
			}
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.MouseDown, Button: events.Left, Target: id, TargetIncarnation: target.Incarnation}); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			if _, err := config.Clock.Advance(core.TimingSettings.SlideStartTime + time.Nanosecond); err != nil {
				t.Fatal(err)
			}
			position := func(value float32) image.Point {
				p := c.handle.Geom.TotalBBox.Min.Add(c.handle.Geom.TotalBBox.Size().Div(2))
				coordinate := int(c.handle.Min + value*(c.handle.Max-c.handle.Min))
				origin := r.widget.AsWidget().Geom.ContentBBox.Min
				if axis == math32.X {
					p.X = origin.X + coordinate
				} else {
					p.Y = origin.Y + coordinate
				}
				return p
			}
			for _, value := range []float32{0.65, 0.75} {
				if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.MouseMove, Position: position(value)}); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
			}
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.MouseUp, Button: events.Left, Position: position(0.75)}); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			if math.Abs(float64(fraction-0.75)) > 0.015 || r.control.split != c || !sameSplitDraft(field.EditingState(), draft) {
				t.Fatal("actual pointer resizing failed accepted fraction/retention", fraction, calls)
			}
			first, second := c.panes[0].Geom.TotalBBox.Dx(), c.panes[1].Geom.TotalBBox.Dx()
			if axis == math32.Y {
				first, second = c.panes[0].Geom.TotalBBox.Dy(), c.panes[1].Geom.TotalBBox.Dy()
			}
			if first+second == 0 || math.Abs(float64(first)/float64(first+second)-0.75) > 0.015 || !c.handle.Geom.TotalBBox.In(r.widget.AsWidget().Geom.TotalBBox) {
				t.Fatal("completed pane geometry does not match accepted resizing", first, second, c.handle.Geom.TotalBBox)
			}
			after, err := m.Capture(testContext, "one")
			if err != nil || bytes.Equal(before.Pix, after.Pix) {
				t.Fatal("accepted split resize produced no completed pixels", err)
			}
			bounds := accessibilityBounds(w.state.root)[id]
			if bounds[2] <= 0 || bounds[3] <= 0 {
				t.Fatal("split accessible divider bounds missing", bounds)
			}
			send(key.CodeTab, 1<<key.Shift)
			snapshot, err = m.Snapshot(testContext)
			if err != nil || snapshot.Windows[0].FocusedID != "left" {
				t.Fatal("reverse divider traversal did not reach first pane", snapshot, err)
			}
			send(key.CodeTab, 0)
			send(key.CodeTab, 0)
			snapshot, err = m.Snapshot(testContext)
			if err != nil || snapshot.Windows[0].FocusedID != "right" {
				t.Fatal("divider traversal did not reach second pane", snapshot, err)
			}
		})
	}
}

func TestSplitManualOrientationChangeKeepsHandleAndChildren(t *testing.T) {
	orientation := ui.Horizontal
	m, a := simpleManual(t, manualConfig(t, 7), func() ui.Node {
		n := splitFixture()
		n.Orientation, n.Grow = orientation, true
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{n}}
	})
	unitSettle(t, m)
	w := m.implementation.find("one")
	r := findRecord(w.state.root, "split")
	c, child := r.control.split, r.children[0].widget
	orientation = ui.Vertical
	if err := a.Invalidate("one"); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if r.control.split != c || r.children[0].widget != child || c.handle != r.control.split.native.Parts.Child(0) || c.panes[0].Geom.TotalBBox.Min.Y >= c.panes[1].Geom.TotalBBox.Min.Y {
		t.Fatal("orientation update lost native identity or top/bottom layout")
	}
	nodes, _ := accessibilitySnapshot(w.state.root, nil)
	found := false
	for _, n := range nodes {
		if n.Role == "split-handle" {
			found = true
			if n.Orientation != "horizontal" {
				t.Fatal("divider semantic orientation stale", n)
			}
		}
	}
	if !found {
		t.Fatal("divider missing from updated semantics")
	}
}

func TestSplitManualSmallPositiveFractionDoesNotCollapsePane(t *testing.T) {
	for _, fraction := range []float32{0.005, 0.995, 0.0000005, 0.9999995} {
		t.Run(fmt.Sprint(fraction), func(t *testing.T) {
			m, _ := simpleManual(t, manualConfig(t, 7), func() ui.Node {
				return ui.Node{ID: "split", Kind: ui.SplitPane, Number: fraction, Grow: true, Numeric: ui.Numeric{Min: 0.0000001, Max: 0.9999999, Step: 0.0000001}, Children: []ui.Node{{ID: "left", Kind: ui.Label, Text: "Left"}, {ID: "right", Kind: ui.Label, Text: "Right"}}}
			})
			unitSettle(t, m)
			w := m.implementation.find("one")
			r := w.state.root
			for _, pane := range r.control.split.panes {
				if !pane.IsDisplayable() || pane.Geom.TotalBBox.Empty() {
					t.Fatal("positive accepted pane was implicitly collapsed", fraction, pane.Geom.TotalBBox)
				}
			}
			if r.control.split.native.Split(0) != fraction {
				t.Fatal("pixel layout changed accepted fraction")
			}
		})
	}
}

func TestSplitManualRegisteredDigitCommandKeepsPriority(t *testing.T) {
	for _, focus := range []string{"left", "split"} {
		t.Run(focus, func(t *testing.T) {
			calls := 0
			registry := command.New()
			if err := registry.Register(command.Command{ID: "digit", Label: "Digit", Shortcut: "Meta+1", Execute: func(context.Context) error { calls++; return nil }}); err != nil {
				t.Fatal(err)
			}
			config := manualConfig(t, 7)
			config.Commands = registry
			m, _ := simpleManual(t, config, func() ui.Node {
				n := splitFixture()
				return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: focus, Sequence: 1}, Children: []ui.Node{n}}
			})
			unitSettle(t, m)
			w := m.implementation.find("one")
			before, err := m.Snapshot(testContext)
			if err != nil || before.Windows[0].FocusedID != focus {
				t.Fatal("command fixture did not focus requested target", focus, before, err)
			}
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Rune: '1', Code: key.Code1, Modifiers: 1 << key.Meta}); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			if calls != 1 || findRecord(w.state.root, "split").control.split.native.Split(0) != 0.5 {
				t.Fatal("native digit guard swallowed application command", focus, calls)
			}
		})
	}
}

func TestSplitManualReadOnlyAndDisabledAncestor(t *testing.T) {
	for _, mode := range []string{"read-only", "disabled-ancestor"} {
		t.Run(mode, func(t *testing.T) {
			disabled, calls := false, 0
			m, a := simpleManual(t, manualConfig(t, 7), func() ui.Node {
				n := splitFixture()
				if mode != "read-only" {
					n.Change = func(ui.ChangeEvent) { calls++ }
				}
				return ui.Node{ID: "root", Kind: ui.Column, Disabled: disabled, Grow: true, Focus: ui.FocusRequest{ID: "split", Sequence: 1}, Children: []ui.Node{n}}
			})
			unitSettle(t, m)
			w := m.implementation.find("one")
			before, err := m.Snapshot(testContext)
			if err != nil || before.Windows[0].FocusedID != "split" {
				t.Fatal("divider initial focus missing", before, err)
			}
			if mode == "disabled-ancestor" {
				disabled = true
				if err := a.Invalidate("one"); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
			}
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: key.CodeRightArrow}); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			r := findRecord(w.state.root, "split")
			if calls != 0 || r.control.split.native.Split(0) != 0.5 {
				t.Fatal("non-writable divider changed accepted fraction", calls)
			}
		})
	}
}
