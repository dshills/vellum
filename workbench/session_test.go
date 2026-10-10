package workbench

import (
	"bytes"
	"math"
	"path/filepath"
	"strings"
	"testing"
)

func sessionFixture(t *testing.T) *Session {
	t.Helper()
	s, err := NewSession(SessionLayout{Version: SessionVersion, Windows: []WindowLayout{{ID: "main", Title: "Main", Bounds: Bounds{Width: 900, Height: 640}, Root: &Group{ID: "documents", Tabs: []string{"one", "two", "three"}, Active: "one"}}}, Focus: Focus{Window: "main", Panel: "one"}}, []string{"one", "two", "three"})
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func mustEncodeSession(t *testing.T, s *Session) []byte {
	t.Helper()
	data, err := s.Encode()
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestSessionFloatRedockClosePersistence(t *testing.T) {
	s := sessionFixture(t)
	if err := s.Float("two", WindowLayout{ID: "float", Title: "Floating two", Bounds: Bounds{X: 120, Y: 80, Width: 500, Height: 400, Positioned: true}}); err != nil {
		t.Fatal(err)
	}
	if err := s.Split("three", "float", "float.tabs", "side", Vertical, true); err != nil {
		t.Fatal(err)
	}
	if err := s.Dock("two", "main", "documents", 0); err != nil {
		t.Fatal(err)
	}
	if err := s.CloseWindow("float", "main", "documents"); err != nil {
		t.Fatal(err)
	}
	l := s.Layout()
	if len(l.Windows) != 1 || len(l.Windows[0].Root.Tabs) != 3 || l.Focus.Panel != "two" {
		t.Fatal(l)
	}
	data := mustEncodeSession(t, s)
	path := filepath.Join(t.TempDir(), "session.json")
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}
	other := sessionFixture(t)
	if err := other.Load(path); err != nil {
		t.Fatal(err)
	}
	copy := mustEncodeSession(t, other)
	if !bytes.Equal(data, copy) {
		t.Fatal("session roundtrip")
	}
}
func TestSessionRejectedTransactionsLeaveEveryWindowUnchanged(t *testing.T) {
	s := sessionFixture(t)
	if err := s.Float("two", WindowLayout{ID: "float", Title: "Float", Bounds: Bounds{Width: 400, Height: 300}}); err != nil {
		t.Fatal(err)
	}
	cases := []func() error{func() error { return s.Dock("two", "main", "documents", 8) }, func() error { return s.Split("two", "main", "documents", "documents", Horizontal, false) }, func() error { return s.Resize("main", "documents", math.NaN()) }, func() error {
		return s.Float("one", WindowLayout{ID: "float", Title: "Duplicate", Bounds: Bounds{Width: 400, Height: 300}})
	}, func() error { return s.CloseWindow("main", "main", "documents") }, func() error { return s.SetBounds("main", Bounds{Width: -1, Height: 300}) }, func() error { l := s.Layout(); l.Windows[0].Root.First = l.Windows[0].Root; return s.Restore(l) }, func() error {
		return s.Decode([]byte(`{"version":2,"windows":[],"focus":{"window":"main","panel":"one"}}`))
	}, func() error { b := mustEncodeSession(t, s); return s.Decode(append(b, []byte(` {}`)...)) }}
	for i, act := range cases {
		before := mustEncodeSession(t, s)
		if err := act(); err == nil {
			t.Fatalf("case %d accepted", i)
		}
		after := mustEncodeSession(t, s)
		if !bytes.Equal(before, after) {
			t.Fatalf("case %d mutated session", i)
		}
	}
}
func TestSessionReorderAndClosedFocusedWindow(t *testing.T) {
	s := sessionFixture(t)
	if err := s.Dock("one", "main", "documents", 2); err != nil {
		t.Fatal(err)
	}
	if got := s.Layout().Windows[0].Root.Tabs; got[0] != "two" || got[2] != "one" {
		t.Fatal(got)
	}
	if err := s.Float("two", WindowLayout{ID: "float", Title: "Float", Bounds: Bounds{Width: 400, Height: 300}}); err != nil {
		t.Fatal(err)
	}
	if err := s.CloseWindow("float", "main", "documents"); err != nil {
		t.Fatal(err)
	}
	l := s.Layout()
	if l.Focus.Window != "main" || l.Focus.Panel != "two" || l.Windows[0].Root.Active != "two" {
		t.Fatal("focus transfer", l)
	}
	snapshot := s.Layout()
	snapshot.Windows[0].Root.Tabs[0] = "corrupted"
	if s.Layout().Windows[0].Root.Tabs[0] == "corrupted" {
		t.Fatal("snapshot aliases session")
	}
}
func TestSessionRejectsGlobalGroupDuplicatesAndMalformedNames(t *testing.T) {
	s := sessionFixture(t)
	l := s.Layout()
	l.Windows[0].Root.Tabs = []string{"one", "three"}
	l.Windows = append(l.Windows, WindowLayout{ID: "float", Title: "Float", Bounds: Bounds{Width: 400, Height: 300}, Root: &Group{ID: "documents", Tabs: []string{"two"}, Active: "two"}})
	if s.Restore(l) == nil {
		t.Fatal("global group collision")
	}
	l = s.Layout()
	l.Windows[0].Title = "bad\x00title"
	if s.Restore(l) == nil {
		t.Fatal("NUL title")
	}
	if s.Save(filepath.Join(t.TempDir(), "missing", "session.json")) == nil {
		t.Fatal("missing directory write succeeded")
	}
}

func TestSessionCannotEncodeBeyondItsDecodeLimit(t *testing.T) {
	s := sessionFixture(t)
	before := mustEncodeSession(t, s)
	l := s.Layout()
	l.Windows[0].Title = strings.Repeat("x", maxSessionBytes)
	if s.Restore(l) == nil {
		t.Fatal("unloadable serialized session accepted")
	}
	after := mustEncodeSession(t, s)
	if !bytes.Equal(before, after) {
		t.Fatal("oversized restore changed session")
	}
}
