package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/placement"
	"github.com/dshills/vellum/ui"
)

func savedRevertFixture(t *testing.T) (*fileFixture, drawingFile) {
	t.Helper()
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{X: -80, Y: 10})
	f.place(t, "note", canvas.Point{X: 90, Y: 60})
	var saved drawingFile
	f.owner(t, func() {
		f.m.setSelection([]string{"block-1", "note-2"}, "block-1")
		if err := f.m.changeGrouping(false); err != nil {
			t.Fatal(err)
		}
		if err := f.m.renameLayer("group-1", "Saved assembly"); err != nil {
			t.Fatal(err)
		}
		if err := f.m.beginFillColor(); err != nil {
			t.Fatal(err)
		}
		f.m.fillDraft.value = "#204080"
		f.m.prompt = ""
		if err := f.m.applyFillColor(f.m.fillDraft); err != nil {
			t.Fatal(err)
		}
		f.m.fillDraft = nil
		f.m.scene.View.Center = canvas.Point{X: 32, Y: 45}
		f.m.scene.View.Zoom = 1.75
		f.m.snap = true
		var err error
		saved, err = drawingCodec(f.m).Capture()
		if err != nil {
			t.Fatal(err)
		}
		if err := f.m.fileAction(f.a, f.ref, "save", false); err != nil {
			t.Fatal(err)
		}
	})
	f.wait(t)
	f.owner(t, func() {
		if state := f.m.files.State(); state.Path != f.path || state.Dirty || state.Last.Err != nil {
			t.Fatal("fixture did not save", state)
		}
	})
	return f, saved
}

func TestRevertSavedCancelThenReloadRestoresMetadataAndResetsHistory(t *testing.T) {
	f, saved := savedRevertFixture(t)
	activePath := f.path
	// A chooser call would use this different path and fail; Revert must reload
	// the active session's path without asking for a destination.
	f.path = filepath.Join(t.TempDir(), "must-not-choose.vellum")
	f.place(t, "connector", canvas.Point{X: 220, Y: 110})
	f.owner(t, func() {
		if err := f.m.renameLayer("group-1", "Unsaved assembly"); err != nil {
			t.Fatal(err)
		}
		f.m.scene.View.Center = canvas.Point{X: 900, Y: 100}
		f.m.scene.View.Zoom = .6
		f.m.snap = false
		before, err := drawingCodec(f.m).Capture()
		if err != nil {
			t.Fatal(err)
		}
		state := f.m.journal.Snapshot()
		ids := slices.Clone(f.m.selectedIDs())
		if err := f.m.commands.Invoke(f.ref.Context(), "revert-saved"); err != nil {
			t.Fatal(err)
		}
		dialog := f.m.revertPrompt(f.a, f.ref)
		if !dialog.Open || f.m.files.State().Busy {
			t.Fatal("dirty revert did not wait for explicit confirmation")
		}
		if err := ui.Validate(dialog); err != nil {
			t.Fatal(err)
		}
		geometryDialogControl(t, dialog, "revert-saved-cancel").Activate()
		after, _ := drawingCodec(f.m).Capture()
		if f.m.revertDraft != nil || f.m.prompt != "" || !reflect.DeepEqual(before, after) || state != f.m.journal.Snapshot() || !slices.Equal(ids, f.m.selectedIDs()) || f.m.files.State().Path != activePath {
			t.Fatal("Cancel changed drawing, history, selection or path")
		}
		if err := f.m.revertSaved(f.a, f.ref, false); err != nil {
			t.Fatal(err)
		}
		newDraft := f.m.revertDraft
		geometryDialogControl(t, dialog, "revert-saved-confirm").Activate()
		dialog.Dismiss()
		if f.m.revertDraft != newDraft || f.m.prompt != "revert-saved" || f.m.files.State().Busy {
			t.Fatal("retired dialog affected its replacement")
		}
		current := f.m.revertPrompt(f.a, f.ref)
		geometryDialogControl(t, current, "revert-saved-confirm").Activate()
		geometryDialogControl(t, current, "revert-saved-confirm").Activate()
		if f.m.revertDraft != nil || f.m.prompt != "" || !f.m.files.State().Busy {
			t.Fatal("confirmed revert was not admitted exactly once")
		}
	})
	f.wait(t)
	f.owner(t, func() {
		got, _ := drawingCodec(f.m).Capture()
		state, files := f.m.journal.Snapshot(), f.m.files.State()
		if !reflect.DeepEqual(got, saved) || files.Path != activePath || files.Last.Err != nil || state.Dirty || state.UndoCount != 0 || state.RedoCount != 0 {
			t.Fatal("revert did not restore saved colors/groups/names/camera/snap/IDs and reset history", got, saved, state, files)
		}
		if !errors.Is(f.m.journal.Undo(), history.ErrEmpty) || !errors.Is(f.m.journal.Redo(), history.ErrEmpty) {
			t.Fatal("discarded edits remained replayable")
		}
	})
}

func TestRevertSavedCleanDrawingReloadsDiskWithoutSelectionOrChooser(t *testing.T) {
	f, saved := savedRevertFixture(t)
	saved.Objects[0].X = -400
	saved.Layers["group-1"] = layerMetadata{Name: "Changed on disk"}
	data, err := drawingCodec(f.m).Encode(saved)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f.path, data, 0600); err != nil {
		t.Fatal(err)
	}
	f.owner(t, func() {
		f.m.setSelection(nil, "")
		if err := f.m.commands.Invoke(f.ref.Context(), "revert-saved"); err != nil {
			t.Fatal(err)
		}
		if f.m.prompt != "" || f.m.revertDraft != nil || !f.m.files.State().Busy {
			t.Fatal("clean reload unexpectedly required a selection or confirmation")
		}
	})
	f.wait(t)
	f.owner(t, func() {
		got, _ := drawingCodec(f.m).Capture()
		if !reflect.DeepEqual(got, saved) || f.m.files.State().Last.Err != nil || f.m.journal.Snapshot().Dirty {
			t.Fatal("clean revert did not reload disk", got, saved)
		}
	})
}

func TestRevertSavedReadAndDecodeFailuresPreserveDocumentAndHistory(t *testing.T) {
	for _, failure := range []string{"missing", "malformed", "invalid metadata"} {
		t.Run(failure, func(t *testing.T) {
			f, _ := savedRevertFixture(t)
			f.place(t, "connector", canvas.Point{X: 300, Y: 100})
			f.place(t, "block", canvas.Point{X: 400, Y: 100})
			f.owner(t, func() {
				if err := f.m.journal.Undo(); err != nil {
					t.Fatal(err)
				}
			})
			if failure == "missing" {
				if err := os.Remove(f.path); err != nil {
					t.Fatal(err)
				}
			} else {
				data := []byte("{malformed")
				if failure == "invalid metadata" {
					data = []byte(`{"format":"vellum-placement","version":4,"objects":[],"center":{"X":0,"Y":0},"zoom":1,"snap":false,"next_id":0,"layers":{"missing":{"name":"bad"}}}`)
				}
				if err := os.WriteFile(f.path, data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			var before drawingFile
			var state history.State
			var ids []string
			f.owner(t, func() {
				before, _ = drawingCodec(f.m).Capture()
				state, ids = f.m.journal.Snapshot(), slices.Clone(f.m.selectedIDs())
				if err := f.m.revertSaved(f.a, f.ref, false); err != nil {
					t.Fatal(err)
				}
				geometryDialogControl(t, f.m.revertPrompt(f.a, f.ref), "revert-saved-confirm").Activate()
			})
			f.wait(t)
			f.owner(t, func() {
				after, _ := drawingCodec(f.m).Capture()
				if f.m.files.State().Last.Err == nil || f.m.files.State().Path != f.path || !reflect.DeepEqual(before, after) || state != f.m.journal.Snapshot() || !slices.Equal(ids, f.m.selectedIDs()) {
					t.Fatal("failed revert changed active document/history/path")
				}
			})
		})
	}
}

type revertBlockingStore struct {
	document.FileStore
	entered, canceled chan struct{}
}

func (s revertBlockingStore) Read(ctx context.Context, _ string, _ int64) ([]byte, error) {
	close(s.entered)
	<-ctx.Done()
	close(s.canceled)
	return nil, ctx.Err()
}

func TestRevertSavedInFlightWindowClosureCannotInstall(t *testing.T) {
	f, _ := savedRevertFixture(t)
	// Keep the owner available to drain the retired window's completion.
	if _, err := f.a.OpenWindow(app.Window{ID: "keepalive", Title: "Keep owner alive", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "keepalive-root", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	f.place(t, "connector", canvas.Point{X: 300})
	store := revertBlockingStore{entered: make(chan struct{}), canceled: make(chan struct{})}
	var before drawingFile
	var state history.State
	f.owner(t, func() {
		before, _ = drawingCodec(f.m).Capture()
		state = f.m.journal.Snapshot()
		var err error
		f.m.files, err = document.New(document.Config[drawingFile]{Owner: f.a, Window: "placement", Journal: f.m.journal, Codec: drawingCodec(f.m), Path: f.path, MaxBytes: drawingByteLimit, Store: store})
		if err != nil {
			t.Fatal(err)
		}
		if err := f.m.revertSaved(f.a, f.ref, false); err != nil {
			t.Fatal(err)
		}
		geometryDialogControl(t, f.m.revertPrompt(f.a, f.ref), "revert-saved-confirm").Activate()
	})
	cutAwait(t, store.entered)
	f.owner(t, func() {
		if err := f.a.CloseWindow(f.ref); err != nil {
			t.Fatal(err)
		}
	})
	cutAwait(t, store.canceled)
	f.wait(t)
	f.owner(t, func() {
		after, _ := drawingCodec(f.m).Capture()
		if !f.m.files.State().Closed || !reflect.DeepEqual(before, after) || state != f.m.journal.Snapshot() || f.m.files.State().Path != f.path {
			t.Fatal("closed-window revert installed a document or reset history")
		}
	})
}

func TestRevertSavedConfirmationRejectsStaleDocumentSnapshots(t *testing.T) {
	for _, change := range []string{"selection", "objects", "metadata", "camera", "journal", "session"} {
		t.Run(change, func(t *testing.T) {
			f, _ := savedRevertFixture(t)
			f.place(t, "connector", canvas.Point{X: 300})
			f.owner(t, func() {
				if err := f.m.revertSaved(f.a, f.ref, false); err != nil {
					t.Fatal(err)
				}
				dialog := f.m.revertPrompt(f.a, f.ref)
				switch change {
				case "selection":
					f.m.setSelection(nil, "")
				case "objects":
					f.m.scene.Objects[0].X++
				case "metadata":
					f.m.layerInfo["group-1"] = layerMetadata{Name: "Later name"}
				case "camera":
					f.m.scene.View.Center.X++
				case "journal":
					if err := f.m.journal.Reset(false); err != nil {
						t.Fatal(err)
					}
				case "session":
					var err error
					f.m.files, err = f.m.newFileSession()
					if err != nil {
						t.Fatal(err)
					}
				}
				before, _ := drawingCodec(f.m).Capture()
				state := f.m.journal.Snapshot()
				geometryDialogControl(t, dialog, "revert-saved-confirm").Activate()
				after, _ := drawingCodec(f.m).Capture()
				if f.m.files.State().Busy || f.m.revertDraft != nil || f.m.prompt != "" || !reflect.DeepEqual(before, after) || state != f.m.journal.Snapshot() {
					t.Fatal("stale confirmation reopened or replaced a newer document")
				}
			})
		})
	}
}

func TestRevertSavedAdmissionAndWindowLifetime(t *testing.T) {
	f := fileModel(t, false)
	f.owner(t, func() {
		if !errors.Is(f.m.revertSaved(f.a, f.ref, false), app.ErrUnavailable) || f.m.commands.Enabled("revert-saved") == nil {
			t.Fatal("untitled drawing offered Revert")
		}
	})
	f, _ = savedRevertFixture(t)
	for _, gate := range []string{"modal", "move", "layer drag", "clipboard", "preferences", "recovery", "placement"} {
		f.owner(t, func() {
			switch gate {
			case "modal":
				f.m.prompt = "close"
			case "move":
				f.m.move = &objectMove{}
			case "layer drag":
				f.m.layerDrag = &layerDrag{}
			case "clipboard":
				f.m.clipboardJob = &clipboardOperation{}
			case "preferences":
				f.m.preferencesLoading = true
			case "recovery":
				f.m.recovery = &placementRecovery{checking: true}
			case "placement":
				f.m.pending = &placement.Request[component]{}
			}
			if !errors.Is(f.m.revertSaved(f.a, f.ref, false), document.ErrBusy) || f.m.commands.Enabled("revert-saved") == nil || f.m.files.State().Busy {
				t.Fatal("Revert admitted during", gate)
			}
			f.m.prompt, f.m.move, f.m.layerDrag, f.m.clipboardJob, f.m.preferencesLoading, f.m.recovery, f.m.pending = "", nil, nil, nil, false, nil, nil
		})
	}
	f.owner(t, func() {
		if err := f.m.revertSaved(f.a, f.ref, false); err != nil {
			t.Fatal(err)
		}
		if !errors.Is(f.m.revertSaved(f.a, f.ref, false), document.ErrBusy) {
			t.Fatal("concurrent file operation admitted a second revert")
		}
	})
	f.wait(t)
	f.place(t, "connector", canvas.Point{X: 300})
	var dialog ui.Node
	f.owner(t, func() {
		if err := f.m.revertSaved(f.a, f.ref, false); err != nil {
			t.Fatal(err)
		}
		dialog = f.m.revertPrompt(f.a, f.ref)
		if err := f.a.CloseWindow(f.ref); err != nil {
			t.Fatal(err)
		}
	})
	before, state, status := slices.Clone(f.m.scene.Objects), f.m.journal.Snapshot(), f.m.status
	geometryDialogControl(t, dialog, "revert-saved-confirm").Activate()
	geometryDialogControl(t, dialog, "revert-saved-cancel").Activate()
	dialog.Dismiss()
	if err := f.m.revertSaved(f.a, f.ref, false); !errors.Is(err, app.ErrWindowClosed) {
		t.Fatal("closed lifetime admitted revert", err)
	}
	if !slices.Equal(before, f.m.scene.Objects) || state != f.m.journal.Snapshot() || status != f.m.status {
		t.Fatal("retained closed-window callbacks mutated model")
	}
}
