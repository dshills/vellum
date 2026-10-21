package main

import (
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/history"
	"os"
	"path/filepath"
)

type detachedFileSession interface {
	State() document.State
	Cancel()
}
type detachedSaveConfig[T any] struct {
	title, filename, extension, success string
	codec                               document.Codec[T]
	store                               document.Store
	recordRecent                        bool
}

// A private journal/session uses the existing bounded worker and lifetime
// machinery. Saving this snapshot never marks the active journal saved or
// changes the active session's path. Only one output operation can be active.
func startDetachedSave[T any](m *model, a *app.Application, ref app.WindowRef, c detachedSaveConfig[T]) error {
	if ref.Context().Err() != nil {
		return app.ErrWindowClosed
	}
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	activePath := ""
	if m.files != nil {
		activePath = m.files.State().Path
	}
	choose := m.chooseFile
	if choose == nil {
		return app.ErrUnavailable
	}
	chooser := func(ctx context.Context, o app.FileOptions) (string, error) {
		o.Title = c.title
		if activePath != "" {
			o.Directory = filepath.Dir(activePath)
		}
		path, err := choose(ctx, o)
		if err != nil {
			return "", err
		}
		if err = separateDrawingDestination(activePath, path); err != nil {
			return "", err
		}
		return path, nil
	}
	journal, err := history.New(1)
	if err != nil {
		return err
	}
	var session *document.Session[T]
	session, err = document.New(document.Config[T]{Owner: a, Window: "placement", Journal: journal, Codec: c.codec, Filename: c.filename, Extensions: []string{c.extension}, MaxBytes: drawingByteLimit, Store: c.store, Choose: chooser, OnComplete: func(r document.Result) {
		if m.output != session {
			return
		}
		m.output = nil
		switch {
		case errors.Is(r.Err, app.ErrCanceled), errors.Is(r.Err, context.Canceled):
			m.status = c.title + " canceled; drawing kept"
		case r.Err != nil:
			m.status = c.title + " failed: " + r.Err.Error()
			if r.Written {
				m.status += " • destination may already contain the snapshot"
			}
		default:
			m.status = c.success + " " + filepath.Base(r.Path)
			if c.recordRecent && m.recent != nil {
				if err := m.recent.Record(r.Path); err != nil {
					m.status += " • Recent drawings unavailable: " + err.Error()
				}
			}
		}
		_ = a.Invalidate("placement")
	}})
	if err != nil {
		return err
	}
	m.output = session
	if err = session.Save(ref.Context(), document.SaveOptions{Choose: true}); err != nil {
		m.output = nil
		return err
	}
	m.status = c.title + "…"
	_ = a.Invalidate("placement")
	return nil
}

// Check path and existing file identity on the worker, including hard links or
// aliases through parent directories. The atomic store also rejects symlinks.
func separateDrawingDestination(active, path string) error {
	if active == "" {
		return nil
	}
	if filepath.Clean(active) == filepath.Clean(path) {
		return errors.New("choose a destination different from the active drawing")
	}
	a, err := os.Stat(active)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	b, other := os.Stat(path)
	if other != nil && !os.IsNotExist(other) {
		return other
	}
	if err == nil && other == nil && os.SameFile(a, b) {
		return errors.New("destination refers to the active drawing; choose another file")
	}
	return nil
}
func (m *model) saveCopy(a *app.Application, ref app.WindowRef, store document.Store) error {
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	codec := drawingCodec(m)
	image, err := codec.Capture()
	if err != nil {
		return err
	}
	codec.Capture = func() (drawingFile, error) { return image, nil }
	codec.Install = func(drawingFile) error { return app.ErrUnavailable }
	name := "Drawing"
	if m.files != nil && m.files.State().Path != "" {
		name = drawingSuggestion(filepath.Base(m.files.State().Path))
	}
	return startDetachedSave(m, a, ref, detachedSaveConfig[drawingFile]{title: "Save a Copy", filename: name + " Copy", extension: "vellum", success: "Saved copy", codec: codec, store: store, recordRecent: true})
}
