package model

import (
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestSelectionAndProjectionIsolation(t *testing.T) {
	s := New(1000)
	s.Select(17)
	if err := s.Execute("view.sort_value", "", nil); err != nil {
		t.Fatal(err)
	}
	if s.Selected().ID != 17 {
		t.Fatal("sorting changed selection")
	}
	s.SetFilter("NET_16")
	if s.Selected().ID != 17 {
		t.Fatal("filtering changed selection")
	}
	s.View[0].Value = "projection only"
	if s.Selected().Value == "projection only" {
		t.Fatal("projection mutated document")
	}
	if err := s.Execute("component.set_value", "22k", nil); err != nil {
		t.Fatal(err)
	}
	if s.Selected().Value != "22k" || s.Doc().Revision != 1 {
		t.Fatal("edit not committed")
	}
}
func TestCommandValidation(t *testing.T) {
	s := New(20)
	if s.Select(0) == nil {
		t.Fatal("invalid selection accepted")
	}
	if s.Execute("unknown", "", nil) == nil {
		t.Fatal("unknown command accepted")
	}
	if s.Execute("component.set_value", " ", nil) == nil {
		t.Fatal("empty value accepted")
	}
}
func waitTask(t *testing.T, s *Session) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for !s.Poll() {
		if time.Now().After(deadline) {
			t.Fatal("worker did not complete")
		}
		runtime.Gosched()
	}
}
func TestTaskRejectsStaleResult(t *testing.T) {
	s := New(1000)
	s.WorkDelay = 0
	if err := s.Execute("analysis.start", "", nil); err != nil {
		t.Fatal(err)
	}
	if err := s.Execute("component.rotate", "", nil); err != nil {
		t.Fatal(err)
	}
	waitTask(t, s)
	if s.LastChecksum != 0 || s.Status != "Discarded stale calculation" {
		t.Fatal(s.Status)
	}
}
func TestTaskCommitsToOriginalDocument(t *testing.T) {
	s := New(1000)
	s.WorkDelay = 0
	s.Execute("analysis.start", "", nil)
	s.Switch(1)
	waitTask(t, s)
	if s.LastChecksum == 0 || !strings.Contains(s.Status, "Board A") {
		t.Fatal(s.Status)
	}
}
func TestCancellationAndRepeatedStart(t *testing.T) {
	s := New(100000)
	s.WorkDelay = time.Millisecond
	s.Execute("analysis.start", "", nil)
	if s.Execute("analysis.start", "", nil) == nil {
		t.Fatal("duplicate task accepted")
	}
	s.Execute("analysis.cancel", "", nil)
	waitTask(t, s)
	if s.LastChecksum != 0 || !strings.Contains(s.Status, "canceled") {
		t.Fatal(s.Status)
	}
}
