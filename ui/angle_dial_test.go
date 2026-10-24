package ui

import (
	"math"
	"testing"
)

func TestAngleDialWrappingSnappingAndQueuedEdits(t *testing.T) {
	for _, c := range []struct{ value, step, want float32 }{{360, 1, 0}, {-15, 15, 345}, {721, 1, 1}, {88, 15, 90}, {359, 15, 0}} {
		got, ok := NormalizeAngle(c.value, c.step)
		if !ok || got != c.want {
			t.Fatal(c, got, ok)
		}
	}
	for _, step := range []float32{7, .001, 181, float32(math.NaN())} {
		if _, ok := NormalizeAngle(0, step); ok {
			t.Fatal("invalid step", step)
		}
	}
	d := &AngleDial{ID: "rotation", Label: "Rotation", Degrees: 345, Step: 15}
	calls := 0
	d.Change = func(e AngleEvent) { calls++; d.Degrees = e.Degrees }
	n, err := d.View()
	if err != nil {
		t.Fatal(err)
	}
	dial := pickerNode(n, AngleDialButtonID(d.ID))
	dial.AngleDial.Adjust(1)
	dial.AngleDial.Adjust(1)
	if calls != 2 || d.Degrees != 15 {
		t.Fatal("queued deltas did not compose", calls, d.Degrees)
	}
	pickerNode(n, AngleDialFieldID(d.ID)).Change(ChangeEvent{Number: 360})
	if d.Degrees != 0 {
		t.Fatal("field failed to wrap")
	}
	d.Disabled = true
	dial.AngleDial.Set(180)
	if d.Degrees != 0 {
		t.Fatal("disabled input acted")
	}
	d.Disabled = false
	d.ID = "replacement"
	dial.Activate()
	if d.Degrees != 0 {
		t.Fatal("stale identity acted")
	}
}
