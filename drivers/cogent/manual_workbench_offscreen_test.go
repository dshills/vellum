//go:build offscreen

package cogent

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"github.com/dshills/vellum/workbench"
)

func manualDockFixture(t *testing.T, batch int) (*Manual, *app.Application, *Workbench, map[string]string) {
	t.Helper()
	a, err := app.New(app.Config{QueueCapacity: 16})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(a.Close)
	model, err := workbench.New(workbench.Layout{Version: workbench.Version, Root: &workbench.Group{ID: "documents", Tabs: []string{"left", "right"}, Active: "left"}}, []string{"left", "right"})
	if err != nil {
		t.Fatal(err)
	}
	dock := &Workbench{Owner: a, Window: "one", Workspace: model}
	values := map[string]string{"left": "Left", "right": "Right"}
	calls := map[string]int{}
	for _, panel := range []string{"left", "right"} {
		panel := panel
		dock.Panels = append(dock.Panels, Panel{ID: panel, Title: panel, View: func() ui.Node {
			calls[panel]++
			return ui.Node{ID: "page", Kind: ui.Column, Grow: true, Children: []ui.Node{{ID: "name", Kind: ui.TextField, Text: values[panel], Accessibility: ui.Accessibility{Name: panel + " name"}, Change: func(e ui.ChangeEvent) { values[panel] = e.Text; a.Invalidate("one") }}}}
		}})
	}
	if err = a.Open(app.Window{ID: "one", Title: "Dock replay", Width: 900, Height: 620, View: func() ui.Node { return ui.Node{ID: "dock", Kind: ui.Extension, Custom: dock, Grow: true} }}); err != nil {
		t.Fatal(err)
	}
	m, err := NewManual(testContext, a, manualConfig(t, batch))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := m.Close(testContext); err != nil {
			t.Error(err)
		}
	})
	if calls["left"] != 1 || calls["right"] != 1 {
		t.Fatalf("panel factory evaluated more than once during initial reconciliation: %v", calls)
	}
	if dock.manualViews != nil {
		t.Fatal("captured panel views retained")
	}
	manualDockSettle(t, m, batch)
	return m, a, dock, values
}

func TestManualWorkbenchRebuildRevokesHeadersAndRetainsDraft(t *testing.T) {
	m, a, dock, _ := manualDockFixture(t, 7)
	leftID := ManualWorkbenchControlID("dock", "left", "name")
	manualDockClick(t, m, 7, "one", leftID)
	w, field := manualDockElement(t, m, "one", leftID)
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Rune: 'x', Code: key.CodeX}); err != nil {
		t.Fatal(err)
	}
	manualDockSettle(t, m, 7)
	_, draft := manualDockElement(t, m, "one", leftID)
	oldWidget := dock.records["left"].children[0].widget
	_, header := manualDockElement(t, m, "one", ManualWorkbenchTabID("dock", "left"))
	if err := a.Do(func() {
		if err := dock.Workspace.Move("right", "documents", 0); err != nil {
			t.Error(err)
		}
		a.Invalidate("one")
	}); err != nil {
		t.Fatal(err)
	}
	manualDockSettle(t, m, 7)
	newWindow, newHeader := manualDockElement(t, m, "one", header.ID)
	_, after := manualDockElement(t, m, "one", leftID)
	if header.Incarnation == newHeader.Incarnation || field.Incarnation != after.Incarnation || dock.records["left"].children[0].widget != oldWidget || after.Value != draft.Value {
		t.Fatal("header/draft lifetime not preserved correctly", header, newHeader, draft, after)
	}
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: newWindow.Incarnation, Target: header.ID, TargetIncarnation: header.Incarnation, Type: events.MouseDown, Button: events.Left}); !errors.Is(err, ErrManualStale) {
		t.Fatal("destroyed header target admitted", err)
	}
}

func TestManualWorkbenchGroupIndependentWindowsAndExplicitPolicies(t *testing.T) {
	a, group := groupFixture(t)
	m, err := NewManual(testContext, a, manualConfig(t, 7))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close(testContext)
	manualDockSettle(t, m, 7)
	mainID, ok := group.WindowID("main")
	if !ok {
		t.Fatal("missing main lifetime")
	}
	main := group.windows["main"]
	if !main.isManual() || main.boundsQueued.Load() || main.pendingBounds.Load() != nil {
		t.Fatal("manual mode installed native bounds work")
	}
	main.queueWindowBounds(workbench.Bounds{X: 90, Y: 100, Width: 900, Height: 640, Positioned: true})
	if main.boundsQueued.Load() || main.pendingBounds.Load() != nil {
		t.Fatal("manual path scheduled a desktop bounds timer")
	}
	before, _ := group.Session.Encode()
	path := filepath.Join(t.TempDir(), "session.json")
	if err = group.Save(path); !errors.Is(err, ErrManualUnsupported) {
		t.Fatal("manual save admitted", err)
	}
	if err = group.Load(path); !errors.Is(err, ErrManualUnsupported) {
		t.Fatal("manual load admitted", err)
	}
	group.PersistencePath = path
	if err = group.Apply(func(s *workbench.Session) error { return s.Activate("main", "documents", "two") }); !errors.Is(err, ErrManualUnsupported) {
		t.Fatal("late autosave admitted", err)
	}
	group.PersistencePath = ""
	if _, err = os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("manual persistence wrote a file", err)
	}
	after, _ := group.Session.Encode()
	if !reflect.DeepEqual(before, after) {
		t.Fatal("refused policy mutated session")
	}
	if err = group.Apply(func(s *workbench.Session) error {
		return s.SetBounds("main", workbench.Bounds{Width: 800, Height: 600})
	}); !errors.Is(err, ErrManualUnsupported) {
		t.Fatal("implicit viewport resize admitted", err)
	}
	originalTwo := main.records["two"].widget
	if err = group.Float("two", workbench.Bounds{Width: 500, Height: 400}); err != nil {
		t.Fatal(err)
	}
	manualDockSettle(t, m, 7)
	if len(a.Windows()) != 2 {
		t.Fatal("floating did not register a separate manual window")
	}
	var floating string
	for _, w := range group.Session.Layout().Windows {
		if w.ID != "main" {
			floating = w.ID
		}
	}
	floatID, _ := group.WindowID(floating)
	snapshot, err := m.Snapshot(testContext)
	if err != nil || len(snapshot.Windows) != 2 {
		t.Fatal("missing completed independent window", snapshot, err)
	}
	if mainID == floatID || group.windows[floating].records["two"].widget == originalTwo {
		t.Fatal("native records crossed window ownership")
	}
	// Exercise the scene close request; Core.ManualWindow.Close is forced cleanup.
	closing := m.implementation.find(floatID)
	closing.state.body.Scene.RenderWindow().SystemWindow.CloseReq()
	manualDockSettle(t, m, 7)
	if _, exists := group.WindowID(floating); exists || len(a.Windows()) != 1 || len(group.Session.Layout().Windows) != 1 {
		t.Fatal("native close did not notify session before unregistering")
	}
	if _, err = m.Capture(testContext, mainID); err != nil {
		t.Fatal("closing sibling damaged survivor", err)
	}
}

func TestManualWorkbenchStoredGridScopedInspectionAndRoutedScroll(t *testing.T) {
	a, err := app.New(app.Config{QueueCapacity: 32})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	source := &replayGridSource{replayGridIndex: replayGridIndex{rows: 10000, revision: 1}, data: make([][20]uint32, 10000)}
	for row := range source.data {
		for column := range 20 {
			source.data[row][column] = uint32(row*20 + column)
		}
	}
	columns := make([]virtual.GridColumn, 20)
	for i := range columns {
		columns[i] = virtual.GridColumn{ID: fmt.Sprintf("c%d", i), Title: fmt.Sprint(i), Width: 90}
	}
	grid, err := virtual.NewGrid(source, columns, virtual.GridConfig{RowHeight: 24, HeaderHeight: 26}, virtual.GridViewport{Width: 700, Height: 400})
	if err != nil {
		t.Fatal(err)
	}
	model, err := workbench.New(workbench.Layout{Version: workbench.Version, Root: &workbench.Group{ID: "documents", Tabs: []string{"data"}, Active: "data"}}, []string{"data"})
	if err != nil {
		t.Fatal(err)
	}
	dock := &Workbench{Owner: a, Window: "one", Workspace: model}
	dock.Panels = []Panel{{ID: "data", Title: "Data", View: func() ui.Node {
		return ui.Node{ID: "grid", Kind: ui.DataGrid, Grid: grid, Grow: true, GridChange: func(change ui.GridEvent) { a.Invalidate("one") }}
	}}}

	if err = a.Open(app.Window{ID: "one", Title: "Stored dock grid", Width: 900, Height: 620, View: func() ui.Node { return ui.Node{ID: "dock", Kind: ui.Extension, Custom: dock, Grow: true} }}); err != nil {
		t.Fatal(err)
	}
	config := manualConfig(t, 7)
	config.LogicalDPI = 96
	m, err := NewManual(testContext, a, config)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close(testContext)
	manualDockSettle(t, m, 7)
	s, err := m.Snapshot(testContext)
	if err != nil || len(s.Windows[0].Grids) != 1 {
		t.Fatal("scoped grid missing", s, err)
	}
	inspection := s.Windows[0].Grids[0]
	qualified := ManualWorkbenchControlID("dock", "data", "grid")
	if inspection.ID != qualified || len(inspection.Cells) == 0 {
		t.Fatal("grid inspection identity missing", inspection)
	}
	var cell GridCellGeometry
	for _, c := range inspection.Cells {
		if !c.Bounds.Empty() {
			cell = c
			break
		}
	}
	manualDockClick(t, m, 7, "one", cell.ID)
	window, target := manualDockElement(t, m, "one", qualified)
	if err = m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: window.Incarnation, Target: target.ID, TargetIncarnation: target.Incarnation, Type: events.Scroll, Delta: math32.Vec2(0, 48)}); err != nil {
		t.Fatal(err)
	}
	manualDockSettle(t, m, 7)
	raw, err := InspectGrid(a, grid)
	if err != nil || raw.ID != "grid" || raw.Viewport.ScrollY != 48 {
		t.Fatal("routed scroll or detached scoped inspection corrupted binding", raw.ID, raw.Viewport, err)
	}
	for _, c := range raw.Cells {
		if c.ID != GridCellID("grid", c.RowKey, c.ColumnID) {
			t.Fatal("manual inspection mutated original cell ID")
		}
	}
}

func manualDockSettle(t *testing.T, m *Manual, batch int) {
	t.Helper()
	for range 4096 {
		s, err := m.Snapshot(testContext)
		if err != nil {
			t.Fatal(err)
		}
		ready := s.Application.Ordinary+s.Application.ReadyCompletions > 0
		for _, w := range s.Windows {
			ready = ready || w.Events > 0 || w.ReadyPaint
			if !ready && w.LayoutPending {
				if err = m.RequestFrame(testContext, w.ID); err != nil {
					t.Fatal(err)
				}
				ready = true
			}
		}
		if !ready {
			return
		}
		work, err := m.Step(testContext, batch)
		if err != nil {
			t.Fatal(err)
		}
		if work.Events+work.OwnerCallbacks > batch {
			t.Fatal("dispatch budget exceeded", work)
		}
	}
	t.Fatal("dock did not settle within explicit dispatch budget")
}

func manualDockElement(t *testing.T, m *Manual, window, id string) (ManualWindowSnapshot, ManualElement) {
	t.Helper()
	s, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range s.Windows {
		if w.ID == window {
			for _, e := range w.Elements {
				if e.ID == id {
					return w, e
				}
			}
		}
	}
	t.Fatalf("missing dock target %q in %q", id, window)
	return ManualWindowSnapshot{}, ManualElement{}
}
func manualDockClick(t *testing.T, m *Manual, batch int, window, id string) {
	t.Helper()
	w, e := manualDockElement(t, m, window, id)
	if !e.Visible || e.Bounds.Empty() || e.Incarnation == 0 {
		t.Fatalf("invalid routed target: %+v", e)
	}
	for _, event := range []events.Types{events.MouseDown, events.MouseUp} {
		if err := m.Input(testContext, ManualInput{Window: window, WindowIncarnation: w.Incarnation, Target: id, TargetIncarnation: e.Incarnation, Type: event, Button: events.Left}); err != nil {
			t.Fatal(err)
		}
	}
	manualDockSettle(t, m, batch)
}

func TestManualWorkbenchRoutedTabsAcrossBatches(t *testing.T) {
	var oracle []byte
	for _, batch := range []int{1, 7, 64} {
		t.Run(fmt.Sprint(batch), func(t *testing.T) {
			m, _, dock, values := manualDockFixture(t, batch)
			leftID := accessibilityWorkbenchNamespace("dock", "left") + "name"
			rightID := accessibilityWorkbenchNamespace("dock", "right") + "name"
			w, left := manualDockElement(t, m, "one", leftID)
			_, right := manualDockElement(t, m, "one", rightID)
			if !left.Visible || right.Visible || left.Incarnation == 0 || right.Incarnation == 0 || left.ID == right.ID {
				t.Fatal("panel namespace/visibility missing", left, right)
			}
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Target: rightID, TargetIncarnation: right.Incarnation, Type: events.MouseDown, Button: events.Left}); !errors.Is(err, ErrManualDisabled) {
				t.Fatal("inactive panel admitted", err)
			}
			manualDockClick(t, m, batch, "one", leftID)
			w, _ = manualDockElement(t, m, "one", leftID)
			for _, input := range []ManualInput{{Type: events.KeyChord, Code: key.CodeA, Modifiers: key.Modifiers(1 << key.Control)}, {Type: events.KeyChord, Rune: 'Z', Code: key.CodeZ}, {Type: events.KeyChord, Code: key.CodeReturnEnter}} {
				input.Window = "one"
				input.WindowIncarnation = w.Incarnation
				if err := m.Input(testContext, input); err != nil {
					t.Fatal(err)
				}
				manualDockSettle(t, m, batch)
			}
			if values["left"] != "Z" || values["right"] != "Right" {
				t.Fatal("actual field route did not edit only left", values)
			}
			manualDockClick(t, m, batch, "one", accessibilityDockOptionID("dock", "right"))
			if dock.Workspace.Layout().Root.Active != "right" {
				t.Fatal("native tab input bypassed or failed owner transition")
			}
			_, hidden := manualDockElement(t, m, "one", leftID)
			_, visible := manualDockElement(t, m, "one", rightID)
			if hidden.Visible || !visible.Visible || hidden.Incarnation != left.Incarnation || visible.Incarnation != right.Incarnation {
				t.Fatal("hidden records lost identity or visibility", hidden, visible)
			}
			image, err := m.Capture(testContext, "one")
			if err != nil {
				t.Fatal(err)
			}
			digest := sha256.Sum256(image.Pix)
			s, err := m.Snapshot(testContext)
			if err != nil {
				t.Fatal(err)
			}
			elements := s.Windows[0].Elements
			for i := range elements {
				elements[i].Incarnation = 0
			}
			state := struct {
				Values   map[string]string
				Elements []ManualElement
				Image    string
			}{values, elements, hex.EncodeToString(digest[:])}
			data, err := json.Marshal(state)
			if err != nil {
				t.Fatal(err)
			}
			if oracle == nil {
				oracle = data
			} else if !reflect.DeepEqual(oracle, data) {
				t.Fatal("dock model/geometry/semantics/image changed with batching")
			}
		})
	}
}

func TestManualWorkbenchRefusesWholeHostedGraphBeforeMount(t *testing.T) {
	for _, variant := range []string{"inactive-extension", "recursive", "budget", "foreign-owner", "reserved-id", "shared-model", "depth"} {
		t.Run(variant, func(t *testing.T) {
			a, err := app.New(app.Config{})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Close()
			model, err := workbench.New(workbench.Layout{Version: workbench.Version, Root: &workbench.Group{ID: "documents", Tabs: []string{"left", "right"}, Active: "left"}}, []string{"left", "right"})
			if err != nil {
				t.Fatal(err)
			}
			dock := &Workbench{Owner: a, Window: "one", Workspace: model}
			dock.Panels = []Panel{{ID: "left", Title: "Left", View: func() ui.Node { return ui.Node{ID: "name", Kind: ui.Label} }}, {ID: "right", Title: "Right", View: func() ui.Node { return ui.Node{ID: "name", Kind: ui.Label} }}}
			config := manualConfig(t, 7)
			switch variant {
			case "inactive-extension":
				dock.Panels[1].View = func() ui.Node { return ui.Node{ID: "unsafe", Kind: ui.Extension, Custom: struct{}{}} }
			case "recursive":
				dock.Panels[1].View = func() ui.Node { return ui.Node{ID: "recursive", Kind: ui.Extension, Custom: dock} }
			case "budget":
				config.MaxControls = 4
			case "foreign-owner":
				dock.Window = "other"
			case "reserved-id":
				dock.Panels[1].View = func() ui.Node { return ui.Node{ID: "workbench:collision", Kind: ui.Label} }
			case "shared-model":
				grid, _ := testNativeGrid(t, 10000)
				for i := range dock.Panels {
					dock.Panels[i].View = func() ui.Node { return ui.Node{ID: "grid", Kind: ui.DataGrid, Grid: grid} }
				}
			case "depth":
				dock.Panels[1].View = func() ui.Node {
					n := ui.Node{ID: "leaf", Kind: ui.Label}
					for i := range 128 {
						n = ui.Node{ID: fmt.Sprintf("level%d", i), Kind: ui.Column, Children: []ui.Node{n}}
					}
					return n
				}
			}
			if err = a.Open(app.Window{ID: "one", Title: "refusal", Width: 500, Height: 400, View: func() ui.Node { return ui.Node{ID: "dock", Kind: ui.Extension, Custom: dock} }}); err != nil {
				t.Fatal(err)
			}
			m, err := NewManual(testContext, a, config)
			if m != nil || err == nil {
				if m != nil {
					m.Close(testContext)
				}
				t.Fatal("unsafe dock graph admitted")
			}
			if dock.host != nil || len(dock.records) != 0 || dock.manualViews != nil || a.Context().Err() != nil || len(a.Windows()) != 1 {
				t.Fatal("refusal mounted or damaged domain")
			}
		})
	}
}

func TestManualWorkbenchNestedQualifiedRoutingAndHiddenLifetime(t *testing.T) {
	a, err := app.New(app.Config{QueueCapacity: 16})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	newModel := func() *workbench.Workspace {
		model, err := workbench.New(workbench.Layout{Version: workbench.Version, Root: &workbench.Group{ID: "documents", Tabs: []string{"left", "right"}, Active: "left"}}, []string{"left", "right"})
		if err != nil {
			t.Fatal(err)
		}
		return model
	}
	inner := &Workbench{Owner: a, Window: "one", Workspace: newModel()}
	activated := 0
	inner.Panels = []Panel{{ID: "left", Title: "Left", View: func() ui.Node {
		return ui.Node{ID: "action", Kind: ui.Button, Text: "Nested", Activate: func() { activated++; a.Invalidate("one") }}
	}}, {ID: "right", Title: "Right", View: func() ui.Node { return ui.Node{ID: "action", Kind: ui.Label, Text: "Other"} }}}
	outer := &Workbench{Owner: a, Window: "one", Workspace: newModel()}
	outer.Panels = []Panel{{ID: "left", Title: "Left", View: func() ui.Node { return ui.Node{ID: "inner", Kind: ui.Extension, Custom: inner, Grow: true} }}, {ID: "right", Title: "Right", View: func() ui.Node { return ui.Node{ID: "action", Kind: ui.Label, Text: "Outside"} }}}
	if err = a.Open(app.Window{ID: "one", Title: "Nested", Width: 900, Height: 620, View: func() ui.Node { return ui.Node{ID: "outer", Kind: ui.Extension, Custom: outer, Grow: true} }}); err != nil {
		t.Fatal(err)
	}
	m, err := NewManual(testContext, a, manualConfig(t, 7))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close(testContext)
	manualDockSettle(t, m, 7)
	target := ManualWorkbenchControlID("outer", "left", ManualWorkbenchControlID("inner", "left", "action"))
	manualDockClick(t, m, 7, "one", target)
	if activated != 1 {
		t.Fatal("nested panel route missed actual callback", activated)
	}
	_, original := manualDockElement(t, m, "one", target)
	manualDockClick(t, m, 7, "one", ManualWorkbenchTabID("outer", "right"))
	w, hidden := manualDockElement(t, m, "one", target)
	if hidden.Visible || hidden.Incarnation != original.Incarnation {
		t.Fatal("hidden nested lifetime lost", hidden, original)
	}
	if err = m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Target: target, TargetIncarnation: hidden.Incarnation, Type: events.MouseDown, Button: events.Left}); !errors.Is(err, ErrManualDisabled) {
		t.Fatal("hidden nested input admitted", err)
	}
	manualDockClick(t, m, 7, "one", ManualWorkbenchTabID("outer", "left"))
	manualDockClick(t, m, 7, "one", target)
	if activated != 2 {
		t.Fatal("nested route did not recover", activated)
	}
}
