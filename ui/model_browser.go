package ui

import (
	"fmt"
	"github.com/dshills/vellum/canvas"
	"image/color"
)

type ModelBrowserItem struct {
	ID, Parent, Label, Badge       string
	Icon                           Symbol
	IconImage                      *canvas.TiledImage
	Visible, Activatable, Disabled bool
}
type ModelBrowserEventKind string

const (
	ModelBrowserSelect     ModelBrowserEventKind = "select"
	ModelBrowserVisibility ModelBrowserEventKind = "visibility"
	ModelBrowserActivate   ModelBrowserEventKind = "activate"
	ModelBrowserExpand     ModelBrowserEventKind = "expand"
)

type ModelBrowserEvent struct {
	ID, ItemID string
	Kind       ModelBrowserEventKind
	Value      bool
}

// ModelBrowser supplies independent selection, visibility, activation and
// disclosure targets per model row. It is intentionally bounded to 256 items and 16 ancestors per item;
// the separate DataTree remains the indexed path for larger datasets.
type ModelBrowser struct {
	ID, Label, SelectedID, ActiveID string
	Items                           []ModelBrowserItem
	ExpandedIDs                     []string
	Disabled                        bool
	Change                          func(ModelBrowserEvent)
	Style                           Style
}

func ModelBrowserRowID(root, item string) string { return root + "/browser/" + ribbonPart(item) }
func (p *ModelBrowser) items() (map[string]ModelBrowserItem, map[string]bool, error) {
	if p == nil || !validTokenText(p.ID) || !validTokenText(p.Label) || len(p.Items) > 256 {
		return nil, nil, fmt.Errorf("ui: invalid model browser")
	}
	by := map[string]ModelBrowserItem{}
	expanded := map[string]bool{}
	for _, i := range p.Items {
		if !validTokenText(i.ID) || !validTokenText(i.Label) || by[i.ID].ID != "" {
			return nil, nil, fmt.Errorf("ui: duplicate/invalid browser item")
		}
		by[i.ID] = i
	}
	for _, i := range p.Items {
		visited := map[string]bool{}
		for parent := i.Parent; parent != ""; parent = by[parent].Parent {
			if by[parent].ID == "" || visited[parent] || parent == i.ID || len(visited) >= 16 {
				return nil, nil, fmt.Errorf("ui: invalid browser ancestry")
			}
			visited[parent] = true
		}
	}
	for _, id := range p.ExpandedIDs {
		if by[id].ID == "" || expanded[id] {
			return nil, nil, fmt.Errorf("ui: invalid browser expansion")
		}
		expanded[id] = true
	}
	if p.SelectedID != "" && by[p.SelectedID].ID == "" || p.ActiveID != "" && (by[p.ActiveID].ID == "" || !by[p.ActiveID].Activatable) {
		return nil, nil, fmt.Errorf("ui: invalid browser selection/active item")
	}
	return by, expanded, nil
}
func (p *ModelBrowser) visible(id string, by map[string]ModelBrowserItem, expanded map[string]bool) bool {
	for parent := by[id].Parent; parent != ""; parent = by[parent].Parent {
		if !expanded[parent] {
			return false
		}
	}
	return true
}
func (p *ModelBrowser) propose(id string, kind ModelBrowserEventKind, value bool) {
	if p.Disabled || p.Change == nil {
		return
	}
	by, expanded, e := p.items()
	if e != nil || by[id].ID == "" || by[id].Disabled || !p.visible(id, by, expanded) {
		return
	}
	for parent := by[id].Parent; parent != ""; parent = by[parent].Parent {
		if by[parent].Disabled {
			return
		}
	}
	if kind == ModelBrowserExpand {
		hasChild := false
		for _, item := range by {
			hasChild = hasChild || item.Parent == id
		}
		if !hasChild {
			return
		}
	}
	if kind == ModelBrowserActivate && !by[id].Activatable {
		return
	}
	p.Change(ModelBrowserEvent{p.ID, id, kind, value})
}
func (p *ModelBrowser) View() (Node, error) {
	by, expanded, e := p.items()
	if e != nil {
		return Node{}, e
	}
	rootID := p.ID
	propose := func(id string, k ModelBrowserEventKind, v bool) {
		if p.ID == rootID {
			p.propose(id, k, v)
		}
	}
	children := map[string][]ModelBrowserItem{}
	for _, i := range p.Items {
		children[i.Parent] = append(children[i.Parent], i)
	}
	rows := Node{ID: p.ID + "/browser/rows", Kind: Column, Style: Style{Gap: Some(float32(0)), GrowX: Some(float32(1))}}
	var appendItems func(string, int)
	appendItems = func(parent string, depth int) {
		for _, item := range children[parent] {
			id := item.ID
			rid := ModelBrowserRowID(p.ID, id)
			disabled := item.Disabled || p.Change == nil
			for parent := item.Parent; parent != ""; parent = by[parent].Parent {
				disabled = disabled || by[parent].Disabled
			}
			row := Node{ID: rid, Kind: Row, Disabled: disabled, Style: Style{Padding: Some(Insets{Left: float32(3 + depth*10), Right: 3, Top: 0, Bottom: 0}), Gap: Some(float32(4)), CornerRadius: Some(float32(2)), GrowX: Some(float32(1)), GrowY: Some(float32(0))}}
			if p.SelectedID == id {
				row.Style.Background = Some(color.NRGBA{R: 222, G: 234, B: 249, A: 255})
				row.Style.Foreground = Some(color.NRGBA{R: 26, G: 48, B: 76, A: 255})
			}
			quiet := Style{Padding: Some(Insets{}), Border: Some(Border{}), Background: Some(color.NRGBA{}), Typography: Some(MacTypography(10, 400)), MinSize: Some(Size{Width: 20, Height: 24}), MaxSize: Some(Size{Width: 20, Height: 24}), GrowX: Some(float32(0))}
			if len(children[id]) > 0 {
				icon := SymbolRight
				if expanded[id] {
					icon = SymbolDown
				}
				row.Children = append(row.Children, Node{ID: rid + "/expand", Kind: Button, Icon: icon, Appearance: ButtonTool, Style: quiet, Accessibility: Accessibility{Name: "Expand " + item.Label}, Activate: func() {
					_, current, err := p.items()
					if err == nil {
						propose(id, ModelBrowserExpand, !current[id])
					}
				}})
			} else {
				row.Children = append(row.Children, Node{ID: rid + "/indent", Kind: Column, Style: quiet})
			}
			row.Children = append(row.Children, Node{ID: rid + "/visibility", Kind: ToggleButton, Icon: SymbolVisibility, Checked: item.Visible, Style: quiet, Accessibility: Accessibility{Name: "Visibility of " + item.Label}, Change: func(e ChangeEvent) { propose(id, ModelBrowserVisibility, e.Checked) }})
			selectStyle := Style{Padding: Some(Insets{Left: 1, Right: 1}), Gap: Some(float32(5)), Border: Some(Border{}), Background: Some(color.NRGBA{}), Typography: Some(MacTypography(12, 400)), GrowX: Some(float32(1)), TextAlign: Some(AlignStart), MinSize: Some(Size{Width: 48, Height: 24})}
			if p.SelectedID == id {
				selectStyle.Foreground = Some(color.NRGBA{R: 26, G: 48, B: 76, A: 255})
			}
			row.Children = append(row.Children, Node{ID: rid + "/select", Kind: Button, Text: item.Label, Icon: item.Icon, IconImage: item.IconImage, Appearance: ButtonTool, Style: selectStyle, Accessibility: Accessibility{Name: "Select " + item.Label}, Activate: func() { propose(id, ModelBrowserSelect, true) }})
			if item.Badge != "" {
				row.Children = append(row.Children, Node{ID: rid + "/badge", Kind: Label, Text: item.Badge, Style: Style{Typography: Some(MacTypography(10, 400)), Padding: Some(Insets{}), GrowX: Some(float32(0))}})
			}
			if item.Activatable {
				row.Children = append(row.Children, Node{ID: rid + "/active", Kind: ToggleButton, Icon: func() Symbol {
					if p.ActiveID == id {
						return SymbolCircleCheck
					}
					return SymbolCircle
				}(), Checked: p.ActiveID == id, Style: quiet, Accessibility: Accessibility{Name: "Activate " + item.Label}, Change: func(ChangeEvent) { propose(id, ModelBrowserActivate, true) }})
			}
			rows.Children = append(rows.Children, row)
			if expanded[id] {
				appendItems(id, depth+1)
			}
		}
	}
	appendItems("", 0)
	root := Node{ID: p.ID, Kind: Column, Disabled: p.Disabled, Accessibility: Accessibility{Name: p.Label}, Style: Style{Padding: Some(UniformInsets(4)), Gap: Some(float32(3)), GrowX: Some(float32(1)), GrowY: Some(float32(1))}.Overlay(p.Style), Children: []Node{{ID: p.ID + "/browser/title", Kind: Label, Text: p.Label, Style: Style{Typography: Some(MacTypography(10, 600)), Padding: Some(Insets{Left: 4, Top: 3, Bottom: 3}), GrowY: Some(float32(0)), GrowX: Some(float32(1))}}, {ID: p.ID + "/browser/scroll", Kind: Scroll, Grow: true, Children: []Node{rows}}}}
	if e := Validate(root); e != nil {
		return Node{}, e
	}
	return root, nil
}
