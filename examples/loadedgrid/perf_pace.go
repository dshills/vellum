// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
package main

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// Identical in the independently consumable grid/tree examples. These clocks
// describe scheduling and observation, never native presentation timestamps.
const perfPaceHz = 60
const perfPaceMaxSlots = 2048
const perfPaceMaxEvidenceBytes = 2 << 20

type perfPaceView struct {
	ScrollX, ScrollY, Width, Height                int
	Bounds                                         [4]int
	Scale                                          float32
	Visible, Rows, Columns, Retained, Materialized int
	LayoutSHA256                                   string
}
type perfPaceObservation struct {
	Tag, Generation, Frame uint64
	View                   perfPaceView
}
type perfPaceSlot struct {
	Index                                                               int
	Tag                                                                 uint64
	ScheduledNS                                                         int64
	State                                                               string
	StartNS, DispatchNS, EnqueuedNS, AcknowledgedNS, InspectedNS, EndNS int64
	Observation                                                         perfPaceObservation
}
type perfPacingJournal struct {
	Hz                     int
	Clock, Method          string
	PlannedEndNS           int64
	Started, Completed     bool
	HoldStartNS, HoldEndNS int64
	Failure                string
	Initial                perfPaceObservation
	Slots                  []perfPaceSlot
}
type perfPaceClock struct {
	now  func() time.Time
	wait func(context.Context, time.Time) error
}

func perfPaceOffset(index int) int64 { return int64(index) * int64(time.Second) / perfPaceHz }
func newPerfPacing(count int, firstTag uint64) (*perfPacingJournal, error) {
	if count < 1 || count > perfPaceMaxSlots || firstTag == 0 || firstTag > ^uint64(0)-uint64(count) {
		return nil, errors.New("paced workload: slot/tag bound")
	}
	result := &perfPacingJournal{Hz: perfPaceHz, Clock: "Go monotonic relative nanoseconds; separate from native host clock", Method: "absolute rational deadlines; serial CPU acknowledgement; retained backlog", PlannedEndNS: perfPaceOffset(count), Slots: make([]perfPaceSlot, count)}
	for i := range result.Slots {
		result.Slots[i] = perfPaceSlot{Index: i, Tag: firstTag + uint64(i), ScheduledNS: perfPaceOffset(i), State: "planned"}
	}
	return result, nil
}
func realPerfPaceClock() perfPaceClock {
	return perfPaceClock{now: time.Now, wait: func(ctx context.Context, target time.Time) error {
		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			remaining := time.Until(target)
			if remaining <= 0 {
				return nil
			}
			timer := time.NewTimer(remaining)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
	}}
}

// A step is ephemeral: journals retain only scalar timestamps and detached view
// summaries. The existing 1ms Poll/owner inspection costs stay inside the cohort.
type perfPaceStep struct {
	slot    *perfPaceSlot
	origin  time.Time
	clock   perfPaceClock
	stage   int
	failure error
}

func (s *perfPaceStep) mark(stage int) {
	if s.failure != nil {
		return
	}
	if stage != s.stage+1 {
		s.failure = errors.New("paced workload: observation stage order")
		return
	}
	value := s.clock.now().Sub(s.origin).Nanoseconds()
	var previous int64
	switch stage {
	case 1:
		previous = s.slot.StartNS
		s.slot.DispatchNS = value
	case 2:
		previous = s.slot.DispatchNS
		s.slot.EnqueuedNS = value
	case 3:
		previous = s.slot.EnqueuedNS
		s.slot.AcknowledgedNS = value
	case 4:
		previous = s.slot.AcknowledgedNS
		s.slot.InspectedNS = value
	default:
		s.failure = errors.New("paced workload: observation stage bound")
		return
	}
	if value < previous {
		s.failure = errors.New("paced workload: observation clock regressed")
		return
	}
	s.stage = stage
}
func (s *perfPaceStep) dispatch()     { s.mark(1) }
func (s *perfPaceStep) enqueued()     { s.mark(2) }
func (s *perfPaceStep) acknowledged() { s.mark(3) }
func (s *perfPaceStep) inspected()    { s.mark(4) }

func runPerfPacing(ctx context.Context, journal *perfPacingJournal, clock perfPaceClock, work func(context.Context, int, *perfPaceStep) (perfPaceObservation, error)) (terminal error) {
	if ctx == nil || journal == nil || len(journal.Slots) == 0 || journal.Started || clock.now == nil || clock.wait == nil || work == nil {
		return errors.New("paced workload: invalid run")
	}
	// The count-bounded plan already exists before the monotonic origin.
	expected, err := newPerfPacing(len(journal.Slots), journal.Slots[0].Tag)
	if err != nil || journal.PlannedEndNS != expected.PlannedEndNS {
		return errors.New("paced workload: plan differs")
	}
	for i := range journal.Slots {
		if journal.Slots[i] != expected.Slots[i] {
			return errors.New("paced workload: plan differs")
		}
	}
	defer func() {
		if terminal != nil {
			journal.Failure = terminal.Error()
			if len(journal.Failure) > 4096 {
				journal.Failure = journal.Failure[:4096]
			}
		}
	}()
	bounded, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	origin := clock.now()
	journal.Started = true
	var previousEnd int64
	for i := range journal.Slots {
		slot := &journal.Slots[i]
		if err := clock.wait(bounded, origin.Add(time.Duration(slot.ScheduledNS))); err != nil {
			return err
		}
		if err := bounded.Err(); err != nil {
			return err
		}
		slot.StartNS = clock.now().Sub(origin).Nanoseconds()
		if slot.StartNS < slot.ScheduledNS || slot.StartNS < previousEnd {
			return errors.New("paced workload: early or regressed start")
		}
		slot.State = "running"
		step := &perfPaceStep{slot: slot, origin: origin, clock: clock}
		observed, err := work(bounded, i, step)
		slot.Observation = observed
		slot.EndNS = clock.now().Sub(origin).Nanoseconds()
		previousEnd = slot.EndNS
		if err == nil {
			err = step.failure
		}
		if err == nil && (step.stage != 4 || slot.EndNS < slot.InspectedNS || observed.Tag != slot.Tag || observed.Generation == 0) {
			err = errors.New("paced workload: incomplete observation")
		}
		if err == nil {
			err = bounded.Err()
		}
		if err != nil {
			slot.State = "failed"
			return err
		}
		slot.State = "completed"
	}
	// The planned right edge never moves. A held final state is idle tail, not an
	// additional changing frame or a manufactured presentation interval.
	journal.HoldStartNS = clock.now().Sub(origin).Nanoseconds()
	if journal.HoldStartNS < previousEnd {
		return errors.New("paced workload: hold clock regressed")
	}
	if err := clock.wait(bounded, origin.Add(time.Duration(journal.PlannedEndNS))); err != nil {
		return err
	}
	journal.HoldEndNS = clock.now().Sub(origin).Nanoseconds()
	if journal.HoldEndNS < journal.PlannedEndNS || journal.HoldEndNS < journal.HoldStartNS {
		return errors.New("paced workload: hold clock regressed")
	}
	journal.Completed = true
	return nil
}

func (c *perfPresentationCapture) storePacing(journal *perfPacingJournal) error {
	if c == nil || journal == nil {
		return errors.New("paced workload: capture required")
	}
	if len(journal.Slots) < 1 || len(journal.Slots) > perfPaceMaxSlots {
		return errors.New("paced workload: evidence count bound")
	}
	// Copy/serialize outside the scheduled cohort. No live model/native references.
	copy := *journal
	copy.Slots = append([]perfPaceSlot(nil), journal.Slots...)
	raw, err := json.Marshal(copy)
	if err != nil {
		return err
	}
	if len(copy.Slots) > perfPaceMaxSlots || len(raw) > perfPaceMaxEvidenceBytes {
		return errors.New("paced workload: evidence byte/count bound")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.finished || c.sealed {
		return errors.New("paced workload: capture already sealed")
	}
	c.pacing = &copy
	return nil
}
