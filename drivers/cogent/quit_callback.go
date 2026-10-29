package cogent

import (
	"github.com/dshills/vellum/app"
	"sync/atomic"
)

// One foreign Quit request may await the UI owner. Native main never waits for
// it, and repeated requests cannot accumulate goroutines behind a busy owner.
type quitCallback struct {
	pending  atomic.Bool
	owner    *app.Application
	callback func()
}

func (q *quitCallback) request() {
	if !q.pending.CompareAndSwap(false, true) {
		return
	}
	go func() {
		defer q.pending.Store(false)
		_ = q.owner.Do(func() {
			if q.owner.Context().Err() == nil {
				q.callback()
			}
		})
	}()
}
