//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	activeRef, e := a.WindowRef("active")
	if e != nil || activeRef.Context().Err() != nil {
		t.Fatal("capture live opening", e)
	}
	before := observeNativeActionFailure(a, "active")
	if before.ApplicationCanceled || !before.DomainWindowPresent || before.WindowCanceled || before.WindowContextError != "" {
		t.Fatal(before)
	}
	var policyError error
	if e = a.Do(func() { policyError = a.RequestCloseWindow(activeRef) }); e != nil || policyError != nil {
		t.Fatal(e, policyError)
	}
	after := observeNativeActionFailure(a, "active")
	if after.ApplicationCanceled || after.DomainWindowPresent || after.WindowCanceled || after.WindowContextError == "" {
		t.Fatal(after)
	}
	if !errors.Is(activeRef.Context().Err(), context.Canceled) {
		t.Fatal("captured opening did not cancel", activeRef.Context().Err())
	}
	if _, err := a.WindowContext("active"); !errors.Is(err, app.ErrWindowClosed) {
		t.Fatal("retired name lookup must refuse", err)
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
