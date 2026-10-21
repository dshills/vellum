package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"image"
	"image/color"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestQualificationCannotPromoteSmokeToFullSupport(t *testing.T) {
	for _, os := range []string{"windows", "linux", "darwin"} {
		if len(qualificationGaps(os, "x11")) == 0 {
			t.Fatal(os)
		}
	}
	for _, display := range []string{"xwayland", "wayland"} {
		if !strings.Contains(strings.Join(qualificationGaps("linux", display), " "), "native Wayland") {
			t.Fatal(display)
		}
	}
}

func TestFailureReportSnapshotDuringWorkerProgress(t *testing.T) {
	recorder := &reportRecorder{value: report{QualificationGaps: []string{"pending"}}}
	var worker sync.WaitGroup
	worker.Go(func() {
		for range 1000 {
			recorder.add("progress")
		}
	})
	for range 1000 {
		snapshot := recorder.snapshot()
		snapshot.Event = "failure"
		if _, err := json.Marshal(snapshot); err != nil {
			t.Fatal(err)
		}
		if len(snapshot.Checks) > 0 {
			snapshot.Checks[0] = "external mutation"
		}
		snapshot.QualificationGaps[0] = "external mutation"
	}
	worker.Wait()
	final := recorder.snapshot()
	if len(final.Checks) != 1000 || final.Checks[0] != "progress" || final.QualificationGaps[0] != "pending" {
		t.Fatal("snapshot shared report storage")
	}
}

func TestSnapshotProbeRejectsStaleDimensionsAndCamera(t *testing.T) {
	view := canvas.Viewport{Width: 200, Height: 100, Zoom: 1, Center: canvas.Point{X: 50, Y: 30}}
	img := image.NewNRGBA(image.Rect(0, 0, 200, 100))
	img.SetNRGBA(100, 50, color.NRGBA{R: 67, G: 197, B: 166, A: 255})
	if err := validateSnapshot(img, view); err != nil {
		t.Fatal(err)
	}
	if err := validateSnapshot(image.NewNRGBA(image.Rect(0, 0, 199, 100)), view); err == nil {
		t.Fatal("stale allocation accepted")
	}
	view.Zoom = 1.25
	view.Center = canvas.Point{X: 10, Y: 10}
	if err := validateSnapshot(img, view); err == nil {
		t.Fatal("stale camera projection accepted")
	}
	p := view.Screen(canvas.Point{X: 50, Y: 30})
	img.SetNRGBA(int(p.X), int(p.Y), color.NRGBA{R: 67, G: 197, B: 166, A: 255})
	if err := validateSnapshot(img, view); err != nil {
		t.Fatal(err)
	}
}

func TestCollectRunResultAvailablePreservesBothFailures(t *testing.T) {
	nativeErr := errors.New("native loop failed")
	workerErr := errors.New("GPU check failed")
	for _, tc := range []struct {
		name              string
		runErr, workerErr error
	}{
		{"success", nil, nil},
		{"native", nativeErr, nil},
		{"worker", nil, workerErr},
		{"both", nativeErr, workerErr},
		{"native and worker cancellation", nativeErr, context.Canceled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := make(chan error, 1)
			result <- tc.workerErr
			// A ready worker result takes precedence over caller cancellation,
			// but a genuine native failure remains discoverable via errors.Is.
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			err := collectRunResult(ctx, tc.runErr, result)
			if tc.runErr == nil && tc.workerErr == nil && err != nil {
				t.Fatal(err)
			}
			for _, want := range []error{tc.runErr, tc.workerErr} {
				if want != nil && !errors.Is(err, want) {
					t.Fatalf("lost %v in %v", want, err)
				}
			}
			if tc.workerErr != context.Canceled && errors.Is(err, context.Canceled) {
				t.Fatalf("caller cancellation masked ready result: %v", err)
			}
		})
	}
}

func TestCollectRunResultPendingNativeFailureDoesNotWait(t *testing.T) {
	nativeErr := errors.New("native loop failed")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := make(chan error, 1)
	if err := collectRunResult(ctx, nativeErr, result); err != nativeErr {
		t.Fatalf("native error replaced by pending worker/caller cancellation: %v", err)
	}
}

func TestCollectRunResultPendingCallerCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := collectRunResult(ctx, nil, make(chan error, 1)); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

// observedWaitContext signals when collection reaches its blocking branch.
// The worker result is then delivered without scheduler timing or sleeps.
type observedWaitContext struct {
	context.Context
	waiting chan struct{}
	once    sync.Once
}

func (c *observedWaitContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.waiting) })
	return c.Context.Done()
}

func TestCollectRunResultWaitsOnlyAfterSuccessfulRun(t *testing.T) {
	workerErr := errors.New("worker check failed")
	for _, want := range []error{nil, workerErr} {
		ctx := &observedWaitContext{Context: context.Background(), waiting: make(chan struct{})}
		result := make(chan error, 1)
		collected := make(chan error, 1)
		go func() { collected <- collectRunResult(ctx, nil, result) }()
		<-ctx.waiting
		result <- want
		if err := <-collected; err != want {
			t.Fatalf("pending worker result: got %v want %v", err, want)
		}
	}
}

func TestResizeAcceptanceRequiresObservedNativeAndCanvasTransition(t *testing.T) {
	before := canvas.Viewport{Width: 600, Height: 400}
	for _, tc := range []struct {
		name                                   string
		beforeNative, currentNative, requested image.Point
		current                                canvas.Viewport
		want                                   bool
	}{
		{"both axes", image.Pt(640, 480), image.Pt(800, 600), image.Pt(800, 600), canvas.Viewport{Width: 760, Height: 520}, true},
		{"width only", image.Pt(640, 480), image.Pt(800, 480), image.Pt(800, 480), canvas.Viewport{Width: 760, Height: 400}, true},
		{"height only", image.Pt(640, 480), image.Pt(640, 600), image.Pt(640, 600), canvas.Viewport{Width: 600, Height: 520}, true},
		{"already requested", image.Pt(800, 600), image.Pt(800, 600), image.Pt(800, 600), canvas.Viewport{Width: 760, Height: 520}, false},
		{"ignored", image.Pt(640, 480), image.Pt(640, 480), image.Pt(800, 600), canvas.Viewport{Width: 760, Height: 520}, false},
		{"clamped", image.Pt(640, 480), image.Pt(720, 540), image.Pt(800, 600), canvas.Viewport{Width: 680, Height: 460}, false},
		{"stale width", image.Pt(640, 480), image.Pt(800, 600), image.Pt(800, 600), canvas.Viewport{Width: 600, Height: 520}, false},
		{"stale height", image.Pt(640, 480), image.Pt(800, 600), image.Pt(800, 600), canvas.Viewport{Width: 760, Height: 400}, false},
		{"not allocated", image.Pt(640, 480), image.Pt(800, 600), image.Pt(800, 600), canvas.Viewport{}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := resizeAccepted(tc.beforeNative, tc.currentNative, tc.requested, before, tc.current); got != tc.want {
				t.Fatalf("accepted=%v want=%v", got, tc.want)
			}
		})
	}
}

func TestProbeViewRetainsCanvasExtensionIdentity(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	var views atomic.Int32
	var shown atomic.Bool
	w, renderer, model := newProbeWindow(a, "first", &views, &shown)
	first, next := w.View(), w.View()
	p, ok := first.Children[1].Custom.(*shownCanvasProbe)
	if !ok || p != next.Children[1].Custom || p.renderer != renderer || renderer.Model != model || renderer.Owner != a || p.shown != &shown || views.Load() != 2 {
		t.Fatal("refresh replaced the mounted canvas extension")
	}
	other, _, _ := newProbeWindow(a, "second", &views, &shown)
	if other.View().Children[1].Custom == p {
		t.Fatal("windows shared extension identity")
	}
}
