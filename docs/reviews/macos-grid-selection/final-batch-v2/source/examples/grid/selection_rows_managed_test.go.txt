//go:build offscreen

package main

import (
	"errors"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/virtual"
)

func rowSelectionFixture(t *testing.T, batch int, variable bool) (*gridProgram, *cogent.Manual) {
	t.Helper()
	cfg := manualConfig(t, batch, 120)
	p, err := newProgramModes(1280, 950, variable, true)
	if err != nil {
		t.Fatal(err)
	}
	m, err := cogent.NewManual(background, p.owner, cfg)
	if err != nil {
		p.owner.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := m.Close(background); err != nil {
			t.Error(err)
		}
	})
	settle(t, m, batch)
	return p, m
}

func rowModifiers(flags ...key.Modifiers) key.Modifiers {
	var mods key.Modifiers
	for _, flag := range flags {
		mods.SetFlag(true, flag)
	}
	return mods
}

func rowClick(t *testing.T, m *cogent.Manual, batch int, windowID, target string, modifiers key.Modifiers) {
	t.Helper()
	w := window(t, settle(t, m, batch), windowID)
	e := element(t, w, target)
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		mods := modifiers
		if typ == events.MouseUp {
			mods = 0
		} // Intent must survive modifier release.
		if err := m.Input(background, cogent.ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Target: e.ID, TargetIncarnation: e.Incarnation, Type: typ, Button: events.Left, Modifiers: mods}); err != nil {
			t.Fatal(err)
		}
	}
	settle(t, m, batch)
}

func rowKeyInput(t *testing.T, m *cogent.Manual, batch int, windowID string, code key.Codes, mods key.Modifiers) {
	t.Helper()
	w := window(t, settle(t, m, batch), windowID)
	if err := m.Input(background, cogent.ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: code, Modifiers: mods}); err != nil {
		t.Fatal(err)
	}
	settle(t, m, batch)
}

func rowTabFocus(t *testing.T, m *cogent.Manual, batch int, windowID string) {
	t.Helper()
	for range 24 {
		w := window(t, settle(t, m, batch), windowID)
		if w.FocusedID == "data" {
			return
		}
		primitiveKey(t, m, batch, windowID, key.CodeTab)
	}
	t.Fatal("actual Tab route did not focus grid")
}

func rowMeasuredSelection(t *testing.T, p *gridProgram, w *gridWindow) virtual.GridSelection {
	t.Helper()
	var selected virtual.GridSelection
	if err := owned(p.owner, func() error { var err error; selected, err = w.grid.Selection(); return err }); err != nil {
		t.Fatal(err)
	}
	g := inspection(t, p, w)
	if !g.RowSelection || g.Selection != selected.State() {
		t.Fatal("measured selection state disagrees", g.Selection, selected.State())
	}
	for _, cell := range g.Cells {
		want, err := selected.Contains(cell.RowKey)
		if err != nil || cell.Selected != want {
			t.Fatal("measured row selection flag disagrees", cell, want, err)
		}
	}
	return selected
}

func TestActualRowSelectionModifiersKeysAndCapturedUniverse(t *testing.T) {
	for _, batch := range []int{1, 7, 64} {
		t.Run(stringBatch(batch), func(t *testing.T) {
			p, m := rowSelectionFixture(t, batch, false)
			w := p.windows[1]
			cell := func(row int) string { return cogent.GridCellID("data", rowKey(row), columnID(0)) }
			rowClick(t, m, batch, w.id, cell(2), 0)
			if s := state(t, w); s.Selection.CapturedRows != 1 || s.Selection.AnchorKey != rowKey(2) {
				t.Fatal("plain click did not replace", s)
			}
			rowClick(t, m, batch, w.id, cell(5), rowModifiers(key.Shift))
			if s := state(t, w); s.Selection.CapturedRows != 4 || s.Active.RowKey != rowKey(5) {
				t.Fatal("Shift press did not extend across captured rows", s)
			}
			rowClick(t, m, batch, w.id, cell(3), rowModifiers(key.Meta))
			snapshot := rowMeasuredSelection(t, p, w)
			if snapshot.State().CapturedRows != 3 || snapshot.State().Overrides != 1 {
				t.Fatal("Command press did not toggle", snapshot.State())
			}
			if chosen, e := snapshot.Contains(rowKey(3)); e != nil || chosen {
				t.Fatal("active toggled row stayed selected", chosen, e)
			}
			if state(t, w).Active.RowKey != rowKey(3) {
				t.Fatal("active cursor conflated with selected state")
			}
			rowKeyInput(t, m, batch, w.id, key.CodeDownArrow, rowModifiers(key.Shift))
			if s := state(t, w); s.Active.RowKey != rowKey(4) || s.Selection.CapturedRows != 3 || s.Selection.Overrides != 0 {
				t.Fatal("Shift key did not capture anchor-to-target range", s)
			}
			primitiveKey(t, m, batch, w.id, key.CodeDownArrow)
			if s := state(t, w); s.Active.RowKey != rowKey(5) || s.Selection.CapturedRows != 1 {
				t.Fatal("plain key did not replace row selection", s)
			}
			rowKeyInput(t, m, batch, w.id, key.CodeA, rowModifiers(key.Meta))
			if s := state(t, w); s.Selection.CapturedRows != 1_000_000 || s.Selection.Scope.Rows != 1_000_000 {
				t.Fatal("Command+A did not capture All", s)
			}
			rowClick(t, m, batch, w.id, cell(1), rowModifiers(key.Meta))
			rowClick(t, m, batch, w.id, cell(2), rowModifiers(key.Meta))
			captured := rowMeasuredSelection(t, p, w)
			if captured.State().CapturedRows != 999_998 {
				t.Fatal("All minus two", captured.State())
			}
			click(t, m, batch, w.id, "reverse")
			click(t, m, batch, w.id, "filter")
			current := rowMeasuredSelection(t, p, w)
			if current.State() != captured.State() || state(t, w).Stamp.Rows != 500_000 {
				t.Fatal("actual filter reinterpreted captured All universe", current.State(), captured.State())
			}
			if selected, e := current.Contains(rowKey(3)); e != nil || !selected {
				t.Fatal("filtered-out captured key lost", selected, e)
			}
			rowTabFocus(t, m, batch, w.id)
			primitiveKey(t, m, batch, w.id, key.CodeEscape)
			if rowMeasuredSelection(t, p, w).State().CapturedRows != 0 || state(t, p.windows[0]).Selection.CapturedRows != 0 {
				t.Fatal("Escape clear/independent window failed")
			}
			if selected, e := captured.Contains(rowKey(999_999)); e != nil || !selected {
				t.Fatal("Clear invalidated detached snapshot", selected, e)
			}
		})
	}
}

func stringBatch(batch int) string { return map[int]string{1: "1", 7: "7", 64: "64"}[batch] }

func TestActualRowSelectionVariablePageControlsAndStalePress(t *testing.T) {
	p, m := rowSelectionFixture(t, 7, true)
	w := p.windows[1]
	rowClick(t, m, 7, w.id, cogent.GridCellID("data", rowKey(2), columnID(0)), 0)
	before := state(t, w)
	want := variablePageDestination(t, w, before)
	rowKeyInput(t, m, 7, w.id, key.CodePageDown, rowModifiers(key.Shift))
	after := state(t, w)
	last, found := w.source.rowIndex(want)
	if !found || after.Active.RowKey != want || after.Selection.CapturedRows != last-2+1 || !after.VariableHeights {
		t.Fatal("actual variable Shift Page did not extend by pixel destination", before, after, want)
	}
	variableGeometry(t, w, inspection(t, p, w))
	click(t, m, 7, w.id, "select-all")
	captured := rowMeasuredSelection(t, p, w)
	click(t, m, 7, w.id, "row-heights")
	if after := state(t, w); after.VariableHeights || after.Selection != captured.State() {
		t.Fatal("height revision changed captured identities", after)
	}
	click(t, m, 7, w.id, "clear-selection")
	if rowMeasuredSelection(t, p, w).State().CapturedRows != 0 {
		t.Fatal("actual Clear button failed")
	}
	click(t, m, 7, w.id, "first")
	g := inspection(t, p, w)
	cell := visible(t, g, false)
	s := window(t, settle(t, m, 7), w.id)
	e := element(t, s, cell.ID)
	down := cogent.ManualInput{Window: w.id, WindowIncarnation: s.Incarnation, Target: e.ID, TargetIncarnation: e.Incarnation, Type: events.MouseDown, Button: events.Left, Modifiers: rowModifiers(key.Shift)}
	if err := m.Input(background, down); err != nil {
		t.Fatal(err)
	}
	if r, err := m.Step(background, 1); err != nil || r.Events != 1 {
		t.Fatal("row selection press not routed", r, err)
	}
	beforeSelection := state(t, w).Selection
	if err := owned(p.owner, func() error { err := w.projection(true, false); w.report(err); return err }); err != nil {
		t.Fatal(err)
	}
	settle(t, m, 7)
	up := down
	up.Type = events.MouseUp
	if err := m.Input(background, up); !errors.Is(err, cogent.ErrManualStale) {
		t.Fatal("retired row-selection target admitted", err)
	}
	up.Target = ""
	up.Position = e.Bounds.Min.Add(e.Bounds.Size().Div(2))
	if err := m.Input(background, up); err != nil {
		t.Fatal(err)
	}
	settle(t, m, 7)
	if state(t, w).Selection != beforeSelection {
		t.Fatal("stale physical release changed selection")
	}
}
