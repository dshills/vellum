package main

import (
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"time"
)

// Application retirement can drop an admitted owner callback. Cancel the
// scenario after Run returns so its private reply wait also loses authority.
type catalogueOutcome[T any] struct {
	report T
	err    error
}

func catalogueWorkerResult[T any](cancel context.CancelFunc, done <-chan catalogueOutcome[T]) (catalogueOutcome[T], bool) {
	cancel() // Release owner waits even when Run exits before their callbacks.
	select {
	case result := <-done:
		return result, true
	case <-time.After(6 * time.Second):
		return catalogueOutcome[T]{}, false
	}
}

// Final teardown is already authorized after all observations. Keep its close
// acknowledgement and closed-reference check bounded but independent of the
// scenario cancellation triggered when closing the last window ends Run.
func catalogueFinishWindow(ctx context.Context, a *app.Application, s *cogent.DesktopSession, ref app.WindowRef, in cogent.DesktopInput) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	finalCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
	defer cancel()
	if err := catalogueOwner(finalCtx, a, ref, func() error { return a.CloseWindow(ref) }); err != nil {
		return ref.Context().Err() != nil, err
	}
	closed := ref.Context().Err() != nil
	err := s.Input(finalCtx, ref, in)
	if !closed || !errors.Is(err, cogent.ErrDesktopClosed) {
		return closed, errors.Join(errors.New("original-window close validation failed"), err)
	}
	return closed, nil
}
