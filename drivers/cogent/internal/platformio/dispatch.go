package platformio

import (
	"context"
	"example.com/vellum/app"
)

type Result[T any] struct {
	Value T
	Err   error
}

// Dispatch returns values through a private buffered channel. A canceled caller
// never leaves a native callback blocked or writing into returned caller storage.
func Dispatch[T any](ctx context.Context, stopped <-chan struct{}, enqueue func(context.Context, func()) error, action func() (T, error)) (T, error) {
	var zero T
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	select {
	case <-stopped:
		return zero, app.ErrClosed
	default:
	}
	result := make(chan Result[T], 1)
	run := func() {
		select {
		case <-stopped:
			result <- Result[T]{Err: app.ErrClosed}
			return
		default:
		}
		if err := ctx.Err(); err != nil {
			result <- Result[T]{Err: err}
			return
		}
		value, err := action()
		result <- Result[T]{Value: value, Err: err}
	}
	if err := enqueue(ctx, run); err != nil {
		return zero, err
	}
	select {
	case r := <-result:
		return r.Value, r.Err
	case <-ctx.Done():
		return zero, ctx.Err()
	case <-stopped:
		return zero, app.ErrClosed
	}
}
