package main

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"
	"unsafe"

	"github.com/dshills/vellum/virtual"
)

func TestEditableSourceSparseImmutableVersionsAndProjection(t *testing.T) {
	s, err := newSource("editable", 1_000_000)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := any(s).(virtual.GridEditSource); ok {
		t.Fatal("default source unexpectedly opted into editing")
	}
	key, column := rowKey(999_999), columnID(19)
	base, err := s.cellValue(key, column)
	if err != nil || base.version != 1 || len(s.overrides) != 0 {
		t.Fatal(base, err)
	}
	large := strings.Repeat("x", 1<<20)
	text := large[100:105]
	next, err := s.withCell(key, column, text, base.version)
	if err != nil {
		t.Fatal(err)
	}
	value, err := next.cellValue(key, column)
	if err != nil || value.text != text || value.version != 2 || next.revision != 2 || len(next.overrides) != 1 {
		t.Fatal(value, next.revision, err)
	}
	if unsafe.StringData(value.text) == unsafe.StringData(text) {
		t.Fatal("bounded override retained the large caller backing")
	}
	if prior, e := s.cellValue(key, column); e != nil || prior != base || s.revision != 1 || len(s.overrides) != 0 {
		t.Fatal("candidate mutated its original revision", prior, e)
	}
	projected, err := next.projected(true, true)
	if err != nil {
		t.Fatal(err)
	}
	if hidden, e := (&editableGridSource{projected}).EditCell(key, column); e != nil || hidden.Text != text || hidden.Version != 2 || !hidden.Editable {
		t.Fatal("stable domain value lost while filtered out", hidden, e)
	}
	if _, e := projected.Cell(key, column); !errors.Is(e, virtual.ErrGridRange) {
		t.Fatal("projected Cell admitted a hidden row", e)
	}
	readOnly, err := (&editableGridSource{s}).EditCell(rowKey(0), columnID(0))
	if err != nil || readOnly.Editable || readOnly.Version != 1 {
		t.Fatal("source-controlled read-only column", readOnly, err)
	}
	metrics := &rowMetricCalls{}
	variable := cellSource(projected, metrics, true, true)
	if _, ok := variable.(virtual.GridRowMetrics); !ok {
		t.Fatal("editable wrapper lost optional row metrics")
	}
	if _, ok := variable.(virtual.GridSelectionSource); !ok {
		t.Fatal("editable wrapper lost immutable selection capture")
	}
	if _, ok := cellSource(s, metrics, true, false).(virtual.GridEditSource); ok {
		t.Fatal("variable default source unexpectedly opted into editing")
	}
}

func TestEditableSourceRefusalsAndBoundedOverrides(t *testing.T) {
	s, err := newSource("editable", 10_000)
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"", "  \t", "bad\nline", "bad\rline", "bad\x00value", string([]byte{0xff}), strings.Repeat("a", maxCellTextBytes+1)} {
		if _, err := s.withCell(rowKey(0), columnID(1), text, 1); !errors.Is(err, errCellValidation) || s.revision != 1 || len(s.overrides) != 0 {
			t.Fatal("validation altered source", text, err)
		}
	}
	if _, err := s.withCell(rowKey(0), columnID(1), "valid", 2); !errors.Is(err, errCellConflict) {
		t.Fatal("wrong base version accepted", err)
	}
	for _, address := range []cellAddress{{"invalid", columnID(1)}, {rowKey(10_000), columnID(1)}, {rowKey(0), "invalid"}} {
		if _, err := s.withCell(address.key, address.column, "valid", 1); !errors.Is(err, virtual.ErrGridRange) {
			t.Fatal(address, err)
		}
	}
	full := *s
	full.overrides = make(map[cellAddress]cellOverride, maxCellOverrides)
	for i := 0; i < maxCellOverrides; i++ {
		full.overrides[cellAddress{rowKey(i), columnID(1)}] = cellOverride{"existing", 2}
	}
	if _, err := full.withCell(rowKey(maxCellOverrides), columnID(1), "new", 1); !errors.Is(err, virtual.ErrGridCapacity) || full.revision != 1 || len(full.overrides) != maxCellOverrides {
		t.Fatal("override capacity exceeded", err)
	}
	updated, err := full.withCell(rowKey(0), columnID(1), "updated", 2)
	if err != nil || len(updated.overrides) != maxCellOverrides || full.overrides[cellAddress{rowKey(0), columnID(1)}].text != "existing" {
		t.Fatal("existing override could not update atomically at capacity", err)
	}
	exhausted := *s
	exhausted.revision = math.MaxUint64
	if _, err := exhausted.withCell(rowKey(0), columnID(1), "new", 1); !errors.Is(err, virtual.ErrGridExhausted) {
		t.Fatal("source revision wrapped", err)
	}
	exhausted = *s
	exhausted.overrides = map[cellAddress]cellOverride{{rowKey(0), columnID(1)}: {"existing", math.MaxUint64}}
	if _, err := exhausted.withCell(rowKey(0), columnID(1), "new", math.MaxUint64); !errors.Is(err, virtual.ErrGridExhausted) {
		t.Fatal("cell version wrapped", err)
	}
}

func TestCellEditCommandAtomicSaveConflictAndHiddenRecovery(t *testing.T) {
	p, err := newProgramOptions(1024, 760, true, true, true)
	if err != nil {
		t.Fatal(err)
	}
	defer p.owner.Close()
	if err = p.owner.Do(func() {
		w := p.windows[1]
		if e := w.grid.JumpToKey(rowKey(1), columnID(1)); e != nil {
			t.Fatal(e)
		}
		if e := w.beginCell(); e != nil {
			t.Fatal(e)
		}
		originalSource := w.source
		original, e := w.grid.Edit()
		if e != nil {
			t.Fatal(e)
		}
		if e = w.grid.SetEditText("edited 😀 é עברית"); e != nil {
			t.Fatal(e)
		}
		if e = w.projection(true, false); e != nil {
			t.Fatal(e)
		}
		if e = w.setVariableRows(false); e != nil {
			t.Fatal(e)
		}
		if e = w.projection(true, true); e != nil {
			t.Fatal(e)
		}
		hidden, e := w.grid.Edit()
		if e != nil || hidden.Status != virtual.GridEditUnavailable || hidden.Text != "edited 😀 é עברית" || hidden.Original != original.Original {
			t.Fatal("hidden draft not recoverable", hidden, e)
		}
		before := w.source
		if e = w.commitCell(); !errors.Is(e, virtual.ErrGridEditConflict) || w.source != before {
			t.Fatal("hidden edit saved or mutated source", e)
		}
		if e = w.projection(true, false); e != nil {
			t.Fatal(e)
		}
		if e = w.grid.JumpToKey(original.RowKey, original.ColumnID); e != nil {
			t.Fatal(e)
		}
		request, e := w.grid.PrepareEdit()
		if e != nil {
			t.Fatal(e)
		}
		r := request.State()
		candidate, e := w.source.withCell(r.RowKey, r.ColumnID, r.Text, r.BaseVersion)
		if e != nil {
			t.Fatal(e)
		}
		if e = w.grid.SetEditText("latest draft"); e != nil {
			t.Fatal(e)
		}
		before = w.source
		if e = w.grid.CompleteEdit(request, cellSource(candidate, w.metricCalls, w.variableRows, true)); !errors.Is(e, virtual.ErrGridStale) || w.source != before {
			t.Fatal("retired request accepted", e)
		}
		if e = w.commitCell(); e != nil {
			t.Fatal(e)
		}
		saved, e := w.source.cellValue(rowKey(1), columnID(1))
		if e != nil || saved.text != "latest draft" || saved.version != 2 || len(w.source.overrides) != 1 {
			t.Fatal("atomic committed source", saved, e)
		}
		if old, e := originalSource.cellValue(rowKey(1), columnID(1)); e != nil || old.text != original.Original || old.version != 1 {
			t.Fatal("commit mutated prior source", old, e)
		}
		if e = w.beginCell(); e != nil {
			t.Fatal(e)
		}
		if e = w.grid.SetEditText("draft must not overwrite external"); e != nil {
			t.Fatal(e)
		}
		if e = w.externalCell(); e != nil {
			t.Fatal(e)
		}
		conflict, e := w.grid.Edit()
		if e != nil || conflict.Status != virtual.GridEditConflict || conflict.CurrentVersion != 3 || conflict.BaseVersion != 2 || conflict.Text != "draft must not overwrite external" {
			t.Fatal("external version not surfaced", conflict, e)
		}
		before = w.source
		if e = w.commitCell(); !errors.Is(e, virtual.ErrGridEditConflict) || w.source != before {
			t.Fatal("conflict overwritten", e)
		}
		if e = w.cancelCell(); e != nil {
			t.Fatal(e)
		}
		if e = w.beginCell(); e != nil {
			t.Fatal(e)
		}
		if e = w.grid.SetEditText("   "); e != nil {
			t.Fatal(e)
		}
		before = w.source
		if e = w.commitCell(); !errors.Is(e, errCellValidation) || w.source != before {
			t.Fatal("domain validation did not preserve draft/source", e)
		}
		if e = w.grid.SetEditText("corrected draft"); e != nil {
			t.Fatal(e)
		}
		if e = w.commitCell(); e != nil {
			t.Fatal(e)
		}
		current, e := w.grid.Edit()
		other, otherErr := p.windows[0].grid.Edit()
		if e != nil || otherErr != nil || current.Active || other.Active || len(p.windows[0].source.overrides) != 0 {
			t.Fatal("independent edit lifetime", current, other, e, otherErr)
		}
	}); err != nil {
		t.Fatal(err)
	}
}

func TestCellCommitSettlesBeforeCaptureAndRejectsRetiredSession(t *testing.T) {
	p, err := newProgramOptions(1024, 760, false, false, true)
	if err != nil {
		t.Fatal(err)
	}
	defer p.owner.Close()
	if err := p.owner.Do(func() {
		w := p.windows[1]
		if err := w.grid.JumpToKey(rowKey(2), columnID(1)); err != nil {
			t.Fatal(err)
		}
		if err := w.beginCell(); err != nil {
			t.Fatal(err)
		}
		if err := w.grid.SetEditText("first draft"); err != nil {
			t.Fatal(err)
		}
		var ack func(error)
		var canceled <-chan struct{}
		w.settleCell = func(ctx context.Context, callback func(error)) error {
			ack, canceled = callback, ctx.Done()
			return nil
		}
		before := w.source
		if err := w.requestCellCommit(); err != nil || !w.cellCommitPending || w.source != before {
			t.Fatal("settle did not defer domain capture", err)
		}
		if err := w.requestCellCommit(); !errors.Is(err, virtual.ErrGridEditBusy) {
			t.Fatal("duplicate pending Save admitted", err)
		}
		// A preceding originating edit may become owner-visible only before the
		// asynchronous acknowledgment, so Save must read after that boundary.
		if err := w.grid.SetEditText("last earlier admitted edit"); err != nil {
			t.Fatal(err)
		}
		ack(nil)
		value, err := w.source.cellValue(rowKey(2), columnID(1))
		if err != nil || value.text != "last earlier admitted edit" || value.version != 2 || w.cellCommitPending {
			t.Fatal("Save captured before settlement", value, err)
		}
		select {
		case <-canceled:
		default:
			t.Fatal("completed settlement context retained")
		}
		if err := w.beginCell(); err != nil {
			t.Fatal(err)
		}
		if err := w.grid.SetEditText("cancel this request"); err != nil {
			t.Fatal(err)
		}
		if err := w.requestCellCommit(); err != nil {
			t.Fatal(err)
		}
		oldAck := ack
		if err := w.cancelCell(); err != nil {
			t.Fatal(err)
		}
		select {
		case <-canceled:
		default:
			t.Fatal("Cancel did not cancel pending originating settlement")
		}
		if err := w.beginCell(); err != nil {
			t.Fatal(err)
		}
		if err := w.grid.SetEditText("next session"); err != nil {
			t.Fatal(err)
		}
		if err := w.requestCellCommit(); err != nil {
			t.Fatal(err)
		}
		before = w.source
		oldAck(nil)
		if !w.cellCommitPending || w.source != before {
			t.Fatal("retired settlement crossed new pending request")
		}
		ack(context.Canceled)
		if w.cellCommitPending || w.source != before {
			t.Fatal("canceled acknowledgment mutated source")
		}
		if edit, err := w.grid.Edit(); err != nil || !edit.Active || edit.Text != "next session" {
			t.Fatal("canceled settlement dropped recoverable draft", edit, err)
		}
		refused := errors.New("settle refused")
		w.settleCell = func(context.Context, func(error)) error { return refused }
		if err := w.requestCellCommit(); !errors.Is(err, refused) || w.cellCommitPending || w.source != before {
			t.Fatal("refused settlement left pending/domain state", err)
		}
	}); err != nil {
		t.Fatal(err)
	}
}

func TestCellEditIndexedWorkIndependentOfDomainRows(t *testing.T) {
	p, err := newProgramOptions(1024, 760, false, false, true)
	if err != nil {
		t.Fatal(err)
	}
	defer p.owner.Close()
	if err := p.owner.Do(func() {
		var previous sourceCalls
		for i, w := range p.windows {
			if err := w.grid.JumpToKey(rowKey(w.source.total/2), columnID(1)); err != nil {
				t.Fatal(err)
			}
			*w.source.calls = sourceCalls{}
			if err := w.beginCell(); err != nil {
				t.Fatal(err)
			}
			if err := w.grid.SetEditText("same bounded edit"); err != nil {
				t.Fatal(err)
			}
			if err := w.commitCell(); err != nil {
				t.Fatal(err)
			}
			calls := *w.source.calls
			if calls.RowKey > 4096 || calls.IndexOf > 4096 || calls.Cell > 4096 || len(w.source.overrides) != 1 {
				t.Fatal("one-cell edit performed source-sized work", calls)
			}
			if i != 0 && calls != previous {
				t.Fatal("10k/1M same viewport edit callback counts differ", previous, calls)
			}
			previous = calls
		}
	}); err != nil {
		t.Fatal(err)
	}
}
