package main

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
	"maps"
	"strings"
	"testing"
)

func TestLayerRenameHistoryIdentityRoundTripAndPruning(t *testing.T) {
	m := groupedModel(t)
	before := append([]canvas.Rect(nil), m.scene.Objects...)
	count := m.journal.Snapshot().UndoCount
	if err := m.renameLayer("group-1", "Assembly"); err != nil {
		t.Fatal(err)
	}
	if err := m.renameLayer("block-1", "Housing"); err != nil {
		t.Fatal(err)
	}
	if m.layerLabel("group-1") != "Assembly · group-1" || m.scene.Objects[0] != before[0] || m.journal.Snapshot().UndoCount != count+2 {
		t.Fatal("identity/history")
	}
	f, err := drawingCodec(m).Capture()
	if err != nil || f.Version != 4 {
		t.Fatal(f.Version, err)
	}
	b, err := drawingCodec(m).Encode(f)
	if err != nil {
		t.Fatal(err)
	}
	n, err := newModel()
	if err != nil {
		t.Fatal(err)
	}
	defer n.palette.Close()
	decoded, err := drawingCodec(n).Decode(b)
	if err != nil {
		t.Fatal(err)
	}
	if err = drawingCodec(n).Install(decoded); err != nil || !maps.Equal(n.layerInfo, m.layerInfo) {
		t.Fatal("roundtrip", err)
	}
	if err = m.journal.Undo(); err != nil || m.layerInfo["block-1"].Name != "" {
		t.Fatal("undo", err)
	}
	if err = m.journal.Redo(); err != nil || m.layerInfo["block-1"].Name != "Housing" {
		t.Fatal("redo", err)
	}
	m.setSelection([]string{"block-1"}, "block-1")
	if err = m.deleteSelected(); err != nil {
		t.Fatal(err)
	}
	if len(m.layerInfo) != 0 {
		t.Fatal("deleted metadata retained")
	}
	if err = m.journal.Undo(); err != nil || m.layerInfo["group-1"].Name != "Assembly" {
		t.Fatal("delete undo", err)
	}
}
func TestLayerRenameValidationNoOpAndBusy(t *testing.T) {
	m := editModel(t)
	count := m.journal.Snapshot().UndoCount
	for _, name := range []string{strings.Repeat("a", 129), "bad\nname", string([]byte{0xff})} {
		if m.renameLayer("block-1", name) == nil {
			t.Fatal("bad name accepted")
		}
	}
	if m.renameLayer("missing", "Name") == nil {
		t.Fatal("missing ID")
	}
	if err := m.renameLayer("block-1", ""); err != nil || m.journal.Snapshot().UndoCount != count {
		t.Fatal("noop")
	}
	m.prompt = "rename"
	if m.renameLayer("block-1", "X") == nil {
		t.Fatal("busy accepted")
	}
	m.prompt = ""
	f, err := drawingCodec(m).Capture()
	if err != nil {
		t.Fatal(err)
	}
	f.Layers = map[string]layerMetadata{"missing": {Name: "X"}}
	f.Version = 4
	if _, err = f.scene(); err == nil {
		t.Fatal("unknown metadata")
	}
	f.Layers = map[string]layerMetadata{"block-1": {Name: "X"}}
	f.Version = 3
	if _, err = f.scene(); err == nil {
		t.Fatal("legacy properties accepted")
	}
}

func TestLayerRenameTargetsFollowCanvasAndOpen(t *testing.T) {
	m := groupedModel(t)
	_ = m.selectLayer("group-1")
	if m.layerActionTarget() != "group-1" {
		t.Fatal("group target")
	}
	m.scene.Selected = "connector-2"
	m.managedSelection("connector-2")
	if m.layerActionTarget() != "connector-2" {
		t.Fatal("stale canvas target")
	}
	f, err := drawingCodec(m).Capture()
	if err != nil {
		t.Fatal(err)
	}
	f.Selected = "block-1"
	if err = drawingCodec(m).Install(f); err != nil || m.layerActionTarget() != "block-1" {
		t.Fatal("stale Open target", err)
	}
}
func TestLayerRenameCannotExceedEncodedFileLimit(t *testing.T) {
	m, err := newModel()
	if err != nil {
		t.Fatal(err)
	}
	defer m.palette.Close()
	m.groups = make(map[string]string)
	m.layerInfo = make(map[string]layerMetadata)
	for i := 1; i <= 512; i++ {
		id := fmt.Sprintf("block-%d", i)
		m.scene.Objects = append(m.scene.Objects, canvas.Rect{ID: id, Width: 80, Height: 50})
		m.groups[id] = fmt.Sprintf("group-%d", (i-1)/2+1)
	}
	m.next = 512
	m.nextGroup = 256
	rejected := false
	for id := range liveLayerIDs(m.scene.Objects, m.groups) {
		prior := maps.Clone(m.layerInfo)
		count := m.journal.Snapshot().UndoCount
		err = m.renameLayer(id, strings.Repeat("<", 80))
		if err != nil {
			if !maps.Equal(prior, m.layerInfo) || m.journal.Snapshot().UndoCount != count {
				t.Fatal("failed rename mutated")
			}
			rejected = true
			break
		}
	}
	if !rejected {
		t.Fatal("fixture did not reach encoded capacity")
	}
	f, err := drawingCodec(m).Capture()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = drawingCodec(m).Encode(f); err != nil {
		t.Fatal("accepted state unsaveable", err)
	}
}
func TestStaleRenameDialogCannotClearNewPromptOrSession(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() {
		f.m.prompt = "rename"
		f.m.renameID = "block-1"
		f.m.renameEpoch = 1
		old := f.m.renameDialog(f.a, f.ref)
		f.m.prompt = "close"
		old.Dismiss()
		old.Children[2].Activate()
		if f.m.prompt != "close" || len(f.m.layerInfo) != 0 {
			t.Fatal("stale dialog cleared Close")
		}
		f.m.prompt = "rename"
		f.m.renameEpoch = 2
		f.m.renameDraft = "New"
		old.Children[1].Change(ui.ChangeEvent{Text: "Stale"})
		old.Children[3].Activate()
		if f.m.renameDraft != "New" || f.m.prompt != "rename" {
			t.Fatal("stale rename session")
		}
	})
}
