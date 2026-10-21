package main

import (
	"errors"
	"image/color"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
)

func TestRepeatDuplicatesGroupsOffsetsMetadataAndReplay(t *testing.T) {
	m := groupedModel(t)
	m.scene.Objects[0].Color = color.RGBA{17, 34, 51, 255}
	m.scene.Objects[2].Color = color.RGBA{90, 60, 30, 255}
	m.layerInfo = map[string]layerMetadata{"block-1": {Name: "Housing"}, "note-3": {Name: "Label"}, "group-1": {Name: "Assembly"}}
	m.setSelection([]string{"block-1"}, "block-1")
	before, groups, info := slices.Clone(m.scene.Objects), maps.Clone(m.groups), maps.Clone(m.layerInfo)
	ids, primary := slices.Clone(m.selectedIDs()), m.scene.Selected
	count := m.journal.Snapshot().UndoCount
	if err := m.repeatDuplicates(3, 15.125, -30.25); err != nil {
		t.Fatal(err)
	}
	if len(m.scene.Objects) != len(before)+6 || len(m.selectedIDs()) != 6 || m.journal.Snapshot().UndoCount != count+1 || !slices.Equal(before, m.scene.Objects[:len(before)]) {
		t.Fatal("not one insertion preserving originals", m.scene.Objects)
	}
	newGroups := map[string]bool{}
	for n := 0; n < 3; n++ {
		copyA, copyB := m.scene.Objects[len(before)+2*n], m.scene.Objects[len(before)+2*n+1]
		var a, b canvas.Rect
		for _, o := range before {
			if o.ID == "block-1" {
				a = o
			}
			if o.ID == "note-3" {
				b = o
			}
		}
		if copyA.X != a.X+float64(n+1)*15.125 || copyA.Y != a.Y-float64(n+1)*30.25 || copyB.X-copyA.X != b.X-a.X || copyB.Y-copyA.Y != b.Y-a.Y || copyA.Color != a.Color || copyB.Color != b.Color || copyA.Width != a.Width || copyB.Height != b.Height {
			t.Fatal("copies lost geometry, offsets or colors", n, copyA, copyB)
		}
		g := m.groups[copyA.ID]
		if g == "" || g == "group-1" || newGroups[g] || m.groups[copyB.ID] != g || m.layerInfo[g].Name != "Assembly" || m.layerInfo[copyA.ID].Name != "Housing" || m.layerInfo[copyB.ID].Name != "Label" {
			t.Fatal("groups not independent or metadata lost", g, m.groups, m.layerInfo)
		}
		newGroups[g] = true
	}
	after, afterGroups, afterInfo := slices.Clone(m.scene.Objects), maps.Clone(m.groups), maps.Clone(m.layerInfo)
	newIDs := slices.Clone(m.selectedIDs())
	if err := m.journal.Undo(); err != nil || !slices.Equal(before, m.scene.Objects) || !maps.Equal(groups, m.groups) || !maps.Equal(info, m.layerInfo) || !slices.Equal(ids, m.selectedIDs()) || m.scene.Selected != primary {
		t.Fatal("Undo did not restore document and selection", err)
	}
	if err := m.journal.Redo(); err != nil || !slices.Equal(after, m.scene.Objects) || !maps.Equal(afterGroups, m.groups) || !maps.Equal(afterInfo, m.layerInfo) || !slices.Equal(newIDs, m.selectedIDs()) {
		t.Fatal("Redo lost identities", err)
	}
	_ = m.journal.Undo()
	if err := m.repeatDuplicates(1, 0, 0); err != nil {
		t.Fatal("intentional overlap rejected", err)
	}
	for _, id := range m.selectedIDs() {
		if slices.Contains(newIDs, id) {
			t.Fatal("branch reused component identity", id)
		}
	}
	for _, id := range m.selectedIDs() {
		if newGroups[m.groups[id]] {
			t.Fatal("branch reused group identity", m.groups[id])
		}
	}
}

func TestRepeatDuplicatesClosedWindowCallbacksAreInert(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	var dialog ui.Node
	f.owner(t, func() {
		if err := f.m.beginRepeatDuplicates(); err != nil {
			t.Fatal(err)
		}
		dialog = f.m.repeatDuplicatesDialog(f.a, f.ref)
	})
	if err := f.a.CloseWindow(f.ref); err != nil {
		t.Fatal(err)
	}
	before, state := slices.Clone(f.m.scene.Objects), f.m.journal.Snapshot()
	func() {
		geometryDialogControl(t, dialog, "repeat-count").Change(ui.ChangeEvent{Text: "10"})
		geometryDialogControl(t, dialog, "repeat-apply").Activate()
		dialog.Dismiss()
	}()
	if !slices.Equal(before, f.m.scene.Objects) || state != f.m.journal.Snapshot() {
		t.Fatal("retired window callback mutated drawing")
	}
}

func TestRepeatDuplicatesValidationIsAtomic(t *testing.T) {
	for _, tc := range []struct {
		count  int
		dx, dy float64
	}{{0, 1, 1}, {101, 1, 1}, {1, math.NaN(), 1}, {1, 1, math.Inf(1)}, {1, 20001, 1}, {1, 1, -20001}, {100, 100, 1}, {1, 10000, 10000}} {
		m := groupedModel(t)
		before, groups, info, state := slices.Clone(m.scene.Objects), maps.Clone(m.groups), maps.Clone(m.layerInfo), m.journal.Snapshot()
		next, nextGroup := m.next, m.nextGroup
		if err := m.repeatDuplicates(tc.count, tc.dx, tc.dy); err == nil || !slices.Equal(before, m.scene.Objects) || !maps.Equal(groups, m.groups) || !maps.Equal(info, m.layerInfo) || state != m.journal.Snapshot() || m.next != next || m.nextGroup != nextGroup {
			t.Fatal("invalid repeat not atomic", tc, err)
		}
	}
	m := groupedModel(t)
	m.rejectNotes = true
	before, state := slices.Clone(m.scene.Objects), m.journal.Snapshot()
	if m.repeatDuplicates(2, 20, 30) == nil || !slices.Equal(before, m.scene.Objects) || state != m.journal.Snapshot() {
		t.Fatal("rejected member partially inserted")
	}
	m.rejectNotes = false
	m.next = math.MaxUint64 - 1
	if m.repeatDuplicates(1, 20, 20) == nil || !slices.Equal(before, m.scene.Objects) {
		t.Fatal("ID exhaustion partially inserted")
	}
}

func TestRepeatDuplicatesCapacityAndDialogLifecycle(t *testing.T) {
	m := groupedModel(t)
	m.setSelection([]string{"block-1"}, "block-1")
	for len(m.scene.Objects) < 511 {
		i := len(m.scene.Objects)
		o := m.scene.Objects[1]
		o.ID = "connector-" + strconv.Itoa(i+1)
		m.scene.Objects = append(m.scene.Objects, o)
		m.next = uint64(i + 1)
	}
	before, state := slices.Clone(m.scene.Objects), m.journal.Snapshot()
	if m.repeatDuplicates(1, 0, 0) == nil || !slices.Equal(before, m.scene.Objects) || state != m.journal.Snapshot() {
		t.Fatal("capacity partially inserted")
	}
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() {
		if err := f.m.beginRepeatDuplicates(); err != nil {
			t.Fatal(err)
		}
		dialog := f.m.repeatDuplicatesDialog(f.a, f.ref)
		geometryDialogControl(t, dialog, "repeat-count").Change(ui.ChangeEvent{Text: strings.Repeat("9", 65)})
		geometryDialogControl(t, dialog, "repeat-apply").Activate()
		if !f.m.repeatDraft.invalidInput[0] || f.m.repeatDraft.message == "" || len(f.m.scene.Objects) != 1 {
			t.Fatal("oversized count changed drawing")
		}
		geometryDialogControl(t, dialog, "repeat-count").Change(ui.ChangeEvent{Text: "1.5"})
		geometryDialogControl(t, dialog, "repeat-apply").Activate()
		if f.m.prompt != "repeat-duplicates" || f.m.repeatDraft.message == "" || len(f.m.scene.Objects) != 1 {
			t.Fatal("invalid count not visible/atomic")
		}
		geometryDialogControl(t, dialog, "repeat-cancel").Activate()
		f.m.prompt = "close"
		geometryDialogControl(t, dialog, "repeat-apply").Activate()
		dialog.Dismiss()
		if f.m.prompt != "close" || len(f.m.scene.Objects) != 1 {
			t.Fatal("old dialog changed replacement prompt")
		}
		f.m.prompt = ""
		if err := f.m.beginRepeatDuplicates(); err != nil {
			t.Fatal(err)
		}
		d := f.m.repeatDraft
		f.m.prompt = ""
		f.m.scene.Objects[0].X++
		if err := f.m.applyRepeatDuplicates(d); !errors.Is(err, history.ErrStale) || len(f.m.scene.Objects) != 1 {
			t.Fatal("stale snapshot accepted", err)
		}
	})
}

func TestRepeatDuplicatesCorrectiveEditRetainsOtherInvalidFields(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() {
		_ = f.m.beginRepeatDuplicates()
		dialog := f.m.repeatDuplicatesDialog(f.a, f.ref)
		d := f.m.repeatDraft
		geometryDialogControl(t, dialog, "repeat-count").Change(ui.ChangeEvent{Text: strings.Repeat("9", 65)})
		geometryDialogControl(t, dialog, "repeat-dx").Change(ui.ChangeEvent{Text: strings.Repeat("9", 65)})
		if d.count != "3" || d.dx != "20" || !d.invalidInput[0] || !d.invalidInput[1] {
			t.Fatal("accepted text lost")
		}
		geometryDialogControl(t, dialog, "repeat-count").Change(ui.ChangeEvent{Text: "2"})
		geometryDialogControl(t, dialog, "repeat-apply").Activate()
		if !d.invalidInput[1] || f.m.prompt != "repeat-duplicates" || len(f.m.scene.Objects) != 1 {
			t.Fatal("other invalidfield bypassed")
		}
		geometryDialogControl(t, dialog, "repeat-dx").Change(ui.ChangeEvent{Text: "30"})
		geometryDialogControl(t, dialog, "repeat-apply").Activate()
		if f.m.repeatDraft != nil || f.m.prompt != "" || len(f.m.scene.Objects) != 3 || f.m.scene.Objects[1].X != 30 || f.m.scene.Objects[2].X != 60 {
			t.Fatal("corrected fields did not recover")
		}
	})
}
