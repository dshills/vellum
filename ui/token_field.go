package ui

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"strings"
	"unicode/utf8"
)

const MaxTokens = 32

// Token has application-assigned identity independent of its editable label.
type Token struct{ ID, Text string }

// TokenEvent requests one addition (TokenID empty) or removal. SuggestionID is
// set only for a catalog choice. The application assigns IDs to accepted additions.
type TokenEvent struct {
	ID, TokenID, Text, SuggestionID string
	Remove                          bool
}

// TokenField is owner-held. Return commits one trimmed token; suggestion choice
// uses the same Change path. Chip buttons remove via click or Tab/Return/Space.
// The owner accepts by changing Tokens, and controls Draft separately. No token
// rearrangement or in-place chip editing is implemented. Retain on one owner.
type TokenField struct {
	ID, Label, Draft, Placeholder string
	Tokens                        []Token
	Suggestions                   []Option
	Disabled                      bool
	Change                        func(TokenEvent)
	Invalid                       func(error)
	Style, InputStyle, ChipStyle  Style
	input                         ComboBox
}

func TokenInputID(root string) string { return ComboInputID(root + "/tokens/entry") }
func TokenRemoveID(root, token string) string {
	return root + "/tokens/remove/" + base64.RawURLEncoding.EncodeToString([]byte(token))
}
func (f *TokenField) validate() error {
	if f == nil || f.ID == "" || len(f.Tokens) > MaxTokens || len(f.Suggestions) > MaxMenuOptions {
		return fmt.Errorf("ui: token field requires ID, at most %d tokens and %d suggestions", MaxTokens, MaxMenuOptions)
	}
	seen := map[string]bool{}
	for _, t := range f.Tokens {
		if t.ID == "" || seen[t.ID] || !validTokenText(t.Text) {
			return fmt.Errorf("ui: tokens require unique IDs and nonblank valid labels")
		}
		seen[t.ID] = true
	}
	return nil
}
func validTokenText(s string) bool {
	return strings.TrimSpace(s) != "" && len(s) <= 256 && utf8.ValidString(s) && !strings.ContainsAny(s, "\x00\r\n")
}
func (f *TokenField) View() (Node, error) {
	if err := f.validate(); err != nil {
		return Node{}, err
	}
	id := f.ID
	label := f.Label
	if label == "" {
		label = "Tags"
	}
	available := func() bool { return f.ID == id && !f.Disabled && f.Change != nil && f.validate() == nil }
	chipStyle := Style{Background: Some(color.NRGBA{R: 218, G: 233, B: 251, A: 255}), Foreground: Some(color.NRGBA{R: 25, G: 57, B: 96, A: 255}), Border: Some(Border{}), CornerRadius: Some(float32(6)), Padding: Some(Insets{Top: 3, Bottom: 3, Left: 8, Right: 8}), MinSize: Some(Size{Height: 24}), GrowX: Some(float32(0))}.Overlay(f.ChipStyle)
	chips := Node{ID: id + "/tokens/chips", Kind: Toolbar, Style: Style{Background: Some(color.NRGBA{}), Gap: Some(float32(4)), Padding: Some(Insets{})}}
	for _, t := range f.Tokens {
		shown := t
		chips.Children = append(chips.Children, Node{ID: TokenRemoveID(id, t.ID), Kind: Button, Text: t.Text + " ×", Style: chipStyle, Accessibility: Accessibility{Name: "Remove " + t.Text, Help: "Remove this tag with Return or Space"}, Activate: func() {
			if !available() {
				return
			}
			for _, current := range f.Tokens {
				if current == shown {
					f.Change(TokenEvent{ID: id, TokenID: shown.ID, Text: shown.Text, Remove: true})
					return
				}
			}
		}})
	}
	shownSuggestions := append([]Option(nil), f.Suggestions...)
	f.input = ComboBox{ID: id + "/tokens/entry", Label: label, Text: f.Draft, Placeholder: f.Placeholder, Options: shownSuggestions, Disabled: !available() || len(f.Tokens) == MaxTokens, Style: f.InputStyle}
	if f.input.Placeholder == "" {
		f.input.Placeholder = "Add a tag…"
	}
	f.input.Change = func(e ComboEvent) {
		if !available() || len(f.Tokens) == MaxTokens {
			return
		}
		// The embedded combo sees its last rendered catalog; the owner may already
		// have changed Suggestions before the next reconciliation.
		if e.OptionID != "" {
			valid := false
			for _, shown := range shownSuggestions {
				if shown.ID != e.OptionID || shown.Label != e.Text || shown.Disabled {
					continue
				}
				for _, current := range f.Suggestions {
					if current == shown {
						valid = true
						break
					}
				}
			}
			if !valid {
				return
			}
		}
		text := strings.TrimSpace(e.Text)
		if !validTokenText(text) {
			if f.Invalid != nil {
				f.Invalid(fmt.Errorf("ui: enter a nonblank tag of at most 256 bytes"))
			}
			return
		}
		// Refuse duplicate labels without mutating accepted tokens or draft.
		for _, t := range f.Tokens {
			if t.Text == text {
				if f.Invalid != nil {
					f.Invalid(fmt.Errorf("ui: tag already exists"))
				}
				return
			}
		}
		f.Change(TokenEvent{ID: id, Text: text, SuggestionID: e.OptionID})
	}
	input, err := f.input.View()
	if err != nil {
		return Node{}, err
	}
	children := []Node{}
	if len(chips.Children) > 0 {
		children = append(children, chips)
	}
	children = append(children, input)
	style := Style{Background: Some(color.NRGBA{}), Gap: Some(float32(6)), GrowX: Some(float32(1)), GrowY: Some(float32(0))}.Overlay(f.Style)
	n := Node{ID: id, Kind: Column, Disabled: !available(), Style: style, Accessibility: Accessibility{Name: label}, Children: children}
	if err := Validate(n); err != nil {
		return Node{}, err
	}
	return n, nil
}
