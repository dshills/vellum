package main

import (
	"context"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/ui"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNotebookEditingRoundTripAndHistory(t *testing.T) {
	m := newModel()
	m.amount = "2 in"
	if e := m.add(); e != nil {
		t.Fatal(e)
	}
	if m.book.Entries[0].Millimeters != 50.8 {
		t.Fatal("conversion")
	}
	saved := clone(m.book)
	data, e := encode(saved)
	if e != nil {
		t.Fatal(e)
	}
	b, e := decode(data)
	if e != nil || !reflect.DeepEqual(saved, b) {
		t.Fatal("round trip", e)
	}
	if e = m.journal.Undo(); e != nil || len(m.book.Entries) != 0 {
		t.Fatal("undo", e)
	}
	if e = m.journal.Redo(); e != nil || !reflect.DeepEqual(saved, m.book) {
		t.Fatal("redo", e)
	}
	if e = ui.Validate(m.view()); e != nil {
		t.Fatal(e)
	}
	m.selected = "1"
	if e = m.remove(); e != nil || len(m.book.Entries) != 0 {
		t.Fatal("delete", e)
	}
}
func TestInvalidInputPreservesNotebook(t *testing.T) {
	m := newModel()
	before := clone(m.book)
	for _, text := range []string{"invalid", "-1 mm", "0", "1000001 mm"} {
		m.amount = text
		if m.add() == nil {
			t.Fatal("invalid accepted", text)
		}
		if !reflect.DeepEqual(before, m.book) {
			t.Fatal("partial change")
		}
	}
	if _, e := decode([]byte(`{"Version":2,"Next":1}`)); e == nil {
		t.Fatal("version accepted")
	}
}

func TestFileStoreRoundTrip(t *testing.T) {
	m := newModel()
	if e := m.add(); e != nil {
		t.Fatal(e)
	}
	data, e := encode(m.book)
	if e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(t.TempDir(), "Measurements.vmeasure")
	store := document.FileStore{}
	if written, e := store.Write(context.Background(), path, data, 65536); e != nil || !written {
		t.Fatal("save", e)
	}
	got, e := store.Read(context.Background(), path, 65536)
	if e != nil {
		t.Fatal(e)
	}
	loaded, e := decode(got)
	if e != nil || !reflect.DeepEqual(loaded, m.book) {
		t.Fatal("file round trip", e)
	}
}
