package notes

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func compatFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "compatibility", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func TestHistoricalNotebooksRemainReadableAndReplayable(t *testing.T) {
	for _, name := range []string{"notes-build13.vellumnotes", "notes-build22.vellumnotes"} {
		t.Run(name, func(t *testing.T) {
			original := compatFixture(t, name)
			b, err := DecodeNotebook(original)
			if err != nil {
				t.Fatal(err)
			}
			m := New()
			if err := m.journal.Replace(true, func() error { return m.installNotebook(b) }); err != nil {
				t.Fatal(err)
			}
			n := m.lookup("n10")
			if n == nil || n.value.Title != "Résumé 😀" || n.value.Body != "Historical fixture\r\ne\u0301 אבג\nLast line\n" || !n.value.Pinned || m.Selected() != "n10" || m.History().Dirty || m.History().UndoCount != 0 {
				t.Fatal("historical content/selection/history changed")
			}
			captured, err := m.CaptureNotebook()
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := EncodeNotebook(captured)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := DecodeNotebook(encoded)
			if err != nil || !reflect.DeepEqual(b, decoded) {
				t.Fatal("round trip changed historical notebook", err)
			}
			if path := os.Getenv("VELLUM_COMPAT_ROUNDTRIP"); path != "" {
				if err := os.WriteFile(filepath.Join(path, name), encoded, 0600); err != nil {
					t.Fatal(err)
				}
			}
			id, err := m.Add()
			if err != nil || id != "n11" {
				t.Fatal("historical next identity reused", id, err)
			}
			if err := m.Undo(); err != nil || m.lookup(id) != nil {
				t.Fatal("Undo", err)
			}
			if err := m.Redo(); err != nil || m.lookup(id) == nil {
				t.Fatal("Redo", err)
			}
			if !bytes.Equal(original, compatFixture(t, name)) {
				t.Fatal("fixture overwritten")
			}
		})
	}
}
func TestHistoricalRecoveryUsesExplicitRestoreAndSave(t *testing.T) {
	for _, name := range []string{"recovery-build13.vellumnotes", "recovery-build22.vellumnotes"} {
		t.Run(name, func(t *testing.T) {
			dir := privateDirectory(t)
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			original := compatFixture(t, name)
			expected, err := decodeRecovery(original)
			if err != nil || expected.Notebook == nil {
				t.Fatal("invalid historical recovery", err)
			}
			if err := os.WriteFile(filepath.Join(dir, "Recovery.vellumnotes"), original, 0600); err != nil {
				t.Fatal(err)
			}
			x := filesFixture(t, dir)
			x.owner(t, func() {
				if x.f.prompt != "recovery" || len(x.m.Items()) != 6 {
					t.Fatal("recovery silently installed")
				}
				node, ok := find(x.f.promptView(), "notes-recovery-restore")
				if !ok {
					t.Fatal("restore action absent")
				}
				node.Activate()
				if len(x.m.Items()) != 7 || !x.m.History().Dirty || x.m.History().UndoCount != 0 {
					t.Fatal("recovery was not dirty/fresh history")
				}
				if err := x.f.Save(false); err != nil {
					t.Fatal(err)
				}
			})
			x.finish(t)
			data, err := os.ReadFile(x.r.Path)
			if err != nil {
				t.Fatal(err)
			}
			b, err := DecodeNotebook(data)
			if err != nil || !reflect.DeepEqual(b, *expected.Notebook) {
				t.Fatal("saved recovery differs", err)
			}
			if !bytes.Equal(original, compatFixture(t, name)) {
				t.Fatal("historical recovery overwritten")
			}
		})
	}
}
func TestHistoricalWindowSessionCodecRoundTrip(t *testing.T) {
	b, err := decodeWindowSession(compatFixture(t, "window-session-build22.json"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	after, err := decodeWindowSession(data)
	if err != nil || b != after || b.Version != 1 || b.Activated != 1 {
		t.Fatal("session metadata changed", err)
	}
}
