package main

import (
	"fmt"
	"github.com/dshills/vellum/app"
	vdoc "github.com/dshills/vellum/document"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestRecentStudiesDedupBoundedAndPersistent(t *testing.T) {
	m := newWorkbench()
	dir := t.TempDir()
	for i := 0; i < 15; i++ {
		m.recordRecent(filepath.Join(dir, fmt.Sprintf("%d.vforge", i)))
	}
	if len(m.recentStudies) != 10 {
		t.Fatal("bound")
	}
	path := m.recentStudies[4]
	m.recordRecent(path)
	if m.recentStudies[0] != path || len(m.recentStudies) != 10 {
		t.Fatal("MRU")
	}
	m.recovery = &recoveryStore{path: filepath.Join(dir, "recovery.vforge")}
	m.workspaceLoaded = true
	m.saveWorkspace(m.recovery.path)
	n := newWorkbench()
	n.recovery = m.recovery
	n.loadWorkspace()
	if !reflect.DeepEqual(m.recentStudies, n.recentStudies) {
		t.Fatal("recent persistence")
	}
}
func TestFilePromptEpochRejectsOldDiscard(t *testing.T) {
	m := newWorkbench()
	m.showFilePrompt("open", "/private/tmp/one.vforge")
	old := m.filePrompt()
	m.showFilePrompt("open", "/private/tmp/two.vforge")
	for _, n := range old.Children {
		if n.ID == "study-cancel" {
			n.Activate()
		}
	}
	if m.pendingStudyPath != "/private/tmp/two.vforge" || m.prompt != "open" {
		t.Fatal("stale prompt changed current operation")
	}
}

func TestRevertConfirmedReloadAndCancel(t *testing.T) {
	a, e := app.New(app.Config{})
	if e != nil {
		t.Fatal(e)
	}
	defer a.Close()
	m := newWorkbench()
	m.application = a
	m.ref, e = a.OpenWindow(app.Window{ID: "forge", Title: "Forge", Width: 800, Height: 600, View: m.View})
	if e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(t.TempDir(), "saved.vforge")
	b, _ := encodeStudy(m.captureStudy())
	if e = os.WriteFile(path, b, 0600); e != nil {
		t.Fatal(e)
	}
	m.files, e = vdoc.New(vdoc.Config[studyFile]{Owner: a, Window: "forge", Journal: m.journal, Path: path, Codec: vdoc.Codec[studyFile]{Capture: func() (studyFile, error) { return m.captureStudy(), nil }, Encode: encodeStudy, Decode: decodeStudy, Install: m.installStudy}})
	if e != nil {
		t.Fatal(e)
	}
	if e = m.add(); e != nil {
		t.Fatal(e)
	}
	if e = m.fileAction("revert", false); e != nil {
		t.Fatal(e)
	}
	dialog := m.filePrompt()
	dialog.Dismiss()
	if len(m.state.Objects) != 8 {
		t.Fatal("cancel lost changes")
	}
	_ = m.fileAction("revert", false)
	for _, n := range m.filePrompt().Children {
		if n.ID == "study-confirm" {
			n.Activate()
		}
	}
	deadline := time.Now().Add(3 * time.Second)
	for m.files.State().Busy && time.Now().Before(deadline) {
		_ = a.Drain()
		time.Sleep(time.Millisecond)
	}
	if m.files.State().Busy || m.files.State().Last.Err != nil || len(m.state.Objects) != 7 || m.journal.Snapshot().UndoCount != 0 {
		t.Fatalf("revert failed: %+v", m.files.State())
	}
}
