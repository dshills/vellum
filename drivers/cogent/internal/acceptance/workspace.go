// Package acceptance builds a deterministic all-component workspace for native
// acceptance runners. Each instance owns fresh mutable models on one UI owner.
package acceptance

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"maps"
	"strconv"
	"strings"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/tasks"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"github.com/dshills/vellum/workbench"
)

const DenseLabels = 1000

type Workspace struct {
	Owner                       *app.Application
	Window                      string
	Editor                      *editor.Model
	Grid                        *virtual.Grid
	Tree                        *virtual.Outline
	Color                       *ui.ColorPicker
	Breadcrumb                  *ui.Breadcrumb
	Drawing                     *canvas.Scene
	listSelection               string
	Dock                        *cogent.Workbench
	checked, open, dialog       bool
	number                      float32
	text, choice, tab, action   string
	split                       float32
	picture                     *canvas.TiledImage
	canvas                      any
	actions                     int
	changes                     map[string]int
	treeSelections, disclosures int
	manager                     *tasks.Manager
	executor                    execution.Executor
	workerError                 error
}

func New(owner *app.Application, window string, gpuCanvas any) (*Workspace, error) {
	if owner == nil || window == "" || gpuCanvas == nil {
		return nil, fmt.Errorf("acceptance: owner, window and canvas required")
	}
	w := &Workspace{Owner: owner, Window: window, canvas: gpuCanvas, open: true, number: 25, text: "Native workspace", choice: "a", tab: "live", split: .5, changes: map[string]int{}}
	w.Color = &ui.ColorPicker{ID: "color", Label: "Workspace color", Color: color.NRGBA{R: 32, G: 96, B: 224, A: 255}}
	w.Color.Change = func(e ui.ColorEdit) {
		w.Color.Color = e.Apply(w.Color.Color)
		w.changes[ui.ColorFieldID(w.Color.ID, e.Channel)]++
	}
	w.Color.Invalid = w.recordFailure
	w.Breadcrumb = &ui.Breadcrumb{ID: "workspace-path", Label: "Workspace location", Items: []ui.BreadcrumbItem{{ID: "home", Label: "Workspace"}, {ID: "models", Label: "Models"}, {ID: "part", Label: "Current"}}}
	w.Breadcrumb.Navigate = func(e ui.BreadcrumbEvent) {
		w.changes[w.Breadcrumb.ID]++
		// The root request is deliberately rejected; the model ancestor is
		// accepted. The native runner must observe the callback in both cases.
		if e.ItemID == "models" {
			w.Breadcrumb.Items = append([]ui.BreadcrumbItem(nil), w.Breadcrumb.Items[:2]...)
		}
	}
	var err error
	w.Drawing, err = newSelectionScene()
	if err != nil {
		return nil, err
	}
	w.listSelection = "b"
	w.Editor, err = editor.New("Native document\nGo owner and shared history", editor.Config{MaxBytes: 4096, HistoryLimit: 16})
	if err != nil {
		return nil, err
	}
	columns := make([]virtual.GridColumn, 20)
	for i := range columns {
		columns[i] = virtual.GridColumn{ID: fmt.Sprintf("c%d", i), Title: fmt.Sprintf("C%d", i), Width: 44}
	}
	w.Grid, err = virtual.NewGrid(source{}, columns, virtual.GridConfig{RowHeight: 18, HeaderHeight: 20, OverscanRows: 1, OverscanColumns: 1}, virtual.GridViewport{Width: 700, Height: 200})
	if err != nil {
		return nil, err
	}
	nodes := make([]virtual.TreeNode, 64)
	for i := range nodes {
		nodes[i] = virtual.TreeNode{ID: fmt.Sprintf("n%d", i), Label: fmt.Sprintf("Node %02d", i)}
		if i >= 1 && i <= 4 {
			nodes[i].Parent = "n0"
		}
	}
	data, err := virtual.NewTreeData(context.Background(), "acceptance-tree:"+window, 1, nodes, virtual.TreeLimits{})
	if err != nil {
		return nil, err
	}
	index, err := virtual.BuildTreeIndex(context.Background(), data, virtual.TreeIndexOptions{ExpandAll: true})
	if err != nil {
		return nil, err
	}
	w.Tree, err = virtual.NewOutline(index, virtual.OutlineConfig{RowHeight: 18, MaxRows: 64}, virtual.OutlineViewport{Width: 140, Height: 200})
	if err != nil {
		return nil, err
	}
	pixels := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	for y := range 8 {
		for x := range 8 {
			pixels.SetNRGBA(x, y, color.NRGBA{R: 220, G: 160, B: 40, A: 255})
		}
	}
	w.picture, err = canvas.NewTiledImage(pixels)
	if err != nil {
		return nil, err
	}
	layout := workbench.Layout{Version: 1, Root: &workbench.Group{ID: "workspace", Axis: workbench.Horizontal, Ratio: .3, First: &workbench.Group{ID: "canvas-panel", Tabs: []string{"canvas"}, Active: "canvas"}, Second: &workbench.Group{ID: "models-panel", Tabs: []string{"models"}, Active: "models"}}}
	dock, err := workbench.New(layout, []string{"canvas", "models"})
	if err != nil {
		return nil, err
	}
	w.Dock = &cogent.Workbench{Owner: owner, Window: window, Workspace: dock, Panels: []cogent.Panel{
		{ID: "canvas", Title: "GPU canvas", View: func() ui.Node { return ui.Node{ID: "gpu", Kind: ui.Extension, Grow: true, Custom: w.canvas} }},
		{ID: "models", Title: "Grid, outline and document", View: w.models},
	}}
	return w, nil
}

func (w *Workspace) models() ui.Node {
	return ui.Node{ID: "models", Kind: ui.Row, Grow: true, Children: []ui.Node{
		{ID: "outline", Kind: ui.DataTree, Tree: w.Tree, TreeChange: func(e ui.TreeEvent) {
			if e.Reason == ui.TreeSelection {
				w.treeSelections++
			}
		}, TreeDisclosure: func(e ui.TreeDisclosureEvent) { w.recordFailure(w.disclose(e.Change)) }, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 130}), MaxSize: ui.Some(ui.Size{Width: 160})}},
		{ID: "model-split", Kind: ui.SplitPane, Grow: true, Orientation: ui.Horizontal, Number: w.split, Change: func(e ui.ChangeEvent) { w.split = e.Number }, Children: []ui.Node{
			{ID: "grid", Kind: ui.DataGrid, Grid: w.Grid, Grow: true},
			{ID: "document", Kind: ui.MultilineText, Editor: w.Editor, SoftWrap: true, Grow: true},
		}},
	}}
}

func (w *Workspace) View() ui.Node {
	choice := []ui.Option{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}, {ID: "disabled", Label: "Disabled", Disabled: true}}
	checked := func(e ui.ChangeEvent) { w.checked = e.Checked; w.changes[e.ID]++ }
	number := func(e ui.ChangeEvent) { w.number = e.Number; w.changes[e.ID]++ }
	choose := func(e ui.ChangeEvent) { w.choice = e.SelectedID; w.changes[e.ID]++ }
	compact := ui.Style{Padding: ui.Some(ui.UniformInsets(0)), Gap: ui.Some(float32(0))}
	breadcrumb, err := w.Breadcrumb.View()
	if err != nil {
		w.recordFailure(err)
	}
	controls := ui.Node{ID: "controls", Kind: ui.Column, Style: compact, Children: []ui.Node{
		breadcrumb,
		{ID: "toolbar", Kind: ui.Toolbar, Children: []ui.Node{
			{ID: "button", Kind: ui.Button, Text: "Action", Activate: func() { w.actions++ }},
			{ID: "check", Kind: ui.Checkbox, Text: "Check", Checked: w.checked, Change: checked},
			{ID: "switch", Kind: ui.Switch, Text: "Switch", Checked: w.checked, Change: checked},
			{ID: "toggle", Kind: ui.ToggleButton, Text: "Toggle", Checked: w.checked, Change: checked},
			{ID: "menu", Kind: ui.MenuButton, Text: "Actions", Options: choice, Change: func(e ui.ChangeEvent) { w.action = e.SelectedID; w.actions++ }},
			{ID: "selector", Kind: ui.Selector, SelectedID: w.choice, Options: choice, Change: choose},
			{ID: "dialog-open", Kind: ui.Button, Text: "Dialog", Activate: func() { w.dialog = true }},
		}},
		{ID: "form", Kind: ui.FormGroup, Text: "Controlled values", Children: []ui.Node{{ID: "form-row", Kind: ui.Row, Children: []ui.Node{
			{ID: "radio", Kind: ui.RadioGroup, Orientation: ui.Horizontal, SelectedID: w.choice, Options: choice, Change: choose},
			{ID: "segment", Kind: ui.SegmentedControl, SelectedID: w.choice, Options: choice, Change: choose},
			{ID: "slider", Kind: ui.Slider, Number: w.number, Numeric: ui.Numeric{Min: 0, Max: 100, Step: 1}, Change: number, MinWidth: 160},
			{ID: "number", Kind: ui.NumberField, Number: w.number, Numeric: ui.Numeric{Min: 0, Max: 100, Step: 1}, Change: number},
			{ID: "text", Kind: ui.TextField, Text: w.text, Change: func(e ui.ChangeEvent) { w.text = e.Text; w.changes[e.ID]++ }},
		}}}},
		{ID: "disclosure", Kind: ui.Disclosure, Text: "Indicators", Checked: w.open, Change: func(e ui.ChangeEvent) { w.open = e.Checked; w.changes[e.ID]++ }, Children: []ui.Node{{ID: "indicators", Kind: ui.Row, Children: []ui.Node{
			{ID: "icon", Kind: ui.Icon, Icon: ui.SymbolCheck, Text: "Check icon"},
			{ID: "image", Kind: ui.Image, Image: w.picture, Text: "Local amber square", Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 24, Height: 24}), MaxSize: ui.Some(ui.Size{Width: 24, Height: 24})}},
			{ID: "progress", Kind: ui.ProgressBar, Number: w.number / 100, MinWidth: 200},
			{ID: "progress-ring", Kind: ui.ProgressRing, Number: w.number / 100, Text: fmt.Sprintf("%.0f%%", w.number), Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 80, Height: 80}), MaxSize: ui.Some(ui.Size{Width: 80, Height: 80})}, Accessibility: ui.Accessibility{Name: "Circular progress"}},
			{ID: "progress-gauge", Kind: ui.ProgressGauge, Number: w.number / 100, Text: fmt.Sprintf("%.0f%%", w.number), Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 88, Height: 56}), MaxSize: ui.Some(ui.Size{Width: 88, Height: 56})}, Accessibility: ui.Accessibility{Name: "Semicircular progress"}},
			{ID: "separator", Kind: ui.Separator, Orientation: ui.Vertical},
			{ID: "status", Kind: ui.StatusBar, Text: fmt.Sprintf("Actions %d · %s", w.actions, w.action)},
		}}}},
	}}
	dense := ui.Node{ID: "dense", Kind: ui.Column, Style: compact}
	for row := range 10 {
		line := ui.Node{ID: fmt.Sprintf("dense-row-%02d", row), Kind: ui.Row, Style: compact}
		for column := range 100 {
			i := row*100 + column
			line.Children = append(line.Children, ui.Node{ID: fmt.Sprintf("dense-%04d", i), Kind: ui.Label, Text: fmt.Sprintf("%03d", i), Style: ui.Style{Typography: ui.Some(ui.Typography{Size: 9}), Padding: ui.Some(ui.UniformInsets(0)), MinSize: ui.Some(ui.Size{Width: 18, Height: 12}), MaxSize: ui.Some(ui.Size{Width: 18, Height: 14})}})
		}
		dense.Children = append(dense.Children, line)
	}
	return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Style: compact, Children: []ui.Node{
		{ID: "density-scroll", Kind: ui.Scroll, Grow: true, MinHeight: 160, Style: ui.Style{MinSize: ui.Some(ui.Size{Height: 160}), MaxSize: ui.Some(ui.Size{Height: 170})}, Children: []ui.Node{dense}},
		controls,
		{ID: "tabs", Kind: ui.Tabs, Grow: true, SelectedID: w.tab, Options: []ui.Option{{ID: "live", Label: "Workspace"}, {ID: "about", Label: "About"}, {ID: "colors", Label: "Colors"}, {ID: "selections", Label: "Selections"}}, Change: func(e ui.ChangeEvent) { w.tab = e.SelectedID; w.changes[e.ID]++ }, Children: []ui.Node{
			{ID: "live", Kind: ui.Extension, Grow: true, MinHeight: 240, Custom: w.Dock},
			{ID: "about", Kind: ui.Label, Text: "All-component acceptance workspace"},
			{ID: "colors", Kind: ui.Scroll, Grow: true, Children: []ui.Node{w.Color.View()}},
			w.selectionView(),
		}},
		{ID: "dialog", Kind: ui.Dialog, Text: "Acceptance dialog", Open: w.dialog, Dismiss: func() { w.dialog = false }, Children: []ui.Node{{ID: "dialog-label", Kind: ui.Label, Text: "Composed modal lifecycle"}, {ID: "dialog-close", Kind: ui.Button, Text: "Close", Activate: func() { w.dialog = false }}}},
	}}
}

// Exercise mutates each fresh model and verifies history replay on the owner.
// Native selectors and completed geometry/readback belong to the runner.
func (w *Workspace) Exercise() error {
	w.checked = true
	w.number = 75
	w.choice = "b"
	w.text = "Accepted composed edit"
	w.actions++
	const next = "Native document\nEdited composed workspace"
	if err := w.Editor.Replace(next, editor.Selection{}); err != nil {
		return err
	}
	if err := w.Editor.Journal().Undo(); err != nil {
		return err
	}
	if err := w.Editor.Journal().Redo(); err != nil {
		return err
	}
	if w.Editor.Snapshot().Text != next {
		return fmt.Errorf("acceptance: editor replay mismatch")
	}
	if err := w.Grid.Scroll(44, 18); err != nil {
		return err
	}
	return w.Tree.Select("n1")
}
func (w *Workspace) SetDialog(open bool) { w.dialog = open }
func (w *Workspace) Text() string        { return w.text }

// State is a detached owner-only observation, separate from native admission.
type ControlState struct {
	Checked, IndicatorsOpen, DialogOpen                bool
	Color                                              color.NRGBA
	Number                                             float32
	Choice, Tab                                        string
	ListSelection, DrawingSelection                    string
	BreadcrumbCurrent                                  string
	BreadcrumbItems                                    int
	Actions, TreeSelections, Disclosures, WorkerActive int
	Changes                                            map[string]int
	Failure                                            error
}

func (w *Workspace) State() ControlState {
	active := 0
	if w.manager != nil {
		active = w.manager.Pending().Active
	}
	return ControlState{Color: w.Color.Color, Checked: w.checked, IndicatorsOpen: w.open, DialogOpen: w.dialog, Number: w.number, Choice: w.choice, Tab: w.tab, ListSelection: w.listSelection, DrawingSelection: w.Drawing.Selected, BreadcrumbCurrent: w.Breadcrumb.Items[len(w.Breadcrumb.Items)-1].ID, BreadcrumbItems: len(w.Breadcrumb.Items), Actions: w.actions, TreeSelections: w.treeSelections, Disclosures: w.disclosures, WorkerActive: active, Changes: maps.Clone(w.changes), Failure: w.workerError}
}

func (w *Workspace) recordFailure(err error) {
	if err != nil && w.workerError == nil {
		w.workerError = err
	}
}

func (w *Workspace) disclose(change virtual.OutlineChange) error {
	if w.manager == nil {
		manager, err := tasks.NewWithConfig(tasks.Config{Owner: w.Owner, Window: w.Window, Limit: 1, Executor: w.executor, Invalidate: func() { _ = w.Owner.Invalidate(w.Window) }})
		if err != nil {
			return err
		}
		w.manager = manager
	}
	lifetime, err := w.Owner.WindowContext(w.Window)
	if err != nil {
		return err
	}
	// Reserve both worker and completion credit before marking the model pending.
	input := make(chan virtual.OutlineRequest, 1)
	var request virtual.OutlineRequest
	prepared := false
	var prepareError error
	task, err := tasks.Start(w.manager, lifetime, func(ctx context.Context, _ func(float64, string)) (*virtual.TreeIndex, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case captured := <-input:
			builder, err := virtual.NewOutlineBuilder(ctx, captured)
			if err != nil {
				return nil, err
			}
			for {
				done, err := builder.Step(64)
				if err != nil {
					return nil, err
				}
				if done {
					return builder.Result()
				}
			}
		}
	}, func(index *virtual.TreeIndex, err error) {
		if !prepared {
			w.recordFailure(prepareError)
			return
		}
		if err == nil {
			err = w.Tree.Complete(request, index)
		}
		if err != nil {
			err = errors.Join(err, w.Tree.Cancel(request))
		}
		w.recordFailure(err)
	})
	if err != nil {
		return err
	}
	request, prepareError = w.Tree.Prepare(change)
	if prepareError != nil {
		task.Cancel()
		w.recordFailure(prepareError)
		return prepareError
	}
	prepared = true
	w.disclosures++
	input <- request
	return nil
}

type source struct{}

func (source) Identity() string { return "acceptance-grid" }
func (source) Revision() uint64 { return 1 }
func (source) RowCount() int    { return 64 }
func (source) RowKey(i int) (string, error) {
	if i < 0 || i >= 64 {
		return "", virtual.ErrGridRange
	}
	return fmt.Sprintf("r%d", i), nil
}
func (source) IndexOf(key string) (int, bool, error) {
	if !strings.HasPrefix(key, "r") {
		return 0, false, nil
	}
	i, e := strconv.Atoi(key[1:])
	return i, e == nil && i >= 0 && i < 64 && fmt.Sprintf("r%d", i) == key, nil
}
func (s source) Cell(key, column string) (string, error) {
	_, ok, _ := s.IndexOf(key)
	if !ok {
		return "", virtual.ErrGridRange
	}
	i, e := strconv.Atoi(strings.TrimPrefix(column, "c"))
	if e != nil || i < 0 || i >= 20 || fmt.Sprintf("c%d", i) != column {
		return "", virtual.ErrGridRange
	}
	return key + "/" + column, nil
}
