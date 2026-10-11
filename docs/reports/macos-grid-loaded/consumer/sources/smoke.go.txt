package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/virtual"
	"time"
)

func owned(p *program, fn func() error) error {
	var e error
	if err := p.owner.Do(func() { e = fn() }); err != nil {
		return err
	}
	return e
}
func emit(stage string, v any) error {
	b, e := json.Marshal(struct {
		Stage string `json:"stage"`
		Data  any    `json:"data"`
	}{stage, v})
	if e != nil {
		return e
	}
	fmt.Println(string(b))
	return nil
}
func wait(ctx context.Context, p *program, fn func() (bool, error)) error {
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		var done bool
		e := owned(p, func() error { var e error; done, e = fn(); return e })
		if e != nil {
			return e
		}
		if done {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
func inspect(ctx context.Context, w *window) (cogent.GridInspection, error) {
	var out cogent.GridInspection
	e := wait(ctx, w.p, func() (bool, error) {
		var e error
		out, e = cogent.InspectGrid(w.p.owner, w.grid)
		if errors.Is(e, cogent.ErrGridClosed) || errors.Is(e, cogent.ErrGridLayout) {
			return false, nil
		}
		if e != nil {
			return false, e
		}
		state, e := w.grid.State()
		if e != nil {
			return false, e
		}
		return out.Frames > 0 && !out.Bounds.Empty() && out.Stamp == state.Stamp && out.Generation == state.Generation, e
	})
	return out, e
}
func desktopSmoke(ctx context.Context, p *program) error {
	for _, w := range p.windows {
		g, e := inspect(ctx, w)
		if e != nil {
			return e
		}
		if g.Stamp.Rows != 0 {
			return errors.New("loadedgrid smoke: initial model was not empty")
		}
	}
	if e := emit("empty-ready", p.rows); e != nil {
		return e
	}
	if e := owned(p, p.load); e != nil {
		return e
	}
	if e := wait(ctx, p, func() (bool, error) { return p.base != nil && p.loadTask == nil, nil }); e != nil {
		return fmt.Errorf("load: %w", e)
	}
	for i, w := range p.windows {
		g, e := inspect(ctx, w)
		if e != nil {
			return e
		}
		want := p.rows
		if i == 0 {
			want = min(want, 10_000)
		}
		if g.Stamp.Rows != want || len(g.Cells) > 4096 {
			return errors.New("loadedgrid smoke: stored input/visible capacity mismatch")
		}
	}
	w := p.windows[1]
	var logicalBytes uint64
	if e := owned(p, func() error { logicalBytes = p.baseSize(); return nil }); e != nil {
		return e
	}
	if e := emit("stored-loaded", struct {
		Rows         int
		LogicalBytes uint64
	}{p.rows, logicalBytes}); e != nil {
		return e
	}
	var initial *fullSource
	var beforeSelection virtual.GridSelection
	if e := owned(p, func() error {
		initial = w.domain
		if e := w.grid.JumpToKey(key(min(7, w.domain.rows-1)), column(1)); e != nil {
			return e
		}
		if e := w.grid.SelectAll(); e != nil {
			return e
		}
		var e error
		beforeSelection, e = w.grid.Selection()
		if e != nil {
			return e
		}
		return w.queryWork(virtual.GridQuery{Sort: []virtual.GridSortColumn{{ColumnID: column(1), Kind: virtual.GridSortInteger, Descending: true}}, Filters: []virtual.GridFilter{{ColumnID: column(2), Kind: virtual.GridFilterContains, Text: "7"}}})
	}); e != nil {
		return e
	}
	if e := wait(ctx, p, func() (bool, error) { return w.queryTask == nil && w.view.Revision() > 2, nil }); e != nil {
		return e
	}
	if _, e := inspect(ctx, w); e != nil {
		return e
	}
	if e := owned(p, func() error {
		state, e := w.grid.State()
		if e != nil {
			return e
		}
		if w.domain != initial || state.Selection.CapturedRows != initial.rows {
			return errors.New("loadedgrid smoke: query changed domain/captured selection")
		}
		return w.queryWork(virtual.GridQuery{})
	}); e != nil {
		return e
	}
	if e := wait(ctx, p, func() (bool, error) { return w.queryTask == nil && w.view.RowCount() == initial.rows, nil }); e != nil {
		return e
	}
	if ok, e := beforeSelection.Contains(key(0)); e != nil || !ok {
		return errors.New("loadedgrid smoke: immutable captured selection lost")
	}
	var projectionStats virtual.GridProjectionStats
	if e := owned(p, func() error { projectionStats = w.view.Stats(); return nil }); e != nil {
		return e
	}
	if e := emit("worker-projection", projectionStats); e != nil {
		return e
	}
	var baseline int64
	var original *fullSource
	if e := owned(p, func() error {
		if e := w.grid.JumpToKey(key(0), column(1)); e != nil {
			return e
		}
		if e := w.begin(); e != nil {
			return e
		}
		original = w.domain
		var e error
		baseline, _, e = original.value(key(0), column(1))
		if e != nil {
			return e
		}
		if e = w.grid.SetEditText("-12345"); e != nil {
			return e
		}
		w.report(nil)
		return nil
	}); e != nil {
		return e
	}
	if _, e := inspect(ctx, w); e != nil {
		return e
	}
	if e := owned(p, w.save); e != nil {
		return e
	}
	if e := wait(ctx, p, func() (bool, error) {
		ed, e := w.grid.Edit()
		if e != nil {
			return false, e
		}
		if w.settling || w.editTask != nil {
			return false, nil
		}
		v, version, e := w.domain.value(key(0), column(1))
		if e != nil {
			return false, e
		}
		if ed.Active || v != -12345 || version != 2 {
			return false, fmt.Errorf("loadedgrid smoke: async save: %+v value%d version%d status%s", ed, v, version, w.message)
		}
		return true, nil
	}); e != nil {
		return e
	}
	if _, e := inspect(ctx, w); e != nil {
		return e
	}
	old, version, e := original.value(key(0), column(1))
	if e != nil || old != baseline || version != 1 {
		return errors.New("loadedgrid smoke: worker mutated old full domain")
	}
	if e = owned(p, func() error {
		if e := w.begin(); e != nil {
			return e
		}
		if e := w.grid.SetEditText("42"); e != nil {
			return e
		}
		return w.external()
	}); e != nil {
		return e
	}
	if e = wait(ctx, p, func() (bool, error) { return w.editTask == nil, nil }); e != nil {
		return e
	}
	if e = owned(p, func() error {
		ed, e := w.grid.Edit()
		if e != nil || ed.Status != virtual.GridEditConflict {
			return errors.New("loadedgrid smoke: external update discarded draft conflict")
		}
		if e = w.editWork(ed); !errors.Is(e, virtual.ErrGridEditConflict) {
			return fmt.Errorf("loadedgrid smoke: conflict save returned%v", e)
		}
		return w.cancelEdit()
	}); e != nil {
		return e
	}
	if _, e = inspect(ctx, w); e != nil {
		return e
	}
	var versions struct{ Domain, View uint64 }
	if e = owned(p, func() error { versions.Domain, versions.View = w.domain.Revision(), w.view.Revision(); return nil }); e != nil {
		return e
	}
	if e = emit("async-edit-conflict", versions); e != nil {
		return e
	}
	if e = owned(p, func() error {
		if e := p.owner.CloseWindow(p.windows[0].id); e != nil {
			return e
		}
		return p.owner.CloseWindow(w.id)
	}); e != nil {
		return e
	}
	return emit("pass", "Cocoa owner commands, worker preparation, native settlement, measured frames and runtime exit; no physical input/performance claim")
}
func (p *program) baseSize() uint64 {
	if p.base == nil {
		return 0
	}
	return uint64(len(p.base.Records)) * 160
}
