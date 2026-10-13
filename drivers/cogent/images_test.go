package cogent

import (
	"image"
	"image/color"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
)

func managedImageSource(t *testing.T) *canvas.TiledImage {
	t.Helper()
	pixels := image.NewNRGBA(image.Rect(0, 0, 12, 6))
	for y := range 6 {
		for x := range 12 {
			value := color.NRGBA{R: 240, A: 255}
			if x >= 4 && x < 8 {
				value = color.NRGBA{G: 240, A: 255}
			}
			if x >= 8 {
				value = color.NRGBA{B: 240, A: 255}
			}
			pixels.SetNRGBA(x, y, value)
		}
	}
	source, err := canvas.NewTiledImage(pixels)
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestManagedImageRetainedSourceFitSemanticsAndRetirement(t *testing.T) {
	a, b := newAXTest(t, 8)
	host := core.NewFrame(core.NewBody("images"))
	source := managedImageSource(t)
	n := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{
		{ID: "decorative", Kind: ui.Image, Image: source},
		{ID: "photo", Kind: ui.Image, Image: source, Text: "<Preview>", Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 80, Height: 80})}},
	}}
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	w := findRecord(r, "photo").widget.(*managedImage)
	w.Style()
	w.SizeUp()
	if w.source != source || w.Geom.Size.Actual.Content != math32.Vec2(80, 80) {
		t.Fatal("image layout box", w.Geom.Size)
	}
	patch := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	patch.SetNRGBA(0, 0, color.NRGBA{A: 255})
	replacement, err := source.WithPatch(image.Rect(0, 0, 1, 1), patch)
	if err != nil {
		t.Fatal(err)
	}
	n.Children[1].Image, n.Children[1].Fit = replacement, ui.ImageCover
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	w.Style()
	w.SizeUp()
	if findRecord(r, "photo").widget != w || w.source != replacement || w.fit != ui.ImageCover || w.Geom.Size.Actual.Content != math32.Vec2(80, 80) {
		t.Fatal("fit/source replacement lost box or identity")
	}
	n.Children[1].Style = ui.Style{MinSize: ui.Some(ui.Size{Width: 60})}
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	w.Style()
	w.SizeUp()
	if w.Geom.Size.Actual.Content != math32.Vec2(60, 30) {
		t.Fatal("one-axis aspect sizing", w.Geom.Size)
	}
	n.Children[1].Style = ui.Style{}
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	w.Style()
	w.SizeUp()
	if w.Geom.Size.Actual.Content != math32.Vec2(12, 6) {
		t.Fatal("natural dimensions not restored", w.Geom.Size)
	}
	nodes, handlers := accessibilitySnapshot(r, b)
	found := false
	for _, node := range nodes {
		if node.DomainID == "decorative" {
			t.Fatal("decorative image became accessible")
		}
		if node.DomainID == "photo" {
			found = true
			if node.Role != "image" || node.Name != "<Preview>" || node.Focusable || len(node.Actions) != 0 || handlers[node.ID] != nil || node.Value != nil {
				t.Fatal("passive image semantics", node)
			}
		}
	}
	if !found {
		t.Fatal("named image missing")
	}
	n.Children = nil
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	if w.source != nil || replacement.NRGBAAt(5, 2).G != 240 {
		t.Fatal("native reference retirement or caller ownership")
	}
}

func TestManagedImageMaskClippingAndComposition(t *testing.T) {
	parent := image.NewAlpha(image.Rect(0, 0, 20, 20))
	parent.SetAlpha(4, 5, color.Alpha{A: 128})
	mask := imageViewportMask{bounds: image.Rect(2, 2, 8, 8), parent: parent, parentP: image.Pt(1, 1), alpha: 32768}
	_, _, _, a := mask.At(3, 4).RGBA()
	_, _, _, outside := mask.At(1, 4).RGBA()
	if a < 16400 || a > 16460 || outside != 0 {
		t.Fatal("clip/ancestor alpha multiplication", a, outside)
	}
}

func TestManagedImagePreferredCapAndMaximum(t *testing.T) {
	a, _ := newAXTest(t, 8)
	source, err := canvas.NewTiledImage(image.NewNRGBA(image.Rect(0, 0, 512, 256)))
	if err != nil {
		t.Fatal(err)
	}
	host := core.NewFrame(core.NewBody("preferred image size"))
	n := ui.Node{ID: "photo", Kind: ui.Image, Image: source}
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	w := r.widget.(*managedImage)
	w.Style()
	w.SizeUp()
	if w.Geom.Size.Actual.Content != math32.Vec2(256, 128) {
		t.Fatal("preferred image long-edge cap", w.Geom.Size)
	}
	n.Style.MaxSize = ui.Some(ui.Size{Width: 40, Height: 20})
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	w.Style()
	w.SizeUp()
	if w.Geom.Size.Actual.Content != math32.Vec2(40, 20) {
		t.Fatal("preferred image ignored maximum", w.Geom.Size)
	}
}
