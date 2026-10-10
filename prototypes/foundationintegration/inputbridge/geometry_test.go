package inputbridge

import (
	"math"
	"reflect"
	"testing"
)

func TestGeometryValidationAndEncoding(t *testing.T) {
	source := "A👩‍💻éZ"
	if got := UTF16Offsets(source); !reflect.DeepEqual(got, []int{0, 1, 3, 4, 6, 7, 8, 9}) {
		t.Fatalf("UTF16 map %v", got)
	}
	good := Geometry{Viewport: [4]float64{0, 0, 100, 30}, Cells: []TextCell{{1, 6, 0, [4]float64{10, 0, 20, 23}}}, Carets: []TextCaret{{6, 0, [4]float64{30, 0, 0, 23}}}}
	if good.Validate(source) != nil {
		t.Fatal("valid emoji cell rejected")
	}
	for _, edit := range []func(*Geometry){
		func(g *Geometry) { g.Cells[0].End = 3 }, func(g *Geometry) { g.Carets[0].Offset = 4 },
		func(g *Geometry) { g.Viewport[0] = math.NaN() }, func(g *Geometry) { g.Cells[0].Rect[3] = math.Inf(1) },
		func(g *Geometry) { g.Cells[0].Line = -1 }, func(g *Geometry) { g.Carets = append(g.Carets, g.Carets[0]) },
	} {
		g := good
		g.Cells = append([]TextCell(nil), good.Cells...)
		g.Carets = append([]TextCaret(nil), good.Carets...)
		edit(&g)
		if g.Validate(source) == nil {
			t.Fatal("invalid geometry admitted")
		}
	}
}
