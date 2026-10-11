package transport

import (
	"encoding/json"
	"errors"
	"sync"

	"cogentcore.org/core/events"
	"github.com/dshills/vellum/drivers/cogent/internal/mailbox"
)

// OrderedPump reserves a payload-free FIFO node along with each mailbox credit.
// Adjacent committed nodes merge only at dequeue, up to batch. Legacy events
// therefore separate batches: later native updates cannot cross those events.
// Deque.Mu precedes mu and the mailbox lock. All dispatch/drain calls run on one
// UI owner; admission/commit/cancel may run on producers. Native mutation and
// shutdown must still be serialized by the native view owner.
type OrderedPump[T any] struct {
	mu             sync.Mutex
	q              *mailbox.Queue
	deque          *events.Deque
	batch, max     int
	closed         bool
	latest         []byte
	latestWake     *orderedWake[T]
	refused, wakes uint64
}

type OrderedTicket[T any] struct {
	p     *OrderedPump[T]
	r     *mailbox.Reservation
	data  []byte
	event *orderedWake[T]
}

type orderedWake[T any] struct {
	events.CustomEvent
	owner  *OrderedPump[T]
	ready  bool // protected by Deque.Mu; immutable after successful commit
	count  int  // merges only under Deque.Mu, before delivery to the UI owner
	latest bool
}

func (w *orderedWake[T]) Ready() bool { return w.ready }
func (w *orderedWake[T]) Merge(e events.DeferredEvent) bool {
	other, ok := e.(*orderedWake[T])
	if !ok || other.owner != w.owner || w.latest || other.latest || w.count+other.count > w.owner.batch {
		return false
	}
	w.count += other.count
	return true
}
func (p *OrderedPump[T]) newWake(latest bool) *orderedWake[T] {
	w := &orderedWake[T]{owner: p, latest: latest, ready: latest, count: 1}
	w.Typ = events.Custom
	w.Init()
	w.SetUnique()
	return w
}
func NewOrdered[T any](limits mailbox.Limits, deque *events.Deque, batch int) (*OrderedPump[T], error) {
	if deque == nil || batch <= 0 {
		return nil, errors.New("deque and positive batch required")
	}
	q, err := mailbox.New(limits)
	if err != nil {
		return nil, err
	}
	return &OrderedPump[T]{q: q, deque: deque, batch: batch, max: limits.MessageBytes}, nil
}
func (p *OrderedPump[T]) encode(v T) ([]byte, error) {
	b, err := json.Marshal(v)
	if err == nil && len(b) > p.max {
		err = mailbox.ErrSize
	}
	return b, err
}
func (p *OrderedPump[T]) Preflight(v T) (*OrderedTicket[T], error) {
	b, encodeErr := p.encode(v)
	var t *OrderedTicket[T]
	err := p.deque.SendDeferred(func() (events.DeferredEvent, error) {
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.closed {
			return nil, mailbox.ErrClosed
		}
		if encodeErr != nil {
			p.refused++
			return nil, encodeErr
		}
		r, err := p.q.TryReserve(len(b))
		if err != nil {
			p.refused++
			return nil, err
		}
		w := p.newWake(false)
		t = &OrderedTicket[T]{p: p, r: r, data: b, event: w}
		p.wakes++
		return w, nil
	})
	return t, err
}
func (t *OrderedTicket[T]) Commit() error {
	return t.p.deque.UpdateDeferred(func() error {
		t.p.mu.Lock()
		defer t.p.mu.Unlock()
		if err := t.r.Commit(t.data); err != nil {
			return err
		}
		t.data = nil
		t.event.ready = true
		return nil
	}, nil)
}
func (t *OrderedTicket[T]) Cancel() error {
	return t.p.deque.UpdateDeferred(func() error {
		t.p.mu.Lock()
		defer t.p.mu.Unlock()
		if err := t.r.Cancel(); err != nil {
			return err
		}
		t.data = nil
		return nil
	}, func(e events.DeferredEvent) bool { return e == t.event })
}
func (p *OrderedPump[T]) Send(v T) error {
	t, err := p.Preflight(v)
	if err != nil {
		return err
	}
	return t.Commit()
}

// Latest retains one replaceable publication trigger, separately from actions.
// It never consumes ordered credits or crosses a legacy event with an action.
func (p *OrderedPump[T]) Latest(v T) error {
	b, encodeErr := p.encode(v)
	return p.deque.SendDeferred(func() (events.DeferredEvent, error) {
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.closed {
			return nil, mailbox.ErrClosed
		}
		if encodeErr != nil {
			p.refused++
			return nil, encodeErr
		}
		p.latest = b
		if p.latestWake != nil {
			return nil, nil
		}
		p.latestWake = p.newWake(true)
		p.wakes++
		return p.latestWake, nil
	})
}

// HandleEvent applies only the records represented by this dequeued batch.
// All captured records are acknowledged even if one fails decoding, so a bad
// encoding cannot strand later records whose nodes merged into the same event.
func (p *OrderedPump[T]) HandleEvent(e events.Event, apply func(T)) (bool, error) {
	w, ok := e.(*orderedWake[T])
	if !ok || w.owner != p {
		return false, nil
	}
	e.SetHandled()
	if w.latest {
		p.mu.Lock()
		b := p.latest
		p.latest = nil
		p.latestWake = nil
		p.mu.Unlock()
		if b == nil {
			return true, nil
		}
		var v T
		if err := json.Unmarshal(b, &v); err != nil {
			return true, err
		}
		apply(v)
		return true, nil
	}
	return true, p.drain(w.count, apply, true)
}
func (p *OrderedPump[T]) drain(batch int, apply func(T), strict bool) error {
	var result error
	for n := 0; n < batch; n++ {
		d, ok := p.q.TryNext()
		if !ok {
			if strict {
				result = errors.Join(result, errors.New("deferred batch missing a committed record"))
			}
			break
		}
		var v T
		err := json.Unmarshal(d.Data, &v)
		if err == nil {
			apply(v)
		}
		result = errors.Join(result, err, p.q.Ack(d.ID))
	}
	return result
}

// Close removes this owner's queued nodes atomically, cancels reservations and
// retains committed records. Shutdown Drain is legal only after Close, with the
// event loop quiescent on the same UI owner. It cannot bypass a live FIFO barrier.
func (p *OrderedPump[T]) Close() {
	_ = p.deque.UpdateDeferred(func() error {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.closed = true
		p.latest = nil
		p.latestWake = nil
		p.q.Close()
		return nil
	}, func(e events.DeferredEvent) bool { w, ok := e.(*orderedWake[T]); return ok && w.owner == p })
}
func (p *OrderedPump[T]) Drain(batch int, apply func(T)) error {
	p.mu.Lock()
	closed := p.closed
	p.mu.Unlock()
	if !closed || batch <= 0 || batch > p.batch {
		return errors.New("bounded shutdown drain requires a closed ordered pump")
	}
	return p.drain(batch, apply, false)
}
func (p *OrderedPump[T]) Stats() Stats {
	p.mu.Lock()
	defer p.mu.Unlock()
	return Stats{p.q.Stats(), p.refused, p.wakes, len(p.latest), p.closed}
}
