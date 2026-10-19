// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakePerfPaceClock struct {
	at    time.Time
	waits []time.Time
	early bool
}

func (f *fakePerfPaceClock) clock() perfPaceClock {
	return perfPaceClock{now: func() time.Time { return f.at }, wait: func(ctx context.Context, target time.Time) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		f.waits = append(f.waits, target)
		if f.at.Before(target) && !f.early {
			f.at = target
		}
		return nil
	}}
}
func fakePaceWork(f *fakePerfPaceClock, duration time.Duration) func(context.Context, int, *perfPaceStep) (perfPaceObservation, error) {
	return func(ctx context.Context, i int, s *perfPaceStep) (perfPaceObservation, error) {
		s.dispatch()
		f.at = f.at.Add(duration / 4)
		s.enqueued()
		f.at = f.at.Add(duration / 4)
		s.acknowledged()
		f.at = f.at.Add(duration / 4)
		s.inspected()
		f.at = f.at.Add(duration / 4)
		return perfPaceObservation{Tag: s.slot.Tag, Generation: uint64(i + 1), Frame: uint64(i + 1)}, nil
	}
}
func TestPaceAbsoluteRationalDeadlinesAndIdleTail(t *testing.T) {
	plan, err := newPerfPacing(61, 100)
	if err != nil {
		t.Fatal(err)
	}
	f := &fakePerfPaceClock{at: time.Unix(1000, 0)}
	origin := f.at
	if err = runPerfPacing(context.Background(), plan, f.clock(), fakePaceWork(f, 4*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	if !plan.Completed || len(plan.Slots) != 61 || len(f.waits) != 62 {
		t.Fatal(plan)
	}
	if plan.Slots[60].ScheduledNS != int64(time.Second) || plan.Slots[60].StartNS != int64(time.Second) {
		t.Fatal("period rounding drift", plan.Slots[60])
	}
	for i, s := range plan.Slots {
		if s.Tag != 100+uint64(i) || s.State != "completed" || s.ScheduledNS != int64(i)*1e9/60 || !f.waits[i].Equal(origin.Add(time.Duration(s.ScheduledNS))) {
			t.Fatal("rebased/omitted slot", i, s)
		}
	}
	if plan.HoldStartNS != int64(time.Second+4*time.Millisecond) || plan.HoldEndNS != plan.PlannedEndNS {
		t.Fatal("held tail changed planned duration")
	}
}
func TestPaceOverloadRetainsEveryOriginalSlot(t *testing.T) {
	plan, _ := newPerfPacing(4, 1)
	f := &fakePerfPaceClock{at: time.Unix(1000, 0)}
	if err := runPerfPacing(context.Background(), plan, f.clock(), fakePaceWork(f, 25*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	for i, s := range plan.Slots {
		if s.StartNS != int64(i)*int64(25*time.Millisecond) || s.ScheduledNS != int64(i)*1e9/60 || s.State != "completed" {
			t.Fatal("overdue slot dropped/rebased", i, s)
		}
	}
	if plan.HoldEndNS != int64(100*time.Millisecond) || plan.PlannedEndNS != 4*int64(time.Second)/60 {
		t.Fatal("overrun hidden", plan)
	}
}
func TestPaceCancellationAndFailureRetainUnstartedSlots(t *testing.T) {
	for _, kind := range []string{"cancel", "work", "missing-ack"} {
		t.Run(kind, func(t *testing.T) {
			plan, _ := newPerfPacing(4, 1)
			f := &fakePerfPaceClock{at: time.Unix(1000, 0)}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ordinary := fakePaceWork(f, time.Millisecond)
			err := runPerfPacing(ctx, plan, f.clock(), func(ctx context.Context, i int, s *perfPaceStep) (perfPaceObservation, error) {
				if i != 1 {
					return ordinary(ctx, i, s)
				}
				if kind == "missing-ack" {
					return perfPaceObservation{Tag: s.slot.Tag, Generation: 2}, nil
				}
				out, err := ordinary(ctx, i, s)
				if kind == "cancel" {
					cancel()
				} else {
					err = errors.New("work failed")
				}
				return out, err
			})
			if err == nil || plan.Completed || plan.Failure == "" || plan.Slots[0].State != "completed" || plan.Slots[1].State != "failed" || plan.Slots[2].State != "planned" || plan.Slots[3].State != "planned" {
				t.Fatal("failed prefix masqueraded as complete", plan, err)
			}
		})
	}
}
func TestPaceRefusesBoundsEarlyClockAndReentry(t *testing.T) {
	for _, n := range []int{-1, 0, 2049} {
		if _, err := newPerfPacing(n, 1); err == nil {
			t.Fatal("slot bound", n)
		}
	}
	if _, err := newPerfPacing(1, 0); err == nil {
		t.Fatal("zero tag")
	}
	if _, err := newPerfPacing(2, ^uint64(0)); err == nil {
		t.Fatal("tag overflow")
	}
	f := &fakePerfPaceClock{at: time.Unix(1000, 0), early: true}
	plan, _ := newPerfPacing(2, 1)
	if runPerfPacing(context.Background(), plan, f.clock(), fakePaceWork(f, time.Millisecond)) == nil {
		t.Fatal("early wake accepted")
	}
	if runPerfPacing(context.Background(), plan, f.clock(), fakePaceWork(f, time.Millisecond)) == nil {
		t.Fatal("reentry accepted")
	}
	if runPerfPacing(context.Background(), &perfPacingJournal{}, f.clock(), fakePaceWork(f, time.Millisecond)) == nil {
		t.Fatal("empty plan accepted")
	}
}
func TestPaceStoredEvidenceDetachedBoundedAndRevoked(t *testing.T) {
	capture := &perfPresentationCapture{}
	plan, _ := newPerfPacing(2, 1)
	if err := capture.storePacing(plan); err != nil {
		t.Fatal(err)
	}
	plan.Slots[0].State = "changed"
	if capture.pacing.Slots[0].State != "planned" {
		t.Fatal("mutable evidence retained")
	}
	plan.Clock = strings.Repeat("x", perfPaceMaxEvidenceBytes)
	if capture.storePacing(plan) == nil {
		t.Fatal("byte bound")
	}
	plan, _ = newPerfPacing(2, 1)
	capture.finished = true
	if capture.storePacing(plan) == nil {
		t.Fatal("write after finish")
	}
}
