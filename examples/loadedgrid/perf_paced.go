// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
package main

import (
	"cogentcore.org/core/system"
	"context"
	"errors"
	"fmt"
	"github.com/dshills/vellum/drivers/cogent"
	"time"
)

func (p *perfProgram) pacedScroll(ctx context.Context, m *cogent.FrameMetrics, dx, dy int, tag uint64, step *perfPaceStep) (cogent.FrameSample, cogent.GridInspection, error) {
	var empty cogent.FrameSample
	step.dispatch()
	err := p.owned(func() error { return cogent.EnqueueGridScroll(ctx, p.owner, p.grid, float64(dx), float64(dy), tag) })
	step.enqueued()
	if err != nil {
		return empty, cogent.GridInspection{}, err
	}
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		poll := m.Poll(tag)
		if poll.Failure != "" {
			return empty, cogent.GridInspection{}, errors.New(poll.Failure)
		}
		if poll.Dropped != 0 {
			return empty, cogent.GridInspection{}, errors.New("loadedgrid perf: frame recorder overflow")
		}
		if poll.Found {
			step.acknowledged()
			s := poll.Sample
			if !s.Completed || s.Failure != "" {
				return s, cogent.GridInspection{}, fmt.Errorf("loadedgrid perf: incomplete/refused frame: %s", s.Failure)
			}
			g, err := p.inspect(ctx)
			step.inspected()
			if err != nil {
				return s, g, err
			}
			if g.Generation != s.Generation {
				return s, g, errors.New("loadedgrid perf: measured frame generation mismatch")
			}
			return s, g, nil
		}
		select {
		case <-ctx.Done():
			return empty, cogent.GridInspection{}, ctx.Err()
		case <-ticker.C:
		}
	}
}

// Keep the serial consumer untouched. Only this opt-in branch omits per-event
// output/ReadMemStats. Validation, geometry hashing and owner probes are timed.
func (p *perfProgram) measurePaced(ctx context.Context, m *cogent.FrameMetrics, c perfConfig, capture *perfPresentationCapture, window perfWindowGeometry) error {
	plan, err := newPerfPacing(c.Samples, 1)
	if err != nil {
		return err
	}
	previous, err := p.inspect(ctx)
	if err != nil {
		return err
	}
	initialGeometry, err := checkPerfGeometry(previous)
	if err != nil {
		return err
	}
	plan.Initial = perfPaceObservation{Tag: (1 << 32) + uint64(c.Warmup), Generation: previous.Generation, Frame: previous.Frames, View: perfPaceView{ScrollX: previous.Viewport.ScrollX, ScrollY: previous.Viewport.ScrollY, Width: previous.Viewport.Width, Height: previous.Viewport.Height, Bounds: [4]int{previous.Bounds.Min.X, previous.Bounds.Min.Y, previous.Bounds.Max.X, previous.Bounds.Max.Y}, Scale: previous.Scale, Visible: initialGeometry.Cells, Rows: initialGeometry.Rows, Columns: initialGeometry.Columns, Retained: initialGeometry.RetainedCells, Materialized: initialGeometry.MaterializedCells, LayoutSHA256: initialGeometry.LayoutSHA256}}
	var baseCallbacks uint64
	if err = p.owned(func() error { baseCallbacks = p.callbacks; return nil }); err != nil {
		return err
	}
	runError := runPerfPacing(ctx, plan, realPerfPaceClock(), func(ctx context.Context, i int, step *perfPaceStep) (perfPaceObservation, error) {
		dx, dy := perfDelta(i)
		sample, g, err := p.pacedScroll(ctx, m, dx, dy, uint64(i+1), step)
		if err != nil {
			return perfPaceObservation{}, err
		}
		if err = verifyPerfCells(p.source, g); err != nil {
			return perfPaceObservation{}, err
		}
		geometry, err := checkPerfGeometry(g)
		if err != nil {
			return perfPaceObservation{}, err
		}
		if dx == 0 && dy == 0 || g.Generation != previous.Generation+1 || g.Viewport.ScrollX != previous.Viewport.ScrollX+dx || g.Viewport.ScrollY != previous.Viewport.ScrollY+dy || g.Viewport.Width != previous.Viewport.Width || g.Viewport.Height != previous.Viewport.Height || g.Bounds != previous.Bounds || g.Scale != previous.Scale || g.Frames <= previous.Frames {
			return perfPaceObservation{}, errors.New("paced grid: clamped/no-op or geometry/generation drift")
		}
		var count uint64
		if err = p.owned(func() error { count = p.callbacks; return nil }); err != nil {
			return perfPaceObservation{}, err
		}
		if count != baseCallbacks+uint64(i+1) {
			return perfPaceObservation{}, errors.New("paced grid: callback count differs")
		}
		previous = g
		return perfPaceObservation{Tag: sample.Tag, Generation: sample.Generation, Frame: g.Frames, View: perfPaceView{ScrollX: g.Viewport.ScrollX, ScrollY: g.Viewport.ScrollY, Width: g.Viewport.Width, Height: g.Viewport.Height, Bounds: [4]int{g.Bounds.Min.X, g.Bounds.Min.Y, g.Bounds.Max.X, g.Bounds.Max.Y}, Scale: g.Scale, Visible: geometry.Cells, Rows: geometry.Rows, Columns: geometry.Columns, Retained: geometry.RetainedCells, Materialized: geometry.MaterializedCells, LayoutSHA256: geometry.LayoutSHA256}}, nil
	})
	if err = capture.storePacing(plan); err != nil {
		return errors.Join(runError, err)
	}
	if runError != nil {
		return runError
	}
	final := m.Snapshot()
	if final.Pending != 0 || final.Dropped != 0 || final.Failure != "" || len(final.Samples) != c.Warmup+c.Samples {
		return errors.New("paced grid: terminal CPU cohort differs")
	}
	actualWindow, err := observePerfWindow(system.TheApp)
	if err != nil {
		return err
	}
	if actualWindow != window {
		return errors.New("paced grid: native window endpoints changed")
	}
	if err = capture.Seal(ctx); err != nil {
		return err
	}
	return c.emit("grid.perf.paced-finished", map[string]any{"samples": c.Samples, "hz": perfPaceHz, "planned_duration_ns": plan.PlannedEndNS, "observed_hold_end_ns": plan.HoldEndNS, "native_window": actualWindow, "full_mac_qualified": false, "schedule_and_native_times": "separate clock domains; see raw journal"})
}
