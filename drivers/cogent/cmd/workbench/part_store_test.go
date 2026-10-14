package main

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"sync"
	"testing"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/virtual"
)

func TestPartSnapshotsAcrossPagesAndConcurrentEdits(t *testing.T) {
	source := newParts(1000000)
	frozen := source.Snapshot()
	if frozen.parts.root != source.parts.root {
		t.Fatal("snapshot enumerated/copied store")
	}
	table, _ := virtual.NewTable(source, 20)
	journal, _ := history.New(256)
	edit := &partEditor{source: source, table: table, journal: journal}
	var readers sync.WaitGroup
	errors := make(chan int, 4)
	for n := 0; n < 4; n++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for j := 0; j < 10000; j++ {
				i := j * 997 % frozen.Len()
				p := frozen.parts.at(i)
				if p != (part{uint32(100 + i%997), uint16(i % 4 * 90), uint16(i % 2)}) {
					errors <- i
					return
				}
			}
		}()
	}
	for _, index := range []int{0, 255, 256, 16383, 16384, 999999} {
		oldRoot := source.parts.root
		next := source.parts.at(index)
		next.Value++
		if err := edit.change(index, "value", next); err != nil {
			t.Fatal(err)
		}
		a, b := index>>14, (index>>8)&63
		if oldRoot == source.parts.root || oldRoot[a] == source.parts.root[a] || oldRoot[a][b] == source.parts.root[a][b] {
			t.Fatal("edit mutated shared path")
		}
		for j := range oldRoot {
			if j != a && oldRoot[j] != source.parts.root[j] {
				t.Fatal("edit copied unrelated branch")
			}
		}
		for j := range oldRoot[a] {
			if j != b && oldRoot[a][j] != source.parts.root[a][j] {
				t.Fatal("edit copied unrelated leaf")
			}
		}
	}
	readers.Wait()
	close(errors)
	for i := range errors {
		t.Fatalf("frozen worker changed at%d", i)
	}
	if frozen.Revision() != 1 || source.Revision() != 7 {
		t.Fatal("snapshot revision changed")
	}
	for n := 0; n < 6; n++ {
		if err := journal.Undo(); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(source.parts.values(), frozen.parts.values()) {
		t.Fatal("undo changed untouched data")
	}
	for n := 0; n < 6; n++ {
		if err := journal.Redo(); err != nil {
			t.Fatal(err)
		}
	}
	if frozen.parts.at(999999).Value == source.parts.at(999999).Value {
		t.Fatal("redo mutated old worker")
	}
}

func TestPartStoreRandomUpdatesAgainstFlatOracle(t *testing.T) {
	source := newParts(100000)
	oracle := source.parts.values()
	original := source.Snapshot()
	random := rand.New(rand.NewSource(942))
	for n := 0; n < 1000; n++ {
		i := random.Intn(len(oracle))
		p := oracle[i]
		p.Value = uint32(random.Uint64())
		p.Rotation = uint16(random.Intn(4) * 90)
		p.Package = uint16(random.Intn(2))
		source.parts = source.parts.with(i, p)
		oracle[i] = p
	}
	if !reflect.DeepEqual(source.parts.values(), oracle) {
		t.Fatal("path-copy result diverged")
	}
	for i := 0; i < original.Len(); i++ {
		if original.parts.at(i).Value != uint32(100+i%997) {
			t.Fatal("historical root mutated")
		}
	}
	// Bulk construction must not retain a decoded caller slice.
	store := storeParts(oracle)
	before := store.at(0)
	oracle[0].Value++
	if store.at(0) != before {
		t.Fatal("import retained mutable caller values")
	}
}

func TestSharedDocumentCapturePreservesV1WireAndLoadedSnapshot(t *testing.T) {
	source := newParts(100000)
	board, _ := canvas.New([]canvas.Rect{{ID: "component:1", Width: 25, Height: 50}})
	refs := &fileReferences{paths: []string{"/tmp/ref"}}
	frozen := captureDocument(source, board, refs)
	if frozen.Parts != nil || frozen.captured.parts.root != source.parts.root {
		t.Fatal("owner capture enumerated parts")
	}
	expected := struct {
		Version    int
		Parts      []part
		Objects    []canvas.Rect
		References []string
	}{1, source.parts.values(), append([]canvas.Rect(nil), board.Objects...), append([]string(nil), refs.paths...)}
	want, err := json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}
	changed := source.parts.at(0)
	changed.Value++
	source.parts = source.parts.with(0, changed)
	got, err := encodeDocument(frozen)
	if err != nil || string(got) != string(want) {
		t.Fatal("captured v1 wire changed", err)
	}
	loaded, err := decodeDocument(got, 100)
	if err != nil {
		t.Fatal(err)
	}
	beforeLoad := source.Snapshot()
	table, _ := virtual.NewTable(source, 20)
	if err = installDocument(source, table, board, refs, loaded); err != nil {
		t.Fatal(err)
	}
	if source.parts.root != loaded.prepared.root || source.parts.at(0) != expected.Parts[0] || beforeLoad.parts.at(0) != changed {
		t.Fatal("load copied owner data or mutated prior snapshot")
	}
	loaded.Parts[0].Value++
	if source.parts.at(0) != expected.Parts[0] {
		t.Fatal("decoded mutable wire slice retained")
	}
}

type flatPartSnapshot struct {
	parts    []part
	revision uint64
}

var snapshotBenchmarkSink *partSource
var flatSnapshotBenchmarkSink *flatPartSnapshot
var storeBenchmarkSink partStore

func BenchmarkPartSnapshot(b *testing.B) {
	for _, count := range []int{100000, 1000000} {
		name := "100k"
		if count == 1000000 {
			name = "1m"
		}
		source := newParts(count)
		flat := source.parts.values()
		b.Run(name+"/original-flat-copy", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				flatSnapshotBenchmarkSink = &flatPartSnapshot{append([]part(nil), flat...), source.revision}
			}
		})
		b.Run(name+"/shared", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				snapshotBenchmarkSink = source.Snapshot()
			}
		})
		b.Run(name+"/edit-path", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				index := i % source.Len()
				next := source.parts.at(index)
				next.Value++
				source.parts = source.parts.with(index, next)
			}
			storeBenchmarkSink = source.parts
		})
	}
}

var traversalBenchmarkSink uint64

func BenchmarkPartTraversal(b *testing.B) {
	source := newParts(1000000)
	flat := source.parts.values()
	b.Run("original-flat", func(b *testing.B) {
		b.ReportAllocs()
		for n := 0; n < b.N; n++ {
			var sum uint64
			for _, p := range flat {
				sum += uint64(p.Value)
			}
			traversalBenchmarkSink = sum
		}
	})
	b.Run("shared-pages", func(b *testing.B) {
		b.ReportAllocs()
		for n := 0; n < b.N; n++ {
			var sum uint64
			for i := 0; i < source.Len(); i++ {
				sum += uint64(source.parts.at(i).Value)
			}
			traversalBenchmarkSink = sum
		}
	})
}
