package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/tasks"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

var errBusy = errors.New("loadedgrid: work already pending")
var errStale = errors.New("loadedgrid: worker result superseded")

type window struct {
	p                     *program
	id                    string
	windowRef             app.WindowRef
	grid                  *virtual.Grid
	domain                *fullSource
	view                  virtual.GridProjection
	manager               *tasks.Manager
	query                 virtual.GridQuery
	queryToken, editToken uint64
	queryTask, editTask   *tasks.Task
	jobs                  map[*tasks.Task]string
	queryRequest          virtual.GridProjectionRequest
	filterText, message   string
	settling              bool
	settleCancel          context.CancelFunc
	settle                func(context.Context, func(error)) error
	filterToken           uint64
	settlingFilter        bool
	filterCancel          context.CancelFunc
	settleFilter          func(context.Context, func(error)) error
}
type program struct {
	owner       *app.Application
	windows     [2]*window
	rows        int
	variable    bool
	executor    execution.Executor
	loadManager *tasks.Manager
	loadTask    *tasks.Task
	loadToken   uint64
	base        *data
}

func gridColumns() []virtual.GridColumn {
	out := make([]virtual.GridColumn, columns)
	for i := range out {
		out[i] = virtual.GridColumn{ID: column(i), Title: fmt.Sprintf("Value %02d", i), Width: 112}
	}
	return out
}
func newProgram(rows int, variable bool, executor execution.Executor) (*program, error) {
	if rows < 1 || rows > maxRows {
		return nil, errData
	}
	a, e := app.New(app.Config{QueueCapacity: 64, CompletionCapacity: 8})
	if e != nil {
		return nil, e
	}
	p := &program{owner: a, rows: rows, variable: variable, executor: executor}
	for i, id := range []string{"small", "large"} {
		w := &window{p: p, id: id, domain: emptySource("stored:" + id), jobs: map[*tasks.Task]string{}, message: "Empty; Load records starts a bounded worker"}
		w.grid, e = virtual.NewGrid(input(w.domain, variable), gridColumns(), virtual.GridConfig{RowHeight: 24, HeaderHeight: 28, FrozenColumns: 2, OverscanRows: 2, OverscanColumns: 1, MaxRows: 256, MaxCells: 4096, RowSelection: true, CellEditing: true, MaxEditBytes: maxEditBytes}, virtual.GridViewport{Width: 960, Height: 520})
		if e != nil {
			a.Close()
			return nil, e
		}
		p.windows[i] = w
		if w.windowRef, e = a.OpenWindow(app.Window{ID: id, Title: "Vellum Loaded Grid · " + id, Width: 1100, Height: 860, View: w.describe}); e != nil {
			a.Close()
			return nil, e
		}
		w.manager, e = tasks.NewWithConfig(tasks.Config{Owner: a, Window: id, Limit: 2, Executor: executor, Invalidate: func() { _ = a.Invalidate(id) }})
		if e != nil {
			a.Close()
			return nil, e
		}
	}
	p.loadManager, e = tasks.NewWithConfig(tasks.Config{Owner: a, Limit: 1, Executor: executor, Invalidate: func() {
		for _, w := range p.windows {
			_ = a.Invalidate(w.id)
		}
	}})
	if e != nil {
		a.Close()
		return nil, e
	}
	return p, nil
}
func (w *window) lifetime() (context.Context, error) {
	ctx := w.windowRef.Context()
	if ctx.Err() != nil {
		return nil, app.ErrWindowClosed
	}
	return ctx, nil
}
func (w *window) report(e error) {
	if e != nil {
		w.message = e.Error()
	}
	_ = w.p.owner.Invalidate(w.id)
}
func cloneQuery(q virtual.GridQuery) virtual.GridQuery {
	out := virtual.GridQuery{Sort: append([]virtual.GridSortColumn(nil), q.Sort...), Filters: append([]virtual.GridFilter(nil), q.Filters...)}
	for i := range out.Sort {
		out.Sort[i].ColumnID = strings.Clone(out.Sort[i].ColumnID)
	}
	for i := range out.Filters {
		out.Filters[i].ColumnID = strings.Clone(out.Filters[i].ColumnID)
		out.Filters[i].Text = strings.Clone(out.Filters[i].Text)
	}
	return out
}
func (p *program) load() error {
	if p.base != nil || p.loadTask != nil {
		return errBusy
	}
	if p.loadToken == math.MaxUint64 {
		return virtual.ErrGridExhausted
	}
	p.loadToken++
	token := p.loadToken
	ids := [2]string{p.windows[0].domain.Identity(), p.windows[1].domain.Identity()}
	variable, rows := p.variable, p.rows
	task, e := tasks.Start(p.loadManager, p.owner.Context(), func(ctx context.Context, progress func(float64, string)) (loadResult, error) {
		d, e := generate(ctx, rows, progress)
		if e != nil {
			return loadResult{}, e
		}
		out := loadResult{base: d}
		for i := range out.sources {
			count := rows
			if i == 0 {
				count = min(count, 10_000)
			}
			s, e := source(ids[i], d, count)
			if e != nil {
				return loadResult{}, e
			}
			out.sources[i] = s
			out.views[i], e = virtual.BuildGridProjection(ctx, input(s, variable), 2, virtual.GridQuery{}, limits())
			if e != nil {
				return loadResult{}, e
			}
		}
		return out, nil
	}, func(out loadResult, e error) {
		if token != p.loadToken {
			return
		}
		p.loadTask = nil
		if e != nil {
			for _, w := range p.windows {
				w.report(e)
			}
			return
		}
		accepted := false
		for i, w := range p.windows {
			if _, e := w.lifetime(); e != nil {
				continue
			}
			if e := w.grid.ReplaceSource(out.views[i]); e != nil {
				w.report(e)
				continue
			}
			w.domain, w.view = out.sources[i], out.views[i]
			accepted = true
			w.message = fmt.Sprintf("Loaded %d stored rows; domain v%d / view v%d", w.domain.rows, w.domain.Revision(), w.view.Revision())
			w.report(nil)
		}
		if accepted {
			p.base = out.base
		}
	})
	if e != nil {
		return e
	}
	p.loadTask = task
	for _, w := range p.windows {
		w.message = "Loading stored records and full immutable indices"
		w.report(nil)
	}
	return nil
}
func (p *program) cancelLoad() error {
	if p.loadToken == math.MaxUint64 {
		return virtual.ErrGridExhausted
	}
	p.loadToken++
	if p.loadTask != nil {
		p.loadTask.Cancel()
		p.loadTask = nil
	}
	for _, w := range p.windows {
		w.message = "Load canceled; accepted worker still owns its credit until terminal"
		w.report(nil)
	}
	return nil
}
func (w *window) queryWork(q virtual.GridQuery) error {
	ctx, e := w.lifetime()
	if e != nil {
		return e
	}
	if w.view == nil {
		return errData
	}
	if len(q.Sort) > 4 || len(q.Filters) > 4 {
		return virtual.ErrGridInvalid
	}
	queryBytes := 0
	for _, sort := range q.Sort {
		if len(sort.ColumnID) > 4096-queryBytes {
			return virtual.ErrGridCapacity
		}
		queryBytes += len(sort.ColumnID)
	}
	for _, filter := range q.Filters {
		if len(filter.ColumnID) > 4096-queryBytes {
			return virtual.ErrGridCapacity
		}
		queryBytes += len(filter.ColumnID)
		if len(filter.Text) > 4096-queryBytes {
			return virtual.ErrGridCapacity
		}
		queryBytes += len(filter.Text)
	}
	if queryBytes > 4096 {
		return virtual.ErrGridCapacity
	}
	if w.manager.Pending().Active >= 2 {
		return tasks.ErrCapacity
	}
	if w.queryToken == math.MaxUint64 {
		return virtual.ErrGridExhausted
	}
	state, e := w.grid.State()
	if e != nil {
		return e
	}
	if state.Stamp.Revision == math.MaxUint64 {
		return virtual.ErrGridExhausted
	}
	token, revision := w.queryToken+1, state.Stamp.Revision+1
	domain := w.domain
	q = cloneQuery(q)
	variable := w.p.variable
	var req virtual.GridProjectionRequest
	var task *tasks.Task
	// Submit never executes inline, and the owner callback cannot run until this
	// owner turn ends. Work reads only captured immutable input/revision; req is
	// assigned only on owner before any possible owner completion delivery.
	task, e = tasks.Start(w.manager, ctx, func(ctx context.Context, _ func(float64, string)) (virtual.GridProjection, error) {
		return virtual.BuildGridProjection(ctx, input(domain, variable), revision, q, limits())
	}, func(result virtual.GridProjection, e error) {
		delete(w.jobs, task)
		if token != w.queryToken {
			return
		}
		w.queryTask = nil
		if w.domain != domain {
			_ = w.grid.CancelProjection(req)
			w.report(errStale)
			return
		}
		if e != nil {
			_ = w.grid.CancelProjection(req)
			w.report(e)
			return
		}
		if e = w.grid.CompleteProjection(req, result); e != nil {
			w.report(e)
			return
		}
		w.view = result
		w.query = cloneQuery(q)
		w.message = fmt.Sprintf("Projection accepted: %d rows, view v%d, domain v%d", result.RowCount(), result.Revision(), domain.Revision())
		w.report(nil)
	})
	if e != nil {
		return e
	}
	w.jobs[task] = "projection"
	req, e = w.grid.PrepareProjection()
	if e != nil {
		task.Cancel()
		return e
	}
	if req.OutputRevision() != revision {
		task.Cancel()
		_ = w.grid.CancelProjection(req)
		return errStale
	}
	if w.queryTask != nil {
		w.queryTask.Cancel()
	}
	w.queryToken = token
	w.queryTask = task
	w.queryRequest = req
	w.message = "Projection worker admitted"
	return nil
}

// applyFilter reads the controlled field only after its originating native FIFO
// acknowledges prior input and seals composition. Admission is asynchronous;
// cancellation, source/query changes, or a retired field refuse publication.
func (w *window) applyFilter() error {
	if w.settlingFilter {
		return errBusy
	}
	if w.view == nil || w.settleFilter == nil {
		return errData
	}
	ctx, err := w.lifetime()
	if err != nil {
		return err
	}
	if w.filterToken == math.MaxUint64 {
		return virtual.ErrGridExhausted
	}
	w.filterToken++
	token, queryToken, domain := w.filterToken, w.queryToken, w.domain
	ctx, cancel := context.WithCancel(ctx)
	w.settlingFilter, w.filterCancel = true, cancel
	err = w.settleFilter(ctx, func(err error) {
		if token != w.filterToken || !w.settlingFilter {
			return
		}
		w.settlingFilter, w.filterCancel = false, nil
		cancel()
		if err != nil {
			w.report(err)
			return
		}
		if w.domain != domain || w.queryToken != queryToken {
			w.report(errStale)
			return
		}
		q := cloneQuery(w.query)
		q.Filters = []virtual.GridFilter{{ColumnID: column(2), Kind: virtual.GridFilterContains, Text: w.filterText}}
		w.report(w.queryWork(q))
	})
	if err != nil {
		w.settlingFilter, w.filterCancel = false, nil
		cancel()
		return err
	}
	w.message = "Settling the originating filter field"
	return nil
}

func (w *window) cancelWork() error {
	if w.view == nil {
		return w.p.cancelLoad()
	}
	if w.queryToken == math.MaxUint64 || w.editToken == math.MaxUint64 || w.filterToken == math.MaxUint64 {
		return virtual.ErrGridExhausted
	}
	w.queryToken++
	w.editToken++
	w.filterToken++
	if w.filterCancel != nil {
		w.filterCancel()
		w.filterCancel = nil
	}
	w.settlingFilter = false
	if w.queryTask != nil {
		w.queryTask.Cancel()
		w.queryTask = nil
		_ = w.grid.CancelProjection(w.queryRequest)
	}
	if w.editTask != nil {
		w.editTask.Cancel()
		w.editTask = nil
	}
	if w.settleCancel != nil {
		w.settleCancel()
		w.settleCancel = nil
	}
	w.settling = false
	w.message = "Canceled work; credits remain until worker and owner terminals"
	return nil
}
func (w *window) begin() error { return w.grid.BeginActiveEdit() }
func (w *window) cancelEdit() error {
	if w.editToken == math.MaxUint64 {
		return virtual.ErrGridExhausted
	}
	w.editToken++
	if w.editTask != nil {
		w.editTask.Cancel()
		w.editTask = nil
	}
	if w.settleCancel != nil {
		w.settleCancel()
		w.settleCancel = nil
	}
	w.settling = false
	return w.grid.CancelEdit()
}
func (w *window) save() error {
	if w.settling || w.editTask != nil {
		return errBusy
	}
	edit, e := w.grid.Edit()
	if e != nil {
		return e
	}
	if !edit.Active {
		return virtual.ErrGridRange
	}
	if w.settle == nil {
		return errData
	}
	ctx, e := w.lifetime()
	if e != nil {
		return e
	}
	if w.editToken == math.MaxUint64 {
		return virtual.ErrGridExhausted
	}
	w.editToken++
	token := w.editToken
	session := edit.Session
	ctx, cancel := context.WithCancel(ctx)
	w.settling = true
	w.settleCancel = cancel
	e = w.settle(ctx, func(e error) {
		if token != w.editToken || !w.settling {
			return
		}
		w.settling = false
		w.settleCancel = nil
		cancel()
		if e != nil {
			w.report(e)
			return
		}
		current, e := w.grid.Edit()
		if e != nil || !current.Active || current.Session != session {
			w.report(errStale)
			return
		}
		w.report(w.editWork(current))
	})
	if e != nil {
		w.settling = false
		w.settleCancel = nil
		cancel()
		return e
	}
	w.message = "Settling the originating cell editor"
	return nil
}

type editResult struct {
	domain *fullSource
	view   virtual.GridProjection
}

func (w *window) editWork(edit virtual.GridEditState) error {
	if w.editTask != nil {
		return errBusy
	}
	if _, e := numeric(edit.Text); e != nil {
		return e
	}
	req, e := w.grid.PrepareEdit()
	if e != nil {
		return e
	}
	if req.State() != edit {
		return errStale
	}
	ctx, e := w.lifetime()
	if e != nil {
		return e
	}
	if w.editToken == math.MaxUint64 {
		return virtual.ErrGridExhausted
	}
	w.editToken++
	token := w.editToken
	domain, queryToken, q := w.domain, w.queryToken, cloneQuery(w.query)
	variable := w.p.variable
	state, e := w.grid.State()
	if e != nil {
		return e
	}
	if state.Stamp.Revision == math.MaxUint64 {
		return virtual.ErrGridExhausted
	}
	revision := state.Stamp.Revision + 1
	var task *tasks.Task
	task, e = tasks.Start(w.manager, ctx, func(ctx context.Context, _ func(float64, string)) (editResult, error) {
		next, e := edited(ctx, domain, edit.RowKey, edit.ColumnID, edit.Text, edit.BaseVersion)
		if e != nil {
			return editResult{}, e
		}
		view, e := virtual.BuildGridProjection(ctx, input(next, variable), revision, q, limits())
		return editResult{next, view}, e
	}, func(out editResult, e error) {
		delete(w.jobs, task)
		if token != w.editToken {
			return
		}
		w.editTask = nil
		if w.domain != domain || w.queryToken != queryToken {
			w.report(errStale)
			return
		}
		if e != nil {
			w.report(e)
			return
		}
		if e = w.grid.CompleteEdit(req, out.view); e != nil {
			w.report(e)
			return
		}
		w.domain, w.view = out.domain, out.view
		w.message = fmt.Sprintf("Cell saved: domain v%d / view v%d", out.domain.Revision(), out.view.Revision())
		w.report(nil)
	})
	if e != nil {
		return e
	}
	w.editTask = task
	w.jobs[task] = "immutable edit candidate"
	w.message = "Copying stored edit candidate and projection on worker"
	return nil
}
func (w *window) external() error {
	if w.settling || w.editTask != nil {
		return errBusy
	}
	state, e := w.grid.State()
	if e != nil {
		return e
	}
	k, c := state.Active.RowKey, state.Active.ColumnID
	if state.Edit.Active {
		k, c = state.Edit.RowKey, state.Edit.ColumnID
	}
	value, version, e := w.domain.value(k, c)
	if e != nil {
		return e
	}
	if value == math.MaxInt64 {
		return virtual.ErrGridExhausted
	}
	ctx, e := w.lifetime()
	if e != nil {
		return e
	}
	if w.editToken == math.MaxUint64 || state.Stamp.Revision == math.MaxUint64 {
		return virtual.ErrGridExhausted
	}
	w.editToken++
	token := w.editToken
	domain, qToken, q := w.domain, w.queryToken, cloneQuery(w.query)
	variable := w.p.variable
	revision := state.Stamp.Revision + 1
	var task *tasks.Task
	task, e = tasks.Start(w.manager, ctx, func(ctx context.Context, _ func(float64, string)) (editResult, error) {
		next, e := edited(ctx, domain, k, c, fmt.Sprint(value+1), version)
		if e != nil {
			return editResult{}, e
		}
		view, e := virtual.BuildGridProjection(ctx, input(next, variable), revision, q, limits())
		return editResult{next, view}, e
	}, func(out editResult, e error) {
		delete(w.jobs, task)
		if token != w.editToken {
			return
		}
		w.editTask = nil
		current, ce := w.grid.State()
		if w.domain != domain || w.queryToken != qToken || ce != nil || current.Stamp != state.Stamp {
			w.report(errStale)
			return
		}
		if e != nil {
			w.report(e)
			return
		}
		if e = w.grid.ReplaceSource(out.view); e != nil {
			w.report(e)
			return
		}
		w.domain, w.view = out.domain, out.view
		w.message = "External domain version accepted; older draft remains recoverable"
		w.report(nil)
	})
	if e != nil {
		return e
	}
	w.editTask = task
	w.jobs[task] = "external immutable candidate"
	return nil
}
func (w *window) onCommit(e ui.GridEditEvent) {
	current, err := w.grid.Edit()
	if err != nil || e.ID != "data" || current != e.State {
		w.report(errStale)
		return
	}
	w.report(w.editWork(current))
}
func (w *window) describe() ui.Node {
	button := func(id, text string, fn func() error) ui.Node {
		return ui.Node{ID: id, Kind: ui.Button, Text: text, Activate: func() { w.report(fn()) }}
	}
	status := w.message
	if state, e := w.grid.State(); e == nil {
		status += fmt.Sprintf(" · domain/view %d/%d · rows %d · selected %d · worker credits %d", w.domain.Revision(), state.Stamp.Revision, state.Stamp.Rows, state.Selection.CapturedRows, w.manager.Pending().Active)
		if state.Edit.Active {
			status += " · draft " + string(state.Edit.Status)
		}
	}
	return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{
		{ID: "title", Kind: ui.Label, Text: "Loaded stored records · " + w.id + " · 20 signed numeric fields"},
		{ID: "tools", Kind: ui.Row, Children: []ui.Node{button("load", "Load records", w.p.load), button("sort-asc", "Sort ↑ value 1", func() error {
			q := cloneQuery(w.query)
			q.Sort = []virtual.GridSortColumn{{ColumnID: column(1), Kind: virtual.GridSortInteger}}
			return w.queryWork(q)
		}), button("sort-desc", "Sort ↓ value 1", func() error {
			q := cloneQuery(w.query)
			q.Sort = []virtual.GridSortColumn{{ColumnID: column(1), Kind: virtual.GridSortInteger, Descending: true}}
			return w.queryWork(q)
		}), button("clear-query", "Clear query", func() error { return w.queryWork(virtual.GridQuery{}) }), button("cancel-work", "Cancel work", w.cancelWork)}},
		{ID: "filters", Kind: ui.Row, Children: []ui.Node{{ID: "filter-text", Kind: ui.TextField, Text: w.filterText, Change: func(e ui.ChangeEvent) { w.filterText = e.Text; w.report(nil) }, Accessibility: ui.Accessibility{Name: "Value 2 contains"}}, button("apply-filter", "Filter value 2", func() error {
			return w.applyFilter()
		}), button("select-all", "Select all view rows", w.grid.SelectAll), button("clear-selection", "Clear selection", w.grid.ClearSelection)}},
		{ID: "editing", Kind: ui.Row, Children: []ui.Node{button("begin-edit", "Edit active cell", w.begin), button("save-edit", "Save cell", w.save), button("cancel-edit", "Cancel edit", w.cancelEdit), button("external", "External +1", w.external), button("last", "Last row + column 19", func() error {
			s, e := w.grid.State()
			if e != nil || s.Stamp.Rows == 0 {
				return virtual.ErrGridRange
			}
			k, e := w.view.RowKey(s.Stamp.Rows - 1)
			if e != nil {
				return e
			}
			return w.grid.JumpToKey(k, column(19))
		})}},
		{ID: "data", Kind: ui.DataGrid, Grid: w.grid, Grow: true, GridChange: func(ui.GridEvent) { w.report(nil) }, GridCommit: w.onCommit, Accessibility: ui.Accessibility{Name: "Stored loaded data grid"}},
		{ID: "status", Kind: ui.Label, Text: status, Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 900})}},
	}}
}
