package cogent

import (
	"fmt"
	"github.com/dshills/vellum/viewport"
	"image"
	"image/color"
	"math"
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

func TestCubeFaceStrokeLabelsStayCenteredAndOnFace(t *testing.T) {
	for _, face := range cubeFaces {
		for _, ch := range face.label {
			lines := cubeStrokes[ch]
			if ch > 127 || len(lines) == 0 {
				t.Fatal("missing fixed ASCII glyph", face.label, ch)
			}
			for _, line := range lines {
				if line[0] == line[2] && line[1] == line[3] {
					t.Fatal("degenerate fixed stroke", ch)
				}
			}
		}
		m := &sceneMesh{}
		m.cubeLabel(face.label, face.normal, face.u, face.v)
		if len(m.indices) == 0 {
			t.Fatal("missing face name", face.label)
		}
		min, max := float64(2), float64(-2)
		for i := 0; i < len(m.vertices); i += 7 {
			p := viewport.Vec3{X: float64(m.vertices[i]), Y: float64(m.vertices[i+1]), Z: float64(m.vertices[i+2])}
			u, v := dotScene(p, face.u), dotScene(p, face.v)
			if math.Abs(dotScene(p, face.normal)-1.012) > 1e-5 || math.Abs(u) > .85 || math.Abs(v) > .2 {
				t.Fatal("label clipped or off face", face.label, p)
			}
			min = math.Min(min, u)
			max = math.Max(max, u)
		}
		if math.Abs(min+max) > .04 {
			t.Fatal("label not centered", face.label, min, max)
		}
	}
}

func dotScene(a, b viewport.Vec3) float64 { return a.X*b.X + a.Y*b.Y + a.Z*b.Z }

// Detached routing at matching physical size/scale, separate from native mouse
// evidence. Cube clicks must not invoke body selection; this view has no command journal.
func TestScene3DCubeAndContextKeepObjectSelection(t *testing.T) {
	for _, scale := range []float64{1, 1.5, 2} {
		t.Run(fmt.Sprint(scale), func(t *testing.T) {
			s := sceneFixture()
			s.Camera = viewport.DefaultCamera()
			before := s.Camera
			selections, contexts := 0, 0
			v := &SceneView3D{Model: s, Select: func(string) { selections++ }, Context: func(GPUEvent) { contexts++ }}
			size := image.Pt(int(800*scale), int(600*scale))
			rect := viewport.CubeRect(size.X, size.Y, scale)
			var target image.Point
			found := false
			for y := rect.Min.Y; y < rect.Max.Y && !found; y++ {
				for x := rect.Min.X; x < rect.Max.X; x++ {
					if face, ok := viewport.CubeViewAt(s.Camera, x, y, size.X, size.Y, scale); ok && face == viewport.Front {
						target = image.Pt(x, y)
						found = true
						break
					}
				}
			}
			if !found {
				t.Fatal("no front face")
			}
			e := GPUEvent{Kind: GPUPointerDown, Sequence: 1, Button: GPUButtonPrimary, Position: target, Size: size, Scale: scale, Inside: true}
			if !v.HandleInput(e) {
				t.Fatal("cube down")
			}
			e.Kind = GPUPointerUp
			if !v.HandleInput(e) || selections != 0 || s.Camera == before {
				t.Fatal("cube click selected object or did not orient")
			}
			front := s.Camera
			e.Button = GPUButtonSecondary
			e.Position = image.Pt(size.X/2, size.Y/2)
			for _, kind := range []GPUEventKind{GPUPointerDown, GPUPointerUp, GPUContextMenu} {
				e.Kind = kind
				v.HandleInput(e)
			}
			if contexts != 1 {
				t.Fatal("context callback count", contexts)
			}
			if selections != 0 {
				t.Fatal("secondary click selected an object", selections)
			}
			if s.Camera != front {
				t.Fatal("context changed camera")
			}
			if v.gesture.Active() {
				t.Fatal("secondary context retained capture")
			}
		})
	}
}

func TestRoundShaftDisplayGeometryStaysInsideProxy(t *testing.T) {
	b := viewport.Box{Min: viewport.Vec3{X: -2, Y: -.1, Z: -.1}, Max: viewport.Vec3{X: 2, Y: .1, Z: .1}}
	m := &sceneMesh{}
	m.cylinderX(b, color.NRGBA{R: 190, G: 200, B: 210, A: 255})
	if len(m.indices) != 48*12 || len(m.vertices)/7 != 48*12 {
		t.Fatal("unbounded cylinder tessellation")
	}
	for j := 0; j < len(m.vertices); j += 7 {
		x, y, z := float64(m.vertices[j]), float64(m.vertices[j+1]), float64(m.vertices[j+2])
		if math.Abs(x) > 2.00001 || y*y+z*z > .010001 {
			t.Fatal("vertex outside round shaft", x, y, z)
		}
	}
	for _, i := range m.indices {
		if int(i) >= len(m.vertices)/7 {
			t.Fatal("invalid cylinder index")
		}
	}
}

func TestCylinderFacesPointOutwardForBackfaceCulling(t *testing.T) {
	m := &sceneMesh{}
	m.cylinderX(viewport.Box{Min: viewport.Vec3{X: -2, Y: -.1, Z: -.1}, Max: viewport.Vec3{X: 2, Y: .1, Z: .1}}, color.NRGBA{A: 255})
	at := func(index uint32) viewport.Vec3 {
		j := int(index) * 7
		return viewport.Vec3{X: float64(m.vertices[j]), Y: float64(m.vertices[j+1]), Z: float64(m.vertices[j+2])}
	}
	for j := 0; j < len(m.indices); j += 3 {
		a, b, c := at(m.indices[j]), at(m.indices[j+1]), at(m.indices[j+2])
		center := a.Add(b).Add(c).Mul(1.0 / 3)
		outward := viewport.Vec3{Y: center.Y, Z: center.Z}
		switch (j / 3) % 4 {
		case 2:
			outward = viewport.Vec3{X: -1}
		case 3:
			outward = viewport.Vec3{X: 1}
		}
		if b.Sub(a).Cross(c.Sub(a)).Dot(outward) <= 0 {
			t.Fatal("inward or degenerate cylinder triangle", j/3)
		}
	}
}
