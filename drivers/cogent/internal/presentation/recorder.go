// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
// Package presentation holds platform-independent accounting for native feedback.
// It does not manufacture presentation timestamps or establish display coverage.
package presentation

import (
	"errors"
	"math"
	"strings"
	"sync"
)

const MaxCapacity = 65536
const MaxScopes = 64

var ErrConfig = errors.New("presentation: invalid or copied recorder")
var ErrBudget = errors.New("presentation: recording capacity exhausted")

type Frame struct {
	Tag, Generation, Composition uint64
	Armed                        bool
}
type Sample struct {
	Sequence, Scope uint64
	Frame
	DrawableID          uint64
	Acquired, Presented float64 // seconds in one native host-time domain
	State               string  // pending, presented, skipped, unavailable, invalid, retired
}
type Input struct {
	Scope, Tag uint64
	Received   float64
}
type Scope struct {
	ID     uint64
	Window string
	Closed bool
}
type Snapshot struct {
	Scopes                                          []Scope
	Inputs                                          []Input
	Samples                                         []Sample
	Pending, Skipped, Unavailable, Invalid, Retired int
	Overflow, Failed                                bool
	Failure                                         string
}
type inputKey struct{ scope, tag uint64 }

// Recorder shares one fixed total capacity across all window incarnations.
// Methods never call platform or application code while holding its mutex.
type Recorder struct {
	self      *Recorder
	mu        sync.Mutex
	capacity  int
	scopes    []Scope
	inputs    []Input
	inputTags map[inputKey]bool
	samples   []Sample
	overflow  bool
	failure   string
}

func New(capacity int) (*Recorder, error) {
	if capacity < 1 || capacity > MaxCapacity {
		return nil, ErrConfig
	}
	r := &Recorder{capacity: capacity, inputTags: make(map[inputKey]bool)}
	r.self = r
	return r, nil
}
func (r *Recorder) valid() bool {
	return r != nil && r.self == r && r.capacity >= 1 && r.capacity <= MaxCapacity
}
func clock(v float64) bool { return v > 0 && !math.IsNaN(v) && !math.IsInf(v, 0) }
func (r *Recorder) fail(message string) {
	if r.failure == "" {
		r.failure = message
	}
}
func (r *Recorder) live(scope uint64) bool {
	return scope > 0 && scope <= uint64(len(r.scopes)) && !r.scopes[scope-1].Closed
}

func (r *Recorder) Open(window string) (uint64, error) {
	if !r.valid() || window == "" || len(window) > 4096 {
		return 0, ErrConfig
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.scopes) == MaxScopes {
		r.overflow = true
		r.fail("window scope capacity exhausted")
		return 0, ErrBudget
	}
	id := uint64(len(r.scopes)) + 1
	r.scopes = append(r.scopes, Scope{ID: id, Window: strings.Clone(window)})
	return id, nil
}

// Receipt precedes enqueueing the traced input. It must use the same native
// host clock as Acquired and Presented, not a Go wall-clock conversion.
func (r *Recorder) Receipt(scope, tag uint64, at float64) bool {
	if !r.valid() {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.live(scope) {
		return false
	}
	if tag == 0 || !clock(at) {
		r.fail("invalid input receipt")
		return false
	}
	key := inputKey{scope, tag}
	if r.inputTags[key] {
		r.fail("duplicate input receipt")
		return false
	}
	if len(r.inputs)+len(r.samples) >= r.capacity {
		r.overflow = true
		r.fail("input receipt capacity exhausted")
		return false
	}
	r.inputTags[key] = true
	r.inputs = append(r.inputs, Input{scope, tag, at})
	return true
}

// Acquire reserves an entry before registering a native presented handler.
// A zero return must prevent installing a handler for an untracked drawable.
func (r *Recorder) Acquire(scope uint64, frame Frame, at float64) uint64 {
	if !r.valid() {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.live(scope) {
		return 0
	}
	if !clock(at) || frame.Armed && (frame.Tag == 0 || frame.Generation == 0 || frame.Composition == 0) {
		r.fail("invalid acquisition metadata")
		return 0
	}
	if len(r.inputs)+len(r.samples) >= r.capacity {
		r.overflow = true
		r.fail("drawable capacity exhausted")
		return 0
	}
	sequence := uint64(len(r.samples)) + 1
	r.samples = append(r.samples, Sample{Sequence: sequence, Scope: scope, Frame: frame, Acquired: at, State: "pending"})
	return sequence
}

// Unavailable records an actual nil drawable without installing a handler.
// A concurrent close may retire the reservation; neither outcome is presented.
func (r *Recorder) Unavailable(scope uint64, frame Frame, at float64) {
	sequence := r.Acquire(scope, frame, at)
	if sequence == 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if s := &r.samples[sequence-1]; r.live(scope) && s.State == "pending" {
		s.State = "unavailable"
	}
}

// Complete may arrive in any callback order. Drawable ID zero is valid.
// Zero presentation time is skipped/unpresented, never a successful sample.
func (r *Recorder) Complete(scope, sequence, drawable uint64, at float64) {
	if !r.valid() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.live(scope) {
		return
	} // retirement revokes every late callback
	if sequence == 0 || sequence > uint64(len(r.samples)) {
		r.fail("unknown acquisition sequence")
		return
	}
	s := &r.samples[sequence-1]
	if s.Scope != scope {
		r.fail("presentation scope mismatch")
		return
	}
	if s.State != "pending" {
		r.fail("duplicate presentation completion")
		return
	}
	s.DrawableID = drawable
	// Preserve finite invalid values for diagnosis, while keeping snapshots JSON
	// encodable when a faulty backend supplies NaN or infinity.
	if !math.IsNaN(at) && !math.IsInf(at, 0) {
		s.Presented = at
	}
	if at == 0 {
		s.State = "skipped"
		return
	}
	if !clock(at) || at < s.Acquired {
		s.State = "invalid"
		r.fail("invalid presentation timestamp")
		return
	}
	s.State = "presented"
}

// Close preserves unresolved records as retired, rather than pretending that
// window teardown presented them. No platform references are retained here.
func (r *Recorder) Close(scope uint64) {
	if !r.valid() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.live(scope) {
		return
	}
	r.scopes[scope-1].Closed = true
	for i := range r.samples {
		s := &r.samples[i]
		if s.Scope == scope && s.State == "pending" {
			s.State = "retired"
		}
	}
}

func (r *Recorder) Snapshot() Snapshot {
	if !r.valid() {
		return Snapshot{Failed: true, Failure: ErrConfig.Error()}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	s := Snapshot{Scopes: append([]Scope(nil), r.scopes...), Inputs: append([]Input(nil), r.inputs...), Samples: append([]Sample(nil), r.samples...), Overflow: r.overflow, Failed: r.failure != "", Failure: r.failure}
	for _, v := range s.Samples {
		switch v.State {
		case "pending":
			s.Pending++
		case "skipped":
			s.Skipped++
		case "unavailable":
			s.Unavailable++
		case "invalid":
			s.Invalid++
		case "retired":
			s.Retired++
		}
	}
	return s
}
