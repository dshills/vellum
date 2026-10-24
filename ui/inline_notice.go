package ui

import (
	"fmt"
	"strings"
)

// NoticeTone selects the symbol and accessible description, not a global palette.
type NoticeTone string

const (
	NoticeInformation NoticeTone = "information"
	NoticeSuccess     NoticeTone = "success"
	NoticeWarning     NoticeTone = "warning"
	NoticeError       NoticeTone = "error"
)

// InlineNotice is a bounded message with an optional application-owned action.
// Text is literal. The component owns ID + "/notice/". It is a semantic group,
// not a modal or live-region announcement; applications decide when to show it.
type InlineNotice struct {
	ID, Text, ActionLabel string
	Tone                  NoticeTone
	Disabled              bool
	Activate              func()
	Style                 Style
}

func (n *InlineNotice) View() (Node, error) {
	if n == nil || n.ID == "" || strings.TrimSpace(n.Text) == "" {
		return Node{}, fmt.Errorf("ui: inline notice requires an ID and text")
	}
	icon := SymbolInfo
	tone := n.Tone
	switch tone {
	case "", NoticeInformation:
		tone = NoticeInformation
	case NoticeSuccess:
		icon = SymbolCheck
	case NoticeWarning, NoticeError:
		icon = SymbolWarning
	default:
		return Node{}, fmt.Errorf("ui: unknown notice tone %q", tone)
	}
	if (n.Activate != nil) != (strings.TrimSpace(n.ActionLabel) != "") {
		return Node{}, fmt.Errorf("ui: inline notice action requires both label and callback")
	}
	id := n.ID
	root := Node{ID: id, Kind: Row, Disabled: n.Disabled, Style: n.Style, Accessibility: Accessibility{Name: string(tone) + ": " + n.Text},
		Children: []Node{
			{ID: id + "/notice/icon", Kind: Icon, Icon: icon},
			{ID: id + "/notice/text", Kind: Label, Text: n.Text, Style: Style{GrowX: Some(float32(1)), GrowY: Some(float32(0))}},
		}}
	if n.Activate != nil {
		root.Children = append(root.Children, Node{ID: id + "/notice/action", Kind: Button, Text: n.ActionLabel, Appearance: ButtonQuiet,
			Activate: func() {
				if n.ID == id && !n.Disabled && n.Activate != nil && n.ActionLabel != "" {
					n.Activate()
				}
			}})
	}
	if err := Validate(root); err != nil {
		return Node{}, err
	}
	return root, nil
}
