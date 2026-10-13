//go:build offscreen

package cogent

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"reflect"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

type gridLayoutProbe struct {
	core.Text
	sizeUps int
}

func (p *gridLayoutProbe) SizeUp() { p.sizeUps++; p.Text.SizeUp() }

// Simultaneous row/column changes interleave new column slots with the incoming
// row. A flat donor queue can move frozen widgets into scrolling panes even
// though each pane's population is unchanged, forcing a full scene layout.
func TestManagedGridCombinedScrollKeepsDonorPanesAndLocalLayout(t *testing.T) {
	for _, dpi := range []float32{96, 120, 144, 192} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			columns := make([]virtual.GridColumn, 20)
			for i := range columns {
				columns[i] = virtual.GridColumn{ID: fmt.Sprintf("c%d", i), Width: 96}
			}
			model, err := virtual.NewGrid(&nativeGridSource{rows: 10000, revision: 1}, columns, virtual.GridConfig{RowHeight: 24, HeaderHeight: 24, FrozenColumns: 2}, virtual.GridViewport{Width: 480, Height: 240})
			if err != nil {
				t.Fatal(err)
			}
			config := manualConfig(t, 7)
			config.LogicalDPI = dpi
			m, a := gridManual(t, config, func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true} })
			gridFrame(t, m)
			var g *gridBinding
			var probe *gridLayoutProbe
			if err = a.Do(func() {
				window := m.implementation.find("one").state
				g = window.root.control.grid
				probe = tree.New[gridLayoutProbe](window.body)
				probe.SetText("unrelated layout probe")
				probe.Styler(func(s *styles.Style) { s.Min.Y = units.Dp(30); s.Grow.SetZero() })
				window.body.Update()
				if e := model.Scroll(192, 72); e != nil {
					t.Error(e)
				}
				_ = a.Invalidate("one")
			}); err != nil {
				t.Fatal(err)
			}
			gridFrame(t, m)
			for _, step := range []image.Point{{96, 24}, {-96, -24}, {96, -24}, {-96, 24}} {
				before := make(map[*core.Text]tree.Node)
				keys := make(map[gridCellKey]*core.Text)
				if err = a.Do(func() {
					for i, w := range g.cells {
						before[w] = w.Parent
						c := g.projection.Cells[i]
						keys[gridCellKey{c.RowKey, c.ColumnID}] = w
					}
					probe.sizeUps = 0
					if e := model.Scroll(step.X, step.Y); e != nil {
						t.Error(e)
					}
					if e := syncGrid(g.r); e != nil {
						t.Error(e)
					}
					if len(g.cells) != len(before) {
						t.Error("fixture changed total cell population")
					}
					for i, w := range g.cells {
						if parent := before[w]; parent != nil && parent != w.Parent {
							t.Errorf("recycled widget crossed frozen/scrolling pane on %v", step)
						}
						c := g.projection.Cells[i]
						if old := keys[gridCellKey{c.RowKey, c.ColumnID}]; old != nil && old != w {
							t.Error("overlapping cell lost domain-key identity")
						}
					}
					if g.host.Scene.GlobalLayoutPending() {
						t.Errorf("combined scroll %v forced global layout", step)
					}
					assertGridRetainedPaneOrder(t, g)
				}); err != nil {
					t.Fatal(err)
				}
				gridFrame(t, m)
				actual, e := m.Capture(testContext, "one")
				if e != nil {
					t.Fatal(e)
				}
				inspection := gridInspection(t, a, model)
				if err = a.Do(func() {
					if probe.sizeUps != 0 {
						t.Error("combined scroll sized unrelated scene content")
					}
					g.host.Update() // Independent ordinary full-layout oracle.
				}); err != nil {
					t.Fatal(err)
				}
				gridFrame(t, m)
				full, e := m.Capture(testContext, "one")
				if e != nil {
					t.Fatal(e)
				}
				if actual.Bounds() != full.Bounds() || !bytes.Equal(actual.Pix, full.Pix) {
					t.Fatalf("combined scroll %v local/full pixels differ", step)
				}
				oracle := gridInspection(t, a, model)
				inspection.Frames, oracle.Frames = 0, 0
				if !reflect.DeepEqual(inspection, oracle) {
					t.Fatalf("combined scroll %v local/full geometry or semantics differ", step)
				}
				if probe.sizeUps == 0 {
					t.Fatal("ordinary full oracle did not exercise unrelated layout")
				}
			}
		})
	}
}

func TestManagedGridDonorShortageFallbackAndRetirement(t *testing.T) {
	a, host := gridContractOwner(t, 8)
	model, _ := testNativeGrid(t, 10000)
	mounted := mountGridContract(t, a, host, ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model}).control.grid
	var frozen, scrolling []*core.Text
	var frozenCells, scrollingCells []virtual.GridCell
	for i, w := range mounted.cells {
		cell := mounted.projection.Cells[i]
		if cell.Frozen && len(frozen) < 2 {
			frozen = append(frozen, w)
			frozenCells = append(frozenCells, cell)
		}
		if !cell.Frozen && len(scrolling) < 2 {
			scrolling = append(scrolling, w)
			scrollingCells = append(scrollingCells, cell)
		}
	}
	if len(frozen) != 2 || len(scrolling) != 2 {
		t.Fatal("fixture lacks both donor panes")
	}
	old := append(append([]*core.Text{}, frozen...), scrolling...)
	oldCells := append(append([]virtual.GridCell{}, frozenCells...), scrollingCells...)
	makeBinding := func() *gridBinding {
		return &gridBinding{cells: old, projection: virtual.GridSnapshot{Cells: oldCells}, panes: mounted.panes}
	}
	makeCells := func(count, frozenCount int) []virtual.GridCell {
		out := make([]virtual.GridCell, count)
		for i := range out {
			out[i] = virtual.GridCell{RowKey: fmt.Sprintf("new-row%d", i), ColumnID: "new-column", Frozen: i < frozenCount}
		}
		return out
	}
	// Growth leaves two frozen holes. All old scrolling donors must first reach
	// the later scrolling holes rather than being stolen by early fallback.
	cells, retired, err := makeBinding().retainCells(virtual.GridSnapshot{Cells: makeCells(6, 4)})
	if err != nil || len(retired) != 0 || cells[2] != nil || cells[3] != nil || cells[4] != scrolling[0] || cells[5] != scrolling[1] {
		t.Fatal("preferred holes lost donors before fallback", cells, retired, err)
	}
	// A true population change still reuses every old widget, including donors
	// from the other pane, rather than temporarily growing the retained pool.
	cells, retired, err = makeBinding().retainCells(virtual.GridSnapshot{Cells: makeCells(4, 0)})
	seen := map[*core.Text]bool{}
	for _, w := range cells {
		if w == nil || seen[w] {
			t.Fatal("fallback lost or duplicated widget")
		}
		seen[w] = true
	}
	if err != nil || len(retired) != 0 || len(seen) != 4 {
		t.Fatal("fallback grew or retired reusable pool", err)
	}
	// Surplus retirement retains old order and contains each unconsumed widget
	// once. Nothing has reparented the borrowed mounted widgets during matching.
	cells, retired, err = makeBinding().retainCells(virtual.GridSnapshot{Cells: makeCells(2, 0)})
	if err != nil || !reflect.DeepEqual(cells, scrolling) || !reflect.DeepEqual(retired, frozen) {
		t.Fatal("surplus retirement changed membership/order", err)
	}
	keyed := makeCells(2, 0)
	keyed[0] = oldCells[0]
	keyed[0].Frozen = false
	cells, _, err = makeBinding().retainCells(virtual.GridSnapshot{Cells: keyed})
	if err != nil || cells[0] != old[0] {
		t.Fatal("pane preference overrode domain-key match", err)
	}
	duplicate := makeCells(2, 0)
	duplicate[1] = duplicate[0]
	if cells, retired, err = makeBinding().retainCells(virtual.GridSnapshot{Cells: duplicate}); !errors.Is(err, ErrGridBinding) || cells != nil || retired != nil {
		t.Fatal("duplicate key was admitted", err)
	}
	for _, w := range frozen {
		if w.Parent != mounted.panes[2] {
			t.Fatal("matching mutated frozen parent")
		}
	}
	for _, w := range scrolling {
		if w.Parent != mounted.panes[3] {
			t.Fatal("matching mutated scrolling parent")
		}
	}
}

func TestManagedGridKeyedLocalLayoutMatchesFullScene(t *testing.T) {
	for _, variableRows := range []bool{false, true} {
		for _, batch := range []int{1, 7, 64} {
			for _, dpi := range []float32{96, 120, 144, 192} {
				t.Run(fmt.Sprintf("variable%t-batch%d-dpi%g", variableRows, batch, dpi), func(t *testing.T) {
					model, _ := testNativeGrid(t, 10_000)
					if variableRows {
						var err error
						model, err = virtual.NewGrid(&variableNativeSource{rows: 10_000, revision: 1}, variableNativeColumns(), virtual.GridConfig{RowHeight: 1, HeaderHeight: 27, FrozenColumns: 1, OverscanRows: 1, OverscanColumns: 1}, virtual.GridViewport{Width: 300, Height: 180})
						if err != nil {
							t.Fatal(err)
						}
					}
					config := manualConfig(t, batch)
					config.LogicalDPI = dpi
					m, a := gridManual(t, config, func() ui.Node {
						return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true}
					})
					gridFrame(t, m)
					var g *gridBinding
					var probe *gridLayoutProbe
					if err := a.Do(func() {
						window := m.implementation.find("one").state
						g = window.root.control.grid
						probe = tree.New[gridLayoutProbe](window.body)
						probe.SetText("unrelated toolbar layout probe")
						probe.Styler(func(s *styles.Style) { s.Min.Y = units.Dp(30); s.Grow.SetZero() })
						window.body.Update()
						if err := model.Scroll(360, 240); err != nil {
							t.Error(err)
						}
						_ = a.Invalidate("one")
					}); err != nil {
						t.Fatal(err)
					}
					gridFrame(t, m)
					localFrames, retainedShapes := 0, 0
					for frame, step := range []image.Point{{0, 24}, {3, 2}, {-3, -2}, {90, 24}, {-90, -24}} {
						before := make(map[gridCellKey]*core.Text)
						shapes := make(map[*core.Text]any)
						if err := a.Do(func() {
							for i, cell := range g.projection.Cells {
								before[gridCellKey{cell.RowKey, cell.ColumnID}] = g.cells[i]
								shapes[g.cells[i]] = g.cells[i].PaintText()
							}
							probe.sizeUps = 0
							if err := model.Scroll(step.X, step.Y); err != nil {
								t.Error(err)
							}
							if err := syncGrid(g.r); err != nil {
								t.Error(err)
							}
							if _, err := InspectGrid(a, model); !errors.Is(err, ErrGridLayout) {
								t.Errorf("frame%d pending layout remained inspectable: %v", frame, err)
							}
							for i, cell := range g.projection.Cells {
								if old := before[gridCellKey{cell.RowKey, cell.ColumnID}]; old != nil && old != g.cells[i] {
									t.Errorf("frame%d overlapping key lost widget identity", frame)
								}
							}
							assertGridRetainedPaneOrder(t, g)
						}); err != nil {
							t.Fatal(err)
						}
						gridFrame(t, m)
						actual, err := m.Capture(testContext, "one")
						if err != nil {
							t.Fatal(err)
						}
						measured := gridInspection(t, a, model)
						if err := a.Do(func() {
							if probe.sizeUps == 0 {
								localFrames++
								for i, cell := range g.projection.Cells {
									w := g.cells[i]
									if before[gridCellKey{cell.RowKey, cell.ColumnID}] == w && shapes[w] == w.PaintText() && w.PaintText() != nil {
										retainedShapes++
									}
								}
							}
							g.host.Update() // independent ordinary full-scene layout oracle
						}); err != nil {
							t.Fatal(err)
						}
						gridFrame(t, m)
						full, err := m.Capture(testContext, "one")
						if err != nil {
							t.Fatal(err)
						}
						if actual.Bounds() != full.Bounds() || !bytes.Equal(actual.Pix, full.Pix) {
							t.Fatalf("frame%d local and ordinary full-scene pixels differ", frame)
						}
						oracle := gridInspection(t, a, model)
						measured.Frames, oracle.Frames = 0, 0
						if !reflect.DeepEqual(measured, oracle) {
							t.Fatalf("frame%d local/full geometry or semantics differ", frame)
						}
						if probe.sizeUps == 0 {
							t.Fatal("full oracle failed to run unrelated layout probe")
						}
					}
					if localFrames == 0 || retainedShapes == 0 {
						t.Fatalf("no actual local/retained shaping path: frames=%d shapes=%d", localFrames, retainedShapes)
					}
				})
			}
		}
	}
}

func TestManagedGridLocalLayoutGlobalFenceAndFallback(t *testing.T) {
	model, _ := testNativeGrid(t, 10_000)
	style := ui.Style{}
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true, Style: style}
	})
	gridFrame(t, m)
	initial := gridInspection(t, a, model)
	var g *gridBinding
	for _, phase := range []string{"same-generation-global", "font-style"} {
		if err := a.Do(func() {
			g = m.implementation.find("one").state.root.control.grid
			if phase == "font-style" {
				style.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 19})
				node := g.r.node
				node.Style = style
				if err := syncRecord(g.r, node); err != nil {
					t.Error(err)
				}
			} else {
				g.host.Scene.NeedsLayout()
			}
			if !g.host.Scene.GlobalLayoutPending() {
				t.Error("full layout fence missing")
			}
			if _, err := InspectGrid(a, model); !errors.Is(err, ErrGridLayout) {
				t.Errorf("pending full layout inspectable: %v", err)
			}
			if g.measurement.Load() == nil {
				t.Error("last completed publication was erased")
			}
			_ = a.Invalidate("one") // pending AX transaction before paint
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := m.Frame(testContext, "one"); err != nil {
			t.Fatal(err)
		}
		if err := a.Do(func() {
			if g.host.Scene.GlobalLayoutPending() {
				t.Error("completed full frame left pending")
			}
			if g.currentMeasurement() == nil {
				t.Error("completed frame was not published")
			}
			if a.Pending().DirtyWindows == 0 {
				t.Error("full completion did not schedule pending AX publication")
			}
		}); err != nil {
			t.Fatal(err)
		}
		gridFrame(t, m)
		current := gridInspection(t, a, model)
		if current.Generation != initial.Generation {
			t.Error("native layout changed portable model generation")
		}
		if phase == "same-generation-global" {
			current.Frames, initial.Frames = 0, 0
			if !reflect.DeepEqual(current, initial) {
				t.Error("same-generation full layout changed public data/geometry")
			}
		}
	}
}

func TestManagedGridInlinePublicationQuiescesWithoutLosingRestore(t *testing.T) {
	m, a, model, _, _ := inlineFixture(t, 144)
	g := gBinding(t, model)
	field := g.editing.record.widget.(*core.TextField)
	updates, stylesRun := 0, 0
	if err := a.Do(func() {
		field.Updater(func() { updates++ })
		field.FinalStyler(func(*styles.Style) { stylesRun++ })
	}); err != nil {
		t.Fatal(err)
	}
	for range 4 {
		if err := a.Do(func() { _ = a.Invalidate("one") }); err != nil {
			t.Fatal(err)
		}
		gridFrame(t, m)
	}
	if err := a.Do(func() {
		if updates != 0 || stylesRun != 0 {
			t.Errorf("stable editor reran full Update: %d/%d", updates, stylesRun)
		}
		if g.currentMeasurement() == nil {
			t.Error("stable editor did not complete geometry")
		}
		assertGridRetainedPaneOrder(t, g)
		if !field.ApplyNativeEdit("private preview", 15, 15, 0, 15, "mark") {
			t.Fatal("native preview refused")
		}
		cacheFieldEditingState(g.editing.record, field.EditingState())
		_ = a.Invalidate("one")
	}); err != nil {
		t.Fatal(err)
	}
	gridFrame(t, m)
	if err := a.Do(func() {
		measured := g.currentMeasurement()
		if measured == nil || measured.Editor == nil || measured.Editor.Text != "private preview" {
			t.Error("render-only preview did not publish actual completed editor data")
		}
		if updates != 0 || stylesRun != 0 {
			t.Errorf("private preview forced a full field update: %d/%d", updates, stylesRun)
		}
		g.editing.record.control.restore = true
		if err := syncGrid(g.r); err != nil {
			t.Error(err)
		}
		if updates != 1 || stylesRun != 1 {
			t.Errorf("restore did not retain full update: %d/%d", updates, stylesRun)
		}
	}); err != nil {
		t.Fatal(err)
	}
	gridFrame(t, m)
	if err := a.Do(func() {
		if updates != 1 || stylesRun != 1 {
			t.Errorf("restore completion retriggered update loop: %d/%d", updates, stylesRun)
		}
		if g.currentMeasurement() == nil {
			t.Error("restored editor did not publish completed geometry")
		}
	}); err != nil {
		t.Fatal(err)
	}
}

func assertGridRetainedPaneOrder(t *testing.T, g *gridBinding) {
	t.Helper()
	if len(g.cells) != len(g.projection.Cells) || len(g.headers) != len(g.projection.Headers) {
		t.Fatal("pool is not the current projection")
	}
	for part, pane := range g.panes {
		var want []tree.Node
		for i, header := range g.projection.Headers {
			p := 1
			if header.Frozen {
				p = 0
			}
			if p == part {
				want = append(want, g.headers[i])
			}
		}
		for i, cell := range g.projection.Cells {
			p := 3
			if cell.Frozen {
				p = 2
			}
			if p == part {
				want = append(want, g.cells[i])
			}
		}
		if g.editing != nil && g.editing.record.widget.AsTree().Parent == pane {
			want = append(want, g.editing.record.widget)
		}
		if len(want) != len(pane.Children) {
			t.Errorf("pane%d retains extra/missing children: %d/%d", part, len(pane.Children), len(want))
			continue
		}
		for i, child := range want {
			if pane.Children[i] != child || child.AsTree().Parent != pane || child.AsTree().This == nil {
				t.Errorf("pane%d painter/parent/lifetime order mismatch at %d", part, i)
			}
		}
	}
}

func TestManagedGridKeyedPoolRetiresShrinkAndRepopulates(t *testing.T) {
	model, _ := testNativeGrid(t, 10_000)
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true} })
	gridFrame(t, m)
	var g *gridBinding
	for revision, rows := range []int{3, 0, 10_000} {
		var old []*core.Text
		if err := a.Do(func() {
			g = m.implementation.find("one").state.root.control.grid
			old = append(old, g.cells...)
			if err := model.ReplaceSource(&nativeGridSource{rows: rows, revision: uint64(revision + 2)}); err != nil {
				t.Error(err)
			}
			_ = a.Invalidate("one")
		}); err != nil {
			t.Fatal(err)
		}
		gridFrame(t, m)
		if err := a.Do(func() {
			assertGridRetainedPaneOrder(t, g)
			current := make(map[*core.Text]bool)
			for _, w := range g.cells {
				current[w] = true
			}
			for _, w := range old {
				if !current[w] && w.This != nil {
					t.Error("surplus widget/private cache lifetime was retained")
				}
			}
			if rows == 0 && len(g.cells) != 0 {
				t.Error("empty source retains cell pool")
			}
			if rows > 0 && len(g.cells) == 0 {
				t.Error("nonempty source failed to repopulate")
			}
			if len(g.layoutDirty) != 0 {
				t.Error("completed frame retains dirty widget references")
			}
		}); err != nil {
			t.Fatal(err)
		}
	}
}
