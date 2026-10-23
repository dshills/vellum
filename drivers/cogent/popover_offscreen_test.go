//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestPopoverAnchoringEscapeRetentionAndRemoval(t *testing.T) {
	open, show := false, true
	dismissed := 0
	view := func() ui.Node {
		p, err := (ui.Popover{ID: "panel", AnchorID: "anchor", Label: "Inspector", Open: open, Dismiss: func() { dismissed++; open = false }, Children: []ui.Node{{ID: "value", Kind: ui.TextField, Text: "Accepted"}, {ID: "done", Kind: ui.Button, Text: "Done", Activate: func() { open = false }}}}).View()
		if err != nil {
			panic(err)
		}
		children := []ui.Node{}
		if show {
			children = append(children, ui.Node{ID: "anchor", Kind: ui.Button, Text: "Inspector", Activate: func() { open = true }})
		}
		return ui.Node{ID: "root", Kind: ui.Column, Children: append(children, p)}
	}
	a, err := app.New(app.Config{QueueCapacity: 32})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(a.Close)
	if err = a.Open(app.Window{ID: "one", Title: "Popover", Width: 600, Height: 600, View: view}); err != nil {
		t.Fatal(err)
	}
	m, err := NewManual(testContext, a, manualConfig(t, 917))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := m.Close(testContext); err != nil {
			t.Error(err)
		}
	})
	unitSettle(t, m)
	w, anchor := imageManualElement(t, m, "anchor")
	click := func(target ManualElement) {
		t.Helper()
		for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: typ, Target: target.ID, TargetIncarnation: target.Incarnation, Button: events.Left}); err != nil {
				t.Fatal(err)
			}
		}
		unitSettle(t, m)
	}
	click(anchor)
	root := m.implementation.find("one").state.root
	r := findRecord(root, "panel")
	value := findRecord(root, "value").widget
	if !r.control.dialog.active || r.control.dialog.body == nil {
		t.Fatal("popover not active")
	}
	stage := r.control.dialog.body.Scene.Stage
	if stage.Scrim || stage.DisplayTitle || stage.Resizable || stage.NewWindow || stage.Title != "" || stage.Scene.Parts != nil {
		t.Fatal("popover has window/dialog chrome")
	}
	rect := r.control.dialog.body.Scene.SceneGeom.Bounds()
	hostSize := r.control.dialog.host.Geom.Size.Actual.Total.Add(stage.Scene.Geom.Size.Space).Add(r.control.dialog.body.Geom.Size.Space).ToPointCeil()
	if rect.Size() != hostSize {
		t.Fatal("popover retains generic dialog outer allowance", rect.Size(), hostSize)
	}
	if stage.Scene.HasScroll[0] || stage.Scene.HasScroll[1] || r.control.dialog.body.HasScroll[0] || r.control.dialog.body.HasScroll[1] {
		t.Fatal("compact popover has spurious scrollbars")
	}
	if rect.Min.Y+int(stage.Scene.Styles.TotalMargin().Top) < anchor.Bounds.Max.Y {
		t.Fatal("not anchored below", rect, anchor.Bounds)
	}
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeEscape}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if open || dismissed != 1 || r.control.dialog.active {
		t.Fatal("Escape did not dismiss", open, dismissed)
	}
	if !findRecord(root, "anchor").widget.AsWidget().ContainsFocus() {
		t.Fatal("anchor focus not restored")
	}
	_, anchor = imageManualElement(t, m, "anchor")
	click(anchor)
	if !r.control.dialog.active || findRecord(root, "value").widget != value {
		t.Fatal("reopen recreated content")
	}
	if err := a.Do(func() { show = false; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if open || r.control.dialog.active || dismissed != 2 {
		t.Fatal("removed anchor left panel", open, dismissed)
	}
}

func TestPopoverInitiallyOpenAnchorValidation(t *testing.T) {
	for _, present := range []bool{true, false} {
		t.Run(map[bool]string{true: "present", false: "missing"}[present], func(t *testing.T) {
			open := true
			dismissed := 0
			view := func() ui.Node {
				p, err := (ui.Popover{ID: "panel", AnchorID: "anchor", Open: open, Dismiss: func() { open = false; dismissed++ }, Children: []ui.Node{{ID: "value", Kind: ui.Button, Text: "Value"}}}).View()
				if err != nil {
					panic(err)
				}
				children := []ui.Node{p}
				if present {
					children = append([]ui.Node{{ID: "anchor", Kind: ui.Button, Text: "Anchor"}}, children...)
				}
				return ui.Node{ID: "root", Kind: ui.Column, Children: children}
			}
			m, _ := simpleManual(t, manualConfig(t, 918), view)
			unitSettle(t, m)
			r := findRecord(m.implementation.find("one").state.root, "panel")
			if present {
				if !open || dismissed != 0 || !r.control.dialog.active {
					t.Fatal("initial valid anchor canceled", open, dismissed)
				}
			} else {
				if open || dismissed != 1 || r.control.dialog.active {
					t.Fatal("missing initial anchor not canceled", open, dismissed)
				}
			}
		})
	}
}
