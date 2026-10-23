package ui

import (
	"fmt"
	"image/color"
)

// SecureField composes a protected TextField and optional explicit reveal action.
// Retain on one owner. Nil Change disables entry and reveal.
// Text is an ordinary Go string; it is not zeroizable storage
// or an OS secure-event-input guarantee. Revealing affects only visible glyphs;
// accessibility and native surrounding-text queries remain protected.
type SecureField struct {
	ID, Label, Text, Placeholder string
	Disabled, AllowReveal        bool
	Change                       func(ChangeEvent)
	Style                        Style
	revealed                     bool
}

func SecureInputID(root string) string { return root + "/secure/input" }
func (s *SecureField) View() (Node, error) {
	if s == nil || s.ID == "" {
		return Node{}, fmt.Errorf("ui: secure field requires ID")
	}
	id := s.ID
	label := s.Label
	if label == "" {
		label = "Password"
	}
	available := func() bool { return s.ID == id && !s.Disabled && s.Change != nil }
	if !s.AllowReveal || !available() {
		s.revealed = false
	}
	clear := Style{Border: Some(Border{}), Background: Some(color.NRGBA{}), MinSize: Some(Size{Height: 26}), MaxSize: Some(Size{}), GrowX: Some(float32(1)), Padding: Some(Insets{Top: 2, Bottom: 2, Left: 6, Right: 6})}
	field := Node{ID: SecureInputID(id), Kind: TextField, Text: s.Text, Placeholder: s.Placeholder, Secure: true, Reveal: s.revealed, Style: clear, Accessibility: Accessibility{Name: label, Help: "Protected text entry"}, Change: func(e ChangeEvent) {
		if available() {
			e.ID = id
			s.Change(e)
		}
	}}
	children := []Node{field}
	if s.AllowReveal {
		text := "Show"
		if s.revealed {
			text = "Hide"
		}
		children = append(children, Node{ID: id + "/secure/reveal", Kind: Button, Text: text, Appearance: ButtonQuiet, Accessibility: Accessibility{Name: text + " " + label}, Style: Style{GrowX: Some(float32(0)), Padding: Some(Insets{Top: 3, Bottom: 3, Left: 6, Right: 6})}, Activate: func() {
			if available() && s.AllowReveal {
				s.revealed = !s.revealed
			}
		}})
	}
	n := Node{ID: id, Kind: Row, Disabled: !available(), Style: Style{Gap: Some(float32(0)), GrowY: Some(float32(0))}.Overlay(s.Style), Children: children}
	if err := Validate(n); err != nil {
		return Node{}, err
	}
	return n, nil
}
