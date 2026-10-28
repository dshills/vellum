package main

import (
	"errors"
	"fmt"
	vdoc "github.com/dshills/vellum/document"
	"github.com/dshills/vellum/ui"
	"path/filepath"
	"slices"
	"strings"
)

func (m *workbench) recordRecent(path string) {
	if !filepath.IsAbs(path) || len(path) > 1024 || strings.ContainsRune(path, 0) {
		return
	}
	path = filepath.Clean(path)
	out := []string{path}
	for _, p := range m.recentStudies {
		if p != path && len(out) < 10 {
			out = append(out, p)
		}
	}
	m.recentStudies = out
}
func (m *workbench) showFilePrompt(kind, path string) {
	m.prompt = kind
	m.pendingStudyPath = path
	m.fileEpoch++
	m.changed()
}
func (m *workbench) openStudy(path string, discard bool) error {
	if m.files == nil {
		return fmt.Errorf("file services unavailable")
	}
	if m.clipboardReady() != nil {
		return vdoc.ErrBusy
	}
	e := m.files.Open(m.ref.Context(), vdoc.OpenOptions{Path: path, DiscardChanges: discard})
	if errors.Is(e, vdoc.ErrUnsaved) {
		m.showFilePrompt("open", path)
		return nil
	}
	return e
}
func (m *workbench) bindRecentCommands() {
	m.register("recent", "Recent Studies…", "", m.clipboardReady, func() error { m.recentOpen = true; m.fileEpoch++; m.changed(); return nil })
	m.register("revert", "Revert to Saved…", "", func() error {
		if e := m.clipboardReady(); e != nil {
			return e
		}
		if m.files.State().Path == "" {
			return fmt.Errorf("save first")
		}
		return nil
	}, func() error { return m.fileAction("revert", false) })
}
func (m *workbench) recentPrompt() ui.Node {
	epoch := m.fileEpoch
	valid := func() bool {
		return m.recentOpen && epoch == m.fileEpoch && (m.application == nil || m.ref.Context().Err() == nil)
	}
	cancel := func() {
		if valid() {
			m.recentOpen = false
			m.fileEpoch++
			m.changed()
		}
	}
	children := []ui.Node{}
	if len(m.recentStudies) == 0 {
		children = append(children, text("recent-empty", "No recent studies", 12, 400, rgb(117, 126, 141)))
	}
	for i, path := range m.recentStudies {
		path := path
		n := button(fmt.Sprintf("recent-study-%d", i), filepath.Base(path), func() {
			if !valid() || !slices.Contains(m.recentStudies, path) {
				return
			}
			cancel()
			m.result(m.openStudy(path, false))
		})
		n.Tooltip = path
		n.IconImage = artwork("open", rgb(236, 168, 52))
		children = append(children, n)
	}
	children = append(children, button("recent-clear", "Clear List", func() {
		if valid() {
			m.recentStudies = nil
			cancel()
		}
	}), button("recent-cancel", "Cancel", cancel))
	return ui.Node{ID: "recent-studies-dialog", Kind: ui.Dialog, Text: "Recent Studies", Open: m.recentOpen, Dismiss: cancel, Children: children}
}
