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
	refs := &fileReferences{paths: []string{filepath.Join(t.TempDir(), "Ω\nfirst"), "/tmp/second"}}
	image := captureDocument(s, board, refs)
	s.parts[0].Value++
	board.Objects[0].X++
	refs.paths[0] = "/tmp/mutated"
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
	if s.Revision() != epoch+1 || s.parts[0] != image.Parts[0] || board.Objects[0] != image.Objects[0] || !reflect.DeepEqual(refs.paths, image.References) {
		t.Fatal("load didn't replace domain state")
	}
	if table.ApplyProjection(projection, token) == nil {
		t.Fatal("old document projection revived")
	}
}
func TestDocumentValidationRejectsMalformedInput(t *testing.T) {
	s := newParts(100000)
	board, _ := canvas.New([]canvas.Rect{{ID: "component:1", Width: 25, Height: 50}})
	baseline := captureDocument(s, board, &fileReferences{})
	valid, err := encodeDocument(baseline)
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*documentImage){func(i *documentImage) { i.Version = 2 }, func(i *documentImage) { i.Parts = i.Parts[:99999] }, func(i *documentImage) { i.Parts[0].Rotation = 45 }, func(i *documentImage) { i.Parts[0].Package = 2 }, func(i *documentImage) { i.Objects[0].ID = "component:2" }, func(i *documentImage) { i.Objects[0].Width = -1 }, func(i *documentImage) { i.References = []string{"relative"} }} {
		image := captureDocument(s, board, &fileReferences{})
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
	table, _ := virtual.NewTable(s, 20)
	s.revision = math.MaxUint64
	before := s.parts[0]
	if err = installDocument(s, table, board, &fileReferences{}, documentImage{outline: baseline.outline}); err == nil || s.parts[0] != before {
		t.Fatal("invalid install changed model")
	}
}
