//go:build offscreen

package cogent

import (
	"context"
	"errors"
	"fmt"
	"image"
	"math"
	"runtime"
	"strings"
	"testing"
	"time"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

func nativeOutline(t *testing.T, count int, expanded bool) *virtual.Outline {
	t.Helper()
	nodes := make([]virtual.TreeNode, count)
	for i := range nodes {
		parent := ""
		if i > 0 {
			parent = "node-0"
		}
		nodes[i] = virtual.TreeNode{ID: fmt.Sprintf("node-%d", i), Parent: parent, Label: fmt.Sprintf("Node %d", i)}
	}
	data, e := virtual.NewTreeData(context.Background(), "native-outline", 1, nodes, virtual.TreeLimits{})
	if e != nil {
		t.Fatal(e)
	}
	index, e := virtual.BuildTreeIndex(context.Background(), data, virtual.TreeIndexOptions{ExpandAll: expanded})
	if e != nil {
		t.Fatal(e)
	}
	o, e := virtual.NewOutline(index, virtual.OutlineConfig{RowHeight: 23, MaxRows: 500}, virtual.OutlineViewport{Width: 400, Height: 200})
	if e != nil {
		t.Fatal(e)
	}
	return o
}
func treeInspection(t *testing.T, a *app.Application, o *virtual.Outline) TreeInspection {
	t.Helper()
	var out TreeInspection
	var e error
	if err := a.Do(func() { out, e = InspectTree(a, o) }); err != nil {
		t.Fatal(err)
	}
	if e != nil {
		t.Fatal(e)
	}
	return out
}
func treeState(t *testing.T, a *app.Application, o *virtual.Outline) virtual.OutlineState {
	t.Helper()
	var s virtual.OutlineState
	var e error
	if err := a.Do(func() { s, e = o.State() }); err != nil {
		t.Fatal(err)
	}
	if e != nil {
		t.Fatal(e)
	}
	return s
}
func treeInput(t *testing.T, m *Manual, input ManualInput) error {
	t.Helper()
	s, e := m.Snapshot(testContext)
	if e != nil {
		return e
	}
	for _, w := range s.Windows {
		if w.ID == input.Window {
			input.WindowIncarnation = w.Incarnation
			break
		}
	}
	return m.Input(testContext, input)
}
func treeKey(t *testing.T, m *Manual, code key.Codes) {
	t.Helper()
	if e := treeInput(t, m, ManualInput{Window: "one", Type: events.KeyChord, Code: code}); e != nil {
		t.Fatal(e)
	}
	unitSettle(t, m)
}
func TestManagedTreeActualGeometryAndKeyboard(t *testing.T) {
	for _, batch := range []int{1, 7, 64} {
		for _, dpi := range []float32{96, 120, 144, 192} {
			t.Run(fmt.Sprintf("batch-%d/dpi-%g", batch, dpi), func(t *testing.T) {
				o := nativeOutline(t, 10_000, true)
				config := manualConfig(t, batch)
				config.LogicalDPI = dpi
				focus := ui.FocusRequest{ID: "outline", Sequence: 1}
				m, a := gridManual(t, config, func() ui.Node { return ui.Node{ID: "outline", Kind: ui.DataTree, Tree: o, Grow: true, Focus: focus} })
				unitSettle(t, m)
				first := treeInspection(t, a, o)
				if first.TotalNodes != 10_000 || first.VisibleRows != 10_000 || len(first.Rows) == 0 || len(first.Rows) > 500 {
					t.Fatal("unbounded/empty rows", first)
				}
				body := image.Rect(first.Bounds.Min.X, first.Bounds.Min.Y, first.Bounds.Min.X+int(math.Round(float64(first.Viewport.Width)*float64(first.Scale))), first.Bounds.Min.Y+int(math.Round(float64(first.Viewport.Height)*float64(first.Scale))))
				for _, r := range first.Rows {
					if r.Bounds.Empty() {
						continue
					}
					y := r.Index*23 - first.Viewport.ScrollY
					expected := image.Rect(body.Min.X, body.Min.Y+int(math.Round(float64(y)*float64(first.Scale))), body.Max.X, body.Min.Y+int(math.Round(float64(y+23)*float64(first.Scale)))).Intersect(body)
					if r.Bounds != expected {
						t.Fatalf("row exterior %s got%v expected%v", r.Key, r.Bounds, expected)
					}
				}
				treeKey(t, m, key.CodeDownArrow)
				if s := treeState(t, a, o); s.Selected != "node-0" {
					t.Fatal("first keyboard selection", s)
				}
				treeKey(t, m, key.CodeDownArrow)
				if s := treeState(t, a, o); s.Selected != "node-1" {
					t.Fatal("move", s)
				}
				treeKey(t, m, key.CodePageDown)
				paged := treeState(t, a, o)
				if paged.Selected == "node-1" || paged.Viewport.ScrollY == 0 {
					t.Fatal("page", paged)
				}
				treeKey(t, m, key.CodeHome)
				if s := treeState(t, a, o); s.Selected != "node-0" {
					t.Fatal("home", s)
				}
				treeKey(t, m, key.CodeEnd)
				last := treeInspection(t, a, o)
				if last.Selected != "node-9999" || last.Viewport.ScrollY == 0 {
					t.Fatal("end", last)
				}
				if first.Rows[0].ID == last.Rows[0].ID || first.RowIncarnation == last.RowIncarnation {
					t.Fatal("pooled identity reused")
				}
				if e := treeInput(t, m, ManualInput{Window: "one", Type: events.Scroll, Position: last.Rows[0].Bounds.Min.Add(image.Pt(4, 4)), Delta: math32.Vector2{Y: -61}}); e != nil {
					t.Fatal(e)
				}
				unitSettle(t, m)
				if treeState(t, a, o).Viewport.ScrollY >= last.Viewport.ScrollY {
					t.Fatal("wheel did not scroll")
				}
			})
		}
	}
}
func TestManagedTreePointerDisclosureAndRetiredPress(t *testing.T) {
	o := nativeOutline(t, 1000, false)
	var request virtual.OutlineRequest
	changes := 0
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "outline", Kind: ui.DataTree, Tree: o, Grow: true, TreeChange: func(ui.TreeEvent) { changes++ }, TreeDisclosure: func(e ui.TreeDisclosureEvent) {
			var err error
			request, err = o.Prepare(e.Change)
			if err != nil {
				t.Error(err)
			}
		}}
	})
	unitSettle(t, m)
	first := treeInspection(t, a, o)
	row := first.Rows[0]
	point := row.DisclosureBounds.Min.Add(image.Pt(2, 2))
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if e := treeInput(t, m, ManualInput{Window: "one", Type: typ, Button: events.Left, Position: point}); e != nil {
			t.Fatal(e)
		}
		unitSettle(t, m)
	}
	if !treeState(t, a, o).Pending || treeState(t, a, o).VisibleRows != 1 {
		t.Fatal("disclosure scanned/committed synchronously")
	}
	builder, e := virtual.NewOutlineBuilder(testContext, request)
	if e != nil {
		t.Fatal(e)
	}
	for {
		done, e := builder.Step(7)
		if e != nil {
			t.Fatal(e)
		}
		if done {
			break
		}
	}
	index, e := builder.Result()
	if e != nil {
		t.Fatal(e)
	}
	if e = a.Do(func() {
		if e := o.Complete(request, index); e != nil {
			t.Error(e)
		}
		_ = a.Invalidate("one")
	}); e != nil {
		t.Fatal(e)
	}
	unitSettle(t, m)
	measured := treeInspection(t, a, o)
	target := measured.Rows[2]
	point = target.Bounds.Min.Add(image.Pt(30, 2))
	if e = treeInput(t, m, ManualInput{Window: "one", Type: events.MouseDown, Button: events.Left, Position: point}); e != nil {
		t.Fatal(e)
	}
	unitSettle(t, m)
	if e = a.Do(func() {
		v := treeStateUnsafe(o)
		v.ScrollY = 120
		if e := o.SetViewport(v); e != nil {
			t.Error(e)
		}
		_ = a.Invalidate("one")
	}); e != nil {
		t.Fatal(e)
	}
	unitSettle(t, m)
	if e = treeInput(t, m, ManualInput{Window: "one", Type: events.MouseUp, Button: events.Left, Position: point}); e != nil {
		t.Fatal(e)
	}
	unitSettle(t, m)
	if s := treeState(t, a, o); s.Selected != "" {
		t.Fatal("stale press selected row", s)
	}
	if changes == 0 {
		t.Fatal("viewport callback absent")
	}
}
func treeStateUnsafe(o *virtual.Outline) virtual.OutlineViewport {
	s, e := o.State()
	if e != nil {
		panic(e)
	}
	return s.Viewport
}
func TestManualNestedGridTreeSnapshotDetached(t *testing.T) {
	o := nativeOutline(t, 50, true)
	g, _ := testNativeGrid(t, 50)
	m, _ := gridManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{{ID: "outline", Kind: ui.DataTree, Tree: o, Grow: true}, {ID: "grid", Kind: ui.DataGrid, Grid: g, Grow: true}}}
	})
	unitSettle(t, m)
	one, e := m.Snapshot(testContext)
	if e != nil {
		t.Fatal(e)
	}
	if len(one.Windows) != 1 || len(one.Windows[0].Trees) != 1 || len(one.Windows[0].Grids) != 1 {
		t.Fatal("missing snapshots", one)
	}
	one.Windows[0].Trees[0].Rows[0].Label = "caller mutation"
	one.Windows[0].Trees[0].Pages[0].Name = "caller mutation"
	one.Windows[0].Grids[0].Cells[0].Text = "caller mutation"
	one.Windows[0].Grids[0].Headers[0].Title = "caller mutation"
	two, e := m.Snapshot(testContext)
	if e != nil {
		t.Fatal(e)
	}
	if two.Windows[0].Trees[0].Rows[0].Label == "caller mutation" || two.Windows[0].Trees[0].Pages[0].Name == "caller mutation" || two.Windows[0].Grids[0].Cells[0].Text == "caller mutation" || two.Windows[0].Grids[0].Headers[0].Title == "caller mutation" {
		t.Fatal("nested snapshot aliased internal cache")
	}
}
func TestManagedTreeBindingCloseAndCaps(t *testing.T) {
	o := nativeOutline(t, 100, true)
	disabled := true
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "outline", Kind: ui.DataTree, Tree: o, Grow: true, Disabled: disabled}
	})
	unitSettle(t, m)
	// Programmatic owner focus cannot enter a disabled tree; no stale row action
	// may bypass that scope. The second window keeps application lifetime alive.
	if e := a.Open(app.Window{ID: "other", Title: "Other", Width: 200, Height: 100, View: func() ui.Node { return ui.Node{ID: "label", Kind: ui.Label, Text: "Alive"} }}); e != nil {
		t.Fatal(e)
	}
	if e := a.CloseWindow("one"); e != nil {
		t.Fatal(e)
	}
	var e error
	if err := a.Do(func() { _, e = InspectTree(a, o) }); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(e, ErrTreeClosed) {
		t.Fatal("closed domain accepted inspection", e)
	}
	unitSettle(t, m)
	outlineRegistry.Lock()
	remaining := outlineRegistry.models[o]
	outlineRegistry.Unlock()
	if remaining != nil {
		t.Fatal("retired registry retained model")
	}
}

func TestManagedTreeModalDisabledAndInactiveTab(t *testing.T) {
	o := nativeOutline(t, 30, true)
	disabled, modal := false, false
	page := "first"
	focus := ui.FocusRequest{ID: "outline", Sequence: 1}
	m, a := gridManual(t, manualConfig(t, 1), func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: focus, Children: []ui.Node{{ID: "tabs", Kind: ui.Tabs, SelectedID: page, Options: []ui.Option{{ID: "first", Label: "Tree"}, {ID: "second", Label: "Other"}}, Children: []ui.Node{{ID: "first", Kind: ui.Column, Grow: true, Children: []ui.Node{{ID: "outline", Kind: ui.DataTree, Tree: o, Grow: true, Disabled: disabled}}}, {ID: "second", Kind: ui.Column, Children: []ui.Node{{ID: "other", Kind: ui.Button, Text: "Other"}}}}}, {ID: "dialog", Kind: ui.Dialog, Open: modal, Children: []ui.Node{{ID: "confirm", Kind: ui.Button, Text: "Confirm"}}}}}
	})
	unitSettle(t, m)
	treeKey(t, m, key.CodeDownArrow)
	if treeState(t, a, o).Selected != "node-0" {
		t.Fatal("initial selection")
	}
	for _, mode := range []string{"disabled", "inactive", "modal"} {
		if e := a.Do(func() {
			disabled = mode == "disabled"
			if mode == "inactive" {
				page = "second"
			} else {
				page = "first"
			}
			modal = mode == "modal"
			if modal {
				focus = ui.FocusRequest{ID: "confirm", Sequence: focus.Sequence + 1}
			}
			_ = a.Invalidate("one")
		}); e != nil {
			t.Fatal(e)
		}
		unitSettle(t, m)
		if e := a.Do(func() {
			outlineRegistry.Lock()
			binding := outlineRegistry.models[o]
			outlineRegistry.Unlock()
			if binding.eligible() {
				t.Error("hidden/refused tree eligible", mode)
			}
		}); e != nil {
			t.Fatal(e)
		}
		treeKey(t, m, key.CodeDownArrow)
		if treeState(t, a, o).Selected != "node-0" {
			t.Fatal("refused scope changed selection", mode)
		}
		if mode == "inactive" {
			if e := a.Do(func() {
				nodes, _ := accessibilitySnapshot(m.implementation.find("one").state.root, nil)
				for _, n := range nodes {
					if n.Role == "tree-row" {
						t.Error("inactive rows exposed", n)
					}
				}
			}); e != nil {
				t.Fatal(e)
			}
		}
	}
	if e := a.Do(func() {
		disabled, modal = false, false
		page = "first"
		focus = ui.FocusRequest{ID: "outline", Sequence: focus.Sequence + 1}
		_ = a.Invalidate("one")
	}); e != nil {
		t.Fatal(e)
	}
	unitSettle(t, m)
	treeKey(t, m, key.CodeDownArrow)
	if treeState(t, a, o).Selected != "node-1" {
		t.Fatal("scope did not resume")
	}
}
func TestManagedTreeAccessibilityBoundedRowsAndRetirement(t *testing.T) {
	o := nativeOutline(t, 1000, true)
	disclosures := 0
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "outline", Kind: ui.DataTree, Tree: o, Grow: true, TreeDisclosure: func(ui.TreeDisclosureEvent) { disclosures++ }}
	})
	unitSettle(t, m)
	var handlers map[string]accessibilityHandler
	var nodes []accessibilityNode
	if e := a.Do(func() { nodes, handlers = accessibilitySnapshot(m.implementation.find("one").state.root, nil) }); e != nil {
		t.Fatal(e)
	}
	measured := treeInspection(t, a, o)
	rows := 0
	for _, n := range nodes {
		if n.Role == "tree-row" {
			rows++
			if n.DomainID == "node-0" || n.Level < 1 || n.RowCount != 1000 {
				t.Fatal("row semantics", n)
			}
		}
	}
	if rows != len(measured.Rows) || len(nodes) > 505 {
		t.Fatal("AX full tree exposed", rows, len(nodes))
	}
	old := handlers[accessibilityManagedID(measured.Rows[0].ID)]
	if old == nil {
		t.Fatal("missing branch action")
	}
	if e := a.Do(func() {
		if !old(accessibilityAction{Kind: "press"}) {
			t.Error("row press refused")
		}
	}); e != nil {
		t.Fatal(e)
	}
	unitSettle(t, m)
	if treeState(t, a, o).Selected != "node-0" {
		t.Fatal("AX select not accepted")
	}
	if e := a.Do(func() {
		if old(accessibilityAction{Kind: "collapse"}) {
			t.Error("retired ref action accepted")
		}
	}); e != nil {
		t.Fatal(e)
	}
	if disclosures != 0 {
		t.Fatal("retired disclosure callback invoked")
	}
	if e := a.Do(func() {
		_, pages := accessibilitySnapshot(m.implementation.find("one").state.root, nil)
		fn := pages[accessibilityManagedID(TreePageID("outline", 1))]
		if fn == nil || !fn(accessibilityAction{Kind: "press"}) {
			t.Error("next page absent/refused")
		}
	}); e != nil {
		t.Fatal(e)
	}
	unitSettle(t, m)
	next := treeInspection(t, a, o)
	if next.Viewport.ScrollY == 0 || next.Rows[0].ID == measured.Rows[0].ID {
		t.Fatal("AX paging did not retire rows")
	}
}
func TestManagedTreeBindingRefusalAndConstructionRollback(t *testing.T) {
	o := nativeOutline(t, 10, true)
	config := manualConfig(t, 7)
	a, e := app.New(app.Config{})
	if e != nil {
		t.Fatal(e)
	}
	defer a.Close()
	secondTree := true
	for _, id := range []string{"one", "two"} {
		id := id
		if e = a.Open(app.Window{ID: id, Title: id, Width: 400, Height: 240, View: func() ui.Node {
			if id == "two" && !secondTree {
				return ui.Node{ID: "label", Kind: ui.Label}
			}
			return ui.Node{ID: "outline", Kind: ui.DataTree, Tree: o, Grow: true}
		}}); e != nil {
			t.Fatal(e)
		}
	}
	if _, e = NewManual(testContext, a, config); !errors.Is(e, ErrTreeBinding) {
		t.Fatal("cross-window live binding admitted", e)
	}
	outlineRegistry.Lock()
	left := outlineRegistry.models[o]
	outlineRegistry.Unlock()
	if left != nil {
		t.Fatal("failed construction left claim")
	}
	secondTree = false
	m, e := NewManual(testContext, a, config)
	if e != nil {
		t.Fatal("refusal poisoned valid remount", e)
	}
	defer m.Close(testContext)
	unitSettle(t, m)
	var wrong error
	if e = a.Do(func() { _, wrong = InspectTree(&app.Application{}, o) }); e != nil {
		t.Fatal(e)
	}
	if !errors.Is(wrong, ErrTreeBinding) {
		t.Fatal(wrong)
	}
	data, e := virtual.NewTreeData(testContext, "cap", 1, []virtual.TreeNode{{ID: "r"}}, virtual.TreeLimits{})
	if e != nil {
		t.Fatal(e)
	}
	index, e := virtual.BuildTreeIndex(testContext, data, virtual.TreeIndexOptions{})
	if e != nil {
		t.Fatal(e)
	}
	big, e := virtual.NewOutline(index, virtual.OutlineConfig{RowHeight: 1, MaxRows: 501}, virtual.OutlineViewport{Width: 100, Height: 100})
	if e != nil {
		t.Fatal(e)
	}
	if e = validateNativeTree(ui.Node{ID: "bad", Kind: ui.DataTree, Tree: big}); !errors.Is(e, ErrTreeBudget) {
		t.Fatal("native pool cap not enforced", e)
	}
}

func TestManagedTreeRefusedPayloadDoesNotGrowOrPublish(t *testing.T) {
	o := nativeOutline(t, 10, true)
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node { return ui.Node{ID: "outline", Kind: ui.DataTree, Tree: o, Grow: true} })
	unitSettle(t, m)
	nodes := []virtual.TreeNode{{ID: strings.Repeat("x", treeNativeMaxKey+1), Label: "Too long"}}
	data, e := virtual.NewTreeData(testContext, "external-long-key", 1, nodes, virtual.TreeLimits{})
	if e != nil {
		t.Fatal(e)
	}
	var req virtual.OutlineRequest
	if e = a.Do(func() { req, e = o.PrepareData(data, virtual.OutlineChange{Kind: virtual.OutlineKeep}) }); e != nil {
		t.Fatal(e)
	}
	if e != nil {
		t.Fatal(e)
	}
	builder, e := virtual.NewOutlineBuilder(testContext, req)
	if e != nil {
		t.Fatal(e)
	}
	for {
		done, e := builder.Step(7)
		if e != nil {
			t.Fatal(e)
		}
		if done {
			break
		}
	}
	index, e := builder.Result()
	if e != nil {
		t.Fatal(e)
	}
	if e = a.Do(func() {
		binding := m.implementation.find("one").state.root.control.outline
		oldRows := len(binding.rows)
		if e = o.Complete(req, index); e != nil {
			t.Error(e)
			return
		}
		if e = syncTree(binding.r); !errors.Is(e, ErrTreeBudget) {
			t.Error("oversized key admitted", e)
		}
		if len(binding.rows) != oldRows {
			t.Error("refusal grew/shrank native pool")
		}
		binding.capture() // A later paint must not republish the rejected old projection.
		if binding.measurement.Load() != nil {
			t.Error("refused model retained completed publication")
		}
		nodes, _ := treeAccessibility(binding.r, "parent", nil, false)
		if len(nodes) != 0 {
			t.Error("refused model exposed old AX rows")
		}
	}); e != nil {
		t.Fatal(e)
	}
}

func TestManagedTreeAbnormalConstructionRetiresClaim(t *testing.T) {
	for _, mode := range []string{"panic", "goexit"} {
		t.Run(mode, func(t *testing.T) {
			o := nativeOutline(t, 30, true)
			a, e := app.New(app.Config{})
			if e != nil {
				t.Fatal(e)
			}
			config := manualConfig(t, 7)
			if e = a.Open(app.Window{ID: "one", Title: "Tree", Width: 400, Height: 240, View: func() ui.Node { return ui.Node{ID: "outline", Kind: ui.DataTree, Tree: o, Grow: true} }}); e != nil {
				t.Fatal(e)
			}
			if e = a.Open(app.Window{ID: "two", Title: "Abnormal", Width: 200, Height: 100, View: func() ui.Node {
				if mode == "panic" {
					panic("tree-second-view")
				}
				runtime.Goexit()
				return ui.Node{}
			}}); e != nil {
				t.Fatal(e)
			}
			done := make(chan struct{})
			var abnormal any
			go func() {
				defer close(done)
				defer func() { abnormal = recover() }()
				_, _ = NewManual(testContext, a, config)
			}()
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("construction did not unwind")
			}
			outlineRegistry.Lock()
			left := outlineRegistry.models[o]
			outlineRegistry.Unlock()
			if left != nil {
				t.Fatal("abnormal construction retained model")
			}
			if a.Context().Err() == nil {
				t.Fatal("abnormal owner callback left domain live")
			}
			if mode == "panic" && abnormal != "tree-second-view" {
				t.Fatal("original panic changed", abnormal)
			}
			if mode == "goexit" && abnormal != nil {
				t.Fatal("Goexit changed to panic", abnormal)
			}
		})
	}
}
func TestManagedTreeRemountRetiresNativeIdentities(t *testing.T) {
	o := nativeOutline(t, 30, true)
	id := "outline"
	focus := ui.FocusRequest{ID: id, Sequence: 1}
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: focus, Children: []ui.Node{{ID: id, Kind: ui.DataTree, Tree: o, Grow: true}}}
	})
	unitSettle(t, m)
	old := treeInspection(t, a, o)
	snapshot, e := m.Snapshot(testContext)
	if e != nil {
		t.Fatal(e)
	}
	nativeWindow := snapshot.Windows[0]
	var oldAction accessibilityHandler
	if e = a.Do(func() {
		_, handlers := accessibilitySnapshot(m.implementation.find("one").state.root, nil)
		oldAction = handlers[accessibilityManagedID(old.Rows[0].ID)]
		id = "remounted"
		focus = ui.FocusRequest{ID: id, Sequence: 2}
		_ = a.Invalidate("one")
	}); e != nil {
		t.Fatal(e)
	}
	unitSettle(t, m)
	current := treeInspection(t, a, o)
	if current.ID != id || current.Incarnation == old.Incarnation || current.Rows[0].ID == old.Rows[0].ID {
		t.Fatal("native incarnation/target did not retire")
	}
	if e = a.Do(func() {
		if oldAction == nil || oldAction(accessibilityAction{Kind: "press"}) {
			t.Error("retired action admitted")
		}
	}); e != nil {
		t.Fatal(e)
	}
	if e = m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: nativeWindow.Incarnation, Target: old.Rows[0].ID, TargetIncarnation: old.RowIncarnation, Type: events.MouseDown, Button: events.Left}); !errors.Is(e, ErrManualStale) {
		t.Fatal("retired target admitted", e)
	}
}
