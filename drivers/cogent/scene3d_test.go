package cogent

import (
	"github.com/dshills/vellum/viewport"
	"image"
	"image/color"
	"testing"
)

func sceneFixture() *viewport.Scene {
	c := viewport.DefaultCamera()
	c.SetView(viewport.Front)
	return &viewport.Scene{Camera: c, Objects: []viewport.Object{{ID: "box", Name: "Box", Bounds: viewport.Box{Min: viewport.Vec3{X: -1, Y: -1, Z: -1}, Max: viewport.Vec3{X: 1, Y: 1, Z: 1}}, Color: color.NRGBA{230, 100, 40, 255}, Visible: true}}, ShowCube: true}
}
func TestScene3DInputSelectionCancelAndCamera(t *testing.T) {
	s := sceneFixture()
	calls := 0
	v := &SceneView3D{Model: s, Select: func(id string) {
		if id != "box" {
			t.Fatal(id)
		}
		calls++
	}}
	e := GPUEvent{Kind: GPUPointerDown, Sequence: 1, Button: GPUButtonPrimary, Position: image.Pt(400, 300), Size: image.Pt(800, 600), Inside: true, Scale: 1}
	v.HandleInput(e)
	beforeClick := s.Camera
	e.Kind = GPUPointerMove
	e.Position.X++
	v.HandleInput(e)
	if s.Camera != beforeClick {
		t.Fatal("click jitter moved camera")
	}
	e.Kind = GPUPointerUp
	v.HandleInput(e)
	v.HandleInput(e)
	if calls != 1 {
		t.Fatal("duplicate selection")
	}
	e.Kind = GPUPointerDown
	e.Sequence = 2
	v.HandleInput(e)
	v.HandleInput(GPUEvent{Kind: GPUCancel, Sequence: 2})
	e.Kind = GPUPointerUp
	v.HandleInput(e)
	if calls != 1 {
		t.Fatal("canceled selection")
	}
	before := s.Camera
	e.Kind = GPUPointerDown
	e.Sequence = 3
	v.HandleInput(e)
	e.Kind = GPUPointerMove
	e.Position.X += 40
	v.HandleInput(e)
	e.Kind = GPUPointerUp
	v.HandleInput(e)
	if s.Camera == before || calls != 1 {
		t.Fatal("orbit drag became click")
	}
	before = s.Camera
	v.HandleInput(GPUEvent{Kind: GPUScroll, ScrollY: 30, Size: image.Pt(800, 600), Inside: true, Scale: 1})
	if s.Camera.Distance == before.Distance {
		t.Fatal("zoom")
	}
}
func TestScene3DGeometryContainsCubeAndRetainsSource(t *testing.T) {
	s := sceneFixture()
	v, i, main, cube := buildScene3DMesh(s)
	if len(v)%7 != 0 || len(i)%3 != 0 || main != 36 || cube < 36 || int(main+cube) != len(i) {
		t.Fatal("geometry counts", len(v), len(i), main, cube)
	}
	for _, index := range i {
		if int(index) >= len(v)/7 {
			t.Fatal("index")
		}
	}
	before := s.Objects[0]
	s.SelectedID = "box"
	selected, _, _, _ := buildScene3DMesh(s)
	same := true
	for j := range v {
		same = same && v[j] == selected[j]
	}
	if same || s.Objects[0] != before {
		t.Fatal("selection did not change pixels or mutated source")
	}
}

func TestScene3DCancelRestoresCameraAndRejectsStaleDown(t *testing.T) {
	s := sceneFixture()
	v := &SceneView3D{Model: s}
	before := s.Camera
	e := GPUEvent{Kind: GPUPointerDown, Sequence: 1, Button: GPUButtonPrimary, Position: image.Pt(400, 300), Size: image.Pt(800, 600), Inside: true, Scale: 1}
	v.HandleInput(e)
	e.Kind = GPUPointerMove
	e.Position.X += 40
	v.HandleInput(e)
	e.Kind = GPUPointerDown
	e.Sequence = 2
	if v.HandleInput(e) {
		t.Fatal("nested down")
	}
	if !v.HandleInput(GPUEvent{Kind: GPUCancel, Sequence: 1}) || s.Camera != before {
		t.Fatal("cancel camera")
	}
	if v.HandleInput(GPUEvent{Kind: GPUCancel, Sequence: 1}) {
		t.Fatal("duplicate cancel")
	}
	var nilView *SceneView3D
	nilView.view(viewport.Front)
	(&SceneView3D{}).view(viewport.Front)
}
