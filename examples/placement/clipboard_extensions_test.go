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
