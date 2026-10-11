package cogent

import (
	"errors"
	"sync/atomic"

	"cogentcore.org/core/gpu"
	"github.com/oliverbestmann/webgpu/wgpu"
)

// GPUTexture describes a tracked one-mip, one-layer 2D texture. Samples is 1 or 4.
// Supported formats have explicit accounting; compressed/array/mip resources can
// be allocated via the borrowed device under the advanced caller contract.
type GPUTexture struct {
	Width, Height uint32
	Format        wgpu.TextureFormat
	Usage         wgpu.TextureUsage
	Samples       uint32
}

// GPUResources provides named checked allocations and counted buffer uploads.
// Methods are valid only within this frame's synchronous Prepare/Draw callbacks.
// Handles are framework-owned; retaining them within the resource generation is
// permitted, releasing them yourself is not. Drop retires a named handle; the
// caller must first retire any program-owned bind group/view referencing it.
// Replacement admits the transient old+new peak before releasing the old handle.
// Pipelines, views, bind groups and raw device allocations are outside this estimate.
type GPUResources struct {
	pool  *gpuResourcePool
	token uint64
}
type trackedBuffer struct {
	desc   wgpu.BufferDescriptor
	handle *wgpu.Buffer
}
type trackedTexture struct {
	desc   GPUTexture
	handle *wgpu.Texture
	bytes  uint64
}
type gpuResourcePool struct {
	lease    atomic.Uint64
	device   *gpu.Device
	limits   wgpu.Limits
	stats    *GPUViewStats
	budget   uint64
	buffers  map[string]trackedBuffer
	textures map[string]trackedTexture
}

func (h *GPUResources) active() (*gpuResourcePool, error) {
	if h == nil || h.pool == nil || h.token == 0 || h.pool.lease.Load() != h.token {
		return nil, errors.New("GPU view: resources outside callback scope")
	}
	return h.pool, nil
}
func (p *gpuResourcePool) admit(extra uint64) error {
	return gpuBudget(p.budget, p.stats.TargetBytes, p.stats.BufferBytes, p.stats.TextureBytes, p.stats.ReadbackBytes, p.stats.AsyncReadbackBytes, extra)
}
func (h *GPUResources) Buffer(name string, desc wgpu.BufferDescriptor) (*wgpu.Buffer, error) {
	p, err := h.active()
	if err != nil {
		return nil, err
	}
	if name == "" || desc.Size == 0 || desc.Usage == 0 || desc.MappedAtCreation {
		return nil, errors.New("GPU view: invalid tracked buffer descriptor")
	}
	if desc.Size > p.limits.MaxBufferSize {
		return nil, errors.New("GPU view: buffer exceeds logical device limit")
	}
	if old, ok := p.buffers[name]; ok && old.desc == desc {
		return old.handle, nil
	}
	if err := p.admit(desc.Size); err != nil {
		return nil, err
	}
	buffer, err := p.device.Device.TryCreateBuffer(&desc)
	if err != nil {
		return nil, err
	}
	if p.buffers == nil {
		p.buffers = make(map[string]trackedBuffer)
	}
	if old, ok := p.buffers[name]; ok {
		old.handle.Release()
		p.stats.BufferBytes -= old.desc.Size
	}
	p.buffers[name] = trackedBuffer{desc, buffer}
	p.stats.BufferBytes += desc.Size
	p.stats.BufferAllocations++
	return buffer, nil
}
func (h *GPUResources) WriteBuffer(name string, offset uint64, data []byte) error {
	p, err := h.active()
	if err != nil {
		return err
	}
	b, ok := p.buffers[name]
	if !ok {
		return errors.New("GPU view: unknown tracked buffer")
	}
	if err := gpuWriteBounds(b.desc, offset, len(data)); err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	if err := p.device.Queue.TryWriteBuffer(b.handle, offset, data); err != nil {
		return err
	}
	p.stats.HelperUploads++
	p.stats.HelperUploadBytes += uint64(len(data))
	return nil
}
func gpuWriteBounds(desc wgpu.BufferDescriptor, offset uint64, size int) error {
	if size < 0 || offset%4 != 0 || size%4 != 0 || desc.Usage&wgpu.BufferUsageCopyDst == 0 || offset > desc.Size || uint64(size) > desc.Size-offset {
		return errors.New("GPU view: invalid/alignment/out-of-bounds buffer upload")
	}
	return nil
}
func gpuTextureBytes(desc GPUTexture, limits wgpu.Limits) (uint64, error) {
	if desc.Width == 0 || desc.Height == 0 || desc.Width > limits.MaxTextureDimension2D || desc.Height > limits.MaxTextureDimension2D || desc.Usage == 0 || (desc.Samples != 1 && desc.Samples != 4) {
		return 0, errors.New("GPU view: invalid/unsupported texture dimensions or samples")
	}
	bpp := uint64(4)
	switch desc.Format {
	case wgpu.TextureFormatRGBA8Unorm, wgpu.TextureFormatRGBA8UnormSrgb, wgpu.TextureFormatBGRA8Unorm, wgpu.TextureFormatBGRA8UnormSrgb, wgpu.TextureFormatDepth32Float:
	case wgpu.TextureFormatRGBA16Float:
		bpp = 8
	default:
		return 0, errors.New("GPU view: texture format is outside tracked contract")
	}
	if desc.Samples == 4 && (desc.Usage&wgpu.TextureUsageRenderAttachment == 0 || desc.Usage&(wgpu.TextureUsageCopySrc|wgpu.TextureUsageCopyDst|wgpu.TextureUsageStorageBinding) != 0) {
		return 0, errors.New("GPU view: invalid multisample texture usage")
	}
	// The native maximum is far below uint32 max, but perform checked arithmetic
	// even when a caller supplies artificial limits during deterministic tests.
	result := uint64(desc.Width)
	for _, factor := range []uint64{uint64(desc.Height), bpp, uint64(desc.Samples)} {
		if result > ^uint64(0)/factor {
			return 0, errors.New("GPU view: texture byte count overflow")
		}
		result *= factor
	}
	return result, nil
}
func (h *GPUResources) Texture(name string, desc GPUTexture) (*wgpu.Texture, error) {
	p, err := h.active()
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, errors.New("GPU view: texture name required")
	}
	bytes, err := gpuTextureBytes(desc, p.limits)
	if err != nil {
		return nil, err
	}
	if old, ok := p.textures[name]; ok && old.desc == desc {
		return old.handle, nil
	}
	if err := p.admit(bytes); err != nil {
		return nil, err
	}
	tex, err := p.device.Device.TryCreateTexture(&wgpu.TextureDescriptor{Label: name, Usage: desc.Usage, Dimension: wgpu.TextureDimension2D, Size: wgpu.Extent3D{Width: desc.Width, Height: desc.Height, DepthOrArrayLayers: 1}, Format: desc.Format, MipLevelCount: 1, SampleCount: desc.Samples})
	if err != nil {
		return nil, err
	}
	if p.textures == nil {
		p.textures = make(map[string]trackedTexture)
	}
	if old, ok := p.textures[name]; ok {
		old.handle.Release()
		p.stats.TextureBytes -= old.bytes
	}
	p.textures[name] = trackedTexture{desc, tex, bytes}
	p.stats.TextureBytes += bytes
	p.stats.TextureAllocations++
	return tex, nil
}
func (h *GPUResources) Drop(name string) error {
	p, err := h.active()
	if err != nil {
		return err
	}
	if b, ok := p.buffers[name]; ok {
		b.handle.Release()
		delete(p.buffers, name)
		p.stats.BufferBytes -= b.desc.Size
	}
	if t, ok := p.textures[name]; ok {
		t.handle.Release()
		delete(p.textures, name)
		p.stats.TextureBytes -= t.bytes
	}
	return nil
}
func (p *gpuResourcePool) release() {
	p.lease.Store(0)
	for _, b := range p.buffers {
		b.handle.Release()
	}
	for _, t := range p.textures {
		t.handle.Release()
	}
	p.buffers = nil
	p.textures = nil
	if p.stats != nil {
		p.stats.BufferBytes = 0
		p.stats.TextureBytes = 0
	}
	p.device = nil
}
