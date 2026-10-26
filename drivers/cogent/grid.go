package cogent

import (
	"errors"
	"image"
	"math"
	"sync"
	"sync/atomic"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/keymap"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/text/text"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

var (
	ErrGridBinding = errors.New("cogent: grid already bound or wrong owner")
	ErrGridLayout  = errors.New("cogent: grid requires a current completed frame")
	ErrGridClosed  = errors.New("cogent: grid binding/window closed")
	ErrGridBudget  = errors.New("cogent: grid native pool budget exceeded")
)

const gridScrollbar = 12
const gridNativeMaxRows = 256
const gridNativeMaxColumns = 64
const gridNativeMaxCells = 4096
const gridNativeMaxText = 8 << 20
const gridNativeMaxDimension = 1 << 16
const gridNativeMaxCoordinate = 1 << 18
const gridNativeMaxScale = 32

var gridRegistry = struct {
	sync.Mutex
	models    map[*virtual.Grid]*gridBinding
	preparing map[*virtual.Grid]*gridBinding
	next      uint64
}{models: map[*virtual.Grid]*gridBinding{}, preparing: map[*virtual.Grid]*gridBinding{}}

type gridRequest struct {
	width, height int
	scale         float32
}

// Native tab topology is captured on the owner after retained reconciliation.
// Managed/workbench tabs are fully constructed and nonclosable; their topology
// retires on that same owner. Only CurrentTab (protected by Tabs.mu) is read
// when a queued intent is drained from another window.
type gridTabOwner struct {
	tabs *core.Tabs
	page core.Widget
}

type gridBinding struct {
	r                    *record
	previous             *gridBinding
	owner                *app.Application
	window               string
	windowRef            app.WindowRef
	lifetimeDone         <-chan struct{}
	model                *virtual.Grid
	incarnation          uint64
	epoch                uint64
	frames               uint64
	closed               atomic.Bool
	host                 *gridWidget
	panes                [4]*gridPane
	cellSurface          *gridCellSurface
	cells                []*core.Text
	headers              []*core.Text
	horizontal, vertical *core.Slider
	projection           virtual.GridSnapshot // originating native scene / owner refresh only
	projectionReady      bool                 // true only after a successful complete sync
	layout               gridLayoutStamp
	layoutDirty          map[*core.Text]bool
	layoutPending        atomic.Bool
	projectedDisabled    bool
	projectedHelp        string
	measurement          atomic.Pointer[GridInspection]
	measurementScene     atomic.Pointer[core.Scene]
	armed                *gridPress // originating scene only; detached opaque identity and scene geometry
	requestMu            sync.Mutex
	requested            gridRequest
	scope                bool           // owner-only active modal visibility
	tabOwners            []gridTabOwner // owner-only retained topology, no native geometry reads
	ancestorRefused      bool
	wheelX, wheelY       float64
	geometryRefused      bool
	editing              *gridEditBinding
	inputs               *gridInputBinding // immutable optional scalar diagnostic sink
}

func validateNativeGrid(n ui.Node) error {
	if n.Kind != ui.DataGrid {
		return nil
	}
	_, err := nativeGridState(n)
	return err
}
func nativeGridState(n ui.Node) (virtual.GridState, error) {
	if n.Grid == nil {
		return virtual.GridState{}, ErrGridBinding
	}
	s, err := n.Grid.State()
	if err != nil {
		return virtual.GridState{}, err
	}
	c := s.Config
	if s.Viewport.Width > gridNativeMaxDimension || s.Viewport.Height > gridNativeMaxDimension || !s.VariableHeights && c.RowHeight > gridNativeMaxDimension || c.HeaderHeight > gridNativeMaxDimension || c.MaxRows > gridNativeMaxRows || c.MaxColumns > gridNativeMaxColumns || c.MaxCells > gridNativeMaxCells || c.MaxTextBytes > gridNativeMaxText {
		return virtual.GridState{}, ErrGridBudget
	}
	return s, nil
}
func mountGrid(parent tree.Node, r *record, a *app.Application, window string) (core.Widget, error) {
	if err := validateNativeGrid(r.node); err != nil {
		return nil, err
	}
	life, err := scopedWindowContext(a, r.scope)
	if err != nil {
		return nil, err
	}
	gridRegistry.Lock()
	existing := gridRegistry.models[r.node.Grid]
	if existing != nil && (gridRegistry.preparing[r.node.Grid] != existing || existing.owner != a || existing.windowRef != r.scope) || gridRegistry.next == math.MaxUint64 {
		gridRegistry.Unlock()
		return nil, ErrGridBinding
	}
	gridRegistry.next++
	g := &gridBinding{r: r, owner: a, window: window, windowRef: r.scope, model: r.node.Grid, incarnation: gridRegistry.next, previous: existing, lifetimeDone: life.Done(), scope: true}
	gridRegistry.models[g.model] = g
	gridRegistry.Unlock()
	r.control.grid = g
	host := tree.New[gridWidget](parent)
	host.binding = g
	g.host = host
	g.inputs = bindGridInputDetails(g)
	for i := range g.panes {
		p := tree.New[gridPane](host)
		p.binding = g
		p.part = uint8(i)
		g.panes[i] = p
	}
	g.horizontal = core.NewSlider(host)
	g.vertical = core.NewSlider(host)
	g.horizontal.SetMin(0).SetMax(1).SetStep(.01)
	g.vertical.SetMin(0).SetMax(1).SetStep(.01)
	g.horizontal.Styler(func(s *styles.Style) { s.SetAbilities(false, abilities.Focusable) })
	g.vertical.Styler(func(s *styles.Style) { s.Direction = styles.Column; s.SetAbilities(false, abilities.Focusable) })
	g.horizontal.OnChange(func(events.Event) {
		value := g.horizontal.Value
		g.postObserved(ui.GridViewport, gridInputOrigin{name: "scrollbar-x"}, func() error {
			s, e := g.model.State()
			if e != nil {
				return e
			}
			v := s.Viewport
			v.ScrollX = normalizedGridOffset(value, gridMaxX(s))
			if v.ScrollX < 0 {
				return virtual.ErrGridInvalid
			}
			return g.model.SetViewport(v)
		})
	})
	g.vertical.OnChange(func(events.Event) {
		value := g.vertical.Value
		g.postObserved(ui.GridViewport, gridInputOrigin{name: "scrollbar-y"}, func() error {
			s, e := g.model.State()
			if e != nil {
				return e
			}
			v := s.Viewport
			v.ScrollY = normalizedGridOffset(value, gridMaxY(s))
			if v.ScrollY < 0 {
				return virtual.ErrGridInvalid
			}
			return g.model.SetViewport(v)
		})
	})
	host.On(events.MouseDown, func(e events.Event) {
		g.armed = nil
		if e.MouseButton() != events.Left || host.StateIs(states.Disabled) {
			return
		}
		if ref, ok := g.refAt(e.Pos()); ok {
			if g.projection.State.Config.RowSelection {
				mode, accepted := gridPointerSelectionMode(e.Modifiers())
				if !accepted {
					return
				}
				ref.rowSelection, ref.mode = true, mode
			}
			g.armed = &ref
			host.SetFocus()
			e.SetHandled()
		}
	})
	host.On(events.MouseUp, func(e events.Event) {
		ref := g.armed
		g.armed = nil
		if ref == nil || e.MouseButton() != events.Left {
			return
		}
		at, ok := g.refAt(e.Pos())
		// Selection intent belongs to the down event. Releasing Shift/Command,
		// or adding one before release, cannot reinterpret an admitted press.
		at.rowSelection, at.mode = ref.rowSelection, ref.mode
		if !ok || at != *ref {
			return
		}
		captured := ref.ref
		mode, rowSelection := ref.mode, ref.rowSelection
		g.post(ui.GridSelection, func() error {
			if rowSelection {
				return g.model.SelectCell(captured, mode)
			}
			return g.model.Activate(captured)
		})
		e.SetHandled()
	})
	host.On(events.KeyChord, func(e events.Event) {
		if g.gridEditBeginKey(e) {
			return
		}
		if g.projection.State.Config.RowSelection {
			g.selectionKey(e)
			return
		}
		rows, cols, page := 0, 0, 0
		switch keymap.Of(e.KeyChord()) {
		case keymap.MoveUp:
			rows = -1
		case keymap.MoveDown:
			rows = 1
		case keymap.MoveLeft:
			cols = -1
		case keymap.MoveRight:
			cols = 1
		case keymap.PageUp:
			page = -1
		case keymap.PageDown:
			page = 1
		default:
			return
		}
		g.post(ui.GridSelection, func() error {
			if page != 0 {
				return g.model.Page(page)
			}
			return g.model.Move(rows, cols)
		})
		e.SetHandled()
	})
	host.On(events.Scroll, func(e events.Event) {
		s, ok := e.(*events.MouseScroll)
		if !ok {
			return
		}
		scale := host.Styles.UnitContext.Px(1)
		if !validGridScale(scale) {
			return
		}
		if math.IsNaN(float64(s.Delta.X)) || math.IsNaN(float64(s.Delta.Y)) || math.IsInf(float64(s.Delta.X), 0) || math.IsInf(float64(s.Delta.Y), 0) {
			return
		}
		var observed GridInputRecord
		if g.inputs != nil {
			_, tag := tracedGridIntent(g)
			observed = GridInputRecord{Kind: "wheel", Tag: tag, Scale: scale, PhysicalX: s.Delta.X, PhysicalY: s.Delta.Y, RemainderBefore: [2]float64{g.wheelX, g.wheelY}}
		}
		g.wheelX += float64(s.Delta.X) / float64(scale)
		g.wheelY += float64(s.Delta.Y) / float64(scale)
		dx, dy := boundedWheel(g.wheelX), boundedWheel(g.wheelY)
		g.wheelX -= math.Trunc(g.wheelX)
		g.wheelY -= math.Trunc(g.wheelY)
		origin := gridInputOrigin{name: "wheel"}
		if g.inputs != nil {
			observed.DeltaX, observed.DeltaY = dx, dy
			observed.RemainderAfter = [2]float64{g.wheelX, g.wheelY}
			origin.sequence = g.inputs.record(observed)
		}
		g.postObserved(ui.GridViewport, origin, func() error { return g.model.Scroll(dx, dy) })
		e.SetHandled()
	})
	return host, nil
}

func gridSemanticModifiers(mods key.Modifiers) key.Modifiers {
	mods.SetFlag(false, key.CapsLock, key.NumLock)
	return mods
}

func gridPointerSelectionMode(mods key.Modifiers) (virtual.GridSelectMode, bool) {
	mods = gridSemanticModifiers(mods)
	var shift, command key.Modifiers
	shift.SetFlag(true, key.Shift)
	command.SetFlag(true, key.Meta)
	switch mods {
	case 0:
		return virtual.GridSelectReplace, true
	case shift:
		return virtual.GridSelectExtend, true
	case command:
		return virtual.GridSelectToggle, true
	default:
		return virtual.GridSelectReplace, false
	}
}

// Row selection uses exact primitive chords rather than configurable keymap
// aliases that might reinterpret Control/Option as selection modifiers.
func (g *gridBinding) selectionKey(e events.Event) {
	if e.IsHandled() {
		return
	}
	mods := gridSemanticModifiers(e.Modifiers())
	var shift, command key.Modifiers
	shift.SetFlag(true, key.Shift)
	command.SetFlag(true, key.Meta)
	if e.KeyCode() == key.CodeA && mods == command {
		g.post(ui.GridSelection, g.model.SelectAll)
		e.SetHandled()
		return
	}
	if e.KeyCode() == key.CodeEscape && mods == 0 {
		// The existing scene-level modal handler owns Escape. An undismissable
		// modal also must not reinterpret it as a row-selection command.
		if scene := g.host.Scene; scene != nil && scene.Stage != nil && scene.Stage.Type == core.DialogStage {
			return
		}
		g.post(ui.GridSelection, g.model.ClearSelection)
		e.SetHandled()
		return
	}
	mode := virtual.GridSelectReplace
	if mods == shift {
		mode = virtual.GridSelectExtend
	} else if mods != 0 {
		return
	}
	rows, cols, page := 0, 0, 0
	switch e.KeyCode() {
	case key.CodeUpArrow:
		rows = -1
	case key.CodeDownArrow:
		rows = 1
	case key.CodeLeftArrow:
		cols = -1
	case key.CodeRightArrow:
		cols = 1
	case key.CodePageUp:
		page = -1
	case key.CodePageDown:
		page = 1
	default:
		return
	}
	g.post(ui.GridSelection, func() error {
		if page != 0 {
			return g.model.PageSelection(page, mode)
		}
		return g.model.MoveSelection(rows, cols, mode)
	})
	e.SetHandled()
}
func gridMaxX(s virtual.GridState) int {
	pane := max(0, s.Viewport.Width-s.FrozenWidth)
	if pane == 0 {
		return 0
	}
	return max(0, s.TotalWidth-s.FrozenWidth-pane)
}
func gridMaxY(s virtual.GridState) int {
	body := max(0, s.Viewport.Height-s.Config.HeaderHeight)
	if body == 0 {
		return 0
	}
	return max(0, s.TotalHeight-body)
}
func boundedWheel(v float64) int {
	const cap = 1 << 20
	if math.IsNaN(v) {
		return 0
	}
	if v > cap {
		return cap
	}
	if v < -cap {
		return -cap
	}
	return int(math.Trunc(v))
}
func normalizedGridOffset(value float32, maximum int) int {
	if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
		return -1
	}
	if value <= 0 {
		return 0
	}
	if value >= 1 {
		return maximum
	}
	return int(math.Round(float64(value) * float64(maximum)))
}
func (g *gridBinding) live() bool {
	if g == nil || g.closed.Load() {
		return false
	}
	select {
	case <-g.lifetimeDone:
		return false
	default:
		return g.owner.Context().Err() == nil
	}
}
func (g *gridBinding) eligible() bool {
	if !g.live() || g.r.control.retired || g.r.control.disabled || !g.scope || g.r.node.Grid != g.model || g.ancestorRefused {
		return false
	}
	for _, owner := range g.tabOwners {
		// A native tab selection may precede its controlled Change. Never apply
		// a retained action to the page that the native Tabs already retired.
		// Do not acquire another scene RC or inspect its geometry under Do.
		current, _ := owner.tabs.CurrentTab()
		if owner.page == nil || current != owner.page {
			return false
		}
	}
	return true
}
func (g *gridBinding) captureTabOwners() {
	clear(g.tabOwners)
	g.tabOwners = g.tabOwners[:0]
	g.ancestorRefused = false
	depth := 0
	for p := tree.Node(g.host); p != nil; p = p.AsTree().Parent {
		if depth >= 1024 {
			g.ancestorRefused = true
			break
		}
		depth++
		if tabs, ok := p.(*core.Tabs); ok {
			page, _ := tabs.CurrentTab()
			if tree.IsNil(page) || g.host.AsTree().ParentLevel(page) < 0 {
				page = nil
			}
			g.tabOwners = append(g.tabOwners, gridTabOwner{tabs: tabs, page: page})
		}
	}
}
func (g *gridBinding) fail(err error) {
	if err != nil && g.r.control.onError != nil {
		g.r.control.onError(err)
	}
}
func (g *gridBinding) applyIntentObserved(reason ui.GridChangeReason, interaction bool, origin gridInputOrigin, fn func() error) {
	if !g.live() || g.r.control.retired || g.r.node.Grid != g.model || interaction && !g.eligible() {
		g.observeTransition(origin, reason, virtual.GridState{}, virtual.GridState{}, "refused")
		return
	}
	before, e := g.model.State()
	if e != nil {
		g.observeTransition(origin, reason, before, virtual.GridState{}, "before-state-error")
		g.fail(e)
		return
	}
	if e = fn(); e != nil {
		g.observeTransition(origin, reason, before, virtual.GridState{}, "operation-error")
		g.fail(e)
		return
	}
	after, e := g.model.State()
	if e != nil {
		g.observeTransition(origin, reason, before, after, "after-state-error")
		g.fail(e)
		return
	}
	if before.Generation == after.Generation {
		g.observeTransition(origin, reason, before, after, "unchanged")
		return
	}
	g.observeTransition(origin, reason, before, after, "changed")
	if callback := g.r.node.GridChange; callback != nil {
		callback(ui.GridEvent{ID: g.r.node.ID, Reason: reason, State: after})
	}
	_ = g.owner.Invalidate(g.window)
}
func (g *gridBinding) post(reason ui.GridChangeReason, fn func() error) {
	if g.inputs != nil {
		g.postObserved(reason, gridInputOrigin{name: "input"}, fn)
		return
	}
	if g.live() {
		trace, tag := tracedGridIntent(g)
		if tag == 0 {
			postControl(g.r, func() { g.applyIntentObserved(reason, true, gridInputOrigin{}, fn) })
			return
		}
		trace.ExpectOwner(tag)
		err := g.owner.PostWindow(g.windowRef, func() {
			returned := false
			defer func() {
				if !returned {
					trace.Fail(tag, "traced owner transition did not return")
					trace.Transition(tag, 0)
				}
			}()
			before, beforeErr := g.model.State()
			g.applyIntentObserved(reason, true, gridInputOrigin{}, fn)
			after, afterErr := g.model.State()
			if beforeErr != nil || afterErr != nil || !g.live() || after.Generation == before.Generation {
				trace.Fail(tag, "traced grid input did not publish an owner transition")
				trace.Transition(tag, 0)
			} else {
				trace.Transition(tag, after.Generation)
			}
			returned = true
		})
		if err != nil {
			trace.Fail(tag, "traced grid owner posting refused")
			trace.Transition(tag, 0)
			trace.OwnerWork(tag, 0) // no owner callback was accepted
		}
	}
}
func (g *gridBinding) postObserved(reason ui.GridChangeReason, origin gridInputOrigin, fn func() error) {
	if g.inputs == nil {
		g.post(reason, fn)
		return
	}
	if g.live() {
		trace, tag := tracedGridIntent(g)
		origin.tag = tag
		if tag == 0 {
			g.postInputControl(origin, reason, func() { g.applyIntentObserved(reason, true, origin, fn) })
			return
		}
		trace.ExpectOwner(tag)
		err := g.owner.PostWindow(g.windowRef, func() {
			returned := false
			defer func() {
				if !returned {
					trace.Fail(tag, "traced owner transition did not return")
					trace.Transition(tag, 0)
				}
			}()
			before, beforeErr := g.model.State()
			g.applyIntentObserved(reason, true, origin, fn)
			after, afterErr := g.model.State()
			if beforeErr != nil || afterErr != nil || !g.live() || after.Generation == before.Generation {
				trace.Fail(tag, "traced grid input did not publish an owner transition")
				trace.Transition(tag, 0)
			} else {
				trace.Transition(tag, after.Generation)
			}
			returned = true
		})
		if err != nil {
			g.observeTransition(origin, reason, virtual.GridState{}, virtual.GridState{}, "post-refused")
			trace.Fail(tag, "traced grid owner posting refused")
			trace.Transition(tag, 0)
			trace.OwnerWork(tag, 0) // no owner callback was accepted
		}
	}
}
func (g *gridBinding) requestViewport(request gridRequest) {
	if !g.live() || request.width < 1 || request.height < 1 || request.scale <= 0 || math.IsNaN(float64(request.scale)) || math.IsInf(float64(request.scale), 0) {
		return
	}
	g.requestMu.Lock()
	if g.requested == request {
		g.requestMu.Unlock()
		return
	}
	g.requested = request
	g.requestMu.Unlock()
	// Layout owns geometry even when the control is disabled or outside modal
	// input scope. No widget/source reads occur in this native admission path.
	err := g.owner.PostWindow(g.windowRef, func() {
		g.applyIntentObserved(ui.GridViewport, false, gridInputOrigin{name: "layout"}, func() error {
			s, e := g.model.State()
			if e != nil {
				return e
			}
			v := s.Viewport
			v.Width, v.Height = request.width, request.height
			return g.model.SetViewport(v)
		})
	})
	if err != nil {
		g.observeTransition(gridInputOrigin{name: "layout"}, ui.GridViewport, virtual.GridState{}, virtual.GridState{}, "post-refused")
		// Queue pressure is transient; allow the next actual frame to retry.
		// A model capacity refusal occurs on the owner and remains remembered,
		// without an invalidate loop until a different measured size is requested.
		g.requestMu.Lock()
		if g.requested == request {
			g.requested = gridRequest{}
		}
		g.requestMu.Unlock()
		if !errors.Is(err, app.ErrClosed) && !errors.Is(err, app.ErrWindowClosed) {
			g.r.control.refused.Store(&err)
			_ = g.owner.Invalidate(g.window)
		}
	}
}

// Total domain extents may be enormous; only each materialized local rectangle
// reaches Core float32 layout. Bound those before growing or moving any pool.
func validateGridProjection(s virtual.GridSnapshot) error {
	if s.State.Viewport.Width < 1 || s.State.Viewport.Height < 1 || s.State.Viewport.Width > gridNativeMaxDimension || s.State.Viewport.Height > gridNativeMaxDimension {
		return ErrGridBudget
	}
	check := func(r virtual.GridRect) bool {
		if r.Width < 0 || r.Height < 0 || r.Width > gridNativeMaxDimension || r.Height > gridNativeMaxDimension || r.X < -gridNativeMaxCoordinate || r.X > gridNativeMaxCoordinate || r.Y < -gridNativeMaxCoordinate || r.Y > gridNativeMaxCoordinate {
			return false
		}
		return r.X+r.Width <= gridNativeMaxCoordinate && r.Y+r.Height <= gridNativeMaxCoordinate
	}
	for _, c := range s.Cells {
		if !check(c.Rect) || !check(c.Clip) {
			return ErrGridBudget
		}
	}
	for _, h := range s.Headers {
		if !check(h.Rect) || !check(h.Clip) {
			return ErrGridBudget
		}
	}
	return nil
}
func validGridScale(scale float32) bool {
	return scale > 0 && scale <= gridNativeMaxScale && !math.IsNaN(float64(scale)) && !math.IsInf(float64(scale), 0)
}
func (g *gridBinding) refuseGeometry() {
	g.withdrawMeasurement()
	if !g.geometryRefused {
		g.geometryRefused = true
		postControl(g.r, func() { g.fail(ErrGridBudget) })
	}
}
func syncGrid(r *record) error {
	g := r.control.grid
	if g == nil || !g.live() {
		return ErrGridClosed
	}
	state, err := nativeGridState(r.node)
	if err != nil {
		return err
	}
	if g.cellSurface != nil && g.projectionReady && g.projection.State.Generation == state.Generation {
		reference, _ := g.host.Scene.Property(gridCellSurfaceProperty).(bool)
		if reference || g.cellSurface.failed || !g.host.Scene.PassiveTextSurfaceAllowed() || !g.layout.valid || g.layoutStamp() != g.layout ||
			g.projectedDisabled != r.control.disabled || g.projectedHelp != r.node.Accessibility.Help || state.Edit.Active {
			g.projectionReady = false // ordinary fallback before rebuilding any native child
		}
	}
	if g.projectionReady && g.projection.State.Generation == state.Generation {
		// Domain data are immutable within their stamp. Owner invalidation after
		// capture only republishes geometry; it need not rematerialize or restyle
		// the same cells. Descriptor/theme/topology changes retain their normal
		// reconciliation updates, and private editing still has an independent
		// selection, restore, preedit and focus lifecycle.
		g.syncScrollbars(state)
		if err := syncGridEditing(g); err != nil {
			g.projectionReady = false
			return err
		}
		if g.projectedDisabled != r.control.disabled || g.projectedHelp != r.node.Accessibility.Help {
			g.layoutPending.Store(true)
			g.layout.valid = false
			g.host.Update()
			g.projectedDisabled = r.control.disabled
			g.projectedHelp = r.node.Accessibility.Help
		}
		return nil
	}
	// A failed partial sync cannot be reused on a later explicit retry.
	wasReady := g.projectionReady
	g.projectionReady = false
	snap, err := g.model.Snapshot()
	if err != nil {
		return err
	}
	if err := validateGridProjection(snap); err != nil {
		return err
	}
	if g.projection.State.Generation != snap.State.Generation {
		gridRegistry.Lock()
		if gridRegistry.next == math.MaxUint64 {
			gridRegistry.Unlock()
			return ErrGridBinding
		}
		gridRegistry.next++
		g.epoch = gridRegistry.next
		gridRegistry.Unlock()
	}
	g.projectionReady = wasReady
	local := g.canLayoutLocally(snap)
	if g.wantsCellSurface(snap) {
		return g.syncCellSurface(snap, local)
	}
	if g.cellSurface != nil {
		g.retireCellSurface()
		local = false
	}
	g.projectionReady = false
	cells, retiredCells, err := g.retainCells(snap)
	if err != nil {
		return err
	}
	headers, retiredHeaders, err := g.retainHeaders(snap)
	if err != nil {
		return err
	}
	g.layoutPending.Store(true)
	g.projection = snap
	g.cells, g.headers = cells, headers
	for _, list := range [][]*core.Text{retiredCells, retiredHeaders} {
		for _, w := range list {
			delete(g.layoutDirty, w)
			w.Delete()
		}
	}
	styleDirty := make(map[*core.Text]bool)
	for i, w := range g.cells {
		created := w == nil
		if created {
			w = core.NewText(g.panes[2])
			g.installTextStyle(w)
			g.cells[i] = w
		}
		c := snap.Cells[i]
		pane := 2
		if !c.Frozen {
			pane = 3
		}
		selected := c.Active
		if snap.State.Config.RowSelection {
			selected = c.Selected
		}
		sty, size, moved := g.placeRetained(w, g.panes[pane], c.Rect, c.Clip, c.Text, selected, c.Active && snap.State.Config.RowSelection)
		if moved && !created {
			local = false
		}
		if sty || created {
			styleDirty[w] = true
		}
		if size {
			if g.layoutDirty == nil {
				g.layoutDirty = make(map[*core.Text]bool)
			}
			g.layoutDirty[w] = true
		}
	}
	for i, w := range g.headers {
		created := w == nil
		if created {
			w = core.NewText(g.panes[0])
			g.installTextStyle(w)
			g.headers[i] = w
		}
		h := snap.Headers[i]
		pane := 0
		if !h.Frozen {
			pane = 1
		}
		sty, size, moved := g.placeRetained(w, g.panes[pane], h.Rect, h.Clip, h.Title, false, false)
		if moved && !created {
			local = false
		}
		if sty || created {
			styleDirty[w] = true
		}
		if size {
			if g.layoutDirty == nil {
				g.layoutDirty = make(map[*core.Text]bool)
			}
			g.layoutDirty[w] = true
		}
	}
	g.syncScrollbars(snap.State)
	if err := syncGridEditing(g); err != nil {
		return err
	}
	g.orderPanes()
	if local && g.host.Scene.NeedsLocalLayout(g.host) {
		for w := range styleDirty {
			if !g.host.Scene.UpdateTextForLocalLayout(g.host, w) {
				w.UpdateWidget()
			}
			w.Style()
		}
	} else {
		g.layout.valid = false
		g.host.Update()
	}
	g.projectedDisabled = r.control.disabled
	g.projectedHelp = r.node.Accessibility.Help
	g.projectionReady = true
	return nil
}
func (g *gridBinding) syncScrollbars(s virtual.GridState) {
	x, y := float32(0), float32(0)
	if maximum := gridMaxX(s); maximum > 0 {
		x = float32(float64(s.Viewport.ScrollX) / float64(maximum))
	}
	if maximum := gridMaxY(s); maximum > 0 {
		y = float32(float64(s.Viewport.ScrollY) / float64(maximum))
	}
	// Native dragging can publish a noncanonical value while its rounded model
	// pixel is unchanged or the intent is refused. Restore only that slider;
	// SetValue marks it for rendering even when the supplied value is identical.
	if g.horizontal.Value != x {
		g.horizontal.SetValue(x)
	}
	if g.vertical.Value != y {
		g.vertical.SetValue(y)
	}
}
func (g *gridBinding) installTextStyle(w *core.Text) {
	w.Styler(func(s *styles.Style) {
		s.Padding.Zero()
		s.Margin.Zero()
		s.Border.Width.Zero()
		// Membership uses the normal selected background; the active cell has
		// a separate inside border even when its row is toggled out of the set.
		cursor, _ := w.Property("vellum.grid.cursor").(bool)
		if cursor {
			s.Border.Style.Set(styles.BorderSolid)
			s.Border.Width.Set(units.Px(1))
			s.Border.Color.Set(colors.Scheme.Primary.Base)
		}
		g.applyScopedCellStyle(s, w.StateIs(states.Selected), cursor)
		s.Grow.SetZero()
		s.Text.WhiteSpace = text.WhiteSpacePre
		s.SetAbilities(false, abilities.Focusable, abilities.Selectable, abilities.Clickable, abilities.Activatable, abilities.DoubleClickable, abilities.TripleClickable, abilities.Draggable, abilities.LongPressable, abilities.LongHoverable, abilities.Slideable)
	})
}

// Mac-themed grids use their scoped palette, rather than the process-global
// light selection scheme. Preserve default rendering for unthemed consumers.
func (g *gridBinding) applyScopedCellStyle(s *styles.Style, selected, cursor bool) {
	if g == nil || g.r == nil || !g.r.theme.Palette.Set {
		return
	}
	p := g.r.theme.Palette.Value
	s.Padding.Set(units.Dp(6), units.Dp(6))
	if selected {
		s.Background = image.NewUniform(p.Accent)
		s.Color = image.NewUniform(p.AccentInk)
	}
	// Read-only row selection has one continuous highlight. An editable grid
	// retains the cell cursor needed to identify its editing destination.
	if cursor && selected && g.projection.State.Config.RowSelection && !g.projection.State.Config.CellEditing {
		s.Border.Width.Zero()
	} else if cursor {
		c := p.Accent
		if selected {
			c = p.AccentInk
		}
		s.Border.Color.Set(image.NewUniform(c))
	}
}

func (g *gridBinding) place(w *core.Text, p *gridPane, rect, clip virtual.GridRect, value string, selected, cursor bool) {
	if w.Parent != p {
		tree.MoveToParent(w, p)
		w.SetScene(p.Scene)
	}
	w.SetProperty("vellum.grid.rect", rect)
	w.SetProperty("vellum.grid.cursor", cursor)
	w.SetText(value)
	w.SetState(clip.Empty(), states.Invisible)
	w.SetState(selected, states.Selected)
	// The retained sync chooses a local update or the full host update after
	// all placements are complete.
}
func disposeGrid(r *record) {
	if r == nil || r.control == nil || r.control.grid == nil {
		return
	}
	g := r.control.grid
	g.retireCellSurface()
	g.closed.Store(true)
	g.disposeEditing()
	g.armed = nil
	g.measurement.Store(nil)
	g.measurementScene.Store(nil)
	clear(g.layoutDirty)
	g.layout = gridLayoutStamp{}
	gridRegistry.Lock()
	if gridRegistry.models[g.model] == g {
		if old := g.previous; old != nil && !old.closed.Load() && !old.r.control.retired {
			gridRegistry.models[g.model] = old
		} else {
			delete(gridRegistry.models, g.model)
		}
	}
	gridRegistry.Unlock()
}

// A preparation may temporarily allocate a replacement, while the old view
// remains the only active binding. Failure restores the predecessor registration.
func prepareGridReplacements(previous *record, next ui.Node) func() {
	models := map[*virtual.Grid]string{}
	var nodes func(ui.Node)
	nodes = func(n ui.Node) {
		if n.Kind == ui.DataGrid {
			models[n.Grid] = n.ID
		}
		for _, c := range n.Children {
			nodes(c)
		}
	}
	nodes(next)
	allowed := map[*virtual.Grid]*gridBinding{}
	var records func(*record)
	records = func(r *record) {
		if r == nil {
			return
		}
		if r.control != nil && r.control.grid != nil {
			g := r.control.grid
			if id, ok := models[g.model]; ok && id != r.node.ID {
				allowed[g.model] = g
			}
		}
		for _, c := range r.children {
			records(c)
		}
	}
	records(previous)
	gridRegistry.Lock()
	for model, g := range allowed {
		if gridRegistry.models[model] == g {
			gridRegistry.preparing[model] = g
		}
	}
	gridRegistry.Unlock()
	return func() {
		gridRegistry.Lock()
		defer gridRegistry.Unlock()
		for model, g := range allowed {
			if gridRegistry.preparing[model] == g {
				delete(gridRegistry.preparing, model)
			}
		}
	}
}
func bindGridScopes(root *record) {
	scope := ui.FocusScope(accessibilityFocusProjection(root, ""))
	allowed := map[*virtual.Grid]bool{}
	var collect func(ui.Node, bool)
	collect = func(n ui.Node, disabled bool) {
		disabled = disabled || n.Disabled
		if disabled || n.Kind == ui.Dialog && !n.Open {
			return
		}
		if n.Grid != nil {
			allowed[n.Grid] = true
		}
		for _, c := range n.Children {
			if n.Kind == ui.Disclosure && !n.Checked || n.Kind == ui.Tabs && c.ID != n.SelectedID {
				continue
			}
			collect(c, disabled)
		}
	}
	collect(scope, false)
	var walk func(*record)
	walk = func(r *record) {
		if r == nil {
			return
		}
		if r.control != nil && r.control.grid != nil {
			r.control.grid.scope = allowed[r.node.Grid]
			r.control.grid.captureTabOwners()
		}
		for _, c := range r.children {
			walk(c)
		}
		for _, c := range activeWorkbenchRecords(r) {
			walk(c)
		}
	}
	walk(root)
}

// gridWidget uses actual Core custom layout; only the single host owns focus.
type gridWidget struct {
	core.Frame
	binding *gridBinding
}

func (w *gridWidget) Init() {
	w.Frame.Init()
	w.Styler(func(s *styles.Style) {
		s.Display = styles.Custom
		s.Overflow.Set(styles.OverflowHidden)
		s.Padding.Zero()
		s.Gap.Zero()
		s.Grow.Set(1, 1)
		s.SetAbilities(true, abilities.Clickable)
		s.SetAbilities(w.binding == nil || w.binding.editing == nil || !w.binding.editing.visible || w.binding.editing.record.node.Disabled, abilities.Focusable)
	})
}
func (w *gridWidget) Position() {
	g := w.binding
	if g == nil {
		w.Frame.Position()
		return
	}
	v := g.projection.State.Viewport
	frozen := min(v.Width, g.projection.State.FrozenWidth)
	header := min(v.Height, g.projection.State.Config.HeaderHeight)
	panes := [4]virtual.GridRect{{Width: frozen, Height: header}, {X: frozen, Width: v.Width - frozen, Height: header}, {Y: header, Width: frozen, Height: v.Height - header}, {X: frozen, Y: header, Width: v.Width - frozen, Height: v.Height - header}}
	scale := w.Styles.UnitContext.Px(1)
	if !validGridScale(scale) {
		g.refuseGeometry()
		return
	}
	for i, p := range g.panes {
		p.rect = panes[i]
		setGridGeometry(&p.WidgetBase, panes[i], scale)
	}
	setGridGeometry(&g.horizontal.WidgetBase, virtual.GridRect{Y: v.Height, Width: v.Width, Height: gridScrollbar}, scale)
	setGridGeometry(&g.vertical.WidgetBase, virtual.GridRect{X: v.Width, Width: gridScrollbar, Height: v.Height}, scale)
	w.Frame.Position()
	clear(g.layoutDirty)
}
func (w *gridWidget) RenderWidget() {
	w.Frame.RenderWidget()
	if w.binding != nil {
		w.binding.capture()
	}
}

type gridPane struct {
	core.Frame
	binding *gridBinding
	part    uint8
	rect    virtual.GridRect
}

func (p *gridPane) Init() {
	p.Frame.Init()
	p.Styler(func(s *styles.Style) {
		s.Display = styles.Custom
		s.Overflow.Set(styles.OverflowHidden)
		s.Padding.Zero()
		s.Gap.Zero()
		s.Grow.SetZero()
	})
}
func (p *gridPane) Position() {
	scale := p.Styles.UnitContext.Px(1)
	p.ForWidgetChildren(func(_ int, w core.Widget, wb *core.WidgetBase) bool {
		r, _ := w.AsTree().Property("vellum.grid.rect").(virtual.GridRect)
		setGridGeometryOrigin(wb, r, p.rect.X, p.rect.Y, scale)
		// A cell's indexed natural rectangle is its exterior, not its text
		// content. Keep the active border inside those shared rounded edges.
		// Panes and scrollbars retain their existing content geometry contract.
		size := &wb.Geom.Size
		size.Actual.Total = size.Actual.Content
		size.Alloc.Total = size.Alloc.Content
		size.SetContentFromTotal(&size.Actual)
		size.SetContentFromTotal(&size.Alloc)
		size.Actual.Content.Set(max(0, size.Actual.Content.X), max(0, size.Actual.Content.Y))
		size.Alloc.Content.Set(max(0, size.Alloc.Content.X), max(0, size.Alloc.Content.Y))
		return tree.Continue
	})
	p.Frame.Position()
}
func setGridGeometry(w *core.WidgetBase, r virtual.GridRect, scale float32) {
	setGridGeometryOrigin(w, r, 0, 0, scale)
}
func setGridGeometryOrigin(w *core.WidgetBase, r virtual.GridRect, originX, originY int, scale float32) {
	// Round global natural edges exactly once; pane-local positions subtract the
	// rounded origin. Independently rounding (edge-origin) changes fractional-DPI
	// pixels at frozen/header boundaries.
	edge := func(v int) float32 { return float32(math.Round(float64(v) * float64(scale))) }
	x0, y0 := edge(r.X)-edge(originX), edge(r.Y)-edge(originY)
	x1, y1 := edge(r.X+r.Width)-edge(originX), edge(r.Y+r.Height)-edge(originY)
	w.Styles.Pos.Set(units.Dot(x0), units.Dot(y0))
	w.Geom.RelPos.Set(x0, y0)
	w.Geom.Size.Actual.Content.Set(x1-x0, y1-y0)
	w.Geom.Size.Alloc.Content.Set(x1-x0, y1-y0)
	w.Geom.Size.SetTotalFromContent(&w.Geom.Size.Actual)
	w.Geom.Size.SetTotalFromContent(&w.Geom.Size.Alloc)
}

type gridPress struct {
	ref                virtual.GridCellRef
	bounds, hostBounds image.Rectangle
	incarnation        uint64
	scale              float32
	rowSelection       bool
	mode               virtual.GridSelectMode
}

func (g *gridBinding) refAt(point image.Point) (gridPress, bool) {
	measured := g.currentMeasurement()
	if measured == nil || !g.live() {
		return gridPress{}, false
	}
	if measured.Generation != g.projection.State.Generation {
		return gridPress{}, false
	}
	for i, c := range measured.Cells {
		if point.In(c.Bounds) && i < len(g.projection.Cells) {
			return gridPress{ref: g.projection.Cells[i].Ref, bounds: c.Bounds, hostBounds: measured.Bounds, incarnation: measured.Incarnation, scale: measured.Scale}, true
		}
	}
	return gridPress{}, false
}
