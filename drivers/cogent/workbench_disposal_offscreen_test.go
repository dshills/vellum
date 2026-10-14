//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"context"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"github.com/dshills/vellum/workbench"
	"testing"
)

func TestWorkbenchDisposalReleasesNestedModelBindings(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.Open(app.Window{ID: "one", Title: "Fixture", Width: 800, Height: 600, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	model, err := editor.New("Nested editor", editor.Config{})
	if err != nil {
		t.Fatal(err)
	}
	grid, _ := testNativeGrid(t, 64)
	data, err := virtual.NewTreeData(context.Background(), "disposal-tree", 1, []virtual.TreeNode{{ID: "item", Label: "Item"}}, virtual.TreeLimits{})
	if err != nil {
		t.Fatal(err)
	}
	index, err := virtual.BuildTreeIndex(context.Background(), data, virtual.TreeIndexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	outline, err := virtual.NewOutline(index, virtual.OutlineConfig{}, virtual.OutlineViewport{Width: 100, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	newDock := func(id string, view func() ui.Node) *Workbench {
		layout, err := workbench.New(workbench.Layout{Version: 1, Root: &workbench.Group{ID: id, Tabs: []string{"panel"}, Active: "panel"}}, []string{"panel"})
		if err != nil {
			t.Fatal(err)
		}
		return &Workbench{Owner: a, Window: "one", Workspace: layout, Panels: []Panel{{ID: "panel", Title: "Panel", View: view}}}
	}
	inner := newDock("inner", func() ui.Node {
		return ui.Node{ID: "content", Kind: ui.Column, Children: []ui.Node{{ID: "editor", Kind: ui.MultilineText, Editor: model}, {ID: "grid", Kind: ui.DataGrid, Grid: grid}, {ID: "tree", Kind: ui.DataTree, Tree: outline}}}
	})
	outer := newDock("outer", func() ui.Node { return ui.Node{ID: "nested", Kind: ui.Extension, Custom: inner} })
	widget := outer.Mount(core.NewBody("Fixture"))
	if err = outer.Sync(widget); err != nil {
		t.Fatal(err)
	}
	if err = inner.Sync(inner.host); err != nil {
		t.Fatal(err)
	}
	r := &record{node: ui.Node{ID: "workbench", Kind: ui.Extension, Custom: outer}, widget: widget}
	if err = bindEditorModels(r); err != nil {
		t.Fatal(err)
	}
	if editorBindings[model] == nil || gridRegistry.models[grid] == nil || outlineRegistry.models[outline] == nil {
		t.Fatal("fixture did not bind private nested models")
	}
	disposeControlTree(r)
	disposeControlTree(r) // normal cleanup can be observed again on failure unwind
	if editorBindings[model] != nil || gridRegistry.models[grid] != nil || outlineRegistry.models[outline] != nil {
		t.Fatal("nested model registry retained closed widget graph")
	}
	if outer.records != nil || inner.records != nil {
		t.Fatal("private panel record graph retained")
	}
	widget.AsTree().Delete()
}

func TestWorkbenchDisposalFailedRebuildReleasesUnadoptedPanel(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	if err := a.Open(app.Window{ID: "one", Title: "Fixture", Width: 800, Height: 600, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	model, _ := editor.New("Unadopted editor", editor.Config{})
	layout := workbench.Layout{Version: 1, Root: &workbench.Group{ID: "old", Tabs: []string{"early", "late"}, Active: "late"}}
	workspace, err := workbench.New(layout, []string{"early", "late"})
	if err != nil {
		t.Fatal(err)
	}
	bad := false
	dock := &Workbench{Owner: a, Window: "one", Workspace: workspace, Panels: []Panel{
		{ID: "early", Title: "Early", View: func() ui.Node {
			kind := ui.Label
			if bad {
				kind = "invalid"
			}
			return ui.Node{ID: "early", Kind: kind}
		}},
		{ID: "late", Title: "Late", View: func() ui.Node { return ui.Node{ID: "editor", Kind: ui.MultilineText, Editor: model} }},
	}}
	widget := dock.Mount(core.NewBody("Fixture"))
	if err = dock.Sync(widget); err != nil {
		t.Fatal(err)
	}
	root := &record{node: ui.Node{ID: "dock", Kind: ui.Extension, Custom: dock}, widget: widget}
	if err = bindEditorModels(root); err != nil || editorBindings[model] == nil {
		t.Fatal("fixture did not publish old active editor", err)
	}
	bad = true
	layout.Root.ID = "new"
	if err = workspace.Restore(layout); err != nil {
		t.Fatal(err)
	}
	if err = dock.Sync(widget); err == nil {
		t.Fatal("invalid early panel did not stop rebuild")
	}
	if editorBindings[model] != nil {
		t.Fatal("failed rebuild retained unadopted panel binding")
	}
	if dock.records != nil {
		t.Fatal("failed build kept partial record map")
	}
	widget.AsTree().Delete()
}

func TestWorkbenchDisposalRevokesQueuedAndLaterInteraction(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	if err := a.Open(app.Window{ID: "one", Title: "One", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	dock := &Workbench{Owner: a, Window: "one"}
	called := 0
	dock.queue(func() error { called++; return nil })
	disposeControlTree(&record{node: ui.Node{ID: "dock", Kind: ui.Extension, Custom: dock}})
	dock.queue(func() error { called++; return nil })
	a.Drain()
	if called != 0 || !a.Pending().Idle() {
		t.Fatal("retired dock interaction delivered", called)
	}
}
