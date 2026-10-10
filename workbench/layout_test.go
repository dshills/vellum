package workbench

import (
	"bytes"
	"math"
	"path/filepath"
	"testing"
)

func fixture(t *testing.T) *Workspace {
	t.Helper()
	w, err := New(Layout{Version: 1, Root: &Group{ID: "split", Axis: Horizontal, Ratio: .7, First: &Group{ID: "documents", Tabs: []string{"board", "notes"}, Active: "board"}, Second: &Group{ID: "right", Tabs: []string{"inspector"}, Active: "inspector"}}}, []string{"board", "notes", "inspector"})
	if err != nil {
		t.Fatal(err)
	}
	return w
}
func TestDockingTransactionsAndPersistence(t *testing.T) {
	w := fixture(t)
	if err := w.Move("notes", "right", 1); err != nil {
		t.Fatal(err)
	}
	if err := w.Split("inspector", "documents", "inspector-dock", Vertical, false); err != nil {
		t.Fatal(err)
	}
	before, _ := w.Encode()
	if w.Resize("inspector-dock", math.NaN()) == nil {
		t.Fatal("nonfinite split")
	}
	after, _ := w.Encode()
	if !bytes.Equal(before, after) {
		t.Fatal("failed transaction mutated layout")
	}
	path := filepath.Join(t.TempDir(), "layout.json")
	if err := w.Save(path); err != nil {
		t.Fatal(err)
	}
	copy := fixture(t)
	if err := copy.Load(path); err != nil {
		t.Fatal(err)
	}
	after, _ = copy.Encode()
	if !bytes.Equal(before, after) {
		t.Fatal("roundtrip")
	}
	if w.Decode([]byte(`{"version":1,"root":null}`)) == nil {
		t.Fatal("invalid restore")
	}
}
func TestCyclesAndMissingPanels(t *testing.T) {
	w := fixture(t)
	l := w.Layout()
	l.Root.First = l.Root
	if w.Restore(l) == nil {
		t.Fatal("cycle accepted")
	}
	if w.Decode([]byte(`{"version":1,"root":{"id":"only","tabs":["board"],"active":"board"}}`)) == nil {
		t.Fatal("lost panels")
	}
}
