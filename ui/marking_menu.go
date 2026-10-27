package ui

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
)

type MarkingAction struct {
	ID, Label string
	IconImage *canvas.TiledImage
	Disabled  bool
}

// MarkingMenu is a controlled, same-window directional action panel. Actions
// map clockwise from north, max8. Choose accepts a direction from another router.
// This version uses an ordinary focus-scoped dialog: keyboard/visible buttons
// are supported; unshown flick gestures and nested radial tracking are not.
type MarkingMenu struct {
	ID, Label      string
	Actions        []MarkingAction
	Open, Disabled bool
	Choose         func(string)
	Dismiss        func()
}

func (p *MarkingMenu) selectAction(index int) {
	if p == nil || !p.Open || p.Disabled || p.Choose == nil || index < 0 || index >= len(p.Actions) {
		return
	}
	a := p.Actions[index]
	if !a.Disabled {
		p.Choose(a.ID)
	}
}
func (p *MarkingMenu) ChooseDirection(index int) { p.selectAction(index) }
func (p *MarkingMenu) View() (Node, error) {
	if p == nil || !validTokenText(p.ID) || !validTokenText(p.Label) || len(p.Actions) > 8 {
		return Node{}, fmt.Errorf("ui: invalid marking menu")
	}
	seen := map[string]bool{}
	for _, a := range p.Actions {
		if !validTokenText(a.ID) || !validTokenText(a.Label) || seen[a.ID] {
			return Node{}, fmt.Errorf("ui: invalid marking action")
		}
		seen[a.ID] = true
	}
	id := p.ID
	n := Node{ID: id, Kind: Dialog, Text: p.Label, Open: p.Open, Disabled: p.Disabled, Dismiss: func() {
		if p.ID == id && p.Dismiss != nil {
			p.Dismiss()
		}
	}, Style: Style{Padding: Some(UniformInsets(16)), Gap: Some(float32(8))}}
	positions := [][]int{{7, 0, 1}, {6, -1, 2}, {5, 4, 3}}
	for y, indices := range positions {
		row := Node{ID: fmt.Sprintf("%s/row/%d", id, y), Kind: Row, Style: Style{Gap: Some(float32(8))}}
		for x, index := range indices {
			cell := Node{ID: fmt.Sprintf("%s/cell/%d/%d", id, y, x), Kind: Column, Style: Style{MinSize: Some(Size{Width: 100, Height: 64}), MaxSize: Some(Size{Width: 100, Height: 64})}}
			if index >= 0 && index < len(p.Actions) {
				a := p.Actions[index]
				actionID := a.ID
				cell.Kind = Button
				cell.Text = a.Label
				cell.IconImage = a.IconImage
				if a.IconImage != nil {
					cell.IconPlacement = IconAboveText
				}
				cell.Disabled = a.Disabled || p.Choose == nil
				cell.Activate = func() {
					if p.ID != id {
						return
					}
					for j, current := range p.Actions {
						if current.ID == actionID {
							p.selectAction(j)
							return
						}
					}
				}
			}
			row.Children = append(row.Children, cell)
		}
		n.Children = append(n.Children, row)
	}
	return n, Validate(n)
}
