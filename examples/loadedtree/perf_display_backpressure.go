// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
package main

import (
	"errors"
	"time"
)

// A zero value preserves the existing workload and journal. Explicit nonzero
// admission is confined to the native presentation experiment, with unchanged
// input, prefix and timing validation.
func validatePerfDisplayBackpressure(perf bool, presentation string, timeout time.Duration) error {
	if timeout == 0 {
		return nil
	}
	if !perf || presentation == "" || timeout < 100*time.Millisecond || timeout > 5*time.Second {
		return errors.New("experimental display backpressure: requires -perf, -perf-presentation and timeout 100ms..5s")
	}
	return nil
}
