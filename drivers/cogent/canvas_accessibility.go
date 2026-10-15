package cogent

import (
	"math"
	"slices"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/placement"
	"github.com/dshills/vellum/ui"
)

// CanvasObjectGeometry describes a labeled object in one completed, clipped
// window-pixel frame. ID is a transport target; ObjectID is the scene identity.
type CanvasObjectGeometry struct {
	ID, ObjectID, Name, Help string
	Selected                 bool
	Bounds                   [4]float32
}

// CanvasInspection is a detached publication for an annotated Canvas2D. Bounds
// and object rectangles use clipped window pixels; View uses logical units.
type CanvasInspection struct {
	Window, ID              string
	Incarnation, Generation uint64
	View                    canvas.Viewport
	Bounds                  [4]float32
	TotalObjects            int
	Objects                 []CanvasObjectGeometry
}

func inspectCanvasFrame(b *placementBinding, f *placementFrame) CanvasInspection {
	g := f.geometry
	return CanvasInspection{Window: b.window.key.window, ID: b.r.node.ID, Incarnation: b.incarnation, Generation: f.picture.generation, View: g.View, Bounds: [4]float32{float32(g.Clip.Min.X), float32(g.Clip.Min.Y), float32(g.Clip.Max.X - g.Clip.Min.X), float32(g.Clip.Max.Y - g.Clip.Min.Y)}, TotalObjects: len(f.picture.semantics), Objects: slices.Clone(f.objects)}
}

// CanvasObjectID separates generated object identities from managed node IDs.
func CanvasObjectID(canvasID, objectID string) string {
	return "canvas-object:" + accessibilityIDPart(canvasID) + ":" + accessibilityIDPart(objectID)
}

type canvasActionEpoch struct{ marker byte }

func sameCanvasPicture(a, b *canvasPicture) bool {
	return a != nil && b != nil && a.view == b.view && a.selected == b.selected && a.grid == b.grid && a.preview == b.preview && slices.Equal(a.objects, b.objects) && slices.Equal(a.semantics, b.semantics)
}
func measureCanvasObjects(p *canvasPicture, g placement.Geometry, hostID string) []CanvasObjectGeometry {
	if len(p.semantics) == 0 {
		return nil
	}
	labels := make(map[string]ui.CanvasObject, len(p.semantics))
	for _, label := range p.semantics {
		labels[label.ID] = label
	}
	out := make([]CanvasObjectGeometry, 0, len(p.semantics))
	// Preserve scene paint order for keyboard traversal and native hit testing.
	for _, o := range p.objects {
		label, ok := labels[o.ID]
		if !ok {
			continue
		}
		point := g.View.Screen(canvas.Point{X: o.X, Y: o.Y})
		x, y := g.Origin.X+point.X*g.Scale, g.Origin.Y+point.Y*g.Scale
		right, bottom := x+o.Width*g.View.Zoom*g.Scale, y+o.Height*g.View.Zoom*g.Scale
		x, y = math.Max(x, g.Clip.Min.X), math.Max(y, g.Clip.Min.Y)
		right, bottom = math.Min(right, g.Clip.Max.X), math.Min(bottom, g.Clip.Max.Y)
		if right <= x || bottom <= y {
			continue
		}
		out = append(out, CanvasObjectGeometry{ID: CanvasObjectID(hostID, o.ID), ObjectID: o.ID, Name: label.Name, Help: label.Help, Selected: o.ID == p.selected, Bounds: [4]float32{float32(x), float32(y), float32(right - x), float32(bottom - y)}})
	}
	return out
}

// All mutable model comparisons occur on the application owner. Drawing sees
// only detached pictures. A completed picture cannot qualify changed objects,
// an unrendered viewport, a retired scope or a live palette capture.
func canvasCompletedFrame(b *placementBinding, f *placementFrame) bool {
	if b == nil || f == nil || b.retired || b.domain.Err() != nil || b.model == nil || f.picture == nil || f.scene == nil || f.scene.GlobalLayoutPending() || b.frame.Load() != f || b.picture.Load() != f.picture {
		return false
	}
	view := b.model.View
	view.Width, view.Height = f.geometry.View.Width, f.geometry.View.Height
	return view == f.geometry.View && b.model.Selected == f.picture.selected && slices.Equal(b.model.Objects, f.picture.objects) && slices.Equal(b.r.node.CanvasObjects, f.picture.semantics)
}

// A pending host without a matching completed cache must not announce an
// unpainted model selection. Existing-cache pending values are restored by
// prepareGridPublication, independently of this candidate host's empty value.
func canvasCompletedSelection(r *record) string {
	if r.control == nil || r.control.placement == nil {
		return ""
	}
	b := r.control.placement
	f := b.frame.Load()
	if !canvasCompletedFrame(b, f) {
		return ""
	}
	for _, label := range f.picture.semantics {
		if label.ID == f.picture.selected {
			return label.Name
		}
	}
	return ""
}

func canvasFrameCurrent(b *placementBinding, f *placementFrame) bool {
	return b != nil && b.live() && b.window.active == nil && canvasCompletedFrame(b, f)
}

func canvasActionGuard(b *placementBinding) func() bool {
	f := b.frame.Load()
	scope := b.scope.Load()
	epoch := b.window.canvasEpoch.Load()
	return func() bool {
		return scope.allowed() && b.window.canvasEpoch.Load() == epoch && canvasFrameCurrent(b, f)
	}
}
func selectCanvasObject(b *placementBinding, id string) bool {
	if b.model.Select(id) != nil {
		return false
	}
	canvasChanged(b.r)
	return true
}
func canvasObjectKey(r *record, e events.Event) bool {
	if len(r.node.CanvasObjects) == 0 || !key.HasAnyModifier(e.Modifiers(), key.Alt) || key.HasAnyModifier(e.Modifiers(), key.Control, key.Meta, key.Shift) {
		return false
	}
	switch e.KeyCode() {
	case key.CodeLeftArrow, key.CodeRightArrow, key.CodeHome, key.CodeEnd:
	default:
		return false
	}
	canvasOwnerEvent(r, e, func(b *placementBinding, f *placementFrame) {
		if len(f.objects) == 0 {
			return
		}
		index := -1
		for i, o := range f.objects {
			if o.ObjectID == b.model.Selected {
				index = i
				break
			}
		}
		switch e.KeyCode() {
		case key.CodeHome:
			index = 0
		case key.CodeEnd:
			index = len(f.objects) - 1
		case key.CodeRightArrow:
			index = (index + 1) % len(f.objects)
		case key.CodeLeftArrow:
			if index <= 0 {
				index = len(f.objects)
			}
			index--
		}
		selectCanvasObject(b, f.objects[index].ObjectID)
	})
	e.SetHandled()
	return true
}
func withdrawCanvasGeometry(b *placementBinding) {
	if b != nil && b.model != nil && b.frame.Swap(nil) != nil {
		_ = b.r.control.a.Invalidate(b.r.control.window)
	}
}

func canvasAccessibility(r *record, parent string, bridge *accessibilityBridge, disabled bool) ([]accessibilityNode, map[string]accessibilityHandler) {
	b := r.control.placement
	f := b.frame.Load()
	nodes := []accessibilityNode{}
	handlers := map[string]accessibilityHandler{}
	if !canvasCompletedFrame(b, f) {
		return nodes, handlers
	}
	for _, o := range f.objects {
		id := accessibilityManagedID(CanvasObjectID(r.node.ID, o.ObjectID))
		n := accessibilityNode{ID: id, DomainID: o.ObjectID, Parent: parent, Role: "button", Name: o.Name, Help: o.Help, Selected: o.Selected, Disabled: disabled, pressFocusID: parent, Actions: []string{}}
		if !disabled {
			n.Actions = []string{"press"}
			objectID := o.ObjectID
			handlers[id] = func(a accessibilityAction) bool {
				if a.Kind != "press" || !selectCanvasObject(b, objectID) {
					return false
				}
				if bridge != nil {
					accessibilityFocus(bridge, parent)
				}
				return true
			}
		}
		nodes = append(nodes, n)
	}
	return nodes, handlers
}
func stampCanvasAccessibility(r *record, nodes []accessibilityNode, handlers map[string]accessibilityHandler) {
	if len(nodes) == 0 {
		return
	}
	b := r.control.placement
	f := b.frame.Load()
	p := b.picture.Load()
	if p == nil {
		return
	}
	pending := !canvasCompletedFrame(b, f)
	frames := map[string][4]float32{}
	if !pending {
		g := f.geometry
		frames[nodes[0].ID] = [4]float32{float32(g.Clip.Min.X), float32(g.Clip.Min.Y), float32(g.Clip.Max.X - g.Clip.Min.X), float32(g.Clip.Max.Y - g.Clip.Min.Y)}
		for _, o := range f.objects {
			frames[accessibilityManagedID(CanvasObjectID(r.node.ID, o.ObjectID))] = o.Bounds
		}
	}
	guard := canvasActionGuard(b)
	for i := range nodes {
		n := &nodes[i]
		n.GridScope = nodes[0].ID
		n.GridIncarnation = b.incarnation
		n.GridTargetGeneration = p.generation
		n.GridPending = pending
		if b.window.active != nil {
			n.Disabled = true
			n.Focused = false
			n.Actions = []string{}
			delete(handlers, n.ID)
		}
		if pending {
			n.Disabled = true
			n.Focused = false
			n.Actions = []string{}
			delete(handlers, n.ID)
		} else {
			n.GridGeneration = p.generation
			frame := frames[n.ID]
			n.GridFrame = &frame
			if handler := handlers[n.ID]; handler != nil {
				n.activationGuard = guard
				handlers[n.ID] = func(a accessibilityAction) bool {
					if !guard() {
						return false
					}
					return handler(a)
				}
			}
		}
	}
}
func canvasAccessibilityBounds(r *record, out map[string][4]float32) {
	if r.control == nil || r.control.placement == nil || len(r.node.CanvasObjects) == 0 {
		return
	}
	b := r.control.placement
	f := b.frame.Load()
	if f == nil || f.picture == nil || b.picture.Load() != f.picture || f.scene == nil || f.scene.GlobalLayoutPending() {
		return
	}
	g := f.geometry
	out[accessibilityManagedID(r.node.ID)] = [4]float32{float32(g.Clip.Min.X), float32(g.Clip.Min.Y), float32(g.Clip.Max.X - g.Clip.Min.X), float32(g.Clip.Max.Y - g.Clip.Min.Y)}
	for _, o := range f.objects {
		out[accessibilityManagedID(CanvasObjectID(r.node.ID, o.ObjectID))] = o.Bounds
	}
}
