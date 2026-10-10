// Package mailbox prototypes bounded admission of ordered application messages.
// It is not yet wired to Cocoa callbacks or the Cogent event pump.
package mailbox

import (
	"context"
	"errors"
	"math"
	"sync"
)

var (
	ErrFull   = errors.New("mailbox capacity exhausted")
	ErrSize   = errors.New("payload exceeds configured size")
	ErrClosed = errors.New("mailbox closed for admission")
	ErrStale  = errors.New("stale reservation or acknowledgement")
	ErrID     = errors.New("receipt IDs exhausted")
)

type state uint8

const (
	empty state = iota
	reserved
	committed
	inFlight
	canceled
)

type slot struct {
	id    uint64
	data  []byte
	state state
}

// Limits bounds queue-owned payload bytes, including reservations and the single
// in-flight delivery. Fixed slot metadata/allocator overhead are additional.
type Limits struct{ Slots, PayloadBytes, MessageBytes int }
type Stats struct {
	Outstanding, PayloadBytes, InFlight int
	Closed                              bool
}
type Queue struct {
	mu                       sync.Mutex
	slots                    []slot
	head, tail, count, bytes int
	limits                   Limits
	next                     uint64
	closed                   bool
	changed                  chan struct{}
}

// Reservation owns a queue credit until committed or canceled. Callers must
// reserve before external mutation. Abandoned tickets can stall the FIFO.
type Reservation struct {
	q     *Queue
	index int
	id    uint64
	size  int
}

// Delivery.Data is borrowed until Ack; the consumer must not retain it after Ack.
// Only one delivery may be in flight; credit is not released merely by dequeue.
type Delivery struct {
	ID   uint64
	Data []byte
}

func New(l Limits) (*Queue, error) {
	if l.Slots <= 0 || l.PayloadBytes <= 0 || l.MessageBytes <= 0 || l.MessageBytes > l.PayloadBytes {
		return nil, ErrSize
	}
	return &Queue{slots: make([]slot, l.Slots), limits: l, changed: make(chan struct{})}, nil
}
func (q *Queue) signal() { close(q.changed); q.changed = make(chan struct{}) }
func (q *Queue) skipCanceled() {
	for q.count > 0 && q.slots[q.head].state == canceled {
		q.slots[q.head] = slot{}
		q.head = (q.head + 1) % len(q.slots)
		q.count--
	}
}
func (q *Queue) reserve(n int) (*Reservation, error) {
	if q.closed {
		return nil, ErrClosed
	}
	if n < 0 || n > q.limits.MessageBytes {
		return nil, ErrSize
	}
	q.skipCanceled()
	if q.count == len(q.slots) || n > q.limits.PayloadBytes-q.bytes {
		return nil, ErrFull
	}
	if q.next == math.MaxUint64 {
		return nil, ErrID
	}
	q.next++
	index := q.tail
	q.slots[index] = slot{q.next, make([]byte, n), reserved}
	q.tail = (q.tail + 1) % len(q.slots)
	q.count++
	q.bytes += n
	return &Reservation{q, index, q.next, n}, nil
}

// TryReserve never waits for capacity. A short mutex protects queue metadata.
func (q *Queue) TryReserve(n int) (*Reservation, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.reserve(n)
}
func (r *Reservation) Commit(data []byte) error {
	q := r.q
	q.mu.Lock()
	defer q.mu.Unlock()
	s := &q.slots[r.index]
	if s.id != r.id || s.state != reserved {
		return ErrStale
	}
	if len(data) != r.size {
		return ErrSize
	}
	copy(s.data, data)
	s.state = committed
	q.signal()
	return nil
}
func (r *Reservation) Cancel() error {
	q := r.q
	q.mu.Lock()
	defer q.mu.Unlock()
	s := &q.slots[r.index]
	if s.id != r.id || s.state != reserved {
		return ErrStale
	}
	q.bytes -= len(s.data)
	s.data = nil
	s.state = canceled
	q.skipCanceled()
	q.signal()
	return nil
}

// Submit waits for credit on worker goroutines only. Never call this from a
// native callback or UI event pump. Cancellation before admission returns an error;
// cancellation after admission does not revoke an accepted message.
func (q *Queue) Submit(ctx context.Context, data []byte) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		q.mu.Lock()
		r, err := q.reserve(len(data))
		changed := q.changed
		q.mu.Unlock()
		if err == nil {
			return r.Commit(data)
		}
		if !errors.Is(err, ErrFull) {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-changed:
		}
	}
}
func (q *Queue) nextDelivery() (Delivery, bool) {
	q.skipCanceled()
	if q.count == 0 || q.slots[q.head].state != committed {
		return Delivery{}, false
	}
	s := &q.slots[q.head]
	s.state = inFlight
	return Delivery{s.id, s.data}, true
}
func (q *Queue) TryNext() (Delivery, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.nextDelivery()
}

// Next waits on a dedicated consumer goroutine, not on the native main thread.
func (q *Queue) Next(ctx context.Context) (Delivery, error) {
	for {
		if err := ctx.Err(); err != nil {
			return Delivery{}, err
		}
		q.mu.Lock()
		d, ok := q.nextDelivery()
		changed := q.changed
		done := q.closed && q.count == 0
		q.mu.Unlock()
		if ok {
			return d, nil
		}
		if done {
			return Delivery{}, ErrClosed
		}
		select {
		case <-ctx.Done():
			return Delivery{}, ctx.Err()
		case <-changed:
		}
	}
}

// Ack releases the head credit after application processing, not before it.
func (q *Queue) Ack(id uint64) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.count == 0 {
		return ErrStale
	}
	s := &q.slots[q.head]
	if s.id != id || s.state != inFlight {
		return ErrStale
	}
	q.bytes -= len(s.data)
	*s = slot{}
	q.head = (q.head + 1) % len(q.slots)
	q.count--
	q.skipCanceled()
	q.signal()
	return nil
}

// Close stops admission and cancels uncommitted reservations. Already accepted
// deliveries remain readable and must be drained/acknowledged by the owner.
// Native mutation/commit and shutdown must be serialized by their view owner.
func (q *Queue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return
	}
	q.closed = true
	for i := range q.slots {
		s := &q.slots[i]
		if s.state == reserved {
			q.bytes -= len(s.data)
			s.data = nil
			s.state = canceled
		}
	}
	q.skipCanceled()
	q.signal()
}
func (q *Queue) Stats() Stats {
	q.mu.Lock()
	defer q.mu.Unlock()
	flight := 0
	if q.count > 0 && q.slots[q.head].state == inFlight {
		flight = 1
	}
	return Stats{q.count, q.bytes, flight, q.closed}
}
