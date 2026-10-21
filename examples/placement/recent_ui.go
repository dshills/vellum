package main

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

func (m *model) recentMenu(a *app.Application, ref app.WindowRef, update func()) ui.Node {
	options := []ui.Option{}
	paths := map[string]string{}
	for _, path := range m.recent.Paths() {
		id := fmt.Sprintf("recent-%x", sha256.Sum256([]byte(path)))
		paths[id] = path
		label := []rune(filepath.Base(path) + " — " + filepath.Dir(path))
		if len(label) > 160 {
			label = append(label[:159], '…')
		}
		options = append(options, ui.Option{ID: id, Label: string(label)})
	}
	_, cacheErr := m.recent.Status()
	help := "The last ten successfully opened or saved drawings."
	if cacheErr != nil {
		help += " Cache unavailable: " + cacheErr.Error()
	}
	empty := len(options) == 0
	if empty {
		options = append(options, ui.Option{ID: "none", Label: "No recent drawings", Disabled: true})
	}
	return ui.Node{ID: "recent-drawings", Kind: ui.MenuButton, Text: "Recent", Options: options, Accessibility: ui.Accessibility{Name: "Recent drawings", Help: help}, Disabled: empty || m.files == nil || m.clipboardAvailable() != nil, Change: func(e ui.ChangeEvent) {
		path, ok := paths[e.SelectedID]
		if !ok || ref.Context().Err() != nil {
			return
		}
		if err := m.openRecent(a, ref, path); err != nil {
			m.status = err.Error()
		}
		update()
	}}
}
