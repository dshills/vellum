package cogent

import (
	"cogentcore.org/core/core"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"strings"
	"sync"
	"testing"
)

func TestGridInputDetailsBoundedDetachedAndConcurrent(t *testing.T) {
	for _, n := range []int{0, -1, 4097} {
		if _, err := NewFrameMetricsWithGridInputDetails(n); !errors.Is(err, ErrFrameMetricsConfig) {
			t.Fatalf("capacity%d: %v", n, err)
		}
	}
	ordinary, _ := NewFrameMetrics(1)
	if ordinary.gridInputs != nil || ordinary.GridInputDetails().Capacity != 0 {
		t.Fatal("ordinary recorder allocated input ledger")
	}
	m, err := NewFrameMetricsWithGridInputDetails(1)
	if err != nil {
		t.Fatal(err)
	}
	b := &gridInputBinding{recorder: m.gridInputs, window: "one", grid: "g", incarnation: 7}
	var workers sync.WaitGroup
	for i := 0; i < 4; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for n := 0; n < 100; n++ {
				b.record(GridInputRecord{Kind: "wheel", Tag: 2})
				_ = m.GridInputDetails()
			}
		}()
	}
	workers.Wait()
	s := m.GridInputDetails()
	if s.Capacity != 259 || len(s.Records) != 259 || s.Dropped != 141 || s.Failure == "" {
		t.Fatalf("lost bounded evidence%+v", s)
	}
	for i, r := range s.Records {
		if r.Sequence != uint64(i+1) || r.Window != "one" || r.Grid != "g" || r.Incarnation != 7 {
			t.Fatalf("invalid binding/sequence%+v", r)
		}
	}
	s.Records[0].Tag = 99
	if m.GridInputDetails().Records[0].Tag != 2 {
		t.Fatal("snapshot aliases recorder")
	}
	if (&FrameMetrics{}).GridInputDetails().Failure == "" {
		t.Fatal("invalid recorder accepted")
	}
}

func TestGridInputDetailsRefusesUnboundedIdentifierRetention(t *testing.T) {
	m, _ := NewFrameMetricsWithGridInputDetails(1)
	a := &app.Application{}
	w := &nativeWindow{window: app.Window{ID: "input-id"}, body: core.NewBody("input-id")}
	b, err := m.bind(a, w)
	if err != nil {
		t.Fatal(err)
	}
	defer b.retire()
	g := &gridBinding{owner: a, window: "input-id", r: &record{node: ui.Node{ID: strings.Repeat("x", 4097)}}}
	if bindGridInputDetails(g) != nil || m.GridInputDetails().Failure == "" {
		t.Fatal("retained unbounded descriptor identifier")
	}
	ordinary, _ := NewFrameMetrics(1)
	if ordinary.GridInputDetails().Failure != "" {
		t.Fatal("diagnostic refusal altered ordinary metrics")
	}
}
