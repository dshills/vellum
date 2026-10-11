package main

import (
	"math"
	"testing"
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
