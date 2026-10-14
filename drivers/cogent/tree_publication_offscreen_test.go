//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"errors"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/workbench"
	"reflect"
	"testing"
)

func TestManagedTreePendingSceneLayoutRevokesGeometryAndAX(t *testing.T) {
	model := nativeOutline(t, 1000, true)
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node { return ui.Node{ID: "outline", Kind: ui.DataTree, Tree: model, Grow: true} })
	unitSettle(t, m)
	binding := m.implementation.find("one").state.root.control.outline
	bridge := &accessibilityBridge{}
	if err := a.Do(func() {
		nodes, handlers := accessibilitySnapshot(binding.r, nil)
		ready, _ := bridge.prepareGridPublication(nodes, handlers)
		if len(ready) < 3 {
			t.Fatal("fixture lacks completed rows")
		}
		var oldAction accessibilityHandler
		for _, node := range ready {
			if node.Role == "tree-row" {
				oldAction = handlers[node.ID]
				break
			}
		}
		if oldAction == nil {
			t.Fatal("fixture lacks row action")
		}
		retained := binding.measurement.Load()
		binding.host.Scene.NeedsLayout()
		if _, err := InspectTree(a, model); !errors.Is(err, ErrTreeLayout) {
			t.Errorf("pending full layout returned stale completed geometry: %v", err)
		}
		if _, ok := binding.refAt(retained.Rows[0].Bounds.Min); ok {
			t.Error("pending layout accepted stale hit geometry")
		}
		if oldAction(accessibilityAction{Kind: "press"}) {
			t.Error("queued old AX action ran during layout")
		}
		if binding.measurement.Load() != retained {
			t.Error("pending layout erased last-readable geometry")
		}
		pending, pendingHandlers := accessibilitySnapshot(binding.r, nil)
		out, actions := bridge.prepareGridPublication(pending, pendingHandlers)
		if len(out) != len(ready) {
			t.Errorf("pending layout discarded readable AX rows: %d -> %d", len(ready), len(out))
		}
		for _, n := range out {
			if !n.GridPending || !n.Disabled || n.Focused || len(n.Actions) != 0 || actions[n.ID] != nil {
				t.Error("pending tree retains interactive AX target", n.ID)
			}
		}
	}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if err := a.Do(func() {
		if _, err := InspectTree(a, model); err != nil {
			t.Fatal("completed layout not restored", err)
		}
		nodes, handlers := accessibilitySnapshot(binding.r, nil)
		ready, actions := bridge.prepareGridPublication(nodes, handlers)
		for _, n := range ready {
			if n.GridPending || n.GridFrame == nil || n.GridGeneration == 0 {
				t.Error("completed tree omitted atomic frame/provenance", n.ID)
			}
			if n.Role == "tree-row" && actions[n.ID] == nil {
				t.Error("completed row action not restored")
			}
		}
	}); err != nil {
		t.Fatal(err)
	}
}

func TestTreePublicationCacheSeparatesScopeKinds(t *testing.T) {
	b := &accessibilityBridge{}
	ready := gridAXCompleted("scope", "", 17, 1)
	ready[0].Role = "outline"
	b.prepareGridPublication(ready, nil)
	next := gridAXPending("scope", "", 17, 1)
	out, _ := b.prepareGridPublication(next, nil)
	if len(out) != 1 || len(b.gridPublications) != 0 || out[0].Role != "table" {
		t.Fatal("independent tree/grid incarnation counters reused an old scope", out)
	}
}

// Every exported publication field except the render ordinal must affect the
// completion comparison, including fields added to TreeInspection in future.
func mutateTreePublicationField(v reflect.Value) {
	switch v.Kind() {
	case reflect.String:
		v.SetString(v.String() + "changed")
	case reflect.Bool:
		v.SetBool(!v.Bool())
	case reflect.Int, reflect.Int64:
		v.SetInt(v.Int() + 1)
	case reflect.Uint64:
		v.SetUint(v.Uint() + 1)
	case reflect.Float32:
		v.SetFloat(v.Float() + 1)
	case reflect.Struct:
		mutateTreePublicationField(v.Field(0))
	case reflect.Slice:
		next := reflect.MakeSlice(v.Type(), 1, 1)
		mutateTreePublicationField(next.Index(0))
		v.Set(next)
	default:
		panic("new publication field kind requires an explicit mutation")
	}
}
func TestTreeCompletedPublicationIncludesAllDataAndGeometry(t *testing.T) {
	base := TreeInspection{}
	kind := reflect.TypeOf(base)
	for i := 0; i < kind.NumField(); i++ {
		name := kind.Field(i).Name
		t.Run(name, func(t *testing.T) {
			next := base
			mutateTreePublicationField(reflect.ValueOf(&next).Elem().Field(i))
			equal := sameTreeAccessibilityMeasurement(&base, &next)
			if equal != (name == "Frames") {
				t.Fatal("publication comparison omitted data/geometry or included render ordinal", name)
			}
		})
	}
}

// Inert ownership sentinels represent a completed old scene; this does not
// claim native frames. Exercise actual dialog host transfers before new paint.
func TestModalTransfersWithdrawAllVirtualViewGeometry(t *testing.T) {
	for _, closeMode := range []string{"native-close", "programmatic-close"} {
		t.Run(closeMode, func(t *testing.T) {
			a, host := gridContractOwner(t, 16)
			outline := nativeOutline(t, 1000, true)
			grid, _ := testNativeGrid(t, 1000)
			r := mountGridContract(t, a, host, ui.Node{ID: "dialog", Kind: ui.Dialog, Open: true, Children: []ui.Node{
				{ID: "tree", Kind: ui.DataTree, Tree: outline}, {ID: "grid", Kind: ui.DataGrid, Grid: grid},
			}})
			t.Cleanup(func() { disposeControlTree(r) })
			treeBinding := findRecord(r, "tree").control.outline
			gridBinding := findRecord(r, "grid").control.grid
			seed := func() {
				treeBinding.measurement.Store(&TreeInspection{Generation: 1})
				treeBinding.measurementScene.Store(treeBinding.host.Scene)
				gridBinding.measurement.Store(&GridInspection{Generation: 1})
				gridBinding.measurementScene.Store(gridBinding.host.Scene)
				gridBinding.layout.valid = true
			}
			withdrawn := func() {
				if treeBinding.measurement.Load() != nil || treeBinding.measurementScene.Load() != nil {
					t.Error("modal transfer retains former tree completion/scene")
				}
				if gridBinding.measurement.Load() != nil || gridBinding.measurementScene.Load() != nil || !gridBinding.layoutPending.Load() || gridBinding.layout.valid {
					t.Error("modal transfer retains former grid completion/scene/layout")
				}
			}
			if err := a.Do(func() {
				original := treeBinding.host.Scene
				seed()
				body := prepareDialogBody(r)
				body.NewDialog(r.widget)
				r.control.dialog.active = true
				if treeBinding.host.Scene == original {
					t.Fatal("fixture did not transfer actual modal child host")
				}
				withdrawn()
				seed()
				if closeMode == "native-close" {
					body.Send(events.Close)
				} else {
					closeDialog(r)
				}
				if treeBinding.host.Scene != original {
					t.Fatal("close did not return retained children to origin")
				}
				withdrawn()
			}); err != nil {
				t.Fatal(err)
			}
			a.Drain()
		})
	}
}

func TestRetainedAncestorMoveWithdrawsDescendantDataGeometry(t *testing.T) {
	a, host := gridContractOwner(t, 16)
	outline := nativeOutline(t, 1000, true)
	grid, _ := testNativeGrid(t, 1000)
	pane := ui.Node{ID: "pane", Kind: ui.Column, Children: []ui.Node{{ID: "tree", Kind: ui.DataTree, Tree: outline}, {ID: "grid", Kind: ui.DataGrid, Grid: grid}}}
	node := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{pane, {ID: "destination", Kind: ui.Column}}}
	r := mountGridContract(t, a, host, node)
	t.Cleanup(func() { disposeControlTree(r) })
	tb := findRecord(r, "tree").control.outline
	gb := findRecord(r, "grid").control.grid
	oldTree, oldGrid := tb.host, gb.host
	if err := a.Do(func() {
		tb.measurement.Store(&TreeInspection{Generation: 1})
		tb.measurementScene.Store(tb.host.Scene)
		gb.measurement.Store(&GridInspection{Generation: 1})
		gb.measurementScene.Store(gb.host.Scene)
		gb.layout.valid = true
		node.Children = []ui.Node{{ID: "destination", Kind: ui.Column, Children: []ui.Node{pane}}}
		var err error
		r, err = reconcile(host, r, node, a, "one")
		if err != nil {
			t.Fatal(err)
		}
		if findRecord(r, "tree").widget != oldTree || findRecord(r, "grid").widget != oldGrid {
			t.Fatal("ancestor move remounted retained views")
		}
		if tb.measurement.Load() != nil || tb.measurementScene.Load() != nil || gb.measurement.Load() != nil || gb.measurementScene.Load() != nil || gb.layout.valid || !gb.layoutPending.Load() {
			t.Fatal("ancestor move retained obsolete descendant geometry")
		}
	}); err != nil {
		t.Fatal(err)
	}
}

func TestWorkbenchParkingRevokesGeometryBeforePanelView(t *testing.T) {
	a, host := gridContractOwner(t, 16)
	outline := nativeOutline(t, 1000, true)
	grid, _ := testNativeGrid(t, 1000)
	model, err := workbench.New(workbench.Layout{Version: 1, Root: &workbench.Group{ID: "group", Tabs: []string{"data", "peer"}, Active: "data"}}, []string{"data", "peer"})
	if err != nil {
		t.Fatal(err)
	}
	var tb *outlineBinding
	var gb *gridBinding
	check, observed := false, false
	w := &Workbench{Owner: a, Window: "one", Workspace: model, Panels: []Panel{
		{ID: "data", Title: "Data", View: func() ui.Node {
			if check {
				observed = true
				if tb.measurement.Load() != nil || tb.measurementScene.Load() != nil || gb.measurement.Load() != nil || gb.measurementScene.Load() != nil {
					t.Error("parked data geometry remains visible to panel.View")
				}
			}
			return ui.Node{ID: "data-root", Kind: ui.Column, Children: []ui.Node{{ID: "tree", Kind: ui.DataTree, Tree: outline}, {ID: "grid", Kind: ui.DataGrid, Grid: grid}}}
		}},
		{ID: "peer", Title: "Peer", View: func() ui.Node { return ui.Node{ID: "peer", Kind: ui.Label} }},
	}}
	widget := w.Mount(host)
	t.Cleanup(w.disposeControlRecords)
	if err := a.Do(func() {
		if err := w.Sync(widget); err != nil {
			t.Fatal(err)
		}
		tb = findRecord(w.records["data"], "tree").control.outline
		gb = findRecord(w.records["data"], "grid").control.grid
		tb.measurement.Store(&TreeInspection{Generation: 1})
		tb.measurementScene.Store(tb.host.Scene)
		gb.measurement.Store(&GridInspection{Generation: 1})
		gb.measurementScene.Store(gb.host.Scene)
		if err := model.Move("data", "group", 1); err != nil {
			t.Fatal(err)
		}
		check = true
		if err := w.Sync(widget); err != nil {
			t.Fatal(err)
		}
		if !observed {
			t.Fatal("fixture did not call retained panel view during parking")
		}
	}); err != nil {
		t.Fatal(err)
	}
}
