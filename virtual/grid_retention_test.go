package virtual

import (
	"errors"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"unsafe"
)

// Inspect the retained graph instead of relying on GC timing/finalizers. An
// interface, closure, unsafe pointer or channel could hide an arbitrary owner;
// a Grid backpointer reaches its borrowed source interface and fails this check.
func gridRefCanRetainDomain(kind reflect.Type, seen map[reflect.Type]bool) bool {
	if seen[kind] {
		return false
	}
	seen[kind] = true
	switch kind.Kind() {
	case reflect.Interface, reflect.Func, reflect.Chan, reflect.UnsafePointer:
		return true
	case reflect.Pointer, reflect.Slice, reflect.Array:
		return gridRefCanRetainDomain(kind.Elem(), seen)
	case reflect.Map:
		return gridRefCanRetainDomain(kind.Key(), seen) || gridRefCanRetainDomain(kind.Elem(), seen)
	case reflect.Struct:
		for i := 0; i < kind.NumField(); i++ {
			if gridRefCanRetainDomain(kind.Field(i).Type, seen) {
				return true
			}
		}
	}
	return false
}

func TestGridReferenceDetachesDomainAndSeparatesIdenticalModels(t *testing.T) {
	if gridRefCanRetainDomain(reflect.TypeOf(GridCellRef{}), map[reflect.Type]bool{}) {
		t.Fatal("detached reference can retain an arbitrary borrowed domain/owner")
	}
	if unsafe.Sizeof(gridRefToken{}) == 0 {
		t.Fatal("identity token permits zero-sized allocation aliasing")
	}
	s := gridSource(1)
	makeGrid := func() *Grid {
		return mustGrid(t, s, gridColumns(1, 100), GridConfig{RowHeight: 20}, GridViewport{Width: 100, Height: 20})
	}
	one, two := makeGrid(), makeGrid()
	ref := gridSnapshot(t, one).Cells[0].Ref
	if err := two.Activate(ref); !errors.Is(err, ErrGridStale) {
		t.Fatal("identical foreign model accepted ref", err)
	}
	if err := one.Activate(ref); err != nil {
		t.Fatal("detached token rejected original model", err)
	}
}

type borrowedGridSource struct {
	identity string
	revision uint64
	keys     []string
	text     string
}

func (s *borrowedGridSource) Identity() string             { return s.identity }
func (s *borrowedGridSource) Revision() uint64             { return s.revision }
func (s *borrowedGridSource) RowCount() int                { return len(s.keys) }
func (s *borrowedGridSource) RowKey(i int) (string, error) { return s.keys[i], nil }
func (s *borrowedGridSource) IndexOf(key string) (int, bool, error) {
	// Three entries are explicit fixture indices, not the public source design.
	for i, k := range s.keys {
		if key == k {
			return i, true, nil
		}
	}
	return 0, false, nil
}
func (s *borrowedGridSource) Cell(string, string) (string, error) { return s.text, nil }

func assertGridDetached(t *testing.T, label, retained, backing string) {
	t.Helper()
	start := uintptr(unsafe.Pointer(unsafe.StringData(backing)))
	address := uintptr(unsafe.Pointer(unsafe.StringData(retained)))
	if address >= start && address-start < uintptr(len(backing)) {
		t.Fatalf("%s retains %d-byte caller backing through %d-byte string", label, len(backing), len(retained))
	}
	runtime.KeepAlive(backing)
}

func TestGridAcceptedStringsDetachLargeCallerBacking(t *testing.T) {
	backing := strings.Repeat("x", 1<<20)
	source := &borrowedGridSource{backing[10:19], 1, []string{backing[30:34], backing[50:56], backing[70:78]}, backing[100:120]}
	columns := []GridColumn{{backing[200:205], backing[220:231], 100}, {backing[240:247], backing[260:273], 100}}
	g := mustGrid(t, source, columns, GridConfig{RowHeight: 20, HeaderHeight: 20, MaxTextBytes: 128}, GridViewport{Width: 100, Height: 40})
	assertGridDetached(t, "source identity", g.stamp.Identity, backing)
	assertGridDetached(t, "initial anchor", g.anchorKey, backing)
	for _, col := range g.columns {
		assertGridDetached(t, "column ID", col.ID, backing)
		assertGridDetached(t, "column title", col.Title, backing)
	}
	callerBacking := strings.Repeat("x", 2<<20)
	if err := g.JumpToKey(callerBacking[300:306], columns[1].ID); err != nil {
		t.Fatal(err)
	}
	assertGridDetached(t, "active caller key", g.active.RowKey, callerBacking)
	assertGridDetached(t, "active source key", g.active.RowKey, backing)
	assertGridDetached(t, "navigation anchor", g.anchorKey, backing)
	if g.active.RowKey != source.keys[1] || g.active.ColumnID != columns[1].ID {
		t.Fatal("cloning changed active identity")
	}
	if err := g.SetViewport(GridViewport{Width: 200, Height: 40, ScrollY: 40}); err != nil {
		t.Fatal(err)
	}
	assertGridDetached(t, "viewport anchor", g.anchorKey, backing)
	snapshot := gridSnapshot(t, g)
	assertGridDetached(t, "snapshot stamp", snapshot.State.Stamp.Identity, backing)
	assertGridDetached(t, "snapshot active", snapshot.State.Active.RowKey, callerBacking)
	if len(snapshot.Cells) != 2 {
		t.Fatal("expected one row and two columns")
	}
	for _, header := range snapshot.Headers {
		assertGridDetached(t, "snapshot header ID", header.ColumnID, backing)
		assertGridDetached(t, "snapshot header title", header.Title, backing)
	}
	for _, cell := range snapshot.Cells {
		assertGridDetached(t, "snapshot row key", cell.RowKey, backing)
		assertGridDetached(t, "snapshot column ID", cell.ColumnID, backing)
		assertGridDetached(t, "snapshot text", cell.Text, backing)
		assertGridDetached(t, "reference row key", cell.Ref.RowKey(), backing)
		assertGridDetached(t, "reference column ID", cell.Ref.ColumnID(), backing)
		if cell.Text != source.text || cell.RowKey != source.keys[2] {
			t.Fatal("cloning changed snapshot bytes")
		}
		if unsafe.StringData(cell.Ref.RowKey()) != unsafe.StringData(cell.RowKey) {
			t.Fatal("reference did not share detached row key")
		}
	}
	if unsafe.StringData(snapshot.Cells[0].RowKey) != unsafe.StringData(snapshot.Cells[1].RowKey) {
		t.Fatal("row key cloned per column")
	}
	source.keys[0], source.keys[2] = source.keys[2], source.keys[0]
	source.revision++
	if err := g.Refresh(); err != nil {
		t.Fatal(err)
	}
	assertGridDetached(t, "refreshed identity", g.stamp.Identity, backing)
	assertGridDetached(t, "refreshed anchor", g.anchorKey, backing)
	assertGridDetached(t, "refreshed active", g.active.RowKey, callerBacking)
	replacementBacking := strings.Repeat("y", 3<<20)
	replacement := &borrowedGridSource{replacementBacking[10:19], 1, []string{replacementBacking[30:34]}, replacementBacking[100:120]}
	if err := g.ReplaceSource(replacement); err != nil {
		t.Fatal(err)
	}
	assertGridDetached(t, "replacement identity", g.stamp.Identity, replacementBacking)
	assertGridDetached(t, "replacement anchor", g.anchorKey, replacementBacking)
	if g.active != (GridActive{}) {
		t.Fatal("new incarnation retained active key")
	}
}

func TestGridEmptyStringsDoNotRetainBacking(t *testing.T) {
	backing := strings.Repeat("x", 1<<20)
	s := &borrowedGridSource{backing[10:19], 1, []string{backing[30:34]}, backing[100:100]}
	g := mustGrid(t, s, []GridColumn{{backing[200:205], backing[220:220], 100}}, GridConfig{RowHeight: 20}, GridViewport{Width: 100, Height: 20})
	assertGridDetached(t, "empty column title", g.columns[0].Title, backing)
	snapshot := gridSnapshot(t, g)
	assertGridDetached(t, "empty header title", snapshot.Headers[0].Title, backing)
	assertGridDetached(t, "empty cell text", snapshot.Cells[0].Text, backing)
	if snapshot.Cells[0].Text != "" || snapshot.Headers[0].Title != "" {
		t.Fatal("empty payload changed")
	}
}
