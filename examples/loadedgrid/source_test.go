package main

import (
	"context"
	"errors"
	"github.com/dshills/vellum/virtual"
	"runtime"
	"sync"
	"testing"
)

func TestStoredRecordsImmutableDenseIndexAndCapabilities(t *testing.T) {
	d, e := generate(context.Background(), 10_000, nil)
	if e != nil {
		t.Fatal(e)
	}
	s, e := source("test", d, len(d.Records))
	if e != nil {
		t.Fatal(e)
	}
	if len(d.Records) != 10_000 || s.logicalBytes() != 1_600_000 {
		t.Fatal("not physically stored20int64 rows")
	}
	for _, r := range []int{0, 7, 9999} {
		k, e := s.RowKey(r)
		if e != nil {
			t.Fatal(e)
		}
		index, found, e := s.IndexOf(k)
		if e != nil || !found || index != r {
			t.Fatal("dense inverse")
		}
		for c := range columns {
			v, _, e := s.value(k, column(c))
			if e != nil || v != d.Records[r].Values[c] {
				t.Fatal("cell not backed by stored records")
			}
		}
	}
	before := s.counts.snapshot()
	idx, e := s.SelectionSnapshot()
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 100 {
				k, _ := idx.RowKey(i)
				n, ok, e := idx.IndexOf(k)
				if e != nil || !ok || n != i {
					t.Error("immutable index")
				}
			}
		}()
	}
	wg.Wait()
	if before != s.counts.snapshot() {
		t.Fatal("selection mutateddiagnostic source")
	}
	old, version, _ := s.value(key(7), column(1))
	next, e := edited(context.Background(), s, key(7), column(1), "-12345", version)
	if e != nil {
		t.Fatal(e)
	}
	value, v, _ := next.value(key(7), column(1))
	unchanged, uv, _ := s.value(key(7), column(1))
	if value != -12345 || v != 2 || unchanged != old || uv != 1 || next.Revision() != s.Revision()+1 {
		t.Fatal("edit mutated immutable revision")
	}
	if _, e = edited(context.Background(), next, key(7), column(1), "42", 1); !errors.Is(e, errConflict) {
		t.Fatal(e)
	}
	for _, text := range []string{"", "+1", "01", " 2", "1\n", "9223372036854775808", "word"} {
		if _, e = numeric(text); !errors.Is(e, errValue) {
			t.Fatal("invalid accepted", text, e)
		}
	}
	variable := &variableSource{s}
	for boundary := 0; boundary <= s.rows; boundary++ {
		pixel, _ := variable.RowOffset(boundary)
		if boundary < s.rows {
			row, e := variable.RowAt(pixel)
			if e != nil || row != boundary {
				t.Fatal("variable inverse")
			}
		}
	}
	if _, e = variable.RowAt(offset(s.rows)); !errors.Is(e, virtual.ErrGridRange) {
		t.Fatal(e)
	}
}
func TestStoredMillionRowsLogicalAndHeapMemory(t *testing.T) {
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	d, e := generate(context.Background(), maxRows, nil)
	if e != nil {
		t.Fatal(e)
	}
	s, e := source("million", d, maxRows)
	if e != nil {
		t.Fatal(e)
	}
	runtime.GC()
	runtime.ReadMemStats(&after)
	if len(d.Records) != maxRows || cap(d.Records) != maxRows || s.logicalBytes() != 160_000_000 {
		t.Fatal("million record payload missing")
	}
	heapDelta := int64(after.HeapAlloc) - int64(before.HeapAlloc)
	t.Logf("domain logicalbytes=%d; observed Go HeapAlloc delta=%d (allocator/runtime effects, not RSS/peak/native/GPU)", s.logicalBytes(), heapDelta)
	if heapDelta < 150_000_000 {
		t.Fatal("stored array not retained in observed heap", heapDelta)
	}
	view, e := virtual.BuildGridProjection(context.Background(), s, 3, virtual.GridQuery{Sort: []virtual.GridSortColumn{{ColumnID: column(1), Kind: virtual.GridSortInteger}}}, limits())
	if e != nil {
		t.Fatal(e)
	}
	if view.RowCount() != maxRows || view.InputStamp().Rows != maxRows || view.Stats().IndexBytes > 64<<20 || view.Stats().WorkingBytes > 256<<20 {
		t.Fatal("projection budget", view.Stats())
	}
	t.Logf("builder stats: %+v; borrowed domain bytes excluded", view.Stats())
	runtime.KeepAlive(d)
	runtime.KeepAlive(s)
	runtime.KeepAlive(view)
}
func TestGenerationAndCopyCancellationBounds(t *testing.T) {
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := generate(canceled, 10_000, nil); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	for _, n := range []int{0, -1, maxRows + 1} {
		if _, e := generate(context.Background(), n, nil); !errors.Is(e, errData) {
			t.Fatal(n, e)
		}
	}
	d, _ := generate(context.Background(), 64, nil)
	s, _ := source("test", d, 64)
	if _, e := edited(canceled, s, key(0), column(1), "1", 1); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if _, e := s.Cell("row:0", column(1)); !errors.Is(e, virtual.ErrGridRange) {
		t.Fatal(e)
	}
	if _, e := s.Cell(key(0), "column:20"); !errors.Is(e, virtual.ErrGridRange) {
		t.Fatal(e)
	}
}
