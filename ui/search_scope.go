package ui

import (
	"encoding/base64"
	"fmt"
	"github.com/dshills/vellum/canvas"
	"image/color"
	"strings"
	"unicode/utf8"
)

type SearchScopeItem struct {
	ID, Label string
	Count     int
	Disabled  bool
	IconImage *canvas.TiledImage
}
type SearchScopeValue struct{ ScopeID, Query string }
type SearchScopeChangeKind uint8

const (
	SearchScopeQuery SearchScopeChangeKind = iota
	SearchScopeSelection
)

type SearchScopeEvent struct {
	ID             string
	Kind           SearchScopeChangeKind
	ScopeID, Query string
}

func (e SearchScopeEvent) Apply(current SearchScopeValue) SearchScopeValue {
	switch e.Kind {
	case SearchScopeQuery:
		current.Query = e.Query
	case SearchScopeSelection:
		current.ScopeID = e.ScopeID
	}
	return current
}

// SearchScopeBar proposes search text and scope independently. Applications own
// results and counts; it never performs filesystem or network search.
type SearchScopeBar struct {
	ID, Label                       string
	Value                           SearchScopeValue
	Items                           []SearchScopeItem
	Small, Disabled                 bool
	Change                          func(SearchScopeEvent)
	Style, ItemStyle, SelectedStyle Style
}

func SearchScopeItemID(root, id string) string {
	return root + "/scope/" + base64.RawURLEncoding.EncodeToString([]byte(id))
}
func (s *SearchScopeBar) validate() error {
	if s == nil || !validTokenText(s.ID) || !validTokenText(s.Label) || len(s.Items) < 1 || len(s.Items) > 16 || len(s.Value.Query) > 4096 || !utf8.ValidString(s.Value.Query) || strings.ContainsRune(s.Value.Query, 0) {
		return fmt.Errorf("ui: search scopes require ID/label, 1..16 items and bounded query")
	}
	seen := map[string]bool{}
	for _, i := range s.Items {
		if !validTokenText(i.ID) || !validTokenText(i.Label) || i.Count < -1 || i.Count > 1000000 || seen[i.ID] {
			return fmt.Errorf("ui: invalid search scope")
		}
		seen[i.ID] = true
	}
	if !seen[s.Value.ScopeID] {
		return fmt.Errorf("ui: selected search scope absent")
	}
	return nil
}
func (s *SearchScopeBar) View() (Node, error) {
	if err := s.validate(); err != nil {
		return Node{}, err
	}
	id := s.ID
	available := func() bool { return s.ID == id && !s.Disabled && s.Change != nil && s.validate() == nil }
	height := float32(30)
	if s.Small {
		height = 24
	}
	fieldStyle := Style{MinSize: Some(Size{Height: height})}
	query := &SearchBox{ID: id + "/scope/query", Label: s.Label, Text: s.Value.Query, Placeholder: "Search documents", Disabled: !available(), Style: Style{GrowX: Some(float32(1))}, InputStyle: Style{MinSize: Some(Size{Width: 180, Height: height})}, ClearStyle: fieldStyle, Change: func(e SearchEvent) {
		if available() && len(e.Text) <= 4096 && utf8.ValidString(e.Text) && !strings.ContainsRune(e.Text, 0) {
			s.Change(SearchScopeEvent{ID: id, Kind: SearchScopeQuery, Query: e.Text})
		}
	}}
	search, err := query.View()
	if err != nil {
		return Node{}, err
	}
	scopes := Node{ID: id + "/scope/items", Kind: Toolbar, Accessibility: Accessibility{Name: "Search in"}, Style: Style{Background: Some(color.NRGBA{}), GrowX: Some(float32(1)), Gap: Some(float32(4)), Padding: Some(Insets{})}}
	for _, item := range s.Items {
		item := item
		selected := s.Value.ScopeID == item.ID
		label := item.Label
		if item.Count >= 0 {
			label = fmt.Sprintf("%s (%d)", label, item.Count)
		}
		appearance := ButtonQuiet
		style := Style{MinSize: Some(Size{Height: height}), Padding: Some(Insets{Top: 3, Bottom: 3, Left: 8, Right: 8}), GrowX: Some(float32(0))}.Overlay(s.ItemStyle)
		help := "Choose search scope"
		if selected {
			appearance = ButtonPrimary
			style = style.Overlay(s.SelectedStyle)
			help = "Selected search scope"
		}
		scopes.Children = append(scopes.Children, Node{ID: SearchScopeItemID(id, item.ID), Kind: Button, Text: label, IconImage: item.IconImage, Appearance: appearance, Style: style, Disabled: !available() || item.Disabled, Accessibility: Accessibility{Name: label, Help: help}, Activate: func() {
			if !available() {
				return
			}
			for _, current := range s.Items {
				if current.ID == item.ID && !current.Disabled && s.Value.ScopeID != current.ID {
					s.Change(SearchScopeEvent{ID: id, Kind: SearchScopeSelection, ScopeID: current.ID})
					return
				}
			}
		}})
	}
	n := Node{ID: id, Kind: Column, Style: Style{Gap: Some(float32(6)), GrowX: Some(float32(1))}.Overlay(s.Style), Accessibility: Accessibility{Name: s.Label}, Children: []Node{search, scopes}}
	return n, Validate(n)
}
