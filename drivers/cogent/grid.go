package cogent

import (
	"errors"
	"html"
	"image"
	"math"
	"sync"
	"sync/atomic"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
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
	lifetimeDone         <-chan struct{}
	model                *virtual.Grid
	incarnation          uint64
	epoch                uint64
	frames               uint64
	closed               atomic.Bool
	host                 *gridWidget
	panes                [4]*gridPane
	cells                []*core.Text
	headers              []*core.Text
	horizontal, vertical *core.Slider
	projection           virtual.GridSnapshot // originating native scene / owner refresh only
	measurement          atomic.Pointer[GridInspection]
	armed                *gridPress // originating scene only; detached opaque identity and scene geometry
	requestMu            sync.Mutex
	requested            gridRequest
	scope                bool           // owner-only active modal visibility
	tabOwners            []gridTabOwner // owner-only retained topology, no native geometry reads
	ancestorRefused      bool
	wheelX, wheelY       float64
	geometryRefused      bool
}

func validateNativeGrid(n ui.Node) error {
	if n.Kind != ui.DataGrid {
		return nil
	}
	if n.Grid == nil {
		return ErrGridBinding
	}
	s, err := n.Grid.State()
	if err != nil {
		return err
	}
	c := s.Config
	if s.Viewport.Width > gridNativeMaxDimension || s.Viewport.Height > gridNativeMaxDimension || !s.VariableHeights && c.RowHeight > gridNativeMaxDimension || c.HeaderHeight > gridNativeMaxDimension || c.MaxRows > gridNativeMaxRows || c.MaxColumns > gridNativeMaxColumns || c.MaxCells > gridNativeMaxCells || c.MaxTextBytes > gridNativeMaxText {
		return ErrGridBudget
	}
	return nil
}
func mountGrid(parent tree.Node, r *record, a *app.Application, window string) (core.Widget, error) {
	if err := validateNativeGrid(r.node); err != nil {
		return nil, err
	}
	life, err := a.WindowContext(window)
	if err != nil {
		return nil, err
	}
	gridRegistry.Lock()
	existing := gridRegistry.models[r.node.Grid]
	if existing != nil && (gridRegistry.preparing[r.node.Grid] != existing || existing.owner != a || existing.window != window) || gridRegistry.next == math.MaxUint64 {
		gridRegistry.Unlock()
		return nil, ErrGridBinding
	}
	gridRegistry.next++
	g := &gridBinding{r: r, owner: a, window: window, model: r.node.Grid, incarnation: gridRegistry.next, previous: existing, lifetimeDone: life.Done(), scope: true}
	gridRegistry.models[g.model] = g
	gridRegistry.Unlock()
	r.control.grid = g
	host := tree.New[gridWidget](parent)
	host.binding = g
	g.host = host
	for i := range g.panes {
		p := tree.New[gridPane](host)
		p.binding = g
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
		g.post(ui.GridViewport, func() error {
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
		g.post(ui.GridViewport, func() error {
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
		if !ok || at != *ref {
			return
		}
		captured := ref.ref
		g.post(ui.GridSelection, func() error { return g.model.Activate(captured) })
		e.SetHandled()
	})
	host.On(events.KeyChord, func(e events.Event) {
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
		g.wheelX += float64(s.Delta.X) / float64(scale)
		g.wheelY += float64(s.Delta.Y) / float64(scale)
		dx, dy := boundedWheel(g.wheelX), boundedWheel(g.wheelY)
		g.wheelX -= math.Trunc(g.wheelX)
		g.wheelY -= math.Trunc(g.wheelY)
		g.post(ui.GridViewport, func() error { return g.model.Scroll(dx, dy) })
		e.SetHandled()
	})
	return host, nil
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
func (g *gridBinding) applyIntent(reason ui.GridChangeReason, interaction bool, fn func() error) {
	if !g.live() || g.r.control.retired || g.r.node.Grid != g.model || interaction && !g.eligible() {
		return
	}
	before, e := g.model.State()
	if e != nil {
		g.fail(e)
		return
	}
	if e = fn(); e != nil {
		g.fail(e)
		return
	}
	after, e := g.model.State()
	if e != nil {
		g.fail(e)
		return
	}
	if before.Generation == after.Generation {
		return
	}
	if callback := g.r.node.GridChange; callback != nil {
		callback(ui.GridEvent{ID: g.r.node.ID, Reason: reason, State: after})
	}
	_ = g.owner.Invalidate(g.window)
}
func (g *gridBinding) post(reason ui.GridChangeReason, fn func() error) {
	if g.live() {
		postControl(g.r, func() { g.applyIntent(reason, true, fn) })
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
	err := g.owner.PostWindow(g.window, func() {
		g.applyIntent(ui.GridViewport, false, func() error {
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
	g.measurement.Store(nil)
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
	if err := validateNativeGrid(r.node); err != nil {
		return err
	}
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
	g.projection = snap
	// Retired slots must release shaped text as well as strings. Destroy surplus
	// widgets rather than retaining private Core text caches for old projections.
	if len(g.cells) > len(snap.Cells) {
		for _, w := range g.cells[len(snap.Cells):] {
			w.Delete()
		}
		clear(g.cells[len(snap.Cells):])
		g.cells = g.cells[:len(snap.Cells)]
	}
	if len(g.headers) > len(snap.Headers) {
		for _, w := range g.headers[len(snap.Headers):] {
			w.Delete()
		}
		clear(g.headers[len(snap.Headers):])
		g.headers = g.headers[:len(snap.Headers)]
	}
	for len(g.cells) < len(snap.Cells) {
		w := core.NewText(g.panes[2])
		g.installTextStyle(w)
		g.cells = append(g.cells, w)
	}
	for len(g.headers) < len(snap.Headers) {
		w := core.NewText(g.panes[0])
		g.installTextStyle(w)
		g.headers = append(g.headers, w)
	}
	for i, w := range g.cells {
		if i >= len(snap.Cells) {
			w.SetState(true, states.Invisible)
			continue
		}
		c := snap.Cells[i]
		pane := 2
		if !c.Frozen {
			pane = 3
		}
		g.place(w, g.panes[pane], c.Rect, c.Clip, html.EscapeString(c.Text), c.Active)
	}
	for i, w := range g.headers {
		if i >= len(snap.Headers) {
			w.SetState(true, states.Invisible)
			continue
		}
		h := snap.Headers[i]
		pane := 0
		if !h.Frozen {
			pane = 1
		}
		g.place(w, g.panes[pane], h.Rect, h.Clip, html.EscapeString(h.Title), false)
	}
	sx := gridMaxX(snap.State)
	sy := gridMaxY(snap.State)
	g.horizontal.SetValue(0)
	if sx > 0 {
		g.horizontal.SetValue(float32(float64(snap.State.Viewport.ScrollX) / float64(sx)))
	}
	g.vertical.SetValue(0)
	if sy > 0 {
		g.vertical.SetValue(float32(float64(snap.State.Viewport.ScrollY) / float64(sy)))
	}
	g.host.Update()
	return nil
}
func (g *gridBinding) installTextStyle(w *core.Text) {
	w.Styler(func(s *styles.Style) {
		s.Padding.Zero()
		s.Margin.Zero()
		s.Border.Width.Zero()
		s.Grow.SetZero()
		s.Text.WhiteSpace = text.WhiteSpacePre
		s.SetAbilities(false, abilities.Focusable, abilities.Selectable, abilities.Clickable, abilities.Activatable, abilities.DoubleClickable, abilities.TripleClickable, abilities.Draggable, abilities.LongPressable, abilities.LongHoverable, abilities.Slideable)
	})
}
func (g *gridBinding) place(w *core.Text, p *gridPane, rect, clip virtual.GridRect, value string, active bool) {
	if w.Parent != p {
		tree.MoveToParent(w, p)
		w.SetScene(p.Scene)
	}
	w.SetProperty("vellum.grid.rect", rect)
	w.SetText(value)
	w.SetState(clip.Empty(), states.Invisible)
	w.SetState(active, states.Selected)
	w.Update()
}
func disposeGrid(r *record) {
	if r == nil || r.control == nil || r.control.grid == nil {
		return
	}
	g := r.control.grid
	g.closed.Store(true)
	g.armed = nil
	g.measurement.Store(nil)
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
		if n.Kind == ui.DataGrid {
			allowed[n.Grid] = true
		}
		for _, c := range n.Children {
			if n.Kind == ui.Tabs && c.ID != n.SelectedID {
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
		s.SetAbilities(true, abilities.Focusable, abilities.Clickable)
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
