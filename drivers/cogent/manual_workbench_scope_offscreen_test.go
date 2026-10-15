//go:build offscreen

package cogent

import (
	"errors"
	"testing"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/workbench"
)

func TestManualWorkbenchQueuedInputScopeRevocation(t *testing.T) {
	for _, variant := range []string{"disabled", "hidden-tab", "dialog-dismissed", "modal-opened"} {
		t.Run(variant, func(t *testing.T) {
			a, err := app.New(app.Config{QueueCapacity: 16})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Close()
			model, err := workbench.New(workbench.Layout{Version: workbench.Version, Root: &workbench.Group{ID: "documents", Tabs: []string{"left", "right"}, Active: "left"}}, []string{"left", "right"})
			if err != nil {
				t.Fatal(err)
			}
			dock := &Workbench{Owner: a, Window: "one", Workspace: model, Panels: []Panel{{ID: "left", Title: "Left", View: func() ui.Node { return ui.Node{ID: "label", Kind: ui.Label, Text: "Left"} }}, {ID: "right", Title: "Right", View: func() ui.Node { return ui.Node{ID: "label", Kind: ui.Label, Text: "Right"} }}}}
			blocked := false
			view := func() ui.Node {
				n := ui.Node{ID: "dock", Kind: ui.Extension, Custom: dock, Grow: true}
				switch variant {
				case "disabled":
					n.Disabled = blocked
					return n
				case "hidden-tab":
					selected := "shown"
					if blocked {
						selected = "hidden"
					}
					return ui.Node{ID: "root", Kind: ui.Tabs, SelectedID: selected, Options: []ui.Option{{ID: "shown", Label: "Shown"}, {ID: "hidden", Label: "Hidden"}}, Children: []ui.Node{{ID: "shown", Kind: ui.Column, Children: []ui.Node{n}}, {ID: "hidden", Kind: ui.Label, Text: "Other"}}}
				case "dialog-dismissed":
					return ui.Node{ID: "modal", Kind: ui.Dialog, Open: !blocked, Text: "Dock dialog", Children: []ui.Node{n}}
				default:
					return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{n, {ID: "modal", Kind: ui.Dialog, Open: blocked, Text: "Blocking dialog", Children: []ui.Node{{ID: "message", Kind: ui.Label, Text: "Modal"}}}}}
				}
			}
			if err = a.Open(app.Window{ID: "one", Title: "Scope", Width: 800, Height: 560, View: view}); err != nil {
				t.Fatal(err)
			}
			m, err := NewManual(testContext, a, manualConfig(t, 7))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close(testContext)
			manualDockSettle(t, m, 7)
			if !dock.inputScope.Load().allowed() {
				t.Fatal("initial scope is not interactive")
			}
			// Queue through the real widget listener, withholding owner drain while
			// applying newer reconciled scope states. Routed positives are separate.
			dock.buttons["right"].Send(events.Click)
			if a.Pending().Ordinary == 0 {
				t.Fatal("tab did not enqueue owner input")
			}
			for _, next := range []bool{true, false} {
				blocked = next
				a.Invalidate("one")
				if err = m.implementation.refresh(m.implementation.find("one")); err != nil {
					t.Fatal(err)
				}
			}
			if err = a.Drain(); err != nil {
				t.Fatal(err)
			}
			manualDockSettle(t, m, 7)
			if model.Layout().Root.Active != "left" {
				t.Fatal("old gesture survived scope revocation and reactivation")
			}
			manualDockClick(t, m, 7, "one", ManualWorkbenchTabID("dock", "right"))
			if model.Layout().Root.Active != "right" {
				t.Fatal("fresh interactive gesture refused")
			}
		})
	}
}

func TestManualWorkbenchControlClickRefusesBeforeAdmission(t *testing.T) {
	m, _, _, _ := manualDockFixture(t, 7)
	w, header := manualDockElement(t, m, "one", ManualWorkbenchTabID("dock", "right"))
	before, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	if err = m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Target: header.ID, TargetIncarnation: header.Incarnation, Type: events.MouseDown, Button: events.Left, Modifiers: key.Modifiers(1 << key.Control)}); !errors.Is(err, ErrManualUnsupported) {
		t.Fatal("unqualified context-menu input admitted", err)
	}
	after, err := m.Snapshot(testContext)
	if err != nil || before.Windows[0].Events != after.Windows[0].Events || before.Application != after.Application {
		t.Fatal("context-menu refusal dispatched work", err)
	}
}

func TestManualWorkbenchOwnerClosedBeforeSceneClose(t *testing.T) {
	m, a, dock, _ := manualDockFixture(t, 7)
	w := m.implementation.find("one")
	a.Close()
	w.state.body.Scene.RenderWindow().SystemWindow.CloseReq()
	if err := m.Close(testContext); err != nil {
		t.Fatal("ordinary scene retirement after owner closure became a sticky manual failure", err)
	}
	if len(dock.records) != 0 || dock.isManual() || dock.inputScope.Load() != nil {
		t.Fatal("owner-closed scene retained manual records, lease or input scope")
	}
}

func TestManualWorkbenchConstructionRollbackReleasesMode(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	model, err := workbench.New(workbench.Layout{Version: workbench.Version, Root: &workbench.Group{ID: "group", Tabs: []string{"page"}, Active: "page"}}, []string{"page"})
	if err != nil {
		t.Fatal(err)
	}
	dock := &Workbench{Owner: a, Window: "one", Workspace: model, Panels: []Panel{{ID: "page", Title: "Page", View: func() ui.Node { return ui.Node{ID: "label", Kind: ui.Label, Text: "Retained"} }}}}
	a.Open(app.Window{ID: "one", Title: "First", Width: 500, Height: 400, View: func() ui.Node { return ui.Node{ID: "dock", Kind: ui.Extension, Custom: dock} }})
	unsafe := true
	a.Open(app.Window{ID: "two", Title: "Second", Width: 500, Height: 400, View: func() ui.Node {
		if unsafe {
			return ui.Node{ID: "unknown", Kind: ui.Extension, Custom: struct{}{}}
		}
		return ui.Node{ID: "label", Kind: ui.Label, Text: "Second"}
	}})
	config := manualConfig(t, 7)
	if m, err := NewManual(testContext, a, config); m != nil || !errors.Is(err, ErrManualUnsupported) {
		if m != nil {
			m.Close(testContext)
		}
		t.Fatal("construction did not refuse second window", err)
	}
	if dock.isManual() || dock.manualViews != nil || a.Context().Err() != nil || len(a.Windows()) != 2 {
		t.Fatal("rollback retained mode, views or damaged domain")
	}
	unsafe = false
	m, err := NewManual(testContext, a, config)
	if err != nil {
		t.Fatal("rolled-back workbench could not be retried", err)
	}
	defer m.Close(testContext)
	manualDockSettle(t, m, 7)
}
