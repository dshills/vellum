package app

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

// Exercise the public Drain/Do chain, not runCompletionHooks in isolation.
// Cleanup is intentionally abnormal to verify unrelated terminal accounting.
func TestOwnerAndCleanupFailureStillCloseAndRetireEveryTicket(t *testing.T) {
	for _, ownerFailure := range []string{"panic", "goexit"} {
		for _, activeFailure := range []string{"panic", "goexit"} {
			for _, revokedFailure := range []string{"normal", "panic", "goexit"} {
				t.Run(ownerFailure+"/"+activeFailure+"/"+revokedFailure, func(t *testing.T) {
					a := completionApp(t, 1, 4)
					var wakes atomic.Int32
					a.SetWake(func() { _ = a.Pending(); wakes.Add(1) })
					var cleaned [4]atomic.Int32
					fail := func(kind, payload string) {
						if kind == "panic" {
							panic(payload)
						}
						if kind == "goexit" {
							runtime.Goexit()
						}
					}
					first := ticket(t, a, "", func(error) {
						_ = a.Pending()
						cleaned[0].Add(1)
						fail(activeFailure, "active cleanup")
					})
					ready := ticket(t, a, "", func(error) { _ = a.Pending(); cleaned[1].Add(1) })
					reserved := ticket(t, a, "", func(error) {
						_ = a.Pending()
						cleaned[2].Add(1)
						fail(revokedFailure, "revoked cleanup")
					})
					remaining := ticket(t, a, "", func(error) { _ = a.Pending(); cleaned[3].Add(1) })
					var ran atomic.Int32
					if err := first.Publish(func() { fail(ownerFailure, "owner") }); err != nil {
						t.Fatal(err)
					}
					if err := ready.Publish(func() { ran.Add(1) }); err != nil {
						t.Fatal(err)
					}
					wakes.Store(0) // Only the shutdown notification is counted below.
					terminal := make(chan any, 1)
					go func() {
						defer func() { terminal <- recover() }()
						_, _ = a.DrainN(7)
					}()
					var value any
					select {
					case value = <-terminal:
					case <-time.After(3 * time.Second):
						t.Fatal("failure unwind did not terminate")
					}
					if revokedFailure == "panic" && value != "revoked cleanup" || revokedFailure == "goexit" && value != nil || revokedFailure == "normal" && activeFailure == "panic" && value != "active cleanup" || revokedFailure == "normal" && activeFailure == "goexit" && value != nil {
						t.Fatalf("unexpected ordinary Go unwinding result %v", value)
					}
					if a.Context().Err() != context.Canceled || !a.Pending().Closed || !a.Pending().Idle() || wakes.Load() != 1 || ran.Load() != 0 {
						t.Fatalf("shutdown stranded state%+v context%v wakes%d later callbacks%d", a.Pending(), a.Context().Err(), wakes.Load(), ran.Load())
					}
					for index, c := range []*Completion{first, ready, reserved, remaining} {
						if cleaned[index].Load() != 1 {
							t.Fatalf("cleanup%d count%d", index, cleaned[index].Load())
						}
						select {
						case <-c.Done():
						default:
							t.Fatalf("ticket%d Done stranded", index)
						}
						expected := ErrClosed
						if index == 0 {
							expected = ErrOwnerCallbackFailed
						}
						if !errors.Is(c.Err(), expected) {
							t.Fatalf("ticket%d error%v", index, c.Err())
						}
					}
					if err := a.Do(func() { t.Error("callback after failure") }); !errors.Is(err, ErrClosed) {
						t.Fatalf("owner remains locked/live: %v", err)
					}
				})
			}
		}
	}
}
