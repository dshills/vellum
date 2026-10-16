package cogent

import (
	"context"
	"errors"
	"fmt"
	"image"
	"sync/atomic"
	"time"

	"cogentcore.org/core/system"
	"github.com/oliverbestmann/webgpu/wgpu"
)

var ErrSnapshotBusy = errors.New("GPU view: asynchronous snapshot already pending")

const defaultSnapshotBytes = 64 << 20

// GPUSnapshot owns a detached straight-sRGB image and capture-time metadata.
// Normal target resize or model changes do not alter the captured frame.
type GPUSnapshot struct {
	Image                                                 *image.NRGBA
	Frame, Revision, ResourceGeneration, TargetGeneration uint64
}

// All fields except ctx/cancel/status/done/reservation belong to resources.mu.
// The sole worker owns image pixels until delivery starts. Queued closures keep
// only this job, so cancellation can discard the image without waiting for Drain.
type gpuSnapshotJob struct {
	ctx                  context.Context
	cancel               context.CancelFunc
	stopWindow           func() bool
	reservation          *atomic.Bool
	done                 chan struct{}
	result               GPUSnapshot
	callback             func(GPUSnapshot, error)
	buffer               *wgpu.Buffer
	bytes                uint64
	status               chan wgpu.MapAsyncStatus
	stride               uint32
	size                 image.Point
	delivering, finished bool
}

// SnapshotAsync captures the last completed frame without waiting for the GPU.
// Call on Owner.Do. One request per view is reserved through callback return or
// canceled worker cleanup, including remount. The worker polls without waiting
// under the originating window's teardown lock, independently of paint/visibility,
// copies mapped rows into one owned image, then decodes on the worker. Delivery
// uses the bounded owner mailbox with GPU/native locks released. Cancellation,
// rebuild/failure, retirement and closure quietly suppress delivery. The callback
// must not nest Do/Drain. MaxSnapshotBytes bounds both padded GPU staging and the
// single tightly packed CPU image (default 64 MiB each). Delivered images retained
// by the application are outside that bound. It never flushes pending changes.
func (g *GPUView) SnapshotAsync(ctx context.Context, callback func(GPUSnapshot, error)) error {
	if ctx == nil || callback == nil {
		return errors.New("GPU view: snapshot context and callback required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := g.validate(); err != nil {
		return err
	}
	window := g.NativeWindow()
	if window == nil || !window.Lock() {
		return ErrFrameNotReady
	}
	defer window.Unlock()
	widget := g.widget
	r := &widget.resources
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.capture != nil || !g.snapshotActive.CompareAndSwap(false, true) {
		return ErrSnapshotBusy
	}
	admitted := false
	defer func() {
		if !admitted {
			g.snapshotActive.Store(false)
		}
	}()
	if err := drawingDeviceError(window); err != nil {
		return err
	}
	if !r.ready || r.target == nil {
		return ErrFrameNotReady
	}
	limit := g.MaxSnapshotBytes
	if limit == 0 {
		limit = defaultSnapshotBytes
	}
	size := r.target.Size()
	stride, bytes, err := snapshotLayout(size, limit, r.limits.MaxBufferSize)
	if err != nil {
		return err
	}
	if err := gpuBudget(g.budget(), r.stats.TargetBytes, r.stats.BufferBytes, r.stats.TextureBytes, r.stats.ReadbackBytes, r.stats.AsyncReadbackBytes, bytes); err != nil {
		return err
	}
	lifetime, err := scopedWindowContext(g.Owner, g.widget.windowRef)
	if err != nil || lifetime.Err() != nil {
		return ErrFrameNotReady
	}
	jobCtx, cancel := context.WithCancel(ctx)
	job := &gpuSnapshotJob{ctx: jobCtx, cancel: cancel, reservation: &g.snapshotActive, done: make(chan struct{}), result: GPUSnapshot{Frame: r.stats.Frames, Revision: r.key.revision, ResourceGeneration: r.stats.ResourceGeneration, TargetGeneration: r.stats.TargetGeneration}, callback: callback, status: make(chan wgpu.MapAsyncStatus, 1), stride: stride, size: size, bytes: bytes}
	job.stopWindow = context.AfterFunc(lifetime, cancel)
	defer func() {
		if !admitted {
			job.cancel()
			job.stopWindow()
			if job.buffer != nil {
				_ = job.buffer.TryUnmap()
				job.buffer.Release()
			}
		}
	}()
	job.buffer, err = r.device.Device.TryCreateBuffer(&wgpu.BufferDescriptor{Label: "vellum.snapshot.async", Size: bytes, Usage: wgpu.BufferUsageCopyDst | wgpu.BufferUsageMapRead})
	if err != nil {
		return err
	}
	encoder, err := r.device.Device.TryCreateCommandEncoder(nil)
	if err != nil {
		return err
	}
	defer encoder.Release()
	if err := r.target.CurrentFrame().CopyToBuffer(encoder, job.buffer, stride); err != nil {
		return err
	}
	commands, err := encoder.TryFinish(nil)
	if err != nil {
		return err
	}
	defer commands.Release()
	r.device.Queue.Submit(commands)
	// Spontaneous map callbacks capture only a bounded, never-closed notifier.
	status := job.status
	if err := job.buffer.TryMapAsync(wgpu.MapModeRead, 0, bytes, func(value wgpu.MapAsyncStatus) {
		select {
		case status <- value:
		default:
		}
	}); err != nil {
		return err
	}
	r.capture = job
	r.stats.AsyncReadbackBytes = bytes
	r.stats.AsyncSnapshotRequests++
	admitted = true
	go widget.runSnapshot(window, job)
	return nil
}

func snapshotLayout(size image.Point, limit, bufferLimit uint64) (uint32, uint64, error) {
	if size.X < 1 || size.Y < 1 || size.X > 16384 || size.Y > 16384 {
		return 0, 0, errors.New("GPU view: invalid snapshot size")
	}
	logical := uint64(size.X) * uint64(size.Y) * 4
	stride := (uint64(size.X)*4 + 255) / 256 * 256
	bytes := stride * uint64(size.Y)
	if logical > limit || bytes > limit {
		return 0, 0, errors.New("GPU view: snapshot exceeds CPU/staging byte limit")
	}
	if bytes > bufferLimit {
		return 0, 0, errors.New("GPU view: snapshot exceeds device buffer limit")
	}
	return uint32(stride), bytes, nil
}

// Called under native window + resource locks, before device/target teardown.
// Cancellation retains the admission slot until the worker has released its image.
func (r *gpuViewResources) cancelSnapshot() {
	if job := r.capture; job != nil {
		job.cancel()
		r.releaseSnapshotBuffer(job)
	}
}
func (r *gpuViewResources) releaseSnapshotBuffer(job *gpuSnapshotJob) {
	if job.buffer != nil {
		defer func() { job.buffer = nil; r.stats.AsyncReadbackBytes = 0 }()
		defer job.buffer.Release()
		_ = job.buffer.TryUnmap()
	}
}
func (r *gpuViewResources) finishSnapshot(job *gpuSnapshotJob, completed bool) {
	if job.finished {
		return
	}
	job.finished = true
	job.result.Image = nil
	job.callback = nil
	if r.capture == job {
		r.capture = nil
	}
	if completed {
		r.stats.AsyncSnapshotCompleted++
	} else {
		r.stats.AsyncSnapshotCanceled++
	}
	job.cancel()
	if job.stopWindow != nil {
		job.stopWindow()
	}
	job.reservation.Store(false)
	close(job.done)
}

// pollSnapshot never enters the application owner or reads public model fields.
// Window.Lock serializes every native handle operation with close/teardown.
func (w *gpuViewWidget) pollSnapshot(window system.Window, job *gpuSnapshotJob) (ready bool, err error) {
	if !window.Lock() {
		// Staging owns a private native device reference. After teardown only
		// this owned buffer may be unmapped/released; never access window/device.
		r := &w.resources
		r.mu.Lock()
		defer r.mu.Unlock()
		job.cancel()
		r.releaseSnapshotBuffer(job)
		return true, ErrFrameNotReady
	}
	defer window.Unlock()
	r := &w.resources
	r.mu.Lock()
	defer r.mu.Unlock()
	defer func() {
		if problem := recover(); problem != nil {
			ready = true
			err = fmt.Errorf("GPU snapshot: %v", problem)
		}
		if ready {
			r.releaseSnapshotBuffer(job)
		}
	}()
	if job.ctx.Err() != nil {
		return true, job.ctx.Err()
	}
	if r.capture != job || r.device == nil {
		return true, ErrFrameNotReady
	}
	if err := drawingDeviceError(window); err != nil {
		return true, err
	}
	r.device.Device.Poll(false, nil)
	r.stats.AsyncSnapshotPolls++
	select {
	case status := <-job.status:
		if status != wgpu.MapAsyncStatusSuccess {
			return true, fmt.Errorf("GPU view: snapshot map status %v", status)
		}
		job.result.Image, err = copySnapshot(job.ctx, job.buffer.GetMappedRange(0, uint(job.bytes)), job.size, job.stride)
		if unmapErr := job.buffer.TryUnmap(); err == nil {
			err = unmapErr
		}
		return true, err
	default:
		return false, nil
	}
}
func (w *gpuViewWidget) runSnapshot(window system.Window, job *gpuSnapshotJob) {
	var resultErr error
	for {
		ready, err := w.pollSnapshot(window, job)
		if ready {
			resultErr = err
			break
		}
		timer := time.NewTimer(5 * time.Millisecond)
		select {
		case <-job.ctx.Done():
			timer.Stop()
		case <-timer.C:
		}
	}
	w.completeSnapshot(job, resultErr)
}

func (w *gpuViewWidget) completeSnapshot(job *gpuSnapshotJob, resultErr error) {
	// Keep the slot while this worker owns pixels or its delivery is running.
	defer func() { r := &w.resources; r.mu.Lock(); defer r.mu.Unlock(); r.finishSnapshot(job, false) }()
	if job.ctx.Err() != nil {
		return
	}
	if resultErr == nil {
		resultErr = decodeSnapshot(job.ctx, job.result.Image)
	}
	if job.ctx.Err() != nil {
		return
	}
	// Only the job is captured, never a local image alias. A canceled queued job
	// can release its image even if Application.Close discards the entire mailbox.
	if err := w.application.PostWait(job.ctx, func() { w.deliverSnapshot(job, resultErr) }); err != nil {
		return
	}
	select {
	case <-job.done:
	case <-job.ctx.Done():
		r := &w.resources
		r.mu.Lock()
		delivering := job.delivering
		r.mu.Unlock()
		if delivering {
			<-job.done
		}
	}
}
func (w *gpuViewWidget) deliverSnapshot(job *gpuSnapshotJob, resultErr error) {
	r := &w.resources
	r.mu.Lock()
	if job.finished || r.capture != job {
		r.mu.Unlock()
		return
	}
	g := w.owner
	native := w.native.Load()
	ctx, err := scopedWindowContext(w.application, w.windowRef)
	live := native != nil && !native.window.IsClosed() && !w.retired.Load() && g != nil && g.widget == w && g.Owner == w.application && g.Window == w.windowID && g.Program == w.program && g.Depth == w.depth && err == nil && ctx.Err() == nil && job.ctx.Err() == nil && r.stats.ResourceGeneration == job.result.ResourceGeneration
	if !live {
		r.finishSnapshot(job, false)
		r.mu.Unlock()
		return
	}
	job.delivering = true
	result, callback := job.result, job.callback
	r.mu.Unlock()
	defer func() { r.mu.Lock(); defer r.mu.Unlock(); r.finishSnapshot(job, true) }()
	callback(result, resultErr)
}
func copySnapshot(ctx context.Context, data []byte, size image.Point, stride uint32) (*image.NRGBA, error) {
	if ctx == nil {
		return nil, errors.New("GPU view: copy context required")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if size.X < 1 || size.Y < 1 || size.X > 16384 || size.Y > 16384 || uint64(stride) < uint64(size.X)*4 || uint64(stride)*uint64(size.Y) != uint64(len(data)) {
		return nil, errors.New("GPU view: invalid mapped snapshot layout")
	}
	result := image.NewNRGBA(image.Rect(0, 0, size.X, size.Y))
	for y := 0; y < size.Y; y++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		copy(result.Pix[y*result.Stride:(y+1)*result.Stride], data[y*int(stride):y*int(stride)+size.X*4])
	}
	return result, nil
}
func decodeSnapshot(ctx context.Context, img *image.NRGBA) error {
	if ctx == nil || img == nil {
		return errors.New("GPU view: decoder context and image required")
	}
	for y := 0; y < img.Rect.Dy(); y++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		row := img.Pix[y*img.Stride : y*img.Stride+img.Rect.Dx()*4]
		straightSnapshot(&image.NRGBA{Pix: row, Stride: img.Stride, Rect: image.Rect(0, 0, img.Rect.Dx(), 1)})
	}
	return nil
}
