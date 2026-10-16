package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/virtual"
)

func owned(p *program, fn func() error) error {
	var e error
	if err := p.owner.Do(func() { e = fn() }); err != nil {
		return err
	}
	return e
}
func emit(stage string, data any) error {
	b, e := json.Marshal(struct {
		Stage string `json:"stage"`
		Data  any    `json:"data"`
	}{stage, data})
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
		done := false
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
func inspect(ctx context.Context, w *window) (cogent.TreeInspection, error) {
	var out cogent.TreeInspection
	e := wait(ctx, w.p, func() (bool, error) {
		var e error
		out, e = cogent.InspectTree(w.p.owner, w.outline)
		if errors.Is(e, cogent.ErrTreeClosed) || errors.Is(e, cogent.ErrTreeLayout) {
			return false, nil
		}
		if e != nil {
			return false, e
		}
		s, e := w.outline.State()
		if e != nil {
			return false, e
		}
		return out.Frames > 0 && !out.Bounds.Empty() && out.Generation == s.Generation && out.Stamp == s.Stamp && out.Pending == s.Pending, nil
	})
	return out, e
}
func workerIdle(ctx context.Context, w *window) error {
	return wait(ctx, w.p, func() (bool, error) {
		s, e := w.outline.State()
		return e == nil && !s.Pending && w.task == nil && w.manager.Pending().Idle(), e
	})
}
func desktopSmoke(ctx context.Context, p *program) error {
	for _, w := range p.windows {
		v, e := inspect(ctx, w)
		if e != nil {
			return e
		}
		if v.TotalNodes != 0 {
			return errors.New("loadedtree: initial model not empty")
		}
	}
	if e := emit("empty-ready", p.nodes); e != nil {
		return e
	}
	if e := owned(p, p.load); e != nil {
		return e
	}
	for _, w := range p.windows {
		if e := wait(ctx, p, func() (bool, error) {
			s, e := w.outline.State()
			return e == nil && s.Stamp.Nodes > 0 && w.task == nil && w.manager.Pending().Idle(), e
		}); e != nil {
			return e
		}
		if _, e := inspect(ctx, w); e != nil {
			return e
		}
	}
	var dataStats [2]virtual.TreeDataStats
	if e := owned(p, func() error {
		for i, w := range p.windows {
			dataStats[i] = w.data.Stats()
		}
		return nil
	}); e != nil {
		return e
	}
	if dataStats[1].Nodes != p.nodes {
		return errors.New("loadedtree: stored size mismatch")
	}
	if e := emit("stored-loaded", dataStats); e != nil {
		return e
	}
	w := p.windows[1]
	if e := owned(p, func() error { return w.change(virtual.OutlineChange{Kind: virtual.OutlineExpandAll}) }); e != nil {
		return e
	}
	if e := workerIdle(ctx, w); e != nil {
		return e
	}
	if e := owned(p, func() error {
		if e := w.outline.End(); e != nil {
			return e
		}
		w.report(nil)
		return nil
	}); e != nil {
		return e
	}
	expanded, e := inspect(ctx, w)
	if e != nil {
		return e
	}
	if expanded.VisibleRows != p.nodes || expanded.Selected != nodeKey(p.nodes-1) || len(expanded.Rows) > 500 {
		return errors.New("loadedtree: expanded far cursor/pool mismatch")
	}
	if e = emit("deep-chain-visible", expanded); e != nil {
		return e
	}
	selected := nodeKey(p.nodes - 1)
	if e = owned(p, func() error { return w.change(virtual.OutlineChange{Kind: virtual.OutlineCollapseAll}) }); e != nil {
		return e
	}
	if e = workerIdle(ctx, w); e != nil {
		return e
	}
	collapsed, e := inspect(ctx, w)
	if e != nil {
		return e
	}
	if collapsed.Selected != selected || p.nodes > 2 && collapsed.SelectedVisible {
		return errors.New("loadedtree: collapse lost hidden selected key")
	}
	if e = emit("hidden-selection-retained", collapsed); e != nil {
		return e
	}
	if e = owned(p, func() error {
		return w.change(virtual.OutlineChange{Kind: virtual.OutlineFilter, Filter: fmt.Sprintf("%06d", p.nodes-1)})
	}); e != nil {
		return e
	}
	if e = workerIdle(ctx, w); e != nil {
		return e
	}
	if e = owned(p, func() error {
		if e := w.outline.Reveal(selected); e != nil {
			return e
		}
		w.report(nil)
		return nil
	}); e != nil {
		return e
	}
	filtered, e := inspect(ctx, w)
	if e != nil {
		return e
	}
	if filtered.Selected != selected || !filtered.SelectedVisible {
		return errors.New("loadedtree: filtered reveal lost key")
	}
	if e = emit("filter-reveals-ancestors", filtered); e != nil {
		return e
	}
	// The ordinary filter field uses the same native FIFO settlement as user Apply.
	if e = owned(p, func() error { w.filterText = "absent-label"; w.report(nil); return nil }); e != nil {
		return e
	}
	if _, e = inspect(ctx, w); e != nil {
		return e
	}
	if e = owned(p, w.applyFilter); e != nil {
		return e
	}
	if e = wait(ctx, p, func() (bool, error) {
		s, e := w.outline.State()
		return e == nil && !w.settling && s.Filter == "absent-label" && w.task == nil && w.manager.Pending().Idle(), e
	}); e != nil {
		return e
	}
	empty, e := inspect(ctx, w)
	if e != nil {
		return e
	}
	if empty.VisibleRows != 0 || empty.Selected != selected {
		return errors.New("loadedtree: settled filter/hidden selection mismatch")
	}
	if e = emit("settled-filter-empty", empty); e != nil {
		return e
	}
	if e = owned(p, func() error {
		if e := w.change(virtual.OutlineChange{Kind: virtual.OutlineExpandAll}); e != nil {
			return e
		}
		return w.cancel()
	}); e != nil {
		return e
	}
	if e = workerIdle(ctx, w); e != nil {
		return e
	}
	if e = emit("canceled-worker-dual-terminal", nil); e != nil {
		return e
	}
	var lifetime context.Context
	if e = owned(p, func() error {
		lifetime = p.windows[0].windowRef.Context()
		return p.owner.CloseWindow(p.windows[0].windowRef)
	}); e != nil {
		return e
	}
	if lifetime.Err() == nil {
		return errors.New("loadedtree: domain close did not cancel window")
	}
	if _, e = inspect(ctx, w); e != nil {
		return e
	}
	if e = emit("survivor-independent", nil); e != nil {
		return e
	}
	if e = owned(p, func() error { return p.owner.CloseWindow(p.windows[1].windowRef) }); e != nil {
		return e
	}
	return emit("pass", map[string]any{"nodes": p.nodes, "desktop_input": "owner commands and real filter settlement", "native_teardown": "not joined or claimed"})
}
