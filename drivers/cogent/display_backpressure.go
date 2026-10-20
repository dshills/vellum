// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
package cogent

import (
	"cogentcore.org/core/system/driver/desktop"
	"errors"
	"sync"
	"time"
)

// DisplayBackpressureError is the typed terminal refusal of the Mac experiment.
type DisplayBackpressureError = desktop.DisplayBackpressureError

type displayExperimentTarget interface {
	EnableDisplayBackpressure(time.Duration) error
	DisplayBackpressureRefusals() <-chan error
	RetireDisplayBackpressure()
}

func attachDisplayBackpressure(target displayExperimentTarget, timeout time.Duration, fail func(error)) (*displayFailureMonitor, error) {
	if target == nil {
		return nil, errors.New("cogent: missing display experiment target")
	}
	if err := target.EnableDisplayBackpressure(timeout); err != nil {
		return nil, err
	}
	monitor, err := startDisplayFailureMonitor(target.DisplayBackpressureRefusals(), target.RetireDisplayBackpressure, fail)
	if err != nil {
		target.RetireDisplayBackpressure()
		return nil, err
	}
	return monitor, nil
}

func validateDisplayBackpressure(timeout time.Duration) error {
	if timeout != 0 && (timeout < 100*time.Millisecond || timeout > 5*time.Second) {
		return errors.New("cogent: experimental display backpressure requires 100ms..5s")
	}
	return nil
}

// One monitor belongs to one native attachment. The backend owns and closes its
// buffer1 refusal channel. No callback enters native/main or waits on this owner.
type displayFailureMonitor struct {
	stop, done chan struct{}
	once       sync.Once
	retire     func()
}

func startDisplayFailureMonitor(refusals <-chan error, retire func(), fail func(error)) (*displayFailureMonitor, error) {
	if refusals == nil || retire == nil || fail == nil {
		return nil, errors.New("cogent: display refusal monitor requires lifetime and failure sink")
	}
	m := &displayFailureMonitor{stop: make(chan struct{}), done: make(chan struct{}), retire: retire}
	go func() {
		defer close(m.done)
		select {
		case <-m.stop:
			return
		case err, ok := <-refusals:
			if !ok {
				return
			}
			select {
			case <-m.stop:
				return
			default:
			}
			if err != nil {
				fail(err)
			}
		}
	}()
	return m, nil
}
func (m *displayFailureMonitor) close() {
	if m == nil {
		return
	}
	m.once.Do(func() { close(m.stop); m.retire(); <-m.done })
}
