package ui

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/dshills/vellum/canvas"
)

// ImageFit changes source placement within the image's allocated content box.
type ImageFit string

const (
	ImageContain         ImageFit = "contain"
	ImageCover           ImageFit = "cover"
	ImageStretch         ImageFit = "stretch"
	MaxEncodedImageBytes          = 32 << 20
)

func validateImage(n Node) error {
	if n.Kind != Image {
		if n.Image != nil || n.Fit != "" {
			return fmt.Errorf("ui: Image/Fit belong only on image nodes")
		}
		return nil
	}
	if n.Image == nil || n.Image.TileCount() == 0 || n.Activate != nil || n.Change != nil {
		return fmt.Errorf("ui: image requires an immutable nonempty source and no input callbacks")
	}
	switch n.Fit {
	case "", ImageContain, ImageCover, ImageStretch:
	default:
		return fmt.Errorf("ui: invalid image fit %q", n.Fit)
	}
	return nil
}

// DecodeImage reads bounded PNG/JPEG data and returns detached immutable pixels.
// The caller owns the reader and its cancellation. Large decoding belongs on a
// worker; decoded/conversion temporaries are outside the retained pixel budget.
func DecodeImage(reader io.Reader) (*canvas.TiledImage, error) {
	if reader == nil {
		return nil, fmt.Errorf("ui: image reader required")
	}
	data, err := io.ReadAll(io.LimitReader(reader, MaxEncodedImageBytes+1))
	if err != nil {
		return nil, fmt.Errorf("ui: read image: %w", err)
	}
	if len(data) > MaxEncodedImageBytes {
		return nil, fmt.Errorf("ui: encoded image exceeds 32 MiB")
	}
	var config image.Config
	var decode func(io.Reader) (image.Image, error)
	switch {
	case bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n")):
		config, err = png.DecodeConfig(bytes.NewReader(data))
		decode = png.Decode
	case bytes.HasPrefix(data, []byte{0xff, 0xd8}):
		config, err = jpeg.DecodeConfig(bytes.NewReader(data))
		decode = jpeg.Decode
	default:
		return nil, fmt.Errorf("ui: image format must be PNG or JPEG")
	}
	if err != nil {
		return nil, fmt.Errorf("ui: image header: %w", err)
	}
	if config.Width < 1 || config.Height < 1 || config.Width > canvas.MaxImageDimension || config.Height > canvas.MaxImageDimension {
		return nil, fmt.Errorf("ui: image dimensions outside 1..8192")
	}
	source, err := decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("ui: decode image: %w", err)
	}
	if source.Bounds().Size() != image.Pt(config.Width, config.Height) {
		return nil, fmt.Errorf("ui: decoded image dimensions changed")
	}
	if pixels, ok := source.(*image.NRGBA); ok {
		return canvas.NewTiledImage(pixels)
	}
	pixels := image.NewNRGBA(image.Rect(0, 0, config.Width, config.Height))
	draw.Draw(pixels, pixels.Bounds(), source, source.Bounds().Min, draw.Src)
	return canvas.NewTiledImage(pixels)
}
