package cogent

import "sync"

// nativeQuitRequest retains terminal intent independently of scene Show events.
// A native window is safe to quit only after the whole construction batch ends.
// topology is the existing Driver.Run topology mutex, never an application owner.
type nativeQuitRequest struct {
	mu                                sync.Mutex
	pending, ready, batch, dispatched bool
	dispatch                          func()
}

func (q *nativeQuitRequest) takeLocked() func() {
	if !q.pending || !q.ready || q.batch || q.dispatched {
		return nil
	}
	q.dispatched = true
	return q.dispatch
}

func (q *nativeQuitRequest) request() {
	q.mu.Lock()
	q.pending = true
	run := q.takeLocked()
	q.mu.Unlock()
	if run != nil {
		run()
	}
}

// markReady preserves Show as a fallback for asynchronous/non-Mac creation.
// It cannot authorize quit while any window in the current batch is pending.
func (q *nativeQuitRequest) markReady() {
	q.mu.Lock()
	q.ready = true
	run := q.takeLocked()
	q.mu.Unlock()
	if run != nil {
		run()
	}
}

func (q *nativeQuitRequest) beginBatch(topology *sync.Mutex) bool {
	topology.Lock()
	q.mu.Lock()
	if q.dispatched {
		q.mu.Unlock()
		topology.Unlock()
		return false
	}
	q.batch = true
	q.mu.Unlock()
	return true
}

func (q *nativeQuitRequest) endBatch(topology *sync.Mutex, nativeReady bool) {
	q.mu.Lock()
	q.ready = q.ready || nativeReady
	q.batch = false
	run := q.takeLocked()
	// Publish completion and unlock topology before another request can dispatch
	// or another construction batch can begin. Neither action loses a wake.
	topology.Unlock()
	q.mu.Unlock()
	if run != nil {
		run()
	}
}
