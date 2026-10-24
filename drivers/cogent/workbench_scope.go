package cogent

import (
	"slices"
	"sync/atomic"

	"github.com/dshills/vellum/ui"
)

// Immutable owner-published input context. Revocation and dialog closure are
// visible to already queued gestures without reading native ancestors at drain.
type workbenchInputScope struct {
	visible, disabled bool
	scopes            []*tabCloseScope
	revoked           atomic.Bool
}

func (s *workbenchInputScope) allowed() bool {
	if s == nil || s.revoked.Load() || !s.visible || s.disabled {
		return false
	}
	for _, scope := range s.scopes {
		if scope == nil || scope.closed.Load() {
			return false
		}
	}
	return true
}
func (w *Workbench) setInputScope(visible, disabled bool, scopes []*tabCloseScope) {
	old := w.inputScope.Load()
	if old != nil && !old.revoked.Load() && old.visible == visible && old.disabled == disabled && slices.Equal(old.scopes, scopes) {
		return
	}
	if old != nil {
		old.revoked.Store(true)
	}
	w.inputScope.Store(&workbenchInputScope{visible: visible, disabled: disabled, scopes: slices.Clone(scopes)})
}
func (w *Workbench) queueInput(change func() error) {
	origin := w.inputScope.Load()
	if !origin.allowed() {
		return
	}
	w.queue(func() error {
		if !origin.allowed() || !w.inputScope.Load().allowed() {
			return nil
		}
		return change()
	})
}

// Include hidden retained pages when allocating identities, revoking input or
// retiring a window. Visible focus/wake projection has a separate traversal.
func retainedWorkbenchRecords(r *record, visit func(string, *record)) {
	if r == nil || r.node.Kind != ui.Extension {
		return
	}
	w, ok := r.node.Custom.(*Workbench)
	if !ok || w == nil {
		return
	}
	for _, panel := range w.Panels {
		if child := w.records[panel.ID]; child != nil {
			visit(panel.ID, child)
		}
	}
}

func bindWorkbenchInputScopes(root *record) {
	allowed := map[string]bool{}
	var project func(ui.Node, bool)
	project = func(n ui.Node, disabled bool) {
		disabled = disabled || n.Disabled
		if disabled || n.Kind == ui.Dialog && !n.Open {
			return
		}
		allowed[n.ID] = true
		for _, child := range n.Children {
			if n.Kind == ui.Disclosure && !n.Checked || n.Kind == ui.Tabs && child.ID != n.SelectedID {
				continue
			}
			project(child, disabled)
		}
	}
	project(ui.FocusScope(accessibilityFocusProjection(root, "")), false)
	var visit func(*record, string)
	visit = func(r *record, prefix string) {
		if r == nil {
			return
		}
		if r.node.Kind == ui.Button || r.node.Kind == ui.MenuButton || r.node.Kind == ui.Checkbox {
			next := uint32(2)
			if allowed[prefix+accessibilityManagedID(r.node.ID)] {
				next = 1
			}
			if r.focusVisibility.Swap(next) != next {
				r.widget.AsWidget().Restyle()
			}
		}
		if r.control != nil && r.control.placement != nil {
			visible, scopes := tabCloseInheritedContext(r)
			setPlacementScope(r.control.placement, visible && allowed[prefix+accessibilityManagedID(r.node.ID)], controlDisabled(r), scopes)
		}
		if r.control != nil && r.node.Kind == ui.ListBox {
			context := r.control.choice.listContext
			setListBoxScope(r, context.visible && allowed[prefix+accessibilityManagedID(r.node.ID)], controlDisabled(r), context.scopes)
		}
		if dock, ok := r.node.Custom.(*Workbench); r.node.Kind == ui.Extension && ok && dock != nil {
			visible, scopes := tabCloseInheritedContext(r)
			dock.setInputScope(visible && allowed[prefix+accessibilityManagedID(r.node.ID)], controlDisabled(r), scopes)
			retainedWorkbenchRecords(r, func(panel string, child *record) {
				visit(child, prefix+accessibilityWorkbenchNamespace(r.node.ID, panel))
			})
		}
		for _, child := range r.children {
			visit(child, prefix)
		}
	}
	visit(root, "")
}
