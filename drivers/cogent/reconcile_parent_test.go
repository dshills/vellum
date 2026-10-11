package cogent

import (
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

type parentProbeExtension struct {
	parent        tree.Node
	widget        *core.Frame
	mounts, syncs int
}

func (e *parentProbeExtension) Mount(parent tree.Node) core.Widget {
	e.parent = parent
	e.mounts++
	e.widget = core.NewFrame(parent)
	return e.widget
}
func (e *parentProbeExtension) Sync(core.Widget) error { e.syncs++; return nil }

func TestKeyedEmbeddedParentIdentityDoesNotReparentOrUpdate(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	w := &nativeWindow{body: core.NewBody("embedded parent")}
	host := newAccessibilityHost(w)
	defer w.body.Scene.Destroy()
	probe := &parentProbeExtension{}
	n := ui.Node{ID: "root", Kind: ui.Extension, Custom: probe}
	r, err := reconcile(host, nil, n, a, "one")
	if err != nil {
		t.Fatal(err)
	}
	if probe.parent != host || r.widget.AsTree().Parent != host.AsTree().This {
		t.Fatal("mount argument or canonical native parent changed")
	}
	updates, stylePasses, additions := 0, 0, 0
	probe.widget.Updater(func() { updates++ })
	probe.widget.FinalStyler(func(*styles.Style) { stylePasses++ })
	host.AsTree().OnChildAdded = func(tree.Node) { additions++ }
	for range 3 {
		before := probe.syncs
		next, err := reconcile(host, r, n, a, "one")
		if err != nil {
			t.Fatal(err)
		}
		if next != r || probe.mounts != 1 || probe.syncs != before+1 || updates != 0 || stylePasses != 0 || additions != 0 {
			t.Fatalf("unchanged refresh rebuilt native parent: retained=%t mounts=%d syncs=%d/%d updates=%d styles=%d additions=%d", next == r, probe.mounts, probe.syncs, before+1, updates, stylePasses, additions)
		}
		if len(host.Children) != 1 || host.Children[0] != r.widget || r.widget.AsTree().Parent != host.AsTree().This {
			t.Fatal("unchanged native topology was not retained")
		}
	}
	// A real style change still updates the retained widget without reparenting.
	n.Style.Padding = ui.Some(ui.Insets{Top: 3})
	if _, err := reconcile(host, r, n, a, "one"); err != nil {
		t.Fatal(err)
	}
	if updates == 0 || stylePasses == 0 || additions != 0 || probe.widget.Styles.Padding.Top.Value != 3 {
		t.Fatal("style update was suppressed or treated as a parent move")
	}
	// A different embedded host remains a genuine topology transition.
	other := newAccessibilityHost(w)
	next, err := reconcile(other, r, n, a, "one")
	if err != nil {
		t.Fatal(err)
	}
	if next != r || probe.mounts != 1 || len(host.Children) != 0 || len(other.Children) != 1 || other.Children[0] != r.widget || r.widget.AsTree().Parent != other.AsTree().This {
		t.Fatal("real embedded-host move lost retained identity or native topology")
	}
}
