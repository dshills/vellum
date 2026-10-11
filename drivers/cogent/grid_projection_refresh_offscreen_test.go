//go:build offscreen

package cogent

import (
	"crypto/sha256"
	"fmt"
	"html"
	"image/color"
	"sync/atomic"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/styles/states"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

// Domain values stay immutable; this separate diagnostic counts materialized
// cells, not metadata validation or state reads.
type refreshGridSource struct {
	selectionNativeIndex
	calls  *atomic.Int64
	prefix string
}

func (s refreshGridSource) Cell(row, col string) (string, error) {
	s.calls.Add(1)
	return s.prefix + row + " · " + col, nil
}
func (s refreshGridSource) SelectionSnapshot() (virtual.GridRowIndex, error) {
	return s.selectionNativeIndex, nil
}
func refreshGridModel(t *testing.T) (*virtual.Grid, refreshGridSource) {
	t.Helper()
	s := refreshGridSource{selectionNativeIndex: selectionNativeIndex{total: 10_000, revision: 1}, calls: new(atomic.Int64)}
	columns := make([]virtual.GridColumn, 20)
	for i := range columns {
		columns[i] = virtual.GridColumn{ID: fmt.Sprint("c", i), Title: fmt.Sprint("Column ", i), Width: 90}
	}
	g, err := virtual.NewGrid(s, columns, virtual.GridConfig{RowHeight: 24, HeaderHeight: 26, FrozenColumns: 2, OverscanRows: 1, OverscanColumns: 1, RowSelection: true}, virtual.GridViewport{Width: 300, Height: 180})
	if err != nil {
		t.Fatal(err)
	}
	return g, s
}

type refreshGridCounts struct {
	host  gridUpdatePass
	slots map[*core.Text]*gridUpdatePass
}

func watchRefreshGrid(g *gridBinding) *refreshGridCounts {
	p := &refreshGridCounts{slots: map[*core.Text]*gridUpdatePass{}}
	g.host.Updater(func() { p.host.updates++ })
	g.host.FinalStyler(func(*styles.Style) { p.host.styles++ })
	p.observeCurrentSlots(g)
	return p
}

// A changed projection may legitimately grow or shrink the pool. Establish its
// accepted retained set before measuring only the following unchanged tail.
func (p *refreshGridCounts) observeCurrentSlots(g *gridBinding) {
	current := make(map[*core.Text]*gridUpdatePass, len(g.cells)+len(g.headers))
	for _, w := range append(append([]*core.Text(nil), g.cells...), g.headers...) {
		count := p.slots[w]
		if count == nil {
			count = new(gridUpdatePass)
			w.Updater(func() { count.updates++ })
			w.FinalStyler(func(*styles.Style) { count.styles++ })
		}
		current[w] = count
	}
	p.slots = current
}
func (p *refreshGridCounts) reset() {
	p.host = gridUpdatePass{}
	for _, slot := range p.slots {
		*slot = gridUpdatePass{}
	}
}
func (p *refreshGridCounts) assertZero(t *testing.T, label string, g *gridBinding) {
	t.Helper()
	if p.host != (gridUpdatePass{}) {
		t.Errorf("%s unnecessary retained host update/style: %+v", label, p.host)
	}
	if len(g.cells)+len(g.headers) != len(p.slots) {
		t.Errorf("%s unchanged projection resized native pool", label)
	}
	for _, w := range append(append([]*core.Text(nil), g.cells...), g.headers...) {
		pass, ok := p.slots[w]
		if !ok || *pass != (gridUpdatePass{}) {
			t.Errorf("%s unnecessary retained cell/header update/style: present=%t passes=%+v", label, ok, pass)
		}
	}
}

// This drains only eligible originating events and owner callbacks. It does not
// force a frame, advance time, or clear Manual's conservative layoutPending.
func refreshGridDrain(t *testing.T, m *Manual) {
	t.Helper()
	for range 64 {
		result, err := m.Step(testContext, 64)
		if err != nil {
			t.Fatal(err)
		}
		if result.Events+result.OwnerCallbacks == 0 {
			return
		}
	}
	t.Fatal("bounded native/owner refresh drain exhausted")
}

func TestManagedGridUnchangedProjectionRefreshDoesNotMaterializeOrUpdate(t *testing.T) {
	for _, dpi := range []float32{96, 144} {
		t.Run(fmt.Sprint(dpi), func(t *testing.T) {
			model, source := refreshGridModel(t)
			config := manualConfig(t, 7)
			config.LogicalDPI = dpi
			m, a := gridManual(t, config, func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true} })
			gridFrame(t, m)
			binding := gBinding(t, model)
			var counts *refreshGridCounts
			if err := a.Do(func() { counts = watchRefreshGrid(binding) }); err != nil {
				t.Fatal(err)
			}
			before := gridInspection(t, a, model)
			raster, err := m.Capture(testContext, "one")
			if err != nil {
				t.Fatal(err)
			}
			digest := sha256.Sum256(raster.Pix)
			if err := a.Do(func() {
				counts.reset()
				source.calls.Store(0)
				if err := a.Invalidate("one"); err != nil {
					t.Error(err)
				}
			}); err != nil {
				t.Fatal(err)
			}
			refreshGridDrain(t, m)
			if err := a.Do(func() {
				counts.assertZero(t, "same descriptor", binding)
				if got := source.calls.Load(); got != 0 {
					t.Errorf("same descriptor unnecessarily fetched %d source cells", got)
				}
				current, err := InspectGrid(a, model)
				if err != nil || current.Generation != before.Generation || current.CellIncarnation != before.CellIncarnation || current.Bounds != before.Bounds || current.Frames != before.Frames {
					t.Errorf("same descriptor changed measurement: generation=%d/%d incarnation=%d/%d bounds=%v/%v frames=%d/%d error=%v", before.Generation, current.Generation, before.CellIncarnation, current.CellIncarnation, before.Bounds, current.Bounds, before.Frames, current.Frames, err)
				}
			}); err != nil {
				t.Fatal(err)
			}
			// The underlying completed software image remains unchanged even though
			// public Manual.Capture conservatively refuses layoutPending after refresh.
			retained, err := m.implementation.find("one").core.Image(testContext)
			if err != nil || sha256.Sum256(retained.Pix) != digest {
				t.Error("no-op refresh changed completed software pixels", err)
			}

			// A legitimate changed model is first rendered. Capture then invalidates
			// the owner to republish that measurement; isolate that tail immediately.
			if err := a.Do(func() {
				if err := model.Scroll(90, 24); err != nil {
					t.Error(err)
				}
				if err := model.SelectToKey("selection-2", "c3", virtual.GridSelectReplace); err != nil {
					t.Error(err)
				}
				_ = a.Invalidate("one")
			}); err != nil {
				t.Fatal(err)
			}
			captured := false
			for range 64 {
				if _, err := m.Frame(testContext, "one"); err != nil {
					t.Fatal(err)
				}
				if err := a.Do(func() {
					state, err := model.State()
					if err != nil {
						t.Error(err)
						return
					}
					in, err := InspectGrid(a, model)
					captured = err == nil && in.Generation == state.Generation && in.Frames > before.Frames
				}); err != nil {
					t.Fatal(err)
				}
				if captured {
					break
				}
			}
			if !captured || a.Pending().DirtyWindows == 0 {
				t.Fatal("fixture did not isolate a real completed capture invalidation")
			}
			before = gridInspection(t, a, model)
			if err := a.Do(func() { counts.observeCurrentSlots(binding); counts.reset(); source.calls.Store(0) }); err != nil {
				t.Fatal(err)
			}
			refreshGridDrain(t, m)
			if err := a.Do(func() {
				counts.assertZero(t, "capture invalidation tail", binding)
				if got := source.calls.Load(); got != 0 {
					t.Errorf("capture invalidation tail unnecessarily fetched %d source cells", got)
				}
				current, err := InspectGrid(a, model)
				if err != nil || current.Generation != before.Generation || current.Frames != before.Frames {
					t.Errorf("capture invalidation tail changed projection/frame: %v", err)
				}
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagedGridProjectionRefreshKeepsModelDirtyChanges(t *testing.T) {
	model, source := refreshGridModel(t)
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true} })
	gridFrame(t, m)
	binding := gBinding(t, model)
	for _, phase := range []struct {
		name   string
		change func() error
	}{
		{"source", func() error { source.revision++; source.prefix = "<new & source>"; return model.ReplaceSource(source) }},
		{"selection", func() error { return model.SelectToKey("selection-2", "c2", virtual.GridSelectReplace) }},
		{"viewport", func() error { return model.Scroll(90, 24) }},
		{"cursor-only", func() error { return model.Move(1, 1) }},
	} {
		before := gridInspection(t, a, model)
		if err := a.Do(func() {
			source.calls.Store(0)
			if err := phase.change(); err != nil {
				t.Error(err)
			}
			_ = a.Invalidate("one")
		}); err != nil {
			t.Fatal(err)
		}
		gridFrame(t, m)
		current := gridInspection(t, a, model)
		if source.calls.Load() == 0 || current.Generation <= before.Generation || current.Cells[0].Text != source.prefix+current.Cells[0].RowKey+" · "+current.Cells[0].ColumnID {
			t.Errorf("%s skipped legitimate changed projection", phase.name)
		}
		if err := a.Do(func() { assertGridUpdateStyles(t, binding, phase.name) }); err != nil {
			t.Fatal(err)
		}
	}
}

func TestManagedGridProjectionRefreshRetainsInheritedStyleDisabledAndTopology(t *testing.T) {
	model, source := refreshGridModel(t)
	config := manualConfig(t, 7)
	config.LogicalDPI = 144
	fontSize := float32(14)
	foreground := color.NRGBA{R: 30, G: 60, B: 90, A: 255}
	disabled, moved, resized := false, false, false
	nodeTypography := false
	help := ""
	view := func() ui.Node {
		grid := ui.Node{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true, Accessibility: ui.Accessibility{Help: help}}
		if nodeTypography {
			grid.Style.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 18})
		}
		if resized {
			grid.Style.MaxSize = ui.Some(ui.Size{Width: 320, Height: 200})
		}
		left := ui.Node{ID: "left", Kind: ui.Column, Grow: true, Disabled: disabled, Style: ui.Style{Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: fontSize}), Foreground: ui.Some(foreground)}}
		right := ui.Node{ID: "right", Kind: ui.Column, Grow: true, Disabled: disabled, Style: left.Style}
		if moved {
			right.Children = []ui.Node{grid}
		} else {
			left.Children = []ui.Node{grid}
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{left, right}}
	}
	m, a := gridManual(t, config, view)
	gridFrame(t, m)
	binding := gBinding(t, model)
	originalHost := binding.host
	originalCell := binding.cells[0]
	for _, phase := range []struct {
		name   string
		change func()
	}{
		{"inherited typography", func() { fontSize = 20; foreground = color.NRGBA{R: 91, G: 23, B: 47, A: 255} }},
		{"node typography", func() { nodeTypography = true }},
		{"theme", func() {
			theme := m.implementation.config.Theme.Clone()
			theme.Kinds = map[ui.Kind]ui.Style{ui.DataGrid: {Background: ui.Some(color.NRGBA{R: 220, G: 230, B: 240, A: 255})}}
			m.implementation.config.Theme = theme
		}},
		{"help added", func() { help = "Use <Grid> & rows" }},
		{"help removed", func() { help = "" }},
		{"ancestor disabled", func() { disabled = true }},
		{"ancestor enabled", func() { disabled = false }},
		{"reparent", func() { moved = true }},
		{"resize", func() { resized = true }},
	} {
		before := gridInspection(t, a, model)
		if err := a.Do(func() { source.calls.Store(0); phase.change(); _ = a.Invalidate("one") }); err != nil {
			t.Fatal(err)
		}
		gridFrame(t, m)
		current := gridInspection(t, a, model)
		if phase.name != "reparent" && phase.name != "resize" && current.Generation != before.Generation {
			t.Error(phase.name, "fixture must exercise native-only synchronization at unchanged model generation")
		}
		if err := a.Do(func() {
			if binding.host != originalHost || binding.cells[0] != originalCell {
				t.Error(phase.name, "replaced retained native identity")
			}
			// Manual requires explicit defaults. The managed grid therefore resolves
			// its own defaults/node theme, while pooled texts inherit that host.
			resolved, err := ui.ResolveStyle(binding.r.node, binding.r.theme)
			if err != nil {
				t.Error(err)
				return
			}
			for _, cell := range append(append([]*core.Text(nil), binding.cells...), binding.headers...) {
				if cell.Styles.Font.Size.Value != resolved.Typography.Value.Size || cell.Styles.Color.At(0, 0) != resolved.Foreground.Value {
					t.Error(phase.name, "lost resolved host typography/color inheritance", cell.Styles.Font.Size, cell.Styles.Color.At(0, 0))
					break
				}
			}
			parentID := "left"
			if moved {
				parentID = "right"
			}
			parent := findRecord(m.implementation.find("one").state.root, parentID)
			if parent == nil || parent.widget.AsWidget().Styles.Font.Size.Value != fontSize || parent.widget.AsWidget().Styles.Color.At(0, 0) != foreground {
				t.Error(phase.name, "parent typography/color did not update")
			}
			if binding.host.Tooltip != html.EscapeString(help) {
				t.Error(phase.name, "stale or unescaped accessibility help tooltip", binding.host.Tooltip)
			}
			if binding.host.AbilityIs(abilities.LongHoverable) != (help != "") || binding.host.AbilityIs(abilities.LongPressable) != (help != "") {
				t.Error(phase.name, "stale tooltip-derived native abilities")
			}
			if binding.host.StateIs(states.Disabled) != disabled || binding.eligible() == disabled {
				t.Error(phase.name, "effective disabled state/admission stale")
			}
			if current.Generation == before.Generation && source.calls.Load() != 0 {
				t.Error(phase.name, "unchanged domain projection fetched source cells")
			}
			if phase.name == "theme" && binding.host.Styles.Background.At(0, 0) != (color.NRGBA{R: 220, G: 230, B: 240, A: 255}) {
				t.Error("unchanged projection skipped current node theme")
			}
			assertGridUpdateStyles(t, binding, phase.name)
		}); err != nil {
			t.Fatal(err)
		}
		if phase.name == "reparent" && current.Bounds == before.Bounds {
			t.Error("reparent did not move measured native host")
		}
		if phase.name == "resize" && current.Bounds.Size() == before.Bounds.Size() {
			t.Error("resize did not change completed native viewport")
		}
	}
}

func TestManagedGridProjectionRefreshKeepsPrivateSelectionRestoreAndFocus(t *testing.T) {
	m, a, model, _, _ := inlineFixture(t, 144)
	binding := gBinding(t, model)
	edit := binding.editing
	r := edit.record
	field := r.widget.(*core.TextField)
	client, session := r.control.client, edit.session
	var generation uint64
	if err := a.Do(func() {
		s, e := model.State()
		if e != nil {
			t.Error(e)
		}
		generation = s.Generation
		initial := field.EditingState()
		if initial.Anchor != 0 || initial.Focus != 4 || initial.Text != "A😀é" {
			t.Error("fixture needs the initial full forward selection", initial)
		}
	}); err != nil {
		t.Fatal(err)
	}
	// Shift+Home reverses the real selection to the logical beginning without
	// changing draft bytes; the fixture observed plain End retaining selection.
	selectionNativeKey(t, m, key.CodeHome, gridSelectionModifiers(key.Shift))
	if err := a.Do(func() {
		s, e := model.State()
		if e != nil || s.Generation != generation {
			t.Error("selection-only key changed Grid generation", e)
		}
		state := field.EditingState()
		if state.Anchor != 4 || state.Focus != 0 || state.Cursor != 0 || state.Text != "A😀é" {
			t.Error("actual selection-only key did not reverse native selection", state)
		}
		r.control.fieldSelection = &editor.Selection{Anchor: 4, Focus: 1}
		_ = a.Invalidate("one")
	}); err != nil {
		t.Fatal(err)
	}
	gridFrame(t, m)
	if err := a.Do(func() {
		state := field.EditingState()
		if state.Anchor != 4 || state.Focus != 1 {
			t.Error("same-generation queued directional selection skipped", state)
		}
		field.SetText("unaccepted native local bytes")
		requestControlRestore(r)
	}); err != nil {
		t.Fatal(err)
	}
	gridFrame(t, m)
	if err := a.Do(func() {
		s, e := model.State()
		state := field.EditingState()
		if e != nil || s.Generation != generation || state.Text != s.Edit.Text {
			t.Error("same-generation private restore skipped", state, s.Edit, e)
		}
		if binding.editing != edit || r.control.client != client || edit.session != session {
			t.Error("private refresh replaced edit/native incarnation")
		}
		if !field.StateIs(states.Focused) || binding.host.AbilityIs(abilities.Focusable) {
			t.Error("private restore lost current editor focus/Tab stop")
		}
	}); err != nil {
		t.Fatal(err)
	}
	// Return remains routed through the real editor and originating settle queue.
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: selectionNativeWindow(t, m).Incarnation, Type: events.KeyChord, Code: key.CodeReturnEnter}); err != nil {
		t.Fatal(err)
	}
	gridFrame(t, m)
}

func TestManagedGridProjectionRefreshRestoresCanonicalScrollbarAfterChange(t *testing.T) {
	model, source := refreshGridModel(t)
	disabled, modal := false, false
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{
			{ID: "g", Kind: ui.DataGrid, Grid: model, Grow: true, Disabled: disabled},
			{ID: "dialog", Kind: ui.Dialog, Text: "Modal", Open: modal, Children: []ui.Node{{ID: "modal-close", Kind: ui.Button, Text: "Close"}}},
		}}
	})
	gridFrame(t, m)
	binding := gBinding(t, model)
	if err := a.Do(func() {
		if err := model.Scroll(90, 24); err != nil {
			t.Error(err)
		}
		_ = a.Invalidate("one")
	}); err != nil {
		t.Fatal(err)
	}
	gridFrame(t, m)
	var counts *refreshGridCounts
	if err := a.Do(func() { counts = watchRefreshGrid(binding) }); err != nil {
		t.Fatal(err)
	}
	for _, phase := range []struct {
		name    string
		prepare func()
		value   func(float32, int) float32
	}{
		{"same canonical value", func() {}, func(v float32, _ int) float32 { return v }},
		{"subpixel same logical offset", func() {}, func(v float32, extent int) float32 { return v + .2/float32(extent) }},
		{"disabled admission", func() { disabled = true }, func(_ float32, _ int) float32 { return .8 }},
		{"outside modal input scope", func() { disabled = false; modal = true }, func(_ float32, _ int) float32 { return .8 }},
	} {
		if err := a.Do(func() { phase.prepare(); _ = a.Invalidate("one") }); err != nil {
			t.Fatal(err)
		}
		gridFrame(t, m)
		var before virtual.GridState
		var expected float32
		if err := a.Do(func() {
			var err error
			before, err = model.State()
			if err != nil {
				t.Error(err)
				return
			}
			extent := gridMaxX(before)
			if extent == 0 {
				t.Error("fixture requires horizontal scroll range")
				return
			}
			expected = float32(float64(before.Viewport.ScrollX) / float64(extent))
			if binding.horizontal.Value != expected {
				t.Error("fixture slider not canonical before change")
			}
			counts.reset()
			source.calls.Store(0)
			binding.horizontal.SetValue(phase.value(expected, extent))
			binding.horizontal.SendChange()
			_ = a.Invalidate("one")
		}); err != nil {
			t.Fatal(err)
		}
		refreshGridDrain(t, m)
		if err := a.Do(func() {
			current, err := model.State()
			if err != nil || current.Generation != before.Generation || current.Viewport != before.Viewport {
				t.Error(phase.name, "no-op/refused scrollbar changed accepted model", current.Viewport, err)
			}
			if binding.horizontal.Value != expected {
				t.Error(phase.name, "did not restore canonical accepted scrollbar", binding.horizontal.Value, expected)
			}
			counts.assertZero(t, phase.name, binding)
			if source.calls.Load() != 0 {
				t.Error(phase.name, "canonical scrollbar refresh fetched unchanged source cells")
			}
		}); err != nil {
			t.Fatal(err)
		}
	}
}
