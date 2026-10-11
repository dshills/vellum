// Package transport connects pre-mutation admission to a bounded UI drain.
package transport

import (
	"encoding/json"
	"errors"
	"github.com/dshills/vellum/drivers/cogent/internal/mailbox"
	"sync"
)

// Pump holds ordered actions separately from one replaceable latest snapshot.
// Wake must enqueue only a wake token, never payloads. Drain runs on one UI owner.
// Wake cannot run inline: it must schedule Drain on that owner.
type Pump[T any] struct {
	mu             sync.Mutex
	q              *mailbox.Queue
	wake           func()
	armed, closed  bool
	latest         []byte
	max            int
	refused, wakes uint64
}
type Ticket[T any] struct {
	p    *Pump[T]
	r    *mailbox.Reservation
	data []byte
}
type Stats struct {
	Queue          mailbox.Stats
	Refused, Wakes uint64
	LatestBytes    int
	Closed         bool
}

func New[T any](limits mailbox.Limits, wake func()) (*Pump[T], error) {
	if wake == nil {
		return nil, errors.New("wake callback required")
	}
	q, err := mailbox.New(limits)
	if err != nil {
		return nil, err
	}
	return &Pump[T]{q: q, wake: wake, max: limits.MessageBytes}, nil
}
func (p *Pump[T]) encode(v T) ([]byte, error) {
	b, e := json.Marshal(v)
	if e == nil && len(b) > p.max {
		e = mailbox.ErrSize
	}
	return b, e
}

// Preflight owns both credits and the immutable encoded prospective state.
// The producer must Commit or Cancel; lifetime shutdown is serialized with native mutation.
func (p *Pump[T]) Preflight(v T) (*Ticket[T], error) {
	b, e := p.encode(v)
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, mailbox.ErrClosed
	}
	if e != nil {
		p.refused++
		return nil, e
	}
	r, e := p.q.TryReserve(len(b))
	if e != nil {
		p.refused++
		return nil, e
	}
	return &Ticket[T]{p, r, b}, nil
}
func (t *Ticket[T]) Commit() error {
	p := t.p
	p.mu.Lock()
	e := t.r.Commit(t.data)
	wake := false
	if e == nil {
		t.data = nil
		wake = p.arm()
	}
	p.mu.Unlock()
	if wake {
		p.wake()
	}
	return e
}
func (t *Ticket[T]) Cancel() error {
	p := t.p
	p.mu.Lock()
	e := t.r.Cancel()
	t.data = nil
	wake := p.arm()
	p.mu.Unlock()
	if wake {
		p.wake()
	}
	return e
}
func (p *Pump[T]) arm() bool {
	if p.armed {
		return false
	}
	p.armed = true
	p.wakes++
	return true
}
func (p *Pump[T]) Send(v T) error {
	t, e := p.Preflight(v)
	if e != nil {
		return e
	}
	return t.Commit()
}

// Latest replaces the sole snapshot (caret/progress). It never replaces actions.
func (p *Pump[T]) Latest(v T) error {
	b, e := p.encode(v)
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return mailbox.ErrClosed
	}
	if e != nil {
		p.refused++
		p.mu.Unlock()
		return e
	}
	p.latest = b
	wake := p.arm()
	p.mu.Unlock()
	if wake {
		p.wake()
	}
	return nil
}

// Drain processes at most batch ordered messages plus one latest snapshot.
// Credit remains held while apply runs. Concurrent commits rearm under the same lock.
// A blocked reservation is woken by Commit/Cancel, so it does not spin the UI.
func (p *Pump[T]) Drain(batch int, apply func(T)) error {
	defer p.rearm()
	if batch <= 0 {
		return errors.New("positive batch required")
	}
	for n := 0; n < batch; n++ {
		d, ok := p.q.TryNext()
		if !ok {
			break
		}
		var v T
		if e := json.Unmarshal(d.Data, &v); e != nil {
			_ = p.q.Ack(d.ID)
			return e
		}
		apply(v)
		if e := p.q.Ack(d.ID); e != nil {
			return e
		}
	}
	p.mu.Lock()
	b := p.latest
	p.latest = nil
	p.mu.Unlock()
	if b != nil {
		var v T
		if e := json.Unmarshal(b, &v); e != nil {
			return e
		}
		apply(v)
	}

	return nil
}

// Close retains committed actions for drain, drops replaceable latest snapshots,
// and cancels reservations. Native owners detach before closing the pump.
func (p *Pump[T]) Close() {
	p.mu.Lock()
	p.closed = true
	p.latest = nil
	p.q.Close()
	wake := p.arm()
	p.mu.Unlock()
	if wake {
		p.wake()
	}
}
func (p *Pump[T]) Stats() Stats {
	p.mu.Lock()
	defer p.mu.Unlock()
	return Stats{p.q.Stats(), p.refused, p.wakes, len(p.latest), p.closed}
}

func (p *Pump[T]) rearm() {
	p.mu.Lock()
	p.armed = false
	// Peek by acquiring a delivery, then process it on a subsequent wake would
	// strand an in-flight receipt. Instead inspect a queue readiness accessor.
	ready := p.q.Ready() || p.latest != nil
	wake := false
	if ready {
		wake = p.arm()
	}
	p.mu.Unlock()
	if wake {
		p.wake()
	}
}
