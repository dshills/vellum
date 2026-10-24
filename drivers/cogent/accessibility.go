package cogent

import (
	"encoding/base64"
	"encoding/json"
	"errors"
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

// Grid-prefixed fields retain the existing bounded virtual-view wire protocol;
// both managed tables and outlines use its pending/completed transaction.
type accessibilityNode struct {
	// pressFocusID is Go-only metadata for the managed Grid cell action.
	// Workbench namespace wrapping qualifies this exact pending host focus.
	activationGuard      func() bool // Go-only originating managed input lease
	pressFocusID         string
	ID                   string      `json:"id"`
	DomainID             string      `json:"domain_id"`
	Parent               string      `json:"parent"`
	Role                 string      `json:"role"`
	Secure               bool        `json:"secure,omitempty"`
	Name                 string      `json:"name"`
	Help                 string      `json:"help"`
	Value                any         `json:"value,omitempty"`
	Disabled             bool        `json:"disabled"`
	Focusable            bool        `json:"focusable"`
	Focused              bool        `json:"focused"`
	Selected             bool        `json:"selected"`
	Actions              []string    `json:"actions"`
	Orientation          string      `json:"orientation,omitempty"`
	RowIndex             int         `json:"row_index,omitempty"`
	RowCount             int         `json:"row_count,omitempty"`
	ColumnIndex          int         `json:"column_index,omitempty"`
	ColumnCount          int         `json:"column_count,omitempty"`
	GridGeneration       uint64      `json:"grid_generation,omitempty"`
	GridScope            string      `json:"grid_scope,omitempty"`
	GridIncarnation      uint64      `json:"grid_incarnation,omitempty"`
	GridPending          bool        `json:"grid_pending,omitempty"`
	GridTargetGeneration uint64      `json:"grid_target_generation,omitempty"`
	GridFrame            *[4]float32 `json:"grid_frame,omitempty"`
	Level                int         `json:"level,omitempty"`
	Expanded             *bool       `json:"expanded,omitempty"`
	Min                  *float32    `json:"min,omitempty"`
	TextClient           uint64      `json:"text_client,omitempty"`
	TextSelection        *[2]int     `json:"text_selection,omitempty"`
	TextFocus            int         `json:"text_focus,omitempty"`
	Max                  *float32    `json:"max,omitempty"`
}
type accessibilityAction struct {
	Kind  string
	Value string
}
type accessibilityHandler func(accessibilityAction) bool

type accessibilityBridge struct {
	activation        *windowActivation // immutable after construction
	visibility        *windowVisibility // immutable after native attachment construction
	token             uint64
	owner             *app.Application
	mu                sync.Mutex
	window            string
	windowRef         app.WindowRef
	handlers          map[string]accessibilityHandler
	activationGuards  map[string]func() bool
	lastSnapshot      string
	completedSnapshot string // semantic identity of the detached ready cache
	gridPublications  map[string][]accessibilityNode
	pendingBaseline   []accessibilityNode    // detached fallback wire records, no callbacks
	readyBaseline     []accessibilityNodeRef // ordered refs to immutable detached ready cache records
	pendingExhausted  bool                   // never rely on an epoch after the full counter wraps
	geometryMembers   map[string]struct{}
	geometryVersion   *accessibilityGeometryVersion
	geometryCache     accessibilityGeometryCache
	generation        atomic.Uint64
	closed            atomic.Bool
	retirement        atomic.Pointer[AccessibilityRetirement]
	pendingFocus      atomic.Pointer[string]
	done              chan struct{}
}

var accessibilityToken atomic.Uint64
var accessibilityRegistry = struct {
	sync.Mutex
	bridges map[uint64]*accessibilityBridge
	windows map[string]*accessibilityBridge
}{bridges: map[uint64]*accessibilityBridge{}, windows: map[string]*accessibilityBridge{}}

func newAccessibilityBridge(owner *app.Application) *accessibilityBridge {
	return newAccessibilityBridgeWithVisibility(owner, nil)
}

// All callback-readable fields are initialized before registry publication.
func newAccessibilityBridgeWithVisibility(owner *app.Application, visibility *windowVisibility) *accessibilityBridge {
	b := &accessibilityBridge{activation: &windowActivation{}, token: accessibilityToken.Add(1), owner: owner, visibility: visibility, handlers: map[string]accessibilityHandler{}, done: make(chan struct{})}
	accessibilityRegistry.Lock()
	accessibilityRegistry.bridges[b.token] = b
	accessibilityRegistry.Unlock()
	return b
}
func (b *accessibilityBridge) bindScoped(ref app.WindowRef) {
	window := ref.ID()
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed.Load() || (b.windowRef != (app.WindowRef{}) && b.windowRef != ref) {
		return
	}
	if _, err := scopedWindowContext(b.owner, ref); err != nil {
		return
	}
	b.windowRef = ref
	b.window = window
	// Serialize publication with retirement; a late refresh must not recreate a
	// registry entry after close removed it.
	accessibilityRegistry.Lock()
	accessibilityRegistry.windows[window] = b
	accessibilityRegistry.Unlock()
}
func (b *accessibilityBridge) close() {
	b.closeFrom("unspecified")
}

// AccessibilityRetirement identifies the first bridge teardown path. It records
// lifecycle provenance, not the event that caused that path to run.
type AccessibilityRetirement struct {
	Window     string
	Token      uint64
	Generation uint64
	Source     string
}

var ErrAccessibilityClosed = errors.New("cogent: native accessibility attachment closed")

// AccessibilityClosedError exposes a copied retirement record without retaining
// the window, bridge, owner, controls or action handlers.
type AccessibilityClosedError struct {
	Retirement AccessibilityRetirement
}

func (e *AccessibilityClosedError) Error() string {
	return fmt.Sprintf("%s (window=%q token=%d generation=%d source=%s)", ErrAccessibilityClosed, e.Retirement.Window, e.Retirement.Token, e.Retirement.Generation, e.Retirement.Source)
}
func (e *AccessibilityClosedError) Unwrap() error { return ErrAccessibilityClosed }

func (b *accessibilityBridge) closedError() error {
	if b != nil {
		if r := b.retirement.Load(); r != nil {
			return &AccessibilityClosedError{Retirement: *r}
		}
	}
	return &AccessibilityClosedError{Retirement: AccessibilityRetirement{Source: "unavailable"}}
}

func (b *accessibilityBridge) closeFrom(source string) {
	b.mu.Lock()
	if b.closed.Load() {
		b.mu.Unlock()
		return
	}
	b.retirement.Store(&AccessibilityRetirement{Window: b.window, Token: b.token, Generation: b.generation.Load(), Source: source})
	// Publish the record before signaling closure so every canceled request sees
	// the same first-close provenance, even when a later cleanup path runs.
	b.closed.Store(true)
	b.activation.close()
	b.visibility.close()
	b.generation.Add(1)
	b.handlers = nil
	b.activationGuards = nil
	b.gridPublications = nil
	b.pendingBaseline = nil
	b.readyBaseline = nil
	b.completedSnapshot = ""
	b.geometryMembers = nil
	b.geometryVersion = nil
	b.geometryCache = accessibilityGeometryCache{}
	window := b.window
	b.mu.Unlock()
	close(b.done)
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
	guard := b.activationGuards[id]
	window := b.window
	ref := b.windowRef
	b.mu.Unlock()
	if handler == nil || window == "" {
		return false
	}
	// Native main-thread callbacks never enter Do or wait for the owner. The
	// generation is rechecked when an admitted action reaches the bounded queue.
	err := b.owner.PostWindow(ref, func() {
		if b.closed.Load() || b.generation.Load() != generation || (guard != nil && !guard()) {
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
	b.updateSnapshot(nodes, handlers, false)
}

// updateOwned consumes fresh owner snapshot records and callbacks. Its caller
// must not read or mutate them afterward. The completed grid cache is detached
// even on this route, and native transport owns encoded bytes rather than nodes.
func (b *accessibilityBridge) updateOwned(nodes []accessibilityNode, handlers map[string]accessibilityHandler) {
	b.updateSnapshot(nodes, handlers, true)
}

func (b *accessibilityBridge) updateSnapshot(nodes []accessibilityNode, handlers map[string]accessibilityHandler, owned bool) {
	if b.closed.Load() {
		return
	}
	if owned && b.reuseReadySnapshot(nodes, handlers) {
		return
	}
	// A custom marshaler can mutate captured input while encoding. Eligibility
	// must stay false even if it leaves only builtin values behind afterward.
	immutableValues := accessibilityImmutableValues(nodes)
	// Reject unencodable records before they can become a retained grid view.
	semantic, err := accessibilitySemanticHinted(nodes, b.semanticAllocationHint(len(nodes)))
	if err != nil {
		return
	}
	pendingGrid := false
	for _, node := range nodes {
		pendingGrid = pendingGrid || node.GridPending
		// Only builtin scalars can safely reuse a detached completed view.
		// Custom marshalers and mutable values retain normal preparation.
		if !accessibilityImmutableScalar(node.Value) {
			immutableValues = false
		}
	}
	if owned && !pendingGrid && immutableValues {
		b.mu.Lock()
		if !b.closed.Load() && b.lastSnapshot == semantic && b.completedSnapshot == semantic && b.sameCompletedGridMetadata(nodes) {
			// Keep the detached cache, but resolve admitted actions against the
			// newest owner callbacks without retiring their semantic generation.
			b.handlers = handlers
			b.activationGuards = accessibilityActivationGuards(nodes)
			b.mu.Unlock()
			return
		}
		b.mu.Unlock()
	}
	completedSemantic := ""
	if !pendingGrid && immutableValues {
		completedSemantic = semantic
	}
	nodes, handlers, readyBaseline := b.prepareGridPublicationWithBaseline(nodes, handlers, owned, completedSemantic)
	if pendingGrid {
		// Ready records are unchanged by the scoped cache. Only pending records
		// replace input semantics with their last-completed read-only subtree.
		// Restored completed records may contain custom values even when the
		// incoming pending hosts were builtin. Check before invoking them too.
		immutableValues = immutableValues && accessibilityImmutableValues(nodes)
		semantic, err = accessibilitySemanticHinted(nodes, b.semanticAllocationHint(len(nodes)))
		if err != nil {
			return
		}
		// Pending preparation can restore values from an older completed cache.
		// Raw host-only input cannot certify those retained values as immutable.
		for _, node := range nodes {
			if !accessibilityImmutableScalar(node.Value) {
				immutableValues = false
				break
			}
		}
	}
	b.mu.Lock()
	if b.closed.Load() {
		b.mu.Unlock()
		return
	}
	b.handlers = handlers
	b.activationGuards = accessibilityActivationGuards(nodes)
	snapshot := semantic
	unchanged := b.lastSnapshot == snapshot
	if b.generation.Load() == math.MaxUint64 {
		// The legacy full path may wrap its publication counter. Permanently
		// revoke this optimization before that can alias an old native epoch.
		b.pendingExhausted = true
	}
	var pendingPatch string
	if !unchanged && owned && immutableValues && pendingGrid && !b.pendingExhausted {
		pendingPatch = accessibilityPendingPatchBaseline(b.wireBaseline(), nodes, b.generation.Load())
	}
	if !owned || !immutableValues || b.pendingExhausted {
		b.pendingBaseline = nil
		b.readyBaseline = nil
	} else if readyBaseline != nil {
		// These references share only immutable detached grid cache records;
		// ordinary records have their own detached storage. No model is kept.
		b.pendingBaseline = nil
		b.readyBaseline = readyBaseline
	} else if !unchanged {
		b.pendingBaseline = accessibilityPendingBaseline(nodes)
		b.readyBaseline = nil
	}
	b.lastSnapshot = snapshot
	if !unchanged {
		b.updateGeometryMembership(nodes)
	}
	// Focus handlers can enqueue a duplicate invalidation after the originating
	// key event already published the same tree. An identical refresh must not
	// retire a valid native action admitted between these two publications.
	if unchanged {
		b.mu.Unlock()
		return
	}
	generation := b.generation.Add(1)
	payload := accessibilityPayload(generation, snapshot)
	if pendingPatch != "" {
		nativeAccessibilityUpdatePending(b.token, payload, pendingPatch)
	} else {
		nativeAccessibilityUpdate(b.token, payload)
	}
	b.mu.Unlock()
}

// An exact match against the already validated, detached wire baseline needs
// no second JSON encoding. The completed-cache identity also guards direct
// preparation and pending transitions; Go-only cell metadata is checked apart.
// Empty snapshots retain normal encoding to distinguish null from [].
func (b *accessibilityBridge) reuseReadySnapshot(nodes []accessibilityNode, handlers map[string]accessibilityHandler) bool {
	if len(nodes) == 0 {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	baseline := b.wireBaseline()
	if b.closed.Load() || b.completedSnapshot == "" || b.completedSnapshot != b.lastSnapshot || baseline.len() != len(nodes) {
		return false
	}
	for i := range nodes {
		// SameWireNode accepts only exact builtin scalar values. Custom
		// marshalers, mutable values and every changed wire field fall back
		// to validation/encoding before retained state can be mutated.
		if nodes[i].GridPending || !accessibilitySameWireNode(baseline.at(i), nodes[i]) {
			return false
		}
	}
	if !b.sameCompletedGridMetadata(nodes) {
		return false
	}
	b.handlers = handlers
	b.activationGuards = accessibilityActivationGuards(nodes)
	return true
}

func accessibilityImmutableValues(nodes []accessibilityNode) bool {
	for _, node := range nodes {
		if !accessibilityImmutableScalar(node.Value) {
			return false
		}
	}
	return true
}

func accessibilityActivationGuards(nodes []accessibilityNode) map[string]func() bool {
	guards := map[string]func() bool{}
	for _, n := range nodes {
		if n.activationGuard != nil {
			guards[n.ID] = n.activationGuard
		}
	}
	return guards
}

// accessibilitySemantic preserves Marshal's compact representation in an owned
// immutable string. Builtin scalars use the fixed wire schema; any unsupported
// value falls back for the entire snapshot to the original standard encoder,
// preserving its custom callbacks, errors and trailing-delimiter handling.
func accessibilitySemantic(nodes []accessibilityNode) (string, error) {
	return accessibilitySemanticHinted(nodes, 0)
}

func accessibilitySemanticHinted(nodes []accessibilityNode, hint int) (string, error) {
	if encoded, ok := accessibilityBuiltinSemanticHinted(nodes, hint); ok {
		return encoded, nil
	}
	var buffer strings.Builder
	if err := json.NewEncoder(&buffer).Encode(nodes); err != nil {
		return "", err
	}
	return strings.TrimSuffix(buffer.String(), "\n"), nil
}

// semantic is the complete successful encoding result above, including the
// second validation after a pending-grid replacement. Only a generated decimal
// integer and fixed delimiters surround it; no raw application text is appended.
// This preserves the full envelope without re-encoding/copying its RawMessage.
func accessibilityPayload(generation uint64, semantic string) string {
	return `{"generation":` + strconv.FormatUint(generation, 10) + `,"nodes":` + semantic + `}`
}
func accessibilityFocus(b *accessibilityBridge, id string) {
	copy := id
	b.pendingFocus.Store(&copy)
}
func accessibilityResolveFocus(root *record, id string) (accessibilityFocusTarget, bool) {
	targets := map[string]accessibilityFocusTarget{}
	accessibilityNativeFocusTargets(root, "", targets)
	if _, ok := targets[id]; !ok {
		id = accessibilityManagedID(id)
	}
	scopeID := id
	if target, ok := targets[id]; ok && target.scopeID != "" {
		scopeID = target.scopeID
	}
	if _, ok := ui.FocusTarget(ui.FocusScope(accessibilityFocusProjection(root, "")), scopeID); !ok {
		return accessibilityFocusTarget{}, false
	}
	target, ok := targets[id]
	if !ok || accessibilityTreeNil(target.widget) {
		return accessibilityFocusTarget{}, false
	}
	return target, true
}
func accessibilityApplyFocus(root *record, id string) bool {
	target, ok := accessibilityResolveFocus(root, id)
	if !ok {
		return false
	}

	widget := target.widget.AsWidget()
	if widget.Scene == nil || widget.Scene.RenderWindow() == nil {
		widget.StartFocus()
	} else {
		widget.SetFocus()
	}
	if target.reveal {
		widget.ScrollToThis()
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
	b.bindScoped(w.window.Ref())
	if request := b.pendingFocus.Swap(nil); request != nil {
		accessibilityApplyFocus(w.root, *request)
	}
	accessibilityBindFocusInvalidation(w.root, a, w.window.ID)
	accessibilityStampVirtualGeometry(w.root)
	accessibilityStampDialogNamespaces(w.root, "")
	nodes, handlers := accessibilitySnapshot(w.root, b)
	// A modal projection retires background Cocoa elements and their geometry.
	// Removing a modal can recompose a cached main Scene without rendering its
	// accessibility host. Render the restored scope once so its newly recreated
	// elements receive current frames; unchanged scopes keep demand painting.
	scope := ""
	if len(nodes) != 0 {
		scope = nodes[0].ID
	}
	b.updateOwned(nodes, handlers)
	if scope != w.axScope {
		w.axScope = scope
		if !accessibilityTreeNil(w.host) {
			w.host.NeedsRender()
		}
	}
}
func closeAccessibility(w *nativeWindow, source string) {
	if b := windowAccessibilityBridge(w); b != nil {
		b.closeFrom(source)
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
		if toolbarOverflowHidden(r) {
			return
		}
		start := len(nodes)
		n := r.node
		disabled = disabled || n.Disabled
		if n.Kind == ui.MenuButton {
			disabled = disabled || !menuAvailable(r)
		}
		if n.Kind == ui.Dialog && (!n.Open || disabled || (r.control != nil && r.control.dialog != nil && r.control.dialog.dismissed)) {
			return
		}
		if n.Kind == ui.Separator || ((n.Kind == ui.Icon || n.Kind == ui.Image) && strings.TrimSpace(n.Accessibility.Name) == "" && strings.TrimSpace(n.Text) == "") {
			return
		} // decorative, no accessible target
		name := n.Accessibility.Name
		if name == "" || ((n.Kind == ui.Icon || n.Kind == ui.Image) && strings.TrimSpace(name) == "") {
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
		case ui.Canvas2D:
			if len(n.CanvasObjects) > 0 {
				x.Role = "canvas"
				x.Value = canvasCompletedSelection(r)
			}
		case ui.Icon:
			x.Role = "image"
		case ui.SplitPane:
			x.Focusable, x.Focused = false, false // The divider owns focus.
		case ui.StatusBar:
			x.Value = n.Text
		case ui.DataTree:
			x.Role = "outline"
			if r.control != nil && r.control.outline != nil {
				x.RowCount = r.control.outline.projection.State.VisibleRows
			}
		case ui.DataGrid:
			x.Role = "table"
			if r.control != nil && r.control.grid != nil {
				state := r.control.grid.projection.State
				x.RowCount = state.Stamp.Rows
				x.ColumnCount = state.Columns
				x.GridGeneration = state.Generation
			}
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
		case ui.MenuButton:
			opened := menuOpen(r)
			x.Expanded = &opened
			if opened {
				x.Focused = false
			}
		case ui.Checkbox, ui.Switch, ui.ToggleButton:
			x.Value = n.Checked
		case ui.ProgressBar, ui.ProgressRing, ui.ProgressGauge:
			x.Role = "progress-bar"
			x.Value = n.Number
			minimum, maximum := float32(0), float32(1)
			x.Min, x.Max = &minimum, &maximum
		case ui.Slider, ui.NumberField:
			x.Value = n.Number
			minimum, maximum := n.Numeric.Min, n.Numeric.Max
			x.Min = &minimum
			x.Max = &maximum
		case ui.RadioGroup, ui.SegmentedControl, ui.ListBox:
			x.Role = "radio-group"
			if n.Kind == ui.ListBox {
				x.Role = "list-box"
				x.RowCount = len(n.Options)
				x.activationGuard = listBoxHostGuard(r)
			}
			x.Value = n.SelectedID
			for _, option := range n.Options {
				if option.ID == n.SelectedID {
					x.Value = option.Label
					break
				}
			}
			if n.Multiple {
				x.Value = fmt.Sprintf("%d selected", len(n.SelectedIDs))
			}
			_, x.Focusable = ui.FocusTarget(n, n.ID)
			// Keyboard focus belongs to one option; group focus is a managed proxy.
			x.Focused = false
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
			if n.ContextMenu != nil && menuAvailable(r) {
				x.Actions = append(x.Actions, "show-menu")
			}
			if x.Focusable {
				x.Actions = append(x.Actions, "focus")
			}
			if n.Kind == ui.Button && n.Activate != nil {
				x.Actions = append(x.Actions, "press")
			}
			if n.Kind == ui.MenuButton {
				x.Actions = append(x.Actions, "press", "expand")
				if menuOpen(r) {
					x.Actions = append(x.Actions, "cancel", "collapse")
				}
			}
			if (n.Kind == ui.Checkbox || n.Kind == ui.Switch || n.Kind == ui.ToggleButton) && n.Change != nil {
				x.Actions = append(x.Actions, "press", "set")
			}
			if n.Kind == ui.MultilineText {
				x.Actions = append(x.Actions, "set", "selection")
			}
			if n.Kind == ui.TextField && n.Change != nil && !n.Secure {
				x.Actions = append(x.Actions, "selection")
			}
			if (n.Kind == ui.TextField || n.Kind == ui.Slider || n.Kind == ui.NumberField) && n.Change != nil {
				x.Actions = append(x.Actions, "set")
			}
			if (n.Kind == ui.Slider || n.Kind == ui.NumberField) && n.Change != nil {
				x.Actions = append(x.Actions, "increment", "decrement")
			}
			if (n.Kind == ui.Selector || n.Kind == ui.Tabs || n.Kind == ui.RadioGroup || n.Kind == ui.SegmentedControl || n.Kind == ui.ListBox) && (n.Change != nil || n.Multiple && n.SelectionChange != nil) {
				x.Actions = append(x.Actions, "set")
			}
			if n.Kind == ui.Dialog && n.Dismiss != nil {
				x.Actions = append(x.Actions, "cancel")
			}
		}
		if n.Kind == ui.Button && r.control != nil {
			x.activationGuard = placementActivationGuard(r)
			if n.ShortcutRecorder != nil {
				x.activationGuard = shortcutAccessibilityGuard(r)
			}
		}
		protectSecureAccessibility(n, &x)
		nodes = append(nodes, x)
		if len(x.Actions) > 0 {
			allowed := append([]string(nil), x.Actions...)
			handlers[x.ID] = func(action accessibilityAction) bool {
				if (n.Kind == ui.ListBox || n.ShortcutRecorder != nil) && (x.activationGuard == nil || !x.activationGuard()) {
					return false
				}
				if !containsAccessibilityAction(allowed, action.Kind) {
					return false
				}
				if action.Kind == "focus" {
					accessibilityFocus(b, x.ID)
					return true
				}
				accepted := applyAccessibilityAction(r, action)
				if accepted && action.Kind == "press" && n.ShortcutRecorder != nil && b != nil {
					accessibilityFocus(b, x.ID)
				}
				return accepted
			}
		}
		if n.Kind == ui.Disclosure && r.control != nil && r.control.disclosure != nil {
			nodes[len(nodes)-1].Focusable = false
			nodes[len(nodes)-1].Focused = false
			nodes[len(nodes)-1].Actions = []string{}
			delete(handlers, x.ID)
			heading, handler := disclosureAccessibility(r, x.ID, b, disabled)
			nodes = append(nodes, heading)
			if handler != nil {
				handlers[heading.ID] = handler
			}
		}
		if n.Kind == ui.Canvas2D && len(n.CanvasObjects) > 0 && r.control != nil && r.control.placement != nil {
			extra, actions := canvasAccessibility(r, x.ID, b, disabled)
			nodes = append(nodes, extra...)
			for id, action := range actions {
				handlers[id] = action
			}
		}
		if n.Kind == ui.DataTree && r.control != nil && r.control.outline != nil {
			extra, actions := treeAccessibility(r, x.ID, b, disabled)
			nodes = append(nodes, extra...)
			for id, action := range actions {
				handlers[id] = action
			}
		}
		if n.Kind == ui.DataGrid && r.control != nil && r.control.grid != nil {
			extra, actions := gridAccessibility(r, x.ID, b, disabled)
			nodes = append(nodes, extra...)
			for id, action := range actions {
				handlers[id] = action
			}
		}
		if n.Kind == ui.Extension {
			extra, actions := extensionAccessibility(r, b, disabled)
			nodes = append(nodes, extra...)
			for id, action := range actions {
				handlers[id] = action
			}
		}
		if r.control != nil && r.control.choice != nil {
			extra, actions := choiceAccessibility(r, x.ID, b, disabled)
			nodes = append(nodes, extra...)
			for id, handler := range actions {
				handlers[id] = handler
			}
		}
		if r.control != nil && r.control.menu != nil {
			extra, actions := menuAccessibility(r, x.ID, b)
			nodes = append(nodes, extra...)
			for id, handler := range actions {
				handlers[id] = handler
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
				if n.Kind == ui.Tabs && o.Closable {
					close, handler := tabCloseAccessibility(r, o, id, disabled)
					nodes = append(nodes, close)
					if handler != nil {
						handlers[close.ID] = handler
					}
				}
			}
		}
		if edit := gridEditingRecord(r); edit != nil {
			visit(edit, x.ID, disabled)
		}
		for i, c := range r.children {
			if n.Kind == ui.SplitPane && i == 1 && r.control != nil && r.control.split != nil {
				handle, action := splitAccessibility(r, x.ID, b, disabled)
				nodes = append(nodes, handle)
				if action != nil {
					handlers[handle.ID] = action
				}
			}
			if n.Kind == ui.Disclosure && !n.Checked || n.Kind == ui.Tabs && c.node.ID != n.SelectedID {
				continue
			}
			visit(c, x.ID, disabled)
		}
		if n.Kind == ui.Canvas2D && len(n.CanvasObjects) > 0 && r.control != nil && r.control.placement != nil {
			stampCanvasAccessibility(r, nodes[start:], handlers)
		}
		if n.Kind == ui.DataTree && r.control != nil && r.control.outline != nil {
			stampTreeAccessibility(r, nodes[start:])
			guardTreeAccessibility(r.control.outline, nodes[start:], handlers)
		}
		if n.Kind == ui.DataGrid && r.control != nil && r.control.grid != nil {
			stampGridAccessibility(r, nodes[start:])
			guardGridAccessibility(r.control.grid, nodes[start:], handlers)
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
	if toolbarOverflowHidden(r) || n.Disabled || (r.control != nil && (r.control.disabled || r.control.retired)) {
		return false
	}
	if n.ContextMenu != nil && action.Kind == "show-menu" {
		if r.control == nil || r.control.menu == nil || !menuAvailable(r) {
			return false
		}
		r.control.menu.position = nil
		r.control.menu.request = true
		_ = r.control.a.Invalidate(r.control.window)
		return true
	}
	if n.Kind == ui.MenuButton {
		if !menuAvailable(r) {
			return false
		}
		switch action.Kind {
		case "press", "expand":
			r.control.menu.request = true
		case "cancel", "collapse":
			if !menuOpen(r) {
				return false
			}
			requestMenuClose(r)
		default:
			return false
		}
		_ = r.control.a.Invalidate(r.control.window)
		return true
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

	if n.Kind == ui.ListBox && n.Multiple && action.Kind == "set" {
		if n.SelectionChange == nil {
			return false
		}
		for _, o := range n.Options {
			if o.ID == action.Value && !o.Disabled {
				n.SelectionChange(ui.ListSelectionEvent{ID: n.ID, TargetID: o.ID, IDs: []string{o.ID}})
				requestControlRestore(r)
				return true
			}
		}
		return false
	}
	if n.Change == nil {
		return false
	}
	change := ui.ChangeEvent{ID: n.ID, Kind: n.Kind}
	switch n.Kind {
	case ui.Disclosure:
		expanded := n.Checked
		switch action.Kind {
		case "press":
			expanded = !expanded
		case "expand":
			expanded = true
		case "collapse":
			expanded = false
		default:
			return false
		}
		if expanded == n.Checked {
			return true
		}
		change.Checked = expanded
	case ui.Checkbox, ui.Switch, ui.ToggleButton:
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
		if b := r.control.gridEdit; b != nil {
			if !b.writable() || !validGridDraft(action.Value, b.maximum()) {
				return false
			}
			if err := b.grid.model.SetEditText(action.Value); err != nil {
				return false
			}
			state, err := b.grid.model.State()
			if err != nil {
				return false
			}
			if cb := b.grid.r.node.GridChange; cb != nil {
				cb(ui.GridEvent{ID: b.grid.r.node.ID, Reason: ui.GridEditing, State: state})
			}
			_ = b.grid.owner.Invalidate(b.grid.window)
			return true
		}
		change.Text = action.Value
	case ui.RadioGroup, ui.SegmentedControl, ui.ListBox:
		if action.Kind != "set" || !choiceOptionEnabled(n, action.Value) {
			return false
		}
		if n.Kind != ui.ListBox && n.SelectedID == action.Value {
			return true
		}
		change.SelectedID = action.Value
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
	case ui.Slider, ui.NumberField, ui.SplitPane:
		bounds := n.Numeric
		if n.Kind == ui.SplitPane {
			bounds = ui.SplitRange(bounds)
		}
		value := float64(n.Number)
		switch action.Kind {
		case "increment":
			value += float64(bounds.Step)
		case "decrement":
			value -= float64(bounds.Step)
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
		number, ok := ui.ClampNumeric(float32(value), bounds)
		if !ok {
			return false
		}
		change.Number = number
		if n.Kind == ui.SplitPane && number == n.Number {
			return true
		}
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
			if w.root == nil || w.closing.Load() || w.window.Ref().Context().Err() != nil {
				return
			}
			if chord == "escape" {
				if r := focusedManagedTextRecord(w.root); r != nil && r.control.gridEdit != nil {
					field := r.widget.(*core.TextField)
					if _, _, marked := field.NativeCompositionRange(); marked {
						state := field.EditingState()
						applyGridFieldNative(r, state.Text, state.Anchor, state.Focus, -1, -1, "cancel")
						e.SetHandled()
						requestTextPublicationScoped(w.window.Ref())
						return
					}
				}
				if r := focusedManagedTextRecord(w.root); r != nil && r.control.editor != nil && r.node.Editor.Snapshot().Composing {
					if err := r.node.Editor.CancelComposition(); err != nil {
						editorError(r, err)
					} else {
						syncEditor(r)
					}
					e.SetHandled()
					requestTextPublicationScoped(w.window.Ref())
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
	order := ui.FocusOrder(projection)
	if r.node.Sheet && len(order) > 0 && order[0] == accessibilityManagedID(ui.SheetCancelID(r.node.ID)) {
		if target, ok := targets[accessibilityManagedID(r.node.DialogDefaultID)]; ok && !accessibilityTreeNil(target.widget) && !target.widget.AsWidget().IsDisabled() {
			return target.widget
		}
	}
	for _, id := range order {
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
	finishWindowFocusReveal(h.window, h.Scene)
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
	bounds := nativeAccessibilityBounds(w.root)
	textBounds := accessibilityTextGeometry(w.root)
	w.rootMu.Unlock()
	b.publishGeometry(bounds)
	b.publishTextGeometry(textBounds)
}
func accessibilityBounds(root *record) map[string][4]float32 {
	return accessibilityBoundsIncludingScoped(root, true)
}

// Scoped grid and outline frames are published atomically with their semantics. Cocoa
// deliberately ignores these IDs in the unversioned geometry channel, so do
// not construct or encode a second copy of them on every render. Inspection
// keeps the full projection above, including private grid editor bounds.
func nativeAccessibilityBounds(root *record) map[string][4]float32 {
	return accessibilityBoundsIncludingScoped(root, false)
}

func accessibilityBoundsIncludingScoped(root *record, includeScoped bool) map[string][4]float32 {
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
		if !includeScoped && r.control != nil && (r.node.Kind == ui.DataGrid && r.control.grid != nil || r.node.Kind == ui.DataTree && r.control.outline != nil || r.node.Kind == ui.Canvas2D && len(r.node.CanvasObjects) > 0 && r.control.placement != nil) {
			return // The bound host and its materialized descendants share a scope.
		}
		widget := r.widget
		if r.control != nil && r.control.dialog != nil {
			widget = r.control.dialog.host
		}
		bounds(accessibilityManagedID(r.node.ID), widget)
		menuPopupBounds(r, bounds)
		if r.node.Kind == ui.Tabs && r.control != nil && r.control.tabs != nil {
			for _, o := range r.node.Options {
				bounds(accessibilityOptionID(r.node.ID, o.ID), r.control.tabs.buttons[o.ID])
				if o.Closable {
					if close := tabCloseButton(r.control.tabs.bindings[o.ID]); close != nil {
						bounds(tabCloseID(r.node.ID, o.ID), close)
					}
				}
			}
		}
		if r.control != nil && r.control.choice != nil {
			for _, option := range r.node.Options {
				bounds(accessibilityOptionID(r.node.ID, option.ID), r.control.choice.buttons[option.ID])
			}
		}
		if r.control != nil && r.control.split != nil {
			bounds(accessibilitySplitHandleID(r.node.ID), r.control.split.handle)
		}
		if r.control != nil && r.control.disclosure != nil {
			bounds(accessibilityDisclosureID(r.node.ID), r.control.disclosure.header)
		}
		if r.node.Kind == ui.Selector {
			for _, o := range r.node.Options {
				bounds(accessibilityOptionID(r.node.ID, o.ID), r.widget)
			}
		}
		extensionAccessibilityBounds(r, out, bounds, includeScoped)
		gridAccessibilityBounds(r, out)
		treeAccessibilityBounds(r, out)
		canvasAccessibilityBounds(r, out)
		if edit := gridEditingRecord(r); edit != nil {
			visit(edit)
		}
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
	// Grid-prefixed metadata covers managed tables and outlines. Pending records
	// preserve last-completed data with actions revoked; target and completed
	// generations remain distinct.
	GridScope            string     `json:"grid_scope"`
	GridIncarnation      uint64     `json:"grid_incarnation"`
	GridPending          bool       `json:"grid_pending"`
	GridTargetGeneration uint64     `json:"grid_target_generation"`
	GridGeneration       uint64     `json:"grid_generation"`
	ObjectIdentity       uint64     `json:"object_identity"`
	ID                   string     `json:"id"`
	Role                 string     `json:"role"`
	Name                 string     `json:"name"`
	Value                any        `json:"value"`
	Enabled              bool       `json:"enabled"`
	RowCount             int        `json:"row_count"`
	ColumnCount          int        `json:"column_count"`
	RowIndex             [2]int     `json:"row_index"`
	ColumnIndex          [2]int     `json:"column_index"`
	TextSelection        [2]int     `json:"text_selection"`
	InsertionLine        int        `json:"insertion_line"`
	Frame                [4]float64 `json:"frame"`
	Selected             bool       `json:"selected"`
	Expanded             bool       `json:"expanded"`
	Parent               string     `json:"parent"`
	DomainID             string     `json:"domain_id"`
	HitTestID            string     `json:"hit_test_id"`
}

func lookupAccessibility(window string) (*accessibilityBridge, error) {
	accessibilityRegistry.Lock()
	b := accessibilityRegistry.windows[window]
	accessibilityRegistry.Unlock()
	if b == nil || b.closed.Load() {
		return nil, fmt.Errorf("cogent: native accessibility window %q unavailable", window)
	}
	b.mu.Lock()
	ref := b.windowRef
	b.mu.Unlock()
	if _, err := scopedWindowContext(b.owner, ref); err != nil {
		return nil, fmt.Errorf("cogent: native accessibility window %q unavailable: %w", window, err)
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
	b.mu.Lock()
	sameScope := b.windowRef == r.scope
	b.mu.Unlock()
	if !sameScope {
		return
	}
	bounds := nativeAccessibilityBounds(r)
	if prefix, _ := h.Property("vellum.accessibility-namespace").(string); prefix != "" {
		qualified := make(map[string][4]float32, len(bounds))
		for id, rect := range bounds {
			qualified[prefix+id] = rect
		}
		bounds = qualified
	}
	b.publishGeometry(bounds)
	textBounds := accessibilityTextGeometry(r)
	if prefix, _ := h.Property("vellum.accessibility-namespace").(string); prefix != "" {
		qualified := map[string]any{}
		for id, value := range textBounds {
			qualified[prefix+id] = value
		}
		textBounds = qualified
	}
	b.publishTextGeometry(textBounds)
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
	n.Disabled = n.Disabled || toolbarOverflowHidden(root) || n.Kind == ui.MenuButton && !menuAvailable(root)
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

// All ordinary semantic and wire outputs share this final projection boundary.
// Never derive the label from a protected value; explicit app labels remain labels.
func protectSecureAccessibility(n ui.Node, x *accessibilityNode) {
	if !n.Secure {
		return
	}
	x.Secure = true
	x.Value = ""
	x.TextClient = 0
	x.TextSelection = nil
	x.TextFocus = 0
	if n.Accessibility.Name == "" {
		x.Name = "Password"
	}
	var actions []string
	if x.Actions != nil {
		actions = make([]string, 0, len(x.Actions))
	}
	for _, a := range x.Actions {
		if a != "selection" {
			actions = append(actions, a)
		}
	}
	x.Actions = actions
}
