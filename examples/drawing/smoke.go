package main

import (
	"cogentcore.org/core/core"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"
	"time"

	"cogentcore.org/core/system/driver/desktop"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/drivers/cogent"
)

func owner(a *app.Application, fn func() error) error {
	var result error
	err := a.Do(func() { result = fn() })
	if err != nil {
		return err
	}
	return result
}
func waitFrame(ctx context.Context, a *app.Application, g *cogent.MeshCanvas, after uint64) (cogent.DrawingStats, error) {
	for {
		var stats cogent.DrawingStats
		err := owner(a, func() error { var e error; stats, e = g.Stats(); return e })
		if err == nil && stats.Ready && stats.Frames > after {
			return stats, nil
		}
		select {
		case <-ctx.Done():
			return stats, fmt.Errorf("waiting for submitted frame after %d (last=%+v): %w", after, stats, ctx.Err())
		case <-time.After(5 * time.Millisecond):
		}
	}
}
func pixels(a *app.Application, g *cogent.MeshCanvas) error {
	return owner(a, func() error {
		img, err := g.Snapshot()
		if err != nil {
			return err
		}
		checks := []struct {
			world     canvas.Point
			color     color.NRGBA
			tolerance int
		}{{canvas.Point{X: -80, Y: -40}, color.NRGBA{156, 156, 156, 255}, 5}, {canvas.Point{X: 100, Y: 0}, color.NRGBA{188, 0, 187, 255}, 3}, {canvas.Point{X: 180, Y: 60}, color.NRGBA{255, 0, 0, 128}, 3}, {canvas.Point{X: -180, Y: -130}, color.NRGBA{}, 0}}
		for _, check := range checks {
			p := g.View.Screen(check.world)
			point := image.Pt(int(p.X), int(p.Y))
			if !point.In(img.Bounds()) {
				return fmt.Errorf("probe outside rendered fixture")
			}
			got := img.NRGBAAt(point.X, point.Y)
			for _, pair := range [][2]uint8{{got.R, check.color.R}, {got.G, check.color.G}, {got.B, check.color.B}, {got.A, check.color.A}} {
				delta := int(pair[0]) - int(pair[1])
				if delta < 0 {
					delta = -delta
				}
				if delta > check.tolerance {
					return fmt.Errorf("GPU pixel at %+v: got %v want %v", check.world, got, check.color)
				}
			}
		}
		return nil
	})
}
func press(ctx context.Context, id string) error {
	for {
		ok, err := cogent.PerformAccessibilityAction(ctx, "drawing", id, "press", "")
		if err == nil && ok {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(20 * time.Millisecond):
		}
	}
}
func drawingSmoke(ctx context.Context, a *app.Application, otherRef app.WindowRef, g, other *cogent.MeshCanvas, m *drawingModel, nativeErrors <-chan error, chrome func()) error {
	initial, err := waitFrame(ctx, a, g, 0)
	if err != nil {
		return err
	}
	if _, err := waitFrame(ctx, a, other, 0); err != nil {
		return err
	}
	initial, err = stableFrame(ctx, a, g)
	if err != nil {
		return err
	}
	if err := pixels(a, g); err != nil {
		return err
	}
	if err := pixels(a, other); err != nil {
		return err
	}
	// Native UI redraws reuse both the uploaded mesh and its completed texture.
	if err := owner(a, func() error { chrome(); return a.Invalidate("drawing") }); err != nil {
		return err
	}
	for {
		visited := false
		if err := owner(a, func() error { stats, err := g.Stats(); visited = stats.RenderVisits > initial.RenderVisits; return err }); err != nil {
			return err
		}
		if visited {
			break
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("chrome invalidation was not rendered: %w", ctx.Err())
		case <-time.After(5 * time.Millisecond):
		}
	}
	var unchanged cogent.DrawingStats
	if err := owner(a, func() error { var e error; unchanged, e = g.Stats(); return e }); err != nil {
		return err
	}
	if unchanged.MeshUploads != initial.MeshUploads || unchanged.Frames != initial.Frames {
		return fmt.Errorf("unchanged drawing re-uploaded or re-rendered: initial=%+v current=%+v", initial, unchanged)
	}
	if err := owner(a, func() error { g.View.Center.X += 10; return a.Invalidate("drawing") }); err != nil {
		return err
	}
	panned, err := waitFrame(ctx, a, g, initial.Frames)
	if err != nil {
		return err
	}
	if panned.MeshUploads != initial.MeshUploads || panned.UniformUploadBytes <= initial.UniformUploadBytes {
		return fmt.Errorf("viewport pan re-uploaded static mesh")
	}
	if err := pixels(a, g); err != nil {
		return err
	}
	if err := press(ctx, "schematic"); err != nil {
		return err
	}
	schematic, err := waitFrame(ctx, a, g, panned.Frames)
	if err != nil {
		return err
	}
	var query canvas.QueryStats
	if err := owner(a, func() error {
		query = m.query
		if schematic.SubmittedTriangles != query.VisibleObjects*2 || query.TotalObjects != 100000 || query.TestedObjects > 10000 {
			return fmt.Errorf("schematic submission/culling mismatch: %+v, %+v", schematic, query)
		}
		img, err := g.Snapshot()
		if err != nil {
			return err
		}
		p := g.View.Screen(canvas.Point{X: 506, Y: 404})
		if !image.Pt(int(p.X), int(p.Y)).In(img.Bounds()) || img.NRGBAAt(int(p.X), int(p.Y)).G < 150 {
			return fmt.Errorf("schematic visible GPU primitive missing")
		}
		return nil
	}); err != nil {
		return err
	}
	if err := press(ctx, "waveform"); err != nil {
		return err
	}
	wave, err := waitFrame(ctx, a, g, schematic.Frames)
	if err != nil {
		return err
	}
	columns := 0
	if err := owner(a, func() error {
		columns = m.columns
		if wave.SubmittedTriangles != columns*2 || m.waveform.Samples() != 1000000 {
			return fmt.Errorf("waveform submission/source mismatch")
		}
		img, err := g.Snapshot()
		if err != nil {
			return err
		}
		p := g.View.Screen(canvas.Point{})
		if img.NRGBAAt(int(p.X), int(p.Y)).G < 150 {
			return fmt.Errorf("waveform envelope missing from GPU frame")
		}
		return nil
	}); err != nil {
		return err
	}
	imageStats, imageReport, err := imageSmoke(ctx, a, g, other, m, wave.Frames, chrome)
	if err != nil {
		return err
	}
	if err := owner(a, func() error { g.MaxGPUBytes = 1; return a.Invalidate("drawing") }); err != nil {
		return err
	}
	select {
	case err := <-nativeErrors:
		if !strings.Contains(err.Error(), "budget exceeded") {
			return err
		}
	case <-ctx.Done():
		return fmt.Errorf("no budget refusal: %w", ctx.Err())
	}
	if err := owner(a, func() error {
		stats, err := g.Stats()
		if err != nil {
			return err
		}
		if !stats.Ready || stats.Frames != imageStats.Frames {
			return fmt.Errorf("budget refusal discarded completed frame")
		}
		return nil
	}); err != nil {
		return err
	}
	if err := owner(a, func() error { g.MaxGPUBytes = 0; return g.RebuildResources() }); err != nil {
		return err
	}
	rebuilt, err := waitFrame(ctx, a, g, imageStats.Frames)
	if err != nil {
		return err
	}
	if rebuilt.Generation <= imageStats.Generation || rebuilt.MeshUploads <= imageStats.MeshUploads || rebuilt.ImageTileUploads-imageStats.ImageTileUploads != 1024 {
		return fmt.Errorf("resource rebuild did not re-upload geometry")
	}
	var window *desktop.Window
	if err := owner(a, func() error {
		var ok bool
		window, ok = g.NativeWindow().(*desktop.Window)
		if !ok {
			return fmt.Errorf("native window unavailable")
		}
		return nil
	}); err != nil {
		return err
	}
	prior := rebuilt
	for _, size := range []image.Point{{1100, 850}, {1300, 1000}, {1200, 950}} {
		// Native window units suffice for allocation churn; SetWinSize dispatches
		// the mutation without SetSize's worker-side native screen queries.
		window.SetWinSize(size)
		next, err := waitFrame(ctx, a, g, prior.Frames)
		if err != nil {
			return err
		}
		if next.Size == prior.Size || next.Generation <= prior.Generation || next.ImageTileUploads != prior.ImageTileUploads || next.ImageTextureAllocations != prior.ImageTextureAllocations {
			return fmt.Errorf("resize did not replace target allocation")
		}
		if err := owner(a, func() error { _, err := g.Snapshot(); return err }); err != nil {
			return err
		}
		prior = next
	}
	empty, _ := canvas.NewMesh(nil, nil)
	if err := owner(a, func() error {
		g.Prepare = nil
		g.Mesh = empty
		g.Image = nil
		g.Revision++
		return a.Invalidate("drawing")
	}); err != nil {
		return err
	}
	cleared, err := waitFrame(ctx, a, g, prior.Frames)
	if err != nil {
		return err
	}
	if cleared.SubmittedTriangles != 0 || cleared.MeshBytes != 0 || cleared.ImageBytes != 0 || cleared.ImageUniformBytes != 0 || cleared.MeshUploads != prior.MeshUploads {
		return fmt.Errorf("empty frame retained geometry or falsely counted upload")
	}
	if err := owner(a, func() error {
		img, err := g.Snapshot()
		if err != nil {
			return err
		}
		if img.NRGBAAt(img.Rect.Dx()/2, img.Rect.Dy()/2) != (color.NRGBA{}) {
			return fmt.Errorf("empty GPU frame did not clear prior scene")
		}
		return nil
	}); err != nil {
		return err
	}
	if err := owner(a, func() error {
		if err := a.CloseWindow(otherRef); err != nil {
			return err
		}
		_, err := other.Snapshot()
		if err != cogent.ErrFrameNotReady {
			return fmt.Errorf("closed window resource API survived: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := owner(a, func() error { var e error; prior, e = g.Stats(); return e }); err != nil {
		return err
	}
	select {
	case err := <-nativeErrors:
		return err
	default:
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.drawing.smoke", "geometry_persistence_disabled_in_fixture": core.DebugSettings.DisableWindowGeometrySaver, "gradient_and_linear_alpha_pixels": true, "transparent_snapshot": true, "unchanged_mesh_uploads": unchanged.MeshUploads, "unchanged_offscreen_frames": unchanged.Frames, "static_pan_retains_mesh": true, "independent_windows": true, "schematic_source_objects": query.TotalObjects, "schematic_visible_objects": query.VisibleObjects, "schematic_tested_objects": query.TestedObjects, "waveform_source_samples": 1000000, "waveform_columns": columns, "waveform_submitted_triangles": wave.SubmittedTriangles, "empty_frame_clear": true, "retained_image": imageReport, "budget_refusal": true, "error_callback_resource_query": true, "resource_rebuild": true, "repeated_resize": true, "final_stats": prior, "physical_device_loss_verified": false, "presentation_latency_measured": false})
}

// Readiness requires an actual frame and a stable submitted allocation across
// startup layout passes. The later chrome assertion separately awaits a native
// Render visit, so ignored invalidation cannot pass retention checks.
func stableFrame(ctx context.Context, a *app.Application, g *cogent.MeshCanvas) (cogent.DrawingStats, error) {
	var last cogent.DrawingStats
	stableSince := time.Now()
	for {
		var now cogent.DrawingStats
		err := owner(a, func() error { var e error; now, e = g.Stats(); return e })
		if err != nil {
			return now, err
		}
		if !now.Ready || now.Frames != last.Frames || now.Size != last.Size {
			last = now
			stableSince = time.Now()
		} else if time.Since(stableSince) >= 100*time.Millisecond {
			return now, nil
		}
		select {
		case <-ctx.Done():
			return now, ctx.Err()
		case <-time.After(5 * time.Millisecond):
		}
	}
}
