//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/styles/units"
	"github.com/dshills/vellum/ui"
	"math"
	"testing"
)

func TestToolbarSpacesGrowHideAndStayOutOfOverflow(t *testing.T) {
	width := float32(360)
	source := fullColorButtonSource(t, false)
	view := func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{ui.WithToolbarOverflow(ui.Node{ID: "tools", Kind: ui.Toolbar, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: width}), MaxSize: ui.Some(ui.Size{Width: width}), GrowX: ui.Some(float32(1))}, Children: []ui.Node{
			{ID: "first", Kind: ui.Button, Text: "First", Activate: func() {}},
			{ID: "fixed", Kind: ui.ToolbarSpace, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 24, Height: 1}), MaxSize: ui.Some(ui.Size{Width: 24, Height: 1})}}, {ID: "flex", Kind: ui.ToolbarFlexibleSpace},
			{ID: "last", Kind: ui.Button, Text: "Last", IconImage: source, Activate: func() {}},
		}})}}
	}
	m, a := simpleManual(t, manualConfig(t, 7), view)
	unitSettle(t, m)
	root := m.implementation.find("one").state.root
	fixed := findRecord(root, "fixed")
	flex := findRecord(root, "flex")
	if fixed.widget.AsWidget().Geom.TotalBBox.Dx() <= 0 || flex.widget.AsWidget().Geom.TotalBBox.Dx() <= fixed.widget.AsWidget().Geom.TotalBBox.Dx() {
		t.Fatal("flex space did not expand", fixed.widget.AsWidget().Geom.Size, flex.widget.AsWidget().Geom.Size)
	}
	expected := units.Dp(24)
	expected.ToDots(&fixed.widget.AsWidget().Styles.UnitContext)
	if math.Abs(float64(fixed.widget.AsWidget().Geom.Size.Actual.Total.X-expected.Dots)) > 1 {
		t.Fatal("explicit fixed-space width ignored", fixed.widget.AsWidget().Geom.Size, expected.Dots)
	}
	bar0 := findRecord(root, "tools").widget.AsWidget()
	last := findRecord(root, "last").widget.AsWidget()
	if last.Geom.TotalBBox.Empty() || !last.Geom.TotalBBox.In(bar0.Geom.TotalBBox) {
		t.Fatal("flex spacer pushed trailing action out of view", last.Geom.TotalBBox, bar0.Geom.TotalBBox, last.Geom.Pos.Total, last.Geom.Size, flex.widget.AsWidget().Geom.Size, findRecord(root, "tools").widget.(*overflowToolbarFrame).widths, findRecord(root, "tools").widget.(*overflowToolbarFrame).hidden)
	}
	if flex.widget.AsWidget().Styles.RenderBox || fixed.widget.AsWidget().Styles.RenderBox {
		t.Fatal("space paints a surface over actions")
	}
	frame, e := m.Capture(testContext, "one")
	if e != nil {
		t.Fatal(e)
	}
	picture := findRecord(root, "last").widget.(*core.Button).ChildByName(buttonImageName).(*managedImage)
	box := picture.Geom.ContentBBox
	left := frame.RGBAAt(box.Min.X+box.Dx()/6, box.Min.Y+box.Dy()/2)
	right := frame.RGBAAt(box.Max.X-1-box.Dx()/6, box.Min.Y+box.Dy()/2)
	if left.R < 180 || left.B > 60 || right.B < 180 || right.R > 60 {
		t.Fatal("space obscured trailing full-color action", left, right, box)
	}
	original := flex.widget
	nodes, _ := accessibilitySnapshot(root, nil)
	for _, n := range nodes {
		if n.DomainID == "fixed" || n.DomainID == "flex" {
			t.Fatal("decorative spacer exposed to AX")
		}
	}
	width = 90
	a.Invalidate("one")
	unitSettle(t, m)
	bar := findRecord(root, "tools")
	more := findRecord(root, ui.ToolbarOverflowID("tools"))
	opts := toolbarOverflowOptions(bar, more.node.Options)
	for _, o := range opts {
		if o.ID == "fixed" || o.ID == "flex" {
			t.Fatal("space entered overflow menu")
		}
	}
	width = 360
	a.Invalidate("one")
	unitSettle(t, m)
	if findRecord(root, "flex").widget != original {
		t.Fatal("resize replaced spacer")
	}
	if findRecord(root, "flex").widget.AsWidget().Styles.Grow.X != 1 {
		t.Fatal("lost flex growth")
	}
}
