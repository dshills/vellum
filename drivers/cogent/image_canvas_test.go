package cogent

import (
	"errors"
	"image"
	"image/color"
	"math"
	"testing"

	"cogentcore.org/core/core"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/oliverbestmann/webgpu/wgpu"
)

func TestImageOnlyAdapterAndPreallocationAdmission(t *testing.T) {
	source, _ := canvas.NewTiledImage(image.NewNRGBA(image.Rect(0, 0, 129, 130)))
	layer := canvas.ImageLayer{Image: source, Bounds: canvas.Bounds{Max: canvas.Point{X: 129, Y: 130}}}
	limits := wgpu.DefaultLimits()
	limits.MaxTextureDimension2D = 128
	if imageAdmission(layer, limits) == nil {
		t.Fatal("image texture limit ignored")
	}
	limits.MaxTextureDimension2D = 8192
	limits.MaxBufferSize = 31
	if imageAdmission(layer, limits) == nil {
		t.Fatal("image uniform buffer limit ignored")
	}
	limits.MaxBufferSize = 32
	if err := imageAdmission(layer, limits); err != nil {
		t.Fatal(err)
	}
	a, _ := app.New(app.Config{})
	defer a.Close()
	if err := a.Open(scopeDeclaration("standalone-unit")); err != nil {
		t.Fatal(err)
	}
	view := canvas.Viewport{Width: 200, Height: 100, Zoom: 1}
	g := &MeshCanvas{Owner: a, Window: "standalone-unit", View: &view, Image: &layer}
	if g.Mount(core.NewBody("image-only")) == nil {
		t.Fatal("image-only mount")
	}
	var resources meshResources
	if err := resources.render(nil, image.Pt(10, 10), emptyDrawingMesh, view, color.NRGBA{}, imageBytes(layer), layer); err == nil || resources.frame != nil || resources.image != nil {
		t.Fatal("budget must refuse before any allocation")
	}
}
func TestImageUploadBaselineRepairsAllTilesAfterPartialFailure(t *testing.T) {
	source, _ := canvas.NewTiledImage(image.NewNRGBA(image.Rect(0, 0, 300, 130)))
	var resources imageResources
	var stats DrawingStats
	failed := errors.New("controlled upload fault")
	writes := 0
	err := resources.uploadTiles(source, &stats, func(canvas.ImageTile) error {
		writes++
		if writes == 2 {
			return failed
		}
		return nil
	})
	if !errors.Is(err, failed) || resources.source != nil || stats.ImageTileUploads != 1 {
		t.Fatal("partial baseline published", err, stats)
	}
	writes = 0
	if err := resources.uploadTiles(source, &stats, func(canvas.ImageTile) error { writes++; return nil }); err != nil || writes != 6 || resources.source != source {
		t.Fatal("retry failed to repair full source", writes, err)
	}
	patch := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	patch.Pix[0] = 200
	next, _ := source.WithPatch(image.Rect(129, 1, 130, 2), patch)
	writes = 0
	if err := resources.uploadTiles(next, &stats, func(canvas.ImageTile) error { writes++; return nil }); err != nil || writes != 1 {
		t.Fatal("changed tile upload", writes, err)
	}
	before := stats
	if err := resources.uploadTiles(next, &stats, func(canvas.ImageTile) error { t.Fatal("unchanged write"); return nil }); err != nil || stats != before {
		t.Fatal("unchanged baseline")
	}
	var independent imageResources
	writes = 0
	independent.uploadTiles(next, &stats, func(canvas.ImageTile) error { writes++; return nil })
	if writes != 6 {
		t.Fatal("window reused another upload baseline")
	}
}
func TestImagePlacementRebasesBeforeFloat32(t *testing.T) {
	source, _ := canvas.NewTiledImage(image.NewNRGBA(image.Rect(0, 0, 2, 1)))
	var result [][8]float32
	for _, translation := range []float64{0, 999999} {
		layer := canvas.ImageLayer{Image: source, Bounds: canvas.Bounds{Min: canvas.Point{X: translation + .01}, Max: canvas.Point{X: translation + .03, Y: .01}}, Filter: canvas.ImageLinear}
		result = append(result, imageTransform(layer, canvas.Viewport{Width: 200, Height: 100, Zoom: 100, Center: canvas.Point{X: translation + .015}}))
	}
	for i := range result[0] {
		if math.Abs(float64(result[0][i]-result[1][i])) > 1e-7 {
			t.Fatal("translated placement lost", result)
		}
	}
}
