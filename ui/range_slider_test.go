package ui

import (
	"math"
	"testing"
)

func TestRangeEndpointMergeSnapAndCrossingClamp(t *testing.T) {
	bounds := Numeric{Min: -10, Max: 10, Step: .5}
	v := RangeValue{Lower: -5, Upper: 5}
	v, e := (RangeEvent{Handle: RangeLower, Value: 2.24}).Apply(v, bounds)
	if e != nil || v.Lower != 2 || v.Upper != 5 {
		t.Fatal(v, e)
	}
	v, e = (RangeEvent{Handle: RangeUpper, Value: -4}).Apply(v, bounds)
	if e != nil || v.Upper != 2 || v.Lower != 2 {
		t.Fatal(v, e)
	}
	if _, e = (RangeEvent{Handle: RangeLower, Value: float32(math.Inf(1))}).Apply(v, bounds); e == nil {
		t.Fatal("Inf")
	}
	if _, e = (RangeEvent{Handle: "bad", Value: 1}).Apply(v, bounds); e == nil {
		t.Fatal("handle")
	}
}
func TestRangeCurrentHandleQueueAndDisabledEpoch(t *testing.T) {
	r := &RangeSlider{ID: "range", Label: "Range", Numeric: Numeric{Min: 0, Max: 100, Step: 1}, Value: RangeValue{Lower: 25, Upper: 75}, TickCount: 5}
	calls := 0
	r.Change = func(e RangeEvent) {
		v, err := e.Apply(r.Value, r.Numeric)
		if err != nil {
			t.Error(err)
			return
		}
		r.Value = v
		calls++
	}
	n, e := r.View()
	if e != nil {
		t.Fatal(e)
	}
	old := pickerNode(n, RangeTrackID(r.ID)).Track
	old.Select(string(RangeUpper))
	old.Adjust(old.Current(), 1)
	old.Adjust(old.Current(), 1)
	if r.Value != (RangeValue{Lower: 25, Upper: 77}) || calls != 2 {
		t.Fatal(r.Value, calls)
	}
	r.Disabled = true
	_, _ = r.View()
	r.Disabled = false
	_, _ = r.View()
	old.Set(string(RangeLower), .9)
	if calls != 2 {
		t.Fatal("revived stale callback")
	}
	r.Value = RangeValue{Lower: 10, Upper: 10}
	if _, e = r.View(); e != nil {
		t.Fatal("equal endpoints", e)
	}
}
