package virtual

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
)

type editableGridFixture struct {
	*indexedGridSource
	values   map[GridActive]GridEditValue
	editHook func(string, string) (GridEditValue, error)
}

func newEditableFixture(rows int) *editableGridFixture {
	return &editableGridFixture{indexedGridSource: gridSource(rows), values: map[GridActive]GridEditValue{}}
}
func (s *editableGridFixture) EditCell(key, col string) (GridEditValue, error) {
	if s.editHook != nil {
		return s.editHook(key, col)
	}
	if v, ok := s.values[GridActive{key, col}]; ok {
		return v, nil
	}
	return GridEditValue{Text: key + "/" + col, Version: 1, Editable: true}, nil
}
func (s *editableGridFixture) Cell(key, col string) (string, error) {
	s.calls.Cell++
	v, e := s.EditCell(key, col)
	return v.Text, e
}
func (s *editableGridFixture) candidate(key, col, text string, version uint64) *editableGridFixture {
	x := *s.indexedGridSource
	x.revision++
	x.calls = gridCalls{}
	n := &editableGridFixture{indexedGridSource: &x, values: map[GridActive]GridEditValue{}}
	for k, v := range s.values {
		n.values[k] = v
	}
	n.values[GridActive{key, col}] = GridEditValue{text, version, true}
	return n
}
func editableFixtureGrid(t *testing.T, rows int) (*Grid, *editableGridFixture) {
	t.Helper()
	s := newEditableFixture(rows)
	g := mustGrid(t, s, gridColumns(4, 100), GridConfig{RowHeight: 24, HeaderHeight: 28, FrozenColumns: 1, CellEditing: true}, GridViewport{Width: 300, Height: 124})
	return g, s
}
func beginFixtureEdit(t *testing.T, g *Grid) GridEditState {
	t.Helper()
	snap := gridSnapshot(t, g)
	if err := g.BeginEdit(snap.Cells[0].Ref); err != nil {
		t.Fatal(err)
	}
	e, err := g.Edit()
	if err != nil {
		t.Fatal(err)
	}
	return e
}
func TestGridEditAtomicCommitAndDetachedRequests(t *testing.T) {
	g, s := editableFixtureGrid(t, 1_000_000)
	e := beginFixtureEdit(t, g)
	if e.Status != GridEditReady || e.BaseVersion != 1 || e.Text != e.Original {
		t.Fatal(e)
	}
	if s.calls.IndexOf > 8 || s.calls.RowKey > 15 {
		t.Fatal("source-sized edit work", s.calls)
	}
	if err := g.SetEditText("café 👩‍👧"); err != nil {
		t.Fatal(err)
	}
	r, err := g.PrepareEdit()
	if err != nil {
		t.Fatal(err)
	}
	if r.State().Text != "café 👩‍👧" || r.Stamp().Revision != 1 {
		t.Fatal(r.State(), r.Stamp())
	}
	before := gridState(t, g)
	bad := s.candidate(e.RowKey, e.ColumnID, "wrong", 2)
	if err := g.CompleteEdit(r, bad); !errors.Is(err, ErrGridEditConflict) {
		t.Fatal(err)
	}
	if gridState(t, g) != before || g.source != s {
		t.Fatal("failed candidate published")
	}
	next := s.candidate(e.RowKey, e.ColumnID, r.State().Text, 2)
	if err := g.CompleteEdit(r, next); err != nil {
		t.Fatal(err)
	}
	if gridState(t, g).Edit.Active || g.source != next || g.generation != before.Generation+1 {
		t.Fatal("commit did not publish once", gridState(t, g))
	}
	if err := g.CompleteEdit(r, next); !errors.Is(err, ErrGridStale) {
		t.Fatal("request replay", err)
	}
	if r.State().Text != "café 👩‍👧" || r.State().BaseVersion != 1 {
		t.Fatal("request mutated", r)
	}
}
func TestGridEditProjectionAndConflictRecovery(t *testing.T) {
	g, s := editableFixtureGrid(t, 100)
	e := beginFixtureEdit(t, g)
	_ = g.SetEditText("local draft")
	n := s.candidate("r:99", "c:2", "unrelated", 2)
	n.reverse = true
	if err := g.ReplaceSource(n); err != nil {
		t.Fatal(err)
	}
	if state, _ := g.Edit(); state.Status != GridEditReady || state.Text != "local draft" || state.RowKey != e.RowKey {
		t.Fatal(state)
	}
	r, err := g.PrepareEdit()
	if err != nil {
		t.Fatal(err)
	}
	remote := n.candidate(e.RowKey, e.ColumnID, "remote", 2)
	if err := g.ReplaceSource(remote); err != nil {
		t.Fatal(err)
	}
	if state, _ := g.Edit(); state.Status != GridEditConflict || state.Text != "local draft" || state.CurrentText != "remote" {
		t.Fatal(state)
	}
	if _, err := g.PrepareEdit(); !errors.Is(err, ErrGridEditConflict) {
		t.Fatal(err)
	}
	if err := g.CompleteEdit(r, remote); !errors.Is(err, ErrGridStale) {
		t.Fatal(err)
	}
	aba := remote.candidate(e.RowKey, e.ColumnID, e.Original, 3)
	if err := g.ReplaceSource(aba); err != nil {
		t.Fatal(err)
	}
	if state, _ := g.Edit(); state.Status != GridEditConflict {
		t.Fatal("cell ABA ignored", state)
	}
	other := aba.candidate(e.RowKey, e.ColumnID, e.Original, 4)
	other.id = "new-universe"
	if err := g.ReplaceSource(other); err != nil {
		t.Fatal(err)
	}
	if state, _ := g.Edit(); state.Status != GridEditIncarnation || state.Text != "local draft" {
		t.Fatal(state)
	}
	if err := g.CancelEdit(); err != nil {
		t.Fatal(err)
	}
}
func TestGridEditHiddenAndReadOnlyTarget(t *testing.T) {
	g, s := editableFixtureGrid(t, 100)
	e := beginFixtureEdit(t, g)
	hidden := s.candidate("r:99", "c:3", "unrelated", 2)
	hidden.rows = 0
	if err := g.ReplaceSource(hidden); err != nil {
		t.Fatal(err)
	}
	if state, _ := g.Edit(); state.Status != GridEditUnavailable || state.CurrentVersion != 0 {
		t.Fatal(state)
	}
	if _, err := g.PrepareEdit(); !errors.Is(err, ErrGridEditConflict) {
		t.Fatal(err)
	}
	back := s.candidate("r:99", "c:3", "unrelated", 2)
	back.revision = hidden.revision + 1
	if err := g.ReplaceSource(back); err != nil {
		t.Fatal(err)
	}
	if state, _ := g.Edit(); state.Status != GridEditReady {
		t.Fatal(state)
	}
	locked := back.candidate(e.RowKey, e.ColumnID, e.Original, 2)
	v := locked.values[GridActive{e.RowKey, e.ColumnID}]
	v.Editable = false
	locked.values[GridActive{e.RowKey, e.ColumnID}] = v
	if err := g.ReplaceSource(locked); err != nil {
		t.Fatal(err)
	}
	if _, err := g.PrepareEdit(); !errors.Is(err, ErrGridEditReadOnly) {
		t.Fatal(err)
	}
}
func TestGridEditStaleRequestDraftABAAndRestart(t *testing.T) {
	g, s := editableFixtureGrid(t, 100)
	e := beginFixtureEdit(t, g)
	r, _ := g.PrepareEdit()
	_ = g.SetEditText("B")
	_ = g.SetEditText(e.Text)
	if err := g.CompleteEdit(r, s.candidate(e.RowKey, e.ColumnID, e.Text, 2)); !errors.Is(err, ErrGridStale) {
		t.Fatal(err)
	}
	r, _ = g.PrepareEdit()
	_ = g.CancelEdit()
	_ = beginFixtureEdit(t, g)
	if err := g.CompleteEdit(r, s.candidate(e.RowKey, e.ColumnID, e.Text, 2)); !errors.Is(err, ErrGridStale) {
		t.Fatal(err)
	}
	var foreign GridEditRequest
	if err := g.CompleteEdit(foreign, s); !errors.Is(err, ErrGridStale) {
		t.Fatal(err)
	}
}
func TestGridEditBoundsNoopsExhaustionAndStaleRecovery(t *testing.T) {
	g, s := editableFixtureGrid(t, 100)
	e := beginFixtureEdit(t, g)
	before := gridState(t, g)
	for _, text := range []string{"a\nb", "a\rb", "a\x00b", string([]byte{0xff}), strings.Repeat("x", g.config.MaxEditBytes+1)} {
		if err := g.SetEditText(text); err == nil {
			t.Fatal("invalid draft", len(text))
		}
		if gridState(t, g) != before {
			t.Fatal("refusal changed state")
		}
	}
	if err := g.SetEditText(e.Text); err != nil || gridState(t, g) != before {
		t.Fatal("noop changed state", err)
	}
	g.edit.state.Revision = math.MaxUint64
	if err := g.SetEditText("new"); !errors.Is(err, ErrGridExhausted) {
		t.Fatal(err)
	}
	g.edit.state.Revision = before.Edit.Revision
	g.generation = math.MaxUint64
	if err := g.CancelEdit(); !errors.Is(err, ErrGridExhausted) {
		t.Fatal(err)
	}
	g.generation = before.Generation
	s.revision++ // unacknowledged source drift
	if _, err := g.State(); !errors.Is(err, ErrGridStale) {
		t.Fatal(err)
	}
	if state, err := g.Edit(); err != nil || state.Text != e.Text {
		t.Fatal("draft recovery refused", state, err)
	}
	if err := g.SetEditText("recoverable"); err != nil {
		t.Fatal(err)
	}
	if err := g.CancelEdit(); err != nil {
		t.Fatal(err)
	}
	copy := *g
	if _, err := copy.Edit(); !errors.Is(err, ErrGridInvalid) {
		t.Fatal("copiedgrid", err)
	}
	for _, typ := range []reflect.Type{reflect.TypeOf(GridEditRequest{}), reflect.TypeOf(gridRefToken{})} {
		for i := 0; i < typ.NumField(); i++ {
			if typ.Field(i).Type == reflect.TypeOf(g) || typ.Field(i).Type == reflect.TypeOf(s) {
				t.Fatal("request retains model/source")
			}
		}
	}
}
func TestGridEditCandidateAdmissionFailureKeepsDraft(t *testing.T) {
	g, s := editableFixtureGrid(t, 100)
	e := beginFixtureEdit(t, g)
	_ = g.SetEditText("valid draft")
	r, _ := g.PrepareEdit()
	before := gridState(t, g)
	next := s.candidate(e.RowKey, e.ColumnID, r.State().Text, 2)
	next.rowHook = func(int) (string, error) { return "", errors.New("malformed rowindex") }
	if err := g.CompleteEdit(r, next); err == nil {
		t.Fatal("malformed candidate accepted")
	}
	if gridState(t, g) != before || g.source != s {
		t.Fatal("partial commit")
	}
}

func TestGridEditSessionIdentityAndSourcePermissionReuse(t *testing.T) {
	g, s := editableFixtureGrid(t, 100)
	first := beginFixtureEdit(t, g)
	if first.Session != 1 {
		t.Fatal(first)
	}
	_ = g.CancelEdit()
	second := beginFixtureEdit(t, g)
	if second.Session <= first.Session {
		t.Fatal("sessionABA", first, second)
	}
	locked := s.candidate(second.RowKey, second.ColumnID, second.Text, 1)
	v := locked.values[GridActive{second.RowKey, second.ColumnID}]
	v.Editable = false
	locked.values[GridActive{second.RowKey, second.ColumnID}] = v
	before := gridState(t, g)
	if err := g.ReplaceSource(locked); !errors.Is(err, ErrGridSource) {
		t.Fatal("permissionversionreuse", err)
	}
	if gridState(t, g) != before {
		t.Fatal("invalid source admitted")
	}
	_ = g.CancelEdit()
	g.editSerial = math.MaxUint64
	before = gridState(t, g)
	if err := g.BeginEdit(gridSnapshot(t, g).Cells[0].Ref); !errors.Is(err, ErrGridExhausted) {
		t.Fatal(err)
	}
	if gridState(t, g) != before {
		t.Fatal("exhaustion mutated state")
	}
}
