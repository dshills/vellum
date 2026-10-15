package workspace

import (
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/virtual"
	"github.com/dshills/vellum/workbench"
)

type TaskSnapshot struct {
	Kind, Error                                                      string
	Present, Running, AwaitingOwner, WorkerDone, OwnerDone, Canceled bool
}
type Snapshot struct {
	Loaded, Sorted, Filtered, Modal            bool
	Name, Status, Failure, DataSHA256          string
	StoredRows, StoredColumns, StoredTreeNodes int
	HistoryRevision                            uint64
	History                                    history.State
	Grid                                       virtual.GridState
	Tree                                       virtual.OutlineState
	Projection                                 virtual.GridProjectionStats
	FirstKey, FirstValue, ActiveValue          string
	Editor                                     editor.Snapshot
	Canvas                                     canvas.Scene
	Layout                                     workbench.Layout
	Task                                       TaskSnapshot
	Outcomes                                   []outcome
	Attempts                                   map[string]int
}

func (f *Fixture) Snapshot() Snapshot {
	failure := f.failure
	grid, e := f.Grid.State()
	if e != nil {
		failure = e.Error()
	}
	tree, e := f.Tree.State()
	if e != nil {
		failure = e.Error()
	}
	// Generation is an input lease affected by internal viewport paint batching;
	// stable domain revisions, viewport, cursor, selection and content are retained.
	grid.Generation = 0
	tree.Generation = 0
	h := f.Journal.Snapshot()
	s := Snapshot{Loaded: f.loaded, Sorted: f.sorted, Filtered: f.filtered, Modal: f.modal, Name: f.name, Status: f.status, Failure: failure, DataSHA256: f.data.digest, StoredRows: len(f.data.rows), StoredColumns: Columns, HistoryRevision: h.Revision, History: h, Grid: grid, Tree: tree, Editor: f.Editor.Snapshot(), Canvas: *f.Canvas, Layout: f.Dock.Workspace.Layout(), Outcomes: append([]outcome(nil), f.outcomes...), Attempts: map[string]int{}}
	s.Canvas.Objects = append([]canvas.Rect(nil), f.Canvas.Objects...)
	if f.treeData != nil {
		s.StoredTreeNodes = f.treeData.Stats().Nodes
	}
	if f.projection != nil {
		s.Projection = f.projection.Stats()
		if f.projection.RowCount() > 0 {
			s.FirstKey, _ = f.projection.RowKey(0)
			s.FirstValue, _ = f.projection.Cell(s.FirstKey, "c00")
		}
	}
	if grid.Active.RowKey != "" {
		s.ActiveValue, _ = f.data.Cell(grid.Active.RowKey, grid.Active.ColumnID)
	}
	if f.task != nil {
		task := f.task.Snapshot()
		s.Task = TaskSnapshot{Kind: f.job, Present: true, Running: task.Running, AwaitingOwner: task.AwaitingOwner, WorkerDone: task.WorkerDone, OwnerDone: task.OwnerDone, Canceled: f.canceled}
		if task.Err != nil {
			s.Task.Error = task.Err.Error()
		}
	}
	for k, v := range f.attempts {
		s.Attempts[k] = v
	}
	return s
}
