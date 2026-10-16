//go:build darwin && cgo && vellum_native_checks

package main

import (
	"fmt"
	"math"
	"time"

	cogent "github.com/dshills/vellum/drivers/cogent"
)

// Deltas require two snapshots from the same live owned process. The controller
// must additionally match PID to its actual process receipt; this is not a
// cross-process replay/provenance validator or an OS wake/presentation measure.
type soakActivityObservation struct {
	Version                     int                        `json:"version"`
	Available                   bool                       `json:"available"`
	Start                       cogent.NativeActivityCheck `json:"start"`
	End                         cogent.NativeActivityCheck `json:"end"`
	StartRead                   soakActivityReadWindow     `json:"start_read"`
	EndRead                     soakActivityReadWindow     `json:"end_read"`
	EmptyEventPostAttemptsDelta uint64                     `json:"empty_event_post_attempts_delta"`
	SurfacePresentAttemptsDelta uint64                     `json:"surface_present_attempts_delta"`
}

type soakActivityReadWindow struct {
	StartUTC time.Time `json:"start_utc"`
	EndUTC   time.Time `json:"end_utc"`
}

type soakActivityBounds struct {
	StartRead, EndRead   soakActivityReadWindow
	QuietStart, QuietEnd time.Time
}

func soakActivityDelta(expectedPID int, before, after cogent.NativeActivityCheck, bounds soakActivityBounds) (soakActivityObservation, error) {
	// Validate the exact wall times published in JSON. UTC also strips any
	// monotonic readings, which could otherwise hide a wall-clock adjustment.
	bounds.StartRead = soakActivityReadWindow{bounds.StartRead.StartUTC.UTC(), bounds.StartRead.EndUTC.UTC()}
	bounds.EndRead = soakActivityReadWindow{bounds.EndRead.StartUTC.UTC(), bounds.EndRead.EndUTC.UTC()}
	bounds.QuietStart, bounds.QuietEnd = bounds.QuietStart.UTC(), bounds.QuietEnd.UTC()
	value := soakActivityObservation{Version: 1, Start: before, End: after,
		StartRead: bounds.StartRead, EndRead: bounds.EndRead}
	if !before.Available || !after.Available || expectedPID <= 0 || before.ProcessID != expectedPID || after.ProcessID != expectedPID {
		return value, fmt.Errorf("soak native API attempt counters unavailable or process scope changed")
	}
	if bounds.StartRead.StartUTC.IsZero() || bounds.StartRead.EndUTC.IsZero() || bounds.EndRead.StartUTC.IsZero() || bounds.EndRead.EndUTC.IsZero() || bounds.QuietStart.IsZero() || bounds.QuietEnd.IsZero() || bounds.StartRead.EndUTC.Before(bounds.StartRead.StartUTC) || bounds.QuietStart.Before(bounds.StartRead.EndUTC) || !bounds.QuietEnd.After(bounds.QuietStart) || bounds.EndRead.StartUTC.Before(bounds.QuietEnd) || bounds.EndRead.EndUTC.Before(bounds.EndRead.StartUTC) {
		return value, fmt.Errorf("soak native API attempt reads do not bracket the quiet interval")
	}
	if before.Saturated || after.Saturated || before.EmptyEventPostAttempts == math.MaxUint64 || before.SurfacePresentAttempts == math.MaxUint64 || after.EmptyEventPostAttempts == math.MaxUint64 || after.SurfacePresentAttempts == math.MaxUint64 {
		return value, fmt.Errorf("soak native API attempt counter saturated")
	}
	if after.EmptyEventPostAttempts < before.EmptyEventPostAttempts || after.SurfacePresentAttempts < before.SurfacePresentAttempts {
		return value, fmt.Errorf("soak native API attempt counters decreased")
	}
	value.Available = true
	value.EmptyEventPostAttemptsDelta = after.EmptyEventPostAttempts - before.EmptyEventPostAttempts
	value.SurfacePresentAttemptsDelta = after.SurfacePresentAttempts - before.SurfacePresentAttempts
	return value, nil
}
