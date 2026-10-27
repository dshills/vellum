package quantity

import (
	"math"
	"strings"
	"testing"
)

func TestLengthExpressions(t *testing.T) {
	for _, c := range []struct {
		s, u string
		want float64
	}{{"2 in + 5 mm", "mm", 55.8}, {"width/2", "mm", 50}, {"(2+3)*4", "cm", 200}, {"-2.5 mm", "mm", -2.5}, {"1e-2 m", "mm", 10}} {
		v, e := Length(c.s, c.u, map[string]float64{"width": 100})
		if e != nil || math.Abs(v-c.want) > 1e-8 {
			t.Fatal(c, v, e)
		}
	}
}
func TestRejectLengthExpressions(t *testing.T) {
	for _, s := range []string{"", "1/0", "2 mm + 3", "width()", "os.Exit(1)", "2 mm * 3 mm", "unknown", "1e999", strings.Repeat("(", 30) + "1" + strings.Repeat(")", 30), "0x10", "1 // comment"} {
		if _, e := Length(s, "mm", nil); e == nil {
			t.Fatal("accepted", s)
		}
	}
	if _, e := Length("width", "mm", map[string]float64{"width": math.NaN()}); e == nil {
		t.Fatal("NaN variable")
	}
}

func TestUnaryPrecedenceAndGroupUnits(t *testing.T) {
	for _, c := range []struct {
		s    string
		want float64
	}{{"-2*3", -6}, {"-(2+3) cm", -50}, {"(1+2) mm", 3}, {"--2", 2}, {"2 + 3*4", 14}} {
		v, e := Length(c.s, "mm", nil)
		if e != nil || v != c.want {
			t.Fatal(c, v, e)
		}
	}
	if _, e := Length("(2 mm) cm", "mm", nil); e == nil {
		t.Fatal("double unit suffix")
	}
}
