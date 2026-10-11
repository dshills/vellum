//go:build offscreen

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/text/fonts"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

var background = context.Background()

func manualConfig(t *testing.T, batch int, dpi float32) cogent.ManualConfig {
	t.Helper()
	clock, err := execution.NewManualClock(time.Unix(1730000000, 0).UTC(), 256, 256)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(clock.Close)
	executor, err := execution.NewManual(8)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(executor.Close)
	cfg := cogent.ManualConfig{Clock: clock, Executor: executor, OwnerBatch: batch, MaxWindows: 2, EventCapacity: 128, LogicalDPI: dpi, Theme: ui.Theme{Default: ui.Style{Foreground: ui.Some(color.NRGBA{R: 24, G: 24, B: 24, A: 255}), Background: ui.Some(color.NRGBA{R: 248, G: 248, B: 248, A: 255}), Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: 16})}}}
	err = fs.WalkDir(fonts.Default, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".ttf") && !strings.HasSuffix(path, ".otf") {
			return nil
		}
		data, err := fs.ReadFile(fonts.Default, path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		cfg.Fonts = append(cfg.Fonts, cogent.ManualFont{Name: path, Data: data, SHA256: hex.EncodeToString(sum[:])})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}
func fixture(t *testing.T, batch int, dpi float32) (*gridProgram, *cogent.Manual) {
	t.Helper()
	cfg := manualConfig(t, batch, dpi)
	// Width/height scale together so both DPI fixtures have the same logical
	// viewport. No monitor-fit-dependent desktop dimension is used as evidence.
	p, err := newProgramSize(int(1024*dpi/96), int(760*dpi/96))
	if err != nil {
		t.Fatal(err)
	}
	if dpi == 120 || dpi == 144 {
		// Exercise fractional joins with dimensions that do not multiply to
		// integral scene pixels. This remains a single fixed height per Grid.
		if err := owned(p.owner, func() error {
			for _, w := range p.windows {
				columns := sourceColumns()
				for i := range columns {
					columns[i].Width = 121
				}
				g, err := virtual.NewGrid(w.source, columns, virtual.GridConfig{RowHeight: 23, HeaderHeight: 27, FrozenColumns: 2, OverscanRows: 1, OverscanColumns: 1}, virtual.GridViewport{Width: 800, Height: 508})
				if err != nil {
					return err
				}
				w.grid = g
			}
			return nil
		}); err != nil {
			p.owner.Close()
			t.Fatal(err)
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
func settle(t *testing.T, m *cogent.Manual, batch int) cogent.ManualSnapshot {
	t.Helper()
	for range 1024 {
		s, err := m.Snapshot(background)
		if err != nil {
			t.Fatal(err)
		}
		ready := s.Application.Ordinary+s.Application.ReadyCompletions > 0
		for _, w := range s.Windows {
			ready = ready || w.Events > 0 || w.ReadyPaint
			if w.Blocked && !w.ReadyPaint {
				t.Fatal("blocked originating FIFO", w)
			}
		}
		if ready {
			r, err := m.Step(background, batch)
			if err != nil {
				t.Fatal(err)
			}
			if r.Events+r.OwnerCallbacks > batch {
				t.Fatal("combined step budget exceeded", r, batch)
			}
			continue
		}
		framed := false
		for _, w := range s.Windows {
			if w.LayoutPending {
				r, err := m.Frame(background, w.ID)
				if err != nil {
					t.Fatal(err)
				}
				framed = framed || r.Events > 0
			}
		}
		if framed {
			continue
		}
		return s
	}
	t.Fatal("bounded explicit settle exhausted")
	return cogent.ManualSnapshot{}
}
func window(t *testing.T, s cogent.ManualSnapshot, id string) cogent.ManualWindowSnapshot {
	t.Helper()
	for _, w := range s.Windows {
		if w.ID == id {
			return w
		}
	}
	t.Fatal("missing window", id)
	return cogent.ManualWindowSnapshot{}
}
func element(t *testing.T, w cogent.ManualWindowSnapshot, id string) cogent.ManualElement {
	t.Helper()
	for _, e := range w.Elements {
		if e.ID == id {
			return e
		}
	}
	t.Fatal("missing measured element", w.ID, id)
	return cogent.ManualElement{}
}
func inspection(t *testing.T, p *gridProgram, w *gridWindow) cogent.GridInspection {
	t.Helper()
	var g cogent.GridInspection
	if err := owned(p.owner, func() error { var err error; g, err = cogent.InspectGrid(p.owner, w.grid); return err }); err != nil {
		t.Fatal(err)
	}
	if err := checkInspection(g); err != nil {
		t.Fatal(err)
	}
	return g
}
func state(t *testing.T, w *gridWindow) virtual.GridState {
	t.Helper()
	var s virtual.GridState
	if err := owned(w.owner, func() error { var err error; s, err = w.grid.State(); return err }); err != nil {
		t.Fatal(err)
	}
	return s
}
func click(t *testing.T, m *cogent.Manual, batch int, windowID, target string) {
	t.Helper()
	w := window(t, settle(t, m, batch), windowID)
	e := element(t, w, target)
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if err := m.Input(background, cogent.ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Target: e.ID, TargetIncarnation: e.Incarnation, Type: typ, Button: events.Left}); err != nil {
			t.Fatal(err)
		}
	}
	settle(t, m, batch)
}
func primitiveKey(t *testing.T, m *cogent.Manual, batch int, windowID string, code key.Codes) {
	t.Helper()
	w := window(t, settle(t, m, batch), windowID)
	if err := m.Input(background, cogent.ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: code}); err != nil {
		t.Fatal(err)
	}
	settle(t, m, batch)
}
func visible(t *testing.T, g cogent.GridInspection, nonFrozen bool) cogent.GridCellGeometry {
	t.Helper()
	for _, c := range g.Cells {
		if !c.Bounds.Empty() && (!nonFrozen || !c.Frozen) {
			return c
		}
	}
	t.Fatal("no visible cell", g)
	return cogent.GridCellGeometry{}
}
func compareProjectionGeometry(t *testing.T, w *gridWindow, g cogent.GridInspection) {
	t.Helper()
	var projection virtual.GridSnapshot
	if err := owned(w.owner, func() error { var err error; projection, err = w.grid.Snapshot(); return err }); err != nil {
		t.Fatal(err)
	}
	if len(projection.Cells) != len(g.Cells) || len(projection.Headers) != len(g.Headers) {
		t.Fatal("native pool differs from bounded projection")
	}
	bounds := func(r virtual.GridRect) image.Rectangle {
		if r.Empty() {
			return image.Rectangle{}
		}
		edge := func(v int) int { return int(math.Round(float64(float32(v) * g.Scale))) }
		return image.Rect(g.Bounds.Min.X+edge(r.X), g.Bounds.Min.Y+edge(r.Y), g.Bounds.Min.X+edge(r.X+r.Width), g.Bounds.Min.Y+edge(r.Y+r.Height))
	}
	for i, c := range projection.Cells {
		actual := g.Cells[i]
		if actual.RowKey != c.RowKey || actual.ColumnID != c.ColumnID || actual.Text != c.Text || actual.Active != c.Active || g.Scale == float32(int(g.Scale)) && actual.Bounds != bounds(c.Clip) {
			t.Fatal("actual cell geometry/identity differs", actual, c, bounds(c.Clip))
		}
	}
	for i, h := range projection.Headers {
		actual := g.Headers[i]
		if actual.ColumnID != h.ColumnID || actual.Title != h.Title || g.Scale == float32(int(g.Scale)) && actual.Bounds != bounds(h.Clip) {
			t.Fatal("actual header geometry/identity differs", actual, h, bounds(h.Clip))
		}
	}
	// Fractional scene origins are not exposed in GridInspection. Comparing a
	// second, locally rounded reconstruction would invent an origin. Shared
	// actual edges instead prove there are no cracks/overlaps at pane/row joins.
	rows := map[int][]cogent.GridCellGeometry{}
	columns := map[int][]cogent.GridCellGeometry{}
	headers := map[int]cogent.GridHeaderGeometry{}
	for _, h := range g.Headers {
		if !h.Bounds.Empty() {
			headers[h.Column] = h
		}
	}
	for _, c := range g.Cells {
		if c.Bounds.Empty() {
			continue
		}
		rows[c.Row] = append(rows[c.Row], c)
		columns[c.Column] = append(columns[c.Column], c)
		if h, ok := headers[c.Column]; ok && (h.Bounds.Min.X != c.Bounds.Min.X || h.Bounds.Max.X != c.Bounds.Max.X) {
			t.Fatal("header/cell column edges disagree", h, c)
		}
	}
	for _, row := range rows {
		for i := 1; i < len(row); i++ {
			if row[i-1].Bounds.Max.X != row[i].Bounds.Min.X || row[i-1].Bounds.Min.Y != row[i].Bounds.Min.Y || row[i-1].Bounds.Max.Y != row[i].Bounds.Max.Y {
				t.Fatal("actual horizontal grid crack/overlap", row[i-1], row[i])
			}
		}
	}
	for _, column := range columns {
		for i := 1; i < len(column); i++ {
			if column[i-1].Row+1 == column[i].Row && column[i-1].Bounds.Max.Y != column[i].Bounds.Min.Y {
				t.Fatal("actual vertical grid crack/overlap", column[i-1], column[i])
			}
		}
	}
}

type routedCapture struct {
	States                []virtual.GridState
	Grids                 []cogent.GridInspection
	Selections, Viewports []uint64
	Images                []string
}

func capture(t *testing.T, p *gridProgram, m *cogent.Manual, batch int) routedCapture {
	t.Helper()
	settle(t, m, batch)
	out := routedCapture{}
	for _, w := range p.windows {
		g := inspection(t, p, w)
		compareProjectionGeometry(t, w, g)
		g.Frames = 0
		g.Incarnation = 0
		g.CellIncarnation = 0 // dispatch/cache clocks are reported separately, not compared as pixels.
		out.Grids = append(out.Grids, g)
		out.States = append(out.States, state(t, w))
		out.Selections = append(out.Selections, w.selectionEvents)
		out.Viewports = append(out.Viewports, w.viewportEvents)
		img, err := m.Capture(background, w.id)
		if err != nil {
			t.Fatal(err)
		}
		if img.Bounds().Empty() || !g.Bounds.In(img.Bounds()) {
			t.Fatal("grid escaped actual software image", g.Bounds, img.Bounds())
		}
		sum := sha256.Sum256(img.Pix)
		out.Images = append(out.Images, hex.EncodeToString(sum[:]))
	}
	return out
}
func scenario(t *testing.T, batch int) routedCapture {
	t.Helper()
	p, m := fixture(t, batch, 96)
	large, small := p.windows[1], p.windows[0]
	initial := inspection(t, p, large)
	cell0, header0, err := firstFrozen(initial)
	if err != nil {
		t.Fatal(err)
	}
	click(t, m, batch, "large", "last") // actual managed button router calls the public JumpToKey command.
	far := inspection(t, p, large)
	want := virtual.GridActive{RowKey: rowKey(999_999), ColumnID: columnID(19)}
	if !measuredActive(far, want) || far.Viewport.ScrollX == 0 || far.Viewport.ScrollY == 0 {
		t.Fatal("far jump not measured", far.Viewport)
	}
	cell1, header1, err := firstFrozen(far)
	if err != nil || cell1.Min.X != cell0.Min.X || header1.Min.Y != header0.Min.Y {
		t.Fatal("frozen region moved", cell0, cell1, header0, header1, err)
	}
	// Focus an actual visible cell; then physical keys navigate the focused grid.
	activeID := cogent.GridCellID("data", want.RowKey, want.ColumnID)
	click(t, m, batch, "large", activeID)
	primitiveKey(t, m, batch, "large", key.CodeUpArrow)
	if state(t, large).Active.RowKey != rowKey(999_998) {
		t.Fatal("key did not route through grid focus")
	}
	before := state(t, large)
	primitiveKey(t, m, batch, "large", key.CodePageUp)
	if state(t, large).Active.RowKey == before.Active.RowKey {
		t.Fatal("page key did not move")
	}
	// Reset with a routed button, select a stable even key, then sort/filter with
	// routed buttons. The indexed immutable source is replaced, not copied.
	click(t, m, batch, "large", "last")
	far = inspection(t, p, large)
	click(t, m, batch, "large", cogent.GridCellID("data", rowKey(999_999), columnID(19)))
	primitiveKey(t, m, batch, "large", key.CodeUpArrow)
	click(t, m, batch, "large", "reverse")
	click(t, m, batch, "large", "filter")
	s := state(t, large)
	if s.Active.RowKey != rowKey(999_998) || s.Stamp.Rows != 500_000 || s.Stamp.Revision != 3 {
		t.Fatal("sort/filter changed stable selected identity", s)
	}
	if s2 := state(t, small); s2.Active != (virtual.GridActive{}) || s2.Stamp.Rows != 10_000 || s2.Stamp.Revision != 1 {
		t.Fatal("independent model changed", s2)
	}
	click(t, m, batch, "small", "pan")
	return capture(t, p, m, batch)
}
func TestActualGridRouterBatchesEquivalent(t *testing.T) {
	var expected routedCapture
	for _, batch := range []int{1, 7, 64} {
		t.Run(fmt.Sprint(batch), func(t *testing.T) {
			got := scenario(t, batch)
			if batch == 1 {
				expected = got
			} else if !reflect.DeepEqual(expected, got) {
				a, _ := json.Marshal(expected)
				b, _ := json.Marshal(got)
				t.Fatalf("routed capture differs at batch %d\n%s\n%s", batch, a, b)
			}
		})
	}
}
func TestActualGridFrozenGeometryAndIndexedMaterialization(t *testing.T) {
	for _, dpi := range []float32{96, 120, 144, 192} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			p, m := fixture(t, 7, dpi)
			var first cogent.GridInspection
			for i, w := range p.windows {
				g := inspection(t, p, w)
				compareProjectionGeometry(t, w, g)
				if g.Scale != dpi/96 {
					t.Fatal("wrong DPI", g.Scale)
				}
				if i == 0 {
					first = g
				} else if g.Viewport != first.Viewport || len(g.Cells) != len(first.Cells) || len(g.Headers) != len(first.Headers) {
					t.Fatal("same logical viewport has row-count-dependent geometry", first.Viewport, g.Viewport)
				}
				var projection virtual.GridSnapshot
				var calls sourceCalls
				if err := owned(p.owner, func() error {
					*w.source.calls = sourceCalls{}
					var err error
					projection, err = w.grid.Snapshot()
					calls = *w.source.calls
					return err
				}); err != nil {
					t.Fatal(err)
				}
				if calls.Cell != uint64(len(projection.Cells)) || calls.RowKey > 256 || calls.IndexOf > 256 || len(projection.Cells) > 4096 || len(projection.Headers) > 20 {
					t.Fatal("snapshot exceeded bounded indexed work", calls, len(projection.Cells))
				}
			}
			// Real wheel dispatch, including both axes and a frozen region at 2 DPI.
			w := p.windows[1]
			g := inspection(t, p, w)
			cell := visible(t, g, true)
			beforeViewportEvents := w.viewportEvents
			s := window(t, settle(t, m, 7), w.id)
			if err := m.Input(background, cogent.ManualInput{Window: w.id, WindowIncarnation: s.Incarnation, Type: events.Scroll, Position: cell.Bounds.Min.Add(cell.Bounds.Size().Div(2)), Delta: math32.Vector2{X: 240 * g.Scale, Y: 48 * g.Scale}}); err != nil {
				t.Fatal(err)
			}
			settle(t, m, 7)
			after := inspection(t, p, w)
			if after.Viewport.ScrollX != 240 || after.Viewport.ScrollY != 48 || w.viewportEvents != beforeViewportEvents+1 {
				t.Fatal("wheel did not route both axes", after.Viewport, w.viewportEvents)
			}
			compareProjectionGeometry(t, w, after)
			beforeCell, beforeHeader, err := firstFrozen(g)
			if err != nil {
				t.Fatal(err)
			}
			afterCell, afterHeader, err := firstFrozen(after)
			if err != nil || beforeCell.Min.X != afterCell.Min.X || beforeHeader.Min.Y != afterHeader.Min.Y {
				t.Fatal("frozen wheel geometry moved", err)
			}
		})
	}
}
func TestActualGridStalePressDisabledModalAndRetirement(t *testing.T) {
	p, m := fixture(t, 7, 96)
	w := p.windows[1]
	g := inspection(t, p, w)
	cell := visible(t, g, true)
	s := window(t, settle(t, m, 7), w.id)
	e := element(t, s, cell.ID)
	down := cogent.ManualInput{Window: w.id, WindowIncarnation: s.Incarnation, Target: e.ID, TargetIncarnation: e.Incarnation, Type: events.MouseDown, Button: events.Left}
	if err := m.Input(background, down); err != nil {
		t.Fatal(err)
	}
	if r, err := m.Step(background, 1); err != nil || r.Events != 1 {
		t.Fatal("down not explicitly dispatched", r, err)
	}
	beforeEvents := w.selectionEvents
	if err := owned(p.owner, func() error { err := w.grid.Scroll(0, 24); w.report(err); return err }); err != nil {
		t.Fatal(err)
	}
	settle(t, m, 7)
	stale := down
	stale.Type = events.MouseUp
	if err := m.Input(background, stale); !errors.Is(err, cogent.ErrManualStale) {
		t.Fatal("stale measured cell admitted", err)
	}
	stale.Target = ""
	stale.Position = e.Bounds.Min.Add(e.Bounds.Size().Div(2))
	if err := m.Input(background, stale); err != nil {
		t.Fatal(err)
	}
	settle(t, m, 7)
	if w.selectionEvents != beforeEvents || state(t, w).Active != (virtual.GridActive{}) {
		t.Fatal("recycled slot press activated replacement")
	}
	// A domain source replacement while a real press is armed also retires the
	// old identity, independently of the vertical-slot recycling case above.
	g = inspection(t, p, w)
	cell = visible(t, g, true)
	s = window(t, settle(t, m, 7), w.id)
	e = element(t, s, cell.ID)
	down.WindowIncarnation, down.Target, down.TargetIncarnation = s.Incarnation, e.ID, e.Incarnation
	if err := m.Input(background, down); err != nil {
		t.Fatal(err)
	}
	if r, err := m.Step(background, 1); err != nil || r.Events != 1 {
		t.Fatal("source-case down not dispatched", r, err)
	}
	if err := owned(p.owner, func() error { err := w.projection(true, false); w.report(err); return err }); err != nil {
		t.Fatal(err)
	}
	settle(t, m, 7)
	up := down
	up.Type, up.Target, up.Position = events.MouseUp, "", e.Bounds.Min.Add(e.Bounds.Size().Div(2))
	if err := m.Input(background, up); err != nil {
		t.Fatal(err)
	}
	settle(t, m, 7)
	if w.selectionEvents != beforeEvents || state(t, w).Active != (virtual.GridActive{}) {
		t.Fatal("source-retired press activated")
	}

	click(t, m, 7, w.id, "disable")
	g = inspection(t, p, w)
	cell = visible(t, g, true)
	s = window(t, settle(t, m, 7), w.id)
	e = element(t, s, cell.ID)
	blocked := cogent.ManualInput{Window: w.id, WindowIncarnation: s.Incarnation, Target: e.ID, TargetIncarnation: e.Incarnation, Type: events.MouseDown, Button: events.Left}
	if err := m.Input(background, blocked); !errors.Is(err, cogent.ErrManualDisabled) {
		t.Fatal("disabled cell admitted", err)
	}
	// Raw physical events also route through the real disabled widget tree.
	blocked.Target = ""
	blocked.Position = e.Bounds.Min.Add(e.Bounds.Size().Div(2))
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		blocked.Type = typ
		if err := m.Input(background, blocked); err != nil {
			t.Fatal(err)
		}
	}
	settle(t, m, 7)
	if w.selectionEvents != beforeEvents {
		t.Fatal("disabled physical route activated")
	}
	click(t, m, 7, w.id, "disable")
	click(t, m, 7, w.id, "modal-open")
	s = window(t, settle(t, m, 7), w.id)
	blocked.WindowIncarnation = s.Incarnation
	blocked.Target = "data"
	blocked.TargetIncarnation = element(t, s, "data").Incarnation
	blocked.Type = events.MouseDown
	if err := m.Input(background, blocked); !errors.Is(err, cogent.ErrManualDisabled) {
		t.Fatal("modal-covered grid admitted", err)
	}
	click(t, m, 7, w.id, "modal-close")
	if w.modal {
		t.Fatal("actual dialog button did not dismiss")
	}
	click(t, m, 7, w.id, "close")
	for _, remaining := range settle(t, m, 7).Windows {
		if remaining.ID == w.id {
			t.Fatal("closed manual window still retained", remaining)
		}
	}
	if p.owner.Context().Err() != nil {
		t.Fatal("survivor application closed")
	}
	if _, err := m.Capture(background, w.id); !errors.Is(err, cogent.ErrManualClosed) {
		t.Fatal("closed window capture accepted", err)
	}
	if err := owned(p.owner, func() error { _, err := cogent.InspectGrid(p.owner, w.grid); return err }); !errors.Is(err, cogent.ErrGridClosed) {
		t.Fatal("retired grid still inspectable", err)
	}
	if err := m.Input(background, down); !errors.Is(err, cogent.ErrManualClosed) {
		t.Fatal("closed window input admitted", err)
	}
	// Close the remaining window through its real button; this is the terminal
	// domain/native manual lifetime test rather than a private resource guess.
	click(t, m, 7, "small", "close")
	if p.owner.Context().Err() == nil {
		t.Fatal("last-window lifetime remained live")
	}
}

// Compile-time check: the independent example uses the public owner API only.
var _ *app.Application

func TestActualGridStatusWrapDoesNotWidenLayout(t *testing.T) {
	p, m := fixture(t, 7, 96)
	w := p.windows[1]
	before := inspection(t, p, w)
	longStatus := strings.Repeat("Long selected identity and projection status ", 8)
	if err := owned(p.owner, func() error { w.status = longStatus; w.report(nil); return nil }); err != nil {
		t.Fatal(err)
	}
	s := window(t, settle(t, m, 7), w.id)
	status := element(t, s, "status")
	if status.Value != longStatus+fmt.Sprintf(" · %d rows × %d columns · revision %d · active %s/%s", 1_000_000, columnCount, 1, "", "") {
		t.Fatal("wrapped status lost semantic text", status.Value)
	}
	if status.Bounds.Empty() || status.Bounds.Dx() > statusMaxWidth {
		t.Fatal("status exceeded logical width cap at scale one", status.Bounds)
	}
	after := inspection(t, p, w)
	if after.Bounds.Dx() != before.Bounds.Dx() || after.Viewport.Width != before.Viewport.Width {
		t.Fatal("status widened the actual grid viewport", before.Viewport, after.Viewport)
	}
}
