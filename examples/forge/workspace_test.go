package main

import (
	"context"
	vdoc "github.com/dshills/vellum/document"
	"github.com/dshills/vellum/ui"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspacePersistenceAndReset(t *testing.T) {
	m := newWorkbench()
	m.workspaceLoaded = true
	path := filepath.Join(t.TempDir(), "recovery.vforge")
	m.recovery = &recoveryStore{path: path}
	m.left = .27
	m.right = .61
	m.hideBrowser = true
	m.saveWorkspace(path)
	m.left = .19
	m.right = .72
	m.hideBrowser = false
	m.loadWorkspace()
	if m.left != .27 || m.right != .61 || !m.hideBrowser {
		t.Fatal("workspace roundtrip")
	}
	m.resetWorkspace()
	if m.left != .19 || m.right != .72 || m.hideBrowser || m.hideInspector {
		t.Fatal("reset")
	}
	m.saveWorkspace(path)
	if _, e := (vdoc.FileStore{}).Read(context.Background(), workspacePath(path), 4096); e != nil {
		t.Fatal(e)
	}
}
func TestInvalidWorkspacePreserved(t *testing.T) {
	m := newWorkbench()
	path := filepath.Join(t.TempDir(), "recovery.vforge")
	m.recovery = &recoveryStore{path: path}
	bad := []byte(`{"Version":99}`)
	if e := os.WriteFile(workspacePath(path), bad, 0600); e != nil {
		t.Fatal(e)
	}
	m.loadWorkspace()
	m.saveWorkspace(path)
	b, _ := os.ReadFile(workspacePath(path))
	if string(b) != string(bad) || !m.preferencesBlocked {
		t.Fatal("invalid preferences overwritten")
	}
}

func TestAllPanelVisibilityLayouts(t *testing.T) {
	m := newWorkbench()
	for _, b := range []bool{false, true} {
		for _, i := range []bool{false, true} {
			m.hideBrowser = b
			m.hideInspector = i
			if e := ui.Validate(m.View()); e != nil {
				t.Fatalf("browser %v inspector %v: %v", b, i, e)
			}
		}
	}
}
