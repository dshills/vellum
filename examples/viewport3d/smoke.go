package main

import (
	"cogentcore.org/core/system/driver/desktop"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"image"
	"image/color"
	"os"
	"strings"
	"time"
)

func owner(a *app.Application, fn func() error) error {
	var result error
	err := a.Do(func() { result = fn() })
	if err != nil {
		return err
	}
	return result
}
func waitStats(ctx context.Context, a *app.Application, g *cogent.GPUView, predicate func(cogent.GPUViewStats) bool) (cogent.GPUViewStats, error) {
	var stats cogent.GPUViewStats
	for {
		err := owner(a, func() error { var e error; stats, e = g.Stats(); return e })
		if err == nil && predicate(stats) {
			return stats, nil
		}
		select {
		case <-ctx.Done():
			return stats, fmt.Errorf("GPU wait last=%+v: %w", stats, ctx.Err())
		case <-time.After(5 * time.Millisecond):
		}
	}
}
func waitFrame(ctx context.Context, a *app.Application, g *cogent.GPUView, after uint64) (cogent.GPUViewStats, error) {
	return waitStats(ctx, a, g, func(s cogent.GPUViewStats) bool { return s.Ready && s.Frames > after })
}
func depthPixel(a *app.Application, g *cogent.GPUView) error {
	return owner(a, func() error {
		img, err := g.Snapshot()
		if err != nil {
			return err
		}
		c := img.NRGBAAt(img.Bounds().Dx()/2, img.Bounds().Dy()/2)
		if c.R < 250 || c.G > 3 || c.B > 3 || c.A < 250 {
			return fmt.Errorf("near-first/far-second depth pixel: got %v want red", c)
		}
		if img.NRGBAAt(2, 2) != (color.NRGBA{}) {
			return fmt.Errorf("transparent corner missing")
		}
		return nil
	})
}
func press(ctx context.Context, id string) error {
	for {
		ok, err := cogent.PerformAccessibilityAction(ctx, "viewport3d", id, "press", "")
		if err == nil && ok {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
}
func viewportSmoke(ctx context.Context, a *app.Application, g, other *cogent.GPUView, m, otherModel *cubeProgram, errors <-chan error, chrome func()) error {
	initial, err := waitFrame(ctx, a, g, 0)
	if err != nil {
		return err
	}
	independent, err := waitFrame(ctx, a, other, 0)
	if err != nil {
		return err
	}
	if err := depthPixel(a, g); err != nil {
		return err
	}
	if err := depthPixel(a, other); err != nil {
		return err
	}
	if err := owner(a, func() error {
		if m.vertices == otherModel.vertices || m.indices == otherModel.indices || m.device == nil || otherModel.device == nil {
			return fmt.Errorf("independent program buffer namespace missing")
		}
		if g.NativeWindow() == other.NativeWindow() || g.NativeWindow() == nil {
			return fmt.Errorf("windows share native identity")
		}
		var e error
		initial, e = g.Stats()
		if e != nil {
			return e
		}
		chrome()
		return a.Invalidate(g.Window)
	}); err != nil {
		return err
	}
	unchanged, err := waitStats(ctx, a, g, func(s cogent.GPUViewStats) bool { return s.RenderVisits > initial.RenderVisits })
	if err != nil {
		return err
	}
	if unchanged.Frames != initial.Frames || unchanged.Prepares != initial.Prepares || unchanged.HelperUploads != initial.HelperUploads {
		return fmt.Errorf("acknowledged chrome redraw rerendered program: initial=%+v now=%+v", initial, unchanged)
	}
	if initial.BufferBytes != 880 || initial.TargetBytes != uint64(initial.Size.X*initial.Size.Y)*36 || initial.TextureBytes != 0 || initial.HelperUploadBytes < 880 {
		return fmt.Errorf("tracked 3D byte accounting: %+v", initial)
	}
	var beforeOrbit [32]byte
	if err := owner(a, func() error {
		img, e := g.Snapshot()
		if e != nil {
			return e
		}
		beforeOrbit = sha256.Sum256(img.Pix)
		return nil
	}); err != nil {
		return err
	}
	if err := press(ctx, "orbit"); err != nil {
		return err
	}
	orbited, err := waitFrame(ctx, a, g, initial.Frames)
	if err != nil {
		return err
	}
	if orbited.BufferAllocations != initial.BufferAllocations || orbited.HelperUploadBytes-initial.HelperUploadBytes != 64 {
		return fmt.Errorf("orbit reallocated/reuploaded geometry: before=%+v now=%+v", initial, orbited)
	}
	if err := owner(a, func() error {
		img, e := g.Snapshot()
		if e != nil {
			return e
		}
		if beforeOrbit == sha256.Sum256(img.Pix) {
			return fmt.Errorf("orbit camera changed counters without changing pixels")
		}
		return nil
	}); err != nil {
		return err
	}
	if err := press(ctx, "dolly"); err != nil {
		return err
	}
	dollied, err := waitFrame(ctx, a, g, orbited.Frames)
	if err != nil {
		return err
	}
	if dollied.HelperUploadBytes-orbited.HelperUploadBytes != 64 {
		return fmt.Errorf("dolly upload size mismatch")
	}
	if err := press(ctx, "reset"); err != nil {
		return err
	}
	resetCamera, err := waitFrame(ctx, a, g, dollied.Frames)
	if err != nil {
		return err
	}
	if err := depthPixel(a, g); err != nil {
		return err
	}
	if err := owner(a, func() error { g.MaxGPUBytes = 1; return a.Invalidate(g.Window) }); err != nil {
		return err
	}
	select {
	case e := <-errors:
		if !strings.Contains(e.Error(), "budget exceeded") {
			return e
		}
	case <-ctx.Done():
		return ctx.Err()
	}
	if err := owner(a, func() error {
		s, e := g.Stats()
		if e != nil || !s.Ready || s.Frames != resetCamera.Frames {
			return fmt.Errorf("budget refusal discarded completed frame: %+v %v", s, e)
		}
		g.MaxGPUBytes = 0
		return nil
	}); err != nil {
		return err
	}
	last := resetCamera
	for _, phase := range []string{"prepare", "draw"} {
		if err := owner(a, func() error {
			m.failPrepare = phase == "prepare"
			m.failDraw = phase == "draw"
			g.Revision++
			return a.Invalidate(g.Window)
		}); err != nil {
			return err
		}
		select {
		case e := <-errors:
			if !strings.Contains(e.Error(), "controlled "+phase) {
				if phase != "prepare" || !strings.Contains(e.Error(), "controlled preparation") {
					return e
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
		if err := owner(a, func() error {
			s, e := g.Stats()
			if e != nil || s.Ready || s.BufferBytes != 0 || s.TextureBytes != 0 || s.TargetBytes != 0 || s.Frames != last.Frames {
				return fmt.Errorf("failed %s exposed/retained incomplete GPU frame: %+v %v", phase, s, e)
			}
			m.failPrepare = false
			m.failDraw = false
			g.Revision++
			return a.Invalidate(g.Window)
		}); err != nil {
			return err
		}
		repaired, e := waitFrame(ctx, a, g, last.Frames)
		if e != nil {
			return e
		}
		if repaired.ResourceGeneration <= last.ResourceGeneration || repaired.BufferBytes != 880 {
			return fmt.Errorf("%s did not repair resources", phase)
		}
		last = repaired
		if err := depthPixel(a, g); err != nil {
			return err
		}
	}
	if err := owner(a, func() error { return g.RebuildResources() }); err != nil {
		return err
	}
	rebuilt, err := waitFrame(ctx, a, g, last.Frames)
	if err != nil {
		return err
	}
	if rebuilt.ResourceGeneration <= last.ResourceGeneration || rebuilt.BufferAllocations != last.BufferAllocations+3 {
		return fmt.Errorf("full resource rebuild failed: %+v %+v", last, rebuilt)
	}
	last, err = stableFrame(ctx, a, g)
	if err != nil {
		return err
	}
	for _, size := range []image.Point{{1100, 850}, {1300, 1000}, {1200, 950}} {
		before := last
		if err := owner(a, func() error {
			w, ok := g.NativeWindow().(*desktop.Window)
			if !ok {
				return fmt.Errorf("desktop window unavailable")
			}
			w.SetWinSize(size)
			return nil
		}); err != nil {
			return err
		}
		last, err = waitStats(ctx, a, g, func(s cogent.GPUViewStats) bool { return s.Ready && s.Frames > before.Frames && s.Size != before.Size })
		if err != nil {
			return err
		}
		last, err = stableFrame(ctx, a, g)
		if err != nil {
			return err
		}
		if last.Size == before.Size || last.ResourceGeneration != before.ResourceGeneration || last.TargetGeneration <= before.TargetGeneration || last.BufferAllocations != before.BufferAllocations || last.BufferBytes != 880 || last.TargetBytes != uint64(last.Size.X*last.Size.Y)*36 {
			return fmt.Errorf("resize failed retained resources/matching depth: %+v %+v", before, last)
		}
		if err := depthPixel(a, g); err != nil {
			return err
		}
	}
	var beforeClear cogent.GPUViewStats
	if err := owner(a, func() error {
		var e error
		beforeClear, e = g.Stats()
		if e != nil {
			return e
		}
		m.clear = true
		g.Revision++
		return a.Invalidate(g.Window)
	}); err != nil {
		return err
	}
	cleared, err := waitFrame(ctx, a, g, beforeClear.Frames)
	if err != nil {
		return err
	}
	if err := owner(a, func() error {
		img, e := g.Snapshot()
		if e != nil {
			return e
		}
		if img.NRGBAAt(img.Bounds().Dx()/2, img.Bounds().Dy()/2) != (color.NRGBA{}) {
			return fmt.Errorf("empty Draw retained cube pixels")
		}
		if e := m.saved.Drop("vertices"); e == nil {
			return fmt.Errorf("expired GPUResources accepted outside callbacks")
		}
		return nil
	}); err != nil {
		return err
	}
	var otherAfter cogent.GPUViewStats
	if err := owner(a, func() error {
		var e error
		otherAfter, e = other.Stats()
		if e != nil {
			return e
		}
		if otherAfter.Frames != independent.Frames || otherAfter.BufferAllocations != independent.BufferAllocations {
			return fmt.Errorf("first-window changes affected independent resources")
		}
		return a.CloseWindow(g.Window)
	}); err != nil {
		return err
	}
	if err := owner(a, func() error {
		if _, e := g.Stats(); e != cogent.ErrFrameNotReady {
			return fmt.Errorf("closed GPU view Stats: %v", e)
		}
		if _, e := g.Snapshot(); e != cogent.ErrFrameNotReady {
			return fmt.Errorf("closed GPU view Snapshot: %v", e)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := depthPixel(a, other); err != nil {
		return err
	}
	if err := owner(a, func() error { return a.CloseWindow(other.Window) }); err != nil {
		return err
	}
	report := map[string]any{"initial": initial, "independentBufferNamespaces": true, "orbitPixelsChanged": true, "chrome": unchanged, "orbit": orbited, "dolly": dollied, "rebuilt": rebuilt, "resized": last, "cleared": cleared, "independent": otherAfter, "depth": "near red face before far blue face; both center pixels red", "callbackFaults": "controlled Prepare and Draw failures; partial resources retired and rebuilt", "readback": "explicit synchronous snapshots only", "closure": "closed adapters unavailable; other window remains usable", "result": "pass"}
	return json.NewEncoder(os.Stdout).Encode(report)
}

// Resize notifications can schedule multiple layout passes. Establish the actual
// settled target before requesting another size, preserving the changed-size
// assertion rather than accepting an unrelated pending frame.
func stableFrame(ctx context.Context, a *app.Application, g *cogent.GPUView) (cogent.GPUViewStats, error) {
	var last cogent.GPUViewStats
	stableSince := time.Now()
	for {
		var now cogent.GPUViewStats
		err := owner(a, func() error { var e error; now, e = g.Stats(); return e })
		if err != nil {
			return now, err
		}
		if !now.Ready || now.Size != last.Size || now.Frames != last.Frames {
			last = now
			stableSince = time.Now()
		} else if time.Since(stableSince) >= 100*time.Millisecond {
			return now, nil
		}
		select {
		case <-ctx.Done():
			return now, fmt.Errorf("settling viewport target: %w", ctx.Err())
		case <-time.After(5 * time.Millisecond):
		}
	}
}
