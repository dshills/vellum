// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/dshills/vellum/drivers/cogent"
	"time"
)

func (p *treePerfProgram) pacedScroll(ctx context.Context, m *cogent.FrameMetrics, dy int, tag uint64, step *perfPaceStep) (cogent.FrameSample, cogent.TreeInspection, error) {
	var empty cogent.FrameSample
	before, err := p.inspect(ctx)
	if err != nil {
		return empty, before, err
	}
	step.dispatch()
	err = p.owned(ctx, func() error { return cogent.EnqueueTreeScroll(ctx, p.owner, p.outline, float64(dy), tag) })
	step.enqueued()
	if err != nil {
		return empty, cogent.TreeInspection{}, err
	}
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		poll := m.Poll(tag)
		if poll.Failure != "" || poll.Dropped != 0 {
			return empty, cogent.TreeInspection{}, errors.New("loadedtree perf: invalid or overflowing recorder")
		}
		if poll.Found {
			step.acknowledged()
			sample := poll.Sample
			if !sample.Completed || sample.Failure != "" {
				return sample, cogent.TreeInspection{}, fmt.Errorf("loadedtree perf: refused frame: %s", sample.Failure)
			}
			g, err := p.inspect(ctx)
			step.inspected()
			if err == nil && (g.Generation != sample.Generation || g.Generation != before.Generation+1 || g.Viewport.Width != before.Viewport.Width || g.Viewport.Height != before.Viewport.Height || g.Viewport.ScrollY != before.Viewport.ScrollY+dy || g.Bounds != before.Bounds || g.Scale != before.Scale || g.Stamp != before.Stamp) {
				err = errors.New("loadedtree perf: event offset/generation/geometry drift")
			}
			return sample, g, err
		}
		select {
		case <-ctx.Done():
			return empty, cogent.TreeInspection{}, ctx.Err()
		case <-ticker.C:
		}
	}
}

func (p *treePerfProgram) measurePaced(ctx context.Context, m *cogent.FrameMetrics, c treePerfConfig, capture *perfPresentationCapture) error {
	firstTag := uint64(c.Warmup + 3)
	plan, err := newPerfPacing(c.Samples, firstTag)
	if err != nil {
		return err
	}
	previous, err := p.inspect(ctx)
	if err != nil {
		return err
	}
	initialGeometry, err := treePerfGeometryCheck(previous)
	if err != nil {
		return err
	}
	plan.Initial = perfPaceObservation{Tag: uint64(c.Warmup + 2), Generation: previous.Generation, Frame: previous.Frames, View: perfPaceView{ScrollY: previous.Viewport.ScrollY, Width: previous.Viewport.Width, Height: previous.Viewport.Height, Bounds: [4]int{previous.Bounds.Min.X, previous.Bounds.Min.Y, previous.Bounds.Max.X, previous.Bounds.Max.Y}, Scale: previous.Scale, Visible: initialGeometry.ClippedRows, Rows: initialGeometry.ClippedRows, Columns: 1, Retained: initialGeometry.RetainedRows, Materialized: initialGeometry.MaterializedRows, LayoutSHA256: initialGeometry.LayoutSHA256}}
	baseCallbacks, err := p.callbackCount(ctx)
	if err != nil {
		return err
	}
	runError := runPerfPacing(ctx, plan, realPerfPaceClock(), func(ctx context.Context, i int, step *perfPaceStep) (perfPaceObservation, error) {
		dy := treePerfDelta(i)
		sample, g, err := p.pacedScroll(ctx, m, dy, firstTag+uint64(i), step)
		if err != nil {
			return perfPaceObservation{}, err
		}
		geometry, err := treePerfGeometryCheck(g)
		if err != nil {
			return perfPaceObservation{}, err
		}
		if dy == 0 || g.Generation != previous.Generation+1 || g.Viewport.ScrollY != previous.Viewport.ScrollY+dy || g.Viewport.Width != previous.Viewport.Width || g.Viewport.Height != previous.Viewport.Height || g.Bounds != previous.Bounds || g.Scale != previous.Scale || g.Frames <= previous.Frames {
			return perfPaceObservation{}, errors.New("paced tree: clamped/no-op or geometry/generation drift")
		}
		count, err := p.callbackCount(ctx)
		if err != nil {
			return perfPaceObservation{}, err
		}
		if count != baseCallbacks+uint64(i+1) {
			return perfPaceObservation{}, errors.New("paced tree: callback count differs")
		}
		previous = g
		return perfPaceObservation{Tag: sample.Tag, Generation: sample.Generation, Frame: g.Frames, View: perfPaceView{ScrollY: g.Viewport.ScrollY, Width: g.Viewport.Width, Height: g.Viewport.Height, Bounds: [4]int{g.Bounds.Min.X, g.Bounds.Min.Y, g.Bounds.Max.X, g.Bounds.Max.Y}, Scale: g.Scale, Visible: geometry.ClippedRows, Rows: geometry.ClippedRows, Columns: 1, Retained: geometry.RetainedRows, Materialized: geometry.MaterializedRows, LayoutSHA256: geometry.LayoutSHA256}}, nil
	})
	if err = capture.storePacing(plan); err != nil {
		return errors.Join(runError, err)
	}
	if runError != nil {
		return runError
	}
	final := m.Snapshot()
	if final.Pending != 0 || final.Dropped != 0 || final.Failure != "" || len(final.Samples) != c.Warmup+c.Samples+2 {
		return errors.New("paced tree: terminal CPU cohort differs")
	}
	if err = capture.Seal(ctx); err != nil {
		return err
	}
	return c.emit("tree.perf.paced-finished", map[string]any{"samples": c.Samples, "hz": perfPaceHz, "planned_duration_ns": plan.PlannedEndNS, "observed_hold_end_ns": plan.HoldEndNS, "full_mac_qualified": false, "schedule_and_native_times": "separate clock domains; see raw journal"})
}
