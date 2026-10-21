package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/document"
)

func TestClipboardNamedGroupDuplicateAndPasteMetadata(t *testing.T) {
	m := groupedModel(t)
	for id, name := range map[string]string{"block-1": "Housing", "note-3": "Label", "group-1": "Assembly"} {
		if err := m.renameLayer(id, name); err != nil {
			t.Fatal(err)
		}
	}
	m.setSelection([]string{"block-1"}, "block-1")
	objects, err := m.copiedSelection()
	if err != nil {
		t.Fatal(err)
	}
	text, err := encodeSelection(objects)
	if err != nil {
		t.Fatal(err)
	}
	var wire selectionClipboard
	if err = json.Unmarshal([]byte(text), &wire); err != nil || wire.Version != 2 || len(wire.Objects) != 2 || wire.Objects[0].Name != "Housing" || wire.Objects[1].Name != "Label" || wire.Objects[0].GroupName != "Assembly" {
		t.Fatal("named Copy payload", err, wire)
	}
	for _, action := range []string{"paste", "duplicate"} {
		t.Run(action, func(t *testing.T) {
			before, info := slices.Clone(m.scene.Objects), maps.Clone(m.layerInfo)
			if action == "paste" {
				err = m.pasteSelection(text)
			} else {
				err = m.duplicateSelected()
			}
			if err != nil {
				t.Fatal(err)
			}
			ids := m.selectedIDs()
			if len(ids) != 2 || m.layerInfo[ids[0]].Name != "Housing" || m.layerInfo[ids[1]].Name != "Label" || m.groups[ids[0]] == "group-1" || m.layerInfo[m.groups[ids[0]]].Name != "Assembly" {
				t.Fatal("metadata did not follow fresh identities", ids, m.layerInfo)
			}
			if err = m.journal.Undo(); err != nil || !slices.Equal(before, m.scene.Objects) || !maps.Equal(info, m.layerInfo) {
				t.Fatal("metadata Undo", err)
			}
		})
	}
}

func TestClipboardMetadataVersionValidationAndVisibilityLockPreservation(t *testing.T) {
	objects := []copiedComponent{
		{Kind: "block", Width: 80, Height: 50, Group: 1, Name: "Housing", GroupName: "Assembly", GroupHidden: true, GroupLocked: true},
		{Kind: "note", Width: 90, Height: 65, Group: 1, Name: "Label", Hidden: true, Locked: true, GroupName: "Assembly", GroupHidden: true, GroupLocked: true},
	}
	encode := func(v int, objects []copiedComponent) string {
		b, e := json.Marshal(selectionClipboard{"vellum-placement-selection", v, objects})
		if e != nil {
			t.Fatal(e)
		}
		return string(b)
	}
	text := encode(2, objects)
	decoded, err := decodeSelection(text)
	if err != nil || !slices.Equal(decoded, objects) {
		t.Fatal("metadata decode", err)
	}
	n, err := newModel()
	if err != nil {
		t.Fatal(err)
	}
	defer n.palette.Close()
	if err = n.pasteSelection(text); err != nil {
		t.Fatal(err)
	}
	g := n.groups["block-1"]
	if n.layerInfo["block-1"] != objects[0].componentMetadata() || n.layerInfo["note-2"] != objects[1].componentMetadata() || n.layerInfo[g] != objects[0].groupMetadata() || len(n.selectedIDs()) != 0 || n.scene.Selected != "" {
		t.Fatal("flags or filtered selection were lost", n.layerInfo)
	}
	f, err := drawingCodec(n).Capture()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = drawingCodec(n).Encode(f); err != nil {
		t.Fatal("metadata drawing is not saveable", err)
	}
	legacy := []copiedComponent{{Kind: "block", Width: 80, Height: 50}}
	if _, err = decodeSelection(encode(1, legacy)); err != nil {
		t.Fatal("legacy clipboard refused", err)
	}
	for _, mode := range []string{"legacy-metadata", "future-version", "long-name", "control-name", "trimmed-name", "inconsistent-group", "orphan-group", "incomplete-group"} {
		t.Run(mode, func(t *testing.T) {
			bad := slices.Clone(objects)
			version := 2
			switch mode {
			case "legacy-metadata":
				version = 1
			case "future-version":
				version = 4
			case "long-name":
				bad[0].Name = strings.Repeat("a", 129)
			case "control-name":
				bad[0].Name = "bad\nname"
			case "trimmed-name":
				bad[0].Name = " Name "
			case "inconsistent-group":
				bad[1].GroupName = "Other"
			case "orphan-group":
				bad[0].Group = 0
			case "incomplete-group":
				bad = bad[:1]
			}
			if _, err := decodeSelection(encode(version, bad)); err == nil {
				t.Fatal("invalid clipboard metadata accepted")
			}
		})
	}
	badUTF8 := strings.Replace(text, "Housing", string([]byte{0xff}), 1)
	if _, err := decodeSelection(badUTF8); err == nil {
		t.Fatal("invalid UTF8 clipboard accepted")
	}
}

func TestClipboardInsertionChecksActualEncodedDestinationBeforeMutation(t *testing.T) {
	m, err := newModel()
	if err != nil {
		t.Fatal(err)
	}
	defer m.palette.Close()
	m.groups = make(map[string]string)
	m.layerInfo = make(map[string]layerMetadata)
	ids := make([]string, 0, 200)
	for i := 1; i <= 200; i++ {
		id, group := fmt.Sprintf("block-%d", i), fmt.Sprintf("group-%d", (i-1)/2+1)
		m.scene.Objects = append(m.scene.Objects, canvas.Rect{ID: id, Width: 80, Height: 50})
		m.groups[id] = group
		ids = append(ids, id)
		m.layerInfo[id] = layerMetadata{Name: strings.Repeat("<", 80)}
		m.layerInfo[group] = layerMetadata{Name: strings.Repeat("<", 80)}
	}
	m.next, m.nextGroup = 200, 100
	m.setSelection(ids, ids[len(ids)-1])
	f, err := drawingCodec(m).Capture()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = drawingCodec(m).Encode(f); err != nil {
		t.Fatal("initial drawing must fit", err)
	}
	objects, err := m.copiedSelection()
	if err != nil {
		t.Fatal(err)
	}
	text, err := encodeSelection(objects)
	if err != nil {
		t.Fatal("clipboard must fit", err)
	}
	before, groups, info, state := slices.Clone(m.scene.Objects), maps.Clone(m.groups), maps.Clone(m.layerInfo), m.journal.Snapshot()
	for _, action := range []string{"paste", "duplicate"} {
		if action == "paste" {
			err = m.pasteSelection(text)
		} else {
			err = m.duplicateSelected()
		}
		if !errors.Is(err, document.ErrTooLarge) || !slices.Equal(before, m.scene.Objects) || !maps.Equal(groups, m.groups) || !maps.Equal(info, m.layerInfo) || m.next != 200 || m.nextGroup != 100 || m.journal.Snapshot() != state || m.pasteText != "" || m.pasteCount != 0 {
			t.Fatal("oversized destination partially inserted", action, err)
		}
	}
}

func TestClipboardProspectiveDrawingUsesProposedIDCounters(t *testing.T) {
	m, err := newModel()
	if err != nil {
		t.Fatal(err)
	}
	defer m.palette.Close()
	m.next, m.nextGroup = math.MaxUint64-2, math.MaxUint64-1
	objects := []copiedComponent{{Kind: "block", Width: 80, Height: 50, Group: 1, Name: "A", GroupName: "Pair"}, {Kind: "note", Width: 90, Height: 65, Group: 1, Name: "B", GroupName: "Pair"}}
	if err = m.insertCopies(objects, 20, "Paste"); err != nil {
		t.Fatal(err)
	}
	if m.next != math.MaxUint64 || m.nextGroup != math.MaxUint64 {
		t.Fatal("proposed counters not accepted")
	}
	f, err := drawingCodec(m).Capture()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = drawingCodec(m).Encode(f); err != nil {
		t.Fatal("last representable IDs not saveable", err)
	}
}
