package main

import (
	"context"
	"errors"
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
)

// The write is held before accepting bytes so tests can distinguish clipboard
// success, owner admission and document mutation without scheduling sleeps.
type cutClipboard struct {
	appClipboard
	started, release, returned chan struct{}
	writeErr                   error
	ignoreCancellation         bool
}

func newCutClipboard() *cutClipboard {
	return &cutClipboard{started: make(chan struct{}), release: make(chan struct{}), returned: make(chan struct{})}
}
func (s *cutClipboard) WriteClipboard(ctx context.Context, text string) error {
	close(s.started)
	defer close(s.returned)
	if s.ignoreCancellation {
		<-s.release
	} else {
		select {
		case <-s.release:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if s.writeErr != nil {
		return s.writeErr
	}
	s.mu.Lock()
	s.text = text
	s.mu.Unlock()
	return nil
}
func cutAwait(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Cut worker did not reach the expected boundary")
	}
}
func cutAwaitReady(t *testing.T, f *fileFixture) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if f.a.Pending().ReadyCompletions == 1 {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("Cut completion was not queued")
}
func cutFixture(t *testing.T) (*fileFixture, *cutClipboard) {
	t.Helper()
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.place(t, "note", canvas.Point{X: 150, Y: 50})
	s := newCutClipboard()
	f.a.SetServices(s)
	f.owner(t, func() {
		if err := f.m.bindClipboard(f.a, f.ref); err != nil {
			t.Fatal(err)
		}
		f.m.setSelection([]string{"block-1", "note-2"}, "note-2")
		if err := f.m.journal.MarkSaved(); err != nil {
			t.Fatal(err)
		}
	})
	return f, s
}

func TestCutWritesBeforeOneUndoableDeletionAndUsesAfterFocusShortcut(t *testing.T) {
	f, s := cutFixture(t)
	f.place(t, "connector", canvas.Point{X: 350, Y: 100})
	var before []canvas.Rect
	var groups map[string]string
	var info map[string]layerMetadata
	var selected []string
	var prior history.State
	var payload string
	f.owner(t, func() {
		f.m.setSelection([]string{"block-1", "note-2"}, "note-2")
		if err := f.m.changeGrouping(false); err != nil {
			t.Fatal(err)
		}
		for id, name := range map[string]string{"block-1": "Housing", "note-2": "Label", "connector-3": "Untouched", "group-1": "Assembly"} {
			if err := f.m.renameLayer(id, name); err != nil {
				t.Fatal(err)
			}
		}
		if err := f.m.journal.MarkSaved(); err != nil {
			t.Fatal(err)
		}
		before, groups, selected = slices.Clone(f.m.scene.Objects), maps.Clone(f.m.groups), f.m.selectedIDs()
		info, prior = maps.Clone(f.m.layerInfo), f.m.journal.Snapshot()
		objects, err := f.m.copiedSelection()
		if err != nil {
			t.Fatal(err)
		}
		payload, err = encodeSelection(objects)
		if err != nil {
			t.Fatal(err)
		}
		f.m.pasteText, f.m.pasteCount = "older clipboard", 7
		if handled, err := f.m.commands.ShortcutInScope(f.ref.Context(), "Meta+X", command.WindowShortcuts); err != nil || handled {
			t.Fatal("Cut intercepted the pre-focus shortcut phase", handled, err)
		}
		if handled, err := f.m.commands.ShortcutInScope(f.ref.Context(), "Meta+X", command.AfterFocusShortcuts); err != nil || !handled {
			t.Fatal("Cut after-focus shortcut was unavailable", handled, err)
		}
		if f.m.commands.Enabled("cut") == nil || f.m.editAvailable() == nil {
			t.Fatal("Cut did not gate concurrent editing")
		}
	})
	cutAwait(t, s.started)
	f.owner(t, func() {
		if !slices.Equal(f.m.scene.Objects, before) || f.m.journal.Snapshot() != prior {
			t.Fatal("Cut removed objects before clipboard success")
		}
	})
	close(s.release)
	cutAwaitReady(t, f)
	f.owner(t, func() {
		if !slices.Equal(f.m.scene.Objects, before) {
			t.Fatal("clipboard worker mutated the owner-held drawing")
		}
	})
	waitClipboard(t, f)
	s.mu.Lock()
	text := s.text
	s.mu.Unlock()
	if text != payload {
		t.Fatal("Cut did not write the frozen Copy payload")
	}
	f.owner(t, func() {
		state := f.m.journal.Snapshot()
		if len(f.m.scene.Objects) != 1 || f.m.scene.Objects[0] != before[2] || len(f.m.groups) != 0 || f.m.scene.Selected != "" || len(f.m.selectedIDs()) != 0 ||
			state.UndoCount != prior.UndoCount+1 || state.UndoLabel != "Cut selection" || !state.Dirty || f.m.next != 3 || f.m.nextGroup != 1 || f.m.pasteCount != 0 || f.m.pasteText != "" ||
			!maps.Equal(f.m.layerInfo, map[string]layerMetadata{"connector-3": info["connector-3"]}) {
			t.Fatal("Cut was not one complete deletion", state)
		}
		if err := f.m.journal.Undo(); err != nil || !slices.Equal(f.m.scene.Objects, before) || !maps.Equal(f.m.groups, groups) ||
			!maps.Equal(f.m.layerInfo, info) || !slices.Equal(f.m.selectedIDs(), selected) || f.m.scene.Selected != "note-2" || f.m.journal.Snapshot().Dirty {
			t.Fatal("Cut Undo lost identity, order, groups, selection or saved state", err)
		}
		if err := f.m.journal.Redo(); err != nil || len(f.m.scene.Objects) != 1 || f.m.scene.Objects[0] != before[2] {
			t.Fatal("Cut Redo", err)
		}
		if err := f.m.pasteSelection(text); err != nil || len(f.m.scene.Objects) != 3 || f.m.scene.Objects[0] != before[2] || f.m.scene.Objects[1].ID != "block-4" || f.m.scene.Objects[2].ID != "note-5" ||
			f.m.scene.Objects[1].X != before[0].X+20 || f.m.scene.Objects[2].Y != before[1].Y+20 || f.m.groups["block-4"] == "" || f.m.groups["block-4"] != f.m.groups["note-5"] || f.m.groups["block-4"] == groups["block-1"] {
			t.Fatal("Cut clipboard did not paste independent grouped objects", err)
		}
		if f.m.layerInfo["block-4"].Name != "Housing" || f.m.layerInfo["note-5"].Name != "Label" || f.m.layerInfo[f.m.groups["block-4"]].Name != "Assembly" || f.m.layerInfo["connector-3"].Name != "Untouched" {
			t.Fatal("Cut/Paste lost component, group or surviving metadata")
		}
	})
}

func TestCutWriteFailureCancellationAndCloseKeepDrawing(t *testing.T) {
	for _, mode := range []string{"write-failure", "cancel-writing", "cancel-queued", "request-close", "window-close"} {
		t.Run(mode, func(t *testing.T) {
			f, s := cutFixture(t)
			if mode == "write-failure" {
				s.writeErr = errors.New("write refused")
			}
			if mode == "request-close" {
				s.ignoreCancellation = true // A native write may finish after cancellation.
			}
			if mode == "window-close" {
				// Keep the Application owner alive while retiring the target window.
				if _, err := f.a.OpenWindow(app.Window{ID: "keeper", Title: "Keeper", Width: 100, Height: 100,
					View: func() ui.Node { return ui.Node{ID: "keeper-root", Kind: ui.Label} }}); err != nil {
					t.Fatal(err)
				}
			}
			var before []canvas.Rect
			var prior history.State
			f.owner(t, func() {
				before, prior = slices.Clone(f.m.scene.Objects), f.m.journal.Snapshot()
				if err := f.m.commands.Invoke(f.ref.Context(), "cut"); err != nil {
					t.Fatal(err)
				}
			})
			cutAwait(t, s.started)
			switch mode {
			case "cancel-writing":
				f.owner(t, func() { f.m.clipboardJob.cancel() })
			case "cancel-queued":
				close(s.release)
				cutAwaitReady(t, f)
				f.owner(t, func() { f.m.clipboardJob.cancel() })
			case "request-close":
				f.owner(t, func() {
					if !f.m.requestClose(f.a) || f.m.clipboardJob != nil {
						t.Fatal("close did not retire the pending Cut")
					}
				})
				close(s.release)
			case "window-close":
				f.owner(t, func() {
					if err := f.a.CloseWindow(f.ref); err != nil {
						t.Fatal(err)
					}
				})
			default:
				close(s.release)
			}
			cutAwait(t, s.returned)
			if mode == "window-close" || mode == "request-close" {
				if err := f.a.Drain(); err != nil {
					t.Fatal(err)
				}
			} else {
				waitClipboard(t, f)
			}
			f.owner(t, func() {
				if !slices.Equal(f.m.scene.Objects, before) || f.m.journal.Snapshot() != prior ||
					!slices.Equal(f.m.selectedIDs(), []string{"block-1", "note-2"}) || f.m.scene.Selected != "note-2" {
					t.Fatal("refused or retired Cut changed the document", f.m.status)
				}
			})
		})
	}
}

func TestCutRechecksFrozenDocumentAndSelectionAtOwnerAdmission(t *testing.T) {
	for _, mode := range []string{"selection", "primary", "geometry", "group", "counter", "revision"} {
		t.Run(mode, func(t *testing.T) {
			f, s := cutFixture(t)
			f.owner(t, func() {
				if err := f.m.commands.Invoke(f.ref.Context(), "cut"); err != nil {
					t.Fatal(err)
				}
			})
			cutAwait(t, s.started)
			close(s.release)
			cutAwaitReady(t, f)
			var current []canvas.Rect
			var currentGroups map[string]string
			var currentIDs []string
			var primary string
			var prior history.State
			f.owner(t, func() {
				// Simulate an external owner-held change bypassing the normal busy
				// gate. A late Cut must refuse, rather than target the new state.
				switch mode {
				case "selection":
					f.m.setSelection([]string{"block-1"}, "block-1")
				case "primary":
					f.m.setSelection([]string{"block-1", "note-2"}, "block-1")
				case "geometry":
					f.m.scene.Objects[0].Width++
				case "group":
					f.m.groups = map[string]string{"block-1": "group-1", "note-2": "group-1"}
					f.m.nextGroup = 1
				case "counter":
					f.m.next++
				case "revision":
					if err := f.m.journal.Reset(true); err != nil {
						t.Fatal(err)
					}
				}
				current, currentGroups, currentIDs = slices.Clone(f.m.scene.Objects), maps.Clone(f.m.groups), f.m.selectedIDs()
				primary, prior = f.m.scene.Selected, f.m.journal.Snapshot()
			})
			waitClipboard(t, f)
			f.owner(t, func() {
				if !slices.Equal(f.m.scene.Objects, current) || !maps.Equal(f.m.groups, currentGroups) ||
					!slices.Equal(f.m.selectedIDs(), currentIDs) || f.m.scene.Selected != primary || f.m.journal.Snapshot() != prior ||
					!strings.Contains(f.m.status, history.ErrStale.Error()) {
					t.Fatal("stale Cut changed or hid the accepted state", f.m.status)
				}
			})
		})
	}
}

func TestCutAdmissionFailureKeepsDocument(t *testing.T) {
	f, _ := cutFixture(t)
	f.a.SetServices(nil)
	f.owner(t, func() {
		before, prior := slices.Clone(f.m.scene.Objects), f.m.journal.Snapshot()
		if err := f.m.commands.Invoke(f.ref.Context(), "cut"); !errors.Is(err, app.ErrUnavailable) || f.m.clipboardJob != nil || !slices.Equal(before, f.m.scene.Objects) || f.m.journal.Snapshot() != prior {
			t.Fatal("failed service admission mutated the drawing", err)
		}
		f.m.setSelection(nil, "")
		if err := f.m.commands.Enabled("cut"); !errors.Is(err, errNoSelection) {
			t.Fatal("Cut enabled without a selection", err)
		}
	})
}
