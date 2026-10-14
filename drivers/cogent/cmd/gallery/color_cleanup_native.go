//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"encoding/json"
	"errors"
	"github.com/dshills/vellum/drivers/cogent"
	"os"
)

func beginColorCleanup() (func(), error) { return cogent.ObserveNativeServiceCleanup() }
func finishColorCleanup() error {
	cleanup, observed := cogent.NativeServiceCleanupObservation()
	encodeErr := json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.color.cleanup", "observed": observed, "native_cleanup": cleanup})
	if !observed || !cleanup.MainThread || !cleanup.CleanupObserved || !cleanup.CallbacksStopped || cleanup.PendingCallbacks != 0 || cleanup.NativeOutstanding != 0 || cleanup.AccessibilityWindows != 0 || len(cleanup.WindowTitles) != 0 {
		return errors.Join(encodeErr, errors.New("color native cleanup incomplete"))
	}
	return encodeErr
}
