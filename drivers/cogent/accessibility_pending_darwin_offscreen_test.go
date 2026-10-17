//go:build darwin && cgo && offscreen

package cogent

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"testing"
)

func pendingAXNativePatch(previous, generation, incarnation, completed, target uint64) map[string]any {
	return map[string]any{"previous": previous, "generation": generation, "scopes": []any{map[string]any{"id": "grid", "incarnation": incarnation, "completed": completed, "target": target}}}
}

func pendingAXNativeRun(t *testing.T, steps []map[string]any) []map[string]any {
	t.Helper()
	for _, step := range steps {
		step["include_nodes"] = true
	}
	payload, err := json.Marshal(steps)
	if err != nil {
		t.Fatal(err)
	}
	var result []map[string]any
	if err := json.Unmarshal([]byte(nativeAccessibilityScopedFixture(string(payload))), &result); err != nil || len(result) != len(steps) {
		t.Fatalf("native pending fixture: %v", err)
	}
	// Object IDs are monotonic across fixture invocations. Preserve all identity
	// relationships within this trace while normalizing its initial serials.
	identities := map[float64]float64{}
	var normalize func(any)
	normalize = func(value any) {
		switch v := value.(type) {
		case map[string]any:
			if id, ok := v["object_identity"].(float64); ok {
				if _, exists := identities[id]; !exists {
					identities[id] = float64(len(identities) + 1)
				}
				v["object_identity"] = identities[id]
			}
			// Fixed traversal order makes serial normalization independent of map order.
			for _, key := range []string{"elements", "retired", "pending_guards"} {
				normalize(v[key])
			}
		case []any:
			for _, item := range v {
				normalize(item)
			}
		}
	}
	for _, result := range result {
		normalize(result)
	}
	return result
}

func TestNativeAccessibilityPendingMatchesFullPublication(t *testing.T) {
	makeSteps := func(fast bool) []map[string]any {
		ready := axEnvelopeNodes(7, 9, 9, false, "", "A😀")
		ready[2]["text_client"], ready[2]["text_selection"], ready[2]["text_focus"] = 13, []int{1, 2}, 3
		pending := axEnvelopeNodes(7, 10, 9, true, "", "A😀")
		repending := axEnvelopeNodes(7, 11, 9, true, "", "A😀")
		readyAgain := axEnvelopeNodes(7, 11, 11, false, "", "B🙂")
		ordinary := map[string]any{"id": "button", "parent": "", "role": "button", "actions": []string{"press"}, "value": "unchanged"}
		for _, nodes := range []*[]map[string]any{&ready, &pending, &repending, &readyAgain} {
			*nodes = append(*nodes, ordinary)
		}
		invalid := axEnvelopeNodes(7, 11, 11, false, "", "bad")
		invalid[2]["parent"] = "missing"
		steps := []map[string]any{
			axEnvelopeUpdate(1, ready), axEnvelopeUpdate(2, pending),
			{"kind": "geometry", "bounds": map[string]any{"cell": []int{999, 999, 1, 1}, "button": []int{5, 6, 7, 8}}},
			axEnvelopeUpdate(3, repending), axEnvelopeUpdate(4, invalid), axEnvelopeUpdate(4, readyAgain),
			axEnvelopeUpdate(5, []map[string]any{ordinary}), axEnvelopeUpdate(6, ready), {"kind": "close"}, axEnvelopeUpdate(7, pending),
		}
		if fast {
			for _, pair := range [][2]int{{1, 10}, {3, 11}} {
				i, target := pair[0], pair[1]
				previous := uint64(1)
				if i == 3 {
					previous = 2
				}
				steps[i]["kind"] = "pending"
				steps[i]["patch"] = pendingAXNativePatch(previous, previous+1, 7, 9, uint64(target))
			}
		}
		return steps
	}
	full, fast := pendingAXNativeRun(t, makeSteps(false)), pendingAXNativeRun(t, makeSteps(true))
	for _, i := range []int{1, 3} {
		if fast[i]["fast_path"] != true {
			t.Fatal("pending path did not execute", i)
		}
		delete(fast[i], "fast_path")
	}
	if !reflect.DeepEqual(full, fast) {
		t.Fatalf("pending transaction differs from full publication:\nfull=%v\nfast=%v", full, fast)
	}
	if fast[4]["accepted"] != false || fast[4]["generation"] != float64(3) || fast[9]["accepted"] != false {
		t.Fatal("refusal or closure changed")
	}
}

func TestNativeAccessibilityPendingPreflightAndFallback(t *testing.T) {
	ready := axEnvelopeNodes(7, 9, 9, false, "", "old")
	pending := axEnvelopeNodes(7, 10, 9, true, "", "old")
	for name, mutate := range map[string]func(map[string]any){
		"stale baseline":   func(p map[string]any) { p["previous"] = 2; p["generation"] = 3 },
		"zero previous":    func(p map[string]any) { p["previous"] = 0 },
		"same generation":  func(p map[string]any) { p["generation"] = 1 },
		"jump generation":  func(p map[string]any) { p["generation"] = 3 },
		"malformed number": func(p map[string]any) { p["generation"] = true },
		"unknown field":    func(p map[string]any) { p["extra"] = 1 },
		"no scopes":        func(p map[string]any) { p["scopes"] = []any{} },
		"malformed scopes": func(p map[string]any) { p["scopes"] = "grid" },
		"missing host":     func(p map[string]any) { p["scopes"].([]any)[0].(map[string]any)["id"] = "missing" },
		"incarnation":      func(p map[string]any) { p["scopes"].([]any)[0].(map[string]any)["incarnation"] = 8 },
		"completed":        func(p map[string]any) { p["scopes"].([]any)[0].(map[string]any)["completed"] = 8 },
		"backward target":  func(p map[string]any) { p["scopes"].([]any)[0].(map[string]any)["target"] = 8 },
		"zero target":      func(p map[string]any) { p["scopes"].([]any)[0].(map[string]any)["target"] = 0 },
		"duplicate host":   func(p map[string]any) { s := p["scopes"].([]any); p["scopes"] = append(s, s[0]) },
		"late invalid scope": func(p map[string]any) {
			p["scopes"] = append(p["scopes"].([]any), map[string]any{"id": "missing", "incarnation": 7, "completed": 9, "target": 10})
		},
	} {
		t.Run(name, func(t *testing.T) {
			patch := pendingAXNativePatch(1, 2, 7, 9, 10)
			mutate(patch)
			out := pendingAXNativeRun(t, []map[string]any{axEnvelopeUpdate(1, ready), {"kind": "pending", "patch": patch}})
			if out[1]["accepted"] != false || out[1]["fast_path"] != false {
				t.Fatal("invalid pending patch accepted")
			}
			delete(out[1], "fast_path")
			out[0]["accepted"] = false
			if !reflect.DeepEqual(out[0], out[1]) {
				t.Fatal("refused patch changed retained state")
			}
			fallback := pendingAXNativeRun(t, []map[string]any{axEnvelopeUpdate(1, ready), {"kind": "pending", "patch": patch, "snapshot": axEnvelopeSnapshot(2, pending)}})
			if fallback[1]["accepted"] != true || fallback[1]["fast_path"] != false || fallback[1]["generation"] != float64(2) {
				t.Fatal("full fallback failed")
			}
		})
	}
	// A rejected native full publication leaves Go ahead of the native epoch.
	// The next patch must fall back to the full payload and recover that gap.
	invalid := axEnvelopeNodes(7, 10, 10, false, "", "bad")
	invalid[2]["parent"] = "missing"
	out := pendingAXNativeRun(t, []map[string]any{axEnvelopeUpdate(1, ready), axEnvelopeUpdate(2, invalid), {"kind": "pending", "patch": pendingAXNativePatch(2, 3, 7, 9, 10), "snapshot": axEnvelopeSnapshot(3, pending)}})
	if out[1]["accepted"] != false || out[2]["accepted"] != true || out[2]["fast_path"] != false || out[2]["generation"] != float64(3) {
		t.Fatal("native rejection recovery failed")
	}
	// Closed and removed hosts cannot be recreated through a stale patch.
	for _, removal := range []map[string]any{{"kind": "close"}, axEnvelopeUpdate(2, nil)} {
		if removal["kind"] == "update" {
			removal["snapshot"] = axEnvelopeSnapshot(2, []map[string]any{})
		}
		out := pendingAXNativeRun(t, []map[string]any{axEnvelopeUpdate(1, ready), removal, {"kind": "pending", "patch": pendingAXNativePatch(2, 3, 7, 9, 10)}})
		if out[2]["accepted"] != false || out[2]["fast_path"] != false {
			t.Fatal("stale pending resurrected host")
		}
	}
}

func TestNativeAccessibilityPendingExactUIntGeneration(t *testing.T) {
	ready := axEnvelopeNodes(7, 9, 9, false, "", "old")
	pending := axEnvelopeNodes(7, 10, 9, true, "", "old")
	for _, previous := range []uint64{1<<53 + 1, 1<<63 + 1, math.MaxUint64 - 1} {
		steps := []map[string]any{axEnvelopeUpdate(previous, ready), {"kind": "pending", "patch": pendingAXNativePatch(previous, previous+1, 7, 9, 10)}}
		payload, err := json.Marshal(steps)
		if err != nil {
			t.Fatal(err)
		}
		var out []struct {
			Accepted   bool
			Generation uint64
			FastPath   bool `json:"fast_path"`
		}
		if err := json.Unmarshal([]byte(nativeAccessibilityScopedFixture(string(payload))), &out); err != nil || len(out) != 2 || !out[1].Accepted || !out[1].FastPath || out[1].Generation != previous+1 {
			t.Fatalf("exact native pending epoch %d: %+v %v", previous, out, err)
		}
	}
	out := axEnvelopeRun(t, axEnvelopeUpdate(math.MaxUint64, ready), map[string]any{"kind": "pending", "patch": pendingAXNativePatch(math.MaxUint64, 0, 7, 9, 10), "snapshot": axEnvelopeSnapshot(0, pending)})
	if out[1].Accepted || out[1].Generation != math.MaxUint64 {
		t.Fatal("overflow wrapped native generation")
	}
}

func TestNativeAccessibilityPendingPrunesAncillaryCachesLikeFull(t *testing.T) {
	makeSteps := func(fast bool) []map[string]any {
		ready := axEnvelopeNodes(7, 9, 9, false, "", "old")
		button := map[string]any{"id": "button", "parent": "", "role": "button", "actions": []string{"press"}}
		steps := []map[string]any{axEnvelopeUpdate(1, append(ready, button))}
		entry := map[string]any{"client": 13, "text": "A😀", "geometry": map[string]any{"viewport": []int{0, 0, 100, 50}, "focus_line": 0, "cells": []any{}, "carets": []any{}}}
		for generation := uint64(2); generation < 6; generation++ {
			orphan := fmt.Sprintf("decorative-%d", generation)
			steps = append(steps, map[string]any{"kind": "geometry", "bounds": map[string]any{orphan: []int{1, 2, 3, 4}, "button": []int{5, 6, 7, 8}}}, map[string]any{"kind": "text_geometry", "entries": map[string]any{orphan: entry, "cell": entry}})
			pending := append(axEnvelopeNodes(7, 9+generation, 9, true, "", "old"), button)
			// Refusal must preserve even unknown caches; cleanup is part of a
			// successful transaction, after every preflight/allocation succeeds.
			if generation == 3 {
				bad := map[string]any{"kind": "update", "snapshot": axEnvelopeSnapshot(generation, []map[string]any{{"id": "bad", "parent": "missing", "role": "button", "actions": []string{}}})}
				if fast {
					bad = map[string]any{"kind": "pending", "patch": pendingAXNativePatch(generation-1, generation, 99, 9, 9+generation)}
				}
				steps = append(steps, bad)
			}
			step := axEnvelopeUpdate(generation, pending)
			if fast {
				step["kind"] = "pending"
				step["patch"] = pendingAXNativePatch(generation-1, generation, 7, 9, 9+generation)
			}
			steps = append(steps, step)
		}
		return steps
	}
	full, fast := pendingAXNativeRun(t, makeSteps(false)), pendingAXNativeRun(t, makeSteps(true))
	for i, result := range fast {
		if branch, exists := result["fast_path"]; exists {
			if branch != result["accepted"] {
				t.Fatal("unexpected fallback in cache parity fixture")
			}
			delete(result, "fast_path")
		}
		if !reflect.DeepEqual(full[i], result) {
			t.Fatalf("ancillary state differs at step %d: full=%v fast=%v", i, full[i], result)
		}
		if result["accepted"] == true && i > 0 {
			geometry, text := result["geometry"].(map[string]any), result["text_geometry"].(map[string]any)
			if len(geometry) != 1 || geometry["button"] == nil || len(text) != 1 || text["cell"] == nil {
				t.Fatal("successful pending update retained orphan or discarded current cache", result)
			}
		}
	}
}
