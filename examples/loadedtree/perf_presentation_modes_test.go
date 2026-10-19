package main

import "testing"

func TestTreePerfPresentationSeparateMode(t *testing.T) {
	c := treePerfConfig{Nodes: 10000, Samples: 120, Warmup: 128, PresentationDetails: "new.json"}
	if err := c.validate(); err != nil {
		t.Fatal(err)
	}
	c.SourceDetails = true
	if c.validate() == nil {
		t.Fatal("combined diagnostics")
	}
}

func TestPacedModeRequiresBoundedSeparateJournal(t *testing.T) {
	c := treePerfConfig{Nodes: 10000, Samples: 600, Warmup: 128, Paced: true, PresentationDetails: "new-paced.json"}
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
