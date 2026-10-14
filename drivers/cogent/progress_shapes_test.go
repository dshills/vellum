package cogent

import (
	"html"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/styles"
	"github.com/dshills/vellum/ui"
)

func TestProgressShapesRetainValuesLabelsAndPassiveSemantics(t *testing.T) {
	for _, tc := range []struct {
		kind  ui.Kind
		shape core.MeterTypes
	}{
		{ui.ProgressRing, core.MeterCircle}, {ui.ProgressGauge, core.MeterSemicircle},
	} {
		t.Run(string(tc.kind), func(t *testing.T) {
			a, b := newAXTest(t, 8)
			host := core.NewFrame(core.NewBody("Progress shapes"))
			n := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "shape", Kind: tc.kind, Number: 0.25, Text: "<b>25%</b>", Accessibility: ui.Accessibility{Name: "Loading"}}}}
			r, err := reconcile(host, nil, n, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			defer func() { disposeControlTree(r) }()
			meter := findRecord(r, "shape").widget.(*core.Meter)
			if meter.Type != tc.shape || meter.Value != 0.25 || meter.Text != html.EscapeString(n.Children[0].Text) {
				t.Fatal("wrong shape/value/literal label", meter.Type, meter.Value, meter.Text)
			}
			n.Children[0].Number = 0.75
			n.Children[0].Text = "75%"
			n.Children[0].Disabled = true
			r, err = reconcile(host, r, n, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			if findRecord(r, "shape").widget != meter || meter.Value != 0.75 || meter.Text != "75%" {
				t.Fatal("retained value/label update lost identity")
			}
			nodes, handlers := accessibilitySnapshot(r, b)
			found := false
			for _, node := range nodes {
				if node.DomainID == "shape" {
					found = true
					if node.Role != "progress-bar" || node.Name != "Loading" || node.Value != float32(0.75) || node.Min == nil || *node.Min != 0 || node.Max == nil || *node.Max != 1 || node.Focusable || len(node.Actions) != 0 || handlers[node.ID] != nil || !node.Disabled {
						t.Fatal("wrong passive progress semantics", node)
					}
				}
			}
			if !found {
				t.Fatal("progress shape missing from accessibility")
			}
		})
	}
}

func TestProgressShapesStyleOverrideRestoreAndKindReplacement(t *testing.T) {
	a, _ := newAXTest(t, 8)
	host := core.NewFrame(core.NewBody("Progress shape styles"))
	for _, tc := range []struct {
		kind          ui.Kind
		width, height float32
	}{{ui.ProgressRing, 128, 128}, {ui.ProgressGauge, 112, 64}} {
		n := ui.Node{ID: "shape", Kind: tc.kind, Number: 0.5, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 80, Height: 60}), Grow: ui.Some(float32(0))}}
		r, err := reconcile(host, nil, n, a, "ax")
		if err != nil {
			t.Fatal(err)
		}
		meter := r.widget.(*core.Meter)
		meter.Style()
		if meter.Styles.Min.X.Value != 80 || meter.Styles.Min.Y.Value != 60 || meter.Styles.Grow.X != 0 || meter.Styles.Grow.Y != 0 {
			t.Fatal("shape defaults overrode explicit style", tc.kind, meter.Styles)
		}
		n.Style = ui.Style{}
		r, err = reconcile(host, r, n, a, "ax")
		if err != nil {
			t.Fatal(err)
		}
		meter.Style()
		if meter.Styles.Min.X.Value != tc.width || meter.Styles.Min.Y.Value != tc.height {
			t.Fatal("removing shape style did not restore defaults", tc.kind, meter.Styles.Min)
		}
		n.Kind = ui.ProgressBar
		n.Orientation = ui.Vertical
		r, err = reconcile(host, r, n, a, "ax")
		if err != nil {
			t.Fatal(err)
		}
		bar := r.widget.(*core.Meter)
		bar.Style()
		if bar == meter || bar.Type != core.MeterLinear || bar.Styles.Direction != styles.Column {
			t.Fatal("kind replacement retained old shape")
		}
		disposeControlTree(r)
	}
}
