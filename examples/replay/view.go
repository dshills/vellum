package replayconsumer

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/dshills/vellum/ui"
)

func (f *Fixture) uiAction(id, value string) {
	f.uiAttempts[id]++
	if err := f.Invoke(f.lifetime, Action{ID: id, Value: value}); err != nil {
		f.status = "Refused: " + err.Error()
		f.touch()
	}
}
func (f *Fixture) button(id, label string) ui.Node {
	// Background controls remain enabled while the modal is open, enabling a
	// native router test to prove the modal actually intercepts pointer/key input.
	// Registry availability independently refuses a bypassed background action.
	disabled := f.available(id, true) != nil
	return ui.Node{ID: id, Kind: ui.Button, Text: label, Disabled: disabled, Accessibility: ui.Accessibility{Name: label}, Activate: func() { f.uiAction(id, "") }}
}

// View declares managed public widgets. The table is four keyed pooled row
// buttons backed by virtual.Table, with bounded public Scroll/Row primitives.
// It requires no consumer-local native extension or callback dispatch bypass.
func (f *Fixture) View() ui.Node {
	rows := f.Table.View()
	tableChildren := []ui.Node{{ID: "table-header", Kind: ui.Row, Children: []ui.Node{{ID: "header-name", Kind: ui.Label, Text: "Name", Grow: true}, {ID: "header-priority", Kind: ui.Label, Text: "Priority"}}}}
	for i, row := range rows {
		index := i
		label := row.Cells[0] + "  " + row.Cells[1]
		if f.Table.Selected == row.ID {
			label = "→ " + label
		}
		tableChildren = append(tableChildren, ui.Node{ID: fmt.Sprintf("table-row-%d", i), Kind: ui.Button, Text: label, Grow: true, MinHeight: 24, Accessibility: ui.Accessibility{Name: row.Cells[0]}, Activate: func() { f.uiAction(SelectRow, strconv.Itoa(index)) }})
	}
	property := ui.Node{ID: f.PropertyID(), Kind: ui.TextField, Text: f.property, Disabled: !f.loaded || f.propertyDisabled, Grow: true, MinWidth: 260, Accessibility: ui.Accessibility{Name: "Project name"}, Change: func(event ui.ChangeEvent) { f.uiAction(EditProperty, event.Text) }}
	return ui.Node{ID: "root", Kind: ui.Column, Focus: f.focus, Children: []ui.Node{
		{ID: "toolbar", Kind: ui.Row, Children: []ui.Node{f.button(OpenProject, "Open fixture"), f.button(Undo, "Undo"), f.button(Redo, "Redo"), f.button(OpenModal, "Open modal")}},
		{ID: "properties", Kind: ui.Row, Children: []ui.Node{{ID: "property-label", Kind: ui.Label, Text: "Project"}, property, {ID: "property-disabled", Kind: ui.Checkbox, Text: "Disable name", Checked: f.propertyDisabled, Disabled: !f.loaded, Change: func(event ui.ChangeEvent) { f.uiAction(DisableProperty, strconv.FormatBool(event.Checked)) }}, f.button(RemountProperty, "Remount name")}},
		{ID: "table-tools", Kind: ui.Row, Children: []ui.Node{f.button(SortTable, "Sort names"), {ID: "row-count", Kind: ui.Label, Text: fmt.Sprintf("%d rows / %d pooled", f.Table.Len(), f.Table.VisibleCount())}}},
		{ID: "table", Kind: ui.Scroll, Grow: true, MinHeight: 150, Children: tableChildren},
		{ID: "task-tools", Kind: ui.Row, Children: []ui.Node{f.button(StartTask, "Start task"), f.button(CancelTask, "Cancel task")}},
		{ID: "status", Kind: ui.Label, Text: f.status, Accessibility: ui.Accessibility{Name: "Status"}},
		{ID: "dialog", Kind: ui.Dialog, Open: f.modal, Text: "Replay modal", Dismiss: func() { f.uiAction(CloseModal, "") }, Children: []ui.Node{{ID: "dialog-text", Kind: ui.Label, Text: "Background actions must be intercepted by the real router."}, f.button(CloseModal, "Close modal")}},
	}}
}

type Semantic struct {
	Grow                    bool
	MinWidth, MinHeight     float32
	ID                      string
	Kind                    ui.Kind
	Text, Name              string
	Disabled, Open, Checked bool
	Children                []string
}
type Counter struct {
	ID    string
	Count int
}
type TaskState struct {
	Present, Running, AwaitingOwner, WorkerDone, OwnerDone bool
	Progress                                               float64
	Message, Error                                         string
}
type Snapshot struct {
	DeclaredViewport                       [2]int
	Loaded, Closed                         bool
	Property                               string
	PropertyID                             string
	PropertyIncarnation, ViewEpoch         uint64
	HistoryRevision                        uint64
	UndoCount, RedoCount                   int
	Dirty, Sorted, Modal, PropertyDisabled bool
	Selected                               string
	VisibleIDs                             []string
	VisibleCells                           [][]string
	TotalRows                              int
	Status                                 string
	Task                                   TaskState
	Results                                []TaskResult
	WorkerCleanups                         uint64
	Semantics                              []Semantic
	UIAttempts                             []Counter
	DeclaredFocus                          ui.FocusRequest
}

// Snapshot contains owner model and declared semantics, not native focus,
// measured layout, compositor completion, OS accessibility, or image pixels.
// Native acceptance must append actual post-frame captures from its backend.
func (f *Fixture) Snapshot() Snapshot {
	history := f.Journal.Snapshot()
	s := Snapshot{Loaded: f.loaded, Closed: f.lifetime.Err() != nil, Property: f.property, PropertyID: f.PropertyID(), PropertyIncarnation: f.propertyMount, ViewEpoch: f.viewEpoch, HistoryRevision: history.Revision, UndoCount: history.UndoCount, RedoCount: history.RedoCount, Dirty: history.Dirty, Sorted: f.sorted, Modal: f.modal, PropertyDisabled: f.propertyDisabled, Selected: f.Table.Selected, TotalRows: f.Table.Len(), Status: f.status, Results: append([]TaskResult{}, f.results...), WorkerCleanups: f.cleanups.Load(), DeclaredFocus: f.focus, UIAttempts: []Counter{}}
	s.DeclaredViewport = [2]int{ViewportWidth, ViewportHeight}
	for _, row := range f.Table.View() {
		s.VisibleIDs = append(s.VisibleIDs, row.ID)
		s.VisibleCells = append(s.VisibleCells, append([]string{}, row.Cells...))
	}
	if f.task != nil {
		state := f.task.Snapshot()
		failure := ""
		if state.Err != nil {
			failure = state.Err.Error()
		}
		s.Task = TaskState{Present: true, Running: state.Running, AwaitingOwner: state.AwaitingOwner, WorkerDone: state.WorkerDone, OwnerDone: state.OwnerDone, Progress: state.Progress, Message: state.Message, Error: failure}
	}
	var walk func(ui.Node)
	walk = func(node ui.Node) {
		v := Semantic{ID: node.ID, Kind: node.Kind, Text: node.Text, Name: node.Accessibility.Name, Disabled: node.Disabled, Open: node.Open, Checked: node.Checked, Children: []string{}}
		v.Grow, v.MinWidth, v.MinHeight = node.Grow, node.MinWidth, node.MinHeight
		for _, child := range node.Children {
			v.Children = append(v.Children, child.ID)
		}
		s.Semantics = append(s.Semantics, v)
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(f.View())
	sort.Slice(s.Semantics, func(i, j int) bool { return s.Semantics[i].ID < s.Semantics[j].ID })
	for id, count := range f.uiAttempts {
		s.UIAttempts = append(s.UIAttempts, Counter{id, count})
	}
	sort.Slice(s.UIAttempts, func(i, j int) bool { return s.UIAttempts[i].ID < s.UIAttempts[j].ID })
	return s
}

// Validate checks both the public descriptor and the command presentation
// agreement, without invoking Change/Activate or pretending a routed event.
func (f *Fixture) Validate() error {
	root := f.View()
	if err := ui.Validate(root); err != nil {
		return err
	}
	nodes := map[string]ui.Node{}
	var visit func(ui.Node)
	visit = func(node ui.Node) {
		nodes[node.ID] = node
		for _, child := range node.Children {
			visit(child)
		}
	}
	visit(root)
	for _, id := range []string{StartTask, CancelTask, Undo, Redo} {
		availability := f.Commands.Enabled(id)
		node, ok := nodes[id]
		if !ok || node.Kind != ui.Button || !f.modal && node.Disabled != (availability != nil) {
			return fmt.Errorf("consumer: command presentation diverged %s", id)
		}
	}
	active := 0
	if f.activeTask() {
		active = 1
	}
	if f.manager.Pending().Active != active {
		return fmt.Errorf("consumer: task limit disagrees with terminal state")
	}
	return nil
}
