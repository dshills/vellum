//go:build offscreen

// Copyright (c) 2026, Davin Hills.
// SPDX-License-Identifier: MIT

package cogent

import (
	"bytes"
	"cogentcore.org/core/core"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/text/fonts"
	"cogentcore.org/core/text/rich"
	"cogentcore.org/core/text/shaped"
	"cogentcore.org/core/text/shaped/shapers/shapedgt"
	"cogentcore.org/core/text/text"
	"cogentcore.org/core/text/textpos"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"html"
	"image"
	"io/fs"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

type gridIntrinsicWordSource struct{ tileBinningSource }

func (s *gridIntrinsicWordSource) Cell(row, column string) (string, error) {
	values := []string{"", "short value", "<literal & markup>", "tab\t spaced  value", "first\nsecond\nthird\nfourth", "e\u0301 combining \U0001f642", "abc \u05d0\u05d1\u05d2 123 \u0639\u0631\u0628\u064a", "averylongwordthatmustremainunwrappedinthefixedgridcell", "line \u2028 separator"}
	n := int(row[len(row)-1]) + int(column[len(column)-1])
	return values[n%len(values)], nil
}

type gridIntrinsicVariableWords struct {
	*gridIntrinsicWordSource
	metrics *tileBinningVariableSource
}

func (s *gridIntrinsicVariableWords) RowOffset(i int) (int, error) { return s.metrics.RowOffset(i) }
func (s *gridIntrinsicVariableWords) RowAt(i int) (int, error)     { return s.metrics.RowAt(i) }

type gridIntrinsicRunState struct {
	Range   textpos.Range
	Advance float32
	Bounds  math32.Box2
	Font    text.Font
	Runes   []math32.Box2
}
type gridIntrinsicLineState struct {
	Range  textpos.Range
	Offset math32.Vector2
	Bounds math32.Box2
	Runs   []gridIntrinsicRunState
}
type gridIntrinsicLabelState struct {
	Size       core.GeomSize
	Position   any
	Bounds     math32.Box2
	Offset     math32.Vector2
	LineHeight float32
	Truncated  bool
	Lines      []gridIntrinsicLineState
}

func gridIntrinsicLabelOracle(w *core.Text) gridIntrinsicLabelState {
	result := gridIntrinsicLabelState{Size: w.Geom.Size, Position: w.Geom.Pos}
	if p := w.PaintText(); p != nil {
		result.Bounds = p.Bounds
		result.Offset = p.Offset
		result.LineHeight = p.LineHeight
		result.Truncated = p.Truncated
		for _, line := range p.Lines {
			ln := gridIntrinsicLineState{Range: line.SourceRange, Offset: line.Offset, Bounds: line.Bounds}
			for _, run := range line.Runs {
				r := gridIntrinsicRunState{Range: run.Runes(), Advance: run.Advance(), Bounds: run.LineBounds(), Font: run.AsBase().Font}
				for i := r.Range.Start; i < r.Range.End; i++ {
					r.Runes = append(r.Runes, run.RuneBounds(i))
				}
				ln.Runs = append(ln.Runs, r)
			}
			result.Lines = append(result.Lines, ln)
		}
	}
	return result
}
func gridIntrinsicPoolOracle(g *gridBinding) []gridIntrinsicLabelState {
	var state []gridIntrinsicLabelState
	for i, c := range g.projection.Cells {
		if !c.Clip.Empty() {
			state = append(state, gridIntrinsicLabelOracle(g.cells[i]))
		}
	}
	for i, h := range g.projection.Headers {
		if !h.Clip.Empty() {
			state = append(state, gridIntrinsicLabelOracle(g.headers[i]))
		}
	}
	return state
}

// Same-state independent full-scene and frozen-original comparison fixture.
func TestGridIntrinsicSizeMatchesFullLayout(t *testing.T) {
	for _, dpi := range []float32{96, 120, 144, 162, 192} {
		for _, variable := range []bool{false, true} {
			t.Run(fmt.Sprintf("%g/%v", dpi, variable), func(t *testing.T) {
				source := &gridIntrinsicWordSource{tileBinningSource{nativeGridSource: nativeGridSource{rows: 1000000, revision: 1}}}
				var input virtual.GridSource = source
				if variable {
					input = &gridIntrinsicVariableWords{source, &tileBinningVariableSource{&source.tileBinningSource}}
				}
				columns := make([]virtual.GridColumn, 20)
				for i := range columns {
					columns[i] = virtual.GridColumn{ID: fmt.Sprint("c", i), Title: fmt.Sprint("C", i), Width: 96}
				}
				model, err := virtual.NewGrid(input, columns, virtual.GridConfig{RowHeight: 12, HeaderHeight: 24, FrozenColumns: 2, OverscanRows: 1, OverscanColumns: 1}, virtual.GridViewport{Width: 768, Height: 624})
				if err != nil {
					t.Fatal(err)
				}
				config := manualConfig(t, 7)
				config.LogicalDPI = dpi
				config.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 15})
				a, err := app.New(app.Config{QueueCapacity: 32})
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(a.Close)
				view := func() ui.Node {
					return ui.Node{ID: "root", Kind: ui.Column, Style: ui.Style{Padding: ui.Some(ui.Insets{}), Gap: ui.Some(float32(0))}, Children: []ui.Node{{ID: "g", Kind: ui.DataGrid, Grid: model, Style: ui.Style{Grow: ui.Some(float32(0)), MinSize: ui.Some(ui.Size{Width: 1299, Height: 1060}), MaxSize: ui.Some(ui.Size{Width: 1299, Height: 1060})}}}}
				}
				if err = a.Open(app.Window{ID: "one", Title: "Size probe", Width: float32(math.Ceil(float64(1500 * dpi / 160))), Height: float32(math.Ceil(float64(1250 * dpi / 160))), View: view}); err != nil {
					t.Fatal(err)
				}
				m, err := NewManual(testContext, a, config)
				if err != nil {
					t.Fatal(err)
				}
				if err = a.Do(func() { m.implementation.find("one").state.body.Scene.SetProperty(gridCellSurfaceProperty, true) }); err != nil {
					t.Fatal(err)
				} // legacy intrinsic widget oracle
				t.Cleanup(func() {
					if e := m.Close(testContext); e != nil {
						t.Error(e)
					}
				})
				gridFrame(t, m)
				var g *gridBinding
				if err = a.Do(func() { g = findRecord(m.implementation.find("one").state.root, "g").control.grid }); err != nil {
					t.Fatal(err)
				}
				type prior struct {
					widget        *core.Text
					paint         any
					width, height int
				}
				for frame, step := range []image.Point{{0, 12}, {0, 12}, {0, -12}, {0, 1}, {0, -1}, {96, 0}, {-96, 0}} {
					old := map[gridCellKey]prior{}
					record := map[string]any{"event": "grid.size.probe", "dpi": dpi, "variable": variable, "frame": frame, "step": step}
					if err = a.Do(func() {
						for i, c := range g.projection.Cells {
							w, h := gridPhysicalSize(c.Rect, g.host.Styles.UnitContext.Px(1))
							old[gridCellKey{c.RowKey, c.ColumnID}] = prior{g.cells[i], g.cells[i].PaintText(), w, h}
						}
						if frame == 3 {
							if e := model.Move(1, 1); e != nil {
								t.Error(e)
							}
						}
						if e := model.Scroll(step.X, step.Y); e != nil {
							t.Error(e)
						}
						if e := syncGrid(g.r); e != nil {
							t.Error(e)
						}
						overlap, dirty, widths, heights := 0, 0, 0, 0
						for i, c := range g.projection.Cells {
							if previous, ok := old[gridCellKey{c.RowKey, c.ColumnID}]; ok && previous.widget == g.cells[i] {
								overlap++
								if g.layoutDirty[g.cells[i]] {
									dirty++
								}
								w, h := gridPhysicalSize(c.Rect, g.host.Styles.UnitContext.Px(1))
								if previous.width != w {
									widths++
								}
								if previous.height != h {
									heights++
								}
							}
						}
						record["cells"] = len(g.cells)
						record["dirty_total"] = len(g.layoutDirty)
						record["overlap"] = overlap
						record["dirty_overlap"] = dirty
						record["width_changed_overlap"] = widths
						record["height_changed_overlap"] = heights
						record["global_layout_pending"] = g.host.Scene.GlobalLayoutPending()
						record["scale"] = g.host.Styles.UnitContext.Px(1)
					}); err != nil {
						t.Fatal(err)
					}
					gridFrame(t, m)
					if err = a.Do(func() {
						changed, retained := 0, 0
						for i, c := range g.projection.Cells {
							if previous, ok := old[gridCellKey{c.RowKey, c.ColumnID}]; ok && previous.widget == g.cells[i] {
								if previous.paint != g.cells[i].PaintText() {
									changed++
								} else {
									retained++
								}
							}
						}
						record["changed_shape_overlap"] = changed
						record["retained_shape_overlap"] = retained
					}); err != nil {
						t.Fatal(err)
					}
					actual, e := m.Capture(testContext, "one")
					if e != nil {
						t.Fatal(e)
					}
					var intrinsic []gridIntrinsicLabelState
					if err = a.Do(func() { intrinsic = gridIntrinsicPoolOracle(g) }); err != nil {
						t.Fatal(err)
					}
					inspection := gridInspection(t, a, model)
					inspection.Frames = 0
					if err = a.Do(func() { g.host.Update() }); err != nil {
						t.Fatal(err)
					}
					gridFrame(t, m)
					full, e := m.Capture(testContext, "one")
					if e != nil {
						t.Fatal(e)
					}
					oracle := gridInspection(t, a, model)
					oracle.Frames = 0
					var intrinsicFull []gridIntrinsicLabelState
					if err = a.Do(func() { intrinsicFull = gridIntrinsicPoolOracle(g) }); err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(intrinsic, intrinsicFull) {
						t.Fatalf("frame%d local/full intrinsic size or glyph interaction geometry differ", frame)
					}
					if actual.Bounds() != full.Bounds() || !bytes.Equal(actual.Pix, full.Pix) {
						t.Fatalf("frame%d local/full pixels differ", frame)
					}
					if !reflect.DeepEqual(inspection, oracle) {
						t.Fatalf("frame%d local/full geometry/semantics differ", frame)
					}
					intrinsicBytes, e := json.Marshal(intrinsic)
					if e != nil {
						t.Fatal(e)
					}
					record["intrinsic_sha256"] = fmt.Sprintf("%x", sha256.Sum256(intrinsicBytes))
					record["pixels_sha256"] = fmt.Sprintf("%x", sha256.Sum256(actual.Pix))
					geometry, e := json.Marshal(inspection)
					if e != nil {
						t.Fatal(e)
					}
					record["geometry_sha256"] = fmt.Sprintf("%x", sha256.Sum256(geometry))
					if dpi == 162 && !variable && frame == 0 {
						if record["height_changed_overlap"].(int) < 200 || record["changed_shape_overlap"].(int) > 50 || record["global_layout_pending"].(bool) {
							t.Fatal("fractional-DPI managed cells did not reuse completed intrinsic sizing")
						}
					}
					raw, e := json.Marshal(record)
					if e != nil {
						t.Fatal(e)
					}
					fmt.Println(string(raw))
				}
			})
		}
	}
}

// Test-only reference frozen from grid_retained.go at 99b632b. It preserves
// the prior physical-allocation sizing rule independently of the candidate.
func (g *gridBinding) gridPriorPlaceForTest(w *core.Text, pane *gridPane, rect, clip virtual.GridRect, value string, selected, cursor bool) (style, size, moved bool) {
	oldRect, hasRect := w.Property("vellum.grid.rect").(virtual.GridRect)
	oldCursor, _ := w.Property("vellum.grid.cursor").(bool)
	value = html.EscapeString(value)
	moved = w.Parent != pane
	style = moved || w.Text != value || w.StateIs(states.Selected) != selected || w.StateIs(states.Invisible) != clip.Empty() || oldCursor != cursor
	oldWidth, oldHeight := gridPhysicalSize(oldRect, g.host.Styles.UnitContext.Px(1))
	width, height := gridPhysicalSize(rect, g.host.Styles.UnitContext.Px(1))
	size = style || !hasRect || oldWidth != width || oldHeight != height || !clip.Empty() && w.PaintText() == nil
	g.place(w, pane, rect, clip, value, selected, cursor)
	return
}

// The reference comparison uses real Text sizing, including unit compilation,
// intrinsic bounds and glyph interaction geometry. Refused styles must execute
// the same dirty sizing as the frozen prior rule, rather than only return a flag.
func TestGridIntrinsicDependentStylesMatchPrior(t *testing.T) {
	cases := []struct {
		name         string
		change       func(*core.Text)
		natural      bool
		eligible     bool
		geometryOnly bool
	}{
		{name: "owned-preformatted", eligible: true},
		{name: "natural-height", natural: true},
		{name: "wrap", change: func(w *core.Text) { w.Styles.Text.WhiteSpace = text.WrapAlways }},
		{name: "minimum", change: func(w *core.Text) { w.Styles.Min.Y = units.Dot(10) }},
		{name: "maximum", change: func(w *core.Text) { w.Styles.Max.Y = units.Dot(100) }},
		{name: "grow", change: func(w *core.Text) { w.Styles.Grow.Y = 1 }},
		{name: "interactive", change: func(w *core.Text) { w.Styles.SetAbilities(true, abilities.Focusable) }},
		{name: "font-element-height", geometryOnly: true, change: func(w *core.Text) { w.Styles.Font.Size = units.Eh(90) }},
		{name: "font-custom", geometryOnly: true, change: func(w *core.Text) {
			w.Styles.Font.Size = units.Custom(func(c *units.Context) float32 { return c.Elh * .9 })
		}},
		{name: "indent-element-width", change: func(w *core.Text) { w.Styles.Text.Indent = units.Ew(5) }},
		{name: "hanging-parent-width", change: func(w *core.Text) { w.Styles.Text.Hanging = units.Pw(5) }},
		{name: "padding-element-height", change: func(w *core.Text) { w.Styles.Padding.Top = units.Eh(15) }},
		{name: "margin-parent-height", change: func(w *core.Text) { w.Styles.Margin.Top = units.Ph(5) }},
		{name: "border-element-height", change: func(w *core.Text) { w.Styles.Border.Width.Top = units.Eh(10) }},
		{name: "border-offset-relative", change: func(w *core.Text) { w.Styles.Border.Offset.Top = units.Eh(10) }},
		{name: "radius-relative", change: func(w *core.Text) { w.Styles.Border.Radius.Top = units.Eh(10) }},
		{name: "max-border-relative", change: func(w *core.Text) { w.Styles.MaxBorder.Width.Top = units.Eh(10) }},
		{name: "shadow", change: func(w *core.Text) { w.Styles.BoxShadow = []styles.Shadow{{Blur: units.Eh(10)}} }},
		{name: "parts", change: func(w *core.Text) { w.Parts = core.NewFrame(w) }},
		{name: "children", change: func(w *core.Text) { core.NewText(w).SetText("child") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			model, _ := testNativeGrid(t, 1000)
			config := manualConfig(t, 7)
			config.LogicalDPI = 162
			m, a := gridManual(t, config, func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true} })
			gridFrame(t, m)
			if err := a.Do(func() {
				g := m.implementation.find("one").state.root.control.grid
				template := g.cells[len(g.cells)/2]
				pane := g.panes[3]
				var pair [2]*core.Text
				old := virtual.GridRect{X: 100, Y: 24, Width: 96, Height: 12}
				value := "sizing one\nsecond\nthird\nfourth"
				for i := range pair {
					w := core.NewText(pane)
					pair[i] = w
					w.Styles = template.Styles
					w.Geom = template.Geom
					w.OverrideStyle = true
					if tc.change != nil {
						tc.change(w)
					}
					w.Styles.ToDots()
					g.place(w, pane, old, old, value, false, false)
					w.UpdateWidget()
					w.SizeUp()
					w.SizeDown(0)
					w.SizeFinal()
				}
				defer pair[0].Delete()
				defer pair[1].Delete()
				pane.Position()
				g.host.ApplyScenePos()
				reused := 0
				for step, y := range []int{25, 26, 27, 28, 29, 30, 24} {
					rect := old
					rect.Y = y
					if tc.natural {
						rect.Height += step % 2
					}
					sty, size, _ := g.placeRetained(pair[0], pane, rect, rect, value, false, false)
					wantStyle, wantSize, _ := g.gridPriorPlaceForTest(pair[1], pane, rect, rect, value, false, false)
					if sty != wantStyle || !tc.eligible && !tc.geometryOnly && size != wantSize {
						t.Errorf("step%d refused-style sizing differs from prior rule", step)
						return
					}
					if wantSize && !size {
						reused++
					}
					for i, dirty := range []bool{size, wantSize} {
						if dirty {
							pair[i].SizeUp()
							pair[i].SizeDown(0)
							pair[i].SizeFinal()
						}
					}
					pane.Position()
					g.host.ApplyScenePos()
					got, want := gridIntrinsicLabelOracle(pair[0]), gridIntrinsicLabelOracle(pair[1])
					if !reflect.DeepEqual(got, want) {
						t.Errorf("step%d intrinsic/glyph geometry differs from frozen prior sizing", step)
						return
					}
				}
				if tc.eligible && reused == 0 {
					t.Error("rounded-allocation labels did not reuse intrinsic sizing")
					return
				}
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

type gridIntrinsicPendingSource struct {
	nativeGridSource
	prefix string
}

func (s *gridIntrinsicPendingSource) Cell(row, column string) (string, error) {
	return s.prefix + " " + row + " " + column, nil
}

// This shaper deliberately makes allocation-derived and intrinsic-derived
// output visibly different, while keeping identical inputs deterministic.
type gridIntrinsicSensitiveShaper struct {
	shaped.Shaper
	last map[string]math32.Vector2
}

func (s *gridIntrinsicSensitiveShaper) WrapLines(tx rich.Text, style *rich.Style, ts *text.Style, size math32.Vector2) *shaped.Lines {
	out := s.Shaper.WrapLines(tx, style, ts, size)
	key := string(tx.Join())
	if strings.HasPrefix(key, "pending-") {
		s.last[key] = size
		out.Offset.X += size.Y * .125
	}
	return out
}

func TestGridIntrinsicPendingLazyShapeIsResized(t *testing.T) {
	a, host := gridContractOwner(t, 32)
	source := &gridIntrinsicPendingSource{nativeGridSource{rows: 1000, revision: 1}, "pending-old"}
	model, err := virtual.NewGrid(source, []virtual.GridColumn{{ID: "c0", Title: "A", Width: 96}, {ID: "c1", Title: "B", Width: 96}, {ID: "c2", Title: "C", Width: 96}}, virtual.GridConfig{RowHeight: 24, HeaderHeight: 26, FrozenColumns: 1, OverscanRows: 1}, virtual.GridViewport{Width: 300, Height: 180})
	if err != nil {
		t.Fatal(err)
	}
	record := mountGridContract(t, a, host, ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true})
	g := record.control.grid
	host.Styler(func(s *styles.Style) { s.Grow.Set(1, 1) })
	config := manualConfig(t, 7)
	probe := &gridIntrinsicSensitiveShaper{Shaper: shapedgt.NewShaperWithFonts([]fs.FS{fonts.Default}), last: map[string]math32.Vector2{}}
	manual, err := host.Scene.Body.NewWindow().RunManual(testContext, core.ManualWindowConfig{Size: image.Pt(500, 360), LogicalDPI: 162, EventCapacity: 128, Now: config.Clock.Now, NewShaper: func() (shaped.Shaper, error) { return probe, nil }, AfterFunc: func(c context.Context, d time.Duration, f func()) (core.ManualTimer, error) {
		return config.Clock.AfterFunc(c, d, f)
	}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if e := manual.Close(testContext); e != nil {
			t.Error(e)
		}
	})
	for range 6 {
		if _, e := a.DrainN(64); e != nil {
			t.Fatal(e)
		}
		if e := a.Do(func() {
			if e := syncGrid(record); e != nil {
				t.Error(e)
			}
		}); e != nil {
			t.Fatal(e)
		}
		if _, e := manual.Frame(testContext); e != nil {
			t.Fatal(e)
		}
	}
	var target *core.Text
	var pendingKept bool
	if err = a.Do(func() {
		if !g.layout.valid || !g.projectionReady {
			t.Error("custom-shaper grid never completed its initial layout")
			return
		}
		next := &gridIntrinsicPendingSource{nativeGridSource{rows: 1000, revision: 2}, "pending-new"}
		if e := model.ReplaceSource(next); e != nil {
			t.Error(e)
			return
		}
		if e := syncGrid(record); e != nil {
			t.Error(e)
			return
		}
		for i, c := range g.projection.Cells {
			if !c.Clip.Empty() {
				target = g.cells[i]
				break
			}
		}
		if target == nil || target.PaintText() != nil || !g.layoutDirty[target] {
			t.Error("content update did not retain deferred intrinsic sizing")
			return
		}
		// Invoke the same lazy Render fallback as a canceled local request. The
		// Core cancellation mechanism itself has separate owned-stage0055 tests.
		target.Render()
		if target.PaintText() == nil || target.PaintText().Offset.X == 0 {
			t.Error("size-sensitive lazy allocation shape was not exercised")
			return
		}
		if e := model.Scroll(0, 1); e != nil {
			t.Error(e)
			return
		}
		if e := syncGrid(record); e != nil {
			t.Error(e)
			return
		}
		pendingKept = g.layoutDirty[target]
	}); err != nil {
		t.Fatal(err)
	}
	if t.Failed() {
		return
	}
	if _, err = manual.Frame(testContext); err != nil {
		t.Fatal(err)
	}
	var actual gridIntrinsicLabelState
	if err = a.Do(func() {
		actual = gridIntrinsicLabelOracle(target)
		key := string(target.TextRunes())
		if probe.last[key] != (math32.Vector2{}) || target.PaintText().Offset.X != 0 {
			t.Error("lazy allocation shape survived the required intrinsic sizing pass")
		}
		if !pendingKept {
			t.Error("later size-only sync discarded earlier intrinsic sizing")
		}
	}); err != nil {
		t.Fatal(err)
	}
	if t.Failed() {
		return
	}
	pixels, err := manual.Image(testContext)
	if err != nil {
		t.Fatal(err)
	}
	host.Scene.Body.Update()
	if _, err = manual.Frame(testContext); err != nil {
		t.Fatal(err)
	}
	full, err := manual.Image(testContext)
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Do(func() {
		if !reflect.DeepEqual(actual, gridIntrinsicLabelOracle(target)) {
			t.Error("custom-shaper local/full intrinsic geometry differs")
		}
	}); err != nil {
		t.Fatal(err)
	}
	if pixels.Bounds() != full.Bounds() || !bytes.Equal(pixels.Pix, full.Pix) {
		t.Fatal("custom-shaper pending local/full pixels differ")
	}
}
