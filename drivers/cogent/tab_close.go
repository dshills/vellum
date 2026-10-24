package cogent

import (
	"html"
	"sync/atomic"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/ui"
)

type tabCloseScope struct{ closed atomic.Bool }
type tabCloseEpoch struct {
	marker byte
	scopes []*tabCloseScope
}

func tabCloseEpochLive(epoch *tabCloseEpoch) bool {
	if epoch == nil {
		return false
	}
	for _, scope := range epoch.scopes {
		if scope == nil || scope.closed.Load() {
			return false
		}
	}
	return true
}
func tabCloseSameScopes(a, b []*tabCloseScope) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

type tabBinding struct {
	button               *core.Tab
	closeEpoch           atomic.Pointer[tabCloseEpoch]
	selectEpoch          atomic.Pointer[tabCloseEpoch]
	selectWritable       bool
	closeWidget          atomic.Pointer[core.Button]
	presented            bool
	presentationType     core.TabTypes
	presentationLabel    string
	presentationClosable bool
	manualIncarnation    uint64 // manual originating-owner inspection only
}

func bindTab(r *record, id string, button *core.Tab) *tabBinding {
	binding := &tabBinding{button: button}
	r.control.tabs.bindings[id] = binding
	// Never restore nil after retirement: a late private close gesture must not
	// fall through to the backend's destructive unmanaged deletion path.
	button.CloseRequest = func(origin *core.Button) {
		// The private close click clears focus in Core. Restore its live header
		// here, on the originating native event, without selecting that page.
		// Owner delivery (including a peer-window drain) must never touch focus.
		if origin != nil && binding.closeWidget.Load() == origin && tabCloseEpochLive(binding.closeEpoch.Load()) && tabCloseEnabled(r, id, binding) {
			button.SetFocusQuiet()
		}
		queueTabCloseFrom(r, id, binding, origin)
	}
	button.OnFirst(events.KeyChord, func(e events.Event) {
		if e.IsHandled() || e.Modifiers() != 0 || !tabCloseEpochLive(binding.closeEpoch.Load()) {
			return
		}
		if e.KeyCode() != key.CodeBackspace && e.KeyCode() != key.CodeDelete {
			return
		}
		e.SetHandled()
		queueTabClose(r, id, binding)
	})
	return binding
}
func tabCloseEnabled(r *record, id string, binding *tabBinding) bool {
	if r == nil || r.control == nil || r.control.retired || r.control.disabled || r.control.tabs == nil || !r.control.tabs.visible || r.control.tabs.bindings[id] != binding || r.node.TabClose == nil {
		return false
	}
	for _, option := range r.node.Options {
		if option.ID == id {
			return option.Closable && !option.Disabled
		}
	}
	return false
}
func queueTabClose(r *record, id string, binding *tabBinding) {
	queueTabCloseFrom(r, id, binding, binding.closeWidget.Load())
}

func queueTabCloseFrom(r *record, id string, binding *tabBinding, origin *core.Button) {
	epoch := binding.closeEpoch.Load()
	if !tabCloseEpochLive(epoch) || origin == nil || binding.closeWidget.Load() != origin {
		return
	}
	postControl(r, func() {
		if !tabCloseEpochLive(epoch) || binding.closeEpoch.Load() != epoch || binding.closeWidget.Load() != origin || !tabCloseEnabled(r, id, binding) {
			return
		}
		r.node.TabClose(ui.TabCloseEvent{ID: r.node.ID, PageID: id})
		_ = r.control.a.Invalidate(r.control.window)
	})
}
func tabCloseButton(binding *tabBinding) *core.Button {
	if binding == nil || tree.IsNil(binding.button) {
		return nil
	}
	button, _ := binding.button.AsTree().ChildByName("close").(*core.Button)
	return button
}
func syncTabsPresentation(r *record) {
	typ := core.StandardTabs
	for _, o := range r.node.Options {
		if o.Closable {
			typ = core.FunctionalTabs
			break
		}
	}
	w := r.widget.(*core.Tabs)
	w.SetType(typ)
	if len(r.node.Options) > 0 {
		button := r.control.tabs.buttons[r.node.Options[0].ID]
		if bar, ok := button.AsTree().Parent.(core.Widget); ok && bar.AsTree().Property("vellum.settings.tabbar") == nil {
			bar.AsTree().SetProperty("vellum.settings.tabbar", true)
			bar.AsWidget().FinalStyler(func(s *styles.Style) {
				if r.node.TabsHideHeader {
					s.Display = styles.DisplayNone
				}
			})
		}
	}

	for _, o := range r.node.Options {
		button := r.control.tabs.buttons[o.ID]
		binding := r.control.tabs.bindings[o.ID]
		label := html.EscapeString(o.Label)
		button.SetType(typ).SetText(label)
		icon := icons.None
		if o.Closable {
			icon = icons.Close
		}
		button.SetCloseIcon(icon)
		// Rebuild only the native tab's private artwork, preserving page widgets.
		if !binding.presented || binding.presentationType != typ || binding.presentationLabel != label || binding.presentationClosable != o.Closable {
			button.UpdateWidget()
			binding.presented = true
			binding.presentationType = typ
			binding.presentationLabel = label
			binding.presentationClosable = o.Closable
		}
		close := tabCloseButton(binding)
		if binding.closeWidget.Load() != close {
			binding.closeEpoch.Store(nil)
			binding.closeWidget.Store(close)
			binding.manualIncarnation = 0
		}
		if close != nil {
			close.Tooltip = html.EscapeString("Close " + o.Label)
		}
	}
}

// Reconciliation publishes owner-held visibility, so a callback drained by a
// different window never walks originating native widget parents. Hiding then
// showing a tab strip retires its previous close capability even if IDs recur.
// syncTabVisibility reuses the reconciler's existing ordered traversal. It
// publishes only owner values/atomic capabilities, never a second tree walk.
func syncTabVisibility(r *record, visible bool, scopes []*tabCloseScope) {
	syncListBoxVisibility(r, visible, scopes)
	if r.control == nil || r.control.tabs == nil {
		return
	}
	r.control.tabs.visible = visible
	for id, binding := range r.control.tabs.bindings {
		writable := r.node.Change != nil
		if binding.selectWritable != writable {
			binding.selectEpoch.Store(nil)
			binding.selectWritable = writable
		}
		if !tabSelectionEnabled(r, id, binding) {
			binding.selectEpoch.Store(nil)
		} else if previous := binding.selectEpoch.Load(); previous == nil || !tabCloseSameScopes(previous.scopes, scopes) {
			binding.selectEpoch.Store(&tabCloseEpoch{scopes: append([]*tabCloseScope(nil), scopes...)})
		}
		enabled := tabCloseEnabled(r, id, binding) && binding.closeWidget.Load() != nil
		if !enabled {
			binding.closeEpoch.Store(nil)
		} else if previous := binding.closeEpoch.Load(); previous == nil || !tabCloseSameScopes(previous.scopes, scopes) {
			binding.closeEpoch.Store(&tabCloseEpoch{scopes: append([]*tabCloseScope(nil), scopes...)})
		}
	}
	reconcileTabSelection(r)
}
func tabCloseID(tabs, page string) string {
	return "tab-close:" + accessibilityIDPart(tabs) + ":" + accessibilityIDPart(page)
}
func tabCloseAccessibility(r *record, option ui.Option, parent string, disabled bool) (accessibilityNode, accessibilityHandler) {
	binding := r.control.tabs.bindings[option.ID]
	n := accessibilityNode{ID: tabCloseID(r.node.ID, option.ID), Parent: parent, DomainID: option.ID, Role: "button", Name: "Close " + option.Label, Disabled: disabled || option.Disabled, Actions: []string{}}
	if n.Disabled || !tabCloseEnabled(r, option.ID, binding) {
		return n, nil
	}
	epoch := binding.closeEpoch.Load()
	n.Actions = []string{"press"}
	return n, func(action accessibilityAction) bool {
		if action.Kind != "press" || !tabCloseEpochLive(epoch) || binding.closeEpoch.Load() != epoch || !tabCloseEnabled(r, option.ID, binding) {
			return false
		}
		queueTabClose(r, option.ID, binding)
		return true
	}
}

const tabCloseContextProperty = "vellum.tab-close.context"

type tabCloseContext struct {
	visible bool
	scopes  []*tabCloseScope
}

// Independent managed roots inside an extension inherit its complete owner
// snapshot during originating reconciliation, never during queued delivery.
func tabCloseInheritedContext(r *record) (bool, []*tabCloseScope) {
	for parent := r.widget.AsTree().Parent; parent != nil; parent = parent.AsTree().Parent {
		if context, ok := parent.AsTree().Property(tabCloseContextProperty).(*tabCloseContext); ok && context != nil {
			return context.visible, context.scopes
		}
	}
	return true, nil
}
