// Copyright (c) 2026, Davin Hills.
// SPDX-License-Identifier: MIT

package main

import (
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/automation"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestAutomationCheckIsolationAdmission(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		enabled, api         bool
		exit                 time.Duration
		conflicting, refused bool
	}{
		{"ordinary", false, false, 0, false, false},
		{"ordinary-flags", false, true, time.Hour, true, false},
		{"minimum", true, true, 5 * time.Second, false, false},
		{"maximum", true, true, 30 * time.Second, false, false},
		{"no-api", true, false, 20 * time.Second, false, true},
		{"short", true, true, 5*time.Second - time.Nanosecond, false, true},
		{"long", true, true, 30*time.Second + time.Nanosecond, false, true},
		{"other-smoke", true, true, 20 * time.Second, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if (validateAutomationCheck(tc.enabled, tc.api, tc.exit, tc.conflicting) != nil) != tc.refused {
				t.Fatal("incorrect isolation admission")
			}
		})
	}
}

func TestAutomationCheckReadinessFailureStillClosesEndpoint(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Mac endpoint")
	}
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	r := automation.NewRegistry(func() (uint64, error) { return 1, nil })
	s, err := automation.Start(automation.Config{Owner: a, Registry: r})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	path := s.Endpoint().Path
	expected := errors.New("fixture not ready")
	ready := make(chan error, 1)
	ready <- expected
	close(ready)
	if err := finishAutomationNativeCheck(ready, s); !errors.Is(err, expected) {
		t.Fatalf("lost readiness failure: %v", err)
	}
	select {
	case <-s.Done():
	default:
		t.Fatal("transport was not joined")
	}
	for _, p := range []string{path, filepath.Dir(path)} {
		if _, err := os.Lstat(p); !os.IsNotExist(err) {
			t.Fatalf("endpoint resource remains: %v", err)
		}
	}
}

type failedEndpointWriter struct{}

func (failedEndpointWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestAutomationCheckMissingReadinessAndWriteFailureCloseEndpoint(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Mac endpoint")
	}
	for _, kind := range []string{"missing", "empty", "write"} {
		t.Run(kind, func(t *testing.T) {
			a, err := app.New(app.Config{})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Close()
			s, err := automation.Start(automation.Config{Owner: a, Registry: automation.NewRegistry(func() (uint64, error) { return 1, nil })})
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			var got error
			switch kind {
			case "missing":
				got = finishAutomationNativeCheck(nil, s)
			case "empty":
				ch := make(chan error)
				close(ch)
				got = finishAutomationNativeCheck(ch, s)
			case "write":
				got = writeAutomationEndpoint(failedEndpointWriter{}, s)
				if !errors.Is(got, io.ErrClosedPipe) {
					t.Fatal("lost write failure")
				}
			}
			if got == nil {
				t.Fatal("failure accepted")
			}
			select {
			case <-s.Done():
			default:
				t.Fatal("transport not joined")
			}
			for _, p := range []string{s.Endpoint().Path, filepath.Dir(s.Endpoint().Path)} {
				if _, err := os.Lstat(p); !os.IsNotExist(err) {
					t.Fatal("endpoint remains")
				}
			}
		})
	}
}

func TestAutomationReadinessRequiresWorkerRetirement(t *testing.T) {
	ready := make(chan error)
	deadline := make(chan time.Time, 1)
	finished := make(chan error, 1)
	go func() { finished <- joinAutomationReadiness(ready, deadline) }()
	// Unbuffered send proves the consumer has observed the result. Keep closure
	// withheld and expire the join; accepting the result alone must fail.
	ready <- nil
	deadline <- time.Now()
	select {
	case err := <-finished:
		if err == nil {
			t.Fatal("result accepted before worker retirement")
		}
	case <-time.After(time.Second):
		t.Fatal("join did not honor deadline")
	}
	close(ready)
}
