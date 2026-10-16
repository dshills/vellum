//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"encoding/json"
	"image"
	"testing"
	"time"

	"cogentcore.org/core/core"
	"github.com/dshills/vellum/drivers/cogent"
)

// All cases return before Application.New. No window, native driver, clipboard
// or OS setting is used by these argument/framing checks.
func TestInvalidConfigurationFailsBeforeNativeSetup(t *testing.T) {
	before := core.DebugSettings.DisableWindowGeometrySaver
	for _, test := range []struct {
		name    string
		rows    int
		timeout time.Duration
	}{
		{"row-count", 42, time.Minute}, {"too-short", 10000, 9 * time.Second}, {"too-long", 1000000, 5*time.Minute + time.Second},
	} {
		t.Run(test.name, func(t *testing.T) {
			out, err := run(test.rows, false, test.timeout)
			if err == nil || out.Result != "FAILED_NATIVE_WORKSPACE" || out.Error == "" || out.GeometryPersistenceDisabled || out.WorkerJoined || len(out.Actions) != 0 {
				t.Fatalf("invalid configuration became execution evidence: %+v %v", out, err)
			}
			if core.DebugSettings.DisableWindowGeometrySaver != before {
				t.Fatal("invalid configuration changed geometry persistence")
			}
		})
	}
}

func TestDistinctVisibleCellSkipsRetainedOverscanAndCurrentCursor(t *testing.T) {
	cell := func(id string, bounds image.Rectangle) cogent.GridCellGeometry {
		return cogent.GridCellGeometry{ID: id, RowKey: id, ColumnID: "c00", Bounds: bounds}
	}
	visible := cell("visible", image.Rect(492, 174, 600, 208))
	current := cell("current", image.Rect(492, 208, 600, 248))
	for _, test := range []struct {
		name  string
		cells []cogent.GridCellGeometry
	}{
		{"leading-overscan", []cogent.GridCellGeometry{cell("overscan", image.Rectangle{}), visible}},
		{"current-cursor", []cogent.GridCellGeometry{current, visible}},
		{"zero-width", []cogent.GridCellGeometry{cell("zero-width", image.Rect(492, 174, 492, 208)), visible}},
		{"zero-height", []cogent.GridCellGeometry{cell("zero-height", image.Rect(492, 174, 600, 174)), visible}},
		{"inverted-bounds", []cogent.GridCellGeometry{cell("inverted", image.Rectangle{Min: image.Pt(600, 208), Max: image.Pt(492, 174)}), visible}},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, ok := distinctVisibleCell(test.cells, current.RowKey, current.ColumnID)
			if !ok || got != visible {
				t.Fatalf("selected non-visible/current cell: %+v %v", got, ok)
			}
		})
	}
}

func TestDistinctVisibleCellRetainsPartialClippingAndRefusesNoTarget(t *testing.T) {
	partial := cogent.GridCellGeometry{ID: "partial", RowKey: "r1", ColumnID: "c01", Bounds: image.Rect(492, 174, 493, 175)}
	if got, ok := distinctVisibleCell([]cogent.GridCellGeometry{partial}, "r1", "c00"); !ok || got != partial {
		t.Fatalf("partially visible distinct column refused: %+v %v", got, ok)
	}
	for _, cells := range [][]cogent.GridCellGeometry{nil, {{ID: "overscan"}}, {partial}} {
		if got, ok := distinctVisibleCell(cells, partial.RowKey, partial.ColumnID); ok || got != (cogent.GridCellGeometry{}) {
			t.Fatalf("invented distinct visible target: %+v %v", got, ok)
		}
	}
}

func TestResultUsesNativeEventFramingWithoutQualificationClaims(t *testing.T) {
	out, err := run(0, false, time.Minute)
	if err == nil {
		t.Fatal("invalid configuration accepted")
	}
	raw, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]json.RawMessage
	if err = json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	if string(value["event"]) != `"vellum.workspace-native.result"` || string(value["result"]) != `"FAILED_NATIVE_WORKSPACE"` || string(value["planned_actions"]) != "18" {
		t.Fatalf("native event framing changed: %s", raw)
	}
	for _, key := range []string{"full_p9_qualified", "full_p11_qualified", "physical_pointer_verified"} {
		if string(value[key]) != "false" {
			t.Fatal("missing/incorrect qualification limit", key, string(value[key]))
		}
	}
	for _, key := range []string{"Event", "Result", "Error"} {
		if _, ok := value[key]; ok {
			t.Fatal("noncanonical root field", key)
		}
	}
}
