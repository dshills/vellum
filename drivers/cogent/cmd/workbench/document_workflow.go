package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/ui"
)

type documentPersistence interface {
	State() document.State
	Save(context.Context, document.SaveOptions) error
	Open(context.Context, document.OpenOptions) error
	Cancel()
}
type documentWorkflow struct {
	session        documentPersistence
	owner          *app.Application
	window         string
	status         func(string)
	pending        string
	savingDecision bool
}

func (w *documentWorkflow) requestClose() bool {
	state := w.session.State()
	if !state.Dirty && !state.Busy {
		w.cancel()
		return true
	}
	if w.pending == "" {
		w.pending = "close"
	}
	w.owner.Invalidate(w.window)
	return false
}
func (w *documentWorkflow) open(ctx context.Context) error {
	if w.pending != "" {
		return app.ErrClosePending
	}
	err := w.session.Open(ctx, document.OpenOptions{})
	if errors.Is(err, document.ErrUnsaved) {
		w.pending = "open"
		w.owner.Invalidate(w.window)
		return nil
	}
	return err
}
func (w *documentWorkflow) cancel() {
	w.pending = ""
	w.savingDecision = false
	w.owner.Invalidate(w.window)
}
func (w *documentWorkflow) saveDecision() {
	if w.pending == "" || w.session.State().Busy {
		return
	}
	if err := w.session.Save(w.owner.Context(), document.SaveOptions{}); err != nil {
		w.status(err.Error())
		return
	}
	w.savingDecision = true
}
func (w *documentWorkflow) discard() {
	if w.pending == "" || w.session.State().Busy {
		return
	}
	pending := w.pending
	w.cancel()
	if pending == "close" {
		if err := w.owner.CloseWindow(w.window); err != nil {
			w.status(err.Error())
		}
		return
	}
	if err := w.session.Open(w.owner.Context(), document.OpenOptions{DiscardChanges: true}); err != nil {
		w.status(err.Error())
	}
}
func (w *documentWorkflow) complete(r document.Result) {
	if r.Err != nil {
		w.savingDecision = false
		w.status(fmt.Sprintf("%s failed: %v", r.Action, r.Err))
		return
	}
	w.status(fmt.Sprintf("%s complete: %s", r.Action, r.Path))
	if r.Action != document.Save || !w.savingDecision {
		return
	}
	w.savingDecision = false
	if w.pending == "" || w.session.State().Dirty {
		return
	}
	pending := w.pending
	w.cancel()
	if pending == "close" {
		if err := w.owner.CloseWindow(w.window); err != nil {
			w.status(err.Error())
		}
		return
	}
	if err := w.session.Open(w.owner.Context(), document.OpenOptions{}); err != nil {
		w.status(err.Error())
	}
}
func (w *documentWorkflow) view() ui.Node {
	busy := w.session.State().Busy
	message := "Save changes before closing?"
	if w.pending == "open" {
		message = "Save changes before opening another document?"
	}
	if busy {
		message = "A document operation is pending. You can cancel it or keep the window open."
	}
	return ui.Node{ID: "document.confirm", Kind: ui.Dialog, Text: "Unsaved document", Open: w.pending != "", Dismiss: w.cancel, Children: []ui.Node{
		{ID: "document.question", Kind: ui.Label, Text: message},
		{ID: "document.decisions", Kind: ui.Row, Children: []ui.Node{
			{ID: "document.save", Kind: ui.Button, Text: "Save", Disabled: busy, Activate: w.saveDecision},
			{ID: "document.discard", Kind: ui.Button, Text: "Discard", Disabled: busy, Activate: w.discard},
			{ID: "document.cancel", Kind: ui.Button, Text: "Cancel", Activate: w.cancel},
			{ID: "document.cancel-operation", Kind: ui.Button, Text: "Cancel operation", Disabled: !busy, Activate: func() { w.session.Cancel() }},
		}},
	}}
}
