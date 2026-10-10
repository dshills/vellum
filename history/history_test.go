package history

import (
	"errors"
	"math"
	"testing"
)

func journal(t *testing.T, limit int) *Journal {
	t.Helper()
	j, err := New(limit)
	if err != nil {
		t.Fatal(err)
	}
	return j
}
func set(value *int, next int) Change {
	before := *value
	return Change{Label: "Set value", Apply: func() error { *value = next; return nil }, Revert: func() error { *value = before; return nil }}
}
func check(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSavedIdentitySurvivesUndoRedoAndBranching(t *testing.T) {
	j := journal(t, 4)
	v := 0
	check(t, j.Apply(set(&v, 1)))
	check(t, j.MarkSaved())
	saved := j.Snapshot().Revision
	check(t, j.Apply(set(&v, 2)))
	check(t, j.Undo())
	if v != 1 || j.Snapshot().Dirty || j.Snapshot().RedoCount != 1 {
		t.Fatal(v, j.Snapshot())
	}
	check(t, j.Undo())
	if v != 0 || !j.Snapshot().Dirty {
		t.Fatal(v, j.Snapshot())
	}
	check(t, j.Apply(set(&v, 7)))
	if v != 7 || !j.Snapshot().Dirty || j.Snapshot().Revision == saved || j.Snapshot().RedoCount != 0 {
		t.Fatal(v, j.Snapshot())
	}
	if !errors.Is(j.Redo(), ErrEmpty) {
		t.Fatal("abandoned branch replayed")
	}
	if !errors.Is(j.MarkSavedRevision(saved), ErrStale) || !j.Snapshot().Dirty {
		t.Fatal("stale save acknowledged")
	}
	check(t, j.MarkSavedRevision(j.Snapshot().Revision))
	if j.Snapshot().Dirty {
		t.Fatal("current save not acknowledged")
	}
}

func TestBoundedEvictionAndClearPreserveDirtyState(t *testing.T) {
	j := journal(t, 2)
	v := 0
	check(t, j.Apply(set(&v, 1)))
	check(t, j.MarkSaved())
	for _, n := range []int{2, 3} {
		check(t, j.Apply(set(&v, n)))
	}
	if len(j.entries) != 2 {
		t.Fatal("limit", len(j.entries))
	}
	check(t, j.Undo())
	check(t, j.Undo())
	if v != 1 || j.Snapshot().Dirty {
		t.Fatal("evicted transition lost reachable save point", v, j.Snapshot())
	}
	if !errors.Is(j.Undo(), ErrEmpty) {
		t.Fatal("evicted callback remains")
	}
	check(t, j.Redo())
	check(t, j.Redo())
	if v != 3 {
		t.Fatal(v)
	}
	check(t, j.Clear())
	if !j.Snapshot().Dirty || j.Snapshot().UndoCount != 0 || j.entries != nil {
		t.Fatal(j.Snapshot())
	}
	stale := j.Snapshot().Revision
	check(t, j.Reset(true))
	if j.Snapshot().Dirty || !errors.Is(j.MarkSavedRevision(stale), ErrStale) {
		t.Fatal("replaced model accepted old save")
	}
	check(t, j.Reset(false))
	if !j.Snapshot().Dirty {
		t.Fatal("unsaved replacement clean")
	}
}

func TestFailedChangeUndoRedoKeepBothStacks(t *testing.T) {
	j := journal(t, 3)
	v := 0
	reject := errors.New("rejected")
	applyFails, revertFails := false, false
	c := Change{Label: "Change", Apply: func() error {
		if applyFails {
			return reject
		}
		v = 1
		return nil
	}, Revert: func() error {
		if revertFails {
			return reject
		}
		v = 0
		return nil
	}}
	check(t, j.Apply(c))
	before := j.Snapshot()
	revertFails = true
	if !errors.Is(j.Undo(), reject) || before != j.Snapshot() || v != 1 {
		t.Fatal("failed undo advanced", j.Snapshot(), v)
	}
	revertFails = false
	check(t, j.Undo())
	before = j.Snapshot()
	applyFails = true
	if !errors.Is(j.Redo(), reject) || before != j.Snapshot() || v != 0 {
		t.Fatal("failed redo advanced")
	}
	if !errors.Is(j.Apply(Change{Label: "Fail", Apply: func() error { return reject }, Revert: func() error { return nil }}), reject) || before != j.Snapshot() {
		t.Fatal("failed new edit discarded redo")
	}
	applyFails = false
	check(t, j.Redo())
}

func TestReentrantOperationsRefusedAndPanicsInvalidate(t *testing.T) {
	j := journal(t, 3)
	c := Change{Label: "Guard", Revert: func() error { return nil }, Apply: func() error {
		if !j.Snapshot().Busy {
			t.Fatal("missing busy state")
		}
		for _, err := range []error{j.Undo(), j.Redo(), j.Clear(), j.MarkSaved(), j.Reset(true), j.Apply(Change{})} {
			if !errors.Is(err, ErrBusy) {
				t.Fatal("nested operation", err)
			}
		}
		return nil
	}}
	check(t, j.Apply(c))
	func() {
		defer func() {
			if recover() != "callback panic" {
				t.Fatal("panic did not propagate")
			}
		}()
		j.Apply(Change{Label: "Panic", Apply: func() error { panic("callback panic") }, Revert: func() error { return nil }})
	}()
	if !j.Snapshot().Invalid || !j.Snapshot().Dirty || j.Snapshot().Busy || j.entries != nil {
		t.Fatal(j.Snapshot())
	}
	for _, err := range []error{j.Undo(), j.Redo(), j.MarkSaved(), j.Clear(), j.Apply(c)} {
		if !errors.Is(err, ErrInvalid) {
			t.Fatal("invalid replay", err)
		}
	}
	check(t, j.Reset(true))
	check(t, j.Apply(c))
}

func TestInvalidAdmissionAndRevisionExhaustionNeverCallModel(t *testing.T) {
	if _, err := New(0); err == nil {
		t.Fatal("zero limit")
	}
	j := journal(t, 2)
	calls := 0
	for _, c := range []Change{{}, {Label: " ", Apply: func() error { calls++; return nil }, Revert: func() error { return nil }}, {Label: "missing reverse", Apply: func() error { calls++; return nil }}} {
		if j.Apply(c) == nil {
			t.Fatal("invalid change accepted")
		}
	}
	j.next = math.MaxUint64
	if !errors.Is(j.Apply(Change{Label: "Exhausted", Apply: func() error { calls++; return nil }, Revert: func() error { return nil }}), ErrRevisionExhausted) {
		t.Fatal("revision wrap")
	}
	if !errors.Is(j.Reset(true), ErrRevisionExhausted) || calls != 0 {
		t.Fatal("exhausted callback/reset mutated")
	}
}

func TestReplaceAtomicallyResetsAndRejectsBeforeInstall(t *testing.T) {
	j := journal(t, 4)
	v := 0
	check(t, j.Apply(set(&v, 1)))
	check(t, j.Undo())
	before := j.Snapshot()
	failure := errors.New("install")
	if err := j.Replace(true, func() error { return failure }); !errors.Is(err, failure) || j.Snapshot() != before || v != 0 {
		t.Fatal("failed install changed history")
	}
	check(t, j.Replace(true, func() error {
		if !errors.Is(j.Replace(true, func() error { t.Fatal("nested install"); return nil }), ErrBusy) {
			t.Fatal("reentrant install")
		}
		v = 90
		return nil
	}))
	if v != 90 || j.Snapshot().Dirty || j.Snapshot().Revision == before.Revision || j.entries != nil {
		t.Fatal(j.Snapshot())
	}
	calls := 0
	j.next = math.MaxUint64
	if !errors.Is(j.Replace(true, func() error { calls++; return nil }), ErrRevisionExhausted) || calls != 0 {
		t.Fatal("exhausted install executed")
	}
}
func TestReplacementPanicInvalidatesAndCanRecover(t *testing.T) {
	j := journal(t, 4)
	func() {
		defer func() {
			if recover() != "install panic" {
				t.Fatal("panic suppressed")
			}
		}()
		j.Replace(true, func() error { panic("install panic") })
	}()
	if !j.Snapshot().Invalid || j.Snapshot().Busy || !j.Snapshot().Dirty {
		t.Fatal(j.Snapshot())
	}
	check(t, j.Replace(true, func() error { return nil }))
	if j.Snapshot().Invalid || j.Snapshot().Dirty {
		t.Fatal(j.Snapshot())
	}
}
