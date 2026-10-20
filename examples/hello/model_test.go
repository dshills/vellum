package main

import (
	"cogentcore.org/core/core"
	"errors"
	"github.com/dshills/vellum/ui"
	"math"
	"testing"
)

func TestCounterActionPreservesIDs(t *testing.T) {
	c := &counter{}
	content := ui.Node{ID: "body", Kind: ui.Label, Text: "Ready"}
	before := c.view(content)
	if err := ui.Validate(before); err != nil {
		t.Fatal(err)
	}
	before.Children[2].Activate()
	after := c.view(content)
	if after.Children[1].Text != "Count: 1" {
		t.Fatal(after.Children[1].Text)
	}
	for i := range before.Children {
		if before.Children[i].ID != after.Children[i].ID {
			t.Fatal("unstable widget IDs")
		}
	}
}
func BenchmarkCounterViewValidation(b *testing.B) {
	c := &counter{}
	content := ui.Node{ID: "body", Kind: ui.Label, Text: "Ready"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := ui.Validate(c.view(content)); err != nil {
			b.Fatal(err)
		}
	}
}

// In-memory settings use Core's real LoadAllSettings/NewBody path without opening
// user settings, creating native windows, or calling a settings-save operation.
type benchmarkGeometrySettings struct {
	core.SettingsBase
	opens    int
	err      error
	disabled bool
	zoom     float32
}

func (f *benchmarkGeometrySettings) Open() error { f.opens++; return f.err }
func (f *benchmarkGeometrySettings) Apply() {
	core.AppearanceSettings.Zoom = f.zoom
	core.DebugSettings.DisableWindowGeometrySaver = f.disabled
}
func preserveBenchmarkSettings(t *testing.T) {
	t.Helper()
	name, all := core.TheApp.Name(), core.AllSettings
	zoom, disabled := core.AppearanceSettings.Zoom, core.DebugSettings.DisableWindowGeometrySaver
	t.Cleanup(func() {
		core.TheApp.SetName(name)
		core.AllSettings = all
		core.AppearanceSettings.Zoom = zoom
		core.DebugSettings.DisableWindowGeometrySaver = disabled
	})
}
func TestBenchmarkGeometrySurvivesLazyBodySettings(t *testing.T) {
	preserveBenchmarkSettings(t)
	for _, disabled := range []bool{false, true} {
		f := &benchmarkGeometrySettings{disabled: disabled, zoom: 100}
		core.AllSettings = []core.Settings{f}
		core.AppearanceSettings.Zoom = 0
		core.DebugSettings.DisableWindowGeometrySaver = !disabled
		restore, err := isolateBenchmarkGeometry()
		if err != nil {
			t.Fatal(err)
		}
		if !core.DebugSettings.DisableWindowGeometrySaver {
			t.Fatal("benchmark override not installed")
		}
		body := core.NewBody("hello-benchmark-geometry-regression")
		if f.opens != 1 || !core.DebugSettings.DisableWindowGeometrySaver {
			t.Fatal("lazy Body reloaded settings over benchmark override", f.opens)
		}
		body.Destroy()
		if !core.DebugSettings.DisableWindowGeometrySaver {
			t.Fatal("override ended before body teardown")
		}
		restore()
		if core.DebugSettings.DisableWindowGeometrySaver != disabled {
			t.Fatal("loaded policy not restored")
		}
	}
}
func TestBenchmarkGeometryLoadFailureAndZeroRefuseOverride(t *testing.T) {
	for _, loadFailure := range []bool{true, false} {
		name := "zero-loaded-zoom"
		if loadFailure {
			name = "settings-load-error"
		}
		t.Run(name, func(t *testing.T) {
			preserveBenchmarkSettings(t)
			marker := errors.New("in-memory settings failed")
			f := &benchmarkGeometrySettings{}
			if loadFailure {
				f.err, f.zoom = marker, 100
			}
			core.AllSettings = []core.Settings{f}
			core.AppearanceSettings.Zoom = 0
			core.DebugSettings.DisableWindowGeometrySaver = true
			restore, err := isolateBenchmarkGeometry()
			if err == nil || restore != nil || f.opens != 1 || core.DebugSettings.DisableWindowGeometrySaver {
				t.Fatal("failed settings admitted override or rolled back loaded policy", err)
			}
			if loadFailure && !errors.Is(err, marker) {
				t.Fatal("settings error identity lost", err)
			}
			// No app/body follows a failed initialization; tests restore global fixtures.
		})
	}
}
func TestBenchmarkGeometryInitializedAndInvalidZoom(t *testing.T) {
	preserveBenchmarkSettings(t)
	f := &benchmarkGeometrySettings{err: errors.New("initialized settings must not reopen")}
	core.AllSettings = []core.Settings{f}
	core.AppearanceSettings.Zoom = 100
	core.DebugSettings.DisableWindowGeometrySaver = false
	restore, err := isolateBenchmarkGeometry()
	if err != nil || f.opens != 0 || !core.DebugSettings.DisableWindowGeometrySaver {
		t.Fatal("initialized settings reopened", err)
	}
	restore()
	if core.DebugSettings.DisableWindowGeometrySaver {
		t.Fatal("initialized policy not restored")
	}
	for _, zoom := range []float32{-1, float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))} {
		core.AppearanceSettings.Zoom = zoom
		restore, err = isolateBenchmarkGeometry()
		if err == nil || restore != nil || f.opens != 0 || core.DebugSettings.DisableWindowGeometrySaver {
			t.Fatal("invalid initialized zoom admitted", zoom, err)
		}
	}
}
