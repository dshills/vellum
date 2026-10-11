package execution

import (
	"errors"
	"time"
)

var ErrInvalidClock = errors.New("execution: use a constructor to initialize a clock")

// ClockReadiness captures registered work relative to one captured Now under
// the clock lock. Inspection never advances time, starts callbacks, or retires
// timers. A canceled timer awaiting its stop observer still counts as pending.
type ClockReadiness struct {
	ClockSnapshot
	Now, NextDeadline time.Time
	Due, Future       int
	HasDeadline       bool
}

// CurrentIdle means quiescent at Now. Future timers alone do not require work
// until the caller explicitly advances manual time. Idle retains the inherited
// ClockSnapshot meaning: no registered pending/running timer or active advance.
func (s ClockReadiness) CurrentIdle() bool {
	return s.Due == 0 && s.Running == 0 && !s.Advancing
}

func (c *clockBase) readiness() (ClockReadiness, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.capacity < 1 || c.timers == nil {
		return ClockReadiness{}, ErrInvalidClock
	}
	s := ClockReadiness{ClockSnapshot: c.s, Now: c.nowLocked()}
	s.Advancing = c.advancing
	for timer := range c.timers {
		if !timer.pending {
			continue
		}
		if timer.deadline.After(s.Now) {
			s.Future++
		} else {
			s.Due++
		}
		if !s.HasDeadline || timer.deadline.Before(s.NextDeadline) {
			s.HasDeadline = true
			s.NextDeadline = timer.deadline
		}
	}
	return s, nil
}

// ReadySnapshot is a checked, read-only current-instant inspection. A closed
// initialized clock remains inspectable so active callback teardown is visible.
func (c *ManualClock) ReadySnapshot() (ClockReadiness, error) {
	if c == nil {
		return ClockReadiness{}, ErrInvalidClock
	}
	return c.base.readiness()
}

func (c *RealClock) ReadySnapshot() (ClockReadiness, error) {
	if c == nil {
		return ClockReadiness{}, ErrInvalidClock
	}
	return c.base.readiness()
}
