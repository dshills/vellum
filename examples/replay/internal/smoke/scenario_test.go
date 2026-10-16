package smoke

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestScenarioHandoffWaitsForRetirement(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	retiring := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	releaseRetirement := func() { once.Do(func() { close(release) }) }
	defer releaseRetirement()
	want := errors.New("retained scenario failure")
	state := 0
	result, done := StartScenario(func() error { state = 1; return want }, func() { close(retiring); <-release; state = 2 })
	// Even a failed assertion releases and joins this actual owned goroutine.
	defer func() {
		releaseRetirement()
		select {
		case <-done:
		case <-ctx.Done():
			t.Error("injected worker did not join")
		}
	}()
	select {
	case <-retiring:
	case <-ctx.Done():
		t.Fatal("retirement not reached")
	}
	select {
	case <-done:
		t.Fatal("completion published before retirement returned")
	default:
	}
	select {
	case <-result:
		t.Fatal("outcome handed off while retirement held")
	default:
	}
	releaseRetirement()
	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("completion did not follow retirement")
	}
	if err := <-result; !errors.Is(err, want) {
		t.Fatal("scenario error lost", err)
	}
	if state != 2 {
		t.Fatal("controller state observed before retirement completed", state)
	}
}
