package main

import (
	"context"
	"github.com/dshills/vellum/app"
	vdoc "github.com/dshills/vellum/document"
	"github.com/dshills/vellum/ui"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestForgeRestoresLastSuccessfulStudy(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	original := newWorkbench()
	_ = original.add()
	original.scene.Camera.Distance = 13
	path := filepath.Join(t.TempDir(), "last.vforge")
	data, _ := encodeStudy(original.captureStudy())
	if e := os.WriteFile(path, data, 0600); e != nil {
		t.Fatal(e)
	}
	r, e := openRecovery()
	if e != nil {
		t.Fatal(e)
	}
	prior := newWorkbench()
	prior.workspaceLoaded = true
	prior.lastStudy = path
	r.stop(prior)
	a, _ := app.New(app.Config{})
	defer a.Close()
	m := newWorkbench()
	m.application = a
	m.ref, e = a.OpenWindow(app.Window{ID: "forge", Title: "Forge fixture", Width: 800, Height: 600, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }})
	if e != nil {
		t.Fatal(e)
	}
	if e = m.bindFiles(); e != nil {
		t.Fatal(e)
	}
	defer func() { a.Close(); m.recovery.stop(m) }()
	end := time.Now().Add(3 * time.Second)
	for m.files.State().Busy && time.Now().Before(end) {
		_ = a.Drain()
		time.Sleep(time.Millisecond)
	}
	_ = a.Do(func() {
		if !reflect.DeepEqual(m.captureStudy(), original.captureStudy()) || m.journal.Snapshot().Dirty || m.files.State().Path != path {
			t.Fatal("study restore differs")
		}
	})
	if _, e := (vdoc.FileStore{}).Read(context.Background(), path, studyLimit); e != nil {
		t.Fatal(e)
	}
}
func TestForgeWorkspaceRejectsNonStudyRestore(t *testing.T) {
	_, e := decodeWorkspace([]byte(`{"Version":1,"Left":0.2,"Right":0.7,"LastStudy":"relative.vforge"}`))
	if e == nil {
		t.Fatal("relative restore accepted")
	}
}

func TestForgeMissingRestoredStudyPreservesDefaultAndRecord(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	missing := filepath.Join(t.TempDir(), "missing.vforge")
	r, e := openRecovery()
	if e != nil {
		t.Fatal(e)
	}
	prior := newWorkbench()
	prior.workspaceLoaded = true
	prior.lastStudy = missing
	r.stop(prior)
	a, _ := app.New(app.Config{})
	defer a.Close()
	m := newWorkbench()
	before := m.captureStudy()
	m.application = a
	m.ref, e = a.OpenWindow(app.Window{ID: "forge", Title: "Forge missing fixture", Width: 800, Height: 600, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }})
	if e != nil {
		t.Fatal(e)
	}
	if e = m.bindFiles(); e != nil {
		t.Fatal(e)
	}
	defer func() { a.Close(); m.recovery.stop(m) }()
	end := time.Now().Add(3 * time.Second)
	for m.files.State().Busy && time.Now().Before(end) {
		_ = a.Drain()
		time.Sleep(time.Millisecond)
	}
	_ = a.Do(func() {
		if !reflect.DeepEqual(before, m.captureStudy()) || m.files.State().Path != "" || m.lastStudy != missing {
			t.Fatal("failed restore changed live study")
		}
	})
}
