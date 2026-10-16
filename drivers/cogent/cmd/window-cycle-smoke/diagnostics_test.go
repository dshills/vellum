//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"bytes"
	"encoding/json"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestFailureContextSeparatesDomainRetirementAndUnknownWindow(t *testing.T) {
	a, e := app.New(app.Config{})
	if e != nil {
		t.Fatal(e)
	}
	defer a.Close()
	var log bytes.Buffer
	for _, id := range []string{"active", "survivor"} {
		if e = a.Open(app.Window{ID: id, Title: id, Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }, CloseRequest: closeRequestDiagnostic(id, &log)}); e != nil {
			t.Fatal(e)
		}
	}
	before := observeNativeActionFailure(a, "active")
	if before.ApplicationCanceled || !before.DomainWindowPresent || before.WindowCanceled || before.WindowContextError != "" {
		t.Fatal(before)
	}
	var policyError error
	if e = a.Do(func() { policyError = a.RequestCloseWindow(currentFixtureRef(a, "active")) }); e != nil || policyError != nil {
		t.Fatal(e, policyError)
	}
	after := observeNativeActionFailure(a, "active")
	if after.ApplicationCanceled || after.DomainWindowPresent || !after.WindowCanceled || after.WindowContextError != "" {
		t.Fatal(after)
	}
	var event map[string]any
	if e = json.Unmarshal(log.Bytes(), &event); e != nil || event["window"] != "active" || event["policy_accepts"] != true || event["trigger_actor_known"] != false {
		t.Fatal(e, event)
	}
	log.Reset()
	if e = a.CloseWindow(currentFixtureRef(a, "survivor")); e != nil {
		t.Fatal(e)
	}
	if log.Len() != 0 {
		t.Fatal("forced teardown falsely logged a close policy request")
	}
	if !observeNativeActionFailure(a, "survivor").ApplicationCanceled {
		t.Fatal("application closure missed")
	}
	unknown := observeNativeActionFailure(a, "unknown")
	if unknown.WindowContextError == "" || unknown.DomainWindowPresent || unknown.WindowCanceled {
		t.Fatal(unknown)
	}
}
