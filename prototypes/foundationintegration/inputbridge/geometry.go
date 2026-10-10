package inputbridge

import (
	"cogentcore.org/core/text/textpos"
	"errors"
	"math"
	"strings"
	"unicode/utf8"
)

// Geometry is an owned layout snapshot in top-left view points. Publish it in
// the same transaction as its text and acknowledgement; native edits invalidate it.
// A zero-width caret records a strong bidi edge, with explicit wrapped-line affinity.
type Geometry struct {
	Cells     []TextCell
	Carets    []TextCaret
	Viewport  [4]float64
	FocusLine int
}
type TextCell struct {
	Start, End, Line int
	Rect             [4]float64
}
type TextCaret struct {
	Offset, Line int
	Rect         [4]float64
}

func validRect(r [4]float64, caret bool) bool {
	for _, v := range r {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}
	return r[2] >= 0 && r[3] > 0 && (!caret || r[2] == 0) && !math.IsInf(r[0]+r[2], 0) && !math.IsInf(r[1]+r[3], 0)
}
func (g Geometry) Validate(text string) error {
	if !utf8.ValidString(text) || strings.ContainsRune(text, 0) || len(text) > 1<<20 {
		return errors.New("invalid geometry text")
	}
	offsets := UTF16Offsets(text)
	bounds := map[int]bool{}
	for _, r := range textpos.GraphemeBounds([]rune(text)) {
		u := offsets[r]
		bounds[u] = true
	}
	if len(g.Cells) > 262144 || len(g.Carets) > 262144 || g.FocusLine < -1 || g.FocusLine > 1<<30 || len(g.Cells) > len(bounds)*2 || len(g.Carets) > len(bounds)*2 || !validRect(g.Viewport, false) {
		return errors.New("invalid geometry size or viewport")
	}
	for _, c := range g.Cells {
		if !bounds[c.Start] || !bounds[c.End] || c.Start >= c.End || c.Line < 0 || c.Line > 1<<30 || !validRect(c.Rect, false) {
			return errors.New("invalid geometry cell")
		}
	}
	seen := map[[2]int]bool{}
	for _, c := range g.Carets {
		key := [2]int{c.Offset, c.Line}
		if !bounds[c.Offset] || c.Line < 0 || c.Line > 1<<30 || !validRect(c.Rect, true) || seen[key] {
			return errors.New("invalid geometry caret")
		}
		seen[key] = true
	}
	return nil
}
