package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"strings"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/drivers/cogent"
)

func imagePixel(a *app.Application, g *cogent.MeshCanvas, world canvas.Point, want color.NRGBA, tolerance int) error {
	return owner(a, func() error {
		snapshot, err := g.Snapshot()
		if err != nil {
			return err
		}
		p := g.View.Screen(world)
		point := image.Pt(int(p.X), int(p.Y))
		if !point.In(snapshot.Bounds()) {
			return fmt.Errorf("image probe outside frame")
		}
		got := snapshot.NRGBAAt(point.X, point.Y)
		for _, pair := range [][2]uint8{{got.R, want.R}, {got.G, want.G}, {got.B, want.B}, {got.A, want.A}} {
			delta := int(pair[0]) - int(pair[1])
			if delta < 0 {
				delta = -delta
			}
			if delta > tolerance {
				return fmt.Errorf("image pixel %+v: got %v want %v", world, got, want)
			}
		}
		return nil
	})
}
func imageSmoke(ctx context.Context, a *app.Application, g, other *cogent.MeshCanvas, m *drawingModel, after uint64, chrome func()) (cogent.DrawingStats, map[string]any, error) {
	var empty cogent.DrawingStats
	if err := press(ctx, "image"); err != nil {
		return empty, nil, err
	}
	initial, err := waitFrame(ctx, a, g, after)
	if err != nil {
		return empty, nil, err
	}
	if initial.ImageTileUploads != 1024 || initial.ImageUploadBytes != 4096*4096*4 || initial.ImageBytes != 4096*4096*4 || initial.ImageTextureAllocations != 1 {
		return empty, nil, fmt.Errorf("4K initial image workload mismatch: %+v", initial)
	}
	var original *canvas.TiledImage
	if err := owner(a, func() error { original = m.imageLayer.Image; return nil }); err != nil {
		return empty, nil, err
	}
	if err := imagePixel(a, g, canvas.Point{X: 8, Y: 8}, color.NRGBA{40, 100, 160, 255}, 1); err != nil {
		return empty, nil, err
	}
	if err := imagePixel(a, g, canvas.Point{X: 100, Y: 0}, color.NRGBA{188, 0, 187, 255}, 3); err != nil {
		return empty, nil, fmt.Errorf("image/mesh ordering: %w", err)
	}
	patch := image.NewNRGBA(image.Rect(0, 0, 128, 128))
	for i := 0; i < len(patch.Pix); i += 4 {
		copy(patch.Pix[i:i+4], []byte{230, 40, 80, 255})
	}
	next, err := original.WithPatch(image.Rect(2048, 2048, 2176, 2176), patch)
	if err != nil {
		return empty, nil, err
	}
	if err := owner(a, func() error { m.imageLayer.Image = next; return a.Invalidate("drawing") }); err != nil {
		return empty, nil, err
	}
	patched, err := waitFrame(ctx, a, g, initial.Frames)
	if err != nil {
		return empty, nil, err
	}
	if patched.ImageTileUploads-initial.ImageTileUploads != 1 || patched.ImageUploadBytes-initial.ImageUploadBytes != 128*128*4 || patched.ImageTextureAllocations != initial.ImageTextureAllocations {
		return empty, nil, fmt.Errorf("single tile update rebuilt/full-uploaded: %+v", patched)
	}
	if err := imagePixel(a, g, canvas.Point{X: 8, Y: 8}, color.NRGBA{230, 40, 80, 255}, 1); err != nil {
		return empty, nil, err
	}
	if err := imagePixel(a, g, canvas.Point{X: 40, Y: 100}, color.NRGBA{40, 100, 160, 255}, 1); err != nil {
		return empty, nil, err
	}
	var otherBefore cogent.DrawingStats
	if err := owner(a, func() error {
		var e error
		otherBefore, e = other.Stats()
		if e != nil {
			return e
		}
		other.Mesh = nil
		copyLayer := *m.imageLayer
		copyLayer.Image = original
		other.Image = &copyLayer
		return a.Invalidate("drawing-other")
	}); err != nil {
		return empty, nil, err
	}
	independent, err := waitFrame(ctx, a, other, otherBefore.Frames)
	if err != nil {
		return empty, nil, err
	}
	if independent.ImageTileUploads != 1024 || independent.ImageUploadBytes != 4096*4096*4 || independent.MeshBytes != 0 || independent.SubmittedTriangles != 2 {
		return empty, nil, fmt.Errorf("image-only independent resources: %+v", independent)
	}
	if err := imagePixel(a, other, canvas.Point{X: 8, Y: 8}, color.NRGBA{40, 100, 160, 255}, 1); err != nil {
		return empty, nil, err
	}
	// Snapshot admission must include the retained source, not just the output.
	// This budget includes the image uniform but excludes the source texels.
	if err := owner(a, func() error {
		stats, err := g.Stats()
		if err != nil {
			return err
		}
		budget := g.MaxGPUBytes
		g.MaxGPUBytes = stats.TargetBytes + stats.MeshBytes + stats.ImageUniformBytes + 16 + stats.ReadbackBytes
		defer func() { g.MaxGPUBytes = budget }()
		if _, err := g.Snapshot(); err == nil || !strings.Contains(err.Error(), "readback exceeds GPU resource budget") {
			return fmt.Errorf("image snapshot budget omitted texels: %v", err)
		}
		return nil
	}); err != nil {
		return empty, nil, err
	}
	// Capture the baseline in the owner transaction requesting this chrome visit.
	var chromeBefore cogent.DrawingStats
	if err := owner(a, func() error {
		var err error
		chromeBefore, err = g.Stats()
		if err != nil {
			return err
		}
		chrome()
		return a.Invalidate("drawing")
	}); err != nil {
		return empty, nil, err
	}
	for {
		var current cogent.DrawingStats
		if err := owner(a, func() error { var e error; current, e = g.Stats(); return e }); err != nil {
			return empty, nil, err
		}
		if current.RenderVisits > chromeBefore.RenderVisits {
			if current.Frames != chromeBefore.Frames || current.ImageTileUploads != chromeBefore.ImageTileUploads {
				return empty, nil, fmt.Errorf("image chrome re-upload/re-render")
			}
			break
		}
		select {
		case <-ctx.Done():
			return empty, nil, ctx.Err()
		case <-time.After(5 * time.Millisecond):
		}
	}
	prior := patched
	changes := []func(){func() { g.View.Center.X += 10 }, func() { m.imageLayer.Bounds.Min.X += 20; m.imageLayer.Bounds.Max.X += 20 }, func() { m.imageLayer.Filter = canvas.ImageLinear }}
	for _, change := range changes {
		if err := owner(a, func() error { change(); return a.Invalidate("drawing") }); err != nil {
			return empty, nil, err
		}
		current, err := waitFrame(ctx, a, g, prior.Frames)
		if err != nil {
			return empty, nil, err
		}
		if current.ImageTileUploads != patched.ImageTileUploads || current.ImageTextureAllocations != patched.ImageTextureAllocations {
			return empty, nil, fmt.Errorf("camera/placement/filter uploaded texels")
		}
		prior = current
	}
	if err := imagePixel(a, g, canvas.Point{X: 28, Y: 8}, color.NRGBA{230, 40, 80, 255}, 1); err != nil {
		return empty, nil, err
	}
	// Filtering opaque red next to transparent saturated blue must produce red
	// at half alpha, with no purple halo. The small texture also exercises row8 uploads.
	source := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	source.SetNRGBA(0, 0, color.NRGBA{255, 0, 0, 255})
	source.SetNRGBA(1, 0, color.NRGBA{0, 0, 255, 0})
	halo, _ := canvas.NewTiledImage(source)
	haloLayer := &canvas.ImageLayer{Image: halo, Bounds: canvas.Bounds{Min: canvas.Point{X: -128, Y: -32}, Max: canvas.Point{X: 128, Y: 32}}, Filter: canvas.ImageLinear}
	if err := owner(a, func() error {
		g.Prepare = nil
		g.Mesh = nil
		g.Image = haloLayer
		g.View.Center = canvas.Point{}
		return a.Invalidate("drawing")
	}); err != nil {
		return empty, nil, err
	}
	filtered, err := waitFrame(ctx, a, g, prior.Frames)
	if err != nil {
		return empty, nil, err
	}
	if err := imagePixel(a, g, canvas.Point{}, color.NRGBA{255, 0, 0, 128}, 3); err != nil {
		return empty, nil, fmt.Errorf("linear alpha halo: %w", err)
	}
	if err := owner(a, func() error { haloLayer.Filter = canvas.ImageNearest; return a.Invalidate("drawing") }); err != nil {
		return empty, nil, err
	}
	nearest, err := waitFrame(ctx, a, g, filtered.Frames)
	if err != nil {
		return empty, nil, err
	}
	if nearest.ImageTileUploads != filtered.ImageTileUploads {
		return empty, nil, fmt.Errorf("nearest filter uploaded texels")
	}
	if err := imagePixel(a, g, canvas.Point{}, color.NRGBA{}, 0); err != nil {
		return empty, nil, fmt.Errorf("nearest selection: %w", err)
	}
	if err := owner(a, func() error {
		haloLayer.Filter = canvas.ImageLinear
		haloLayer.Bounds.Min.X += 999800
		haloLayer.Bounds.Max.X += 999800
		g.View.Center.X = 999800.125
		return a.Invalidate("drawing")
	}); err != nil {
		return empty, nil, err
	}
	translated, err := waitFrame(ctx, a, g, nearest.Frames)
	if err != nil {
		return empty, nil, err
	}
	if translated.ImageTileUploads != filtered.ImageTileUploads {
		return empty, nil, fmt.Errorf("translated placement uploaded texels")
	}
	if err := imagePixel(a, g, canvas.Point{X: 999800}, color.NRGBA{255, 0, 0, 128}, 3); err != nil {
		return empty, nil, fmt.Errorf("translated fractional image: %w", err)
	}
	if err := owner(a, func() error {
		m.imageLayer.Bounds = canvas.Bounds{Min: canvas.Point{X: -256, Y: -256}, Max: canvas.Point{X: 256, Y: 256}}
		m.imageLayer.Filter = canvas.ImageNearest
		g.Image = m.imageLayer
		g.Prepare = m.prepare
		g.Mesh = m.fixture
		g.View.Center = canvas.Point{}
		g.Revision++
		return a.Invalidate("drawing")
	}); err != nil {
		return empty, nil, err
	}
	restored, err := waitFrame(ctx, a, g, translated.Frames)
	if err != nil {
		return empty, nil, err
	}
	if restored.ImageTileUploads-translated.ImageTileUploads != 1024 {
		return empty, nil, fmt.Errorf("replacement did not upload complete 4K source")
	}
	return restored, map[string]any{"source_dimensions": image.Pt(4096, 4096), "initial_tiles": 1024, "initial_upload_bytes": 4096 * 4096 * 4, "changed_tiles": 1, "changed_upload_bytes": 128 * 128 * 4, "image_snapshot_budget_refusal": true, "independent_image_only_window": true, "image_below_mesh": true, "chrome_camera_placement_filter_retention": true, "premultiplied_linear_filter_pixels": true, "nearest_pixels": true, "translated_fractional_pixels": true}, nil
}
