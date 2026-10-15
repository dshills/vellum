//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"cogentcore.org/core/system/driver/desktop"
	"context"
	"errors"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"image"
	"math"
	"sync"
	"testing"
	"time"
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
