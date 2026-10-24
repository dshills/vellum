package ui

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"slices"
	"strings"
)

// ToolbarLayout is caller-owned, serializable toolbar configuration. Order may
// be partial: omitted current items follow in catalogue order. Hidden contains
// action/separator IDs, not labels. Unknown/duplicate IDs are rejected. Apply it
// to the fresh, complete Toolbar before WithToolbarOverflow. No disk I/O occurs.
type ToolbarLayout struct {
	Order  []string `json:"order,omitempty"`
	Hidden []string `json:"hidden,omitempty"`
}

func resolveToolbarLayout(toolbar Node, value ToolbarLayout) (ToolbarLayout, []Node, error) {
	if toolbar.Kind != Toolbar || toolbar.ToolbarOverflow || len(toolbar.Children) < 1 || len(toolbar.Children) > MaxMenuOptions {
		return ToolbarLayout{}, nil, fmt.Errorf("ui: toolbar customization requires a complete horizontal Toolbar with 1..64 actions, separators or spaces before overflow")
	}
	if toolbar.Orientation == Vertical {
		return ToolbarLayout{}, nil, fmt.Errorf("ui: toolbar customization requires horizontal actions")
	}
	if err := Validate(toolbar); err != nil {
		return ToolbarLayout{}, nil, err
	}
	byID := make(map[string]Node, len(toolbar.Children))
	actions := 0
	for _, item := range toolbar.Children {
		if (item.Kind != Button && item.Kind != Separator && !IsToolbarSpace(item.Kind)) || item.ContextMenu != nil || item.PlacementSource != nil || item.ToolbarReorder != nil || len(item.Children) > 0 {
			return ToolbarLayout{}, nil, fmt.Errorf("ui: customizable toolbar items must be ordinary Buttons, Separators or toolbar spaces")
		}
		if item.Kind == Button {
			actions++
			if strings.TrimSpace(toolbarItemLabel(item)) == "" {
				return ToolbarLayout{}, nil, fmt.Errorf("ui: customizable action needs a caption or accessible name")
			}
		}
		byID[item.ID] = item
	}
	if actions == 0 {
		return ToolbarLayout{}, nil, fmt.Errorf("ui: customizable toolbar needs at least one action")
	}
	if len(value.Order) > len(byID) || len(value.Hidden) > len(byID) {
		return ToolbarLayout{}, nil, fmt.Errorf("ui: toolbar layout exceeds catalogue size")
	}
	order := make([]string, 0, len(byID))
	seen := map[string]bool{}
	for _, id := range value.Order {
		if _, ok := byID[id]; !ok || seen[id] {
			return ToolbarLayout{}, nil, fmt.Errorf("ui: duplicate or unknown toolbar order ID %q", id)
		}
		seen[id] = true
		order = append(order, id)
	}
	for _, item := range toolbar.Children {
		if !seen[item.ID] {
			order = append(order, item.ID)
		}
	}
	hidden := map[string]bool{}
	for _, id := range value.Hidden {
		if _, ok := byID[id]; !ok || hidden[id] {
			return ToolbarLayout{}, nil, fmt.Errorf("ui: duplicate or unknown hidden toolbar ID %q", id)
		}
		hidden[id] = true
	}
	resolved := ToolbarLayout{Order: order}
	items := make([]Node, 0, len(order))
	for _, id := range order {
		items = append(items, byID[id])
		if hidden[id] {
			resolved.Hidden = append(resolved.Hidden, id)
		}
	}
	return resolved, items, nil
}

// Apply returns a detached ordered/visible child list. Captions, full-color
// artwork, styles and callbacks remain unchanged. Leading/trailing/adjacent
// separators are normalized. Hiding every action is valid and produces an empty
// Toolbar; keep the customization opener outside it and skip overflow when empty.
func (value ToolbarLayout) Apply(toolbar Node) (Node, error) {
	layout, items, err := resolveToolbarLayout(toolbar, value)
	if err != nil {
		return Node{}, err
	}
	toolbar.Children = make([]Node, 0, len(items))
	hidden := map[string]bool{}
	for _, id := range layout.Hidden {
		hidden[id] = true
	}
	for _, item := range items {
		if hidden[item.ID] {
			continue
		}
		if item.Kind == Separator && (len(toolbar.Children) == 0 || toolbar.Children[len(toolbar.Children)-1].Kind == Separator) {
			continue
		}
		toolbar.Children = append(toolbar.Children, item)
	}
	if n := len(toolbar.Children); n > 0 && toolbar.Children[n-1].Kind == Separator {
		toolbar.Children = toolbar.Children[:n-1]
	}
	if !slices.ContainsFunc(toolbar.Children, func(n Node) bool { return n.Kind == Button }) {
		toolbar.Children = nil
	}
	return toolbar, nil
}

func toolbarItemLabel(item Node) string {
	if IsToolbarSpace(item.Kind) {
		if item.Kind == ToolbarFlexibleSpace {
			return "Flexible Space"
		}
		return "Space"
	}
	if item.Kind == Separator {
		return "Separator"
	}
	if strings.TrimSpace(item.Text) != "" {
		return item.Text
	}
	return item.Accessibility.Name
}
func toolbarCustomizationItemID(root, item, action string) string {
	return root + "/toolbar-customization/" + action + "/" + base64.RawURLEncoding.EncodeToString([]byte(item))
}
func ToolbarCustomizationVisibilityID(root, item string) string {
	return toolbarCustomizationItemID(root, item, "show")
}
func ToolbarCustomizationEarlierID(root, item string) string {
	return toolbarCustomizationItemID(root, item, "earlier")
}
func ToolbarCustomizationLaterID(root, item string) string {
	return toolbarCustomizationItemID(root, item, "later")
}
func ToolbarCustomizationDoneID(root string) string { return root + "/toolbar-customization/done" }
func ToolbarCustomizationRestoreID(root string) string {
	return root + "/toolbar-customization/restore"
}

// ToolbarCustomizer is a retained, single-owner, controlled window-modal panel.
// Toolbar is always the COMPLETE catalogue, never the result of Layout.Apply or
// WithToolbarOverflow. Accept Change by assigning the proposed Value. Edits are
// live; Done/Escape dismiss without rolling them back. Retain one instance and
// render closed/disabled states to revoke sessions. The application owns optional
// persistence. This is a same-window content sheet, not NSToolbar customization.
type ToolbarCustomizer struct {
	ID, Title, ReturnFocusID string
	Toolbar                  Node
	Value                    ToolbarLayout
	Open, Disabled           bool
	Change                   func(ToolbarLayout)
	Dismiss                  func()
	Style                    Style
	epoch                    *toolbarCustomizationEpoch
	identity                 string
}
type toolbarCustomizationEpoch struct {
	marker  byte
	order   []string
	handles map[string]*ToolbarReorderHandle
}

// ToolbarReorderHandle is an immutable native-drag capability generated by
// ToolbarCustomizer.View. Applications use the customizer, rather than building
// handles directly. DropBefore runs on the application owner and proposes one
// reorder; an empty target means the end. Hidden actions remain in the order.
type ToolbarReorderHandle struct {
	id, listID string
	rows       []string
	drop       func(string)
}

func (h *ToolbarReorderHandle) ListID() string   { return h.listID }
func (h *ToolbarReorderHandle) RowIDs() []string { return slices.Clone(h.rows) }
func (h *ToolbarReorderHandle) DropBefore(rowID string) {
	if h != nil && h.drop != nil {
		h.drop(rowID)
	}
}
func ToolbarCustomizationDragID(root, item string) string {
	return toolbarCustomizationItemID(root, item, "drag")
}

// ValidateToolbarReorderNode confines this capability to its generated handle.
func ValidateToolbarReorderNode(n Node) error {
	h := n.ToolbarReorder
	if h == nil {
		return nil
	}
	if n.Kind != Button || h.id != n.ID || h.drop == nil || n.PlacementSource != nil || len(h.rows) == 0 || len(h.rows) > MaxMenuOptions {
		return fmt.Errorf("ui: invalid toolbar reorder handle")
	}
	return nil
}

func (c *ToolbarCustomizer) View() (result Node, err error) {
	if c == nil {
		return Node{}, fmt.Errorf("ui: toolbar customizer required")
	}
	defer func() {
		if err != nil {
			c.epoch = nil
		}
	}()
	if strings.TrimSpace(c.ID) == "" || c.Dismiss == nil {
		return Node{}, fmt.Errorf("ui: toolbar customizer needs an ID and Dismiss")
	}
	layout, items, err := resolveToolbarLayout(c.Toolbar, c.Value)
	if err != nil {
		return Node{}, err
	}
	identity := c.ID + "\x00" + c.Toolbar.ID
	if c.identity != identity || !c.Open || c.Disabled || c.Change == nil {
		c.epoch = nil
		c.identity = identity
	}
	available := c.Open && !c.Disabled && c.Change != nil
	if available && c.epoch == nil {
		c.epoch = &toolbarCustomizationEpoch{}
	}
	if available && !slices.Equal(c.epoch.order, layout.Order) {
		c.epoch.order = slices.Clone(layout.Order)
		c.epoch.handles = make(map[string]*ToolbarReorderHandle, len(items))
	}
	id, toolbarID, epoch := c.ID, c.Toolbar.ID, c.epoch
	live := func() bool {
		return c.ID == id && c.Toolbar.ID == toolbarID && c.Open && !c.Disabled && c.Change != nil && c.epoch == epoch && epoch != nil
	}
	emit := func(edit func(*ToolbarLayout) bool) {
		if !live() {
			return
		}
		current, _, e := resolveToolbarLayout(c.Toolbar, c.Value)
		if e != nil {
			return
		}
		before := ToolbarLayout{Order: slices.Clone(current.Order), Hidden: slices.Clone(current.Hidden)}
		if !edit(&current) {
			return
		}
		normalized, _, e := resolveToolbarLayout(c.Toolbar, current)
		if e != nil {
			return
		}
		if !slices.Equal(before.Order, normalized.Order) || !slices.Equal(before.Hidden, normalized.Hidden) {
			c.Change(normalized)
		}
	}
	close := func() {
		if live() {
			c.epoch = nil
			if c.Dismiss != nil {
				c.Dismiss()
			}
		}
	}
	hidden := map[string]bool{}
	for _, item := range layout.Hidden {
		hidden[item] = true
	}
	rows := make([]Node, 0, len(items))
	for index, item := range items {
		itemID, label := item.ID, toolbarItemLabel(item)
		art := Node{ID: toolbarCustomizationItemID(id, itemID, "art"), Kind: Icon, Icon: SymbolRemove, Style: Style{MinSize: Some(Size{Width: 22, Height: 22}), MaxSize: Some(Size{Width: 22, Height: 22}), GrowX: Some(float32(0))}}
		if item.IconImage != nil {
			art.Kind = Image
			art.Icon = ""
			art.Image = item.IconImage
			art.Fit = ImageContain
		} else if item.Icon != "" {
			art.Kind = Icon
			art.Icon = item.Icon
		}
		show := Node{ID: ToolbarCustomizationVisibilityID(id, itemID), Kind: Checkbox, Text: label, Checked: !hidden[itemID], Style: Style{GrowX: Some(float32(1)), Justify: Some(AlignStart)}, Accessibility: Accessibility{Name: "Show " + label, Help: "Include this item in the toolbar"}, Change: func(e ChangeEvent) {
			emit(func(value *ToolbarLayout) bool {
				if slices.Index(value.Order, itemID) < 0 {
					return false
				}
				value.Hidden = slices.DeleteFunc(value.Hidden, func(id string) bool { return id == itemID })
				if !e.Checked {
					value.Hidden = append(value.Hidden, itemID)
				}
				return true
			})
		}}
		move := func(delta int) func() {
			return func() {
				emit(func(value *ToolbarLayout) bool {
					at := slices.Index(value.Order, itemID)
					next := at + delta
					if at < 0 || next < 0 || next >= len(value.Order) {
						return false
					}
					value.Order[at], value.Order[next] = value.Order[next], value.Order[at]
					return true
				})
			}
		}
		arrowStyle := Style{GrowX: Some(float32(0)), MinSize: Some(Size{Width: 22, Height: 22}), MaxSize: Some(Size{Width: 22, Height: 22}), Padding: Some(UniformInsets(2))}
		var drag *ToolbarReorderHandle
		if available {
			drag = epoch.handles[itemID]
			if drag == nil {
				order := slices.Clone(layout.Order)
				rowIDs := make([]string, len(order))
				for i, key := range order {
					rowIDs[i] = toolbarCustomizationItemID(id, key, "row")
				}
				drag = &ToolbarReorderHandle{id: ToolbarCustomizationDragID(id, itemID), listID: id + "/items", rows: rowIDs}
				drag.drop = func(beforeRow string) {
					emit(func(value *ToolbarLayout) bool {
						// Reordering/removal during capture revokes the old geometry snapshot.
						if epoch.handles[itemID] != drag || !slices.Equal(value.Order, order) {
							return false
						}
						before := len(order)
						if beforeRow != "" {
							before = slices.Index(rowIDs, beforeRow)
							if before < 0 {
								return false
							}
						}
						from := slices.Index(value.Order, itemID)
						if from < before {
							before--
						}
						if before == from {
							return false
						}
						value.Order = slices.Delete(value.Order, from, from+1)
						value.Order = slices.Insert(value.Order, before, itemID)
						return true
					})
				}
				epoch.handles[itemID] = drag
			}
		}
		grip := Node{ID: ToolbarCustomizationDragID(id, itemID), Kind: Button, Icon: SymbolMenu, Appearance: ButtonTool, Style: arrowStyle, Disabled: !available, ToolbarReorder: drag, Accessibility: Accessibility{Name: "Drag " + label + " to reorder", Help: "Drag to a new position. Use the adjacent arrow buttons with the keyboard."}}
		rows = append(rows, Node{ID: toolbarCustomizationItemID(id, itemID, "row"), Kind: Row, Style: Style{GrowX: Some(float32(1)), Align: Some(AlignCenter), Gap: Some(float32(8)), Padding: Some(Insets{Top: 4, Bottom: 4})}, Children: []Node{grip, art, show,
			{ID: ToolbarCustomizationEarlierID(id, itemID), Kind: Button, Icon: SymbolUp, Appearance: ButtonTool, Disabled: index == 0, Style: arrowStyle, Accessibility: Accessibility{Name: "Move " + label + " earlier", Help: "Move one place toward the start"}, Activate: move(-1)},
			{ID: ToolbarCustomizationLaterID(id, itemID), Kind: Button, Icon: SymbolDown, Appearance: ButtonTool, Disabled: index == len(items)-1, Style: arrowStyle, Accessibility: Accessibility{Name: "Move " + label + " later", Help: "Move one place toward the end"}, Activate: move(1)},
		}})
	}
	title := c.Title
	if strings.TrimSpace(title) == "" {
		title = "Customize Toolbar"
	}
	done := ToolbarCustomizationDoneID(id)
	n := Node{ID: id, Kind: Dialog, Sheet: true, Text: title, Open: available, Disabled: c.Disabled, Dismiss: close, DialogDefaultID: done, DialogReturnFocusID: c.ReturnFocusID,
		Style: Style{Padding: Some(UniformInsets(20)), Gap: Some(float32(12)), MinSize: Some(Size{Width: 380}), MaxSize: Some(Size{Width: 480}), CornerRadius: Some(float32(12))}.Overlay(c.Style),
		Children: []Node{
			{ID: id + "/heading", Kind: Label, Text: title, Style: Style{Typography: Some(Typography{Size: 17, Weight: 600})}},
			{ID: id + "/help", Kind: Label, Text: "Choose the actions to show. Drag the grips or use the arrows to reorder. Changes apply immediately."},
			{ID: id + "/items", Kind: Scroll, Style: Style{GrowX: Some(float32(1)), MaxSize: Some(Size{Height: 300}), Gap: Some(float32(2))}, Children: rows},
			{ID: id + "/actions", Kind: Row, Style: Style{GrowX: Some(float32(1)), Justify: Some(AlignEnd), Gap: Some(float32(8))}, Children: []Node{
				{ID: ToolbarCustomizationRestoreID(id), Kind: Button, Text: "Restore Defaults", Appearance: ButtonSecondary, Activate: func() {
					emit(func(value *ToolbarLayout) bool { *value = ToolbarLayout{}; return true })
				}},
				{ID: done, Kind: Button, Text: "Done", Appearance: ButtonPrimary, Style: Style{Background: Some(color.NRGBA{R: 0, G: 112, B: 245, A: 255}), Foreground: Some(color.NRGBA{R: 255, G: 255, B: 255, A: 255})}, Activate: close},
			}},
		}}
	if err := Validate(n); err != nil {
		return Node{}, err
	}
	return n, nil
}
