//go:build offscreen

package cogent

import (
	"testing"

	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

type identityReviewGridSource struct {
	*nativeGridSource
	identity string
}

func (s identityReviewGridSource) Identity() string { return s.identity }
func (s identityReviewGridSource) Cell(row, col string) (string, error) {
	value, err := s.nativeGridSource.Cell(row, col)
	return s.identity + ": " + value, err
}

// Generations are local to a model. Reconciliation must remount a changed
// model pointer before syncGrid can consider its completed projection reusable.
func TestManagedGridEqualGenerationModelReplacementRemounts(t *testing.T) {
	a, host := gridContractOwner(t, 8)
	makeModel := func(identity string) (*virtual.Grid, identityReviewGridSource) {
		source := identityReviewGridSource{nativeGridSource: &nativeGridSource{rows: 100, revision: 1}, identity: identity}
		model, err := virtual.NewGrid(source, []virtual.GridColumn{{ID: "c", Width: 90}},
			virtual.GridConfig{RowHeight: 24, HeaderHeight: 26}, virtual.GridViewport{Width: 300, Height: 180})
		if err != nil {
			t.Fatal(err)
		}
		return model, source
	}
	first, _ := makeModel("first")
	second, source := makeModel("second")
	old := mountGridContract(t, a, host, ui.Node{ID: "g", Kind: ui.DataGrid, Grid: first})
	oldBinding := old.control.grid
	var next *record
	var reconcileErr error
	if err := a.Do(func() {
		state, err := second.State()
		if err != nil || state.Generation != oldBinding.projection.State.Generation {
			t.Errorf("fixture must replace a different model with equal generation: %v", err)
			return
		}
		next, reconcileErr = reconcile(host, old, ui.Node{ID: "g", Kind: ui.DataGrid, Grid: second}, a, "one")
		if reconcileErr != nil {
			return
		}
		binding := next.control.grid
		if next == old || binding == oldBinding || binding.model != second || !old.control.retired || !oldBinding.closed.Load() || !binding.projectionReady {
			t.Error("equal-generation replacement did not retire and remount the model binding")
		}
		if source.cells == 0 || len(binding.projection.Cells) == 0 || binding.projection.Cells[0].Text != "second: row-0 · c" || binding.cells[0].Text != "second: row-0 · c" {
			t.Error("equal-generation replacement kept cells from the previous model")
		}
		if len(host.Children) != 1 || host.Children[0] != next.widget || gridRegistry.models[first] != nil || gridRegistry.models[second] != binding {
			t.Error("model replacement left stale native children or model registration")
		}
	}); err != nil {
		t.Fatal(err)
	}
	if next != nil {
		t.Cleanup(func() { disposeControl(next) })
	}
	if reconcileErr != nil {
		t.Fatal(reconcileErr)
	}
}
