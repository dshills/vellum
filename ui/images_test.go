package ui

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/dshills/vellum/canvas"
)

func imageSource(t *testing.T) *canvas.TiledImage {
	t.Helper()
	source, err := canvas.NewTiledImage(image.NewNRGBA(image.Rect(0, 0, 8, 4)))
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestImagePortableValidationAndFocus(t *testing.T) {
	source := imageSource(t)
	for _, fit := range []ImageFit{"", ImageContain, ImageCover, ImageStretch} {
		if err := Validate(Node{ID: "image", Kind: Image, Image: source, Fit: fit}); err != nil {
			t.Fatal(fit, err)
		}
	}
	n := Node{ID: "root", Kind: Column, Children: []Node{
		{ID: "photo", Kind: Image, Image: source, Text: "Preview"},
		{ID: "save", Kind: Button, Text: "Save"},
	}}
	if got := FocusOrder(n); !reflect.DeepEqual(got, []string{"save"}) {
		t.Fatal("image added a focus stop", got)
	}
	for _, n := range []Node{
		{ID: "bad", Kind: Image},
		{ID: "bad", Kind: Image, Image: &canvas.TiledImage{}},
		{ID: "bad", Kind: Image, Image: source, Fit: "unknown"},
		{ID: "bad", Kind: Image, Image: source, Activate: func() {}},
		{ID: "bad", Kind: Image, Image: source, Change: func(ChangeEvent) {}},
		{ID: "bad", Kind: Image, Image: source, Children: []Node{{ID: "child", Kind: Label}}},
		{ID: "bad", Kind: Label, Image: source},
		{ID: "bad", Kind: Label, Fit: ImageContain},
	} {
		if err := Validate(n); err == nil {
			t.Fatal("invalid image descriptor accepted", n)
		}
	}
}

func TestDecodeImagePNGJPEGAndImmutableOwnership(t *testing.T) {
	input := image.NewNRGBA(image.Rect(2, 3, 10, 7))
	input.SetNRGBA(5, 4, color.NRGBA{R: 200, G: 80, B: 40, A: 128})
	for _, format := range []string{"png", "jpeg"} {
		t.Run(format, func(t *testing.T) {
			var encoded bytes.Buffer
			var err error
			if format == "png" {
				err = png.Encode(&encoded, input)
			} else {
				err = jpeg.Encode(&encoded, input, nil)
			}
			if err != nil {
				t.Fatal(err)
			}
			data := append([]byte(nil), encoded.Bytes()...)
			asset, err := DecodeImage(bytes.NewReader(data))
			if err != nil || asset.Size() != image.Pt(8, 4) {
				t.Fatal("decoded normalized bounds", asset, err)
			}
			before := asset.NRGBAAt(3, 1)
			clear(data)
			input.SetNRGBA(5, 4, color.NRGBA{A: 255})
			if asset.NRGBAAt(3, 1) != before || (format == "png" && before != (color.NRGBA{R: 200, G: 80, B: 40, A: 128})) {
				t.Fatal("source ownership or straight-alpha normalization", before)
			}
		})
	}
}

type countedZeroReader struct{ read int }

func (r *countedZeroReader) Read(p []byte) (int, error) {
	clear(p)
	r.read += len(p)
	return len(p), nil
}

type failedImageReader struct{ err error }

func (r failedImageReader) Read([]byte) (int, error) { return 0, r.err }

func TestDecodeImageRefusesEncodingAndCapacity(t *testing.T) {
	for _, reader := range []io.Reader{nil, bytes.NewReader(nil), bytes.NewBufferString("GIF89a"), bytes.NewBufferString("\x89PNG\r\n\x1a\n"), bytes.NewReader([]byte{0xff, 0xd8})} {
		if result, err := DecodeImage(reader); result != nil || err == nil {
			t.Fatal("invalid encoding accepted", result, err)
		}
	}
	sentinel := errors.New("reader stopped")
	if _, err := DecodeImage(failedImageReader{sentinel}); !errors.Is(err, sentinel) {
		t.Fatal("reader error lost", err)
	}
	counted := &countedZeroReader{}
	if result, err := DecodeImage(counted); result != nil || err == nil || counted.read != MaxEncodedImageBytes+1 {
		t.Fatal("encoded admission not bounded", counted.read, err)
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, image.NewNRGBA(image.Rect(0, 0, 8, 4))); err != nil {
		t.Fatal(err)
	}
	data := append([]byte(nil), encoded.Bytes()...)
	binary.BigEndian.PutUint32(data[16:20], canvas.MaxImageDimension+1)
	binary.BigEndian.PutUint32(data[29:33], crc32.ChecksumIEEE(data[12:29]))
	if result, err := DecodeImage(bytes.NewReader(data)); result != nil || err == nil || !strings.Contains(err.Error(), "dimensions outside") {
		t.Fatal("oversized header entered decode", result, err)
	}
}
