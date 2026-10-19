// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/dshills/vellum/drivers/cogent"
)

// Kept byte-identical in both standalone stored-data consumers. This diagnostic
// exports actual native feedback; the serial workload does not qualify cadence.
const presentationCaptureMaxBytes = 64 << 20

type presentationCapturePoint struct {
	Frames       cogent.FrameMetricsSnapshot
	Presentation cogent.PresentationSnapshot
}
type presentationCaptureConfig struct {
	Method                              string
	Samples, Warmup, DataSize           int
	Variable, Deep, Novel, GPURequested bool
	Paced                               bool `json:",omitempty"`
}
type perfPresentationCapture struct {
	metrics                *cogent.PresentationMetrics
	frames                 *cogent.FrameMetrics
	file                   *os.File
	config                 presentationCaptureConfig
	build                  map[string]string
	binarySHA256           string
	cutoff                 uint64
	receiptCutoff          int
	mu                     sync.Mutex
	before                 *presentationCapturePoint
	sealed, finished       bool
	sealError, finishError error
	pacing                 *perfPacingJournal
}

func validPresentationPath(path string) bool {
	return len(path) <= 4096 && !strings.ContainsRune(path, 0)
}
func startPerfPresentation(path string, frames *cogent.FrameMetrics, config presentationCaptureConfig) (*perfPresentationCapture, error) {
	if path == "" {
		return nil, nil
	}
	if !validPresentationPath(path) || frames == nil || config.Method != "routed-grid-scroll" && config.Method != "routed-tree-scroll" {
		return nil, errors.New("presentation capture: invalid configuration")
	}
	metrics, err := cogent.NewPresentationMetrics(65536)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	executable, err := os.Executable()
	if err != nil {
		return nil, errors.Join(err, file.Close())
	}
	binarySHA256, err := hashPresentationFile(executable)
	if err != nil {
		return nil, errors.Join(err, file.Close())
	}
	build := map[string]string{}
	if info, ok := debug.ReadBuildInfo(); ok {
		build["go_version"] = info.GoVersion
		for _, setting := range info.Settings {
			build[setting.Key] = setting.Value
		}
	}
	return &perfPresentationCapture{metrics: metrics, frames: frames, file: file, config: config, build: build, binarySHA256: binarySHA256}, nil
}
func (c *perfPresentationCapture) Metrics() *cogent.PresentationMetrics {
	if c == nil {
		return nil
	}
	return c.metrics
}
func (c *perfPresentationCapture) point() presentationCapturePoint {
	return presentationCapturePoint{Frames: c.frames.Snapshot(), Presentation: c.metrics.Snapshot()}
}
func presentationPointError(p presentationCapturePoint, cutoff uint64) error {
	s := p.Presentation
	f := p.Frames
	if s.Failed || s.Overflow || f.Failure != "" || f.Dropped != 0 {
		return errors.New("presentation capture: recorder failure; retain journal for diagnosis")
	}
	if len(s.Scopes) != 1 || s.Scopes[0].Window != "perf" || s.Scopes[0].Closed {
		return errors.New("presentation capture: missing or closed native scope")
	}
	if f.Pending != 0 {
		return errors.New("presentation capture: pending CPU frame")
	}
	if cutoff == 0 || cutoff > uint64(len(s.Samples)) || len(s.Inputs) == 0 || len(f.Samples) == 0 {
		return errors.New("presentation capture: no native feedback or traced input")
	}
	for _, sample := range s.Samples[:cutoff] {
		if sample.State == "pending" {
			return errors.New("presentation capture: pending prefix feedback")
		}
		if sample.State != "presented" {
			return errors.New("presentation capture: lost prefix feedback")
		}
	}
	return nil
}

// Check the completed CPU cohort against the frozen acquisition prefix. Missing
// final acquisitions fail capture; never move the cutoff to hide missing work.
func presentationCohortError(p presentationCapturePoint, cutoff uint64) error {
	if cutoff == 0 || cutoff > uint64(len(p.Presentation.Samples)) || len(p.Frames.Samples) == 0 {
		return errors.New("presentation capture: no completed acquisition cohort")
	}
	expected := make(map[uint64]uint64, len(p.Frames.Samples))
	for _, frame := range p.Frames.Samples {
		if frame.Tag == 0 || frame.Window != "perf" || frame.Generation == 0 || !frame.Completed || frame.Failure != "" {
			return errors.New("presentation capture: incomplete CPU cohort")
		}
		if _, duplicate := expected[frame.Tag]; duplicate {
			return errors.New("presentation capture: duplicate CPU tag")
		}
		expected[frame.Tag] = frame.Generation
	}
	receipts := make(map[uint64]bool, len(p.Presentation.Inputs))
	for _, input := range p.Presentation.Inputs {
		if _, ok := expected[input.Tag]; !ok || input.Scope != 1 || receipts[input.Tag] {
			return errors.New("presentation capture: input cohort differs")
		}
		receipts[input.Tag] = true
	}
	if len(receipts) != len(expected) {
		return errors.New("presentation capture: missing input receipt")
	}
	armed := make(map[uint64]bool, len(expected))
	for _, sample := range p.Presentation.Samples[:cutoff] {
		if !sample.Armed {
			continue
		}
		generation, ok := expected[sample.Tag]
		if !ok || sample.Scope != 1 || sample.Generation != generation || armed[sample.Tag] {
			return errors.New("presentation capture: final armed acquisition differs")
		}
		armed[sample.Tag] = true
	}
	if len(armed) != len(expected) {
		return errors.New("presentation capture: missing final armed acquisition")
	}
	return nil
}

// Seal occurs outside measured events and before window teardown. It never
// claims that a Present call or retirement was an actual presentation.
func (c *perfPresentationCapture) Seal(ctx context.Context) error {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.finished {
		return errors.New("presentation capture: already finished")
	}
	if c.sealed {
		return c.sealError
	}
	c.sealed = true
	if ctx == nil {
		c.sealError = errors.New("presentation capture: nil context")
		return c.sealError
	}
	drain, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	ticker := time.NewTicker(2 * time.Millisecond)
	defer ticker.Stop()
	if err := ctx.Err(); err != nil {
		c.sealError = err
		return err
	}
	initial := c.point()
	c.cutoff = uint64(len(initial.Presentation.Samples))
	c.receiptCutoff = len(initial.Presentation.Inputs)
	c.before = &initial
	if err := presentationCohortError(initial, c.cutoff); err != nil {
		c.sealError = err
		return err
	}
	for {
		point := c.point()
		c.before = &point
		pending := false
		if c.cutoff > 0 && c.cutoff <= uint64(len(point.Presentation.Samples)) {
			for _, sample := range point.Presentation.Samples[:c.cutoff] {
				if sample.State == "pending" {
					pending = true
				}
			}
		}
		if !pending || point.Presentation.Failed || point.Presentation.Overflow || point.Frames.Pending != 0 || point.Frames.Failure != "" || point.Frames.Dropped != 0 {
			c.sealError = presentationPointError(point, c.cutoff)
			return c.sealError
		}
		select {
		case <-drain.Done():
			c.sealError = drain.Err()
			return c.sealError
		case <-ticker.C:
		}
	}
}

type presentationCaptureWriter struct {
	target    io.Writer
	remaining int
}

func (w *presentationCaptureWriter) Write(p []byte) (int, error) {
	if len(p) > w.remaining {
		return 0, errors.New("presentation capture: file budget exceeded")
	}
	n, err := w.target.Write(p)
	w.remaining -= n
	if n != len(p) && err == nil {
		err = io.ErrShortWrite
	}
	return n, err
}

// Finish is called only after Run and the measurement worker have retired.
// Both failed and successful runs retain evidence, including late close records.
func (c *perfPresentationCapture) Finish(runError error, workerRetired, runtimeReturned bool) error {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.finished {
		return c.finishError
	}
	c.finished = true
	if !c.sealed {
		c.sealError = errors.New("presentation capture: pre-close snapshot was not sealed")
	}
	if !workerRetired {
		c.sealError = errors.Join(c.sealError, errors.New("presentation capture: measurement worker is not retired"))
	}
	if !runtimeReturned {
		c.sealError = errors.Join(c.sealError, errors.New("presentation capture: native runtime did not return"))
	}
	var after *presentationCapturePoint
	if workerRetired && runtimeReturned {
		point := c.point()
		after = &point
		if point.Presentation.Failed || point.Presentation.Overflow || point.Presentation.Pending != 0 || point.Frames.Failure != "" || point.Frames.Pending != 0 || point.Frames.Dropped != 0 {
			c.sealError = errors.Join(c.sealError, errors.New("presentation capture: post-close recorder is not terminal"))
		}
	}
	if c.config.Paced && (c.pacing == nil || !c.pacing.Completed || c.pacing.Failure != "") {
		c.sealError = errors.Join(c.sealError, errors.New("paced workload: cohort not completed"))
	}
	terminal := errors.Join(runError, c.sealError)
	result := "CAPTURED_DIAGNOSTIC"
	message := ""
	if terminal != nil {
		result = "FAILED_DIAGNOSTIC"
		message = terminal.Error()
		if len(message) > 4096 {
			message = message[:4096]
		}
	}
	value := struct {
		Schema, Result, Window, RunError, CutoffMethod, BinarySHA256 string
		Config                                                       presentationCaptureConfig
		Build                                                        map[string]string
		ContinuousWork, WorkerRetired, RuntimeReturned               bool
		AcquisitionCutoff                                            uint64
		ReceiptCutoff                                                int
		BeforeClose, AfterClose                                      *presentationCapturePoint
		Pacing                                                       *perfPacingJournal `json:",omitempty"`
	}{Schema: "vellum-presentation-capture-v1", Result: result, Window: "perf", RunError: message, CutoffMethod: "acquisition-sequence prefix at completed CPU cohort", BinarySHA256: c.binarySHA256, Config: c.config, Build: c.build, ContinuousWork: c.config.Paced, WorkerRetired: workerRetired, RuntimeReturned: runtimeReturned, AcquisitionCutoff: c.cutoff, ReceiptCutoff: c.receiptCutoff, BeforeClose: c.before, AfterClose: after, Pacing: c.pacing}
	if c.config.Paced {
		value.Schema = "vellum-presentation-capture-v2"
	}
	writer := &presentationCaptureWriter{target: c.file, remaining: presentationCaptureMaxBytes}
	c.finishError = errors.Join(c.sealError, json.NewEncoder(writer).Encode(value), c.file.Close())
	return c.finishError
}

// The result channel alone precedes worker cleanup. Require its terminal ack.
func joinPerfPresentationWorker(done <-chan struct{}, cancel context.CancelFunc) bool {
	select {
	case <-done:
		return true
	default:
	}
	cancel()
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	select {
	case <-done:
		return true
	case <-timer.C:
		return false
	}
}

func perfPresentationStatus(path string) string {
	if path == "" {
		return "unavailable"
	}
	return "separate serial diagnostic journal"
}

// Hash the executable once outside measured work. External launch/build binding
// is still required: an absent VCS stamp is common with an external modfile.
func hashPresentationFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	const limit = 512 << 20
	if !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > limit {
		return "", errors.New("presentation capture: executable file exceeds scope")
	}
	hash := sha256.New()
	count, err := io.Copy(hash, io.LimitReader(file, limit+1))
	if err != nil {
		return "", err
	}
	if count != info.Size() {
		return "", errors.New("presentation capture: executable size changed")
	}
	final, err := file.Stat()
	if err != nil {
		return "", err
	}
	if final.Size() != info.Size() || !final.ModTime().Equal(info.ModTime()) {
		return "", errors.New("presentation capture: executable changed")
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func markPresentationRecord(value map[string]any, enabled bool) map[string]any {
	if enabled {
		value["presentation_capture"] = true
		value["cadence_qualified"] = false
	}
	return value
}
