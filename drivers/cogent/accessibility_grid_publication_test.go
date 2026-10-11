//go:build offscreen

package cogent

import (
	"encoding/json"
	"errors"
	"image"
	"math"
	"testing"

	"cogentcore.org/core/core"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

// These tests exercise owner publications and actual software-rendered managed
// grids. They do not measure Cocoa delivery, screen-reader behavior, or GPU work.
func gridAXCompleted(scope, parent string, incarnation, generation uint64) []accessibilityNode {
	frame := func(x, y, w, h float32) *[4]float32 { v := [4]float32{x, y, w, h}; return &v }
	selection, expanded, minimum, maximum := [2]int{1, 2}, true, float32(0), float32(100)
	nodes := []accessibilityNode{
		{ID: scope, Parent: parent, Role: "table", Name: "Completed table", RowCount: 10000, ColumnCount: 20, Focusable: true, Focused: true, Actions: []string{"focus"}, GridFrame: frame(10, 20, 300, 180)},
		{ID: scope + ":row", Parent: scope, Role: "grid-row", Name: "row-7", RowIndex: 8, Selected: true, Actions: []string{}, GridFrame: frame(10, 70, 240, 24)},
		{ID: scope + ":cell", Parent: scope + ":row", Role: "grid-cell", Value: "old cell", Selected: true, Actions: []string{"press"}, GridFrame: frame(10, 70, 90, 24)},
		{ID: scope + ":header", Parent: scope, Role: "grid-column", Name: "Column 1", Actions: []string{}, GridFrame: frame(10, 20, 90, 26)},
		{ID: scope + ":editor", Parent: scope, Role: string(ui.TextField), Value: "A😀", Focused: true, Focusable: true, Actions: []string{"focus", "set", "selection"}, TextClient: 9, TextSelection: &selection, TextFocus: 3, Expanded: &expanded, Min: &minimum, Max: &maximum, GridFrame: frame(10, 70, 90, 24)},
	}
	for i := range nodes {
		nodes[i].GridScope = scope
		nodes[i].GridIncarnation = incarnation
		nodes[i].GridGeneration = generation
		nodes[i].GridTargetGeneration = generation
	}
	return nodes
}

func gridAXPending(scope, parent string, incarnation, target uint64) []accessibilityNode {
	n := gridAXCompleted(scope, parent, incarnation, target)[0]
	n.GridPending, n.GridGeneration, n.GridFrame = true, 0, nil
	return []accessibilityNode{n}
}

func gridAXHandlers(nodes []accessibilityNode, calls *int) map[string]accessibilityHandler {
	handlers := map[string]accessibilityHandler{}
	for _, n := range nodes {
		if len(n.Actions) > 0 {
			handlers[n.ID] = func(accessibilityAction) bool {
				if calls != nil {
					*calls++
				}
				return true
			}
		}
	}
	return handlers
}

func gridAXByID(t *testing.T, nodes []accessibilityNode) map[string]accessibilityNode {
	t.Helper()
	out := map[string]accessibilityNode{}
	for _, n := range nodes {
		if _, duplicate := out[n.ID]; duplicate {
			t.Fatalf("duplicate published ID %q", n.ID)
		}
		out[n.ID] = n
	}
	return out
}

func gridAXLast(t *testing.T, b *accessibilityBridge) []accessibilityNode {
	t.Helper()
	b.mu.Lock()
	payload := b.lastSnapshot
	b.mu.Unlock()
	var nodes []accessibilityNode
	if err := json.Unmarshal([]byte(payload), &nodes); err != nil {
		t.Fatal(err)
	}
	return nodes
}

func gridAXAssertPending(t *testing.T, nodes []accessibilityNode, scope string, completed, target uint64, handlers map[string]accessibilityHandler) {
	t.Helper()
	count := 0
	for _, n := range nodes {
		if n.GridScope != scope {
			continue
		}
		count++
		if !n.GridPending || n.GridGeneration != completed || n.GridTargetGeneration != target || !n.Disabled || n.Focused || len(n.Actions) != 0 || n.TextClient != 0 || n.TextSelection != nil || n.TextFocus != 0 || handlers[n.ID] != nil {
			t.Errorf("pending node retained action/current-text authority: %+v", n)
		}
	}
	if count == 0 {
		t.Fatal("pending host vanished")
	}
}

func TestAccessibilityGridPublicationRetainsCompletedAndDetachesCache(t *testing.T) {
	b := &accessibilityBridge{}
	ready := gridAXCompleted("grid-one", "root", 11, 5)
	_, _ = b.prepareGridPublication(ready, gridAXHandlers(ready, nil))
	// Mutating an input retained by the caller must not mutate the completed cache.
	ready[0].Name = "caller mutation"
	ready[2].Actions[0] = "caller action"
	ready[2].GridFrame[0] = 900
	ready[4].TextSelection[0] = 900
	*ready[4].Expanded = false
	*ready[4].Min, *ready[4].Max = 900, 901
	pending := gridAXPending("grid-one", "root", 11, 6)
	pending[0].Name, pending[0].Help = "Current label", "Current help"
	out, handlers := b.prepareGridPublication(pending, gridAXHandlers(pending, nil))
	gridAXAssertPending(t, out, "grid-one", 5, 6, handlers)
	if len(out) != 5 || out[0].Name != "Current label" || out[0].Help != "Current help" || out[0].RowCount != 10000 || out[2].Value != "old cell" || *out[2].GridFrame != [4]float32{10, 70, 90, 24} {
		t.Fatalf("completed data/geometry were replaced by pending data: %+v", out)
	}
	if !*out[4].Expanded || *out[4].Min != 0 || *out[4].Max != 100 {
		t.Fatal("completed pointer fields alias input")
	}
	// Returned read-only records also cannot corrupt the next retained publication.
	out[2].GridFrame[0] = 800
	*out[4].Expanded = false
	*out[4].Min, *out[4].Max = 800, 801
	again, _ := b.prepareGridPublication(pending, nil)
	if *again[2].GridFrame != [4]float32{10, 70, 90, 24} || !*again[4].Expanded || *again[4].Min != 0 || *again[4].Max != 100 {
		t.Fatal("returned pending pointers alias completed cache")
	}
	cached := b.gridPublications["grid-one"]
	if cached[2].Actions[0] != "press" || *cached[4].TextSelection != [2]int{1, 2} || cached[4].TextClient != 9 || cached[0].GridPending {
		t.Fatal("pending transformation mutated completed cache")
	}
}

func TestAccessibilityGridPublicationStartupAndScopeRetirement(t *testing.T) {
	for _, tc := range []struct {
		name    string
		pending func() []accessibilityNode
		remove  bool
	}{
		{"startup", func() []accessibilityNode { return gridAXPending("g", "root", 1, 2) }, true},
		{"incarnation", func() []accessibilityNode { return gridAXPending("g", "root", 2, 2) }, false},
		{"parent", func() []accessibilityNode { return gridAXPending("g", "other", 1, 2) }, false},
		{"namespace", func() []accessibilityNode { return gridAXPending("panel-b:g", "panel-b:root", 1, 2) }, false},
		{"removed", func() []accessibilityNode { return gridAXPending("g", "root", 1, 2) }, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := &accessibilityBridge{}
			if tc.name != "startup" {
				ready := gridAXCompleted("g", "root", 1, 1)
				b.prepareGridPublication(ready, gridAXHandlers(ready, nil))
			}
			if tc.remove {
				b.prepareGridPublication([]accessibilityNode{{ID: "button", Role: "button"}}, nil)
			}
			pending := tc.pending()
			out, handlers := b.prepareGridPublication(pending, gridAXHandlers(pending, nil))
			gridAXAssertPending(t, out, pending[0].GridScope, 0, 2, handlers)
			if len(out) != 1 || out[0].GridFrame != nil || len(b.gridPublications) != 0 || out[0].GridIncarnation != pending[0].GridIncarnation || out[0].Parent != pending[0].Parent {
				t.Fatalf("retired or foreign subtree resurrected: %+v cache=%v", out, b.gridPublications)
			}
		})
	}
}

func TestAccessibilityGridPublicationMixedScopesAndLatestCompletion(t *testing.T) {
	b := &accessibilityBridge{}
	one := gridAXCompleted("one", "root", 10, 1)
	two := gridAXCompleted("two", "root", 20, 1)
	button := accessibilityNode{ID: "button", Role: "button", Name: "Old", Actions: []string{"press"}}
	initial := append(append(one, two...), button)
	b.prepareGridPublication(initial, gridAXHandlers(initial, nil))
	two = gridAXCompleted("two", "root", 20, 2)
	two[2].Value = "current second grid"
	button.Name = "Current"
	input := append(append(gridAXPending("one", "root", 10, 3), two...), button)
	out, handlers := b.prepareGridPublication(input, gridAXHandlers(input, nil))
	gridAXAssertPending(t, out, "one", 1, 3, handlers)
	byID := gridAXByID(t, out)
	if byID["two:cell"].Value != "current second grid" || byID["two:cell"].GridPending || handlers["two:cell"] == nil || byID["button"].Name != "Current" || handlers["button"] == nil {
		t.Fatal("pending grid starved ready grid or ordinary controls")
	}
	// Completing a new visible range must replace, rather than append to, the cache.
	completed := gridAXCompleted("one", "root", 10, 3)
	completed[1].ID = "one:new-row"
	completed[2].ID, completed[2].Parent = "one:new-cell", "one:new-row"
	completed[2].Value = "new cell"
	input = append(append(completed, two...), button)
	out, handlers = b.prepareGridPublication(input, gridAXHandlers(input, nil))
	byID = gridAXByID(t, out)
	if _, stale := byID["one:cell"]; stale || handlers["one:cell"] != nil || byID["one:new-cell"].Value != "new cell" || handlers["one:new-cell"] == nil {
		t.Fatal("older completed cell/handler survived a new completion")
	}
	out, handlers = b.prepareGridPublication(append(two, button), gridAXHandlers(append(two, button), nil))
	if len(b.gridPublications) != 1 || len(out) != len(two)+1 || handlers["one:new-cell"] != nil {
		t.Fatal("removed scope retained cache or action")
	}
}

func TestAccessibilityGridPublicationDuplicatePendingAndQueuedActions(t *testing.T) {
	a, b := newAXTest(t, 16)
	calls, readyCalls := 0, 0
	one := gridAXCompleted("g", "root", 7, 4)
	control := accessibilityNode{ID: "button", Role: "button", Actions: []string{"press"}}
	ready := append(one, control)
	b.update(ready, gridAXHandlers(ready, &calls))
	completedGeneration := b.generation.Load()
	for _, id := range []string{"g", "g:cell", "g:editor"} {
		if !accessibilityDeliver(b.token, completedGeneration, id, accessibilityAction{Kind: "press"}) {
			t.Fatalf("completed fixture action %q not admitted", id)
		}
	}
	pending := append(gridAXPending("g", "root", 7, 5), control)
	b.update(pending, map[string]accessibilityHandler{"button": func(accessibilityAction) bool { readyCalls++; return true }})
	pendingGeneration := b.generation.Load()
	if pendingGeneration == completedGeneration {
		t.Fatal("pending transition did not retire admitted completed actions")
	}
	for _, id := range []string{"g", "g:cell", "g:editor"} {
		if accessibilityDeliver(b.token, pendingGeneration, id, accessibilityAction{Kind: "press"}) {
			t.Errorf("read-only retained action %q admitted", id)
		}
	}
	if !accessibilityDeliver(b.token, pendingGeneration, "button", accessibilityAction{Kind: "press"}) {
		t.Fatal("ordinary action blocked by pending grid")
	}
	before := b.lastSnapshot
	b.update(pending, map[string]accessibilityHandler{"button": func(accessibilityAction) bool { readyCalls++; return true }})
	if b.generation.Load() != pendingGeneration || b.lastSnapshot != before {
		t.Fatal("identical pending target churned semantic/global generation")
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if calls != 0 || readyCalls != 1 {
		t.Fatalf("queued pending transition: retired=%d ordinary=%d", calls, readyCalls)
	}
	newer := append(gridAXPending("g", "root", 7, 6), control)
	b.update(newer, gridAXHandlers(newer, nil))
	if b.generation.Load() == pendingGeneration {
		t.Fatal("new pending target did not supersede old publication")
	}
	nodes := gridAXLast(t, b)
	gridAXAssertPending(t, nodes, "g", 4, 6, b.handlers)
	if nodes[2].Value != "old cell" {
		t.Fatal("pending target overwrote last completed data")
	}
	b.close()
	if len(b.gridPublications) != 0 {
		t.Fatal("closed bridge retained completed grid")
	}
	if out, handlers := b.prepareGridPublication(one, gridAXHandlers(one, nil)); out != nil || handlers != nil || len(b.gridPublications) != 0 {
		t.Fatal("closed bridge resurrected cache")
	}
}

func TestAccessibilityGridPublicationRejectedEncodingDoesNotPoisonCache(t *testing.T) {
	_, b := newAXTest(t, 4)
	ready := gridAXCompleted("g", "root", 1, 1)
	b.update(ready, nil)
	generation := b.generation.Load()
	bad := gridAXCompleted("g", "root", 2, 2)
	bad = append(bad, accessibilityNode{ID: "bad", Value: math.NaN()})
	b.update(bad, nil)
	if b.generation.Load() != generation || b.gridPublications["g"][0].GridIncarnation != 1 {
		t.Fatal("unencodable publication changed completed cache or generation")
	}
	b.update(gridAXPending("g", "root", 1, 3), nil)
	gridAXAssertPending(t, gridAXLast(t, b), "g", 1, 3, b.handlers)
}

func TestAccessibilityGridPublicationNamespaceMetadata(t *testing.T) {
	ready := gridAXCompleted("g", "root", 4, 8)
	prefix := accessibilityWorkbenchNamespace("workbench", "panel")
	nodes, handlers := namespaceAccessibility(ready, gridAXHandlers(ready, nil), prefix, &accessibilityBridge{})
	for _, n := range nodes {
		if n.GridScope != prefix+"g" || n.ID[:len(prefix)] != prefix || n.Parent[:len(prefix)] != prefix || n.GridGeneration != 8 || n.GridIncarnation != 4 || n.GridFrame == nil {
			t.Errorf("namespace separated grid provenance from node identity: %+v", n)
		}
	}
	if handlers[prefix+"g:cell"] == nil || handlers["g:cell"] != nil {
		t.Fatal("namespace left an unqualified grid action")
	}
	b := &accessibilityBridge{}
	b.prepareGridPublication(nodes, handlers)
	pending := gridAXPending("g", "root", 4, 9)
	pending, actions := namespaceAccessibility(pending, gridAXHandlers(pending, nil), prefix, nil)
	out, actions := b.prepareGridPublication(pending, actions)
	gridAXAssertPending(t, out, prefix+"g", 8, 9, actions)
	if len(out) != 5 {
		t.Fatal("qualified scope did not retain its own completed rows")
	}
}

func TestAccessibilityGridPublicationCompletedMeasurementChanges(t *testing.T) {
	baseline := &GridInspection{Window: "one", ID: "g", Generation: 3, Incarnation: 4, CellIncarnation: 5, Frames: 10, Bounds: image.Rect(10, 20, 310, 200), Scale: 1,
		Cells:   []GridCellGeometry{{ID: "cell", RowKey: "row-1", ColumnID: "c1", Text: "text", Bounds: image.Rect(10, 40, 100, 64)}},
		Headers: []GridHeaderGeometry{{ID: "header", Title: "Column", Bounds: image.Rect(10, 20, 100, 40)}},
		Editor:  &GridEditorGeometry{ID: "editor", Session: 6, Text: "draft", Focused: true, Bounds: image.Rect(10, 40, 100, 64)}}
	clone := func() *GridInspection {
		out := *baseline
		out.Cells = append([]GridCellGeometry(nil), baseline.Cells...)
		out.Headers = append([]GridHeaderGeometry(nil), baseline.Headers...)
		e := *baseline.Editor
		out.Editor = &e
		return &out
	}
	frames := clone()
	frames.Frames++
	if !sameGridAccessibilityMeasurement(baseline, frames) || !sameGridAccessibilityMeasurement(nil, nil) || sameGridAccessibilityMeasurement(nil, baseline) || sameGridAccessibilityMeasurement(baseline, nil) {
		t.Fatal("frame ordinal/no-frame comparison is wrong")
	}
	for _, tc := range []struct {
		name   string
		change func(*GridInspection)
	}{
		{"window", func(m *GridInspection) { m.Window = "other" }},
		{"ID", func(m *GridInspection) { m.ID = "other" }},
		{"generation", func(m *GridInspection) { m.Generation++ }},
		{"cell incarnation", func(m *GridInspection) { m.CellIncarnation++ }},
		{"stamp", func(m *GridInspection) { m.Stamp.Revision++ }},
		{"viewport", func(m *GridInspection) { m.Viewport.ScrollY++ }},
		{"column count", func(m *GridInspection) { m.Columns++ }},
		{"retained cells", func(m *GridInspection) { m.RetainedCells++ }},
		{"retained headers", func(m *GridInspection) { m.RetainedHeaders++ }},
		{"row selection mode", func(m *GridInspection) { m.RowSelection = !m.RowSelection }},
		{"selection state", func(m *GridInspection) { m.Selection.Revision++ }},
		{"edit state", func(m *GridInspection) { m.Edit.Revision++ }},
		{"host movement", func(m *GridInspection) { m.Bounds = m.Bounds.Add(image.Pt(1, 0)) }},
		{"cell movement", func(m *GridInspection) { m.Cells[0].Bounds = m.Cells[0].Bounds.Add(image.Pt(0, 1)) }},
		{"header movement", func(m *GridInspection) { m.Headers[0].Bounds = m.Headers[0].Bounds.Add(image.Pt(0, 1)) }},
		{"scale", func(m *GridInspection) { m.Scale = 1.25 }},
		{"cell value", func(m *GridInspection) { m.Cells[0].Text = "new" }},
		{"selection", func(m *GridInspection) { m.Cells[0].Selected = true }},
		{"editor movement", func(m *GridInspection) { m.Editor.Bounds = m.Editor.Bounds.Add(image.Pt(1, 0)) }},
		{"editor text", func(m *GridInspection) { m.Editor.Text = "new" }},
		{"editor focus", func(m *GridInspection) { m.Editor.Focused = false }},
		{"editor retirement", func(m *GridInspection) { m.Editor = nil }},
		{"binding", func(m *GridInspection) { m.Incarnation++ }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed := clone()
			tc.change(changed)
			if sameGridAccessibilityMeasurement(baseline, changed) {
				t.Fatal("completed metadata change was suppressed")
			}
		})
	}
}

func TestAccessibilityGridPublicationActualCompletedFramesAndEditor(t *testing.T) {
	m, a, g, _, _ := inlineFixture(t, 120)
	gridFrame(t, m)
	inspection := gridInspection(t, a, g)
	if inspection.Editor == nil || inspection.Editor.Bounds.Empty() {
		t.Fatal("fixture has no completed inline editor")
	}
	if err := a.Do(func() {
		binding := gBinding(t, g)
		nodes, handlers := accessibilitySnapshot(binding.r, nil)
		frames := map[string][4]float32{}
		put := func(id string, rect image.Rectangle) {
			frames[id] = [4]float32{float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy())}
		}
		put(accessibilityManagedID("g"), inspection.Bounds)
		rows := map[string]image.Rectangle{}
		for _, c := range inspection.Cells {
			if c.Bounds.Empty() {
				continue
			}
			put(accessibilityManagedID(c.ID), c.Bounds)
			rows[c.RowKey] = rows[c.RowKey].Union(c.Bounds)
		}
		for key, rect := range rows {
			put(accessibilityRowID("g", key), rect)
		}
		for _, h := range inspection.Headers {
			if !h.Bounds.Empty() {
				put(accessibilityManagedID(h.ID), h.Bounds)
			}
		}
		put(accessibilityManagedID(inspection.Editor.ID), inspection.Editor.Bounds)
		if len(nodes) != len(frames) {
			t.Errorf("completed scope rows/header/editor mismatch: nodes=%d frames=%d", len(nodes), len(frames))
		}
		for _, n := range nodes {
			frame, found := frames[n.ID]
			if !found || n.GridFrame == nil || *n.GridFrame != frame || n.GridScope != accessibilityManagedID("g") || n.GridPending || n.GridGeneration != inspection.Generation || n.GridTargetGeneration != inspection.Generation || n.GridIncarnation != inspection.Incarnation {
				t.Errorf("node does not describe its completed measured frame: %+v expected=%v", n, frame)
			}
		}
		editorNode := gridAXByID(t, nodes)[accessibilityManagedID(inspection.Editor.ID)]
		if editorNode.TextSelection == nil || editorNode.Value != inspection.Editor.Text || handlers[editorNode.ID] == nil {
			t.Error("completed inline editor semantics/actions missing")
		}
		// Frame metadata must be a copy, not a writable reference into measurement.
		nodes[0].GridFrame[0]++
		if binding.measurement.Load().Bounds != inspection.Bounds {
			t.Error("published frame pointer mutated measurement")
		}
		previous := binding.measurement.Load()
		withoutEditor := *previous
		withoutEditor.Editor = nil
		binding.measurement.Store(&withoutEditor)
		pending, pendingActions := accessibilitySnapshot(binding.r, nil)
		for _, n := range pending {
			if !n.GridPending || n.GridGeneration != 0 || n.GridFrame != nil {
				t.Errorf("unmeasured editor mixed into completed grid: %+v", n)
			}
		}
		b := &accessibilityBridge{}
		// Obtain the unmodified completed snapshot before priming the cache.
		binding.measurement.Store(previous)
		complete, completeActions := accessibilitySnapshot(binding.r, nil)
		b.prepareGridPublication(complete, completeActions)
		retained, actions := b.prepareGridPublication(pending, pendingActions)
		gridAXAssertPending(t, retained, accessibilityManagedID("g"), inspection.Generation, inspection.Generation, actions)
		if len(retained) != len(complete) {
			t.Error("pending editor measurement lost completed subtree")
		}
	}); err != nil {
		t.Fatal(err)
	}
}

func TestAccessibilityGridPublicationActualReconcilePendingAndCompletion(t *testing.T) {
	g, _ := testNativeGrid(t, 10000)
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{{ID: "button", Kind: ui.Button, Text: "Current control", Activate: func() {}}, {ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true}}}
	})
	gridFrame(t, m)
	b := newAccessibilityBridge(a)
	b.bind("one")
	t.Cleanup(b.close)
	window := m.implementation.find("one")
	var complete []accessibilityNode
	if err := a.Do(func() {
		var handlers map[string]accessibilityHandler
		complete, handlers = accessibilitySnapshot(window.state.root, b)
		b.update(complete, handlers)
	}); err != nil {
		t.Fatal(err)
	}
	scope := accessibilityManagedID("g")
	oldCells := map[string]bool{}
	var cellID string
	for _, n := range complete {
		if n.GridScope == scope && n.Role == "grid-cell" {
			oldCells[n.ID] = true
			cellID = n.ID
		}
	}
	if len(oldCells) == 0 {
		t.Fatalf("fixture completed no native cells: measurement=%+v nodes=%+v", gridInspection(t, a, g), complete)
	}
	completedGeneration := b.generation.Load()
	if !accessibilityDeliver(b.token, completedGeneration, cellID, accessibilityAction{Kind: "press"}) {
		t.Fatal("completed actual cell action not queued")
	}
	if err := a.Do(func() {
		if e := g.Scroll(270, 2400); e != nil {
			t.Error(e)
		}
		_ = a.Invalidate("one")
	}); err != nil {
		t.Fatal(err)
	}
	// Use the actual Manual owner reconciliation boundary, before an explicit
	// frame, rather than fabricating a pending model or invoking an action directly.
	if err := m.implementation.refresh(window); err != nil {
		t.Fatal(err)
	}
	var inspectErr error
	var targetGeneration uint64
	if err := a.Do(func() {
		_, inspectErr = InspectGrid(a, g)
		targetGeneration = gBinding(t, g).projection.State.Generation
		nodes, handlers := accessibilitySnapshot(window.state.root, b)
		b.update(nodes, handlers)
	}); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(inspectErr, ErrGridLayout) {
		t.Fatal("pending InspectGrid returned old completed geometry", inspectErr)
	}
	pending := gridAXLast(t, b)
	var completedGridGeneration uint64
	for _, n := range complete {
		if n.ID == scope {
			completedGridGeneration = n.GridGeneration
		}
	}
	gridAXAssertPending(t, pending, scope, completedGridGeneration, targetGeneration, b.handlers)
	if targetGeneration <= completedGridGeneration {
		t.Fatal("scroll did not advance accepted target generation")
	}
	for id := range oldCells {
		if _, ok := gridAXByID(t, pending)[id]; !ok {
			t.Errorf("last drawn cell %q withdrawn during pending", id)
		}
	}
	if b.handlers[accessibilityManagedID("button")] == nil {
		t.Fatal("pending projection removed ordinary control")
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if err := a.Do(func() {
		state, e := g.State()
		if e != nil || state.Active.RowKey != "" {
			t.Error("queued old cell changed pending selection", state.Active, e)
		}
	}); err != nil {
		t.Fatal(err)
	}
	gridFrame(t, m)
	inspection := gridInspection(t, a, g)
	if err := a.Do(func() { nodes, handlers := accessibilitySnapshot(window.state.root, b); b.update(nodes, handlers) }); err != nil {
		t.Fatal(err)
	}
	completed := gridAXByID(t, gridAXLast(t, b))
	if completed[scope].GridPending || completed[scope].GridGeneration != inspection.Generation || completed[scope].GridFrame == nil {
		t.Fatal("completed actual frame did not replace pending publication")
	}
	for id := range oldCells {
		if _, retained := completed[id]; retained || b.handlers[id] != nil {
			t.Errorf("old range cell %q resurrected after new frame", id)
		}
	}
	// Cache retirement is keyed to the currently traversed scopes, including an
	// otherwise still-live grid that leaves the visible accessibility tree.
	b.update([]accessibilityNode{{ID: accessibilityManagedID("button"), Role: "button", Actions: []string{"press"}}}, nil)
	if len(b.gridPublications) != 0 {
		t.Fatal("actual scope removal retained completed grid")
	}
}

func TestAccessibilityGridPublicationPrivateEditorSameGenerationCoherence(t *testing.T) {
	m, a, g, _, _ := inlineFixture(t, 120)
	gridFrame(t, m)
	binding := gBinding(t, g)
	if err := a.Do(func() {
		previous := binding.measurement.Load()
		if previous == nil || previous.Editor == nil {
			t.Fatal("no completed editor fixture")
		}
		generation := binding.projection.State.Generation
		for _, tc := range []struct {
			name   string
			change func(*GridInspection)
		}{
			{"session", func(v *GridInspection) { v.Editor.Session++ }},
			{"identity", func(v *GridInspection) { v.Editor.ID = "retired-editor" }},
			{"focus", func(v *GridInspection) { v.Editor.Focused = !v.Editor.Focused }},
			{"preview", func(v *GridInspection) { v.Editor.Text = "previous preview" }},
		} {
			changed := *previous
			editor := *previous.Editor
			changed.Editor = &editor
			tc.change(&changed)
			binding.measurement.Store(&changed)
			nodes, _ := accessibilitySnapshot(binding.r, nil)
			for _, n := range nodes {
				if !n.GridPending || n.GridGeneration != 0 || n.GridFrame != nil {
					t.Errorf("%s mixed stale editor metadata into completed scope: %+v", tc.name, n)
				}
			}
			binding.measurement.Store(previous)
		}
		binding.editing.visible = false
		nodes, _ := accessibilitySnapshot(binding.r, nil)
		for _, n := range nodes {
			if !n.GridPending {
				t.Errorf("hidden editor reused visible editor measurement: %+v", n)
			}
		}
		binding.editing.visible = true
		nodes, _ = accessibilitySnapshot(binding.r, nil)
		for _, n := range nodes {
			if n.GridPending || n.GridGeneration != generation {
				t.Errorf("restored exact editor measurement remained pending: %+v", n)
			}
		}
	}); err != nil {
		t.Fatal(err)
	}

	var completed []accessibilityNode
	var completedGeneration uint64
	b := newAccessibilityBridge(a)
	b.bind("one")
	t.Cleanup(b.close)
	if err := a.Do(func() {
		var handlers map[string]accessibilityHandler
		completed, handlers = accessibilitySnapshot(binding.r, b)
		b.update(completed, handlers)
		completedGeneration = binding.projection.State.Generation
		edit := gridEditingRecord(binding.r)
		field := edit.widget.(*core.TextField)
		if !accessibilityDeliver(b.token, b.generation.Load(), accessibilityManagedID(edit.node.ID), accessibilityAction{Kind: "set", Value: "old admitted value"}) {
			t.Fatal("old completed editor action was not admitted")
		}
		// Use Core's actual field preedit seam. The marked preview is private:
		// the committed portable draft and generation do not advance yet.
		if !field.ApplyNativeEdit("preedit", 7, 7, 0, 7, "mark") {
			t.Fatal("real field rejected marked fixture")
		}
		cacheFieldEditingState(edit, field.EditingState())
		for _, n := range completed {
			if h := handlers[n.ID]; h != nil && len(n.Actions) > 0 && h(accessibilityAction{Kind: n.Actions[0], Value: "old admitted value"}) {
				t.Errorf("old %s handler accepted before pending publication", n.Role)
			}
		}
		pending, handlers := accessibilitySnapshot(binding.r, nil)
		for _, n := range pending {
			if !n.GridPending {
				t.Errorf("new private preview published against old frame: %+v", n)
			}
		}
		retained, handlers := b.prepareGridPublication(pending, handlers)
		gridAXAssertPending(t, retained, accessibilityManagedID("g"), completedGeneration, completedGeneration, handlers)
		oldEditor := gridAXByID(t, retained)[accessibilityManagedID(edit.node.ID)]
		if oldEditor.Value != binding.measurement.Load().Editor.Text || oldEditor.Value == "preedit" {
			t.Error("pending preedit replaced last-completed text")
		}
		state, e := g.State()
		if e != nil || state.Generation != completedGeneration {
			t.Error("private preedit changed portable generation", state.Generation, e)
		}
		_ = a.Invalidate("one")
	}); err != nil {
		t.Fatal(err)
	}
	// The old callback runs before a new bridge publication. A same-generation
	// native preview must revoke it through the captured handler's live guard.
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if err := a.Do(func() {
		state, e := g.State()
		if e != nil || state.Generation != completedGeneration || state.Edit.Text == "old admitted value" {
			t.Error("queued old editor action changed marked draft before pending publication", state.Edit, e)
		}
	}); err != nil {
		t.Fatal(err)
	}
	gridFrame(t, m)
	if err := a.Do(func() {
		measurement := binding.currentMeasurement()
		if measurement == nil || measurement.Editor == nil || measurement.Editor.Text != "preedit" || measurement.Generation != completedGeneration {
			t.Fatalf("actual completed preview frame missing: %+v", measurement)
		}
		nodes, handlers := accessibilitySnapshot(binding.r, nil)
		for _, n := range nodes {
			if n.GridPending || n.GridGeneration != completedGeneration {
				t.Errorf("completed same-generation preview stayed pending: %+v", n)
			}
		}
		nodes, _ = b.prepareGridPublication(nodes, handlers)
		if gridAXByID(t, nodes)[accessibilityManagedID(measurement.Editor.ID)].Value != "preedit" {
			t.Error("new complete frame did not publish actual preview")
		}
	}); err != nil {
		t.Fatal(err)
	}
}

func TestAccessibilityGridPublicationWithdrawnMeasurementRevokesCapturedHandlers(t *testing.T) {
	m, a, g, _, _ := inlineFixture(t, 120)
	gridFrame(t, m)
	binding := gBinding(t, g)
	if err := a.Do(func() {
		measurement := binding.currentMeasurement()
		if measurement == nil {
			t.Fatal("fixture has no completed frame")
		}
		nodes, handlers := accessibilitySnapshot(binding.r, &accessibilityBridge{})
		if len(handlers) < 3 {
			t.Fatal("fixture lacks host, cell, and editor handlers")
		}
		_ = a.TakeDirty("one")
		binding.withdrawMeasurement()
		if binding.measurement.Load() != nil || !a.TakeDirty("one") {
			t.Error("withdrawal did not request pending publication")
		}
		binding.withdrawMeasurement()
		if a.TakeDirty("one") {
			t.Error("already-nil measurement repeatedly invalidated owner")
		}
		for _, n := range nodes {
			if h := handlers[n.ID]; h != nil && len(n.Actions) > 0 && h(accessibilityAction{Kind: n.Actions[0], Value: "obsolete"}) {
				t.Errorf("withdrawn %s handler accepted before owner publication", n.Role)
			}
		}
		state, e := g.State()
		if e != nil || state.Generation != measurement.Generation {
			t.Error("withdrawal/captured handler mutated portable state", state.Generation, e)
		}
		// Even a nonnil measurement may describe an already retired native binding.
		for _, tc := range []struct {
			name   string
			change func(*GridInspection)
		}{
			{"incarnation", func(v *GridInspection) { v.Incarnation++ }},
			{"cell incarnation", func(v *GridInspection) { v.CellIncarnation++ }},
		} {
			v := *measurement
			tc.change(&v)
			binding.measurement.Store(&v)
			for _, n := range nodes {
				if h := handlers[n.ID]; h != nil && len(n.Actions) > 0 && h(accessibilityAction{Kind: n.Actions[0], Value: "obsolete"}) {
					t.Errorf("%s mismatch admitted %s", tc.name, n.Role)
				}
			}
		}
		binding.measurement.Store(measurement)
		if e := g.Scroll(0, 23); e != nil {
			t.Fatal(e)
		}
		beforeStale, e := g.State()
		if e != nil {
			t.Fatal(e)
		}
		func() {
			originalOnError, originalChange := binding.r.control.onError, binding.r.node.GridChange
			defer func() {
				binding.r.control.onError = originalOnError
				binding.r.node.GridChange = originalChange
			}()
			staleErrors, expectedErrors, notifications := 0, 0, 0
			binding.r.control.onError = func(err error) {
				staleErrors++
				if !errors.Is(err, virtual.ErrGridStale) {
					t.Errorf("captured cell handler reported %v, want ErrGridStale", err)
				}
			}
			binding.r.node.GridChange = func(ui.GridEvent) { notifications++ }
			for _, n := range nodes {
				h := handlers[n.ID]
				if h == nil || len(n.Actions) == 0 {
					continue
				}
				beforeErrors, wantErrors := staleErrors, 0
				if n.Role == "grid-cell" {
					wantErrors = 1
					expectedErrors++
				}
				if h(accessibilityAction{Kind: n.Actions[0], Value: "obsolete"}) {
					t.Errorf("changed owner model admitted old %s before reconciliation", n.Role)
				}
				if got := staleErrors - beforeErrors; got != wantErrors {
					t.Errorf("captured %s handler reported %d errors, want %d", n.Role, got, wantErrors)
				}
			}
			if expectedErrors == 0 || staleErrors != expectedErrors {
				t.Errorf("captured cell stale errors=%d, advertised handlers=%d", staleErrors, expectedErrors)
			}
			afterStale, err := g.State()
			if err != nil || afterStale.Generation != beforeStale.Generation || afterStale.Active != beforeStale.Active || afterStale.Selection != beforeStale.Selection || notifications != 0 {
				t.Errorf("stale handlers changed selection or notified: before=%+v after=%+v notifications=%d error=%v", beforeStale, afterStale, notifications, err)
			}
		}()
		_ = a.Invalidate("one")
	}); err != nil {
		t.Fatal(err)
	}
	gridFrame(t, m)
}

func TestAccessibilityGridPublicationCaptureWithdrawalPaths(t *testing.T) {
	g, _ := testNativeGrid(t, 10000)
	m, a := gridManual(t, manualConfig(t, 7), func() ui.Node { return ui.Node{ID: "g", Kind: ui.DataGrid, Grid: g, Grow: true} })
	gridFrame(t, m)
	binding := gBinding(t, g)
	if err := a.Do(func() {
		measurement := binding.currentMeasurement()
		if measurement == nil || len(measurement.Cells) == 0 {
			t.Fatal("fixture lacks completed cells")
		}
		frames, box := binding.frames, binding.host.Geom.ContentBBox
		viewport := binding.projection.State.Viewport
		defer func() {
			binding.frames = frames
			binding.host.Geom.ContentBBox = box
			binding.projection.State.Viewport = viewport
			binding.measurement.Store(measurement)
		}()
		_ = a.TakeDirty("one")
		binding.frames = math.MaxUint64
		binding.capture()
		if binding.measurement.Load() != nil || !a.TakeDirty("one") {
			t.Error("frame exhaustion silently withdrew last completed publication")
		}
		binding.capture()
		if a.TakeDirty("one") {
			t.Error("frame exhaustion repeated pending invalidation")
		}
		binding.frames = frames
		binding.measurement.Store(measurement)
		// No new viewport post is generated: native request is already equal to
		// the measured rectangle, but the accepted projection has not caught up.
		before := a.Pending().Ordinary
		binding.projection.State.Viewport.Width--
		binding.capture()
		if binding.measurement.Load() != nil || !a.TakeDirty("one") || a.Pending().Ordinary != before {
			t.Error("no-op viewport request failed to publish pending withdrawal")
		}
		binding.capture()
		if a.TakeDirty("one") {
			t.Error("viewport mismatch repeatedly invalidated owner")
		}
		binding.projection.State.Viewport = viewport
		binding.measurement.Store(measurement)
		binding.host.Geom.ContentBBox = image.Rect(0, 0, gridNativeMaxDimension*4, 120)
		binding.geometryRefused = false
		binding.capture()
		if binding.measurement.Load() != nil || !a.TakeDirty("one") {
			t.Error("refused geometry silently withdrew completed frame")
		}
		queued := a.Pending().Ordinary
		binding.capture()
		if a.TakeDirty("one") || a.Pending().Ordinary != queued {
			t.Error("refused geometry repeatedly notified owner/error queue")
		}
	}); err != nil {
		t.Fatal(err)
	}
}
