package ui

import "github.com/dshills/vellum/virtual"

// DataGrid is a managed, fixed-height two-dimensional viewport.
const DataGrid Kind = "data-grid"

type GridChangeReason string

const (
	GridSelection GridChangeReason = "selection"
	GridViewport  GridChangeReason = "viewport"
)

// GridEvent reports an already accepted owner-held transition. It never asks
// the caller to apply the selection or viewport a second time.
type GridEvent struct {
	ID     string
	Reason GridChangeReason
	State  virtual.GridState
}
