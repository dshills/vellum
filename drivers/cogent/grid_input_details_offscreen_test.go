//go:build offscreen

package cogent

import (
	"image"
	"math"
	"testing"

	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/ui"
)

// Observe real untraced fractional delivery followed by reserved FIFO input.
// This is an event/owner proof, never a software/native timing qualification.
func TestGridInputDetailsActualFractionalAndTaggedRouting(t *testing.T) {
	grid, _ := testNativeGrid(t, 1000)
	notifications := 0
	m, a := gridManual(t, manualConfig(t, 1), func() ui.Node {
		return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: grid, Grow: true, GridChange: func(ui.GridEvent) { notifications++ }}
	})
	gridFrame(t, m)
	metrics, err := NewFrameMetricsWithGridInputDetails(2)
	if err != nil {
		t.Fatal(err)
	}
	w := m.implementation.find("one").state
	b, err := metrics.bind(a, w)
	if err != nil {
		t.Fatal(err)
	}
	defer b.retire()
	g := w.root.control.grid
	if g == nil {
		t.Fatal("missing grid binding")
	}
	// Manual binds its observer after construction; desktop binds before mount.
	g.inputs = bindGridInputDetails(g)
	first := gridInspection(t, a, grid)
	position := image.Pt((first.Bounds.Min.X+first.Bounds.Max.X)/2, (first.Bounds.Min.Y+first.Bounds.Max.Y)/2)
	notifications = 0 // Initial viewport notification precedes these inputs.
	for i := 0; i < 4; i++ {
		w.source.Load().Scroll(position, math32.Vector2{Y: .25 * first.Scale})
		gridFrame(t, m)
	}
	var enqueue error
	if err := a.Do(func() { enqueue = EnqueueGridScroll(testContext, a, grid, 0, 12.25, 42) }); err != nil || enqueue != nil {
		t.Fatal(err, enqueue)
	}
	gridFrame(t, m)
	b.trace.OwnerWork(42, 0)
	s := metrics.GridInputDetails()
	if s.Failure != "" || s.Dropped != 0 {
		t.Fatal(s)
	}
	var wheels, transitions []GridInputRecord
	var admission GridInputRecord
	for _, r := range s.Records {
		switch r.Kind {
		case "wheel":
			wheels = append(wheels, r)
		case "enqueue":
			admission = r
		case "transition":
			if r.Origin == "wheel" {
				transitions = append(transitions, r)
			}
		}
	}
	if len(wheels) != 5 || len(transitions) != 5 || notifications != 2 {
		t.Fatalf("missing actual event/owner evidence: wheels%d transitions%d callbacks%d", len(wheels), len(transitions), notifications)
	}
	for i, r := range wheels[:4] {
		if r.Tag != 0 || r.Scale != first.Scale || r.RemainderBefore[1] != float64(i)*.25 || r.RemainderAfter[1] != float64((i+1)%4)*.25 || r.DeltaY != i/3 {
			t.Fatalf("fractional input%d %+v", i, r)
		}
		if transitions[i].Input != r.Sequence || transitions[i].Tag != 0 {
			t.Fatal("untraced correlation lost")
		}
	}
	last := wheels[4]
	if admission.Tag != 42 || admission.RequestedY != 12.25 || admission.Scale != first.Scale || admission.PhysicalY != last.PhysicalY || last.Tag != 42 || last.DeltaY != 12 || math.Abs(last.RemainderAfter[1]-.25) > 1e-6 || transitions[4].Input != last.Sequence || transitions[4].Tag != 42 || transitions[4].ViewportBefore.ScrollY != 1 || transitions[4].ViewportAfter.ScrollY != 13 {
		t.Fatalf("tagged scalar provenance differs: %+v %+v %+v", admission, last, transitions[4])
	}
}
