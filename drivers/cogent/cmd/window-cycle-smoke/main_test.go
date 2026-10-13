//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
)

func TestProbeDeadlineReachesBlockingCheck(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := waitWithin(parent, 20*time.Millisecond, func(inner context.Context) (bool, error) {
		if _, ok := inner.Deadline(); !ok {
			t.Fatal("probe received unbounded parent context")
		}
		<-inner.Done()
		return false, inner.Err()
	})
	if !errors.Is(err, context.DeadlineExceeded) || parent.Err() != nil {
		t.Fatalf("probe deadline/caller lifetime: %v %v", err, parent.Err())
	}
}

func TestOwnerAdmissionCannotSubstituteForAcknowledgement(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	called := false
	_, err = ownerProbe(ctx, a, func() ownerReply { called = true; return ownerReply{} })
	if !errors.Is(err, context.DeadlineExceeded) || called {
		t.Fatalf("admitted but undrained callback qualified: %v called=%t", err, called)
	}
	// An expired request may still be queued. It must skip native work and its
	// buffered late acknowledgment must not strand the application owner.
	drained := make(chan struct{})
	go func() { a.Drain(); close(drained) }()
	select {
	case <-drained:
	case <-time.After(time.Second):
		t.Fatal("late acknowledgment blocked owner")
	}
	if called || !a.Pending().Idle() {
		t.Fatal("expired owner probe executed or remained pending")
	}
}

func TestOwnerAcknowledgementCarriesCompletedFailure(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	marker := errors.New("completed owner check")
	finished := make(chan ownerReply, 1)
	go func() {
		value, e := ownerProbe(ctx, a, func() ownerReply { return ownerReply{err: marker} })
		if e != nil {
			value.err = e
		}
		finished <- value
	}()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case value := <-finished:
			if !errors.Is(value.err, marker) {
				t.Fatalf("lost completed owner result: %v", value.err)
			}
			return
		case <-ticker.C:
			a.Drain()
		case <-ctx.Done():
			t.Fatal("acknowledged probe did not join")
		}
	}
}

func TestSampledMemoryGuardsRetainBoundary(t *testing.T) {
	if memoryGuard(observation{HeapAlloc: 512 << 20, RSSBytes: 1 << 30}) != nil {
		t.Fatal("exact sampled bounds refused")
	}
	for _, s := range []observation{{HeapAlloc: 512<<20 + 1}, {RSSBytes: 1<<30 + 1}} {
		if memoryGuard(s) == nil {
			t.Fatal("sample above safety bound accepted")
		}
	}
}
func TestCycleConfigurationBounds(t *testing.T) {
	if err := (config{cycles: 1000, every: 100, timeout: 12 * time.Minute}).validate(); err != nil {
		t.Fatal(err)
	}
	for _, c := range []config{{cycles: 0, every: 1, timeout: time.Minute}, {cycles: 1001, every: 100, timeout: time.Minute}, {cycles: 10, every: 0, timeout: time.Minute}, {cycles: 10, every: 11, timeout: time.Minute}, {cycles: 1, every: 1, timeout: time.Second}, {cycles: 1, every: 1, timeout: 21 * time.Minute}, {cycles: 1, every: 1, timeout: time.Minute, heapProfile: "relative.profile"}} {
		if c.validate() == nil {
			t.Fatalf("accepted invalid config: %+v", c)
		}
	}
}

func TestHeapProfileRefusesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "existing.profile")
	if err := os.WriteFile(path, []byte("preserve existing bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := writeHeapProfile(path); !errors.Is(err, os.ErrExist) {
		t.Fatalf("profile overwrote/refused wrongly: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "preserve existing bytes" {
		t.Fatal("existing profile bytes changed")
	}
}
