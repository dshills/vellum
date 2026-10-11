package cogent

import (
	_ "embed"
	"encoding/binary"
	"fmt"
	"image"
	"math"

	"cogentcore.org/core/gpu"
	"github.com/dshills/vellum/canvas"
	"github.com/oliverbestmann/webgpu/wgpu"
)

//go:embed image_canvas.wgsl
var imageShader string

const imageUniformBytes uint64 = 32

// imageResources borrows the window device. Its texel upload baseline is separate
// from meshResources.layer, which identifies only a completed offscreen frame.
type imageResources struct {
	pipeline *wgpu.RenderPipeline
	layout   *wgpu.BindGroupLayout
	uniform  *wgpu.Buffer
	texture  *wgpu.Texture
	view     *wgpu.TextureView
	group    *wgpu.BindGroup
	size     image.Point
	source   *canvas.TiledImage
}

func (p *imageResources) releaseTexture() {
	p.source = nil
	p.size = image.Point{}
	if p.group != nil {
		p.group.Release()
		p.group = nil
	}
	if p.view != nil {
		p.view.Release()
		p.view = nil
	}
	if p.texture != nil {
		p.texture.Release()
		p.texture = nil
	}
}
func (p *imageResources) release() {
	p.releaseTexture()
	if p.pipeline != nil {
		p.pipeline.Release()
		p.pipeline = nil
	}
	if p.uniform != nil {
		p.uniform.Release()
		p.uniform = nil
	}
	if p.layout != nil {
		p.layout.Release()
		p.layout = nil
	}
}
func newImageResources(device *gpu.Device, format wgpu.TextureFormat, samples int) (p *imageResources, err error) {
	p = &imageResources{}
	allocated := p
	defer func() {
		if err != nil {
			allocated.release()
		}
	}()
	p.uniform, err = device.Device.TryCreateBuffer(&wgpu.BufferDescriptor{Label: "vellum.image.view", Size: imageUniformBytes, Usage: wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst})
	if err != nil {
		return nil, err
	}
	p.layout, err = device.Device.TryCreateBindGroupLayout(&wgpu.BindGroupLayoutDescriptor{Label: "vellum.image.layout", Entries: []wgpu.BindGroupLayoutEntry{
		{Binding: 0, Visibility: wgpu.ShaderStageVertex | wgpu.ShaderStageFragment, Buffer: wgpu.BufferBindingLayout{Type: wgpu.BufferBindingTypeUniform, MinBindingSize: imageUniformBytes}},
		{Binding: 1, Visibility: wgpu.ShaderStageFragment, Texture: wgpu.TextureBindingLayout{SampleType: wgpu.TextureSampleTypeFloat, ViewDimension: wgpu.TextureViewDimension2D}},
	}})
	if err != nil {
		return nil, err
	}
	layout, err := device.Device.TryCreatePipelineLayout(&wgpu.PipelineLayoutDescriptor{Label: "vellum.image.pipeline.layout", BindGroupLayouts: []*wgpu.BindGroupLayout{p.layout}})
	if err != nil {
		return nil, err
	}
	defer layout.Release()
	shader, err := device.Device.TryCreateShaderModule(&wgpu.ShaderModuleDescriptor{Label: "vellum.image.shader", WGSLSource: &wgpu.ShaderSourceWGSL{Code: imageShader}})
	if err != nil {
		return nil, err
	}
	defer shader.Release()
	p.pipeline, err = device.Device.TryCreateRenderPipeline(&wgpu.RenderPipelineDescriptor{
		Label: "vellum.image", Layout: layout,
		Vertex:      wgpu.VertexState{Module: shader, EntryPoint: "vs_image"},
		Fragment:    &wgpu.FragmentState{Module: shader, EntryPoint: "fs_image", Targets: []wgpu.ColorTargetState{{Format: format, Blend: &wgpu.BlendStatePremultipliedAlphaBlending, WriteMask: wgpu.ColorWriteMaskAll}}},
		Primitive:   wgpu.PrimitiveState{Topology: wgpu.PrimitiveTopologyTriangleList, FrontFace: wgpu.FrontFaceCCW, CullMode: wgpu.CullModeNone},
		Multisample: wgpu.MultisampleState{Count: uint32(samples), Mask: 0xffffffff},
	})
	if err != nil {
		return nil, err
	}
	return p, nil
}
func imageBytes(layer canvas.ImageLayer) uint64 {
	size := layer.Image.Size()
	if size.X == 0 {
		return 0
	}
	return uint64(size.X)*uint64(size.Y)*4 + imageUniformBytes
}
func imageAdmission(layer canvas.ImageLayer, limits wgpu.Limits) error {
	if layer.Image == nil {
		return nil
	}
	if err := layer.Validate(); err != nil {
		return err
	}
	size := layer.Image.Size()
	if uint64(size.X) > uint64(limits.MaxTextureDimension2D) || uint64(size.Y) > uint64(limits.MaxTextureDimension2D) || imageUniformBytes > limits.MaxBufferSize {
		return fmt.Errorf("drawing: image exceeds logical device limits")
	}
	return nil
}
func imageTransform(layer canvas.ImageLayer, view canvas.Viewport) [8]float32 {
	x, y := 2*view.Zoom/view.Width, -2*view.Zoom/view.Height
	return [8]float32{
		float32((layer.Bounds.Max.X - layer.Bounds.Min.X) * x), float32((layer.Bounds.Max.Y - layer.Bounds.Min.Y) * y),
		float32((layer.Bounds.Min.X - view.Center.X) * x), float32((layer.Bounds.Min.Y - view.Center.Y) * y),
		float32(layer.Filter), 0, 0, 0,
	}
}
func (p *imageResources) upload(device *gpu.Device, layer canvas.ImageLayer, stats *DrawingStats) error {
	size := layer.Image.Size()
	stats.ImageUniformBytes = imageUniformBytes
	if p.texture == nil || p.size != size {
		// Retire the old allocation before replacing it; budget tracks one image.
		p.releaseTexture()
		stats.ImageBytes = 0
		texture, err := device.Device.TryCreateTexture(&wgpu.TextureDescriptor{Label: "vellum.image.texels", Size: wgpu.Extent3D{Width: uint32(size.X), Height: uint32(size.Y), DepthOrArrayLayers: 1}, MipLevelCount: 1, SampleCount: 1, Dimension: wgpu.TextureDimension2D, Format: wgpu.TextureFormatRGBA8UnormSrgb, Usage: wgpu.TextureUsageTextureBinding | wgpu.TextureUsageCopyDst})
		if err != nil {
			return err
		}
		p.texture = texture
		p.view, err = texture.TryCreateView(nil)
		if err != nil {
			p.releaseTexture()
			return err
		}
		p.group, err = device.Device.TryCreateBindGroup(&wgpu.BindGroupDescriptor{Label: "vellum.image.binding", Layout: p.layout, Entries: []wgpu.BindGroupEntry{{Binding: 0, Buffer: p.uniform, Size: imageUniformBytes}, {Binding: 1, TextureView: p.view}}})
		if err != nil {
			p.releaseTexture()
			return err
		}
		p.size = size
		stats.ImageTextureAllocations++
	}
	stats.ImageBytes = uint64(size.X) * uint64(size.Y) * 4
	stats.ImageUniformBytes = imageUniformBytes
	return p.uploadTiles(layer.Image, stats, func(tile canvas.ImageTile) error {
		bounds := tile.Bounds()
		return device.Queue.TryWriteTexture(&wgpu.TexelCopyTextureInfo{Texture: p.texture, Aspect: wgpu.TextureAspectAll, Origin: wgpu.Origin3D{X: uint32(bounds.Min.X), Y: uint32(bounds.Min.Y)}}, tile.Pixels(),
			&wgpu.TexelCopyBufferLayout{BytesPerRow: uint32(bounds.Dx() * 4), RowsPerImage: uint32(bounds.Dy())},
			&wgpu.Extent3D{Width: uint32(bounds.Dx()), Height: uint32(bounds.Dy()), DepthOrArrayLayers: 1})
	})
}

// The injectable writer tests failure recovery at the real upload-baseline seam;
// native rendering supplies checked Queue.TryWriteTexture, never a mock renderer.
func (p *imageResources) uploadTiles(source *canvas.TiledImage, stats *DrawingStats, write func(canvas.ImageTile) error) error {
	previous := p.source
	p.source = nil
	for _, tile := range source.ChangedTiles(previous) {
		if err := write(tile); err != nil {
			return err
		}
		stats.ImageTileUploads++
		stats.ImageUploadBytes += uint64(tile.Bounds().Dx() * tile.Bounds().Dy() * 4)
	}
	p.source = source
	return nil
}

func (p *imageResources) draw(device *gpu.Device, pass *wgpu.RenderPassEncoder, layer canvas.ImageLayer, view canvas.Viewport, stats *DrawingStats) error {
	transform := imageTransform(layer, view)
	var bytes [32]byte
	for i, v := range transform {
		binary.LittleEndian.PutUint32(bytes[i*4:], math.Float32bits(v))
	}
	if err := device.Queue.TryWriteBuffer(p.uniform, 0, bytes[:]); err != nil {
		return err
	}
	stats.UniformUploadBytes += imageUniformBytes
	pass.SetPipeline(p.pipeline)
	pass.SetBindGroup(0, p.group, nil)
	pass.Draw(6, 1, 0, 0)
	return nil
}
