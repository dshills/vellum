package main

import (
	"context"
	"errors"
	"testing"
)

func TestReadinessMatchesExpectedNativeLifetime(t *testing.T) {
	ready := make(chan string, 4)
	ready <- "main.1"
	ready <- "floating.2"
	ready <- "floating.2"
	ready <- "floating.3"
	id, err := waitNativeReady(context.Background(), ready, make(chan error), "floating.3")
	if err != nil || id != "floating.3" {
		t.Fatal("stale window readiness accepted", id, err)
	}
}
func TestReadinessCannotSucceedWithoutExpectedLifetime(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ready := make(chan string, 1)
	ready <- "retired"
	cancel()
	_, err := waitNativeReady(ctx, ready, make(chan error), "new")
	if !errors.Is(err, context.Canceled) {
		t.Fatal("wrong-window readiness satisfied wait", err)
	}
}

func TestRunReturnsArgumentErrorsWithoutExiting(t *testing.T) {
	for i := 0; i < 2; i++ {
		if err := run([]string{"-timeout", "invalid-duration"}); err == nil {
			t.Fatal("invalid duration accepted")
		}
	}
}
