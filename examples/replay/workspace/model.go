package workspace

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/placement"
	"github.com/dshills/vellum/tasks"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"github.com/dshills/vellum/workbench"
)

const Width = 1600
const Height = 1000

var ErrUnavailable = errors.New("workspace: action unavailable")

type jobResult struct {
	kind       string
	data       *source
	tree       *virtual.TreeData
	index      *virtual.TreeIndex
	projection virtual.GridProjection
}
type outcome struct{ Kind, Result string }
type Fixture struct {
	pendingPlacement                *placement.Request[string]
	canceled                        bool
	Owner                           *app.Application
	Commands                        *command.Registry
	Grid                            *virtual.Grid
	Tree                            *virtual.Outline
	Editor                          *editor.Model
	Canvas                          *canvas.Scene
	Dock                            *cogent.Workbench
	Palette                         *placement.Controller[string]
	Target                          *placement.Target
	Journal                         *history.Journal
	rows                            int
	data                            *source
	treeData                        *virtual.TreeData
	projection                      virtual.GridProjection
	manager                         *tasks.Manager
	task                            *tasks.Task
	job                             string
	token                           uint64
	gridRequest                     virtual.GridProjectionRequest
	treeRequest                     virtual.OutlineRequest
	loaded, sorted, filtered, modal bool
	name, status, failure           string
	argument                        string
	invoking                        bool
	life                            context.Context
	outcomes                        []outcome
	nextObject                      uint64
	attempts                        map[string]int
}

func New(a *app.Application, ex execution.Executor, rows int) (*Fixture, error) {
	var err error
	if a == nil || ex == nil || (rows != 10_000 && rows != MaxRows) {
		return nil, ErrUnavailable
	}
	f := &Fixture{Owner: a, Commands: command.New(), rows: rows, data: &source{}, name: "North Star", status: "No project", attempts: map[string]int{}}
	failed := true
	defer func() {
		if failed {
			if f.Palette != nil {
				f.Palette.Close()
			}
			a.Close()
		}
	}()
	f.Journal, err = history.New(64)
	if err != nil {
		return nil, err
	}
	columns := make([]virtual.GridColumn, Columns)
	for i := range columns {
		columns[i] = virtual.GridColumn{ID: columnID(i), Title: fmt.Sprintf("V%02d", i), Width: 64}
	}
	f.Grid, err = virtual.NewGrid(f.data, columns, virtual.GridConfig{RowHeight: 24, HeaderHeight: 28, FrozenColumns: 1, OverscanRows: 1, OverscanColumns: 1, MaxRows: 128, MaxCells: 4096, RowSelection: true}, virtual.GridViewport{Width: 900, Height: 600})
	if err != nil {
		return nil, err
	}
	empty, err := virtual.NewTreeData(context.Background(), "workspace:stored-tree", 1, nil, virtual.TreeLimits{})
	if err != nil {
		return nil, err
	}
	index, err := virtual.BuildTreeIndex(context.Background(), empty, virtual.TreeIndexOptions{})
	if err != nil {
		return nil, err
	}
	f.Tree, err = virtual.NewOutline(index, virtual.OutlineConfig{RowHeight: 24, MaxRows: 128}, virtual.OutlineViewport{Width: 220, Height: 700})
	if err != nil {
		return nil, err
	}
	f.Editor, err = editor.New("Workspace notes\n", editor.Config{MaxBytes: 65536, HistoryLimit: 32})
	if err != nil {
		return nil, err
	}
	f.Canvas, err = canvas.New(nil)
	if err != nil {
		return nil, err
	}
	f.Palette, err = placement.New(placement.Config[string]{Items: []placement.Item[string]{{ID: "block", Label: "Block", Payload: "block", Preview: canvas.Rect{Width: 80, Height: 50, Color: color.RGBA{45, 115, 210, 255}}}}, Command: func(ctx context.Context, r placement.Request[string]) error {
		f.pendingPlacement = &r
		defer func() { f.pendingPlacement = nil }()
		return f.Invoke(ctx, "place", "")
	}})
	if err != nil {
		return nil, err
	}
	f.Target, err = f.Palette.NewTarget(placement.TargetConfig[string]{ID: "drawing", Snap: func(r placement.Request[string]) (canvas.Point, error) {
		return canvas.Point{X: math.Round(r.Position.X/20) * 20, Y: math.Round(r.Position.Y/20) * 20}, nil
	}, Accept: func(placement.Request[string]) error {
		if !f.loaded || f.modal {
			return placement.ErrRejected
		}
		return nil
	}})
	if err != nil {
		return nil, err
	}
	if err = f.register(); err != nil {
		return nil, err
	}
	layout, err := workbench.New(workbench.Layout{Version: workbench.Version, Root: &workbench.Group{ID: "outer", Axis: workbench.Horizontal, Ratio: .15, First: &workbench.Group{ID: "navigation", Tabs: []string{"tree"}, Active: "tree"}, Second: &workbench.Group{ID: "inner", Axis: workbench.Horizontal, Ratio: .82, First: &workbench.Group{ID: "documents", Tabs: []string{"canvas", "grid", "notes"}, Active: "canvas"}, Second: &workbench.Group{ID: "properties", Tabs: []string{"properties"}, Active: "properties"}}}}, []string{"tree", "canvas", "grid", "notes", "properties"})
	if err != nil {
		return nil, err
	}
	f.Dock = &cogent.Workbench{Owner: a, Window: "workspace", Workspace: layout, OnError: func(e error) { f.report(e) }}
	for _, id := range []string{"tree", "canvas", "grid", "notes", "properties"} {
		id := id
		f.Dock.Panels = append(f.Dock.Panels, cogent.Panel{ID: id, Title: map[string]string{"tree": "Outline", "canvas": "Canvas", "grid": "Table", "notes": "Notes", "properties": "Properties"}[id], View: func() ui.Node { return f.panel(id) }})
	}
	if err = a.Open(app.Window{ID: "workspace", Title: "Vellum professional workspace", Width: Width, Height: Height, View: f.View}); err != nil {
		return nil, err
	}
	if err = a.Open(app.Window{ID: "inspector", Title: "Shared project inspector", Width: 500, Height: 320, View: f.Inspector}); err != nil {
		return nil, err
	}
	f.life, err = a.WindowContext("workspace")
	if err != nil {
		return nil, err
	}
	f.manager, err = tasks.NewWithConfig(tasks.Config{Owner: a, Window: "workspace", Executor: ex, Limit: 1, Invalidate: f.touch})
	if err != nil {
		return nil, err
	}
	failed = false
	return f, nil
}
func (f *Fixture) touch() { _ = f.Owner.Invalidate("workspace"); _ = f.Owner.Invalidate("inspector") }
func (f *Fixture) report(e error) {
	if e != nil {
		f.failure = e.Error()
		f.status = e.Error()
	}
	f.touch()
}
func (f *Fixture) busy() bool { return f.task != nil && !f.task.Snapshot().OwnerDone }
func (f *Fixture) available(id string) error {
	if f.life != nil && f.life.Err() != nil {
		return app.ErrWindowClosed
	}
	if f.modal && id != "close-modal" {
		return ErrUnavailable
	}
	return f.baseAvailable(id)
}
func (f *Fixture) baseAvailable(id string) error {
	switch id {
	case "open":
		if f.loaded || f.busy() {
			return ErrUnavailable
		}
	case "cancel":
		if f.canceled || !f.busy() {
			return ErrUnavailable
		}
	case "undo":
		if f.Journal.Snapshot().UndoCount == 0 {
			return history.ErrEmpty
		}
	case "redo":
		if f.Journal.Snapshot().RedoCount == 0 {
			return history.ErrEmpty
		}
	case "close-modal":
		if !f.modal {
			return ErrUnavailable
		}
	default:
		if !f.loaded {
			return ErrUnavailable
		}
		if (id == "sort" || id == "filter" || id == "tree-filter") && f.busy() {
			return tasks.ErrCapacity
		}
	}
	return nil
}
func (f *Fixture) register() error {
	definitions := []command.Command{
		{ID: "place", Label: "Place block", Execute: func(ctx context.Context) error { return f.place(ctx) }},
		{ID: "open", Label: "Open project", Execute: func(ctx context.Context) error { return f.open(ctx) }},
		{ID: "name", Label: "Project name", Execute: func(context.Context) error { return f.editName(f.argument) }},
		{ID: "sort", Label: "Sort table", Execute: func(ctx context.Context) error { return f.query(ctx, true, f.filtered) }},
		{ID: "filter", Label: "Filter table", Execute: func(ctx context.Context) error { return f.query(ctx, f.sorted, !f.filtered) }},
		{ID: "tree-filter", Label: "Filter outline", Execute: func(ctx context.Context) error {
			return f.changeTree(ctx, virtual.OutlineChange{Kind: virtual.OutlineFilter, Filter: "Component 000100"})
		}},
		{ID: "cancel", Label: "Cancel worker", Execute: func(context.Context) error { f.cancel(); return nil }},
		{ID: "undo", Label: "Undo", Shortcut: "Meta+Z", ShortcutScope: command.AfterFocusShortcuts, Execute: func(context.Context) error { e := f.Journal.Undo(); f.touch(); return e }},
		{ID: "redo", Label: "Redo", Shortcut: "Meta+Shift+Z", ShortcutScope: command.AfterFocusShortcuts, Execute: func(context.Context) error { e := f.Journal.Redo(); f.touch(); return e }},
		{ID: "modal", Label: "Open modal", Execute: func(context.Context) error { f.modal = true; f.touch(); return nil }},
		{ID: "close-modal", Label: "Close modal", Execute: func(context.Context) error { f.modal = false; f.touch(); return nil }},
		{ID: "close-main", Label: "Close main window", Execute: func(context.Context) error { return f.Owner.CloseWindow("workspace") }},
	}
	for _, c := range definitions {
		id := c.ID
		c.Available = func() error { return f.available(id) }
		if e := f.Commands.Register(c); e != nil {
			return e
		}
	}
	return nil
}
func (f *Fixture) Invoke(ctx context.Context, id, value string) error {
	if f.invoking {
		return ErrUnavailable
	}
	if id != "name" && value != "" {
		return ErrUnavailable
	}
	f.invoking = true
	f.argument = value
	defer func() { f.invoking = false; f.argument = "" }()
	return f.Commands.Invoke(ctx, id)
}
func (f *Fixture) uiAction(id, value string) {
	f.attempts[id]++
	if e := f.Invoke(f.life, id, value); e != nil {
		f.report(e)
	}
}
func (f *Fixture) editName(value string) error {
	if value == "" || len(value) > 128 || !utf8.ValidString(value) || strings.ContainsRune(value, 0) {
		return ErrUnavailable
	}
	if value == f.name {
		return nil
	}
	before := f.name
	e := f.Journal.Apply(history.Change{Label: "Project name", Apply: func() error { f.name = value; return nil }, Revert: func() error { f.name = before; return nil }})
	f.touch()
	return e
}
func (f *Fixture) start(ctx context.Context, kind string, work func(context.Context) (jobResult, error), finish func(jobResult) error) error {
	if f.busy() || f.token == math.MaxUint64 {
		return tasks.ErrCapacity
	}
	token := f.token + 1
	task, e := tasks.Start(f.manager, ctx, func(ctx context.Context, _ func(float64, string)) (jobResult, error) {
		if e := ctx.Err(); e != nil {
			return jobResult{kind: kind}, e
		}
		return work(ctx)
	}, func(result jobResult, e error) {
		defer f.retireRequests()
		if f.canceled || token != f.token || errors.Is(e, context.Canceled) {
			f.recordOutcome(outcome{kind, "canceled"})
			f.status = kind + " canceled"
			f.touch()
			return
		}
		if e == nil {
			e = finish(result)
		}
		if e != nil {
			f.recordOutcome(outcome{kind, "failed"})
			f.report(e)
			return
		}
		f.recordOutcome(outcome{kind, "complete"})
		f.status = kind + " complete"
		f.touch()
	})
	if e != nil {
		return e
	}
	f.task = task
	f.canceled = false
	f.token = token
	f.job = kind
	f.status = kind + " pending"
	f.touch()
	return nil
}
func (f *Fixture) open(ctx context.Context) error {
	count := f.rows
	return f.start(ctx, "open", func(ctx context.Context) (jobResult, error) {
		data, tree, index, e := buildData(ctx, count)
		if e != nil {
			return jobResult{}, e
		}
		projection, e := virtual.BuildGridProjection(ctx, data, 2, virtual.GridQuery{}, projectionLimits())
		return jobResult{kind: "open", data: data, tree: tree, index: index, projection: projection}, e
	}, func(r jobResult) error {
		tree, e := virtual.NewOutline(r.index, virtual.OutlineConfig{RowHeight: 24, MaxRows: 128}, virtual.OutlineViewport{Width: 220, Height: 700})
		if e != nil {
			return e
		}
		return f.Journal.Replace(true, func() error {
			if e := f.Grid.ReplaceSource(r.projection); e != nil {
				return e
			}
			f.data = r.data
			f.treeData = r.tree
			f.Tree = tree
			f.projection = r.projection
			f.loaded = true
			return nil
		})
	})
}
func (f *Fixture) query(ctx context.Context, sorted, filtered bool) error {
	request, e := f.Grid.PrepareProjection()
	if e != nil {
		return e
	}
	q := virtual.GridQuery{}
	if sorted {
		q.Sort = []virtual.GridSortColumn{{ColumnID: "c00", Kind: virtual.GridSortInteger, Descending: true}}
	}
	if filtered {
		q.Filters = []virtual.GridFilter{{ColumnID: "c01", Kind: virtual.GridFilterContains, Text: "7"}}
	}
	data := f.data
	e = f.start(ctx, "query", func(ctx context.Context) (jobResult, error) {
		p, e := virtual.BuildGridProjection(ctx, data, request.OutputRevision(), q, projectionLimits())
		return jobResult{kind: "query", projection: p}, e
	}, func(r jobResult) error {
		if f.data != data {
			return virtual.ErrGridStale
		}
		if e := f.Grid.CompleteProjection(request, r.projection); e != nil {
			return e
		}
		f.projection = r.projection
		f.sorted = sorted
		f.filtered = filtered
		return nil
	})
	if e != nil {
		_ = f.Grid.CancelProjection(request)
		return e
	}
	f.gridRequest = request
	return nil
}
func (f *Fixture) changeTree(ctx context.Context, change virtual.OutlineChange) error {
	if f.busy() {
		return tasks.ErrCapacity
	}
	request, e := f.Tree.Prepare(change)
	if e != nil {
		return e
	}
	tree := f.Tree
	e = f.start(ctx, "tree", func(ctx context.Context) (jobResult, error) {
		b, e := virtual.NewOutlineBuilder(ctx, request)
		if e != nil {
			return jobResult{}, e
		}
		for {
			done, e := b.Step(4096)
			if e != nil {
				return jobResult{}, e
			}
			if done {
				index, e := b.Result()
				return jobResult{kind: "tree", index: index}, e
			}
		}
	}, func(r jobResult) error {
		if f.Tree != tree {
			return virtual.ErrTreeStale
		}
		return tree.Complete(request, r.index)
	})
	if e != nil {
		_ = tree.Cancel(request)
		return e
	}
	f.treeRequest = request
	return nil
}
func (f *Fixture) cancel() {
	if !f.busy() {
		return
	}
	f.task.Cancel()
	f.canceled = true
	if f.token < math.MaxUint64 {
		f.token++
	}
	f.retireRequests()
	f.status = "Cancellation requested"
	f.touch()
}

// There is one admitted task until its owner callback has returned. Retire its
// exact capability even on parent cancellation, worker failure or finish refusal.
func (f *Fixture) retireRequests() {
	_ = f.Grid.CancelProjection(f.gridRequest)
	_ = f.Tree.Cancel(f.treeRequest)
	f.gridRequest = virtual.GridProjectionRequest{}
	f.treeRequest = virtual.OutlineRequest{}
}

func (f *Fixture) place(ctx context.Context) error {
	r := f.pendingPlacement
	if r == nil {
		return ErrUnavailable
	}
	if !f.loaded || f.modal || len(f.Canvas.Objects) >= 256 || f.nextObject == math.MaxUint64 {
		return ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	object := canvas.Rect{ID: fmt.Sprintf("block-%d", f.nextObject+1), X: r.Position.X, Y: r.Position.Y, Width: 80, Height: 50, Color: color.RGBA{45, 115, 210, 255}}
	before := append([]canvas.Rect(nil), f.Canvas.Objects...)
	after := append(append([]canvas.Rect(nil), before...), object)
	selection := f.Canvas.Selected
	err := f.Journal.Apply(history.Change{Label: "Place block", Apply: func() error {
		f.Canvas.Objects = append([]canvas.Rect(nil), after...)
		f.Canvas.Selected = object.ID
		return nil
	}, Revert: func() error {
		f.Canvas.Objects = append([]canvas.Rect(nil), before...)
		f.Canvas.Selected = selection
		return nil
	}})
	if err == nil {
		f.nextObject++
		f.touch()
	}
	return err
}

func (f *Fixture) recordOutcome(o outcome) {
	if len(f.outcomes) == 32 {
		copy(f.outcomes, f.outcomes[1:])
		f.outcomes = f.outcomes[:31]
	}
	f.outcomes = append(f.outcomes, o)
}

// WaitWorker joins the sole admitted task after the application's driver has
// stopped and Owner.Close has canceled its windows. A manual-executor caller
// must explicitly select pending work first; this method never selects work.
func (f *Fixture) WaitWorker(ctx context.Context) error {
	if ctx == nil {
		return ErrUnavailable
	}
	if f.Owner.Context().Err() == nil {
		return ErrUnavailable
	}
	if f.task == nil {
		return nil
	}
	select {
	case <-f.task.Done():
		f.retireRequests()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
