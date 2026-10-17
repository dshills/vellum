// Copyright (c) 2026, Davin Hills.
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dshills/vellum/automation"
)

var testEndpoint = automation.Endpoint{Path: "/private/tmp/vellum-api-test/api.sock", Token: strings.Repeat("ab", 32), Version: automation.Version}

const readyLine = `{"event":"vellum.automation.native.ready","geometry_persistence_disabled_in_fixture":true,"native_window":true,"width":600,"height":400}`

func descriptor() string { data, _ := json.Marshal(testEndpoint); return string(data) }
func writeCapture(t *testing.T, c *capture, lines ...string) {
	t.Helper()
	for _, line := range lines {
		if _, err := c.Write([]byte(line + "\n")); err != nil {
			t.Fatal(err)
		}
	}
}
func TestCaptureBoundedPrivateFraming(t *testing.T) {
	c := newCapture()
	wire := descriptor() + "\n" + readyLine + "\n"
	for _, b := range []byte(wire) {
		if _, err := c.Write([]byte{b}); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ep, err := c.waitReady(ctx)
	if err != nil || ep != testEndpoint {
		t.Fatal("fragmented descriptor not admitted")
	}
	event := observation{"check:01", "inspect", "component", "ok"}
	terminal, _ := json.Marshal(struct {
		Event string           `json:"event"`
		Done  bool             `json:"transport_done"`
		Audit automation.Audit `json:"audit"`
	}{"vellum.automation.native.closed", true, automation.Audit{Events: []automation.Event{{ID: event.ID, Operation: event.Operation, Target: event.Target, Code: event.Code}}}})
	// Even diagnostics containing an arbitrary secret are discarded, not logged.
	writeCapture(t, c, "diagnostic "+testEndpoint.Token, string(terminal))
	if err := c.complete([]observation{event}); err != nil {
		t.Fatal(err)
	}
	if err := c.complete([]observation{{ID: "wrong"}}); err == nil {
		t.Fatal("mismatched audit accepted")
	}
	encoded, _ := json.Marshal(summary{Status: "PASS"})
	if bytes.Contains(encoded, []byte(testEndpoint.Token)) || bytes.Contains(encoded, []byte(testEndpoint.Path)) {
		t.Fatal("private descriptor published")
	}
}
func TestCaptureRejectsBadReadinessAndEndpointSequences(t *testing.T) {
	for name, lines := range map[string][]string{
		"duplicate-endpoint":    {descriptor(), descriptor()},
		"second-secret":         {descriptor(), strings.Replace(descriptor(), testEndpoint.Token, strings.Repeat("cd", 32), 1)},
		"duplicate-key":         {strings.Replace(descriptor(), `"token":`, `"token":"hidden","token":`, 1)},
		"extra-field":           {strings.TrimSuffix(descriptor(), "}") + `,"extra":true}`},
		"malformed-descriptor":  {`{"token":17}`},
		"escaped-capability":    {strings.Replace(descriptor(), `"abab`, `"\u0061bab`, 1)},
		"ready-before-endpoint": {readyLine},
		"ready-not-native":      {descriptor(), strings.Replace(readyLine, `"native_window":true`, `"native_window":false`, 1)},
		"ready-not-isolated":    {descriptor(), strings.Replace(readyLine, `"geometry_persistence_disabled_in_fixture":true`, `"geometry_persistence_disabled_in_fixture":false`, 1)},
		"ready-zero-size":       {descriptor(), strings.Replace(readyLine, `"width":600`, `"width":0`, 1)},
		"duplicate-ready-field": {descriptor(), strings.Replace(readyLine, `"native_window":true`, `"native_window":false,"native_window":true`, 1)},
		"duplicate-ready":       {descriptor(), readyLine, readyLine},
		"close-before-ready":    {descriptor(), `{"event":"vellum.automation.native.closed","transport_done":true,"audit":{}}`},
	} {
		t.Run(name, func(t *testing.T) {
			c := newCapture()
			var err error
			for _, line := range lines {
				_, err = c.Write([]byte(line + "\n"))
				if err != nil {
					break
				}
			}
			if err == nil {
				t.Fatal("bad framing admitted")
			}
			if strings.Contains(err.Error(), testEndpoint.Token) {
				t.Fatal("credential leaked in error")
			}
		})
	}
}
func TestCaptureLimitsAndMissingEvidence(t *testing.T) {
	c := newCapture()
	if _, err := c.Write(bytes.Repeat([]byte{'x'}, maxLineBytes+1)); err == nil {
		t.Fatal("long line admitted")
	}
	c = newCapture()
	line := bytes.Repeat([]byte{'x'}, maxLineBytes)
	line = append(line, '\n')
	var err error
	for i := 0; i < 129; i++ {
		_, err = c.Write(line)
		if err != nil {
			break
		}
	}
	if err == nil {
		t.Fatal("unbounded stream admitted")
	}
	c = newCapture()
	writeCapture(t, c, descriptor(), strings.Replace(readyLine, "native.ready", "native.wrong", 1))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = c.waitReady(ctx); err == nil {
		t.Fatal("missing ready accepted")
	}
	if c.complete(nil) == nil {
		t.Fatal("missing close accepted")
	}
	d := &discardBounded{}
	if _, err = d.Write(make([]byte, maxStreamBytes+1)); err == nil {
		t.Fatal("unbounded stderr admitted")
	}
}
func TestInvokeTransportFailureIsNotRetried(t *testing.T) {
	calls := 0
	s := &session{send: func(context.Context, automation.Endpoint, automation.Request) (automation.Response, error) {
		calls++
		return automation.Response{}, errors.New("response lost after commit")
	}}
	if s.invoke(context.Background(), "component.rotate", `{}`, "ok", 1) == nil || calls != 1 {
		t.Fatal("invoke retried or transport failure accepted")
	}
}

func TestRetireChildIndependentFailurePaths(t *testing.T) {
	done := make(chan struct{})
	terms, kills := 0, 0
	err := retireChild(func() error { terms++; return errors.New("private-term-detail") }, func() error { kills++; close(done); return nil }, done, time.Millisecond)
	if err == nil || err.Error() != "native child TERM failed" || terms != 1 || kills != 1 {
		t.Fatal("TERM failure suppressed or fallback omitted")
	}
	done = make(chan struct{})
	terms, kills = 0, 0
	err = retireChild(func() error { terms++; return errors.New("private") }, func() error { kills++; return errors.New("private") }, done, time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "native child final join unproven") || !strings.Contains(err.Error(), "native child Kill failed") || strings.Contains(err.Error(), "private") || terms != 1 || kills != 1 {
		t.Fatal("final reap not bounded or error lost")
	}
	close(done)
	if retireChild(func() error { t.Fatal("signaled retired process"); return nil }, func() error { t.Fatal("killed retired process"); return nil }, done, time.Millisecond) != nil {
		t.Fatal("retired process refused")
	}
	done = make(chan struct{})
	if retireChild(func() error { close(done); return nil }, func() error { t.Fatal("unnecessary escalation"); return nil }, done, time.Millisecond) != nil {
		t.Fatal("normal TERM refused")
	}
}

func TestStrictRejectsNestedDuplicates(t *testing.T) {
	var out any
	for _, raw := range []string{`{"audit":{"Events":[],"Events":[]}}`, `{"items":[{"a":1,"a":2}]}`} {
		if strict([]byte(raw), &out) == nil {
			t.Fatal("nested duplicate admitted")
		}
	}
}

func TestPostCloseProbeCannotPassFromCancellationOrOtherErrors(t *testing.T) {
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	call := func(ctx context.Context, _ automation.Endpoint, _ automation.Request) (automation.Response, error) {
		called = true
		if ctx.Err() != nil {
			t.Fatal("probe reused canceled context")
		}
		return automation.Response{}, os.ErrNotExist
	}
	if probeClosedEndpoint(canceled, testEndpoint, call) == nil || called {
		t.Fatal("canceled run reported success")
	}
	if err := probeClosedEndpoint(context.Background(), testEndpoint, call); err != nil || !called {
		t.Fatal("missing socket not verified")
	}
	active, stop := context.WithCancel(context.Background())
	defer stop()
	if probeClosedEndpoint(active, testEndpoint, func(ctx context.Context, _ automation.Endpoint, _ automation.Request) (automation.Response, error) {
		if ctx.Err() != nil {
			t.Fatal("probe context canceled")
		}
		stop()
		return automation.Response{}, os.ErrNotExist
	}) == nil {
		t.Fatal("cancellation during probe passed")
	}
	for _, failure := range []error{nil, context.DeadlineExceeded, errors.New("protocol")} {
		if probeClosedEndpoint(context.Background(), testEndpoint, func(context.Context, automation.Endpoint, automation.Request) (automation.Response, error) {
			return automation.Response{}, failure
		}) == nil {
			t.Fatal("unrelated error passed")
		}
	}
}
