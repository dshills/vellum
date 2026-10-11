//go:build offscreen

package cogent

import (
	"errors"
	"fmt"
	"image"
	"math"
	"strconv"
	"strings"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles/states"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

// Captured selection indices own only immutable values. They do not borrow the
// source's counters, Grid, application, or mutable projection configuration.
type selectionNativeIndex struct {
	total         int
	revision      uint64
	reverse, even bool
}

func (s selectionNativeIndex) Identity() string { return "native-selection-grid" }
func (s selectionNativeIndex) Revision() uint64 { return s.revision }
func (s selectionNativeIndex) RowCount() int {
	if s.even {
		return (s.total + 1) / 2
	}
	return s.total
}
func (s selectionNativeIndex) original(row int) int {
	if s.reverse {
		row = s.RowCount() - 1 - row
	}
	if s.even {
		row *= 2
	}
	return row
}
func (s selectionNativeIndex) RowKey(row int) (string, error) {
	if row < 0 || row >= s.RowCount() {
		return "", virtual.ErrGridRange
	}
	return fmt.Sprintf("selection-%d", s.original(row)), nil
}
func (s selectionNativeIndex) IndexOf(k string) (int, bool, error) {
	if !strings.HasPrefix(k, "selection-") {
		return 0, false, nil
	}
	row, e := strconv.Atoi(strings.TrimPrefix(k, "selection-"))
	if e != nil || row < 0 || row >= s.total || s.even && row%2 != 0 {
		return 0, false, nil
	}
	if s.even {
		row /= 2
	}
	if s.reverse {
		row = s.RowCount() - 1 - row
	}
	return row, true, nil
}

type selectionNativeSource struct {
	selectionNativeIndex
	cells, keys, indices, captures, metrics int
}

func (s *selectionNativeSource) RowKey(row int) (string, error) {
	s.keys++
	return s.selectionNativeIndex.RowKey(row)
}
func (s *selectionNativeSource) IndexOf(k string) (int, bool, error) {
	s.indices++
	return s.selectionNativeIndex.IndexOf(k)
}
func (s *selectionNativeSource) Cell(k, c string) (string, error) {
	s.cells++
	return k + " · " + c, nil
}
func (s *selectionNativeSource) SelectionSnapshot() (virtual.GridRowIndex, error) {
	s.captures++
	return s.selectionNativeIndex, nil
}

type selectionNativeMetrics struct{ *selectionNativeSource }

var selectionNativeHeights = [...]int{23, 35, 47, 29}

func (s *selectionNativeSource) offset(end int) int {
	period := 4
	if s.even {
		period = 2
	}
	period = min(period, s.RowCount())
	if period == 0 {
		return 0
	}
	var cycle int
	for i := 0; i < period; i++ {
		cycle += selectionNativeHeights[s.original(i)%4]
	}
	total := end / period * cycle
	for i := 0; i < end%period; i++ {
		total += selectionNativeHeights[s.original(i)%4]
	}
	return total
}
func (s selectionNativeMetrics) RowOffset(boundary int) (int, error) {
	s.metrics++
	if boundary < 0 || boundary > s.RowCount() {
		return 0, virtual.ErrGridRange
	}
	return s.offset(boundary), nil
}
func (s selectionNativeMetrics) RowAt(pixel int) (int, error) {
	s.metrics++
	if pixel < 0 || pixel >= s.offset(s.RowCount()) {
		return 0, virtual.ErrGridRange
	}
	lo, hi := 0, s.RowCount()
	for lo < hi {
		mid := lo + (hi-lo)/2
		if s.offset(mid+1) <= pixel {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo, nil
}
func selectionNativeGrid(t *testing.T, rows int, variable bool, limit int) (*virtual.Grid, *selectionNativeSource) {
	t.Helper()
	s := &selectionNativeSource{selectionNativeIndex: selectionNativeIndex{total: rows, revision: 1}}
	var source virtual.GridSource = s
	if variable {
		source = selectionNativeMetrics{s}
	}
	cols := make([]virtual.GridColumn, 20)
	for i := range cols {
		cols[i] = virtual.GridColumn{ID: fmt.Sprintf("c%d", i), Title: fmt.Sprintf("Column %d", i), Width: 90}
	}
	g, e := virtual.NewGrid(source, cols, virtual.GridConfig{RowHeight: 24, HeaderHeight: 27, FrozenColumns: 2, OverscanRows: 1, OverscanColumns: 1, RowSelection: true, MaxSelectionKeys: limit}, virtual.GridViewport{Width: 300, Height: 180})
	if e != nil {
		t.Fatal(e)
	}
	return g, s
}
func selectionNativeSnapshot(t *testing.T, a *app.Application, g *virtual.Grid) virtual.GridSelection {
	t.Helper()
	var out virtual.GridSelection
	var e error
	if err := a.Do(func() { out, e = g.Selection() }); err != nil {
		t.Fatal(err)
	}
	if e != nil {
		t.Fatal(e)
	}
	return out
}
func selectionNativeContains(t *testing.T, s virtual.GridSelection, key string, want bool) {
	t.Helper()
	got, e := s.Contains(key)
	if e != nil || got != want {
		t.Fatalf("captured membership %s: got %v want %v error %v", key, got, want, e)
	}
}
func selectionNativeCell(t *testing.T, m GridInspection, row, col int) GridCellGeometry {
	t.Helper()
	for _, c := range m.Cells {
		if c.Row == row && c.Column == col && !c.Bounds.Empty() {
			return c
		}
	}
	t.Fatalf("visible grid cell absent row%d col%d", row, col)
	return GridCellGeometry{}
}
func selectionNativeWindow(t *testing.T, m *Manual) ManualWindowSnapshot {
	t.Helper()
	s, e := m.Snapshot(testContext)
	if e != nil || len(s.Windows) != 1 {
		t.Fatal("selection fixture window", e)
	}
	return s.Windows[0]
}
func selectionNativePointer(t *testing.T, m *Manual, c GridCellGeometry, downMods, upMods key.Modifiers) {
	t.Helper()
	w := selectionNativeWindow(t, m)
	pos := c.Bounds.Min.Add(c.Bounds.Size().Div(2))
	for _, v := range []struct {
		kind events.Types
		mods key.Modifiers
	}{{events.MouseDown, downMods}, {events.MouseUp, upMods}} {
		if e := m.Input(testContext, ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: v.kind, Button: events.Left, Position: pos, Modifiers: v.mods}); e != nil {
			t.Fatal(e)
		}
	}
	gridFrame(t, m)
}
func selectionNativeKey(t *testing.T, m *Manual, code key.Codes, mods key.Modifiers) {
	t.Helper()
	w := selectionNativeWindow(t, m)
	if e := m.Input(testContext, ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: code, Modifiers: mods}); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
}
func selectionNativeProjection(t *testing.T, m *Manual, a *app.Application, g *virtual.Grid) {
	t.Helper()
	inspection := gridInspection(t, a, g)
	if !inspection.RowSelection {
		t.Fatal("inspection lost row-mode contract")
	}
	if e := a.Do(func() {
		binding := m.implementation.find("one").state.root.control.grid
		if inspection.Selection != binding.projection.State.Selection {
			t.Error("selection metadata diverges from accepted frame")
		}
		for i, c := range inspection.Cells {
			if c.Selected != binding.projection.Cells[i].Selected || binding.cells[i].StateIs(states.Selected) != c.Selected {
				t.Error("rendered row membership diverges", c)
			}
			cursor, _ := binding.cells[i].Property("vellum.grid.cursor").(bool)
			if cursor != c.Active {
				t.Error("active cursor was merged with row membership", c, cursor)
			}
		}
	}); e != nil {
		t.Fatal(e)
	}
}

func TestManagedGridRowSelectionActualPointerKeysAllAndPaging(t *testing.T) {
	for _, rows := range []int{10000, 1000000} {
		for _, variable := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/variable-%v", rows, variable), func(t *testing.T) {
				g, source := selectionNativeGrid(t, rows, variable, 0)
				changes := 0
				m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
					return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true, GridChange: func(e ui.GridEvent) {
						if e.Reason == ui.GridSelection {
							changes++
						}
					}}
				})
				gridFrame(t, m)
				first := selectionNativeCell(t, gridInspection(t, a, g), 2, 0)
				selectionNativePointer(t, m, first, 0, gridSelectionModifiers(key.Meta))
				s := selectionNativeSnapshot(t, a, g)
				if s.State().CapturedRows != 1 || changes != 1 {
					t.Fatal("plain down reinterpreted by Command release", s.State(), changes)
				}
				selectionNativeContains(t, s, first.RowKey, true)
				end := selectionNativeCell(t, gridInspection(t, a, g), 5, 3)
				selectionNativePointer(t, m, end, gridSelectionModifiers(key.Shift), 0)
				s = selectionNativeSnapshot(t, a, g)
				if s.State().CapturedRows != 4 || s.State().AnchorKey != first.RowKey || changes != 2 {
					t.Fatal("Shift range/down intent", s.State(), changes)
				}
				for row := 2; row <= 5; row++ {
					selectionNativeContains(t, s, fmt.Sprintf("selection-%d", row), true)
				}
				middle := selectionNativeCell(t, gridInspection(t, a, g), 3, 0)
				selectionNativePointer(t, m, middle, gridSelectionModifiers(key.Meta), gridSelectionModifiers(key.Shift))
				s = selectionNativeSnapshot(t, a, g)
				if s.State().CapturedRows != 3 || changes != 3 {
					t.Fatal("Command toggle/down intent", s.State(), changes)
				}
				selectionNativeContains(t, s, middle.RowKey, false)
				selectionNativeProjection(t, m, a, g)
				before := variableNativeState(t, a, g)
				selectionNativeKey(t, m, key.CodeDownArrow, gridSelectionModifiers(key.Shift))
				shifted := variableNativeState(t, a, g)
				if shifted.Active.RowKey != "selection-4" || shifted.Selection.CapturedRows != 3 || shifted.Selection.AnchorKey != first.RowKey || changes != 4 {
					t.Fatal("actual Shift Arrow", before, shifted, changes)
				}
				body := max(1, shifted.Viewport.Height-shifted.Config.HeaderHeight)
				want := 4 + max(1, body/24)
				if variable {
					metric := selectionNativeMetrics{source}
					var e error
					want, e = metric.RowAt(min(source.offset(rows)-1, source.offset(4)+body))
					if e != nil {
						t.Fatal(e)
					}
				}
				selectionNativeKey(t, m, key.CodePageDown, gridSelectionModifiers(key.Shift))
				paged := variableNativeState(t, a, g)
				if paged.Active.RowKey != fmt.Sprintf("selection-%d", want) || paged.Selection.CapturedRows != want-2+1 || changes != 5 {
					t.Fatal("actual Shift Page did not extend pixel/fixed interval", paged, want, changes)
				}
				selectionNativeKey(t, m, key.CodeA, gridSelectionModifiers(key.Meta))
				all := selectionNativeSnapshot(t, a, g)
				if all.State().CapturedRows != rows || changes != 6 {
					t.Fatal("actual Command A", all.State(), changes)
				}
				selectionNativeContains(t, all, fmt.Sprintf("selection-%d", rows-1), true)
				inspected := gridInspection(t, a, g)
				var visible GridCellGeometry
				for _, c := range inspected.Cells {
					if !c.Bounds.Empty() {
						visible = c
						break
					}
				}
				selectionNativePointer(t, m, visible, gridSelectionModifiers(key.Meta), 0)
				toggled := selectionNativeSnapshot(t, a, g)
				if toggled.State().CapturedRows != rows-1 || changes != 7 {
					t.Fatal("All then Toggle count", toggled.State(), changes)
				}
				selectionNativeContains(t, toggled, visible.RowKey, false)
				selectionNativeContains(t, all, visible.RowKey, true)
				selectionNativeProjection(t, m, a, g)
				selectionNativeKey(t, m, key.CodeEscape, 0)
				empty := selectionNativeSnapshot(t, a, g)
				if empty.State().CapturedRows != 0 || changes != 8 {
					t.Fatal("Escape clear", empty.State(), changes)
				}
				if source.keys > 20000 || source.cells > gridNativeMaxCells*32 {
					t.Fatal("native selection enumerated domain", source.keys, source.cells)
				}
			})
		}
	}
}

func TestManagedGridRowSelectionSourceReplaceAndStalePress(t *testing.T) {
	g, source := selectionNativeGrid(t, 1000000, true, 0)
	changes := 0
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true, GridChange: func(e ui.GridEvent) {
			if e.Reason == ui.GridSelection {
				changes++
			}
		}}
	})
	gridFrame(t, m)
	selectionNativePointer(t, m, selectionNativeCell(t, gridInspection(t, a, g), 2, 0), 0, 0)
	selectionNativePointer(t, m, selectionNativeCell(t, gridInspection(t, a, g), 5, 3), gridSelectionModifiers(key.Shift), 0)
	captured := selectionNativeSnapshot(t, a, g)
	old := gridInspection(t, a, g)
	cell := selectionNativeCell(t, old, 3, 0)
	w := selectionNativeWindow(t, m)
	pos := cell.Bounds.Min.Add(cell.Bounds.Size().Div(2))
	if e := m.Input(testContext, ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: events.MouseDown, Button: events.Left, Position: pos, Modifiers: gridSelectionModifiers(key.Shift)}); e != nil {
		t.Fatal(e)
	}
	if _, e := m.Step(testContext, 1); e != nil {
		t.Fatal(e)
	}
	reversed := &selectionNativeSource{selectionNativeIndex: selectionNativeIndex{total: source.total, revision: 2, reverse: true}}
	var replaceErr error
	if e := a.Do(func() {
		replaceErr = g.ReplaceSource(selectionNativeMetrics{reversed})
		if replaceErr == nil {
			replaceErr = a.Invalidate("one")
		}
	}); e != nil {
		t.Fatal(e)
	}
	if replaceErr != nil {
		t.Fatal(replaceErr)
	}
	gridFrame(t, m)
	if e := m.Input(testContext, ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: events.MouseUp, Button: events.Left, Position: pos}); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
	reversedState := variableNativeState(t, a, g)
	if changes != 2 || reversedState.Active.RowKey != "selection-5" || reversedState.Selection != captured.State() {
		t.Fatal("revised projection reinterpreted stale press or selection", changes, reversedState, captured.State())
	}
	// Filtering changes the live row count/order, not the already captured set.
	filtered := &selectionNativeSource{selectionNativeIndex: selectionNativeIndex{total: source.total, revision: 3, reverse: true, even: true}}
	if e := a.Do(func() {
		replaceErr = g.ReplaceSource(selectionNativeMetrics{filtered})
		if replaceErr == nil {
			replaceErr = a.Invalidate("one")
		}
	}); e != nil {
		t.Fatal(e)
	}
	if replaceErr != nil {
		t.Fatal(replaceErr)
	}
	gridFrame(t, m)
	after := selectionNativeSnapshot(t, a, g)
	if after.State() != captured.State() {
		t.Fatal("filter changed immutable captured scope", after.State(), captured.State())
	}
	for row := 2; row <= 5; row++ {
		selectionNativeContains(t, after, fmt.Sprintf("selection-%d", row), true)
	}
	inspected := gridInspection(t, a, g)
	if inspected.Stamp.Rows != 500000 || inspected.Selection.CapturedRows != 4 {
		t.Fatal("captured count confused with live filtered rows", inspected)
	}
	for _, c := range inspected.Cells {
		if !c.Bounds.Empty() && (c.RowKey == "selection-2" || c.RowKey == "selection-4") && !c.Selected {
			t.Fatal("retained visible row membership omitted", c)
		}
	}
	// Captured source state survives retiring its application/window binding.
	if e := m.Close(testContext); e != nil {
		t.Fatal(e)
	}
	selectionNativeContains(t, captured, "selection-3", true)
	var inspectErr error
	if e := a.Do(func() { _, inspectErr = InspectGrid(a, g) }); e != nil && !errors.Is(e, app.ErrClosed) {
		t.Fatal(e)
	}
	if inspectErr != nil && !errors.Is(inspectErr, ErrGridClosed) {
		t.Fatal(inspectErr)
	}
}

func TestManagedGridRowSelectionRefusalAndEligibility(t *testing.T) {
	g, _ := selectionNativeGrid(t, 10000, false, 1)
	changes := 0
	disabled := false
	modal := false
	dismisses := 0
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		children := []ui.Node{{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true, Disabled: disabled, GridChange: func(e ui.GridEvent) {
			if e.Reason == ui.GridSelection {
				changes++
			}
		}}}
		if modal {
			children = append(children, ui.Node{ID: "modal", Kind: ui.Dialog, Open: true, Dismiss: func() { dismisses++; modal = false }, Children: []ui.Node{{ID: "modal-dismiss", Kind: ui.Button, Text: "Close", Activate: func() { dismisses++; modal = false }}}})
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: children}
	})
	gridFrame(t, m)
	selectionNativePointer(t, m, selectionNativeCell(t, gridInspection(t, a, g), 2, 0), 0, 0)
	selected := selectionNativeSnapshot(t, a, g)
	// Two sparse exclusions exceed MaxSelectionKeys=1, but All remains an O(1)
	// immutable interval. Capacity refusal does not lose the first exclusion.
	selectionNativeKey(t, m, key.CodeA, gridSelectionModifiers(key.Meta))
	selectionNativePointer(t, m, selectionNativeCell(t, gridInspection(t, a, g), 2, 0), gridSelectionModifiers(key.Meta), 0)
	before := variableNativeState(t, a, g)
	beforeChanges := changes
	for _, mods := range []key.Modifiers{gridSelectionModifiers(key.Control), gridSelectionModifiers(key.Alt), gridSelectionModifiers(key.Shift, key.Meta)} {
		selectionNativePointer(t, m, selectionNativeCell(t, gridInspection(t, a, g), 4, 3), mods, 0)
		selectionNativeKey(t, m, key.CodeDownArrow, mods)
		if got := variableNativeState(t, a, g); got != before || changes != beforeChanges {
			t.Fatal("unsupported semantic modifiers changed selection", mods, got)
		}
	}
	if e := a.Do(func() { disabled = true; _ = a.Invalidate("one") }); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
	inspection := gridInspection(t, a, g)
	c := selectionNativeCell(t, inspection, 4, 3)
	w := selectionNativeWindow(t, m)
	pos := c.Bounds.Min.Add(c.Bounds.Size().Div(2))
	if e := m.Input(testContext, ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: events.MouseDown, Button: events.Left, Position: pos, Modifiers: gridSelectionModifiers(key.Meta)}); e != nil {
		t.Fatal(e)
	}
	if e := m.Input(testContext, ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: events.MouseUp, Button: events.Left, Position: pos}); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
	if got := variableNativeState(t, a, g); got.Selection != before.Selection || changes != beforeChanges {
		t.Fatal("disabled input changed row selection", got)
	}
	if e := a.Do(func() { disabled = false; modal = true; _ = a.Invalidate("one") }); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
	selectionNativeKey(t, m, key.CodeEscape, 0)
	if dismisses != 1 || modal || variableNativeState(t, a, g).Selection != before.Selection || changes != beforeChanges {
		t.Fatal("modal Escape lost precedence", dismisses, modal, changes)
	}
	selectionNativeContains(t, selected, "selection-2", true)
	// The public Manual error sink deliberately retains delivered control errors.
	// Check the actual refused primitive event at the end instead of replacing
	// a sink that ordinary reconciliation restores during input/focus refresh.
	c = selectionNativeCell(t, gridInspection(t, a, g), 3, 3)
	w = selectionNativeWindow(t, m)
	pos = c.Bounds.Min.Add(c.Bounds.Size().Div(2))
	if e := m.Input(testContext, ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: events.MouseDown, Button: events.Left, Position: pos, Modifiers: gridSelectionModifiers(key.Meta)}); e != nil {
		t.Fatal(e)
	}
	if e := m.Input(testContext, ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: events.MouseUp, Button: events.Left, Position: pos}); e != nil {
		t.Fatal(e)
	}
	if _, e := m.Step(testContext, 256); !errors.Is(e, virtual.ErrGridCapacity) {
		t.Fatal("missing delivered selection capacity refusal", e)
	}
	if after := variableNativeState(t, a, g); before != after || changes != beforeChanges {
		t.Fatal("selection capacity refusal not atomic", before, after, changes)
	}
	if e := m.Close(testContext); !errors.Is(e, virtual.ErrGridCapacity) {
		t.Fatal("manual error disposition lost", e)
	}
}

func TestManagedGridLegacySelectionKeepsActiveCellBehavior(t *testing.T) {
	g, _ := testNativeGrid(t, 10000)
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true} })
	gridFrame(t, m)
	cell := selectionNativeCell(t, gridInspection(t, a, g), 2, 0)
	selectionNativePointer(t, m, cell, gridSelectionModifiers(key.Shift), 0)
	inspected := gridInspection(t, a, g)
	if inspected.RowSelection {
		t.Fatal("legacy grid silently opted into row mode")
	}
	active := 0
	for _, c := range inspected.Cells {
		if c.Selected {
			t.Fatal("legacy row membership fabricated", c)
		}
		if c.Active {
			active++
		}
	}
	if active != 1 {
		t.Fatal("legacy cursor changed", active)
	}
	var selectionErr error
	if e := a.Do(func() { _, selectionErr = g.Selection() }); e != nil {
		t.Fatal(e)
	}
	if !errors.Is(selectionErr, virtual.ErrGridSelection) {
		t.Fatal("legacy selection API unexpectedly active", selectionErr)
	}
}

func TestManagedGridRowSelectionQueuedIntentNativeTabAndWindowClose(t *testing.T) {
	t.Run("native-tab", func(t *testing.T) {
		g, _ := selectionNativeGrid(t, 10000, false, 0)
		changes := 0
		m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
			return ui.Node{ID: "tabs", Kind: ui.Tabs, SelectedID: "one-page", Options: []ui.Option{{ID: "one-page", Label: "One"}, {ID: "two-page", Label: "Two"}}, Children: []ui.Node{
				{ID: "one-page", Kind: ui.Column, Children: []ui.Node{{ID: "g", Kind: ui.DataGrid, Grid: g, GridChange: func(e ui.GridEvent) {
					if e.Reason == ui.GridSelection {
						changes++
					}
				}, Style: ui.Style{Grow: ui.Some(float32(0)), MinSize: ui.Some(ui.Size{Width: 400, Height: 240}), MaxSize: ui.Some(ui.Size{Width: 400, Height: 240})}}}},
				{ID: "two-page", Kind: ui.Column, Children: []ui.Node{{ID: "other", Kind: ui.Label, Text: "Other"}}},
			}}
		})
		gridFrame(t, m)
		inspected := gridInspection(t, a, g)
		visible := selectionNativeCell(t, inspected, 2, 0)
		before := variableNativeState(t, a, g)
		var pressAccepted bool
		var handlerMissing bool
		if e := a.Do(func() {
			root := m.implementation.find("one").state.root
			_, handlers := accessibilitySnapshot(root, nil)
			press := handlers[accessibilityManagedID(visible.ID)]
			handlerMissing = press == nil
			binding := root.children[0].children[0].control.grid
			binding.post(ui.GridSelection, g.SelectAll)
			root.widget.(*core.Tabs).SelectTabByName("two-page")
			// Preserve retained descriptor scope. The queued and AX paths must check
			// actual native tab ownership before the controlled Change refresh.
			if press != nil {
				pressAccepted = press(accessibilityAction{Kind: "press"})
			}
		}); e != nil {
			t.Fatal(e)
		}
		if handlerMissing || pressAccepted {
			t.Fatal("inactive native tab accepted retained AX selection", handlerMissing, pressAccepted)
		}
		if e := a.Drain(); e != nil {
			t.Fatal(e)
		}
		if got := variableNativeState(t, a, g); got != before || changes != 0 {
			t.Fatal("queued All crossed native tab lifetime", got, before, changes)
		}
	})
	t.Run("logical-close", func(t *testing.T) {
		g, _ := selectionNativeGrid(t, 10000, false, 0)
		changes := 0
		m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
			return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true, GridChange: func(e ui.GridEvent) {
				if e.Reason == ui.GridSelection {
					changes++
				}
			}}
		})
		gridFrame(t, m)
		if e := a.Open(app.Window{ID: "survivor", Title: "Survivor", Width: 400, Height: 300, View: func() ui.Node { return ui.Node{ID: "other", Kind: ui.Label, Text: "Other"} }}); e != nil {
			t.Fatal(e)
		}
		before := variableNativeState(t, a, g)
		var closeErr error
		if e := a.Do(func() {
			binding := m.implementation.find("one").state.root.control.grid
			binding.post(ui.GridSelection, g.SelectAll)
			closeErr = a.CloseWindow("one")
		}); e != nil {
			t.Fatal(e)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
		if e := a.Drain(); e != nil {
			t.Fatal(e)
		}
		if got := variableNativeState(t, a, g); got != before || changes != 0 {
			t.Fatal("queued All mutated logically closed originating window", got, before, changes)
		}
	})
}

func TestManagedGridRowSelectionOnlyTransitionRetiresPressAndAX(t *testing.T) {
	g, _ := selectionNativeGrid(t, 10000, false, 0)
	changes := 0
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true, GridChange: func(e ui.GridEvent) {
			if e.Reason == ui.GridSelection {
				changes++
			}
		}}
	})
	gridFrame(t, m)
	selectionNativePointer(t, m, selectionNativeCell(t, gridInspection(t, a, g), 2, 0), 0, 0)
	before := variableNativeState(t, a, g)
	cell := selectionNativeCell(t, gridInspection(t, a, g), 3, 0)
	var oldAction accessibilityHandler
	if e := a.Do(func() {
		_, handlers := accessibilitySnapshot(m.implementation.find("one").state.root, nil)
		oldAction = handlers[accessibilityManagedID(cell.ID)]
	}); e != nil {
		t.Fatal(e)
	}
	w := selectionNativeWindow(t, m)
	pos := cell.Bounds.Min.Add(cell.Bounds.Size().Div(2))
	if e := m.Input(testContext, ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: events.MouseDown, Button: events.Left, Position: pos}); e != nil {
		t.Fatal(e)
	}
	if _, e := m.Step(testContext, 1); e != nil {
		t.Fatal(e)
	}
	var changeErr error
	if e := a.Do(func() {
		changeErr = g.SelectToKey(before.Active.RowKey, before.Active.ColumnID, virtual.GridSelectToggle)
		if changeErr == nil {
			changeErr = a.Invalidate("one")
		}
	}); e != nil {
		t.Fatal(e)
	}
	if changeErr != nil {
		t.Fatal(changeErr)
	}
	gridFrame(t, m)
	after := variableNativeState(t, a, g)
	if after.Active != before.Active || after.Viewport != before.Viewport || after.Generation != before.Generation+1 || after.Selection.CapturedRows != 0 {
		t.Fatal("selection-only transition did not retire its projection once", before, after)
	}
	if e := m.Input(testContext, ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: events.MouseUp, Button: events.Left, Position: pos}); e != nil {
		t.Fatal(e)
	}
	gridFrame(t, m)
	if got := variableNativeState(t, a, g); got != after || changes != 1 {
		t.Fatal("old press reused a selection-only generation", got, changes)
	}
	var accepted bool
	var refused error
	if e := a.Do(func() {
		m.implementation.find("one").state.root.control.onError = func(e error) { refused = e }
		if oldAction != nil {
			accepted = oldAction(accessibilityAction{Kind: "press"})
		}
	}); e != nil {
		t.Fatal(e)
	}
	if oldAction == nil || accepted || !errors.Is(refused, virtual.ErrGridStale) || changes != 1 {
		t.Fatal("old AX press reused a selection-only generation", accepted, refused, changes)
	}
}

// The first indexed row is shorter than a cursor border. Its exterior still
// belongs to the portable shared-edge geometry; only inner text may clamp.
type selectionNativeThinMetrics struct{ *selectionNativeSource }

func (s selectionNativeThinMetrics) RowOffset(boundary int) (int, error) {
	if boundary < 0 || boundary > s.RowCount() {
		return 0, virtual.ErrGridRange
	}
	if boundary == 0 {
		return 0, nil
	}
	return 1 + (boundary-1)*23, nil
}
func (s selectionNativeThinMetrics) RowAt(pixel int) (int, error) {
	total, _ := s.RowOffset(s.RowCount())
	if pixel < 0 || pixel >= total {
		return 0, virtual.ErrGridRange
	}
	if pixel == 0 {
		return 0, nil
	}
	return 1 + (pixel-1)/23, nil
}

func TestManagedGridRowSelectionCursorBorderInsideIndexedGeometry(t *testing.T) {
	for _, dpi := range []float32{96, 120, 144, 192} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			source := &selectionNativeSource{selectionNativeIndex: selectionNativeIndex{total: 8, revision: 1}}
			g, err := virtual.NewGrid(selectionNativeThinMetrics{source}, []virtual.GridColumn{{ID: "c0", Width: 121}, {ID: "c1", Width: 119}, {ID: "c2", Width: 123}}, virtual.GridConfig{RowHeight: 23, HeaderHeight: 27, FrozenColumns: 1, OverscanRows: 1, RowSelection: true}, virtual.GridViewport{Width: 300, Height: 180})
			if err != nil {
				t.Fatal(err)
			}
			config := manualConfig(t, 7)
			config.LogicalDPI = dpi
			m, a := gridManual(t, config, func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true} })
			gridFrame(t, m)
			check := func() {
				t.Helper()
				measurement := gridInspection(t, a, g)
				var projection virtual.GridSnapshot
				var stateErr error
				if err := a.Do(func() { projection, stateErr = g.Snapshot() }); err != nil {
					t.Fatal(err)
				}
				if stateErr != nil {
					t.Fatal(stateErr)
				}
				if len(measurement.Cells) != len(projection.Cells) {
					t.Fatal("measurement cardinality", len(measurement.Cells), len(projection.Cells))
				}
				edge := func(v int) int { return int(math.Round(float64(v) * float64(measurement.Scale))) }
				for i, cell := range projection.Cells {
					want := image.Rectangle{}
					if !cell.Clip.Empty() {
						want = image.Rect(edge(cell.Clip.X), edge(cell.Clip.Y), edge(cell.Clip.X+cell.Clip.Width), edge(cell.Clip.Y+cell.Clip.Height)).Add(measurement.Bounds.Min)
					}
					if got := measurement.Cells[i].Bounds; got != want {
						t.Fatalf("cell border changed indexed shared edges: cell=%+v got=%v want=%v", cell, got, want)
					}
				}
				if err := a.Do(func() {
					binding := m.implementation.find("one").state.root.control.grid
					for _, cell := range binding.cells {
						if cell.Geom.Size.Actual.Content.X < 0 || cell.Geom.Size.Actual.Content.Y < 0 || cell.Geom.Size.Alloc.Content.X < 0 || cell.Geom.Size.Alloc.Content.Y < 0 {
							t.Error("short indexed row created negative inner text geometry", cell.Geom.Size)
						}
					}
				}); err != nil {
					t.Fatal(err)
				}
			}
			check()
			// Select the one-pixel indexed row, then an odd-height adjacent row.
			// Both cursor and ordinary selected peers must retain exact exteriors.
			selectionNativePointer(t, m, selectionNativeCell(t, gridInspection(t, a, g), 0, 0), 0, 0)
			check()
			selectionNativePointer(t, m, selectionNativeCell(t, gridInspection(t, a, g), 1, 0), gridSelectionModifiers(key.Shift), 0)
			check()
			selectionNativeProjection(t, m, a, g)
		})
	}
}
