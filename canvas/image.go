package canvas

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"
)

const ImageTileSize = 128
const MaxImageDimension = 8192

// TiledImage is immutable straight-alpha sRGB RGBA8 data. Patches share unchanged
// tiles without retaining a version chain. It is safe to share between workers
// and UI owners. Source construction belongs on a worker for large images.
type TiledImage struct {
	size    image.Point
	columns int
	tiles   []*imageTileData
}
type imageTileData struct{ pixels []byte }

// ImageTile is an immutable upload-sized tile. Bounds are in normalized source
// image coordinates; Pixels returns a copy with stride Bounds().Dx()*4.
type ImageTile struct {
	bounds image.Rectangle
	data   *imageTileData
}

func (t ImageTile) Bounds() image.Rectangle { return t.bounds }
func (t ImageTile) Pixels() []byte {
	if t.data == nil {
		return nil
	}
	return slices.Clone(t.data.pixels)
}

// NewTiledImage copies a valid NRGBA image, including nonzero origins/subimages.
// Dimensions are 1..8192; retained pixel data is at most 256 MiB.
func NewTiledImage(source *image.NRGBA) (*TiledImage, error) {
	if err := validNRGBA(source); err != nil {
		return nil, err
	}
	size := source.Rect.Size()
	out := &TiledImage{size: size, columns: (size.X + ImageTileSize - 1) / ImageTileSize}
	rows := (size.Y + ImageTileSize - 1) / ImageTileSize
	out.tiles = make([]*imageTileData, out.columns*rows)
	for i := range out.tiles {
		bounds := out.tileBounds(i)
		pixels := make([]byte, bounds.Dx()*bounds.Dy()*4)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			start := y*source.Stride + bounds.Min.X*4
			dst := (y - bounds.Min.Y) * bounds.Dx() * 4
			copy(pixels[dst:dst+bounds.Dx()*4], source.Pix[start:start+bounds.Dx()*4])
		}
		out.tiles[i] = &imageTileData{pixels}
	}
	return out, nil
}
func validNRGBA(source *image.NRGBA) error {
	if source == nil {
		return fmt.Errorf("canvas: image source required")
	}
	if source.Rect.Empty() {
		return fmt.Errorf("canvas: empty or reversed image rectangle")
	}
	size := source.Rect.Size()
	if size.X < 1 || size.Y < 1 || size.X > MaxImageDimension || size.Y > MaxImageDimension {
		return fmt.Errorf("canvas: image dimensions outside 1..8192")
	}
	row := size.X * 4
	// Division avoids overflow for caller-constructed Stride/Rect headers.
	if source.Stride < row || len(source.Pix) < row || (size.Y > 1 && source.Stride > (len(source.Pix)-row)/(size.Y-1)) {
		return fmt.Errorf("canvas: malformed NRGBA pixel layout")
	}
	return nil
}
func (m *TiledImage) Size() image.Point {
	if m == nil {
		return image.Point{}
	}
	return m.size
}
func (m *TiledImage) Bounds() image.Rectangle { return image.Rectangle{Max: m.Size()} }
func (*TiledImage) ColorModel() color.Model   { return color.NRGBAModel }
func (m *TiledImage) At(x, y int) color.Color { return m.NRGBAAt(x, y) }
func (m *TiledImage) NRGBAAt(x, y int) color.NRGBA {
	if m == nil || !image.Pt(x, y).In(m.Bounds()) {
		return color.NRGBA{}
	}
	index := (y/ImageTileSize)*m.columns + x/ImageTileSize
	bounds := m.tileBounds(index)
	offset := ((y-bounds.Min.Y)*bounds.Dx() + x - bounds.Min.X) * 4
	pixels := m.tiles[index].pixels
	return color.NRGBA{R: pixels[offset], G: pixels[offset+1], B: pixels[offset+2], A: pixels[offset+3]}
}
func (m *TiledImage) TileCount() int {
	if m == nil {
		return 0
	}
	return len(m.tiles)
}
func (m *TiledImage) tileBounds(index int) image.Rectangle {
	min := image.Pt(index%m.columns*ImageTileSize, index/m.columns*ImageTileSize)
	return image.Rectangle{Min: min, Max: image.Pt(min.X+ImageTileSize, min.Y+ImageTileSize)}.Intersect(m.Bounds())
}

// WithPatch copies source into a nonempty, fully contained normalized rectangle.
// Source dimensions must match. Only intersecting tiles are cloned; identical
// pixels retain their tile identity, and a completely unchanged patch returns m.
func (m *TiledImage) WithPatch(rect image.Rectangle, source *image.NRGBA) (*TiledImage, error) {
	if m == nil || rect.Empty() || !rect.In(m.Bounds()) {
		return nil, fmt.Errorf("canvas: image patch outside bounds")
	}
	if err := validNRGBA(source); err != nil {
		return nil, err
	}
	if source.Rect.Size() != rect.Size() {
		return nil, fmt.Errorf("canvas: patch dimensions do not match")
	}
	tiles := slices.Clone(m.tiles)
	changed := false
	for y := rect.Min.Y / ImageTileSize; y <= (rect.Max.Y-1)/ImageTileSize; y++ {
		for x := rect.Min.X / ImageTileSize; x <= (rect.Max.X-1)/ImageTileSize; x++ {
			index := y*m.columns + x
			bounds := m.tileBounds(index)
			overlap := rect.Intersect(bounds)
			pixels := slices.Clone(tiles[index].pixels)
			for py := overlap.Min.Y; py < overlap.Max.Y; py++ {
				src := (py-rect.Min.Y)*source.Stride + (overlap.Min.X-rect.Min.X)*4
				dst := ((py-bounds.Min.Y)*bounds.Dx() + overlap.Min.X - bounds.Min.X) * 4
				count := overlap.Dx() * 4
				copy(pixels[dst:dst+count], source.Pix[src:src+count])
			}
			if !bytes.Equal(pixels, tiles[index].pixels) {
				tiles[index] = &imageTileData{pixels}
				changed = true
			}
		}
	}
	if !changed {
		return m, nil
	}
	return &TiledImage{size: m.size, columns: m.columns, tiles: tiles}, nil
}

// ChangedTiles compares tile identities in O(tile count), without scanning pixel
// data. A nil/different-sized prior image requires every tile; unchanged versions
// return no work. The returned tile list and pixel copies cannot mutate m.
func (m *TiledImage) ChangedTiles(previous *TiledImage) []ImageTile {
	if m == nil || m == previous {
		return nil
	}
	compatible := previous != nil && previous.size == m.size
	var changed []ImageTile
	for i, tile := range m.tiles {
		if compatible && previous.tiles[i] == tile {
			continue
		}
		changed = append(changed, ImageTile{bounds: m.tileBounds(i), data: tile})
	}
	return changed
}

type ImageFilter uint8

const (
	ImageNearest ImageFilter = iota
	ImageLinear
)

// ImageLayer places a TiledImage in world coordinates below MeshCanvas geometry.
// Image data is immutable; placement fields belong to the application UI owner.
// Linear filtering interpolates premultiplied linear colors, avoiding alpha halos.
type ImageLayer struct {
	Image  *TiledImage
	Bounds Bounds
	Filter ImageFilter
}

func (l ImageLayer) Validate() error {
	if l.Image == nil || l.Image.TileCount() == 0 {
		return fmt.Errorf("canvas: image layer source required")
	}
	if err := l.Bounds.Validate(); err != nil {
		return err
	}
	for _, v := range []float64{l.Bounds.Min.X, l.Bounds.Min.Y, l.Bounds.Max.X, l.Bounds.Max.Y} {
		if math.Abs(v) > 1e6 {
			return fmt.Errorf("canvas: image bounds outside coordinate range")
		}
	}
	if l.Filter > ImageLinear {
		return fmt.Errorf("canvas: unsupported image filter")
	}
	return nil
}
