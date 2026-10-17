//go:build offscreen

// Copyright (c) 2026, Davin Hills.
// SPDX-License-Identifier: MIT

package cogent

import (
	"bytes"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint/renderers/rasterx"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"image"
	"reflect"
	"testing"
)

// This read-only numeric probe verifies the private optimization was exercised
// across the renderer package boundary without adding a public testing API.
func gridPlacementHits(r *rasterx.Renderer) uint64 {
	coverage := reflect.ValueOf(r).Elem().FieldByName("glyphCoverage")
	if !coverage.IsValid() || coverage.IsNil() {
		return 0
	}
	placements := coverage.Elem().FieldByName("placements")
	if !placements.IsValid() || placements.IsNil() {
		return 0
	}
	return placements.Elem().FieldByName("hits").Uint()
}

const gridPlacementBaseline = false

func TestVellumGlyphPlacementActualGridWarm(t *testing.T) {
	for _, dpi := range []float32{96, 160, 192} {
		for _, variable := range []bool{false, true} {
			t.Run(fmt.Sprintf("%g/%v", dpi, variable), func(t *testing.T) {
				source := &tileBinningSource{nativeGridSource: nativeGridSource{rows: 1000000, revision: 1}}
				var input virtual.GridSource = source
				if variable {
					input = &tileBinningVariableSource{source}
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
				if err = a.Open(app.Window{ID: "one", Title: "Grid encoder", Width: 1500 * dpi / 160, Height: 1250 * dpi / 160, View: view}); err != nil {
					t.Fatal(err)
				}
				m, err := NewManual(testContext, a, config)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if e := m.Close(testContext); e != nil {
						t.Error(e)
					}
				})
				var rs *rasterx.Renderer
				t.Cleanup(func() {
					if rs != nil {
						rs.Release()
					}
				})
				for frame, step := range []image.Point{{}, {0, 12}, {96, 12}, {-96, -12}} {
					if frame > 0 {
						if err = a.Do(func() {
							if e := model.Scroll(step.X, step.Y); e != nil {
								t.Error(e)
							}
							_ = a.Invalidate("one")
						}); err != nil {
							t.Fatal(err)
						}
					}
					gridFrame(t, m)
					captured, err := m.Capture(testContext, "one")
					if err != nil {
						t.Fatal(err)
					}
					if err = a.Do(func() {
						scene := m.implementation.find("one").state.body.Scene
						scene.RenderWidget()
						commands := scene.Painter.RenderDone()
						if rs == nil {
							rs = rasterx.New(math32.FromPoint(scene.SceneGeom.Size)).(*rasterx.Renderer)
						}
						beforeHits := gridPlacementHits(rs)
						cold := rs.ScenePacket(commands)
						packet := rs.ScenePacket(commands)
						if dpi == 96 && !gridPlacementBaseline && gridPlacementHits(rs) <= beforeHits {
							t.Fatal("actual grid never reused an exact placement")
						}
						if cold == nil || packet == nil {
							t.Fatal("first capture refused")
						}
						ca, coa := cold.EncodeGPU()
						wa, woa := packet.EncodeGPU()
						if coa != woa || !reflect.DeepEqual(ca, wa) {
							t.Fatal("warm packet bytes differ")
						}
						if packet == nil {
							t.Fatal("actual packet refused")
						}
						encoded, ok := packet.EncodeGPU()
						if !ok {
							t.Fatal("actual encoded refused")
						}
						original, ok := packet.Encode()
						if !ok {
							t.Fatal("original layout refused")
						}
						expected := image.NewRGBA(packet.Bounds())
						if !packet.Replay(expected) {
							t.Fatal("packet replay refused")
						}
						// The existing tile-binning fixture executes the serialized GPU
						// integer oracle. This companion proves warm bytes and captured
						// pixels without duplicating its expensive per-pixel replay.
						if !bytes.Equal(expected.Pix, captured.Pix) {
							t.Fatal("warm packet/actual grid pixels differ")
						}
						h := sha256.New()
						for _, data := range [][]byte{encoded.Ops, encoded.Rows, encoded.Refs, encoded.Data} {
							h.Write(data)
						}
						record := map[string]any{"dpi": dpi, "variable": variable, "frame": frame, "bounds": packet.Bounds(), "operations": packet.Operations(), "packet_bytes": packet.Bytes(), "tiles": encoded.Tiles, "original_refs": len(original.Refs) / 4, "retained_refs": len(encoded.Refs) / 4, "dense_selected": len(encoded.Refs)/4 > 2*encoded.Tiles.X*encoded.Tiles.Y, "encoded_sha256": fmt.Sprintf("%x", h.Sum(nil)), "pixels_sha256": fmt.Sprintf("%x", sha256.Sum256(expected.Pix))}
						line, _ := json.Marshal(record)
						t.Log("PLACEMENT_WARM", string(line))
					}); err != nil {
						t.Fatal(err)
					}
				}
			})
		}
	}
}
