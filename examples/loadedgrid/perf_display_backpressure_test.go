// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDisplayBackpressureExplicitPresentationMode(t *testing.T) {
	if err := validatePerfDisplayBackpressure(false, "", 0); err != nil {
		t.Fatal(err)
	}
	for _, timeout := range []time.Duration{100 * time.Millisecond, time.Second, 5 * time.Second} {
		c := perfConfig{Rows: 10000, Samples: 120, Warmup: 128}
		c.PresentationDetails = "new.json"
		c.DisplayBackpressure = timeout
		if err := c.validate(); err != nil {
			t.Fatal(err)
		}
		c.PresentationDetails = ""
		if c.validate() == nil {
			t.Fatal("display experiment without capture")
		}
		if validatePerfDisplayBackpressure(false, "new.json", timeout) == nil {
			t.Fatal("display experiment without perf")
		}
	}
	for _, timeout := range []time.Duration{-time.Second, 1, 100*time.Millisecond - 1, 5*time.Second + 1} {
		c := perfConfig{Rows: 10000, Samples: 120, Warmup: 128}
		c.PresentationDetails = "new.json"
		c.DisplayBackpressure = timeout
		if c.validate() == nil {
			t.Fatal("invalid display timeout", timeout)
		}
	}
}
func TestDisplayBackpressureJournalOmissionAndExactDuration(t *testing.T) {
	// Default JSON stays byte-identical to the existing Config schema.
	c := presentationCaptureConfig{Method: "routed-grid-scroll", Samples: 120, Warmup: 128, DataSize: 10000}
	before, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err = json.Unmarshal(before, &decoded); err != nil {
		t.Fatal(err)
	}
	if _, ok := decoded["DisplayBackpressureNS"]; ok {
		t.Fatal("default experiment declaration emitted")
	}
	c.DisplayBackpressureNS = int64(time.Second)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["DisplayBackpressureNS"] != float64(time.Second) {
		t.Fatal("duration declaration differs")
	}
}
