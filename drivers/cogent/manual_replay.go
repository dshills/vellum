package cogent

import (
	"context"
	"github.com/dshills/vellum/replay"
	"time"
)

// ReplayReadiness observes current manual work without dispatching, choosing a
// worker or moving time. It reports window layout fences and future-only paint
// deadlines separately from ready work. Desktop adapters refuse as Snapshot does.
func (m *Manual) ReplayReadiness(ctx context.Context) (replay.Readiness, error) {
	s, err := m.Snapshot(ctx)
	if err != nil {
		return replay.Readiness{}, err
	}
	if s.Failure != nil {
		return replay.Readiness{}, s.Failure
	}
	r := replay.Readiness{Now: s.Clock.Now, Ready: s.Application.Ordinary + s.Application.ReadyCompletions, Running: s.Executor.Running + s.Clock.Running + s.Application.Dispatching + s.Application.Cleanup, Awaiting: s.Executor.Pending + s.Application.Reserved, Blocked: s.Clock.Advancing, DueTimers: s.Clock.Due, FutureTimers: s.Clock.Future, NextDeadline: s.Clock.NextDeadline}
	deadline := func(t time.Time) {
		if r.NextDeadline.IsZero() || t.Before(r.NextDeadline) {
			r.NextDeadline = t
		}
	}
	for _, w := range s.Windows {
		if !w.Now.Equal(r.Now) {
			return r, replay.ErrTime
		}
		r.Ready += w.Events
		r.Blocked = r.Blocked || w.Blocked || w.LayoutPending
		if w.ReadyPaint {
			r.Ready++
		}
		if w.FuturePaint {
			r.FutureTimers++
			deadline(w.NextPaint)
		}
	}
	return r, nil
}
