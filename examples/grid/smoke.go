package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/virtual"
)

func owned(a *app.Application, fn func() error) error {
	var result error
	if err := a.Do(func() { result = fn() }); err != nil {
		return err
	}
	return result
}
func emit(stage string, data any) {
	record, err := json.Marshal(struct {
		Stage string `json:"stage"`
		Data  any    `json:"data"`
	}{stage, data})
	if err != nil {
		fmt.Println("grid smoke record encoding failed:", err)
		return
	}
	fmt.Println(string(record))
}
func waitInspection(ctx context.Context, w *gridWindow) (cogent.GridInspection, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		var out cogent.GridInspection
		err := owned(w.owner, func() error { var e error; out, e = cogent.InspectGrid(w.owner, w.grid); return e })
		if err == nil && out.Frames > 0 && !out.Bounds.Empty() && len(out.Cells) > 0 {
			return out, nil
		}
		if err != nil && !errors.Is(err, cogent.ErrGridClosed) && !errors.Is(err, cogent.ErrGridLayout) {
			return out, err
		}
		select {
		case <-ctx.Done():
			return out, fmt.Errorf("grid smoke: %s completed frame: %w", w.id, ctx.Err())
		case <-ticker.C:
		}
	}
}
func checkInspection(g cogent.GridInspection) error {
	if g.Columns != columnCount || g.Stamp.Rows < 1 || g.Scale <= 0 || g.Bounds.Empty() || g.Incarnation == 0 || g.CellIncarnation == 0 || g.Frames == 0 {
		return fmt.Errorf("grid smoke: invalid measured grid: %+v", g)
	}
	if len(g.Cells) > 4096 || len(g.Headers) > columnCount {
		return fmt.Errorf("grid smoke: source-sized native pool: %d cells, %d headers", len(g.Cells), len(g.Headers))
	}
	seen := make(map[string]bool, len(g.Cells))
	for _, c := range g.Cells {
		if c.ID != cogent.GridCellID(g.ID, c.RowKey, c.ColumnID) || seen[c.ID] || c.Row < 0 || c.Row >= g.Stamp.Rows || c.Column < 0 || c.Column >= columnCount || c.Frozen != (c.Column < 2) {
			return fmt.Errorf("grid smoke: invalid measured cell %+v", c)
		}
		seen[c.ID] = true
		if !c.Bounds.Empty() && !c.Bounds.In(g.Bounds) {
			return fmt.Errorf("grid smoke: cell %s escaped grid: %v / %v", c.ID, c.Bounds, g.Bounds)
		}
	}
	for _, h := range g.Headers {
		if h.Column < 0 || h.Column >= columnCount || h.Frozen != (h.Column < 2) || !h.Bounds.Empty() && !h.Bounds.In(g.Bounds) {
			return fmt.Errorf("grid smoke: invalid measured header %+v", h)
		}
	}
	return nil
}
func firstFrozen(g cogent.GridInspection) (image.Rectangle, image.Rectangle, error) {
	var cell, header image.Rectangle
	for _, c := range g.Cells {
		if c.Column == 0 && !c.Bounds.Empty() {
			cell = c.Bounds
			break
		}
	}
	for _, h := range g.Headers {
		if h.Column == 0 && !h.Bounds.Empty() {
			header = h.Bounds
			break
		}
	}
	if cell.Empty() || header.Empty() {
		return cell, header, errors.New("grid smoke: frozen leading column/header not measured")
	}
	return cell, header, nil
}
func measuredActive(g cogent.GridInspection, want virtual.GridActive) bool {
	for _, c := range g.Cells {
		if c.Active && c.RowKey == want.RowKey && c.ColumnID == want.ColumnID && !c.Bounds.Empty() {
			return true
		}
	}
	return false
}

// Desktop smoke uses explicit model commands, not synthesized human input.
// Actual key/pointer routing is separately exercised by offscreen tests.
func desktopSmoke(ctx context.Context, p *gridProgram) error {
	first, second := p.windows[0], p.windows[1]
	firstLifetime, err := p.owner.WindowContext(first.id)
	if err != nil {
		return err
	}
	secondLifetime, err := p.owner.WindowContext(second.id)
	if err != nil {
		return err
	}
	initial := make([]cogent.GridInspection, len(p.windows))
	expectedRows := []int{10_000, 1_000_000}
	for i, w := range p.windows {
		initial[i], err = waitInspection(ctx, w)
		if err != nil {
			return err
		}
		if err = checkInspection(initial[i]); err != nil {
			return err
		}
		if initial[i].Stamp.Rows != expectedRows[i] {
			return errors.New("grid smoke: domain row count mismatch")
		}
	}
	emit("ready", struct {
		Rows  []int
		Cells []int
	}{[]int{10_000, 1_000_000}, []int{len(initial[0].Cells), len(initial[1].Cells)}})
	frozenCell, frozenHeader, err := firstFrozen(initial[1])
	if err != nil {
		return err
	}
	if err = owned(p.owner, func() error { err := second.last(); second.report(err); return err }); err != nil {
		return err
	}
	far, err := waitInspection(ctx, second)
	if err != nil {
		return err
	}
	if err = checkInspection(far); err != nil {
		return err
	}
	if !measuredActive(far, virtual.GridActive{RowKey: rowKey(999_999), ColumnID: columnID(19)}) || far.Viewport.ScrollX <= 0 || far.Viewport.ScrollY <= 0 {
		return errors.New("grid smoke: far row/column not in measured frame")
	}
	farCell, farHeader, err := firstFrozen(far)
	if err != nil {
		return err
	}
	if farCell.Min.X != frozenCell.Min.X || farHeader.Min.Y != frozenHeader.Min.Y {
		return errors.New("grid smoke: frozen leading column/header moved with far scroll")
	}
	emit("far-both-axes", struct {
		Viewport virtual.GridViewport
		Cells    int
		Frames   uint64
	}{far.Viewport, len(far.Cells), far.Frames})
	wantRetained := virtual.GridActive{RowKey: rowKey(999_998), ColumnID: columnID(19)}
	var retained virtual.GridState
	if err = owned(p.owner, func() error {
		if err := second.grid.JumpToKey(wantRetained.RowKey, wantRetained.ColumnID); err != nil {
			return err
		}
		if err := second.projection(true, true); err != nil {
			return err
		}
		var err error
		retained, err = second.grid.State()
		if err != nil {
			return err
		}
		if retained.Stamp.Rows != 500_000 || retained.Stamp.Revision != 2 || retained.Active != wantRetained {
			return errors.New("grid smoke: projection lost stable selected identity")
		}
		// Retaining the active key is independent of the best-effort top-row
		// anchor. A reordered active row may be outside the current viewport.
		// Explicitly reveal it before asserting measured visibility.
		err = second.grid.JumpToKey(wantRetained.RowKey, wantRetained.ColumnID)
		second.report(err)
		return err
	}); err != nil {
		return err
	}
	projected, err := waitInspection(ctx, second)
	if err != nil {
		return err
	}
	if projected.Stamp != retained.Stamp || !measuredActive(projected, wantRetained) {
		return errors.New("grid smoke: explicit retained-key jump not in measured frame")
	}
	if err = owned(p.owner, func() error {
		state, err := first.grid.State()
		if err != nil {
			return err
		}
		if state.Active != (virtual.GridActive{}) || state.Stamp.Rows != 10_000 || state.Stamp.Revision != 1 {
			return errors.New("grid smoke: independent window changed")
		}
		return nil
	}); err != nil {
		return err
	}
	emit("immutable-projection", struct {
		Rows             int
		Revision         uint64
		Active           virtual.GridActive
		RetainedViewport virtual.GridViewport
		RevealedViewport virtual.GridViewport
	}{projected.Stamp.Rows, projected.Stamp.Revision, retained.Active, retained.Viewport, projected.Viewport})
	if p.variableRows {
		if err := desktopVariablePage(ctx, p, second); err != nil {
			return err
		}
	}
	if p.rowSelection {
		if err := desktopRowSelection(ctx, p, second); err != nil {
			return err
		}
	}
	if err = owned(p.owner, func() error { return p.owner.CloseWindow(first.id) }); err != nil {
		return err
	}
	if firstLifetime.Err() == nil || secondLifetime.Err() != nil || p.owner.Context().Err() != nil {
		return errors.New("grid smoke: closing one domain window affected survivor")
	}
	if _, err = waitInspection(ctx, second); err != nil {
		return err
	}
	// InspectGrid refuses after logical close. This observation does not claim
	// that the native registry or compositor has already completed teardown.
	err = owned(p.owner, func() error { _, err := cogent.InspectGrid(p.owner, first.grid); return err })
	if !errors.Is(err, cogent.ErrGridClosed) {
		return fmt.Errorf("grid smoke: logically closed binding inspection: %v", err)
	}

	if err = owned(p.owner, func() error { return p.owner.CloseWindow(second.id) }); err != nil {
		return err
	}
	if secondLifetime.Err() == nil || p.owner.Context().Err() == nil {
		return errors.New("grid smoke: last domain window lifetime remains active")
	}
	emit("pass", struct{ Scope string }{"desktop measured geometry, explicit model commands, closed-binding refusal and domain lifetimes; no human input/performance claim"})
	return nil
}

// This remains a model-command desktop check. The optional source computes the
// expected page destination arithmetically; the completed native inspection must
// contain that exact selected stable key after the public Page command.
func desktopVariablePage(ctx context.Context, p *gridProgram, w *gridWindow) error {
	var before, after virtual.GridState
	var expected virtual.GridActive
	var metricCounts rowMetricCalls
	if err := owned(p.owner, func() error {
		var err error
		before, err = w.grid.State()
		if err != nil {
			return err
		}
		if !before.VariableHeights || !w.variableRows {
			return errors.New("grid smoke: variable source not installed")
		}
		index, found := w.source.rowIndex(before.Active.RowKey)
		if !found {
			return errors.New("grid smoke: variable page requires current active key")
		}
		metrics := &variableGridSource{gridSource: w.source}
		start, err := metrics.RowOffset(index)
		if err != nil {
			return err
		}
		extent, err := metrics.RowOffset(w.source.RowCount())
		if err != nil {
			return err
		}
		pixel := min(extent-1, start+max(1, before.Viewport.Height-before.Config.HeaderHeight))
		target, err := metrics.RowAt(pixel)
		if err != nil {
			return err
		}
		target = max(target, min(index+1, w.source.RowCount()-1))
		key, err := w.source.RowKey(target)
		if err != nil {
			return err
		}
		expected = virtual.GridActive{RowKey: key, ColumnID: before.Active.ColumnID}
		if err = w.grid.Page(1); err != nil {
			return err
		}
		after, err = w.grid.State()
		if err != nil {
			return err
		}
		if !after.VariableHeights || after.Stamp != before.Stamp || after.Active != expected {
			return errors.New("grid smoke: variable Page did not use active-top plus body pixels")
		}
		metricCounts = *w.metricCalls
		w.report(nil)
		return nil
	}); err != nil {
		return err
	}
	measured, err := waitInspection(ctx, w)
	if err != nil {
		return err
	}
	if err = checkInspection(measured); err != nil {
		return err
	}
	if !measuredActive(measured, expected) {
		return errors.New("grid smoke: variable pixel Page selection not in completed measured frame")
	}
	emit("variable-pixel-page", struct {
		Rows         int
		TotalHeight  int
		Active       virtual.GridActive
		Viewport     virtual.GridViewport
		MetricCalls  rowMetricCalls
		SourcePolicy string
	}{after.Stamp.Rows, after.TotalHeight, expected, measured.Viewport, metricCounts, "stable-key periodic heights 23/35/47/29; indexed prefix/inverse; explicit owner Page"})
	return nil
}

// Capturing all rows is an owner command over immutable indexed metadata, not
// an enumeration or an input surrogate. The detached snapshot survives Clear.
func desktopRowSelection(ctx context.Context, p *gridProgram, w *gridWindow) error {
	var captured virtual.GridSelection
	var scope virtual.GridSelectionState
	var current virtual.GridState
	if err := owned(p.owner, func() error {
		// The earlier projection check used the even-only reversed view. Capture
		// the full million-key universe explicitly before filtering it again.
		if err := w.projection(false, false); err != nil {
			return err
		}
		if err := w.grid.SelectAll(); err != nil {
			return err
		}
		for _, key := range []string{rowKey(1), rowKey(2)} {
			if err := w.grid.SelectToKey(key, columnID(0), virtual.GridSelectToggle); err != nil {
				return err
			}
		}
		var err error
		captured, err = w.grid.Selection()
		if err != nil {
			return err
		}
		scope = captured.State()
		if scope.CapturedRows != 999_998 || scope.Scope.Rows != 1_000_000 || scope.Overrides != 2 {
			return errors.New("grid smoke: all-minus-two captured selection count/scope mismatch")
		}
		if err := w.projection(true, true); err != nil {
			return err
		}
		current, err = w.grid.State()
		if err != nil {
			return err
		}
		if current.Stamp.Rows != 500_000 || current.Selection != scope {
			return errors.New("grid smoke: projection changed captured selection universe")
		}
		// A filtered-out odd key remains selected in the earlier captured view.
		for key, want := range map[string]bool{rowKey(0): true, rowKey(1): false, rowKey(2): false, rowKey(3): true, rowKey(999_999): true} {
			got, err := captured.Contains(key)
			if err != nil || got != want {
				return fmt.Errorf("grid smoke: captured membership %s = %t, want %t: %v", key, got, want, err)
			}
		}
		w.report(nil)
		return nil
	}); err != nil {
		return err
	}
	measured, err := waitInspection(ctx, w)
	if err != nil {
		return err
	}
	if measured.Selection != scope || !measured.RowSelection {
		return errors.New("grid smoke: completed selection inspection disagrees with owner")
	}
	for _, cell := range measured.Cells {
		want, err := captured.Contains(cell.RowKey)
		if err != nil || cell.Selected != want {
			return errors.New("grid smoke: measured selected flags disagree with captured identities")
		}
	}
	if err := owned(p.owner, func() error {
		if err := w.grid.ClearSelection(); err != nil {
			return err
		}
		var err error
		current, err = w.grid.State()
		if err != nil {
			return err
		}
		if current.Selection.CapturedRows != 0 {
			return errors.New("grid smoke: Clear did not clear current selection")
		}
		other, err := p.windows[0].grid.State()
		if err != nil || other.Selection.CapturedRows != 0 {
			return errors.New("grid smoke: selecting one window changed independent window")
		}
		w.report(nil)
		return nil
	}); err != nil {
		return err
	}
	// Query the immutable snapshot outside the owner after its source was
	// replaced and current selection cleared. No full-million visit is needed.
	if selected, err := captured.Contains(rowKey(999_999)); err != nil || !selected || captured.State() != scope {
		return errors.New("grid smoke: cleared owner invalidated detached captured selection")
	}
	cleared, err := waitInspection(ctx, w)
	if err != nil {
		return err
	}
	if cleared.Selection.CapturedRows != 0 {
		return errors.New("grid smoke: completed Clear frame retained selected rows")
	}
	for _, cell := range cleared.Cells {
		if cell.Selected {
			return errors.New("grid smoke: completed Clear frame retained selected cell flags")
		}
	}
	emit("captured-row-selection", struct {
		Captured     virtual.GridSelectionState
		CurrentView  virtual.GridStamp
		ClearedCount int
		Scope        string
	}{scope, current.Stamp, current.Selection.CapturedRows, "owner All1M minus two; stable captured identities across filter; detached snapshot after Clear; no full enumeration/desktop input"})
	return nil
}
