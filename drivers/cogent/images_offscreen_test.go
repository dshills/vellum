//go:build offscreen

package cogent

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
)

func imageManualElement(t *testing.T, m *Manual, id string) (ManualWindowSnapshot, ManualElement) {
	t.Helper()
	s, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range s.Windows[0].Elements {
		if e.ID == id {
			return s.Windows[0], e
		}
	}
	t.Fatal("image target missing", id)
	return ManualWindowSnapshot{}, ManualElement{}
}

func TestManagedImageManualCenteredFitAndCompletedPixels(t *testing.T) {
	for _, fit := range []ui.ImageFit{ui.ImageContain, ui.ImageCover, ui.ImageStretch} {
		t.Run(string(fit), func(t *testing.T) {
			source := managedImageSource(t)
			background := color.NRGBA{R: 8, G: 16, B: 24, A: 255}
			m, _ := simpleManual(t, manualConfig(t, 7), func() ui.Node {
				return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "preview", Kind: ui.Image, Image: source, Fit: fit, Text: "Alternate preview", Accessibility: ui.Accessibility{Name: "Preview"}, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 120, Height: 120}), Background: ui.Some(background)}}}}
			})
			unitSettle(t, m)
			_, target := imageManualElement(t, m, "preview")
			if target.Bounds.Size() != image.Pt(120, 120) || target.Kind != ui.Image || target.Name != "Preview" {
				t.Fatal("fit changed layout/semantics", target)
			}
			frame, err := m.Capture(testContext, "one")
			if err != nil {
				t.Fatal(err)
			}
			at := func(x, y int) color.RGBA { p := target.Bounds.Min.Add(image.Pt(x, y)); return frame.RGBAAt(p.X, p.Y) }
			want := color.RGBA{R: 240, A: 255}
			if fit == ui.ImageContain {
				want = color.RGBA{R: 8, G: 16, B: 24, A: 255}
			}
			if fit == ui.ImageCover {
				want = color.RGBA{G: 240, A: 255}
			}
			if got := at(30, 10); got != want {
				t.Fatal("independent centered fit pixel", fit, got, want)
			}
			for _, point := range []struct {
				x int
				c color.RGBA
			}{{5, color.RGBA{R: 240, A: 255}}, {60, color.RGBA{G: 240, A: 255}}, {115, color.RGBA{B: 240, A: 255}}} {
				if got := at(point.x, 60); got != point.c {
					t.Fatal("source stripes not painted", fit, point.x, got)
				}
			}
			if fit == ui.ImageCover {
				outside := target.Bounds.Min.Add(image.Pt(136, 60))
				if got := frame.RGBAAt(outside.X, outside.Y); got != (color.RGBA{R: 248, G: 248, B: 248, A: 255}) {
					t.Fatal("cover paint escaped image content", got)
				}
			}
			if fit == ui.ImageCover && os.Getenv("VELLUM_IMAGE_CAPTURE") != "" {
				f, err := os.Create(os.Getenv("VELLUM_IMAGE_CAPTURE"))
				if err != nil {
					t.Fatal(err)
				}
				err = png.Encode(f, frame)
				closeErr := f.Close()
				if err != nil || closeErr != nil {
					t.Fatal(err, closeErr)
				}
			}
		})
	}
}

func TestManagedImageManualAlphaReplacementAndIdentity(t *testing.T) {
	pixels := image.NewNRGBA(image.Rect(0, 0, 8, 4))
	for y := range 4 {
		for x := range 8 {
			pixels.SetNRGBA(x, y, color.NRGBA{G: 200, A: 128})
		}
	}
	source, err := canvas.NewTiledImage(pixels)
	if err != nil {
		t.Fatal(err)
	}
	view := func() ui.Node {
		return ui.Node{ID: "preview", Kind: ui.Image, Image: source, Text: "Alpha", Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 80, Height: 40}), Background: ui.Some(color.NRGBA{R: 100, G: 100, B: 100, A: 255})}}
	}
	m, a := simpleManual(t, manualConfig(t, 7), view)
	unitSettle(t, m)
	_, beforeTarget := imageManualElement(t, m, "preview")
	before, err := m.Capture(testContext, "one")
	if err != nil {
		t.Fatal(err)
	}
	point := beforeTarget.Bounds.Min.Add(image.Pt(40, 20))
	c := before.RGBAAt(point.X, point.Y)
	if c.R < 49 || c.R > 50 || c.G < 149 || c.G > 151 || c.B < 49 || c.B > 50 || c.A != 255 {
		t.Fatal("straight-alpha blend", c)
	}
	prior := source
	for y := range 4 {
		for x := range 8 {
			pixels.SetNRGBA(x, y, color.NRGBA{B: 200, A: 128})
		}
	}
	replacement, err := canvas.NewTiledImage(pixels)
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Post(func() { source = replacement; a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	_, afterTarget := imageManualElement(t, m, "preview")
	after, err := m.Capture(testContext, "one")
	if err != nil {
		t.Fatal(err)
	}
	c = after.RGBAAt(point.X, point.Y)
	if beforeTarget.Incarnation != afterTarget.Incarnation || beforeTarget.Bounds != afterTarget.Bounds || c.B < 149 || c.B > 151 || c.G < 49 || c.G > 50 || prior.NRGBAAt(3, 2).G != 200 {
		t.Fatal("completed immutable replacement", c, afterTarget)
	}
}

func TestManagedImageManualScrollClipPreservesSourceMapping(t *testing.T) {
	source := managedImageSource(t)
	m, a := simpleManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "scroll", Kind: ui.Scroll, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 80, Height: 50}), MaxSize: ui.Some(ui.Size{Width: 80, Height: 50})}, Children: []ui.Node{{ID: "preview", Kind: ui.Image, Image: source, Text: "Scrolling preview", Fit: ui.ImageStretch, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 160, Height: 80})}}}}
	})
	unitSettle(t, m)
	window, target := imageManualElement(t, m, "preview")
	before, err := m.Capture(testContext, "one")
	if err != nil {
		t.Fatal(err)
	}
	var w *managedImage
	if err = a.Do(func() { w = findRecord(m.implementation.find("one").state.root, "preview").widget.(*managedImage) }); err != nil {
		t.Fatal(err)
	}
	fullBefore := w.Geom.Pos.Content
	if err = m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: window.Incarnation, Type: events.Scroll, Position: target.Bounds.Min.Add(image.Pt(20, 20)), Delta: math32.Vec2(60, 0)}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	_, afterTarget := imageManualElement(t, m, "preview")
	after, err := m.Capture(testContext, "one")
	if err != nil {
		t.Fatal(err)
	}
	if w.Geom.Pos.Content.X >= fullBefore.X || w.Geom.ContentBBox.Empty() {
		t.Fatal("real wheel did not clip/move image", w.Geom)
	}
	x, y := afterTarget.Bounds.Min.X+20, afterTarget.Bounds.Min.Y+20
	if before.RGBAAt(x, y) == after.RGBAAt(x, y) {
		t.Fatal("clipped source was reanchored instead of scrolling")
	}
	// Image transforms must not draw into the sibling/outer scene beyond the
	// visible clipped content. Sample a point still inside the unscrolled image.
	x, y = afterTarget.Bounds.Max.X+8, afterTarget.Bounds.Min.Y+20
	if c := after.RGBAAt(x, y); c.R == 240 || c.G == 240 || c.B == 240 {
		t.Fatal("image paint escaped viewport", c)
	}
}

func TestManagedImageManualSharedWindowsAndDPIScaling(t *testing.T) {
	source := managedImageSource(t)
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(a.Close)
	view := func() ui.Node {
		return ui.Node{ID: "preview", Kind: ui.Image, Image: source, Text: "Shared preview", Fit: ui.ImageStretch, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 48, Height: 32})}}
	}
	for _, id := range []string{"one", "two"} {
		if err = a.Open(app.Window{ID: id, Title: "Image fixture", Width: 400, Height: 240, View: view}); err != nil {
			t.Fatal(err)
		}
	}
	config := manualConfig(t, 7)
	config.LogicalDPI = 320
	m, err := NewManual(testContext, a, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := m.Close(testContext); err != nil {
			t.Error(err)
		}
	})
	unitSettle(t, m)
	snapshot, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	var retired, surviving *managedImage
	if err = a.Do(func() {
		retired = m.implementation.find("one").state.root.widget.(*managedImage)
		surviving = m.implementation.find("two").state.root.widget.(*managedImage)
	}); err != nil {
		t.Fatal(err)
	}
	for _, window := range snapshot.Windows {
		if len(window.Elements) != 1 || window.Elements[0].Bounds.Size() != image.Pt(96, 64) {
			t.Fatal("logical DPI did not scale image box", window.Elements)
		}
		frame, err := m.Capture(testContext, window.ID)
		if err != nil {
			t.Fatal(err)
		}
		p := window.Elements[0].Bounds.Min.Add(image.Pt(48, 32))
		if got := frame.RGBAAt(p.X, p.Y); got != (color.RGBA{G: 240, A: 255}) {
			t.Fatal("independent image frame", window.ID, got)
		}
	}
	if retired.source != source || surviving.source != source {
		t.Fatal("shared immutable source copied or replaced")
	}
	if err = a.CloseWindow("one"); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if retired.source != nil || surviving.source != source {
		t.Fatal("one window retirement disturbed another image")
	}
	frame, err := m.Capture(testContext, "two")
	if err != nil || frame == nil {
		t.Fatal("surviving window image unavailable", err)
	}
}
