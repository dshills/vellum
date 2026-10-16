package cogent

import (
	"image/color"
	"reflect"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/text/rich"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/workbench"
)

func TestNativeStyleAppliesThemeAndNode(t *testing.T) {
	theme := ui.Theme{Default: ui.Style{Typography: ui.Some(ui.Typography{Size: 18, Weight: 600}), Foreground: ui.Some(color.NRGBA{R: 19, A: 255})}, Kinds: map[ui.Kind]ui.Style{ui.Row: {Gap: ui.Some(float32(8))}}}
	n := ui.Node{ID: "r", Kind: ui.Row, Grow: true, MinWidth: 40, Style: ui.Style{Padding: ui.Some(ui.Insets{Top: 1, Right: 2, Bottom: 3, Left: 4}), Align: ui.Some(ui.AlignCenter), Justify: ui.Some(ui.AlignEnd), MaxSize: ui.Some(ui.Size{Width: 100, Height: 200}), Background: ui.Some(color.NRGBA{B: 25, A: 255})}}
	var s styles.Style
	s.Defaults()
	applyStyle(&s, n, theme)
	if s.Direction != styles.Row || s.Gap.X.Value != 8 || s.Padding.Left.Value != 4 || s.Padding.Top.Value != 1 || s.Min.X.Value != 40 || s.Max.Y.Value != 200 || s.Grow.X != 1 || s.Align.Items != styles.Center || s.Align.Content != styles.Center || s.Justify.Content != styles.End {
		t.Fatalf("native layout not applied: %+v", s)
	}
	if s.Font.Size.Value != 18 || s.Font.Weight != rich.Semibold || s.Color.At(0, 0) != (color.NRGBA{R: 19, A: 255}) || s.Background.At(0, 0) != (color.NRGBA{B: 25, A: 255}) {
		t.Fatal("native theme appearance not applied")
	}
}

func TestStyleRemovalRestoresNativeDefaultsAndIdentity(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	n := ui.Node{ID: "label", Kind: ui.Label, Text: "text", Style: ui.Style{Padding: ui.Some(ui.UniformInsets(30)), Gap: ui.Some(float32(23)), MaxSize: ui.Some(ui.Size{Width: 500, Height: 400}), Background: ui.Some(color.NRGBA{R: 255, A: 255}), Typography: ui.Some(ui.Typography{Family: "Example", Size: 22, Weight: 700, Italic: true})}}
	r, err := mount(core.NewBody("headless"), n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	widget := r.widget
	widget.AsWidget().Style()
	if widget.AsWidget().Styles.Padding.Left.Value != 30 || widget.AsWidget().Styles.Font.CustomFont != "Example" {
		t.Fatal("style not installed in native widget")
	}
	n.Style = ui.Style{}
	if err := syncRecord(r, n); err != nil {
		t.Fatal(err)
	}
	widget.AsWidget().Style()
	baseline, err := mount(core.NewBody("baseline"), n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	baseline.widget.AsWidget().Style()
	actual, want := widget.AsWidget().Styles, baseline.widget.AsWidget().Styles
	if r.widget != widget || !reflect.DeepEqual(actual.Padding, want.Padding) || !reflect.DeepEqual(actual.Gap, want.Gap) || !reflect.DeepEqual(actual.Max, want.Max) || !reflect.DeepEqual(actual.Font, want.Font) || actual.Background != want.Background {
		t.Fatal("removing style retained overrides or replaced native widget")
	}
}

func TestExtensionLegacySizingPreservesOtherNativeAxis(t *testing.T) {
	var s styles.Style
	s.Defaults()
	s.Min.Set(units.Dp(12), units.Dp(24))
	s.Grow.Set(2, 3)
	applyStyle(&s, ui.Node{ID: "ext", Kind: ui.Extension, MinWidth: 15}, ui.Theme{})
	if s.Min.X.Value != 15 || s.Min.Y.Value != 24 || s.Grow.X != 2 || s.Grow.Y != 3 {
		t.Fatal("legacy extension dimensions or native growth lost")
	}
	applyStyle(&s, ui.Node{ID: "ext", Kind: ui.Extension, Style: ui.Style{MinSize: ui.Some(ui.Size{}), Grow: ui.Some(float32(0))}}, ui.Theme{})
	if s.Min.X.Value != 0 || s.Min.Y.Value != 0 || s.Grow.X != 0 || s.Grow.Y != 0 {
		t.Fatal("explicit zero did not clear extension dimensions and growth")
	}
}

func TestThemeConflictRejectsWholeKeyedUpdate(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	host := core.NewFrame(core.NewBody("headless"))
	theme := ui.Theme{Default: ui.Style{MaxSize: ui.Some(ui.Size{Width: 50})}}
	n := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "keep", Kind: ui.Label, Text: "before"}}}
	r, err := reconcile(host, nil, n, a, "main", theme)
	if err != nil {
		t.Fatal(err)
	}
	keep := r.children[0].widget
	next := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "keep", Kind: ui.Label, Text: "after"}, {ID: "invalid", Kind: ui.Button, MinWidth: 100}}}
	actual, err := reconcile(host, r, next, a, "main", theme)
	if err == nil || actual != r || len(r.children) != 1 || r.children[0].widget != keep || keep.(*core.Text).Text != "before" {
		t.Fatal("theme conflict partially mutated retained tree", err)
	}
}

func TestWorkbenchPanelReceivesIndependentThemeSnapshot(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	model, err := workbench.New(workbench.Layout{Version: 1, Root: &workbench.Group{ID: "dock", Tabs: []string{"panel"}, Active: "panel"}}, []string{"panel"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixtureMountScope(a, "main"); err != nil {
		t.Fatal(err)
	}
	w := &Workbench{Owner: a, Window: "main", Workspace: model, Panels: []Panel{{ID: "panel", Title: "Panel", View: func() ui.Node {
		return ui.Node{ID: "row", Kind: ui.Row, Children: []ui.Node{{ID: "label", Kind: ui.Label, Text: "value"}}}
	}}}}
	theme := ui.Theme{Kinds: map[ui.Kind]ui.Style{ui.Row: {Gap: ui.Some(float32(8))}}}
	w.SetTheme(theme)
	theme.Kinds[ui.Row] = ui.Style{Gap: ui.Some(float32(99))}
	widget := w.Mount(core.NewBody("headless"))
	if err := w.Sync(widget); err != nil {
		t.Fatal(err)
	}
	r := w.records["panel"]
	r.widget.AsWidget().Style()
	if r.widget.AsWidget().Styles.Gap.X.Value != 8 {
		t.Fatal("panel theme missing or aliased")
	}
}

func TestReconcileNewThemeRestylesRetainedNodesAndPanels(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	host := core.NewFrame(core.NewBody("headless"))
	n := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "label", Kind: ui.Label, Text: "unchanged"}}}
	first := ui.Theme{Default: ui.Style{Typography: ui.Some(ui.Typography{Size: 14})}}
	second := ui.Theme{Default: ui.Style{Typography: ui.Some(ui.Typography{Size: 22})}}
	r, err := reconcile(host, nil, n, a, "main", first)
	if err != nil {
		t.Fatal(err)
	}
	widget := r.children[0].widget
	r, err = reconcile(host, r, n, a, "main", second)
	if err != nil {
		t.Fatal(err)
	}
	widget.AsWidget().Style()
	if r.children[0].widget != widget || widget.AsWidget().Styles.Font.Size.Value != 22 {
		t.Fatal("retained record kept stale theme")
	}
	model, err := workbench.New(workbench.Layout{Version: 1, Root: &workbench.Group{ID: "dock", Tabs: []string{"panel"}, Active: "panel"}}, []string{"panel"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixtureMountScope(a, "main"); err != nil {
		t.Fatal(err)
	}
	w := &Workbench{Owner: a, Window: "main", Workspace: model, Panels: []Panel{{ID: "panel", Title: "Panel", View: func() ui.Node { return ui.Node{ID: "label", Kind: ui.Label, Text: "value"} }}}}
	n = ui.Node{ID: "ext", Kind: ui.Extension, Custom: w}
	r, err = reconcile(host, r, n, a, "main", first)
	if err != nil {
		t.Fatal(err)
	}
	panel := w.records["panel"].widget
	r, err = reconcile(host, r, n, a, "main", second)
	if err != nil {
		t.Fatal(err)
	}
	panel.AsWidget().Style()
	if w.records["panel"].widget != panel || panel.AsWidget().Styles.Font.Size.Value != 22 {
		t.Fatal("theme extension kept stale theme")
	}
}

func TestSliderManagedSizingOverridesFinalDefaultsAndRestores(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	host := core.NewBody("headless")
	n := ui.Node{ID: "slider", Kind: ui.Slider, Numeric: ui.Numeric{Min: 0, Max: 255, Step: 1}, MinWidth: 160, Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 240})}}
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	widget := r.widget
	check := func(min float32, unit units.Units, max float32) {
		t.Helper()
		widget.AsWidget().Style()
		s := widget.AsWidget().Styles
		if s.Min.X.Value != min || s.Min.X.Unit != unit || s.Max.X.Value != max {
			t.Fatalf("managed sizing after native final style: min=%+v max=%+v", s.Min.X, s.Max.X)
		}
	}
	check(160, units.UnitDp, 240)
	n.MinWidth, n.Style = 0, ui.Style{}
	theme := ui.Theme{Kinds: map[ui.Kind]ui.Style{ui.Slider: {MinSize: ui.Some(ui.Size{Width: 110}), MaxSize: ui.Some(ui.Size{Width: 220})}}}
	r, err = reconcile(host, r, n, a, "main", theme)
	if err != nil {
		t.Fatal(err)
	}
	check(110, units.UnitDp, 220)
	r, err = reconcile(host, r, n, a, "main", ui.Theme{})
	if err != nil {
		t.Fatal(err)
	}
	check(20, units.UnitEm, 0)
	if r.widget != widget {
		t.Fatal("sizing changes replaced retained slider")
	}
}
