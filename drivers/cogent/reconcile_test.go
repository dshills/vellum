package cogent

import (
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/tree"
	"example.com/vellum/app"
	"example.com/vellum/ui"
)

type retainedExtension struct {
	mounts, syncs int
	widget        *core.Frame
}

func (e *retainedExtension) Mount(parent tree.Node) core.Widget {
	e.mounts++
	e.widget = core.NewFrame(parent)
	return e.widget
}
func (e *retainedExtension) Sync(core.Widget) error { e.syncs++; return nil }
func label(id string) ui.Node                       { return ui.Node{ID: id, Kind: ui.Label, Text: id} }
func frame(id string, children ...ui.Node) ui.Node {
	return ui.Node{ID: id, Kind: ui.Column, Children: children}
}
func findRecord(r *record, id string) *record {
	if r.node.ID == id {
		return r
	}
	for _, c := range r.children {
		if result := findRecord(c, id); result != nil {
			return result
		}
	}
	return nil
}

func TestKeyedInsertionReorderRemovalAndCallback(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	host := core.NewFrame(core.NewBody("headless"))
	clicks := 0
	button := ui.Node{ID: "button", Kind: ui.Button, Activate: func() { clicks++ }}
	ext := &retainedExtension{}
	n := frame("root", button, ui.Node{ID: "gpu", Kind: ui.Extension, Custom: ext}, label("remove"))
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	oldButton := findRecord(r, "button").widget
	oldButton.AsWidget().SetState(true, states.Focused)
	oldGPU := ext.widget
	oldGPU.SetProperty("scroll-position", 42)
	removed := findRecord(r, "remove").widget
	button.Activate = func() { clicks += 10 }
	next := frame("root", label("insert"), ui.Node{ID: "gpu", Kind: ui.Extension, Custom: ext}, button)
	r, err = reconcile(host, r, next, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if findRecord(r, "button").widget != oldButton || !oldButton.AsWidget().StateIs(states.Focused) {
		t.Fatal("focus or identity lost")
	}
	if ext.mounts != 1 || ext.widget != oldGPU || ext.widget.Property("scroll-position") != 42 {
		t.Fatal("retained native state lost")
	}
	if !tree.IsNil(removed) {
		t.Fatal("removed widget not destroyed")
	}
	for i, c := range r.children {
		if r.widget.AsTree().Children[i] != c.widget {
			t.Fatal("native child order differs")
		}
	}
	oldButton.AsWidget().Send(events.Click)
	if clicks != 10 {
		t.Fatalf("stale callback: %d", clicks)
	}
	if len(host.Children) != 1 {
		t.Fatal("host leaked parking or stale root")
	}
}

func TestKeyedReparentAndAncestorInversion(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	host := core.NewFrame(core.NewBody("headless"))
	r, err := reconcile(host, nil, frame("a", frame("b", label("keep")), frame("c")), a, "main")
	if err != nil {
		t.Fatal(err)
	}
	aWidget, bWidget, keepWidget := r.widget, findRecord(r, "b").widget, findRecord(r, "keep").widget
	r, err = reconcile(host, r, frame("b", frame("a", frame("c", label("keep")))), a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if r.widget != bWidget || findRecord(r, "a").widget != aWidget || findRecord(r, "keep").widget != keepWidget {
		t.Fatal("cross-parent identities lost")
	}
	if keepWidget.AsTree().Parent != findRecord(r, "c").widget {
		t.Fatal("wrong new parent")
	}
	if len(host.Children) != 1 {
		t.Fatal("parking leaked")
	}
}

func TestKeyedInvalidDescriptionAndMountFailurePreserveTree(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	host := core.NewFrame(core.NewBody("headless"))
	old, err := reconcile(host, nil, frame("root", label("keep")), a, "main")
	if err != nil {
		t.Fatal(err)
	}
	keep := old.children[0].widget
	for _, bad := range []ui.Node{
		frame("root", label("keep"), label("keep")),
		frame("root", label("new"), ui.Node{ID: "bad", Kind: ui.Extension, Custom: nilExtension{}}),
		frame("root", ui.Node{ID: "bad", Kind: ui.Extension, Custom: (*retainedExtension)(nil)}),
	} {
		actual, err := reconcile(host, old, bad, a, "main")
		if err == nil || actual != old || old.children[0].widget != keep || tree.IsNil(keep) || len(old.widget.AsTree().Children) != 1 {
			t.Fatal("failed update mutated old tree", err)
		}
	}
}

func TestExtensionReplacementDestroysOnlyReplacedInstance(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	host := core.NewFrame(core.NewBody("headless"))
	first, second := &retainedExtension{}, &retainedExtension{}
	r, err := reconcile(host, nil, frame("root", label("keep"), ui.Node{ID: "ext", Kind: ui.Extension, Custom: first}), a, "main")
	if err != nil {
		t.Fatal(err)
	}
	keep := r.children[0].widget
	r, err = reconcile(host, r, frame("root", label("keep"), ui.Node{ID: "ext", Kind: ui.Extension, Custom: second}), a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if !tree.IsNil(first.widget) || second.mounts != 1 || r.children[0].widget != keep {
		t.Fatal("incorrect replacement lifetime")
	}
}
