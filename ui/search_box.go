package ui

import (
	"fmt"
	"strings"
)

// SearchEvent proposes a query. Clear requests the empty query through the same path.
type SearchEvent struct{ ID, Text string }

// SearchBox is an owner-held controlled search input. Nil Change makes it disabled.
// It owns no search engine, timers, history or goroutines. Retain one instance on
// one application owner; accept or reject proposals by updating Text in Change.
type SearchBox struct {
	ID, Label, Text, Help, Placeholder string
	Disabled                           bool
	Change                             func(SearchEvent)
	Style                              Style
	InputStyle, ClearStyle             Style
}

// SearchInputID returns the managed field target for normal FocusRequest use.
// The component owns root + "/search/".
func SearchInputID(root string) string { return root + "/search/input" }
func SearchClearID(root string) string { return root + "/search/clear" }

func (s *SearchBox) View() (Node, error) {
	if s == nil || s.ID == "" {
		return Node{}, fmt.Errorf("ui: search box requires an ID")
	}
	id := s.ID
	label := s.Label
	if strings.TrimSpace(label) == "" {
		label = "Search"
	}
	placeholder := s.Placeholder
	if placeholder == "" {
		placeholder = label
	}
	style := s.Style
	if !style.Gap.Set {
		style.Gap = Some(float32(4))
	}
	available := func() bool { return s.ID == id && !s.Disabled && s.Change != nil }
	root := Node{ID: id, Kind: Row, Disabled: s.Disabled || s.Change == nil, Style: style, Accessibility: Accessibility{Name: label}}
	root.Children = []Node{
		{ID: id + "/search/icon", Kind: Icon, Icon: SymbolSearch, Style: Style{MinSize: Some(Size{Width: 16, Height: 16}), MaxSize: Some(Size{Width: 16, Height: 16})}},
		{ID: SearchInputID(id), Kind: TextField, Text: s.Text, Placeholder: placeholder, Style: s.InputStyle, Grow: true, Accessibility: Accessibility{Name: label, Help: s.Help},
			Change: func(e ChangeEvent) {
				if available() {
					s.Change(SearchEvent{ID: id, Text: e.Text})
				}
			}},
		{ID: SearchClearID(id), Kind: Button, Icon: SymbolClose, Appearance: ButtonTool, Style: s.ClearStyle, Disabled: s.Text == "",
			Accessibility: Accessibility{Name: "Clear " + label, Help: "Clear the search terms"},
			Activate: func() {
				if available() && s.Text != "" {
					s.Change(SearchEvent{ID: id})
				}
			}},
	}
	if err := Validate(root); err != nil {
		return Node{}, err
	}
	return root, nil
}
