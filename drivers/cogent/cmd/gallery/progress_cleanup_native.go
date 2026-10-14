//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/dshills/vellum/drivers/cogent"
)

func beginProgressCleanup() (func(), error) { return cogent.ObserveNativeServiceCleanup() }
func finishProgressCleanup() error {
	cleanup, observed := cogent.NativeServiceCleanupObservation()
	err := json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.progress.cleanup", "observed": observed, "native_cleanup": cleanup})
	if !observed || !cleanup.MainThread || !cleanup.CleanupObserved || !cleanup.CallbacksStopped || cleanup.PendingCallbacks != 0 || cleanup.NativeOutstanding != 0 || cleanup.AccessibilityWindows != 0 || len(cleanup.WindowTitles) != 0 {
		return errors.Join(err, errors.New("progress native cleanup incomplete"))
	}
	return err
}
