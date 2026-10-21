package main

import (
	"encoding/json"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/virtual"
	"math"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDocumentRoundtripDetachedGeometryReferencesAndEpoch(t *testing.T) {
	s := newParts(100000)
	board, _ := canvas.New([]canvas.Rect{{ID: "component:1", X: 2, Y: 4, Width: 25, Height: 50}})
	refs := &fileReferences{paths: []string{filepath.Join(t.TempDir(), "Ω\nfirst"), filepath.Join(t.TempDir(), "second")}}
	image := materializeDocument(captureDocument(s, board, refs))
	changed := s.parts.at(0)
	changed.Value++
	s.parts = s.parts.with(0, changed)
	board.Objects[0].X++
	refs.paths[0] = filepath.Join(t.TempDir(), "mutated")
	data, err := encodeDocument(image)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := decodeDocument(data, 2000)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(image.Parts, loaded.Parts) || !reflect.DeepEqual(image.Objects, loaded.Objects) || !reflect.DeepEqual(image.References, loaded.References) {
		t.Fatal("domain roundtrip changed")
	}
	table, _ := virtual.NewTable(s, 20)
	projection, err := virtual.Project(s.Snapshot(), "", 1, false)
	if err != nil {
		t.Fatal(err)
	}
	token := table.ProjectionToken()
	epoch := s.Revision()
	if err = installDocument(s, table, board, refs, loaded); err != nil {
		t.Fatal(err)
	}
	if s.Revision() != epoch+1 || s.parts.at(0) != image.Parts[0] || board.Objects[0] != image.Objects[0] || !reflect.DeepEqual(refs.paths, image.References) {
		t.Fatal("load didn't replace domain state")
	}
	if table.ApplyProjection(projection, token) == nil {
		t.Fatal("old document projection revived")
	}
}
func TestDocumentValidationRejectsMalformedInput(t *testing.T) {
	s := newParts(100000)
	board, _ := canvas.New([]canvas.Rect{{ID: "component:1", Width: 25, Height: 50}})
	baseline := materializeDocument(captureDocument(s, board, &fileReferences{}))
	valid, err := encodeDocument(baseline)
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*documentImage){func(i *documentImage) { i.Version = 2 }, func(i *documentImage) { i.Parts = i.Parts[:99999] }, func(i *documentImage) { i.Parts[0].Rotation = 45 }, func(i *documentImage) { i.Parts[0].Package = 2 }, func(i *documentImage) { i.Objects[0].ID = "component:2" }, func(i *documentImage) { i.Objects[0].Width = -1 }, func(i *documentImage) { i.References = []string{"relative"} }} {
		image := materializeDocument(captureDocument(s, board, &fileReferences{}))
		change(&image)
		data, err := json.Marshal(image)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = decodeDocument(data, 100); err == nil {
			t.Fatal("invalid domain accepted")
		}
	}
	for _, data := range [][]byte{append(append([]byte{}, valid...), []byte(" {}")...), []byte(`{"Version":1,"unknown":42}`), {0xff}, []byte("{")} {
		if _, err := decodeDocument(data, 100); err == nil {
			t.Fatal("malformed file accepted")
		}
	}
	loaded, err := decodeDocument(valid, 100)
	if err != nil {
		t.Fatal(err)
	}
	table, _ := virtual.NewTable(s, 20)
	beforeRoot := s.parts.root
	before := s.parts.at(0)
	refs := &fileReferences{}
	missing := loaded
	missing.prepared = partStore{}
	if err = installDocument(s, table, board, refs, missing); err == nil || s.parts.root != beforeRoot || s.parts.at(0) != before || s.Revision() != 1 || refs.revision != 0 {
		t.Fatal("missing prepared data changed model")
	}
	for _, sourceExhausted := range []bool{true, false} {
		s.revision, refs.revision = 1, 0
		if sourceExhausted {
			s.revision = math.MaxUint64
		} else {
			refs.revision = math.MaxUint64
		}
		sourceRevision, refsRevision := s.Revision(), refs.revision
		if err = installDocument(s, table, board, refs, loaded); err == nil || err.Error() != "component revision exhausted" || s.parts.root != beforeRoot || s.parts.at(0) != before || s.Revision() != sourceRevision || refs.revision != refsRevision {
			t.Fatalf("exhausted revision admitted valid prepared document (source=%t): %v", sourceExhausted, err)
		}
	}
}
