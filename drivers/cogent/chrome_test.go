package cogent

import (
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"github.com/dshills/vellum/ui"
)

func TestChromeRetainsManagedChildrenAndPrivateCaptions(t *testing.T) {
	for _, kind := range []ui.Kind{ui.Toolbar, ui.StatusBar, ui.FormGroup} {
		t.Run(string(kind), func(t *testing.T) {
			a, b := newAXTest(t, 8)
			host := core.NewFrame(core.NewBody("chrome"))
			edits := 0
			n := ui.Node{ID: "chrome", Kind: kind, Text: "<Title>", Children: []ui.Node{{ID: "field", Kind: ui.TextField, Text: "Accepted"}, {ID: "check", Kind: ui.Switch, Change: func(ui.ChangeEvent) { edits++ }}}}
			r, err := reconcile(host, nil, n, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			c := r.control.chrome
			content := c.host
			caption := c.caption
			wantsCaption := kind != ui.Toolbar
			if (caption != nil) != wantsCaption {
				t.Fatal("private caption presence does not match container kind", kind)
			}
			field := r.children[0].widget.(*core.TextField)
			check := r.children[1].widget.(*core.Switch)
			if wantsCaption && (caption.Text != "&lt;Title&gt;" || r.widget.AsTree().NumChildren() != 2) {
				t.Fatal("caption literal/private structure missing")
			}
			if !field.ApplyNativeEdit("Pending", 7, 7, -1, -1, "insert") {
				t.Fatal("fixture edit refused")
			}
			before := field.EditingState()
			check.SetChecked(true).SendChange()
			n.Text = "New title"
			n.Children[0], n.Children[1] = n.Children[1], n.Children[0]
			n.Children[0].Change = func(ui.ChangeEvent) { edits += 10 }
			r, err = reconcile(host, r, n, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			a.Drain()
			if c.host != content || c.caption != caption || r.children[0].widget != check || r.children[1].widget != field || content.Child(0) != check || content.Child(1) != field || field.EditingState() != before || edits != 10 {
				t.Fatal("retained reorder changed binding, draft, or latest callback", edits)
			}
			if wantsCaption && (caption.Text != "New title" || r.widget.AsTree().Child(0) != caption || r.widget.AsTree().Child(1) != content) {
				t.Fatal("managed reorder replaced private caption/content")
			}
			n.Text = ""
			n.Disabled = true
			r, err = reconcile(host, r, n, a, "ax")
			if err != nil {
				t.Fatal(err)
			}
			if wantsCaption {
				caption.Style()
				if caption.Styles.Display != styles.DisplayNone || !caption.StateIs(states.Disabled) {
					t.Fatal("empty/disabled caption still visible or enabled")
				}
			}
			check.SetChecked(true).SendChange()
			a.Drain()
			if edits != 10 {
				t.Fatal("disabled container delivered child change")
			}
			nodes, handlers := accessibilitySnapshot(r, b)
			if len(nodes) != 3 || nodes[0].Role != string(kind) || nodes[0].Focusable || len(nodes[0].Actions) != 0 || !nodes[1].Disabled || len(handlers) != 0 {
				t.Fatal("container semantics/private caption/disabled actions", nodes, len(handlers))
			}
			if kind == ui.StatusBar && nodes[0].Value != "" {
				t.Fatal("status value stale", nodes[0].Value)
			}
		})
	}
}
func TestChromeMoveChildrenOutBeforeOldHostRemoval(t *testing.T) {
	a, _ := newAXTest(t, 8)
	host := core.NewFrame(core.NewBody("move"))
	n := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "old", Kind: ui.FormGroup, Text: "Old", Children: []ui.Node{{ID: "field", Kind: ui.TextField, Text: "Value"}}}}}
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	field := findRecord(r, "field").widget.(*core.TextField)
	if !field.ApplyNativeEdit("Draft", 5, 5, -1, -1, "insert") {
		t.Fatal("edit refused")
	}
	before := field.EditingState()
	n.Children = []ui.Node{{ID: "next", Kind: ui.StatusBar, Text: "Next", Children: n.Children[0].Children}}
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	if findRecord(r, "field").widget != field || field.EditingState() != before || field.Parent != r.children[0].control.chrome.host {
		t.Fatal("move removed retained field/draft")
	}
}

func TestChromeContentStyleOverrideAndRestoration(t *testing.T) {
	for _, kind := range []ui.Kind{ui.FormGroup, ui.StatusBar} {
		a, _ := newAXTest(t, 8)
		host := core.NewFrame(core.NewBody("content style"))
		n := ui.Node{ID: "group", Kind: kind, Style: ui.Style{Gap: ui.Some(float32(23)), Align: ui.Some(ui.AlignEnd), Justify: ui.Some(ui.AlignCenter)}, Children: []ui.Node{{ID: "one", Kind: ui.Label, Text: "One"}, {ID: "two", Kind: ui.Label, Text: "Two"}}}
		r, err := reconcile(host, nil, n, a, "ax")
		if err != nil {
			t.Fatal(err)
		}
		content := r.control.chrome.host
		content.Style()
		if content.Styles.Gap.X.Value != 23 || content.Styles.Align.Items != styles.End || content.Styles.Justify.Content != styles.Center {
			t.Fatal("managed child spacing/alignment ignored")
		}
		n.Style = ui.Style{}
		r, err = reconcile(host, r, n, a, "ax")
		if err != nil {
			t.Fatal(err)
		}
		content.Style()
		if content.Styles.Gap.X.Value != 8 || content.Styles.Align.Items == styles.End || content.Styles.Justify.Content == styles.Center {
			t.Fatal("removed content style not restored")
		}
	}
}

func TestChromeCaptionPreservesFontWeightAndRestoresDefault(t *testing.T) {
	for _, kind := range []ui.Kind{ui.StatusBar, ui.FormGroup} {
		a, _ := newAXTest(t, 8)
		host := core.NewFrame(core.NewBody("font"))
		n := ui.Node{ID: "group", Kind: kind, Text: "Caption", Style: ui.Style{Typography: ui.Some(ui.Typography{Size: 20, Weight: 600})}}
		r, err := reconcile(host, nil, n, a, "ax")
		if err != nil {
			t.Fatal(err)
		}
		outer := r.widget.AsWidget()
		caption := r.control.chrome.caption
		outer.Style()
		caption.Style()
		weighted := outer.Styles.Font.Weight
		if caption.Styles.Font.Weight != weighted {
			t.Fatal("caption lost inherited font weight", caption.Styles.Font.Weight, weighted)
		}
		n.Style = ui.Style{}
		r, err = reconcile(host, r, n, a, "ax")
		if err != nil {
			t.Fatal(err)
		}
		outer.Style()
		caption.Style()
		if caption.Styles.Font.Weight != outer.Styles.Font.Weight || caption.Styles.Font.Weight == weighted {
			t.Fatal("caption did not restore inherited default weight")
		}
	}
}

func TestChromeCaptionInheritsSizeWithProcessLocalFontScaling(t *testing.T) {
	// This is an isolated Go style fixture; no settings API or persistent OS/app
	// preference is called. Other tests in this package do not run in parallel.
	body := core.NewBody("initialize scaled style fixture")
	original := core.AppearanceSettings.FontSize
	t.Cleanup(func() { core.AppearanceSettings.FontSize = original })
	for _, kind := range []ui.Kind{ui.StatusBar, ui.FormGroup} {
		a, _ := newAXTest(t, 8)
		host := core.NewFrame(body)
		r, err := reconcile(host, nil, ui.Node{ID: "group", Kind: kind, Text: "Caption", Style: ui.Style{Typography: ui.Some(ui.Typography{Size: 20, Weight: 600})}}, a, "ax")
		if err != nil {
			t.Fatal(err)
		}
		core.AppearanceSettings.FontSize = 125
		outer := r.widget.AsWidget()
		caption := r.control.chrome.caption
		outer.Style()
		caption.Style()
		if core.AppearanceSettings.FontSize != 125 || outer.Styles.Font.Size.Value != 25 {
			t.Fatal("scaled fixture did not establish 125%/25dp", core.AppearanceSettings.FontSize, outer.Styles.Font.Size.Value)
		}
		if caption.Styles.Font.Size.Value != outer.Styles.Font.Size.Value {
			t.Fatal("caption double-scaled inherited font", caption.Styles.Font.Size.Value, outer.Styles.Font.Size.Value)
		}
	}
}
