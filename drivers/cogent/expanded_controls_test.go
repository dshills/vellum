package cogent

import (
	"reflect"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/styles"
	"github.com/dshills/vellum/ui"
)

func TestExpandedBooleanControlsRetainedOwnerDelivery(t *testing.T) {
	for _, kind := range []ui.Kind{ui.Switch, ui.ToggleButton} {
		t.Run(string(kind), func(t *testing.T) {
			a, b := newAXTest(t, 8)
			host := core.NewFrame(core.NewBody("expanded controls"))
			var changes []ui.ChangeEvent
			description := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "control", Kind: kind, Text: "<literal>", Change: func(e ui.ChangeEvent) { changes = append(changes, e) }}}}
			r, err := reconcile(host, nil, description, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			control := findRecord(r, "control")
			w := control.widget.(*core.Switch)
			want := core.SwitchSwitch
			if kind == ui.ToggleButton {
				want = core.SwitchSegmentedButton
			}
			if w.Type != want || w.Text != "&lt;literal&gt;" {
				t.Fatal("wrong native shape or markup escaping", w.Type, w.Text)
			}
			w.SetChecked(true).SendChange()
			if len(changes) != 0 {
				t.Fatal("native event bypassed owner queue")
			}
			a.Drain()
			if len(changes) != 1 || changes[0].ID != "control" || changes[0].Kind != kind || !changes[0].Checked {
				t.Fatal(changes)
			}
			r, err = reconcile(host, r, description, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			if control.widget != w || w.IsChecked() {
				t.Fatal("rejected value not restored or identity lost")
			}
			w.SetChecked(true).SendChange()
			description.Children[0].Change = func(e ui.ChangeEvent) { e.ID = "latest"; changes = append(changes, e) }
			r, err = reconcile(host, r, description, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			a.Drain()
			if changes[len(changes)-1].ID != "latest" || !changes[len(changes)-1].Checked {
				t.Fatal("queued event used stale callback/value", changes)
			}
			nodes, handlers := accessibilitySnapshot(r, b)
			var node accessibilityNode
			for _, n := range nodes {
				if n.DomainID == "control" {
					node = n
				}
			}
			if node.Value != false || !node.Focusable || !reflect.DeepEqual(node.Actions, []string{"focus", "press", "set"}) {
				t.Fatal("checkable semantics", node)
			}
			if !handlers[node.ID](accessibilityAction{Kind: "set", Value: "true"}) || changes[len(changes)-1].Kind != kind {
				t.Fatal("semantic set refused")
			}
			before := len(changes)
			description.Disabled = true
			r, err = reconcile(host, r, description, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			w.SetChecked(true).SendChange()
			a.Drain()
			if len(changes) != before {
				t.Fatal("disabled ancestor delivered event")
			}
			description.Disabled = false
			r, err = reconcile(host, r, description, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			w.SetChecked(true).SendChange()
			description.Children = nil
			r, err = reconcile(host, r, description, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			a.Drain()
			if len(changes) != before {
				t.Fatal("retired control delivered queued event")
			}
		})
	}
}

func TestPassiveControlsReconcileAndSemantics(t *testing.T) {
	a, b := newAXTest(t, 8)
	host := core.NewFrame(core.NewBody("passive controls"))
	n := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "progress", Kind: ui.ProgressBar, Number: 0.25, Accessibility: ui.Accessibility{Name: "Loading"}}, {ID: "divider", Kind: ui.Separator}}}
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	progress := findRecord(r, "progress")
	meter := progress.widget.(*core.Meter)
	divider := findRecord(r, "divider").widget.(*core.Separator)
	if meter.Value != 0.25 || meter.Min != 0 || meter.Max != 1 || meter.Styles.Direction != styles.Row || divider.Styles.Direction != styles.Row {
		t.Fatal("passive initial state", meter.Value, meter.Styles.Direction, divider.Styles.Direction)
	}
	n.Children[0].Number = 0.75
	n.Children[0].Orientation = ui.Vertical
	n.Children[1].Orientation = ui.Vertical
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	if findRecord(r, "progress").widget != meter || findRecord(r, "divider").widget != divider || meter.Value != 0.75 || meter.Styles.Direction != styles.Column || divider.Styles.Direction != styles.Column {
		t.Fatal("orientation/value update lost identity")
	}
	nodes, handlers := accessibilitySnapshot(r, b)
	var found bool
	for _, node := range nodes {
		if node.DomainID == "divider" {
			t.Fatal("decorative separator became accessible target")
		}
		if node.DomainID == "progress" {
			found = true
			if node.Focusable || len(node.Actions) != 0 || handlers[node.ID] != nil || node.Value != float32(0.75) || node.Min == nil || *node.Min != 0 || node.Max == nil || *node.Max != 1 {
				t.Fatal("progress not read-only", node)
			}
		}
	}
	if !found {
		t.Fatal("progress inaccessible")
	}
}

func TestPassiveControlStyleOverrideAndRestore(t *testing.T) {
	a, _ := newAXTest(t, 8)
	host := core.NewFrame(core.NewBody("passive style"))
	n := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "progress", Kind: ui.ProgressBar, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 120, Height: 12}), Grow: ui.Some(float32(0))}}, {ID: "separator", Kind: ui.Separator, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 100, Height: 3}), Grow: ui.Some(float32(0))}}}}
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		id            string
		width, height float32
	}{{"progress", 120, 12}, {"separator", 100, 3}} {
		w := findRecord(r, test.id).widget.AsWidget()
		w.Style()
		if w.Styles.Min.X.Value != test.width || w.Styles.Min.Y.Value != test.height || w.Styles.Grow.X != 0 || w.Styles.Grow.Y != 0 {
			t.Fatal("native final defaults overrode managed style", test.id, w.Styles.Min, w.Styles.Grow)
		}
	}
	for i := range n.Children {
		n.Children[i].Style = ui.Style{}
	}
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	progress := findRecord(r, "progress").widget.AsWidget()
	progress.Style()
	separator := findRecord(r, "separator").widget.AsWidget()
	separator.Style()
	if progress.Styles.Min.X.Value != 320 || progress.Styles.Min.Y.Value != 8 || separator.Styles.Min.Y.Value != 1 || separator.Styles.Grow.X != 1 {
		t.Fatal("removing explicit style did not restore native defaults")
	}
}
