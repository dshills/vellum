package main

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/ui"
)

func (m *model) fileBusy() bool {
	return m.layerDrag != nil || m.clipboardJob != nil || (m.files != nil && m.files.State().Busy)
}

func (m *model) bindFiles(a *app.Application, ref app.WindowRef, choose func(context.Context, app.FileOptions) (string, error)) error {
	var err error
	m.files, err = document.New(document.Config[drawingFile]{Owner: a, Window: "placement", Journal: m.journal, Codec: drawingCodec(m), Filename: "Drawing.vellum", Extensions: []string{"vellum"}, MaxBytes: drawingByteLimit, Choose: choose, OnComplete: func(r document.Result) {
		switch {
		case errors.Is(r.Err, app.ErrCanceled), errors.Is(r.Err, context.Canceled):
			m.status = "File operation canceled; drawing kept"
		case r.Err != nil:
			m.status = "File operation failed: " + r.Err.Error()
		case r.Action == document.Save:
			m.status = "Saved " + filepath.Base(r.Path)
		case r.Action == document.Open:
			m.status = "Opened " + filepath.Base(r.Path)
		}
		if r.Err == nil {
			m.recovery.Request()
		}
		_ = a.Invalidate("placement")
	}})
	if err != nil {
		return err
	}
	for _, c := range []struct{ id, label, shortcut string }{{"open", "Open…", "Meta+O"}, {"save", "Save", "Meta+S"}, {"save-as", "Save As…", "Meta+Shift+S"}} {
		if err := m.commands.Register(command.Command{ID: c.id, Label: c.label, Shortcut: c.shortcut, Available: func() error {
			if m.fileBusy() || m.move != nil || m.prompt != "" || m.palette.State().Active {
				return document.ErrBusy
			}
			return nil
		}, Execute: func(context.Context) error { return m.fileAction(a, ref, c.id, false) }}); err != nil {
			return err
		}
	}
	return nil
}

func (m *model) fileAction(a *app.Application, ref app.WindowRef, action string, discard bool) error {
	if m.files == nil {
		return app.ErrUnavailable
	}
	if m.recovery.pending() {
		return document.ErrBusy
	}
	var err error
	switch action {
	case "open":
		err = m.files.Open(ref.Context(), document.OpenOptions{DiscardChanges: discard})
		if errors.Is(err, document.ErrUnsaved) {
			m.prompt = "open"
			err = nil
		} else if err == nil {
			m.status = "Opening drawing…"
		}
	case "save", "save-as":
		err = m.files.Save(ref.Context(), document.SaveOptions{Choose: action == "save-as"})
		if err == nil {
			m.status = "Saving drawing…"
		}
	}
	_ = a.Invalidate("placement")
	return err
}

func (m *model) requestClose(a *app.Application) bool {
	m.removeRecoveryOnClose = false
	if m.recovery.pending() {
		// Closing a startup decision keeps the interrupted session's copy.
		return true
	}
	m.cancelLayerDrag(m.layerDrag)
	m.cancelMove(m.move)
	if job := m.clipboardJob; job != nil {
		m.clipboardJob = nil
		job.cancel()
		job.ticket.Abort()
	}
	if m.fileBusy() {
		m.files.Cancel()
		m.status = "Canceling file operation; close again when it finishes"
		_ = a.Invalidate("placement")
		return false
	}
	if m.journal.Snapshot().Dirty {
		m.prompt = "close"
		_ = a.Invalidate("placement")
		return false
	}
	m.removeRecoveryOnClose = true
	return true
}

func (m *model) filePrompt(a *app.Application, ref app.WindowRef) ui.Node {
	cancel := func() { m.prompt = ""; _ = a.Invalidate("placement") }
	text := "Discard unsaved changes and open another drawing?"
	label := "Discard and Open…"
	if m.prompt == "close" {
		text, label = "Discard unsaved changes and close? Cancel to save first.", "Discard and Close"
	}
	return ui.Node{ID: "file-decision", Kind: ui.Dialog, Text: "Unsaved drawing", Open: m.prompt == "open" || m.prompt == "close", Dismiss: cancel, Children: []ui.Node{
		{ID: "file-decision-text", Kind: ui.Label, Text: text},
		{ID: "file-discard", Kind: ui.Button, Text: label, Activate: func() {
			prompt := m.prompt
			m.prompt = ""
			if prompt == "close" {
				m.removeRecoveryOnClose = true
				_ = a.CloseWindow(ref)
			} else if err := m.fileAction(a, ref, "open", true); err != nil {
				m.status = err.Error()
			}
			_ = a.Invalidate("placement")
		}},
		{ID: "file-keep", Kind: ui.Button, Text: "Cancel", Activate: cancel},
	}}
}
