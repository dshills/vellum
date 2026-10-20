package cogent

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func quitFixture(t *testing.T) (*nativeQuitRequest, *sync.Mutex, *atomic.Int64) {
	t.Helper()
	var topology sync.Mutex
	var calls atomic.Int64
	q := &nativeQuitRequest{}
	q.dispatch = func() {
		calls.Add(1)
		q.request() // reentry must neither deadlock nor dispatch twice
	}
	return q, &topology, &calls
}

func TestNativeQuitRetainsPreReadyFailureWithoutOnShow(t *testing.T) {
	q, topology, calls := quitFixture(t)
	if !q.beginBatch(topology) {
		t.Fatal("initial construction refused")
	}
	q.request() // terminal error before any scene Show
	if calls.Load() != 0 {
		t.Fatal("quit entered partial native construction")
	}
	q.endBatch(topology, true) // valid synchronous native stage, no Show needed
	if calls.Load() != 1 {
		t.Fatal("pre-show failure quit intent was lost")
	}
}

func TestNativeQuitRetainsPreReadyCancellationWithoutOnShow(t *testing.T) {
	q, topology, calls := quitFixture(t)
	if !q.beginBatch(topology) {
		t.Fatal("initial construction refused")
	}
	ctx, cancel := context.WithCancel(context.Background())
	requested := make(chan struct{})
	stop := context.AfterFunc(ctx, func() { q.request(); close(requested) })
	defer stop()
	cancel()
	select {
	case <-requested:
	case <-time.After(time.Second):
		t.Fatal("cancellation callback did not finish")
	}
	if calls.Load() != 0 {
		t.Fatal("cancel quit entered partial construction")
	}
	q.endBatch(topology, true)
	if calls.Load() != 1 {
		t.Fatal("pre-show cancellation quit intent was lost")
	}
}

func TestNativeQuitReadyBeforeRequest(t *testing.T) {
	q, topology, calls := quitFixture(t)
	if !q.beginBatch(topology) {
		t.Fatal("construction refused")
	}
	q.endBatch(topology, true)
	if calls.Load() != 0 {
		t.Fatal("readiness alone requested quit")
	}
	q.request()
	if calls.Load() != 1 {
		t.Fatal("ready quit request did not dispatch")
	}
}

func TestNativeQuitWaitsForCompleteTwoWindowBatch(t *testing.T) {
	q, topology, calls := quitFixture(t)
	if !q.beginBatch(topology) {
		t.Fatal("construction refused")
	}
	q.markReady() // first window's Show is not a complete two-window batch
	q.request()
	if calls.Load() != 0 {
		t.Fatal("first window readiness allowed premature quit")
	}
	q.markReady() // second window can Show while its Stage.Run is still unwinding
	if calls.Load() != 0 {
		t.Fatal("Show bypassed batch completion")
	}
	q.endBatch(topology, true)
	if calls.Load() != 1 {
		t.Fatal("complete batch failed to forward quit")
	}
}

func TestNativeQuitLaterConstructionBlocksReadyQuit(t *testing.T) {
	q, topology, calls := quitFixture(t)
	if !q.beginBatch(topology) {
		t.Fatal("first construction refused")
	}
	q.endBatch(topology, true)
	if !q.beginBatch(topology) {
		t.Fatal("later construction refused before quit")
	}
	q.request()
	if calls.Load() != 0 {
		t.Fatal("old readiness bypassed later construction")
	}
	q.endBatch(topology, false) // earlier native readiness remains valid
	if calls.Load() != 1 {
		t.Fatal("later batch lost retained request")
	}
}

func TestNativeQuitNoWindowNeverEntersCoreQuit(t *testing.T) {
	q, topology, calls := quitFixture(t)
	if !q.beginBatch(topology) {
		t.Fatal("construction refused")
	}
	q.request()
	q.endBatch(topology, false) // Driver.Run preserves its early direct error return
	q.request()
	if calls.Load() != 0 {
		t.Fatal("quit dispatched without a constructed native window")
	}
}

func TestNativeQuitShowFallbackAfterAsynchronousBatch(t *testing.T) {
	q, topology, calls := quitFixture(t)
	if !q.beginBatch(topology) {
		t.Fatal("construction refused")
	}
	q.endBatch(topology, false)
	q.request()
	if calls.Load() != 0 {
		t.Fatal("asynchronous window became ready prematurely")
	}
	q.markReady()
	if calls.Load() != 1 {
		t.Fatal("Show fallback failed to forward retained intent")
	}
}

func TestNativeQuitConcurrentRequestsDispatchExactlyOnce(t *testing.T) {
	q, topology, calls := quitFixture(t)
	if !q.beginBatch(topology) {
		t.Fatal("construction refused")
	}
	var workers sync.WaitGroup
	for range 64 {
		workers.Add(1)
		go func() { defer workers.Done(); q.request(); q.markReady() }()
	}
	workers.Wait()
	if calls.Load() != 0 {
		t.Fatal("racing request escaped construction batch")
	}
	q.endBatch(topology, true)
	workers = sync.WaitGroup{}
	for range 64 {
		workers.Add(1)
		go func() { defer workers.Done(); q.request(); q.markReady() }()
	}
	workers.Wait()
	if calls.Load() != 1 {
		t.Fatalf("dispatches=%d, want1", calls.Load())
	}
	if q.beginBatch(topology) {
		q.endBatch(topology, true)
		t.Fatal("new construction admitted after quit dispatch")
	}
}

func TestNativeQuitBatchHandoffNeverDropsNextConstruction(t *testing.T) {
	q, topology, calls := quitFixture(t)
	if !q.beginBatch(topology) {
		t.Fatal("first construction refused")
	}
	second := make(chan bool, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		ok := q.beginBatch(topology)
		second <- ok
		if ok {
			q.endBatch(topology, true)
		}
	}()
	q.endBatch(topology, true)
	select {
	case ok := <-second:
		if !ok {
			t.Fatal("next construction was dropped during handoff")
		}
	case <-time.After(time.Second):
		t.Fatal("next construction did not acquire topology")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("second batch did not finish")
	}
	q.request()
	if calls.Load() != 1 {
		t.Fatal("handoff left quit blocked or duplicated")
	}
}

func TestNativeQuitDispatchRunsOutsideBothOwners(t *testing.T) {
	q, topology, calls := quitFixture(t)
	q.dispatch = func() {
		if !q.mu.TryLock() {
			t.Fatal("dispatch under request owner")
		}
		q.mu.Unlock()
		if !topology.TryLock() {
			t.Fatal("dispatch under topology owner")
		}
		topology.Unlock()
		calls.Add(1)
	}
	if !q.beginBatch(topology) {
		t.Fatal("construction refused")
	}
	q.request()
	q.endBatch(topology, true)
	if calls.Load() != 1 {
		t.Fatal("retained quit not dispatched")
	}
}
