package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
)

func TestColorOwnerQueuedCancellationAndExecutingJoin(t *testing.T) {
	for _, executing := range []bool{false, true} {
		t.Run(map[bool]string{false: "queued", true: "executing"}[executing], func(t *testing.T) {
			a, _ := app.New(app.Config{})
			defer a.Close()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			queued := make(chan struct{})
			var once sync.Once
			a.SetWake(func() { once.Do(func() { close(queued) }) })
			entered, release := make(chan struct{}), make(chan struct{})
			sentinel := errors.New("callback result")
			done := make(chan error, 1)
			calls := 0
			go func() {
				done <- colorOwner(ctx, a, func() error { calls++; close(entered); <-release; return sentinel })
			}()
			select {
			case <-queued:
			case <-time.After(time.Second):
				t.Fatal("not queued")
			}
			drained := make(chan error, 1)
			if executing {
				go func() { drained <- a.Drain() }()
				select {
				case <-entered:
				case <-time.After(time.Second):
					t.Fatal("not entered")
				}
			}
			cancel()
			if executing {
				select {
				case <-done:
					t.Fatal("executing callback was not joined")
				case <-time.After(10 * time.Millisecond):
				}
				close(release)
			}
			select {
			case err := <-done:
				if !errors.Is(err, context.Canceled) || (executing && !errors.Is(err, sentinel)) {
					t.Fatal("lost cancellation/result", err)
				}
			case <-time.After(time.Second):
				t.Fatal("no terminal owner acknowledgment")
			}
			if executing {
				if err := <-drained; err != nil {
					t.Fatal(err)
				}
			} else {
				if err := a.Drain(); err != nil {
					t.Fatal(err)
				}
			}
			if calls != map[bool]int{false: 0, true: 1}[executing] {
				t.Fatal("revoked/duplicate callback", calls)
			}
		})
	}
}
