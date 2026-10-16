//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"encoding/json"
	"testing"
	"time"

	"cogentcore.org/core/core"
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
