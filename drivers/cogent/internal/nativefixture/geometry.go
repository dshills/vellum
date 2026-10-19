// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
package nativefixture

import (
	"cogentcore.org/core/core"
	"errors"
	"math"
)

// IsolateGeometry loads persisted Core settings before disabling geometry persistence.
// Core lazily loads persisted settings in the first NewBody while Zoom is zero.
// Initialize them before the process-local override, so that reload cannot
// re-enable geometry persistence. This function does not save settings.
// It changes process-global fixture state: call once before creating the app and
// defer restore before app.Close is deferred, so restoration follows teardown.
// It is not a concurrent application preference API.
// Any error is terminal for this disposable fixture: stop before creating windows.
// Loading may apply defaults before failing; this helper does not roll back Core
// settings and must not be retried or ignored in the same process.
func IsolateGeometry() (restore func(), err error) {
	if core.AppearanceSettings.Zoom == 0 {
		if err := core.LoadAllSettings(); err != nil {
			return nil, err
		}
	}
	zoom := float64(core.AppearanceSettings.Zoom)
	if zoom <= 0 || math.IsNaN(zoom) || math.IsInf(zoom, 0) {
		return nil, errors.New("native fixture: positive initialized Core zoom required")
	}
	previous := core.DebugSettings.DisableWindowGeometrySaver
	core.DebugSettings.DisableWindowGeometrySaver = true
	return func() { core.DebugSettings.DisableWindowGeometrySaver = previous }, nil
}
