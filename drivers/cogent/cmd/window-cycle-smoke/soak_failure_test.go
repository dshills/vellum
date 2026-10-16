//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"encoding/json"
	"errors"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"math"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func soakEndpointFixture() soakEndpoint {
	return soakEndpoint{target: cogent.NativeServiceWindowCheck{Title: "Vellum composed idle soak", Visible: true, Frame: [4]float64{20, 30, 100, 100}, OrderedIndex: 1, Opaque: true, Alpha: 1}, cover: cogent.NativeServiceWindowCheck{Title: "Vellum cycle survivor", Visible: true, Key: true, Frame: [4]float64{10, 20, 120, 120}, OrderedIndex: 0, Opaque: true, Alpha: 1}, at: time.Date(2026, 10, 16, 4, 0, 0, 0, time.UTC)}
}
func TestSoakFailureKeepsStartAndEndDrift(t *testing.T) {
	for _, which := range []string{"start", "end"} {
		t.Run(which, func(t *testing.T) {
			start, end := soakEndpointFixture(), soakEndpointFixture()
			end.at = end.at.Add(time.Minute)
			if which == "start" {
				start.target.Key = true
			} else {
				end.target.Key = true
			}
			var records []soakFailureRecord
			err := checkSoakEndpoints("occluded", 7, start, end, start.at, end.at, func(v any) error { records = append(records, v.(soakFailureRecord)); return nil })
			if err == nil || err.Error() != "soak occluded state drift" || len(records) != 1 {
				t.Fatalf("failure vanished: %v %+v", err, records)
			}
			r := records[0]
			if r.Stage != "state-drift" || r.Attempt != 7 || r.Start.TargetStateAccepted != (which != "start") || r.End.TargetStateAccepted != (which != "end") || r.Start.Target.Key != (which == "start") || r.End.Target.Key != (which == "end") {
				t.Fatalf("failed endpoint replaced: %+v", r)
			}
			if !r.Start.CoverContainsTarget || !r.End.CoverOcclusionAccepted || r.Start.Target.Frame != [4]string{"20", "30", "100", "100"} || r.QuietStartUTC != start.at.Format(time.RFC3339Nano) || r.QuietEndUTC != end.at.Format(time.RFC3339Nano) {
				t.Fatalf("captured context lost: %+v", r)
			}
		})
	}
}
func TestSoakFailureRetainsQueryAndEmissionErrors(t *testing.T) {
	for _, which := range []string{"start", "end"} {
		original, writeErr := errors.New("query failed"), errors.New("stdout failed")
		start, end := soakEndpointFixture(), soakEndpointFixture()
		if which == "start" {
			start.err = original
		} else {
			end.err = original
		}
		var record soakFailureRecord
		err := checkSoakEndpoints("occluded", 1, start, end, time.Time{}, time.Time{}, func(v any) error { record = v.(soakFailureRecord); return writeErr })
		if !errors.Is(err, original) || !errors.Is(err, writeErr) || record.Stage != which+"-query" {
			t.Fatalf("failure identity lost: %v %+v", err, record)
		}
		failed := record.Start
		if which == "end" {
			failed = record.End
		}
		if failed.QuerySucceeded || failed.TargetStateAccepted || failed.QueryError != "query failed" {
			t.Fatalf("query failure presented as observed success: %+v", failed)
		}
	}
}
func TestSoakFailureSerializationIsBoundedAndPreservesNonfiniteValues(t *testing.T) {
	start, end := soakEndpointFixture(), soakEndpointFixture()
	end.target.Title = strings.Repeat("é\x00", 4096) + "\xff"
	end.target.Frame[0] = math.NaN()
	end.target.Alpha = math.Inf(1)
	var encoded []byte
	var record soakFailureRecord
	err := checkSoakEndpoints("occluded", 1, start, end, start.at, end.at, func(v any) error { record = v.(soakFailureRecord); var e error; encoded, e = json.Marshal(v); return e })
	if err == nil || !json.Valid(encoded) || len(encoded) > 64<<10 || len(record.End.Target.Title) > 512 || !utf8.ValidString(record.End.Target.Title) || record.End.Target.Frame[0] != "NaN" || record.End.Target.Alpha != "+Inf" {
		t.Fatalf("invalid diagnostic: %v len=%d %+v", err, len(encoded), record)
	}
}
func TestSoakFailureBoundsQueryTextAndConvertsTimestampsToUTC(t *testing.T) {
	start, end := soakEndpointFixture(), soakEndpointFixture()
	end.err = errors.New(strings.Repeat("é", 4096))
	end.at = end.at.In(time.FixedZone("offset", 10800))
	var record soakFailureRecord
	err := checkSoakEndpoints("occluded", 1, start, end, start.at, end.at, func(v any) error { record = v.(soakFailureRecord); return nil })
	if !errors.Is(err, end.err) || len(record.End.QueryError) > 512 || !utf8.ValidString(record.End.QueryError) || !strings.HasSuffix(record.End.ObservedUTC, "Z") {
		t.Fatalf("invalid text/time: %v %+v", err, record)
	}
}
func TestSoakEndpointSuccessEmitsNothing(t *testing.T) {
	start, end := soakEndpointFixture(), soakEndpointFixture()
	called := false
	if err := checkSoakEndpoints("occluded", 1, start, end, start.at, end.at, func(any) error { called = true; return nil }); err != nil || called {
		t.Fatalf("successful endpoints changed: %v emit=%t", err, called)
	}
}
func TestSoakEndpointRequiresCapturedTimeAndStrictOcclusion(t *testing.T) {
	for _, change := range []func(*soakEndpoint){func(e *soakEndpoint) { e.at = time.Time{} }, func(e *soakEndpoint) { e.target.OcclusionVisible = true }} {
		start, end := soakEndpointFixture(), soakEndpointFixture()
		change(&end)
		called := false
		if err := checkSoakEndpoints("occluded", 1, start, end, start.at, end.at, func(any) error { called = true; return nil }); err == nil || !called {
			t.Fatalf("unobserved/visible endpoint qualified: %v emit=%t", err, called)
		}
	}
}
