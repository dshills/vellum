//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"cogentcore.org/core/system/driver/desktop"
	"context"
	"errors"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"image"
	"math"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"
)

func TestSoakRejectsImpossibleWindowsAndUnboundedDeadlines(t *testing.T) {
	c := config{cycles: 1, every: 1, composed: true, timeout: time.Minute + 40*time.Second, soak: soakConfig{Duration: 40 * time.Second, Quiet: time.Second, Settle: time.Second, Interval: 2 * time.Second}}
	if e := c.validate(); e != nil {
		t.Fatal(e)
	}
	for _, mutate := range []func(*config){func(c *config) { c.composed = false }, func(c *config) { c.cycles = 1000 }, func(c *config) { c.every = 2 }, func(c *config) { c.soak.Duration = 13 * time.Hour }, func(c *config) { c.soak.Quiet = 0 }, func(c *config) { c.soak.Interval = time.Second }, func(c *config) { c.timeout = c.soak.Duration }, func(c *config) { c.timeout = 24 * time.Hour }, func(c *config) { c.heapProfile = "relative.heap" }, func(c *config) { c.soak.Quiet = time.Duration(1<<63 - 1) }} {
		bad := c
		mutate(&bad)
		if bad.validate() == nil {
			t.Fatal("invalid soak admitted", bad)
		}
	}
}

func TestSoakErrorTextPreservesBoundedUTF8Prefix(t *testing.T) {
	for _, pair := range [][2]string{
		{"", ""}, {"context deadline exceeded", "context deadline exceeded"},
		{strings.Repeat("x", 513), strings.Repeat("x", 512)},
		{strings.Repeat("a", 511) + "é", strings.Repeat("a", 511)},
		{strings.Repeat("é", 257), strings.Repeat("é", 256)},
		{strings.Repeat("a", 510) + "🧭", strings.Repeat("a", 510)},
		{"bad\xffwire", "badwire"},
	} {
		got := soakErrorText(pair[0])
		if got != pair[1] || len(got) > 512 || !utf8.ValidString(got) {
			t.Fatalf("bounded error text: got %q, want %q", got, pair[1])
		}
	}
}
func TestSoakStateRequiresObservedVisibilityAndOcclusion(t *testing.T) {
	s := cogent.NativeServiceWindowCheck{Visible: true, OcclusionVisible: true, Frame: [4]float64{0, 0, 100, 100}}
	if !soakState(s, "visible") || soakState(s, "occluded") || soakState(s, "minimized") {
		t.Fatal("visible state confused")
	}
	s.OcclusionVisible = false
	if !soakState(s, "occluded") || soakState(s, "visible") {
		t.Fatal("occlusion confused")
	}
	s.Visible = false
	s.Miniaturized = true
	if !soakState(s, "minimized") {
		t.Fatal("minimized refused")
	}
	s.Key = true
	if soakState(s, "minimized") {
		t.Fatal("focused primary qualified")
	}
}
func TestSoakWaitCancellationIsImmediate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if e := soakSleep(ctx, time.Hour); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
}

func TestSoakWorkerResultSurfacesFailure(t *testing.T) {
	marker := errors.New("worker failed")
	if !errors.Is(soakResult(133693440, marker), marker) {
		t.Fatal("worker error lost")
	}
	if soakResult(0, nil) == nil {
		t.Fatal("wrong checksum accepted")
	}
	if e := soakResult(133693440, nil); e != nil {
		t.Fatal(e)
	}
}

func TestSoakRejectsPhaseWithoutRoomForOneQuietSample(t *testing.T) {
	c := config{cycles: 1, every: 1, composed: true, timeout: time.Minute + 40*time.Second, soak: soakConfig{Duration: 40 * time.Second, Quiet: 5 * time.Second, Settle: 5 * time.Second, Interval: 10 * time.Second}}
	if c.validate() == nil {
		t.Fatal("phase equal to settle plus quiet admits no actual sample")
	}
}

func TestSoakRejectsNonfiniteOrEmptyObservedFrames(t *testing.T) {
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		for index := range 4 {
			s := cogent.NativeServiceWindowCheck{Visible: true, OcclusionVisible: true, Frame: [4]float64{0, 0, 100, 100}}
			s.Frame[index] = value
			if soakState(s, "visible") {
				t.Fatalf("nonfinite coordinate %d qualified", index)
			}
		}
	}
	for _, phase := range []string{"visible", "occluded", "minimized", "background", "unknown"} {
		if soakState(cogent.NativeServiceWindowCheck{}, phase) {
			t.Fatalf("empty window qualified %s", phase)
		}
	}
}

func TestSoakIntervalSeparatesCPUFromMeasurementEligibility(t *testing.T) {
	for _, c := range []struct {
		name                        string
		requested                   time.Duration
		elapsed, wall, cpu, percent float64
		eligible                    bool
	}{
		{"sixty seconds", time.Minute, 60, 60, 0.15, 0.25, true},
		{"parallel CPU", time.Minute, 60, 60, 90, 150, true},
		{"short preflight", time.Second, 1, 1, 0.005, 0.5, false},
		{"underslept", time.Minute, 59, 59, 0.1475, 0.25, false},
		{"overslept", time.Minute, 67, 67, 0.1675, 0.25, false},
		{"wall clock shift", time.Minute, 60, 120, 0.15, 0.25, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			eligible, percent, err := soakInterval(c.requested, c.elapsed, c.wall, c.cpu)
			if err != nil || eligible != c.eligible || math.Abs(percent-c.percent) > 1e-9 {
				t.Fatal(eligible, percent, err)
			}
		})
	}
	for _, values := range [][3]float64{{0, 60, 1}, {60, 0, 1}, {-1, 60, 1}, {60, -1, 1}, {60, 60, -1}, {math.NaN(), 60, 1}, {60, math.Inf(1), 1}, {60, 60, math.NaN()}, {math.SmallestNonzeroFloat64, 60, math.MaxFloat64}} {
		eligible, _, err := soakInterval(time.Minute, values[0], values[1], values[2])
		if err == nil || eligible {
			t.Fatal("invalid CPU measurement qualified", values, err)
		}
	}
}

func TestSoakWorkCancellationAndChecksum(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := soakWork(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("pre-canceled worker did not refuse", err)
	}
	value, err := soakWork(context.Background())
	// Multiplication by odd31 permutes each byte-valued block; 1MiB has4096 blocks.
	var block uint64
	for v := uint64(0); v < 256; v++ {
		block += v
	}
	if err != nil || value != block*4096 {
		t.Fatal("allocated worker checksum", value, err)
	}
	if err = soakResult(value, nil); err != nil {
		t.Fatal(err)
	}
}

func TestSoakGeometryReadSharesCallbackMutex(t *testing.T) {
	// This is a backend record, with no GLFW/Cocoa window or native main loop.
	window := &desktop.Window{}
	window.WnSize = image.Pt(1, 1)
	start := make(chan struct{})
	var done sync.WaitGroup
	done.Add(1)
	go func() {
		defer done.Done()
		<-start
		for i := 1; i <= 10000; i++ {
			window.Mu.Lock()
			window.WnSize = image.Pt(i, i)
			window.Mu.Unlock()
		}
	}()
	close(start)
	for range 10000 {
		size := soakWindowSize(window)
		if size.X < 1 || size.X != size.Y {
			t.Errorf("torn callback geometry: %v", size)
			break
		}
	}
	done.Wait()
	// Returning with Mu still held would strand native dispatch/resize callbacks.
	if !window.Mu.TryLock() {
		t.Fatal("geometry copy retained native mutex")
	}
	window.Mu.Unlock()
}

func TestSoakPairRefusesAmbiguousOrNonMainObservations(t *testing.T) {
	target := cogent.NativeServiceWindowCheck{Title: "target", Frame: [4]float64{10, 20, 100, 80}}
	cover := cogent.NativeServiceWindowCheck{Title: "cover", Frame: [4]float64{5, 15, 120, 100}}
	s := cogent.NativeServiceCheckState{MainThread: true, Windows: []cogent.NativeServiceWindowCheck{target, cover}}
	a, b, err := soakPairState(s, "target", "cover")
	if err != nil || a != target || b != cover {
		t.Fatal("pair did not preserve native DTOs", a, b, err)
	}
	for _, bad := range []cogent.NativeServiceCheckState{
		{Windows: s.Windows},
		{MainThread: true, Windows: []cogent.NativeServiceWindowCheck{target}},
		{MainThread: true, Windows: []cogent.NativeServiceWindowCheck{cover}},
		{MainThread: true, Windows: []cogent.NativeServiceWindowCheck{target, cover, target}},
		{MainThread: true, Windows: []cogent.NativeServiceWindowCheck{target, cover, cover}},
	} {
		if _, _, err := soakPairState(bad, "target", "cover"); err == nil {
			t.Fatal("ambiguous or non-main pair accepted", bad)
		}
	}
	if _, _, err := soakPairState(s, "target", "target"); err == nil {
		t.Fatal("identical target and cover accepted")
	}
}

func TestSoakContainmentDoesNotQualifyOcclusion(t *testing.T) {
	inner := [4]float64{-100, -50, 200, 150}
	outer := [4]float64{-110, -60, 220, 170}
	if !soakContains(outer, inner) || !soakContains(inner, inner) {
		t.Fatal("full outer frame containment refused")
	}
	for _, frame := range [][4]float64{
		{-99, -50, 200, 150}, {-100, -49, 200, 150}, {-100, -50, 199, 150}, {-100, -50, 200, 149},
	} {
		if soakContains(frame, inner) {
			t.Fatal("partial containment accepted", frame)
		}
	}
	for _, frame := range [][4]float64{
		{math.NaN(), 0, 100, 100}, {0, 0, math.Inf(1), 100}, {math.MaxFloat64, 0, math.MaxFloat64, 100},
		{0, 0, 0, 100}, {0, 0, 100, -1},
	} {
		if soakContains(frame, inner) || soakContains(outer, frame) {
			t.Fatal("invalid containment accepted", frame)
		}
	}
	target := cogent.NativeServiceWindowCheck{Visible: true, OcclusionVisible: true, Frame: inner}
	if soakState(target, "occluded") {
		t.Fatal("covered frame substituted for effective native occlusion")
	}
}

func TestSoakCoverRequiresObservedGeometryOrderingAndOpacity(t *testing.T) {
	target := cogent.NativeServiceWindowCheck{Frame: [4]float64{10, 10, 100, 100}, OrderedIndex: 1}
	cover := cogent.NativeServiceWindowCheck{Visible: true, Opaque: true, Alpha: 1, Frame: [4]float64{0, 0, 120, 120}, OrderedIndex: 0}
	if !soakCoverState(target, cover) {
		t.Fatal("valid native cover refused")
	}
	for _, mutate := range []func(*cogent.NativeServiceWindowCheck){
		func(w *cogent.NativeServiceWindowCheck) { w.Visible = false },
		func(w *cogent.NativeServiceWindowCheck) { w.Miniaturized = true },
		func(w *cogent.NativeServiceWindowCheck) { w.Opaque = false },
		func(w *cogent.NativeServiceWindowCheck) { w.Alpha = 0.5 },
		func(w *cogent.NativeServiceWindowCheck) { w.Level = 1 },
		func(w *cogent.NativeServiceWindowCheck) { w.OrderedIndex = -1 },
		func(w *cogent.NativeServiceWindowCheck) { w.OrderedIndex = 1 },
		func(w *cogent.NativeServiceWindowCheck) { w.OrderedIndex = 2 },
		func(w *cogent.NativeServiceWindowCheck) { w.Frame[2] = 100 },
	} {
		bad := cover
		mutate(&bad)
		if soakCoverState(target, bad) {
			t.Fatal("invalid native cover accepted", bad)
		}
	}
}
