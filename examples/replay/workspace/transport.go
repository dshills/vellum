package workspace

import (
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/replay"
	"unicode/utf8"
)

// Key names are this fixture's declared transport: one Unicode scalar for text,
// or one of the explicit special codes. Modifier bits are Shift/Control/Alt/Meta.
func transportKey(value string) (rune, key.Codes, error) {
	codes := map[string]key.Codes{"Enter": key.CodeReturnEnter, "Escape": key.CodeEscape, "Tab": key.CodeTab, "Backspace": key.CodeBackspace, "Delete": key.CodeDelete, "Left": key.CodeLeftArrow, "Right": key.CodeRightArrow, "Up": key.CodeUpArrow, "Down": key.CodeDownArrow, "Home": key.CodeHome, "End": key.CodeEnd, "F2": key.CodeF2}
	if code, ok := codes[value]; ok {
		return 0, code, nil
	}
	if utf8.RuneCountInString(value) != 1 {
		return 0, 0, replay.ErrInvalid
	}
	r, _ := utf8.DecodeRuneInString(value)
	code := key.CodeUnknown
	if r == 'a' || r == 'A' {
		code = key.CodeA
	}
	if r == 'z' || r == 'Z' {
		code = key.CodeZ
	}
	return r, code, nil
}
func transportModifiers(bits uint32) key.Modifiers {
	var m key.Modifiers
	for i, flag := range []key.Modifiers{key.Shift, key.Control, key.Alt, key.Meta} {
		if bits&(1<<i) != 0 {
			m.SetFlag(true, flag)
		}
	}
	return m
}
