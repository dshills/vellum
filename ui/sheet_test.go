package ui

import "testing"

func TestSheetControlledActionsAndRetirement(t *testing.T) {
	s := &Sheet{ID: "task", Title: "Export drawing", Open: true}
	confirms, cancels := 0, 0
	s.Confirm = func() { confirms++ }
	s.Cancel = func() { cancels++; s.Open = false }
	n, err := s.View()
	if err != nil {
		t.Fatal(err)
	}
	row := n.Children[len(n.Children)-1]
	ok, cancel := row.Children[1].Activate, row.Children[0].Activate
	ok()
	if confirms != 1 || !s.Open {
		t.Fatal("rejected confirmation must remain open")
	}
	s.ConfirmDisabled = true
	ok()
	if confirms != 1 {
		t.Fatal("disabled confirm")
	}
	cancel()
	if cancels != 1 || s.Open {
		t.Fatal("cancel")
	}
	if _, err = s.View(); err != nil {
		t.Fatal(err)
	}
	s.Open = true
	if _, err = s.View(); err != nil {
		t.Fatal(err)
	}
	ok()
	cancel()
	if confirms != 1 || cancels != 1 {
		t.Fatal("stale session")
	}
	if n.PopoverAnchor != "" || !n.Sheet || n.DialogDefaultID != SheetConfirmID("task") {
		t.Fatal("not a sheet")
	}
}
func TestSheetRejectsMalformedDescriptors(t *testing.T) {
	s := &Sheet{ID: "x", Title: "Task"}
	if _, err := s.View(); err == nil {
		t.Fatal("missing cancel")
	}
	if ValidateControl(Node{Kind: Button, Sheet: true}) == nil {
		t.Fatal("sheet button")
	}
	if ValidateControl(Node{Kind: Dialog, DialogDefaultID: "x"}) == nil {
		t.Fatal("bare target")
	}
}

func TestSheetInertDefaultDescriptor(t *testing.T) {
	n := Node{ID: "sheet", Kind: Dialog, Sheet: true, DialogDefaultID: "done", Dismiss: func() {}, Children: []Node{{ID: "done", Kind: Button, Disabled: true}}}
	if err := Validate(n); err != nil {
		t.Fatal(err)
	}
}
