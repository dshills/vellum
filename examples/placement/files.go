package main

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/ui"
)

func (m *model) fileBusy() bool {
	return m.preferencesLoading || m.layerDrag != nil || m.clipboardJob != nil || (m.output != nil && m.output.State().Busy) || (m.files != nil && m.files.State().Busy)
}

func (m *model) bindFiles(a *app.Application, ref app.WindowRef, choose func(context.Context, app.FileOptions) (string, error)) error {
	var err error
	nativeChooser := func(ctx context.Context, options app.FileOptions) (string, error) {
		if choose != nil {
			return choose(ctx, options)
		}
		services, err := a.Services()
		if err != nil {
			return "", err
		}
		return services.ChooseFile(ctx, options)
	}
	m.chooseFile = nativeChooser
	// Normalize only a suggestion, never the confirmed destination.
	chooseDrawing := func(ctx context.Context, options app.FileOptions) (string, error) {
		if options.Mode == app.SaveFile {
			options.Filename = drawingSuggestion(options.Filename)
		}
		return nativeChooser(ctx, options)
	}
	m.newFileSession = func() (*document.Session[drawingFile], error) {
		return document.New(document.Config[drawingFile]{Owner: a, Window: "placement", Journal: m.journal, Codec: drawingCodec(m), Filename: "Drawing.vellum", Extensions: []string{"vellum"}, MaxBytes: drawingByteLimit, Choose: chooseDrawing, OnComplete: func(r document.Result) {
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
				if m.recent != nil && (r.Action == document.Open || r.Action == document.Save) {
					if err := m.recent.Record(r.Path); err != nil {
						m.status += " • Recent drawings unavailable: " + err.Error()
					}
				}
				m.recovery.Request()
			}
			_ = a.Invalidate("placement")
		}})
	}
	m.files, err = m.newFileSession()
	if err != nil {
		return err
	}
	for _, c := range []struct{ id, label, shortcut string }{{"new", "New", "Meta+N"}, {"open", "Open…", "Meta+O"}, {"save", "Save", "Meta+S"}, {"save-as", "Save As…", "Meta+Shift+S"}, {"save-copy", "Save a Copy…", ""}, {"export-svg", "Export SVG…", ""}, {"export-selected-svg", "Export Selected SVG…", ""}} {
		if err := m.commands.Register(command.Command{ID: c.id, Label: c.label, Shortcut: c.shortcut, Available: func() error {
			if c.id == "export-selected-svg" {
				return m.selectedSVGAvailable()
			}
			if c.id == "export-svg" {
				return m.svgAvailable()
			}
			if m.fileBusy() || m.move != nil || m.prompt != "" || m.palette.State().Active {
				return document.ErrBusy
			}
			return nil
		}, Execute: func(context.Context) error { return m.fileAction(a, ref, c.id, false) }}); err != nil {
			return err
		}
	}
	return m.bindRevertSaved(a, ref)
}

func drawingSuggestion(name string) string {
	const suffix = ".vellum"
	for len(name) >= len(suffix) && strings.EqualFold(name[len(name)-len(suffix):], suffix) {
		name = name[:len(name)-len(suffix)]
	}
	if name == "" {
		return "Drawing"
	}
	return name
}

func (m *model) fileAction(a *app.Application, ref app.WindowRef, action string, discard bool) error {
	if m.files == nil {
		return app.ErrUnavailable
	}
	if m.recovery.pending() || m.clipboardAvailable() != nil {
		return document.ErrBusy
	}
	var err error
	switch action {
	case "export-selected-svg":
		return m.exportSelectedSVG(a, ref, nil)
	case "export-svg":
		return m.exportSVG(a, ref, nil)
	case "save-copy":
		return m.saveCopy(a, ref, nil)
	case "new":
		return m.newDrawing(a, ref, discard)
	case "open":
		path := ""
		if discard {
			path = m.pendingOpenPath
		}
		return m.openDrawing(a, ref, path, discard)
	case "save", "save-as":
		err = m.files.Save(ref.Context(), document.SaveOptions{Choose: action == "save-as"})
		if err == nil {
			m.status = "Saving drawing…"
		}
	}
	_ = a.Invalidate("placement")
	return err
}

func (m *model) openRecent(a *app.Application, ref app.WindowRef, path string) error {
	if m.recent == nil || !slicesContainsPath(m.recent.Paths(), path) {
		return app.ErrUnavailable
	}
	return m.openDrawing(a, ref, path, false)
}

func (m *model) openDrawing(a *app.Application, ref app.WindowRef, path string, discard bool) error {
	if m.files == nil {
		return app.ErrUnavailable
	}
	if ref.Context().Err() != nil {
		return app.ErrWindowClosed
	}
	if m.recovery.pending() || m.clipboardAvailable() != nil {
		return document.ErrBusy
	}
	err := m.files.Open(ref.Context(), document.OpenOptions{Path: path, DiscardChanges: discard})
	if errors.Is(err, document.ErrUnsaved) {
		m.pendingOpenPath, m.prompt = path, "open"
		err = nil
	} else {
		m.pendingOpenPath = ""
		if err == nil {
			m.status = "Opening drawing…"
		}
	}
	_ = a.Invalidate("placement")
	return err
}

func (m *model) requestClose(a *app.Application) bool {
	m.removeRecoveryOnClose = false
	m.cancelPreferenceStartup()
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
		if m.output != nil {
			m.output.Cancel()
		}
		if m.files != nil {
			m.files.Cancel()
		}
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
	expectedPrompt, expectedPath := m.prompt, m.pendingOpenPath
	current := func() bool {
		return ref.Context().Err() == nil && m.prompt == expectedPrompt && m.pendingOpenPath == expectedPath
	}
	cancel := func() {
		if !current() {
			return
		}
		m.prompt, m.pendingOpenPath = "", ""
		_ = a.Invalidate("placement")
	}
	text := "Discard unsaved changes and open another drawing?"
	label := "Discard and Open…"
	if m.prompt == "new" {
		text, label = "Discard unsaved changes and start a new drawing? Cancel to save first.", "Discard and New"
	}
	if m.prompt == "close" {
		text, label = "Discard unsaved changes and close? Cancel to save first.", "Discard and Close"
	}
	return ui.Node{ID: "file-decision", Kind: ui.Dialog, Text: "Unsaved drawing", Open: m.prompt == "open" || m.prompt == "close" || m.prompt == "new", Dismiss: cancel, Children: []ui.Node{
		{ID: "file-decision-text", Kind: ui.Label, Text: text},
		{ID: "file-discard", Kind: ui.Button, Text: label, Activate: func() {
			if !current() {
				return
			}
			prompt := m.prompt
			m.prompt = ""
			if prompt == "close" {
				m.pendingOpenPath = ""
				m.removeRecoveryOnClose = true
				_ = a.CloseWindow(ref)
			} else if err := m.fileAction(a, ref, prompt, true); err != nil {
				m.status = err.Error()
			}
			_ = a.Invalidate("placement")
		}},
		{ID: "file-keep", Kind: ui.Button, Text: "Cancel", Activate: cancel},
	}}
}

// A fresh idle session forgets the previous Save destination without changing the
// public document API. Admission precedes replacement; no file IO is needed.
func (m *model) newDrawing(a *app.Application, ref app.WindowRef, discard bool) error {
	if ref.Context().Err() != nil {
		return app.ErrWindowClosed
	}
	if m.newFileSession == nil {
		return app.ErrUnavailable
	}
	if err := m.clipboardAvailable(); err != nil {
		return err
	}
	if m.journal.Snapshot().Dirty && !discard {
		m.prompt, m.pendingOpenPath = "new", ""
		_ = a.Invalidate("placement")
		return nil
	}
	session, err := m.newFileSession()
	if err != nil {
		return err
	}
	blank := drawingFile{Format: "vellum-placement", Version: 2, Zoom: 1, Snap: m.snap}
	if err = m.journal.Replace(true, func() error { return drawingCodec(m).Install(blank) }); err != nil {
		return err
	}
	m.files = session
	m.pendingOpenPath, m.renameID, m.renameDraft = "", "", ""
	m.geometry = nil
	m.recovery.Request()
	m.status = "New drawing"
	_ = a.Invalidate("placement")
	return nil
}
