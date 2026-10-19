package cogent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"math"
	"slices"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/keymap"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/placement"
	"github.com/dshills/vellum/ui"
)

var (
	ErrDesktopSession = errors.New("cogent: invalid desktop session")
	ErrDesktopBusy    = errors.New("cogent: desktop operation pending")
	ErrDesktopClosed  = errors.New("cogent: desktop session or window closed")
	ErrDesktopPending = errors.New("cogent: desktop window needs a completed frame")
	ErrDesktopBudget  = errors.New("cogent: desktop inspection budget exceeded")
)

// DesktopSession is an opt-in, single-consumer integration capability installed
// on Driver. It routes initialized Go events through the real native FIFO and
// waits for actual compositions. It does not synthesize OS input, drive an IME,
// control the native scheduler, or prove physical input/presentation. Do not call
// its blocking methods while holding the application owner or a native callback.
// One request and at most 64 live windows are retained. A session runs once.
type DesktopSession struct {
	self         *DesktopSession
	mu           sync.Mutex
	owner        *app.Application
	used, closed bool
	windows      map[app.WindowRef]*desktopWindow
	pending      *desktopRequest
	next         uint64
	measurements *desktopMeasurements
}
type desktopWindow struct {
	window                             *nativeWindow
	trace                              *core.FrameTrace
	source                             events.Source // own event-generation state; never mutate OS Source.Last
	revision, completedRevision, frame uint64
}
type desktopRequest struct {
	session  *DesktopSession
	binding  *desktopWindow
	ref      app.WindowRef
	ctx      context.Context
	kind     uint8
	tag      uint64
	input    DesktopInput
	reply    chan desktopReply
	done     atomic.Bool
	measured *desktopMeasuredRequest
}
type desktopReply struct {
	snapshot    DesktopSnapshot
	visibility  DesktopWindowVisibility
	err         error
	measurement DesktopInputMeasurement
}
type desktopFIFO struct {
	events.Event
	request *desktopRequest
}

func (*desktopFIFO) Ready() bool                     { return true }
func (*desktopFIFO) Merge(events.DeferredEvent) bool { return false }
func (e *desktopFIFO) FrameTraceEvent() events.Event { return e.Event }

// DesktopInput coordinates and deltas are scene pixels. Frame is the latest
// completed snapshot capability for the original WindowRef. Target is an
// optional exact DesktopElement.ID; it never resolves a reused domain name.
// Cancellation removes undispatched request-owned FIFO nodes. Already dequeued
// input may have effects and cannot be retracted; close the fixture on failure.
type DesktopInput struct {
	Frame     uint64
	Target    string
	Type      events.Types
	Rune      rune
	Code      key.Codes
	Modifiers key.Modifiers
	Button    events.Buttons
	Position  image.Point
	Delta     math32.Vector2
}
type DesktopElement struct {
	ID, DomainID, Parent, Role, Name string
	Value                            json.RawMessage
	Bounds                           image.Rectangle
	Enabled, Focused, Selected       bool
}
type DesktopSnapshot struct {
	Frame      uint64
	Size       image.Point
	LogicalDPI float32
	Elements   []DesktopElement
	Grids      []GridInspection
	Trees      []TreeInspection
	Canvases   []DesktopCanvasInspection
	Popups     []DesktopPopup
}

// DesktopPopup contains actual current popup geometry, separate from persistent
// semantic option nodes. IDs include frame, Scene and Button incarnations; they
// are not application domain IDs. Only top eligible direct standard Buttons are
// enabled targets. Custom/nested menus keep their existing semantic inspectors.
type DesktopPopup struct {
	ID, Kind    string
	Bounds      image.Rectangle
	Interactive bool
	Items       []DesktopElement
}

func desktopTargetElement(out DesktopSnapshot, id string) (DesktopElement, bool) {
	for _, el := range out.Elements {
		if el.ID == id {
			return el, true
		}
	}
	for _, p := range out.Popups {
		for _, el := range p.Items {
			if el.ID == id {
				return el, true
			}
		}
	}
	return DesktopElement{}, false
}

// DesktopCanvasInspection keeps the existing annotation count separate from
// actual painted objects and accepted placement preview. All values are detached.
type DesktopCanvasInspection struct {
	CanvasInspection
	Objects  []canvas.Rect
	Selected string
	Preview  placement.State
}

func NewDesktopSession() *DesktopSession {
	s := &DesktopSession{windows: make(map[app.WindowRef]*desktopWindow)}
	s.self = s
	return s
}
func (s *DesktopSession) start(a *app.Application) error {
	if s == nil || s.self != s || a == nil {
		return ErrDesktopSession
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.used {
		return ErrDesktopSession
	}
	s.used = true
	s.owner = a
	return nil
}
func (s *DesktopSession) bind(w *nativeWindow) error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.owner == nil || len(s.windows) >= 64 {
		return ErrDesktopBudget
	}
	ref := w.window.Ref()
	if s.windows[ref] != nil {
		return ErrDesktopSession
	}
	b := &desktopWindow{window: w, trace: &core.FrameTrace{}, revision: 1}
	s.windows[ref] = b
	w.desktop = s
	w.body.Scene.SetFrameTrace(b.trace)
	return nil
}
func (s *DesktopSession) invalidate(w *nativeWindow) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if b := s.windows[w.window.Ref()]; b != nil {
		if r := s.pending; r != nil && r.measured != nil && r.binding == b && r.measured.accepted != 0 {
			// A new view cannot paint under the earlier armed generation. Fail only
			// pending work; a completed historic composition remains immutable.
			b.trace.Fail(r.tag, "measured view invalidated after final arming")
		}
		if b.revision == math.MaxUint64 {
			s.closed = true
		} else {
			b.revision++
		}
	}
}
func (s *DesktopSession) retire(w *nativeWindow) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	b := s.windows[w.window.Ref()]
	if b == nil {
		return
	}
	if r := s.pending; r != nil && r.binding == b {
		b.trace.Retire(r.tag, "native window retired")
		if sample, ok := b.trace.Take(); ok && r.measured != nil {
			s.completeMeasuredLocked(r, sample)
		}
		s.finishLocked(r, desktopReply{err: ErrDesktopClosed})
	}
	w.body.Scene.SetOriginCustomHandler(nil)
	w.body.Scene.SetFrameTrace(nil)
	b.window = nil
	delete(s.windows, w.window.Ref())
	w.desktop = nil
}
func (s *DesktopSession) close() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	if r := s.pending; r != nil {
		r.binding.trace.Retire(r.tag, "native driver retired")
		if sample, ok := r.binding.trace.Take(); ok && r.measured != nil {
			s.completeMeasuredLocked(r, sample)
		}
		s.finishLocked(r, desktopReply{err: ErrDesktopClosed})
	}
	for ref, b := range s.windows {
		if b.window != nil {
			b.window.body.Scene.SetOriginCustomHandler(nil)
			b.window.body.Scene.SetFrameTrace(nil)
			b.window.desktop = nil
		}
		b.window = nil
		delete(s.windows, ref)
	}
	s.owner = nil
}
func (s *DesktopSession) finishLocked(r *desktopRequest, reply desktopReply) {
	if r.done.Swap(true) {
		return
	}
	if s.pending == r {
		s.pending = nil
	}
	if r.measured != nil {
		r.measured.check = nil
		if !r.measured.completed {
			reply.measurement = s.refuseMeasuredLocked(r, reply.err)
		} else {
			reply.measurement = s.measurements.samples[r.measured.index]
		}
	}
	r.reply <- reply
}
func (s *DesktopSession) reapLocked() {
	r := s.pending
	if r == nil || r.kind != 2 && r.kind != 6 {
		return
	}
	sample, ok := r.binding.trace.Take()
	if !ok {
		return
	}
	reply := desktopReply{}
	if r.measured != nil {
		r.measured.check = nil
		reply.measurement = s.completeMeasuredLocked(r, sample)
	}
	if err := r.ctx.Err(); err != nil {
		reply.err = err
	} else if !sample.Completed || sample.Failure != "" || sample.Unavailable {
		reply.err = &DesktopCompositionError{Failure: sample.Failure, Unavailable: sample.Unavailable && !sample.Completed && sample.Failure != ""}
	} else {
		r.binding.frame = r.tag
		if r.measured != nil {
			r.binding.completedRevision = sample.Generation
		}
		reply.snapshot.Frame = r.tag
	}
	s.finishLocked(r, reply)
}
func (s *DesktopSession) submit(ctx context.Context, ref app.WindowRef, kind uint8, input DesktopInput) (*desktopRequest, error) {
	return s.submitWithExpectation(ctx, ref, kind, input, nil)
}
func (s *DesktopSession) submitWithExpectation(ctx context.Context, ref app.WindowRef, kind uint8, input DesktopInput, check DesktopInputExpectation) (*desktopRequest, error) {
	if s == nil || s.self != s || ctx == nil || ref == (app.WindowRef{}) {
		return nil, ErrDesktopSession
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reapLocked()
	if !s.used || s.closed || s.owner == nil {
		return nil, ErrDesktopClosed
	}
	if s.pending != nil {
		return nil, ErrDesktopBusy
	}
	b := s.windows[ref]
	if b == nil || b.window == nil || ref.Context().Err() != nil || b.window.closing.Load() {
		return nil, ErrDesktopClosed
	}
	source := b.window.source.Load()
	if source == nil {
		return nil, ErrDesktopPending
	}
	if s.next == math.MaxUint64 {
		return nil, ErrDesktopBudget
	}
	if kind == 5 && (s.measurements == nil || check == nil) {
		return nil, ErrDesktopSession
	}
	if kind == 5 && len(s.measurements.samples) >= s.measurements.capacity {
		return nil, ErrDesktopBudget
	}
	s.next++
	r := &desktopRequest{session: s, binding: b, ref: ref, ctx: ctx, kind: kind, tag: s.next, input: input, reply: make(chan desktopReply, 1)}
	if kind == 5 {
		r.measured = &desktopMeasuredRequest{index: len(s.measurements.samples), check: check, before: b.revision}
		s.measurements.samples = append(s.measurements.samples, DesktopInputMeasurement{Frame: FrameSample{Tag: r.tag, Window: b.window.window.ID}, ViewRevisionBefore: b.revision})
	}
	event := &events.CustomEvent{}
	event.Data = r
	event.Typ = events.Custom
	event.Init()
	if kind == 2 {
		if err := b.trace.Begin(r.tag, event); err != nil {
			return nil, err
		}
	}
	if kind == 2 {
		b.frame = 0
	}
	s.pending = r
	source.Deque.Send(&desktopFIFO{Event: event, request: r})
	return r, nil
}
func (s *DesktopSession) cancel(r *desktopRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pending != r || r.done.Load() {
		return
	}
	if r.kind == 6 {
		r.measured.check = nil
		r.binding.trace.Fail(r.tag, "measured input canceled")
		s.reapLocked()
		return
	}
	if r.kind == 2 {
		r.binding.trace.Fail(r.tag, "desktop observation canceled")
		s.reapLocked()
		return
	}
	if n := r.binding.window; n != nil {
		if source := n.source.Load(); source != nil {
			_ = source.Deque.UpdateDeferred(func() error { return nil }, func(e events.DeferredEvent) bool { x, ok := e.(*desktopFIFO); return ok && x.request == r })
		}
	}
	s.finishLocked(r, desktopReply{err: r.ctx.Err()})
}
func (s *DesktopSession) wait(r *desktopRequest) (DesktopSnapshot, error) {
	reply := s.waitReply(r)
	return reply.snapshot, reply.err
}
func (s *DesktopSession) waitReply(r *desktopRequest) desktopReply {
	tick := time.NewTicker(2 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case out := <-r.reply:
			return out
		case <-r.ctx.Done():
			s.cancel(r)
			return desktopReply{err: r.ctx.Err()}
		case <-r.ref.Context().Done():
			return desktopReply{err: ErrDesktopClosed}
		case <-tick.C:
			s.mu.Lock()
			s.reapLocked()
			s.mu.Unlock()
		}
	}
}

// Frame forces a paint after preceding FIFO input and owner reconciliation, and
// waits for its completed native composition. It is not an OS presentation fence.
// A checked drawable-acquisition yield returns a DesktopCompositionError matching
// ErrDesktopFrameUnavailable. It publishes no frame capability and does not retry.
// A caller may make a new request after a meaningful lifecycle/application event;
// neither a shown/focused window nor this error proves drawable readiness.
func (s *DesktopSession) Frame(ctx context.Context, ref app.WindowRef) error {
	r, e := s.submit(ctx, ref, 2, DesktopInput{})
	if e != nil {
		return e
	}
	_, e = s.wait(r)
	return e
}

// Visibility observes the latest native Mac occlusion state after preceding FIFO
// work. It does not require or publish a completed frame, acquire a drawable,
// wait for visibility, or retry Frame. Known is false for an unsupported driver
// (including Manual). A later Frame may still return ErrDesktopFrameUnavailable.
func (s *DesktopSession) Visibility(ctx context.Context, ref app.WindowRef) (DesktopWindowVisibility, error) {
	r, err := s.submit(ctx, ref, 4, DesktopInput{})
	if err != nil {
		return DesktopWindowVisibility{}, err
	}
	reply := s.waitReply(r)
	return reply.visibility, reply.err
}

// Snapshot reads settled geometry/semantics without creating a frame. Call Frame
// first, and repeat it when asynchronous application work invalidates that frame.
func (s *DesktopSession) Snapshot(ctx context.Context, ref app.WindowRef) (DesktopSnapshot, error) {
	r, e := s.submit(ctx, ref, 0, DesktopInput{})
	if e != nil {
		return DesktopSnapshot{}, e
	}
	return s.wait(r)
}
func (s *DesktopSession) Input(ctx context.Context, ref app.WindowRef, input DesktopInput) error {
	if err := validateDesktopInput(input); err != nil {
		return err
	}
	r, e := s.submit(ctx, ref, 1, input)
	if e != nil {
		return e
	}
	_, e = s.wait(r)
	return e
}
func validateDesktopInput(in DesktopInput) error {
	if in.Frame == 0 || len(in.Target) > 4096 || !utf8.ValidRune(in.Rune) || in.Modifiers < 0 || in.Modifiers & ^(key.Modifiers(1<<key.ModifiersN)-1) != 0 || in.Position.X < math.MinInt32 || in.Position.X > math.MaxInt32 || in.Position.Y < math.MinInt32 || in.Position.Y > math.MaxInt32 {
		return ErrDesktopSession
	}
	if desktopUnqualifiedInput(in) {
		return ErrDesktopSession
	}
	switch in.Type {
	case events.MouseDown, events.MouseUp:
		if in.Button != events.Left || key.HasAnyModifier(in.Modifiers, key.Control) {
			return ErrDesktopSession
		}
	case events.MouseMove:
	case events.Scroll:
		if math.IsNaN(float64(in.Delta.X)) || math.IsNaN(float64(in.Delta.Y)) || math.IsInf(float64(in.Delta.X), 0) || math.IsInf(float64(in.Delta.Y), 0) || math.Abs(float64(in.Delta.X)) > 1e6 || math.Abs(float64(in.Delta.Y)) > 1e6 {
			return ErrDesktopSession
		}
	case events.KeyDown, events.KeyUp, events.KeyChord:
		valid := false
		for _, c := range key.CodesValues() {
			valid = valid || c == in.Code
		}
		if !valid {
			return ErrDesktopSession
		}
	default:
		return ErrDesktopSession
	}
	return nil
}

// Called only in the originating native event loop, after ordinary owner drain
// and refresh. No session mutex is held while acquiring the application owner.
func (s *DesktopSession) handle(w *nativeWindow, e events.Event) bool {
	ce, ok := e.(*events.CustomEvent)
	if !ok {
		return false
	}
	r, ok := ce.Data.(*desktopRequest)
	if !ok || r.session != s {
		return false
	}
	e.SetHandled()
	s.mu.Lock()
	current := s.pending == r && !r.done.Load() && r.binding.window == w
	owner := s.owner
	s.mu.Unlock()
	if !current {
		return true
	}
	b := r.binding
	if r.kind == 2 {
		b.trace.ExpectOwner(r.tag)
		b.trace.Transition(r.tag, r.tag)
		b.trace.OwnerWork(r.tag, 0)
		if r.ctx.Err() != nil {
			b.trace.Fail(r.tag, "desktop observation canceled")
		} else {
			s.mu.Lock()
			b.completedRevision = b.revision
			s.mu.Unlock()
			b.trace.Reconciled(r.tag, 0)
			w.body.Scene.NeedsRender()
		}
		return true
	}
	if r.kind == 4 {
		value, err := w.visibility.Load().snapshot()
		if canceled := r.ctx.Err(); canceled != nil {
			err = canceled
		}
		s.mu.Lock()
		s.finishLocked(r, desktopReply{visibility: value, err: err})
		s.mu.Unlock()
		return true
	}
	if r.kind == 6 {
		s.finishMeasuredInput(w, r)
		return true
	}
	if r.kind == 3 {
		s.mu.Lock()
		s.finishLocked(r, desktopReply{})
		s.mu.Unlock()
		return true
	}
	var out DesktopSnapshot
	var failure error
	err := withWindowOwner(owner, r.ref, func() {
		if r.ctx.Err() != nil {
			failure = r.ctx.Err()
			return
		}
		w.rootMu.Lock()
		defer w.rootMu.Unlock()
		s.mu.Lock()
		ready := b.frame != 0 && b.revision == b.completedRevision
		frame := b.frame
		s.mu.Unlock()
		if !ready || w.body.Scene.GlobalLayoutPending() {
			failure = ErrDesktopPending
			return
		}
		out, failure = desktopSnapshot(w, frame)
		if failure != nil {
			return
		}
		if r.kind == 1 || r.kind == 5 {
			if r.input.Frame != frame {
				failure = ErrDesktopPending
				return
			}
			if r.input.Target != "" {
				el, found := desktopTargetElement(out, r.input.Target)
				if !found || !el.Enabled || el.Bounds.Empty() {
					failure = ErrDesktopPending
				} else if r.input.Type.IsKey() {
					if !el.Focused {
						failure = ErrDesktopPending
					}
				} else if !r.input.Position.In(el.Bounds) {
					failure = ErrDesktopPending
				}
			}
		}
	})
	failure = errors.Join(err, failure)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pending != r || r.done.Load() {
		return true
	}
	if failure != nil {
		s.finishLocked(r, desktopReply{err: failure})
		return true
	}
	if r.kind == 0 {
		s.finishLocked(r, desktopReply{snapshot: out})
		return true
	}
	source := w.source.Load()
	if source == nil || r.ctx.Err() != nil || r.ref.Context().Err() != nil {
		s.finishLocked(r, desktopReply{err: ErrDesktopClosed})
		return true
	}

	in := r.input
	previousSource := b.source.Last
	previousFrame := b.frame
	b.frame = 0 // Native input can change focus or paint without an application dirty view.
	b.source.Last.Mods = in.Modifiers
	switch in.Type {
	case events.KeyDown, events.KeyUp, events.KeyChord:
		b.source.Key(in.Type, in.Rune, in.Code, in.Modifiers)
	case events.MouseDown, events.MouseUp:
		b.source.MouseButton(in.Type, in.Button, in.Position, in.Modifiers)
	case events.MouseMove:
		b.source.MouseMove(in.Position)
	case events.Scroll:
		b.source.Scroll(in.Position, in.Delta)
	}
	group := make([]events.Event, 0, 2)
	for {
		ev, ok := b.source.Deque.PollEvent()
		if !ok {
			break
		}
		group = append(group, ev)
	}
	if r.measured != nil {
		if err := s.beginMeasuredLocked(r, group); err != nil {
			b.source.Last = previousSource
			b.frame = previousFrame
			s.finishLocked(r, desktopReply{err: err})
			return true
		}
	}
	for _, ev := range group {
		source.Deque.Send(&desktopFIFO{Event: ev, request: r})
	}

	r.kind = 3
	if r.measured != nil {
		r.kind = 6
	}
	ack := &events.CustomEvent{}
	ack.Data = r
	ack.Typ = events.Custom
	ack.Init()
	source.Deque.Send(&desktopFIFO{Event: ack, request: r})
	return true
}
func desktopSnapshot(w *nativeWindow, frame uint64) (DesktopSnapshot, error) {
	out := DesktopSnapshot{Frame: frame, Size: w.body.Scene.SceneGeom.Size, LogicalDPI: w.body.Scene.Styles.UnitContext.DPI}
	nodes, _ := accessibilitySnapshot(w.root, nil)
	if len(nodes) > 8192 {
		return out, ErrDesktopBudget
	}
	bounds := accessibilityBounds(w.root)
	focus := map[string]accessibilityFocusTarget{}
	accessibilityNativeFocusTargets(w.root, "", focus)
	visible := map[string]bool{}
	elementIndex := map[string]int{}
	bytes := 0
	for _, n := range nodes {
		visible[n.ID] = true
		if target, ok := focus[n.ID]; ok {
			n.Focused = target.focused
		}
		if n.GridPending {
			return out, ErrDesktopPending
		}
		v, err := json.Marshal(n.Value)
		if err != nil {
			return out, err
		}
		bytes += len(v) + len(n.ID) + len(n.DomainID) + len(n.Parent) + len(n.Name)
		if len(v) > 1<<20 || bytes > 8<<20 {
			return out, ErrDesktopBudget
		}
		b := bounds[n.ID]
		if n.GridFrame != nil {
			b = *n.GridFrame
		}
		elementIndex[n.ID] = len(out.Elements)
		out.Elements = append(out.Elements, DesktopElement{ID: n.ID, DomainID: n.DomainID, Parent: n.Parent, Role: n.Role, Name: n.Name, Value: v, Bounds: image.Rect(int(b[0]), int(b[1]), int(b[0]+b[2]), int(b[1]+b[3])), Enabled: !n.Disabled, Focused: n.Focused, Selected: n.Selected})
	}

	var collect func(*record, string) error
	retained := 0
	collect = func(r *record, prefix string) error {
		if r == nil {
			return nil
		}
		if visible[prefix+accessibilityManagedID(r.node.ID)] && r.control != nil {
			if r.control.grid != nil {
				g, e := InspectGrid(r.control.a, r.node.Grid)
				if e != nil {
					return errors.Join(ErrDesktopPending, e)
				}
				retained += len(g.Cells) + len(g.Headers)
				if retained > 8192 {
					return ErrDesktopBudget
				}
				g.ID = prefix + g.ID
				out.Grids = append(out.Grids, g)
			}
			if r.control.outline != nil {
				t, e := InspectTree(r.control.a, r.node.Tree)
				if e != nil {
					return errors.Join(ErrDesktopPending, e)
				}
				retained += len(t.Rows) + len(t.Pages)
				if retained > 8192 {
					return ErrDesktopBudget
				}
				t.ID = prefix + t.ID
				out.Trees = append(out.Trees, t)
			}
			if r.control.placement != nil && r.control.placement.model != nil {
				b := r.control.placement
				f := b.frame.Load()
				if !canvasCompletedFrame(b, f) {
					return ErrDesktopPending
				}
				c := inspectCanvasFrame(b, f)
				retained += len(c.Objects) + len(f.picture.objects)
				if retained > 8192 {
					return ErrDesktopBudget
				}
				c.ID = prefix + c.ID
				out.Canvases = append(out.Canvases, DesktopCanvasInspection{CanvasInspection: c, Objects: slices.Clone(f.picture.objects), Selected: f.picture.selected, Preview: f.picture.preview})
			}
		}
		// Per-tab focus belongs to the actual retained header, rather than the
		// selected panel. Publish it only when the current semantic/modal scope
		// has admitted that option. Keep the native AX wire contract unchanged.
		if dock, ok := r.node.Custom.(*Workbench); r.node.Kind == ui.Extension && ok {
			for panel, tab := range dock.buttons {
				id := prefix + accessibilityDockOptionID(r.node.ID, panel)
				if index, admitted := elementIndex[id]; admitted && tab != nil {
					out.Elements[index].Focused = tab.ContainsFocus()
				}
			}
		}
		var failure error
		retainedWorkbenchRecords(r, func(panel string, child *record) {
			if failure == nil {
				failure = collect(child, prefix+accessibilityWorkbenchNamespace(r.node.ID, panel))
			}
		})
		if failure != nil {
			return failure
		}
		if edit := gridEditingRecord(r); edit != nil {
			if e := collect(edit, prefix); e != nil {
				return e
			}
		}
		for _, c := range r.children {
			if e := collect(c, prefix); e != nil {
				return e
			}
		}
		return nil
	}
	if e := collect(w.root, ""); e != nil {
		return out, e
	}
	popups, err := w.body.Scene.InspectPopups(8192-len(out.Elements), (8<<20)-bytes)
	if err != nil {
		if errors.Is(err, core.ErrPopupInspectionBudget) {
			return out, ErrDesktopBudget
		}
		return out, errors.Join(ErrDesktopPending, err)
	}
	remaining := (8 << 20) - bytes
	for _, p := range popups {
		v, err := desktopPopupInspection(p, frame, &remaining)
		if err != nil {
			return out, err
		}
		out.Popups = append(out.Popups, v)
	}

	return out, nil
}

// Charge the generated public representation too: Core's own bounded metadata
// budget does not account for these frame/parent IDs and DesktopElement fields.
// 512 per record conservatively reserves fixed records/slice overhead; strings
// are charged separately. This shares the existing 8MiB semantic output budget.
func desktopPopupInspection(p core.PopupInspection, frame uint64, remaining *int) (DesktopPopup, error) {
	id := fmt.Sprintf("popup:%d:%d", frame, p.ID)
	kind := p.Type.String()
	*remaining -= 512 + len(id) + len(kind)
	if *remaining < 0 {
		return DesktopPopup{}, ErrDesktopBudget
	}
	v := DesktopPopup{ID: id, Kind: kind, Bounds: p.Bounds, Interactive: p.Interactive}
	for _, item := range p.Items {
		itemID := fmt.Sprintf("%s:%d", id, item.ID)
		*remaining -= 512 + len(itemID) + len(id) + len("popup-item") + len(item.Name)
		if *remaining < 0 {
			return DesktopPopup{}, ErrDesktopBudget
		}
		v.Items = append(v.Items, DesktopElement{ID: itemID, Parent: id, Role: "popup-item", Name: item.Name, Bounds: item.Bounds, Enabled: item.Enabled, Focused: item.Focused, Selected: item.Selected})
	}
	return v, nil
}

// DesktopState is bounded terminal accounting, without retained window roots.
type DesktopState struct {
	Started, Closed, Pending bool
	LiveWindows              int
	Issued                   uint64
}

func (s *DesktopSession) State() DesktopState {
	if s == nil || s.self != s {
		return DesktopState{Closed: true}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reapLocked()
	return DesktopState{Started: s.used, Closed: s.closed, Pending: s.pending != nil, LiveWindows: len(s.windows), Issued: s.next}
}

func desktopUnqualifiedInput(in DesktopInput) bool {
	if !in.Type.IsKey() {
		return false
	}
	if (in.Code == key.CodeC || in.Code == key.CodeX || in.Code == key.CodeV) && key.HasAnyModifier(in.Modifiers, key.Control, key.Meta) {
		return true
	}
	switch keymap.Of(key.NewChord(in.Rune, in.Code, in.Modifiers)) {
	case keymap.Copy, keymap.Cut, keymap.Paste, keymap.PasteHist, keymap.WinSnapshot, keymap.ZoomIn, keymap.ZoomOut, keymap.Refresh, keymap.Find, keymap.Menu, keymap.WinFocusNext:
		return true
	}
	return false
}

func (s *DesktopSession) owns(w *nativeWindow, e events.Event) bool {
	ce, ok := e.(*events.CustomEvent)
	if !ok {
		return false
	}
	r, ok := ce.Data.(*desktopRequest)
	return ok && r.session == s && r.ref == w.window.Ref()
}
