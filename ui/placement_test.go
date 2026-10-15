package ui

import (
	"context"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/placement"
	"testing"
)

func TestPlacementDescriptorContracts(t *testing.T) {
	scene, _ := canvas.New(nil)
	controller, err := placement.New(placement.Config[string]{Items: []placement.Item[string]{{ID: "box", Label: "Box", Preview: canvas.Rect{Width: 10, Height: 10}}}, Command: func(context.Context, placement.Request[string]) error { return nil }})
	if err != nil {
		t.Fatal(err)
	}
	defer controller.Close()
	target, err := controller.NewTarget(placement.TargetConfig[string]{ID: "canvas"})
	if err != nil {
		t.Fatal(err)
	}
	source := Node{ID: "source", Kind: Button, PlacementSource: controller.Source("box"), Activate: func() {}}
	drawing := Node{ID: "canvas", Kind: Canvas2D, Canvas: scene, PlacementTarget: target, CanvasGrid: 20}
	root := Node{ID: "root", Kind: Row, Children: []Node{source, drawing}}
	if err := Validate(root); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name string
		node Node
	}{
		{"source-no-keyboard", Node{ID: "bad", Kind: Button, PlacementSource: controller.Source("box")}},
		{"source-wrong-kind", Node{ID: "bad", Kind: Label, PlacementSource: controller.Source("box"), Activate: func() {}}},
		{"canvas-missing", Node{ID: "bad", Kind: Canvas2D}},
		{"canvas-grid", Node{ID: "bad", Kind: Canvas2D, Canvas: scene, CanvasGrid: -1}},
		{"target-wrong-kind", Node{ID: "bad", Kind: Button, PlacementTarget: target}},
		{"canvas-child", Node{ID: "bad", Kind: Canvas2D, Canvas: scene, Children: []Node{{ID: "child", Kind: Label}}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if Validate(test.node) == nil {
				t.Fatal("invalid placement descriptor admitted")
			}
		})
	}
	for _, node := range []Node{source, drawing} {
		copy := node
		copy.ID = "duplicate"
		candidate := root
		candidate.Children = append(append([]Node{}, root.Children...), copy)
		if Validate(candidate) == nil {
			t.Fatal("model or endpoint multiply bound")
		}
	}
}
