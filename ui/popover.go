package ui

import "fmt"

// Popover is a controlled, same-window anchored panel with managed children.
// AnchorID names a visible enabled managed Button in the same window. Missing,
// removed or hidden anchors cancel the panel. Set Open false in Dismiss. Closed
// content retains its normal keyed identity; reopening creates a new focus scope.
// No detachable or cross-window panels are created.
type Popover struct {
	ID, AnchorID, Label string
	Open, Disabled      bool
	Children            []Node
	Dismiss             func()
	Style               Style
}

func (p Popover) View() (Node, error) {
	if p.ID == "" || p.AnchorID == "" || p.ID == p.AnchorID {
		return Node{}, fmt.Errorf("ui: popover requires distinct ID and anchor")
	}
	n := Node{ID: p.ID, Kind: Dialog, Text: p.Label, PopoverAnchor: p.AnchorID, Open: p.Open, Disabled: p.Disabled, Dismiss: p.Dismiss, Style: p.Style}
	n.Children = append([]Node{{ID: p.ID + "/popover/arrow", Kind: Label, Text: "▴", Style: Style{TextAlign: Some(AlignStart), Padding: Some(Insets{Left: 8})}}}, p.Children...)
	if err := Validate(n); err != nil {
		return Node{}, err
	}
	return n, nil
}
