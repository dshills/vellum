package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/gpu"
	"cogentcore.org/core/system"
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/oliverbestmann/webgpu/wgpu"
	"image"
	"image/color"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSnapshotPaddingCapsAndDetachedLinearAlpha(t *testing.T) {
	stride, bytes, err := snapshotLayout(image.Pt(65, 2), 1024, 1024)
	if err != nil || stride != 512 || bytes != 1024 {
		t.Fatal(stride, bytes, err)
	}
	for _, limits := range [][2]uint64{{1023, 1024}, {1024, 1023}} {
		if _, _, err := snapshotLayout(image.Pt(65, 2), limits[0], limits[1]); err == nil {
			t.Fatal("padded cap not enforced")
		}
	}
	for _, size := range []image.Point{{}, {-1, 2}, {16385, 1}} {
		if _, _, err := snapshotLayout(size, 1<<30, 1<<30); err == nil {
			t.Fatal("size accepted", size)
		}
	}
	raw := make([]byte, bytes)
	straight := color.NRGBA{R: 210, G: 130, B: 30, A: 128}
	for y := 0; y < 2; y++ {
		for x := 0; x < 65; x++ {
			i := y*int(stride) + x*4
			raw[i] = srgbByte(float64(linearByte(straight.R)) * 128 / 255)
			raw[i+1] = srgbByte(float64(linearByte(straight.G)) * 128 / 255)
			raw[i+2] = srgbByte(float64(linearByte(straight.B)) * 128 / 255)
			raw[i+3] = 128
		}
	}
	img, err := copySnapshot(context.Background(), raw, image.Pt(65, 2), stride)
	if err != nil {
		t.Fatal(err)
	}
	if len(img.Pix) != 65*2*4 || img.Stride != 260 {
		t.Fatal("CPU copy retains padding", len(img.Pix), img.Stride)
	}
	raw[0] = 0 // mapped bytes are independent from the worker image.
	if err := decodeSnapshot(context.Background(), img); err != nil {
		t.Fatal(err)
	}
	got := img.NRGBAAt(0, 0)
	for _, delta := range []int{int(got.R) - 210, int(got.G) - 130, int(got.B) - 30} {
		if delta < -3 || delta > 3 {
			t.Fatal("linear alpha mismatch", got)
		}
	}
	if got.A != 128 || got != img.NRGBAAt(64, 1) {
		t.Fatal(got)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := copySnapshot(canceled, raw, image.Pt(65, 2), stride); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := decodeSnapshot(canceled, img); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := copySnapshot(context.Background(), raw[:100], image.Pt(65, 2), stride); err == nil {
		t.Fatal("truncated map accepted")
	}
	transparent := &image.NRGBA{Pix: []byte{255, 255, 255, 0}, Stride: 4, Rect: image.Rect(0, 0, 1, 1)}
	decodeSnapshot(context.Background(), transparent)
	if transparent.NRGBAAt(0, 0) != (color.NRGBA{}) {
		t.Fatal("transparent RGB survives")
	}
}
func TestSnapshotBudgetIncludesAsyncWithHelpers(t *testing.T) {
	p := gpuResourcePool{budget: 100, stats: &GPUViewStats{TargetBytes: 10, BufferBytes: 20, TextureBytes: 30, ReadbackBytes: 10, AsyncReadbackBytes: 25}}
	if p.admit(5) != nil || p.admit(6) == nil {
		t.Fatal("async staging omitted from admission")
	}
}

func snapshotFixture(t *testing.T) (*app.Application, *GPUView, *gpuViewWidget, *gpuSnapshotJob) {
	t.Helper()
	a, _ := app.New(app.Config{QueueCapacity: 1})
	if err := a.Open(app.Window{ID: "capture", Title: "capture", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	g := &GPUView{Owner: a, Window: "capture", Program: &GPUProgram{Draw: func(GPUFrame, *wgpu.RenderPassEncoder) error { return nil }}}
	g.Mount(core.NewBody("capture"))
	w := g.widget
	w.native.Store(&drawingWindow{&snapshotTestWindow{}})
	ctx, cancel := context.WithCancel(context.Background())
	lifetime, _ := a.WindowContext(g.Window)
	job := &gpuSnapshotJob{ctx: ctx, cancel: cancel, reservation: &g.snapshotActive, done: make(chan struct{}), result: GPUSnapshot{Image: image.NewNRGBA(image.Rect(0, 0, 3, 2)), Frame: 7, Revision: 4, ResourceGeneration: 10, TargetGeneration: 11}}
	job.stopWindow = context.AfterFunc(lifetime, cancel)
	g.snapshotActive.Store(true)
	w.resources.capture = job
	w.resources.stats.ResourceGeneration = 10
	t.Cleanup(func() { cancel(); a.Close(); w.Destroy() })
	return a, g, w, job
}
func awaitSnapshotDone(t *testing.T, job *gpuSnapshotJob) {
	t.Helper()
	select {
	case <-job.done:
	case <-time.After(time.Second):
		t.Fatal("snapshot worker failed to release reservation")
	}
}
func TestSnapshotCapturedResizeAndCallbackReservation(t *testing.T) {
	a, g, w, job := snapshotFixture(t)
	// A normal resize changes target generation without invalidating the capture.
	w.resources.stats.TargetGeneration = 99
	w.resources.key.revision = 5
	calls := 0
	job.callback = func(result GPUSnapshot, err error) {
		calls++
		if err != nil || result.Revision != 4 || result.Frame != 7 || result.TargetGeneration != 11 || result.Image.Bounds() != image.Rect(0, 0, 3, 2) {
			t.Error(result, err)
		}
		if !g.snapshotActive.Load() || g.snapshotActive.CompareAndSwap(false, true) {
			t.Error("busy slot released before callback return")
		}
		// A callback may take resource locks: they must not be held during delivery.
		w.resources.mu.Lock()
		w.resources.mu.Unlock()
	}
	a.Do(func() { w.deliverSnapshot(job, nil) })
	awaitSnapshotDone(t, job)
	if calls != 1 || g.snapshotActive.Load() || job.result.Image != nil || w.resources.stats.AsyncSnapshotCompleted != 1 {
		t.Fatal("delivery reservation", calls, w.resources.stats)
	}
	// Stale delivery/cleanup cannot clear a successor's reservation.
	g.snapshotActive.Store(true)
	w.resources.mu.Lock()
	w.resources.finishSnapshot(job, false)
	w.resources.mu.Unlock()
	w.deliverSnapshot(job, nil)
	if !g.snapshotActive.Load() || calls != 1 {
		t.Fatal("stale completion altered successor")
	}
	g.snapshotActive.Store(false)
}
func TestSnapshotCanceledMailboxAndQueuedPayload(t *testing.T) {
	for _, queued := range []bool{false, true} {
		t.Run(map[bool]string{false: "full-mailbox", true: "queued-undrained"}[queued], func(t *testing.T) {
			a, g, w, job := snapshotFixture(t)
			var callback atomic.Bool
			job.callback = func(GPUSnapshot, error) { callback.Store(true) }
			enqueued := make(chan struct{}, 1)
			a.SetWake(func() {
				select {
				case enqueued <- struct{}{}:
				default:
				}
			})
			waiting := make(chan struct{})
			if !queued {
				job.ctx = &observedSnapshotWait{Context: job.ctx, entered: waiting}
				a.Post(func() {})
			}
			go w.completeSnapshot(job, nil)
			if queued {
				select {
				case <-enqueued:
				case <-time.After(time.Second):
					t.Fatal("no queued completion")
				}
			}
			if !queued {
				select {
				case <-waiting:
				case <-time.After(time.Second):
					t.Fatal("worker did not enter mailbox credit wait")
				}
			}
			job.cancel()
			awaitSnapshotDone(t, job)
			if g.snapshotActive.Load() || job.result.Image != nil || callback.Load() {
				t.Fatal("cancel retained image/reservation or delivered")
			}
			a.Drain()
			if callback.Load() {
				t.Fatal("stale queued closure delivered canceled pixels")
			}
		})
	}
}
func TestSnapshotCancelDuringCallbackKeepsReservation(t *testing.T) {
	a, g, w, job := snapshotFixture(t)
	entered, release := make(chan struct{}), make(chan struct{})
	job.callback = func(GPUSnapshot, error) { close(entered); <-release }
	go w.completeSnapshot(job, nil)
	drained := make(chan struct{})
	go func() {
		for g.snapshotActive.Load() {
			a.Drain()
			time.Sleep(time.Millisecond)
		}
		close(drained)
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("callback not reached")
	}
	job.cancel()
	time.Sleep(10 * time.Millisecond)
	if !g.snapshotActive.Load() {
		t.Fatal("cancel released a running callback")
	}
	close(release)
	awaitSnapshotDone(t, job)
	<-drained
}

// Blocks the worker's first decode row so reset cannot assume cancellation has
// already released a CPU image. Only production completeSnapshot owns cleanup.
type heldSnapshotContext struct {
	context.Context
	calls            atomic.Int32
	entered, release chan struct{}
}

func (c *heldSnapshotContext) Err() error {
	if c.calls.Add(1) == 2 {
		close(c.entered)
		<-c.release
	}
	return c.Context.Err()
}
func TestSnapshotCanceledWorkerHoldsReservationUntilExit(t *testing.T) {
	_, g, w, job := snapshotFixture(t)
	held := &heldSnapshotContext{Context: job.ctx, entered: make(chan struct{}), release: make(chan struct{})}
	job.ctx = held
	go w.completeSnapshot(job, nil)
	<-held.entered
	w.resources.mu.Lock()
	w.resources.reset()
	w.resources.mu.Unlock()
	if !g.snapshotActive.Load() || job.result.Image == nil {
		t.Fatal("reset freed a worker-owned image")
	}
	close(held.release)
	awaitSnapshotDone(t, job)
	if g.snapshotActive.Load() || job.result.Image != nil {
		t.Fatal("canceled worker retained image")
	}
}
func TestSnapshotCloseDropsQueuedPayload(t *testing.T) {
	a, g, w, job := snapshotFixture(t)
	called := false
	job.callback = func(GPUSnapshot, error) { called = true }
	enqueued := make(chan struct{}, 1)
	a.SetWake(func() {
		select {
		case enqueued <- struct{}{}:
		default:
		}
	})
	go w.completeSnapshot(job, nil)
	<-enqueued
	a.Close()
	awaitSnapshotDone(t, job)
	if called || g.snapshotActive.Load() || job.result.Image != nil {
		t.Fatal("closed app retained queued payload")
	}
}

// IsClosed and a failed Lock must never inspect a withdrawn device wrapper.
type snapshotTestWindow struct {
	system.Window
	closed atomic.Bool
	locks  atomic.Uint64
}

func (w *snapshotTestWindow) IsClosed() bool { return w.closed.Load() }
func (w *snapshotTestWindow) Lock() bool     { w.locks.Add(1); return false }
func TestSnapshotForcedNativeCloseWithoutWidgetDestruction(t *testing.T) {
	_, g, w, job := snapshotFixture(t)
	native := &snapshotTestWindow{}
	native.closed.Store(true)
	w.native.Store(&drawingWindow{native})
	// An invalid device would panic if the close fallback attempted polling.
	w.resources.device = &gpu.Device{Device: &wgpu.Device{}}
	go w.runSnapshot(native, job)
	awaitSnapshotDone(t, job)
	if w.retired.Load() || native.locks.Load() != 1 || g.snapshotActive.Load() || job.result.Image != nil || w.resources.stats.AsyncSnapshotPolls != 0 {
		t.Fatal("forced close requires widget destruction or device access")
	}
	// Native closure also suppresses an already queued CPU completion.
	a2, g2, w2, job2 := snapshotFixture(t)
	w2.native.Store(&drawingWindow{native})
	job2.callback = func(GPUSnapshot, error) { t.Error("closed native delivery") }
	a2.Do(func() { w2.deliverSnapshot(job2, nil) })
	awaitSnapshotDone(t, job2)
	if g2.snapshotActive.Load() {
		t.Fatal("closed native payload retained reservation")
	}
}

// Done is requested by PostWait only after finding a full mailbox.
type observedSnapshotWait struct {
	context.Context
	once    sync.Once
	entered chan struct{}
}

func (c *observedSnapshotWait) Done() <-chan struct{} {
	c.once.Do(func() { close(c.entered) })
	return c.Context.Done()
}

// Mounted resource ownership must reject an occupied capture even if a model
// alias has an independent atomic admission flag. No GPU API may be reached.
type admittedSnapshotWindow struct{ snapshotTestWindow }

func (w *admittedSnapshotWindow) Lock() bool { return true }
func (w *admittedSnapshotWindow) Unlock()    {}
func TestSnapshotMountedCaptureCannotBeOverwritten(t *testing.T) {
	_, g, w, job := snapshotFixture(t)
	w.native.Store(&drawingWindow{&admittedSnapshotWindow{}})
	g.snapshotActive.Store(false)
	if err := g.SnapshotAsync(context.Background(), func(GPUSnapshot, error) {}); !errors.Is(err, ErrSnapshotBusy) {
		t.Fatal("existing mounted capture replaced", err)
	}
	if w.resources.capture != job || g.snapshotActive.Load() {
		t.Fatal("busy admission mutated ownership")
	}
}
