package ui

import (
	"fmt"
	"strings"
)

// ToolbarOverflowID is the stable More actions button identity added by
// WithToolbarOverflow. Reserve this child ID in the application toolbar.
func ToolbarOverflowID(id string) string { return id + "/overflow" }

// WithToolbarOverflow returns a detached child list with a managed More actions
// menu. The native adapter fits a prefix and moves the remaining actions into
// this menu as available width changes. Supply a horizontal Toolbar containing
// 1..64 Buttons/Separators/toolbar spaces, with a caption or accessible name on each action.
// No editors, nested controls, placement sources or context menus are supported.
// Call once per fresh toolbar description, after applying per-item presentation.
// Validation rejects unsupported descriptions; ordinary Toolbar keeps scrolling.
func WithToolbarOverflow(toolbar Node) Node {
	toolbar.ToolbarOverflow = true
	items := append([]Node(nil), toolbar.Children...)
	options := make([]Option, 0, len(items))
	for _, item := range items {
		label := item.Text
		if strings.TrimSpace(label) == "" {
			label = item.Accessibility.Name
		}
		options = append(options, Option{ID: item.ID, Label: label, Icon: item.Icon, IconImage: item.IconImage,
			Disabled: item.Disabled || item.Activate == nil, Separator: item.Kind == Separator})
		if item.Kind == Separator || IsToolbarSpace(item.Kind) {
			options[len(options)-1] = Option{ID: item.ID, Separator: true}
		}
	}
	more := Node{ID: ToolbarOverflowID(toolbar.ID), Kind: MenuButton, Text: "»",
		Appearance: ButtonTool, Accessibility: Accessibility{Name: "More actions", Help: "Show toolbar actions that do not fit"},
		Style:   Style{GrowX: Some(float32(0)), GrowY: Some(float32(0)), MinSize: Some(Size{Width: 28, Height: 28}), Padding: Some(UniformInsets(4))},
		Options: options, Change: func(e ChangeEvent) {
			if toolbar.Disabled {
				return
			}
			for _, item := range items {
				if item.ID == e.SelectedID && item.Kind == Button && !item.Disabled && item.Activate != nil {
					item.Activate()
					return
				}
			}
		}}
	toolbar.Children = append(items, more)
	return toolbar
}

func validateToolbarOverflow(n Node) error {
	if !n.ToolbarOverflow {
		return nil
	}
	if n.Kind != Toolbar || n.Orientation == Vertical || len(n.Children) < 2 || len(n.Children) > MaxMenuOptions+1 {
		return fmt.Errorf("ui: toolbar overflow requires a horizontal action Toolbar with 1..64 items")
	}
	more := n.Children[len(n.Children)-1]
	if more.Kind != MenuButton || more.ID != ToolbarOverflowID(n.ID) || len(more.Options) != len(n.Children)-1 {
		return fmt.Errorf("ui: use WithToolbarOverflow to create the managed overflow menu")
	}
	for i, child := range n.Children[:len(n.Children)-1] {
		if (child.Kind != Button && child.Kind != Separator && !IsToolbarSpace(child.Kind)) || child.ContextMenu != nil || child.PlacementSource != nil || len(child.Children) != 0 {
			return fmt.Errorf("ui: overflow toolbar items must be ordinary Buttons, Separators or toolbar spaces")
		}
		if child.Kind == Button && strings.TrimSpace(child.Text) == "" && strings.TrimSpace(child.Accessibility.Name) == "" {
			return fmt.Errorf("ui: overflow toolbar actions need a label or accessible name")
		}
		if more.Options[i].ID != child.ID {
			return fmt.Errorf("ui: overflow options must match toolbar item order")
		}
	}
	return nil
}
