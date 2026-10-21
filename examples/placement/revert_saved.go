package main

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"slices"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
)

// A confirmation belongs to one complete document snapshot, including selection
// changes that do not advance history. It never carries a path to a later file.
type revertDraft struct {
	files   *document.Session[drawingFile]
	scene   *canvas.Scene
	journal *history.Journal
	state   history.State
	path    string
	drawing drawingFile
	ids     []string
	message string
}

func (m *model) bindRevertSaved(a *app.Application, ref app.WindowRef) error {
	return m.commands.Register(command.Command{ID: "revert-saved", Label: "Revert to Saved…", Available: func() error {
		return m.revertSavedAvailable(ref)
	}, Execute: func(context.Context) error { return m.revertSaved(a, ref, false) }})
}

func (m *model) revertSavedAvailable(ref app.WindowRef) error {
	if ref.Context().Err() != nil {
		return app.ErrWindowClosed
	}
	if m.files == nil || m.files.State().Closed || m.files.State().Path == "" {
		return app.ErrUnavailable
	}
	if m.pending != nil {
		return document.ErrBusy
	}
	return m.selectionAvailable()
}

// Revert deliberately uses the active document session's validated Open path.
// Failed reads/decode/install preserve the drawing, active path and Undo/Redo;
// only a successful replacement clears history and marks the drawing saved.
func (m *model) revertSaved(a *app.Application, ref app.WindowRef, discard bool) error {
	if err := m.revertSavedAvailable(ref); err != nil {
		return err
	}
	path := m.files.State().Path
	err := m.files.Open(ref.Context(), document.OpenOptions{Path: path, DiscardChanges: discard})
	if errors.Is(err, document.ErrUnsaved) {
		if m.modalEpoch == ^uint64(0) {
			return errors.New("dialog focus sequence exhausted")
		}
		drawing, captureErr := drawingCodec(m).Capture()
		if captureErr != nil {
			return captureErr
		}
		m.revertDraft = &revertDraft{files: m.files, scene: m.scene, journal: m.journal, state: m.journal.Snapshot(), path: path, drawing: drawing, ids: slices.Clone(m.selectedIDs())}
		m.prompt = "revert-saved"
		m.modalEpoch++
		err = nil
	} else if err == nil {
		m.revertDraft = nil
		m.status = "Reverting to saved drawing…"
	}
	_ = a.Invalidate("placement")
	return err
}

func (m *model) validateRevertDraft(d *revertDraft) error {
	if d == nil || m.revertDraft != d || m.files != d.files || m.scene != d.scene || m.journal != d.journal || m.journal.Snapshot() != d.state || m.files.State().Path != d.path || !slices.Equal(m.selectedIDs(), d.ids) {
		return history.ErrStale
	}
	current, err := drawingCodec(m).Capture()
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(current, d.drawing) {
		return history.ErrStale
	}
	return nil
}

func (m *model) revertPrompt(a *app.Application, ref app.WindowRef) ui.Node {
	d := m.revertDraft
	active := func() bool {
		return ref.Context().Err() == nil && d != nil && m.revertDraft == d && m.prompt == "revert-saved"
	}
	cancel := func() {
		if active() {
			m.revertDraft, m.prompt = nil, ""
			m.status = "Revert canceled; drawing kept"
			_ = a.Invalidate("placement")
		}
	}
	name, message := "drawing", ""
	if d != nil {
		name, message = filepath.Base(d.path), d.message
	}
	return ui.Node{ID: "revert-saved-dialog", Kind: ui.Dialog, Text: "Revert to Saved?", Open: active(), Dismiss: cancel, Children: []ui.Node{
		{ID: "revert-saved-help", Kind: ui.Label, Text: "Reload " + name + " from disk? Unsaved changes will be discarded. A successful reload clears Undo/Redo. If the file cannot be read, the current drawing and history are kept.", Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 460})}},
		{ID: "revert-saved-error", Kind: ui.Label, Text: message},
		{ID: "revert-saved-cancel", Kind: ui.Button, Text: "Cancel", Activate: cancel},
		{ID: "revert-saved-confirm", Kind: ui.Button, Text: "Discard and Revert", Activate: func() {
			if !active() {
				return
			}
			// Admission runs synchronously on the UI owner. No intermediate View
			// observes the cleared prompt before an admission failure restores it.
			m.prompt = ""
			err := m.validateRevertDraft(d)
			if err == nil {
				err = m.revertSavedAvailable(ref)
			}
			if err == nil {
				err = m.revertSaved(a, ref, true)
			}
			if err != nil {
				m.status = err.Error()
				if errors.Is(err, history.ErrStale) {
					m.revertDraft = nil
				} else {
					d.message = err.Error()
					m.prompt = "revert-saved"
				}
			}
			_ = a.Invalidate("placement")
		}},
	}}
}
