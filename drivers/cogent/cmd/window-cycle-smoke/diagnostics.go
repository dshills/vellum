//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"encoding/json"
	"fmt"
	"github.com/dshills/vellum/app"
	"io"
	"os"
	"time"
)

// These scalar observations add context to an acceptance failure. The public
// registry reads are individually synchronized, not one atomic lifecycle snapshot.
type nativeActionFailure struct {
	ObservedUTC                                              time.Time
	ApplicationCanceled, DomainWindowPresent, WindowCanceled bool
	WindowContextError                                       string `json:",omitempty"`
}

func observeNativeActionFailure(a *app.Application, window string) nativeActionFailure {
	f := nativeActionFailure{ObservedUTC: time.Now().UTC(), ApplicationCanceled: a.Context().Err() != nil}
	for _, w := range a.Windows() {
		if w.ID == window {
			f.DomainWindowPresent = true
			break
		}
	}
	lifetime, err := a.WindowContext(window)
	if err != nil {
		f.WindowContextError = err.Error()
	} else {
		f.WindowCanceled = lifetime.Err() != nil
	}
	return f
}

// Policy invocation proves a request reached the owner, not who triggered it
// or that Cocoa has already completed closure. Forced teardown bypasses policy.
func closeRequestDiagnostic(window string, out io.Writer) func() bool {
	return func() bool {
		if err := json.NewEncoder(out).Encode(map[string]any{"event": "vellum.window-close-request", "window": window, "observed_utc": time.Now().UTC(), "policy_accepts": true, "trigger_actor_known": false}); err != nil {
			fmt.Fprintln(os.Stderr, "window close diagnostic:", err)
		}
		return true
	}
}
