package main

import (
	"context"
	"fmt"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/ui"
)

type macFindGallery struct {
	parent *macComponentsGallery
	model  *editor.Model
	bar    *ui.FindBar
	focus  ui.FocusRequest
}

func (m *macComponentsGallery) findView() ui.Node {
	if m.find == nil {
		model, e := editor.New("A precise plan starts with a clear idea.\nThe plan includes a PLAN for every room.\nTry finding plan, replacing it with drawing, and Undo.\n", editor.Config{})
		if e != nil {
			panic(e)
		}
		f := &macFindGallery{parent: m, model: model}
		f.bar = &ui.FindBar{ID: "preview/find", Label: "Find and replace", Open: true, ReplaceOpen: true, Query: "plan", Replacement: "drawing", Model: model, IconImage: m.artwork[ui.SymbolDocument]}
		f.bar.Change = func(e ui.FindBarChange) {
			switch e.Field {
			case "query":
				f.bar.Query = e.Text
			case "replacement":
				f.bar.Replacement = e.Text
			case "case":
				f.bar.Options.CaseSensitive = e.Checked
			case "word":
				f.bar.Options.WholeWord = e.Checked
			case "replace-open":
				f.bar.ReplaceOpen = e.Checked
			}
		}
		f.bar.Request = func(q ui.FindRequest) {
			e := q.Apply(model)
			if e != nil {
				m.status = e.Error()
			} else {
				m.count++
				m.status = "Find operation complete"
			}
		}
		f.bar.Close = func() {
			f.bar.Open = false
			f.focus = ui.FocusRequest{ID: "preview/find-editor", Sequence: f.focus.Sequence + 1}
		}
		m.find = f
	}
	f := m.find
	n, e := f.bar.View()
	if e != nil {
		panic(e)
	}
	return ui.Node{ID: "preview/find-page", Kind: ui.Column, Style: ui.Style{GrowX: ui.Some(float32(1))}, Children: []ui.Node{{ID: "preview/find-help", Kind: ui.Label, Text: "⌘F opens Find; ⌘G / ⇧⌘G navigate; Escape closes. Literal Unicode search. Replace All is one undoable editor operation."}, {ID: "preview/find-controls", Kind: ui.Row, Children: []ui.Node{{ID: "preview/find-open", Kind: ui.Button, Text: "Find…", IconImage: m.artwork[ui.SymbolDocument], Activate: f.open}, {ID: "preview/find-small", Kind: ui.Checkbox, Text: "Small controls", Checked: f.bar.Small, Change: func(e ui.ChangeEvent) { f.bar.Small = e.Checked }}, {ID: "preview/find-undo", Kind: ui.Button, Text: "Undo", Activate: func() {
		if e := f.model.Journal().Undo(); e != nil {
			m.status = e.Error()
		}
	}}, {ID: "preview/find-redo", Kind: ui.Button, Text: "Redo", Activate: func() {
		if e := f.model.Journal().Redo(); e != nil {
			m.status = e.Error()
		}
	}}}}, n, {ID: "preview/find-editor", Kind: ui.MultilineText, Editor: f.model, SoftWrap: true, Style: ui.Style{GrowX: ui.Some(float32(1)), MinSize: ui.Some(ui.Size{Height: 200})}}}}
}
func (f *macFindGallery) open() {
	f.bar.Open = true
	f.focus = ui.FocusRequest{ID: ui.FindBarInputID(f.bar.ID), Sequence: f.focus.Sequence + 1}
}
func (m *macWidgetsGallery) findCommands() *command.Registry {
	r := command.New()
	active := func() error {
		if m.page != "mac-components" || m.components.page != "find-replace" {
			return fmt.Errorf("find page is inactive")
		}
		return nil
	}
	for _, entry := range []struct {
		id, label, chord string
		action           int
	}{{"find", "Find", "Meta+F", -1}, {"next", "Find Next", "Meta+G", int(ui.FindNext)}, {"previous", "Find Previous", "Meta+Shift+G", int(ui.FindPrevious)}, {"close", "Close Find", "Escape", -2}} {
		c := entry
		scope := command.WindowShortcuts
		e := r.Register(command.Command{ID: "gallery-find-" + c.id, Label: c.label, Shortcut: c.chord, ShortcutScope: scope, Available: func() error {
			if e := active(); e != nil {
				return e
			}
			if c.action == -2 && (m.components.find == nil || !m.components.find.bar.Open) {
				return fmt.Errorf("find is closed")
			}
			return nil
		}, Execute: func(context.Context) error {
			if m.components.find == nil {
				m.components.findView()
			}
			f := m.components.find
			if c.action == -1 {
				f.open()
			} else if c.action == -2 {
				f.bar.Close()
			} else {
				f.bar.Request(ui.FindRequest{Action: ui.FindAction(c.action), Query: f.bar.Query, Replacement: f.bar.Replacement, Options: f.bar.Options})
			}
			return nil
		}})
		if e != nil {
			panic(e)
		}
	}
	return r
}
