package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/tasks"
	"github.com/dshills/vellum/virtual"
)

var bg = context.Background()

func drain(t *testing.T, p *program, batch int) {
	t.Helper()
	for i := 0; i < 1024; i++ {
		pending := p.owner.Pending()
		if pending.Ordinary+pending.ReadyCompletions == 0 {
			return
		}
		if _, e := p.owner.DrainN(batch); e != nil {
			t.Fatal(e)
		}
	}
	t.Fatal("owner drain bound")
}
func runWorker(t *testing.T, ex *execution.Manual, id execution.ID) {
	t.Helper()
	if e := ex.Step(bg, id); e != nil {
		t.Fatal(e)
	}
}
func fixture(t *testing.T, n int) (*program, *execution.Manual) {
	t.Helper()
	ex, e := execution.NewManual(8)
	if e != nil {
		t.Fatal(e)
	}
	p, e := newProgram(n, ex)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		p.owner.Close()
		ex.Close()
		for _, id := range ex.PendingIDs() {
			runWorker(t, ex, id)
		}
	})
	return p, ex
}
func loadFixture(t *testing.T, p *program, ex *execution.Manual, batch int) {
	t.Helper()
	if e := owned(p, p.load); e != nil {
		t.Fatal(e)
	}
	runWorker(t, ex, ex.PendingIDs()[0])
	drain(t, p, batch)
	ids := ex.PendingIDs()
	if len(ids) != 2 {
		t.Fatal("initial index workers", ids)
	}
	for _, id := range ids {
		runWorker(t, ex, id)
	}
	drain(t, p, batch)
}
func state(t *testing.T, w *window) virtual.OutlineState {
	t.Helper()
	var s virtual.OutlineState
	if e := owned(w.p, func() error { var e error; s, e = w.outline.State(); return e }); e != nil {
		t.Fatal(e)
	}
	return s
}
func TestStoredTreeDataHugeBranchDeepChain(t *testing.T) {
	for _, n := range []int{1, 10_000, 100_000} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			d, e := makeData(bg, "fixture", n)
			if e != nil {
				t.Fatal(e)
			}
			if d.Len() != n || d.Stats().StorageBytes <= 0 {
				t.Fatal(d.Stats())
			}
			index, e := virtual.BuildTreeIndex(bg, d, virtual.TreeIndexOptions{ExpandAll: true})
			if e != nil {
				t.Fatal(e)
			}
			if index.Len() != n || index.Stats().Operations == 0 {
				t.Fatal(index.Stats())
			}
			last, e := d.Node(nodeKey(n - 1))
			if e != nil || last.ID != nodeKey(n-1) {
				t.Fatal(last, e)
			}
			if n >= 4 && d.Stats().Depth != n-n*3/4-1 {
				t.Fatal("stored chain depth", d.Stats())
			}
		})
	}
	ctx, cancel := context.WithCancel(bg)
	cancel()
	if _, e := makeData(ctx, "canceled", 100_000); !errors.Is(e, context.Canceled) {
		t.Fatal("canceled storage accepted", e)
	}
}
func TestWorkerLatestCancellationHiddenSelection(t *testing.T) {
	for _, batch := range []int{1, 7, 64} {
		t.Run(fmt.Sprintf("batch-%d", batch), func(t *testing.T) {
			p, ex := fixture(t, 128)
			loadFixture(t, p, ex, batch)
			w := p.windows[1]
			if e := owned(p, func() error { return w.change(virtual.OutlineChange{Kind: virtual.OutlineExpandAll}) }); e != nil {
				t.Fatal(e)
			}
			old := ex.PendingIDs()[0]
			if e := owned(p, func() error { return w.change(virtual.OutlineChange{Kind: virtual.OutlineCollapseAll}) }); e != nil {
				t.Fatal(e)
			}
			newest := ex.PendingIDs()[1]
			if e := owned(p, func() error { return w.change(virtual.OutlineChange{Kind: virtual.OutlineExpandAll}) }); !errors.Is(e, tasks.ErrCapacity) {
				t.Fatal("capacity should preserve latest request", e)
			}
			runWorker(t, ex, newest)
			drain(t, p, batch)
			runWorker(t, ex, old)
			drain(t, p, batch)
			if s := state(t, w); s.VisibleRows != 2 || s.Pending || !w.manager.Pending().Idle() {
				t.Fatal("old canceled result installed or credit stranded", s, w.manager.Pending())
			}
			if e := owned(p, func() error { return w.change(virtual.OutlineChange{Kind: virtual.OutlineExpandAll}) }); e != nil {
				t.Fatal(e)
			}
			runWorker(t, ex, ex.PendingIDs()[0])
			drain(t, p, batch)
			if e := owned(p, func() error {
				if e := w.outline.End(); e != nil {
					return e
				}
				return w.change(virtual.OutlineChange{Kind: virtual.OutlineCollapseAll})
			}); e != nil {
				t.Fatal(e)
			}
			runWorker(t, ex, ex.PendingIDs()[0])
			drain(t, p, batch)
			if s := state(t, w); s.Selected != nodeKey(127) || s.SelectedVisible {
				t.Fatal("hidden selection lost", s)
			}
			if e := owned(p, func() error { return w.change(virtual.OutlineChange{Kind: virtual.OutlineFilter, Filter: "000127"}) }); e != nil {
				t.Fatal(e)
			}
			runWorker(t, ex, ex.PendingIDs()[0])
			drain(t, p, batch)
			if s := state(t, w); s.Selected != nodeKey(127) || !s.SelectedVisible || s.VisibleRows != 32 {
				t.Fatal("filter ancestors", s)
			}
			if e := owned(p, func() error { return w.outline.Reveal(nodeKey(127)) }); e != nil {
				t.Fatal(e)
			}
			if e := owned(p, func() error { return w.cancel() }); e != nil {
				t.Fatal(e)
			}
		})
	}
}
func TestCanceledLoadAndWindowResultRefusal(t *testing.T) {
	p, ex := fixture(t, 32)
	if e := owned(p, p.load); e != nil {
		t.Fatal(e)
	}
	old := ex.PendingIDs()[0]
	if e := owned(p, p.cancelLoad); e != nil {
		t.Fatal(e)
	}
	if e := owned(p, p.load); !errors.Is(e, tasks.ErrCapacity) {
		t.Fatal("accepted canceled load released worker credit", e)
	}
	runWorker(t, ex, old)
	drain(t, p, 7)
	if p.base[1] != nil {
		t.Fatal("canceled stored data installed")
	}
	loadFixture(t, p, ex, 7)
	w := p.windows[0]
	if e := owned(p, func() error { return w.change(virtual.OutlineChange{Kind: virtual.OutlineExpandAll}) }); e != nil {
		t.Fatal(e)
	}
	id := ex.PendingIDs()[0]
	if e := owned(p, func() error { return p.owner.CloseWindow(w.id) }); e != nil {
		t.Fatal(e)
	}
	if w.manager.Pending().Idle() {
		t.Fatal("window close hid accepted pending worker")
	}
	runWorker(t, ex, id)
	if !w.manager.Pending().Idle() {
		t.Fatal("closed worker terminal did not refund")
	}
	if s := state(t, p.windows[1]); s.Stamp.Nodes != 32 {
		t.Fatal("survivor changed", s)
	}
}

func TestIndexCommitRetainsCurrentCursorAndViewport(t *testing.T) {
	p, ex := fixture(t, 128)
	loadFixture(t, p, ex, 7)
	w := p.windows[1]
	if e := owned(p, func() error { return w.change(virtual.OutlineChange{Kind: virtual.OutlineExpandAll}) }); e != nil {
		t.Fatal(e)
	}
	runWorker(t, ex, ex.PendingIDs()[0])
	drain(t, p, 7)
	if e := owned(p, func() error {
		if e := w.change(virtual.OutlineChange{Kind: virtual.OutlineKeep}); e != nil {
			return e
		}
		return w.outline.End()
	}); e != nil {
		t.Fatal(e)
	}
	before := state(t, w)
	runWorker(t, ex, ex.PendingIDs()[0])
	drain(t, p, 7)
	after := state(t, w)
	if after.Selected != before.Selected || after.Viewport != before.Viewport || after.Pending {
		t.Fatal("worker-start state overwrote current cursor", before, after)
	}
}
func TestConsumerDomainCASAndHeldFilterCancellation(t *testing.T) {
	p, ex := fixture(t, 128)
	loadFixture(t, p, ex, 7)
	w := p.windows[1]
	if e := owned(p, func() error { return w.change(virtual.OutlineChange{Kind: virtual.OutlineExpandAll}) }); e != nil {
		t.Fatal(e)
	}
	id := ex.PendingIDs()[0]
	external, e := makeData(bg, "remote-incarnation", 128)
	if e != nil {
		t.Fatal(e)
	}
	if e = owned(p, func() error { w.data = external; return nil }); e != nil {
		t.Fatal(e)
	}
	runWorker(t, ex, id)
	drain(t, p, 7)
	if s := state(t, w); s.Stamp.Identity == external.Stamp().Identity || s.Pending || w.task != nil || !w.manager.Pending().Idle() {
		t.Fatal("domain-CAS refusal mutated accepted view or stranded credit", s)
	}
	var held func(error)
	var ctx context.Context
	if e = owned(p, func() error {
		w.settleFilter = func(c context.Context, cb func(error)) error { ctx = c; held = cb; return nil }
		return w.applyFilter()
	}); e != nil {
		t.Fatal(e)
	}
	if e = owned(p, func() error { return w.change(virtual.OutlineChange{Kind: virtual.OutlineExpandAll}) }); !errors.Is(e, errBusy) {
		t.Fatal("change must refuse while native filter settling", e)
	}
	if e = owned(p, w.cancel); e != nil {
		t.Fatal(e)
	}
	if ctx.Err() == nil {
		t.Fatal("cancel left native settlement lifetime alive")
	}
	if e = owned(p, func() error { held(context.Canceled); return nil }); e != nil {
		t.Fatal(e)
	}
	if w.settling || w.filterCancel != nil || len(ex.PendingIDs()) != 0 {
		t.Fatal("canceled settlement started stale work")
	}
}

func TestCompleteRefusalCancelAndRequestRetirement(t *testing.T) {
	p, ex := fixture(t, 128)
	loadFixture(t, p, ex, 7)
	w := p.windows[1]
	if w.request != (virtual.OutlineRequest{}) || p.windows[0].request != (virtual.OutlineRequest{}) {
		t.Fatal("accepted requests retained prior indexes")
	}
	index, e := virtual.BuildTreeIndex(bg, w.data, virtual.TreeIndexOptions{})
	if e != nil {
		t.Fatal(e)
	}
	constrained, e := virtual.NewOutline(index, virtual.OutlineConfig{RowHeight: 1, MaxRows: 16}, virtual.OutlineViewport{Width: 200, Height: 200})
	if e != nil {
		t.Fatal(e)
	}
	if e = owned(p, func() error {
		w.outline = constrained
		return w.change(virtual.OutlineChange{Kind: virtual.OutlineExpandAll})
	}); e != nil {
		t.Fatal(e)
	}
	runWorker(t, ex, ex.PendingIDs()[0])
	drain(t, p, 7)
	if s := state(t, w); !s.Pending || w.task != nil || !w.manager.Pending().Idle() || w.request == (virtual.OutlineRequest{}) {
		t.Fatal("Complete capacity-refusal oracle", s, w.message)
	}
	if e = owned(p, w.cancel); e != nil {
		t.Fatal(e)
	}
	if state(t, w).Pending || w.request != (virtual.OutlineRequest{}) {
		t.Fatal("Cancel after terminal Complete refusal stranded pending/request")
	}
	if e = owned(p, func() error { return w.change(virtual.OutlineChange{Kind: virtual.OutlineKeep}) }); e != nil {
		t.Fatal(e)
	}
	runWorker(t, ex, ex.PendingIDs()[0])
	drain(t, p, 7)
	if w.request != (virtual.OutlineRequest{}) {
		t.Fatal("accepted keep request retained prior index")
	}
	if e = owned(p, func() error { return w.change(virtual.OutlineChange{Kind: virtual.OutlineKeep}) }); e != nil {
		t.Fatal(e)
	}
	task := w.task
	task.Cancel()
	runWorker(t, ex, ex.PendingIDs()[0])
	drain(t, p, 7)
	if state(t, w).Pending || w.request != (virtual.OutlineRequest{}) || w.task != nil {
		t.Fatal("worker error retained request/index")
	}
	if e = owned(p, func() error { return w.change(virtual.OutlineChange{Kind: virtual.OutlineKeep}) }); e != nil {
		t.Fatal(e)
	}
	id := ex.PendingIDs()[0]
	if e = owned(p, w.cancel); e != nil {
		t.Fatal(e)
	}
	if state(t, w).Pending || w.request != (virtual.OutlineRequest{}) {
		t.Fatal("canceled request retained previous data")
	}
	runWorker(t, ex, id)
	drain(t, p, 7)
}

func TestCancelPublishedLoaderRejectsOlderToken(t *testing.T) {
	for _, batch := range []int{1, 7, 64} {
		t.Run(fmt.Sprintf("batch-%d", batch), func(t *testing.T) {
			p, ex := fixture(t, 128)
			if e := owned(p, p.load); e != nil {
				t.Fatal(e)
			}
			old := ex.PendingIDs()[0]
			runWorker(t, ex, old)
			if p.loader.Pending().Active != 1 {
				t.Fatal("owner delivery was not retained")
			}
			if e := owned(p, p.cancelLoad); e != nil {
				t.Fatal(e)
			}
			drain(t, p, batch)
			if !p.loader.Pending().Idle() || len(ex.PendingIDs()) != 0 {
				t.Fatal("published canceled load kept credit or admitted index jobs")
			}
			if e := owned(p, func() error {
				if p.loadTask != nil || p.base[0] != nil || p.base[1] != nil {
					t.Fatal("old token published stored data")
				}
				return nil
			}); e != nil {
				t.Fatal(e)
			}
			for _, w := range p.windows {
				if s := state(t, w); s.Stamp.Nodes != 0 || s.Pending {
					t.Fatal("old canceled loader changed outline", s)
				}
			}
			loadFixture(t, p, ex, batch)
			if s := state(t, p.windows[1]); s.Stamp.Nodes != 128 || s.Pending {
				t.Fatal("retry after terminal cleanup failed", s)
			}
		})
	}
}

func TestFailedPreparationCannotRetireNewerAdmission(t *testing.T) {
	for _, batch := range []int{1, 7, 64} {
		t.Run(fmt.Sprintf("batch-%d", batch), func(t *testing.T) {
			p, ex := fixture(t, 128)
			loadFixture(t, p, ex, batch)
			w := p.windows[1]
			if e := owned(p, func() error {
				return w.change(virtual.OutlineChange{Kind: virtual.OutlineFilter, Filter: strings.Repeat("x", 4097)})
			}); e == nil {
				t.Fatal("oversized request preparation accepted")
			}
			failed := ex.PendingIDs()[0]
			if e := owned(p, func() error { return w.change(virtual.OutlineChange{Kind: virtual.OutlineKeep}) }); e != nil {
				t.Fatal(e)
			}
			ids := ex.PendingIDs()
			if len(ids) != 2 {
				t.Fatal("expected failed and latest accepted workers", ids)
			}
			var latest *tasks.Task
			var req virtual.OutlineRequest
			if e := owned(p, func() error { latest = w.task; req = w.request; return nil }); e != nil {
				t.Fatal(e)
			}
			runWorker(t, ex, failed)
			drain(t, p, batch)
			if e := owned(p, func() error {
				s, e := w.outline.State()
				if e != nil {
					return e
				}
				if w.task != latest || w.request != req || !s.Pending {
					t.Fatal("failed-preparation cleanup retired newer pending request", s, w.task, w.request)
				}
				return w.cancel()
			}); e != nil {
				t.Fatal(e)
			}
			if s := state(t, w); s.Pending {
				t.Fatal("Cancel lost authority for newest request", s)
			}
			runWorker(t, ex, ids[1])
			drain(t, p, batch)
			if !w.manager.Pending().Idle() {
				t.Fatal("dual terminal credits stranded after canceled newest worker")
			}
		})
	}
}
