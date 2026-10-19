// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
package main

import (
	"cogentcore.org/core/core"
	"errors"
	"math"
)

// Core lazily loads persisted settings in the first NewBody while Zoom is zero.
// Initialize them before the process-local override, so that reload cannot
// re-enable geometry persistence. This function does not save settings.
func isolatePerfGeometry() (restore func(), err error) {
	if core.AppearanceSettings.Zoom == 0 {
		if err := core.LoadAllSettings(); err != nil {
			return nil, err
		}
	}
	zoom := float64(core.AppearanceSettings.Zoom)
	if zoom <= 0 || math.IsNaN(zoom) || math.IsInf(zoom, 0) {
		return nil, errors.New("performance fixture: positive initialized Core zoom required")
	}
	previous := core.DebugSettings.DisableWindowGeometrySaver
	core.DebugSettings.DisableWindowGeometrySaver = true
	return func() { core.DebugSettings.DisableWindowGeometrySaver = previous }, nil
}
