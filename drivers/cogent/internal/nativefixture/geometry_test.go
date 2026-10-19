// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
package nativefixture

import (
	"cogentcore.org/core/core"
	"errors"
	"math"
	"testing"
)

type geometrySettingsFixture struct {
	core.SettingsBase
	opens    int
	err      error
	disabled bool
	zoom     float32
}

func (f *geometrySettingsFixture) Open() error { f.opens++; return f.err }
func (f *geometrySettingsFixture) Apply() {
	core.AppearanceSettings.Zoom = f.zoom
	core.DebugSettings.DisableWindowGeometrySaver = f.disabled
}

// Uses the actual lazy NewBody/LoadAllSettings path with an in-memory settings
// opener. No user's settings files or native windows participate in this test.
func TestIsolateGeometrySurvivesLazyBodySettings(t *testing.T) {
	originalName := core.TheApp.Name()
	originalAll := core.AllSettings
	originalZoom := core.AppearanceSettings.Zoom
	originalDisabled := core.DebugSettings.DisableWindowGeometrySaver
	defer func() {
		core.TheApp.SetName(originalName)
		core.AllSettings = originalAll
		core.AppearanceSettings.Zoom = originalZoom
		core.DebugSettings.DisableWindowGeometrySaver = originalDisabled
	}()
	for _, disabled := range []bool{false, true} {
		f := &geometrySettingsFixture{disabled: disabled, zoom: 1}
		core.AllSettings = []core.Settings{f}
		core.AppearanceSettings.Zoom = 0
		core.DebugSettings.DisableWindowGeometrySaver = !disabled
		restore, err := IsolateGeometry()
		if err != nil {
			t.Fatal(err)
		}
		if f.opens != 1 || !core.DebugSettings.DisableWindowGeometrySaver {
			t.Fatal("initial settings not isolated", f.opens)
		}
		body := core.NewBody("geometry-settings-regression")
		if f.opens != 1 || !core.DebugSettings.DisableWindowGeometrySaver {
			t.Fatal("body reloaded settings after override", f.opens)
		}
		body.Destroy()
		restore()
		if core.DebugSettings.DisableWindowGeometrySaver != disabled {
			t.Fatal("loaded policy not restored")
		}
	}
}
func TestIsolateGeometrySettingsFailureRefusesOverride(t *testing.T) {
	originalAll := core.AllSettings
	originalZoom := core.AppearanceSettings.Zoom
	originalDisabled := core.DebugSettings.DisableWindowGeometrySaver
	defer func() {
		core.AllSettings = originalAll
		core.AppearanceSettings.Zoom = originalZoom
		core.DebugSettings.DisableWindowGeometrySaver = originalDisabled
	}()
	marker := errors.New("fixture settings could not load")
	f := &geometrySettingsFixture{err: marker, zoom: 1}
	core.AllSettings = []core.Settings{f}
	core.AppearanceSettings.Zoom = 0
	core.DebugSettings.DisableWindowGeometrySaver = false
	restore, err := IsolateGeometry()
	if !errors.Is(err, marker) || restore != nil || core.DebugSettings.DisableWindowGeometrySaver {
		t.Fatal("failed settings admitted geometry override", err)
	}
	f.err = nil
	f.zoom = 0
	core.AppearanceSettings.Zoom = 0
	restore, err = IsolateGeometry()
	if err == nil || restore != nil || core.DebugSettings.DisableWindowGeometrySaver {
		t.Fatal("zero zoom admitted")
	}
}

func TestIsolateGeometryAlreadyInitializedAndInvalidZoom(t *testing.T) {
	originalAll := core.AllSettings
	originalZoom := core.AppearanceSettings.Zoom
	originalDisabled := core.DebugSettings.DisableWindowGeometrySaver
	defer func() {
		core.AllSettings = originalAll
		core.AppearanceSettings.Zoom = originalZoom
		core.DebugSettings.DisableWindowGeometrySaver = originalDisabled
	}()
	f := &geometrySettingsFixture{err: errors.New("already-loaded settings must not reopen")}
	core.AllSettings = []core.Settings{f}
	core.AppearanceSettings.Zoom = 100
	core.DebugSettings.DisableWindowGeometrySaver = false
	restore, err := IsolateGeometry()
	if err != nil || f.opens != 0 || !core.DebugSettings.DisableWindowGeometrySaver {
		t.Fatal("initialized settings reopened", f.opens, err)
	}
	restore()
	if core.DebugSettings.DisableWindowGeometrySaver {
		t.Fatal("initialized policy not restored")
	}
	for _, zoom := range []float32{-1, float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))} {
		core.AppearanceSettings.Zoom = zoom
		restore, err = IsolateGeometry()
		if err == nil || restore != nil || f.opens != 0 || core.DebugSettings.DisableWindowGeometrySaver {
			t.Fatal("invalid preloaded zoom admitted", zoom, err)
		}
	}
}
