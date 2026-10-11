package ui

import (
	"testing"

	"github.com/dshills/vellum/virtual"
)

type gridUISource struct{}

func (gridUISource) Identity() string                    { return "ui-test" }
func (gridUISource) Revision() uint64                    { return 1 }
func (gridUISource) RowCount() int                       { return 1 }
func (gridUISource) RowKey(int) (string, error)          { return "a", nil }
func (gridUISource) IndexOf(string) (int, bool, error)   { return 0, true, nil }
func (gridUISource) Cell(string, string) (string, error) { return "A", nil }
func TestDataGridDescriptor(t *testing.T) {
	g, e := virtual.NewGrid(gridUISource{}, []virtual.GridColumn{{ID: "c", Width: 80}}, virtual.GridConfig{RowHeight: 24}, virtual.GridViewport{Width: 80, Height: 100})
	if e != nil {
		t.Fatal(e)
	}
	n := Node{ID: "g", Kind: DataGrid, Grid: g}
	if e = Validate(n); e != nil {
		t.Fatal(e)
	}
	if !IsFocusable(DataGrid) {
		t.Fatal("grid not focusable")
	}
	n.Kind = Label
	if e = Validate(n); e == nil {
		t.Fatal("foreign grid property accepted")
	}
	n.Kind = DataGrid
	n.Grid = nil
	if e = Validate(n); e == nil {
		t.Fatal("nil grid accepted")
	}
	n.Grid = g
	n.Children = []Node{{ID: "x", Kind: Label}}
	if e = Validate(n); e == nil {
		t.Fatal("grid children accepted")
	}
	n.Children = nil
	root := Node{ID: "root", Kind: Column, Children: []Node{n, {ID: "other", Kind: DataGrid, Grid: g}}}
	if e = Validate(root); e == nil {
		t.Fatal("duplicate binding accepted")
	}
}

func TestDataGridGeneratedNamespaceReserved(t *testing.T) {
	for _, id := range []string{"grid:g", "grid:Zw:cell:cm93LTA:YzA", "grid:Zw:header:YzA", "grid:Zw:row:cm93LTA"} {
		t.Run(id, func(t *testing.T) {
			// The bounded prefix check also applies before a grid is visible or
			// mounted and to nested/modal descriptors; no source row scan is needed.
			root := Node{ID: "root", Kind: Column, Children: []Node{{ID: "dialog", Kind: Dialog, Children: []Node{{ID: id, Kind: Button}}}}}
			if err := Validate(root); err == nil {
				t.Fatal("generated identity admitted as ordinary descriptor", id)
			}
		})
	}
	if err := Validate(Node{ID: "grid", Kind: Button}); err != nil {
		t.Fatal("unreserved ID refused", err)
	}
}
