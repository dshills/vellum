// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dshills/vellum/drivers/cogent"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestPresentationCaptureExclusiveFileAndFailurePreservation(t *testing.T) {
	frames, _ := cogent.NewFrameMetrics(1)
	path := filepath.Join(t.TempDir(), "presentation.json")
	c, err := startPerfPresentation(path, frames, presentationCaptureConfig{Method: "routed-grid-scroll"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := startPerfPresentation(path, frames, presentationCaptureConfig{Method: "routed-grid-scroll"}); err == nil {
		t.Fatal("overwrote evidence")
	}
	if err := c.Seal(context.Background()); err == nil {
		t.Fatal("empty metrics claimed native feedback")
	}
	if err := c.Finish(errors.New("failed setup"), true, true); err == nil {
		t.Fatal("missing preclose native feedback accepted")
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Finish(nil, true, true); err == nil {
		t.Fatal("second finish cleared failure")
	}
	after, _ := os.ReadFile(path)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("finish rewrote prior evidence")
	}
	var record map[string]any
	if err := json.Unmarshal(before, &record); err != nil || record["Schema"] != "vellum-presentation-capture-v1" || record["Result"] != "FAILED_DIAGNOSTIC" || record["ContinuousWork"] != false || record["AfterClose"] == nil {
		t.Fatal(err, record)
	}
	if counts, ok := record["PrefixObservations"].(map[string]any); !ok || counts["Records"] != float64(0) || counts["Presented"] != float64(0) || counts["StartupSkipped"] != float64(0) {
		t.Fatal("failed capture omitted explicit prefix accounting", record)
	}
	if c.Seal(context.Background()) == nil {
		t.Fatal("seal after finish")
	}
}
func TestPresentationCaptureNoAfterCloseClaimBeforeWorkerRetirement(t *testing.T) {
	frames, _ := cogent.NewFrameMetrics(1)
	path := filepath.Join(t.TempDir(), "notjoined.json")
	c, err := startPerfPresentation(path, frames, presentationCaptureConfig{Method: "routed-tree-scroll"})
	if err != nil {
		t.Fatal(err)
	}
	if c.Finish(errors.New("worker did not retire"), false, true) == nil {
		t.Fatal("unjoined worker accepted")
	}
	raw, _ := os.ReadFile(path)
	var record map[string]any
	if err := json.Unmarshal(raw, &record); err != nil || record["AfterClose"] != nil || record["WorkerRetired"] != false {
		t.Fatal(err, record)
	}
}
func TestPresentationCaptureAdmissionDisabledSymlinkAndCancellation(t *testing.T) {
	frames, _ := cogent.NewFrameMetrics(1)
	disabled, err := startPerfPresentation("", frames, presentationCaptureConfig{})
	if err != nil || disabled != nil || disabled.Metrics() != nil || disabled.Seal(nil) != nil || disabled.Finish(nil, true, true) != nil {
		t.Fatal("disabled capture")
	}
	for _, path := range []string{"a\x00b", string(make([]byte, 4097))} {
		if _, err := startPerfPresentation(path, frames, presentationCaptureConfig{Method: "routed-grid-scroll"}); err == nil {
			t.Fatal("invalid path")
		}
	}
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink("missing-target", link); err != nil {
		t.Fatal(err)
	}
	if _, err := startPerfPresentation(link, frames, presentationCaptureConfig{Method: "routed-grid-scroll"}); err == nil {
		t.Fatal("followed dangling evidence symlink")
	}
	c, err := startPerfPresentation(filepath.Join(t.TempDir(), "canceled.json"), frames, presentationCaptureConfig{Method: "routed-grid-scroll"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if !errors.Is(c.Seal(ctx), context.Canceled) {
		t.Fatal("canceled seal")
	}
	if c.Finish(context.Canceled, true, true) == nil {
		t.Fatal("cancellation suppressed")
	}
}
func TestPresentationPrefixLossCannotBeFilteredByDraining(t *testing.T) {
	p := presentationCapturePoint{Frames: cogent.FrameMetricsSnapshot{Samples: []cogent.FrameSample{{Tag: 1}}}, Presentation: cogent.PresentationSnapshot{Scopes: []cogent.PresentationScope{{ID: 1, Window: "perf"}}, Inputs: []cogent.PresentationInput{{Tag: 1}}, Samples: []cogent.PresentationSample{{Sequence: 1, State: "presented"}, {Sequence: 2, State: "pending"}}}}
	if err := presentationPointError(p, 1); err != nil {
		t.Fatal("explicit post-cutoff tail altered prefix", err)
	}
	for _, state := range []string{"pending", "retired", "skipped", "unavailable", "invalid"} {
		p.Presentation.Samples[0].State = state
		if presentationPointError(p, 1) == nil {
			t.Fatal("lost prefix accepted", state)
		}
	}
	p.Presentation.Samples[0].State = "presented"
	p.Presentation.Failed = true
	if presentationPointError(p, 1) == nil {
		t.Fatal("global recorder failure")
	}
	writer := &presentationCaptureWriter{target: io.Discard, remaining: 2}
	if n, err := writer.Write([]byte("abc")); err == nil || n != 0 {
		t.Fatal("byte budget")
	}
}
func TestPresentationWorkerJoinRequiresTerminalAck(t *testing.T) {
	done := make(chan struct{})
	closed := false
	cancel := func() { closed = true; close(done) }
	if !joinPerfPresentationWorker(done, cancel) || !closed {
		t.Fatal("cancellation/terminal ack")
	}
	if !joinPerfPresentationWorker(done, func() { t.Fatal("already retired worker canceled") }) {
		t.Fatal("retired worker")
	}
}

func TestPresentationExecutableHashIsBoundedAndStable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "binary")
	if err := os.WriteFile(path, []byte("abc"), 0600); err != nil {
		t.Fatal(err)
	}
	sum, err := hashPresentationFile(path)
	if err != nil || sum != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatal(sum, err)
	}
	if _, err := hashPresentationFile(filepath.Dir(path)); err == nil {
		t.Fatal("directory accepted")
	}
	if err := os.Truncate(path, (512<<20)+1); err != nil {
		t.Fatal(err)
	}
	if _, err := hashPresentationFile(path); err == nil {
		t.Fatal("executable read budget")
	}
}

func TestPresentationRecordsCannotMasqueradeAsOrdinary(t *testing.T) {
	original := map[string]any{"event": "sample", "old_value": 1}
	if got := markPresentationRecord(map[string]any{"event": "sample", "old_value": 1}, false); !reflect.DeepEqual(got, original) {
		t.Fatal("ordinary schema changed", got)
	}
	got := markPresentationRecord(map[string]any{"presentation_capture": false, "cadence_qualified": true}, true)
	if got["presentation_capture"] != true || got["cadence_qualified"] != false {
		t.Fatal("diagnostic labels missing", got)
	}
}

func TestPresentationCaptureSetupFailureHasNoAfterCloseClaim(t *testing.T) {
	frames, _ := cogent.NewFrameMetrics(1)
	path := filepath.Join(t.TempDir(), "setup-failed.json")
	c, err := startPerfPresentation(path, frames, presentationCaptureConfig{Method: "routed-grid-scroll"})
	if err != nil {
		t.Fatal(err)
	}
	if c.Finish(errors.New("setup output failed"), true, false) == nil {
		t.Fatal("pre-run failure accepted")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if err := json.Unmarshal(raw, &record); err != nil {
		t.Fatal(err)
	}
	if record["Result"] != "FAILED_DIAGNOSTIC" || record["AfterClose"] != nil || record["WorkerRetired"] != true || record["RuntimeReturned"] != false {
		t.Fatal("setup failure manufactured a runtime/teardown claim", record)
	}
}

func TestPresentationFrozenPrefixRequiresEntireCompletedCohort(t *testing.T) {
	makePoint := func() presentationCapturePoint {
		return presentationCapturePoint{
			Frames: cogent.FrameMetricsSnapshot{Samples: []cogent.FrameSample{
				{Tag: 1, Window: "perf", Generation: 9, Completed: true},
				{Tag: 2, Window: "perf", Generation: 10, Completed: true},
			}},
			Presentation: cogent.PresentationSnapshot{
				Inputs: []cogent.PresentationInput{{Scope: 1, Tag: 1}, {Scope: 1, Tag: 2}},
				Samples: []cogent.PresentationSample{
					{Scope: 1, Frame: cogent.PresentationFrame{Tag: 1, Generation: 8, Armed: false}},
					{Scope: 1, Frame: cogent.PresentationFrame{Tag: 1, Generation: 9, Armed: true}, State: "presented"},
					{Scope: 1, Frame: cogent.PresentationFrame{Tag: 2, Generation: 10, Armed: true}, State: "pending"},
				},
			},
		}
	}
	p := makePoint()
	// The last callback may still be pending: Seal drains this exact prefix.
	if err := presentationCohortError(p, 3); err != nil {
		t.Fatal(err)
	}
	// A later final acquisition cannot repair a prefix that already froze.
	if presentationCohortError(p, 2) == nil {
		t.Fatal("cutoff moved to admit missing final frame")
	}
	for _, kind := range []string{"missing-final", "wrong-generation", "duplicate-final", "extra-final", "missing-receipt", "duplicate-receipt", "foreign-scope", "incomplete-cpu", "duplicate-cpu"} {
		t.Run(kind, func(t *testing.T) {
			p := makePoint()
			switch kind {
			case "missing-final":
				p.Presentation.Samples[2].Armed = false
			case "wrong-generation":
				p.Presentation.Samples[2].Generation++
			case "duplicate-final":
				p.Presentation.Samples = append(p.Presentation.Samples, p.Presentation.Samples[2])
			case "extra-final":
				p.Presentation.Samples[2].Tag = 3
			case "missing-receipt":
				p.Presentation.Inputs = p.Presentation.Inputs[:1]
			case "duplicate-receipt":
				p.Presentation.Inputs[1] = p.Presentation.Inputs[0]
			case "foreign-scope":
				p.Presentation.Samples[2].Scope = 2
			case "incomplete-cpu":
				p.Frames.Samples[1].Completed = false
			case "duplicate-cpu":
				p.Frames.Samples[1] = p.Frames.Samples[0]
			}
			if presentationCohortError(p, uint64(len(p.Presentation.Samples))) == nil {
				t.Fatal("incomplete final acquisition cohort accepted")
			}
		})
	}
}

func TestPresentationStartupSkipsRemainExplicitAndOutsideTiming(t *testing.T) {
	makePoint := func() presentationCapturePoint {
		return presentationCapturePoint{
			Frames: cogent.FrameMetricsSnapshot{Samples: []cogent.FrameSample{{Tag: 1, Window: "perf", Generation: 9, Completed: true}}},
			Presentation: cogent.PresentationSnapshot{
				Scopes: []cogent.PresentationScope{{ID: 1, Window: "perf"}},
				Inputs: []cogent.PresentationInput{{Scope: 1, Tag: 1, Received: 10}},
				Samples: []cogent.PresentationSample{
					{Sequence: 1, Scope: 1, Frame: cogent.PresentationFrame{Composition: 1}, Acquired: 9, State: "skipped"},
					{Sequence: 2, Scope: 1, Frame: cogent.PresentationFrame{Tag: 1, Generation: 9, Composition: 2, Armed: true}, Acquired: 11, Presented: 12, State: "presented"},
				},
			},
		}
	}
	p := makePoint()
	before, _ := json.Marshal(p)
	if err := presentationPointError(p, 2); err != nil {
		t.Fatal("actual pre-input terminal skipped observation refused", err)
	}
	if err := presentationCohortError(p, 2); err != nil {
		t.Fatal("startup observation changed final acquisition cohort", err)
	}
	if counts := presentationPrefixCounts(p, 2); counts != (presentationPrefixAccounting{Records: 2, Presented: 1, StartupSkipped: 1}) {
		t.Fatal("skipped startup counted as presented or omitted", counts)
	}
	after, _ := json.Marshal(p)
	if !reflect.DeepEqual(before, after) || p.Presentation.Samples[0].Presented != 0 || p.Presentation.Samples[0].State != "skipped" {
		t.Fatal("startup classification rewrote evidence")
	}
	for _, kind := range []string{"at-first-receipt", "after-first-receipt", "armed", "tagged", "generation", "foreign-scope", "nonzero-presented", "missing-receipt", "foreign-receipt", "zero-receipt", "nonfinite-receipt", "regressed-receipt", "zero-acquired", "nan-acquired", "infinite-acquired", "pending", "unavailable", "invalid", "retired", "missing-final"} {
		t.Run(kind, func(t *testing.T) {
			p := makePoint()
			sample := &p.Presentation.Samples[0]
			switch kind {
			case "at-first-receipt":
				sample.Acquired = 10
			case "after-first-receipt":
				sample.Acquired = 10.5
			case "armed":
				sample.Armed = true
			case "tagged":
				sample.Tag = 1
			case "generation":
				sample.Generation = 8
			case "foreign-scope":
				sample.Scope = 2
			case "nonzero-presented":
				sample.Presented = 9.5
			case "missing-receipt":
				p.Presentation.Inputs = nil
			case "foreign-receipt":
				p.Presentation.Inputs[0].Scope = 2
			case "zero-receipt":
				p.Presentation.Inputs[0].Received = 0
			case "nonfinite-receipt":
				p.Presentation.Inputs[0].Received = math.Inf(1)
			case "regressed-receipt":
				p.Presentation.Inputs = append(p.Presentation.Inputs, cogent.PresentationInput{Scope: 1, Tag: 2, Received: 9.5})
			case "zero-acquired":
				sample.Acquired = 0
			case "nan-acquired":
				sample.Acquired = math.NaN()
			case "infinite-acquired":
				sample.Acquired = math.Inf(1)
			case "pending", "unavailable", "invalid", "retired":
				sample.State = kind
			case "missing-final":
				p.Presentation.Samples[1].Armed = false
			}
			if kind == "missing-final" {
				if presentationCohortError(p, 2) == nil {
					t.Fatal("startup exemption repaired missing final acquisition")
				}
			} else if presentationPointError(p, 2) == nil {
				t.Fatal("invalid prefix loss received startup exemption")
			}
		})
	}
}
