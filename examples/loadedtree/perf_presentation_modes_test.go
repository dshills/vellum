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
