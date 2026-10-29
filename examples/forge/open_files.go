package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func (m *workbench) handleOpenFiles(paths []string, initial bool) {
	if len(paths) != 1 || !filepath.IsAbs(paths[0]) || len(paths[0]) > 1024 || !strings.EqualFold(filepath.Ext(paths[0]), ".vforge") || !utf8.ValidString(paths[0]) || strings.ContainsRune(paths[0], 0) {
		m.status = "Open one valid Forge study at a time"
		m.changed()
		return
	}
	if m.startupRestore && m.files != nil && m.files.State().Busy {
		if m.pendingExternalStudy != "" {
			m.status = "A Finder open is already pending"
		} else {
			m.pendingExternalStudy = paths[0]
			m.files.Cancel()
		}
		m.changed()
		return
	}
	if e := m.openStudy(paths[0], false); e != nil {
		m.status = fmt.Sprint(e)
	}
	m.changed()
}
