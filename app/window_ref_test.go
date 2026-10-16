package app

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"testing"
	"weak"
)

func TestWindowRefReopeningFencesAllOriginatingWork(t *testing.T) {
	a, _ := New(Config{})
	defer a.Close()
	_ = a.Open(sampleWindow("survivor"))
	old, err := a.OpenWindow(sampleWindow("document"))
	if err != nil {
		t.Fatal(err)
	}
	copy := old
	if old != copy || old != a.Windows()[1].Ref() || (Window{}).Ref() != (WindowRef{}) {
		t.Fatal("snapshot/copy lifetime changed")
	}
	if err = a.PostWindow(copy, func() { t.Error("old queued work ran in reopened window") }); err != nil {
		t.Fatal(err)
	}
	cleaned := 0
	ready, err := a.ReserveWindowCompletion(old, func(err error) {
		if !errors.Is(err, ErrWindowClosed) {
			t.Error(err)
		}
		cleaned++
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = ready.Publish(func() { t.Error("old completion ran in reopened window") }); err != nil {
		t.Fatal(err)
	}
	if err = a.CloseWindow(old); err != nil {
		t.Fatal(err)
	}
	current, err := a.OpenWindow(sampleWindow("document"))
	if err != nil || current == old || current.ID() != old.ID() || old.Context().Err() != context.Canceled || current.Context().Err() != nil {
		t.Fatal("reopened lifetime alias", err)
	}
	for _, action := range []func() error{
		func() error { return a.PostWindow(old, func() { t.Error("stale post ran") }) },
		func() error { return a.CloseWindow(copy) },
		func() error { return a.CheckCloseWindow(old) },
		func() error { return a.RequestCloseWindow(old) },
		func() error { _, err := a.ReserveWindowCompletion(old, nil); return err },
	} {
		if err := action(); !errors.Is(err, ErrWindowClosed) {
			t.Fatal("stale action admitted", err)
		}
	}
	if a.TakeDirty(old) || !a.TakeDirty(current) {
		t.Fatal("old driver consumed new dirty state")
	}
	delivered := 0
	if err = a.PostWindow(current, func() { delivered++ }); err != nil {
		t.Fatal(err)
	}
	if err = a.Drain(); err != nil || delivered != 1 || cleaned != 1 || current.Context().Err() != nil {
		t.Fatal("current work or revocation lost", delivered, cleaned, err)
	}
}

func TestWindowRefRejectsForeignAndZeroWithoutApplicationFallback(t *testing.T) {
	a, _ := New(Config{})
	defer a.Close()
	b, _ := New(Config{})
	defer b.Close()
	current, _ := a.OpenWindow(sampleWindow("same"))
	foreign, _ := b.OpenWindow(sampleWindow("same"))
	if (WindowRef{}).Context().Err() != context.Canceled || (WindowRef{}).ID() != "" {
		t.Fatal("zero reference has an active context")
	}
	for _, ref := range []WindowRef{{}, foreign} {
		for _, action := range []func() error{
			func() error { return a.PostWindow(ref, func() { t.Error("invalid delivery") }) },
			func() error { return a.CloseWindow(ref) },
			func() error { return a.CheckCloseWindow(ref) },
			func() error { return a.RequestCloseWindow(ref) },
			func() error { _, err := a.ReserveWindowCompletion(ref, nil); return err },
		} {
			if err := action(); !errors.Is(err, ErrInvalidWindowRef) {
				t.Fatal("invalid reference admitted", err)
			}
		}
		if a.TakeDirty(ref) {
			t.Fatal("invalid dirty consumption")
		}
	}
	if !a.TakeDirty(current) || current.Context().Err() != nil || foreign.Context().Err() != nil || a.Pending().Ordinary != 0 || a.Pending().Reserved != 0 {
		t.Fatal("invalid attempts changed owner state")
	}
	ticket, err := a.ReserveCompletion(nil)
	if err != nil || ticket == nil {
		t.Fatal("explicit application reservation refused", err)
	}
	ticket.Abort()
}

func TestOldClosePolicyCannotApproveOrBlockReopenedName(t *testing.T) {
	a, _ := New(Config{})
	defer a.Close()
	_ = a.Open(sampleWindow("survivor"))
	var old, current WindowRef
	oldSpec := sampleWindow("document")
	newCalls := 0
	oldSpec.CloseRequest = func() bool {
		if !errors.Is(a.CheckCloseWindow(old), ErrClosePending) {
			t.Fatal("recursive policy admitted")
		}
		if err := a.CloseWindow(old); err != nil {
			t.Fatal(err)
		}
		next := sampleWindow("document")
		next.CloseRequest = func() bool { newCalls++; return false }
		var err error
		current, err = a.OpenWindow(next)
		if err != nil {
			t.Fatal(err)
		}
		if !errors.Is(a.CheckCloseWindow(current), ErrCloseRefused) {
			t.Fatal("old policy blocked new policy")
		}
		return true
	}
	var err error
	old, err = a.OpenWindow(oldSpec)
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Do(func() {
		if !errors.Is(a.RequestCloseWindow(old), ErrWindowClosed) {
			t.Fatal("old permission closed replacement")
		}
	}); err != nil {
		t.Fatal(err)
	}
	if current.Context().Err() != nil || newCalls != 1 || len(a.closeRequests) != 0 || len(a.Windows()) != 2 {
		t.Fatal("reentrant policy lost replacement")
	}
}

func TestConcurrentWindowRefRetirementAndNameReuse(t *testing.T) {
	a, _ := New(Config{})
	defer a.Close()
	_ = a.Open(sampleWindow("survivor"))
	for i := 0; i < 100; i++ {
		old, err := a.OpenWindow(sampleWindow("document"))
		if err != nil {
			t.Fatal(err)
		}
		var wg sync.WaitGroup
		wg.Add(2)
		var closed, posted error
		go func() { defer wg.Done(); closed = a.CloseWindow(old) }()
		go func() {
			defer wg.Done()
			posted = a.PostWindow(old, func() { t.Error("retired queued work executed") })
		}()
		wg.Wait()
		if closed != nil || posted != nil && !errors.Is(posted, ErrWindowClosed) {
			t.Fatal(closed, posted)
		}
		current, err := a.OpenWindow(sampleWindow("document"))
		if err != nil {
			t.Fatal(err)
		}
		if err = a.Drain(); err != nil {
			t.Fatal(err)
		}
		if err = a.CloseWindow(old); !errors.Is(err, ErrWindowClosed) {
			t.Fatal("late close aliased current", err)
		}
		if current.Context().Err() != nil {
			t.Fatal("new context canceled")
		}
		if err = a.CloseWindow(current); err != nil {
			t.Fatal(err)
		}
	}
}

//go:noinline
func referenceWithoutApplication(t *testing.T) (WindowRef, weak.Pointer[Application], weak.Pointer[closedWindowPayload]) {
	t.Helper()
	a, _ := New(Config{})
	owner := weak.Make(a)
	payload, err := openRetainedWindow(a)
	if err != nil {
		t.Fatal(err)
	}
	ref := windowRef(t, a, "temporary")
	a.Close()
	if len(a.Windows()) != 0 || a.lifetimes != nil || a.windowCancels != nil || a.closeRequests != nil || a.dirty != nil {
		t.Fatal("application close retained window state")
	}
	return ref, owner, payload
}

func TestWindowRefDoesNotRetainApplicationOrClosedView(t *testing.T) {
	ref, owner, payload := referenceWithoutApplication(t)
	runtime.GC()
	if owner.Value() != nil || payload.Value() != nil {
		t.Fatal("reference retained the application or a closed view")
	}
	if ref.Context().Err() != context.Canceled || ref.ID() != "temporary" {
		t.Fatal("reference lost captured identity")
	}
	runtime.KeepAlive(ref)
}

func TestOpenWindowReturnsOriginalRefAfterWakeReplacement(t *testing.T) {
	a, _ := New(Config{})
	defer a.Close()
	_ = a.Open(sampleWindow("survivor"))
	armed := false
	var first, replacement WindowRef
	a.SetWake(func() {
		if !armed {
			return
		}
		armed = false
		var err error
		first, err = a.WindowRef("document")
		if err != nil {
			t.Fatal(err)
		}
		if err = a.CloseWindow(first); err != nil {
			t.Fatal(err)
		}
		replacement, err = a.OpenWindow(sampleWindow("document"))
		if err != nil {
			t.Fatal(err)
		}
	})
	armed = true
	returned, err := a.OpenWindow(sampleWindow("document"))
	if err != nil || returned != first || returned == replacement || returned.Context().Err() != context.Canceled || replacement.Context().Err() != nil {
		t.Fatal("open returned a later current-name lookup", err)
	}
	a.SetWake(nil)
}
