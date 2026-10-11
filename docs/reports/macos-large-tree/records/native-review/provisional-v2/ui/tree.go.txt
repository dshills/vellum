package ui

import "github.com/dshills/vellum/virtual"

// DataTree is a managed bounded viewport over an immutable indexed outline.
// It does not change the legacy virtual.Tree or driver VirtualTree APIs.
const DataTree Kind = "data-tree"

type TreeChangeReason string

const (
	TreeSelection TreeChangeReason = "selection"
	TreeViewport  TreeChangeReason = "viewport"
)

// TreeEvent reports an already accepted owner-held selection or viewport change.
// Callers must not apply that transition a second time.
type TreeEvent struct {
	ID     string
	Reason TreeChangeReason
	State  virtual.OutlineState
}

// TreeDisclosureEvent requests a worker-built expansion change. The native
// driver validates the captured row reference and originating lifetime first;
// it does not scan the tree or prepare an index on the owner. The application
// admits a bounded task, prepares the model request, builds the immutable index
// off-owner, and calls Outline.Complete before publishing its accepted state.
// It can refuse or cancel the request while the old visible index stays usable.
type TreeDisclosureEvent struct {
	ID     string
	Change virtual.OutlineChange
	Row    virtual.OutlineRow
	State  virtual.OutlineState
}
