// Measure Pad is an independent application grown from scripts/new-app.py.
package main

import (
	"context"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
	"os"
)

func run() error {
	a, e := app.New(app.Config{})
	if e != nil {
		return e
	}
	defer a.Close()
	m := newModel()
	registry := command.New()
	var files *document.Session[notebook]
	var ref app.WindowRef
	closing := false
	ready := func() error {
		if files == nil || files.State().Busy || closing {
			return document.ErrBusy
		}
		return nil
	}
	action := func(id, label, key string, scope command.ShortcutScope, f func() error) error {
		return registry.Register(command.Command{ID: id, Label: label, Shortcut: key, ShortcutScope: scope, Available: func() error {
			if e := ready(); e != nil {
				return e
			}
			s := m.journal.Snapshot()
			switch id {
			case "delete":
				if m.selected == "" {
					return fmt.Errorf("select a measurement")
				}
			case "undo":
				if s.UndoCount == 0 {
					return history.ErrEmpty
				}
			case "redo":
				if s.RedoCount == 0 {
					return history.ErrEmpty
				}
			}
			return nil
		}, Execute: func(context.Context) error { e := f(); m.result(e); return e }})
	}
	for _, v := range []struct {
		id, label, key string
		scope          command.ShortcutScope
		f              func() error
	}{
		{"add", "Add measurement", "", command.WindowShortcuts, m.add}, {"delete", "Delete", "Backspace", command.AfterFocusShortcuts, m.remove},
		{"undo", "Undo", "Meta+Z", command.AfterFocusShortcuts, m.journal.Undo}, {"redo", "Redo", "Meta+Shift+Z", command.AfterFocusShortcuts, m.journal.Redo},
		{"save", "Save…", "Meta+S", command.WindowShortcuts, func() error { return files.Save(ref.Context(), document.SaveOptions{}) }},
		{"open", "Open…", "Meta+O", command.WindowShortcuts, func() error { return files.Open(ref.Context(), document.OpenOptions{}) }},
	} {
		if e := action(v.id, v.label, v.key, v.scope, v.f); e != nil {
			return e
		}
	}
	m.changed = func() { _ = a.Invalidate("measurepad") }
	ref, e = a.OpenWindow(app.Window{ID: "measurepad", Title: "Measure Pad", Width: 640, Height: 640, CloseRequest: func() bool {
		if files != nil && files.State().Busy {
			files.Cancel()
			m.status = "Canceling file operation; close again when complete"
			m.changed()
			return false
		}
		if m.journal.Snapshot().Dirty {
			closing = true
			m.changed()
			return false
		}
		return true
	}, View: func() ui.Node {
		buttons := []ui.Node{}
		for _, id := range []string{"add", "delete", "undo", "redo", "open", "save"} {
			id := id
			c, _ := registry.Lookup(id)
			buttons = append(buttons, ui.Node{ID: id, Kind: ui.Button, Text: c.Label, Disabled: registry.Enabled(id) != nil, Activate: func() { _ = registry.Invoke(ref.Context(), id) }})
		}
		n := m.view(ui.Node{ID: "actions", Kind: ui.Row, Children: buttons})
		if closing {
			n.Children = append(n.Children, ui.Node{ID: "close-confirm", Kind: ui.Dialog, Text: "Unsaved measurements", Open: true, Dismiss: func() { closing = false; m.changed() }, Children: []ui.Node{{ID: "close-help", Kind: ui.Label, Text: "Cancel and Save to keep your changes, or discard them."}, {ID: "cancel-close", Kind: ui.Button, Text: "Cancel", Activate: func() { closing = false; m.changed() }}, {ID: "discard-close", Kind: ui.Button, Text: "Discard and Close", Activate: func() { _ = a.CloseWindow(ref) }}}})
		}
		return n
	}})
	if e != nil {
		return e
	}
	files, e = document.New(document.Config[notebook]{Owner: a, Window: ref.ID(), Journal: m.journal, Filename: "Measurements.vmeasure", Extensions: []string{"vmeasure"}, MaxBytes: 65536, Codec: document.Codec[notebook]{Capture: func() (notebook, error) { return clone(m.book), nil }, Encode: encode, Decode: decode, Install: m.install}, OnComplete: func(r document.Result) { m.result(r.Err) }})
	if e != nil {
		return e
	}
	return a.Run(context.Background(), cogent.Driver{FollowSystemAppearance: true, Commands: registry, CommandMenus: []app.MenuGroup{{Label: "File", Commands: []string{"open", "save"}}, {Label: "Edit", Commands: []string{"undo", "redo", "delete"}}}, OnError: func(e error) { fmt.Fprintln(os.Stderr, e) }})
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
