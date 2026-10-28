package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/tree"
	_ "embed"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/viewport"
	"github.com/oliverbestmann/webgpu/wgpu"
	"image"
	"image/color"
	"math"
	"slices"
	"strings"
)

//go:embed scene3d.wgsl
var scene3DShader string

type SceneTool string

const (
	SceneOrbit  SceneTool = "orbit"
	ScenePan    SceneTool = "pan"
	SceneSelect SceneTool = "select"
)

// SceneView3D is an experimental reusable small-scene GPU extension. It borrows
// GPUView's existing device/queue and capture contracts. Model and all methods
// belong to the application owner; model/owner/window are fixed while mounted.
// Selection is proposed to Select; camera/hover are transient model state.
type SceneView3D struct {
	Owner         *app.Application
	Window        string
	Model         *viewport.Scene
	Tool          SceneTool
	Select        func(string)
	CameraChanged func(viewport.Camera)
	Context       func(GPUEvent)
	OnError       func(error)
	gpu           *GPUView
	renderer      *scene3DRenderer
	gesture       viewport.Gesture
	mode          SceneTool
	cubeGesture   bool
	sequence      uint64
	beforeCamera  viewport.Camera
}

func (v *SceneView3D) Mount(parent tree.Node) core.Widget {
	if v == nil || v.Owner == nil || v.Model.Validate() != nil {
		return nil
	}
	if v.gpu != nil && v.gpu.widget != nil && !v.gpu.widget.retired.Load() {
		return nil
	}
	v.gesture.Reset()
	v.renderer = &scene3DRenderer{scene: v.Model}
	program := &GPUProgram{Prepare: v.renderer.prepare, Draw: v.renderer.draw, Release: v.renderer.release}
	v.gpu = &GPUView{Owner: v.Owner, Window: v.Window, Program: program, Depth: true, Background: color.NRGBA{241, 243, 247, 255}, Revision: v.Model.Revision, OnError: v.OnError, InputEvent: v.HandleInput}
	return v.gpu.Mount(parent)
}
func (v *SceneView3D) Sync(w core.Widget) error {
	if v == nil || v.gpu == nil || v.renderer.scene != v.Model || v.gpu.Owner != v.Owner || v.gpu.Window != v.Window {
		return fmt.Errorf("scene3d: mounted identity changed")
	}
	if e := v.Model.Validate(); e != nil {
		return e
	}
	v.gpu.Revision = v.Model.Revision
	return v.gpu.Sync(w)
}
func (v *SceneView3D) Stats() (GPUViewStats, error) {
	if v == nil || v.gpu == nil {
		return GPUViewStats{}, fmt.Errorf("scene3d: not mounted")
	}
	return v.gpu.Stats()
}
func (v *SceneView3D) Snapshot() (*image.NRGBA, error) {
	if v == nil || v.gpu == nil {
		return nil, fmt.Errorf("scene3d: not mounted")
	}
	return v.gpu.Snapshot()
}
func (v *SceneView3D) changed() {
	if e := v.Model.Touch(); e != nil {
		if v.OnError != nil {
			v.OnError(e)
		}
		return
	}
	if v.gpu != nil {
		v.gpu.Revision = v.Model.Revision
	}
	if v.Owner != nil {
		_ = v.Owner.Invalidate(v.Window)
	}
}
func (v *SceneView3D) view(face viewport.View) {
	if v == nil || v.Model == nil {
		return
	}
	if v.Model.Camera.SetView(face) == nil {
		v.changed()
		if v.CameraChanged != nil {
			v.CameraChanged(v.Model.Camera)
		}
	}
}

// HandleInput consumes typed adapter input and may also be called by an external
// router on this application owner. It does not synthesize physical OS input.
func (v *SceneView3D) HandleInput(e GPUEvent) bool {
	if v == nil || v.Model == nil || v.Model.Validate() != nil {
		return false
	}
	s := v.Model
	scale := e.Scale
	if scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
		scale = 1
	}
	switch e.Kind {
	case GPUPointerDown:
		if e.Button != GPUButtonPrimary {
			return false
		}
		if v.gesture.Active() || !v.gesture.Begin(e.Sequence, float64(e.Position.X), float64(e.Position.Y), scale) {
			return false
		}
		v.sequence = e.Sequence
		v.mode = v.Tool
		if v.mode == "" {
			v.mode = SceneOrbit
		}
		if e.Modifiers&GPUShift != 0 {
			v.mode = ScenePan
		}
		if e.Modifiers&GPUOption != 0 {
			v.mode = SceneOrbit
		}
		v.beforeCamera = s.Camera
		v.cubeGesture = s.ShowCube && image.Pt(e.Position.X, e.Position.Y).In(viewport.CubeRect(e.Size.X, e.Size.Y, scale))
		return true
	case GPUPointerMove:
		if v.gesture.Active() {
			dx, dy, ok := v.gesture.Move(e.Sequence, float64(e.Position.X), float64(e.Position.Y))
			if !ok {
				return false
			}
			if !v.gesture.Dragging() {
				return true
			}
			if v.mode == ScenePan && !v.cubeGesture {
				_ = s.Camera.Pan(dx, dy, float64(e.Size.X), float64(e.Size.Y))
			} else if v.mode != SceneSelect || v.cubeGesture {
				_ = s.Camera.Orbit(-dx/scale*.008, dy/scale*.008)
			}
			v.changed()
			if v.CameraChanged != nil {
				v.CameraChanged(s.Camera)
			}
			return true
		}
		if !e.Inside {
			return false
		}
		id, err := s.Pick(float64(e.Position.X), float64(e.Position.Y), float64(e.Size.X), float64(e.Size.Y))
		if err == nil && id != s.HoveredID {
			s.HoveredID = id
			v.changed()
		}
		return false
	case GPUPointerUp:
		if e.Button != GPUButtonPrimary {
			return false
		}
		if !v.gesture.Active() || e.Sequence != v.sequence {
			return false
		}
		cubeGesture := v.cubeGesture
		v.cubeGesture = false
		click := v.gesture.End(e.Sequence, float64(e.Position.X), float64(e.Position.Y), e.Inside)
		if !click {
			return true
		}
		if cubeGesture {
			if face, ok := viewport.CubeViewAt(s.Camera, e.Position.X, e.Position.Y, e.Size.X, e.Size.Y, scale); ok {
				v.view(face)
			}
			return true
		}
		if id, err := s.Pick(float64(e.Position.X), float64(e.Position.Y), float64(e.Size.X), float64(e.Size.Y)); err == nil && v.Select != nil {
			v.Select(id)
		}
		return true
	case GPUCancel:
		if !v.gesture.Cancel(e.Sequence) {
			return false
		}
		v.cubeGesture = false
		s.Camera = v.beforeCamera
		v.changed()
		return true
	case GPUPointerLeave:
		if s.HoveredID != "" {
			s.HoveredID = ""
			v.changed()
		}
		return false
	case GPUScroll:
		if !e.Inside {
			return false
		}
		if e.Modifiers&GPUShift != 0 || math.Abs(float64(e.ScrollX)) > math.Abs(float64(e.ScrollY)) {
			_ = s.Camera.Pan(float64(e.ScrollX), float64(e.ScrollY), float64(e.Size.X), float64(e.Size.Y))
		} else {
			_ = s.Camera.Zoom(math.Exp(math.Max(-5, math.Min(5, float64(e.ScrollY)/scale*.003))))
		}
		v.changed()
		if v.CameraChanged != nil {
			v.CameraChanged(s.Camera)
		}
		return true
	case GPUContextMenu:
		if v.Context != nil {
			v.Context(e)
			return true
		}
		return false
	case GPUKey:
		switch strings.ToLower(e.Key) {
		case "leftarrow":
			_ = s.Camera.Orbit(-.1, 0)
		case "rightarrow":
			_ = s.Camera.Orbit(.1, 0)
		case "uparrow":
			_ = s.Camera.Orbit(0, .1)
		case "downarrow":
			_ = s.Camera.Orbit(0, -.1)
		case "home":
			v.view(viewport.Isometric)
			return true
		default:
			return false
		}
		v.changed()
		return true
	}
	return false
}

type scene3DRenderer struct {
	scene                   *viewport.Scene
	previous                []viewport.Object
	selected, hover         string
	grid                    bool
	generation              uint64
	main, cube              *wgpu.RenderPipeline
	mainGroup, cubeGroup    *wgpu.BindGroup
	vertices, indices       *wgpu.Buffer
	vertexBytes, indexBytes uint64
	mainCount, cubeCount    uint32
}

func (r *scene3DRenderer) release() {
	if r.mainGroup != nil {
		r.mainGroup.Release()
		r.mainGroup = nil
	}
	if r.cubeGroup != nil {
		r.cubeGroup.Release()
		r.cubeGroup = nil
	}
	if r.main != nil {
		r.main.Release()
		r.main = nil
	}
	if r.cube != nil {
		r.cube.Release()
		r.cube = nil
	}
	r.generation = 0
	r.previous = nil
	r.vertices = nil
	r.indices = nil
}
func (r *scene3DRenderer) prepare(f GPUFrame) error {
	if e := r.scene.Validate(); e != nil {
		return e
	}
	if f.Size.X <= 0 || f.Size.Y <= 0 {
		return fmt.Errorf("scene3d: empty target")
	}
	fresh := r.main == nil || r.generation != f.ResourceGeneration
	if fresh {
		var err error
		mainBuffer, err := f.Resources.Buffer("scene3d.camera", wgpu.BufferDescriptor{Size: 64, Usage: wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst})
		if err != nil {
			return err
		}
		cubeBuffer, err := f.Resources.Buffer("scene3d.cube-camera", wgpu.BufferDescriptor{Size: 64, Usage: wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst})
		if err != nil {
			return err
		}
		shader, err := f.Device.TryCreateShaderModule(&wgpu.ShaderModuleDescriptor{WGSLSource: &wgpu.ShaderSourceWGSL{Code: scene3DShader}})
		if err != nil {
			return err
		}
		defer shader.Release()
		desc := wgpu.RenderPipelineDescriptor{Label: "vellum.scene3d", Vertex: wgpu.VertexState{Module: shader, EntryPoint: "vs_main", Buffers: []wgpu.VertexBufferLayout{{ArrayStride: 28, StepMode: wgpu.VertexStepModeVertex, Attributes: []wgpu.VertexAttribute{{Format: wgpu.VertexFormatFloat32x3, ShaderLocation: 0}, {Format: wgpu.VertexFormatFloat32x4, Offset: 12, ShaderLocation: 1}}}}}, Fragment: &wgpu.FragmentState{Module: shader, EntryPoint: "fs_main", Targets: []wgpu.ColorTargetState{{Format: f.Format, Blend: &wgpu.BlendStatePremultipliedAlphaBlending, WriteMask: wgpu.ColorWriteMaskAll}}}, Primitive: wgpu.PrimitiveState{Topology: wgpu.PrimitiveTopologyTriangleList, FrontFace: wgpu.FrontFaceCCW, CullMode: wgpu.CullModeBack}, Multisample: wgpu.MultisampleState{Count: f.Samples, Mask: 0xffffffff}, DepthStencil: &wgpu.DepthStencilState{Format: f.DepthFormat, DepthWriteEnabled: wgpu.OptionalBoolTrue, DepthCompare: wgpu.CompareFunctionLess}}
		r.main, err = f.Device.TryCreateRenderPipeline(&desc)
		if err != nil {
			return err
		}
		desc.Label = "vellum.scene3d.orientation"
		desc.DepthStencil = &wgpu.DepthStencilState{Format: f.DepthFormat, DepthWriteEnabled: wgpu.OptionalBoolFalse, DepthCompare: wgpu.CompareFunctionAlways}
		r.cube, err = f.Device.TryCreateRenderPipeline(&desc)
		if err != nil {
			return err
		}
		layout := r.main.GetBindGroupLayout(0)
		r.mainGroup, err = f.Device.TryCreateBindGroup(&wgpu.BindGroupDescriptor{Layout: layout, Entries: []wgpu.BindGroupEntry{{Binding: 0, Buffer: mainBuffer, Size: 64}}})
		layout.Release()
		if err != nil {
			return err
		}
		layout = r.cube.GetBindGroupLayout(0)
		r.cubeGroup, err = f.Device.TryCreateBindGroup(&wgpu.BindGroupDescriptor{Layout: layout, Entries: []wgpu.BindGroupEntry{{Binding: 0, Buffer: cubeBuffer, Size: 64}}})
		layout.Release()
		if err != nil {
			return err
		}
		r.generation = f.ResourceGeneration
	}
	if fresh || !slices.Equal(r.previous, r.scene.Objects) || r.selected != r.scene.SelectedID || r.hover != r.scene.HoveredID || r.grid != r.scene.Grid {
		vertices, indices, main, cube := buildScene3DMesh(r.scene)
		var err error
		r.vertexBytes = uint64(len(vertices) * 4)
		r.indexBytes = uint64(len(indices) * 4)
		r.vertices, err = f.Resources.Buffer("scene3d.vertices", wgpu.BufferDescriptor{Size: r.vertexBytes, Usage: wgpu.BufferUsageVertex | wgpu.BufferUsageCopyDst})
		if err != nil {
			return err
		}
		r.indices, err = f.Resources.Buffer("scene3d.indices", wgpu.BufferDescriptor{Size: r.indexBytes, Usage: wgpu.BufferUsageIndex | wgpu.BufferUsageCopyDst})
		if err != nil {
			return err
		}
		if err = f.Resources.WriteBuffer("scene3d.vertices", 0, wgpu.ToBytes(vertices)); err != nil {
			return err
		}
		if err = f.Resources.WriteBuffer("scene3d.indices", 0, wgpu.ToBytes(indices)); err != nil {
			return err
		}
		r.mainCount, r.cubeCount = main, cube
		r.previous = slices.Clone(r.scene.Objects)
		r.selected, r.hover, r.grid = r.scene.SelectedID, r.scene.HoveredID, r.scene.Grid
	}
	matrix, err := r.scene.Camera.Matrix(float64(f.Size.X), float64(f.Size.Y))
	if err != nil {
		return err
	}
	if err = f.Resources.WriteBuffer("scene3d.camera", 0, wgpu.ToBytes(matrix[:])); err != nil {
		return err
	}
	matrix, err = viewport.CubeCamera(r.scene.Camera).Matrix(1, 1)
	if err != nil {
		return err
	}
	return f.Resources.WriteBuffer("scene3d.cube-camera", 0, wgpu.ToBytes(matrix[:]))
}
func (r *scene3DRenderer) draw(f GPUFrame, p *wgpu.RenderPassEncoder) error {
	p.SetPipeline(r.main)
	p.SetBindGroup(0, r.mainGroup, nil)
	p.SetVertexBuffer(0, r.vertices, 0, r.vertexBytes)
	p.SetIndexBuffer(r.indices, wgpu.IndexFormatUint32, 0, r.indexBytes)
	p.DrawIndexed(r.mainCount, 1, 0, 0, 0)
	if r.scene.ShowCube {
		rect := viewport.CubeRect(f.Size.X, f.Size.Y, float64(f.DpScale))
		if !rect.Empty() {
			p.SetViewport(float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), 0, 1)
			p.SetPipeline(r.cube)
			p.SetBindGroup(0, r.cubeGroup, nil)
			p.DrawIndexed(r.cubeCount, 1, r.mainCount, 0, 0)
		}
	}
	return nil
}

type sceneMesh struct {
	vertices []float32
	indices  []uint32
}

func linearSceneColor(c uint8) float32 {
	x := float64(c) / 255
	if x <= .04045 {
		return float32(x / 12.92)
	}
	return float32(math.Pow((x+.055)/1.055, 2.4))
}
func (m *sceneMesh) quad(center, u, v viewport.Vec3, c color.NRGBA) {
	base := uint32(len(m.vertices) / 7)
	for _, p := range []viewport.Vec3{center.Sub(u).Sub(v), center.Add(u).Sub(v), center.Add(u).Add(v), center.Sub(u).Add(v)} {
		m.vertices = append(m.vertices, float32(p.X), float32(p.Y), float32(p.Z), linearSceneColor(c.R), linearSceneColor(c.G), linearSceneColor(c.B), 1)
	}
	m.indices = append(m.indices, base, base+1, base+2, base, base+2, base+3)
}

var cubeFaces = []struct {
	normal, u, v viewport.Vec3
	label        string
}{
	{viewport.Vec3{Z: 1}, viewport.Vec3{X: 1}, viewport.Vec3{Y: 1}, "FRONT"},
	{viewport.Vec3{Z: -1}, viewport.Vec3{X: -1}, viewport.Vec3{Y: 1}, "BACK"},
	{viewport.Vec3{X: 1}, viewport.Vec3{Z: -1}, viewport.Vec3{Y: 1}, "RIGHT"},
	{viewport.Vec3{X: -1}, viewport.Vec3{Z: 1}, viewport.Vec3{Y: 1}, "LEFT"},
	{viewport.Vec3{Y: 1}, viewport.Vec3{X: 1}, viewport.Vec3{Z: -1}, "TOP"},
	{viewport.Vec3{Y: -1}, viewport.Vec3{X: 1}, viewport.Vec3{Z: 1}, "BOTTOM"},
}

// Original stroke geometry keeps face names continuous and centered, without
// embedding a font or changing the existing rendering/picking pipeline.
var cubeStrokes = map[rune][][4]float64{
	'A': {{0, 0, .5, 1}, {.5, 1, 1, 0}, {.2, .4, .8, .4}},
	'B': {{0, 0, 0, 1}, {0, 1, .8, 1}, {.8, 1, 1, .75}, {1, .75, .8, .5}, {.8, .5, 0, .5}, {.8, .5, 1, .25}, {1, .25, .8, 0}, {.8, 0, 0, 0}},
	'C': {{1, 1, 0, 1}, {0, 1, 0, 0}, {0, 0, 1, 0}},
	'E': {{0, 0, 0, 1}, {0, 1, 1, 1}, {0, .5, .8, .5}, {0, 0, 1, 0}},
	'F': {{0, 0, 0, 1}, {0, 1, 1, 1}, {0, .5, .8, .5}},
	'G': {{1, 1, 0, 1}, {0, 1, 0, 0}, {0, 0, 1, 0}, {1, 0, 1, .5}, {1, .5, .5, .5}},
	'H': {{0, 0, 0, 1}, {1, 0, 1, 1}, {0, .5, 1, .5}},
	'I': {{0, 1, 1, 1}, {.5, 1, .5, 0}, {0, 0, 1, 0}},
	'K': {{0, 0, 0, 1}, {0, .5, 1, 1}, {0, .5, 1, 0}},
	'L': {{0, 1, 0, 0}, {0, 0, 1, 0}},
	'M': {{0, 0, 0, 1}, {0, 1, .5, .4}, {.5, .4, 1, 1}, {1, 1, 1, 0}},
	'N': {{0, 0, 0, 1}, {0, 1, 1, 0}, {1, 0, 1, 1}},
	'O': {{.15, 0, .85, 0}, {.85, 0, 1, .15}, {1, .15, 1, .85}, {1, .85, .85, 1}, {.85, 1, .15, 1}, {.15, 1, 0, .85}, {0, .85, 0, .15}, {0, .15, .15, 0}},
	'P': {{0, 0, 0, 1}, {0, 1, 1, 1}, {1, 1, 1, .5}, {1, .5, 0, .5}},
	'R': {{0, 0, 0, 1}, {0, 1, 1, 1}, {1, 1, 1, .5}, {1, .5, 0, .5}, {.4, .5, 1, 0}},
	'T': {{0, 1, 1, 1}, {.5, 1, .5, 0}},
}

// cubeLabel lays out only the six fixed ASCII names in cubeFaces.
func (m *sceneMesh) cubeLabel(label string, normal, u, v viewport.Vec3) {
	const width, height, gap, stroke = .18, .34, .055, .014
	left := -(float64(len(label))*width + float64(len(label)-1)*gap) / 2
	for i, ch := range label {
		for _, line := range cubeStrokes[ch] {
			ax, ay := left+float64(i)*(width+gap)+line[0]*width, (line[1]-.5)*height
			bx, by := left+float64(i)*(width+gap)+line[2]*width, (line[3]-.5)*height
			dx, dy := bx-ax, by-ay
			length := math.Hypot(dx, dy)
			if length == 0 {
				continue
			}
			center := normal.Mul(1.012).Add(u.Mul((ax + bx) / 2)).Add(v.Mul((ay + by) / 2))
			along := u.Mul(dx / 2).Add(v.Mul(dy / 2))
			across := u.Mul(-dy / length * stroke).Add(v.Mul(dx / length * stroke))
			m.quad(center, along, across, color.NRGBA{R: 45, G: 55, B: 70, A: 255})
		}
	}
}

func (m *sceneMesh) box(b viewport.Box, c color.NRGBA) {
	center := b.Min.Add(b.Max).Mul(.5)
	half := b.Max.Sub(b.Min).Mul(.5)
	for i, f := range cubeFaces {
		extent := func(a viewport.Vec3) float64 {
			return math.Abs(a.X)*half.X + math.Abs(a.Y)*half.Y + math.Abs(a.Z)*half.Z
		}
		shade := []float64{.94, .74, .78, .86, 1, .65}[i]
		color := color.NRGBA{uint8(float64(c.R) * shade), uint8(float64(c.G) * shade), uint8(float64(c.B) * shade), 255}
		m.quad(center.Add(f.normal.Mul(extent(f.normal))), f.u.Mul(extent(f.u)), f.v.Mul(extent(f.v)), color)
	}
}

// cylinderX is display geometry only. Scene.Pick deliberately retains Bounds.
func (m *sceneMesh) cylinderX(b viewport.Box, c color.NRGBA) {
	const segments = 48
	center, half := b.Min.Add(b.Max).Mul(.5), b.Max.Sub(b.Min).Mul(.5)
	triangle := func(a, b, cpoint viewport.Vec3, tint color.NRGBA) {
		base := uint32(len(m.vertices) / 7)
		for _, p := range []viewport.Vec3{a, b, cpoint} {
			m.vertices = append(m.vertices, float32(p.X), float32(p.Y), float32(p.Z), linearSceneColor(tint.R), linearSceneColor(tint.G), linearSceneColor(tint.B), 1)
		}
		m.indices = append(m.indices, base, base+1, base+2)
	}
	shade := func(f float64) color.NRGBA {
		return color.NRGBA{uint8(float64(c.R) * f), uint8(float64(c.G) * f), uint8(float64(c.B) * f), 255}
	}
	point := func(x, angle float64) viewport.Vec3 {
		return viewport.Vec3{X: x, Y: center.Y + half.Y*math.Cos(angle), Z: center.Z + half.Z*math.Sin(angle)}
	}
	for i := 0; i < segments; i++ {
		a, next := float64(i)*2*math.Pi/segments, float64(i+1)*2*math.Pi/segments
		leftA, leftB, rightA, rightB := point(center.X-half.X, a), point(center.X-half.X, next), point(center.X+half.X, a), point(center.X+half.X, next)
		tint := shade(.65 + .35*math.Max(0, .8*math.Cos((a+next)/2)+.6*math.Sin((a+next)/2)))
		triangle(leftA, rightB, rightA, tint)
		triangle(leftA, leftB, rightB, tint)
		triangle(viewport.Vec3{X: center.X - half.X, Y: center.Y, Z: center.Z}, leftB, leftA, shade(.78))
		triangle(viewport.Vec3{X: center.X + half.X, Y: center.Y, Z: center.Z}, rightA, rightB, shade(.94))
	}
}

func buildScene3DMesh(s *viewport.Scene) ([]float32, []uint32, uint32, uint32) {
	m := &sceneMesh{}
	if s.Grid {
		for i := -10; i <= 10; i++ {
			n := float64(i) * .5
			c := color.NRGBA{215, 222, 230, 255}
			if i == 0 {
				c = color.NRGBA{179, 190, 205, 255}
			}
			m.box(viewport.Box{Min: viewport.Vec3{X: n, Y: -.225, Z: -5}, Max: viewport.Vec3{X: n + .008, Y: -.22, Z: 5}}, c)
			m.box(viewport.Box{Min: viewport.Vec3{X: -5, Y: -.225, Z: n}, Max: viewport.Vec3{X: 5, Y: -.22, Z: n + .008}}, c)
		}
	}
	for _, o := range s.Objects {
		if !o.Visible {
			continue
		}
		c := o.Color
		if o.ID == s.SelectedID {
			c = color.NRGBA{65, 133, 225, 255}
		} else if o.ID == s.HoveredID {
			c.R = uint8((int(c.R)*3 + 255) / 4)
			c.G = uint8((int(c.G)*3 + 255) / 4)
			c.B = uint8((int(c.B)*3 + 255) / 4)
		}
		if o.Shape == viewport.CylinderXShape {
			m.cylinderX(o.Bounds, c)
		} else {
			m.box(o.Bounds, c)
		}
	}
	mainCount := uint32(len(m.indices))
	// Original geometry only; existing pipeline, scale and pick rectangles stay.
	m.box(viewport.Box{Min: viewport.Vec3{X: -1, Y: -1, Z: -1}, Max: viewport.Vec3{X: 1, Y: 1, Z: 1}}, color.NRGBA{R: 241, G: 243, B: 247, A: 255})
	for _, face := range cubeFaces {
		m.cubeLabel(face.label, face.normal, face.u, face.v)
	}

	return m.vertices, m.indices, mainCount, uint32(len(m.indices)) - mainCount
}
