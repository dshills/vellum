package cogent

import (
	"errors"
	"fmt"
	"image"
	"reflect"
	"sync"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/app"
)

func TestFrameMetricsCapacityAndInvalidCopies(t *testing.T) {
	for _, n := range []int{-1, 0, 65537} {
		if _, err := NewFrameMetrics(n); !errors.Is(err, ErrFrameMetricsConfig) {
			t.Fatalf("capacity %d: %v", n, err)
		}
	}
	m, err := NewFrameMetrics(2)
	if err != nil {
		t.Fatal(err)
	}
	// Deliberately simulate a forbidden copy before either mutex is ever used.
	// Reflection avoids teaching go vet that copying live mutexes is acceptable.
	var copy FrameMetrics
	reflect.ValueOf(&copy).Elem().Set(reflect.ValueOf(m).Elem())
	if copy.valid() || copy.Snapshot().Failure == "" {
		t.Fatal("copied recorder accepted")
	}
	var zero FrameMetrics
	if zero.valid() || zero.Snapshot().Failure == "" {
		t.Fatal("zero recorder accepted")
	}
	if got := m.Snapshot(); got.Failure != "" || got.Pending != 0 || got.Dropped != 0 || len(got.Samples) != 0 {
		t.Fatalf("new recorder: %+v", got)
	}
}

func TestFrameMetricsPollCollectsFailureAndRetirement(t *testing.T) {
	m, err := NewFrameMetrics(2)
	if err != nil {
		t.Fatal(err)
	}
	a := &app.Application{}
	w := &nativeWindow{window: app.Window{ID: "poll-unit"}, body: core.NewBody("poll-unit")}
	b, err := m.bind(a, w)
	if err != nil {
		t.Fatal(err)
	}
	defer b.retire()
	begin := func(tag uint64) {
		e := events.NewScroll(image.Pt(1, 2), math32.Vector2{Y: 12}, 0)
		e.Init()
		m.mu.Lock()
		defer m.mu.Unlock()
		if err := b.trace.Begin(tag, e); err != nil {
			t.Fatal(err)
		}
		b.tag = tag
		m.tags[tag] = 0
	}
	begin(900)
	for _, tag := range []uint64{900, 33} {
		if p := m.Poll(tag); p.Found || p.Pending != 1 || p.Dropped != 0 || p.Failure != "" {
			t.Fatalf("pending/unknown: %+v", p)
		}
	}
	b.trace.Fail(900, "injected refusal")
	if p := m.Poll(900); p.Found || p.Pending != 1 {
		t.Fatal("refusal collected before input/owner terminal")
	}
	// Retirement supplies the terminal for this undispatched unit fixture,
	// preserving the refusal recorded before that terminal.
	b.trace.Retire(900, "fixture terminal")
	p := m.Poll(900)
	if !p.Found || p.Pending != 0 || p.Sample.Completed || p.Sample.Tag != 900 || p.Sample.Failure != "injected refusal" {
		t.Fatalf("failure: %+v", p)
	}
	s := m.Snapshot()
	if len(s.Samples) != 1 || !reflect.DeepEqual(p.Sample, s.Samples[0]) {
		t.Fatal("poll/snapshot failed-record mismatch")
	}
	begin(3)
	b.retire()
	p = m.Poll(3)
	if !p.Found || p.Pending != 0 || p.Sample.Completed || p.Sample.Failure != "native window retired" {
		t.Fatalf("retirement: %+v", p)
	}
	p.Sample.Tag = 999
	if m.Poll(3).Sample.Tag != 3 || len(m.Snapshot().Samples) != 2 || !m.Poll(900).Found {
		t.Fatal("poll removes or aliases history")
	}
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for range 100 {
				if !m.Poll(3).Found || len(m.Snapshot().Samples) != 2 {
					t.Error("concurrent observer lost history")
					return
				}
			}
		})
	}
	wg.Wait()
}

func TestFrameMetricsPollInvalidAndDetachedDetails(t *testing.T) {
	var zero FrameMetrics
	if zero.Poll(1).Failure == "" || (*FrameMetrics)(nil).Poll(1).Failure == "" {
		t.Fatal("invalid recorder accepted")
	}
	m, _ := NewFrameMetricsWithSourceDetails(2)
	var copied FrameMetrics
	reflect.ValueOf(&copied).Elem().Set(reflect.ValueOf(m).Elem())
	if copied.Poll(1).Failure == "" || m.Poll(0).Failure == "" {
		t.Fatal("copied recorder or zero tag accepted")
	}
	if p := m.Poll(17); p.Found || p.Pending != 0 || p.Dropped != 0 || p.Failure != "" {
		t.Fatal("unknown tag changed status")
	}
	m.samples = append(m.samples, FrameSample{Tag: 17, Completed: true, SourceDetails: &FrameSourceDetails{Compositions: 1, Records: []FrameSourceSample{{DrawNS: 100}}}})
	m.tags[17] = 1
	m.dropped = 2
	p := m.Poll(17)
	if !p.Found || !p.Sample.Completed || p.Dropped != 2 {
		t.Fatal("sample/drop status lost")
	}
	p.Sample.SourceDetails.Compositions = 9
	p.Sample.SourceDetails.Records[0].DrawNS = 999
	if q := m.Poll(17); q.Sample.SourceDetails.Compositions != 1 || q.Sample.SourceDetails.Records[0].DrawNS != 100 || m.Snapshot().Dropped != 2 {
		t.Fatal("details alias observer")
	}
}

func BenchmarkFrameMetricsObserverHistory(b *testing.B) {
	for _, n := range []int{128, 512, 4096} {
		m, _ := NewFrameMetrics(n)
		for i := 0; i < n; i++ {
			tag := uint64(i + 1)
			m.samples = append(m.samples, FrameSample{Tag: tag, Completed: true, Generation: tag})
			m.tags[tag] = i + 1
		}
		b.Run(fmt.Sprintf("Snapshot/%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				if len(m.Snapshot().Samples) != n {
					b.Fatal("lost history")
				}
			}
		})
		b.Run(fmt.Sprintf("Poll/%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				if !m.Poll(uint64(n)).Found {
					b.Fatal("lost ticket")
				}
			}
		})
	}
}

func TestFrameMetricsRetirementDetachesNativeOwnerAndSample(t *testing.T) {
	m, err := NewFrameMetrics(2)
	if err != nil {
		t.Fatal(err)
	}
	a := &app.Application{}
	w := &nativeWindow{window: app.Window{ID: "metrics-unit"}, body: core.NewBody("metrics-unit")}
	b, err := m.bind(a, w)
	if err != nil {
		t.Fatal(err)
	}
	if w.body.Scene.FrameTrace() != b.trace {
		t.Fatal("trace not published")
	}
	other := &nativeWindow{window: app.Window{ID: "metrics-other"}, body: core.NewBody("metrics-other")}
	if _, err := m.bind(a, other); !errors.Is(err, ErrFrameMetricsConfig) {
		t.Fatalf("admitted another live window: %v", err)
	}
	e := events.NewScroll(image.Pt(1, 2), math32.Vector2{Y: 12}, 0)
	e.Init()
	if err := b.trace.Begin(1, e); err != nil {
		t.Fatal(err)
	}
	b.tag = 1
	m.tags[1] = 0
	if got := m.Snapshot(); got.Pending != 1 || len(got.Samples) != 0 {
		t.Fatal("queued event not pending")
	}
	b.retire()
	b.retire()
	got := m.Snapshot()
	if got.Pending != 0 || len(got.Samples) != 1 || got.Samples[0].Completed || got.Samples[0].Failure != "native window retired" {
		t.Fatalf("wrong retirement: %+v", got)
	}
	if b.window != nil || b.key.owner != nil || w.body.Scene.FrameTrace() != nil {
		t.Fatal("retained native owner")
	}
	frameMetricsRegistry.Lock()
	leaked := frameMetricsRegistry.windows[frameWindowKey{a, "metrics-unit", w.window.Ref()}]
	frameMetricsRegistry.Unlock()
	if leaked != nil {
		t.Fatal("retained global registration")
	}
	got.Samples[0].Tag = 99
	if m.Snapshot().Samples[0].Tag != 1 {
		t.Fatal("sample aliases observer")
	}
	b2, err := m.bind(a, other)
	if err != nil {
		t.Fatal(err)
	}
	b2.retire()
}
