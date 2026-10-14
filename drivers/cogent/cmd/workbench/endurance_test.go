package main

import (
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestEnduranceConfigurationAndTraceBounds(t *testing.T) {
	if err := (enduranceConfig{}).validate(); err != nil {
		t.Fatal(err)
	}
	good := enduranceConfig{duration: 10 * time.Minute, output: filepath.Join(t.TempDir(), "new")}
	if err := good.validate(); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []enduranceConfig{{duration: time.Second, output: good.output}, {duration: 13 * time.Hour, output: good.output}, {duration: good.duration, output: "relative"}, {output: good.output}, {duration: good.duration}} {
		if bad.validate() == nil {
			t.Fatalf("invalid configuration accepted:%+v", bad)
		}
	}
	f, err := os.CreateTemp(t.TempDir(), "bounded")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := &cappedTrace{file: f, remaining: 3}
	if n, err := w.Write([]byte("abc")); n != 3 || err != nil {
		t.Fatal(n, err)
	}
	if n, err := w.Write([]byte("d")); n != 0 || err == nil || w.failure() == nil {
		t.Fatal("trace limit not retained")
	}
	data, err := os.ReadFile(f.Name())
	if err != nil || string(data) != "abc" {
		t.Fatal("trace limit overwrote file", err)
	}
}
func TestEnduranceOwnerWaitsForExecution(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	called := false
	if err := enduranceOwner(ctx, a, func() error { called = true; return nil }); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("queuedadmission substituted for execution:%v", err)
	}
	if called {
		t.Fatal("owner action executed without drain")
	}
	// Completion of an admitted callback returns its own error.
	marker := errors.New("owner error")
	reply := make(chan error, 1)
	go func() { reply <- enduranceOwner(context.Background(), a, func() error { return marker }) }()
	limit := time.After(time.Second)
	for a.Pending().Ordinary < 2 {
		select {
		case <-limit:
			t.Fatal("callback neveradmitted")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	a.Drain()
	if called {
		t.Fatal("expired owner callback mutated after controller timeout")
	}
	select {
	case err := <-reply:
		if !errors.Is(err, marker) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("owner acknowledgement missing")
	}
}

func TestEnduranceJoinsRunningOwnerCallback(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	result := make(chan error, 1)
	go func() { result <- enduranceOwner(ctx, a, func() error { close(entered); <-release; return nil }) }()
	deadline := time.After(time.Second)
	for a.Pending().Ordinary == 0 {
		select {
		case <-deadline:
			t.Fatal("notadmitted")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	drained := make(chan struct{})
	go func() { a.Drain(); close(drained) }()
	<-entered
	cancel()
	select {
	case <-result:
		t.Fatal("controller returned while owner callback was running")
	case <-time.After(20 * time.Millisecond):
	}
	unblock()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("runningowner callback never joined")
	}
	<-drained
}

func TestEndurancePreservesNativeFailureBeforeCancellation(t *testing.T) {
	native := errors.New("native loop failure")
	done := make(chan error, 1)
	done <- context.Canceled
	err := enduranceRunResult(native, done, time.Second)
	if !errors.Is(err, native) || !errors.Is(err, context.Canceled) || !strings.HasPrefix(err.Error(), native.Error()) {
		t.Fatalf("native cause hidden:%v", err)
	}
	done = make(chan error, 1)
	done <- nil
	if err := enduranceRunResult(nil, done, time.Second); err != nil {
		t.Fatal(err)
	}
	if err := enduranceRunResult(native, make(chan error), time.Nanosecond); !errors.Is(err, native) {
		t.Fatalf("join timeout hid native cause:%v", err)
	}
}
