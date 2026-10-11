package cogent

import (
	"errors"
	"image"
	"reflect"
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
	m.tags[1] = true
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
	leaked := frameMetricsRegistry.windows[frameWindowKey{a, "metrics-unit"}]
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
