package viewport

import "math"

// Gesture tracks one primary pointer sequence. Coordinates and scale must use
// matching units; the threshold is six logical pixels. Sequence is supplied by
// the input adapter. Stale cancellation cannot revoke a newer gesture. Reset
// when replacing the adapter/mount, whose sequence numbers may start again.
type Gesture struct {
	sequence                            uint64
	down, moved                         bool
	startX, startY, lastX, lastY, scale float64
}

func (g *Gesture) Begin(sequence uint64, x, y, scale float64) bool {
	if g == nil || sequence == 0 || sequence <= g.sequence || !finite(x) || !finite(y) || !finite(scale) || scale <= 0 {
		return false
	}
	*g = Gesture{sequence: sequence, down: true, startX: x, startY: y, lastX: x, lastY: y, scale: scale}
	return true
}
func (g *Gesture) Move(sequence uint64, x, y float64) (dx, dy float64, ok bool) {
	if g == nil || !g.down || sequence != g.sequence || !finite(x) || !finite(y) {
		return
	}
	dx, dy = x-g.lastX, y-g.lastY
	g.lastX, g.lastY = x, y
	g.moved = g.moved || math.Hypot(x-g.startX, y-g.startY) >= 6*g.scale
	return dx, dy, true
}

// End returns one click only for an inside release below the movement threshold.
// A second release or stale sequence is a no-op. Movement may still update a
// camera continuously; this click result is for selection/command admission.
func (g *Gesture) End(sequence uint64, x, y float64, inside bool) bool {
	if _, _, ok := g.Move(sequence, x, y); !ok {
		return false
	}
	click := inside && !g.moved
	g.down = false
	return click
}
func (g *Gesture) Cancel(sequence uint64) bool {
	if g == nil || !g.down || g.sequence != sequence {
		return false
	}
	g.down = false
	return true
}
func (g *Gesture) Reset() {
	if g != nil {
		*g = Gesture{}
	}
}
func (g *Gesture) Active() bool { return g != nil && g.down }

// Dragging reports whether the active gesture crossed the logical threshold.
func (g *Gesture) Dragging() bool { return g != nil && g.down && g.moved }
