package ui

import (
	"fmt"
	"strings"
)

// FormField labels one focusable managed leaf and displays caller-owned help and
// validation text. Required is descriptive; the application performs validation.
// Content keeps its own ID, value and callbacks. View does not mutate Content.
type FormField struct {
	ID, Label, Help, ValidationMessage string
	Required, Disabled                 bool
	Content                            Node
	Style                              Style
}

func (f FormField) View() (Node, error) {
	if f.ID == "" || strings.TrimSpace(f.Label) == "" || !IsFocusable(f.Content.Kind) || IsContainer(f.Content.Kind) {
		return Node{}, fmt.Errorf("ui: form field requires an ID, label and focusable leaf content")
	}
	content := f.Content
	if strings.TrimSpace(content.Accessibility.Name) == "" {
		content.Accessibility.Name = f.Label
	}
	help := []string{}
	if content.Accessibility.Help != "" {
		help = append(help, content.Accessibility.Help)
	}
	if f.Required {
		help = append(help, "Required")
	}
	if f.Help != "" {
		help = append(help, f.Help)
	}
	if f.ValidationMessage != "" {
		help = append(help, "Error: "+f.ValidationMessage)
	}
	content.Accessibility.Help = strings.Join(help, ". ")
	heading := f.Label
	if f.Required {
		heading += " (required)"
	}
	root := Node{ID: f.ID, Kind: FormGroup, Text: heading, Disabled: f.Disabled, Style: f.Style, Accessibility: Accessibility{Name: heading}, Children: []Node{content}}
	if f.Help != "" {
		root.Children = append(root.Children, Node{ID: f.ID + "/field/help", Kind: Label, Text: f.Help})
	}
	if f.ValidationMessage != "" {
		root.Children = append(root.Children, Node{ID: f.ID + "/field/error", Kind: Label, Text: "Error: " + f.ValidationMessage, Accessibility: Accessibility{Name: "Error: " + f.ValidationMessage}})
	}
	if err := Validate(root); err != nil {
		return Node{}, err
	}
	return root, nil
}
