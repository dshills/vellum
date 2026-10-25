//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
	"image"
	"testing"
)

func TestDocumentListNativeSelectionActivationAndResize(t *testing.T) {
	d := &ui.DocumentList{ID: "files", Label: "Files", Columns: ui.DefaultDocumentColumns(), Items: []ui.DocumentItem{{ID: "a", Name: "Alpha"}, {ID: "b", Name: "Beta"}}}
	calls := 0
	sortCalls := 0
	d.SortChange = func(e ui.DocumentSortEvent) { sortCalls++; d.Sort = e.Sort }
	d.Change = func(e ui.ListSelectionEvent) { d.SelectedIDs = e.Apply(d.SelectedIDs) }
	d.Activate = func(ui.DocumentEvent) { calls++ }
	d.Resize = func(e ui.DocumentResizeEvent) { d.Columns[0].Width = e.Width }
	focus := ui.DocumentListID(d.ID)
	sequence := uint64(1)
	view := func() ui.Node {
		n, err := d.View()
		if err != nil {
			t.Fatal(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: focus, Sequence: sequence}, Children: []ui.Node{n}}
	}
	m, a := simpleManual(t, manualConfig(t, 911), view)
	unitSettle(t, m)
	w := m.implementation.find("one")
	input := func(code key.Codes) {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.KeyChord, Code: code, Modifiers: 1<<key.CapsLock | 1<<key.NumLock}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	input(key.CodeDownArrow)
	input(key.CodeReturnEnter)
	if len(d.SelectedIDs) != 1 || d.SelectedIDs[0] != "b" || calls != 1 {
		t.Fatal("selection/activation", d.SelectedIDs, calls)
	}
	if err := a.Do(func() { focus = ui.DocumentHeaderID(d.ID, "name") + "/resize"; sequence++; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	input(key.CodeRightArrow)
	if d.Columns[0].Width != 221 {
		t.Fatal("column keyboard resize", d.Columns)
	}
	_, handle := imageManualElement(t, m, ui.DocumentHeaderID(d.ID, "name")+"/resize")
	at := handle.Bounds.Min.Add(handle.Bounds.Size().Div(2))
	end := at.Add(image.Pt(24, 0))
	for _, step := range []struct {
		typ events.Types
		p   image.Point
	}{{events.MouseDown, at}, {events.MouseMove, at.Add(image.Pt(12, 0))}, {events.MouseMove, end}, {events.MouseUp, end}} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: step.typ, Button: events.Left, Position: step.p}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	if d.Columns[0].Width <= 221 || calls != 1 {
		t.Fatal("pointer resize or release consumption", d.Columns[0].Width, calls)
	}

	_, currentHandle := imageManualElement(t, m, ui.DocumentHeaderID(d.ID, "name")+"/resize")
	startLoss := currentHandle.Bounds.Min.Add(currentHandle.Bounds.Size().Div(2))
	lossEnd := startLoss.Add(image.Pt(24, 0))
	widthBeforeLoss := d.Columns[0].Width
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.MouseDown, Button: events.Left, Position: startLoss}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	findRecord(w.state.root, ui.DocumentHeaderID(d.ID, "name")+"/resize").widget.AsWidget().Scene.Events.CancelPointerCapture()
	for _, typ := range []events.Types{events.MouseMove, events.MouseUp} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: typ, Button: events.Left, Position: lossEnd}); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	if d.Columns[0].Width != widthBeforeLoss || sortCalls != 0 {
		t.Fatal("capture loss resized or clicked a header", d.Columns[0].Width, sortCalls)
	}

	if err := a.Do(func() {
		d.Columns = []ui.DocumentColumn{{ID: "kind", Title: "Kind", Width: 100}, {ID: "name", Title: "Name", Width: 240}}
		_ = a.Invalidate("one")
	}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	row := findRecord(w.state.root, ui.DocumentListID(d.ID)).control.choice.buttons["b"]
	if len(row.Children) != 2 || row.Children[0].AsTree().Name != "document-cell-kind" || row.Children[1].AsTree().Name != "document-cell-name" {
		t.Fatal("column plan was not rebuilt", row.Children)
	}

}
