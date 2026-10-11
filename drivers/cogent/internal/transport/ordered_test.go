package transport

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/drivers/cogent/internal/mailbox"
)

func orderedTestPump[T any](t *testing.T, source *events.Source, slots, batch int) *OrderedPump[T] {
	t.Helper()
	p, err := NewOrdered[T](mailbox.Limits{Slots: slots, PayloadBytes: slots * 256, MessageBytes: 256}, &source.Deque, batch)
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func sendRecords(t *testing.T, p *OrderedPump[msg], first, last int) {
	t.Helper()
	for i := first; i <= last; i++ {
		if err := p.Send(msg{i, "native"}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestOrderedLegacyBoundaries(t *testing.T) {
	var source events.Source
	p := orderedTestPump[msg](t, &source, 64, 16)
	sendRecords(t, p, 1, 32)
	source.Key(events.KeyDown, 'z', key.CodeZ, key.Meta)
	sendRecords(t, p, 33, 40)
	source.KeyChord('s', key.CodeS, key.Meta)
	sendRecords(t, p, 41, 47)
	applied, maxBatch := 0, 0
	var shortcuts []int
	var batches []int
	for {
		e, ok := source.Deque.PollEvent()
		if !ok {
			break
		}
		before := applied
		handled, err := p.HandleEvent(e, func(v msg) {
			if v.N != applied+1 {
				t.Fatalf("FIFO: got %d after %d", v.N, applied)
			}
			if p.Stats().Queue.InFlight != 1 {
				t.Fatal("credit released before apply")
			}
			applied++
		})
		if err != nil {
			t.Fatal(err)
		}
		if handled {
			size := applied - before
			maxBatch = max(maxBatch, size)
			batches = append(batches, size)
		} else if e.Type() == events.KeyChord {
			shortcuts = append(shortcuts, applied)
		}
	}
	if applied != 47 || maxBatch != 16 || !reflect.DeepEqual(shortcuts, []int{32, 40}) || !reflect.DeepEqual(batches, []int{16, 16, 8, 7}) || p.Stats().Queue.Outstanding != 0 {
		t.Fatal(applied, maxBatch, shortcuts, batches, p.Stats())
	}
}

func TestOrderedNewAdmissionsDuringDispatchStayAfterKey(t *testing.T) {
	var source events.Source
	p := orderedTestPump[msg](t, &source, 64, 16)
	sendRecords(t, p, 1, 32)
	source.KeyChord('z', key.CodeZ, key.Meta)
	applied, atKey := 0, -1
	for {
		e, ok := source.Deque.PollEvent()
		if !ok {
			break
		}
		handled, err := p.HandleEvent(e, func(v msg) {
			if v.N != applied+1 {
				t.Fatal(v, applied)
			}
			applied++
			if applied == 1 {
				sendRecords(t, p, 33, 40)
			}
		})
		if err != nil {
			t.Fatal(err)
		}
		if !handled && e.Type() == events.KeyChord {
			atKey = applied
		}
	}
	if atKey != 32 || applied != 40 {
		t.Fatal(atKey, applied)
	}
}

func TestOrderedPendingCommitAndCancel(t *testing.T) {
	for _, cancel := range []bool{false, true} {
		t.Run(map[bool]string{false: "commit", true: "cancel"}[cancel], func(t *testing.T) {
			var source events.Source
			p := orderedTestPump[msg](t, &source, 4, 16)
			first, err := p.Preflight(msg{1, "reserved"})
			if err != nil {
				t.Fatal(err)
			}
			source.KeyChord('z', key.CodeZ, key.Meta)
			sendRecords(t, p, 2, 2)
			if e, ok := source.Deque.PollEvent(); ok {
				t.Fatal("unready head overtaken", e)
			}
			priority := events.NewKey(events.KeyUp, 0, key.CodeZ, 0)
			source.Deque.SendFirst(priority)
			if e, ok := source.Deque.PollEvent(); !ok || e != priority {
				t.Fatal("priority blocked", e, ok)
			}
			delivered := make(chan events.Event, 1)
			go func() { delivered <- source.Deque.NextEvent() }()
			if cancel {
				err = first.Cancel()
			} else {
				err = first.Commit()
			}
			if err != nil {
				t.Fatal(err)
			}
			var e events.Event
			select {
			case e = <-delivered:
			case <-time.After(time.Second):
				t.Fatal("consumer not signaled")
			}
			var got []int
			atKey := -1
			for {
				handled, err := p.HandleEvent(e, func(v msg) { got = append(got, v.N) })
				if err != nil {
					t.Fatal(err)
				}
				if !handled && e.Type() == events.KeyChord {
					atKey = len(got)
				}
				var ok bool
				e, ok = source.Deque.PollEvent()
				if !ok {
					break
				}
			}
			want := []int{1, 2}
			keyCount := 1
			if cancel {
				want = []int{2}
				keyCount = 0
			}
			if !reflect.DeepEqual(got, want) || atKey != keyCount || p.Stats().Queue.Outstanding != 0 {
				t.Fatal(got, atKey, p.Stats())
			}
		})
	}
}

func TestOrderedSaturationCancellationAndLatestBound(t *testing.T) {
	var source events.Source
	p := orderedTestPump[msg](t, &source, 2, 16)
	source.KeyChord('z', key.CodeZ, key.Meta)
	for i := 0; i < 10000; i++ {
		ticket, err := p.Preflight(msg{i, "cancel"})
		if err != nil {
			t.Fatal(err)
		}
		if err = ticket.Cancel(); err != nil {
			t.Fatal(err)
		}
	}
	if len(source.Deque.Back) != 1 || p.Stats().Queue.Outstanding != 0 {
		t.Fatal("canceled nodes retained", len(source.Deque.Back), p.Stats())
	}
	sendRecords(t, p, 1, 2)
	if _, err := p.Preflight(msg{3, "refused"}); !errors.Is(err, mailbox.ErrFull) {
		t.Fatal(err)
	}
	if len(source.Deque.Back) != 3 {
		t.Fatal("refusal queued a node")
	}
	for i := 0; i < 10000; i++ {
		if err := p.Latest(msg{i, "caret"}); err != nil {
			t.Fatal(err)
		}
	}
	if len(source.Deque.Back) != 4 || p.Stats().LatestBytes > 256 {
		t.Fatal("publication not coalesced", len(source.Deque.Back), p.Stats())
	}
	var got []int
	for {
		e, ok := source.Deque.PollEvent()
		if !ok {
			break
		}
		if _, err := p.HandleEvent(e, func(v msg) { got = append(got, v.N) }); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(got, []int{1, 2, 9999}) {
		t.Fatal(got)
	}
}

func TestOrderedCloseAndDecodeFailure(t *testing.T) {
	var source events.Source
	p := orderedTestPump[msg](t, &source, 4, 16)
	pending, err := p.Preflight(msg{1, "pending"})
	if err != nil {
		t.Fatal(err)
	}
	source.KeyChord('z', key.CodeZ, key.Meta)
	sendRecords(t, p, 2, 3)
	if err := p.Drain(16, func(msg) {}); err == nil {
		t.Fatal("live barrier bypassed")
	}
	p.Close()
	p.Close()
	if err := pending.Commit(); !errors.Is(err, mailbox.ErrStale) {
		t.Fatal(err)
	}
	if err := p.Send(msg{4, "late"}); !errors.Is(err, mailbox.ErrClosed) {
		t.Fatal(err)
	}
	if e, ok := source.Deque.PollEvent(); !ok || e.Type() != events.KeyChord {
		t.Fatal("owner nodes survived close", e, ok)
	}
	if _, ok := source.Deque.PollEvent(); ok {
		t.Fatal("closed node retained")
	}
	var got []int
	if err := p.Drain(16, func(v msg) { got = append(got, v.N) }); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []int{2, 3}) || p.Stats().Queue.Outstanding != 0 {
		t.Fatal(got, p.Stats())
	}
	bad := orderedTestPump[brokenJSON](t, &source, 4, 2)
	for i := 0; i < 3; i++ {
		if err := bad.Send(brokenJSON{}); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		e, ok := source.Deque.PollEvent()
		if !ok {
			t.Fatal("stranded decode batch")
		}
		handled, err := bad.HandleEvent(e, func(brokenJSON) { t.Fatal("invalid record applied") })
		if !handled || err == nil {
			t.Fatal(handled, err)
		}
	}
	if bad.Stats().Queue.Outstanding != 0 || bad.Stats().Queue.InFlight != 0 {
		t.Fatal(bad.Stats())
	}
}

func TestOrderedConcurrentProducers(t *testing.T) {
	var source events.Source
	p := orderedTestPump[msg](t, &source, 64, 16)
	var wg sync.WaitGroup
	for k := 0; k < 4; k++ {
		wg.Add(1)
		go func(k int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				for p.Send(msg{k*200 + i, "x"}) != nil {
				}
			}
		}(k)
	}
	seen := map[int]bool{}
	for len(seen) < 800 {
		e := source.Deque.NextEvent()
		_, err := p.HandleEvent(e, func(v msg) {
			if seen[v.N] {
				t.Fatal("duplicate", v.N)
			}
			seen[v.N] = true
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	wg.Wait()
	if p.Stats().Queue.Outstanding != 0 {
		t.Fatal(p.Stats())
	}
}

func TestOrderedOwnersAndOrdinaryDequeBehavior(t *testing.T) {
	var source events.Source
	first := orderedTestPump[msg](t, &source, 4, 16)
	second := orderedTestPump[msg](t, &source, 4, 16)
	sendRecords(t, first, 1, 1)
	sendRecords(t, second, 2, 2)
	source.KeyChord('z', key.CodeZ, key.Meta)
	first.Close()
	var got []int
	e, ok := source.Deque.PollEvent()
	if !ok {
		t.Fatal("other owner's event removed")
	}
	if handled, err := first.HandleEvent(e, func(msg) { t.Fatal("wrong owner") }); handled || err != nil {
		t.Fatal(handled, err)
	}
	if handled, err := second.HandleEvent(e, func(v msg) { got = append(got, v.N) }); !handled || err != nil {
		t.Fatal(handled, err)
	}
	e, ok = source.Deque.PollEvent()
	if !ok || e.Type() != events.KeyChord {
		t.Fatal(e, ok)
	}
	if err := first.Drain(16, func(v msg) { got = append(got, v.N) }); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []int{2, 1}) {
		t.Fatal(got)
	}
	// Non-deferred events retain compression and priority behavior.
	a, b := &events.CustomEvent{}, &events.CustomEvent{}
	a.Typ = events.Custom
	b.Typ = events.Custom
	a.Init()
	b.Init()
	source.Deque.Send(a)
	source.Deque.Send(b)
	source.Deque.SendFirst(e)
	next, ok := source.Deque.PollEvent()
	if !ok || next != e {
		t.Fatal("priority changed")
	}
	next, ok = source.Deque.PollEvent()
	if !ok || next != b {
		t.Fatal("compression changed")
	}
	if _, ok := source.Deque.PollEvent(); ok {
		t.Fatal("compressed event retained")
	}
}

func TestOrderedOrdinaryCustomCannotReplaceDeferred(t *testing.T) {
	for _, reserved := range []bool{false, true} {
		var source events.Source
		p := orderedTestPump[msg](t, &source, 4, 16)
		var ticket *OrderedTicket[msg]
		var err error
		if reserved {
			ticket, err = p.Preflight(msg{1, "input"})
		} else {
			err = p.Latest(msg{1, "publication"})
		}
		if err != nil {
			t.Fatal(err)
		}
		source.Custom("vellum.wake")
		if reserved {
			if e, ok := source.Deque.PollEvent(); ok {
				t.Fatalf("uncommitted FIFO overwritten by %T", e)
			}
			if err = ticket.Commit(); err != nil {
				t.Fatal(err)
			}
		}
		first, ok := source.Deque.PollEvent()
		if !ok {
			t.Fatal("native event missing")
		}
		var values []int
		if handled, err := p.HandleEvent(first, func(m msg) { values = append(values, m.N) }); err != nil || !handled || !reflect.DeepEqual(values, []int{1}) {
			t.Fatalf("native FIFO replaced: %T %v %v %v", first, handled, values, err)
		}
		second, ok := source.Deque.PollEvent()
		if !ok {
			t.Fatal("ordinary wake missing")
		}
		if ce, ok := second.(*events.CustomEvent); !ok || ce.Data != "vellum.wake" {
			t.Fatalf("ordinary wake changed: %T", second)
		}
		if _, ok := source.Deque.PollEvent(); ok || p.Stats().Queue.Outstanding != 0 || p.Stats().LatestBytes != 0 {
			t.Fatal("publication/credits stranded", p.Stats())
		}
	}
}
