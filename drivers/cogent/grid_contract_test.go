//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/tree"
	"errors"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"image"
	"math"
	"runtime"
	"testing"
)

func gridContractOwner(t *testing.T, capacity int) (*app.Application, *core.Frame) {
	t.Helper()
	a, e := app.New(app.Config{QueueCapacity: capacity})
	if e != nil {
		t.Fatal(e)
	}
	if e = a.Open(app.Window{ID: "one", Title: "Grid", Width: 860, Height: 620, View: func() ui.Node { return label("stub") }}); e != nil {
		t.Fatal(e)
	}
	host := core.NewFrame(core.NewBody("grid contracts"))
	t.Cleanup(func() { host.Delete(); a.Close() })
	return a, host
}
func mountGridContract(t *testing.T, a *app.Application, host *core.Frame, n ui.Node) *record {
	t.Helper()
	var r *record
	var e error
	if err := a.Do(func() { r, e = reconcile(host, nil, n, a, "one") }); err != nil {
		t.Fatal(err)
	}
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { disposeControl(r) })
	return r
}
func TestManagedGridRemountRollbackAndCrossWindowRefusal(t *testing.T) {
	a, host := gridContractOwner(t, 8)
	g, s := testNativeGrid(t, 10000)
	old := mountGridContract(t, a, host, ui.Node{ID: "old", Kind: ui.DataGrid, Grid: g})
	binding := old.control.grid
	var next *record
	var e error
	s.failCell = true
	if err := a.Do(func() { next, e = reconcile(host, old, ui.Node{ID: "new", Kind: ui.DataGrid, Grid: g}, a, "one") }); err != nil {
		t.Fatal(err)
	}
	if e == nil || next != old || old.control.retired || gridRegistry.models[g] != binding || len(host.Children) != 1 {
		t.Fatal("failed prepare poisoned existing binding", e)
	}
	s.failCell = false
	if err := a.Open(app.Window{ID: "two", Title: "Other", Width: 400, Height: 300, View: func() ui.Node { return label("other") }}); err != nil {
		t.Fatal(err)
	}
	otherHost := core.NewFrame(core.NewBody("other"))
	defer otherHost.Delete()
	if err := a.Do(func() { _, e = reconcile(otherHost, nil, ui.Node{ID: "peer", Kind: ui.DataGrid, Grid: g}, a, "two") }); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(e, ErrGridBinding) || gridRegistry.models[g] != binding {
		t.Fatal("cross-window duplicate admitted", e)
	}
	if err := a.Do(func() { next, e = reconcile(host, old, ui.Node{ID: "new", Kind: ui.DataGrid, Grid: g}, a, "one") }); err != nil {
		t.Fatal(err)
	}
	if e != nil || next == old || !old.control.retired || !binding.closed.Load() || next.control.grid.incarnation == binding.incarnation {
		t.Fatal("remount did not retire old incarnation", e)
	}
	defer disposeControl(next)
	if tree.IsNil(next.widget) || next.control.grid.previous != nil || gridRegistry.models[g] != next.control.grid {
		t.Fatal("new binding not sole registration")
	}
}
func TestManagedGridNativeBudgetRefusedBeforeAllocation(t *testing.T) {
	a, host := gridContractOwner(t, 8)
	for _, c := range []virtual.GridConfig{{MaxRows: 257}, {MaxColumns: 65}, {MaxCells: 1 << 20}, {MaxTextBytes: 9 << 20}} {
		c.RowHeight = 24
		s := &nativeGridSource{rows: 0, revision: 1}
		g, e := virtual.NewGrid(s, []virtual.GridColumn{{ID: "c", Width: 80}}, c, virtual.GridViewport{Width: 80, Height: 80})
		if e != nil {
			t.Fatal(e)
		}
		var r *record
		if err := a.Do(func() { r, e = reconcile(host, nil, ui.Node{ID: "bad", Kind: ui.DataGrid, Grid: g}, a, "one") }); err != nil {
			t.Fatal(err)
		}
		if !errors.Is(e, ErrGridBudget) || r != nil || len(host.Children) != 0 || gridRegistry.models[g] != nil || s.cells != 0 || s.keys != 0 {
			t.Fatal("overbudget allocated or read source", e)
		}
	}
}
func TestManagedGridQueueLifetimeAndScopeRefusal(t *testing.T) {
	a, host := gridContractOwner(t, 1)
	g, _ := testNativeGrid(t, 10000)
	changes, failures := 0, 0
	r := mountGridContract(t, a, host, ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, GridChange: func(ui.GridEvent) { changes++ }})
	b := r.control.grid
	r.control.onError = func(e error) {
		if !errors.Is(e, app.ErrQueueFull) {
			t.Error(e)
		}
		failures++
	}
	before, _ := g.State()
	if e := a.PostWindow("one", func() {}); e != nil {
		t.Fatal(e)
	}
	b.post(ui.GridSelection, func() error { return g.Move(1, 0) })
	if e := a.Drain(); e != nil {
		t.Fatal(e)
	}
	if e := a.Do(func() {
		if e := syncControl(r, r.node); e != nil {
			t.Error(e)
		}
	}); e != nil {
		t.Fatal(e)
	}
	after, _ := g.State()
	if before != after || changes != 0 || failures != 1 {
		t.Fatal("queue refusal changed model or hid error")
	}
	for _, mode := range []string{"disabled", "modal"} {
		if e := a.Do(func() { r.control.disabled = mode == "disabled"; b.scope = mode != "modal" }); e != nil {
			t.Fatal(e)
		}
		b.post(ui.GridSelection, func() error { return g.Move(1, 0) })
		if e := a.Drain(); e != nil {
			t.Fatal(e)
		}
		after, _ = g.State()
		if before != after || changes != 0 {
			t.Fatal("inactive input accepted", mode)
		}
	}
	b.requestViewport(gridRequest{320, 190, 1})
	if e := a.Drain(); e != nil {
		t.Fatal(e)
	}
	after, _ = g.State()
	if after.Viewport.Width != 320 || after.Viewport.Height != 190 || changes != 1 {
		t.Fatal("hidden layout request lost")
	}
	if e := a.Open(app.Window{ID: "two", Title: "Other", Width: 400, Height: 300, View: func() ui.Node { return label("other") }}); e != nil {
		t.Fatal(e)
	}
	if e := a.Do(func() { r.control.disabled = false; b.scope = true }); e != nil {
		t.Fatal(e)
	}
	b.post(ui.GridSelection, func() error { return g.Move(1, 0) })
	if e := a.CloseWindow("one"); e != nil {
		t.Fatal(e)
	}
	if e := a.Drain(); e != nil {
		t.Fatal(e)
	}
	closed, _ := g.State()
	if closed != after || changes != 1 {
		t.Fatal("logically closed delivery mutated model")
	}
	var inspectionErr error
	if e := a.Do(func() { _, inspectionErr = InspectGrid(a, g) }); e != nil {
		t.Fatal(e)
	}
	if !errors.Is(inspectionErr, ErrGridClosed) {
		t.Fatal(inspectionErr)
	}
	disposeControl(r)
	if gridRegistry.models[g] != nil || b.measurement.Load() != nil {
		t.Fatal("closed binding retained")
	}
}
func TestManagedGridViewportCapacityRefusalDoesNotLoop(t *testing.T) {
	a, host := gridContractOwner(t, 8)
	s := &nativeGridSource{rows: 10000, revision: 1}
	g, e := virtual.NewGrid(s, []virtual.GridColumn{{ID: "c", Width: 80}}, virtual.GridConfig{RowHeight: 24, MaxCells: 4}, virtual.GridViewport{Width: 80, Height: 48})
	if e != nil {
		t.Fatal(e)
	}
	failures := 0
	r := mountGridContract(t, a, host, ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g})
	r.control.onError = func(e error) {
		if !errors.Is(e, virtual.ErrGridCapacity) {
			t.Error(e)
		}
		failures++
	}
	before, _ := g.State()
	b := r.control.grid
	for range 5 {
		b.requestViewport(gridRequest{800, 600, 1})
		if e := a.Drain(); e != nil {
			t.Fatal(e)
		}
	}
	after, _ := g.State()
	if after != before || failures != 1 {
		t.Fatal("capacity refusal changed viewport or repeated work", failures)
	}
	b.requestViewport(gridRequest{80, 72, 1})
	if e := a.Drain(); e != nil {
		t.Fatal(e)
	}
	after, _ = g.State()
	if after.Viewport.Height != 72 {
		t.Fatal("valid request could not recover")
	}
}
func TestManagedGridNumericAdmission(t *testing.T) {
	for _, v := range []float32{float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))} {
		if normalizedGridOffset(v, 100) != -1 {
			t.Fatal("nonfinite scrollbar admitted")
		}
	}
	if normalizedGridOffset(-1, 100) != 0 || normalizedGridOffset(2, 100) != 100 || normalizedGridOffset(.5, 100) != 50 {
		t.Fatal("scrollbar clamp")
	}
	if boundedWheel(.9) != 0 || boundedWheel(-.9) != 0 || boundedWheel(1.2) != 1 || boundedWheel(1e300) != (1<<20) {
		t.Fatal("wheel bound")
	}
}

func TestManagedGridLocalGeometryRefusal(t *testing.T) {
	for _, dpi := range []float32{96, 192} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			a, host := gridContractOwner(t, 8)
			for _, shape := range []string{"width", "height", "header", "overscan"} {
				source := &nativeGridSource{rows: 2, revision: 1}
				config := virtual.GridConfig{RowHeight: 24}
				width := 80
				switch shape {
				case "width":
					width = math.MaxInt
				case "height":
					config.RowHeight = math.MaxInt / 2
				case "header":
					config.HeaderHeight = math.MaxInt / 2
				case "overscan":
					config.RowHeight = gridNativeMaxDimension
					config.OverscanRows = 6
					source.rows = 1000
				}
				viewport := virtual.GridViewport{Width: 80, Height: 48}
				if shape == "overscan" {
					viewport.ScrollY = 500 * config.RowHeight
				}
				g, e := virtual.NewGrid(source, []virtual.GridColumn{{ID: "c", Width: width}}, config, viewport)
				if e != nil {
					t.Fatal("portable-valid shape refused", shape, e)
				}
				var r *record
				host.Styles.UnitContext.DPI = dpi
				if e := a.Do(func() { r, e = reconcile(host, nil, ui.Node{ID: "bad", Kind: ui.DataGrid, Grid: g}, a, "one") }); e != nil {
					t.Fatal(e)
				}
				if !errors.Is(e, ErrGridBudget) || r != nil || len(host.Children) != 0 || gridRegistry.models[g] != nil {
					t.Fatal("local native geometry allocated", shape, e)
				}
			}
		})
	}
}

func TestManagedGridShrinkingProjectionReleasesSurplusWidgets(t *testing.T) {
	a, host := gridContractOwner(t, 8)
	g, _ := testNativeGrid(t, 10000)
	r := mountGridContract(t, a, host, ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g})
	binding := r.control.grid
	old := append([]*core.Text(nil), binding.cells...)
	if len(old) < 2 {
		t.Fatal("fixture not materialized")
	}
	if e := a.Do(func() {
		state, e := g.State()
		if e != nil {
			t.Error(e)
			return
		}
		v := state.Viewport
		v.Height = 24
		v.Width = 180
		if e = g.SetViewport(v); e != nil {
			t.Error(e)
			return
		}
		if e = syncGrid(r); e != nil {
			t.Error(e)
		}
	}); e != nil {
		t.Fatal(e)
	}
	if len(binding.cells) >= len(old) {
		t.Fatal("fixture did not shrink")
	}
	for _, w := range old[len(binding.cells):] {
		if !tree.IsNil(w) {
			t.Fatal("surplus cache widget retained")
		}
	}
	for _, w := range binding.cells[len(binding.cells):cap(binding.cells)] {
		if w != nil {
			t.Fatal("backing array retained retired widget")
		}
	}
}

func TestManagedGridViewportAndTinyScaleGeometryRefusal(t *testing.T) {
	a, host := gridContractOwner(t, 8)
	for _, rows := range []int{0, 1} {
		s := &nativeGridSource{rows: rows, revision: 1}
		g, e := virtual.NewGrid(s, []virtual.GridColumn{{ID: "c", Width: 80}}, virtual.GridConfig{RowHeight: 24}, virtual.GridViewport{Width: math.MaxInt, Height: math.MaxInt})
		if e != nil {
			t.Fatal(e)
		}
		var r *record
		if err := a.Do(func() { r, e = reconcile(host, nil, ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g}, a, "one") }); err != nil {
			t.Fatal(err)
		}
		if !errors.Is(e, ErrGridBudget) || r != nil || len(host.Children) != 0 || s.cells != 0 {
			t.Fatal("giant sparse viewport allocated", rows, e)
		}
	}
	for _, scale := range []float32{math.SmallestNonzeroFloat32, 0, 33, float32(math.NaN()), float32(math.Inf(1))} {
		if _, e := gridViewportRequest(image.Rect(0, 0, 860, 620), scale); !errors.Is(e, ErrGridBudget) {
			t.Fatal("unsafe scale converted", scale, e)
		}
	}
	for _, scale := range []float32{1, 2} {
		request, e := gridViewportRequest(image.Rect(0, 0, 860, 620), scale)
		if e != nil || request.width != int(860/scale)-gridScrollbar || request.height != int(620/scale)-gridScrollbar {
			t.Fatal("valid measured conversion refused", request, e)
		}
	}
	g, _ := testNativeGrid(t, 10000)
	r := mountGridContract(t, a, host, ui.Node{ID: "valid", Kind: ui.DataGrid, Grid: g})
	binding := r.control.grid
	failures := 0
	r.control.onError = func(e error) {
		if !errors.Is(e, ErrGridBudget) {
			t.Error(e)
		}
		failures++
	}
	old, _ := g.State()
	binding.host.Styles.UnitContext.DPI = math.SmallestNonzeroFloat32
	binding.host.Geom.ContentBBox = image.Rect(0, 0, 860, 620)
	for range 5 {
		binding.capture()
		if e := a.Drain(); e != nil {
			t.Fatal(e)
		}
	}
	now, _ := g.State()
	if now != old || failures != 1 || binding.measurement.Load() != nil {
		t.Fatal("invalid measured geometry mutated/looped", failures)
	}
}

type gridGoexitSource struct {
	nativeGridSource
	exit bool
}

func (s *gridGoexitSource) Cell(row string, column string) (string, error) {
	if s.exit {
		runtime.Goexit()
	}
	return s.nativeGridSource.Cell(row, column)
}
func TestManagedGridMountGoexitRetiresRegistry(t *testing.T) {
	a, host := gridContractOwner(t, 8)
	source := &gridGoexitSource{nativeGridSource: nativeGridSource{rows: 10000, revision: 1}}
	g, err := virtual.NewGrid(source, []virtual.GridColumn{{ID: "c", Width: 80}}, virtual.GridConfig{RowHeight: 24}, virtual.GridViewport{Width: 80, Height: 80})
	if err != nil {
		t.Fatal(err)
	}
	before, err := g.State()
	if err != nil {
		t.Fatal(err)
	}
	source.exit = true
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = a.Do(func() { _, _ = reconcile(host, nil, ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g}, a, "one") })
	}()
	<-done
	gridRegistry.Lock()
	leaked := gridRegistry.models[g]
	gridRegistry.Unlock()
	if leaked != nil {
		disposeGrid(leaked.r)
		t.Fatal("Goexit retained unfinished Grid registry claim")
	}
	if a.Context().Err() == nil || len(host.Children) != 0 {
		t.Fatal("Goexit did not unwind owner and unfinished widgets")
	}
	source.exit = false
	after, err := g.State()
	if err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatal("mount failure changed caller-owned Grid state")
	}
	fresh, freshHost := gridContractOwner(t, 8)
	r := mountGridContract(t, fresh, freshHost, ui.Node{ID: "fresh", Kind: ui.DataGrid, Grid: g})
	if r.control.grid == nil {
		t.Fatal("fresh application could not remount Grid after Goexit")
	}
}
