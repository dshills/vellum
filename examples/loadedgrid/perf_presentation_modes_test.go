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

func TestPacedModeRequiresBoundedSeparateJournal(t *testing.T) {
	c := perfConfig{Rows: 10000, Samples: 600, Warmup: 128, Paced: true, PresentationDetails: "new-paced.json"}
	if err := c.validate(); err != nil {
		t.Fatal(err)
	}
	c.PresentationDetails = ""
	if c.validate() == nil {
		t.Fatal("paced mode without journal")
	}
	c.PresentationDetails = "new.json"
	c.Samples = 599
	if c.validate() == nil {
		t.Fatal("short paced cohort")
	}
	c.Samples = 2049
	if c.validate() == nil {
		t.Fatal("oversized paced cohort")
	}
}

func TestPacedGridDeclaredTrajectoryOffsets(t *testing.T) {
	expected := map[int][2]int{7: {96, 96}, 15: {192, 192}, 23: {288, 288}, 31: {384, 384}, 39: {288, 288}, 47: {192, 192}, 55: {96, 96}, 63: {0, 0}}
	x, y := 0, 0
	for i := 0; i < 64; i++ {
		dx, dy := perfDelta(i)
		x += dx
		y += dy
		if i < 7 && x != 0 {
			t.Fatal("early horizontal movement", i, x)
		}
		if want, ok := expected[i]; ok && [2]int{x, y} != want {
			t.Fatal(i, x, y, want)
		}
	}
}
