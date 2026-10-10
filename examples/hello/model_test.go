package main

import (
	"example.com/vellum/ui"
	"testing"
)

func TestCounterActionPreservesIDs(t *testing.T) {
	c := &counter{}
	content := ui.Node{ID: "body", Kind: ui.Label, Text: "Ready"}
	before := c.view(content)
	if err := ui.Validate(before); err != nil {
		t.Fatal(err)
	}
	before.Children[2].Activate()
	after := c.view(content)
	if after.Children[1].Text != "Count: 1" {
		t.Fatal(after.Children[1].Text)
	}
	for i := range before.Children {
		if before.Children[i].ID != after.Children[i].ID {
			t.Fatal("unstable widget IDs")
		}
	}
}
func BenchmarkCounterViewValidation(b *testing.B) {
	c := &counter{}
	content := ui.Node{ID: "body", Kind: ui.Label, Text: "Ready"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := ui.Validate(c.view(content)); err != nil {
			b.Fatal(err)
		}
	}
}
