//go:build offscreen

package cogent

import (
	"image/color"
	"reflect"
	"testing"

	"cogentcore.org/core/styles/units"
	"github.com/dshills/vellum/ui"
)

// Exercise retained native widgets and their completed layout. Optional zero
// clears an override; removing it must restore the original native appearance
// and preserve the focused control rather than reconstructing that control.
func TestAppSurfaceStyleGrowthOutlineRadiusAndFocusedReset(t *testing.T) {
	base := ui.Style{MinSize: ui.Some(ui.Size{Width: 130, Height: 40}), Padding: ui.Some(ui.UniformInsets(4))}
	current := base
	view := func() ui.Node {
		return ui.Node{ID: "surface", Kind: ui.Row, Grow: true, Focus: ui.FocusRequest{ID: "other", Sequence: 1}, Children: []ui.Node{
			{ID: "action", Kind: ui.Button, Text: "Keep focus", Appearance: ui.ButtonSecondary, Style: current, Activate: func() {}},
			{ID: "other", Kind: ui.Button, Text: "Other", Appearance: ui.ButtonSecondary, Style: base, Activate: func() {}},
		}}
	}
	m, a := simpleManual(t, manualConfig(t, 7), view)
	unitSettle(t, m)
	record := findRecord(m.implementation.find("one").state.root, "action")
	widget := record.widget
	native := widget.AsWidget()
	baseline := native.Styles
	baselineSize := native.Geom.Size.Actual.Total
	baselineFocusedBorder := findRecord(m.implementation.find("one").state.root, "other").widget.AsWidget().Styles.Border
	expectedFocus := "other"
	assertFocus := func() {
		t.Helper()
		snapshot, err := m.Snapshot(testContext)
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Windows[0].FocusedID != expectedFocus || native.ContainsFocus() != (expectedFocus == "action") {
			t.Fatalf("styling lost keyboard focus: %+v", snapshot.Windows[0])
		}
		if findRecord(m.implementation.find("one").state.root, "action") != record || record.widget != widget {
			t.Fatal("styling replaced retained focused control")
		}
	}
	focus := func(id string) {
		t.Helper()
		if err := a.Post(func() { findRecord(m.implementation.find("one").state.root, id).widget.AsWidget().SetFocus() }); err != nil {
			t.Fatal(err)
		}
		expectedFocus = id
		unitSettle(t, m)
		assertFocus()
	}
	update := func(value ui.Style) {
		t.Helper()
		if err := a.Post(func() { current = value; _ = a.Invalidate("one") }); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
		assertFocus()
	}
	assertFocus()
	tint := color.NRGBA{R: 181, G: 42, B: 74, A: 255}
	update(base.Overlay(ui.Style{Grow: ui.Some(float32(1)), GrowX: ui.Some(float32(2)), GrowY: ui.Some(float32(.25)), Border: ui.Some(ui.Border{Width: 3, Color: tint}), CornerRadius: ui.Some(float32(12))}))
	s := native.Styles
	if s.Grow.X != 2 || s.Grow.Y != .25 {
		t.Fatalf("axis overrides lost to all-axis growth: %v", s.Grow)
	}
	if s.Border.Width.Top.Value != 3 || s.Border.Width.Top.Unit != units.UnitDp || s.Border.Color.Top.At(0, 0) != tint || s.Border.Radius.Top.Value != 12 || s.Border.Radius.Top.Unit != units.UnitDp {
		t.Fatalf("actual native outline/radius differs: %+v", s.Border)
	}
	if size := native.Geom.Size.Actual.Total; size.X < baselineSize.X+20 || size.Y < baselineSize.Y+20 {
		t.Fatalf("positive axis growth did not change allocated layout: baseline=%v grown=%v", baselineSize, size)
	}
	// The resting outline is application-owned; keyboard focus retains the
	// native indicator even when the application later requests zero width.
	focus("action")
	if native.Styles.Border.Width.Top.Value <= 0 {
		t.Fatal("custom resting outline removed native focus indicator")
	}
	update(base.Overlay(ui.Style{Grow: ui.Some(float32(1)), GrowX: ui.Some(float32(0)), GrowY: ui.Some(float32(0)), Border: ui.Some(ui.Border{}), CornerRadius: ui.Some(float32(0))}))
	s = native.Styles
	if s.Grow.X != 0 || s.Grow.Y != 0 || s.Border.Width.Top.Value <= 0 || s.Border.Radius.Top.Value != 0 {
		t.Fatalf("explicit zero failed axis/radius reset or removed native focus indicator: grow=%v border=%+v", s.Grow, s.Border)
	}
	focus("other")
	if native.Styles.Border.Width.Top.Value != 0 {
		t.Fatal("explicit zero did not remove the unfocused resting outline")
	}
	focus("action")
	if native.Styles.Border.Width.Top.Value <= 0 {
		t.Fatal("zero resting outline suppressed restored native focus indicator")
	}
	update(base)
	s = native.Styles
	if s.Grow != baseline.Grow || !reflect.DeepEqual(s.Border, baselineFocusedBorder) || native.Geom.Size.Actual.Total != baselineSize {
		t.Fatalf("removing overrides failed native property/layout reset: baseline grow=%v focused border=%+v size=%v; got grow=%v border=%+v size=%v", baseline.Grow, baselineFocusedBorder, baselineSize, s.Grow, s.Border, native.Geom.Size.Actual.Total)
	}
}
