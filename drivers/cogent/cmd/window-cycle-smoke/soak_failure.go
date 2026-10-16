//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"strconv"
	"time"
)

// Endpoints are captured outside the quiet CPU interval. A failed query retains
// its original error and partial values; later queries never replace it.
type soakEndpoint struct {
	target, cover cogent.NativeServiceWindowCheck
	err           error
	at            time.Time
}
type soakFailureWindow struct {
	Title            string    `json:"title"`
	Visible          bool      `json:"visible"`
	Miniaturized     bool      `json:"miniaturized"`
	Key              bool      `json:"key"`
	OcclusionVisible bool      `json:"occlusion_visible"`
	Frame            [4]string `json:"frame_points"`
	OrderedIndex     int       `json:"ordered_index"`
	Opaque           bool      `json:"opaque"`
	Alpha            string    `json:"alpha"`
	Level            int       `json:"level"`
}
type soakFailureEndpoint struct {
	QuerySucceeded         bool              `json:"query_succeeded"`
	QueryError             string            `json:"query_error"`
	ObservedUTC            string            `json:"observed_utc"`
	Target                 soakFailureWindow `json:"target"`
	Cover                  soakFailureWindow `json:"cover"`
	TargetStateAccepted    bool              `json:"target_state_accepted"`
	CoverContainsTarget    bool              `json:"cover_contains_target"`
	CoverOcclusionAccepted bool              `json:"cover_occlusion_accepted"`
}
type soakFailureRecord struct {
	Event         string              `json:"event"`
	Phase         string              `json:"phase"`
	Attempt       int                 `json:"attempt"`
	Stage         string              `json:"stage"`
	Cause         string              `json:"cause"`
	QuietStartUTC string              `json:"quiet_start_utc"`
	QuietEndUTC   string              `json:"quiet_end_utc"`
	Start         soakFailureEndpoint `json:"start"`
	End           soakFailureEndpoint `json:"end"`
}

func soakFailureTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}
func soakFailureWindowValue(w cogent.NativeServiceWindowCheck) soakFailureWindow {
	f := soakFailureWindow{Title: soakErrorText(w.Title), Visible: w.Visible, Miniaturized: w.Miniaturized, Key: w.Key, OcclusionVisible: w.OcclusionVisible, OrderedIndex: w.OrderedIndex, Opaque: w.Opaque, Alpha: strconv.FormatFloat(w.Alpha, 'g', -1, 64), Level: w.Level}
	// Strings preserve nonfinite diagnostic values without invalid JSON or
	// replacing an impossible native frame with fabricated finite geometry.
	for i, v := range w.Frame {
		f.Frame[i] = strconv.FormatFloat(v, 'g', -1, 64)
	}
	return f
}
func soakFailureEndpointValue(e soakEndpoint, phase string) soakFailureEndpoint {
	f := soakFailureEndpoint{QuerySucceeded: e.err == nil && !e.at.IsZero(), ObservedUTC: soakFailureTime(e.at), Target: soakFailureWindowValue(e.target), Cover: soakFailureWindowValue(e.cover)}
	if e.err != nil {
		f.QueryError = soakErrorText(e.err.Error())
	}
	f.TargetStateAccepted = f.QuerySucceeded && soakState(e.target, phase)
	f.CoverContainsTarget = f.QuerySucceeded && soakContains(e.cover.Frame, e.target.Frame)
	f.CoverOcclusionAccepted = f.QuerySucceeded && soakCoverState(e.target, e.cover)
	return f
}
func checkSoakEndpoints(phase string, attempt int, start, end soakEndpoint, t0, t1 time.Time, emit func(any) error) error {
	var cause error
	stage := "state-drift"
	switch {
	case start.err != nil:
		cause = start.err
		stage = "start-query"
	case end.err != nil:
		cause = end.err
		stage = "end-query"
	case start.at.IsZero() || end.at.IsZero() || !soakState(start.target, phase) || !soakState(end.target, phase):
		cause = fmt.Errorf("soak %s state drift", soakErrorText(phase))
	default:
		return nil
	}
	record := soakFailureRecord{Event: "vellum.soak.interval-failure", Phase: soakErrorText(phase), Attempt: attempt, Stage: stage, Cause: soakErrorText(cause.Error()), QuietStartUTC: soakFailureTime(t0), QuietEndUTC: soakFailureTime(t1), Start: soakFailureEndpointValue(start, phase), End: soakFailureEndpointValue(end, phase)}
	encoded, err := json.Marshal(record)
	if err != nil {
		return errors.Join(cause, fmt.Errorf("soak failure record: %w", err))
	}
	if len(encoded) > 64<<10 {
		return errors.Join(cause, errors.New("soak failure record exceeds64KiB"))
	}
	return errors.Join(cause, emit(record))
}
