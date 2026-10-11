//go:build offscreen

package main

import (
	"context"
	"errors"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/virtual"
)

func editingFixture(t *testing.T, batch int, variable bool) (*gridProgram, *cogent.Manual) {
	t.Helper()
	cfg := manualConfig(t, batch, 120)
	p, err := newProgramOptions(1280, 1100, variable, true, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range p.windows {
		w.settleCell = func(ctx context.Context, callback func(error)) error {
			return cogent.SettleGridEdit(ctx, p.owner, w.id, w.grid, callback)
		}
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

func editingRune(t *testing.T, m *cogent.Manual, batch int, windowID string, value rune) {
	t.Helper()
	w := window(t, settle(t, m, batch), windowID)
	if err := m.Input(background, cogent.ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: events.KeyChord, Rune: value}); err != nil {
		t.Fatal(err)
	}
	settle(t, m, batch)
}

func editingText(t *testing.T, m *cogent.Manual, batch int, windowID, value string) {
	t.Helper()
	rowKeyInput(t, m, batch, windowID, key.CodeA, rowModifiers(key.Meta))
	for _, r := range value {
		editingRune(t, m, batch, windowID, r)
	}
}

func measuredEditor(t *testing.T, p *gridProgram, w *gridWindow) cogent.GridEditorGeometry {
	t.Helper()
	g := inspection(t, p, w)
	edit := state(t, w).Edit
	if g.Edit != edit || g.Editor == nil || g.Editor.ID != cogent.GridEditorID("data") || g.Editor.Session != edit.Session || g.Editor.Text != edit.Text || g.Editor.Bounds.Empty() || !g.Editor.Bounds.In(g.Bounds) {
		t.Fatal("measured inline editor differs from accepted draft", g.Edit, g.Editor, edit)
	}
	return *g.Editor
}

func acceptedCell(t *testing.T, w *gridWindow, row, column int, text string, version uint64) {
	t.Helper()
	if err := owned(w.owner, func() error {
		v, err := w.source.cellValue(rowKey(row), columnID(column))
		if err != nil || v.text != text || v.version != version {
			t.Fatalf("accepted cell %s/%s = %+v, want %q v%d: %v", rowKey(row), columnID(column), v, text, version, err)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestActualCellEditingCommitCancelAndOriginatingSave(t *testing.T) {
	for _, batch := range []int{1, 7, 64} {
		t.Run(stringBatch(batch), func(t *testing.T) {
			p, m := editingFixture(t, batch, false)
			w := p.windows[1]
			click(t, m, batch, w.id, cogent.GridCellID("data", rowKey(2), columnID(1)))
			primitiveKey(t, m, batch, w.id, key.CodeF2)
			firstEditor := measuredEditor(t, p, w)
			if !firstEditor.Focused {
				t.Fatal("F2 did not route focus to real inline field")
			}
			editingText(t, m, batch, w.id, "Routed 😀")
			primitiveKey(t, m, batch, w.id, key.CodeReturnEnter)
			acceptedCell(t, w, 2, 1, "Routed 😀", 2)
			if state(t, w).Edit.Active || inspection(t, p, w).Editor != nil {
				t.Fatal("Return did not accept and retire editor")
			}
			click(t, m, batch, w.id, cogent.GridCellID("data", rowKey(2), columnID(1)))
			primitiveKey(t, m, batch, w.id, key.CodeReturnEnter)
			secondEditor := measuredEditor(t, p, w)
			if secondEditor.Session <= firstEditor.Session {
				t.Fatal("cancel/rebegin reused private editor session")
			}
			editingText(t, m, batch, w.id, "cancel")
			primitiveKey(t, m, batch, w.id, key.CodeEscape)
			acceptedCell(t, w, 2, 1, "Routed 😀", 2)
			if state(t, w).Edit.Active {
				t.Fatal("Escape did not cancel real field draft")
			}
			click(t, m, batch, w.id, "edit-cell")
			measuredEditor(t, p, w)
			editingText(t, m, batch, w.id, "earlier routed value")
			// Admit one more real key before the Save button, without a settle or
			// owner snapshot between them. Coordinates come from the last frame.
			snapshot := window(t, settle(t, m, batch), w.id)
			save := element(t, snapshot, "save-cell")
			if err := m.Input(background, cogent.ManualInput{Window: w.id, WindowIncarnation: snapshot.Incarnation, Type: events.KeyChord, Rune: '!'}); err != nil {
				t.Fatal(err)
			}
			for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
				if err := m.Input(background, cogent.ManualInput{Window: w.id, WindowIncarnation: snapshot.Incarnation, Type: typ, Button: events.Left, Position: save.Bounds.Min.Add(save.Bounds.Size().Div(2))}); err != nil {
					t.Fatal(err)
				}
			}
			settle(t, m, batch)
			acceptedCell(t, w, 2, 1, "earlier routed value!", 3)
			if state(t, w).Edit.Active || w.cellCommitPending {
				t.Fatal("toolbar Save did not complete originating settlement")
			}
			click(t, m, batch, w.id, "edit-cell")
			editingText(t, m, batch, w.id, " ")
			primitiveKey(t, m, batch, w.id, key.CodeReturnEnter)
			if edit := state(t, w).Edit; !edit.Active || edit.Text != " " {
				t.Fatal("validation did not preserve actual typed draft", edit)
			}
			acceptedCell(t, w, 2, 1, "earlier routed value!", 3)
			click(t, m, batch, w.id, "cancel-cell")
			if state(t, p.windows[0]).Edit.Active || len(p.windows[0].source.overrides) != 0 {
				t.Fatal("independent window domain changed")
			}
		})
	}
}

func TestActualCellEditingConflictHiddenDraftGeometryAndModal(t *testing.T) {
	p, m := editingFixture(t, 7, true)
	w := p.windows[1]
	click(t, m, 7, w.id, cogent.GridCellID("data", rowKey(3), columnID(1)))
	primitiveKey(t, m, 7, w.id, key.CodeF2)
	editingText(t, m, 7, w.id, "recoverable")
	initial := measuredEditor(t, p, w)
	oldWindow := window(t, settle(t, m, 7), w.id)
	oldTarget := element(t, oldWindow, initial.ID)
	click(t, m, 7, w.id, "reverse")
	click(t, m, 7, w.id, "filter")
	if edit := state(t, w).Edit; edit.Status != virtual.GridEditUnavailable || edit.Text != "recoverable" || edit.Session != initial.Session {
		t.Fatal("actual filter lost unavailable draft", edit)
	}
	if err := m.Input(background, cogent.ManualInput{Window: w.id, WindowIncarnation: oldWindow.Incarnation, Target: oldTarget.ID, TargetIncarnation: oldTarget.Incarnation, Type: events.MouseDown, Button: events.Left}); !errors.Is(err, cogent.ErrManualStale) {
		t.Fatal("hidden private editor target remained routable", err)
	}
	click(t, m, 7, w.id, "filter")
	if err := owned(p.owner, func() error { err := w.grid.JumpToKey(rowKey(3), columnID(1)); w.report(err); return err }); err != nil {
		t.Fatal(err)
	}
	settle(t, m, 7)
	if editor := measuredEditor(t, p, w); editor.Text != "recoverable" || editor.Session != initial.Session {
		t.Fatal("reappearing source changed cell draft", editor)
	}
	variableGeometry(t, w, inspection(t, p, w))
	click(t, m, 7, w.id, "row-heights")
	if editor := measuredEditor(t, p, w); editor.Session != initial.Session || state(t, w).VariableHeights {
		t.Fatal("height source revision changed edit session", editor)
	}
	click(t, m, 7, w.id, "external-cell")
	conflict := state(t, w).Edit
	if conflict.Status != virtual.GridEditConflict || conflict.Text != "recoverable" || conflict.CurrentVersion != 2 || conflict.BaseVersion != 1 {
		t.Fatal("actual external update did not expose conflict", conflict)
	}
	click(t, m, 7, w.id, "save-cell")
	acceptedCell(t, w, 3, 1, "External update v2", 2)
	if !state(t, w).Edit.Active {
		t.Fatal("conflict silently closed draft")
	}
	click(t, m, 7, w.id, "cancel-cell")
	click(t, m, 7, w.id, "edit-cell")
	measuredEditor(t, p, w)
	beforeModal := window(t, settle(t, m, 7), w.id)
	coveredEditor := element(t, beforeModal, cogent.GridEditorID("data"))
	click(t, m, 7, w.id, "modal-open")
	s := window(t, settle(t, m, 7), w.id)
	if err := m.Input(background, cogent.ManualInput{Window: w.id, WindowIncarnation: s.Incarnation, Target: coveredEditor.ID, TargetIncarnation: coveredEditor.Incarnation, Type: events.MouseDown, Button: events.Left}); !errors.Is(err, cogent.ErrManualDisabled) {
		t.Fatal("modal did not cover inline editor", err)
	}
	before := state(t, w).Edit
	primitiveKey(t, m, 7, w.id, key.CodeEscape)
	if w.modal || state(t, w).Edit != before {
		t.Fatal("modal Escape canceled background grid draft")
	}
	click(t, m, 7, w.id, "cancel-cell")
	if state(t, w).Edit.Active {
		t.Fatal("Cancel button did not retire draft")
	}
}
