// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
package presentation

import (
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"
)

func setup(t *testing.T, capacity int) (*Recorder, uint64) {
	t.Helper()
	r, err := New(capacity)
	if err != nil {
		t.Fatal(err)
	}
	s, err := r.Open("main")
	if err != nil {
		t.Fatal(err)
	}
	return r, s
}
func armed(tag uint64) Frame { return Frame{Tag: tag, Generation: 7, Composition: 3, Armed: true} }

func TestAcquisitionIsNotPresentation(t *testing.T) {
	r, s := setup(t, 4)
	if !r.Receipt(s, 1, 10) {
		t.Fatal("receipt")
	}
	first := r.Acquire(s, Frame{Tag: 1, Composition: 1}, 10.01)
	final := r.Acquire(s, armed(1), 10.02)
	if first == 0 || final != first+1 || r.Snapshot().Pending != 2 {
		t.Fatal(r.Snapshot())
	}
	r.Complete(s, final, 0, 10.04) // native drawable ID zero is valid
	r.Complete(s, first, 9, 10.03) // callback order need not equal acquisition order
	snap := r.Snapshot()
	if snap.Failed || snap.Pending != 0 || snap.Samples[1].State != "presented" || snap.Samples[1].DrawableID != 0 || !snap.Samples[1].Armed || snap.Samples[0].Armed || snap.Inputs[0].Received != 10 {
		t.Fatal(snap)
	}
	// A detached earlier snapshot cannot become a new presentation after reuse.
	snap.Samples[0].Presented = 999
	snap.Inputs[0].Received = 999
	snap.Scopes[0].Window = "changed"
	if r.Snapshot().Samples[0].Presented == 999 || r.Snapshot().Inputs[0].Received == 999 || r.Snapshot().Scopes[0].Window != "main" {
		t.Fatal("mutable snapshot alias")
	}
}
func TestSkippedInvalidAndDuplicateAreNotSuccessfulFrames(t *testing.T) {
	for _, at := range []float64{0, -1, math.NaN(), math.Inf(1), 9} {
		r, s := setup(t, 1)
		seq := r.Acquire(s, Frame{}, 10)
		r.Complete(s, seq, 4, at)
		v := r.Snapshot()
		if v.Samples[0].State == "presented" || v.Pending != 0 {
			t.Fatal(at, v)
		}
		if at == 0 && (v.Skipped != 1 || v.Samples[0].Presented != 0) {
			t.Fatal(v)
		}
		if at != 0 && (v.Invalid != 1 || !v.Failed) {
			t.Fatal(v)
		}
		if _, err := json.Marshal(v); err != nil {
			t.Fatal("invalid feedback broke diagnostic JSON", err)
		}
	}
	r, s := setup(t, 1)
	seq := r.Acquire(s, Frame{}, 10)
	r.Complete(s, seq, 4, 11)
	r.Complete(s, seq, 5, 12)
	v := r.Snapshot()
	if !v.Failed || v.Samples[0].DrawableID != 4 || v.Samples[0].Presented != 11 {
		t.Fatal(v)
	}
}
func TestScopeAndCapacityCannotSilentlyDropWork(t *testing.T) {
	r, s := setup(t, 3)
	if !r.Receipt(s, 1, 1) {
		t.Fatal("first input")
	}
	seq := r.Acquire(s, Frame{}, 2)
	if seq == 0 || !r.Receipt(s, 2, 3) {
		t.Fatal("mixed entries within capacity")
	}
	if r.Receipt(s, 3, 4) || r.Acquire(s, Frame{}, 4) != 0 {
		t.Fatal("combined capacity")
	}
	if v := r.Snapshot(); !v.Overflow || !v.Failed || v.Pending != 1 || len(v.Inputs) != 2 || len(v.Samples) != 1 {
		t.Fatal(v)
	}
	// Each entry kind consumes the shared final slot, regardless of order.
	for _, inputFirst := range []bool{true, false} {
		limited, scope := setup(t, 1)
		if inputFirst {
			if !limited.Receipt(scope, 1, 1) || limited.Acquire(scope, Frame{}, 2) != 0 {
				t.Fatal("receipt must exhaust shared budget")
			}
		} else if limited.Acquire(scope, Frame{}, 1) == 0 || limited.Receipt(scope, 1, 2) {
			t.Fatal("acquisition must exhaust shared budget")
		}
		v := limited.Snapshot()
		if !v.Overflow || !v.Failed || len(v.Inputs)+len(v.Samples) != 1 {
			t.Fatal(v)
		}
	}
	r.Close(s)
	before := r.Snapshot()
	r.Complete(s, seq, 1, 5)
	if after := r.Snapshot(); !reflect.DeepEqual(before, after) || after.Retired != 1 || after.Pending != 0 || r.Receipt(s, 3, 3) || r.Acquire(s, Frame{}, 6) != 0 {
		t.Fatal(before, after)
	}
	newScope, err := r.Open("main")
	if err != nil || newScope == s {
		t.Fatal(newScope, err)
	}
	if r.Snapshot().Scopes[1].Closed {
		t.Fatal("replacement scope inherited retirement")
	}
}
func TestInvalidMetadataAndReceiptsRefuseWithoutInstallingHandlers(t *testing.T) {
	for _, frame := range []Frame{{Armed: true}, {Tag: 1, Armed: true}, {Tag: 1, Generation: 1, Armed: true}} {
		r, s := setup(t, 1)
		if r.Acquire(s, frame, 1) != 0 || !r.Snapshot().Failed || len(r.Snapshot().Samples) != 0 {
			t.Fatal(frame, r.Snapshot())
		}
	}
	for _, at := range []float64{0, -1, math.Inf(-1), math.NaN()} {
		r, s := setup(t, 1)
		if r.Acquire(s, Frame{}, at) != 0 || r.Receipt(s, 1, at) || !r.Snapshot().Failed {
			t.Fatal(at, r.Snapshot())
		}
	}
	r, s := setup(t, 3)
	r.Receipt(s, 1, 1)
	if r.Receipt(s, 1, 2) || !r.Snapshot().Failed || len(r.Snapshot().Inputs) != 1 {
		t.Fatal(r.Snapshot())
	}
	other, err := r.Open("other")
	if err != nil {
		t.Fatal(err)
	}
	seq := r.Acquire(s, Frame{}, 3)
	r.Complete(other, seq, 4, 5)
	if v := r.Snapshot(); v.Pending != 1 || !v.Failed {
		t.Fatal(v)
	}
}
func TestConfigurationScopeLimitAndCopiedRecorder(t *testing.T) {
	for _, capacity := range []int{-1, 0, MaxCapacity + 1} {
		if _, err := New(capacity); !errors.Is(err, ErrConfig) {
			t.Fatal(capacity, err)
		}
	}
	r, err := New(1)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < MaxScopes; i++ {
		s, err := r.Open("window")
		if err != nil {
			t.Fatal(err)
		}
		r.Close(s)
	}
	if _, err := r.Open("overflow"); !errors.Is(err, ErrBudget) || !r.Snapshot().Overflow {
		t.Fatal(err, r.Snapshot())
	}
	for _, invalid := range []*Recorder{nil, {}, {self: r, capacity: 1}} {
		if _, err := invalid.Open("bad"); !errors.Is(err, ErrConfig) || !invalid.Snapshot().Failed || invalid.Acquire(1, Frame{}, 1) != 0 {
			t.Fatal(err)
		}
		invalid.Complete(1, 1, 1, 1)
		invalid.Close(1)
	}
}
func TestCompletionAndRetirementRacePreservesTerminalAccounting(t *testing.T) {
	for cycle := 0; cycle < 32; cycle++ {
		r, s := setup(t, 64)
		seqs := make([]uint64, 64)
		for i := range seqs {
			seqs[i] = r.Acquire(s, Frame{}, 1)
		}
		var wg sync.WaitGroup
		wg.Add(3)
		go func() {
			defer wg.Done()
			for _, seq := range seqs {
				r.Complete(s, seq, seq, 2)
			}
		}()
		go func() { defer wg.Done(); r.Close(s) }()
		go func() {
			defer wg.Done()
			for i := 0; i < 64; i++ {
				_ = r.Snapshot()
			}
		}()
		wg.Wait()
		v := r.Snapshot()
		presented := 0
		for _, sample := range v.Samples {
			if sample.State == "presented" {
				presented++
			}
		}
		if v.Failed || v.Pending != 0 || v.Retired+presented != 64 {
			t.Fatal(v)
		}
	}
}
