package viewport

import "testing"

func TestGestureExactlyOnceThresholdAndStaleCancellation(t *testing.T) {
	var g Gesture
	if !g.Begin(1, 10, 20, 2) || !g.End(1, 12, 22, true) || g.End(1, 12, 22, true) {
		t.Fatal("click admission")
	}
	if !g.Begin(2, 10, 20, 2) {
		t.Fatal("begin")
	}
	g.Move(2, 22, 20)
	if g.End(2, 22, 20, true) {
		t.Fatal("drag became click")
	}
	if !g.Begin(3, 10, 20, 1) || g.Cancel(2) || !g.Active() {
		t.Fatal("stale cancel")
	}
	if !g.Cancel(3) || g.End(3, 10, 20, true) {
		t.Fatal("canceled release")
	}
	if !g.Begin(4, 0, 0, 1) || g.End(4, 0, 0, false) {
		t.Fatal("outside release")
	}
	g.Reset()
	if !g.Begin(1, 0, 0, 1) {
		t.Fatal("new mount")
	}
}
