package main

import (
	"fmt"
	"github.com/dshills/vellum/ui"
	"strings"
	"unicode"
	"unicode/utf8"
)

func (m *workbench) renameBody(id, name string) error {
	name = strings.TrimSpace(name)
	if !utf8.ValidString(name) || utf8.RuneCountInString(name) < 1 || utf8.RuneCountInString(name) > 128 {
		return fmt.Errorf("use a name of 1–128 characters")
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return fmt.Errorf("name cannot contain control characters")
		}
	}
	s := clone(m.state)
	found := false
	changed := false
	for i := range s.Objects {
		if s.Objects[i].ID == id {
			changed = changed || s.Objects[i].Name != name
			s.Objects[i].Name = name
			found = true
		}
	}
	if !found {
		return fmt.Errorf("body no longer exists")
	}
	for i := range s.Features {
		if s.Features[i].Object == id {
			changed = changed || s.Features[i].Label != name
			s.Features[i].Label = name
		}
	}
	if !changed {
		return nil
	}
	return m.apply("Rename body", s)
}
func (m *workbench) beginRename(id string) {
	if m.clipboardReady() != nil {
		return
	}
	for _, o := range m.state.Objects {
		if o.ID == id {
			m.renameID = id
			m.renameText = o.Name
			m.renameError = ""
			m.renameEpoch++
			m.changed()
			return
		}
	}
}
func (m *workbench) renamePrompt() ui.Node {
	id, epoch := m.renameID, m.renameEpoch
	valid := func() bool {
		return id != "" && m.renameID == id && m.renameEpoch == epoch && (m.application == nil || m.ref.Context().Err() == nil)
	}
	cancel := func() {
		if valid() {
			m.renameID = ""
			m.renameEpoch++
			m.changed()
		}
	}
	return ui.Node{ID: "rename-body-dialog", Kind: ui.Dialog, Text: "Rename Body", Open: id != "", Dismiss: cancel, Children: []ui.Node{{ID: "rename-body-name", Kind: ui.TextField, Text: m.renameText, Accessibility: ui.Accessibility{Name: "Body name"}, Change: func(e ui.ChangeEvent) {
		if valid() {
			m.renameText = e.Text
			m.renameError = ""
			m.changed()
		}
	}}, text("rename-body-error", m.renameError, 11, 400, rgb(188, 55, 55)), button("rename-body-apply", "Rename", func() {
		if !valid() {
			return
		}
		if e := m.renameBody(id, m.renameText); e != nil {
			m.renameError = e.Error()
			m.changed()
			return
		}
		cancel()
	}), button("rename-body-cancel", "Cancel", cancel)}}
}
func (m *workbench) browserContext(n ui.Node) ui.Node {
	for _, o := range m.state.Objects {
		if n.ID == ui.ModelBrowserRowID("browser", o.ID)+"/select" {
			id := o.ID
			n.ContextMenu = &ui.ContextMenu{Options: []ui.Option{{ID: "rename", Label: "Rename…"}}, Change: func(e ui.ChangeEvent) {
				if e.SelectedID == "rename" {
					m.beginRename(id)
				}
			}}
		}
	}
	for i := range n.Children {
		n.Children[i] = m.browserContext(n.Children[i])
	}
	return n
}
