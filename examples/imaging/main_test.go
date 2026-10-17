package main

import (
	"context"
	"errors"
	"image"
	"sync"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
)

func TestCaptureCompletionNeverBlocksFullDiagnosticMailbox(t *testing.T) {
	for _, scenario := range []string{"admission error", "callback error", "callback success"} {
		t.Run(scenario, func(t *testing.T) {
			a, err := app.New(app.Config{})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Close()
			w, spec := newImageWindow(a, "capture", false, make(chan error, 1))
			if w.windowRef, err = a.OpenWindow(spec); err != nil {
				t.Fatal(err)
			}
			prior := errors.New("prior unread diagnostic")
			w.captureResults <- captureResult{err: prior}
			callbackErr := errors.New("asynchronous capture failed")
			done := make(chan error, 1)
			go func() {
				done <- a.Do(func() {
					switch scenario {
					case "admission error":
						// An unmounted view rejects SnapshotAsync synchronously.
						w.capture()
					case "callback error":
						w.captureBusy = true
						w.completeCapture(context.Background(), cogent.GPUSnapshot{}, callbackErr)
					case "callback success":
						w.captureBusy = true
						w.completeCapture(context.Background(), cogent.GPUSnapshot{Frame: 7, Image: image.NewNRGBA(image.Rect(0, 0, 2, 2))}, nil)
					}
				})
			}()
			select {
			case err := <-done:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(2 * time.Second):
				// Release an old blocking implementation before failing the test.
				<-w.captureResults
				<-done
				t.Fatal("capture completion blocked the UI owner on its diagnostic channel")
			}
			if err := a.Do(func() {
				if w.captureBusy {
					t.Error("terminal capture left button busy")
				}
				if scenario == "admission error" && w.status != cogent.ErrFrameNotReady.Error() {
					t.Errorf("missing admission error: %q", w.status)
				}
				if scenario == "callback error" && w.status != callbackErr.Error() {
					t.Errorf("missing callback error: %q", w.status)
				}
				if scenario == "callback success" && w.status != "Captured frame 7: 2 × 2" {
					t.Errorf("missing success status: %q", w.status)
				}
			}); err != nil {
				t.Fatal(err)
			}
			if got := <-w.captureResults; !errors.Is(got.err, prior) {
				t.Fatalf("unread diagnostic replaced: %+v", got)
			}
			if len(w.captureResults) != 0 {
				t.Fatal("new diagnostic retained over full prior slot")
			}
		})
	}
}

func TestSmokeWorkerCompletionFollowsCleanup(t *testing.T) {
	for _, scenario := range []string{"success", "failure"} {
		t.Run(scenario, func(t *testing.T) {
			workCtx, cancel := context.WithCancel(context.Background())
			defer cancel()
			cleanupEntered := make(chan struct{})
			allowCleanup := make(chan struct{})
			var release sync.Once
			t.Cleanup(func() { release.Do(func() { close(allowCleanup) }) })
			cleanupFinished := false
			want := error(nil)
			if scenario == "failure" {
				want = errors.New("native worker failed")
			}
			result, done := startSmokeWorker(func() {
				close(cleanupEntered)
				<-allowCleanup
				cleanupFinished = true
			}, func() error {
				defer cancel()
				return want
			})
			select {
			case got := <-result:
				if got != want || workCtx.Err() == nil {
					t.Fatalf("result preceded work/context cleanup: got=%v want=%v context=%v", got, want, workCtx.Err())
				}
			case <-time.After(2 * time.Second):
				t.Fatal("worker result unavailable")
			}
			select {
			case <-cleanupEntered:
			case <-time.After(2 * time.Second):
				t.Fatal("stop callback not entered")
			}
			// The stop callback is now demonstrably blocked. Result delivery must
			// not have published worker completion before that callback returns.
			select {
			case <-done:
				t.Fatal("worker completion preceded application cleanup")
			default:
			}
			release.Do(func() { close(allowCleanup) })
			select {
			case <-done:
				if !cleanupFinished {
					t.Fatal("completion did not follow cleanup")
				}
			case <-time.After(2 * time.Second):
				t.Fatal("worker did not join after cleanup")
			}
		})
	}
}
