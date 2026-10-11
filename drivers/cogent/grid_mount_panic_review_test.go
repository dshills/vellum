//go:build offscreen

package cogent

import (
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"testing"
)

type reviewPanickingGridSource struct {
	nativeGridSource
	fault bool
}

func (s *reviewPanickingGridSource) Cell(row, col string) (string, error) {
	if s.fault {
		panic("grid-source-fault")
	}
	return s.nativeGridSource.Cell(row, col)
}
func TestReviewGridMountPanicRetiresRegistry(t *testing.T) {
	a, host := gridContractOwner(t, 8)
	source := &reviewPanickingGridSource{nativeGridSource: nativeGridSource{rows: 100, revision: 1}}
	g, err := virtual.NewGrid(source, []virtual.GridColumn{{ID: "c", Width: 80}}, virtual.GridConfig{RowHeight: 24}, virtual.GridViewport{Width: 80, Height: 72})
	if err != nil {
		t.Fatal(err)
	}
	source.fault = true
	func() {
		defer func() {
			if got := recover(); got != "grid-source-fault" {
				t.Fatalf("wrong panic %#v", got)
			}
		}()
		_ = a.Do(func() { _, _ = reconcile(host, nil, ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g}, a, "one") })
	}()
	if a.Context().Err() == nil {
		t.Fatal("owner failure did not close application")
	}
	gridRegistry.Lock()
	old := gridRegistry.models[g]
	gridRegistry.Unlock()
	defer func() {
		if old != nil {
			disposeGrid(old.r)
		}
	}()
	if old != nil {
		t.Errorf("failed Grid mount retained native registry claim closed=%v", old.closed.Load())
	}
	source.fault = false
	fresh, newHost := gridContractOwner(t, 8)
	var mountErr error
	var freshRecord *record
	defer func() {
		if freshRecord != nil {
			disposeControl(freshRecord)
		}
	}()
	if err = fresh.Do(func() {
		freshRecord, mountErr = reconcile(newHost, nil, ui.Node{ID: "remount", Kind: ui.DataGrid, Grid: g}, fresh, "one")
	}); err != nil {
		t.Fatal(err)
	}
	if mountErr != nil {
		t.Errorf("valid caller-owned Grid cannot bind after failed old mount: %v", mountErr)
	}
}
