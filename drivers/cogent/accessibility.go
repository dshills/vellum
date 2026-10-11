package cogent

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent/internal/textinput"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/workbench"
)

type accessibilityNode struct {
	ID            string   `json:"id"`
	DomainID      string   `json:"domain_id"`
	Parent        string   `json:"parent"`
	Role          string   `json:"role"`
	Name          string   `json:"name"`
	Help          string   `json:"help"`
	Value         any      `json:"value,omitempty"`
	Disabled      bool     `json:"disabled"`
	Focusable     bool     `json:"focusable"`
	Focused       bool     `json:"focused"`
	Selected      bool     `json:"selected"`
	Actions       []string `json:"actions"`
	RowIndex      int      `json:"row_index,omitempty"`
	RowCount      int      `json:"row_count,omitempty"`
	Level         int      `json:"level,omitempty"`
	Expanded      *bool    `json:"expanded,omitempty"`
	Min           *float32 `json:"min,omitempty"`
	TextClient    uint64   `json:"text_client,omitempty"`
	TextSelection *[2]int  `json:"text_selection,omitempty"`
	TextFocus     int      `json:"text_focus,omitempty"`
	Max           *float32 `json:"max,omitempty"`
}
type accessibilityAction struct {
	Kind  string
	Value string
}
type accessibilityHandler func(accessibilityAction) bool

type accessibilityBridge struct {
	token        uint64
	owner        *app.Application
	mu           sync.Mutex
	window       string
	handlers     map[string]accessibilityHandler
	lastSnapshot string
	generation   atomic.Uint64
	closed       atomic.Bool
	pendingFocus atomic.Pointer[string]
	done         chan struct{}
}

var accessibilityToken atomic.Uint64
var accessibilityRegistry = struct {
	sync.Mutex
	bridges map[uint64]*accessibilityBridge
	windows map[string]*accessibilityBridge
}{bridges: map[uint64]*accessibilityBridge{}, windows: map[string]*accessibilityBridge{}}

func newAccessibilityBridge(owner *app.Application) *accessibilityBridge {
	b := &accessibilityBridge{token: accessibilityToken.Add(1), owner: owner, handlers: map[string]accessibilityHandler{}, done: make(chan struct{})}
	accessibilityRegistry.Lock()
	accessibilityRegistry.bridges[b.token] = b
	accessibilityRegistry.Unlock()
	return b
}
func (b *accessibilityBridge) bind(window string) {
	b.mu.Lock()
	b.window = window
	b.mu.Unlock()
	accessibilityRegistry.Lock()
	accessibilityRegistry.windows[window] = b
	accessibilityRegistry.Unlock()
}
func (b *accessibilityBridge) close() {
	if !b.closed.CompareAndSwap(false, true) {
		return
	}
	close(b.done)
	b.generation.Add(1)
	b.mu.Lock()
	b.handlers = nil
	window := b.window
	b.mu.Unlock()
	accessibilityRegistry.Lock()
	delete(accessibilityRegistry.bridges, b.token)
	if accessibilityRegistry.windows[window] == b {
		delete(accessibilityRegistry.windows, window)
	}
	accessibilityRegistry.Unlock()
}
func accessibilityDeliver(token, generation uint64, id string, action accessibilityAction) bool {
	accessibilityRegistry.Lock()
	b := accessibilityRegistry.bridges[token]
	accessibilityRegistry.Unlock()
	if b == nil || b.closed.Load() || b.generation.Load() != generation {
		return false
	}
	b.mu.Lock()
	handler := b.handlers[id]
	window := b.window
	b.mu.Unlock()
	if handler == nil || window == "" {
		return false
	}
	// Native main-thread callbacks never enter Do or wait for the owner. The
	// generation is rechecked when an admitted action reaches the bounded queue.
	err := b.owner.PostWindow(window, func() {
		if b.closed.Load() || b.generation.Load() != generation {
			return
		}
		b.mu.Lock()
		current := b.handlers[id]
		b.mu.Unlock()
		if current != nil && current(action) {
			_ = b.owner.Invalidate(window)
		}
	})
	return err == nil
}
func (b *accessibilityBridge) update(nodes []accessibilityNode, handlers map[string]accessibilityHandler) {
	if b.closed.Load() {
		return
	}
	semantic, err := json.Marshal(nodes)
	if err != nil {
		return
	}
	b.mu.Lock()
	if b.closed.Load() {
		b.mu.Unlock()
		return
	}
	b.handlers = handlers
	unchanged := b.lastSnapshot == string(semantic)
	b.lastSnapshot = string(semantic)
	b.mu.Unlock()
	// Focus handlers can enqueue a duplicate invalidation after the originating
	// key event already published the same tree. An identical refresh must not
	// retire a valid native action admitted between these two publications.
	if unchanged {
		return
	}
	generation := b.generation.Add(1)
	payload, _ := json.Marshal(struct {
		Generation uint64              `json:"generation"`
		Nodes      []accessibilityNode `json:"nodes"`
	}{generation, nodes})
	nativeAccessibilityUpdate(b.token, string(payload))
}
func accessibilityFocus(b *accessibilityBridge, id string) {
	copy := id
	b.pendingFocus.Store(&copy)
}
func accessibilityApplyFocus(root *record, id string) bool {
	targets := map[string]accessibilityFocusTarget{}
	accessibilityNativeFocusTargets(root, "", targets)
	if _, ok := targets[id]; !ok {
		id = accessibilityManagedID(id)
	}
	if _, ok := ui.FocusTarget(ui.FocusScope(accessibilityFocusProjection(root, "")), id); !ok {
		return false
	}
	target, ok := targets[id]
	if !ok || accessibilityTreeNil(target.widget) {
		return false
	}
	widget := target.widget.AsWidget()
	if widget.Scene == nil || widget.Scene.RenderWindow() == nil {
		widget.StartFocus()
	} else {
		widget.SetFocus()
	}
	return true
}

func syncAccessibility(w *nativeWindow, a *app.Application) {
	if w == nil || w.root == nil || w.closing.Load() {
		return
	}
	b := windowAccessibilityBridge(w)
	if b == nil || b.closed.Load() {
		return
	}
	b.bind(w.window.ID)
	if request := b.pendingFocus.Swap(nil); request != nil {
		accessibilityApplyFocus(w.root, *request)
	}
	accessibilityBindFocusInvalidation(w.root, a, w.window.ID)
	accessibilityStampVirtualGeometry(w.root)
	accessibilityStampDialogNamespaces(w.root, "")
	nodes, handlers := accessibilitySnapshot(w.root, b)
	b.update(nodes, handlers)
}
func closeAccessibility(w *nativeWindow) {
	if b := windowAccessibilityBridge(w); b != nil {
		b.close()
	}
}

func accessibilitySnapshot(root *record, b *accessibilityBridge) ([]accessibilityNode, map[string]accessibilityHandler) {
	return accessibilitySnapshotInherited(root, b, false)
}

func accessibilitySnapshotInherited(root *record, b *accessibilityBridge, inheritedDisabled bool) ([]accessibilityNode, map[string]accessibilityHandler) {
	nodes := []accessibilityNode{}
	handlers := map[string]accessibilityHandler{}
	scoped, namespace := root, ""
	if !inheritedDisabled {
		scoped, namespace = visibleAccessibilityScope(root)
	}
	if scoped == nil {
		return nodes, handlers
	}
	var visit func(*record, string, bool)
	visit = func(r *record, parent string, disabled bool) {
		n := r.node
		disabled = disabled || n.Disabled
		if n.Kind == ui.Dialog && (!n.Open || disabled || (r.control != nil && r.control.dialog != nil && r.control.dialog.dismissed)) {
			return
		}
		name := n.Accessibility.Name
		if name == "" {
			name = n.Text
		}
		x := accessibilityNode{ID: accessibilityManagedID(n.ID), DomainID: n.ID, Parent: parent, Role: string(n.Kind), Name: name, Help: n.Accessibility.Help, Disabled: disabled, Focusable: ui.IsFocusable(n.Kind), Actions: []string{}}
		if widget, err := focusWidget(r); err == nil && !accessibilityTreeNil(widget) {
			x.Focused = widget.AsWidget().StateIs(states.Focused)
			if r.control != nil && r.control.tabs != nil {
				for _, button := range r.control.tabs.buttons {
					if !accessibilityTreeNil(button) && button.StateIs(states.Focused) {
						x.Focused = true
						break
					}
				}
			}
		}
		switch extension := n.Custom.(type) {
		case *VirtualTable:
			if extension != nil {
				x.Role = "table"
			}
		case *VirtualTree:
			if extension != nil {
				x.Role = "outline"
			}
		}
		switch n.Kind {
		case ui.Label:
			x.Value = n.Text
		case ui.TextField, ui.MultilineText:
			x.Value = n.Text
			selection := editor.Selection{}
			if n.Kind == ui.MultilineText {
				snapshot := n.Editor.Snapshot()
				x.Value = snapshot.VisibleText
				selection = snapshot.Selection
			}
			if r.control != nil {
				x.TextClient = r.control.client
				if n.Kind == ui.TextField && r.control.fieldState != nil {
					x.Value = r.control.fieldState.Text
					selection = editor.Selection{Anchor: r.control.fieldState.Anchor, Focus: r.control.fieldState.Focus}
				}
			}
			text, _ := x.Value.(string)
			anchor, _ := textinput.UTF16Offset(text, selection.Anchor)
			focus, _ := textinput.UTF16Offset(text, selection.Focus)
			rangeValue := [2]int{min(anchor, focus), max(anchor, focus) - min(anchor, focus)}
			x.TextSelection = &rangeValue
			x.TextFocus = focus
		case ui.Checkbox:
			x.Value = n.Checked
		case ui.Slider, ui.NumberField:
			x.Value = n.Number
			minimum, maximum := n.Numeric.Min, n.Numeric.Max
			x.Min = &minimum
			x.Max = &maximum
		case ui.Selector, ui.Tabs:
			x.Value = n.SelectedID
			for _, o := range n.Options {
				if o.ID == n.SelectedID {
					x.Value = o.Label
					break
				}
			}
		}
		if !disabled {
			if x.Focusable {
				x.Actions = append(x.Actions, "focus")
			}
			if n.Kind == ui.Button && n.Activate != nil {
				x.Actions = append(x.Actions, "press")
			}
			if n.Kind == ui.Checkbox && n.Change != nil {
				x.Actions = append(x.Actions, "press", "set")
			}
			if n.Kind == ui.MultilineText {
				x.Actions = append(x.Actions, "set", "selection")
			}
			if n.Kind == ui.TextField && n.Change != nil {
				x.Actions = append(x.Actions, "selection")
			}
			if (n.Kind == ui.TextField || n.Kind == ui.Slider || n.Kind == ui.NumberField) && n.Change != nil {
				x.Actions = append(x.Actions, "set")
			}
			if (n.Kind == ui.Slider || n.Kind == ui.NumberField) && n.Change != nil {
				x.Actions = append(x.Actions, "increment", "decrement")
			}
			if (n.Kind == ui.Selector || n.Kind == ui.Tabs) && n.Change != nil {
				x.Actions = append(x.Actions, "set")
			}
			if n.Kind == ui.Dialog && n.Dismiss != nil {
				x.Actions = append(x.Actions, "cancel")
			}
		}
		nodes = append(nodes, x)
		if len(x.Actions) > 0 {
			allowed := append([]string(nil), x.Actions...)
			handlers[x.ID] = func(action accessibilityAction) bool {
				if !containsAccessibilityAction(allowed, action.Kind) {
					return false
				}
				if action.Kind == "focus" {
					accessibilityFocus(b, x.ID)
					return true
				}
				return applyAccessibilityAction(r, action)
			}
		}
		if n.Kind == ui.Extension {
			extra, actions := extensionAccessibility(r, b, disabled)
			nodes = append(nodes, extra...)
			for id, action := range actions {
				handlers[id] = action
			}
		}
		if n.Kind == ui.Tabs || n.Kind == ui.Selector {
			for _, o := range n.Options {
				id := accessibilityOptionID(n.ID, o.ID)
				child := accessibilityNode{ID: id, Parent: x.ID, Role: "option", DomainID: o.ID, Name: o.Label, Selected: o.ID == n.SelectedID, Value: o.ID == n.SelectedID, Disabled: disabled || o.Disabled, Actions: []string{}}
				if !child.Disabled && n.Change != nil {
					child.Actions = []string{"press"}
					optionID := o.ID
					handlers[id] = func(action accessibilityAction) bool {
						if action.Kind != "press" {
							return false
						}
						return applyAccessibilityAction(r, accessibilityAction{Kind: "set", Value: optionID})
					}
				}
				nodes = append(nodes, child)
			}
		}
		for _, c := range r.children {
			if n.Kind == ui.Tabs && c.node.ID != n.SelectedID {
				continue
			}
			visit(c, x.ID, disabled)
		}
	}
	visit(scoped, "", inheritedDisabled)
	return namespaceAccessibility(nodes, handlers, namespace, b)
}
func accessibilityIDPart(id string) string    { return base64.RawURLEncoding.EncodeToString([]byte(id)) }
func accessibilityManagedID(id string) string { return "node:" + accessibilityIDPart(id) }
func accessibilityRowID(parent, id string) string {
	return "row:" + accessibilityIDPart(parent) + ":" + accessibilityIDPart(id)
}
func accessibilityPageID(parent, id string) string {
	return "page:" + accessibilityIDPart(parent) + ":" + accessibilityIDPart(id)
}
func accessibilityOptionID(parent, option string) string {
	return "option:" + accessibilityIDPart(parent) + ":" + accessibilityIDPart(option)
}
func containsAccessibilityAction(actions []string, kind string) bool {
	for _, s := range actions {
		if s == kind {
			return true
		}
	}
	return false
}
func applyAccessibilityAction(r *record, action accessibilityAction) bool {
	n := r.node
	if n.Disabled || (r.control != nil && (r.control.disabled || r.control.retired)) {
		return false
	}
	if action.Kind == "press" && n.Kind == ui.Button && n.Activate != nil {
		n.Activate()
		return true
	}
	if action.Kind == "cancel" && n.Kind == ui.Dialog && n.Open && n.Dismiss != nil {
		n.Dismiss()
		return true
	}
	if n.Kind == ui.MultilineText || (n.Kind == ui.TextField && action.Kind == "selection") {
		if action.Kind == "selection" {
			var request struct {
				Text          string `json:"text"`
				Start, Length int
			}
			if json.Unmarshal([]byte(action.Value), &request) != nil || request.Start < 0 || request.Length < 0 || request.Start > int(^uint(0)>>1)-request.Length {
				return false
			}
			u, err := textinput.Decode(textinput.Update{Text: request.Text, BeforeText: request.Text, Anchor: request.Start, Focus: request.Start + request.Length, MarkedStart: -1, MarkedEnd: -1})
			if err != nil {
				return false
			}
			selection := editor.Selection{Anchor: u.RuneAnchor, Focus: u.RuneFocus}
			if n.Kind == ui.MultilineText {
				snapshot := n.Editor.Snapshot()
				if snapshot.Composing || snapshot.VisibleText != request.Text || n.Editor.Select(selection) != nil {
					return false
				}
			} else {
				if r.control.fieldState == nil || r.control.fieldState.Text != request.Text {
					return false
				}
				r.control.fieldSelection = &selection
				r.control.textEpoch++
			}
			_ = r.control.a.Invalidate(r.control.window)
			return true
		}
		if n.Kind == ui.MultilineText && action.Kind == "set" {
			if err := n.Editor.Replace(action.Value, editor.Selection{}); err != nil {
				return false
			}
			committedEditorChange(r)
			return true
		}
		return false
	}
	if n.Change == nil {
		return false
	}
	change := ui.ChangeEvent{ID: n.ID, Kind: n.Kind}
	switch n.Kind {
	case ui.Checkbox:
		if action.Kind == "press" {
			change.Checked = !n.Checked
		} else if action.Kind == "set" && (action.Value == "true" || action.Value == "false") {
			change.Checked = action.Value == "true"
		} else {
			return false
		}
	case ui.TextField:
		if action.Kind != "set" || len(action.Value) > textinput.MaxDocumentBytes || strings.ContainsRune(action.Value, 0) {
			return false
		}
		change.Text = action.Value
	case ui.Selector, ui.Tabs:
		if action.Kind != "set" {
			return false
		}
		selected := ""
		for _, o := range n.Options {
			if o.ID == action.Value {
				if o.Disabled {
					return false
				}
				selected = o.ID
				break
			}
		}
		if selected == "" {
			matches := 0
			for _, o := range n.Options {
				if o.Label == action.Value && !o.Disabled {
					matches++
					selected = o.ID
				}
			}
			if matches != 1 {
				return false
			}
		}
		change.SelectedID = selected
	case ui.Slider, ui.NumberField:
		value := float64(n.Number)
		switch action.Kind {
		case "increment":
			value += float64(n.Numeric.Step)
		case "decrement":
			value -= float64(n.Numeric.Step)
		case "set":
			parsed, err := strconv.ParseFloat(strings.TrimSpace(action.Value), 64)
			if err != nil {
				return false
			}
			value = parsed
		default:
			return false
		}
		if math.IsNaN(value) || math.IsInf(value, 0) || value > math.MaxFloat32 || value < -math.MaxFloat32 {
			return false
		}
		number, ok := ui.ClampNumeric(float32(value), n.Numeric)
		if !ok {
			return false
		}
		change.Number = number
	default:
		return false
	}
	n.Change(change)
	if r.control != nil {
		r.control.committed = true
	}
	return true
}

// bindAccessibilityKeyboard runs on the originating native event loop. Tab and
// Shift-Tab use descriptor order; Escape dismisses the active managed dialog.
func bindAccessibilityKeyboard(w *nativeWindow, a *app.Application, onError func(error)) {
	handle := func(e events.Event) {
		chord := strings.ToLower(string(e.KeyChord()))
		if chord != "tab" && chord != "shift+tab" && chord != "escape" {
			return
		}
		err := a.Do(func() {
			w.rootMu.Lock()
			defer w.rootMu.Unlock()
			if w.root == nil || w.closing.Load() {
				return
			}
			if chord == "escape" {
				if r := focusedManagedTextRecord(w.root); r != nil && r.control.editor != nil && r.node.Editor.Snapshot().Composing {
					if err := r.node.Editor.CancelComposition(); err != nil {
						editorError(r, err)
					} else {
						syncEditor(r)
					}
					e.SetHandled()
					requestTextPublication(w.window.ID)
					return
				}
				scope := ui.FocusScope(accessibilityFocusProjection(w.root, ""))
				if scope.Kind == ui.Dialog && scope.Dismiss != nil {
					scope.Dismiss()
					_ = a.Invalidate(w.window.ID)
					e.SetHandled()
				}
				return
			}
			if accessibilityExtensionFocused(w.root) {
				return
			}
			if editor := focusedManagedTextRecord(w.root); editor != nil && editor.node.Kind == ui.MultilineText && editor.node.TabInserts {
				return
			}
			current := accessibilityFocusedID(w.root)
			id, ok := ui.NextFocus(accessibilityFocusProjection(w.root, ""), current, chord == "shift+tab")
			if ok && accessibilityApplyFocus(w.root, id) {
				e.SetHandled()
				syncAccessibility(w, a)
			}
		})
		if err != nil && onError != nil {
			onError(err)
		}
	}
	w.body.Scene.OnFirst(events.KeyChord, handle)
	// Modal Scenes do not have the main Scene as an ancestor; integration calls
	// bindAccessibilityDialogKeyboard on each newly created modal body as well.
	w.body.SetProperty("vellum.accessibility-keyboard", handle)
}
func bindAccessibilityDialogKeyboard(r *record, body *core.Body) {
	if r == nil || accessibilityTreeNil(r.widget) || body == nil || body.Scene == nil {
		return
	}
	for parent := r.widget.AsTree().Parent; !accessibilityTreeNil(parent); parent = parent.AsTree().Parent {
		if handler, ok := parent.AsTree().Property("vellum.accessibility-keyboard").(func(events.Event)); ok {
			body.Scene.OnFirst(events.KeyChord, handler)
			// Scene.OnFirst is reached through the event focus's ancestors.
			// Demand layout can precede deferred Shown/default focus, so seed
			// the explicit start target before the first queued key arrives.
			if target := accessibilityDialogInitialFocus(r, body); !accessibilityTreeNil(target) {
				body.Scene.Events.SetStartFocus(target)
			}
			return
		}
	}
}

// Choose only this modal's visible enabled targets; a nested modal owns a
// different Scene and will establish its own start focus. With no managed
// focus target the body still anchors the Escape ancestor-handler chain.
func accessibilityDialogInitialFocus(r *record, body *core.Body) core.Widget {
	if r == nil || accessibilityTreeNil(r.widget) || body == nil || r.node.Kind != ui.Dialog || !r.node.Open || controlDisabled(r) {
		return nil
	}
	projection := accessibilityFocusProjection(r, "")
	var closeNested func(*ui.Node)
	closeNested = func(n *ui.Node) {
		if n.Kind == ui.Dialog {
			n.Open = false
		}
		for i := range n.Children {
			closeNested(&n.Children[i])
		}
	}
	for i := range projection.Children {
		closeNested(&projection.Children[i])
	}
	targets := map[string]accessibilityFocusTarget{}
	accessibilityNativeFocusTargets(r, "", targets)
	for _, id := range ui.FocusOrder(projection) {
		if target, ok := targets[id]; ok && !accessibilityTreeNil(target.widget) {
			return target.widget
		}
	}
	return body
}

func accessibilityFocusedID(root *record) string {
	targets := map[string]accessibilityFocusTarget{}
	accessibilityNativeFocusTargets(root, "", targets)
	for _, id := range ui.FocusOrder(accessibilityFocusProjection(root, "")) {
		if targets[id].focused {
			return id
		}
	}
	return ""
}

// accessibilityHost publishes geometry only after children complete their
// layout/render pass. It never enters the application owner or reads models.
type accessibilityHost struct {
	core.Frame
	window *nativeWindow
}

func newAccessibilityHost(w *nativeWindow) *core.Frame {
	host := tree.New[accessibilityHost](w.body)
	host.window = w
	return &host.Frame
}
func (h *accessibilityHost) RenderWidget() {
	h.Frame.RenderWidget()
	publishAccessibilityGeometry(h.window)
}
func publishAccessibilityGeometry(w *nativeWindow) {
	if w == nil || w.closing.Load() {
		return
	}
	b := windowAccessibilityBridge(w)
	if b == nil || b.closed.Load() {
		return
	}
	w.rootMu.Lock()
	bounds := accessibilityBounds(w.root)
	textBounds := accessibilityTextGeometry(w.root)
	w.rootMu.Unlock()
	payload, _ := json.Marshal(bounds)
	nativeAccessibilityGeometry(b.token, string(payload))
	textPayload, _ := json.Marshal(textBounds)
	nativeAccessibilityTextGeometry(b.token, string(textPayload))
}
func accessibilityBounds(root *record) map[string][4]float32 {
	out := map[string][4]float32{}
	var bounds func(string, core.Widget)
	bounds = func(id string, widget core.Widget) {
		if accessibilityTreeNil(widget) {
			return
		}
		wb := widget.AsWidget()
		box := wb.Geom.TotalBBox
		offsetX, offsetY := 0, 0
		if wb.Scene != nil {
			offsetX = wb.Scene.SceneGeom.Pos.X
			offsetY = wb.Scene.SceneGeom.Pos.Y
		}
		out[id] = [4]float32{float32(box.Min.X + offsetX), float32(box.Min.Y + offsetY), float32(box.Dx()), float32(box.Dy())}
	}
	var visit func(*record)
	visit = func(r *record) {
		if r == nil {
			return
		}
		widget := r.widget
		if r.control != nil && r.control.dialog != nil {
			widget = r.control.dialog.host
		}
		bounds(accessibilityManagedID(r.node.ID), widget)
		if r.node.Kind == ui.Tabs && r.control != nil && r.control.tabs != nil {
			for _, o := range r.node.Options {
				bounds(accessibilityOptionID(r.node.ID, o.ID), r.control.tabs.buttons[o.ID])
			}
		}
		if r.node.Kind == ui.Selector {
			for _, o := range r.node.Options {
				bounds(accessibilityOptionID(r.node.ID, o.ID), r.widget)
			}
		}
		extensionAccessibilityBounds(r, out, bounds)
		for _, c := range r.children {
			visit(c)
		}
	}
	visit(root)
	return out
}

// AccessibilityInspection is an in-process query of the actual native tree.
// It is useful for acceptance tests; it does not certify screen-reader behavior.
type AccessibilityInspection struct {
	Generation   uint64                 `json:"generation"`
	Reachable    bool                   `json:"reachable"`
	FocusedID    string                 `json:"focused_id"`
	WindowRole   string                 `json:"window_role"`
	NativeWindow bool                   `json:"native_window"`
	Elements     []AccessibilityElement `json:"elements"`
}
type AccessibilityElement struct {
	ID            string     `json:"id"`
	Role          string     `json:"role"`
	Name          string     `json:"name"`
	Value         any        `json:"value"`
	Enabled       bool       `json:"enabled"`
	TextSelection [2]int     `json:"text_selection"`
	InsertionLine int        `json:"insertion_line"`
	Frame         [4]float64 `json:"frame"`
	Selected      bool       `json:"selected"`
	Expanded      bool       `json:"expanded"`
	Parent        string     `json:"parent"`
	DomainID      string     `json:"domain_id"`
	HitTestID     string     `json:"hit_test_id"`
}

func lookupAccessibility(window string) (*accessibilityBridge, error) {
	accessibilityRegistry.Lock()
	b := accessibilityRegistry.windows[window]
	accessibilityRegistry.Unlock()
	if b == nil || b.closed.Load() {
		return nil, fmt.Errorf("cogent: native accessibility window %q unavailable", window)
	}
	return b, nil
}

// A modal's retained host renders in its own Scene. Publishing its geometry
// here handles resize and scroll independently of the main Scene's host.
type accessibilityDialogHost struct {
	core.Frame
	record *record
}

func newAccessibilityDialogHost(parent tree.Node, r *record) *accessibilityDialogHost {
	host := tree.New[accessibilityDialogHost](parent)
	host.record = r
	return host
}
func (h *accessibilityDialogHost) RenderWidget() {
	h.Frame.RenderWidget()
	r := h.record
	if r == nil || r.control == nil {
		return
	}
	b, err := lookupAccessibility(r.control.window)
	if err != nil {
		return
	}
	bounds := accessibilityBounds(r)
	if prefix, _ := h.Property("vellum.accessibility-namespace").(string); prefix != "" {
		qualified := make(map[string][4]float32, len(bounds))
		for id, rect := range bounds {
			qualified[prefix+id] = rect
		}
		bounds = qualified
	}
	payload, _ := json.Marshal(bounds)
	nativeAccessibilityGeometry(b.token, string(payload))
	textBounds := accessibilityTextGeometry(r)
	if prefix, _ := h.Property("vellum.accessibility-namespace").(string); prefix != "" {
		qualified := map[string]any{}
		for id, value := range textBounds {
			qualified[prefix+id] = value
		}
		textBounds = qualified
	}
	textPayload, _ := json.Marshal(textBounds)
	nativeAccessibilityTextGeometry(b.token, string(textPayload))
}

func accessibilityBindFocusInvalidation(root *record, a *app.Application, window string) {
	targets := map[string]accessibilityFocusTarget{}
	accessibilityNativeFocusTargets(root, "", targets)
	for _, target := range targets {
		widget := target.widget
		if accessibilityTreeNil(widget) {
			continue
		}
		wb := widget.AsWidget()
		if bound, _ := wb.Property("vellum.accessibility-focus-bound").(bool); !bound {
			wb.SetProperty("vellum.accessibility-focus-bound", true)
			wb.OnFocus(func(events.Event) { _ = a.Invalidate(window) })
			wb.OnFocusLost(func(events.Event) { _ = a.Invalidate(window) })
		}
	}
}

func accessibilityFocusDescriptor(root *record) ui.Node {
	if root == nil {
		return ui.Node{}
	}
	n := root.node
	if n.Kind == ui.Dialog && root.control != nil && root.control.dialog != nil && root.control.dialog.dismissed {
		n.Open = false
	}
	n.Children = make([]ui.Node, 0, len(root.children))
	for _, c := range root.children {
		n.Children = append(n.Children, accessibilityFocusDescriptor(c))
	}
	return n
}

func accessibilityExtensionFocused(root *record) bool {
	if root == nil {
		return false
	}
	_, isWorkbench := root.node.Custom.(*Workbench)
	if root.node.Kind == ui.Extension && !isWorkbench && !accessibilityTreeNil(root.widget) && root.widget.AsWidget().ContainsFocus() {
		return true
	}
	if w, ok := root.node.Custom.(*Workbench); ok {
		focused := false
		accessibilityWorkbenchLeaves(w, func(g *workbench.Group) {
			if accessibilityExtensionFocused(w.records[g.Active]) {
				focused = true
			}
		})
		if focused {
			return true
		}
	}
	for _, child := range root.children {
		if accessibilityExtensionFocused(child) {
			return true
		}
	}
	return false
}

// applyAccessibleFocus accepts ordinary root IDs or qualified native identifiers
// returned by inspection for managed controls inside active dock panels.
func applyAccessibleFocus(root *record, request ui.FocusRequest, last *ui.FocusRequest) error {
	if request.ID == "" || request == *last {
		return nil
	}
	if !accessibilityApplyFocus(root, accessibilityRequestedFocusID(root, request.ID)) {
		return fmt.Errorf("cogent: focus target %q is hidden, absent, disabled, or not focusable", request.ID)
	}
	*last = request
	return nil
}

// Application focus requests prefer a root descriptor's domain ID when its
// spelling also happens to equal another control's native identifier. Native
// accessibility actions already carry exact identifiers and bypass this rule.
func accessibilityRequestedFocusID(root *record, id string) string {
	targets := map[string]accessibilityFocusTarget{}
	accessibilityNativeFocusTargets(root, "", targets)
	managed := accessibilityManagedID(id)
	if _, exists := targets[managed]; exists {
		return managed
	}
	return id
}

// tree.IsNil handles retired tree.This values, but its promoted AsTree call
// dereferences a typed-nil embedded widget pointer. Reject that pointer first.
func accessibilityTreeNil(node tree.Node) bool {
	if node == nil {
		return true
	}
	value := reflect.ValueOf(node)
	if value.Kind() == reflect.Pointer && value.IsNil() {
		return true
	}
	return tree.IsNil(node)
}
