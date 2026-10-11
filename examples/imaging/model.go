package main

import (
	_ "embed"
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/dshills/vellum/drivers/cogent"
	"github.com/oliverbestmann/webgpu/wgpu"
)

const imageSize = 256
const storageBytes = imageSize * imageSize * 16
const floatTextureBytes = imageSize * imageSize * 8
const uniformBytes = 32

//go:embed imaging.wgsl
var imagingShader string

type imagingProgram struct {
	mu                                     sync.Mutex
	exposure                               float32
	generation                             uint64
	compute                                *wgpu.ComputePipeline
	floatPass, display                     *wgpu.RenderPipeline
	computeGroup, floatGroup, displayGroup *wgpu.BindGroup
	hdrView                                *wgpu.TextureView
	storage, uniform                       *wgpu.Buffer  // borrowed helper handles
	hdr                                    *wgpu.Texture // borrowed helper handle
	releases                               uint64
}

func newImagingProgram() (*imagingProgram, *cogent.GPUProgram) {
	m := &imagingProgram{exposure: 1}
	return m, &cogent.GPUProgram{Prepare: m.prepare, Draw: m.draw, Release: m.release}
}
func (m *imagingProgram) setExposure(value float32) error {
	if value <= 0 || value > 16 || math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
		return errors.New("exposure must be finite and in (0,16]")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.exposure = value
	return nil
}
func (m *imagingProgram) exposureValue() float32 { m.mu.Lock(); defer m.mu.Unlock(); return m.exposure }
func uniformData(exposure float32, size image.Point) [8]float32 {
	return [8]float32{exposure, imageSize, imageSize, 0, float32(size.X), float32(size.Y), 0, 0}
}

func (m *imagingProgram) release() { m.mu.Lock(); defer m.mu.Unlock(); m.releaseLocked() }
func (m *imagingProgram) releaseLocked() {
	// Every dependent program-owned group/pipeline/view retires before the
	// framework frees its helper-owned buffers and float attachment.
	for _, group := range []*wgpu.BindGroup{m.computeGroup, m.floatGroup, m.displayGroup} {
		if group != nil {
			group.Release()
		}
	}
	m.computeGroup, m.floatGroup, m.displayGroup = nil, nil, nil
	if m.compute != nil {
		m.compute.Release()
		m.compute = nil
	}
	for _, pipeline := range []*wgpu.RenderPipeline{m.floatPass, m.display} {
		if pipeline != nil {
			pipeline.Release()
		}
	}
	m.floatPass, m.display = nil, nil
	if m.hdrView != nil {
		m.hdrView.Release()
		m.hdrView = nil
	}
	m.storage, m.uniform, m.hdr = nil, nil, nil
	m.generation = 0
	m.releases++
}
func group(device *wgpu.Device, layout *wgpu.BindGroupLayout, entries []wgpu.BindGroupEntry) (*wgpu.BindGroup, error) {
	defer layout.Release()
	return device.TryCreateBindGroup(&wgpu.BindGroupDescriptor{Layout: layout, Entries: entries})
}
func (m *imagingProgram) initialize(f cogent.GPUFrame) error {
	var err error
	m.storage, err = f.Resources.Buffer("hdr.pixels", wgpu.BufferDescriptor{Size: storageBytes, Usage: wgpu.BufferUsageStorage})
	if err != nil {
		return err
	}
	m.uniform, err = f.Resources.Buffer("hdr.params", wgpu.BufferDescriptor{Size: uniformBytes, Usage: wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst})
	if err != nil {
		return err
	}
	m.hdr, err = f.Resources.Texture("hdr.float", cogent.GPUTexture{Width: imageSize, Height: imageSize, Format: wgpu.TextureFormatRGBA16Float, Samples: 1, Usage: wgpu.TextureUsageRenderAttachment | wgpu.TextureUsageTextureBinding})
	if err != nil {
		return err
	}
	m.hdrView, err = m.hdr.TryCreateView(nil)
	if err != nil {
		return err
	}
	shader, err := f.Device.TryCreateShaderModule(&wgpu.ShaderModuleDescriptor{Label: "vellum.imaging", WGSLSource: &wgpu.ShaderSourceWGSL{Code: imagingShader}})
	if err != nil {
		return err
	}
	defer shader.Release()
	m.compute, err = f.Device.TryCreateComputePipeline(&wgpu.ComputePipelineDescriptor{Label: "vellum.imaging.compute", Compute: wgpu.ProgrammableStageDescriptor{Module: shader, EntryPoint: "cs_main"}})
	if err != nil {
		return err
	}
	m.computeGroup, err = group(f.Device, m.compute.GetBindGroupLayout(0), []wgpu.BindGroupEntry{{Binding: 0, Buffer: m.uniform, Size: uniformBytes}, {Binding: 1, Buffer: m.storage, Size: storageBytes}})
	if err != nil {
		return err
	}
	descriptor := func(format wgpu.TextureFormat, samples uint32, entry string, blend *wgpu.BlendState) *wgpu.RenderPipelineDescriptor {
		return &wgpu.RenderPipelineDescriptor{Label: "vellum.imaging." + entry, Vertex: wgpu.VertexState{Module: shader, EntryPoint: "vs_main"}, Fragment: &wgpu.FragmentState{Module: shader, EntryPoint: entry, Targets: []wgpu.ColorTargetState{{Format: format, Blend: blend, WriteMask: wgpu.ColorWriteMaskAll}}}, Primitive: wgpu.PrimitiveState{Topology: wgpu.PrimitiveTopologyTriangleList, FrontFace: wgpu.FrontFaceCCW, CullMode: wgpu.CullModeNone}, Multisample: wgpu.MultisampleState{Count: samples, Mask: 0xffffffff}}
	}
	m.floatPass, err = f.Device.TryCreateRenderPipeline(descriptor(wgpu.TextureFormatRGBA16Float, 1, "fs_float", nil))
	if err != nil {
		return err
	}
	m.floatGroup, err = group(f.Device, m.floatPass.GetBindGroupLayout(0), []wgpu.BindGroupEntry{{Binding: 3, Buffer: m.storage, Size: storageBytes}})
	if err != nil {
		return err
	}
	m.display, err = f.Device.TryCreateRenderPipeline(descriptor(f.Format, f.Samples, "fs_display", &wgpu.BlendStatePremultipliedAlphaBlending))
	if err != nil {
		return err
	}
	m.displayGroup, err = group(f.Device, m.display.GetBindGroupLayout(0), []wgpu.BindGroupEntry{{Binding: 0, Buffer: m.uniform, Size: uniformBytes}, {Binding: 2, TextureView: m.hdrView}})
	if err != nil {
		return err
	}
	m.generation = f.ResourceGeneration
	return nil
}
func (m *imagingProgram) prepare(f cogent.GPUFrame) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if f.Version != cogent.GPUContextVersion || f.Encoder == nil || f.Format != wgpu.TextureFormatRGBA8UnormSrgb || f.Samples != 4 {
		return errors.New("imaging GPU context mismatch")
	}
	if m.compute == nil || m.generation != f.ResourceGeneration {
		m.releaseLocked()
		if err := m.initialize(f); err != nil {
			return err
		}
	}
	data := uniformData(m.exposure, f.Size)
	if err := f.Resources.WriteBuffer("hdr.params", 0, wgpu.ToBytes(data[:])); err != nil {
		return err
	}
	if err := m.encodeCompute(f); err != nil {
		return err
	}
	pass, err := f.Encoder.TryBeginRenderPass(&wgpu.RenderPassDescriptor{Label: "vellum.imaging.float", ColorAttachments: []wgpu.RenderPassColorAttachment{{View: m.hdrView, LoadOp: wgpu.LoadOpClear, StoreOp: wgpu.StoreOpStore, ClearValue: wgpu.Color{}}}})
	if err != nil {
		return err
	}
	defer pass.Release()
	pass.SetPipeline(m.floatPass)
	pass.SetBindGroup(0, m.floatGroup, nil)
	pass.Draw(3, 1, 0, 0)
	return pass.TryEnd()
}
func (m *imagingProgram) encodeCompute(f cogent.GPUFrame) (err error) {
	// This pinned binding exposes only BeginComputePass (which can panic), not a
	// checked Try variant. Convert that creation refusal into a program error.
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("compute encoding: %v", p)
		}
	}()
	pass := f.Encoder.BeginComputePass(&wgpu.ComputePassDescriptor{Label: "vellum.imaging.compute"})
	defer pass.Release()
	pass.SetPipeline(m.compute)
	pass.SetBindGroup(0, m.computeGroup, nil)
	pass.DispatchWorkgroups(imageSize/8, imageSize/8, 1)
	return pass.TryEnd()
}
func (m *imagingProgram) draw(f cogent.GPUFrame, pass *wgpu.RenderPassEncoder) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.display == nil || m.generation != f.ResourceGeneration || f.Encoder != nil {
		return errors.New("imaging Draw resource generation mismatch")
	}
	if f.Clip.Empty() {
		return nil
	}
	pass.SetScissorRect(uint32(f.Clip.Min.X), uint32(f.Clip.Min.Y), uint32(f.Clip.Dx()), uint32(f.Clip.Dy()))
	pass.SetPipeline(m.display)
	pass.SetBindGroup(0, m.displayGroup, nil)
	pass.Draw(3, 1, 0, 0)
	return nil
}
func (m *imagingProgram) resourceIdentity() (*wgpu.Buffer, *wgpu.Texture, uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.storage, m.hdr, m.releases
}

// referencePixel computes the expected straight-alpha sRGB readback. It is
// independent of GPU execution and distinguishes HDR preservation from clamping.
func referencePixel(exposure float64, right, transparent bool) color.NRGBA {
	if transparent {
		return color.NRGBA{}
	}
	value := [3]float64{2, .5, 0}
	if right {
		value = [3]float64{0, .25, 3}
	}
	out := [3]uint8{}
	for i, x := range value {
		x *= exposure
		x = x / (1 + x)
		if x <= .0031308 {
			x *= 12.92
		} else {
			x = 1.055*math.Pow(x, 1/2.4) - .055
		}
		out[i] = uint8(math.Round(x * 255))
	}
	return color.NRGBA{R: out[0], G: out[1], B: out[2], A: 128}
}
