// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
package cogent

import (
	"cogentcore.org/core/system/driver/desktop"
	"github.com/dshills/vellum/drivers/cogent/internal/presentation"
	"sync"
)

// PresentationSnapshot is detached native host-time feedback, not CPU work.
// Missing, pending, skipped, unavailable, retired or failed samples do not pass
// a display gate. Snapshots retain no platform, window or application references.
type PresentationSnapshot = presentation.Snapshot
type PresentationSample = presentation.Sample
type PresentationInput = presentation.Input
type PresentationFrame = presentation.Frame
type PresentationScope = presentation.Scope

// PresentationMetrics records optional actual Metal drawable feedback. Input
// receipts currently cover EnqueueGridScroll/EnqueueTreeScroll before FIFO send;
// they are routed requests, not physical input-to-photon measurements.
// One shared budget (1..65536) counts receipts plus drawable records across at
// most 64 window incarnations. Overflow fails the diagnostic; no eviction hides
// missing feedback. Count limits are not exact allocator/process byte limits.
type PresentationMetrics struct {
	self     *PresentationMetrics
	recorder *presentation.Recorder
}

func NewPresentationMetrics(capacity int) (*PresentationMetrics, error) {
	recorder, err := presentation.New(capacity)
	if err != nil {
		return nil, err
	}
	m := &PresentationMetrics{recorder: recorder}
	m.self = m
	return m, nil
}
func (m *PresentationMetrics) valid() bool { return m != nil && m.self == m && m.recorder != nil }
func (m *PresentationMetrics) Snapshot() PresentationSnapshot {
	if !m.valid() {
		return (*presentation.Recorder)(nil).Snapshot()
	}
	return m.recorder.Snapshot()
}

type presentationSink struct {
	recorder    *presentation.Recorder
	scope       uint64
	attach      sync.Once
	attachError error
}

func (m *PresentationMetrics) bind(window string) (*presentationSink, error) {
	scope, err := m.recorder.Open(window)
	if err != nil {
		return nil, err
	}
	return &presentationSink{recorder: m.recorder, scope: scope}, nil
}
func (s *presentationSink) Receipt(tag uint64, at float64) { s.recorder.Receipt(s.scope, tag, at) }
func presentationFrame(f desktop.PresentationFrame) presentation.Frame {
	return presentation.Frame{Tag: f.Tag, Generation: f.Generation, Composition: f.Composition, Armed: f.Armed}
}
func (s *presentationSink) Acquired(f desktop.PresentationFrame, at float64) uint64 {
	return s.recorder.Acquire(s.scope, presentationFrame(f), at)
}
func (s *presentationSink) Unavailable(f desktop.PresentationFrame, at float64) {
	s.recorder.Unavailable(s.scope, presentationFrame(f), at)
}
func (s *presentationSink) Presented(sequence, drawable uint64, at float64) {
	s.recorder.Complete(s.scope, sequence, drawable, at)
}
func (s *presentationSink) Closed() {
	if s != nil {
		s.recorder.Close(s.scope)
	}
}
func recordPresentationReceipt(w *nativeWindow, tag uint64) {
	if sink := w.presentation; sink != nil {
		// The sink owns only recorder+scope. Sampling the native host clock needs
		// no Scene/Stage/window lookup or render lock on this app-owner path.
		sink.Receipt(tag, desktop.PresentationHostTime())
	}
}
func attachPresentation(w *nativeWindow, window any) error {
	if w.presentation == nil {
		return nil
	}
	target, ok := window.(interface {
		SetPresentationObserver(desktop.PresentationObserver) error
	})
	if !ok {
		return presentation.ErrConfig
	}
	w.presentation.attach.Do(func() { w.presentation.attachError = target.SetPresentationObserver(w.presentation) })
	return w.presentation.attachError
}
