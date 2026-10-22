// Copyright (c) 2026 Davin Hills. SPDX-License-Identifier: MIT.
package main

import "testing"

func TestSuccessfulCohortDoesNotCollectFailureContext(t *testing.T) {
	out := map[string]any{"existing": "preserved"}
	for sequence := 1; sequence <= 248; sequence++ {
		if err := cohortStateError(out, sequence, nil, true, func() map[string]any {
			t.Fatal("successful input called native failure observer")
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	if len(out) != 1 || out["existing"] != "preserved" {
		t.Fatal("successful inputs changed journal", out)
	}
}

func TestFailureKeepsFirstStateAcrossLaterTeardown(t *testing.T) {
	out := map[string]any{}
	calls := 0
	context := func() map[string]any { calls++; return map[string]any{"front_bundle": "example.other"} }
	for i, state := range []map[string]any{{"front_pid": 42, "key": false}, {"front_pid": 99, "key": false}} {
		if err := cohortStateError(out, i+33, state, false, context); err == nil || err.Error() != "foreground/configuration changed during cohort" {
			t.Fatal("foreground loss was not rejected", err)
		}
	}
	failure := out["foreground_failure"].(map[string]any)
	if calls != 1 || failure["sequence"] != 33 || failure["state"].(map[string]any)["front_pid"] != 42 || failure["context"].(map[string]any)["front_bundle"] != "example.other" {
		t.Fatal("first failure replaced or context collected twice", out, calls)
	}
}
