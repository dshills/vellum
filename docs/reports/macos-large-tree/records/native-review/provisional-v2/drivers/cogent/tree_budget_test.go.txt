package cogent

import (
	"errors"
	"strings"
	"testing"

	"github.com/dshills/vellum/virtual"
)

func TestTreeNativePayloadBudgetIncludesGeneratedIdentities(t *testing.T) {
	rectangle := virtual.OutlineRect{Width: 200, Height: 23}
	row := virtual.OutlineRow{TreeRow: virtual.TreeRow{TreeNode: virtual.TreeNode{ID: strings.Repeat("k", 64<<10), Label: "Short"}}, Rect: rectangle, Clip: rectangle}
	if e := treeProjectionBudget("outline", virtual.OutlineSnapshot{Rows: []virtual.OutlineRow{row}}); !errors.Is(e, ErrTreeBudget) {
		t.Fatal("long key bypassed native payload budget", e)
	}
	row.ID = "short"
	row.Parent = strings.Repeat("p", treeNativeMaxKey+1)
	if e := treeProjectionBudget("outline", virtual.OutlineSnapshot{Rows: []virtual.OutlineRow{row}}); !errors.Is(e, ErrTreeBudget) {
		t.Fatal("long parent bypassed native payload budget", e)
	}
	row.Parent = ""
	row.Label = strings.Repeat("&", 32<<10)
	rows := make([]virtual.OutlineRow, 100)
	for i := range rows {
		rows[i] = row
	}
	// Raw labels fit 8MiB but escaped/native/semantic payload does not.
	if len(row.Label)*len(rows) >= treeNativeMaxText {
		t.Fatal("invalid budget oracle")
	}
	if e := treeProjectionBudget("outline", virtual.OutlineSnapshot{Rows: rows}); !errors.Is(e, ErrTreeBudget) {
		t.Fatal("cumulative escaped payload admitted", e)
	}
	row.Label = "& safe < label"
	rows = []virtual.OutlineRow{row}
	if e := treeProjectionBudget("outline", virtual.OutlineSnapshot{Rows: rows}); e != nil {
		t.Fatal("bounded payload refused", e)
	}
}
func TestTreeNavigationBoundariesAreNoop(t *testing.T) {
	for _, e := range []error{virtual.ErrTreeRange, virtual.ErrTreeHidden, nil} {
		if e := treeNavigation(func() error { return e }); e != nil {
			t.Fatal(e)
		}
	}
	if e := treeNavigation(func() error { return virtual.ErrTreeStale }); !errors.Is(e, virtual.ErrTreeStale) {
		t.Fatal("genuine native state error swallowed", e)
	}
}
