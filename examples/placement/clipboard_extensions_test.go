package main

import (
	"context"
	"image/color"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/history"
)

type extendedClipboardRead struct {
	appClipboard
	started, release chan struct{}
	ctx              context.Context
}

func newExtendedClipboardRead(text string) *extendedClipboardRead {
	return &extendedClipboardRead{appClipboard: appClipboard{text: text}, started: make(chan struct{}), release: make(chan struct{})}
}
func (s *extendedClipboardRead) ReadClipboard(ctx context.Context) (string, error) {
	s.ctx = ctx
	close(s.started)
	select {
	case <-s.release:
		return s.text, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
func fillClipboardFixture(t *testing.T) *fileFixture {
	t.Helper()
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.place(t, "note", canvas.Point{X: 150, Y: 50})
	f.owner(t, func() {
		if err := f.m.bindClipboard(f.a, f.ref); err != nil {
			t.Fatal(err)
		}
		f.m.setSelection([]string{"block-1", "note-2"}, "note-2")
		f.m.pasteText, f.m.pasteCount = "existing component payload", 9
	})
	return f
}
func TestFillClipboardWorkerCopyPasteAndCounters(t *testing.T) {
	f := fillClipboardFixture(t)
	s := &appClipboard{}
	f.a.SetServices(s)
	var value color.RGBA
	var before []canvas.Rect
	var prior history.State
	f.owner(t, func() {
		value = f.m.scene.Objects[1].Color
		before, prior = slices.Clone(f.m.scene.Objects), f.m.journal.Snapshot()
		if err := f.m.commands.Invoke(f.ref.Context(), "copy-fill"); err != nil {
			t.Fatal(err)
		}
	})
	waitClipboard(t, f)
	s.mu.Lock()
	text := s.text
	s.mu.Unlock()
	got, err := decodeFillStyle(text)
	if err != nil || got != value {
		t.Fatal("native Copy Fill payload", got, err)
	}
	f.owner(t, func() {
		if !slices.Equal(before, f.m.scene.Objects) || prior != f.m.journal.Snapshot() || f.m.pasteText != "existing component payload" || f.m.pasteCount != 9 {
			t.Fatal("Copy Fill mutated document or component paste sequence")
		}
	})
	payload, _ := encodeFillStyle(color.RGBA{R: 31, G: 122, B: 207, A: 255})
	read := newExtendedClipboardRead(payload)
	f.a.SetServices(read)
	f.owner(t, func() {
		if err := f.m.commands.Invoke(f.ref.Context(), "paste-fill"); err != nil {
			t.Fatal(err)
		}
	})
	cutAwait(t, read.started)
	close(read.release)
	cutAwaitReady(t, f)
	if read.ctx.Err() != nil {
		t.Fatal("fill read context retired before owner admission")
	}
	waitClipboard(t, f)
	f.owner(t, func() {
		after := slices.Clone(f.m.scene.Objects)
		if after[0].Color != after[1].Color || after[0].Color != (color.RGBA{R: 31, G: 122, B: 207, A: 255}) || f.m.journal.Snapshot().UndoCount != prior.UndoCount+1 || f.m.journal.Snapshot().UndoLabel != "Paste fill style" || f.m.pasteText != "existing component payload" || f.m.pasteCount != 9 {
			t.Fatal("worker style paste did not commit once or changed paste sequence")
		}
		if err := f.m.journal.Undo(); err != nil || !slices.Equal(before, f.m.scene.Objects) {
			t.Fatal("worker Undo", err)
		}
		if err := f.m.journal.Redo(); err != nil || !slices.Equal(after, f.m.scene.Objects) {
			t.Fatal("worker Redo", err)
		}
	})
	if read.ctx.Err() == nil {
		t.Fatal("completed style read retained context")
	}
}
func TestFillClipboardQueuedReadRejectsStaleAndCanceledAdmission(t *testing.T) {
	for _, mode := range []string{"scene", "journal", "revision", "selection", "primary", "document", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			f := fillClipboardFixture(t)
			payload, _ := encodeFillStyle(color.RGBA{R: 31, G: 122, B: 207, A: 255})
			read := newExtendedClipboardRead(payload)
			f.a.SetServices(read)
			f.owner(t, func() {
				if err := f.m.commands.Invoke(f.ref.Context(), "paste-fill"); err != nil {
					t.Fatal(err)
				}
			})
			cutAwait(t, read.started)
			close(read.release)
			cutAwaitReady(t, f)
			var before []canvas.Rect
			var prior history.State
			var groups map[string]string
			var info map[string]layerMetadata
			var ids []string
			var primary string
			f.owner(t, func() {
				switch mode {
				case "scene":
					scene := *f.m.scene
					f.m.scene = &scene
				case "journal":
					f.m.journal, _ = history.New(100)
				case "revision":
					if err := f.m.journal.Reset(true); err != nil {
						t.Fatal(err)
					}
				case "selection":
					f.m.setSelection([]string{"block-1"}, "block-1")
				case "primary":
					f.m.setSelection([]string{"block-1", "note-2"}, "block-1")
				case "document":
					f.m.layerInfo = map[string]layerMetadata{"block-1": {Name: "Changed externally"}}
				case "cancel":
					f.m.clipboardJob.cancel()
				}
				before, prior = slices.Clone(f.m.scene.Objects), f.m.journal.Snapshot()
				groups, info = maps.Clone(f.m.groups), maps.Clone(f.m.layerInfo)
				ids, primary = f.m.selectedIDs(), f.m.scene.Selected
			})
			waitClipboard(t, f)
			f.owner(t, func() {
				if !slices.Equal(before, f.m.scene.Objects) || prior != f.m.journal.Snapshot() || !maps.Equal(groups, f.m.groups) || !reflect.DeepEqual(info, f.m.layerInfo) || !slices.Equal(ids, f.m.selectedIDs()) || primary != f.m.scene.Selected || f.m.pasteCount != 9 || f.m.pasteText != "existing component payload" {
					t.Fatal("refused read changed current state", f.m.status)
				}
				want := history.ErrStale.Error()
				if mode == "cancel" {
					want = context.Canceled.Error()
				}
				if !strings.Contains(f.m.status, want) {
					t.Fatal("admission failure not reported", f.m.status)
				}
			})
		})
	}
}

func TestPasteInPlaceExactGeometryGroupsHistoryAndOffsetSequence(t *testing.T) {
	m := groupedModel(t)
	m.layerInfo = map[string]layerMetadata{"group-1": {Name: "Assembly"}, "block-1": {Name: "Housing"}}
	objects, err := m.copiedSelection()
	if err != nil {
		t.Fatal(err)
	}
	text, err := encodeSelection(objects)
	if err != nil {
		t.Fatal(err)
	}
	before, groups, info := slices.Clone(m.scene.Objects), maps.Clone(m.groups), maps.Clone(m.layerInfo)
	ids, primary, state := m.selectedIDs(), m.scene.Selected, m.journal.Snapshot()
	m.scene.View.Center = canvas.Point{X: -314.5, Y: 271.25}
	m.scene.View.Zoom = 3.5
	m.snap = true
	m.gridSpacing = 100
	m.pasteText, m.pasteCount = text, 9
	if err := m.pasteSelectionInPlace(text); err != nil {
		t.Fatal(err)
	}
	after := slices.Clone(m.scene.Objects)
	freshIDs := m.selectedIDs()
	freshGroup := m.groups[freshIDs[0]]
	if len(freshIDs) != 2 || freshIDs[0] == ids[0] || freshGroup == "" || freshGroup == groups[ids[0]] || freshGroup != m.groups[freshIDs[1]] || m.layerInfo[freshGroup].Name != "Assembly" || m.layerInfo[freshIDs[0]].Name != "Housing" {
		t.Fatal("fresh group identity/metadata", freshIDs, m.groups, m.layerInfo)
	}
	for i, o := range objects {
		got := after[len(before)+i]
		fill, _ := componentFill(o.Kind, o.Color)
		if got.X != o.X || got.Y != o.Y || got.Width != o.Width || got.Height != o.Height || got.Color != fill {
			t.Fatal("paste in place changed geometry or color", got, o)
		}
	}
	if !slices.Equal(before, after[:len(before)]) || m.journal.Snapshot().UndoCount != state.UndoCount+1 || m.journal.Snapshot().UndoLabel != "Paste in place" || m.pasteText != text || m.pasteCount != 9 {
		t.Fatal("changed existing components/history/paste sequence")
	}
	if err := m.journal.Undo(); err != nil || !slices.Equal(before, m.scene.Objects) || !maps.Equal(groups, m.groups) || !maps.Equal(info, m.layerInfo) || !slices.Equal(ids, m.selectedIDs()) || m.scene.Selected != primary {
		t.Fatal("Undo in place lost original document/selection", err)
	}
	if err := m.journal.Redo(); err != nil || !slices.Equal(after, m.scene.Objects) || !slices.Equal(freshIDs, m.selectedIDs()) || m.groups[freshIDs[0]] != freshGroup {
		t.Fatal("Redo in place changed identity", err)
	}
	if err := m.pasteSelection(text); err != nil {
		t.Fatal(err)
	}
	latest := m.scene.Objects[len(after):]
	if m.pasteCount != 10 || latest[0].X != objects[0].X+200 || latest[0].Y != objects[0].Y+200 {
		t.Fatal("in-place paste advanced ordinary offset sequence", m.pasteCount, latest)
	}
}

func TestPasteInPlaceClipboardWorkerAllowsEmptyDrawingAndKeepsContext(t *testing.T) {
	f := fileModel(t, false)
	text, err := encodeSelection([]copiedComponent{{Kind: "block", X: -63.25, Y: 14.75, Width: 50, Height: 60}})
	if err != nil {
		t.Fatal(err)
	}
	read := newExtendedClipboardRead(text)
	f.a.SetServices(read)
	f.owner(t, func() {
		if err := f.m.bindClipboard(f.a, f.ref); err != nil {
			t.Fatal(err)
		}
		if err := f.m.commands.Enabled("copy-fill"); err == nil {
			t.Fatal("empty drawing enabled Copy Fill")
		}
		if err := f.m.commands.Enabled("paste-fill"); err == nil {
			t.Fatal("empty drawing enabled Paste Fill")
		}
		f.m.pasteText, f.m.pasteCount = "unchanged", 11
		if err := f.m.commands.Invoke(f.ref.Context(), "paste-in-place"); err != nil {
			t.Fatal(err)
		}
	})
	cutAwait(t, read.started)
	close(read.release)
	cutAwaitReady(t, f)
	if read.ctx.Err() != nil {
		t.Fatal("in-place read context retired before owner admission")
	}
	waitClipboard(t, f)
	f.owner(t, func() {
		objects := f.m.scene.Objects
		if len(objects) != 1 || objects[0].X != -63.25 || objects[0].Y != 14.75 || objects[0].ID != "block-1" || f.m.journal.Snapshot().UndoCount != 1 || f.m.pasteText != "unchanged" || f.m.pasteCount != 11 {
			t.Fatal("empty drawing in-place native paste", objects)
		}
		if err := f.m.journal.Undo(); err != nil || len(f.m.scene.Objects) != 0 {
			t.Fatal("empty drawing Undo", err)
		}
		if err := f.m.journal.Redo(); err != nil || !slices.Equal(objects, f.m.scene.Objects) {
			t.Fatal("empty drawing exact Redo", err)
		}
	})
	if read.ctx.Err() == nil {
		t.Fatal("completed in-place read retained context")
	}
}

func TestPasteInPlaceClipboardRejectsStaleAndCanceledReads(t *testing.T) {
	for _, mode := range []string{"scene", "journal", "revision", "selection", "primary", "document", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			f := fillClipboardFixture(t)
			var payload string
			f.owner(t, func() { payload = copiedText(t, f.m) })
			read := newExtendedClipboardRead(payload)
			f.a.SetServices(read)
			f.owner(t, func() {
				if err := f.m.commands.Invoke(f.ref.Context(), "paste-in-place"); err != nil {
					t.Fatal(err)
				}
			})
			cutAwait(t, read.started)
			close(read.release)
			cutAwaitReady(t, f)
			var before []canvas.Rect
			var prior history.State
			var next, nextGroup uint64
			var ids []string
			var primary string
			f.owner(t, func() {
				switch mode {
				case "scene":
					scene := *f.m.scene
					f.m.scene = &scene
				case "journal":
					f.m.journal, _ = history.New(100)
				case "revision":
					if err := f.m.journal.Reset(true); err != nil {
						t.Fatal(err)
					}
				case "selection":
					f.m.setSelection(nil, "")
				case "primary":
					f.m.setSelection([]string{"block-1", "note-2"}, "block-1")
				case "document":
					f.m.scene.Objects[0].Width++
				case "cancel":
					f.m.clipboardJob.cancel()
				}
				before, prior = slices.Clone(f.m.scene.Objects), f.m.journal.Snapshot()
				next, nextGroup = f.m.next, f.m.nextGroup
				ids, primary = f.m.selectedIDs(), f.m.scene.Selected
			})
			waitClipboard(t, f)
			f.owner(t, func() {
				if !slices.Equal(before, f.m.scene.Objects) || prior != f.m.journal.Snapshot() || next != f.m.next || nextGroup != f.m.nextGroup || !slices.Equal(ids, f.m.selectedIDs()) || primary != f.m.scene.Selected || f.m.pasteCount != 9 || f.m.pasteText != "existing component payload" {
					t.Fatal("refused in-place read mutated state", f.m.status)
				}
				want := history.ErrStale.Error()
				if mode == "cancel" {
					want = context.Canceled.Error()
				}
				if !strings.Contains(f.m.status, want) {
					t.Fatal("admission failure not reported", f.m.status)
				}
			})
		})
	}
}

func TestPasteInPlaceRejectsWithoutMutation(t *testing.T) {
	m := groupedModel(t)
	text := copiedText(t, m)
	before, state, next, nextGroup := slices.Clone(m.scene.Objects), m.journal.Snapshot(), m.next, m.nextGroup
	m.pasteText, m.pasteCount = "unchanged", 4
	for _, mode := range []string{"payload", "capacity", "locked", "prompt"} {
		switch mode {
		case "payload":
			if m.pasteSelectionInPlace("plain text") == nil {
				t.Fatal("unrelated clipboard accepted")
			}
		case "capacity":
			many := make([]copiedComponent, 512)
			for i := range many {
				many[i] = copiedComponent{Kind: "block", Width: 50, Height: 50}
			}
			large, err := encodeSelection(many)
			if err != nil {
				t.Fatal(err)
			}
			if m.pasteSelectionInPlace(large) == nil {
				t.Fatal("capacity accepted")
			}
		case "locked":
			m.locked = true
			if m.pasteSelectionInPlace(text) == nil {
				t.Fatal("placement rejection ignored")
			}
			m.locked = false
		case "prompt":
			m.prompt = "close"
			if m.pasteSelectionInPlace(text) == nil {
				t.Fatal("modal gate ignored")
			}
			m.prompt = ""
		}
	}
	if !slices.Equal(before, m.scene.Objects) || state != m.journal.Snapshot() || next != m.next || nextGroup != m.nextGroup || m.pasteText != "unchanged" || m.pasteCount != 4 {
		t.Fatal("refused in-place paste partially mutated")
	}
}

func TestPasteInPlaceCommandModalRefusesBeforeNativeRead(t *testing.T) {
	f := fileModel(t, false)
	read := newExtendedClipboardRead("not read")
	f.a.SetServices(read)
	f.owner(t, func() {
		if err := f.m.bindClipboard(f.a, f.ref); err != nil {
			t.Fatal(err)
		}
		f.m.prompt = "fill-color"
		if f.m.commands.Enabled("paste-in-place") == nil || f.m.commands.Invoke(f.ref.Context(), "paste-in-place") == nil || f.m.clipboardJob != nil {
			t.Fatal("modal command admitted")
		}
		select {
		case <-read.started:
			t.Fatal("modal command read native clipboard")
		default:
		}
		f.m.prompt = ""
	})
}
