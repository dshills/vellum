//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/workbench"
	"math"
	"testing"
)

func TestCanvasRetainedAncestorMoveWithdrawsCompletedFrame(t *testing.T) {
	scene, _ := canvas.New([]canvas.Rect{{ID: "object", Width: 40, Height: 30}})
	moved := false
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		pane := ui.Node{ID: "pane", Kind: ui.Column, Grow: true, Children: []ui.Node{{ID: "drawing", Kind: ui.Canvas2D, Canvas: scene, CanvasObjects: []ui.CanvasObject{{ID: "object", Name: "Object"}}, Grow: true}}}
		destination := ui.Node{ID: "destination", Kind: ui.Column, Grow: true}
		children := []ui.Node{pane, destination}
		if moved {
			destination.Children = []ui.Node{pane}
			children = []ui.Node{destination}
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: children}
	})
	gridFrame(t, m)
	native := m.implementation.find("one")
	r := findRecord(native.state.root, "drawing")
	b := r.control.placement
	if b.frame.Load() == nil {
		t.Fatal("no actual software completion")
	}
	var old accessibilityHandler
	a.Do(func() {
		_, h := accessibilitySnapshot(native.state.root, nil)
		old = h[accessibilityManagedID(CanvasObjectID("drawing", "object"))]
	})
	a.Do(func() { moved = true; a.Invalidate("one") })
	if err := m.implementation.refresh(native); err != nil {
		t.Fatal(err)
	}
	if findRecord(native.state.root, "drawing") != r {
		t.Fatal("ancestor move remounted canvas")
	}
	if b.frame.Load() != nil {
		t.Fatal("retained ancestor move kept old completion")
	}
	a.Do(func() {
		if old(accessibilityAction{Kind: "press"}) {
			t.Error("old action survived transfer")
		}
	})
	gridFrame(t, m)
	a.Do(func() {
		_, h := accessibilitySnapshot(native.state.root, nil)
		if current := h[accessibilityManagedID(CanvasObjectID("drawing", "object"))]; current == nil || !current(accessibilityAction{Kind: "press"}) {
			t.Error("destination completion did not restore action")
		}
	})
}

func TestCanvasDialogTransfersWithdrawBeforeNewPaint(t *testing.T) {
	a, host := gridContractOwner(t, 16)
	scene, _ := canvas.New([]canvas.Rect{{ID: "object", Width: 40, Height: 30}})
	r := mountGridContract(t, a, host, ui.Node{ID: "dialog", Kind: ui.Dialog, Open: true, Children: []ui.Node{{ID: "drawing", Kind: ui.Canvas2D, Canvas: scene, CanvasObjects: []ui.CanvasObject{{ID: "object", Name: "Object"}}}}})
	t.Cleanup(func() { disposeControlTree(r) })
	b := findRecord(r, "drawing").control.placement
	// Identity sentinels test the production dialog transfer hooks, not geometry.
	a.Do(func() {
		original := b.r.widget.AsWidget().Scene
		b.frame.Store(&placementFrame{})
		body := prepareDialogBody(r)
		body.NewDialog(r.widget)
		r.control.dialog.active = true
		if b.r.widget.AsWidget().Scene == original || b.frame.Load() != nil {
			t.Fatal("dialog kept old scene completion")
		}
		b.frame.Store(&placementFrame{})
		body.Send(events.Close)
		if b.r.widget.AsWidget().Scene != original || b.frame.Load() != nil {
			t.Fatal("dialog return kept old completion")
		}
	})
	a.Drain()
}

func TestCanvasNamespacedObjectPressFocusesHost(t *testing.T) {
	f := newCanvasAX(t, 7, 160)
	prefix := accessibilityWorkbenchNamespace("workbench", "panel")
	bridge := &accessibilityBridge{owner: f.owner}
	f.owner.Do(func() {
		nodes, h := accessibilitySnapshot(f.root(), bridge)
		namespaced, wrapped := namespaceAccessibility(nodes, h, prefix, bridge)
		id := prefix + accessibilityManagedID(CanvasObjectID("drawing", "rear:/β"))
		if wrapped[id] == nil || !wrapped[id](accessibilityAction{Kind: "press"}) {
			t.Fatal("namespaced press refused")
		}
		focus := bridge.pendingFocus.Load()
		if focus == nil || *focus != prefix+accessibilityManagedID("drawing") {
			t.Fatal("unqualified host focus", focus)
		}
		for _, n := range namespaced {
			if n.ID == id && (n.GridScope != prefix+accessibilityManagedID("drawing") || n.Parent != n.GridScope) {
				t.Fatal("object scope not namespaced", n)
			}
		}
	})
}

func TestCanvasWorkbenchParkingWithdrawsBeforePanelView(t *testing.T) {
	a, host := gridContractOwner(t, 16)
	scene, _ := canvas.New([]canvas.Rect{{ID: "object", Width: 40, Height: 30}})
	layout, _ := workbench.New(workbench.Layout{Version: 1, Root: &workbench.Group{ID: "group", Tabs: []string{"drawing", "peer"}, Active: "drawing"}}, []string{"drawing", "peer"})
	var binding *placementBinding
	check, observed := false, false
	dock := &Workbench{Owner: a, Window: "one", Workspace: layout, Panels: []Panel{
		{ID: "drawing", Title: "Drawing", View: func() ui.Node {
			if check {
				observed = true
				if binding.frame.Load() != nil {
					t.Error("parked completion visible to panel callback")
				}
			}
			return ui.Node{ID: "drawing", Kind: ui.Canvas2D, Canvas: scene, CanvasObjects: []ui.CanvasObject{{ID: "object", Name: "Object"}}}
		}},
		{ID: "peer", Title: "Peer", View: func() ui.Node { return ui.Node{ID: "peer", Kind: ui.Label} }},
	}}
	widget := dock.Mount(host)
	t.Cleanup(dock.disposeControlRecords)
	a.Do(func() {
		if err := dock.Sync(widget); err != nil {
			t.Fatal(err)
		}
		binding = dock.records["drawing"].control.placement
		binding.frame.Store(&placementFrame{})
		if err := layout.Move("drawing", "group", 1); err != nil {
			t.Fatal(err)
		}
		check = true
		if err := dock.Sync(widget); err != nil {
			t.Fatal(err)
		}
		if !observed {
			t.Fatal("parking callback not exercised")
		}
	})
}

func TestCanvasObjectGeometryUsesActualAncestorScrollClip(t *testing.T) {
	f := newPlacementFixture(t, 7, 240)
	f.app.Do(func() {
		f.scroll = true
		f.metadata = []ui.CanvasObject{{ID: "existing", Name: "Scrolled object"}}
		f.model.Objects[0].Y = 200
		f.model.Objects[0].Height = 300
		f.app.Invalidate("main")
	})
	manualDockSettle(t, f.manual, 7)
	scroller := findRecord(f.manual.implementation.find("main").state.root, "scroller").widget.(*core.Frame)
	if !scroller.ScrollDimToContentEnd(math32.Y) {
		t.Fatal("actual scrollbar did not move")
	}
	manualDockSettle(t, f.manual, 7)
	g := f.frame(t).geometry
	if g.Origin.Y >= g.Clip.Min.Y {
		t.Fatal("fixture has no scrolled origin", g)
	}
	snapshot, err := f.manual.Snapshot(testContext)
	if err != nil || len(snapshot.Windows[0].Canvases) != 1 || len(snapshot.Windows[0].Canvases[0].Objects) != 1 {
		t.Fatal("missing scrolled semantics", err)
	}
	got := snapshot.Windows[0].Canvases[0].Objects[0].Bounds
	object := f.model.Objects[0]
	x := g.Origin.X + (g.View.Width/2+(object.X-g.View.Center.X)*g.View.Zoom)*1.5
	y := g.Origin.Y + (g.View.Height/2+(object.Y-g.View.Center.Y)*g.View.Zoom)*1.5
	right := math.Min(x+object.Width*g.View.Zoom*1.5, g.Clip.Max.X)
	bottom := math.Min(y+object.Height*g.View.Zoom*1.5, g.Clip.Max.Y)
	x = math.Max(x, g.Clip.Min.X)
	y = math.Max(y, g.Clip.Min.Y)
	want := [4]float64{x, y, right - x, bottom - y}
	for i := range got {
		if math.Abs(float64(got[i])-want[i]) > .01 {
			t.Fatal("ancestor scroll transform", got, want)
		}
	}
}
