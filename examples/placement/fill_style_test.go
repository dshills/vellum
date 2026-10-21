package main

import (
	"image/color"
	"maps"
	"slices"
	"strings"
	"testing"
)

func TestFillStyleCopyUsesActualPrimaryAndLeavesState(t *testing.T) {
	m := groupedModel(t)
	m.scene.Objects[0].Color = color.RGBA{R: 17, G: 34, B: 51, A: 255}
	m.scene.Objects[2].Color = color.RGBA{R: 170, G: 187, B: 204, A: 255}
	m.setSelection([]string{"block-1"}, "block-1")
	m.pasteText, m.pasteCount = "existing component paste", 4
	before, state := slices.Clone(m.scene.Objects), m.journal.Snapshot()
	text, err := m.copyFillStyle()
	if err != nil {
		t.Fatal(err)
	}
	value, err := decodeFillStyle(text)
	if err != nil || value != before[0].Color {
		t.Fatal("did not copy primary member", value, err)
	}
	m.setSelection(m.selectedIDs(), "note-3")
	text, err = m.copyFillStyle()
	if err != nil {
		t.Fatal(err)
	}
	value, err = decodeFillStyle(text)
	if err != nil || value != before[2].Color || !slices.Equal(before, m.scene.Objects) || state != m.journal.Snapshot() || m.pasteText != "existing component paste" || m.pasteCount != 4 {
		t.Fatal("copy changed document/history/clipboard paste state", value, err)
	}
}

func TestFillStylePasteWholeSelectionMetadataAndOneUndoRedo(t *testing.T) {
	m := groupedModel(t)
	m.layerInfo = map[string]layerMetadata{"block-1": {Name: "Housing"}, "note-3": {Name: "Label"}, "group-1": {Name: "Assembly"}}
	m.pasteText, m.pasteCount = "component clipboard", 7
	before, groups, info := slices.Clone(m.scene.Objects), maps.Clone(m.groups), maps.Clone(m.layerInfo)
	ids, primary := slices.Clone(m.selectedIDs()), m.scene.Selected
	state := m.journal.Snapshot()
	text, err := encodeFillStyle(color.RGBA{R: 201, G: 80, B: 17, A: 255})
	if err != nil {
		t.Fatal(err)
	}
	if err = m.pasteFillStyle(text); err != nil {
		t.Fatal(err)
	}
	after := slices.Clone(m.scene.Objects)
	for i, o := range after {
		want := before[i]
		if slices.Contains(ids, o.ID) {
			want.Color = color.RGBA{R: 201, G: 80, B: 17, A: 255}
		}
		if o != want {
			t.Fatal("paste changed geometry/identity/order or missed grouped member", o, want)
		}
	}
	if !maps.Equal(m.groups, groups) || !maps.Equal(m.layerInfo, info) || !slices.Equal(ids, m.selectedIDs()) || primary != m.scene.Selected || m.journal.Snapshot().UndoCount != state.UndoCount+1 || m.pasteText != "component clipboard" || m.pasteCount != 7 {
		t.Fatal("style did not preserve metadata, selection or paste sequencing")
	}
	if err = m.journal.Undo(); err != nil || !slices.Equal(before, m.scene.Objects) {
		t.Fatal("Undo style", err)
	}
	if err = m.journal.Redo(); err != nil || !slices.Equal(after, m.scene.Objects) {
		t.Fatal("Redo style", err)
	}
	state = m.journal.Snapshot()
	if err = m.pasteFillStyle(text); err != nil || state != m.journal.Snapshot() {
		t.Fatal("no-op style changed history", err)
	}
	f, err := drawingCodec(m).Capture()
	if err != nil {
		t.Fatal(err)
	}
	data, err := drawingCodec(m).Encode(f)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := drawingCodec(m).Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	scene, err := decoded.scene()
	if err != nil || !slices.Equal(scene.Objects, after) {
		t.Fatal("saved style round trip", err)
	}
}

func TestFillStyleStrictWireAndAtomicRejection(t *testing.T) {
	valid := `{"format":"vellum-placement-fill","version":1,"color":"#aBcD01"}`
	value, err := decodeFillStyle(valid)
	if err != nil || value != (color.RGBA{R: 171, G: 205, B: 1, A: 255}) {
		t.Fatal("valid mixed-case RGB", value, err)
	}
	bad := []string{"", "ordinary text", "#123456", "null", "[]", valid + "{}", strings.Repeat(" ", fillStyleByteLimit+1), strings.Replace(valid, "version\":1", "version\":2", 1), strings.Replace(valid, "#aBcD01", "#abcd", 1), strings.Replace(valid, "#aBcD01", "#abcd01ff", 1), strings.Replace(valid, "#aBcD01", "#+12345", 1), strings.Replace(valid, "#aBcD01", "#xyzxyz", 1), strings.Replace(valid, ",\"version\":1", "", 1), strings.Replace(valid, "\"version\":1", "\"version\":1,\"version\":1", 1), strings.Replace(valid, "\"version\":1", "\"version\":1,\"extra\":true", 1), strings.Replace(valid, "\"version\":1", "\"version\":\"1\"", 1), strings.Replace(valid, "vellum-placement-fill", "vellum-placement-selection", 1), string([]byte{0xff})}
	m := groupedModel(t)
	before, state := slices.Clone(m.scene.Objects), m.journal.Snapshot()
	for _, text := range bad {
		if _, err := decodeFillStyle(text); err == nil {
			t.Fatalf("invalid style accepted: %q", text)
		}
		if err := m.pasteFillStyle(text); err == nil || !slices.Equal(before, m.scene.Objects) || state != m.journal.Snapshot() {
			t.Fatalf("invalid style partially applied: %q", text)
		}
	}
	if _, err = encodeFillStyle(color.RGBA{R: 10, A: 128}); err == nil {
		t.Fatal("encoded nonopaque style")
	}
	m.prompt = "close"
	if _, err = m.copyFillStyle(); err == nil || m.pasteFillStyle(valid) == nil {
		t.Fatal("style ignored modal guard")
	}
	m.prompt = ""
	m.setSelection(nil, "")
	if _, err = m.copyFillStyle(); err == nil || m.pasteFillStyle(valid) == nil {
		t.Fatal("style ignored empty selection")
	}
}
