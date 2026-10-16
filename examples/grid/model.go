package main

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

// The default Label wraps within this application-local logical width.
// Long selection/status text must not become the native window minimum width.
const statusMaxWidth = 640

type gridWindow struct {
	owner                           *app.Application
	id                              string
	windowRef                       app.WindowRef
	source                          *gridSource
	grid                            *virtual.Grid
	status                          string
	selectionEvents, viewportEvents uint64
	disabled, modal                 bool
	variableRows, heightControls    bool
	rowSelection                    bool
	editableCells                   bool
	metricCalls                     *rowMetricCalls
	// Injected by native consumers. Portable commands never pretend an owner
	// model read acknowledges earlier originating native editor records.
	settleCell        func(context.Context, func(error)) error
	cellCommitPending bool
	cellCommitCancel  context.CancelFunc
	cellCommitToken   uint64
}

func newGridWindow(owner *app.Application, id string, rows int) (*gridWindow, error) {
	return newGridWindowMode(owner, id, rows, false)
}

func newGridWindowMode(owner *app.Application, id string, rows int, variableRows bool) (*gridWindow, error) {
	return newGridWindowModes(owner, id, rows, variableRows, false)
}

func newGridWindowModes(owner *app.Application, id string, rows int, variableRows, rowSelection bool) (*gridWindow, error) {
	return newGridWindowOptions(owner, id, rows, variableRows, rowSelection, false)
}

func newGridWindowOptions(owner *app.Application, id string, rows int, variableRows, rowSelection, editableCells bool) (*gridWindow, error) {
	if owner == nil || id == "" {
		return nil, errSourceConfig
	}
	source, err := newSource("grid-example:seed7:"+id, rows)
	if err != nil {
		return nil, err
	}
	metrics := &rowMetricCalls{}
	config := virtual.GridConfig{RowHeight: 24, HeaderHeight: 28, FrozenColumns: 2, OverscanRows: 1, OverscanColumns: 1, RowSelection: rowSelection, CellEditing: editableCells}
	if editableCells {
		config.MaxEditBytes = maxCellTextBytes
	}
	g, err := virtual.NewGrid(cellSource(source, metrics, variableRows, editableCells), sourceColumns(), config, virtual.GridViewport{Width: 800, Height: 508})
	if err != nil {
		return nil, err
	}
	return &gridWindow{owner: owner, id: id, source: source, grid: g, status: "Ready", variableRows: variableRows, heightControls: variableRows, metricCalls: metrics, rowSelection: rowSelection, editableCells: editableCells}, nil
}

func (w *gridWindow) selectAllRows() error {
	if err := w.grid.SelectAll(); err != nil {
		return err
	}
	w.status = "Selected all identities in the current captured view"
	return nil
}
func (w *gridWindow) clearRows() error {
	if err := w.grid.ClearSelection(); err != nil {
		return err
	}
	w.status = "Cleared row selection"
	return nil
}

func (w *gridWindow) beginCell() error {
	if err := w.grid.BeginActiveEdit(); err != nil {
		return err
	}
	w.status = "Editing active cell; Enter saves, Escape cancels"
	return nil
}

func (w *gridWindow) commitCell() error {
	edit, err := w.grid.Edit()
	if err != nil {
		return err
	}
	if !edit.Active {
		return virtual.ErrGridRange
	}
	// Domain validation precedes proposal creation. Failure keeps the exact
	// draft and live source available for correction or cancellation.
	if err = validateCellText(edit.Text); err != nil {
		return err
	}
	request, err := w.grid.PrepareEdit()
	if err != nil {
		return err
	}
	state := request.State()
	candidate, err := w.source.withCell(state.RowKey, state.ColumnID, state.Text, state.BaseVersion)
	if err != nil {
		return err
	}
	if err = w.grid.CompleteEdit(request, cellSource(candidate, w.metricCalls, w.variableRows, w.editableCells)); err != nil {
		return err
	}
	// Publish the domain pointer only after Grid accepted this exact candidate.
	w.source = candidate
	w.status = fmt.Sprintf("Saved %s / %s at cell version %d", state.RowKey, state.ColumnID, state.BaseVersion+1)
	return nil
}

func (w *gridWindow) requestCellCommit() error {
	if w.cellCommitPending {
		return virtual.ErrGridEditBusy
	}
	if w.settleCell == nil {
		return errors.New("grid example: originating editor settlement unavailable")
	}
	edit, err := w.grid.Edit()
	if err != nil {
		return err
	}
	if !edit.Active {
		return virtual.ErrGridRange
	}
	if w.cellCommitToken == math.MaxUint64 {
		return virtual.ErrGridExhausted
	}
	lifetime := w.windowRef.Context()
	if lifetime.Err() != nil {
		return app.ErrWindowClosed
	}
	ctx, cancel := context.WithCancel(lifetime)
	w.cellCommitToken++
	token, session := w.cellCommitToken, edit.Session
	w.cellCommitPending, w.cellCommitCancel = true, cancel
	err = w.settleCell(ctx, func(err error) {
		if !w.cellCommitPending || w.cellCommitToken != token {
			return
		}
		w.cellCommitPending, w.cellCommitCancel = false, nil
		cancel()
		if err != nil {
			w.report(err)
			return
		}
		current, stateErr := w.grid.Edit()
		if stateErr != nil {
			w.report(stateErr)
			return
		}
		if !current.Active || current.Session != session {
			w.report(virtual.ErrGridStale)
			return
		}
		w.report(w.commitCell())
	})
	if err != nil {
		w.cellCommitPending, w.cellCommitCancel = false, nil
		cancel()
		return err
	}
	w.status = "Waiting for originating cell editor acknowledgment"
	return nil
}

func (w *gridWindow) cancelCell() error {
	if w.cellCommitCancel != nil {
		w.cellCommitCancel()
		w.cellCommitCancel = nil
	}
	w.cellCommitPending = false
	if err := w.grid.CancelEdit(); err != nil {
		return err
	}
	w.status = "Canceled cell draft"
	return nil
}

func (w *gridWindow) externalCell() error {
	state, err := w.grid.State()
	if err != nil {
		return err
	}
	key, column := state.Active.RowKey, state.Active.ColumnID
	if state.Edit.Active {
		key, column = state.Edit.RowKey, state.Edit.ColumnID
	}
	current, err := w.source.cellValue(key, column)
	if err != nil {
		return err
	}
	if current.version == math.MaxUint64 {
		return virtual.ErrGridExhausted
	}
	text := fmt.Sprintf("External update v%d", current.version+1)
	candidate, err := w.source.withCell(key, column, text, current.version)
	if err != nil {
		return err
	}
	if err = w.grid.ReplaceSource(cellSource(candidate, w.metricCalls, w.variableRows, w.editableCells)); err != nil {
		return err
	}
	w.source = candidate
	w.status = "External cell version changed; an older draft cannot overwrite it"
	return nil
}
func (w *gridWindow) report(err error) {
	if err != nil {
		w.status = err.Error()
	}
	if invalid := w.owner.Invalidate(w.id); invalid != nil && !errors.Is(invalid, app.ErrClosed) && !errors.Is(invalid, app.ErrWindowClosed) {
		w.status = invalid.Error()
	}
}
func (w *gridWindow) projection(reverse, filtered bool) error {
	if reverse == w.source.reverse && filtered == w.source.filtered {
		return nil
	}
	next, err := w.source.projected(reverse, filtered)
	if err != nil {
		return err
	}
	if err = w.grid.ReplaceSource(cellSource(next, w.metricCalls, w.variableRows, w.editableCells)); err != nil {
		return err
	}
	w.source = next
	w.status = fmt.Sprintf("Revision %d · reverse=%t · even rows=%t", next.revision, next.reverse, next.filtered)
	return nil
}

// rowSource opts in only when requested. The fixed source remains the exact
// original type, preserving default indexed access and callback-count fixtures.
func rowSource(source *gridSource, metrics *rowMetricCalls, variableRows bool) virtual.GridSource {
	if variableRows {
		return &variableGridSource{gridSource: source, metrics: metrics}
	}
	return source
}

func cellSource(source *gridSource, metrics *rowMetricCalls, variableRows, editableCells bool) virtual.GridSource {
	if !editableCells {
		return rowSource(source, metrics, variableRows)
	}
	if variableRows {
		return &editableVariableGridSource{&variableGridSource{gridSource: source, metrics: metrics}}
	}
	return &editableGridSource{source}
}

func (w *gridWindow) setVariableRows(enabled bool) error {
	if enabled == w.variableRows {
		return nil
	}
	// Height changes are immutable source revisions, including a transition
	// back to the fallback fixed height; old source metrics never mutate.
	next, err := w.source.projected(w.source.reverse, w.source.filtered)
	if err != nil {
		return err
	}
	if err = w.grid.ReplaceSource(cellSource(next, w.metricCalls, enabled, w.editableCells)); err != nil {
		return err
	}
	w.source, w.variableRows = next, enabled
	w.status = fmt.Sprintf("Revision %d · variable row heights=%t", next.revision, enabled)
	return nil
}

func (w *gridWindow) last() error {
	key, err := w.source.RowKey(w.source.RowCount() - 1)
	if err != nil {
		return err
	}
	if err = w.grid.JumpToKey(key, columnID(19)); err != nil {
		return err
	}
	w.status = "Jumped to last projected row and column"
	return nil
}
func (w *gridWindow) first() error {
	key, err := w.source.RowKey(0)
	if err != nil {
		return err
	}
	if err = w.grid.JumpToKey(key, columnID(0)); err != nil {
		return err
	}
	w.status = "Jumped to first projected row and column"
	return nil
}
func (w *gridWindow) pan() error {
	if err := w.grid.Scroll(360, 240); err != nil {
		return err
	}
	w.status = "Panned 360×240 logical pixels"
	return nil
}
func (w *gridWindow) onGrid(event ui.GridEvent) {
	if event.ID != "data" {
		w.report(errors.New("grid example: unexpected event identity"))
		return
	}
	switch event.Reason {
	case ui.GridSelection:
		w.selectionEvents++
		w.status = "Selected " + event.State.Active.RowKey + " / " + event.State.Active.ColumnID
	case ui.GridViewport:
		w.viewportEvents++
	case ui.GridEditing:
		w.status = "Cell draft " + string(event.State.Edit.Status)
	default:
		w.report(errors.New("grid example: unexpected event reason"))
		return
	}
	w.report(nil)
}

func (w *gridWindow) onCellCommit(event ui.GridEditEvent) {
	if w.cellCommitPending {
		w.report(virtual.ErrGridEditBusy)
		return
	}
	current, err := w.grid.Edit()
	if err != nil {
		w.report(err)
		return
	}
	if event.ID != "data" || event.State != current || !current.Active {
		w.report(virtual.ErrGridStale)
		return
	}
	w.report(w.commitCell())
}
func (w *gridWindow) view() ui.Node {
	state, err := w.grid.State()
	text := w.status
	if err != nil {
		text = err.Error()
	} else {
		text = fmt.Sprintf("%s · %d rows × %d columns · revision %d · active %s/%s", w.status, state.Stamp.Rows, state.Columns, state.Stamp.Revision, state.Active.RowKey, state.Active.ColumnID)
		if w.rowSelection {
			text += fmt.Sprintf(" · %d captured selected identities · %d sparse overrides", state.Selection.CapturedRows, state.Selection.Overrides)
		}
		if w.editableCells && state.Edit.Active {
			text += fmt.Sprintf(" · draft %s/%s · %s · base/current v%d/v%d", state.Edit.RowKey, state.Edit.ColumnID, state.Edit.Status, state.Edit.BaseVersion, state.Edit.CurrentVersion)
		}
	}
	button := func(id, label string, fn func() error) ui.Node {
		return ui.Node{ID: id, Kind: ui.Button, Text: label, Activate: func() { w.report(fn()) }}
	}
	moreButtons := []ui.Node{
		button("page-up", "Page up", func() error { return w.grid.Page(-1) }), button("page-down", "Page down", func() error { return w.grid.Page(1) }),
		button("disable", "Toggle disabled", func() error { w.disabled = !w.disabled; return nil }),
		button("modal-open", "Open modal", func() error { w.modal = true; return nil }),
		button("close", "Close window", func() error { return w.owner.CloseWindow(w.windowRef) }),
	}
	if w.heightControls {
		moreButtons = append(moreButtons, button("row-heights", "Toggle row heights", func() error { return w.setVariableRows(!w.variableRows) }))
	}
	var selectionControls []ui.Node
	if w.rowSelection {
		selectionControls = []ui.Node{{ID: "selection-tools", Kind: ui.Row, Children: []ui.Node{
			button("select-all", "Select all rows", w.selectAllRows), button("clear-selection", "Clear row selection", w.clearRows),
		}}}
	}
	children := []ui.Node{
		{ID: "title", Kind: ui.Label, Text: fmt.Sprintf("%s · %d domain rows · procedural seed 7", w.id, w.source.total)},
		{ID: "tools", Kind: ui.Row, Children: []ui.Node{
			button("reverse", "Reverse rows", func() error { return w.projection(!w.source.reverse, w.source.filtered) }),
			button("filter", "Even rows", func() error { return w.projection(w.source.reverse, !w.source.filtered) }),
			button("first", "First", w.first), button("last", "Last + column 19", w.last), button("pan", "Pan", w.pan),
		}},
		{ID: "more", Kind: ui.Row, Children: moreButtons},
	}
	children = append(children, selectionControls...)
	if w.editableCells {
		children = append(children, ui.Node{ID: "editing-tools", Kind: ui.Row, Children: []ui.Node{
			button("edit-cell", "Edit active cell", w.beginCell), button("save-cell", "Save cell", w.requestCellCommit),
			button("cancel-cell", "Cancel edit", w.cancelCell), button("external-cell", "External update", w.externalCell),
		}})
	}
	children = append(children,
		ui.Node{ID: "data", Kind: ui.DataGrid, Grow: true, Grid: w.grid, GridChange: w.onGrid, GridCommit: w.onCellCommit, Disabled: w.disabled, Accessibility: ui.Accessibility{Name: "Procedural data grid"}},
		ui.Node{ID: "status", Kind: ui.Label, Text: text, Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: statusMaxWidth})}},
		ui.Node{ID: "modal", Kind: ui.Dialog, Open: w.modal, Text: "Grid is covered by a modal", Dismiss: func() { w.modal = false; w.report(nil) }, Children: []ui.Node{
			{ID: "modal-label", Kind: ui.Label, Text: "Background grid input must be refused while this modal is open."},
			button("modal-close", "Dismiss", func() error { w.modal = false; return nil }),
		}},
	)
	return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: children}
}

type gridProgram struct {
	owner   *app.Application
	windows []*gridWindow
	// Startup mode is immutable; per-window toggles belong to gridWindow.
	variableRows  bool
	rowSelection  bool
	editableCells bool
}

func newProgram() (*gridProgram, error) { return newProgramSize(1024, 760) }

func newProgramSize(width, height int) (*gridProgram, error) {
	return newProgramMode(width, height, false)
}

func newProgramMode(width, height int, variableRows bool) (*gridProgram, error) {
	return newProgramModes(width, height, variableRows, false)
}

func newProgramModes(width, height int, variableRows, rowSelection bool) (*gridProgram, error) {
	return newProgramOptions(width, height, variableRows, rowSelection, false)
}

func newProgramOptions(width, height int, variableRows, rowSelection, editableCells bool) (*gridProgram, error) {
	a, err := app.New(app.Config{QueueCapacity: 64})
	if err != nil {
		return nil, err
	}
	p := &gridProgram{owner: a, variableRows: variableRows, rowSelection: rowSelection, editableCells: editableCells}
	for _, config := range []struct {
		id   string
		rows int
	}{{"small", 10_000}, {"large", 1_000_000}} {
		w, err := newGridWindowOptions(a, config.id, config.rows, variableRows, rowSelection, editableCells)
		if err != nil {
			a.Close()
			return nil, err
		}
		if w.windowRef, err = a.OpenWindow(app.Window{ID: w.id, Title: fmt.Sprintf("Vellum Grid · %s", w.id), Width: float32(width), Height: float32(height), View: w.view}); err != nil {
			a.Close()
			return nil, err
		}
		p.windows = append(p.windows, w)
	}
	return p, nil
}

// Keep these model commands explicit: portable tests do not pretend command
// calls are native input. Offscreen qualification will route physical events.
func (w *gridWindow) saturateScroll() error { return w.grid.Scroll(math.MaxInt, math.MaxInt) }
