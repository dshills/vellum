//go:build offscreen

package cogent

import (
	"fmt"
	"testing"

	"cogentcore.org/core/core"
	"github.com/dshills/vellum/ui"
)

func TestMacFieldInsetsAndOutlineGeometry(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, dpi := range []float32{96, 144, 192} {
			t.Run(fmt.Sprintf("dark-%v-dpi-%.0f", dark, dpi), func(t *testing.T) {
				field, err := (ui.FormField{ID: "group", Label: "Drawing name", Help: "A name for your drawing.",
					Content: ui.Node{ID: "name", Kind: ui.TextField, Text: "Untitled"}}).View()
				if err != nil {
					t.Fatal(err)
				}
				view := func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{field}} }
				config := manualConfig(t, 7)
				background := config.Theme.Default.Background
				config.LogicalDPI = dpi
				config.Theme = ui.MacTheme(dark)
				config.Theme.Default.Background = background
				config.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
				m, _ := simpleManual(t, config, view)
				unitSettle(t, m)
				r := m.implementation.find("one").state.root
				group := findRecord(r, "group")
				input := findRecord(r, "name").widget.(*core.TextField)
				if input.Type != core.TextFieldOutlined {
					t.Fatal("custom full border uses bottom-only field", input.Type)
				}
				if group.widget.AsWidget().Styles.Padding.Left.Value < 12 || group.widget.AsWidget().Styles.Padding.Top.Value < 12 {
					t.Fatal("group inset lost")
				}
				outer := group.widget.AsWidget().Geom.TotalBBox
				caption := group.control.chrome.caption.Geom.TotalBBox
				box := input.Geom.TotalBBox
				if caption.Min.X <= outer.Min.X || caption.Min.Y <= outer.Min.Y || box.Min.X <= outer.Min.X || box.Max.X >= outer.Max.X {
					t.Fatal("caption/input touches group border", outer, caption, box)
				}
				if input.Styles.Padding.Left.Value < 8 || input.Styles.Padding.Top.Value < 5 || input.Styles.MaxBorder.Width.Top.Value < 2 {
					t.Fatal("text inset or full focus border lost")
				}
				if input.Text() != "Untitled" {
					t.Fatal("style changed accepted value")
				}
			})
		}
	}
}

func TestMacFieldRetainedOutlineRestylesWithoutReplacingDraft(t *testing.T) {
	a, _ := newAXTest(t, 8)
	host := core.NewFrame(core.NewBody("retained input outline"))
	n := ui.Node{ID: "field", Kind: ui.TextField, Text: "accepted"}
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	field := r.widget.(*core.TextField)
	field.Style()
	field.SetText("private draft")
	n.Style = ui.MacTheme(false).Kinds[ui.TextField]
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	if r.widget != field || field.Type != core.TextFieldOutlined || field.Styles.MaxBorder.Width.Top.Value != 2 || field.Text() != "private draft" {
		t.Fatal("retained outline did not restyle or draft changed", field.Type, field.Text())
	}
	n.Style = ui.Style{}
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	if r.widget != field || field.Type != core.TextFieldFilled || field.Styles.MaxBorder.Width.Top.Value != 0 || field.Text() != "private draft" {
		t.Fatal("removing outline did not restore field", field.Type, field.Text())
	}
}
