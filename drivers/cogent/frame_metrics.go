package cogent

import (
	"context"
	"errors"
	"image"
	"math"
	"strings"
	"sync"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/virtual"
)

var (
	ErrFrameMetricsConfig   = errors.New("cogent: invalid or copied frame metrics")
	ErrFrameMetricsBudget   = errors.New("cogent: frame metrics capacity exhausted")
	ErrFrameMetricsBusy     = errors.New("cogent: traced interaction pending")
	ErrFrameMetricsDisabled = errors.New("cogent: traced control disabled or out of scope")
)

// FrameSample links one routed scroll to owner work and its first armed native
// composition. Durations measure elapsed CPU-side work, including GC pauses;
// they are not OS thread CPU counters, GPU execution or actual presentation time.
type FrameSample struct {
	Tag                                                     uint64
	Window                                                  string
	Generation                                              uint64
	QueuedNS, InputNS, OwnerNS, ReconcileNS                 int64
	LayoutPaintNS, GatherNS, RasterUploadNS, EncodeSubmitNS int64
	AcquireNS, PresentCallNS, TotalUIWorkNS                 int64
	Completed                                               bool
	Failure                                                 string
	SourceDetails                                           *FrameSourceDetails `json:"SourceDetails,omitempty"`
}
type FrameMetricsSnapshot struct {
	Samples []FrameSample
	Dropped uint64
	Pending int
	Failure string
}

// FrameMetricsPoll is one detached ticket plus recorder-wide refusal/pending
// status. Found distinguishes a completed or refused record from an unknown or
// still-pending tag. Poll never removes records from the retained history.
type FrameMetricsPoll struct {
	Sample  FrameSample
	Found   bool
	Dropped uint64
	Pending int
	Failure string
}
type frameWindowKey struct {
	owner  *app.Application
	window string
	ref    app.WindowRef
}
type frameWindowMetrics struct {
	metrics *FrameMetrics
	window  *nativeWindow
	key     frameWindowKey
	trace   *core.FrameTrace
	tag     uint64      // metrics.mu
	stop    func() bool // metrics.mu
	retired bool        // metrics.mu
}

// FrameMetrics is optional bounded diagnostic storage. Admission refuses when
// full; it never overwrites samples or silently drops them. This per-window
// work attribution requires one active native window. Up to 64 sequential
// window lifetimes may bind one recorder. Snapshot returns detached values.
type FrameMetrics struct {
	self          *FrameMetrics
	mu            sync.Mutex
	capacity      int
	samples       []FrameSample
	tags          map[uint64]int // reserved=0; collected=sample index+1
	windows       []*frameWindowMetrics
	dropped       uint64
	sourceDetails bool               // immutable configuration
	gridInputs    *gridInputRecorder // immutable optional diagnostic configuration
}

var frameMetricsRegistry = struct {
	sync.Mutex
	windows map[frameWindowKey]*frameWindowMetrics
}{windows: map[frameWindowKey]*frameWindowMetrics{}}

func NewFrameMetrics(capacity int) (*FrameMetrics, error) {
	return newFrameMetrics(capacity, false)
}

// NewFrameMetricsWithSourceDetails opts into per-source timing and image API
// attempt counters. Its additional clocks affect elapsed time; these samples
// are diagnostic and cannot qualify ordinary frame latency targets.
func NewFrameMetricsWithSourceDetails(capacity int) (*FrameMetrics, error) {
	return newFrameMetrics(capacity, true)
}

func newFrameMetrics(capacity int, sourceDetails bool) (*FrameMetrics, error) {
	limit := 65536
	if sourceDetails {
		limit = MaxSourceFrameMetricsCapacity
	}
	if capacity < 1 || capacity > limit {
		return nil, ErrFrameMetricsConfig
	}
	m := &FrameMetrics{capacity: capacity, sourceDetails: sourceDetails, samples: make([]FrameSample, 0, capacity), tags: make(map[uint64]int)}
	m.self = m
	return m, nil
}
func (m *FrameMetrics) valid() bool {
	if m == nil || m.self != m {
		return false
	}
	limit := 65536
	if m.sourceDetails || m.gridInputs != nil {
		limit = MaxSourceFrameMetricsCapacity
	}
	return m.capacity >= 1 && m.capacity <= limit
}
func (m *FrameMetrics) bind(a *app.Application, w *nativeWindow) (*frameWindowMetrics, error) {
	if !m.valid() {
		return nil, ErrFrameMetricsConfig
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.windows) >= 64 {
		return nil, ErrFrameMetricsBudget
	}
	if len(w.window.ID) > 4096 {
		return nil, ErrFrameMetricsConfig
	}
	for _, bound := range m.windows {
		if !bound.retired {
			return nil, ErrFrameMetricsConfig
		}
	}
	key := frameWindowKey{a, strings.Clone(w.window.ID), w.window.Ref()}
	frameMetricsRegistry.Lock()
	defer frameMetricsRegistry.Unlock()
	if frameMetricsRegistry.windows[key] != nil {
		return nil, ErrFrameMetricsBusy
	}
	b := &frameWindowMetrics{metrics: m, window: w, key: key, trace: &core.FrameTrace{}}
	if m.sourceDetails {
		b.trace = core.NewFrameTraceWithSourceDetails()
	}
	m.windows = append(m.windows, b)
	frameMetricsRegistry.windows[key] = b
	w.body.Scene.SetFrameTrace(b.trace)
	return b, nil
}
func (m *FrameMetrics) collectLocked() {
	for _, window := range m.windows {
		if sample, ok := window.trace.Take(); ok {
			if window.stop != nil {
				window.stop()
				window.stop = nil
			}
			window.tag = 0
			if len(m.samples) >= m.capacity {
				m.dropped++
				continue
			}
			m.samples = append(m.samples, FrameSample{Tag: sample.Tag, Window: window.key.window, Generation: sample.Generation,
				QueuedNS: sample.QueuedNS, InputNS: sample.InputNS, OwnerNS: sample.OwnerNS, ReconcileNS: sample.ReconcileNS,
				LayoutPaintNS: sample.LayoutPaintNS, GatherNS: sample.GatherNS, RasterUploadNS: sample.RasterUploadNS,
				EncodeSubmitNS: sample.EncodeSubmitNS, AcquireNS: sample.AcquireNS, PresentCallNS: sample.PresentCallNS,
				TotalUIWorkNS: sample.TotalUIWorkNS, Completed: sample.Completed, Failure: sample.Failure,
				SourceDetails: frameSourceDetails(sample.SourceDetails)})
			m.tags[sample.Tag] = len(m.samples)
		}
	}
}

// Poll collects available native completion through the same path as Snapshot,
// then copies only the requested record. Ordinary records require no observer
// allocation; diagnostic source details are separately detached when present.
// Admission, retained history and global drop/pending checks remain unchanged.
func (m *FrameMetrics) Poll(tag uint64) FrameMetricsPoll {
	if !m.valid() || tag == 0 {
		return FrameMetricsPoll{Failure: ErrFrameMetricsConfig.Error()}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.collectLocked()
	out := FrameMetricsPoll{Dropped: m.dropped}
	for _, window := range m.windows {
		if window.tag != 0 {
			out.Pending++
		}
	}
	if index := m.tags[tag]; index > 0 {
		if index > len(m.samples) {
			out.Failure = ErrFrameMetricsConfig.Error()
			return out
		}
		out.Sample, out.Found = m.samples[index-1], true
		out.Sample.SourceDetails = cloneFrameSourceDetails(out.Sample.SourceDetails)
	}
	return out
}
func (m *FrameMetrics) Snapshot() FrameMetricsSnapshot {
	if !m.valid() {
		return FrameMetricsSnapshot{Failure: ErrFrameMetricsConfig.Error()}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.collectLocked()
	out := FrameMetricsSnapshot{Samples: append([]FrameSample(nil), m.samples...), Dropped: m.dropped}
	for i := range out.Samples {
		out.Samples[i].SourceDetails = cloneFrameSourceDetails(out.Samples[i].SourceDetails)
	}
	for _, window := range m.windows {
		if window.tag != 0 {
			out.Pending++
		}
	}
	return out
}
func (b *frameWindowMetrics) retire() {
	if b == nil {
		return
	}
	m := b.metrics
	m.mu.Lock()
	defer m.mu.Unlock()
	if b.retired {
		return
	}
	b.retired = true
	if b.tag != 0 {
		b.trace.Retire(b.tag, "native window retired")
	}
	m.collectLocked()
	b.window.body.Scene.SetFrameTrace(nil)
	frameMetricsRegistry.Lock()
	if frameMetricsRegistry.windows[b.key] == b {
		delete(frameMetricsRegistry.windows, b.key)
	}
	frameMetricsRegistry.Unlock()
	b.window = nil
	b.key.owner = nil
}

// EnqueueGridScroll must run on the application owner (Do/PostWait). It queues
// an initialized actual MouseScroll in the originating native FIFO, using the
// completed grid geometry and logical-pixel deltas. It does not call Scroll or
// descriptor callbacks. Only one traced event per window may be outstanding.
// Context cancellation retires measurement; it cannot retract delivered input.
func EnqueueGridScroll(ctx context.Context, a *app.Application, model *virtual.Grid, dx, dy float64, tag uint64) error {
	if ctx == nil || a == nil || model == nil || tag == 0 || math.IsNaN(dx) || math.IsNaN(dy) || math.IsInf(dx, 0) || math.IsInf(dy, 0) || math.Abs(dx) > 1<<20 || math.Abs(dy) > 1<<20 || dx == 0 && dy == 0 {
		return ErrFrameMetricsConfig
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	gridRegistry.Lock()
	g := gridRegistry.models[model]
	gridRegistry.Unlock()
	if g == nil || g.owner != a || !g.live() {
		return ErrGridClosed
	}
	if !g.eligible() {
		return ErrFrameMetricsDisabled
	}
	// The ledger mutex is acquired only after metrics.mu is released. Native
	// delivery can append first; Tag, not append order, links the admission.
	var admission *GridInputRecord
	if g.inputs != nil {
		defer func() {
			if admission != nil {
				g.inputs.record(*admission)
			}
		}()
	}
	s, err := model.State()
	if err != nil {
		return err
	}
	measurement := g.currentMeasurement()
	if measurement == nil || measurement.Generation != s.Generation || measurement.Stamp != s.Stamp || measurement.Viewport != s.Viewport || !validGridScale(measurement.Scale) {
		return ErrGridLayout
	}
	bounds := measurement.Bounds
	position := image.Pt((bounds.Min.X+bounds.Max.X)/2, (bounds.Min.Y+bounds.Max.Y)/2)
	if bounds.Empty() || position.Y < bounds.Min.Y+int(float32(s.Config.HeaderHeight)*measurement.Scale) {
		return ErrGridLayout
	}
	frameMetricsRegistry.Lock()
	b := frameMetricsRegistry.windows[frameWindowKey{a, g.window, g.windowRef}]
	frameMetricsRegistry.Unlock()
	if b == nil || !b.metrics.valid() {
		return ErrFrameMetricsConfig
	}
	m := b.metrics
	m.mu.Lock()
	defer m.mu.Unlock()
	m.collectLocked()
	if b.retired || b.window.closing.Load() {
		return ErrGridClosed
	}
	if b.tag != 0 {
		return ErrFrameMetricsBusy
	}
	if _, reserved := m.tags[tag]; reserved {
		return ErrFrameMetricsConfig
	}
	if len(m.tags) >= m.capacity {
		return ErrFrameMetricsBudget
	}
	source := b.window.source.Load()
	if source == nil {
		return ErrGridLayout
	}
	event := events.NewScroll(position, math32.Vector2{X: float32(dx) * measurement.Scale, Y: float32(dy) * measurement.Scale}, 0)
	event.Init()
	if err := b.trace.Begin(tag, event); err != nil {
		return err
	}
	b.tag = tag
	m.tags[tag] = 0
	b.stop = context.AfterFunc(ctx, func() { b.trace.Fail(tag, ctx.Err().Error()) })
	// A deferred node owns its FIFO position even when a later ordinary scroll
	// would normally replace/combine the preceding same-type event.
	if g.inputs != nil {
		admission = &GridInputRecord{Kind: "enqueue", Tag: tag, Scale: measurement.Scale, RequestedX: dx, RequestedY: dy,
			PhysicalX: event.Delta.X, PhysicalY: event.Delta.Y, GenerationBefore: s.Generation, ViewportBefore: s.Viewport}
	}
	source.Deque.Send(&tracedScrollEvent{MouseScroll: event})
	return nil
}

type tracedScrollEvent struct{ *events.MouseScroll }

func (e *tracedScrollEvent) Ready() bool                     { return true }
func (e *tracedScrollEvent) Merge(events.DeferredEvent) bool { return false }
func (e *tracedScrollEvent) FrameTraceEvent() events.Event   { return e.MouseScroll }

func tracedGridIntent(g *gridBinding) (*core.FrameTrace, uint64) {
	if g.host == nil || g.host.Scene == nil {
		return nil, 0
	}
	trace := g.host.Scene.FrameTrace()
	return trace, trace.CurrentTag()
}
func (b *frameWindowMetrics) ownerStart() (uint64, time.Time) {
	if b == nil {
		return 0, time.Time{}
	}
	tag := b.trace.OwnerTag()
	if tag == 0 {
		return 0, time.Time{}
	}
	return tag, time.Now()
}
func (b *frameWindowMetrics) ownerEnd(tag uint64, start time.Time) {
	if b != nil && tag != 0 {
		b.trace.OwnerWork(tag, time.Since(start).Nanoseconds())
	}
}
func (b *frameWindowMetrics) reconcileStart() (uint64, time.Time) {
	if b == nil {
		return 0, time.Time{}
	}
	tag := b.trace.ReconcileTag()
	if tag == 0 {
		return 0, time.Time{}
	}
	return tag, time.Now()
}
func (b *frameWindowMetrics) reconcileEnd(tag uint64, start time.Time) {
	if b != nil && tag != 0 {
		b.trace.Reconciled(tag, time.Since(start).Nanoseconds())
	}
}
