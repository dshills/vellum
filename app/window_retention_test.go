package app

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestClosedWindowLifetimesCompactWithoutAcceptingStaleWork(t *testing.T) {
	a, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.Open(sampleWindow("survivor")); err != nil {
		t.Fatal(err)
	}
	survivor, _ := a.WindowContext("survivor")
	markers := make(map[context.Context]bool)
	for i := 0; i < 1000; i++ {
		id := fmt.Sprintf("cycle-%04d", i)
		if err = a.Open(sampleWindow(id)); err != nil {
			t.Fatal(err)
		}
		original, _ := a.WindowContext(id)
		child, cancel := context.WithCancel(original)
		cleaned := 0
		reserved, err := a.ReserveCompletion(id, func(err error) {
			if !errors.Is(err, ErrWindowClosed) {
				t.Errorf("wrong revocation: %v", err)
			}
			cleaned++
		})
		if err != nil {
			t.Fatal(err)
		}
		if err = a.PostWindow(id, func() { t.Error("stale delivery ran") }); err != nil {
			t.Fatal(err)
		}
		if err = a.CloseWindow(id); err != nil {
			t.Fatal(err)
		}
		if !errors.Is(original.Err(), context.Canceled) || !errors.Is(child.Err(), context.Canceled) || cleaned != 1 {
			t.Fatalf("captured work survived cycle %d", i)
		}
		cancel()
		marker, err := a.WindowContext(id)
		if err != nil || !errors.Is(marker.Err(), context.Canceled) {
			t.Fatalf("closed identity lookup: %v", err)
		}
		markers[marker] = true
		if a.Open(sampleWindow(id)) == nil {
			t.Fatal("closed ID reused")
		}
		if a.CloseWindow(id) == nil {
			t.Fatal("duplicate close accepted")
		}
		if err = a.PostWindow(id, func() {}); !errors.Is(err, ErrWindowClosed) {
			t.Fatalf("late delivery accepted: %v", err)
		}
		if err = reserved.Publish(func() {}); !errors.Is(err, ErrWindowClosed) {
			t.Fatalf("late reserved result accepted: %v", err)
		}
		if _, err = a.ReserveCompletion(id, nil); !errors.Is(err, ErrWindowClosed) {
			t.Fatalf("new closed reservation accepted: %v", err)
		}
		if _, err = a.DrainN(1); err != nil {
			t.Fatal(err)
		}
		if survivor.Err() != nil || a.Context().Err() != nil || len(a.Windows()) != 1 {
			t.Fatal("survivor lost")
		}
	}
	if len(a.windowCancels) != 1 || len(markers) != 1 {
		t.Fatalf("closed contexts/cancel functions retained: cancel entries=%d distinct closed contexts=%d; want 1 survivor/1 marker", len(a.windowCancels), len(markers))
	}
	if len(a.lifetimes) != 1001 {
		t.Fatal("ID tombstones lost")
	}
	if _, err = a.WindowContext("unknown"); err == nil {
		t.Fatal("unknown ID accepted")
	}
	if err = a.CloseWindow("survivor"); err != nil {
		t.Fatal(err)
	}
	if survivor.Err() == nil || a.Context().Err() == nil || len(a.windowCancels) != 0 {
		t.Fatal("last close did not retire survivor")
	}
}
