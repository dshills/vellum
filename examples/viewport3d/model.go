package main

import (
	_ "embed"
	"errors"
	"math"

	"github.com/dshills/vellum/drivers/cogent"
	"github.com/oliverbestmann/webgpu/wgpu"
)

//go:embed cube.wgsl
var shaderCode string

type camera struct{ yaw, pitch, distance float64 }

func defaultCamera() camera          { return camera{distance: 5} }
func (c *camera) orbit(x, y float64) { c.yaw += x; c.pitch = math.Max(-1.4, math.Min(1.4, c.pitch+y)) }
func (c *camera) dolly(factor float64) {
	if factor > 0 && !math.IsNaN(factor) && !math.IsInf(factor, 0) {
		c.distance = math.Max(2.5, math.Min(20, c.distance*factor))
	}
}
func (c camera) matrix(aspect float64) [16]float32 {
	sy, cy := math.Sincos(c.yaw)
	sp, cp := math.Sincos(c.pitch)
	view := [16]float64{cy, -sy * sp, sy * cp, 0, 0, cp, sp, 0, -sy, -cy * sp, cy * cp, 0, 0, 0, -c.distance, 1}
	near, far := .1, 100.
	f := 1 / math.Tan(math.Pi/6)
	projection := [16]float64{f / aspect, 0, 0, 0, 0, f, 0, 0, 0, 0, far / (near - far), -1, 0, 0, far * near / (near - far), 0}
	var m [16]float32
	for col := 0; col < 4; col++ {
		for row := 0; row < 4; row++ {
			var sum float64
			for k := 0; k < 4; k++ {
				sum += projection[k*4+row] * view[col*4+k]
			}
			m[col*4+row] = float32(sum)
		}
	}
	return m
}

// Near red face is deliberately submitted before the overlapping far blue face.
// A missing depth test makes the center blue; submission order cannot mask it.
func cube() ([]float32, []uint32) {
	faces := [][4][3]float32{
		{{-1, -1, 1}, {1, -1, 1}, {1, 1, 1}, {-1, 1, 1}},
		{{-1, -1, -1}, {1, -1, -1}, {1, 1, -1}, {-1, 1, -1}},
		{{-1, -1, -1}, {-1, -1, 1}, {-1, 1, 1}, {-1, 1, -1}},
		{{1, -1, -1}, {1, -1, 1}, {1, 1, 1}, {1, 1, -1}},
		{{-1, 1, -1}, {1, 1, -1}, {1, 1, 1}, {-1, 1, 1}},
		{{-1, -1, -1}, {1, -1, -1}, {1, -1, 1}, {-1, -1, 1}},
	}
	colors := [][4]float32{{1, 0, 0, 1}, {0, 0, 1, 1}, {0, 1, 0, 1}, {1, .3, 0, 1}, {1, 1, 0, 1}, {.5, 0, 1, 1}}
	var vertices []float32
	var indices []uint32
	for i, face := range faces {
		base := uint32(len(vertices) / 7)
		for _, p := range face {
			vertices = append(vertices, p[:]...)
			vertices = append(vertices, colors[i][:]...)
		}
		indices = append(indices, base, base+1, base+2, base, base+2, base+3)
	}
	return vertices, indices
}

type cubeProgram struct {
	device                       *wgpu.Device
	camera                       camera
	generation                   uint64
	pipeline                     *wgpu.RenderPipeline
	group                        *wgpu.BindGroup
	vertices, indices            *wgpu.Buffer
	saved                        *cogent.GPUResources
	clear, failPrepare, failDraw bool
	releases                     int
}

func newCubeProgram() (*cubeProgram, *cogent.GPUProgram) {
	m := &cubeProgram{camera: defaultCamera()}
	return m, &cogent.GPUProgram{Prepare: m.prepare, Draw: m.draw, Release: m.release}
}
func (m *cubeProgram) release() {
	if m.group != nil {
		m.group.Release()
		m.group = nil
	}
	if m.pipeline != nil {
		m.pipeline.Release()
		m.pipeline = nil
	}
	m.vertices = nil
	m.indices = nil
	m.generation = 0
	m.device = nil
	m.releases++
}
func (m *cubeProgram) prepare(f cogent.GPUFrame) error {
	if m.saved != nil {
		if err := m.saved.Drop("stale-probe"); err == nil {
			return errors.New("expired resource helper accepted")
		}
	}
	m.saved = f.Resources
	if f.Version != cogent.GPUContextVersion || f.Format != wgpu.TextureFormatRGBA8UnormSrgb || f.DepthFormat != wgpu.TextureFormatDepth32Float || f.Samples != 4 || f.DpScale <= 0 {
		return errors.New("GPU context metadata mismatch")
	}
	if m.pipeline == nil || m.generation != f.ResourceGeneration {
		v, i := cube()
		var err error
		m.vertices, err = f.Resources.Buffer("vertices", wgpu.BufferDescriptor{Size: uint64(len(v) * 4), Usage: wgpu.BufferUsageVertex | wgpu.BufferUsageCopyDst})
		if err != nil {
			return err
		}
		m.indices, err = f.Resources.Buffer("indices", wgpu.BufferDescriptor{Size: uint64(len(i) * 4), Usage: wgpu.BufferUsageIndex | wgpu.BufferUsageCopyDst})
		if err != nil {
			return err
		}
		uniform, err := f.Resources.Buffer("camera", wgpu.BufferDescriptor{Size: 64, Usage: wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst})
		if err != nil {
			return err
		}
		if err := f.Resources.WriteBuffer("vertices", 0, wgpu.ToBytes(v)); err != nil {
			return err
		}
		if err := f.Resources.WriteBuffer("indices", 0, wgpu.ToBytes(i)); err != nil {
			return err
		}
		shader, err := f.Device.TryCreateShaderModule(&wgpu.ShaderModuleDescriptor{WGSLSource: &wgpu.ShaderSourceWGSL{Code: shaderCode}})
		if err != nil {
			return err
		}
		defer shader.Release()
		m.pipeline, err = f.Device.TryCreateRenderPipeline(&wgpu.RenderPipelineDescriptor{
			Label: "vellum.example.cube", Vertex: wgpu.VertexState{Module: shader, EntryPoint: "vs_main", Buffers: []wgpu.VertexBufferLayout{{ArrayStride: 28, StepMode: wgpu.VertexStepModeVertex, Attributes: []wgpu.VertexAttribute{{Format: wgpu.VertexFormatFloat32x3, ShaderLocation: 0}, {Format: wgpu.VertexFormatFloat32x4, Offset: 12, ShaderLocation: 1}}}}},
			Fragment:    &wgpu.FragmentState{Module: shader, EntryPoint: "fs_main", Targets: []wgpu.ColorTargetState{{Format: f.Format, Blend: &wgpu.BlendStatePremultipliedAlphaBlending, WriteMask: wgpu.ColorWriteMaskAll}}},
			Primitive:   wgpu.PrimitiveState{Topology: wgpu.PrimitiveTopologyTriangleList, FrontFace: wgpu.FrontFaceCCW, CullMode: wgpu.CullModeNone},
			Multisample: wgpu.MultisampleState{Count: f.Samples, Mask: 0xffffffff}, DepthStencil: &wgpu.DepthStencilState{Format: f.DepthFormat, DepthWriteEnabled: wgpu.OptionalBoolTrue, DepthCompare: wgpu.CompareFunctionLess},
		})
		if err != nil {
			return err
		}
		layout := m.pipeline.GetBindGroupLayout(0)
		defer layout.Release()
		m.group, err = f.Device.TryCreateBindGroup(&wgpu.BindGroupDescriptor{Layout: layout, Entries: []wgpu.BindGroupEntry{{Binding: 0, Buffer: uniform, Size: 64}}})
		if err != nil {
			return err
		}
		// Exercise a tracked floating-point preparation attachment without claiming
		// compute or floating-point presentation qualification.
		_, err = f.Resources.Texture("scratch", cogent.GPUTexture{Width: 8, Height: 8, Samples: 1, Format: wgpu.TextureFormatRGBA16Float, Usage: wgpu.TextureUsageRenderAttachment})
		if err != nil {
			return err
		}
		if err := f.Resources.Drop("scratch"); err != nil {
			return err
		}
		m.generation = f.ResourceGeneration
		m.device = f.Device
	}
	matrix := m.camera.matrix(float64(f.Size.X) / float64(f.Size.Y))
	if err := f.Resources.WriteBuffer("camera", 0, wgpu.ToBytes(matrix[:])); err != nil {
		return err
	}
	if m.failPrepare {
		return errors.New("controlled preparation failure")
	}
	return nil
}
func (m *cubeProgram) draw(f cogent.GPUFrame, p *wgpu.RenderPassEncoder) error {
	if f.Encoder != nil {
		return errors.New("Draw received an active encoder")
	}
	if err := m.saved.Drop("stale-probe"); err == nil {
		return errors.New("Prepare helper escaped into Draw")
	}
	if m.failDraw {
		return errors.New("controlled draw failure")
	}
	if m.clear {
		return nil
	}
	p.SetPipeline(m.pipeline)
	p.SetBindGroup(0, m.group, nil)
	p.SetVertexBuffer(0, m.vertices, 0, 672)
	p.SetIndexBuffer(m.indices, wgpu.IndexFormatUint32, 0, 144)
	p.DrawIndexed(36, 1, 0, 0, 0)
	return nil
}
