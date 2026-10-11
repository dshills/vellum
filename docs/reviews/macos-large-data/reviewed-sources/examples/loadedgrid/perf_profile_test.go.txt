package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type profileShortWriter struct{}

func (profileShortWriter) Write(p []byte) (int, error) { return len(p) - 1, nil }
func TestPerfProfileWriterBudgetAndStickyFailure(t *testing.T) {
	var dst bytes.Buffer
	w := &perfProfileWriter{target: &dst, remaining: 3}
	if n, e := w.Write([]byte("ab")); n != 2 || e != nil {
		t.Fatal(n, e)
	}
	if n, e := w.Write([]byte("cd")); n != 0 || !errors.Is(e, errPerfProfileBudget) {
		t.Fatal(n, e)
	}
	if n, e := w.Write([]byte("e")); n != 0 || !errors.Is(e, errPerfProfileBudget) {
		t.Fatal(n, e)
	}
	if dst.String() != "ab" || !errors.Is(w.failure(), errPerfProfileBudget) {
		t.Fatal(dst.String(), w.failure())
	}
	short := &perfProfileWriter{target: profileShortWriter{}, remaining: 10}
	if _, e := short.Write([]byte("ab")); !errors.Is(e, io.ErrShortWrite) {
		t.Fatal(e)
	}
	if _, e := short.Write([]byte("c")); !errors.Is(e, io.ErrShortWrite) {
		t.Fatal(e)
	}
}
func TestPerfProfileExclusiveAdmissionAndDisabled(t *testing.T) {
	path := filepath.Join(t.TempDir(), "existing.pprof")
	if e := os.WriteFile(path, []byte("keep"), 0600); e != nil {
		t.Fatal(e)
	}
	if stop, e := startPerfCPUProfile(path); e == nil || stop != nil {
		t.Fatal("overwrote existing profile", e)
	}
	if data, e := os.ReadFile(path); e != nil || string(data) != "keep" {
		t.Fatal(string(data), e)
	}
	stop, e := startPerfCPUProfile("")
	if e != nil {
		t.Fatal(e)
	}
	if e = stop(); e != nil {
		t.Fatal(e)
	}
	if validPerfProfilePath("bad\x00path") || validPerfProfilePath(string(bytes.Repeat([]byte("p"), 4097))) {
		t.Fatal("invalid profile path accepted")
	}
}
