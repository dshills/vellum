// Package replayconsumer declares a real public-UI workload for a future native
// replay router. Its portable command qualification is a separate proof.
package replayconsumer

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"unicode/utf8"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/tasks"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

//go:embed fixture.json
var fixtureBytes []byte

const (
	ViewportWidth   = 860
	ViewportHeight  = 620
	OpenProject     = "project-open"
	EditProperty    = "property-edit"
	SortTable       = "table-sort"
	StartTask       = "task-start"
	CancelTask      = "task-cancel"
	Undo            = "undo"
	Redo            = "redo"
	OpenModal       = "modal-open"
	CloseModal      = "modal-close"
	RemountProperty = "property-remount"
	DisableProperty = "property-disable"
	SelectRow       = "row-select"
)

var (
	ErrNotOpen  = errors.New("consumer: project is not open")
	ErrModal    = errors.New("consumer: modal blocks background commands")
	ErrDisabled = errors.New("consumer: property is disabled")
	ErrStale    = errors.New("consumer: stale model revision")
	ErrAction   = errors.New("consumer: invalid action")
)

type Config struct {
	Owner    *app.Application
	Executor execution.Executor
	Window   string
}

type Action struct {
	ID, Value        string
	CheckRevision    bool
	ExpectedRevision uint64
}

type projectRow struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Priority int    `json:"priority"`
}
type project struct {
	Name string       `json:"name"`
	Rows []projectRow `json:"rows"`
}
type dataset struct{ rows []projectRow }

func (d *dataset) Len() int         { return len(d.rows) }
func (d *dataset) Revision() uint64 { return 1 }
func (d *dataset) Row(index int) virtual.Row {
	r := d.rows[index]
	return virtual.Row{ID: r.ID, Cells: []string{r.Name, fmt.Sprintf("%02d", r.Priority)}}
}

type TaskResult struct {
	Ordinal, CapturedRevision uint64
	Checksum                  int
	Outcome                   string
}

// Fixture model and its Registry/View/Snapshot belong to Config.Owner. Worker
// data is immutable and detached; only cleanup count is atomically published.
type Fixture struct {
	Owner                   *app.Application
	Commands                *command.Registry
	Window                  string
	Journal                 *history.Journal
	Table                   *virtual.Table
	lifetime                context.Context
	manager                 *tasks.Manager
	data                    *dataset
	loaded                  bool
	property                string
	propertyDisabled, modal bool
	propertyMount           uint64
	viewEpoch               uint64
	focus                   ui.FocusRequest
	sorted                  bool
	status                  string
	task                    *tasks.Task
	nextTask                uint64
	results                 []TaskResult
	cleanups                atomic.Uint64
	argument                string
	invoking                bool
	uiAttempts              map[string]int
}

func FixtureBytes() []byte  { return slices.Clone(fixtureBytes) }
func FixtureSHA256() string { h := sha256.Sum256(fixtureBytes); return hex.EncodeToString(h[:]) }

func New(config Config) (*Fixture, error) {
	if config.Owner == nil || config.Executor == nil {
		return nil, fmt.Errorf("consumer: explicit owner and executor required")
	}
	if config.Window == "" {
		config.Window = "replay"
	}
	j, err := history.New(32)
	if err != nil {
		return nil, err
	}
	f := &Fixture{Owner: config.Owner, Window: config.Window, Journal: j, Commands: command.New(), data: &dataset{}, propertyMount: 1, status: "No project", uiAttempts: map[string]int{}}
	f.Table, err = virtual.NewTable(f.data, 4)
	if err != nil {
		return nil, err
	}
	if err = f.register(); err != nil {
		return nil, err
	}
	if err = f.Owner.Open(app.Window{ID: f.Window, Title: "Deterministic project", Width: ViewportWidth, Height: ViewportHeight, View: f.View}); err != nil {
		return nil, err
	}
	f.lifetime, err = f.Owner.WindowContext(f.Window)
	if err != nil {
		return nil, err
	}
	f.manager, err = tasks.NewWithConfig(tasks.Config{Owner: f.Owner, Window: f.Window, Executor: config.Executor, Limit: 1, Invalidate: func() { _ = f.Owner.Invalidate(f.Window) }})
	if err != nil {
		_ = f.Owner.CloseWindow(f.Window)
		return nil, err
	}
	return f, nil
}

func (f *Fixture) touch()             { f.viewEpoch++; _ = f.Owner.Invalidate(f.Window) }
func (f *Fixture) PropertyID() string { return fmt.Sprintf("property-%d", f.propertyMount) }
func (f *Fixture) SetFocus(id string) { f.focus.Sequence++; f.focus.ID = id; f.touch() }
func (f *Fixture) activeTask() bool {
	if f.task == nil {
		return false
	}
	s := f.task.Snapshot()
	return !s.WorkerDone || !s.OwnerDone
}
func (f *Fixture) available(id string, allowModal bool) error {
	if f.lifetime != nil && f.lifetime.Err() != nil {
		return app.ErrWindowClosed
	}
	if f.Owner.Context().Err() != nil {
		return app.ErrClosed
	}
	if f.modal && !allowModal {
		return ErrModal
	}
	if id != OpenProject && !f.loaded {
		return ErrNotOpen
	}
	switch id {
	case EditProperty:
		if f.propertyDisabled {
			return ErrDisabled
		}
	case StartTask:
		if f.activeTask() {
			return tasks.ErrCapacity
		}
	case CancelTask:
		if !f.activeTask() {
			return ErrDisabled
		}
	case Undo:
		if f.Journal.Snapshot().UndoCount == 0 {
			return history.ErrEmpty
		}
	case Redo:
		if f.Journal.Snapshot().RedoCount == 0 {
			return history.ErrEmpty
		}
	case CloseModal:
		if !f.modal {
			return ErrDisabled
		}
	}
	return nil
}

func (f *Fixture) register() error {
	definitions := []command.Command{
		{ID: OpenProject, Label: "Open fixture", Shortcut: "Meta+O", Execute: func(context.Context) error { return f.open() }},
		{ID: EditProperty, Label: "Edit project name", Execute: func(context.Context) error { return f.edit(f.argument) }},
		{ID: SortTable, Label: "Sort names", Shortcut: "Meta+S", Execute: func(context.Context) error { return f.sort() }},
		{ID: StartTask, Label: "Start task", Execute: func(ctx context.Context) error { return f.start(ctx) }},
		{ID: CancelTask, Label: "Cancel task", Execute: func(context.Context) error {
			f.task.Cancel()
			f.status = "Cancellation requested"
			f.touch()
			return nil
		}},
		{ID: Undo, Label: "Undo", Shortcut: "Meta+Z", Execute: func(context.Context) error {
			if err := f.Journal.Undo(); err != nil {
				return err
			}
			f.status = "Undone"
			f.touch()
			return nil
		}},
		{ID: Redo, Label: "Redo", Shortcut: "Meta+Shift+Z", Execute: func(context.Context) error {
			if err := f.Journal.Redo(); err != nil {
				return err
			}
			f.status = "Redone"
			f.touch()
			return nil
		}},
		{ID: OpenModal, Label: "Open modal", Execute: func(context.Context) error { f.modal = true; f.status = "Modal open"; f.touch(); return nil }},
		{ID: CloseModal, Label: "Close modal", Execute: func(context.Context) error { f.modal = false; f.status = "Modal closed"; f.touch(); return nil }},
		{ID: RemountProperty, Label: "Remount property", Execute: func(context.Context) error { f.propertyMount++; f.status = "Property remounted"; f.touch(); return nil }},
		{ID: DisableProperty, Label: "Set property disabled", Execute: func(context.Context) error {
			if f.argument != "true" && f.argument != "false" {
				return ErrAction
			}
			f.propertyDisabled = f.argument == "true"
			f.touch()
			return nil
		}},
		{ID: SelectRow, Label: "Select visible row", Execute: func(context.Context) error {
			i, err := strconv.Atoi(f.argument)
			if err != nil {
				return err
			}
			if _, err = f.Table.SelectVisible(i); err != nil {
				return err
			}
			f.touch()
			return nil
		}},
	}
	for _, definition := range definitions {
		id := definition.ID
		definition.Available = func() error { return f.available(id, id == CloseModal) }
		if err := f.Commands.Register(definition); err != nil {
			return err
		}
	}
	return nil
}

// Invoke is command-only model input, not a native pointer/key/Change simulator.
// It validates payload, revision, context and Registry availability on owner.
func (f *Fixture) Invoke(ctx context.Context, action Action) error {
	if f.invoking {
		return fmt.Errorf("consumer: command is already executing")
	}
	if action.CheckRevision && action.ExpectedRevision != f.Journal.Snapshot().Revision {
		return ErrStale
	}
	if !action.CheckRevision && action.ExpectedRevision != 0 {
		return ErrAction
	}
	switch action.ID {
	case EditProperty:
		if action.Value == "" || len(action.Value) > 256 || !utf8.ValidString(action.Value) || strings.ContainsRune(action.Value, 0) {
			return ErrAction
		}
	case DisableProperty:
		if action.Value != "true" && action.Value != "false" {
			return ErrAction
		}
	case SelectRow:
		i, err := strconv.Atoi(action.Value)
		if err != nil || i < 0 || i >= 4 {
			return ErrAction
		}
	case OpenProject, SortTable, StartTask, CancelTask, Undo, Redo, OpenModal, CloseModal, RemountProperty:
		if action.Value != "" {
			return ErrAction
		}
	default:
		return ErrAction
	}
	f.invoking = true
	f.argument = action.Value
	defer func() { f.invoking = false; f.argument = "" }()
	return f.Commands.Invoke(ctx, action.ID)
}

func (f *Fixture) open() error {
	var p project
	decoder := json.NewDecoder(bytes.NewReader(fixtureBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&p); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return ErrAction
	}
	if p.Name == "" || len(p.Rows) < 4 || len(p.Rows) > 128 {
		return ErrAction
	}
	ids := map[string]bool{}
	for _, r := range p.Rows {
		if r.ID == "" || ids[r.ID] || r.Name == "" || r.Priority < 0 || r.Priority > 99 {
			return ErrAction
		}
		ids[r.ID] = true
	}
	next := &dataset{rows: slices.Clone(p.Rows)}
	if err := f.Journal.Replace(true, func() error {
		if err := f.Table.ReplaceSource(next); err != nil {
			return err
		}
		f.data = next
		f.property = p.Name
		f.loaded = true
		f.sorted = false
		return nil
	}); err != nil {
		return err
	}
	f.status = "Project open"
	f.touch()
	return nil
}
func (f *Fixture) edit(value string) error {
	if value == "" || len(value) > 256 || !utf8.ValidString(value) || strings.ContainsRune(value, 0) {
		return ErrAction
	}
	if value == f.property {
		return nil
	}
	before := f.property
	if err := f.Journal.Apply(history.Change{Label: "Project name", Apply: func() error { f.property = value; return nil }, Revert: func() error { f.property = before; return nil }}); err != nil {
		return err
	}
	f.status = "Property edited"
	f.touch()
	return nil
}
func (f *Fixture) sort() error {
	projection, err := virtual.Project(f.data, "", 0, false)
	if err != nil {
		return err
	}
	if err = f.Table.ApplyProjection(projection, f.Table.ProjectionToken()); err != nil {
		return err
	}
	f.sorted = true
	f.status = "Table sorted"
	f.touch()
	return nil
}
func (f *Fixture) start(ctx context.Context) error {
	ordinal, revision := f.nextTask+1, f.Journal.Snapshot().Revision
	data := f.data // Immutable fixture source; replacement never edits old rows.
	task, err := tasks.Start(f.manager, ctx, func(ctx context.Context, progress func(float64, string)) (TaskResult, error) {
		defer f.cleanups.Add(1)
		result := TaskResult{Ordinal: ordinal, CapturedRevision: revision}
		if err := ctx.Err(); err != nil {
			return result, err
		}
		for _, row := range data.rows {
			result.Checksum += row.Priority
		}
		progress(1, "Checksum complete")
		return result, nil
	}, func(result TaskResult, err error) {
		if errors.Is(err, context.Canceled) {
			result.Outcome = "canceled"
			f.status = "Task canceled"
		} else if err != nil {
			result.Outcome = "failed"
			f.status = "Task failed"
		} else if result.CapturedRevision != f.Journal.Snapshot().Revision {
			result.Outcome = "stale"
			f.status = "Task stale"
		} else {
			result.Outcome = "complete"
			f.status = "Task complete"
		}
		f.results = append(f.results, result)
		f.touch()
	})
	if err != nil {
		return err
	}
	f.task = task
	f.nextTask = ordinal
	f.status = "Task pending"
	f.touch()
	return nil
}
