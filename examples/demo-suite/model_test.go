package main

import (
	"github.com/dshills/vellum/virtual"
	"testing"
)

func TestMillionRowSource(t *testing.T) {
	s := rows{}
	for _, i := range []int{0, 499999, 999999} {
		k, e := s.RowKey(i)
		if e != nil {
			t.Fatal(e)
		}
		got, found, e := s.IndexOf(k)
		if e != nil || !found || got != i {
			t.Fatalf("row identity %d: %d %v %v", i, got, found, e)
		}
		for _, c := range []string{"part", "value", "package", "status"} {
			text, e := s.Cell(k, c)
			if e != nil || text == "" {
				t.Fatalf("cell %s/%s: %q %v", k, c, text, e)
			}
		}
	}
	if _, e := s.RowKey(s.RowCount()); e != virtual.ErrGridRange {
		t.Fatalf("out of range: %v", e)
	}
}
