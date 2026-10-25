package ui

import (
	"image/color"
	"math"
	"testing"
)

func sampleGradient() GradientValue {
	return GradientValue{Stops: []GradientStop{{ID: "a", Position: 0, Color: color.NRGBA{R: 255, A: 255}}, {ID: "b", Position: .5, Color: color.NRGBA{B: 255, A: 255}}, {ID: "c", Position: 1, Color: color.NRGBA{G: 255, A: 255}}}}
}
func TestGradientTypedEditsAndStableIdentity(t *testing.T) {
	v := sampleGradient()
	next, e := (GradientEdit{Kind: GradientMove, StopID: "a", Position: .75}).Apply(v)
	if e != nil || next.Stops[1].ID != "a" || v.Stops[0].Position != 0 {
		t.Fatal(next, e)
	}
	next, e = (GradientEdit{Kind: GradientColor, StopID: "a", Color: ColorEdit{Channel: ColorAlpha, Byte: 123}}).Apply(next)
	if e != nil || next.Stops[1].Color.R != 255 || next.Stops[1].Color.A != 123 {
		t.Fatal(next, e)
	}
	if _, e = (GradientEdit{Kind: GradientMove, StopID: "a", Position: float32(math.NaN())}).Apply(v); e == nil {
		t.Fatal("NaN")
	}
	if _, e = (GradientEdit{Kind: GradientRemove, StopID: "missing"}).Apply(v); e == nil {
		t.Fatal("missing")
	}
}
func TestGradientTrackCloseAndNativeAcceptAreRevoked(t *testing.T) {
	g := &GradientWell{ID: "gradient", Label: "Gradient", Value: sampleGradient()}
	calls := 0
	g.Change = func(e GradientEdit) {
		next, err := e.Apply(g.Value)
		if err != nil {
			t.Error(err)
			return
		}
		g.Value = next
		calls++
	}
	n, e := g.View()
	if e != nil {
		t.Fatal(e)
	}
	pickerNode(n, GradientWellButtonID(g.ID)).Activate()
	n, e = g.View()
	if e != nil {
		t.Fatal(e)
	}
	old := pickerNode(n, GradientTrackID(g.ID)).Track
	old.Adjust("b", 10)
	if g.Value.Stops[1].Position != .6 {
		t.Fatal(g.Value)
	}
	g.close()
	old.Set("b", .2)
	if calls != 1 {
		t.Fatal("closed track")
	}
	g.NativeColor = func(q GradientColorRequest) { g.close(); q.Accept(color.NRGBA{A: 255}) }
	n, _ = g.View()
	pickerNode(n, GradientWellButtonID(g.ID)).Activate()
	n, _ = g.View()
	pickerNode(n, g.ID+"/gradient/native").Activate()
	if calls != 1 {
		t.Fatal("stale native accept")
	}
}
func TestGradientRasterColorsAndBudget(t *testing.T) {
	v := sampleGradient()
	image, e := GradientImage(v, 3, 1)
	if e != nil {
		t.Fatal(e)
	}
	r, g, b, _ := image.At(0, 0).RGBA()
	if r < 60000 || g != 0 || b != 0 {
		t.Fatal(r, g, b)
	}
	_, _, b, _ = image.At(1, 0).RGBA()
	if b < 60000 {
		t.Fatal("midpoint")
	}
	if _, e = GradientImage(v, 1025, 1); e == nil {
		t.Fatal("unbounded raster")
	}
}

func TestGradientEqualStopOrderAndInvalidAdd(t *testing.T) {
	v := sampleGradient()
	n, e := (GradientEdit{Kind: GradientMove, StopID: "a", Position: .5}).Apply(v)
	if e != nil || n.Stops[0].ID != "a" || n.Stops[1].ID != "b" {
		t.Fatal(n, e)
	}
	if _, e = (GradientEdit{Kind: GradientAdd, Stop: GradientStop{ID: "bad", Position: float32(math.NaN())}}).Apply(v); e == nil {
		t.Fatal("nonfinite add")
	}
}
