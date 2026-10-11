package main

import (
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
	source                          *gridSource
	grid                            *virtual.Grid
	status                          string
	selectionEvents, viewportEvents uint64
	disabled, modal                 bool
}

func newGridWindow(owner *app.Application, id string, rows int) (*gridWindow, error) {
	if owner == nil || id == "" {
		return nil, errSourceConfig
	}
	source, err := newSource("grid-example:seed7:"+id, rows)
	if err != nil {
		return nil, err
	}
	g, err := virtual.NewGrid(source, sourceColumns(), virtual.GridConfig{RowHeight: 24, HeaderHeight: 28, FrozenColumns: 2, OverscanRows: 1, OverscanColumns: 1}, virtual.GridViewport{Width: 800, Height: 508})
	if err != nil {
		return nil, err
	}
	return &gridWindow{owner: owner, id: id, source: source, grid: g, status: "Ready"}, nil
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
	if err = w.grid.ReplaceSource(next); err != nil {
		return err
	}
	w.source = next
	w.status = fmt.Sprintf("Revision %d · reverse=%t · even rows=%t", next.revision, next.reverse, next.filtered)
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
	default:
		w.report(errors.New("grid example: unexpected event reason"))
		return
	}
	w.report(nil)
}
func (w *gridWindow) view() ui.Node {
	state, err := w.grid.State()
	text := w.status
	if err != nil {
		text = err.Error()
	} else {
		text = fmt.Sprintf("%s · %d rows × %d columns · revision %d · active %s/%s", w.status, state.Stamp.Rows, state.Columns, state.Stamp.Revision, state.Active.RowKey, state.Active.ColumnID)
	}
	button := func(id, label string, fn func() error) ui.Node {
		return ui.Node{ID: id, Kind: ui.Button, Text: label, Activate: func() { w.report(fn()) }}
	}
	return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{
		{ID: "title", Kind: ui.Label, Text: fmt.Sprintf("%s · %d domain rows · procedural seed 7", w.id, w.source.total)},
		{ID: "tools", Kind: ui.Row, Children: []ui.Node{
			button("reverse", "Reverse rows", func() error { return w.projection(!w.source.reverse, w.source.filtered) }),
			button("filter", "Even rows", func() error { return w.projection(w.source.reverse, !w.source.filtered) }),
			button("first", "First", w.first), button("last", "Last + column 19", w.last), button("pan", "Pan", w.pan),
		}},
		{ID: "more", Kind: ui.Row, Children: []ui.Node{
			button("page-up", "Page up", func() error { return w.grid.Page(-1) }), button("page-down", "Page down", func() error { return w.grid.Page(1) }),
			button("disable", "Toggle disabled", func() error { w.disabled = !w.disabled; return nil }),
			button("modal-open", "Open modal", func() error { w.modal = true; return nil }),
			button("close", "Close window", func() error { return w.owner.CloseWindow(w.id) }),
		}},
		{ID: "data", Kind: ui.DataGrid, Grow: true, Grid: w.grid, GridChange: w.onGrid, Disabled: w.disabled, Accessibility: ui.Accessibility{Name: "Procedural data grid"}},
		{ID: "status", Kind: ui.Label, Text: text, Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: statusMaxWidth})}},
		{ID: "modal", Kind: ui.Dialog, Open: w.modal, Text: "Grid is covered by a modal", Dismiss: func() { w.modal = false; w.report(nil) }, Children: []ui.Node{
			{ID: "modal-label", Kind: ui.Label, Text: "Background grid input must be refused while this modal is open."},
			button("modal-close", "Dismiss", func() error { w.modal = false; return nil }),
		}},
	}}
}

type gridProgram struct {
	owner   *app.Application
	windows []*gridWindow
}

func newProgram() (*gridProgram, error) { return newProgramSize(1024, 760) }

func newProgramSize(width, height int) (*gridProgram, error) {
	a, err := app.New(app.Config{QueueCapacity: 64})
	if err != nil {
		return nil, err
	}
	p := &gridProgram{owner: a}
	for _, config := range []struct {
		id   string
		rows int
	}{{"small", 10_000}, {"large", 1_000_000}} {
		w, err := newGridWindow(a, config.id, config.rows)
		if err != nil {
			a.Close()
			return nil, err
		}
		if err = a.Open(app.Window{ID: w.id, Title: fmt.Sprintf("Vellum Grid · %s", w.id), Width: float32(width), Height: float32(height), View: w.view}); err != nil {
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
