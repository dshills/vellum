package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/history"
)

// The document image is immutable exact UTF-8 bytes held in a Go string. No line
// ending, normalization, or terminal-empty-line conversion happens in this codec.
func documentCodec(model *editor.Model) document.Codec[string] {
	limit := model.MaxBytes() // Immutable worker budget, captured on the owner.
	return document.Codec[string]{
		Capture: func() (string, error) {
			if err := model.Validate(); err != nil {
				return "", err
			}
			snapshot := model.Snapshot()
			if snapshot.Composing {
				return "", editor.ErrComposition
			}
			return snapshot.Text, nil
		},
		Encode: func(value string) ([]byte, error) {
			if !utf8.ValidString(value) || strings.IndexByte(value, 0) >= 0 {
				return nil, editor.ErrInvalidText
			}
			if len(value) > limit {
				return nil, document.ErrTooLarge
			}
			return []byte(value), nil
		},
		Decode: func(value []byte) (string, error) {
			if !utf8.Valid(value) || bytes.IndexByte(value, 0) >= 0 {
				return "", editor.ErrInvalidText
			}
			if len(value) > limit {
				return "", document.ErrTooLarge
			}
			return string(value), nil
		},
		Install: func(value string) error { return model.Install(value, editor.Selection{}) },
	}
}

type settleFunc func(context.Context, *app.Application, string, *editor.Model, func(error)) error

type action string

const (
	actionSave         action = "save"
	actionSaveAs       action = "save-as"
	actionOpen         action = "open"
	actionDiscardOpen  action = "discard-open"
	actionClose        action = "close"
	actionSaveClose    action = "save-close"
	actionSavedClose   action = "saved-close"
	actionDiscardClose action = "discard-close"
	actionUndo         action = "undo"
	actionRedo         action = "redo"
)

// start is UI-owner only. It first acknowledges earlier originating native work,
// then starts a persistence/history action against that settled owner snapshot.
func (w *editorWindow) start(kind action, path string) {
	if w.settling || w.session.State().Busy {
		w.status = "Another document operation is still running"
		w.invalidate()
		return
	}
	w.settling = true
	w.cancelRequested = false
	w.operation++
	token := w.operation
	settleCtx, cancelSettle := context.WithCancel(w.lifetime)
	w.settleCancel = cancelSettle
	w.status = "Finishing text input"
	w.invalidate()
	err := w.settle(settleCtx, w.application, w.id, w.model, func(err error) {
		if w.lifetime.Err() != nil || token != w.operation {
			cancelSettle()
			return
		}
		w.settling = false
		w.settleCancel = nil
		cancelSettle()
		if err != nil {
			w.finish(document.Result{Err: err})
			return
		}
		if w.cancelRequested {
			w.finish(document.Result{Err: context.Canceled})
			return
		}
		switch kind {
		case actionSave, actionSaveAs, actionSaveClose:
			w.closeAfterSave = kind == actionSaveClose
			err = w.session.Save(w.lifetime, document.SaveOptions{Path: path, Choose: kind == actionSaveAs})
		case actionOpen, actionDiscardOpen:
			if kind == actionOpen && w.session.State().Dirty {
				w.prompt = "open"
				w.openPath = path
				w.status = "Choose whether to keep or discard this document"
				w.invalidate()
				return
			}
			err = w.session.Open(w.lifetime, document.OpenOptions{Path: path, DiscardChanges: kind == actionDiscardOpen})
		case actionClose:
			if w.session.State().Dirty {
				w.prompt = "close"
				w.status = "This document has unsaved changes"
				w.invalidate()
				return
			}
			err = w.closeNow()
		case actionSavedClose:
			if w.session.State().Dirty || w.model.Snapshot().Composing {
				w.status = "New changes remain; the window stays open"
			} else {
				err = w.closeNow()
			}
		case actionDiscardClose:
			err = w.closeNow()
		case actionUndo:
			err = w.model.Journal().Undo()
			w.status = "Undo applied"
		case actionRedo:
			err = w.model.Journal().Redo()
			w.status = "Redo applied"
		default:
			err = fmt.Errorf("unknown document action %q", kind)
		}
		if err != nil {
			w.closeAfterSave = false
			w.finish(document.Result{Err: err})
			return
		}
		if kind == actionSave || kind == actionSaveAs || kind == actionSaveClose {
			w.status = "Saving document"
		}
		if kind == actionOpen || kind == actionDiscardOpen {
			w.status = "Opening document"
		}
		w.invalidate()
	})
	if err != nil {
		w.settling = false
		w.settleCancel = nil
		cancelSettle()
		w.finish(document.Result{Err: err})
	}
}
func (w *editorWindow) cancel() {
	w.cancelRequested = true
	w.closeAfterSave = false
	if w.settling && w.settleCancel != nil {
		w.settleCancel()
	}
	if w.session.State().Busy {
		w.session.Cancel()
		w.status = "Canceling document operation"
	} else if w.settling {
		w.status = "Canceling input acknowledgment"
	} else {
		w.status = "No operation to cancel"
	}
	w.invalidate()
}
func (w *editorWindow) finish(result document.Result) {
	w.last = result
	w.resultCount++
	switch {
	case errors.Is(result.Err, history.ErrStale):
		w.status = "The document changed during the operation; it remains unsaved"
	case errors.Is(result.Err, context.Canceled) || errors.Is(result.Err, app.ErrCanceled):
		w.status = "Document operation canceled"
	case result.Err != nil:
		w.status = result.Err.Error()
	case result.Action == document.Open:
		w.status = "Document opened"
		w.focusID = w.editorID
		w.focusSequence++
	case result.Action == document.Save:
		w.status = "Document saved"
	}
	if result.Err != nil {
		w.closeAfterSave = false
	}
	if result.Action == document.Save && result.Err == nil && w.closeAfterSave {
		w.closeAfterSave = false
		// Save captured an earlier committed revision. Input may have been admitted
		// or begun preedit during IO without making that journal revision dirty yet.
		// A second originating FIFO barrier must finish those edits before closure.
		w.start(actionSavedClose, "")
	}

	w.invalidate()
}
func (w *editorWindow) closeNow() error { return w.application.CloseWindow(w.windowRef) }
func (w *editorWindow) invalidate()     { _ = w.application.Invalidate(w.id) }
