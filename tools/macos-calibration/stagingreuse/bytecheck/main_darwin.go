//go:build darwin && !offscreen

package main

import (
	"bytes"
	"cogentcore.org/core/gpu"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/oliverbestmann/webgpu/wgpu"
	"image"
	"os"
)

type spec struct {
	name                string
	width, height, seed int
	bgra                bool
	stride, offset      int
}
type item struct {
	tx                        *wgpu.Texture
	read                      *wgpu.Buffer
	dev                       *gpu.Device
	width, height, pitch, bpp int
	want                      []byte
	row                       map[string]any
}

func (it *item) release() {
	if it.read != nil {
		it.read.Release()
	}
	it.tx.Release()
}
func hash(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }

func write(dev *gpu.Device, s spec, ordinal int) (*item, error) {
	format := gpu.TextureFormat{Size: image.Pt(s.width, s.height), Layers: 1, Samples: 1, Format: wgpu.TextureFormatRGBA8UnormSrgb}
	if s.bgra {
		format.Format = wgpu.TextureFormatBGRA8UnormSrgb
	}
	tx, err := dev.Device.TryCreateTexture(&wgpu.TextureDescriptor{Label: s.name, Size: format.Extent3D(), MipLevelCount: 1, SampleCount: 1, Dimension: wgpu.TextureDimension2D, Format: format.Format, Usage: wgpu.TextureUsageCopySrc | wgpu.TextureUsageCopyDst | wgpu.TextureUsageTextureBinding})
	if err != nil {
		return nil, err
	}
	want := make([]byte, s.width*s.height*4)
	for i := 0; i < len(want); i += 4 {
		want[i] = byte(i/4 + s.seed)
		want[i+1] = byte(i/(s.width*4) + s.seed*7)
		want[i+2] = byte(i/4*17 + s.seed*13)
		want[i+3] = 255
		if s.bgra {
			want[i], want[i+2] = want[i+2], want[i]
		}
	}
	stride := s.stride
	if stride == 0 {
		stride = s.width * 4
	}
	source := bytes.Repeat([]byte{0xa5}, s.offset+stride*s.height)
	for y := 0; y < s.height; y++ {
		copy(source[s.offset+y*stride:s.offset+y*stride+s.width*4], want[y*s.width*4:(y+1)*s.width*4])
	}
	size := format.Extent3D()
	if err := dev.Queue.TryWriteTexture(tx.AsImageCopy(), source, &wgpu.TexelCopyBufferLayout{Offset: uint64(s.offset), BytesPerRow: uint32(stride), RowsPerImage: uint32(s.height)}, &size); err != nil {
		tx.Release()
		return nil, err
	}
	return &item{tx: tx, dev: dev, width: s.width, height: s.height, bpp: 4, pitch: ((s.width*4 + 255) / 256) * 256, want: want, row: map[string]any{"case": s.name, "write": ordinal, "width": s.width, "height": s.height, "bgra": s.bgra, "source_stride": stride, "source_offset": s.offset, "input_bytes": len(source), "staging_bytes": len(want), "bytes_per_pixel": 4, "expected_sha256": hash(want)}}, nil
}
func copyAndSubmit(dev *gpu.Device, items []*item) error {
	cmd, err := dev.Device.TryCreateCommandEncoder(nil)
	if err != nil {
		return err
	}
	defer cmd.Release()
	for _, it := range items {
		it.read, err = dev.Device.TryCreateBuffer(&wgpu.BufferDescriptor{Label: "bytecheck readback", Size: uint64(it.pitch * it.height), Usage: wgpu.BufferUsageMapRead | wgpu.BufferUsageCopyDst})
		if err != nil {
			return err
		}
		size := wgpu.Extent3D{Width: uint32(it.width), Height: uint32(it.height), DepthOrArrayLayers: 1}
		if err := cmd.TryCopyTextureToBuffer(it.tx.AsImageCopy(), &wgpu.TexelCopyBufferInfo{Buffer: it.read, Layout: wgpu.TexelCopyBufferLayout{BytesPerRow: uint32(it.pitch), RowsPerImage: uint32(it.height)}}, &size); err != nil {
			return err
		}
	}
	buffer, err := cmd.TryFinish(nil)
	if err != nil {
		return err
	}
	defer buffer.Release()
	dev.Queue.Submit(buffer)
	return nil
}
func verify(it *item) error {
	if err := gpu.BufferReadSync(it.dev, it.pitch*it.height, it.read); err != nil {
		return err
	}
	mapped := it.read.GetMappedRange(0, uint(it.pitch*it.height))
	actual := make([]byte, len(it.want))
	for y := 0; y < it.height; y++ {
		copy(actual[y*it.width*it.bpp:(y+1)*it.width*it.bpp], mapped[y*it.pitch:y*it.pitch+it.width*it.bpp])
	}
	err := it.read.TryUnmap()
	it.row["readback_bytes"] = len(actual)
	it.row["actual_sha256"] = hash(actual)
	it.row["full_byte_match"] = bytes.Equal(actual, it.want)
	if err != nil {
		return err
	}
	if !bytes.Equal(actual, it.want) {
		return errors.New("full WebGPU readback mismatch")
	}
	return nil
}
func run(out map[string]any) (err error) {
	gp, dev, err := gpu.NoDisplayGPU()
	defer func() {
		if dev != nil {
			dev.Release()
		}
		if gp != nil {
			gp.Release()
		}
		gpu.ReleaseInstance()
	}()
	if err != nil {
		return err
	}
	if gp == nil || dev == nil {
		return errors.New("incomplete headless GPU pair")
	}
	out["backend"] = int(gp.Properties.BackendType)
	out["device"] = gp.DeviceName
	if gp.Properties.BackendType != wgpu.BackendTypeMetal {
		return errors.New("Metal backend required")
	}
	rows := []map[string]any{}
	defer func() { out["writes"] = rows }()
	ordinal := 0
	// Zero-area success must not acquire staging. It creates no GPU copy/readback.
	zeroFormat := gpu.TextureFormat{Size: image.Pt(3, 2), Layers: 1, Samples: 1, Format: wgpu.TextureFormatRGBA8UnormSrgb}
	zero, e := dev.Device.TryCreateTexture(&wgpu.TextureDescriptor{Size: zeroFormat.Extent3D(), MipLevelCount: 1, SampleCount: 1, Dimension: wgpu.TextureDimension2D, Format: zeroFormat.Format, Usage: wgpu.TextureUsageCopySrc | wgpu.TextureUsageCopyDst | wgpu.TextureUsageTextureBinding})
	if e != nil {
		return e
	}
	extent := zeroFormat.Extent3D()
	extent.Width = 0
	ordinal++
	e = dev.Queue.TryWriteTexture(zero.AsImageCopy(), make([]byte, 24), &wgpu.TexelCopyBufferLayout{BytesPerRow: 12, RowsPerImage: 2}, &extent)
	zero.Release()
	if e != nil {
		return e
	}
	rows = append(rows, map[string]any{"case": "zero-area", "write": ordinal, "width": 0, "height": 2, "input_bytes": 24, "staging_bytes": 0, "no_transfer": true})
	sequential := []spec{{name: "small-a", width: 3, height: 2, seed: 1}, {name: "small-b", width: 3, height: 2, seed: 2}, {name: "bgra", width: 3, height: 2, seed: 3, bgra: true}, {name: "unaligned-row", width: 7, height: 5, seed: 4}, {name: "padded-offset-source", width: 7, height: 5, seed: 5, stride: 40, offset: 16}, {name: "large-a", width: 1519, height: 1266, seed: 6}, {name: "large-b", width: 1519, height: 1266, seed: 7}, {name: "above-idle-cap", width: 4096, height: 2049, seed: 8}}
	single := func(device *gpu.Device, s spec) error {
		ordinal++
		it, e := write(device, s, ordinal)
		if e != nil {
			return e
		}
		defer it.release()
		rows = append(rows, it.row)
		if e = copyAndSubmit(device, []*item{it}); e != nil {
			return e
		}
		return verify(it)
	}
	for _, s := range sequential {
		if e := single(dev, s); e != nil {
			return fmt.Errorf("%s: %w", s.name, e)
		}
	}
	// Three same-sized leases coexist in pending writes before one GPU submission.
	group := []*item{}
	groupErr := func() error {
		defer func() {
			for _, it := range group {
				it.release()
			}
		}()
		for i := 0; i < 3; i++ {
			ordinal++
			it, e := write(dev, spec{name: fmt.Sprintf("queued-%d", i+1), width: 3, height: 2, seed: 20 + i}, ordinal)
			if e != nil {
				return e
			}
			group = append(group, it)
			rows = append(rows, it.row)
		}
		if e := copyAndSubmit(dev, group); e != nil {
			return e
		}
		for _, it := range group {
			if e := verify(it); e != nil {
				return e
			}
		}
		return nil
	}()
	if groupErr != nil {
		return groupErr
	}
	for i := 0; i < 9; i++ {
		if e := single(dev, spec{name: fmt.Sprintf("size-cycle-%d", i+1), width: 2 + i, height: 3, seed: 40 + i}); e != nil {
			return e
		}
	}
	other, e := gpu.NewDevice(gp)
	if e != nil {
		return e
	}
	e = single(other, spec{name: "other-device", width: 3, height: 2, seed: 60})
	other.Release()
	if e != nil {
		return e
	}

	// A non-allowlisted one-byte format must stay fresh and outside the reuse ledger.
	ordinal++
	excludedErr := func() error {
		size := wgpu.Extent3D{Width: 3, Height: 2, DepthOrArrayLayers: 1}
		tx, e := dev.Device.TryCreateTexture(&wgpu.TextureDescriptor{Size: size, MipLevelCount: 1, SampleCount: 1, Dimension: wgpu.TextureDimension2D, Format: wgpu.TextureFormatR8Unorm, Usage: wgpu.TextureUsageCopySrc | wgpu.TextureUsageCopyDst | wgpu.TextureUsageTextureBinding})
		if e != nil {
			return e
		}
		want := []byte{7, 29, 53, 101, 173, 239}
		it := &item{tx: tx, dev: dev, width: 3, height: 2, pitch: 256, bpp: 1, want: want, row: map[string]any{"case": "excluded-r8", "write": ordinal, "width": 3, "height": 2, "source_stride": 3, "source_offset": 0, "input_bytes": 6, "staging_bytes": 8, "bytes_per_pixel": 1, "expected_sha256": hash(want)}}
		defer it.release()
		if e := dev.Queue.TryWriteTexture(tx.AsImageCopy(), want, &wgpu.TexelCopyBufferLayout{BytesPerRow: 3, RowsPerImage: 2}, &size); e != nil {
			return e
		}
		rows = append(rows, it.row)
		if e := copyAndSubmit(dev, []*item{it}); e != nil {
			return e
		}
		return verify(it)
	}()
	if excludedErr != nil {
		return excludedErr
	}
	// Discard a pending upload at device/queue release. It must never enter cache
	// merely because its ordinary staging destructor runs.
	ordinal++
	unsent, e := write(dev, spec{name: "unsubmitted-release", width: 3, height: 2, seed: 70}, ordinal)
	if e != nil {
		return e
	}
	unsent.row["unsubmitted"] = true
	rows = append(rows, unsent.row)
	unsent.release()
	out["windows_created"] = 0
	out["surfaces_created"] = 0
	out["physical_gpu_overlap_forced"] = false
	return nil
}
func main() {
	out := map[string]any{"kind": "staging-reuse-byte-check", "release_qualified": false, "runtime_returned": false}
	var err error
	if len(os.Args) != 2 || (os.Args[1] != "off" && os.Args[1] != "on") {
		err = errors.New("off|on required")
	} else {
		out["mode"] = os.Args[1]
		err = reuseConfigure(os.Args[1] == "on")
		if err == nil {
			err = probeConfigure(true)
		}
		if err == nil {
			err = run(out)
		}
	}
	cpu, cpuErr := probeSnapshot()
	pool, poolErr := reuseSnapshot()
	err = errors.Join(err, cpuErr, poolErr)
	out["native_upload_trace"] = cpu
	out["reuse_trace"] = pool
	out["error"] = fmt.Sprint(err)
	out["runtime_returned"] = true
	if e := json.NewEncoder(os.Stdout).Encode(out); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(2)
	}
	if err != nil {
		os.Exit(1)
	}
}
