package cogent

import (
	"math"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/tree"
	"example.com/vellum/app"
	"example.com/vellum/ui"
)

func TestNativeControlsChangeRetainedCallbacksAndDisabled(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	if err := a.Open(app.Window{ID: "main", Title: "headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	host := core.NewFrame(core.NewBody("headless"))
	var changes []ui.ChangeEvent
	callback := func(e ui.ChangeEvent) { changes = append(changes, e) }
	n := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{
		{ID: "check", Kind: ui.Checkbox, Text: "Check", Change: callback},
		{ID: "select", Kind: ui.Selector, SelectedID: "a", Options: []ui.Option{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}, {ID: "disabled", Label: "Disabled", Disabled: true}}, Change: callback},
		{ID: "slider", Kind: ui.Slider, Number: 4, Numeric: ui.Numeric{Min: 0, Max: 10, Step: 1}, Change: callback},
		{ID: "number", Kind: ui.NumberField, Number: 4, Numeric: ui.Numeric{Min: 0, Max: 10, Step: 1}, Change: callback},
		{ID: "text", Kind: ui.TextField, Text: "one", Accessibility: ui.Accessibility{Name: "User name", Help: "Enter a name"}, Change: callback},
	}}
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	check := findRecord(r, "check")
	check.widget.(*core.Switch).SetChecked(true).SendChange()
	a.Drain()
	choice := findRecord(r, "select").widget.(*core.Chooser)
	choice.SetCurrentValue("b").SendChange()
	a.Drain()
	findRecord(r, "slider").widget.(*core.Slider).SetValue(7).SendChange()
	a.Drain()
	findRecord(r, "number").widget.(*core.Spinner).SetValue(8).SendChange()
	a.Drain()
	text := findRecord(r, "text").widget.(*core.TextField)
	text.SetText("two").SendChange()
	a.Drain()
	if len(changes) != 5 || !changes[0].Checked || changes[1].SelectedID != "b" || changes[2].Number != 7 || changes[3].Number != 8 || changes[4].Text != "two" {
		t.Fatalf("wrong change payloads: %+v", changes)
	}
	for _, item := range choice.Items {
		if item.Value == "disabled" {
			t.Fatal("disabled native selector option selectable")
		}
	}
	sem := text.Property(SemanticsProperty).(Semantics)
	if sem.Name != "User name" || sem.Help != "Enter a name" || sem.Role != ui.TextField {
		t.Fatalf("semantic metadata missing: %+v", sem)
	}
	n.Children[0].Change = func(e ui.ChangeEvent) { e.ID = "updated"; changes = append(changes, e) }
	n.Children[4].Accessibility = ui.Accessibility{Name: "Updated label"}
	n.Children[4].Disabled = true
	r, err = reconcile(host, r, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if findRecord(r, "check").widget != check.widget || findRecord(r, "text").widget != text {
		t.Fatal("control identity lost")
	}
	check.widget.(*core.Switch).SetChecked(true).SendChange()
	a.Drain()
	if changes[len(changes)-1].ID != "updated" {
		t.Fatal("retained control invokes stale callback")
	}
	before := len(changes)
	text.SetText("blocked").SendChange()
	a.Drain()
	if len(changes) != before {
		t.Fatal("disabled edit reached model before refresh")
	}
	r, err = reconcile(host, r, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != before || text.Text() != "one" {
		t.Fatal("disabled control changed model or retained invalid input")
	}
	slider := findRecord(r, "slider").widget.(*core.Slider)
	slider.Value = float32(math.NaN())
	slider.SendChange()
	a.Drain()
	if len(changes) != before {
		t.Fatal("nonfinite edit reached model before refresh")
	}
	r, err = reconcile(host, r, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != before || slider.Value != 4 {
		t.Fatal("NaN input reached model")
	}
	n.Disabled = true
	r, err = reconcile(host, r, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	check.widget.(*core.Switch).SetChecked(true).SendChange()
	a.Drain()
	if len(changes) != before {
		t.Fatal("disabled ancestor allowed child callback")
	}
}

func TestNativeTabsReorderReparentAndDisabledOption(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	if err := a.Open(app.Window{ID: "main", Title: "headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	host := core.NewFrame(core.NewBody("headless"))
	selected := "a"
	tabs := ui.Node{ID: "tabs", Kind: ui.Tabs, SelectedID: "a", Options: []ui.Option{{ID: "a", Label: "First"}, {ID: "b", Label: "Second", Disabled: true}}, Change: func(e ui.ChangeEvent) { selected = e.SelectedID }, Children: []ui.Node{{ID: "a", Kind: ui.Column, Children: []ui.Node{{ID: "edit", Kind: ui.TextField, Text: "retained"}}}, {ID: "b", Kind: ui.Column}}}
	r, err := reconcile(host, nil, tabs, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	edit := findRecord(r, "edit").widget
	page := r.control.tabs.pages["a"]
	r.control.tabs.buttons["b"].Send(events.Click)
	a.Drain()
	if selected != "a" {
		t.Fatal("disabled tab option invoked change")
	}
	tabs.Options[1].Disabled = false
	r, err = reconcile(host, r, tabs, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	r.control.tabs.buttons["b"].Send(events.Click)
	a.Drain()
	if selected != "b" {
		t.Fatal("enabled tab didn't invoke change")
	}
	tabs.SelectedID = "b"
	tabs.Options = []ui.Option{{ID: "b", Label: "Renamed"}, {ID: "a", Label: "First"}}
	tabs.Children = []ui.Node{tabs.Children[1], tabs.Children[0]}
	r, err = reconcile(host, r, tabs, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if findRecord(r, "edit").widget != edit || r.control.tabs.pages["a"] != page || r.widget.(*core.Tabs).NumTabs() != 2 || r.widget.(*core.Tabs).Child(0).AsTree().Child(0) != r.control.tabs.buttons["b"] {
		t.Fatal("tab reorder replaced page or corrupted native internals")
	}
	// Move a field out of a removed tab before its native page is destroyed.
	next := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "edit", Kind: ui.TextField, Text: "retained"}, {ID: "tabs", Kind: ui.Tabs, SelectedID: "b", Options: []ui.Option{{ID: "b", Label: "Second"}}, Children: []ui.Node{{ID: "b", Kind: ui.Column}}}}}
	r, err = reconcile(host, r, next, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if findRecord(r, "edit").widget != edit || tree.IsNil(edit) || findRecord(r, "tabs").widget.(*core.Tabs).NumTabs() != 1 {
		t.Fatal("moving child from removed tab destroyed retained widget")
	}
}

func TestDialogSceneLifecyclePreservesChildrenAndDismissesOnce(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	if err := a.Open(app.Window{ID: "main", Title: "headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	host := core.NewFrame(core.NewBody("headless"))
	dismissed := 0
	n := ui.Node{ID: "dialog", Kind: ui.Dialog, Text: "Preferences", Open: true, Dismiss: func() { dismissed++ }, Children: []ui.Node{{ID: "dialog-field", Kind: ui.TextField, Text: "saved"}}}
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	field := findRecord(r, "dialog-field").widget
	body := prepareDialogBody(r)
	stage := body.NewDialog(r.widget)
	if !stage.Modal || stage.Type != core.DialogStage {
		t.Fatal("descriptor didn't create native modal")
	}
	r.control.dialog.active = true
	body.Send(events.Close)
	a.Drain()
	body.Scene.DeleteChildren()
	if dismissed != 1 || tree.IsNil(field) || r.control.dialog.host.Parent != r.widget || field.AsTree().Parent != r.control.dialog.host {
		t.Fatal("native scene close destroyed retained children or failed dismissal")
	}
	n.Open = false
	r, err = reconcile(host, r, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	n.Open = true
	r, err = reconcile(host, r, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	body = prepareDialogBody(r)
	body.NewDialog(r.widget)
	r.control.dialog.active = true
	disposeControl(r)
	if dismissed != 1 || tree.IsNil(field) {
		t.Fatal("programmatic close reentered dismiss or destroyed child")
	}
	next := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "dialog-field", Kind: ui.TextField, Text: "saved"}}}
	r, err = reconcile(host, r, next, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if findRecord(r, "dialog-field").widget != field {
		t.Fatal("field lost when moving from removed dialog")
	}
}

func TestNativeFocusRequestUsesStableIDAndSequence(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	if err := a.Open(app.Window{ID: "main", Title: "headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	host := core.NewFrame(core.NewBody("headless"))
	n := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "field", Kind: ui.TextField}, {ID: "off", Kind: ui.Button, Disabled: true}}}
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	var last ui.FocusRequest
	request := ui.FocusRequest{ID: "field", Sequence: 1}
	if err := applyFocus(r, request, &last); err != nil || last != request {
		t.Fatal("focus ID request failed")
	}
	if err := applyFocus(r, request, &last); err != nil {
		t.Fatal(err)
	}
	request.Sequence++
	if err := applyFocus(r, request, &last); err != nil || last.Sequence != 2 {
		t.Fatal("repeated ID sequence request failed")
	}
	if err := applyFocus(r, ui.FocusRequest{ID: "off", Sequence: 3}, &last); err == nil {
		t.Fatal("disabled focus target accepted")
	}
	if findRecord(r, "off").widget.AsWidget().StateIs(states.Focused) {
		t.Fatal("disabled button focused")
	}
}

func TestControlRefreshPreservesUncommittedTextAndNumericEditing(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	if err := a.Open(app.Window{ID: "main", Title: "headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	host := core.NewFrame(core.NewBody("headless"))
	changes := 0
	n := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "text", Kind: ui.TextField, Text: "one", Change: func(e ui.ChangeEvent) { changes++ }}, {ID: "number", Kind: ui.NumberField, Number: 4, Numeric: ui.Numeric{Min: 0, Max: 10, Step: 1}}}}
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	text := findRecord(r, "text").widget.(*core.TextField)
	if !text.ApplyNativeEdit("draft", 5, 5, -1, -1, "insert") {
		t.Fatal("native edit rejected")
	}
	number := findRecord(r, "number").widget.(*core.Spinner)
	number.SetText("7.25")
	// The refresh runs inside the application owner. A side-effecting Text()
	// getter here would commit and reenter owner through OnChange, deadlocking.
	if err := a.Do(func() {
		var inner error
		r, inner = reconcile(host, r, n, a, "main")
		if inner != nil {
			t.Error(inner)
		}
	}); err != nil {
		t.Fatal(err)
	}
	if changes != 0 || text.EditingState().Text != "draft" || number.EditingState().Text != "7.25" {
		t.Fatal("unrelated redraw committed or overwrote editing")
	}
	text.SendChange()
	a.Drain()
	if changes != 1 {
		t.Fatal("safe snapshot change wasn't delivered")
	}
}

func TestControlMailboxRefusalAndRemovedQueuedCallbacks(t *testing.T) {
	a, _ := app.New(app.Config{QueueCapacity: 1})
	defer a.Close()
	if err := a.Open(app.Window{ID: "main", Title: "headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	host := core.NewFrame(core.NewBody("headless"))
	changes := 0
	reported := 0
	n := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "check", Kind: ui.Checkbox, Change: func(e ui.ChangeEvent) { changes++ }}}}
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	check := findRecord(r, "check")
	bindControlWake(r, func(events.Event) {}, func(error) { reported++ })
	if err := a.Post(func() {}); err != nil {
		t.Fatal(err)
	}
	check.widget.(*core.Switch).SetChecked(true).SendChange()
	if reported != 0 || !check.widget.(*core.Switch).IsChecked() || check.control.refused.Load() == nil || changes != 0 {
		t.Fatal("refusal accessed controlled values or callback before owner refresh")
	}
	if err := a.Do(func() { r, err = reconcile(host, r, n, a, "main") }); err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	if reported != 1 || check.widget.(*core.Switch).IsChecked() || changes != 0 {
		t.Fatal("refused control edit not restored/reported")
	}
	a.Drain()
	check.widget.(*core.Switch).SetChecked(true).SendChange()
	n.Children = nil
	r, err = reconcile(host, r, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if changes != 0 || !tree.IsNil(check.widget) {
		t.Fatal("removed native control delivered queued stale callback")
	}
}

func TestDialogForwardsNativeCustomWake(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	if err := a.Open(app.Window{ID: "main", Title: "headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	host := core.NewFrame(core.NewBody("headless"))
	n := ui.Node{ID: "dialog", Kind: ui.Dialog, Open: true}
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	wakes := 0
	bindControlWake(r, func(e events.Event) { wakes++; a.Drain() })
	body := prepareDialogBody(r)
	processed := false
	a.Post(func() { processed = true })
	event := &events.CustomEvent{}
	event.Typ = events.Custom
	event.Data = "vellum.wake"
	event.Init()
	body.Scene.HandleEvent(event)
	if wakes != 1 || !processed {
		t.Fatal("modal event owner blocked application mailbox")
	}
}

func TestStandardControlLabelsRemainLiteral(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	if err := a.Open(app.Window{ID: "main", Title: "headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	host := core.NewFrame(core.NewBody("headless"))
	literal := "a < b & c"
	escaped := "a &lt; b &amp; c"
	n := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "check", Kind: ui.Checkbox, Text: literal, Accessibility: ui.Accessibility{Help: literal}}, {ID: "selector", Kind: ui.Selector, Options: []ui.Option{{ID: "value", Label: literal}}, SelectedID: "value"}, {ID: "tabs", Kind: ui.Tabs, Options: []ui.Option{{ID: "page", Label: literal}}, SelectedID: "page", Children: []ui.Node{{ID: "page", Kind: ui.Column}}}}}
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if findRecord(r, "check").widget.(*core.Switch).Text != escaped || findRecord(r, "selector").widget.(*core.Chooser).Items[0].Text != escaped || findRecord(r, "tabs").control.tabs.buttons["page"].Text != escaped || findRecord(r, "check").widget.AsWidget().Tooltip != escaped {
		t.Fatal("standard-control label/help interpreted as markup")
	}
}

func TestSelectorDeselectionClearsNativeValue(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	if err := a.Open(app.Window{ID: "main", Title: "headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	host := core.NewFrame(core.NewBody("headless"))
	n := ui.Node{ID: "choice", Kind: ui.Selector, SelectedID: "a", Options: []ui.Option{{ID: "a", Label: "A"}}, Text: "Choose an option"}
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	choice := r.widget.(*core.Chooser)
	n.SelectedID = ""
	n.Options = []ui.Option{{ID: "b", Label: "B"}}
	r, err = reconcile(host, r, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if choice.CurrentIndex != -1 || choice.CurrentItem.Value != nil || choice.CurrentItem.Text != "Choose an option" || r.widget.AsTree().Property(SemanticsProperty).(Semantics).SelectedID != "" {
		t.Fatal("native deselection retained previous item value")
	}
}

func TestRejectedCommittedTextAndNumberRestoreAcceptedModel(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	if err := a.Open(app.Window{ID: "main", Title: "headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	host := core.NewFrame(core.NewBody("headless"))
	noop := func(ui.ChangeEvent) {}
	n := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "text", Kind: ui.TextField, Text: "accepted", Change: noop}, {ID: "number", Kind: ui.NumberField, Number: 4, Numeric: ui.Numeric{Min: 0, Max: 10, Step: 1}, Change: noop}}}
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	text := findRecord(r, "text").widget.(*core.TextField)
	number := findRecord(r, "number").widget.(*core.Spinner)
	text.SetText("rejected").SendChange()
	number.SetValue(7.25).SendChange()
	a.Drain()
	r, err = reconcile(host, r, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if text.EditingState().Text != "accepted" || number.Value != 4 || number.EditingState().Text != "4" {
		t.Fatal("rejected committed value persisted after accepted model redraw")
	}
}

func TestRefusedDialogDismissalRemainsClosedUntilAcknowledged(t *testing.T) {
	a, _ := app.New(app.Config{QueueCapacity: 1})
	defer a.Close()
	if err := a.Open(app.Window{ID: "main", Title: "headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	host := core.NewFrame(core.NewBody("headless"))
	dismissed := 0
	reported := 0
	n := ui.Node{ID: "dialog", Kind: ui.Dialog, Open: true, Dismiss: func() { dismissed++ }, Children: []ui.Node{{ID: "field", Kind: ui.TextField, Text: "retained"}}}
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	bindControlWake(r, func(events.Event) {}, func(error) { reported++ })
	field := findRecord(r, "field").widget
	body := prepareDialogBody(r)
	body.NewDialog(r.widget)
	r.control.dialog.active = true
	a.Post(func() {})
	body.Send(events.Close)
	body.Scene.DeleteChildren()
	a.Drain()
	if err := a.Do(func() { r, err = reconcile(host, r, n, a, "main") }); err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	if dismissed != 0 || reported != 1 || r.control.dialog.active || !r.control.dialog.dismissed || r.control.dialog.body != nil || tree.IsNil(field) {
		t.Fatal("refused dismissal lost content, silently reopened or failed error reporting")
	}
	n.Open = false
	r, err = reconcile(host, r, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if r.control.dialog.dismissed {
		t.Fatal("Open=false did not acknowledge refused dismissal")
	}
	n.Open = true
	r, err = reconcile(host, r, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if findRecord(r, "field").widget != field {
		t.Fatal("acknowledgment/reopen discarded retained content")
	}
}

func TestNumericFieldRetainsTightFiniteBounds(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	if err := a.Open(app.Window{ID: "main", Title: "headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	host := core.NewFrame(core.NewBody("headless"))
	maximum := float32(1.999999)
	n := ui.Node{ID: "number", Kind: ui.NumberField, Number: maximum, Numeric: ui.Numeric{Min: 0, Max: maximum, Step: 0.01}}
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	number := r.widget.(*core.Spinner)
	if number.Value != maximum || number.Value > n.Numeric.Max {
		t.Fatal("native precision rounding crossed declared maximum")
	}
}

func TestTabsReorderAndSelectionAtSameNativeIndex(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	if err := a.Open(app.Window{ID: "main", Title: "headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	host := core.NewFrame(core.NewBody("headless"))
	n := ui.Node{ID: "tabs", Kind: ui.Tabs, SelectedID: "a", Options: []ui.Option{{ID: "a"}, {ID: "b"}}, Children: []ui.Node{{ID: "a", Kind: ui.Column}, {ID: "b", Kind: ui.Column}}}
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if !r.control.tabs.buttons["a"].StateIs(states.Selected) {
		t.Fatal("initial tab not selected")
	}
	n.Options[0], n.Options[1] = n.Options[1], n.Options[0]
	n.Children[0], n.Children[1] = n.Children[1], n.Children[0]
	n.SelectedID = "b"
	r, err = reconcile(host, r, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if !r.control.tabs.buttons["b"].StateIs(states.Selected) || r.control.tabs.buttons["a"].StateIs(states.Selected) {
		t.Fatal("same native index retained stale selected button after reorder")
	}
	if r.widget.(*core.Tabs).Child(1).AsTree().Child(0) != r.control.tabs.pages["b"] {
		t.Fatal("selected tab content does not match reordered model")
	}
}

func TestTabsFocusTargetsEnabledNativeButton(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	if err := a.Open(app.Window{ID: "main", Title: "headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	host := core.NewFrame(core.NewBody("headless"))
	n := ui.Node{ID: "tabs", Kind: ui.Tabs, SelectedID: "b", Options: []ui.Option{{ID: "a", Disabled: true}, {ID: "b"}}, Children: []ui.Node{{ID: "a", Kind: ui.Column}, {ID: "b", Kind: ui.Column}}}
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	focus, err := focusWidget(r)
	if err != nil || focus != r.control.tabs.buttons["b"] {
		t.Fatal("selected enabled native tab not resolved")
	}
	var last ui.FocusRequest
	if err := applyFocus(r, ui.FocusRequest{ID: "tabs", Sequence: 1}, &last); err != nil {
		t.Fatal(err)
	}
	n.SelectedID = "a"
	r, err = reconcile(host, r, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	focus, err = focusWidget(r)
	if err != nil || focus != r.control.tabs.buttons["b"] {
		t.Fatal("disabled selection did not fall back to enabled native tab")
	}
	n.Options[1].Disabled = true
	r, err = reconcile(host, r, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := focusWidget(r); err == nil {
		t.Fatal("all-disabled native tabs accepted focus")
	}
	if err := applyFocus(r, ui.FocusRequest{ID: "tabs", Sequence: 2}, &last); err == nil {
		t.Fatal("all-disabled focus request accepted")
	}
}

func TestRemovedAncestorDisposesDescendantDialog(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	if err := a.Open(app.Window{ID: "main", Title: "headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	host := core.NewFrame(core.NewBody("headless"))
	dismissed := 0
	n := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "ancestor", Kind: ui.Column, Children: []ui.Node{{ID: "modal", Kind: ui.Dialog, Open: true, Dismiss: func() { dismissed++ }, Children: []ui.Node{{ID: "field", Kind: ui.TextField}}}}}}}
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	modal := findRecord(r, "modal")
	field := findRecord(r, "field").widget
	body := prepareDialogBody(modal)
	body.NewDialog(modal.widget)
	modal.control.dialog.active = true
	n.Children = nil
	if _, err := reconcile(host, r, n, a, "main"); err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if modal.control.dialog.body != nil || modal.control.dialog.active || !tree.IsNil(field) || !tree.IsNil(modal.control.dialog.host) || dismissed != 0 {
		t.Fatal("removing ancestor left modal or retained orphan content")
	}
}

func TestDialogDisposalAfterApplicationClose(t *testing.T) {
	a, _ := app.New(app.Config{})
	host := core.NewFrame(core.NewBody("headless"))
	n := ui.Node{ID: "modal", Kind: ui.Dialog, Open: true, Children: []ui.Node{{ID: "field", Kind: ui.TextField}}}
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	body := prepareDialogBody(r)
	body.NewDialog(r.widget)
	r.control.dialog.active = true
	a.Close()
	disposeControlTree(r)
	if r.control.dialog.body != nil || r.control.dialog.active || r.control.dialog.host.Parent != r.widget {
		t.Fatal("application closure prevented native modal disposal")
	}
}
