// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
package cogent

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type fakeDisplayExperiment struct {
	retireEntered    chan struct{}
	refusal          chan error
	enableErr        error
	enabled, retired atomic.Int64
	timeout          time.Duration
}

func (f *fakeDisplayExperiment) EnableDisplayBackpressure(timeout time.Duration) error {
	f.enabled.Add(1)
	f.timeout = timeout
	return f.enableErr
}
func (f *fakeDisplayExperiment) DisplayBackpressureRefusals() <-chan error { return f.refusal }
func (f *fakeDisplayExperiment) RetireDisplayBackpressure() {
	if f.retired.Add(1) == 1 && f.retireEntered != nil {
		close(f.retireEntered)
	}
}
func TestDisplayBackpressureAdapterReportsUntaggedFailureAndJoins(t *testing.T) {
	target := &fakeDisplayExperiment{refusal: make(chan error, 1)}
	reported := make(chan error, 1)
	m, err := attachDisplayBackpressure(target, time.Second, func(err error) { reported <- err })
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.close)
	want := &DisplayBackpressureError{Reason: "controlled lost callback"}
	target.refusal <- want
	select {
	case got := <-reported:
		if got != want {
			t.Fatal(got)
		}
	case <-time.After(time.Second):
		t.Fatal("untagged failure was not reported")
	}
	m.close()
	m.close()
	if target.enabled.Load() != 1 || target.retired.Load() != 1 || target.timeout != time.Second {
		t.Fatal("enable retried or retirement lost")
	}
	select {
	case <-m.done:
	default:
		t.Fatal("failure goroutine still live")
	}
}
func TestDisplayBackpressureAdapterLateEnableRefusalIsNotRetried(t *testing.T) {
	want := &DisplayBackpressureError{Reason: "drawable acquisition already started"}
	target := &fakeDisplayExperiment{enableErr: want}
	m, err := attachDisplayBackpressure(target, time.Second, func(error) { t.Error("unexpected callback") })
	if m != nil || err != want || target.enabled.Load() != 1 || target.retired.Load() != 0 {
		t.Fatal("late enable refusal altered or retried", m, err)
	}
}
func TestDisplayBackpressureAdapterMonitorConstructionFailureRetires(t *testing.T) {
	target := &fakeDisplayExperiment{}
	m, err := attachDisplayBackpressure(target, time.Second, func(error) {})
	if m != nil || err == nil || target.enabled.Load() != 1 || target.retired.Load() != 1 {
		t.Fatal("partial experiment leaked", m, err)
	}
}
func TestDisplayBackpressureAdapterCloseCancelsIdleMonitor(t *testing.T) {
	target := &fakeDisplayExperiment{refusal: make(chan error, 1)}
	var failures atomic.Int64
	m, err := attachDisplayBackpressure(target, time.Second, func(error) { failures.Add(1) })
	if err != nil {
		t.Fatal(err)
	}
	m.close()
	target.refusal <- errors.New("late callback")
	if failures.Load() != 0 || target.retired.Load() != 1 {
		t.Fatal("closed monitor dispatched late failure")
	}
	select {
	case <-m.done:
	default:
		t.Fatal("close did not join monitor")
	}
}
func TestDisplayBackpressureAdapterNativeRetirementClosesMonitor(t *testing.T) {
	target := &fakeDisplayExperiment{refusal: make(chan error, 1)}
	m, err := attachDisplayBackpressure(target, time.Second, func(error) { t.Error("close invented failure") })
	if err != nil {
		t.Fatal(err)
	}
	close(target.refusal)
	select {
	case <-m.done:
	case <-time.After(time.Second):
		t.Fatal("native retirement left monitor blocked")
	}
	m.close()
}
func TestDisplayBackpressureAdapterJoinsInFlightFailureDispatch(t *testing.T) {
	target := &fakeDisplayExperiment{refusal: make(chan error, 1), retireEntered: make(chan struct{})}
	entered, release := make(chan struct{}), make(chan struct{})
	m, err := attachDisplayBackpressure(target, time.Second, func(error) { close(entered); <-release })
	if err != nil {
		t.Fatal(err)
	}
	target.refusal <- errors.New("controlled refusal")
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("dispatch missing")
	}
	stopped := make(chan struct{})
	go func() { defer close(stopped); m.close() }()
	select {
	case <-target.retireEntered:
	case <-time.After(time.Second):
		t.Fatal("close did not reach lifetime retirement")
	}
	select {
	case <-stopped:
		t.Fatal("close failed to join active dispatch")
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("monitor did not retire after dispatch")
	}
}
func TestDisplayBackpressureAdapterOptionBounds(t *testing.T) {
	for _, v := range []time.Duration{0, 100 * time.Millisecond, time.Second, 5 * time.Second} {
		if err := validateDisplayBackpressure(v); err != nil {
			t.Fatal(v, err)
		}
	}
	for _, v := range []time.Duration{-1, 99 * time.Millisecond, 5*time.Second + 1} {
		if err := validateDisplayBackpressure(v); err == nil {
			t.Fatal("invalid experiment timeout accepted", v)
		}
	}
}
