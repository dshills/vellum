package canvas

import (
	"image/color"
	"math"
	"math/rand"
	"testing"
)

func TestMillionSampleEnvelopePreservesPeaks(t *testing.T) {
	samples := make([]float32, 1_000_000)
	samples[123456] = 1
	samples[999999] = -1
	w, err := NewWaveform(samples)
	if err != nil {
		t.Fatal(err)
	}
	samples[123456] = 0
	buckets, err := w.Envelope(0, w.Samples(), 1024)
	if err != nil || len(buckets) != 1024 {
		t.Fatal(err, len(buckets))
	}
	positive, negative := false, false
	last := 0
	for _, b := range buckets {
		if b.Start != last || b.End <= b.Start {
			t.Fatal("missing sample range")
		}
		last = b.End
		positive = positive || b.Max == 1
		negative = negative || b.Min == -1
	}
	if !positive || !negative || last != 1_000_000 {
		t.Fatal("lost peaks or coverage")
	}
	mesh, err := w.EnvelopeMesh(0, w.Samples(), 1024, Bounds{Point{}, Point{1024, 200}}, color.NRGBA{0, 255, 0, 255})
	if err != nil {
		t.Fatal(err)
	}
	if v, ix := mesh.Counts(); v != 4096 || ix != 6144 {
		t.Fatal("submitted source instead of envelope", v, ix)
	}
}
func TestWaveformQueriesMatchSource(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	samples := make([]float32, 1001)
	for i := range samples {
		samples[i] = rng.Float32()*2 - 1
	}
	w, _ := NewWaveform(samples)
	for range 100 {
		start := rng.Intn(len(samples))
		end := start + 1 + rng.Intn(len(samples)-start)
		buckets, err := w.Envelope(start, end, 1+rng.Intn(80))
		if err != nil {
			t.Fatal(err)
		}
		for _, b := range buckets {
			lo, hi := float32(1), float32(-1)
			for _, s := range samples[b.Start:b.End] {
				lo = min(lo, s)
				hi = max(hi, s)
			}
			if b.Min != lo || b.Max != hi {
				t.Fatal("inexact envelope", b, lo, hi)
			}
		}
	}
	for _, samples := range [][]float32{nil, {2}, {float32(math.NaN())}} {
		if _, err := NewWaveform(samples); err == nil {
			t.Fatal("bad source")
		}
	}
	for _, args := range [][3]int{{-1, 2, 1}, {0, 2000, 1}, {2, 2, 1}, {0, 10, 0}, {0, 10, MaxEnvelopeColumns + 1}} {
		if _, err := w.Envelope(args[0], args[1], args[2]); err == nil {
			t.Fatal("bad envelope")
		}
	}
}
func BenchmarkMillionSampleEnvelope(b *testing.B) {
	samples := make([]float32, 1_000_000)
	w, _ := NewWaveform(samples)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		w.Envelope(0, len(samples), 1024)
	}
}
