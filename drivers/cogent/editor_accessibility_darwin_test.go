//go:build darwin && cgo

package cogent

import (
	"cogentcore.org/core/core"
	"encoding/json"
	"github.com/dshills/vellum/drivers/cogent/internal/textinput"
	"github.com/dshills/vellum/ui"
	"strings"
	"testing"
)

func TestNativeAccessibilityTextSelectorsExactUTF16AndStaleGeometry(t *testing.T) {
	node := map[string]any{"id": "text", "role": "multiline-text", "actions": []string{"selection"}, "value": "A😀\r\nB\n\n", "text_client": 1, "text_selection": []int{1, 2}, "text_focus": 6}
	geometry := map[string]any{"client": 1, "text": node["value"], "geometry": map[string]any{"viewport": []int{0, 0, 100, 60}, "focus_line": 1, "cells": []any{map[string]any{"start": 1, "end": 3, "line": 0, "rect": []int{10, 0, 20, 20}}, map[string]any{"start": 5, "end": 6, "line": 1, "rect": []int{0, 20, 10, 20}}}, "carets": []any{map[string]any{"offset": 3, "line": 0, "rect": []int{30, 0, 0, 20}}}}}
	query := func(start, length, line int) map[string]any {
		n, _ := json.Marshal(node)
		g, _ := json.Marshal(geometry)
		var result map[string]any
		if err := json.Unmarshal([]byte(nativeAccessibilityTextFixture(string(n), string(g), start, length, line)), &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	out := query(1, 2, 0)
	if out["substring"] != "😀" || out["selected_start"] != float64(1) || out["selected_length"] != float64(2) || out["insertion_line"] != float64(1) || out["line_start"] != float64(0) || out["line_length"] != float64(5) {
		t.Fatal("native UTF16/CRLF selectors disagree", out)
	}
	frame := out["frame"].([]any)
	if frame[0] != float64(10) || frame[2] != float64(20) || frame[3] != float64(20) {
		t.Fatal("rendered range geometry missing", out)
	}
	out = query(2, 1, 3)
	if out["substring"] != "" || out["frame"].([]any)[3] != float64(0) || out["line_start"] != float64(8) || out["line_length"] != float64(0) {
		t.Fatal("surrogate interior/final blank range accepted incorrectly", out)
	}

	// A caret between CR and LF remains on the line containing the complete
	// CRLF terminator; only the boundary after LF enters the next line.
	node["text_focus"] = 4
	if out = query(4, 0, 0); out["insertion_line"] != float64(0) {
		t.Fatal("CRLF interior advanced line early", out)
	}
	node["text_focus"] = 5
	if out = query(5, 0, 1); out["insertion_line"] != float64(1) {
		t.Fatal("CRLF end did not advance line", out)
	}
	node["text_focus"] = 6
	g := geometry["geometry"].(map[string]any)
	cells := g["cells"]
	g["cells"] = nil
	if out = query(1, 2, 0); len(out) != 0 {
		t.Fatal("null geometry accepted", out)
	}
	g["cells"] = cells
	for _, change := range []func(){func() { geometry["client"] = 2 }, func() { geometry["client"] = 1; geometry["text"] = "Z😀\r\nB\n\n" }, func() { geometry["text"] = node["value"]; node["disabled"] = true }} {
		change()
		out = query(1, 2, 0)
		if out["available"] != false || out["frame"].([]any)[3] != float64(0) {
			t.Fatal("stale or disabled geometry answered range", out)
		}
	}
}

func TestNativeAccessibilityFieldSetBoundsBeforeCStringAdmission(t *testing.T) {
	a, b := newAXTest(t, 8)
	body := core.NewBody("native AX field admission")
	var field *record
	calls := 0
	n := ui.Node{ID: "field", Kind: ui.TextField, Text: "accepted", Change: func(e ui.ChangeEvent) {
		calls++
		field.node.Text = e.Text
	}}
	var err error
	field, err = reconcile(body, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { disposeControlTree(field) })
	nodes, handlers := accessibilitySnapshot(field, b)
	b.mu.Lock()
	b.handlers = handlers
	b.mu.Unlock()
	b.generation.Store(1)
	node, err := json.Marshal(nodes[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{strings.Repeat("x", textinput.MaxDocumentBytes+1), "prefix\x00suffix"} {
		encoded, _ := json.Marshal(value)
		if nativeAccessibilitySetValueFixture(b.token, 1, string(node), string(encoded)) {
			t.Fatal("invalid NSString reached Go admission")
		}
		a.Drain()
		if calls != 0 || field.node.Text != "accepted" || field.control.committed {
			t.Fatal("refused native AX value changed controlled state")
		}
	}
	encoded, _ := json.Marshal("valid🙂")
	if !nativeAccessibilitySetValueFixture(b.token, 1, string(node), string(encoded)) {
		t.Fatal("valid native AX setter refused")
	}
	a.Drain()
	if calls != 1 || field.node.Text != "valid🙂" {
		t.Fatal("valid native AX setter lost exact bytes")
	}
}

func TestNativeAccessibilityEditorSetRejectsNULWithoutTruncatingDocument(t *testing.T) {
	a, w, _, r, m := editorFixture(t, "accepted\r\n\r\n")
	b := newAccessibilityBridge(a)
	b.bind(w.window.ID)
	t.Cleanup(b.close)
	nodes, handlers := accessibilitySnapshot(r, b)
	b.mu.Lock()
	b.handlers = handlers
	b.mu.Unlock()
	b.generation.Store(1)
	node, _ := json.Marshal(nodes[0])
	before := m.Snapshot()
	value, _ := json.Marshal("prefix\x00suffix")
	if nativeAccessibilitySetValueFixture(b.token, 1, string(node), string(value)) {
		t.Fatal("NUL document value admitted as a truncated prefix")
	}
	a.Drain()
	if m.Snapshot() != before {
		t.Fatal("refused NSString changed document or journal")
	}
	value, _ = json.Marshal("valid😀\r\n\r\n")
	if !nativeAccessibilitySetValueFixture(b.token, 1, string(node), string(value)) {
		t.Fatal("valid document value refused")
	}
	a.Drain()
	if m.Snapshot().Text != "valid😀\r\n\r\n" || m.Snapshot().History.UndoCount != before.History.UndoCount+1 {
		t.Fatal("valid NSString lost exact bytes or history")
	}
}
