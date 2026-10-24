package ui

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
	"image/color"
	"slices"
)

// SplitButton presents independent primary and menu affordances. Retain it on
// one owner; callbacks and enabled state are checked again when invoked.
// Small selects a 24dp compact height; regular height is 30dp. Theme controls
// ink, field, border and focus colors. ID owns the /split-button/ namespace.
type SplitButton struct {
	ID, Label       string
	Icon            Symbol
	IconImage       *canvas.TiledImage
	Options         []Option
	Disabled, Small bool
	Primary         func()
	Change          func(ChangeEvent)
	Style           Style
}

func SplitPrimaryID(id string) string { return id + "/split-button/primary" }
func SplitMenuID(id string) string    { return id + "/split-button/menu" }
func (s *SplitButton) View() (Node, error) {
	if s == nil || s.ID == "" || s.Label == "" {
		return Node{}, fmt.Errorf("ui: split button needs ID and label")
	}
	id := s.ID
	options := slices.Clone(s.Options)
	h := float32(30)
	if s.Small {
		h = 24
	}
	compact := Style{MinSize: Some(Size{Height: h}), MaxSize: Some(Size{Height: h}), Padding: Some(Insets{Top: 2, Bottom: 2, Left: 8, Right: 8}), GrowX: Some(float32(0)), CornerRadius: Some(float32(0)), Border: Some(Border{}), Background: Some(color.NRGBA{})}
	available := func() bool { return s.ID == id && !s.Disabled }
	main := Node{ID: SplitPrimaryID(id), Kind: Button, Text: s.Label, Icon: s.Icon, IconImage: s.IconImage, Style: compact, Disabled: s.Disabled || s.Primary == nil, Activate: func() {
		if available() && s.Primary != nil {
			s.Primary()
		}
	}}
	menuStyle := compact
	menuStyle.MinSize = Some(Size{Width: 24, Height: h})
	menuStyle.MaxSize = Some(Size{Width: 28, Height: h})
	menuStyle.Padding = Some(UniformInsets(3))
	menu := Node{ID: SplitMenuID(id), Kind: MenuButton, MenuAnchorID: id, Icon: SymbolDown, Accessibility: Accessibility{Name: s.Label + " options"}, Style: menuStyle, Disabled: s.Disabled || s.Change == nil, Options: options, Change: func(e ChangeEvent) {
		if !available() || s.Change == nil {
			return
		}
		old, shown := MenuAction(options, e.SelectedID)
		current, ok := MenuAction(s.Options, e.SelectedID)
		if shown && ok && old == current {
			e.ID = id
			s.Change(e)
		}
	}}
	style := Style{Gap: Some(float32(0)), Padding: Some(Insets{}), GrowX: Some(float32(0)), Align: Some(AlignCenter)}.Overlay(s.Style)
	root := Node{ID: id, Kind: Row, SplitButtonLayout: true, Style: style, Disabled: s.Disabled, Accessibility: Accessibility{Name: s.Label}, Children: []Node{main, menu}}
	if err := Validate(root); err != nil {
		return Node{}, err
	}
	return root, nil
}
