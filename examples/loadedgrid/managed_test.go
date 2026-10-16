//go:build offscreen

package main

import (
	"cogentcore.org/core/events"
	keyboard "cogentcore.org/core/events/key"

	"cogentcore.org/core/text/fonts"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"image/color"
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"time"
)

func manualFixture(t *testing.T, batch int) (*program, *cogent.Manual, *execution.Manual) {
	t.Helper()
	ex, e := execution.NewManual(8)
	if e != nil {
		t.Fatal(e)
	}
	clock, e := execution.NewManualClock(time.Unix(1730000000, 0).UTC(), 256, 256)
	if e != nil {
		t.Fatal(e)
	}
	cfg := cogent.ManualConfig{Clock: clock, Executor: ex, OwnerBatch: batch, MaxWindows: 2, EventCapacity: 128, LogicalDPI: 96, Theme: ui.Theme{Default: ui.Style{Foreground: ui.Some(color.NRGBA{R: 24, G: 24, B: 24, A: 255}), Background: ui.Some(color.NRGBA{R: 248, G: 248, B: 248, A: 255}), Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: 16})}}}
	e = fs.WalkDir(fonts.Default, ".", func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() || !strings.HasSuffix(path, ".ttf") && !strings.HasSuffix(path, ".otf") {
			return nil
		}
		bytes, e := fs.ReadFile(fonts.Default, path)
		if e != nil {
			return e
		}
		sum := sha256.Sum256(bytes)
		cfg.Fonts = append(cfg.Fonts, cogent.ManualFont{Name: path, Data: bytes, SHA256: hex.EncodeToString(sum[:])})
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	p, e := newProgram(128, true, ex)
	if e != nil {
		t.Fatal(e)
	}
	for _, w := range p.windows {
		w.settle = func(ctx context.Context, cb func(error)) error {
			return cogent.SettleGridEdit(ctx, p.owner, w.id, w.grid, cb)
		}
		w.settleFilter = func(ctx context.Context, cb func(error)) error {
			return cogent.SettleTextField(ctx, p.owner, w.id, "filter-text", cb)
		}
	}
	m, e := cogent.NewManual(bg, p.owner, cfg)
	if e != nil {
		p.owner.Close()
		t.Fatal(e)
	}
	t.Cleanup(func() {
		if e := m.Close(bg); e != nil {
			t.Error(e)
		}
		ex.Close()
		clock.Close()
	})
	settleManual(t, m, batch)
	return p, m, ex
}
func settleManual(t *testing.T, m *cogent.Manual, batch int) cogent.ManualSnapshot {
	t.Helper()
	for range 1024 {
		s, e := m.Snapshot(bg)
		if e != nil {
			t.Fatal(e)
		}
		ready := s.Application.Ordinary+s.Application.ReadyCompletions > 0
		for _, w := range s.Windows {
			ready = ready || w.Events > 0 || w.ReadyPaint
			if w.Blocked && !w.ReadyPaint {
				t.Fatal("origin blocked", w)
			}
		}
		if ready {
			r, e := m.Step(bg, batch)
			if e != nil {
				t.Fatal(e)
			}
			if r.Events+r.OwnerCallbacks > batch {
				t.Fatal("combinedbudget", r)
			}
			continue
		}
		framed := false
		for _, w := range s.Windows {
			if w.LayoutPending {
				r, e := m.Frame(bg, w.ID)
				if e != nil {
					t.Fatal(e)
				}
				framed = framed || r.Events > 0
			}
		}
		if framed {
			continue
		}
		return s
	}
	t.Fatal("manualsettlebound")
	return cogent.ManualSnapshot{}
}
func measured(t *testing.T, s cogent.ManualSnapshot, id string) cogent.ManualWindowSnapshot {
	t.Helper()
	for _, w := range s.Windows {
		if w.ID == id {
			return w
		}
	}
	t.Fatal("missingwindow", id)
	return cogent.ManualWindowSnapshot{}
}
func target(t *testing.T, w cogent.ManualWindowSnapshot, id string) cogent.ManualElement {
	t.Helper()
	for _, e := range w.Elements {
		if e.ID == id {
			return e
		}
	}
	t.Fatal("missingtarget", id)
	return cogent.ManualElement{}
}
func press(t *testing.T, m *cogent.Manual, batch int, windowID, id string) {
	t.Helper()
	w := measured(t, settleManual(t, m, batch), windowID)
	e := target(t, w, id)
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if err := m.Input(bg, cogent.ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Target: e.ID, TargetIncarnation: e.Incarnation, Type: typ, Button: events.Left}); err != nil {
			t.Fatal(err)
		}
	}
	settleManual(t, m, batch)
}
func work(t *testing.T, ex *execution.Manual, id execution.ID, m *cogent.Manual, batch int) {
	t.Helper()
	if e := ex.Step(bg, id); e != nil {
		t.Fatal(e)
	}
	settleManual(t, m, batch)
}
func gridInspect(t *testing.T, w *window) cogent.GridInspection {
	t.Helper()
	var out cogent.GridInspection
	owner(t, w.p, func() error { var e error; out, e = cogent.InspectGrid(w.p.owner, w.grid); return e })
	if out.Frames == 0 || out.Bounds.Empty() {
		t.Fatal("notmeasured")
	}
	return out
}
func routeKey(t *testing.T, m *cogent.Manual, batch int, id string, r rune, code keyboard.Codes) {
	t.Helper()
	w := measured(t, settleManual(t, m, batch), id)
	if e := m.Input(bg, cogent.ManualInput{Window: id, WindowIncarnation: w.Incarnation, Type: events.KeyChord, Rune: r, Code: code}); e != nil {
		t.Fatal(e)
	}
	settleManual(t, m, batch)
}
func TestActualLoadedGridWorkerRoutingAndFilterFIFO(t *testing.T) {
	var final *capture
	for _, batch := range []int{1, 7, 64} {
		t.Run(fmt.Sprintf("batch-%d", batch), func(t *testing.T) {
			p, m, ex := manualFixture(t, batch)
			w := p.windows[1]
			if gridInspect(t, w).Stamp.Rows != 0 {
				t.Fatal("initialemptyframe")
			}
			press(t, m, batch, "large", "load")
			ids := ex.PendingIDs()
			if len(ids) != 1 || p.base != nil {
				t.Fatal("workerimplicitlyran")
			}
			work(t, ex, ids[0], m, batch)
			if gridInspect(t, w).Stamp.Rows != 128 {
				t.Fatal("loadnotmeasured")
			}
			// Start two actual routed queries, execute newest selected job first, then
			// execute canceled older work. The actual owner callback gates publication.
			press(t, m, batch, "large", "sort-asc")
			press(t, m, batch, "large", "sort-desc")
			ids = ex.PendingIDs()
			if len(ids) != 2 {
				t.Fatal(ids)
			}
			work(t, ex, ids[1], m, batch)
			revision := gridInspect(t, w).Stamp.Revision
			work(t, ex, ids[0], m, batch)
			if gridInspect(t, w).Stamp.Revision != revision || !w.manager.Pending().Idle() {
				t.Fatal("olderqueryinstalled/strandedcredit")
			}
			// No owner read or intermediate Step between final field key admission and
			// the Apply mouse events. Existing native-origin boundary/focus loss must
			// publish the controlled field value before its toolbar callback reads it.
			press(t, m, batch, "large", "filter-text")
			s := measured(t, settleManual(t, m, batch), "large")
			apply := target(t, s, "apply-filter")
			if e := m.Input(bg, cogent.ManualInput{Window: s.ID, WindowIncarnation: s.Incarnation, Type: events.KeyChord, Rune: '7', Code: keyboard.Code7}); e != nil {
				t.Fatal(e)
			}
			pos := apply.Bounds.Min.Add(apply.Bounds.Size().Div(2))
			for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
				if e := m.Input(bg, cogent.ManualInput{Window: s.ID, WindowIncarnation: s.Incarnation, Type: typ, Button: events.Left, Position: pos}); e != nil {
					t.Fatal(e)
				}
			}
			settleManual(t, m, batch)
			owner(t, p, func() error {
				if w.filterText != "7" {
					t.Fatal("Apply outran controlledfield", w.filterText)
				}
				return nil
			})
			ids = ex.PendingIDs()
			if len(ids) != 1 {
				t.Fatal("filterworker", ids)
			}
			work(t, ex, ids[0], m, batch)
			owner(t, p, func() error {
				if len(w.query.Filters) != 1 || w.query.Filters[0].Text != "7" {
					t.Fatalf("stalequeryfield accepted=%+v ownerText=%q message=%q token=%d pending=%+v", w.query, w.filterText, w.message, w.queryToken, w.manager.Pending())
				}
				for i := 0; i < w.view.RowCount(); i++ {
					k, _ := w.view.RowKey(i)
					text, _ := w.view.Cell(k, column(2))
					if !strings.Contains(text, "7") {
						t.Fatal("filtermembership")
					}
				}
				return nil
			})
			press(t, m, batch, "large", "select-all")
			g := gridInspect(t, w)
			if g.Selection.CapturedRows != g.Stamp.Rows {
				t.Fatal("viewselection")
			}
			current := &capture{Stamp: g.Stamp, Selection: g.Selection, Cells: g.Cells}
			if final == nil {
				final = current
			} else if !reflect.DeepEqual(final, current) {
				t.Fatal("batchdependentcapture")
			}
			// Closing a window with a selected pending worker revokes completion. Its
			// worker is still explicit and consumes credit until actual terminal cleanup.
			press(t, m, batch, "small", "sort-asc")
			ids = ex.PendingIDs()
			owner(t, p, func() error { return p.owner.CloseWindow(p.windows[0].windowRef) })
			settleManual(t, m, batch)
			work(t, ex, ids[0], m, batch)
			if !p.windows[0].manager.Pending().Idle() {
				t.Fatal("closedworkercredit")
			}
			if _, e := p.windows[0].lifetime(); e == nil {
				t.Fatal("closedwindowlive")
			}
			if _, e := w.lifetime(); e != nil {
				t.Fatal("survivorclosed")
			}
		})
	}
}

type capture struct {
	Stamp     virtual.GridStamp
	Selection virtual.GridSelectionState
	Cells     []cogent.GridCellGeometry
}

func TestActualLoadedGridAsyncEditingDraftAndConflict(t *testing.T) {
	p, m, ex := manualFixture(t, 7)
	w := p.windows[1]
	press(t, m, 7, "large", "load")
	work(t, ex, ex.PendingIDs()[0], m, 7)
	g := gridInspect(t, w)
	var cell cogent.GridCellGeometry
	for _, c := range g.Cells {
		if c.Column == 1 && !c.Bounds.Empty() {
			cell = c
			break
		}
	}
	press(t, m, 7, "large", cell.ID)
	routeKey(t, m, 7, "large", 0, keyboard.CodeF2)
	g = gridInspect(t, w)
	if g.Editor == nil || !g.Editor.Focused {
		t.Fatal("nativeeditornotfocused")
	}
	s := measured(t, settleManual(t, m, 7), "large")
	// Select all and type a canonical numeric value through actual widget keys.
	var flags keyboard.Modifiers
	flags.SetFlag(true, keyboard.Meta)
	if e := m.Input(bg, cogent.ManualInput{Window: s.ID, WindowIncarnation: s.Incarnation, Type: events.KeyChord, Rune: 'a', Code: keyboard.CodeA, Modifiers: flags}); e != nil {
		t.Fatal(e)
	}
	settleManual(t, m, 7)
	for _, r := range "-12345" {
		routeKey(t, m, 7, "large", r, keyboard.CodeUnknown)
	}
	press(t, m, 7, "large", "save-edit")
	if len(ex.PendingIDs()) != 1 {
		t.Fatal("settle didnotadmit editworker", w.message)
	}
	old := w.domain
	work(t, ex, ex.PendingIDs()[0], m, 7)
	owner(t, p, func() error {
		value, v, e := w.domain.value(cell.RowKey, cell.ColumnID)
		if e != nil || value != -12345 || v != 2 {
			t.Fatal("asyncedit", value, v, e, w.message)
		}
		if w.domain == old {
			t.Fatal("noimmutablecandidate")
		}
		return nil
	})
	g = gridInspect(t, w)
	if g.Edit.Active || g.Editor != nil {
		t.Fatal("savededitorretained")
	}
	press(t, m, 7, "large", "begin-edit")
	routeKey(t, m, 7, "large", '9', keyboard.Code9)
	press(t, m, 7, "large", "external")
	work(t, ex, ex.PendingIDs()[0], m, 7)
	g = gridInspect(t, w)
	if g.Edit.Status != virtual.GridEditConflict {
		t.Fatal("conflictmissing")
	}
	press(t, m, 7, "large", "save-edit")
	if len(ex.PendingIDs()) != 0 {
		t.Fatal("conflictworkadmitted")
	}
	press(t, m, 7, "large", "cancel-edit")
	g = gridInspect(t, w)
	if g.Edit.Active {
		t.Fatal("cancelnotretired")
	}
}

func TestActualLoadedGridAsyncEditRemovesFilteredRow(t *testing.T) {
	p, m, ex := manualFixture(t, 7)
	w := p.windows[1]
	press(t, m, 7, w.id, "load")
	work(t, ex, ex.PendingIDs()[0], m, 7)
	var original string
	owner(t, p, func() error {
		var err error
		original, err = w.domain.Cell(key(0), column(1))
		if err != nil {
			return err
		}
		return w.queryWork(virtual.GridQuery{Filters: []virtual.GridFilter{{ColumnID: column(1), Kind: virtual.GridFilterEquals, Text: original}}})
	})
	work(t, ex, ex.PendingIDs()[0], m, 7)
	g := gridInspect(t, w)
	if g.Stamp.Rows != 1 {
		t.Fatal("expected exact one-row view", g.Stamp)
	}
	var cell cogent.GridCellGeometry
	for _, c := range g.Cells {
		if c.RowKey == key(0) && c.Column == 1 {
			cell = c
			break
		}
	}
	if cell.ID == "" {
		t.Fatal("missing filtered cell")
	}
	press(t, m, 7, w.id, cell.ID)
	routeKey(t, m, 7, w.id, 0, keyboard.CodeF2)
	g = gridInspect(t, w)
	if g.Editor == nil || !g.Editor.Focused {
		t.Fatal("filtered editor not focused")
	}
	s := measured(t, settleManual(t, m, 7), w.id)
	var mods keyboard.Modifiers
	mods.SetFlag(true, keyboard.Meta)
	if err := m.Input(bg, cogent.ManualInput{Window: s.ID, WindowIncarnation: s.Incarnation, Type: events.KeyChord, Rune: 'a', Code: keyboard.CodeA, Modifiers: mods}); err != nil {
		t.Fatal(err)
	}
	settleManual(t, m, 7)
	for _, r := range "314" {
		routeKey(t, m, 7, w.id, r, keyboard.CodeUnknown)
	}
	// Return uses the actual private field/scene route and originating FIFO
	// settlement. No direct GridCommit or descriptor callback is invoked here.
	routeKey(t, m, 7, w.id, 0, keyboard.CodeReturnEnter)
	if ids := ex.PendingIDs(); len(ids) != 1 {
		t.Fatal("Return did not admit edit worker", ids, w.message)
	} else {
		work(t, ex, ids[0], m, 7)
	}
	g = gridInspect(t, w)
	if g.Stamp.Rows != 0 || g.Edit.Active || g.Editor != nil || len(g.Cells) != 0 {
		t.Fatal("filtered-out edit did not retire measured cell/editor", g)
	}
	owner(t, p, func() error {
		value, version, err := w.domain.value(key(0), column(1))
		if err != nil || value != 314 || version != 2 || len(w.query.Filters) != 1 || w.query.Filters[0].Text != original {
			t.Fatal("domain/view/filter commit incoherent", value, version, err, w.query, w.message)
		}
		return nil
	})
	last := measured(t, settleManual(t, m, 7), w.id)
	if last.FocusedID == cogent.GridEditorID("data") {
		t.Fatal("retired editor retains focus")
	}
	for _, element := range last.Elements {
		if element.ID == cogent.GridEditorID("data") {
			t.Fatal("retired editor remains in accessibility", element)
		}
	}
}
