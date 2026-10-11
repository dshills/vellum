// Editor is an independent document consumer of Vellum's public multiline control.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/ui"
)

type editorWindow struct {
	application                               *app.Application
	id                                        string
	lifetime                                  context.Context
	model                                     *editor.Model
	session                                   *document.Session[string]
	store                                     document.Store
	settle                                    settleFunc
	settleCancel                              context.CancelFunc
	settling, cancelRequested, closeAfterSave bool
	operation, resultCount, focusSequence     uint64
	searchChanges                             uint64
	status, prompt, openPath, editorID        string
	focusID, search                           string
	last                                      document.Result
}
type windowOptions struct {
	config editor.Config
	text   string
	settle settleFunc
	store  document.Store
	choose func(context.Context, app.FileOptions) (string, error)
}

func newEditorWindow(a *app.Application, id string, o windowOptions) (*editorWindow, error) {
	m, err := editor.New(o.text, o.config)
	if err != nil {
		return nil, err
	}
	w := &editorWindow{application: a, id: id, model: m, settle: o.settle, store: o.store, status: "Ready", editorID: "document", focusID: "document", focusSequence: 1}
	if w.settle == nil {
		w.settle = cogent.SettleEditor
	}
	if err = a.Open(app.Window{ID: id, Title: "Vellum document editor — " + id, Width: 840, Height: 660, View: w.view, CloseRequest: func() bool { w.start(actionClose, ""); return false }}); err != nil {
		return nil, err
	}
	w.lifetime, err = a.WindowContext(id)
	if err != nil {
		return nil, err
	}
	w.session, err = document.New(document.Config[string]{Owner: a, Window: id, Journal: m.Journal(), Codec: documentCodec(m), MaxBytes: int64(m.MaxBytes()), Filename: "Untitled.txt", Extensions: []string{"txt", "md", "go"}, Store: o.store, Choose: o.choose, OnComplete: w.finish})
	if err != nil {
		_ = a.CloseWindow(id)
		return nil, err
	}
	return w, nil
}
func (w *editorWindow) dismissPrompt() {
	w.prompt = ""
	w.openPath = ""
	w.focusID = w.editorID
	w.focusSequence++
	w.status = "Document kept"
	w.invalidate()
}
func (w *editorWindow) searchChanged(e ui.ChangeEvent) {
	w.search = e.Text
	w.searchChanges++
	if w.search == "" {
		w.status = "Find text cleared"
	} else {
		w.status = fmt.Sprintf("Find: %d matches", strings.Count(w.model.Snapshot().Text, w.search))
	}
	w.invalidate()
}
func (w *editorWindow) view() ui.Node {
	state := w.session.State()
	snapshot := w.model.Snapshot()
	path := state.Path
	if path == "" {
		path = "Untitled"
	} else {
		path = filepath.Base(path)
	}
	if state.Dirty {
		path += " • Unsaved"
	}
	busy := w.settling || state.Busy
	root := ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: w.focusID, Sequence: w.focusSequence}, Children: []ui.Node{
		{ID: "tools", Kind: ui.Row, Children: []ui.Node{
			{ID: "open", Kind: ui.Button, Text: "Open…", Disabled: busy, Activate: func() { w.start(actionOpen, "") }},
			{ID: "save", Kind: ui.Button, Text: "Save", Disabled: busy, Activate: func() { w.start(actionSave, "") }},
			{ID: "save-as", Kind: ui.Button, Text: "Save As…", Disabled: busy, Activate: func() { w.start(actionSaveAs, "") }},
			{ID: "undo", Kind: ui.Button, Text: "Undo", Disabled: busy || snapshot.History.UndoCount == 0 && !snapshot.Composing, Activate: func() { w.start(actionUndo, "") }},
			{ID: "redo", Kind: ui.Button, Text: "Redo", Disabled: busy || snapshot.History.RedoCount == 0, Activate: func() { w.start(actionRedo, "") }},
			{ID: "cancel-operation", Kind: ui.Button, Text: "Cancel operation", Disabled: !busy, Activate: w.cancel},
		}},
		{ID: "find-row", Kind: ui.Row, Children: []ui.Node{{ID: "find-label", Kind: ui.Label, Text: "Find"}, {ID: "find", Kind: ui.TextField, Text: w.search, MinWidth: 200, Change: w.searchChanged, Accessibility: ui.Accessibility{Name: "Find text"}}}},
		{ID: "filename", Kind: ui.Label, Text: path, MinHeight: 24},
		{ID: w.editorID, Kind: ui.MultilineText, Editor: w.model, Grow: true, TabInserts: true, SoftWrap: true, Change: func(ui.ChangeEvent) { w.status = "Document edited"; w.invalidate() }, Accessibility: ui.Accessibility{Name: "Document text"}},
		{ID: "status", Kind: ui.Label, Text: w.status, MinHeight: 24},
	}}
	dialog := ui.Node{ID: "decision", Kind: ui.Dialog, Open: w.prompt != "", Dismiss: w.dismissPrompt, Children: []ui.Node{{ID: "decision-text", Kind: ui.Label, Text: "This document has unsaved changes."}}}
	if w.prompt == "open" {
		dialog.Children = append(dialog.Children, ui.Node{ID: "discard-open", Kind: ui.Button, Text: "Discard and Open…", Activate: func() { path := w.openPath; w.prompt = ""; w.start(actionDiscardOpen, path) }})
	} else {
		dialog.Children = append(dialog.Children, ui.Node{ID: "save-close", Kind: ui.Button, Text: "Save and Close", Activate: func() { w.prompt = ""; w.start(actionSaveClose, "") }}, ui.Node{ID: "discard-close", Kind: ui.Button, Text: "Discard and Close", Activate: func() { w.prompt = ""; w.start(actionDiscardClose, "") }})
	}
	dialog.Children = append(dialog.Children, ui.Node{ID: "keep", Kind: ui.Button, Text: "Cancel", Activate: w.dismissPrompt})
	root.Children = append(root.Children, dialog)
	return root
}
func run(smoke bool) error {
	a, err := app.New(app.Config{})
	if err != nil {
		return err
	}
	defer a.Close()
	var fixtureStore document.Store
	if smoke {
		fixtureStore = &smokeStore{}
	}
	first, err := newEditorWindow(a, "editor-one", windowOptions{store: fixtureStore})
	if err != nil {
		return err
	}
	second, err := newEditorWindow(a, "editor-two", windowOptions{config: func() editor.Config {
		if smoke {
			return editor.Config{MaxBytes: 64}
		}
		return editor.Config{}
	}()})
	if err != nil {
		return err
	}
	result := make(chan error, 1)
	if smoke {
		go func() {
			ctx, cancel := context.WithTimeout(a.Context(), 60*time.Second)
			defer cancel()
			result <- editorSmoke(ctx, a, first, second)
			a.Close()
		}()
	}
	nativeErr := a.Run(context.Background(), cogent.Driver{OnError: reportNativeError})
	if !smoke {
		return nativeErr
	}
	select {
	case workerErr := <-result:
		return errors.Join(nativeErr, workerErr)
	default:
	}
	if nativeErr != nil {
		return nativeErr
	}
	select {
	case workerErr := <-result:
		return workerErr
	case <-time.After(2 * time.Second):
		return errors.New("editor smoke result missing after native shutdown")
	}
}

// runAXFixture is a separate bounded observation mode for the external macOS
// accessibility inspector. It exposes known text and a scalar emoji selection;
// it does not grant permissions, change settings, or synthesize user input.
func runAXFixture() error {
	a, err := app.New(app.Config{})
	if err != nil {
		return err
	}
	defer a.Close()
	text := "A😀e\u0301 אבג العربية\r\n" + strings.Repeat("wrapped paragraph ", 24) + "\r\nterminal\r\n\r\n"
	w, err := newEditorWindow(a, "editor-ax", windowOptions{text: text})
	if err != nil {
		return err
	}
	if err := w.model.Select(editor.Selection{Anchor: 1, Focus: 2}); err != nil {
		return err
	}
	fmt.Printf("EDITOR_AX_FIXTURE %s\n", mustJSON(map[string]any{"pid": os.Getpid(), "window": w.id, "title": "Vellum document editor — " + w.id, "accessibility_name": "Document text", "text": text, "selection_scalar": [2]int{1, 2}, "selection_utf16": [2]int{1, 2}, "lifetime_seconds": 60}))
	timer := time.AfterFunc(60*time.Second, a.Close)
	defer timer.Stop()
	return a.Run(context.Background(), cogent.Driver{OnError: reportNativeError})
}

// The driver invokes OnError under the application owner. Logging is deliberately
// non-reentrant: it does not read models or invoke application methods.
func reportNativeError(err error) {
	fmt.Fprintln(os.Stderr, "editor native:", err)
}

func main() {
	smoke := flag.Bool("smoke", false, "run deterministic native document/editor acceptance")
	axFixture := flag.Bool("ax-fixture", false, "open one known document for a bounded external accessibility inspector")
	flag.Parse()
	if *smoke && *axFixture {
		fmt.Fprintln(os.Stderr, "choose only one fixture mode")
		os.Exit(1)
	}
	if *axFixture {
		if err := runAXFixture(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *smoke {
		timer := time.AfterFunc(75*time.Second, func() { fmt.Fprintln(os.Stderr, "editor smoke watchdog"); os.Exit(1) })
		defer timer.Stop()
	}
	if err := run(*smoke); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
