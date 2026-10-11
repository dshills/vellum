//go:build offscreen

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image/color"
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"time"

	"cogentcore.org/core/events"
	keyboard "cogentcore.org/core/events/key"
	"cogentcore.org/core/text/fonts"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
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
	p, e := newProgram(maxNodes, ex)
	if e != nil {
		t.Fatal(e)
	}
	for _, w := range p.windows {
		w.settleFilter = func(ctx context.Context, cb func(error)) error {
			return cogent.SettleTextField(ctx, p.owner, w.id, "filter", cb)
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
func routeKey(t *testing.T, m *cogent.Manual, batch int, id string, r rune, code keyboard.Codes) {
	t.Helper()
	w := measured(t, settleManual(t, m, batch), id)
	if e := m.Input(bg, cogent.ManualInput{Window: id, WindowIncarnation: w.Incarnation, Type: events.KeyChord, Rune: r, Code: code}); e != nil {
		t.Fatal(e)
	}
	settleManual(t, m, batch)
}

func treeInspect(t *testing.T, w *window) cogent.TreeInspection {
	t.Helper()
	var out cogent.TreeInspection
	if e := owned(w.p, func() error { var e error; out, e = cogent.InspectTree(w.p.owner, w.outline); return e }); e != nil {
		t.Fatal(e)
	}
	return out
}
func TestActualStoredTreeRoutesAndWorkers(t *testing.T) {
	type capture struct {
		Selected    string
		VisibleRows int
		Viewport    virtual.OutlineViewport
		Rows        []cogent.TreeRowGeometry
	}
	var final *capture
	for _, batch := range []int{1, 7, 64} {
		t.Run(fmt.Sprintf("batch-%d", batch), func(t *testing.T) {
			p, m, ex := manualFixture(t, batch)
			w := p.windows[1]
			if treeInspect(t, w).TotalNodes != 0 {
				t.Fatal("initial view not empty")
			}
			press(t, m, batch, "large", "load")
			ids := ex.PendingIDs()
			if len(ids) != 1 {
				t.Fatal("load ran implicitly", ids)
			}
			work(t, ex, ids[0], m, batch)
			ids = ex.PendingIDs()
			if len(ids) != 2 {
				t.Fatal("request-bound initial index workers", ids)
			}
			for _, id := range ids {
				work(t, ex, id, m, batch)
			}
			if treeInspect(t, w).TotalNodes != 100_000 || treeInspect(t, p.windows[0]).TotalNodes != 10_000 {
				t.Fatal("actual stored sizes not accepted")
			}
			press(t, m, batch, "large", "expand")
			press(t, m, batch, "large", "collapse")
			ids = ex.PendingIDs()
			if len(ids) != 2 {
				t.Fatal("two bounded workers", ids)
			}
			work(t, ex, ids[1], m, batch)
			accepted := treeInspect(t, w)
			work(t, ex, ids[0], m, batch)
			if treeInspect(t, w).Generation != accepted.Generation || !w.manager.Pending().Idle() {
				t.Fatal("canceled older result installed")
			}
			press(t, m, batch, "large", "expand")
			work(t, ex, ex.PendingIDs()[0], m, batch)
			view := treeInspect(t, w)
			first := view.Rows[1]
			// Real row press owns focus; following End travels the actual event router.
			press(t, m, batch, "large", first.ID)
			routeKey(t, m, batch, "large", 0, keyboard.CodeEnd)
			if got := treeInspect(t, w); got.Selected != nodeKey(maxNodes-1) || got.Viewport.ScrollY == 0 || len(got.Rows) > 500 {
				t.Fatal("far routed selected row", got)
			}
			press(t, m, batch, "large", "collapse")
			work(t, ex, ex.PendingIDs()[0], m, batch)
			if s := state(t, w); s.Selected != nodeKey(maxNodes-1) || s.SelectedVisible {
				t.Fatal("hidden key lost", s)
			}
			press(t, m, batch, "large", "filter")
			snapshot := measured(t, settleManual(t, m, batch), "large")
			apply := target(t, snapshot, "apply")
			if e := m.Input(bg, cogent.ManualInput{Window: snapshot.ID, WindowIncarnation: snapshot.Incarnation, Type: events.KeyChord, Rune: '7', Code: keyboard.Code7}); e != nil {
				t.Fatal(e)
			}
			point := apply.Bounds.Min.Add(apply.Bounds.Size().Div(2))
			for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
				if e := m.Input(bg, cogent.ManualInput{Window: snapshot.ID, WindowIncarnation: snapshot.Incarnation, Type: typ, Button: events.Left, Position: point}); e != nil {
					t.Fatal(e)
				}
			}
			settleManual(t, m, batch)
			ids = ex.PendingIDs()
			if len(ids) != 1 {
				t.Fatal("queued field→Apply did not start exactly one worker", ids)
			}
			work(t, ex, ids[0], m, batch)
			if s := state(t, w); s.Filter != "7" || s.Selected != nodeKey(maxNodes-1) {
				t.Fatalf("queued field value raced Apply: state=%+v ownerfield=%q message=%q", s, w.filterText, w.message)
			}
			got := treeInspect(t, w)
			current := &capture{Selected: got.Selected, VisibleRows: got.VisibleRows, Viewport: got.Viewport, Rows: got.Rows}
			if final == nil {
				final = current
			} else if !reflect.DeepEqual(final, current) {
				t.Fatal("routed capture differs by owner batch")
			}
			press(t, m, batch, "large", "expand")
			id := ex.PendingIDs()[0]
			press(t, m, batch, "large", "cancel")
			if w.manager.Pending().Idle() {
				t.Fatal("cancel silently released pending worker")
			}
			work(t, ex, id, m, batch)
			if !w.manager.Pending().Idle() {
				t.Fatal("dual terminal credit stranded")
			}
			press(t, m, batch, "small", "expand")
			id = ex.PendingIDs()[0]
			if e := owned(p, func() error { return p.owner.CloseWindow("small") }); e != nil {
				t.Fatal(e)
			}
			settleManual(t, m, batch)
			if p.windows[0].manager.Pending().Idle() {
				t.Fatal("close hid pending worker")
			}
			work(t, ex, id, m, batch)
			if !p.windows[0].manager.Pending().Idle() {
				t.Fatal("closed worker did not terminally release")
			}
			if treeInspect(t, w).TotalNodes != 100_000 {
				t.Fatal("closed window damaged survivor")
			}
		})
	}
}
func TestActualTreeDisclosureUsesWorker(t *testing.T) {
	p, m, ex := manualFixture(t, 7)
	press(t, m, 7, "large", "load")
	work(t, ex, ex.PendingIDs()[0], m, 7)
	for _, id := range ex.PendingIDs() {
		work(t, ex, id, m, 7)
	}
	w := p.windows[1]
	before := treeInspect(t, w)
	branch := before.Rows[0]
	press(t, m, 7, "large", branch.DisclosureID)
	pending := treeInspect(t, w)
	if !pending.Pending || pending.VisibleRows != before.VisibleRows || len(ex.PendingIDs()) != 1 {
		t.Fatal("disclosure performed index work on router", pending)
	}
	work(t, ex, ex.PendingIDs()[0], m, 7)
	if treeInspect(t, w).VisibleRows != 75_001 {
		t.Fatal("huge-branch disclosure failed")
	}
	routeKey(t, m, 7, "large", 0, keyboard.CodeHome)
	routeKey(t, m, 7, "large", 0, keyboard.CodeRightArrow)
	if state(t, w).Selected != nodeKey(1) {
		t.Fatal("expanded first child keyboard")
	}
	routeKey(t, m, 7, "large", 0, keyboard.CodeLeftArrow)
	if state(t, w).Selected != nodeKey(0) {
		t.Fatal("parent keyboard")
	}
	routeKey(t, m, 7, "large", 0, keyboard.CodeLeftArrow)
	if len(ex.PendingIDs()) != 1 {
		t.Fatal("collapse intent did not request worker")
	}
	work(t, ex, ex.PendingIDs()[0], m, 7)
	if treeInspect(t, w).VisibleRows != 2 {
		t.Fatal("collapse branch failed")
	}
}

func TestActualCancelInitialLoaderBeforeWorker(t *testing.T) {
	for _, batch := range []int{1, 7, 64} {
		t.Run(fmt.Sprintf("batch-%d", batch), func(t *testing.T) {
			p, m, ex := manualFixture(t, batch)
			press(t, m, batch, "large", "load")
			ids := ex.PendingIDs()
			if len(ids) != 1 {
				t.Fatal("initial load did not remain explicit", ids)
			}
			var token uint64
			if e := owned(p, func() error {
				token = p.loadToken
				if p.loadTask == nil {
					t.Fatal("missing admitted loader")
				}
				return nil
			}); e != nil {
				t.Fatal(e)
			}
			press(t, m, batch, "small", "cancel")
			if e := owned(p, func() error {
				if p.loadTask != nil || p.loadToken == token || p.loader.Pending().Active != 1 {
					t.Fatal("Cancel load did not revoke latest token while retaining worker credit", p.loadToken, p.loader.Pending())
				}
				for _, w := range p.windows {
					s, e := w.outline.State()
					if e != nil {
						return e
					}
					if s.Stamp.Nodes != 0 || s.Pending || w.request != (virtual.OutlineRequest{}) || w.task != nil {
						t.Fatal("canceled initial load changed window", w.id, s)
					}
				}
				return nil
			}); e != nil {
				t.Fatal(e)
			}
			work(t, ex, ids[0], m, batch)
			if e := owned(p, func() error {
				if !p.loader.Pending().Idle() || p.base[0] != nil || p.base[1] != nil {
					t.Fatal("canceled loader published or retained terminal credit", p.loader.Pending())
				}
				return nil
			}); e != nil {
				t.Fatal(e)
			}
			if len(ex.PendingIDs()) != 0 {
				t.Fatal("canceled loader admitted child workers")
			}
			press(t, m, batch, "large", "load")
			work(t, ex, ex.PendingIDs()[0], m, batch)
			ids = ex.PendingIDs()
			if len(ids) != 2 {
				t.Fatal("load retry did not admit independent indexes", ids)
			}
			for _, id := range ids {
				work(t, ex, id, m, batch)
			}
			if treeInspect(t, p.windows[1]).TotalNodes != maxNodes || !p.loader.Pending().Idle() {
				t.Fatal("load retry failed after canceled credit retirement")
			}
		})
	}
}
