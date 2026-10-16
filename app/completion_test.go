package app

import (
	"context"
	"errors"
	"math"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dshills/vellum/ui"
)

func completionApp(t *testing.T, ordinary, reserved int) *Application {
	t.Helper()
	a, err := New(Config{QueueCapacity: ordinary, CompletionCapacity: reserved})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(a.Close)
	return a
}
func completionWindow(t *testing.T, a *Application, id string) {
	t.Helper()
	if err := a.Open(Window{ID: id, Title: id, Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
}
func ticket(t *testing.T, a *Application, window string, cleanup func(error)) *Completion {
	t.Helper()
	var c *Completion
	var err error
	if window == "" {
		c, err = a.ReserveCompletion(cleanup)
	} else {
		c, err = a.ReserveWindowCompletion(windowRef(t, a, window), cleanup)
	}
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestReservedResultsMergeByPublicationArrivalWithoutPlaceholder(t *testing.T) {
	for _, batch := range []int{1, 7, 64} {
		t.Run(string(rune('a'+batch)), func(t *testing.T) {
			a := completionApp(t, 2, 2)
			var order []string
			var cleaned atomic.Int32
			cleanup := func(error) { _ = a.Pending(); cleaned.Add(1) } // Would deadlock under app.mu.
			A, B := ticket(t, a, "", cleanup), ticket(t, a, "", cleanup)
			if a.Pending().Idle() || a.Pending().Reserved != 2 {
				t.Fatal("unpublished work reported idle")
			}
			if err := a.Post(func() { order = append(order, "P") }); err != nil {
				t.Fatal(err)
			}
			if err := B.Publish(func() { order = append(order, "B") }); err != nil {
				t.Fatal(err)
			}
			if err := a.Post(func() { order = append(order, "Q") }); err != nil {
				t.Fatal(err)
			}
			if err := A.Publish(func() { order = append(order, "A") }); err != nil {
				t.Fatal(err)
			}
			if err := a.Post(func() { t.Error("refused ordinary callback ran") }); !errors.Is(err, ErrQueueFull) {
				t.Fatalf("full ordinary queue %v", err)
			}
			for !a.Pending().Idle() {
				n, err := a.DrainN(batch)
				if err != nil || n == 0 {
					t.Fatalf("drain %d %v %+v", n, err, a.Pending())
				}
			}
			if !reflect.DeepEqual(order, []string{"P", "B", "Q", "A"}) || cleaned.Load() != 2 {
				t.Fatalf("order %v cleanup %d", order, cleaned.Load())
			}
		})
	}
}

func TestArrivalCutoffSurvivesAbortedInitialEnvelope(t *testing.T) {
	a := completionApp(t, 2, 1)
	B := ticket(t, a, "", nil)
	var order []string
	if err := a.Post(func() {
		order = append(order, "A")
		if !B.Abort() {
			t.Error("B abort refused")
		}
		if err := a.Post(func() { order = append(order, "C") }); err != nil {
			t.Error(err)
		}
	}); err != nil {
		t.Fatal(err)
	}
	if err := B.Publish(func() { t.Error("aborted B ran") }); err != nil {
		t.Fatal(err)
	}
	if n, err := a.DrainN(64); err != nil || n != 1 || !reflect.DeepEqual(order, []string{"A"}) {
		t.Fatalf("new C entered initial batch: n%d err%v order%v", n, err, order)
	}
	if a.Pending().Ordinary != 1 {
		t.Fatal("C not retained")
	}
	if n, err := a.DrainN(1); err != nil || n != 1 || !reflect.DeepEqual(order, []string{"A", "C"}) {
		t.Fatalf("next batch n%d err%v order%v", n, err, order)
	}
}

func TestWindowAndApplicationCloseRevokeUnpublishedAndReadyTicketsSynchronously(t *testing.T) {
	a := completionApp(t, 2, 4)
	completionWindow(t, a, "one")
	completionWindow(t, a, "two")
	lifetime, _ := a.WindowContext("one")
	var canceled atomic.Int32
	cleanup := func(err error) {
		if !errors.Is(err, ErrWindowClosed) {
			t.Errorf("window reason %v", err)
		}
		canceled.Add(1)
		_ = a.Pending()
	}
	reserved := ticket(t, a, "one", cleanup)
	ready := ticket(t, a, "one", cleanup)
	if err := ready.Publish(func() { t.Error("closed window completion ran") }); err != nil {
		t.Fatal(err)
	}
	survivor := ticket(t, a, "two", nil)
	var delivered bool
	if err := survivor.Publish(func() { delivered = true }); err != nil {
		t.Fatal(err)
	}
	if err := a.CloseWindow(windowRef(t, a, "one")); err != nil {
		t.Fatal(err)
	}
	if lifetime.Err() != context.Canceled || canceled.Load() != 2 {
		t.Fatalf("close not synchronous ctx%v cleanup%d", lifetime.Err(), canceled.Load())
	}
	for _, c := range []*Completion{reserved, ready} {
		select {
		case <-c.Done():
		default:
			t.Fatal("terminal cleanup not observed")
		}
		if err := c.Publish(func() {}); !errors.Is(err, ErrWindowClosed) {
			t.Fatalf("late publish %v", err)
		}
	}
	if err := a.Drain(); err != nil || !delivered {
		t.Fatalf("survivor delivery %v %v", delivered, err)
	}
	whole := ticket(t, a, "", nil)
	a.Close()
	select {
	case <-whole.Done():
	default:
		t.Fatal("app ticket survived close")
	}
	if !errors.Is(whole.Err(), ErrClosed) || !a.Pending().Idle() {
		t.Fatalf("shutdown accounting %+v reason%v", a.Pending(), whole.Err())
	}
}

func TestReservationValidationCreditAndSequenceExhaustion(t *testing.T) {
	a := completionApp(t, 1, 1)
	a.nextArrival = math.MaxUint64 - 1
	c := ticket(t, a, "", nil)
	if _, err := a.ReserveCompletion(nil); !errors.Is(err, ErrCompletionCapacity) {
		t.Fatalf("full reserve %v", err)
	}
	if err := a.Post(func() {}); !errors.Is(err, ErrSequenceExhausted) {
		t.Fatalf("normal post spent reserved sequence credit: %v", err)
	}
	if err := c.Publish(nil); err == nil || a.Pending().Reserved != 1 {
		t.Fatalf("nil publication altered reservation: %v", err)
	}
	if err := c.Publish(func() {}); err != nil {
		t.Fatal(err)
	}
	if err := c.Publish(func() { t.Error("duplicate ran") }); !errors.Is(err, ErrCompletionUsed) {
		t.Fatalf("duplicate %v", err)
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if _, err := a.ReserveCompletion(nil); !errors.Is(err, ErrSequenceExhausted) {
		t.Fatalf("identity recycled %v", err)
	}
	if n, err := a.DrainN(-1); err == nil || n != 0 {
		t.Fatalf("negative budget %d %v", n, err)
	}
}

func TestCopiedCompletionCannotPublishAbortOrRefundOriginalCredit(t *testing.T) {
	a := completionApp(t, 1, 1)
	var cleaned atomic.Int32
	original := ticket(t, a, "", func(error) { cleaned.Add(1) })
	copy := *original
	if err := copy.Publish(func() { t.Error("copied callback ran") }); !errors.Is(err, ErrInvalidCompletion) {
		t.Fatalf("copied publication %v", err)
	}
	if copy.Abort() || !errors.Is(copy.Err(), ErrInvalidCompletion) {
		t.Fatal("copied ticket accepted terminal operation")
	}
	if a.Pending().Reserved != 1 || cleaned.Load() != 0 {
		t.Fatalf("copy stole original credit %+v cleanup%d", a.Pending(), cleaned.Load())
	}
	if _, err := a.ReserveCompletion(nil); !errors.Is(err, ErrCompletionCapacity) {
		t.Fatalf("copy refunded credit %v", err)
	}
	if !original.Abort() {
		t.Fatal("original ticket was invalidated")
	}
	if cleaned.Load() != 1 || !a.Pending().Idle() {
		t.Fatalf("original cleanup %+v count%d", a.Pending(), cleaned.Load())
	}
}

func TestFailedCleanupCannotStrandOtherRevokedReservations(t *testing.T) {
	for _, failure := range []string{"panic", "goexit"} {
		t.Run(failure, func(t *testing.T) {
			a := completionApp(t, 1, 3)
			var cleanups atomic.Int32
			var wakes atomic.Int32
			a.SetWake(func() { _ = a.Pending(); wakes.Add(1) })
			tickets := make([]*Completion, 3)
			for i := range tickets {
				tickets[i] = ticket(t, a, "", func(error) {
					_ = a.Pending() // Accounting callbacks are still outside app.mu.
					cleanups.Add(1)
					if failure == "panic" {
						panic("cleanup failed")
					}
					runtime.Goexit()
				})
			}
			terminated := make(chan any, 1)
			go func() {
				defer func() { terminated <- recover() }()
				a.Close()
			}()
			value := <-terminated
			if failure == "panic" && value != "cleanup failed" {
				t.Fatalf("original panic did not propagate: %v", value)
			}
			if cleanups.Load() != 3 || !a.Pending().Idle() || wakes.Load() != 1 {
				t.Fatalf("cleanup/wake stranded: calls%d pending%+v wakes%d", cleanups.Load(), a.Pending(), wakes.Load())
			}
			for _, c := range tickets {
				select {
				case <-c.Done():
				default:
					t.Fatal("revoked completion Done stranded")
				}
				if !errors.Is(c.Err(), ErrClosed) {
					t.Fatalf("lost revocation reason %v", c.Err())
				}
			}
		})
	}
}

func TestDirectOwnerGoexitClosesAfterUnlockAndRetiresResults(t *testing.T) {
	a := completionApp(t, 1, 1)
	c := ticket(t, a, "", nil)
	terminated := make(chan struct{})
	go func() {
		defer close(terminated)
		if err := a.Do(func() { runtime.Goexit() }); err != nil {
			t.Errorf("unexpected Do refusal %v", err)
		}
		t.Error("Goexit returned")
	}()
	<-terminated
	if a.Context().Err() != context.Canceled || !a.Pending().Idle() {
		t.Fatalf("nonreturn left owner live: %+v context%v", a.Pending(), a.Context().Err())
	}
	select {
	case <-c.Done():
	default:
		t.Fatal("Goexit stranded reserved result")
	}
	if err := a.Do(func() { t.Error("callback after Goexit") }); !errors.Is(err, ErrClosed) {
		t.Fatalf("owner locked/live after Goexit: %v", err)
	}
}

func TestCleanupFailuresFollowDeferredUnwindingAndRetireDistinctHooks(t *testing.T) {
	for _, second := range []string{"panic", "goexit"} {
		t.Run(second, func(t *testing.T) {
			a := completionApp(t, 1, 3)
			var cleaned atomic.Int32
			callbacks := []func(error){
				func(error) { cleaned.Add(1); panic("first") },
				func(error) {
					cleaned.Add(1)
					if second == "panic" {
						panic("second")
					}
					runtime.Goexit()
				},
				func(error) { cleaned.Add(1) },
			}
			var hooks []completionHook
			var tickets []*Completion
			for _, callback := range callbacks {
				c := ticket(t, a, "", callback)
				tickets = append(tickets, c)
				a.mu.Lock()
				hooks = append(hooks, a.terminalLocked(c, ErrCompletionAborted))
				a.mu.Unlock()
			}
			terminal := make(chan any, 1)
			go func() {
				defer func() { terminal <- recover() }()
				runCompletionHooks(hooks)
			}()
			value := <-terminal
			if second == "panic" && value != "second" || second == "goexit" && value != nil {
				t.Fatalf("wrong Go unwinding result: %v", value)
			}
			if cleaned.Load() != 3 || !a.Pending().Idle() {
				t.Fatalf("distinct hooks stranded: calls%d state%+v", cleaned.Load(), a.Pending())
			}
			for _, c := range tickets {
				select {
				case <-c.Done():
				default:
					t.Fatal("terminal hook not finalized")
				}
			}
		})
	}
}

func TestCleanupPhaseIsObservableAfterReservationCreditRetires(t *testing.T) {
	a := completionApp(t, 1, 1)
	entered, release := make(chan struct{}), make(chan struct{})
	c := ticket(t, a, "", func(error) { close(entered); <-release })
	finished := make(chan struct{})
	go func() { defer close(finished); c.Abort() }()
	<-entered
	if p := a.Pending(); p.Cleanup != 1 || p.Reserved != 0 || p.Idle() {
		t.Fatalf("cleanup incorrectly idle %+v", p)
	}
	select {
	case <-c.Done():
		t.Fatal("Done before cleanup return")
	default:
	}
	// This adversarial blocking hook breaches the documented contract, but
	// exposes the precise reservation-vs-cleanup accounting boundary.
	next := ticket(t, a, "", nil)
	next.Abort()
	close(release)
	<-finished
	if !a.Pending().Idle() {
		t.Fatalf("cleanup remained pending %+v", a.Pending())
	}
}

func TestConcurrentReadyAdmissionsAreBoundedAndExactlyOnce(t *testing.T) {
	const count = 32
	a := completionApp(t, count, count)
	tickets := make([]*Completion, count)
	for i := range tickets {
		tickets[i] = ticket(t, a, "", nil)
	}
	counts := make([]atomic.Int32, 2*count)
	var wg sync.WaitGroup
	for i := range count {
		wg.Go(func() {
			if err := a.Post(func() { counts[i].Add(1) }); err != nil {
				t.Error(err)
			}
		})
		wg.Go(func() {
			if err := tickets[i].Publish(func() { counts[count+i].Add(1) }); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	a.mu.Lock()
	last := uint64(0)
	for _, e := range a.queue {
		if e.arrival <= last {
			t.Error("arrival order not monotonic")
		}
		last = e.arrival
	}
	a.mu.Unlock()
	for !a.Pending().Idle() {
		if n, err := a.DrainN(7); err != nil || n == 0 {
			t.Fatalf("drain n%d %v", n, err)
		}
	}
	for i := range counts {
		if counts[i].Load() != 1 {
			t.Errorf("entry %d ran%d", i, counts[i].Load())
		}
	}
}

func TestOwnerCallbackPanicRetiresTicketAndRevokesOtherResults(t *testing.T) {
	a := completionApp(t, 1, 2)
	first, second := ticket(t, a, "", nil), ticket(t, a, "", nil)
	if err := first.Publish(func() { panic("owner") }); err != nil {
		t.Fatal(err)
	}
	if err := second.Publish(func() { t.Error("callback after owner panic") }); err != nil {
		t.Fatal(err)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("owner panic did not propagate")
			}
		}()
		_ = a.Drain()
	}()
	if !errors.Is(first.Err(), ErrOwnerCallbackFailed) || !errors.Is(second.Err(), ErrClosed) || !a.Pending().Idle() {
		t.Fatalf("panic cleanup first%v second%v %+v", first.Err(), second.Err(), a.Pending())
	}
}
