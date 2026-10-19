package main

import "testing"

func TestPerfPresentationSeparateMode(t *testing.T) {
	c := perfConfig{Rows: 10000, Samples: 120, Warmup: 128, PresentationDetails: "new.json"}
	if err := c.validate(); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*perfConfig){func(c *perfConfig) { c.SourceDetails = true }, func(c *perfConfig) { c.InputDetails = "input.json" }, func(c *perfConfig) { c.CPUProfile = "cpu.pprof" }, func(c *perfConfig) { c.AllocationProfile = "alloc" }} {
		invalid := c
		mutate(&invalid)
		if invalid.validate() == nil {
			t.Fatal("combined diagnostics")
		}
	}
	if perfPresentationStatus("") != "unavailable" || perfPresentationStatus("new.json") == "unavailable" {
		t.Fatal("ordinary/diagnostic output")
	}
}
