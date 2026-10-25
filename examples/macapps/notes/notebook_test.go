package notes

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestNotebookUnicodeIdentityAndHistoryRoundTrip(t *testing.T) {
	m := New()
	id, err := m.Add()
	if err != nil {
		t.Fatal(err)
	}
	text := "A😀e\u0301 אבג العربية\r\nterminal\r\n"
	if err = m.Update(id, "Résumé 😀", text); err != nil {
		t.Fatal(err)
	}
	b, err := m.CaptureNotebook()
	if err != nil {
		t.Fatal(err)
	}
	data, err := EncodeNotebook(b)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeNotebook(data)
	if err != nil || !reflect.DeepEqual(b, decoded) {
		t.Fatal(err)
	}
	other := New()
	if err = other.journal.Replace(true, func() error { return other.installNotebook(decoded) }); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m.Items(), other.Items()) || other.Selected() != id || other.History().Dirty || other.History().UndoCount != 0 {
		t.Fatal("installed state/history")
	}
	next, err := other.Add()
	if err != nil || next == id {
		t.Fatal("identity reuse", next, err)
	}
	if err = other.Undo(); err != nil {
		t.Fatal(err)
	}
	if err = other.Redo(); err != nil || other.Selected() != next {
		t.Fatal("replay", err)
	}
	if other.lookup(id).editor.Snapshot().Text != text {
		t.Fatal("Unicode changed")
	}
}
func TestNotebookMalformedDocumentsNeverChangeLiveModel(t *testing.T) {
	m := New()
	b, _ := m.CaptureNotebook()
	good, _ := EncodeNotebook(b)
	before := m.Items()
	cases := [][]byte{nil, []byte("null"), []byte(`{"version":1,"Version":1}`), append(append([]byte{}, good...), good...), bytes.Replace(good, []byte(`"version": 1`), []byte(`"version": 2`), 1), bytes.Replace(good, []byte(`"version": 1`), []byte(`"unknown": 1, "version": 1`), 1), []byte{0xff}}
	for _, data := range cases {
		if _, err := DecodeNotebook(data); err == nil {
			t.Fatalf("accepted %q", data)
		}
		if !reflect.DeepEqual(before, m.Items()) {
			t.Fatal("decode changed live model")
		}
	}
	bad := b
	bad.Notes = append([]Note(nil), b.Notes...)
	bad.Notes[1].ID = bad.Notes[0].ID
	if err := m.installNotebook(bad); err == nil || !reflect.DeepEqual(before, m.Items()) {
		t.Fatal("partial invalid install")
	}
	bad = b
	bad.Selected = "missing"
	if bad.Validate() == nil {
		t.Fatal("missing selection")
	}
	bad = b
	bad.Notes = append([]Note(nil), b.Notes...)
	bad.Notes[0].Body = strings.Repeat("x", MaxBodyBytes+1)
	if bad.Validate() == nil {
		t.Fatal("oversized body")
	}
}
func TestRepositoryRefusesConcurrentAndMalformedStorage(t *testing.T) {
	dir := privateDirectory(t)
	r, err := OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}
	if other, err := OpenRepository(dir); err == nil {
		other.Close()
		t.Fatal("second writer admitted")
	}
	bad := []byte(`{"version":99}`)
	if err = os.WriteFile(r.Path, bad, 0600); err != nil {
		t.Fatal(err)
	}
	m := New()
	before := m.Items()
	if _, err = r.Load(m); err == nil || !reflect.DeepEqual(before, m.Items()) {
		t.Fatal("invalid startup changed model")
	}
	retained, _ := os.ReadFile(r.Path)
	if !bytes.Equal(retained, bad) {
		t.Fatal("invalid file overwritten")
	}
	if err = r.Close(); err != nil {
		t.Fatal(err)
	}
	r, err = OpenRepository(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
}
func TestRepositoryRejectsSymlinkDirectory(t *testing.T) {
	base := t.TempDir()
	target := filepath.Join(base, "private")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(base, "alias")
	if err := os.Symlink(target, link); err != nil {
		t.Skip(err)
	}
	if r, err := OpenRepository(link); err == nil {
		r.Close()
		t.Fatal("symlink repository admitted")
	}
}
func TestRecoveryEnvelopeCannotMasqueradeAsNotebook(t *testing.T) {
	m := New()
	b, _ := m.CaptureNotebook()
	data, _ := encodeRecovery(Recovery{Version: 1, Notebook: &b})
	if _, err := DecodeNotebook(data); err == nil {
		t.Fatal("recovery opened as normal notebook")
	}
	marker, _ := encodeRecovery(Recovery{Version: 1})
	r, err := decodeRecovery(marker)
	if err != nil || r.Notebook != nil {
		t.Fatal("clean marker", err)
	}
	var raw map[string]any
	if err = json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
}

func privateDirectory(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "private-notes")
}
