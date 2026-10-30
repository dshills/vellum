package ui

import (
	"fmt"
	"image/color"
	"strings"
	"unicode/utf8"
)

// ComboEvent proposes literal text. OptionID is empty for typed text, otherwise
// the stable identity of a predefined option. Choosing never adds a new option.
type ComboEvent struct{ ID, Text, OptionID string }

// ComboBox combines the existing native-backed TextField and managed action menu.
// Retain on one owner. Text is controlled; nil Change disables both parts.
// Tab reaches the suggestions button; Down/Return opens its keyboard menu.
// The root owns root+"/combo/". At most MaxMenuOptions suggestions are supported.
type ComboBox struct {
	ID, Label, Text, Placeholder string
	Options                      []Option
	Disabled                     bool
	Change                       func(ComboEvent)
	Style                        Style
}

func ComboInputID(root string) string { return root + "/combo/input" }
func ComboMenuID(root string) string  { return root + "/combo/menu" }
func (c *ComboBox) View() (Node, error) {
	if c == nil || c.ID == "" || len(c.Options) > MaxMenuOptions {
		return Node{}, fmt.Errorf("ui: combo box requires an ID and at most %d suggestions", MaxMenuOptions)
	}
	seen := map[string]bool{}
	for _, o := range c.Options {
		if o.ID == "" || seen[o.ID] || o.Branch != nil || o.Shortcut != "" || o.Closable || o.Separator || o.Checkable || o.Checked || strings.TrimSpace(o.Label) == "" || !utf8.ValidString(o.Label) || strings.ContainsRune(o.Label, 0) {
			return Node{}, fmt.Errorf("ui: invalid combo suggestion")
		}
		seen[o.ID] = true
	}
	options := append([]Option(nil), c.Options...)
	id := c.ID
	label := c.Label
	if strings.TrimSpace(label) == "" {
		label = "Value"
	}
	available := func() bool { return c.ID == id && !c.Disabled && c.Change != nil }
	style := c.Style
	if !style.Gap.Set {
		style.Gap = Some(float32(0))
	}
	if !style.Align.Set {
		style.Align = Some(AlignCenter)
	}
	root := Node{ID: id, Kind: Row, Disabled: !available(), Style: style, Accessibility: Accessibility{Name: label}}
	// The single outer border is supplied by the caller's MacTheme input style.
	clear := Style{Border: Some(Border{}), Background: Some(color.NRGBA{}), MinSize: Some(Size{Height: 16}), MaxSize: Some(Size{}), GrowX: Some(float32(1)), GrowY: Some(float32(0)), Padding: Some(Insets{Top: 2, Bottom: 2, Left: 6, Right: 4})}
	root.Children = []Node{{ID: ComboInputID(id), Kind: TextField, Text: c.Text, Placeholder: c.Placeholder, Style: clear, Accessibility: Accessibility{Name: label}, Change: func(e ChangeEvent) {
		if available() {
			c.Change(ComboEvent{ID: id, Text: e.Text})
		}
	}}}
	if len(c.Options) > 0 {
		root.Children = append(root.Children, Node{ID: ComboMenuID(id), Kind: MenuButton, MenuAnchorID: id, Icon: SymbolDown, Appearance: ButtonTool, Options: options, Style: Style{Border: Some(Border{}), Background: Some(color.NRGBA{}), Padding: Some(UniformInsets(2)), MinSize: Some(Size{Width: 16, Height: 16}), MaxSize: Some(Size{Width: 20}), GrowX: Some(float32(0))}, Accessibility: Accessibility{Name: label + " suggestions"}, Change: func(e ChangeEvent) {
			if !available() {
				return
			}
			for _, o := range c.Options {
				if o.ID == e.SelectedID && !o.Disabled {
					valid := false
					for _, shown := range options {
						if shown == o {
							valid = true
							break
						}
					}
					if !valid {
						return
					}
					c.Change(ComboEvent{ID: id, Text: o.Label, OptionID: o.ID})
					return
				}
			}
		}})
	}
	if err := Validate(root); err != nil {
		return Node{}, err
	}
	return root, nil
}
