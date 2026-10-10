package cogent

import (
	"errors"

	"cogentcore.org/core/events"
	"github.com/dshills/vellum/app"
)

// Gate the Scene before its Close walk reaches body/widget cleanup. Forced
// application/window cancellation bypasses user policy so teardown always ends.
func bindCloseRequest(w *nativeWindow, a *app.Application, onError func(error)) {
	w.body.Scene.OnFirst(events.Close, func(e events.Event) {
		if w.closing.Load() || a.Context().Err() != nil {
			return
		}
		lifetime, err := a.WindowContext(w.window.ID)
		if err != nil || lifetime.Err() != nil {
			return
		}
		var decision error
		err = a.Do(func() { decision = a.CheckCloseWindow(w.window.ID) })
		if errors.Is(err, app.ErrClosed) || errors.Is(decision, app.ErrClosed) || errors.Is(decision, app.ErrWindowClosed) {
			return
		}
		if decision != nil || err != nil {
			e.SetHandled()
			if onError != nil && !errors.Is(decision, app.ErrCloseRefused) && !errors.Is(decision, app.ErrClosePending) {
				// Report on the owner, never while holding a foreign native lock.
				if err != nil {
					decision = err
				}
				a.PostWindow(w.window.ID, func() { onError(decision) })
			}
		}
	})
}
