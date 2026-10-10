package main

import (
	"context"
	"fmt"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/app"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/workbench"
)

func namedElement(s cogent.AccessibilityInspection, name string) (cogent.AccessibilityElement, bool) {
	for _, element := range s.Elements {
		if element.Name == name {
			return element, true
		}
	}
	return cogent.AccessibilityElement{}, false
}
func dockAction(ctx context.Context, id, kind, value string) error {
	accepted, err := cogent.PerformAccessibilityAction(ctx, "ax-dock", id, kind, value)
	if err != nil {
		return err
	}
	if !accepted {
		return fmt.Errorf("dock native accessibility refused %s %s", id, kind)
	}
	return nil
}
func checkWorkbench(ctx context.Context, a *app.Application) (result error) {
	stage := "dock startup modal flush and geometry"
	defer func() {
		if result != nil {
			result = fmt.Errorf("%s: %w", stage, result)
		}
	}()
	values := map[string]string{"one": "One", "two": "Two", "three": "Three"}
	modal := "one"
	workspace, err := workbench.New(workbench.Layout{Version: workbench.Version, Root: &workbench.Group{ID: "split", Axis: workbench.Horizontal, Ratio: .5, First: &workbench.Group{ID: "left", Tabs: []string{"one", "two"}, Active: "one"}, Second: &workbench.Group{ID: "right", Tabs: []string{"three"}, Active: "three"}}}, []string{"one", "two", "three"})
	if err != nil {
		return err
	}
	panels := []cogent.Panel{}
	for _, id := range []string{"one", "two", "three"} {
		id := id
		title := values[id]
		panels = append(panels, cogent.Panel{ID: id, Title: title, View: func() ui.Node {
			return ui.Node{ID: "panel", Kind: ui.Column, Children: []ui.Node{
				{ID: "shared", Kind: ui.TextField, Text: values[id], Accessibility: ui.Accessibility{Name: title + " field"}, Change: func(e ui.ChangeEvent) { values[id] = e.Text }},
				{ID: "show-modal", Kind: ui.Button, Text: title + " preferences", Activate: func() { modal = id }},
				{ID: "modal", Kind: ui.Dialog, Text: title + " preferences", Open: modal == id, Dismiss: func() { modal = "" }, Children: []ui.Node{{ID: "modal-field", Kind: ui.TextField, Text: values[id], Accessibility: ui.Accessibility{Name: title + " modal field"}, Change: func(e ui.ChangeEvent) { values[id] = e.Text }}}},
			}}
		}})
	}
	dock := &cogent.Workbench{Owner: a, Window: "ax-dock", Workspace: workspace, Panels: panels}
	if err := a.Do(func() {
		err = a.Open(app.Window{ID: "ax-dock", Title: "Vellum accessibility docking smoke", Width: 800, Height: 500, View: func() ui.Node { return ui.Node{ID: "workbench", Kind: ui.Extension, Custom: dock, Grow: true} }})
	}); err != nil {
		return err
	}
	if err != nil {
		return err
	}
	initialModal, err := waitWindow(ctx, "ax-dock", func(s cogent.AccessibilityInspection) bool {
		x, visible := namedElement(s, "One modal field")
		return visible && x.Frame[2] > 0 && x.HitTestID == x.ID
	})
	if err != nil {
		return err
	}
	initialField, _ := namedElement(initialModal, "One modal field")
	if err = dockAction(ctx, initialField.ID, "set", "One initial modal accepted"); err != nil {
		return err
	}
	_, err = waitWindow(ctx, "ax-dock", func(s cogent.AccessibilityInspection) bool {
		x, _ := namedElement(s, "One modal field")
		return x.Value == "One initial modal accepted"
	})
	if err != nil {
		return err
	}
	// Window registration precedes SetName in native startup. Query by name
	// only after the actual live modal probe and accepted edit prove startup.
	native := core.TheApp.WindowByName("Vellum accessibility docking smoke")
	if native == nil {
		return fmt.Errorf("dock native window unavailable")
	}
	native.Events().KeyChord(0, key.CodeEscape, 0)
	stage = "dock initial visible panel tree"
	first, err := waitWindow(ctx, "ax-dock", func(s cogent.AccessibilityInspection) bool {
		x, ok := namedElement(s, "One field")
		y, yes := namedElement(s, "Three field")
		return ok && yes && x.Frame[2] > 0 && y.Frame[2] > 0
	})
	if err != nil {
		return err
	}
	if first.WindowRole != "AXWindow" || !first.NativeWindow {
		return fmt.Errorf("dock native window identity lost: %+v", first)
	}
	one, _ := namedElement(first, "One field")
	three, _ := namedElement(first, "Three field")
	if one.ID == three.ID || one.DomainID != "shared" || three.DomainID != "shared" {
		return fmt.Errorf("panel-local native identifiers collide")
	}
	if _, ok := namedElement(first, "Two field"); ok {
		return fmt.Errorf("inactive dock panel accessible")
	}
	stage = "dock controlled edit"
	if err = dockAction(ctx, one.ID, "set", "One updated"); err != nil {
		return err
	}
	_, err = waitWindow(ctx, "ax-dock", func(s cogent.AccessibilityInspection) bool {
		x, _ := namedElement(s, "One field")
		return x.Value == "One updated"
	})
	if err != nil {
		return err
	}
	stage = "dock native focus and Tab"
	if err = dockAction(ctx, one.ID, "focus", ""); err != nil {
		return err
	}
	_, err = waitWindow(ctx, "ax-dock", func(s cogent.AccessibilityInspection) bool {
		x, ok := namedElement(s, "One field")
		return ok && x.ID != "" && x.ID == one.ID && s.FocusedID == x.ID
	})
	if err != nil {
		return err
	}
	if native == nil {
		return fmt.Errorf("dock native window unavailable")
	}
	native.Events().KeyChord('\t', key.CodeTab, 0)
	_, err = waitWindow(ctx, "ax-dock", func(s cogent.AccessibilityInspection) bool {
		x, ok := namedElement(s, "One preferences")
		return ok && x.ID != "" && s.FocusedID == x.ID
	})
	if err != nil {
		return err
	}
	stage = "dock panel activation"
	choice, _ := namedElement(first, "Two")
	if choice.ID == "" {
		return fmt.Errorf("dock choice missing")
	}
	if err = dockAction(ctx, choice.ID, "press", ""); err != nil {
		return err
	}
	second, err := waitWindow(ctx, "ax-dock", func(s cogent.AccessibilityInspection) bool {
		_, visible := namedElement(s, "Two field")
		_, hidden := namedElement(s, "One field")
		return visible && !hidden
	})
	if err != nil {
		return err
	}
	stage = "dock modal open and geometry"
	show, _ := namedElement(second, "Two preferences")
	if err = dockAction(ctx, show.ID, "press", ""); err != nil {
		return err
	}
	modalTree, err := waitWindow(ctx, "ax-dock", func(s cogent.AccessibilityInspection) bool {
		x, ok := namedElement(s, "Two modal field")
		_, outside := namedElement(s, "Three field")
		return ok && !outside && x.Frame[2] > 0 && x.HitTestID == x.ID
	})
	if err != nil {
		return err
	}
	modalField, _ := namedElement(modalTree, "Two modal field")
	stage = "dock modal focus edit and Escape"
	if err = dockAction(ctx, modalField.ID, "focus", ""); err != nil {
		return err
	}
	_, err = waitWindow(ctx, "ax-dock", func(s cogent.AccessibilityInspection) bool {
		x, ok := namedElement(s, "Two modal field")
		return ok && x.ID != "" && x.ID == modalField.ID && s.FocusedID == x.ID
	})
	if err != nil {
		return err
	}
	if err = dockAction(ctx, modalField.ID, "set", "Two modal accepted"); err != nil {
		return err
	}
	_, err = waitWindow(ctx, "ax-dock", func(s cogent.AccessibilityInspection) bool {
		x, _ := namedElement(s, "Two modal field")
		return x.Value == "Two modal accepted"
	})
	if err != nil {
		return err
	}
	native.Events().KeyChord(0, key.CodeEscape, 0)
	_, err = waitWindow(ctx, "ax-dock", func(s cogent.AccessibilityInspection) bool {
		_, modal := namedElement(s, "Two modal field")
		x, normal := namedElement(s, "Two field")
		return !modal && normal && x.Value == "Two modal accepted"
	})
	if err != nil {
		return err
	}
	stage = "dock owner state and route retirement"
	var ownerErr error
	if err = a.Do(func() {
		if values["one"] != "One updated" || values["two"] != "Two modal accepted" || modal != "" {
			ownerErr = fmt.Errorf("dock controlled state incorrect")
		}
		if ownerErr == nil {
			ownerErr = a.CloseWindow("ax-dock")
		}
	}); err != nil {
		return err
	}
	if ownerErr != nil {
		return ownerErr
	}
	for {
		_, err = cogent.InspectAccessibility(ctx, "ax-dock")
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(20 * time.Millisecond):
		}
	}
	if accepted, err := cogent.PerformAccessibilityAction(ctx, "ax-dock", modalField.ID, "set", "late"); accepted || err == nil {
		return fmt.Errorf("closed dock route accepted late native action")
	}
	return nil
}
