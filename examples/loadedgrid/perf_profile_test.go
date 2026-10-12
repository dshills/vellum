package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func TestPerfAllocationExclusivePairAndFailure(t *testing.T) {
	for _, suffix := range []string{".before.pprof", ".after.pprof"} {
		t.Run(suffix, func(t *testing.T) {
			prefix := filepath.Join(t.TempDir(), "alloc")
			if err := os.WriteFile(prefix+suffix, []byte("keep"), 0600); err != nil {
				t.Fatal(err)
			}
			called := false
			stop, err := startPerfAllocationProfileWithCapture(prefix, func(io.Writer) error { called = true; return nil })
			if err == nil || stop != nil || called {
				t.Fatal("existing snapshot admitted", err)
			}
			data, err := os.ReadFile(prefix + suffix)
			if err != nil || string(data) != "keep" {
				t.Fatal(string(data), err)
			}
		})
	}
	failure := errors.New("snapshot failed")
	prefix := filepath.Join(t.TempDir(), "failed")
	stop, err := startPerfAllocationProfileWithCapture(prefix, func(w io.Writer) error {
		_, _ = w.Write([]byte("partial"))
		return failure
	})
	if stop != nil || !errors.Is(err, failure) {
		t.Fatal(stop != nil, err)
	}
	// Failure preserves the owned partial baseline and reserved after file.
	for name, expected := range map[string]string{".before.pprof": "partial", ".after.pprof": ""} {
		data, err := os.ReadFile(prefix + name)
		if err != nil || string(data) != expected {
			t.Fatal(name, string(data), err)
		}
	}
}

func TestPerfAllocationSnapshotsBoundedAndFinishOnce(t *testing.T) {
	for _, fail := range []bool{false, true} {
		prefix := filepath.Join(t.TempDir(), "alloc")
		calls := 0
		stop, err := startPerfAllocationProfileWithCapture(prefix, func(w io.Writer) error {
			calls++
			if fail && calls == 2 {
				// Ignoring an underlying write error must still fail completion.
				_, _ = w.Write(make([]byte, perfProfileMaxBytes+1))
				return nil
			}
			_, e := w.Write([]byte{byte(calls)})
			return e
		})
		if err != nil || calls != 1 {
			t.Fatal(calls, err)
		}
		for i := 0; i < 2; i++ {
			err = stop()
			if fail && !errors.Is(err, errPerfProfileBudget) || !fail && err != nil {
				t.Fatal(err)
			}
		}
		if calls != 2 {
			t.Fatal("finish repeated capture", calls)
		}
		for _, suffix := range []string{".before.pprof", ".after.pprof"} {
			info, err := os.Stat(prefix + suffix)
			if err != nil || info.Mode().Perm() != 0600 || info.Size() > perfProfileMaxBytes {
				t.Fatal(info, err)
			}
		}
	}
}

func TestPerfAllocationDisabledInvalidAndRuntimeSnapshot(t *testing.T) {
	stop, err := startPerfAllocationProfileWithCapture("", func(io.Writer) error { t.Fatal("disabled capture"); return nil })
	if err != nil || stop() != nil {
		t.Fatal(err)
	}
	for _, invalid := range []string{"bad\x00prefix", strings.Repeat("p", 4096-len(".before.pprof")+1)} {
		stop, err = startPerfAllocationProfile(invalid)
		if stop != nil || err == nil {
			t.Fatal("invalid prefix admitted", err)
		}
	}
	prefix := filepath.Join(t.TempDir(), "real")
	rate := runtime.MemProfileRate
	stop, err = startPerfAllocationProfile(prefix)
	if err != nil || stop() != nil || runtime.MemProfileRate != rate {
		t.Fatal("runtime capture/rate", err)
	}
	for _, suffix := range []string{".before.pprof", ".after.pprof"} {
		data, err := os.ReadFile(prefix + suffix)
		if err != nil || len(data) < 2 || data[0] != 0x1f || data[1] != 0x8b {
			t.Fatal("missing gzip pprof snapshot", err)
		}
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
