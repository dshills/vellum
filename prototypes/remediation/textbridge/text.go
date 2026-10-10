// Package textbridge converts owned logical text offsets for native semantics.
package textbridge

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"cogentcore.org/core/text/textpos"
)

type Run struct {
	RuneBounds  []int
	UTF16Bounds []int
	ByteLengths []uint8
}

func New(text string) (Run, error) {
	if !utf8.ValidString(text) {
		return Run{}, fmt.Errorf("invalid UTF-8")
	}
	if strings.ContainsRune(text, 0) {
		return Run{}, fmt.Errorf("embedded NUL cannot be represented by native C strings")
	}
	runes := []rune(text)
	bounds := textpos.GraphemeBounds(runes)
	r := Run{RuneBounds: bounds, UTF16Bounds: []int{0}}
	for i := 1; i < len(bounds); i++ {
		cluster := runes[bounds[i-1]:bounds[i]]
		length := len(string(cluster))
		if length > 255 {
			return Run{}, fmt.Errorf("cluster exceeds AccessKit byte-length representation")
		}
		r.ByteLengths = append(r.ByteLengths, uint8(length))
		r.UTF16Bounds = append(r.UTF16Bounds, r.UTF16Bounds[len(r.UTF16Bounds)-1]+len(utf16.Encode(cluster)))
	}
	return r, nil
}
func (r Run) CharacterAtRune(offset int) (int, bool) {
	index := slices.Index(r.RuneBounds, offset)
	return index, index >= 0
}
func (r Run) RuneAtCharacter(index int) (int, bool) {
	if index < 0 || index >= len(r.RuneBounds) {
		return 0, false
	}
	return r.RuneBounds[index], true
}
func (r Run) CharacterAtUTF16(offset int) (int, bool) {
	index := slices.Index(r.UTF16Bounds, offset)
	return index, index >= 0
}
