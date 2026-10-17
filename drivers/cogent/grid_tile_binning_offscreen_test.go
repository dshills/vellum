//go:build offscreen

// Copyright (c) 2026, Davin Hills.
// SPDX-License-Identifier: MIT

package cogent

import (
	"bytes"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint/renderers/rasterx"
	"cogentcore.org/core/paint/renderers/sceneplan"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"image"
	"image/draw"
	"strconv"
	"testing"
)

func replayEncodedGPU(t *testing.T, e *sceneplan.Encoded, dst *image.RGBA) {
	t.Helper()
	word := func(b []byte, i uint32) uint32 { return binary.LittleEndian.Uint32(b[int(i)*4:]) }
	blend := func(dst, src, alpha, coverage uint32, over bool) uint32 {
		if !over {
			return (src * coverage / 16776960) & 255
		}
		a := (65535 - alpha*coverage/65535) * 257
		return ((dst*a + src*coverage) / 16776960) & 255
	}
	for y := 0; y < dst.Rect.Dy(); y++ {
		for x := 0; x < dst.Rect.Dx(); x++ {
			tile := uint32((y/sceneplan.TileSize)*e.Tiles.X + x/sceneplan.TileSize)
			start := word(e.Rows, tile*2)
			count := word(e.Rows, tile*2+1)
			at := dst.PixOffset(x, y)
			for ri := start; ri < start+count; ri++ {
				oi := word(e.Refs, ri) * 16
				minX, minY := int(word(e.Ops, oi)), int(word(e.Ops, oi+1))
				if x < minX || y < minY || x >= int(word(e.Ops, oi+2)) || y >= int(word(e.Ops, oi+3)) {
					continue
				}
				kind, over := word(e.Ops, oi+8), word(e.Ops, oi+9) == uint32(draw.Over)
				src := [4]uint32{word(e.Ops, oi+4), word(e.Ops, oi+5), word(e.Ops, oi+6), word(e.Ops, oi+7)}
				if kind == sceneplan.Spans {
					row := word(e.Ops, oi+10) + uint32(y-minY)*2
					s0, n := word(e.Data, row), word(e.Data, row+1)
					for si := s0; si < s0+n*3; si += 3 {
						if uint32(x) >= word(e.Data, si) && uint32(x) < word(e.Data, si+1) {
							for ch := 0; ch < 4; ch++ {
								dst.Pix[at+ch] = byte(blend(uint32(dst.Pix[at+ch]), src[ch], src[3], word(e.Data, si+2), over))
							}
						}
					}
					continue
				}
				if kind == sceneplan.Pixels {
					v := word(e.Data, word(e.Ops, oi+10)+uint32(y-minY)*word(e.Ops, oi+11)+uint32(x-minX))
					for ch := 0; ch < 4; ch++ {
						src[ch] = ((v >> uint(ch*8)) & 255) * 257
					}
				}
				mask := word(e.Ops, oi+12)
				if mask != ^uint32(0) {
					byteAt := (uint32(y-minY)+word(e.Ops, oi+15))*word(e.Ops, oi+13) + uint32(x-minX) + word(e.Ops, oi+14)
					a := ((word(e.Data, mask+byteAt/4) >> ((byteAt % 4) * 8)) & 255) * 257
					for ch := 0; ch < 4; ch++ {
						dst.Pix[at+ch] = byte(blend(uint32(dst.Pix[at+ch]), src[ch], src[3], a, over))
					}
				} else {
					for ch := 0; ch < 4; ch++ {
						v := src[ch]
						if over {
							v = (uint32(dst.Pix[at+ch])*(65535-src[3])*257/65535 + src[ch])
						}
						dst.Pix[at+ch] = byte(v >> 8)
					}
				}
			}
		}
	}
}

type tileBinningSource struct{ nativeGridSource }

func (s *tileBinningSource) Cell(row, column string) (string, error) {
	r, err := strconv.Atoi(row[4:])
	if err != nil {
		return "", err
	}
	c, err := strconv.Atoi(column[1:])
	return strconv.Itoa((r*20 + c) * 37 % 10000), err
}

type tileBinningVariableSource struct{ *tileBinningSource }

func (s *tileBinningVariableSource) RowOffset(i int) (int, error) {
	if i < 0 || i > s.rows {
		return 0, virtual.ErrGridRange
	}
	n := i / 4 * 48
	for j := 0; j < i%4; j++ {
		n += []int{9, 11, 13, 15}[j]
	}
	return n, nil
}
func (s *tileBinningVariableSource) RowAt(pixel int) (int, error) {
	max, _ := s.RowOffset(s.rows)
	if pixel < 0 || pixel >= max {
		return 0, virtual.ErrGridRange
	}
	row, rem := pixel/48*4, pixel%48
	for _, h := range []int{9, 11, 13, 15} {
		if rem < h {
			return row, nil
		}
		row++
		rem -= h
	}
	return 0, virtual.ErrGridRange
}
func TestVellumTileBinningActualGrid(t *testing.T) {
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
						rs := rasterx.New(math32.FromPoint(scene.SceneGeom.Size)).(*rasterx.Renderer)
						defer rs.Release()
						packet := rs.ScenePacket(commands)
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
						got := image.NewRGBA(packet.Bounds())
						replayEncodedGPU(t, encoded, got)
						if !bytes.Equal(expected.Pix, got.Pix) || !bytes.Equal(expected.Pix, captured.Pix) {
							t.Fatal("serialized/packet/actual grid pixels differ")
						}
						h := sha256.New()
						for _, data := range [][]byte{encoded.Ops, encoded.Rows, encoded.Refs, encoded.Data} {
							h.Write(data)
						}
						record := map[string]any{"dpi": dpi, "variable": variable, "frame": frame, "bounds": packet.Bounds(), "operations": packet.Operations(), "packet_bytes": packet.Bytes(), "tiles": encoded.Tiles, "original_refs": len(original.Refs) / 4, "retained_refs": len(encoded.Refs) / 4, "dense_selected": len(encoded.Refs)/4 > 2*encoded.Tiles.X*encoded.Tiles.Y, "encoded_sha256": fmt.Sprintf("%x", h.Sum(nil)), "pixels_sha256": fmt.Sprintf("%x", sha256.Sum256(got.Pix))}
						line, _ := json.Marshal(record)
						t.Log("BINNING", string(line))
					}); err != nil {
						t.Fatal(err)
					}
				}
			})
		}
	}
}
