package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/tree"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/ui"
)

func mount(parent tree.Node, n ui.Node, a *app.Application, name string, themes ...ui.Theme) (*record, error) {
	ref, err := fixtureMountScope(a, name)
	if err != nil {
		return nil, err
	}
	bindWindowScope(parent, a, ref)
	return mountScoped(parent, n, a, ref, themes...)
}
func mountShallow(parent tree.Node, n ui.Node, a *app.Application, name string, themes ...ui.Theme) (*record, error) {
	ref, err := fixtureMountScope(a, name)
	if err != nil {
		return nil, err
	}
	bindWindowScope(parent, a, ref)
	return mountShallowScoped(parent, n, a, ref, themes...)
}
func reconcile(parent core.Widget, previous *record, n ui.Node, a *app.Application, name string, themes ...ui.Theme) (*record, error) {
	if a == nil {
		return reconcileScoped(parent, previous, n, a, app.WindowRef{}, themes...)
	}
	if previous != nil {
		return reconcileScoped(parent, previous, n, a, previous.scope, themes...)
	}
	ref, err := fixtureMountScope(a, name)
	if err != nil {
		return previous, err
	}
	bindWindowScope(parent, a, ref)
	return reconcileScoped(parent, previous, n, a, ref, themes...)
}
func bindCommandShortcuts(scene *core.Scene, owner *app.Application, name string, registry *command.Registry, onError func(error)) {
	ref, err := owner.WindowRef(name)
	if err != nil {
		panic(err)
	}
	bindCommandShortcutsScoped(scene, owner, ref, registry, onError)
}
func (b *accessibilityBridge) bind(name string) {
	ref, err := b.owner.WindowRef(name)
	if err != nil {
		panic(err)
	}
	b.bindScoped(ref)
}

func fixtureWindowRef(a *app.Application, name string) app.WindowRef {
	ref, err := a.WindowRef(name)
	if err != nil {
		panic(err)
	}
	return ref
}
func fixtureWindow(a *app.Application, name string) app.Window {
	for _, w := range a.Windows() {
		if w.ID == name {
			return w
		}
	}
	panic(name)
}

// Legacy pure-widget fixtures declare a live native scope at first mount.
// New lifecycle regressions use explicit OpenWindow receipts and Scoped APIs.
func fixtureMountScope(a *app.Application, name string) (app.WindowRef, error) {
	if a == nil {
		return app.WindowRef{}, app.ErrInvalidWindowRef
	}
	ref, err := a.WindowRef(name)
	if errors.Is(err, app.ErrWindowClosed) && a.Context().Err() == nil {
		return a.OpenWindow(scopeDeclaration(name))
	}
	return ref, err
}

func requestTextPublication(name string) {
	if t := lookupTextWindow(name); t != nil {
		requestTextPublicationScoped(t.window.window.Ref())
	}
}
