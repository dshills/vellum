//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"errors"
	"fmt"

	"cogentcore.org/core/system"
	"github.com/dshills/vellum/app"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/virtual"
)

// The same-name fixture keeps only the immediately preceding closed opening.
// It intentionally retains these probe objects until the next cycle completes;
// it is a bounded diagnostic workload, not an empty-window memory baseline.
type retiredNativeWindow struct {
	ref        app.WindowRef
	completion *app.Completion
	native     system.Window
	canvas     *cogent.GPUCanvas
	grid       *virtual.Grid
	outline    *virtual.Outline
	composed   bool
}

// Call on the owner. These are the old objects, never name-based lookups of the
// replacement. Native readback must refuse before any backend work occurs.
func (w *retiredNativeWindow) verify(a *app.Application) error {
	if w == nil || w.native == nil || !w.native.IsClosed() {
		return errors.New("old native window did not remain closed")
	}
	if w.composed && (w.grid == nil || w.outline == nil) {
		return errors.New("closed composed window lacks original data-view probes")
	}
	if w.canvas == nil || w.canvas.NativeWindow() != nil {
		return errors.New("closed canvas exposed native window")
	}
	if _, err := w.canvas.Snapshot(); !errors.Is(err, cogent.ErrFrameNotReady) {
		return fmt.Errorf("closed readback: %v", err)
	}
	if w.grid != nil {
		if _, err := cogent.InspectGrid(a, w.grid); !errors.Is(err, cogent.ErrGridClosed) {
			return fmt.Errorf("closed grid inspection: %v", err)
		}
	}
	if w.outline != nil {
		if _, err := cogent.InspectTree(a, w.outline); !errors.Is(err, cogent.ErrTreeClosed) {
			return fmt.Errorf("closed tree inspection: %v", err)
		}
	}
	return nil
}
