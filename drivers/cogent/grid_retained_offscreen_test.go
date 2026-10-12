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
