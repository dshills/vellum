package main

import (
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/virtual"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSnapshotAndTypedMatcher(t *testing.T) {
	s := newParts(100000)
	snapshot := s.Snapshot()
	changed := s.parts.at(0)
	changed.Value += 100
	s.parts = s.parts.with(0, changed)
	if snapshot.parts.at(0).Value == s.parts.at(0).Value {
		t.Fatal("mutable worker snapshot")
	}
	for _, q := range []string{"r0000001", "0603", "ω", "90°", "not-present"} {
		for _, i := range []int{0, 1, 50, 99999} {
			expected := strings.Contains(strings.ToLower(strings.Join(s.Row(i).Cells, " ")), q)
			if s.Matches(i, q) != expected {
				t.Fatal(q, i)
			}
		}
	}
}
func BenchmarkActualMillionViewport(b *testing.B) {
	source := newParts(1000000)
	table, err := virtual.NewTable(source, 20)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		table.SetViewport(i%999980, 20)
		table.View()
	}
}
func BenchmarkTypedMatcher(b *testing.B) {
	source := newParts(1000000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.Matches(i%1000000, "0603")
	}
}

func TestDocumentUndoCapturesDomainIdentityAndInvalidatesWorkers(t *testing.T) {
	source := newParts(100000)
	table, err := virtual.NewTable(source, 20)
	if err != nil {
		t.Fatal(err)
	}
	board, err := canvas.New([]canvas.Rect{{ID: source.Row(0).ID, X: 2, Y: 4, Width: 50, Height: 25}})
	if err != nil {
		t.Fatal(err)
	}
	journal, err := history.New(256)
	if err != nil {
		t.Fatal(err)
	}
	editor := &partEditor{source: source, table: table, board: board, journal: journal}
	old := source.parts.at(0)
	oldRect := board.Objects[0]
	token := table.ProjectionToken()
	projection, err := virtual.Project(source.Snapshot(), "", 1, false)
	if err != nil {
		t.Fatal(err)
	}
	next := old
	next.Rotation = (next.Rotation + 90) % 360
	if err = editor.change(0, "Rotate", next); err != nil {
		t.Fatal(err)
	}
	if source.parts.at(0) != next || board.Objects[0].Width != 25 || source.Revision() != 2 {
		t.Fatal("edit not reflected")
	}
	if table.ApplyProjection(projection, token) == nil {
		t.Fatal("stale pre-edit worker accepted")
	}
	nextOther := source.parts.at(5)
	nextOther.Value += 100
	if err = editor.change(5, "Value", nextOther); err != nil {
		t.Fatal(err)
	}
	if err = journal.Undo(); err != nil {
		t.Fatal(err)
	}
	if source.parts.at(0) != next || source.parts.at(5) == nextOther {
		t.Fatal("undo used current selection instead of captured component")
	}
	if err = journal.Undo(); err != nil {
		t.Fatal(err)
	}
	if source.parts.at(0) != old || board.Objects[0] != oldRect || journal.Snapshot().Dirty || source.Revision() != 5 {
		t.Fatal("undo did not restore model, geometry and clean state", source.Revision())
	}
	if table.ApplyProjection(projection, token) == nil {
		t.Fatal("undo incorrectly revived old worker token")
	}
	if err = journal.Redo(); err != nil {
		t.Fatal(err)
	}
	if source.parts.at(0) != next || board.Objects[0].Width != 25 || source.Revision() != 6 {
		t.Fatal("redo not reflected")
	}
	before := journal.Snapshot()
	if err = editor.change(0, "No-op", next); err != nil || journal.Snapshot() != before {
		t.Fatal("no-op recorded")
	}
	if editor.change(-1, "Invalid", next) == nil || journal.Snapshot() != before {
		t.Fatal("invalid edit recorded")
	}
}

func TestFileReferencesUndoRedoPreserveCapturedBatches(t *testing.T) {
	journal, err := history.New(8)
	if err != nil {
		t.Fatal(err)
	}
	refs := &fileReferences{}
	first := []string{filepath.Join(t.TempDir(), "first Ω\nfile")}
	original := first[0]
	if err = refs.replace(journal, first); err != nil {
		t.Fatal(err)
	}
	first[0] = "caller mutation"
	if refs.paths[0] != original || !journal.Snapshot().Dirty {
		t.Fatal("accepted batch aliased or clean")
	}
	second := []string{filepath.Join(t.TempDir(), "second"), original}
	if err = refs.replace(journal, second); err != nil {
		t.Fatal(err)
	}
	refs.paths[0] = "retained mutation"
	if err = journal.Undo(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(refs.paths, []string{original}) {
		t.Fatal("undo did not restore first captured batch")
	}
	if err = journal.Redo(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(refs.paths, second) {
		t.Fatal("redo did not restore second captured batch")
	}
	before := journal.Snapshot()
	if err = refs.replace(journal, second); err != nil || journal.Snapshot() != before {
		t.Fatal("no-op changed history")
	}
	if err = refs.replace(journal, []string{"relative"}); err == nil || journal.Snapshot() != before {
		t.Fatal("invalid batch changed history")
	}
	if err = journal.Undo(); err != nil {
		t.Fatal(err)
	}
	if err = journal.Undo(); err != nil || len(refs.paths) != 0 || journal.Snapshot().Dirty {
		t.Fatal("undo to baseline failed")
	}
}

func TestFileReferenceSummaryBoundsChromeWithoutChangingPaths(t *testing.T) {
	refs := &fileReferences{paths: []string{filepath.Join(t.TempDir(), strings.Repeat("Ω", 1000)), filepath.Join(t.TempDir(), "two"), filepath.Join(t.TempDir(), "three"), filepath.Join(t.TempDir(), "four")}}
	original := append([]string(nil), refs.paths...)
	summary := refs.summary()
	if len(summary) > 800 || !strings.Contains(summary, "Referenced 4 files") || !strings.Contains(summary, "…") || strings.Contains(summary, "four") || !reflect.DeepEqual(refs.paths, original) {
		t.Fatal("summary changed the paths or exceeded its display bound", summary)
	}
}
