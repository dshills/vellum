//go:build offscreen

package main

import (
	"errors"
	"image"
	"math"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/virtual"
)

func variableFixture(t *testing.T, batch int, dpi float32) (*gridProgram, *cogent.Manual) {
	t.Helper()
	cfg := manualConfig(t, batch, dpi)
	p, err := newProgramMode(int(1024*dpi/96), int(760*dpi/96), true)
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

// InspectGrid exposes rounded scene bounds rather than a fractional origin.
// Compare exact indexed logical clips to actual edges within that one-pixel
// rounding uncertainty, then require adjacent real rows to share actual edges.
func variableGeometry(t *testing.T, w *gridWindow, g cogent.GridInspection) {
	t.Helper()
	var snapshot virtual.GridSnapshot
	if err := withModelOwner(w.owner, func() error {
		var err error
		snapshot, err = w.grid.Snapshot()
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if !snapshot.State.VariableHeights || len(snapshot.Cells) != len(g.Cells) {
		t.Fatal("variable cached projection mismatch", snapshot.State, len(snapshot.Cells), len(g.Cells))
	}
	byID := make(map[string]virtual.GridCell, len(snapshot.Cells))
	for _, cell := range snapshot.Cells {
		byID[cogent.GridCellID(g.ID, cell.RowKey, cell.ColumnID)] = cell
	}
	lastRow := make(map[int]cogent.GridCellGeometry)
	for _, actual := range g.Cells {
		cell, ok := byID[actual.ID]
		if !ok || actual.RowKey != cell.RowKey || actual.ColumnID != cell.ColumnID || actual.Text != cell.Text || actual.Row != cell.RowIndex || actual.Column != cell.ColumnIndex || actual.Frozen != cell.Frozen || actual.Active != cell.Active {
			t.Fatal("variable measured identity mismatch", actual, cell)
		}
		domain, err := w.source.domainRow(cell.RowIndex)
		if err != nil || cell.Rect.Height != domainRowHeight(domain) {
			t.Fatal("row height followed projected index", cell, domain, err)
		}
		if cell.Clip.Empty() {
			if !actual.Bounds.Empty() {
				t.Fatal("variable overscan cell became visible", actual)
			}
			continue
		}
		clip := cell.Clip
		want := image.Rect(g.Bounds.Min.X+int(math.Round(float64(clip.X)*float64(g.Scale))), g.Bounds.Min.Y+int(math.Round(float64(clip.Y)*float64(g.Scale))), g.Bounds.Min.X+int(math.Round(float64(clip.X+clip.Width)*float64(g.Scale))), g.Bounds.Min.Y+int(math.Round(float64(clip.Y+clip.Height)*float64(g.Scale))))
		if abs(actual.Bounds.Min.X-want.Min.X) > 1 || abs(actual.Bounds.Min.Y-want.Min.Y) > 1 || abs(actual.Bounds.Max.X-want.Max.X) > 1 || abs(actual.Bounds.Max.Y-want.Max.Y) > 1 || !actual.Bounds.In(g.Bounds) {
			t.Fatal("variable measured clip disagrees with indexed geometry", actual.Bounds, want, g.Bounds)
		}
		if previous, found := lastRow[actual.Column]; found && previous.Row+1 == actual.Row && previous.Bounds.Max.Y != actual.Bounds.Min.Y {
			t.Fatal("variable rows leave actual crack/overlap", previous, actual)
		}
		lastRow[actual.Column] = actual
	}
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func variablePageDestination(t *testing.T, w *gridWindow, s virtual.GridState) string {
	t.Helper()
	metrics := &variableGridSource{gridSource: w.source}
	index, found := w.source.rowIndex(s.Active.RowKey)
	if !found {
		t.Fatal("page fixture has no active key", s)
	}
	start, err := metrics.RowOffset(index)
	if err != nil {
		t.Fatal(err)
	}
	extent, err := metrics.RowOffset(w.source.RowCount())
	if err != nil {
		t.Fatal(err)
	}
	row, err := metrics.RowAt(min(extent-1, start+max(1, s.Viewport.Height-s.Config.HeaderHeight)))
	if err != nil {
		t.Fatal(err)
	}
	row = max(row, min(index+1, w.source.RowCount()-1))
	domain, err := w.source.domainRow(row)
	if err != nil {
		t.Fatal(err)
	}
	return rowKey(domain)
}

func TestActualVariableGridGeometryAndBoundedIndexedWork(t *testing.T) {
	for _, dpi := range []float32{96, 120, 144, 192} {
		t.Run(stringDPI(dpi), func(t *testing.T) {
			p, m := variableFixture(t, 7, dpi)
			var firstCells int
			var firstCalls sourceCalls
			var firstMetrics rowMetricCalls
			for i, w := range p.windows {
				g := inspection(t, p, w)
				variableGeometry(t, w, g)
				var snapshot virtual.GridSnapshot
				var calls sourceCalls
				var metricCalls rowMetricCalls
				if err := withModelOwner(p.owner, func() error {
					*w.source.calls, *w.metricCalls = sourceCalls{}, rowMetricCalls{}
					var err error
					snapshot, err = w.grid.Snapshot()
					calls, metricCalls = *w.source.calls, *w.metricCalls
					return err
				}); err != nil {
					t.Fatal(err)
				}
				if len(snapshot.Cells) > 4096 || calls.Cell != uint64(len(snapshot.Cells)) || calls.RowKey > 256 || calls.IndexOf > 256 || metricCalls.Offset+metricCalls.At > uint64(12*len(snapshot.Cells)+32) {
					t.Fatal("variable projection enumerated source-sized data", calls, metricCalls, len(snapshot.Cells))
				}
				if i == 0 {
					firstCells, firstCalls, firstMetrics = len(snapshot.Cells), calls, metricCalls
				} else if firstCells != len(snapshot.Cells) || firstCalls != calls || firstMetrics != metricCalls {
					t.Fatal("same viewport indexed work depends on 10k versus1M rows", firstCells, len(snapshot.Cells), firstCalls, calls, firstMetrics, metricCalls)
				}
			}
			w := p.windows[1]
			before := inspection(t, p, w)
			frozenCell, frozenHeader, err := firstFrozen(before)
			if err != nil {
				t.Fatal(err)
			}
			click(t, m, 7, w.id, "last")
			far := inspection(t, p, w)
			variableGeometry(t, w, far)
			if !measuredActive(far, virtual.GridActive{RowKey: rowKey(999_999), ColumnID: columnID(19)}) || far.Viewport.ScrollX <= 0 || far.Viewport.ScrollY <= 0 {
				t.Fatal("variable last row/column not measured", far)
			}
			farCell, farHeader, err := firstFrozen(far)
			if err != nil || farCell.Min.X != frozenCell.Min.X || farHeader.Min.Y != frozenHeader.Min.Y {
				t.Fatal("variable frozen geometry moved", err)
			}
		})
	}
}

func stringDPI(dpi float32) string {
	return map[float32]string{96: "96", 120: "120", 144: "144", 192: "192"}[dpi]
}

func TestActualVariableGridPixelPageWheelAndHeightRevision(t *testing.T) {
	p, m := variableFixture(t, 7, 120)
	w := p.windows[1]
	click(t, m, 7, w.id, "first")
	click(t, m, 7, w.id, cogent.GridCellID("data", rowKey(0), columnID(0)))
	before := state(t, w)
	want := variablePageDestination(t, w, before)
	primitiveKey(t, m, 7, w.id, key.CodePageDown)
	after := state(t, w)
	if after.Active.RowKey != want || !after.VariableHeights || after.Stamp != before.Stamp {
		t.Fatal("actual Page used fixed row count instead of pixel distance", before, after, want)
	}
	g := inspection(t, p, w)
	variableGeometry(t, w, g)
	if !measuredActive(g, after.Active) {
		t.Fatal("actual variable Page target not measured", after.Active)
	}
	cell := visible(t, g, true)
	s := window(t, settle(t, m, 7), w.id)
	if err := m.Input(background, cogent.ManualInput{Window: w.id, WindowIncarnation: s.Incarnation, Type: events.Scroll, Position: cell.Bounds.Min.Add(cell.Bounds.Size().Div(2)), Delta: math32.Vector2{X: 240 * g.Scale, Y: 61 * g.Scale}}); err != nil {
		t.Fatal(err)
	}
	settle(t, m, 7)
	scrolled := inspection(t, p, w)
	if scrolled.Viewport.ScrollX != g.Viewport.ScrollX+240 || scrolled.Viewport.ScrollY != g.Viewport.ScrollY+61 {
		t.Fatal("variable wheel converted pixels to fixed rows", g.Viewport, scrolled.Viewport)
	}
	variableGeometry(t, w, scrolled)
	cell = visible(t, scrolled, true)
	s = window(t, settle(t, m, 7), w.id)
	e := element(t, s, cell.ID)
	down := cogent.ManualInput{Window: w.id, WindowIncarnation: s.Incarnation, Target: e.ID, TargetIncarnation: e.Incarnation, Type: events.MouseDown, Button: events.Left}
	if err := m.Input(background, down); err != nil {
		t.Fatal(err)
	}
	if r, err := m.Step(background, 1); err != nil || r.Events != 1 {
		t.Fatal("variable press not routed", r, err)
	}
	beforeEvents := w.selectionEvents
	if err := withModelOwner(p.owner, func() error { err := w.setVariableRows(false); w.report(err); return err }); err != nil {
		t.Fatal(err)
	}
	settle(t, m, 7)
	up := down
	up.Type = events.MouseUp
	if err := m.Input(background, up); !errors.Is(err, cogent.ErrManualStale) {
		t.Fatal("height-revision stale cell admitted", err)
	}
	up.Target, up.Position = "", e.Bounds.Min.Add(e.Bounds.Size().Div(2))
	if err := m.Input(background, up); err != nil {
		t.Fatal(err)
	}
	settle(t, m, 7)
	if state(t, w).VariableHeights || state(t, w).Stamp.Revision != 2 || w.selectionEvents != beforeEvents {
		t.Fatal("height revision retained stale press or wrong mode", state(t, w), w.selectionEvents)
	}
	click(t, m, 7, w.id, "row-heights")
	if !state(t, w).VariableHeights || state(t, w).Stamp.Revision != 3 || !state(t, p.windows[0]).VariableHeights || state(t, p.windows[0]).Stamp.Revision != 1 {
		t.Fatal("actual toggle changed independent window or failed revision", state(t, w), state(t, p.windows[0]))
	}
	// Actual source controls retain the selected even domain key while changing
	// projection order/membership. Explicit first-key reveal then permits the
	// same indexed height/clip assertions on the reverse-filtered native frame.
	click(t, m, 7, w.id, "last")
	click(t, m, 7, w.id, cogent.GridCellID("data", rowKey(999_999), columnID(19)))
	primitiveKey(t, m, 7, w.id, key.CodeUpArrow)
	click(t, m, 7, w.id, "reverse")
	click(t, m, 7, w.id, "filter")
	projected := state(t, w)
	if !projected.VariableHeights || projected.Active.RowKey != rowKey(999_998) || projected.Active.ColumnID != columnID(19) || projected.Stamp.Rows != 500_000 || projected.Stamp.Revision != 5 {
		t.Fatal("actual variable source projection lost stable selection", projected)
	}
	click(t, m, 7, w.id, "first")
	revealed := inspection(t, p, w)
	variableGeometry(t, w, revealed)
	if !measuredActive(revealed, virtual.GridActive{RowKey: rowKey(999_998), ColumnID: columnID(0)}) {
		t.Fatal("reverse-filtered variable key not revealed", revealed)
	}
}
