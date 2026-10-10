package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/tree"
	"example.com/vellum/app"
	"example.com/vellum/ui"
	"sync/atomic"
	"testing"
)

type nilExtension struct{}

func (nilExtension) Mount(tree.Node) core.Widget { return nil }
func (nilExtension) Sync(core.Widget) error      { return nil }
func TestRetainedReconcileAndExtensionFailure(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	host := core.NewBody("headless")
	n := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "label", Kind: ui.Label, Text: "one"}}}
	r, err := mount(host, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	original := r.children[0].widget
	n.Children[0].Text = "two"
	if !sameShape(r, n) {
		t.Fatal("text change recreated widget")
	}
	if err := syncRecord(r, n); err != nil {
		t.Fatal(err)
	}
	if r.children[0].widget != original || original.(*core.Text).Text != "two" {
		t.Fatal("identity lost")
	}
	n.Children[0].ID = "other"
	if sameShape(r, n) {
		t.Fatal("identity change ignored")
	}
	if _, err := mount(host, ui.Node{ID: "nil", Kind: ui.Extension, Custom: nilExtension{}}, a, "main"); err == nil {
		t.Fatal("nil extension accepted")
	}
}

func TestOtherCustomEventPreservesWake(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	var source events.Source
	var pending atomic.Bool
	pending.Store(true)
	n := 0
	a.Post(func() { n++ })
	source.Custom("vellum.wake")
	source.Custom("other producer")
	e, ok := source.Deque.PollEvent()
	if !ok {
		t.Fatal("no event")
	}
	drainCustom(a, &pending, func() {}, e)
	if n != 1 || pending.Load() || e.IsHandled() {
		t.Fatal("compressed wake lost or unrelated event consumed")
	}
}

func TestButtonDisplaysLiteralText(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	host := core.NewBody("headless")
	n := ui.Node{ID: "button", Kind: ui.Button, Text: "a < b & c"}
	r, err := mount(host, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if r.widget.(*core.Button).Text != "a &lt; b &amp; c" {
		t.Fatal("markup interpreted")
	}
	n.Text = "<b>literal</b>"
	if err := syncRecord(r, n); err != nil {
		t.Fatal(err)
	}
	if r.widget.(*core.Button).Text != "&lt;b&gt;literal&lt;/b&gt;" {
		t.Fatal("updated markup interpreted")
	}
}
