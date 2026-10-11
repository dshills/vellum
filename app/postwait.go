package app

import (
	"context"
	"fmt"
)

// PostWait waits for mailbox space or cancellation, without losing a completion.
// Call only from bounded background workers, never from the UI owner.
func (a *Application) PostWait(ctx context.Context, fn func()) error {
	if ctx == nil || fn == nil {
		return fmt.Errorf("vellum: context and callback required")
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		a.mu.Lock()
		if a.closed {
			a.mu.Unlock()
			return ErrClosed
		}
		if a.normalPending < a.capacity {
			if err := a.postLocked(fn); err != nil {
				a.mu.Unlock()
				return err
			}
			wake := a.wake
			a.mu.Unlock()
			if wake != nil {
				wake()
			}
			return nil
		}
		changed := a.changed
		a.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-a.ctx.Done():
			return ErrClosed
		case <-changed:
		}
	}
}
