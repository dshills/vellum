package editor

import (
	"errors"
	"math"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/dshills/vellum/history"
)

func model(t *testing.T, text string, config Config) *Model {
	t.Helper()
	m, err := New(text, config)
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func edit(m *Model, r Range, before, insert string, s Selection) Edit {
	n := m.Snapshot()
	return Edit{BaseRevision: n.Revision, BaseVisibleEpoch: n.VisibleEpoch, Range: r, Before: before, Insert: insert, Selection: s}
}
func ok(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func refused(t *testing.T, m *Model, want error, f func() error) {
	t.Helper()
	before := m.Snapshot()
	if err := f(); !errors.Is(err, want) {
		t.Fatalf("got %v, want %v", err, want)
	}
	if after := m.Snapshot(); after != before {
		t.Fatalf("refusal changed model: before %#v after %#v", before, after)
	}
}

func TestScalarEditsPreserveBytesDirectionAndReplay(t *testing.T) {
	m := model(t, "a😀e\u0301\r\n\r\n", Config{})
	ok(t, m.Select(Selection{4, 1}))
	before := m.Snapshot()
	ok(t, m.Apply(edit(m, Range{1, 2}, "😀", "界", Selection{2, 1})))
	after := m.Snapshot()
	if after.Text != "a界e\u0301\r\n\r\n" || after.Selection != (Selection{2, 1}) || after.History.UndoCount != 1 {
		t.Fatal(after)
	}
	ok(t, m.Journal().Undo())
	undone := m.Snapshot()
	if undone.Text != before.Text || undone.Selection != before.Selection || undone.Revision <= after.Revision || undone.VisibleEpoch <= after.VisibleEpoch {
		t.Fatal(undone)
	}
	ok(t, m.Journal().Redo())
	redone := m.Snapshot()
	if redone.Text != after.Text || redone.Selection != after.Selection || redone.Revision <= undone.Revision || redone.VisibleEpoch <= undone.VisibleEpoch {
		t.Fatal(redone)
	}
}

func TestInvalidAdmissionIsAtomic(t *testing.T) {
	m := model(t, "a😀\r\n", Config{MaxBytes: 16})
	base := edit(m, Range{1, 2}, "😀", "b", Selection{2, 2})
	cases := []struct {
		name string
		e    Edit
		want error
	}{
		{"revision", func() Edit { e := base; e.BaseRevision--; return e }(), ErrStale},
		{"visible", func() Edit { e := base; e.BaseVisibleEpoch--; return e }(), ErrStale},
		{"before", func() Edit { e := base; e.Before = "x"; return e }(), ErrStale},
		{"negative", func() Edit { e := base; e.Range.Start = -1; return e }(), ErrRange},
		{"reversed", func() Edit { e := base; e.Range = Range{2, 1}; return e }(), ErrRange},
		{"outside", func() Edit { e := base; e.Range.End = 20; return e }(), ErrRange},
		{"utf8", func() Edit { e := base; e.Insert = "\xff"; return e }(), ErrInvalidText},
		{"nul", func() Edit { e := base; e.Insert = "b\x00"; return e }(), ErrInvalidText},
		{"bytes", func() Edit { e := base; e.Insert = strings.Repeat("界", 5); return e }(), ErrTooLarge},
		{"selection", func() Edit { e := base; e.Selection = Selection{10, 1}; return e }(), ErrRange},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) { refused(t, m, c.want, func() error { return m.Apply(c.e) }) })
	}
	badMarked := base
	badMarked.Marked = Range{0, 20}
	refused(t, m, ErrRange, func() error { return m.Preview(badMarked) })
	outsideMarked := base
	outsideMarked.Marked = Range{0, 1}
	refused(t, m, ErrRange, func() error { return m.Preview(outsideMarked) })
	refused(t, m, ErrRange, func() error { return m.Select(Selection{-1, 0}) })
	validMarked := base
	validMarked.Marked = Range{1, 2}
	ok(t, m.Preview(validMarked))
	refused(t, m, ErrRange, func() error { return m.Select(Selection{0, 0}) })
}

func TestCompositionCommitsOnceWithoutVisibleTransitionOnUnmark(t *testing.T) {
	m := model(t, "one\r\n", Config{})
	ok(t, m.Select(Selection{3, 0}))
	baseline := m.Snapshot()
	e := edit(m, Range{0, 3}, "one", "n", Selection{1, 1})
	e.Marked = Range{0, 1}
	ok(t, m.Preview(e))
	e = edit(m, Range{0, 1}, "n", "你好", Selection{2, 2})
	e.Marked = Range{0, 2}
	ok(t, m.Preview(e))
	preview := m.Snapshot()
	if preview.Text != baseline.Text || preview.VisibleText != "你好\r\n" || preview.Revision != baseline.Revision || preview.History != baseline.History || !preview.Composing {
		t.Fatal(preview)
	}
	refused(t, m, ErrComposition, func() error { return m.Replace("bad", Selection{}) })
	ok(t, m.CommitComposition())
	committed := m.Snapshot()
	if committed.Text != preview.VisibleText || committed.Composing || committed.History.UndoCount != 1 || committed.Revision <= preview.Revision || committed.VisibleEpoch != preview.VisibleEpoch || committed.SelectionEpoch != preview.SelectionEpoch {
		t.Fatal(committed)
	}
	ok(t, m.CommitComposition())
	if m.Snapshot() != committed {
		t.Fatal("duplicate unmark changed model")
	}
	ok(t, m.Journal().Undo())
	if m.Snapshot().Text != baseline.Text || m.Snapshot().Selection != baseline.Selection {
		t.Fatal(m.Snapshot())
	}
	ok(t, m.Journal().Redo())
	if m.Snapshot().Text != committed.Text || m.Snapshot().Selection != committed.Selection {
		t.Fatal(m.Snapshot())
	}
}

func TestNoNetCancelAndSelectionOnlyPreserveRedo(t *testing.T) {
	m := model(t, "a", Config{})
	ok(t, m.Replace("b", Selection{1, 0}))
	ok(t, m.Journal().Undo())
	h := m.Journal().Snapshot()
	rev := m.Snapshot().Revision
	ok(t, m.Apply(edit(m, Range{0, 1}, "a", "a", Selection{1, 0})))
	if m.Journal().Snapshot() != h || m.Snapshot().Revision != rev {
		t.Fatal(m.Snapshot())
	}
	e := edit(m, Range{0, 1}, "a", "あ", Selection{1, 1})
	e.Marked = Range{0, 1}
	ok(t, m.Preview(e))
	refused(t, m, ErrComposition, m.Journal().Redo)
	ok(t, m.CancelComposition())
	if m.Journal().Snapshot() != h || m.Snapshot().Text != "a" || m.Snapshot().Selection != (Selection{1, 0}) {
		t.Fatal(m.Snapshot())
	}
	e = edit(m, Range{0, 1}, "a", "x", Selection{1, 1})
	e.Marked = Range{0, 1}
	ok(t, m.Preview(e))
	e = edit(m, Range{0, 1}, "x", "a", Selection{0, 1})
	e.Marked = Range{0, 1}
	ok(t, m.Preview(e))
	before := m.Snapshot()
	ok(t, m.CommitComposition())
	if m.Journal().Snapshot() != h || m.Snapshot().Revision != rev || m.Snapshot().Selection != before.Selection || m.Snapshot().SelectionEpoch != before.SelectionEpoch {
		t.Fatal(m.Snapshot())
	}
	ok(t, m.Journal().Redo())
	if m.Snapshot().Text != "b" {
		t.Fatal(m.Snapshot())
	}
}

func TestClocksRejectCommittedAndPreviewABA(t *testing.T) {
	m := model(t, "a", Config{})
	old := edit(m, Range{0, 1}, "a", "c", Selection{})
	ok(t, m.Replace("b", Selection{}))
	ok(t, m.Journal().Undo())
	refused(t, m, ErrStale, func() error { return m.Apply(old) })
	e := edit(m, Range{0, 1}, "a", "x", Selection{})
	ok(t, m.Preview(e))
	oldPreview := edit(m, Range{0, 1}, "x", "y", Selection{})
	ok(t, m.CancelComposition())
	e = edit(m, Range{0, 1}, "a", "x", Selection{})
	ok(t, m.Preview(e))
	if m.Snapshot().Revision != oldPreview.BaseRevision {
		t.Fatal("previews changed committed revision")
	}
	refused(t, m, ErrStale, func() error { return m.Preview(oldPreview) })
}

func TestEpochExhaustionPreservesEveryState(t *testing.T) {
	for _, clock := range []string{"committed", "visible", "selection"} {
		t.Run(clock, func(t *testing.T) {
			m := model(t, "a", Config{})
			switch clock {
			case "committed":
				m.revision = math.MaxUint64
			case "visible":
				m.visibleEpoch = math.MaxUint64
			case "selection":
				m.selectionEpoch = math.MaxUint64
			}
			refused(t, m, ErrRevisionExhausted, func() error { return m.Apply(edit(m, Range{0, 1}, "a", "b", Selection{1, 1})) })
		})
	}
	m := model(t, "a", Config{})
	ok(t, m.Replace("b", Selection{}))
	m.revision = math.MaxUint64
	refused(t, m, ErrRevisionExhausted, m.Journal().Undo)
	m = model(t, "a", Config{})
	e := edit(m, Range{0, 1}, "a", "x", Selection{})
	ok(t, m.Preview(e))
	m.visibleEpoch = math.MaxUint64
	refused(t, m, ErrRevisionExhausted, m.CancelComposition)
	// Unmark with unchanged visible bytes does not need a fresh visible epoch.
	ok(t, m.CommitComposition())
	m = model(t, "a", Config{})
	m.selectionEpoch = math.MaxUint64
	refused(t, m, ErrRevisionExhausted, func() error { return m.Select(Selection{1, 0}) })
	// Undo callback exhaustion preserves both journal stacks and current bytes.
	m = model(t, "a", Config{})
	ok(t, m.Replace("b", Selection{}))
	ok(t, m.Journal().Undo())
	m.visibleEpoch = math.MaxUint64
	refused(t, m, ErrRevisionExhausted, m.Journal().Redo)
	m = model(t, "a", Config{})
	e = edit(m, Range{0, 1}, "a", "x", Selection{})
	ok(t, m.Preview(e))
	m.revision = math.MaxUint64
	refused(t, m, ErrRevisionExhausted, m.CommitComposition)
	m = model(t, "a", Config{})
	m.revision = math.MaxUint64
	refused(t, m, ErrRevisionExhausted, func() error {
		return m.Journal().Replace(true, func() error { return m.Install("a", Selection{}) })
	})
}

func TestUndoRefusesCompositionAndSelectionRestorationExhaustion(t *testing.T) {
	m := model(t, "a", Config{})
	ok(t, m.Replace("b", Selection{1, 1}))
	e := edit(m, Range{0, 1}, "b", "x", Selection{1, 1})
	e.Marked = Range{0, 1}
	ok(t, m.Preview(e))
	refused(t, m, ErrComposition, m.Journal().Undo)
	m.selectionEpoch = math.MaxUint64
	// Cancellation restores the baseline selection (already the same), then an
	// undo which needs a different selection must refuse atomically.
	ok(t, m.CancelComposition())
	refused(t, m, ErrRevisionExhausted, m.Journal().Undo)
}

func TestAtomicInstallReplacesHistoryOnceAndInvalidatesSameBytes(t *testing.T) {
	m := model(t, "a", Config{})
	ok(t, m.Replace("b", Selection{1, 0}))
	before := m.Snapshot()
	refused(t, m, ErrInstallContext, func() error { return m.Install("c", Selection{}) })
	refused(t, m, ErrInvalidText, func() error { return m.Journal().Replace(true, func() error { return m.Install("\xff", Selection{}) }) })
	ok(t, m.Journal().Replace(true, func() error { return m.Install("b", Selection{}) }))
	after := m.Snapshot()
	if after.Text != "b" || after.Revision <= before.Revision || after.VisibleEpoch != before.VisibleEpoch || after.History.Revision != before.History.Revision+1 || after.History.Dirty || after.History.UndoCount != 0 {
		t.Fatal(after)
	}
	refused(t, m, ErrStale, func() error {
		return m.Apply(Edit{BaseRevision: before.Revision, BaseVisibleEpoch: before.VisibleEpoch, Range: Range{0, 1}, Before: "b", Insert: "c"})
	})
}

func TestInitializationAndBounds(t *testing.T) {
	var nilModel *Model
	for _, m := range []*Model{nilModel, {}} {
		if !errors.Is(m.Validate(), ErrInvalidModel) || m.Journal() != nil || m.Snapshot() != (Snapshot{}) || m.MaxBytes() != 0 {
			t.Fatal("unsafe zero model")
		}
		if !errors.Is(m.Select(Selection{}), ErrInvalidModel) {
			t.Fatal("zero mutation admitted")
		}
	}
	for _, c := range []Config{{MaxBytes: -1}, {HistoryLimit: -1}} {
		if _, err := New("", c); err == nil {
			t.Fatal("negative config")
		}
	}
	for _, s := range []string{"\xff", "nul\x00"} {
		if _, err := New(s, Config{}); !errors.Is(err, ErrInvalidText) {
			t.Fatal(err)
		}
	}
	if _, err := New("界", Config{MaxBytes: 2}); !errors.Is(err, ErrTooLarge) {
		t.Fatal(err)
	}
	m := model(t, "", Config{HistoryLimit: 2})
	if m.MaxBytes() != DefaultMaxBytes || model(t, "", Config{MaxBytes: 17}).MaxBytes() != 17 {
		t.Fatal("byte budget accessor")
	}
	if m.Snapshot().History.Dirty || m.Snapshot().Revision != 1 || m.Snapshot().VisibleEpoch != 1 || m.Snapshot().SelectionEpoch != 1 {
		t.Fatal(m.Snapshot())
	}
	for _, s := range []string{"a", "b", "c"} {
		ok(t, m.Replace(s, Selection{}))
	}
	if m.Snapshot().History.UndoCount != 2 {
		t.Fatal(m.Snapshot())
	}
	ok(t, m.Journal().Undo())
	ok(t, m.Journal().Undo())
	if m.Snapshot().Text != "a" {
		t.Fatal(m.Snapshot())
	}
	refused(t, m, history.ErrEmpty, m.Journal().Undo)
}

func TestSparseDifferenceIsScalarAligned(t *testing.T) {
	for _, c := range [][2]string{{"prefix😀suffix", "prefix界suffix"}, {"e\u0301", "e"}, {"\r\n", "\n"}, {"", "😀"}, {"ab", "a😀b"}} {
		p := difference(c[0], c[1])
		m := model(t, c[0], Config{})
		next, err := m.replaceSpan(c[0], Range{p.start, p.start + utf8.RuneCountInString(p.before)}, p.before, p.insert)
		ok(t, err)
		if next != c[1] || !utf8.ValidString(p.before) || !utf8.ValidString(p.insert) {
			t.Fatal(c, p, next)
		}
	}
	p := difference(strings.Repeat("a", 10000)+"😀"+strings.Repeat("b", 10000), strings.Repeat("a", 10000)+"界"+strings.Repeat("b", 10000))
	if p.before != "😀" || p.insert != "界" || p.start != 10000 {
		t.Fatal("composition retained unchanged spans", p)
	}
}

func TestCopiedModelCannotAliasHistoryAuthority(t *testing.T) {
	m := model(t, "a", Config{})
	ok(t, m.Replace("b", Selection{1, 1}))
	original := m.Snapshot()
	copied := *m
	if !errors.Is(copied.Validate(), ErrInvalidModel) || copied.Journal() != nil || copied.MaxBytes() != 0 || copied.Snapshot() != (Snapshot{}) {
		t.Fatal("copied model exposes the original model's authority")
	}
	mutations := []func() error{
		func() error { return copied.Replace("c", Selection{}) },
		func() error { return copied.Apply(Edit{}) },
		func() error { return copied.Preview(Edit{}) },
		func() error { return copied.Select(Selection{}) },
		copied.CommitComposition, copied.CancelComposition,
		func() error { return copied.Install("c", Selection{}) },
	}
	for _, mutation := range mutations {
		if err := mutation(); !errors.Is(err, ErrInvalidModel) {
			t.Fatalf("copied mutation returned %v", err)
		}
		if m.Snapshot() != original {
			t.Fatal("copied mutation changed the original")
		}
	}
	ok(t, m.Journal().Undo())
	if m.Snapshot().Text != "a" || m.Snapshot().History.UndoCount != 0 || m.Snapshot().History.RedoCount != 1 {
		t.Fatal("original history no longer belongs to the original model", m.Snapshot())
	}
}
