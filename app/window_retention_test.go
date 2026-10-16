package app

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestClosedWindowLifetimesRetireWithoutHistoricalNames(t *testing.T) {
	for _, reuse := range []bool{false, true} {
		t.Run(fmt.Sprint(reuse), func(t *testing.T) {
			a, _ := New(Config{})
			defer a.Close()
			survivor, err := a.OpenWindow(sampleWindow("survivor"))
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 1000; i++ {
				id := fmt.Sprintf("cycle-%04d", i)
				if reuse {
					id = "cycle"
				}
				ref, err := a.OpenWindow(sampleWindow(id))
				if err != nil {
					t.Fatal(err)
				}
				child, cancel := context.WithCancel(ref.Context())
				cleaned := 0
				reserved, err := a.ReserveWindowCompletion(ref, func(err error) {
					if !errors.Is(err, ErrWindowClosed) {
						t.Errorf("wrong revocation: %v", err)
					}
					cleaned++
				})
				if err != nil {
					t.Fatal(err)
				}
				if err = a.PostWindow(ref, func() { t.Error("retired delivery ran") }); err != nil {
					t.Fatal(err)
				}
				if err = a.CloseWindow(ref); err != nil {
					t.Fatal(err)
				}
				if ref.Context().Err() != context.Canceled || child.Err() != context.Canceled || cleaned != 1 {
					t.Fatalf("captured work survived cycle %d", i)
				}
				cancel()
				if _, err = a.WindowRef(id); !errors.Is(err, ErrWindowClosed) {
					t.Fatal("retired name still registered", err)
				}
				if _, err = a.WindowContext(id); !errors.Is(err, ErrWindowClosed) {
					t.Fatal("historical context lookup admitted", err)
				}
				if err = a.CloseWindow(ref); !errors.Is(err, ErrWindowClosed) {
					t.Fatal("duplicate close admitted", err)
				}
				if err = a.PostWindow(ref, func() {}); !errors.Is(err, ErrWindowClosed) {
					t.Fatal("late delivery admitted", err)
				}
				if err = reserved.Publish(func() {}); !errors.Is(err, ErrWindowClosed) {
					t.Fatal("late reserved result admitted", err)
				}
				if _, err = a.ReserveWindowCompletion(ref, nil); !errors.Is(err, ErrWindowClosed) {
					t.Fatal("new retired reservation admitted", err)
				}
				if _, err = a.DrainN(1); err != nil {
					t.Fatal(err)
				}
				if survivor.Context().Err() != nil || a.Context().Err() != nil || len(a.Windows()) != 1 {
					t.Fatal("survivor lost")
				}
				if len(a.lifetimes) != 1 || len(a.windowCancels) != 1 || len(a.closeRequests) != 0 {
					t.Fatalf("retired registration retained at cycle %d: lifetimes=%d cancels=%d policies=%d", i, len(a.lifetimes), len(a.windowCancels), len(a.closeRequests))
				}
			}
			if err = a.CloseWindow(survivor); err != nil {
				t.Fatal(err)
			}
			if survivor.Context().Err() == nil || a.Context().Err() == nil || len(a.lifetimes) != 0 || len(a.windowCancels) != 0 || len(a.windows) != 0 || len(a.dirty) != 0 {
				t.Fatal("last close retained registrations")
			}
		})
	}
}
