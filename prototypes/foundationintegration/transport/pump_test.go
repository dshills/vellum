package transport

import (
	"errors"
	"example.com/vellum/foundationcontracts/mailbox"
	"sync"
	"testing"
)

type msg struct {
	N    int
	Text string
}

func TestSaturationPrecedesMutation(t *testing.T) {
	wakes := 0
	p, _ := New[msg](mailbox.Limits{Slots: 2, PayloadBytes: 512, MessageBytes: 256}, func() { wakes++ })
	mutations := 0
	for i := 0; i < 100; i++ {
		r, e := p.Preflight(msg{i, "👩中文"})
		if e != nil {
			continue
		}
		mutations++
		if e = r.Commit(); e != nil {
			t.Fatal(e)
		}
	}
	if mutations != 2 || wakes != 1 || p.Stats().Refused != 98 {
		t.Fatal(mutations, wakes, p.Stats())
	}
	var got []int
	if e := p.Drain(1, func(v msg) {
		got = append(got, v.N)
		if p.Stats().Queue.InFlight != 1 {
			t.Fatal("early credit release")
		}
	}); e != nil {
		t.Fatal(e)
	}
	if wakes != 2 {
		t.Fatal("batch not rearmed", wakes)
	}
	p.Drain(1, func(v msg) { got = append(got, v.N) })
	if len(got) != 2 || got[0] != 0 || got[1] != 1 || p.Stats().Queue.Outstanding != 0 {
		t.Fatal(got, p.Stats())
	}
}
func TestReservationOrderCloseAndLatest(t *testing.T) {
	wakes := 0
	p, _ := New[msg](mailbox.Limits{Slots: 3, PayloadBytes: 512, MessageBytes: 256}, func() { wakes++ })
	first, _ := p.Preflight(msg{1, "pending"})
	second, _ := p.Preflight(msg{2, "accepted"})
	second.Commit()
	for i := 0; i < 10000; i++ {
		if e := p.Latest(msg{i, "caret"}); e != nil {
			t.Fatal(e)
		}
	}
	got := []int{}
	p.Drain(3, func(v msg) { got = append(got, v.N) })
	if len(got) != 1 || got[0] != 9999 || wakes != 1 {
		t.Fatal(got, wakes)
	}
	first.Cancel()
	p.Drain(3, func(v msg) { got = append(got, v.N) })
	pending, _ := p.Preflight(msg{3, "canceled at shutdown"})
	p.Send(msg{4, "accepted before shutdown"})
	p.Close()
	if e := pending.Commit(); e != mailbox.ErrStale {
		t.Fatal(e)
	}
	p.Drain(3, func(v msg) { got = append(got, v.N) })
	if len(got) != 3 || got[1] != 2 || got[2] != 4 || p.Stats().Queue.Outstanding != 0 {
		t.Fatal(got, p.Stats())
	}
}
func TestConcurrentWakeRearm(t *testing.T) {
	wake := make(chan struct{}, 1)
	p, _ := New[msg](mailbox.Limits{Slots: 64, PayloadBytes: 16384, MessageBytes: 256}, func() {
		select {
		case wake <- struct{}{}:
		default:
			t.Error("duplicate queued wake")
		}
	})
	var wg sync.WaitGroup
	for j := 0; j < 4; j++ {
		wg.Add(1)
		go func(k int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				for p.Send(msg{k*1000 + i, "x"}) != nil {
				}
			}
		}(j)
	}
	seen := map[int]bool{}
	for len(seen) < 4000 {
		<-wake
		if e := p.Drain(7, func(v msg) {
			if seen[v.N] {
				t.Fatal("duplicate", v.N)
			}
			seen[v.N] = true
		}); e != nil {
			t.Fatal(e)
		}
	}
	wg.Wait()
	if p.Stats().Queue.Outstanding != 0 {
		t.Fatal(p.Stats())
	}
}

type brokenJSON struct{}

func (brokenJSON) MarshalJSON() ([]byte, error) { return []byte(`{"valid":true}`), nil }
func (*brokenJSON) UnmarshalJSON([]byte) error  { return errors.New("deliberate decode failure") }
func TestDecodeFailureReleasesAndRearms(t *testing.T) {
	wakes := 0
	p, _ := New[brokenJSON](mailbox.Limits{Slots: 2, PayloadBytes: 512, MessageBytes: 256}, func() { wakes++ })
	p.Send(brokenJSON{})
	p.Send(brokenJSON{})
	if e := p.Drain(1, func(brokenJSON) { t.Fatal("invalid message applied") }); e == nil {
		t.Fatal("decode failure hidden")
	}
	if p.Stats().Queue.InFlight != 0 || p.Stats().Queue.Outstanding != 1 || wakes != 2 {
		t.Fatal(p.Stats(), wakes)
	}
	if e := p.Drain(1, func(brokenJSON) {}); e == nil {
		t.Fatal("decode failure hidden")
	}
	p.Send(brokenJSON{})
	if wakes != 3 {
		t.Fatal("pump stranded after failure", wakes)
	}
}
