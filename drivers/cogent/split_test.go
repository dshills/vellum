package cogent

import (
	"math"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/ui"
)

func splitFixture() ui.Node {
	return ui.Node{ID: "split", Kind: ui.SplitPane, Number: 0.5, Accessibility: ui.Accessibility{Name: "Workspace divider"}, Children: []ui.Node{{ID: "left", Kind: ui.TextField, Text: "Accepted"}, {ID: "right", Kind: ui.Button, Text: "Right"}}}
}

func TestSplitRetainedHostsDraftAndLatestCallback(t *testing.T) {
	a, _ := newAXTest(t, 16)
	host := core.NewFrame(core.NewBody("split retained"))
	n := splitFixture()
	changes := []ui.ChangeEvent{}
	n.Change = func(e ui.ChangeEvent) { t.Fatal("old callback delivered", e) }
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	c := r.control.split
	field := r.children[0].widget.(*core.TextField)
	if !field.ApplyNativeEdit("Draft", 5, 5, -1, -1, "insert") {
		t.Fatal("draft refused")
	}
	draft := field.EditingState()
	queueSplit(r, 0.99)
	if r.control.split.native.Split(0) != 0.5 {
		t.Fatal("native resize changed accepted fraction before owner")
	}
	n.Numeric = ui.Numeric{Min: 0.2, Max: 0.8, Step: 0.05}
	n.Change = func(e ui.ChangeEvent) { changes = append(changes, e) }
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if len(changes) != 1 || changes[0].ID != "split" || changes[0].Kind != ui.SplitPane || changes[0].Number != 0.8 {
		t.Fatal("latest bounds/callback not used", changes)
	}
	// The application refuses the request and returns the same accepted value.
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil || r.control.split != c || r.children[0].widget != field || field.EditingState() != draft || r.control.split.native.Split(0) != 0.5 {
		t.Fatal("refused resize lost retained state", err)
	}
	queueSplit(r, 0.7)
	n.Children[0], n.Children[1] = n.Children[1], n.Children[0]
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if len(changes) != 1 || r.children[1].widget != field || field.Parent != c.panes[1] || c.panes[1].Child(0) != field || field.EditingState() != draft {
		t.Fatal("reorder applied old resize or lost draft", changes)
	}
}

func TestSplitDisabledRetiredAndSemanticActions(t *testing.T) {
	a, b := newAXTest(t, 16)
	host := core.NewFrame(core.NewBody("split semantics"))
	n := splitFixture()
	changes := []ui.ChangeEvent{}
	n.Change = func(e ui.ChangeEvent) { changes = append(changes, e) }
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	nodes, handlers := accessibilitySnapshot(r, b)
	id := accessibilitySplitHandleID("split")
	if len(nodes) != 4 || nodes[0].Role != "split-pane" || nodes[0].Focusable || nodes[2].ID != id || nodes[2].Role != "split-handle" || nodes[2].Value != float32(0.5) || nodes[2].Orientation != "vertical" || *nodes[2].Min != 0.05 || *nodes[2].Max != 0.95 {
		t.Fatal("split hierarchy/value/bounds", nodes)
	}
	if !handlers[id](accessibilityAction{Kind: "focus"}) || b.pendingFocus.Load() == nil || *b.pendingFocus.Load() != id {
		t.Fatal("split focus action did not qualify divider")
	}
	targets := map[string]accessibilityFocusTarget{}
	accessibilityNativeFocusTargets(r, "", targets)
	if !accessibilityApplyFocus(r, id) || targets[id].widget != r.control.split.handle {
		t.Fatal("split start-focus did not resolve retained native handle")
	}
	for _, action := range []accessibilityAction{{Kind: "increment"}, {Kind: "decrement"}, {Kind: "set", Value: "2"}} {
		if !handlers[id](action) {
			t.Fatal("valid split action refused", action)
		}
	}
	if len(changes) != 3 || changes[0].Number != 0.51 || changes[1].Number != 0.49 || changes[2].Number != 0.95 {
		t.Fatal("semantic numeric changes", changes)
	}
	for _, value := range []string{"NaN", "Inf", "-Inf", "1e300", "bad"} {
		if handlers[id](accessibilityAction{Kind: "set", Value: value}) {
			t.Fatal("invalid native fraction admitted", value)
		}
	}
	queueSplit(r, 0.7)
	n.Disabled = true
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if len(changes) != 3 || handlers[id](accessibilityAction{Kind: "increment"}) {
		t.Fatal("disabled split delivered queued/stale action")
	}
	n.Disabled = false
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	queueSplit(r, 0.7)
	disposeControl(r)
	a.Drain()
	queueSplit(r, float32(math.NaN()))
	if len(changes) != 3 || handlers[id](accessibilityAction{Kind: "set", Value: "0.6"}) {
		t.Fatal("retired split delivered change")
	}
}

func TestSplitKeyboardOrientationAndNativeShortcutGuard(t *testing.T) {
	for _, orientation := range []ui.Orientation{ui.Horizontal, ui.Vertical} {
		a, _ := newAXTest(t, 16)
		n := splitFixture()
		n.Orientation = orientation
		changes := []ui.ChangeEvent{}
		n.Change = func(e ui.ChangeEvent) { changes = append(changes, e) }
		r, err := reconcile(core.NewFrame(core.NewBody("split keys")), nil, n, a, "ax")
		if err != nil {
			t.Fatal(err)
		}
		code := key.CodeRightArrow
		if orientation == ui.Vertical {
			code = key.CodeDownArrow
		}
		e := events.NewKey(events.KeyChord, 0, code, 0)
		r.control.split.handle.HandleEvent(e)
		a.Drain()
		if !e.IsHandled() || len(changes) != 1 || changes[0].Number != 0.51 {
			t.Fatal("divider keyboard change", orientation, e.IsHandled(), changes)
		}
		for _, mods := range []key.Modifiers{1 << key.Meta, 1 << key.Control} {
			e := events.NewKey(events.KeyChord, '1', key.Code1, mods)
			r.widget.AsWidget().HandleEvent(e)
			if !e.IsHandled() || r.control.split.native.Split(0) != 0.5 {
				t.Fatal("native shortcut bypassed accepted fraction")
			}
		}
	}
}

func TestSplitInvalidUpdateAndNilMountPreserveAcceptedPanes(t *testing.T) {
	a, _ := newAXTest(t, 16)
	host := core.NewFrame(core.NewBody("split transactional rejection"))
	n := splitFixture()
	changes := []ui.ChangeEvent{}
	n.Change = func(e ui.ChangeEvent) { changes = append(changes, e) }
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	c := r.control.split
	left, right := r.children[0].widget, r.children[1].widget
	queueSplit(r, 0.7)
	tooMany := n
	tooMany.Children = append(append([]ui.Node(nil), n.Children...), ui.Node{ID: "third", Kind: ui.Label})
	nilMount := n
	nilMount.Children = []ui.Node{n.Children[0], {ID: "nil-mount", Kind: ui.Extension, Custom: nilExtension{}}}
	for _, invalid := range []ui.Node{tooMany, nilMount} {
		actual, err := reconcile(host, r, invalid, a, "ax")
		if err == nil || actual != r || r.control.split != c || c.childIDs != ([2]string{"left", "right"}) || len(r.children) != 2 || r.children[0].widget != left || r.children[1].widget != right || c.panes[0].Child(0) != left || c.panes[1].Child(0) != right || tree.IsNil(c.handle) || c.native.Split(0) != 0.5 {
			t.Fatal("invalid split update mutated accepted native structure", err)
		}
	}
	a.Drain()
	if len(changes) != 1 || changes[0].Number != 0.7 {
		t.Fatal("failed preparation retired a valid accepted-pane resize", changes)
	}
}
