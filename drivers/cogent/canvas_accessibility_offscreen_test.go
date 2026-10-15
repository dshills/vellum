//go:build offscreen

package cogent

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"math"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
)

type canvasAXFixture struct {
	manual                  *Manual
	owner                   *app.Application
	model                   *canvas.Scene
	metadata                []ui.CanvasObject
	disabled, hidden, modal bool
	calls                   int
	last                    ui.ChangeEvent
}

func newCanvasAX(t *testing.T, batch int, dpi float32) *canvasAXFixture {
	t.Helper()
	f := &canvasAXFixture{}
	var err error
	f.model, err = canvas.New([]canvas.Rect{{ID: "rear:/β", X: -30, Y: -20, Width: 100, Height: 60, Color: color.RGBA{R: 30, A: 255}}, {ID: "front", X: -30, Y: -20, Width: 100, Height: 60, Color: color.RGBA{B: 120, A: 255}}, {ID: "edge", X: -1e6, Y: -20, Width: 1e6, Height: 60, Color: color.RGBA{G: 90, A: 255}}, {ID: "offscreen", X: 10000, Y: 10000, Width: 10, Height: 10}})
	if err != nil {
		t.Fatal(err)
	}
	f.metadata = []ui.CanvasObject{{ID: "rear:/β", Name: "Rear block", Help: "Select rear"}, {ID: "front", Name: "Front block"}, {ID: "edge", Name: "Clipped edge"}, {ID: "offscreen", Name: "Outside"}}
	f.model.View.Zoom = 1.7
	f.model.View.Center = canvas.Point{X: 40, Y: 70}
	config := manualConfig(t, batch)
	config.LogicalDPI = dpi
	f.manual, f.owner = gridManual(t, config, func() ui.Node {
		drawing := ui.Node{ID: "drawing", Kind: ui.Canvas2D, Canvas: f.model, CanvasObjects: f.metadata, Disabled: f.disabled, Grow: true, Accessibility: ui.Accessibility{Name: "Components"}, Change: func(e ui.ChangeEvent) { f.calls++; f.last = e }}
		if f.hidden {
			drawing = ui.Node{ID: "placeholder", Kind: ui.Label, Text: "No canvas"}
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{drawing, {ID: "modal", Kind: ui.Dialog, Text: "Modal", Open: f.modal, Children: []ui.Node{{ID: "message", Kind: ui.Label, Text: "Blocked"}}}}}
	})
	gridFrame(t, f.manual)
	return f
}
func (f *canvasAXFixture) root() *record { return f.manual.implementation.find("one").state.root }
func (f *canvasAXFixture) binding(t *testing.T) *placementBinding {
	t.Helper()
	var out *placementBinding
	var walk func(*record)
	walk = func(r *record) {
		if r.node.ID == "drawing" && r.control != nil {
			out = r.control.placement
		}
		for _, c := range r.children {
			walk(c)
		}
	}
	walk(f.root())
	if out == nil {
		t.Fatal("missing canvas binding")
	}
	return out
}
func (f *canvasAXFixture) snapshot(t *testing.T) ([]accessibilityNode, map[string]accessibilityHandler) {
	t.Helper()
	var nodes []accessibilityNode
	var handlers map[string]accessibilityHandler
	if err := f.owner.Do(func() { nodes, handlers = accessibilitySnapshot(f.root(), nil) }); err != nil {
		t.Fatal(err)
	}
	return nodes, handlers
}
func (f *canvasAXFixture) refresh(t *testing.T) {
	t.Helper()
	if err := f.owner.Invalidate("one"); err != nil {
		t.Fatal(err)
	}
	gridFrame(t, f.manual)
}
func (f *canvasAXFixture) input(t *testing.T, in ManualInput) {
	t.Helper()
	s, err := f.manual.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	in.Window = "one"
	in.WindowIncarnation = s.Windows[0].Incarnation
	if err := f.manual.Input(testContext, in); err != nil {
		t.Fatal(err)
	}
	gridFrame(t, f.manual)
}
func invokeCanvas(t *testing.T, f *canvasAXFixture, handler accessibilityHandler) bool {
	t.Helper()
	accepted := false
	if err := f.owner.Do(func() { accepted = handler(accessibilityAction{Kind: "press"}) }); err != nil {
		t.Fatal(err)
	}
	return accepted
}
func TestCanvasCompletedSemanticsAndSelection(t *testing.T) {
	for _, batch := range []int{1, 7, 64} {
		for _, dpi := range []float32{96, 120, 160, 240} {
			t.Run(fmt.Sprintf("%d/%g", batch, dpi), func(t *testing.T) {
				f := newCanvasAX(t, batch, dpi)
				nodes, handlers := f.snapshot(t)
				byID := gridAXByID(t, nodes)
				b := f.binding(t)
				frame := b.frame.Load()
				if frame == nil {
					t.Fatal("no completed frame")
				}
				host := byID[accessibilityManagedID("drawing")]
				if host.Role != "canvas" || host.GridPending || host.GridFrame == nil || host.GridGeneration == 0 || host.GridIncarnation == 0 {
					t.Fatal("canvas host lacks completed transaction", host)
				}
				objectID := accessibilityManagedID(CanvasObjectID("drawing", "rear:/β"))
				rear := byID[objectID]
				if rear.Name != "Rear block" || rear.Help != "Select rear" || rear.DomainID != "rear:/β" || rear.Role != "button" || rear.GridFrame == nil || rear.GridScope != host.ID || rear.GridGeneration != host.GridGeneration {
					t.Fatal("named object not published", rear)
				}
				if _, present := byID[accessibilityManagedID(CanvasObjectID("drawing", "offscreen"))]; present {
					t.Fatal("offscreen object leaked into visible tree")
				}
				edge := byID[accessibilityManagedID(CanvasObjectID("drawing", "edge"))]
				if edge.GridFrame == nil || (*edge.GridFrame)[0] != float32(frame.geometry.Clip.Min.X) {
					t.Fatal("partially clipped object bounds", edge)
				}
				// Independent arithmetic: logical world -> window pixels, then visible clip.
				object := f.model.Objects[0]
				scale := float64(dpi) / 160
				v := frame.geometry.View
				x := (object.X-v.Center.X)*v.Zoom + v.Width/2
				y := (object.Y-v.Center.Y)*v.Zoom + v.Height/2
				left, top := math.Max(frame.geometry.Origin.X+x*scale, frame.geometry.Clip.Min.X), math.Max(frame.geometry.Origin.Y+y*scale, frame.geometry.Clip.Min.Y)
				right := math.Min(frame.geometry.Origin.X+(x+object.Width*v.Zoom)*scale, frame.geometry.Clip.Max.X)
				bottom := math.Min(frame.geometry.Origin.Y+(y+object.Height*v.Zoom)*scale, frame.geometry.Clip.Max.Y)
				expected := [4]float32{float32(left), float32(top), float32(right - left), float32(bottom - top)}
				for i, value := range *rear.GridFrame {
					if math.Abs(float64(value-expected[i])) > .01 {
						t.Fatal("transformed/clipped object bounds", *rear.GridFrame, expected)
					}
				}
				before := f.calls
				if !invokeCanvas(t, f, handlers[objectID]) || f.model.Selected != "rear:/β" || f.calls != before+1 || f.last.SelectedID != "rear:/β" {
					t.Fatal("AX press hit topmost object instead of exact ID", f.model.Selected, f.calls)
				}
				f.refresh(t)
				s, err := f.manual.Snapshot(testContext)
				if err != nil || len(s.Windows[0].Canvases) != 1 || len(s.Windows[0].Canvases[0].Objects) != 3 {
					t.Fatal("public completed canvas inspection", err)
				}
				// Physical pointer chooses the front object, using the same owner notification.
				r := expected
				point := image.Pt(int(r[0]+r[2]/2), int(r[1]+r[3]/2))
				for _, kind := range []events.Types{events.MouseDown, events.MouseUp} {
					f.input(t, ManualInput{Type: kind, Position: point, Button: events.Left})
				}
				if f.model.Selected != "front" || f.last.SelectedID != "front" {
					t.Fatal("pointer paint-order selection", f.model.Selected)
				}
				var alt key.Modifiers
				alt.SetFlag(true, key.Alt)
				f.input(t, ManualInput{Type: events.KeyChord, Code: key.CodeHome, Modifiers: alt})
				if f.model.Selected != "rear:/β" {
					t.Fatal("Alt+Home did not use first semantic object")
				}
				f.input(t, ManualInput{Type: events.KeyChord, Code: key.CodeRightArrow, Modifiers: alt})
				if f.model.Selected != "front" {
					t.Fatal("Alt+Right did not select next object")
				}
				center := f.model.View.Center
				f.input(t, ManualInput{Type: events.KeyChord, Code: key.CodeRightArrow})
				if f.model.View.Center == center || f.model.Selected != "front" {
					t.Fatal("ordinary arrow panning changed")
				}
				native := nativeAccessibilityBounds(f.root())
				if _, exists := native[host.ID]; exists {
					t.Fatal("scoped canvas duplicated in ordinary geometry")
				}
			})
		}
	}
}
func TestCanvasPendingKeepsCompletedTreeAndRejectsOldActions(t *testing.T) {
	f := newCanvasAX(t, 7, 160)
	nodes, handlers := f.snapshot(t)
	bridge := newAccessibilityBridge(f.owner)
	defer bridge.close()
	bridge.bind("one")
	bridge.update(nodes, handlers)
	oldNodes := cloneAccessibilityNodes(bridge.gridPublications[accessibilityManagedID("drawing")])
	if len(oldNodes) != 4 {
		t.Fatal("completed scope missing", len(oldNodes))
	}
	for _, n := range oldNodes {
		if n.activationGuard != nil {
			t.Fatal("completed cache retained a binding closure")
		}
	}
	id := accessibilityManagedID(CanvasObjectID("drawing", "rear:/β"))
	oldHandler := handlers[id]
	if err := f.owner.Do(func() { f.model.Objects[0].X += 100 }); err != nil {
		t.Fatal(err)
	}
	if invokeCanvas(t, f, oldHandler) {
		t.Fatal("changed object accepted old semantic action")
	}
	pending, pendingHandlers := f.snapshot(t)
	bridge.update(pending, pendingHandlers)
	var packet struct {
		Generation uint64
		Nodes      []accessibilityNode
	}
	if err := json.Unmarshal([]byte(bridge.lastSnapshot), &packet.Nodes); err != nil {
		t.Fatal(err)
	}
	retained := gridAXByID(t, packet.Nodes)
	old := retained[id]
	if !old.GridPending || !old.Disabled || len(old.Actions) != 0 || old.GridFrame == nil || *old.GridFrame != *oldNodes[1].GridFrame {
		t.Fatal("pending canvas lost read-only completed data", old)
	}
	f.refresh(t)
	fresh, actions := f.snapshot(t)
	for _, n := range fresh {
		if n.ID == id && n.GridPending {
			t.Fatal("fresh object remained pending")
		}
	}
	if !invokeCanvas(t, f, actions[id]) {
		t.Fatal("new completed action refused")
	}
}

func TestCanvasPendingWithoutMatchingCacheOmitsUnpaintedSelection(t *testing.T) {
	a, host := gridContractOwner(t, 16)
	scene, _ := canvas.New([]canvas.Rect{{ID: "object", Width: 40, Height: 30}})
	scene.Selected = "object"
	r := mountGridContract(t, a, host, ui.Node{ID: "drawing", Kind: ui.Canvas2D, Canvas: scene, CanvasObjects: []ui.CanvasObject{{ID: "object", Name: "Unpainted object"}}})
	t.Cleanup(func() { disposeControlTree(r) })
	a.Do(func() {
		if r.control.placement.frame.Load() != nil {
			t.Fatal("startup fixture already painted")
		}
		nodes, h := accessibilitySnapshot(r, nil)
		bridge := &accessibilityBridge{}
		pending, _ := bridge.prepareGridPublication(nodes, h)
		if len(pending) != 1 || !pending[0].GridPending || pending[0].Value != "" || !pending[0].Disabled {
			t.Fatal("startup announced unpainted selection", pending)
		}
	})
}

func TestCanvasChangedParentWithoutCompletionOmitsUnpaintedSelection(t *testing.T) {
	f := newCanvasAX(t, 7, 160)
	f.owner.Do(func() { f.model.Selected = "rear:/β" })
	f.refresh(t)
	f.owner.Do(func() {
		nodes, h := accessibilitySnapshot(f.root(), nil)
		bridge := &accessibilityBridge{}
		bridge.prepareGridPublication(nodes, h)
		f.model.Selected = "front"
		withdrawCanvasGeometry(f.binding(t))
		pending, actions := accessibilitySnapshot(f.root(), nil)
		host := accessibilityManagedID("drawing")
		// This is the changed-parent scope-cache boundary. Actual transfer hooks
		// are exercised independently by TestCanvasRetainedAncestorMove... .
		for i := range pending {
			if pending[i].ID == host {
				pending[i].Parent = "new-parent"
			}
		}
		out, _ := bridge.prepareGridPublication(pending, actions)
		found := false
		for _, n := range out {
			if n.ID == host {
				found = true
				if !n.GridPending || n.Value != "" || !n.Disabled {
					t.Fatal("new parent reused unpainted selection", n)
				}
			}
		}
		if !found {
			t.Fatal("pending host missing")
		}
	})
}
