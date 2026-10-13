package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/system/composer"
	"errors"
	"github.com/dshills/vellum/app"
	"image"
	"reflect"
	"testing"
)

func TestFrameSourceMetricsExplicitCapacityAndDetachedSnapshot(t *testing.T) {
	for _, n := range []int{0, -1, 4097, 65536} {
		if _, err := NewFrameMetricsWithSourceDetails(n); !errors.Is(err, ErrFrameMetricsConfig) {
			t.Fatalf("capacity%d: %v", n, err)
		}
	}
	m, err := NewFrameMetricsWithSourceDetails(4096)
	if err != nil {
		t.Fatal(err)
	}
	var copied FrameMetrics
	reflect.ValueOf(&copied).Elem().Set(reflect.ValueOf(m).Elem())
	if copied.valid() {
		t.Fatal("copied diagnostic accepted")
	}
	upstream := &composer.SourceDetails{Compositions: 1, Records: []composer.SourceTiming{{Composition: 1, Kind: composer.SourceScene, DrawNS: 100, CopyNS: 40, ImageWidth: 10, ImageHeight: 20, SetFromGoImageAttempts: 1, LogicalUploadRGBABytes: 800}}}
	m.samples = append(m.samples, FrameSample{Tag: 7, SourceDetails: frameSourceDetails(upstream)})
	upstream.Records[0].DrawNS = 999
	first := m.Snapshot()
	first.Samples[0].SourceDetails.Compositions = 99
	first.Samples[0].SourceDetails.Records[0].DrawNS = 888
	second := m.Snapshot()
	d := second.Samples[0].SourceDetails
	if d.Compositions != 1 || len(d.Records) != 1 || d.Records[0].DrawNS != 100 || d.UploadFailuresAvailable {
		t.Fatal("source details alias input or snapshot, or invent error status")
	}
	ordinary, err := NewFrameMetrics(65536)
	if err != nil || ordinary.sourceDetails || frameSourceDetails(nil) != nil {
		t.Fatal("ordinary metrics changed")
	}
}

func TestFrameSourceMetricsBindingRetirementKeepsDiagnosticMode(t *testing.T) {
	m, err := NewFrameMetricsWithSourceDetails(2)
	if err != nil {
		t.Fatal(err)
	}
	a := &app.Application{}
	w := &nativeWindow{window: app.Window{ID: "source-metrics-retire"}, body: core.NewBody("source-metrics-retire")}
	b, err := m.bind(a, w)
	if err != nil {
		t.Fatal(err)
	}
	e := events.NewScroll(image.Pt(1, 2), math32.Vector2{Y: 12}, 0)
	e.Init()
	if err = b.trace.Begin(1, e); err != nil {
		t.Fatal(err)
	}
	b.tag = 1
	m.tags[1] = 0
	b.retire()
	s := m.Snapshot()
	if s.Pending != 0 || len(s.Samples) != 1 || s.Samples[0].Completed || s.Samples[0].Failure != "native window retired" || s.Samples[0].SourceDetails == nil || s.Samples[0].SourceDetails.UploadFailuresAvailable || len(s.Samples[0].SourceDetails.Records) != 0 {
		t.Fatal("wrong diagnostic mode/retirement")
	}
	s.Samples[0].SourceDetails.Compositions = 99
	if m.Snapshot().Samples[0].SourceDetails.Compositions != 0 {
		t.Fatal("retired details alias observer")
	}
}
