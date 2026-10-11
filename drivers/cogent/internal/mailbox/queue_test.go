package mailbox

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
)

func newQueue(t *testing.T, l Limits) *Queue {
	t.Helper()
	q, err := New(l)
	if err != nil {
		t.Fatal(err)
	}
	return q
}
func TestAdmissionAndAckOwnership(t *testing.T) {
	q := newQueue(t, Limits{2, 6, 4})
	first, err := q.TryReserve(4)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = q.TryReserve(3); !errors.Is(err, ErrFull) {
		t.Fatal(err)
	}
	if err := first.Commit([]byte("bad")); !errors.Is(err, ErrSize) {
		t.Fatal(err)
	}
	source := []byte("abcd")
	if err := first.Commit(source); err != nil {
		t.Fatal(err)
	}
	source[0] = 'X'
	second, err := q.TryReserve(2)
	if err != nil {
		t.Fatal(err)
	}
	if err := second.Commit([]byte("ef")); err != nil {
		t.Fatal(err)
	}
	d, ok := q.TryNext()
	if !ok || string(d.Data) != "abcd" {
		t.Fatal(d, ok)
	}
	if _, err = q.TryReserve(0); !errors.Is(err, ErrFull) {
		t.Fatal("inflight credit released", err)
	}
	if _, ok = q.TryNext(); ok {
		t.Fatal("more than one inflight")
	}
	if err := q.Ack(d.ID + 1); !errors.Is(err, ErrStale) {
		t.Fatal(err)
	}
	if err := q.Ack(d.ID); err != nil {
		t.Fatal(err)
	}
	if err := q.Ack(d.ID); !errors.Is(err, ErrStale) {
		t.Fatal(err)
	}
	d, ok = q.TryNext()
	if !ok || string(d.Data) != "ef" {
		t.Fatal(d, ok)
	}
	if err := q.Ack(d.ID); err != nil {
		t.Fatal(err)
	}
	if s := q.Stats(); s.Outstanding != 0 || s.PayloadBytes != 0 {
		t.Fatal(s)
	}
	if err := first.Commit([]byte("abcd")); !errors.Is(err, ErrStale) {
		t.Fatal(err)
	}
}
func TestReservationOrderAndCancel(t *testing.T) {
	q := newQueue(t, Limits{3, 12, 4})
	a, _ := q.TryReserve(4)
	b, _ := q.TryReserve(4)
	c, _ := q.TryReserve(4)
	if err := c.Commit([]byte("last")); err != nil {
		t.Fatal(err)
	}
	if _, ok := q.TryNext(); ok {
		t.Fatal("overtook reservation")
	}
	if err := b.Cancel(); err != nil {
		t.Fatal(err)
	}
	if err := a.Commit([]byte("head")); err != nil {
		t.Fatal(err)
	}
	d, _ := q.TryNext()
	if string(d.Data) != "head" {
		t.Fatal(d)
	}
	if err := q.Ack(d.ID); err != nil {
		t.Fatal(err)
	}
	d, _ = q.TryNext()
	if string(d.Data) != "last" {
		t.Fatal(d)
	}
	if err := q.Ack(d.ID); err != nil {
		t.Fatal(err)
	}
}
func TestCloseDrainsAcceptedAndWakesWorkers(t *testing.T) {
	q := newQueue(t, Limits{2, 8, 4})
	if err := q.Submit(context.Background(), []byte("kept")); err != nil {
		t.Fatal(err)
	}
	r, _ := q.TryReserve(4)
	result := make(chan error, 1)
	go func() { result <- q.Submit(context.Background(), []byte("wait")) }()
	q.Close()
	q.Close()
	if err := <-result; !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	if err := r.Commit([]byte("lost")); !errors.Is(err, ErrStale) {
		t.Fatal(err)
	}
	d, err := q.Next(context.Background())
	if err != nil || string(d.Data) != "kept" {
		t.Fatal(d, err)
	}
	if err := q.Ack(d.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = q.Next(context.Background()); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
}
func TestCancellationAndLimits(t *testing.T) {
	for _, l := range []Limits{{0, 8, 4}, {1, 0, 4}, {1, 8, 9}} {
		if _, err := New(l); !errors.Is(err, ErrSize) {
			t.Fatal(l, err)
		}
	}
	q := newQueue(t, Limits{1, 8, 4})
	for _, size := range []int{-1, 5} {
		if _, err := q.TryReserve(size); !errors.Is(err, ErrSize) {
			t.Fatal(size, err)
		}
	}
	if err := q.Submit(context.Background(), []byte("head")); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := q.Submit(ctx, []byte("no")); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if s := q.Stats(); s.Outstanding != 1 || s.PayloadBytes != 4 {
		t.Fatal(s)
	}
}
func TestConcurrentBoundedDelivery(t *testing.T) {
	q := newQueue(t, Limits{8, 128, 32})
	ctx := context.Background()
	const producers = 4
	const each = 500
	var wg sync.WaitGroup
	wg.Add(producers)
	for p := range producers {
		go func() {
			defer wg.Done()
			for i := range each {
				if err := q.Submit(ctx, []byte(fmt.Sprintf("%d:%d", p, i))); err != nil {
					t.Error(err)
					return
				}
			}
		}()
	}
	go func() { wg.Wait(); q.Close() }()
	seen := make(map[string]bool)
	last := make(map[string]int)
	var id uint64
	for {
		d, err := q.Next(ctx)
		if errors.Is(err, ErrClosed) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if d.ID <= id {
			t.Fatal("receipt order", d.ID, id)
		}
		id = d.ID
		if seen[string(d.Data)] {
			t.Fatal("duplicate")
		}
		seen[string(d.Data)] = true
		fields := bytes.Split(d.Data, []byte(":"))
		var index int
		fmt.Sscanf(string(fields[1]), "%d", &index)
		key := string(fields[0])
		previous, exists := last[key]
		if exists && index != previous+1 {
			t.Fatal("producer order")
		}
		last[key] = index
		if s := q.Stats(); s.Outstanding > 8 || s.PayloadBytes > 128 {
			t.Fatal(s)
		}
		if err := q.Ack(d.ID); err != nil {
			t.Fatal(err)
		}
	}
	if len(seen) != producers*each {
		t.Fatal("lost accepted messages", len(seen))
	}
}

func TestCanceledMiddleReservationDoesNotOvertake(t *testing.T) {
	q := newQueue(t, Limits{2, 8, 4})
	a, _ := q.TryReserve(4)
	b, _ := q.TryReserve(4)
	if err := b.Cancel(); err != nil {
		t.Fatal(err)
	}
	if _, err := q.TryReserve(4); !errors.Is(err, ErrFull) {
		t.Fatal("reuse middle slot before head", err)
	}
	if err := a.Commit([]byte("head")); err != nil {
		t.Fatal(err)
	}
	d, _ := q.TryNext()
	if err := q.Ack(d.ID); err != nil {
		t.Fatal(err)
	}
	c, err := q.TryReserve(4)
	if err != nil {
		t.Fatal(err)
	}
	if err = b.Commit([]byte("nope")); !errors.Is(err, ErrStale) {
		t.Fatal("stale ticket reuses slot", err)
	}
	if err = c.Commit([]byte("next")); err != nil {
		t.Fatal(err)
	}
	d, _ = q.TryNext()
	if string(d.Data) != "next" {
		t.Fatal(d)
	}
}
