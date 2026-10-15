package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dshills/vellum/canvas"
)

func TestCanvasSemanticDescriptors(t *testing.T) {
	scene, err := canvas.New([]canvas.Rect{{ID: "a:/β", Width: 20, Height: 20}, {ID: "b", Width: 20, Height: 20}})
	if err != nil {
		t.Fatal(err)
	}
	node := Node{ID: "drawing", Kind: Canvas2D, Canvas: scene, CanvasObjects: []CanvasObject{{ID: "a:/β", Name: "Resistor", Help: "Select component"}, {ID: "b", Name: "Resistor"}}}
	if err := Validate(node); err != nil {
		t.Fatal(err)
	} // Duplicate labels are fine; identity is separate.
	for _, mutate := range []func(*Node){
		func(n *Node) { n.Kind = Label }, func(n *Node) { n.CanvasObjects[0].ID = "missing" }, func(n *Node) { n.CanvasObjects[1].ID = n.CanvasObjects[0].ID },
		func(n *Node) { n.CanvasObjects[0].Name = " " }, func(n *Node) { n.CanvasObjects[0].Name = "\xff" }, func(n *Node) { n.CanvasObjects[0].Help = "bad\x00help" },
		func(n *Node) { n.CanvasObjects[0].Name = strings.Repeat("x", 1025) }, func(n *Node) { n.CanvasObjects[0].Help = strings.Repeat("x", 4097) },
		func(n *Node) { n.CanvasObjects = make([]CanvasObject, MaxCanvasObjects+1) }, func(n *Node) { n.ID = "canvas-object:reserved" },
	} {
		copy := node
		copy.CanvasObjects = append([]CanvasObject(nil), node.CanvasObjects...)
		mutate(&copy)
		if Validate(copy) == nil {
			t.Fatal("invalid canvas metadata admitted", copy.ID)
		}
	}
	objects := make([]canvas.Rect, 1025)
	metadata := make([]CanvasObject, len(objects))
	for i := range objects {
		id := fmt.Sprint(i)
		objects[i] = canvas.Rect{ID: id, Width: 1, Height: 1}
		metadata[i] = CanvasObject{ID: id, Name: strings.Repeat("n", 1024)}
	}
	scene, _ = canvas.New(objects)
	node.Canvas = scene
	node.CanvasObjects = metadata
	if Validate(node) == nil {
		t.Fatal("aggregate metadata budget ignored")
	}
}

func TestCanvasOversizedSceneRefusedBeforeMetadataAllocation(t *testing.T) {
	scene := &canvas.Scene{View: canvas.Viewport{Width: 800, Height: 600, Zoom: 1}, Objects: make([]canvas.Rect, 100000)}
	node := Node{Kind: Canvas2D, Canvas: scene, CanvasObjects: []CanvasObject{{ID: "a", Name: "A"}}}
	// Unique IDs would allocate per-entry overflow buckets if metadata lookup
	// accidentally scanned this rejected scene before the constant-time cap.
	for i := range scene.Objects {
		scene.Objects[i].ID = fmt.Sprint(i)
	}
	for name, validate := range map[string]func() error{
		"control":  func() error { return ValidateControl(node) },
		"metadata": func() error { return validateCanvasObjects(node) },
	} {
		if allocations := testing.AllocsPerRun(3, func() {
			if validate() == nil {
				t.Fatal("oversized scene accepted")
			}
		}); allocations > 8 {
			t.Fatalf("%s scanned oversized scene: %g allocations", name, allocations)
		}
	}
}
