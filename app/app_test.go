package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/dshills/vellum/ui"
	"sync"
	"testing"
)

func TestBoundedFIFOAndClose(t *testing.T) {
	a, _ := New(Config{QueueCapacity: 2})
	defer a.Close()
	var got []int
	for i := 0; i < 2; i++ {
		i := i
		if err := a.Post(func() { got = append(got, i) }); err != nil {
			t.Fatal(err)
		}
	}
	if !errors.Is(a.Post(func() {}), ErrQueueFull) {
		t.Fatal("unbounded admission")
	}
	a.Drain()
	if len(got) != 2 || got[0] != 0 || got[1] != 1 {
		t.Fatal(got)
	}
	a.Post(func() { t.Fatal("callback after close") })
	a.Close()
	a.Drain()
	if !errors.Is(a.Post(func() {}), ErrClosed) || a.Context().Err() == nil {
		t.Fatal("lifetime still active")
	}
}
func TestSerializedOwner(t *testing.T) {
	a, _ := New(Config{})
	defer a.Close()
	var wg sync.WaitGroup
	n := 0
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				a.Do(func() { n++ })
			}
		}()
	}
	wg.Wait()
	if n != 2000 {
		t.Fatal(n)
	}
}

type testDriver struct{}

func (testDriver) Run(ctx context.Context, a *Application) error { a.Drain(); return nil }
func TestWindowLifetime(t *testing.T) {
	a, _ := New(Config{})
	w := Window{ID: "main", Title: "test", Width: 640, Height: 480, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}
	if err := a.Open(w); err != nil {
		t.Fatal(err)
	}
	if a.Open(w) == nil {
		t.Fatal("duplicate accepted")
	}
	if err := a.Run(context.Background(), testDriver{}); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(a.Run(context.Background(), testDriver{}), ErrClosed) {
		t.Fatal("run again")
	}
}
func TestDrainIsFinite(t *testing.T) {
	a, _ := New(Config{})
	defer a.Close()
	n := 0
	var repeat func()
	repeat = func() { n++; a.Post(repeat) }
	a.Post(repeat)
	a.Drain()
	if n != 1 {
		t.Fatal(n)
	}
	a.Drain()
	if n != 2 {
		t.Fatal(n)
	}
}

func TestPanicClosesOwner(t *testing.T) {
	a, _ := New(Config{})
	a.Post(func() { panic("failed transaction") })
	a.Post(func() { t.Fatal("tail executed") })
	func() {
		defer func() {
			if recover() == nil {
				t.Error("panic suppressed")
			}
		}()
		a.Drain()
	}()
	if !errors.Is(a.Do(func() {}), ErrClosed) {
		t.Fatal("closed action undiagnosed")
	}
}

func TestCloseWakesDriverOnce(t *testing.T) {
	a, _ := New(Config{})
	calls := 0
	a.SetWake(func() { calls++ })
	a.Close()
	a.Close()
	if calls != 1 {
		t.Fatal(calls)
	}
}

func TestDirectActionPanicClosesOwner(t *testing.T) {
	a, _ := New(Config{})
	func() {
		defer func() {
			if recover() == nil {
				t.Error("panic swallowed")
			}
		}()
		a.Do(func() { panic("direct") })
	}()
	if a.Context().Err() == nil {
		t.Fatal("owner left open")
	}
}

func TestCloseRequestRefusalCheckAndForce(t *testing.T) {
	a, _ := New(Config{})
	defer a.Close()
	calls := 0
	accept := false
	var ref WindowRef
	w := Window{ID: "main", Title: "Document", Width: 400, Height: 300, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }, CloseRequest: func() bool {
		calls++
		if !errors.Is(a.CheckCloseWindow(ref), ErrClosePending) {
			t.Fatal("recursive policy admitted")
		}
		return accept
	}}
	if err := a.Open(w); err != nil {
		t.Fatal(err)
	}
	ref = windowRef(t, a, "main")
	lifetime := ref.Context()
	a.Do(func() {
		if !errors.Is(a.RequestCloseWindow(ref), ErrCloseRefused) {
			t.Fatal("refusal ignored")
		}
	})
	if lifetime.Err() != nil || len(a.Windows()) != 1 || calls != 1 {
		t.Fatal("refusal changed lifetime")
	}
	accept = true
	a.Do(func() {
		if err := a.CheckCloseWindow(ref); err != nil {
			t.Fatal(err)
		}
	})
	if lifetime.Err() != nil || len(a.Windows()) != 1 || calls != 2 {
		t.Fatal("permission check committed teardown")
	}
	accept = false
	if err := a.CloseWindow(ref); err != nil {
		t.Fatal(err)
	}
	if calls != 2 || lifetime.Err() == nil {
		t.Fatal("forced close called policy")
	}
}
func TestAcceptedCloseAndPolicyCancellation(t *testing.T) {
	for _, force := range []bool{false, true} {
		t.Run(fmt.Sprint(force), func(t *testing.T) {
			a, _ := New(Config{})
			defer a.Close()
			a.Open(Window{ID: "main", Title: "Document", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }, CloseRequest: func() bool {
				if force {
					a.Close()
				}
				return !force
			}})
			ref := windowRef(t, a, "main")
			a.Do(func() {
				err := a.RequestCloseWindow(ref)
				if force && !errors.Is(err, ErrClosed) {
					t.Fatal(err)
				}
				if !force && err != nil {
					t.Fatal(err)
				}
			})
			if a.Context().Err() == nil {
				t.Fatal("accepted close left app alive")
			}
		})
	}
}
