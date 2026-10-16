//go:build darwin && cgo && vellum_native_checks

package main

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	cogent "github.com/dshills/vellum/drivers/cogent"
)

func TestSoakNativeActivityValidatesScopeSaturationAndMonotonicCounts(t *testing.T) {
	start := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	bounds := soakActivityBounds{StartRead: soakActivityReadWindow{start.Add(-2 * time.Second), start.Add(-time.Second)}, EndRead: soakActivityReadWindow{start.Add(61 * time.Second), start.Add(62 * time.Second)}, QuietStart: start, QuietEnd: start.Add(time.Minute)}
	before := cogent.NativeActivityCheck{Available: true, ProcessID: 123, EmptyEventPostAttempts: 1<<60 + 7, SurfacePresentAttempts: 1<<60 + 17}
	after := before
	after.EmptyEventPostAttempts += 3
	after.SurfacePresentAttempts += 2
	got, err := soakActivityDelta(123, before, after, bounds)
	if err != nil || !got.Available || got.EmptyEventPostAttemptsDelta != 3 || got.SurfacePresentAttemptsDelta != 2 || got.Start != before || got.End != after {
		t.Fatal("exact attempt deltas lost", got, err)
	}
	for _, change := range []func(*cogent.NativeActivityCheck){
		func(c *cogent.NativeActivityCheck) { c.Available = false },
		func(c *cogent.NativeActivityCheck) { c.ProcessID = 0 },
		func(c *cogent.NativeActivityCheck) { c.ProcessID++ },
		func(c *cogent.NativeActivityCheck) { c.Saturated = true },
		func(c *cogent.NativeActivityCheck) { c.EmptyEventPostAttempts = math.MaxUint64 },
		func(c *cogent.NativeActivityCheck) { c.SurfacePresentAttempts = math.MaxUint64 },
		func(c *cogent.NativeActivityCheck) { c.EmptyEventPostAttempts = before.EmptyEventPostAttempts - 1 },
		func(c *cogent.NativeActivityCheck) { c.SurfacePresentAttempts = before.SurfacePresentAttempts - 1 },
	} {
		bad := after
		change(&bad)
		value, err := soakActivityDelta(123, before, bad, bounds)
		if err == nil || value.Available || value.Start != before || value.End != bad || value.EmptyEventPostAttemptsDelta != 0 || value.SurfacePresentAttemptsDelta != 0 {
			t.Fatal("bad counters exposed usable deltas", before, bad, value, err)
		}
	}
	for _, bad := range []cogent.NativeActivityCheck{{}, {Available: true, ProcessID: -1}, {Available: true, ProcessID: 123, Saturated: true}, {Available: true, ProcessID: 123, EmptyEventPostAttempts: math.MaxUint64}, {Available: true, ProcessID: 123, SurfacePresentAttempts: math.MaxUint64}} {
		if value, err := soakActivityDelta(123, bad, after, bounds); err == nil || value.Available {
			t.Fatal("bad start accepted", bad, value, err)
		}
	}
	for _, pid := range []int{-1, 0, 124} {
		if value, err := soakActivityDelta(pid, before, after, bounds); err == nil || value.Available {
			t.Fatal("unbound owned PID accepted", pid, value, err)
		}
	}
	for _, change := range []func(*soakActivityBounds){
		func(b *soakActivityBounds) { b.StartRead.StartUTC = time.Time{} },
		func(b *soakActivityBounds) { b.StartRead.EndUTC = b.StartRead.StartUTC.Add(-time.Second) },
		func(b *soakActivityBounds) { b.StartRead.EndUTC = b.QuietStart.Add(time.Nanosecond) },
		func(b *soakActivityBounds) { b.EndRead.StartUTC = b.QuietEnd.Add(-time.Nanosecond) },
		func(b *soakActivityBounds) { b.EndRead.EndUTC = b.EndRead.StartUTC.Add(-time.Nanosecond) },
		func(b *soakActivityBounds) { b.QuietEnd = b.QuietStart },
	} {
		bad := bounds
		change(&bad)
		if value, err := soakActivityDelta(123, before, after, bad); err == nil || value.Available {
			t.Fatal("unbracketed read accepted", bad, value, err)
		}
	}
	raw, err := json.Marshal(after)
	var detached cogent.NativeActivityCheck
	if err != nil || json.Unmarshal(raw, &detached) != nil || detached != after {
		t.Fatal("large integer counter JSON changed", string(raw), detached, err)
	}
	encoded, err := json.Marshal(got)
	var observation soakActivityObservation
	if err != nil || json.Unmarshal(encoded, &observation) != nil || observation != got || observation.Version != 1 {
		t.Fatal("raw counters/read bounds lost in observation JSON", string(encoded), observation, err)
	}
}

func TestSoakNativeActivityPublishesUTCWallBounds(t *testing.T) {
	// time.Now supplies monotonic readings without changing the host clock.
	start := time.Now()
	bounds := soakActivityBounds{
		StartRead:  soakActivityReadWindow{start.Add(-2 * time.Second), start.Add(-time.Second)},
		EndRead:    soakActivityReadWindow{start.Add(61 * time.Second), start.Add(62 * time.Second)},
		QuietStart: start, QuietEnd: start.Add(time.Minute),
	}
	counters := cogent.NativeActivityCheck{Available: true, ProcessID: 123}
	value, err := soakActivityDelta(123, counters, counters, bounds)
	if err != nil || !value.Available {
		t.Fatal("valid monotonic read bounds rejected", value, err)
	}
	for _, pair := range [][2]time.Time{
		{value.StartRead.StartUTC, bounds.StartRead.StartUTC},
		{value.StartRead.EndUTC, bounds.StartRead.EndUTC},
		{value.EndRead.StartUTC, bounds.EndRead.StartUTC},
		{value.EndRead.EndUTC, bounds.EndRead.EndUTC},
	} {
		if pair[0] != pair[1].UTC() || pair[0].Location() != time.UTC {
			t.Fatal("read retained unpublished monotonic/local data", pair)
		}
	}
	raw, err := json.Marshal(value)
	var decoded soakActivityObservation
	if err != nil || json.Unmarshal(raw, &decoded) != nil || decoded != value {
		t.Fatal("validated read bounds differ from serialized wall times", string(raw), decoded, err)
	}
	// A reversed wall bound must be refused even when other bounds still have
	// monotonic data. The refusal retains the exact wire bounds for diagnosis.
	bounds.EndRead.StartUTC = bounds.QuietEnd.UTC().Add(-time.Nanosecond)
	bad, err := soakActivityDelta(123, counters, counters, bounds)
	if err == nil || bad.Available || bad.EndRead.StartUTC != bounds.EndRead.StartUTC {
		t.Fatal("reversed serialized wall bound accepted or lost", bad, err)
	}
}
