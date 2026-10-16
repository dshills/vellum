//go:build vellum_native_checks

package main

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/dshills/vellum/app"
)

func cycleWindowID(index int, sameName bool) string {
	if sameName {
		return "native-cycle-reused"
	}
	return fmt.Sprintf("native-cycle-%04d", index)
}

func validateSameNameMode(enabled, composed bool, cycles int, soak time.Duration) error {
	if enabled && (!composed || cycles < 2 || soak != 0) {
		return errors.New("same-name requires composed, at least two cycles and no soak")
	}
	return nil
}

type reopenedLifetimeProbe struct {
	unexpected atomic.Bool
}

// Call on the application owner after opening the replacement. The returned
// probe must be checked after a later owner acknowledgment: an admission error
// alone does not prove that no callback was incorrectly enqueued.
func probeReopenedLifetime(a *app.Application, old, current app.WindowRef, completion *app.Completion) (*reopenedLifetimeProbe, error) {
	if old == (app.WindowRef{}) || current == (app.WindowRef{}) || old == current || old.ID() != current.ID() || old.Context().Err() != context.Canceled || current.Context().Err() != nil || completion == nil {
		return nil, errors.New("same-name probe requires a retired opening and a distinct live replacement")
	}
	if ref, err := a.WindowRef(current.ID()); err != nil || ref != current {
		return nil, errors.New("same-name probe replacement is not the registered opening")
	}
	probe := &reopenedLifetimeProbe{}
	for _, action := range []struct {
		name string
		run  func() error
	}{
		{"post", func() error { return a.PostWindow(old, func() { probe.unexpected.Store(true) }) }},
		{"close", func() error { return a.CloseWindow(old) }},
		{"close policy", func() error { return a.CheckCloseWindow(old) }},
		{"request close", func() error { return a.RequestCloseWindow(old) }},
		{"completion publication", func() error { return completion.Publish(func() { probe.unexpected.Store(true) }) }},
	} {
		if err := action.run(); !errors.Is(err, app.ErrWindowClosed) {
			return nil, fmt.Errorf("same-name old %s did not refuse: %v", action.name, err)
		}
	}
	reserved, err := a.ReserveWindowCompletion(old, nil)
	if reserved != nil {
		reserved.Abort()
	}
	if !errors.Is(err, app.ErrWindowClosed) || reserved != nil {
		return nil, fmt.Errorf("same-name old reservation did not refuse: %v", err)
	}
	// Establish a new dirty bit so an old reference cannot pass merely because
	// the native driver already consumed it. Rearm it for the next native frame.
	if err := a.Invalidate(current.ID()); err != nil {
		return nil, err
	}
	preserved := !a.TakeDirty(old) && a.TakeDirty(current)
	rearmErr := a.Invalidate(current.ID())
	if !preserved {
		return nil, errors.Join(errors.New("same-name old reference consumed replacement dirty state"), rearmErr)
	}
	if rearmErr != nil {
		return nil, fmt.Errorf("same-name replacement repaint rearm: %w", rearmErr)
	}
	return probe, nil
}

func (p *reopenedLifetimeProbe) afterOwnerBarrier(a *app.Application, current app.WindowRef) error {
	if p == nil || p.unexpected.Load() {
		return errors.New("same-name old callback reached replacement")
	}
	if ref, err := a.WindowRef(current.ID()); err != nil || ref != current || current.Context().Err() != nil {
		return errors.New("same-name replacement retired during old-reference probes")
	}
	return nil
}
