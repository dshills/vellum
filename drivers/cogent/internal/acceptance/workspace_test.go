package acceptance

import (
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/tasks"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"image/color"
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
	if first.Editor == second.Editor || first.Grid == second.Grid || first.Tree == second.Tree || first.Dock == second.Dock || first.Color == second.Color || first.Breadcrumb == second.Breadcrumb || first.Drawing == second.Drawing {
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
	for _, kind := range []ui.Kind{ui.Column, ui.Row, ui.Label, ui.Button, ui.Icon, ui.Image, ui.Extension, ui.Canvas2D, ui.Checkbox, ui.Switch, ui.ToggleButton, ui.ProgressBar, ui.ProgressRing, ui.ProgressGauge, ui.Separator, ui.MenuButton, ui.RadioGroup, ui.SegmentedControl, ui.Selector, ui.ListBox, ui.Slider, ui.NumberField, ui.TextField, ui.MultilineText, ui.DataGrid, ui.DataTree, ui.Tabs, ui.Scroll, ui.Toolbar, ui.StatusBar, ui.FormGroup, ui.Disclosure, ui.SplitPane, ui.Dialog} {
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

func TestComposedBreadcrumbRejectAcceptAndWindowIsolation(t *testing.T) {
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
	var nodes map[string]ui.Node
	collect := func() {
		nodes = map[string]ui.Node{}
		var walk func(ui.Node)
		walk = func(n ui.Node) {
			nodes[n.ID] = n
			for _, c := range n.Children {
				walk(c)
			}
		}
		walk(first.View())
	}
	collect()
	home := ui.BreadcrumbItemID("workspace-path", "home")
	models := ui.BreadcrumbItemID("workspace-path", "models")
	part := ui.BreadcrumbItemID("workspace-path", "part")
	if nodes[home].Kind != ui.Button || nodes[models].Kind != ui.Button || nodes[part].Kind != ui.Label || nodes[part].Accessibility.Help != "Current location" {
		t.Fatal("composed breadcrumb roles missing")
	}
	stale := nodes[models].Activate
	if err = a.Do(nodes[home].Activate); err != nil {
		t.Fatal(err)
	}
	if s := first.State(); s.Changes["workspace-path"] != 1 || s.BreadcrumbItems != 3 || s.BreadcrumbCurrent != "part" {
		t.Fatal("rejected navigation changed path or lacked callback", s)
	}
	if err = a.Do(nodes[models].Activate); err != nil {
		t.Fatal(err)
	}
	collect()
	if s := first.State(); s.Changes["workspace-path"] != 2 || s.BreadcrumbItems != 2 || s.BreadcrumbCurrent != "models" {
		t.Fatal("accepted path not committed", s)
	}
	if nodes[models].Kind != ui.Label || nodes[models].Activate != nil || nodes[models].Accessibility.Help != "Current location" {
		t.Fatal("accepted current location stayed actionable")
	}
	if _, present := nodes[part]; present {
		t.Fatal("old current location retained")
	}
	if err = a.Do(stale); err != nil {
		t.Fatal(err)
	}
	if first.State().Changes["workspace-path"] != 2 || second.State().Changes["workspace-path"] != 0 || second.State().BreadcrumbCurrent != "part" || second.State().BreadcrumbItems != 3 {
		t.Fatal("stale request accepted or another window mutated")
	}
}

func TestDisclosureUsesReservedWorkerAndJoinsRetirement(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	w, err := New(a, "worker", struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	executor, err := execution.NewManual(2)
	if err != nil {
		t.Fatal(err)
	}
	defer executor.Close()
	w.executor = executor
	if err = a.Open(app.Window{ID: "worker", Title: "worker", Width: 640, Height: 480, View: w.View}); err != nil {
		t.Fatal(err)
	}
	start := func(kind virtual.OutlineChangeKind) {
		t.Helper()
		var failure error
		if err = a.Do(func() { failure = w.disclose(virtual.OutlineChange{Kind: kind, ID: "n0"}) }); err != nil || failure != nil {
			t.Fatal(err, failure)
		}
		if w.workerError != nil {
			t.Fatal(w.workerError)
		}
	}
	// No builder runs before explicit worker selection; the old index stays live.
	var admissionError error
	if err = a.Do(func() { admissionError = w.disclose(virtual.OutlineChange{Kind: virtual.OutlineCollapse, ID: "n0"}) }); err != nil || admissionError != nil {
		t.Fatal(err, admissionError)
	}
	pending, _ := w.Tree.State()
	if !pending.Pending || pending.VisibleRows != 64 || w.State().Disclosures != 1 || w.State().WorkerActive != 1 {
		t.Fatal("builder ran on owner or credit missing", pending, w.State())
	}
	if err = a.Do(func() { admissionError = w.disclose(virtual.OutlineChange{Kind: virtual.OutlineExpand, ID: "n0"}) }); err != nil || !errors.Is(admissionError, tasks.ErrCapacity) {
		t.Fatal("capacity did not refuse before prepare", err, admissionError)
	}
	if _, err = executor.StepNext(context.Background()); err != nil {
		t.Fatal(err)
	}
	a.Drain()
	collapsed, _ := w.Tree.State()
	if collapsed.Pending || collapsed.VisibleRows != 60 || w.State().WorkerActive != 0 || w.State().Failure != nil {
		t.Fatal("accepted worker failed to complete/release", collapsed, w.State())
	}
	start(virtual.OutlineExpand)
	if _, err = executor.StepNext(context.Background()); err != nil {
		t.Fatal(err)
	}
	a.Drain()
	expanded, _ := w.Tree.State()
	if expanded.Pending || expanded.VisibleRows != 64 || w.State().WorkerActive != 0 || w.State().Disclosures != 2 {
		t.Fatal("expansion failed", expanded, w.State())
	}
	start(virtual.OutlineCollapse)
	if err = a.CloseWindow(currentFixtureRef(a, "worker")); err != nil {
		t.Fatal(err)
	}
	if _, err = executor.StepNext(context.Background()); err != nil {
		t.Fatal(err)
	}
	a.Drain()
	retired, _ := w.Tree.State()
	if !retired.Pending || retired.VisibleRows != 64 || w.State().WorkerActive != 0 {
		t.Fatal("retired worker committed or retained credit", retired, w.State())
	}
}

func TestRefusedPreparePreservesFirstFailureAcrossLaterSuccess(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	w, err := New(a, "failure", struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	executor, err := execution.NewManual(2)
	if err != nil {
		t.Fatal(err)
	}
	defer executor.Close()
	w.executor = executor
	if err = a.Open(app.Window{ID: "failure", Title: "failure", Width: 640, Height: 480, View: w.View}); err != nil {
		t.Fatal(err)
	}
	var refused error
	if err = a.Do(func() { refused = w.disclose(virtual.OutlineChange{Kind: virtual.OutlineExpand, ID: "missing"}) }); err != nil || !errors.Is(refused, virtual.ErrTreeRange) {
		t.Fatal(err, refused)
	}
	first := w.State().Failure
	if first != refused {
		t.Fatal("original refusal not retained", first, refused)
	}
	if _, err = executor.StepNext(context.Background()); err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if w.State().Failure != first || errors.Is(w.State().Failure, virtual.ErrTreeStale) || errors.Is(w.State().Failure, context.Canceled) || w.State().WorkerActive != 0 {
		t.Fatal("unprepared cancel or worker cancellation masked original refusal", w.State())
	}
	for _, kind := range []virtual.OutlineChangeKind{virtual.OutlineCollapse, virtual.OutlineExpand} {
		if err = a.Do(func() {
			w.models().Children[0].TreeDisclosure(ui.TreeDisclosureEvent{Change: virtual.OutlineChange{Kind: kind, ID: "n0"}})
		}); err != nil {
			t.Fatal(err)
		}
		if w.State().Failure != first {
			t.Fatal("successful callback cleared failure", w.State())
		}
		if _, err = executor.StepNext(context.Background()); err != nil {
			t.Fatal(err)
		}
		a.Drain()
		if w.State().Failure != first || w.State().WorkerActive != 0 {
			t.Fatal("successful completion cleared failure", w.State())
		}
	}
	state, _ := w.Tree.State()
	if state.Pending || state.VisibleRows != 64 {
		t.Fatal("later successful worker did not complete", state)
	}
}

func TestComposedColorModelAndQueuedChannelsAreIndependent(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	first, err := New(a, "first", struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := New(a, "second", struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	view := first.Color.View()
	var red, green ui.Node
	var find func(ui.Node)
	find = func(n ui.Node) {
		if n.ID == ui.ColorFieldID("color", ui.ColorRed) {
			red = n
		}
		if n.ID == ui.ColorFieldID("color", ui.ColorGreen) {
			green = n
		}
		for _, c := range n.Children {
			find(c)
		}
	}
	find(view)
	if err := a.Do(func() { red.Change(ui.ChangeEvent{Number: 12}); green.Change(ui.ChangeEvent{Number: 97}) }); err != nil {
		t.Fatal(err)
	}
	if first.State().Color != (color.NRGBA{R: 12, G: 97, B: 224, A: 255}) || second.State().Color != (color.NRGBA{R: 32, G: 96, B: 224, A: 255}) || first.State().Changes[red.ID] != 1 || first.State().Changes[green.ID] != 1 {
		t.Fatal("composed color shared/stale model")
	}
}
