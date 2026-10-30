package notes

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

// Mutation inputs are capped at 64 KiB for a bounded developer run. Production
// byte limits remain unchanged; their larger boundary regressions are separate.
func codecSeeds(f *testing.F, names ...string) {
	f.Helper()
	for _, name := range names {
		b, err := os.ReadFile(filepath.Join("testdata", "compatibility", name))
		if err != nil {
			f.Fatal(err)
		}
		f.Add(b)
	}
	for _, b := range [][]byte{nil, []byte("null"), []byte(`{}`), []byte(`{"version":1,"Version":2}`), []byte(`{"version":99}`), []byte{0xff, 0}, []byte(`{"next_id":18446744073709551615}`), []byte(`[[[[[[[[[[[[[0]]]]]]]]]]]]]`)} {
		f.Add(b)
	}
}
func FuzzDecodeNotebook(f *testing.F) {
	codecSeeds(f, "notes-build13.vellumnotes", "notes-build22.vellumnotes")
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64<<10 {
			return
		}
		b, err := DecodeNotebook(data)
		if err != nil {
			return
		}
		if err := b.Validate(); err != nil {
			t.Fatal("accepted invalid notebook", err)
		}
		encoded, err := EncodeNotebook(b)
		if err != nil {
			t.Fatal(err)
		}
		again, err := DecodeNotebook(encoded)
		if err != nil || !reflect.DeepEqual(b, again) {
			t.Fatal("unstable notebook round trip", err)
		}
		m := New()
		if err := m.journal.Replace(true, func() error { return m.installNotebook(b) }); err != nil {
			t.Fatal("accepted notebook cannot install", err)
		}
		if !slices.Equal(m.Items(), b.Notes) || m.Selected() != b.Selected || m.nextID != b.NextID || m.selectedFolder != b.Folder || m.History().Dirty || m.History().UndoCount != 0 {
			t.Fatal("install changed accepted data/history")
		}
	})
}
func FuzzDecodeRecovery(f *testing.F) {
	codecSeeds(f, "recovery-build13.vellumnotes", "recovery-build22.vellumnotes")
	f.Add([]byte(`{"version":1,"notebook":null}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64<<10 {
			return
		}
		r, err := decodeRecovery(data)
		if err != nil {
			return
		}
		encoded, err := encodeRecovery(r)
		if err != nil {
			t.Fatal(err)
		}
		again, err := decodeRecovery(encoded)
		if err != nil || !reflect.DeepEqual(r, again) {
			t.Fatal("unstable recovery", err)
		}
	})
}
func FuzzDecodeWindowSession(f *testing.F) {
	codecSeeds(f, "window-session-build22.json")
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64<<10 {
			return
		}
		s, err := decodeWindowSession(data)
		if err != nil {
			return
		}
		if err := s.validate(); err != nil {
			t.Fatal("invalid accepted session", err)
		}
		encoded, err := json.Marshal(s)
		if err != nil {
			t.Fatal(err)
		}
		again, err := decodeWindowSession(encoded)
		if err != nil || s != again {
			t.Fatal("unstable session", err)
		}
		twice, err := json.Marshal(again)
		if err != nil || !bytes.Equal(encoded, twice) {
			t.Fatal("noncanonical session", err)
		}
	})
}

func TestCodecByteLimitsRejectOtherwiseValidDocuments(t *testing.T) {
	for _, tc := range []struct {
		name   string
		seed   []byte
		limit  int
		decode func([]byte) error
	}{
		{"notebook", []byte(`{"version":1,"next_id":10,"notes":[],"selected":"","folder":"notes"}`), int(MaxNotebookBytes), func(b []byte) error { _, e := DecodeNotebook(b); return e }},
		{"recovery", []byte(`{"version":1,"notebook":null}`), int(MaxNotebookBytes), func(b []byte) error { _, e := decodeRecovery(b); return e }},
		{"window-session", compatFixture(t, "window-session-build22.json"), int(workspaceLimit), func(b []byte) error { _, e := decodeWindowSession(b); return e }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.decode(tc.seed); err != nil {
				t.Fatal("invalid size fixture", err)
			}
			b := append(bytes.Clone(tc.seed), bytes.Repeat([]byte(" "), tc.limit+1-len(tc.seed))...)
			if err := tc.decode(b); err == nil {
				t.Fatal("oversized valid JSON accepted")
			}
		})
	}
}
