package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestAutomationLaunchRetainsSeparatePrivateFailureLogs(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS private app launcher")
	}
	dir := t.TempDir()
	bundle := filepath.Join(dir, "Test.app")
	binary := filepath.Join(bundle, "Contents/MacOS/Vellum Workbench")
	if err := os.MkdirAll(filepath.Dir(binary), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binary, []byte("#!/bin/sh\nexit 7\n"), 0700); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err := launchAutomation(dir, bundle); err != nil {
			t.Fatal(err)
		}
	}
	deadline := time.Now().Add(3 * time.Second)
	for {
		logs, err := filepath.Glob(filepath.Join(dir, "automation-workbench-*.log"))
		if err != nil {
			t.Fatal(err)
		}
		complete := len(logs) == 2
		for _, path := range logs {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(data), "exit status 7") {
				complete = false
			}
			info, err := os.Stat(path)
			if err != nil || info.Mode().Perm() != 0600 {
				t.Fatalf("private log permissions: %v %v", info, err)
			}
		}
		if complete {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("child failure diagnostics were not retained separately")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
