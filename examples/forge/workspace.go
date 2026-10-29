package main

import (
	"context"
	"encoding/json"
	"fmt"
	vdoc "github.com/dshills/vellum/document"
	"github.com/dshills/vellum/drivers/cogent"
	"math"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type workspaceSettings struct {
	LastStudy                  string `json:",omitempty"`
	Version                    int
	Left, Right                float32
	HideBrowser, HideInspector bool
	Recent                     []string `json:",omitempty"`
}

func decodeWorkspace(b []byte) (workspaceSettings, error) {
	var s workspaceSettings
	if len(b) > 16<<10 {
		return s, fmt.Errorf("workspace size")
	}
	if e := json.Unmarshal(b, &s); e != nil {
		return s, e
	}
	if s.Version != 1 || math.IsNaN(float64(s.Left)) || math.IsNaN(float64(s.Right)) || s.Left < .1 || s.Left > .5 || s.Right < .35 || s.Right > .9 {
		return s, fmt.Errorf("workspace range")
	}
	if s.LastStudy != "" && (!filepath.IsAbs(s.LastStudy) || len(s.LastStudy) > 1024 || !strings.EqualFold(filepath.Ext(s.LastStudy), ".vforge") || !utf8.ValidString(s.LastStudy) || strings.ContainsRune(s.LastStudy, 0)) {
		return s, fmt.Errorf("invalid restored study")
	}
	if len(s.Recent) > 10 {
		return s, fmt.Errorf("recent study limit")
	}
	seen := map[string]bool{}
	for _, p := range s.Recent {
		if !filepath.IsAbs(p) || len(p) > 1024 || seen[p] {
			return s, fmt.Errorf("invalid recent study")
		}
		seen[p] = true
	}
	return s, nil
}
func workspacePath(recovery string) string {
	return filepath.Join(filepath.Dir(recovery), "workspace.json")
}
func (m *workbench) loadWorkspace() {
	m.workspaceLoaded = true
	b, e := (vdoc.FileStore{}).Read(context.Background(), workspacePath(m.recovery.path), 16<<10)
	if os.IsNotExist(e) {
		return
	}
	if e == nil {
		var s workspaceSettings
		s, e = decodeWorkspace(b)
		if e == nil {
			m.lastStudy = s.LastStudy
			m.recentStudies = append([]string(nil), s.Recent...)
			m.left = s.Left
			m.right = s.Right
			m.hideBrowser = s.HideBrowser
			m.hideInspector = s.HideInspector
			return
		}
	}
	m.preferencesBlocked = true
	m.status = "Workspace settings preserved; use Reset Workspace to replace invalid settings"
}

// Called after Run retires the owner; only bounded local file IO occurs here.
func (m *workbench) saveWorkspace(recovery string) {
	if !m.workspaceLoaded || m.preferencesBlocked {
		return
	}
	s := workspaceSettings{LastStudy: m.lastStudy, Version: 1, Left: m.left, Right: m.right, HideBrowser: m.hideBrowser, HideInspector: m.hideInspector, Recent: append([]string(nil), m.recentStudies...)}
	b, e := json.Marshal(s)
	if e == nil {
		_, e = (vdoc.FileStore{}).Write(context.Background(), workspacePath(recovery), b, 16<<10)
	}
	if e != nil {
		fmt.Fprintln(os.Stderr, "Forge workspace:", e)
	}
}
func (m *workbench) resetWorkspace() {
	m.left = .19
	m.right = .72
	m.hideBrowser = false
	m.hideInspector = false
	m.preferencesBlocked = false
	m.drag = nil
	m.isolatedIDs = nil
	m.handleMode = ""
	m.refresh()
}
func (m *workbench) bindWorkspaceCommands() {
	m.register("browser", "Show/Hide Browser", "", nil, func() error { m.hideBrowser = !m.hideBrowser; m.changed(); return nil })
	m.register("inspector", "Show/Hide Inspector", "", nil, func() error { m.hideInspector = !m.hideInspector; m.changed(); return nil })
	m.register("customize", "Customize Toolbar…", "", nil, func() error { return cogent.CustomizeTitlebarToolbar(m.ref) })
	m.register("reset-workspace", "Reset Workspace", "", nil, func() error { m.resetWorkspace(); return cogent.ResetTitlebarToolbar(m.ref) })
}
