package ui

import (
	"fmt"
	"image/color"
)

// Sheet is an owner-held, window-modal scoped task. Render and mutate it only
// on its application owner; concurrent View calls or cross-owner use are unsupported.
// A disabled or nil-Confirm default action is valid and inert. Accept Confirm/Cancel by
// setting Open false; rejected confirmation can leave it open for correction.
// Return invokes the default action only if a focused editor did not consume it.
// ReturnFocusID optionally names an enabled visible Button in the parent scene.
// No native window, draggable title or implicit Close button is added.
type Sheet struct {
	ID, Title, ConfirmLabel, CancelLabel, ReturnFocusID string
	Open, Disabled, ConfirmDisabled                     bool
	Children                                            []Node
	Confirm, Cancel                                     func()
	Style                                               Style
	epoch                                               *sheetEpoch
	identity                                            string
}
type sheetEpoch struct{ marker byte }

func SheetConfirmID(id string) string { return id + "/sheet/confirm" }
func SheetCancelID(id string) string  { return id + "/sheet/cancel" }
func (s *Sheet) View() (Node, error) {
	if s == nil || s.ID == "" || !validTokenText(s.Title) || s.Cancel == nil {
		return Node{}, fmt.Errorf("ui: sheet requires ID, title and Cancel")
	}
	if s.identity != s.ID || !s.Open || s.Disabled {
		s.epoch = nil
		s.identity = s.ID
	}
	if s.Open && !s.Disabled && s.epoch == nil {
		s.epoch = &sheetEpoch{}
	}
	id, epoch := s.ID, s.epoch
	live := func() bool { return s.ID == id && s.Open && !s.Disabled && s.epoch == epoch && epoch != nil }
	cancel := func() {
		if live() && s.Cancel != nil {
			s.Cancel()
		}
	}
	primary, secondary := s.ConfirmLabel, s.CancelLabel
	if primary == "" {
		primary = "Done"
	}
	if secondary == "" {
		secondary = "Cancel"
	}
	content := append([]Node{{ID: id + "/sheet/title", Kind: Label, Text: s.Title, Style: Style{Typography: Some(Typography{Size: 17, Weight: 600})}}}, s.Children...)
	actions := Node{ID: id + "/sheet/actions", Kind: Row, Style: Style{GrowX: Some(float32(1)), Justify: Some(AlignEnd), Gap: Some(float32(8)), Padding: Some(Insets{Top: 12})}, Children: []Node{{ID: SheetCancelID(id), Kind: Button, Text: secondary, Appearance: ButtonSecondary, Activate: cancel}, {ID: SheetConfirmID(id), Kind: Button, Text: primary, Appearance: ButtonPrimary, Style: Style{Background: Some(color.NRGBA{R: 0, G: 112, B: 245, A: 255}), Foreground: Some(color.NRGBA{R: 255, G: 255, B: 255, A: 255})}, Disabled: s.Confirm == nil || s.ConfirmDisabled, Activate: func() {
		if live() && !s.ConfirmDisabled && s.Confirm != nil {
			s.Confirm()
		}
	}}}}
	content = append(content, actions)
	n := Node{ID: id, Kind: Dialog, Text: s.Title, Sheet: true, DialogDefaultID: SheetConfirmID(id), DialogReturnFocusID: s.ReturnFocusID, Open: s.Open, Disabled: s.Disabled, Dismiss: cancel, Children: content, Style: Style{Padding: Some(UniformInsets(20)), Gap: Some(float32(12)), MinSize: Some(Size{Width: 300}), MaxSize: Some(Size{Width: 480}), CornerRadius: Some(float32(12))}.Overlay(s.Style)}
	if err := Validate(n); err != nil {
		return Node{}, err
	}
	return n, nil
}
func validateSheet(n Node) error {
	if !n.Sheet {
		if n.DialogDefaultID != "" || n.DialogReturnFocusID != "" {
			return fmt.Errorf("ui: sheet action/focus targets require Sheet")
		}
		return nil
	}
	if n.Kind != Dialog || n.PopoverAnchor != "" || n.Dismiss == nil {
		return fmt.Errorf("ui: sheet requires unanchored dismissible Dialog")
	}
	found := false
	var walk func(Node)
	walk = func(c Node) {
		if c.ID == n.DialogDefaultID && c.Kind == Button {
			found = true
		}
		for _, child := range c.Children {
			walk(child)
		}
	}
	for _, c := range n.Children {
		walk(c)
	}
	if n.DialogDefaultID == "" || !found || n.DialogReturnFocusID == n.ID {
		return fmt.Errorf("ui: sheet requires a default Button child and distinct return target")
	}
	return nil
}
