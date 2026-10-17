package main

import (
	"fmt"
	"github.com/dshills/vellum/drivers/cogent"
	"image"
	"math"
	"testing"
	"time"
)

func TestCameraProjectsNearFarAndOrbit(t *testing.T) {
	c := defaultCamera()
	m := c.matrix(1)
	project := func(x, y, z float32) (float64, float64, float64) {
		v := [4]float32{x, y, z, 1}
		var p [4]float32
		for row := 0; row < 4; row++ {
			for col := 0; col < 4; col++ {
				p[row] += m[col*4+row] * v[col]
			}
		}
		return float64(p[0] / p[3]), float64(p[1] / p[3]), float64(p[2] / p[3])
	}
	x, y, near := project(0, 0, 4.9)
	_, _, far := project(0, 0, -95)
	if math.Abs(x) > 1e-6 || math.Abs(y) > 1e-6 || math.Abs(near) > 1e-4 || math.Abs(far-1) > 1e-5 {
		t.Fatalf("WebGPU depth [0,1]: x/y=%g/%g near/far=%g/%g", x, y, near, far)
	}
	_, _, front := project(0, 0, 1)
	_, _, back := project(0, 0, -1)
	if front >= back || front < 0 || back > 1 {
		t.Fatalf("cube depth ordering %g %g", front, back)
	}
	c.orbit(math.Pi/2, 0)
	m = c.matrix(1)
	x, _, _ = project(1, 0, 0)
	if math.Abs(x) > 1e-6 {
		t.Fatalf("quarter-turn face not centered: %g", x)
	}
	c.dolly(.0001)
	if c.distance != 2.5 {
		t.Fatal("near dolly limit")
	}
	c.dolly(1e9)
	if c.distance != 20 {
		t.Fatal("far dolly limit")
	}
	c.orbit(0, 100)
	if c.pitch != 1.4 {
		t.Fatal("pole singularity not bounded")
	}
}
func TestCubeDepthProbeSubmissionOrder(t *testing.T) {
	v, i := cube()
	if len(v) != 24*7 || len(i) != 36 {
		t.Fatalf("cube sizes %d %d", len(v), len(i))
	}
	for face := 0; face < 2; face++ {
		for vertex := 0; vertex < 4; vertex++ {
			p := v[(face*4+vertex)*7:]
			if face == 0 && (p[2] != 1 || p[3] != 1 || p[5] != 0) {
				t.Fatal("near-first red probe changed")
			}
			if face == 1 && (p[2] != -1 || p[3] != 0 || p[5] != 1) {
				t.Fatal("far-second blue probe changed")
			}
		}
	}
	for _, index := range i {
		if index >= 24 {
			t.Fatal("index outside geometry")
		}
	}
}

// Reproduce a provisional first target, then a later transition in the other
// window. Neither individual window's earlier quiet period admits the pair.
func TestInitialFramesRequireJointQuietInterval(t *testing.T) {
	base := time.Unix(1, 0)
	frame := cogent.GPUViewStats{Ready: true, Frames: 1, Prepares: 1, Draws: 1, Size: image.Pt(5102, 1330), TargetGeneration: 1, BufferAllocations: 3, HelperUploads: 3, HelperUploadBytes: 880}
	pair := [2]cogent.GPUViewStats{frame, frame}
	var stability initialFrameStability
	check := func(ms int, want bool) {
		t.Helper()
		if got := stability.observe(pair, base.Add(time.Duration(ms)*time.Millisecond)); got != want {
			t.Fatalf("at %dms ready=%v want %v", ms, got, want)
		}
	}
	check(0, false)
	pair[0].Size = image.Pt(792, 498)
	pair[0].Frames++
	pair[0].Prepares++
	pair[0].Draws++
	pair[0].TargetGeneration++
	pair[0].HelperUploads++
	pair[0].HelperUploadBytes += 64
	check(10, false)
	check(100, false)
	pair[1].Size = image.Pt(632, 412)
	pair[1].Frames++
	pair[1].TargetGeneration++
	check(105, false)
	check(204, false)
	check(205, true)
	// A resource transition also requires a fresh joint interval even if the
	// observed pixel dimensions and completed-frame count have not changed.
	pair[1].ResourceGeneration++
	pair[1].BufferAllocations += 3
	check(220, false)
	check(319, false)
	check(320, true)
}

func TestInitialFramesRejectInterruptedReadiness(t *testing.T) {
	for window := 0; window < 2; window++ {
		for _, invalid := range []string{"not-ready", "no-frame", "no-width", "no-height"} {
			t.Run(fmt.Sprintf("window%d/%s", window, invalid), func(t *testing.T) {
				base := time.Unix(1, 0)
				frame := cogent.GPUViewStats{Ready: true, Frames: 1, Size: image.Pt(640, 480)}
				pair := [2]cogent.GPUViewStats{frame, frame}
				var stability initialFrameStability
				stability.observe(pair, base)
				broken := pair
				switch invalid {
				case "not-ready":
					broken[window].Ready = false
				case "no-frame":
					broken[window].Frames = 0
				case "no-width":
					broken[window].Size.X = 0
				case "no-height":
					broken[window].Size.Y = 0
				}
				if stability.observe(broken, base.Add(99*time.Millisecond)) {
					t.Fatal("incomplete pair accepted")
				}
				if stability.observe(pair, base.Add(200*time.Millisecond)) || stability.observe(pair, base.Add(299*time.Millisecond)) {
					t.Fatal("readiness interruption did not restart the quiet interval")
				}
				if !stability.observe(pair, base.Add(300*time.Millisecond)) {
					t.Fatal("complete joint interval not accepted")
				}
			})
		}
	}
}

func TestInitialFramesPermitVisitsAndReadbacks(t *testing.T) {
	base := time.Unix(1, 0)
	frame := cogent.GPUViewStats{Ready: true, Frames: 1, Size: image.Pt(640, 480)}
	pair := [2]cogent.GPUViewStats{frame, frame}
	var stability initialFrameStability
	stability.observe(pair, base)
	pair[0].RenderVisits++
	pair[1].SnapshotReadbacks++
	if !stability.observe(pair, base.Add(100*time.Millisecond)) {
		t.Fatal("non-rendering visits/readbacks blocked an otherwise quiet pair")
	}
}
