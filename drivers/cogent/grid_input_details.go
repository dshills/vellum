package cogent

import (
	"cogentcore.org/core/tree"
	"errors"
	"github.com/dshills/vellum/app"
	"strings"
	"sync"

	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

// GridInputRecord contains copied scalar provenance, never an event, model,
// widget or callback. Tag zero means untraced; it does not identify a producer.
// Sequence orders recorder appends, not native delivery or wall-clock time.
type GridInputRecord struct {
	Sequence, Input, Tag, Incarnation  uint64
	Window, Grid, Kind, Origin, Result string
	Reason                             ui.GridChangeReason
	Scale                              float32
	RequestedX, RequestedY             float64
	PhysicalX, PhysicalY               float32
	RemainderBefore, RemainderAfter    [2]float64
	DeltaX, DeltaY                     int
	GenerationBefore, GenerationAfter  uint64
	ViewportBefore, ViewportAfter      virtual.GridViewport
}

type GridInputDetails struct {
	Capacity int
	Records  []GridInputRecord
	Dropped  uint64
	Failure  string
}

type gridInputRecorder struct {
	mu       sync.Mutex
	capacity int
	records  []GridInputRecord
	dropped  uint64
	failure  string
}

type gridInputBinding struct {
	recorder     *gridInputRecorder
	window, grid string
	incarnation  uint64
}

// NewFrameMetricsWithGridInputDetails records admissions, actual wheel
// arithmetic and owner transitions, including untraced input and layout.
// Extra synchronization makes these diagnostic samples ineligible for ordinary
// latency qualification. The ordinary constructor allocates no input ledger.
func NewFrameMetricsWithGridInputDetails(capacity int) (*FrameMetrics, error) {
	if capacity < 1 || capacity > MaxSourceFrameMetricsCapacity {
		return nil, ErrFrameMetricsConfig
	}
	m, err := newFrameMetrics(capacity, false)
	if err != nil {
		return nil, err
	}
	n := 3*capacity + 256
	m.gridInputs = &gridInputRecorder{capacity: n, records: make([]GridInputRecord, 0, n)}
	return m, nil
}

// GridInputDetails returns a detached snapshot. A full ledger refuses further
// records and reports loss; it never overwrites evidence or alters grid input.
func (m *FrameMetrics) GridInputDetails() GridInputDetails {
	if !m.valid() {
		return GridInputDetails{Failure: ErrFrameMetricsConfig.Error()}
	}
	r := m.gridInputs
	if r == nil {
		return GridInputDetails{}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	out := GridInputDetails{Capacity: r.capacity, Records: append([]GridInputRecord(nil), r.records...), Dropped: r.dropped, Failure: r.failure}
	if r.dropped != 0 && out.Failure == "" {
		out.Failure = "cogent: grid input details capacity exhausted"
	}
	return out
}

func bindGridInputDetails(g *gridBinding) *gridInputBinding {
	frameMetricsRegistry.Lock()
	b := frameMetricsRegistry.windows[frameWindowKey{g.owner, g.window, g.windowRef}]
	frameMetricsRegistry.Unlock()
	if b == nil || b.metrics.gridInputs == nil {
		return nil
	}
	if len(g.r.node.ID) > 4096 {
		r := b.metrics.gridInputs
		r.mu.Lock()
		r.failure = "cogent: grid input identifier exceeds4096 bytes"
		r.mu.Unlock()
		return nil
	}
	return &gridInputBinding{recorder: b.metrics.gridInputs, window: strings.Clone(g.window), grid: strings.Clone(g.r.node.ID), incarnation: g.incarnation}
}

func (b *gridInputBinding) record(v GridInputRecord) uint64 {
	if b == nil {
		return 0
	}
	r := b.recorder
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.records) >= r.capacity {
		if r.dropped != ^uint64(0) {
			r.dropped++
		}
		return 0
	}
	v.Sequence = uint64(len(r.records)) + 1
	v.Window, v.Grid, v.Incarnation = b.window, b.grid, b.incarnation
	r.records = append(r.records, v)
	return v.Sequence
}

type gridInputOrigin struct {
	name          string
	sequence, tag uint64
}

func (g *gridBinding) observeTransition(origin gridInputOrigin, reason ui.GridChangeReason, before, after virtual.GridState, result string) {
	if g.inputs != nil {
		g.inputs.record(GridInputRecord{Kind: "transition", Origin: origin.name, Input: origin.sequence, Tag: origin.tag, Reason: reason, Result: result,
			GenerationBefore: before.Generation, GenerationAfter: after.Generation, ViewportBefore: before.Viewport, ViewportAfter: after.Viewport})
	}
}

// Diagnostic equivalent of postControl: preserve its retirement/refusal
// behavior while making an absent delivery distinguishable from admission.
func (g *gridBinding) postInputControl(origin gridInputOrigin, reason ui.GridChangeReason, fn func()) {
	r := g.r
	if tree.IsNil(r.widget) {
		g.observeTransition(origin, reason, virtual.GridState{}, virtual.GridState{}, "widget-refused")
		return
	}
	if err := r.control.a.PostWindow(r.control.windowRef, func() {
		if !r.control.retired {
			fn()
		} else {
			g.observeTransition(origin, reason, virtual.GridState{}, virtual.GridState{}, "retired")
		}
	}); err != nil {
		g.observeTransition(origin, reason, virtual.GridState{}, virtual.GridState{}, "post-refused")
		if errors.Is(err, app.ErrClosed) || errors.Is(err, app.ErrWindowClosed) {
			return
		}
		r.control.refused.Store(&err)
		r.control.a.Invalidate(r.control.window)
	}
}
