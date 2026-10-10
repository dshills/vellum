package cogent

import (
	"fmt"
	"reflect"

	"cogentcore.org/core/core"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

// reconcile preserves compatible native widgets by ID throughout a managed tree.
// Validation and allocation finish before any retained node is moved or removed.
func reconcile(parent core.Widget, previous *record, n ui.Node, a *app.Application, window string, themes ...ui.Theme) (*record, error) {
	var theme ui.Theme
	if previous != nil {
		theme = previous.theme
	}
	if len(themes) > 0 {
		theme = themes[0]
	}
	if err := theme.Validate(); err != nil {
		return previous, err
	}
	theme = theme.Clone()
	var validateStyle func(ui.Node) error
	validateStyle = func(node ui.Node) error {
		if _, err := ui.ResolveStyle(node, theme); err != nil {
			return err
		}
		for _, child := range node.Children {
			if err := validateStyle(child); err != nil {
				return err
			}
		}
		return nil
	}
	if err := validateStyle(n); err != nil {
		return previous, err
	}
	if err := ui.Validate(n); err != nil {
		return previous, err
	}
	if err := validateExtensions(n); err != nil {
		return previous, err
	}
	old := map[string]*record{}
	var collect func(*record)
	collect = func(r *record) {
		if r == nil || tree.IsNil(r.widget) {
			return
		}
		old[r.node.ID] = r
		for _, child := range r.children {
			collect(child)
		}
	}
	collect(previous)
	type plan struct {
		r        *record
		node     ui.Node
		children []*plan
		parent   tree.Node
	}
	used := map[*record]bool{}
	var created []*record
	var prepare func(tree.Node, ui.Node) (*plan, error)
	prepare = func(p tree.Node, node ui.Node) (*plan, error) {
		r := old[node.ID]
		if r == nil || !compatible(r.node, node) {
			var err error
			r, err = mountShallow(p, node, a, window, theme)
			if err != nil {
				return nil, err
			}
			created = append(created, r)
		} else {
			used[r] = true
		}
		if err := prepareControl(r, node); err != nil {
			return nil, err
		}
		next := &plan{r: r, node: node, parent: p}
		for _, child := range node.Children {
			entry, err := prepare(childHost(r, child.ID), child)
			if err != nil {
				return nil, err
			}
			next.children = append(next.children, entry)
		}
		return next, nil
	}
	next, err := prepare(parent, n)
	if err != nil {
		for _, r := range created {
			if !tree.IsNil(r.widget) {
				disposeControl(r)
				r.widget.AsTree().Delete()
			}
		}
		return previous, err
	}
	changed := previous == nil || len(created) > 0
	for _, r := range old {
		if !used[r] {
			changed = true
		}
	}
	var checkOrder func(*plan)
	checkOrder = func(p *plan) {
		if p.r.widget.AsTree().Parent != p.parent || len(p.r.children) != len(p.children) {
			changed = true
		}
		for i, child := range p.children {
			if i < len(p.r.children) && p.r.children[i] != child.r {
				changed = true
			}
			checkOrder(child)
		}
	}
	checkOrder(next)
	// Detach every retained node whose parent changes before making new links.
	// This also handles an ancestor moving beneath its former descendant safely.
	var parking *core.Frame
	var park func(*plan)
	park = func(p *plan) {
		if used[p.r] && p.r.widget.AsTree().Parent != p.parent {
			if parking == nil {
				parking = core.NewFrame(parent)
			}
			tree.MoveToParent(p.r.widget, parking)
		}
		for _, child := range p.children {
			park(child)
		}
	}
	park(next)
	if parking != nil {
		defer parking.Delete()
	}
	var attach func(*plan)
	attach = func(p *plan) {
		if !reflect.DeepEqual(p.r.theme, theme) {
			p.r.theme = theme
			if ext, ok := p.node.Custom.(ThemeExtension); p.node.Kind == ui.Extension && ok {
				ext.SetTheme(theme.Clone())
			}
			p.r.widget.AsWidget().Update()
			changed = true
		}
		if p.r.widget.AsTree().Parent != p.parent {
			tree.MoveToParent(p.r.widget, p.parent)
			p.r.widget.AsWidget().SetScene(p.parent.(core.Widget).AsWidget().Scene)
		}
		p.r.children = make([]*record, len(p.children))
		for i, child := range p.children {
			attach(child)
			p.r.children[i] = child.r
		}
	}
	attach(next)
	// Dispose external modal scenes before deleting any ancestor; map order is
	// unspecified, and an ancestor Delete would otherwise hide live descendants.
	for _, r := range old {
		if !used[r] && !tree.IsNil(r.widget) {
			disposeControl(r)
		}
	}
	for _, r := range old {
		if !used[r] && !tree.IsNil(r.widget) {
			r.widget.AsTree().Delete()
		}
	}
	var order func(*plan)
	order = func(p *plan) {
		// ui.Validate rejects children on leaf/extension kinds. These children
		// therefore occupy the entire managed container; no extension subtree
		// is reordered. A same-parent reorder uses the exported native slice
		// to avoid calling OnAdd and resetting native state.
		if p.node.Kind == ui.Column || p.node.Kind == ui.Row || p.node.Kind == ui.Scroll || p.node.Kind == ui.Dialog {
			children := make([]tree.Node, len(p.children))
			for i, child := range p.children {
				children[i] = child.r.widget
			}
			childHost(p.r, "").AsTree().Children = children
		}
		finishControl(p.r, p.node)
		for _, child := range p.children {
			order(child)
		}
	}
	order(next)
	// The parent is dedicated to this managed root (not an extension's own tree).
	if parking != nil {
		parking.Delete()
	}
	parent.AsTree().Children = []tree.Node{next.r.widget}
	if err := syncRecord(next.r, n); err != nil {
		return next.r, err
	}
	if changed {
		parent.AsWidget().Update()
	}
	return next.r, nil
}

func compatible(old, next ui.Node) bool {
	if old.ID != next.ID || old.Kind != next.Kind {
		return false
	}
	if next.Kind != ui.Extension {
		return true
	}
	if reflect.TypeOf(old.Custom) != reflect.TypeOf(next.Custom) {
		return false
	}
	a, b := reflect.ValueOf(old.Custom), reflect.ValueOf(next.Custom)
	return !a.IsValid() || a.Kind() != reflect.Pointer || a.Pointer() == b.Pointer()
}

func validateExtensions(n ui.Node) error {
	if n.Kind == ui.Extension {
		v := reflect.ValueOf(n.Custom)
		if _, ok := n.Custom.(Extension); !ok || !v.IsValid() || (v.Kind() == reflect.Pointer && v.IsNil()) {
			return fmt.Errorf("cogent: unsupported or nil extension %q", n.ID)
		}
	}
	for _, child := range n.Children {
		if err := validateExtensions(child); err != nil {
			return err
		}
	}
	return nil
}
