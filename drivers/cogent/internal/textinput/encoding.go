// Package textinput adapts Cocoa UTF-16 ranges to owned UI-thread updates.
package textinput

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Update struct {
	Sequence, ClientID                    uint64
	Delta                                 bool
	ReplaceStart, ReplaceEnd              int // UTF-16 range in the prior document when Delta is true.
	Kind, Text, BeforeText                string
	Anchor, Focus, MarkedStart, MarkedEnd int // UTF-16 offsets; -1 means unmarked.
}

type RuneUpdate struct {
	Update
	RuneAnchor, RuneFocus, RuneMarkedStart, RuneMarkedEnd int
}

// Decode rejects malformed strings, surrogate interiors and invalid ranges.
// IME selections may be inside graphemes but must be on Unicode scalar bounds.
func Decode(u Update) (RuneUpdate, error) {
	if u.Delta {
		return RuneUpdate{}, fmt.Errorf("delta requires an owned document shadow")
	}
	result := RuneUpdate{Update: u, RuneMarkedStart: -1, RuneMarkedEnd: -1}
	if !utf8.ValidString(u.Text) || strings.ContainsRune(u.Text, 0) || !utf8.ValidString(u.BeforeText) || strings.ContainsRune(u.BeforeText, 0) {
		return result, fmt.Errorf("unrepresentable text")
	}
	positions := map[int]int{0: 0}
	utf16Offset, runeOffset := 0, 0
	for _, r := range u.Text {
		utf16Offset++
		if r > 0xffff {
			utf16Offset++
		}
		runeOffset++
		// Retain only queried boundaries; a large document must not allocate
		// one hash entry for every rune on each keystroke.
		if utf16Offset == u.Anchor || utf16Offset == u.Focus || utf16Offset == u.MarkedStart || utf16Offset == u.MarkedEnd {
			positions[utf16Offset] = runeOffset
		}
	}
	var ok bool
	if result.RuneAnchor, ok = positions[u.Anchor]; !ok {
		return result, fmt.Errorf("invalid anchor")
	}
	if result.RuneFocus, ok = positions[u.Focus]; !ok {
		return result, fmt.Errorf("invalid focus")
	}
	if u.MarkedStart == -1 && u.MarkedEnd == -1 {
		return result, nil
	}
	if u.MarkedStart < 0 || u.MarkedEnd < u.MarkedStart {
		return result, fmt.Errorf("invalid marked range")
	}
	if result.RuneMarkedStart, ok = positions[u.MarkedStart]; !ok {
		return result, fmt.Errorf("invalid marked start")
	}
	if result.RuneMarkedEnd, ok = positions[u.MarkedEnd]; !ok {
		return result, fmt.Errorf("invalid marked end")
	}
	if u.Anchor < u.MarkedStart || u.Anchor > u.MarkedEnd || u.Focus < u.MarkedStart || u.Focus > u.MarkedEnd {
		return result, fmt.Errorf("selection outside marked range")
	}
	return result, nil
}
func UTF16Offset(text string, runeOffset int) (int, bool) {
	if runeOffset < 0 {
		return 0, false
	}
	offset, index := 0, 0
	for _, r := range text {
		if index == runeOffset {
			return offset, true
		}
		index++
		offset++
		if r > 0xffff {
			offset++
		}
	}
	return offset, index == runeOffset
}

// UTF16Offsets maps each rune boundary in one pass; geometry has many offsets.
func UTF16Offsets(text string) []int {
	offsets := []int{0}
	u := 0
	for _, r := range text {
		if r > 0xffff {
			u += 2
		} else {
			u++
		}
		offsets = append(offsets, u)
	}
	return offsets
}
