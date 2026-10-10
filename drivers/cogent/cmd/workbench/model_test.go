package main

import (
	"github.com/dshills/vellum/virtual"
	"strings"
	"testing"
)

func TestSnapshotAndTypedMatcher(t *testing.T) {
	s := newParts(100000)
	snapshot := s.Snapshot()
	s.parts[0].Value += 100
	if snapshot.parts[0].Value == s.parts[0].Value {
		t.Fatal("mutable worker snapshot")
	}
	for _, q := range []string{"r0000001", "0603", "ω", "90°", "not-present"} {
		for _, i := range []int{0, 1, 50, 99999} {
			expected := strings.Contains(strings.ToLower(strings.Join(s.Row(i).Cells, " ")), q)
			if s.Matches(i, q) != expected {
				t.Fatal(q, i)
			}
		}
	}
}
func BenchmarkActualMillionViewport(b *testing.B) {
	source := newParts(1000000)
	table, err := virtual.NewTable(source, 20)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		table.SetViewport(i%999980, 20)
		table.View()
	}
}
func BenchmarkTypedMatcher(b *testing.B) {
	source := newParts(1000000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.Matches(i%1000000, "0603")
	}
}
