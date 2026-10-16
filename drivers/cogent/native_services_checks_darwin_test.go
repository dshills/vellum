//go:build darwin && cgo && !offscreen && vellum_native_checks

package cogent

import (
	"reflect"
	"testing"
)

func TestNativeWindowSnapshotBooleanWire(t *testing.T) {
	raw := nativeWindowSnapshotFixture()
	state, err := decodeNativeServiceSnapshot(raw)
	if err != nil || !state.MainThread || len(state.Windows) != 4 {
		t.Fatal("production Cocoa window snapshot did not round-trip", raw, err)
	}
	for i, w := range state.Windows {
		if w.Title != "fixture" || !w.Visible || w.Miniaturized || !w.Key || w.OcclusionVisible != (i >= 2) || w.Frame != [4]float64{10, 20, 30, 40} {
			t.Fatal("unexpected window state", i, w)
		}
	}
}

func TestNativeServiceSnapshotRejectsPartialDecode(t *testing.T) {
	for _, raw := range []string{
		`{"MainThread":true,"CleanupObserved":true,"Windows":[{"OcclusionVisible":1}]}`,
		`{"MainThread":true,"CleanupObserved":true,"Windows":[{"Visible":"true"}]}`,
		`{"MainThread":true,"CleanupObserved":true,"Windows":[{"Frame":[0,"bad",20,30]}]}`,
		`{"MainThread":true,"CleanupObserved":true,"Windows":[{"IdentityAvailable":1}]}`,
		`{"MainThread":true,"CleanupObserved":true,"Windows":[{"WindowNumber":9223372036854775808}]}`,
	} {
		state, err := decodeNativeServiceSnapshot(raw)
		if err == nil || !reflect.DeepEqual(state, NativeServiceCheckState{}) {
			t.Fatal("invalid wire exposed a partial usable state", raw, state, err)
		}
	}
}
