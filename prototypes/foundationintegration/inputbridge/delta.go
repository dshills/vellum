package inputbridge

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Native documents and delta payloads have independent budgets. JSON can escape
// each byte of a payload into six bytes; two 4096-byte fragments fit 64 KiB.
const MaxDocumentBytes = 4 << 20
const MaxDeltaFragmentBytes = 4096

// ExpandDelta validates a bounded edit against an owned shadow. Selection and
// marked ranges remain global UTF-16 offsets, preserving Cocoa's range contract.
// The application must Decode the expanded selection and compare this shadow
// with its current document atomically.
func ExpandDelta(u Update, shadow string) (Update, error) {
	if !u.Delta {
		return u, nil
	}
	if !utf8.ValidString(shadow) || len(shadow) > MaxDocumentBytes || strings.ContainsRune(shadow, 0) || len(u.Text) > MaxDeltaFragmentBytes || len(u.BeforeText) > MaxDeltaFragmentBytes || !utf8.ValidString(u.Text) || !utf8.ValidString(u.BeforeText) || strings.ContainsRune(u.Text, 0) || strings.ContainsRune(u.BeforeText, 0) || u.ReplaceStart < 0 || u.ReplaceEnd < u.ReplaceStart {
		return u, fmt.Errorf("invalid delta")
	}
	start, end := -1, -1
	offset := 0
	for byteIndex, r := range shadow {
		if offset == u.ReplaceStart {
			start = byteIndex
		}
		if offset == u.ReplaceEnd {
			end = byteIndex
		}
		offset++
		if r > 0xffff {
			offset++
		}
	}
	if offset == u.ReplaceStart {
		start = len(shadow)
	}
	if offset == u.ReplaceEnd {
		end = len(shadow)
	}
	if start < 0 || end < 0 || shadow[start:end] != u.BeforeText {
		return u, fmt.Errorf("delta range or expected fragment mismatch")
	}
	if len(shadow)-len(u.BeforeText)+len(u.Text) > MaxDocumentBytes {
		return u, fmt.Errorf("document budget exceeded")
	}
	u.BeforeText = shadow
	u.Text = shadow[:start] + u.Text + shadow[end:]
	u.Delta = false
	return u, nil
}
