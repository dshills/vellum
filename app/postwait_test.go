package app

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCompletionWaitsForCredit(t *testing.T) {
	a, _ := New(Config{QueueCapacity: 1})
	defer a.Close()
	a.Post(func() {})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	done := make(chan error, 1)
	n := 0
	go func() { done <- a.PostWait(ctx, func() { n++ }) }()
	a.Drain()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if n != 1 {
		t.Fatal(n)
	}
	a.Post(func() {})
	go func() { done <- a.PostWait(ctx, func() {}) }()
	a.Close()
	if !errors.Is(<-done, ErrClosed) {
		t.Fatal("blocked waiter survived close")
	}
}
