package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecoveryFailuresRemainVisibleUntilSuccessfulWrite(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "blocked")
	if e := os.Mkdir(p, 0700); e != nil {
		t.Fatal(e)
	}
	r := &recoveryStore{path: p}
	m := newWorkbench()
	r.process(m.captureStudy())
	if r.warning() == "" {
		t.Fatal("write error hidden")
	}
	r.path = filepath.Join(dir, "recovery.vforge")
	r.process(m.captureStudy())
	if r.warning() != "" {
		t.Fatal(r.warning())
	}
	r.process(studyFile{Version: 9})
	if r.warning() == "" {
		t.Fatal("encoding error hidden")
	}
	r.process(studyFile{})
	if r.warning() != "" {
		t.Fatal(r.warning())
	}
}
