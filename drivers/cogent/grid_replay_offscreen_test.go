//go:build offscreen

package cogent

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

// This fixture actually retains twenty uint32 fields per row. Its immutable
// index is independent of the live Grid and can outlive a source replacement.
type replayGridIndex struct {
	rows     int
	revision uint64
	reverse  bool
}

func (s replayGridIndex) Identity() string { return "stored-replay" }
func (s replayGridIndex) Revision() uint64 { return s.revision }
func (s replayGridIndex) RowCount() int    { return s.rows }
func (s replayGridIndex) RowKey(i int) (string, error) {
	if i < 0 || i >= s.rows {
		return "", virtual.ErrGridRange
	}
	if s.reverse {
		i = s.rows - 1 - i
	}
	return "record-" + strconv.Itoa(i), nil
}
func (s replayGridIndex) IndexOf(k string) (int, bool, error) {
	i, err := strconv.Atoi(strings.TrimPrefix(k, "record-"))
	if err != nil || k != "record-"+strconv.Itoa(i) || i < 0 || i >= s.rows {
		return 0, false, nil
	}
	if s.reverse {
		i = s.rows - 1 - i
	}
	return i, true, nil
}

type replayGridSource struct {
	replayGridIndex
	data                      [][20]uint32
	editRow, editColumn, text string
	version                   uint64
}

func (s *replayGridSource) Cell(k, column string) (string, error) {
	_, found, err := s.IndexOf(k)
	if err != nil || !found {
		return "", virtual.ErrGridRange
	}
	i, _ := strconv.Atoi(strings.TrimPrefix(k, "record-"))
	c, err := strconv.Atoi(strings.TrimPrefix(column, "c"))
	if err != nil || column != "c"+strconv.Itoa(c) || c < 0 || c >= 20 {
		return "", virtual.ErrGridRange
	}
	if k == s.editRow && column == s.editColumn {
		return s.text, nil
	}
	return strconv.FormatUint(uint64(s.data[i][c]), 10), nil
}
func (s *replayGridSource) EditCell(k, c string) (virtual.GridEditValue, error) {
	text, err := s.Cell(k, c)
	v := uint64(1)
	if k == s.editRow && c == s.editColumn {
		v = s.version
	}
	return virtual.GridEditValue{Text: text, Version: v, Editable: true}, err
}
func (s *replayGridSource) SelectionSnapshot() (virtual.GridRowIndex, error) {
	return s.replayGridIndex, nil
}

type replayGridFixture struct {
	m       *Manual
	a       *app.Application
	g       *virtual.Grid
	source  *replayGridSource
	journal *history.Journal
	commits int
	failure error
}

func newReplayGrid(t *testing.T, rows, batch int) *replayGridFixture {
	t.Helper()
	f := &replayGridFixture{source: &replayGridSource{replayGridIndex: replayGridIndex{rows: rows, revision: 1}, data: make([][20]uint32, rows)}}
	for i := range f.source.data {
		for c := range 20 {
			f.source.data[i][c] = uint32(i*20 + c)
		}
	}
	columns := make([]virtual.GridColumn, 20)
	for c := range columns {
		columns[c] = virtual.GridColumn{ID: fmt.Sprintf("c%d", c), Title: fmt.Sprintf("Field %d", c), Width: 90}
	}
	var err error
	f.g, err = virtual.NewGrid(f.source, columns, virtual.GridConfig{RowHeight: 24, HeaderHeight: 26, FrozenColumns: 2, RowSelection: true, CellEditing: true}, virtual.GridViewport{Width: 700, Height: 400})
	if err != nil {
		t.Fatal(err)
	}
	f.journal, err = history.New(8)
	if err != nil {
		t.Fatal(err)
	}
	f.a, err = app.New(app.Config{QueueCapacity: 32})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(f.a.Close)
	if err = f.a.Open(app.Window{ID: "grid", Title: "Stored replay", Width: 860, Height: 620, View: f.view}); err != nil {
		t.Fatal(err)
	}
	if err = f.a.Open(app.Window{ID: "inspector", Title: "Independent survivor", Width: 400, Height: 220, View: func() ui.Node {
		return ui.Node{ID: "summary", Kind: ui.Label, Text: fmt.Sprintf("Revision %d; commits %d", f.source.revision, f.commits)}
	}}); err != nil {
		t.Fatal(err)
	}
	config := manualConfig(t, batch)
	config.LogicalDPI = 96
	f.m, err = NewManual(testContext, f.a, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.m.Close(testContext); err != nil {
			t.Error(err)
		}
	})
	return f
}

func (f *replayGridFixture) invalidate() {
	for _, w := range f.a.Windows() {
		if err := f.a.Invalidate(w.ID); err != nil {
			f.failure = errors.Join(f.failure, err)
		}
	}
}
func (f *replayGridFixture) replace(text string) error {
	next := *f.source
	next.revision++
	next.version++
	next.text = text
	if err := f.g.ReplaceSource(&next); err != nil {
		return err
	}
	f.source = &next
	f.invalidate()
	return nil
}
func (f *replayGridFixture) commit(e ui.GridEditEvent) {
	request, err := f.g.PrepareEdit()
	if err != nil {
		f.failure = err
		return
	}
	before := e.State.Original
	next := *f.source
	next.revision++
	next.editRow, next.editColumn, next.text = e.State.RowKey, e.State.ColumnID, e.State.Text
	next.version = e.State.BaseVersion + 1
	first := true
	err = f.journal.Apply(history.Change{Label: "Grid field", Apply: func() error {
		if !first {
			return f.replace(e.State.Text)
		}
		if err := f.g.CompleteEdit(request, &next); err != nil {
			return err
		}
		first = false
		f.source = &next
		return nil
	}, Revert: func() error { return f.replace(before) }})
	if err != nil {
		f.failure = err
		return
	}
	f.commits++
	f.invalidate()
}
func (f *replayGridFixture) view() ui.Node {
	button := func(id string, disabled bool, fn func() error) ui.Node {
		return ui.Node{ID: id, Kind: ui.Button, Text: id, Disabled: disabled, Activate: func() {
			if err := fn(); err != nil {
				f.failure = err
			}
			f.invalidate()
		}}
	}
	h := f.journal.Snapshot()
	return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{
		{ID: "tools", Kind: ui.Row, Children: []ui.Node{
			button("reverse", false, func() error {
				next := *f.source
				next.revision++
				next.reverse = !next.reverse
				if err := f.g.ReplaceSource(&next); err != nil {
					return err
				}
				f.source = &next
				return nil
			}),
			button("undo", h.UndoCount == 0, f.journal.Undo), button("redo", h.RedoCount == 0, f.journal.Redo),
		}},
		{ID: "data", Kind: ui.DataGrid, Grid: f.g, Grow: true, GridCommit: f.commit},
	}}
}

// Inputs are measured once in the pilot and replayed verbatim, including raw
// physical coordinates. No replay resolves a fresh cell or invokes a callback.
type replayGridStep struct {
	label  string
	inputs []ManualInput
}
type replayGridCheckpoint struct {
	Label                   string
	Source                  virtual.GridStamp
	Active                  virtual.GridActive
	Selection               virtual.GridSelectionState
	Edit                    virtual.GridEditState
	Viewport                virtual.GridViewport
	Cells                   []GridCellGeometry
	Headers                 []GridHeaderGeometry
	Editor                  *GridEditorGeometry
	Elements                [][]ManualElement
	Pixels                  []replayGridImage
	History                 history.State
	Commits                 int
	EditedRow, EditedColumn string
	EditedCell              virtual.GridEditValue
}

type replayGridImage struct {
	Bounds image.Rectangle
	Stride int
	SHA256 string
}

func replayGridSettle(t *testing.T, f *replayGridFixture, batch int) {
	t.Helper()
	for range 2048 {
		s, err := f.m.Snapshot(testContext)
		if err != nil || s.Failure != nil || f.failure != nil {
			t.Fatal("replay readiness", err, s.Failure, f.failure)
		}
		ready := s.Application.Ordinary+s.Application.ReadyCompletions > 0
		for _, w := range s.Windows {
			ready = ready || w.Events > 0 || w.ReadyPaint
		}
		if ready {
			r, err := f.m.Step(testContext, batch)
			if err != nil || r.Events+r.OwnerCallbacks > batch || r.Events+r.OwnerCallbacks == 0 {
				t.Fatal("bounded replay dispatch", r, err)
			}
			continue
		}
		requested := false
		for _, w := range s.Windows {
			if w.LayoutPending {
				if err := f.m.RequestFrame(testContext, w.ID); err != nil {
					t.Fatal(err)
				}
				requested = true
			}
		}
		if !requested {
			return
		}
	}
	t.Fatal("stored replay exhausted explicit dispatch bound")
}
func replayGridCapture(t *testing.T, f *replayGridFixture, label string) replayGridCheckpoint {
	t.Helper()
	s, err := f.m.Snapshot(testContext)
	if err != nil || len(s.Windows) != 2 {
		t.Fatal("two-window replay", s, err)
	}
	in := gridInspection(t, f.a, f.g)
	wantReversed := label != "initial" && label != "selected" && label != "range" && label != "scrolled"
	if f.source.reverse != wantReversed {
		t.Fatal("source reverse transition absent", label)
	}
	for _, cell := range in.Cells {
		original := cell.Row
		if wantReversed {
			original = len(f.source.data) - 1 - original
		}
		wantKey := "record-" + strconv.Itoa(original)
		wantText := strconv.Itoa(original*20 + cell.Column)
		if wantKey == f.source.editRow && cell.ColumnID == f.source.editColumn {
			wantText = f.source.text
		}
		if cell.RowKey != wantKey || cell.ColumnID != "c"+strconv.Itoa(cell.Column) || cell.Text != wantText {
			t.Fatalf("%s cell row=%d column=%d got %s/%s=%s want %s=%s", label, cell.Row, cell.Column, cell.RowKey, cell.ColumnID, cell.Text, wantKey, wantText)
		}
	}
	var state virtual.GridState
	if err := f.a.Do(func() { state, err = f.g.State() }); err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	c := replayGridCheckpoint{Label: label, Source: state.Stamp, Active: state.Active, Selection: state.Selection, Edit: state.Edit, Viewport: state.Viewport, Cells: in.Cells, Headers: in.Headers, Editor: in.Editor, Commits: f.commits}
	if err := f.a.Do(func() {
		c.History = f.journal.Snapshot()
		c.EditedRow, c.EditedColumn = f.source.editRow, f.source.editColumn
		if c.EditedRow != "" {
			c.EditedCell, err = f.source.EditCell(c.EditedRow, c.EditedColumn)
		}
	}); err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range s.Windows {
		if w.LayoutPending || w.Events != 0 || w.ReadyPaint {
			t.Fatal("capture before explicit completion", label, w.ID)
		}
		// Native incarnations are process-global retirement identities. Keep IDs,
		// measured geometry, values, selection and focus; omit only those tokens.
		for i := range w.Elements {
			w.Elements[i].Incarnation = 0
		}
		c.Elements = append(c.Elements, w.Elements)
		pixels, err := f.m.Capture(testContext, w.ID)
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(pixels.Pix)
		c.Pixels = append(c.Pixels, replayGridImage{Bounds: pixels.Bounds(), Stride: pixels.Stride, SHA256: hex.EncodeToString(h[:])})
	}
	if len(in.Cells) == 0 || in.RetainedCells > gridNativeMaxCells || in.Stamp.Rows != len(f.source.data) {
		t.Fatal("unbounded or empty stored-grid pool", in.RetainedCells, in.Stamp)
	}
	return c
}
func replayGridInput(f *replayGridFixture, typ events.Types) ManualInput {
	return ManualInput{Window: "grid", WindowIncarnation: 1, Type: typ}
}
func replayGridClick(f *replayGridFixture, bounds image.Rectangle) []ManualInput {
	point := bounds.Min.Add(bounds.Size().Div(2))
	down := replayGridInput(f, events.MouseDown)
	down.Position, down.Button = point, events.Left
	up := down
	up.Type = events.MouseUp
	return []ManualInput{down, up}
}
func replayGridKey(f *replayGridFixture, code key.Codes, r rune, modifiers key.Modifiers) ManualInput {
	in := replayGridInput(f, events.KeyChord)
	in.Code, in.Rune, in.Modifiers = code, r, modifiers
	return in
}
func authorReplayGrid(t *testing.T, f *replayGridFixture) ([]replayGridStep, []replayGridCheckpoint) {
	t.Helper()
	var record []replayGridStep
	var captures []replayGridCheckpoint
	emit := func(label string, input ...ManualInput) {
		record = append(record, replayGridStep{label: label, inputs: input})
		for _, in := range input {
			if err := f.m.Input(testContext, in); err != nil {
				t.Fatal(err)
			}
		}
		replayGridSettle(t, f, 7)
		captures = append(captures, replayGridCapture(t, f, label))
	}
	clickButton := func(id string) []ManualInput {
		s, _ := f.m.Snapshot(testContext)
		for _, e := range s.Windows[0].Elements {
			if e.ID == id && e.Enabled && e.Visible && !e.Bounds.Empty() {
				return replayGridClick(f, e.Bounds)
			}
		}
		t.Fatal("pilot button absent", id)
		return nil
	}
	cell := func() GridCellGeometry {
		in := gridInspection(t, f.a, f.g)
		for _, c := range in.Cells {
			if c.Column == 1 && c.Bounds.Dy() == 24 {
				return c
			}
		}
		t.Fatal("pilot fully visible field absent")
		return GridCellGeometry{}
	}
	emit("initial")
	emit("selected", replayGridClick(f, cell().Bounds)...)
	emit("range", replayGridKey(f, key.CodeDownArrow, 0, gridSelectionModifiers(key.Shift)))
	wheel := replayGridInput(f, events.Scroll)
	wheel.Position = cell().Bounds.Min.Add(image.Pt(4, 4))
	wheel.Delta = math32.Vec2(180, 360)
	emit("scrolled", wheel)
	emit("reversed", clickButton("reverse")...)
	emit("deep-selected", replayGridClick(f, cell().Bounds)...)
	emit("draft", replayGridKey(f, key.CodeF2, 0, 0))
	emit("committed", replayGridKey(f, key.CodeUnknown, '4', 0), replayGridKey(f, key.CodeUnknown, '2', 0), replayGridKey(f, key.CodeReturnEnter, 0, 0))
	emit("undone", clickButton("undo")...)
	emit("redone", clickButton("redo")...)
	return record, captures
}

func TestStoredGridRoutedReplayAcrossBatches(t *testing.T) {
	for _, rows := range []int{10_000, 1_000_000} {
		t.Run(strconv.Itoa(rows), func(t *testing.T) {
			var record []replayGridStep
			var expected []replayGridCheckpoint
			t.Run("pilot", func(t *testing.T) { f := newReplayGrid(t, rows, 7); record, expected = authorReplayGrid(t, f) })
			if len(expected) != 10 {
				t.Fatal("pilot checkpoint count", len(expected))
			}
			if expected[2].Selection.CapturedRows != 2 || expected[2].Active.RowKey != "record-1" || expected[3].Viewport.ScrollX == 0 || expected[3].Viewport.ScrollY == 0 || expected[4].Source.Revision != 2 || expected[4].Viewport.ScrollY < (rows-100)*24 || expected[6].Editor == nil || !expected[6].Editor.Focused || !expected[6].Edit.Active || expected[7].Edit.Active || expected[7].History.UndoCount != 1 || expected[7].Commits != 1 || expected[8].History.RedoCount != 1 || expected[9].History.UndoCount != 1 {
				t.Fatal("pilot did not exercise range/scroll/reverse/editor/history transitions")
			}
			originalRow, err := strconv.Atoi(strings.TrimPrefix(expected[6].Edit.RowKey, "record-"))
			originalText := strconv.Itoa(originalRow*20 + 1)
			if err != nil || originalRow < 0 || originalRow >= rows || expected[6].Edit.Original != originalText || expected[6].Edit.ColumnID != "c1" {
				t.Fatal("inline edit did not start from stored original field", expected[6].Edit)
			}
			for i, want := range []string{"42", originalText, "42"} {
				c := expected[7+i]
				if c.EditedRow != expected[6].Edit.RowKey || c.EditedColumn != "c1" || c.EditedCell.Text != want || c.EditedCell.Version != uint64(i+2) || c.Source.Revision != uint64(i+3) || c.Commits != 1 {
					t.Fatalf("checkpoint %s incorrect committed domain/history result: %s v%d rev%d commits%d", c.Label, c.EditedCell.Text, c.EditedCell.Version, c.Source.Revision, c.Commits)
				}
			}
			for _, batch := range []int{1, 7, 64} {
				t.Run(fmt.Sprintf("batch-%d", batch), func(t *testing.T) {
					f := newReplayGrid(t, rows, batch)
					var retired ManualInput
					for i, step := range record {
						for _, in := range step.inputs {
							if err := f.m.Input(testContext, in); err != nil {
								t.Fatalf("step %d %s input: %v", i, step.label, err)
							}
						}
						replayGridSettle(t, f, batch)
						got := replayGridCapture(t, f, step.label)
						if i == 0 {
							s, _ := f.m.Snapshot(testContext)
							for _, element := range s.Windows[0].Elements {
								if element.ID == GridCellID("data", "record-0", "c1") {
									retired = record[1].inputs[0]
									retired.Target, retired.TargetIncarnation = element.ID, element.Incarnation
								}
							}
							if retired.Target == "" {
								t.Fatal("initial cell missing for retirement check")
							}
						}
						if i == 3 {
							if err := f.m.Input(testContext, retired); !errors.Is(err, ErrManualStale) {
								t.Fatal("scrolled-out cell admitted captured input", err)
							}
							if after := replayGridCapture(t, f, step.label); !reflect.DeepEqual(after, got) {
								t.Fatal("refused captured input changed completed state")
							}
						}
						if !reflect.DeepEqual(got, expected[i]) {
							a, _ := json.Marshal(expected[i])
							b, _ := json.Marshal(got)
							t.Fatalf("step %d %s batch %d diverged\nwant %s\ngot  %s", i, step.label, batch, a, b)
						}
					}
					before, err := f.m.Capture(testContext, "inspector")
					if err != nil {
						t.Fatal(err)
					}
					if err := f.a.CloseWindow(fixtureWindowRef(f.a, "grid")); err != nil {
						t.Fatal(err)
					}
					replayGridSettle(t, f, batch)
					s, err := f.m.Snapshot(testContext)
					if err != nil || len(s.Windows) != 1 || s.Windows[0].ID != "inspector" || f.a.Context().Err() != nil {
						t.Fatal("survivor retirement", err, s)
					}
					if err := f.m.Input(testContext, record[1].inputs[0]); !errors.Is(err, ErrManualClosed) {
						t.Fatal("closed grid admitted old input", err)
					}
					after, err := f.m.Capture(testContext, "inspector")
					if err != nil || !reflect.DeepEqual(before, after) {
						t.Fatal("grid teardown changed survivor pixels", err)
					}
					t.Logf("rows=%d batch=%d checkpoints=%d fields=%d commits=%d", rows, batch, len(record), rows*20, f.commits)
				})
			}
		})
	}
}
