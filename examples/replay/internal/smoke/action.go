// Package smoke supplies bounded, single-admission acceptance steps. Native
// transports and owner-state oracles are injected by the native consumer.
package smoke

import (
	"context"
	"errors"
	"time"
)

type Record struct {
	Name                 string
	Attempts, Admissions int
	OwnerCompleted       bool
	ElapsedSeconds       float64
}

type Hooks struct {
	Ready    func(context.Context) (string, error)
	Perform  func(context.Context, string) (bool, error)
	Accepted func(context.Context) (bool, error)
}

// Action refuses an already-satisfied oracle and sends at most one native
// proposal. An admitted action is never retried while publication catches up.
// A transport refusal or an uncertain outcome fails the step conservatively.
func Action(ctx context.Context, name string, h Hooks) (record Record, err error) {
	start := time.Now()
	record.Name = name
	defer func() { record.ElapsedSeconds = time.Since(start).Seconds() }()
	if ctx == nil || h.Ready == nil || h.Perform == nil || h.Accepted == nil {
		return record, errors.New("smoke: missing context or hooks")
	}
	// Hooks are cooperative: native calls and owner waits must honor this
	// context. The helper supplies its own cap and never extends a caller's.
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return record, err
	}
	accepted, err := h.Accepted(ctx)
	if err != nil {
		return record, err
	}
	if accepted {
		return record, errors.New("smoke: oracle already satisfied before proposal")
	}
	var target string
	for target == "" {
		if err = ctx.Err(); err != nil {
			return record, err
		}
		target, err = h.Ready(ctx)
		if err != nil {
			return record, err
		}
		if target == "" {
			if err = pause(ctx); err != nil {
				return record, err
			}
		}
	}
	if err = ctx.Err(); err != nil {
		return record, err
	}
	record.Attempts++
	admitted, err := h.Perform(ctx, target)
	if admitted {
		record.Admissions++
	}
	if err != nil {
		return record, err
	}
	if !admitted {
		return record, errors.New("smoke: native proposal refused")
	}
	for {
		if err = ctx.Err(); err != nil {
			return record, err
		}
		accepted, err = h.Accepted(ctx)
		if err != nil {
			return record, err
		}
		if accepted {
			record.OwnerCompleted = true
			return record, nil
		}
		if err = pause(ctx); err != nil {
			return record, err
		}
	}
}

func pause(ctx context.Context) error {
	timer := time.NewTimer(20 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
