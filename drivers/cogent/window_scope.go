package cogent

import (
	"cogentcore.org/core/tree"
	"context"
	"github.com/dshills/vellum/app"
)

const windowScopeProperty = "vellum.window-scope"

type mountedWindowScope struct {
	owner *app.Application
	ref   app.WindowRef
}

// WindowScope returns the originating lifetime of a managed native subtree.
// Extensions should capture it during Mount and retain it for delayed work.
// A logical name alone cannot identify that opening after it has closed.
func WindowScope(parent tree.Node) (*app.Application, app.WindowRef, bool) {
	for p := parent; !tree.IsNil(p); p = p.AsTree().Parent {
		if s, ok := p.AsTree().Property(windowScopeProperty).(*mountedWindowScope); ok && s != nil {
			return s.owner, s.ref, true
		}
	}
	return nil, app.WindowRef{}, false
}
func bindWindowScope(node tree.Node, owner *app.Application, ref app.WindowRef) {
	node.AsTree().SetProperty(windowScopeProperty, &mountedWindowScope{owner: owner, ref: ref})
}
func scopedWindowContext(owner *app.Application, ref app.WindowRef) (context.Context, error) {
	if owner == nil || ref == (app.WindowRef{}) {
		return nil, app.ErrInvalidWindowRef
	}
	current, err := owner.WindowRef(ref.ID())
	if err != nil {
		return nil, err
	}
	if current != ref || ref.Context().Err() != nil {
		return nil, app.ErrWindowClosed
	}
	return ref.Context(), nil
}

// Direct extension mounts may capture a current opening at setup. Managed
// mounts inherit their parent's original reference and never resolve it anew.
func captureMountScope(parent tree.Node, owner *app.Application, name string) (app.WindowRef, error) {
	if inheritedOwner, ref, ok := WindowScope(parent); ok {
		if owner != inheritedOwner || name != ref.ID() {
			return app.WindowRef{}, app.ErrInvalidWindowRef
		}
		_, err := scopedWindowContext(owner, ref)
		return ref, err
	}
	if owner == nil {
		return app.WindowRef{}, app.ErrInvalidWindowRef
	}
	return owner.WindowRef(name)
}

// withWindowOwner rechecks the original lifetime after acquiring the owner.
// Cleanup must use Do directly: retirement must never skip resource release.
func withWindowOwner(owner *app.Application, ref app.WindowRef, fn func()) error {
	admitted := false
	err := owner.Do(func() {
		if ref.Context().Err() != nil {
			return
		}
		admitted = true
		fn()
	})
	if err != nil {
		return err
	}
	if !admitted {
		return app.ErrWindowClosed
	}
	return nil
}
