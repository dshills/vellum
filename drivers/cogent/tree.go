package cogent

import (
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"math"
	"slices"
	"sync"
	"sync/atomic"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
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
	ErrTreeBinding = errors.New("cogent: outline already bound or wrong owner")
	ErrTreeLayout  = errors.New("cogent: outline requires a current completed frame")
	ErrTreeClosed  = errors.New("cogent: outline binding/window closed")
	ErrTreeBudget  = errors.New("cogent: outline native pool/geometry budget exceeded")
)

const treeNativeMaxRows = 500
const treeNativeMaxText = 8 << 20
const treeNativeMaxLabel = 64 << 10
const treeNativeMaxKey = 4096
const treeFooter = 28
const treeDisclosureWidth = 16
const treeIndent = 14
const treeMaxIndentDepth = 24

var outlineRegistry = struct {
	sync.Mutex
	models    map[*virtual.Outline]*outlineBinding
	preparing map[*virtual.Outline]*outlineBinding
	next      uint64
}{models: map[*virtual.Outline]*outlineBinding{}, preparing: map[*virtual.Outline]*outlineBinding{}}

type outlineRequest struct {
	width, height int
	scale         float32
}
type outlineBinding struct {
	r                          *record
	previous                   *outlineBinding
	owner                      *app.Application
	window                     string
	model                      *virtual.Outline
	lifetimeDone               <-chan struct{}
	incarnation, epoch, frames uint64
	closed                     atomic.Bool
	projectionValid            atomic.Bool
	host                       *outlineWidget
	pane                       *outlinePane
	rows                       []*outlineNativeRow
	pager                      [2]*core.Button
	status                     *core.Text
	vertical                   *core.Slider
	projection                 virtual.OutlineSnapshot
	projectedDisabled          bool
	projectedHelp              string
	measurement                atomic.Pointer[TreeInspection]
	armed                      *outlinePress
	requestMu                  sync.Mutex
	requested                  outlineRequest
	scope                      bool
	tabOwners                  []gridTabOwner
	ancestorRefused            bool
	geometryRefused            bool
	wheel                      float64
}

func validateNativeTree(n ui.Node) error {
	if n.Kind != ui.DataTree {
		return nil
	}
	if len(n.ID) > treeNativeMaxKey {
		return ErrTreeBudget
	}
	if n.Tree == nil {
		return ErrTreeBinding
	}
	s, err := n.Tree.State()
	if err != nil {
		return err
	}
	if s.Config.MaxRows > treeNativeMaxRows || s.Config.RowHeight > gridNativeMaxDimension || s.Viewport.Width > gridNativeMaxDimension || s.Viewport.Height > gridNativeMaxDimension {
		return ErrTreeBudget
	}
	return nil
}
func mountTree(parent tree.Node, r *record, a *app.Application, window string) (core.Widget, error) {
	if err := validateNativeTree(r.node); err != nil {
		return nil, err
	}
	life, err := a.WindowContext(window)
	if err != nil {
		return nil, err
	}
	outlineRegistry.Lock()
	old := outlineRegistry.models[r.node.Tree]
	if old != nil && (outlineRegistry.preparing[r.node.Tree] != old || old.owner != a || old.window != window) || outlineRegistry.next == math.MaxUint64 {
		outlineRegistry.Unlock()
		return nil, ErrTreeBinding
	}
	outlineRegistry.next++
	t := &outlineBinding{r: r, owner: a, window: window, model: r.node.Tree, lifetimeDone: life.Done(), incarnation: outlineRegistry.next, previous: old, scope: true}
	outlineRegistry.models[t.model] = t
	outlineRegistry.Unlock()
	r.control.outline = t
	host := tree.New[outlineWidget](parent)
	host.binding = t
	t.host = host
	t.pane = tree.New[outlinePane](host)
	t.pane.binding = t
	t.vertical = core.NewSlider(host).SetMin(0).SetMax(1).SetStep(.01)
	t.vertical.Styler(func(s *styles.Style) { s.Direction = styles.Column; s.SetAbilities(false, abilities.Focusable) })
	t.vertical.OnChange(func(events.Event) {
		value := t.vertical.Value
		t.post(ui.TreeViewport, func() error {
			s, e := t.model.State()
			if e != nil {
				return e
			}
			v := s.Viewport
			v.ScrollY = normalizedGridOffset(value, treeMaxY(s))
			if v.ScrollY < 0 {
				return ErrTreeBudget
			}
			return t.model.SetViewport(v)
		})
	})
	for i, label := range []string{"Previous page", "Next page"} {
		dir := i*2 - 1
		b := core.NewButton(host).SetText(label)
		t.pager[i] = b
		b.Styler(func(s *styles.Style) {
			s.SetAbilities(false, abilities.Focusable)
			s.Padding.Set(units.Dp(1))
			s.Min.Zero()
		})
		b.OnClick(func(events.Event) {
			host.SetFocus()
			t.post(ui.TreeViewport, func() error { return t.pageViewport(dir) })
		})
	}
	t.status = core.NewText(host)
	t.status.Styler(func(s *styles.Style) {
		s.Text.WhiteSpace = text.WhiteSpacePre
		s.Padding.Zero()
		s.Margin.Zero()
		s.SetAbilities(false, abilities.Focusable, abilities.Selectable)
	})
	host.On(events.MouseDown, func(e events.Event) {
		t.armed = nil
		if e.MouseButton() != events.Left || gridSemanticModifiers(e.Modifiers()) != 0 || host.StateIs(states.Disabled) {
			return
		}
		if ref, ok := t.refAt(e.Pos()); ok {
			t.armed = &ref
			host.SetFocus()
			e.SetHandled()
		}
	})
	host.On(events.MouseUp, func(e events.Event) {
		ref := t.armed
		t.armed = nil
		if ref == nil || e.MouseButton() != events.Left {
			return
		}
		at, ok := t.refAt(e.Pos())
		if !ok || at != *ref {
			return
		}
		captured := ref.ref
		disclosure := ref.disclosure
		postControl(r, func() {
			if !t.eligible() {
				return
			}
			row, err := t.model.Row(captured)
			if err != nil {
				return
			} // retired press is a refused interaction, not a mutation
			if disclosure {
				t.disclose(row, !row.Expanded)
			} else {
				t.apply(ui.TreeSelection, true, func() error { return t.model.SelectRef(captured) })
			}
		})
		e.SetHandled()
	})
	host.On(events.KeyChord, func(e events.Event) { t.key(e) })
	host.On(events.Scroll, func(e events.Event) {
		sc, ok := e.(*events.MouseScroll)
		if !ok || math.IsNaN(float64(sc.Delta.Y)) || math.IsInf(float64(sc.Delta.Y), 0) {
			return
		}
		scale := host.Styles.UnitContext.Px(1)
		if !validGridScale(scale) {
			return
		}
		t.wheel += float64(sc.Delta.Y) / float64(scale)
		dy := boundedWheel(t.wheel)
		t.wheel -= math.Trunc(t.wheel)
		t.post(ui.TreeViewport, func() error {
			s, e := t.model.State()
			if e != nil {
				return e
			}
			v := s.Viewport
			v.ScrollY = clampTreeScroll(s, dy)
			return t.model.SetViewport(v)
		})
		e.SetHandled()
	})
	return host, nil
}
func treeMaxY(s virtual.OutlineState) int { return max(0, s.TotalHeight-s.Viewport.Height) }
func clampTreeScroll(s virtual.OutlineState, delta int) int {
	m := treeMaxY(s)
	v := s.Viewport.ScrollY
	if delta > 0 {
		return v + min(delta, m-v)
	}
	return v - min(-delta, v)
}
func (t *outlineBinding) pageViewport(pages int) error {
	s, e := t.model.State()
	if e != nil {
		return e
	}
	v := s.Viewport
	v.ScrollY = clampTreeScroll(s, pages*max(1, v.Height))
	return t.model.SetViewport(v)
}
func (t *outlineBinding) live() bool {
	if t == nil || t.closed.Load() {
		return false
	}
	select {
	case <-t.lifetimeDone:
		return false
	default:
		return t.owner.Context().Err() == nil
	}
}
func (t *outlineBinding) eligible() bool {
	if !t.live() || !t.projectionValid.Load() || t.r.control.retired || t.r.control.disabled || !t.scope || t.r.node.Tree != t.model || t.ancestorRefused {
		return false
	}
	for _, o := range t.tabOwners {
		p, _ := o.tabs.CurrentTab()
		if o.page == nil || p != o.page {
			return false
		}
	}
	return true
}
func (t *outlineBinding) fail(err error) {
	if err != nil && t.r.control.onError != nil {
		t.r.control.onError(err)
	}
}
func (t *outlineBinding) apply(reason ui.TreeChangeReason, interactive bool, fn func() error) {
	if !t.live() || t.r.control.retired || t.r.node.Tree != t.model || interactive && !t.eligible() {
		return
	}
	before, err := t.model.State()
	if err != nil {
		t.fail(err)
		return
	}
	if err = fn(); err != nil {
		t.fail(err)
		return
	}
	after, err := t.model.State()
	if err != nil {
		t.fail(err)
		return
	}
	if before.Generation == after.Generation {
		return
	}
	if cb := t.r.node.TreeChange; cb != nil {
		cb(ui.TreeEvent{ID: t.r.node.ID, Reason: reason, State: after})
	}
	_ = t.owner.Invalidate(t.window)
}
func (t *outlineBinding) post(reason ui.TreeChangeReason, fn func() error) {
	if t.live() {
		postControl(t.r, func() { t.apply(reason, true, fn) })
	}
}
func (t *outlineBinding) disclose(row virtual.OutlineRow, expanded bool) {
	if !t.eligible() || !row.HasChildren || row.Expanded == expanded {
		return
	}
	current, err := t.model.Row(row.Ref)
	if err != nil {
		return
	}
	if current.ID != row.ID || current.Expanded != row.Expanded {
		return
	}
	cb := t.r.node.TreeDisclosure
	if cb == nil {
		return
	}
	state, err := t.model.State()
	if err != nil {
		t.fail(err)
		return
	}
	kind := virtual.OutlineExpand
	if !expanded {
		kind = virtual.OutlineCollapse
	}
	cb(ui.TreeDisclosureEvent{ID: t.r.node.ID, Change: virtual.OutlineChange{Kind: kind, ID: row.ID}, Row: current, State: state})
	_ = t.owner.Invalidate(t.window)
}
func treeNavigation(fn func() error) error {
	err := fn()
	if errors.Is(err, virtual.ErrTreeRange) || errors.Is(err, virtual.ErrTreeHidden) {
		return nil
	}
	return err
}
func (t *outlineBinding) key(e events.Event) {
	if e.IsHandled() || gridSemanticModifiers(e.Modifiers()) != 0 {
		return
	}
	switch e.KeyCode() {
	case key.CodeUpArrow:
		t.post(ui.TreeSelection, func() error { return treeNavigation(func() error { return t.model.Move(-1) }) })
	case key.CodeDownArrow:
		t.post(ui.TreeSelection, func() error { return treeNavigation(func() error { return t.model.Move(1) }) })
	case key.CodePageUp:
		t.post(ui.TreeSelection, func() error { return treeNavigation(func() error { return t.model.Page(-1) }) })
	case key.CodePageDown:
		t.post(ui.TreeSelection, func() error { return treeNavigation(func() error { return t.model.Page(1) }) })
	case key.CodeHome:
		t.post(ui.TreeSelection, func() error { return treeNavigation(t.model.Home) })
	case key.CodeEnd:
		t.post(ui.TreeSelection, func() error { return treeNavigation(t.model.End) })
	case key.CodeLeftArrow:
		postControl(t.r, func() {
			if !t.eligible() {
				return
			}
			row, err := t.model.Current()
			if err != nil {
				return
			}
			if row.HasChildren && row.Expanded {
				t.disclose(row, false)
			} else {
				t.apply(ui.TreeSelection, true, func() error { return treeNavigation(t.model.Parent) })
			}
		})
	case key.CodeRightArrow:
		postControl(t.r, func() {
			if !t.eligible() {
				return
			}
			row, err := t.model.Current()
			if err != nil {
				return
			}
			if !row.HasChildren {
				return
			}
			if !row.Expanded {
				t.disclose(row, true)
			} else {
				t.apply(ui.TreeSelection, true, func() error { return treeNavigation(t.model.FirstChild) })
			}
		})
	case key.CodeReturnEnter:
		postControl(t.r, func() {
			if !t.eligible() {
				return
			}
			row, err := t.model.Current()
			if err == nil {
				t.disclose(row, !row.Expanded)
			}
		})
	default:
		return
	}
	e.SetHandled()
}
func (t *outlineBinding) requestViewport(req outlineRequest) {
	if !t.live() || req.width < 1 || req.height < 1 || !validGridScale(req.scale) {
		return
	}
	t.requestMu.Lock()
	if t.requested == req {
		t.requestMu.Unlock()
		return
	}
	t.requested = req
	t.requestMu.Unlock()
	err := t.owner.PostWindow(t.window, func() {
		t.apply(ui.TreeViewport, false, func() error {
			s, e := t.model.State()
			if e != nil {
				return e
			}
			v := s.Viewport
			v.Width, v.Height = req.width, req.height
			return t.model.SetViewport(v)
		})
	})
	if err != nil {
		t.requestMu.Lock()
		if t.requested == req {
			t.requested = outlineRequest{}
		}
		t.requestMu.Unlock()
		if !errors.Is(err, app.ErrClosed) && !errors.Is(err, app.ErrWindowClosed) {
			t.r.control.refused.Store(&err)
			_ = t.owner.Invalidate(t.window)
		}
	}
}
func (t *outlineBinding) refuseGeometry() {
	t.measurement.Store(nil)
	if !t.geometryRefused {
		t.geometryRefused = true
		postControl(t.r, func() { t.fail(ErrTreeBudget) })
	}
}
func syncTree(r *record) (result error) {
	t := r.control.outline
	defer func() {
		if result != nil && t != nil {
			t.projectionValid.Store(false)
			t.measurement.Store(nil)
		}
	}()
	if t == nil || !t.live() {
		return ErrTreeClosed
	}
	if err := validateNativeTree(r.node); err != nil {
		return err
	}
	snap, err := t.model.Snapshot()
	if err != nil {
		return err
	}
	if err := treeProjectionBudget(r.node.ID, snap); err != nil {
		return err
	}
	if snap.State.Generation != t.projection.State.Generation {
		outlineRegistry.Lock()
		if outlineRegistry.next == math.MaxUint64 {
			outlineRegistry.Unlock()
			return ErrTreeBinding
		}
		outlineRegistry.next++
		t.epoch = outlineRegistry.next
		outlineRegistry.Unlock()
	}
	// State.Pending can change without a new generation. Compare all validated
	// state and bounded rows; a completed owner publication must not schedule
	// another full scene layout when the projection is unchanged.
	dirty := !t.projectionValid.Load() || t.projection.State != snap.State || !slices.Equal(t.projection.Rows, snap.Rows) || t.projectedDisabled != r.control.disabled || t.projectedHelp != r.node.Accessibility.Help
	t.projection = snap
	if len(t.rows) > len(snap.Rows) {
		for _, row := range t.rows[len(snap.Rows):] {
			row.Delete()
		}
		clear(t.rows[len(snap.Rows):])
		t.rows = t.rows[:len(snap.Rows)]
	}
	for len(t.rows) < len(snap.Rows) {
		row := tree.New[outlineNativeRow](t.pane)
		row.binding = t
		row.arrow = core.NewText(row)
		row.label = core.NewText(row)
		installTreeText(row.arrow)
		installTreeText(row.label)
		t.rows = append(t.rows, row)
	}
	for i, row := range snap.Rows {
		native := t.rows[i]
		rowDirty := dirty || native.row != row
		native.row = row
		value := ""
		if row.HasChildren {
			value = "▸"
			if row.Expanded {
				value = "▾"
			}
		}
		native.arrow.SetText(value)
		native.label.SetText(html.EscapeString(row.Label))
		native.SetState((row.Clip.Width <= 0 || row.Clip.Height <= 0), states.Invisible)
		native.SetState(row.Selected, states.Selected)
		native.label.SetState(row.Selected, states.Selected)
		if rowDirty {
			native.Update()
		}
	}
	value := float32(0)
	if maximum := treeMaxY(snap.State); maximum > 0 {
		value = float32(float64(snap.State.Viewport.ScrollY) / float64(maximum))
	}
	// SetValue requests paint even for an identical value. A native drag may
	// leave a noncanonical value despite an unchanged accepted model pixel.
	if t.vertical.Value != value {
		t.vertical.SetValue(value)
	}
	t.pager[0].SetState(snap.State.Viewport.ScrollY == 0, states.Disabled)
	t.pager[1].SetState(snap.State.Viewport.ScrollY == treeMaxY(snap.State), states.Disabled)
	first, last := 0, 0
	for _, row := range snap.Rows {
		if !(row.Clip.Width <= 0 || row.Clip.Height <= 0) {
			if first == 0 {
				first = row.Index + 1
			}
			last = row.Index + 1
		}
	}
	suffix := ""
	if snap.State.Pending {
		suffix = " · building"
	}
	t.status.SetText(fmtTreeRange(first, last, snap.State.VisibleRows, suffix))
	if dirty {
		t.host.Update()
	}
	t.projectedDisabled = r.control.disabled
	t.projectedHelp = r.node.Accessibility.Help
	t.projectionValid.Store(true)
	return nil
}

// Admission precedes any native pool growth or generated target allocation.
// Counts raw keys/labels, worst-case HTML escaping, row+disclosure target IDs,
// and the encoded AX row ID. Shared strings, shaping/allocator overhead, and
// the sealed immutable data/index are outside this logical payload budget.
func treeProjectionBudget(id string, snap virtual.OutlineSnapshot) error {
	if len(id) > treeNativeMaxKey || len(snap.Rows) > treeNativeMaxRows {
		return ErrTreeBudget
	}
	encodedHost := base64.RawURLEncoding.EncodedLen(len(id))
	total := 256
	for _, row := range snap.Rows {
		if len(row.ID) > treeNativeMaxKey || len(row.Parent) > treeNativeMaxKey || len(row.Label) > treeNativeMaxLabel || !validTreeRect(row.Rect) || !validTreeRect(row.Clip) {
			return ErrTreeBudget
		}
		encodedKey := base64.RawURLEncoding.EncodedLen(len(row.ID))
		rowID := 5 + encodedHost + 5 + encodedKey
		disclosureID := 5 + encodedHost + 12 + encodedKey
		bytes := len(row.ID) + len(row.Parent) + 6*len(row.Label) + rowID + disclosureID + 5 + base64.RawURLEncoding.EncodedLen(rowID)
		if bytes > treeNativeMaxText-total {
			return ErrTreeBudget
		}
		total += bytes
	}
	return nil
}
func installTreeText(w *core.Text) {
	w.Styler(func(s *styles.Style) {
		s.Padding.Zero()
		s.Margin.Zero()
		s.Border.Width.Zero()
		s.Grow.SetZero()
		s.Text.WhiteSpace = text.WhiteSpacePre
		s.SetAbilities(false, abilities.Focusable, abilities.Selectable, abilities.Clickable, abilities.Activatable, abilities.DoubleClickable, abilities.TripleClickable, abilities.Draggable, abilities.LongPressable, abilities.LongHoverable, abilities.Slideable)
	})
}
func validTreeRect(r virtual.OutlineRect) bool {
	return r.Width >= 0 && r.Height >= 0 && r.Width <= gridNativeMaxDimension && r.Height <= gridNativeMaxDimension && r.X >= -gridNativeMaxCoordinate && r.X <= gridNativeMaxCoordinate && r.Y >= -gridNativeMaxCoordinate && r.Y <= gridNativeMaxCoordinate && r.X+r.Width <= gridNativeMaxCoordinate && r.Y+r.Height <= gridNativeMaxCoordinate
}
func disposeTree(r *record) {
	if r == nil || r.control == nil || r.control.outline == nil {
		return
	}
	t := r.control.outline
	t.closed.Store(true)
	t.projectionValid.Store(false)
	t.armed = nil
	t.measurement.Store(nil)
	outlineRegistry.Lock()
	if outlineRegistry.models[t.model] == t {
		if old := t.previous; old != nil && !old.closed.Load() && !old.r.control.retired {
			outlineRegistry.models[t.model] = old
		} else {
			delete(outlineRegistry.models, t.model)
		}
	}
	outlineRegistry.Unlock()
}
func prepareTreeReplacements(previous *record, next ui.Node) func() {
	models := map[*virtual.Outline]string{}
	var nodes func(ui.Node)
	nodes = func(n ui.Node) {
		if n.Kind == ui.DataTree {
			models[n.Tree] = n.ID
		}
		for _, c := range n.Children {
			nodes(c)
		}
	}
	nodes(next)
	allowed := map[*virtual.Outline]*outlineBinding{}
	var records func(*record)
	records = func(r *record) {
		if r == nil {
			return
		}
		if r.control != nil && r.control.outline != nil {
			t := r.control.outline
			if id, ok := models[t.model]; ok && id != r.node.ID {
				allowed[t.model] = t
			}
		}
		for _, c := range r.children {
			records(c)
		}
	}
	records(previous)
	outlineRegistry.Lock()
	for model, t := range allowed {
		if outlineRegistry.models[model] == t {
			outlineRegistry.preparing[model] = t
		}
	}
	outlineRegistry.Unlock()
	return func() {
		outlineRegistry.Lock()
		defer outlineRegistry.Unlock()
		for model, t := range allowed {
			if outlineRegistry.preparing[model] == t {
				delete(outlineRegistry.preparing, model)
			}
		}
	}
}
func bindTreeScopes(root *record) {
	scope := ui.FocusScope(accessibilityFocusProjection(root, ""))
	allowed := map[*virtual.Outline]bool{}
	var collect func(ui.Node, bool)
	collect = func(n ui.Node, disabled bool) {
		disabled = disabled || n.Disabled
		if disabled || n.Kind == ui.Dialog && !n.Open {
			return
		}
		if n.Tree != nil {
			allowed[n.Tree] = true
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
		if r.control != nil && r.control.outline != nil {
			t := r.control.outline
			t.scope = allowed[r.node.Tree]
			t.captureTabOwners()
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
func (t *outlineBinding) captureTabOwners() {
	clear(t.tabOwners)
	t.tabOwners = t.tabOwners[:0]
	t.ancestorRefused = false
	depth := 0
	for p := tree.Node(t.host); p != nil; p = p.AsTree().Parent {
		if depth >= 1024 {
			t.ancestorRefused = true
			break
		}
		depth++
		if tabs, ok := p.(*core.Tabs); ok {
			page, _ := tabs.CurrentTab()
			if tree.IsNil(page) || t.host.AsTree().ParentLevel(page) < 0 {
				page = nil
			}
			t.tabOwners = append(t.tabOwners, gridTabOwner{tabs: tabs, page: page})
		}
	}
}

func fmtTreeRange(first, last, total int, suffix string) string {
	return fmt.Sprintf("%d–%d of %d%s", first, last, total, suffix)
}
