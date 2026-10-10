package canvas

import (
	"fmt"
	"image/color"
	"math"
)

// Waveform retains an immutable min/max segment tree of normalized audio samples.
// It preserves single-sample peaks at overview scale without scanning the full
// source on each frame. Construction is O(n); an envelope is O(columns*log(n)).
type Waveform struct {
	count, base int
	tree        []sampleRange
}
type sampleRange struct{ min, max float32 }
type WaveformBucket struct {
	Start, End int
	Min, Max   float32
}

const MaxWaveformSamples = 4_000_000
const MaxEnvelopeColumns = 16_384

func NewWaveform(samples []float32) (*Waveform, error) {
	if len(samples) == 0 || len(samples) > MaxWaveformSamples {
		return nil, fmt.Errorf("canvas: waveform sample count outside limits")
	}
	for _, s := range samples {
		if math.IsNaN(float64(s)) || math.IsInf(float64(s), 0) || s < -1 || s > 1 {
			return nil, fmt.Errorf("canvas: waveform samples must be finite and normalized")
		}
	}
	base := 1
	for base < len(samples) {
		base *= 2
	}
	w := &Waveform{count: len(samples), base: base, tree: make([]sampleRange, base*2)}
	for i := base; i < len(w.tree); i++ {
		w.tree[i] = sampleRange{float32(math.Inf(1)), float32(math.Inf(-1))}
	}
	for i, s := range samples {
		w.tree[base+i] = sampleRange{s, s}
	}
	for i := base - 1; i > 0; i-- {
		w.tree[i] = combine(w.tree[i*2], w.tree[i*2+1])
	}
	return w, nil
}
func combine(a, b sampleRange) sampleRange { return sampleRange{min(a.min, b.min), max(a.max, b.max)} }
func (w *Waveform) Samples() int {
	if w == nil {
		return 0
	}
	return w.count
}
func (w *Waveform) rangeAt(start, end int) sampleRange {
	r := sampleRange{float32(math.Inf(1)), float32(math.Inf(-1))}
	start += w.base
	end += w.base
	for start < end {
		if start&1 != 0 {
			r = combine(r, w.tree[start])
			start++
		}
		if end&1 != 0 {
			end--
			r = combine(r, w.tree[end])
		}
		start /= 2
		end /= 2
	}
	return r
}
func (w *Waveform) Envelope(start, end, columns int) ([]WaveformBucket, error) {
	if w == nil || start < 0 || end > w.count || start >= end || columns < 1 || columns > MaxEnvelopeColumns {
		return nil, fmt.Errorf("canvas: invalid waveform range or resolution")
	}
	columns = min(columns, end-start)
	out := make([]WaveformBucket, columns)
	for i := range out {
		a := start + int(int64(i)*int64(end-start)/int64(columns))
		b := start + int(int64(i+1)*int64(end-start)/int64(columns))
		r := w.rangeAt(a, b)
		out[i] = WaveformBucket{a, b, r.min, r.max}
	}
	return out, nil
}

// EnvelopeMesh draws one min/max band per visible column into world bounds.
// A constant bucket has zero area; consumers needing a centerline can add a stroke.
func (w *Waveform) EnvelopeMesh(start, end, columns int, b Bounds, c color.NRGBA) (*Mesh, error) {
	if err := b.Validate(); err != nil {
		return nil, err
	}
	buckets, err := w.Envelope(start, end, columns)
	if err != nil {
		return nil, err
	}
	v := make([]Vertex, 0, 4*len(buckets))
	ix := make([]uint32, 0, 6*len(buckets))
	for i, r := range buckets {
		x0 := b.Min.X + float64(i)*(b.Max.X-b.Min.X)/float64(len(buckets))
		x1 := b.Min.X + float64(i+1)*(b.Max.X-b.Min.X)/float64(len(buckets))
		y0 := b.Min.Y + (1-float64(r.Max))*.5*(b.Max.Y-b.Min.Y)
		y1 := b.Min.Y + (1-float64(r.Min))*.5*(b.Max.Y-b.Min.Y)
		v, ix = appendQuad(v, ix, Bounds{Point{x0, y0}, Point{x1, y1}}, c)
	}
	return NewMesh(v, ix)
}
