package acceptance

import (
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"strings"
	"testing"
)

func TestWorkspaceCatalogAndFreshModelIsolation(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	first, err := New(a, "first", struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := New(a, "second", struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	if first.Editor == second.Editor || first.Grid == second.Grid || first.Tree == second.Tree || first.Dock == second.Dock {
		t.Fatal("per-window mutable models aliased")
	}
	kinds := map[ui.Kind]bool{}
	labels := 0
	var inspect func(ui.Node)
	inspect = func(n ui.Node) {
		kinds[n.Kind] = true
		if n.Kind == ui.Label && strings.HasPrefix(n.ID, "dense-") {
			labels++
		}
		for _, child := range n.Children {
			inspect(child)
		}
	}
	if err = ui.Validate(first.View()); err != nil {
		t.Fatal(err)
	}
	inspect(first.View())
	for _, panel := range first.Dock.Panels {
		view := panel.View()
		if err = ui.Validate(view); err != nil {
			t.Fatal(err)
		}
		inspect(view)
	}
	if labels != DenseLabels {
		t.Fatalf("dense fixture has %d labels", labels)
	}
	for _, kind := range []ui.Kind{ui.Column, ui.Row, ui.Label, ui.Button, ui.Icon, ui.Image, ui.Extension, ui.Checkbox, ui.Switch, ui.ToggleButton, ui.ProgressBar, ui.Separator, ui.MenuButton, ui.RadioGroup, ui.SegmentedControl, ui.Selector, ui.Slider, ui.NumberField, ui.TextField, ui.MultilineText, ui.DataGrid, ui.DataTree, ui.Tabs, ui.Scroll, ui.Toolbar, ui.StatusBar, ui.FormGroup, ui.Disclosure, ui.SplitPane, ui.Dialog} {
		if !kinds[kind] {
			t.Errorf("accepted component %s missing", kind)
		}
	}
	before := second.Editor.Snapshot()
	var exerciseErr error
	if err = a.Do(func() { exerciseErr = first.Exercise() }); err != nil || exerciseErr != nil {
		t.Fatal(err, exerciseErr)
	}
	if first.Editor.Snapshot().History.RedoCount != 0 || first.Editor.Snapshot().History.UndoCount != 1 {
		t.Fatal("history replay failed")
	}
	if second.Editor.Snapshot() != before || second.Text() == first.Text() {
		t.Fatal("exercise mutated another workspace")
	}
}
