package cogent

import (
	"github.com/dshills/vellum/app"
	"sync/atomic"
	"testing"
	"time"
)

func TestQuitCallbackCoalescesAndRetires(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	var calls atomic.Int32
	q := &quitCallback{owner: a, callback: func() { calls.Add(1) }}
	_ = a.Do(func() {
		for range 100 {
			q.request()
		}
		if !q.pending.Load() {
			t.Fatal("request not retained")
		}
	})
	end := time.Now().Add(time.Second)
	for q.pending.Load() && time.Now().Before(end) {
		time.Sleep(time.Millisecond)
	}
	if calls.Load() != 1 || q.pending.Load() {
		t.Fatal("duplicated or retained Quit", calls.Load())
	}
	a.Close()
	q.request()
	for q.pending.Load() && time.Now().Before(end) {
		time.Sleep(time.Millisecond)
	}
	if calls.Load() != 1 {
		t.Fatal("closed app callback")
	}
}
