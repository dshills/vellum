// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
package cogent

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

// JSON implementations in supported Go toolchains differ in the byte spelling
// of invalid UTF-8 replacement. Probe this one fixed builtin once; no caller
// value, encoder storage or application callback is retained.
var accessibilityInvalidUTF8JSON = func() string {
	encoded, _ := json.Marshal("\xff")
	return string(encoded[1 : len(encoded)-1])
}()

// The fixed internal wire schema avoids reflection and intermediate JSON buffers
// for builtin scalar records. Custom values retain the original stdlib encoder,
// including its errors and side effects. No source or scratch backing is cached.
func accessibilityBuiltinSemantic(nodes []accessibilityNode) (string, bool) {
	return accessibilityBuiltinSemanticHinted(nodes, 0)
}

func accessibilityBuiltinSemanticHinted(nodes []accessibilityNode, hint int) (string, bool) {
	for i := range nodes {
		n := &nodes[i]
		switch v := n.Value.(type) {
		case nil, string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		case float32:
			if !accessibilityFinite(float64(v)) {
				return "", false
			}
		case float64:
			if !accessibilityFinite(v) {
				return "", false
			}
		default:
			return "", false
		}
		if n.Min != nil && !accessibilityFinite(float64(*n.Min)) || n.Max != nil && !accessibilityFinite(float64(*n.Max)) {
			return "", false
		}
		if n.GridFrame != nil {
			for _, v := range *n.GridFrame {
				if !accessibilityFinite(float64(v)) {
					return "", false
				}
			}
		}
	}
	if nodes == nil {
		return "null", true
	}
	if len(nodes) == 0 {
		return "[]", true
	}
	var e accessibilitySemanticWriter
	// This is an allocation hint, not retained scratch or an admission limit.
	// Bound the initial hint; unusually large strings retain ordinary growth.
	if hint <= 0 {
		hint = min(len(nodes), 2048) * 512
	}
	e.out.Grow(min(hint, 1<<20))
	e.out.WriteByte('[')
	for i := range nodes {
		if i != 0 {
			e.out.WriteByte(',')
		}
		e.node(&nodes[i])
	}
	e.out.WriteByte(']')
	return e.out.String(), true
}
func accessibilityFinite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

type accessibilitySemanticWriter struct {
	out    strings.Builder
	number [64]byte
}

func (e *accessibilitySemanticWriter) boolean(v bool) {
	if v {
		e.out.WriteString("true")
	} else {
		e.out.WriteString("false")
	}
}
func (e *accessibilitySemanticWriter) integer(v int64) {
	e.out.Write(strconv.AppendInt(e.number[:0], v, 10))
}
func (e *accessibilitySemanticWriter) unsigned(v uint64) {
	e.out.Write(strconv.AppendUint(e.number[:0], v, 10))
}
func (e *accessibilitySemanticWriter) floating(v float64, bits int) {
	format := byte('f')
	a := math.Abs(v)
	if a != 0 && (bits == 64 && (a < 1e-6 || a >= 1e21) || bits == 32 && (float32(a) < 1e-6 || float32(a) >= 1e21)) {
		format = 'e'
	}
	b := strconv.AppendFloat(e.number[:0], v, format, -1, bits)
	// JSON's exponent spelling has no redundant leading zero after the minus.
	n := len(b)
	if format == 'e' && n >= 4 && b[n-4] == 'e' && b[n-3] == '-' && b[n-2] == '0' {
		b[n-2] = b[n-1]
		b = b[:n-1]
	}
	e.out.Write(b)
}
func (e *accessibilitySemanticWriter) scalar(value any) {
	switch v := value.(type) {
	case string:
		e.quoted(v)
	case bool:
		e.boolean(v)
	case int:
		e.integer(int64(v))
	case int8:
		e.integer(int64(v))
	case int16:
		e.integer(int64(v))
	case int32:
		e.integer(int64(v))
	case int64:
		e.integer(v)
	case uint:
		e.unsigned(uint64(v))
	case uint8:
		e.unsigned(uint64(v))
	case uint16:
		e.unsigned(uint64(v))
	case uint32:
		e.unsigned(uint64(v))
	case uint64:
		e.unsigned(v)
	case uintptr:
		e.unsigned(uint64(v))
	case float32:
		e.floating(float64(v), 32)
	case float64:
		e.floating(v, 64)
	}
}
func (e *accessibilitySemanticWriter) quoted(s string) {
	e.out.WriteByte('"')
	start := 0
	for i := 0; i < len(s); {
		width := 1
		replacement := ""
		c := s[i]
		if c < utf8.RuneSelf {
			switch c {
			case '"':
				replacement = `\"`
			case '\\':
				replacement = `\\`
			case '\b':
				replacement = `\b`
			case '\f':
				replacement = `\f`
			case '\n':
				replacement = `\n`
			case '\r':
				replacement = `\r`
			case '\t':
				replacement = `\t`
			case '<':
				replacement = `\u003c`
			case '>':
				replacement = `\u003e`
			case '&':
				replacement = `\u0026`
			default:
				if c < 0x20 {
					e.out.WriteString(s[start:i])
					e.out.WriteString(`\u00`)
					e.out.WriteByte("0123456789abcdef"[c>>4])
					e.out.WriteByte("0123456789abcdef"[c&15])
					i++
					start = i
					continue
				}
			}
		} else {
			var r rune
			r, width = utf8.DecodeRuneInString(s[i:])
			if r == utf8.RuneError && width == 1 {
				replacement = accessibilityInvalidUTF8JSON
			} else if r == '\u2028' {
				replacement = `\u2028`
			} else if r == '\u2029' {
				replacement = `\u2029`
			}
		}
		if replacement != "" {
			e.out.WriteString(s[start:i])
			e.out.WriteString(replacement)
			i += width
			start = i
		} else {
			i += width
		}
	}
	e.out.WriteString(s[start:])
	e.out.WriteByte('"')
}
func (e *accessibilitySemanticWriter) node(n *accessibilityNode) {
	e.out.WriteByte('{')
	e.out.WriteString(`"id":`)
	e.quoted(n.ID)
	e.out.WriteString(`,"domain_id":`)
	e.quoted(n.DomainID)
	e.out.WriteString(`,"parent":`)
	e.quoted(n.Parent)
	e.out.WriteString(`,"role":`)
	e.quoted(n.Role)
	if n.Secure {
		e.out.WriteString(`,"secure":true`)
	}
	e.out.WriteString(`,"name":`)
	e.quoted(n.Name)
	e.out.WriteString(`,"help":`)
	e.quoted(n.Help)
	if n.Value != nil {
		e.out.WriteString(`,"value":`)
		e.scalar(n.Value)
	}
	e.out.WriteString(`,"disabled":`)
	e.boolean(n.Disabled)
	e.out.WriteString(`,"focusable":`)
	e.boolean(n.Focusable)
	e.out.WriteString(`,"focused":`)
	e.boolean(n.Focused)
	e.out.WriteString(`,"selected":`)
	e.boolean(n.Selected)
	e.out.WriteString(`,"actions":`)
	if n.Actions == nil {
		e.out.WriteString("null")
	} else {
		e.out.WriteByte('[')
		for i, v := range n.Actions {
			if i != 0 {
				e.out.WriteByte(',')
			}
			e.quoted(v)
		}
		e.out.WriteByte(']')
	}
	if n.Orientation != "" {
		e.out.WriteString(`,"orientation":`)
		e.quoted(n.Orientation)
	}
	if n.RowIndex != 0 {
		e.out.WriteString(`,"row_index":`)
		e.integer(int64(n.RowIndex))
	}
	if n.RowCount != 0 {
		e.out.WriteString(`,"row_count":`)
		e.integer(int64(n.RowCount))
	}
	if n.ColumnIndex != 0 {
		e.out.WriteString(`,"column_index":`)
		e.integer(int64(n.ColumnIndex))
	}
	if n.ColumnCount != 0 {
		e.out.WriteString(`,"column_count":`)
		e.integer(int64(n.ColumnCount))
	}
	if n.GridGeneration != 0 {
		e.out.WriteString(`,"grid_generation":`)
		e.unsigned(n.GridGeneration)
	}
	if n.GridScope != "" {
		e.out.WriteString(`,"grid_scope":`)
		e.quoted(n.GridScope)
	}
	if n.GridIncarnation != 0 {
		e.out.WriteString(`,"grid_incarnation":`)
		e.unsigned(n.GridIncarnation)
	}
	if n.GridPending {
		e.out.WriteString(`,"grid_pending":`)
		e.boolean(n.GridPending)
	}
	if n.GridTargetGeneration != 0 {
		e.out.WriteString(`,"grid_target_generation":`)
		e.unsigned(n.GridTargetGeneration)
	}
	if n.GridFrame != nil {
		e.out.WriteString(`,"grid_frame":`)
		e.out.WriteByte('[')
		for i, v := range *n.GridFrame {
			if i != 0 {
				e.out.WriteByte(',')
			}
			e.floating(float64(v), 32)
		}
		e.out.WriteByte(']')
	}
	if n.Level != 0 {
		e.out.WriteString(`,"level":`)
		e.integer(int64(n.Level))
	}
	if n.Expanded != nil {
		e.out.WriteString(`,"expanded":`)
		e.boolean(*n.Expanded)
	}
	if n.Min != nil {
		e.out.WriteString(`,"min":`)
		e.floating(float64(*n.Min), 32)
	}
	if n.TextClient != 0 {
		e.out.WriteString(`,"text_client":`)
		e.unsigned(n.TextClient)
	}
	if n.TextSelection != nil {
		e.out.WriteString(`,"text_selection":`)
		e.out.WriteByte('[')
		e.integer(int64(n.TextSelection[0]))
		e.out.WriteByte(',')
		e.integer(int64(n.TextSelection[1]))
		e.out.WriteByte(']')
	}
	if n.TextFocus != 0 {
		e.out.WriteString(`,"text_focus":`)
		e.integer(int64(n.TextFocus))
	}
	if n.Max != nil {
		e.out.WriteString(`,"max":`)
		e.floating(float64(*n.Max), 32)
	}
	e.out.WriteByte('}')
}

// Capacity reuse borrows only the length of an existing owned semantic string.
// It certifies no wire semantics, epoch, validity or snapshot identity. Mismatched
// membership uses the normal hint; growth always remains legal. No lock crosses
// the encoder or an application marshaler callback.
func (b *accessibilityBridge) semanticAllocationHint(count int) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	if count <= 0 || b.wireBaseline().len() != count {
		return 0
	}
	hint := min(len(b.lastSnapshot), 1<<20)
	slack := min(count, 65536) * 16
	return hint + min(slack, (1<<20)-hint)
}
