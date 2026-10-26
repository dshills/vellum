package ui

import (
	"encoding/base64"
	"fmt"
	"github.com/dshills/vellum/canvas"
	"image/color"
)

type ChecklistAction string

const (
	ChecklistToggle ChecklistAction = "toggle"
	ChecklistRename ChecklistAction = "rename"
)

type ChecklistItem struct {
	ID, Text, Due       string
	Completed, Disabled bool
	Priority            uint8
	IconImage           *canvas.TiledImage
}
type ChecklistEvent struct {
	ID, ItemID string
	Action     ChecklistAction
	Checked    bool
	Text       string
}

// Checklist is controlled and owner-only. Change proposes completion or rename;
// the app owns data/history. Due is caller-formatted text; no clocks/storage are
// created. Retain one model; callbacks recheck current item identity/availability.
type Checklist struct {
	ID, Label                string
	Items                    []ChecklistItem
	Editable, Disabled, Dark bool
	Change                   func(ChecklistEvent)
	Invalid                  func(error)
	Style                    Style
}

func ChecklistCheckID(id, item string) string {
	return id + "/checklist/check/" + base64.RawURLEncoding.EncodeToString([]byte(item))
}
func ChecklistNameID(id, item string) string {
	return id + "/checklist/name/" + base64.RawURLEncoding.EncodeToString([]byte(item))
}
func (c *Checklist) validate() error {
	if c == nil || !validTokenText(c.ID) || len(c.Items) > 128 {
		return fmt.Errorf("ui: checklist requires ID and at most128 items")
	}
	seen := map[string]bool{}
	for _, i := range c.Items {
		if !validTokenText(i.ID) || !validTokenText(i.Text) || !paletteText(i.Due, 64, true) || i.Priority > 3 || seen[i.ID] {
			return fmt.Errorf("ui: invalid checklist item")
		}
		seen[i.ID] = true
	}
	return nil
}
func (c *Checklist) View() (Node, error) {
	if e := c.validate(); e != nil {
		return Node{}, e
	}
	id := c.ID
	p := MacPalette(c.Dark, color.NRGBA{})
	propose := func(item string, action ChecklistAction, checked bool, text string) {
		if c.ID != id || c.Disabled || c.Change == nil || c.validate() != nil {
			return
		}
		for _, v := range c.Items {
			if v.ID != item || v.Disabled {
				continue
			}
			if action == ChecklistRename {
				if !c.Editable {
					return
				}
				if !validTokenText(text) {
					if c.Invalid != nil {
						c.Invalid(fmt.Errorf("ui: checklist name must be nonblank and at most256 bytes"))
					}
					return
				}
			}
			c.Change(ChecklistEvent{ID: id, ItemID: item, Action: action, Checked: checked, Text: text})
			return
		}
	}
	root := Node{ID: id, Kind: Column, Disabled: c.Disabled, Accessibility: Accessibility{Name: c.Label}, Style: Style{Gap: Some(float32(0)), GrowX: Some(float32(1)), Background: Some(p.Field), Padding: Some(UniformInsets(8)), CornerRadius: Some(float32(8))}.Overlay(c.Style)}
	if c.Label != "" {
		root.Children = append(root.Children, Node{ID: id + "/checklist/title", Kind: Label, Text: c.Label, Style: Style{Typography: Some(MacTypography(17, 600)), Padding: Some(Insets{Bottom: 10})}})
	}
	for _, item := range c.Items {
		key := base64.RawURLEncoding.EncodeToString([]byte(item.ID))
		prefix := id + "/checklist/row/" + key
		ink := p.Ink
		if item.Completed {
			ink = p.Secondary
		}
		symbol := SymbolCircle
		if item.Completed {
			symbol = SymbolCircleCheck
		}
		row := Node{ID: prefix, Kind: Row, Disabled: item.Disabled, Style: Style{Gap: Some(float32(9)), Padding: Some(Insets{Top: 7, Right: 6, Bottom: 7, Left: 6}), Align: Some(AlignCenter), GrowX: Some(float32(1))}}
		row.Children = append(row.Children, Node{ID: ChecklistCheckID(id, item.ID), Kind: ToggleButton, Icon: symbol, Checked: item.Completed, Disabled: c.Change == nil, Style: Style{MinSize: Some(Size{Width: 24, Height: 24}), MaxSize: Some(Size{Width: 24, Height: 24}), Padding: Some(Insets{}), Border: Some(Border{}), Background: Some(color.NRGBA{}), Foreground: Some(p.Accent)}, Accessibility: Accessibility{Name: "Complete " + item.Text, Help: "Space toggles completion"}, Change: func(e ChangeEvent) { propose(item.ID, ChecklistToggle, e.Checked, "") }})
		if item.IconImage != nil {
			row.Children = append(row.Children, Node{ID: prefix + "/icon", Kind: Image, Image: item.IconImage, Fit: ImageContain, Text: "Category", Style: Style{MinSize: Some(Size{Width: 20, Height: 20}), MaxSize: Some(Size{Width: 20, Height: 20})}})
		}
		name := Node{ID: ChecklistNameID(id, item.ID), Kind: Label, Text: item.Text, Grow: true, Style: Style{Foreground: Some(ink), Typography: Some(MacTypography(13, 400)), Padding: Some(Insets{}), Border: Some(Border{}), Background: Some(color.NRGBA{})}}
		if c.Editable {
			name.Kind = TextField
			name.Disabled = c.Change == nil
			name.Accessibility.Name = "Name for " + item.Text
			name.Change = func(e ChangeEvent) { propose(item.ID, ChecklistRename, false, e.Text) }
		}
		row.Children = append(row.Children, name)
		if item.Due != "" {
			row.Children = append(row.Children, Node{ID: prefix + "/due", Kind: Label, Text: item.Due, Style: Style{Foreground: Some(p.Secondary), Typography: Some(MacTypography(11, 400))}})
		}
		if item.Priority > 0 {
			marks := []string{"", "!", "!!", "!!!"}
			row.Children = append(row.Children, Node{ID: prefix + "/priority", Kind: Label, Text: marks[item.Priority], Accessibility: Accessibility{Name: fmt.Sprintf("Priority %d", item.Priority)}, Style: Style{Foreground: Some(p.Accent), Typography: Some(MacTypography(13, 600))}})
		}
		root.Children = append(root.Children, row)
	}
	if e := Validate(root); e != nil {
		return Node{}, e
	}
	return root, nil
}
