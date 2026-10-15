package cogent

import (
	"errors"
	"reflect"
	"testing"

	"cogentcore.org/core/core"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

func listBoxDescriptor(change func(ui.ChangeEvent)) ui.Node {
	return ui.Node{ID: "list", Kind: ui.ListBox, SelectedID: "a", Options: []ui.Option{{ID: "a", Label: "Same"}, {ID: "b", Label: "Same"}, {ID: "c", Label: "Unavailable", Disabled: true}}, Change: change}
}
func TestListBoxRowAndHostOriginEpochs(t *testing.T) {
	for _, transition := range []string{"remove-readd", "disable-reenable", "whole-disable", "hide-show", "scope-close", "retire"} {
		t.Run(transition, func(t *testing.T) {
			a, b := newAXTest(t, 16)
			host := core.NewFrame(core.NewBody("list"))
			calls := 0
			n := listBoxDescriptor(func(ui.ChangeEvent) { calls++ })
			r, err := reconcile(host, nil, n, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			_, handlers := accessibilitySnapshot(r, b)
			oldRow, oldHost := handlers[accessibilityOptionID("list", "b")], handlers[accessibilityManagedID("list")]
			if oldRow == nil || oldHost == nil {
				t.Fatal("missing originating semantic actions")
			}
			button := r.control.choice.buttons["b"]
			queueListBox(r, "b", button)
			sync := func() {
				t.Helper()
				var err error
				r, err = reconcile(host, r, n, a, "ax")
				if err != nil {
					t.Fatal(err)
				}
			}
			switch transition {
			case "remove-readd":
				n.Options = n.Options[:1]
				sync()
				n.Options = append(n.Options, ui.Option{ID: "b", Label: "New B"})
				sync()
			case "disable-reenable":
				n.Options[1].Disabled = true
				sync()
				n.Options[1].Disabled = false
				sync()
			case "whole-disable":
				n.Disabled = true
				sync()
				n.Disabled = false
				sync()
			case "hide-show":
				setListBoxScope(r, false, false, nil)
				setListBoxScope(r, true, false, nil)
			case "scope-close":
				s := &tabCloseScope{}
				setListBoxScope(r, true, false, []*tabCloseScope{s})
				oldRow = nil
				oldHost = nil
				queueListBox(r, "b", button)
				s.closed.Store(true)
			case "retire":
				disposeControlTree(r)
			}
			a.Drain()
			if calls != 0 {
				t.Fatal("revoked native request delivered", calls)
			}
			if oldRow != nil && oldRow(accessibilityAction{Kind: "press"}) {
				t.Fatal("old row admitted after ABA")
			}
			if oldHost != nil && oldHost(accessibilityAction{Kind: "set", Value: "b"}) {
				t.Fatal("old host admitted after ABA")
			}
		})
	}
}
func TestListBoxAbsoluteProposalsRetainCurrentCallback(t *testing.T) {
	a, b := newAXTest(t, 16)
	host := core.NewFrame(core.NewBody("list"))
	var calls []string
	n := listBoxDescriptor(func(ui.ChangeEvent) { t.Fatal("stale callback") })
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	button := r.control.choice.buttons["b"]
	queueListBox(r, "b", button)
	queueListBox(r, "a", r.control.choice.buttons["a"])
	n.Change = func(e ui.ChangeEvent) { calls = append(calls, e.SelectedID) }
	n.Options[0].Label = "Renamed"
	n.Options = []ui.Option{n.Options[1], n.Options[0], n.Options[2]}
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if !reflect.DeepEqual(calls, []string{"b", "a"}) || r.control.choice.buttons["b"] != button || r.widget.AsTree().Child(0) != button {
		t.Fatal("ordered identity/current callback", calls)
	}
	if !r.control.choice.buttons["a"].IsChecked() || button.IsChecked() {
		t.Fatal("rejection changed accepted preview")
	}
	nodes, handlers := accessibilitySnapshot(r, b)
	if nodes[0].Role != "list-box" || nodes[0].RowCount != 3 || nodes[1].Role != "list-option" || nodes[1].RowIndex != 1 || nodes[2].Selected != true {
		t.Fatal(nodes)
	}
	if handlers[accessibilityManagedID("list")](accessibilityAction{Kind: "set", Value: "Same"}) || handlers[accessibilityManagedID("list")](accessibilityAction{Kind: "set", Value: "c"}) {
		t.Fatal("label or disabled set accepted")
	}
	if !handlers[accessibilityOptionID("list", "a")](accessibilityAction{Kind: "press"}) || !reflect.DeepEqual(calls, []string{"b", "a", "a"}) {
		t.Fatal("explicit same-selection proposal missing", calls)
	}
	n.Change = nil
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	_, handlers = accessibilitySnapshot(r, b)
	if handlers[accessibilityManagedID("list")](accessibilityAction{Kind: "set", Value: "b"}) || handlers[accessibilityOptionID("list", "b")](accessibilityAction{Kind: "press"}) {
		t.Fatal("readonly list changed")
	}
}
func TestListBoxQueueRefusalAndWindowClosure(t *testing.T) {
	a, _ := newAXTest(t, 1)
	host := core.NewFrame(core.NewBody("list"))
	n := listBoxDescriptor(func(ui.ChangeEvent) { t.Fatal("refused/closed callback delivered") })
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	reports := 0
	r.control.onError = func(err error) {
		if !errors.Is(err, app.ErrQueueFull) {
			t.Fatal(err)
		}
		reports++
	}
	if err := a.Post(func() {}); err != nil {
		t.Fatal(err)
	}
	r.control.choice.buttons["b"].SetChecked(true).SendChange()
	a.Drain()
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	if reports != 1 || !r.control.choice.buttons["a"].IsChecked() || r.control.choice.buttons["b"].IsChecked() {
		t.Fatal("refusal restore", reports)
	}
	queueListBox(r, "b", r.control.choice.buttons["b"])
	if err := a.CloseWindow("ax"); err != nil {
		t.Fatal(err)
	}
	a.Drain()
}
