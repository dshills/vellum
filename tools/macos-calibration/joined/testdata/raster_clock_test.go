package rasterx

import (
	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint/pimage"
	"cogentcore.org/core/paint/render"
	"image"
	"image/color"
	"image/draw"
	"math"
	"sync"
	"sync/atomic"
	"testing"
)

func joinedClock(t *testing.T) {
	t.Helper()
	if !VellumRasterReset() {
		t.Fatal("pending previous capture")
	}
	var ticks atomic.Uint64
	if !VellumRasterConfigure(func() float64 { return float64(ticks.Add(1)) }) {
		t.Fatal("configure")
	}
	t.Cleanup(func() {
		if !VellumRasterReset() {
			t.Error("unretired clock")
		}
	})
}
func TestVellumJoinedRasterPixelsAndDetachedSnapshot(t *testing.T) {
	joinedClock(t)
	rs := New(math32.Vec2(17, 13))
	cmd := &pimage.Params{Cmd: pimage.Draw, Rect: image.Rect(0, 0, 17, 13), Source: image.NewUniform(color.RGBA{20, 30, 40, 255}), Op: draw.Src}
	rs.Render(render.Render{cmd})
	s := VellumRasterRead()
	if s.Invalid || s.Overflow || s.Pending != 0 || len(s.Records) != 1 || s.Records[0].Begin != 1 || s.Records[0].End != 2 || !s.Records[0].Returned || !s.Records[0].Complete || s.Records[0].Target != image.Rect(0, 0, 17, 13) {
		t.Fatalf("capture %+v", s)
	}
	for y := range 13 {
		for x := range 17 {
			if rs.Image().(*image.RGBA).RGBAAt(x, y) != (color.RGBA{20, 30, 40, 255}) {
				t.Fatal("pixel changed")
			}
		}
	}
	s.Records[0].Begin = 999
	if VellumRasterRead().Records[0].Begin != 1 {
		t.Fatal("snapshot alias")
	}
	if VellumRasterConfigure(func() float64 { return 1 }) {
		t.Fatal("reconfigure accepted")
	}
}
func TestVellumJoinedRasterCapsAndPendingReset(t *testing.T) {
	joinedClock(t)
	i := vellumRasterBegin(nil)
	if VellumRasterReset() || VellumRasterRead().Pending != 1 {
		t.Fatal("reset erased pending")
	}
	vellumRasterEnd(i, true)
	for range vellumRasterCapacity - 1 {
		i := vellumRasterBegin(nil)
		vellumRasterEnd(i, true)
	}
	if vellumRasterBegin(nil) != -1 {
		t.Fatal("overflow admitted")
	}
	s := VellumRasterRead()
	if !s.Overflow || s.Invalid || s.Pending != 0 || len(s.Records) != vellumRasterCapacity {
		t.Fatalf("bad bound %+v", s)
	}
}
func TestVellumJoinedRasterInvalidAndAbnormalBrackets(t *testing.T) {
	for _, clock := range []func() float64{func() float64 { return math.NaN() }, func() float64 { return 0 }, func() float64 { return 1 }} {
		if !VellumRasterReset() || !VellumRasterConfigure(clock) {
			t.Fatal("configure")
		}
		i := vellumRasterBegin(nil)
		vellumRasterEnd(i, false)
		s := VellumRasterRead()
		if !s.Invalid || s.Pending != 0 || !s.Records[0].Complete || s.Records[0].Returned {
			t.Fatalf("invalid/abnormal accepted %+v", s)
		}
	}
	joinedClock(t)
	i := vellumRasterBegin(nil)
	vellumRasterEnd(i, true)
	vellumRasterEnd(i, true)
	if !VellumRasterRead().Invalid {
		t.Fatal("duplicate end accepted")
	}
}
func TestVellumJoinedRasterConcurrentReservation(t *testing.T) {
	joinedClock(t)
	var wg sync.WaitGroup
	for range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 4 {
				i := vellumRasterBegin(nil)
				vellumRasterEnd(i, true)
			}
		}()
	}
	wg.Wait()
	s := VellumRasterRead()
	if s.Pending != 0 || s.Invalid || s.Overflow || len(s.Records) != 128 {
		t.Fatalf("concurrent capture %+v", s)
	}
	for i, r := range s.Records {
		if r.Ordinal != i || !r.Complete || !r.Returned || r.End <= r.Begin {
			t.Fatal("bad reservation")
		}
	}
}

func TestVellumJoinedRasterUnpublishedBeginRefusesEnd(t *testing.T) {
	if !VellumRasterReset() {
		t.Fatal("pending previous trace")
	}
	entered, resume := make(chan struct{}), make(chan struct{})
	var calls atomic.Uint64
	if !VellumRasterConfigure(func() float64 {
		n := calls.Add(1)
		if n == 1 {
			close(entered)
			<-resume
		}
		return float64(n)
	}) {
		t.Fatal("configure")
	}
	token := make(chan int, 1)
	go func() { token <- vellumRasterBegin(nil) }()
	<-entered
	// Fault injection: real Render cannot have the token before Begin returns.
	vellumRasterEnd(0, true)
	s := VellumRasterRead()
	if !s.Invalid || s.Pending != 1 || s.Records[0].Begun || s.Records[0].Complete || VellumRasterReset() {
		t.Fatal("unpublished begin accepted or erased")
	}
	close(resume)
	i := <-token
	vellumRasterEnd(i, true)
	s = VellumRasterRead()
	if !s.Invalid || s.Pending != 0 || !s.Records[0].Begun || !s.Records[0].Complete || !s.Records[0].Returned {
		t.Fatal("reservation did not retire")
	}
	if !VellumRasterReset() {
		t.Fatal("reset")
	}
}
