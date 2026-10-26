package ui

import (
	"bytes"
	"encoding/base64"
	"github.com/dshills/vellum/canvas"
	"image"
	"image/color"
	"image/png"
	"path/filepath"
	"testing"
)

func TestMediaPlayerPosterAndLocalAdmission(t *testing.T) {
	pixels := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	want := color.NRGBA{R: 210, G: 35, B: 180, A: 200}
	pixels.SetNRGBA(0, 0, want)
	poster, _ := canvas.NewTiledImage(pixels)
	p := MediaPlayer{ID: "media", Label: "Clip", Path: filepath.Join(t.TempDir(), "clip.mov"), PosterImage: poster, IconImage: poster, InitialMuted: true}
	n, e := p.View()
	if e != nil {
		t.Fatal(e)
	}
	if n.Children[0].Kind != Row || n.Children[0].Disabled || n.Children[0].Children[0].Image != poster {
		t.Fatal("title should retain full-color decorative art without a disabled button")
	}
	v := n.Children[1].NativeContent
	if v.Type != "media" || !v.InitialMuted || v.PosterPNG == "" {
		t.Fatal(v)
	}
	data, _ := base64.StdEncoding.DecodeString(v.PosterPNG)
	decoded, e := png.Decode(bytes.NewReader(data))
	if e != nil || color.NRGBAModel.Convert(decoded.At(0, 0)) != want {
		t.Fatal("poster colors", e)
	}
	again, _ := p.View()
	if again.Children[1].NativeContent.PosterPNG != v.PosterPNG {
		t.Fatal("poster cache changed")
	}
	patch := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	patch.SetNRGBA(0, 0, color.NRGBA{G: 200, A: 255})
	replacement, e := poster.WithPatch(image.Rect(0, 0, 1, 1), patch)
	if e != nil {
		t.Fatal(e)
	}
	p.PosterImage = replacement
	changed, e := p.View()
	if e != nil || changed.Children[1].NativeContent.PosterPNG == v.PosterPNG || poster.NRGBAAt(0, 0) != want {
		t.Fatal("immutable patch did not refresh the poster", e)
	}
	p.PosterImage = nil
	n, e = p.View()
	if e != nil || n.Children[1].NativeContent.PosterPNG != "" {
		t.Fatal("poster not cleared", e)
	}
	v.PosterPNG = "broken"
	if Validate(Node{ID: "forged", Kind: NativeContent, NativeContent: v}) == nil {
		t.Fatal("invalid poster accepted")
	}
	huge, _ := canvas.NewTiledImage(image.NewNRGBA(image.Rect(0, 0, 257, 1)))
	p.PosterImage = huge
	if _, e = p.View(); e == nil {
		t.Fatal("oversize poster accepted")
	}
	p.PosterImage = nil
	for _, path := range []string{"https://example.org/clip.mp4", "relative.mov", "/tmp/a\x00.mov"} {
		p.Path = path
		if _, e = p.View(); e == nil {
			t.Fatal(path)
		}
	}
	p.Path = ""
	n, e = p.View()
	if e != nil || n.Children[1].Kind != Label {
		t.Fatal(e)
	}
}
