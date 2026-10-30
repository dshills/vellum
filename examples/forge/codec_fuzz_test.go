package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func FuzzDecodeStudy(f *testing.F) {
	seed, err := os.ReadFile(filepath.Join("testdata", "compatibility", "study-build22.vforge"))
	if err != nil {
		f.Fatal(err)
	}
	for _, b := range [][]byte{seed, nil, []byte("null"), []byte(`{"Version":99}`), []byte(`{"Version":1,"Next":1e100}`), []byte{0xff, 0}, []byte(`{"Camera":{"Distance":-1}}`)} {
		f.Add(b)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64<<10 {
			return
		}
		s, err := decodeStudy(data)
		if err != nil {
			return
		}
		if err := validateStudy(s); err != nil {
			t.Fatal("invalid accepted study", err)
		}
		encoded, err := encodeStudy(s)
		if err != nil {
			t.Fatal(err)
		}
		again, err := decodeStudy(encoded)
		if err != nil || !reflect.DeepEqual(s, again) {
			t.Fatal("unstable study", err)
		}
		m := newWorkbench()
		if err := m.journal.Replace(true, func() error { return m.installStudy(s) }); err != nil {
			t.Fatal("accepted study cannot install", err)
		}
		if !reflect.DeepEqual(m.captureStudy(), s) || m.journal.Snapshot().Dirty || m.journal.Snapshot().UndoCount != 0 {
			t.Fatal("install changed study/history")
		}
	})
}
func FuzzDecodeWorkspace(f *testing.F) {
	for _, b := range [][]byte{[]byte(`{"Version":1,"Left":0.19,"Right":0.72}`), []byte(`{"Version":1,"Left":0.2,"Right":0.7,"Recent":[]}`), nil, []byte("null"), []byte(`{"Version":99}`), []byte(`{"Left":1e100}`), []byte{0xff, 0}} {
		f.Add(b)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64<<10 {
			return
		}
		s, err := decodeWorkspace(data)
		if err != nil {
			return
		}
		encoded, err := json.Marshal(s)
		if err != nil {
			t.Fatal(err)
		}
		again, err := decodeWorkspace(encoded)
		if err != nil {
			t.Fatal("accepted settings cannot reload", err)
		}
		twice, err := json.Marshal(again)
		if err != nil || !bytes.Equal(encoded, twice) {
			t.Fatal("unstable workspace", err)
		}
	})
}

func TestStudyAndWorkspaceByteLimits(t *testing.T) {
	study, err := os.ReadFile(filepath.Join("testdata", "compatibility", "study-build22.vforge"))
	if err != nil {
		t.Fatal(err)
	}
	study = append(study, bytes.Repeat([]byte(" "), studyLimit+1-len(study))...)
	if _, err := decodeStudy(study); err == nil {
		t.Fatal("oversized valid study accepted")
	}
	workspace := []byte(`{"Version":1,"Left":0.19,"Right":0.72}`)
	if _, err := decodeWorkspace(workspace); err != nil {
		t.Fatal("invalid workspace size fixture", err)
	}
	workspace = append(workspace, bytes.Repeat([]byte(" "), (16<<10)+1-len(workspace))...)
	if _, err := decodeWorkspace(workspace); err == nil {
		t.Fatal("oversized valid workspace accepted")
	}
}
