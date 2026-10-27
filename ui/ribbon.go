package ui

import (
	"encoding/base64"
	"fmt"
	"github.com/dshills/vellum/canvas"
	"image/color"
	"maps"
	"slices"
	"strings"
)

type RibbonAction struct {
	ID, Label string
	Icon      Symbol
	IconImage *canvas.TiledImage
	Disabled  bool
	Menu      []Option
}
type RibbonGroup struct {
	ID, Label string
	Disabled  bool
	Actions   []RibbonAction
}
type RibbonTab struct {
	ID, Label            string
	Disabled, Contextual bool
	Groups               []RibbonGroup
}
type RibbonEvent struct{ ID, TabID, ActionID, OptionID string }
type ribbonBranch struct {
	items  []Option
	branch *MenuBranch
}

// Ribbon is an owner-held contextual action bar. Whole measured groups overflow
// together, including their split/dropdown actions and full-color artwork.
// Retain the instance for stable submenu identities. Menus retain the ordinary
// 64-entry/depth-four bound; at most eight tabs/eight groups per tab are admitted.
type Ribbon struct {
	ID, Label, SelectedTab string
	Tabs                   []RibbonTab
	Disabled               bool
	SelectTab              func(string)
	Action                 func(RibbonEvent)
	Style                  Style
	branches               map[string]ribbonBranch
}

func ribbonPart(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }
func RibbonActionID(root, tab, action string) string {
	return root + "/ribbon/" + ribbonPart(tab) + "/action/" + ribbonPart(action)
}
func (p *Ribbon) dispatch(tab, action, option string) {
	if p.Disabled || p.SelectedTab != tab || p.Action == nil {
		return
	}
	for _, t := range p.Tabs {
		if t.ID != tab || t.Disabled {
			continue
		}
		for _, g := range t.Groups {
			if g.Disabled {
				continue
			}
			for _, a := range g.Actions {
				if a.ID != action || a.Disabled {
					continue
				}
				if option != "" {
					if _, ok := MenuAction(a.Menu, option); !ok {
						return
					}
				}
				p.Action(RibbonEvent{p.ID, tab, action, option})
				return
			}
		}
	}
}
func (p *Ribbon) branch(key string, items []Option, used map[string]bool) (*MenuBranch, error) {
	used[key] = true
	if c, ok := p.branches[key]; ok && slices.Equal(c.items, items) {
		return c.branch, nil
	}
	b, e := NewMenuBranch(items)
	if e != nil {
		return nil, e
	}
	p.branches[key] = ribbonBranch{slices.Clone(items), b}
	return b, nil
}
func (p *Ribbon) View() (Node, error) {
	if p == nil || !validTokenText(p.ID) || !validTokenText(p.Label) || len(p.Tabs) < 1 || len(p.Tabs) > 8 {
		return Node{}, fmt.Errorf("ui: ribbon needs ID/label and 1..8 tabs")
	}
	oldBranches := p.branches
	p.branches = maps.Clone(oldBranches)
	if p.branches == nil {
		p.branches = map[string]ribbonBranch{}
	}
	success := false
	defer func() {
		if !success {
			p.branches = oldBranches
		}
	}()
	rootID := p.ID
	dispatch := func(tab, action, option string) {
		if p.ID == rootID {
			p.dispatch(tab, action, option)
		}
	}
	used := map[string]bool{}
	seen := map[string]bool{}
	found := false
	root := Node{ID: p.ID, Kind: Tabs, SelectedID: p.SelectedTab, Disabled: p.Disabled, Accessibility: Accessibility{Name: p.Label}, Style: Style{Padding: Some(Insets{}), GrowX: Some(float32(1)), GrowY: Some(float32(0))}.Overlay(p.Style)}
	root.Change = func(e ChangeEvent) {
		if p.Disabled || p.SelectTab == nil {
			return
		}
		for _, t := range p.Tabs {
			if t.ID == e.SelectedID && !t.Disabled {
				p.SelectTab(t.ID)
				return
			}
		}
	}
	for _, tab := range p.Tabs {
		if !validTokenText(tab.ID) || !validTokenText(tab.Label) || seen[tab.ID] || len(tab.Groups) < 1 || len(tab.Groups) > 8 {
			return Node{}, fmt.Errorf("ui: invalid ribbon tab/groups")
		}
		seen[tab.ID] = true
		found = found || tab.ID == p.SelectedTab
		caption := tab.Label
		if tab.Contextual {
			caption += " •"
		}
		pageID := p.ID + "/ribbon/" + ribbonPart(tab.ID)
		root.Options = append(root.Options, Option{ID: pageID, Label: caption, Disabled: tab.Disabled})
		// Page identities stay namespaced; external tab IDs remain application values.
		if tab.ID == p.SelectedTab {
			root.SelectedID = pageID
		}
		bar := Node{ID: pageID, Kind: Toolbar, ToolbarOverflow: true, ToolbarGroups: true, Disabled: tab.Disabled, Style: Style{Padding: Some(Insets{Top: 3, Right: 6, Bottom: 2, Left: 6}), Gap: Some(float32(2)), GrowX: Some(float32(1)), GrowY: Some(float32(0))}}
		var groupOptions []Option
		groupSeen := map[string]bool{}
		actionSeen := map[string]bool{}
		routes := map[string]struct{ action, option string }{}
		for groupIndex, group := range tab.Groups {
			if !validTokenText(group.ID) || !validTokenText(group.Label) || groupSeen[group.ID] || len(group.Actions) < 1 || len(group.Actions) > 12 {
				return Node{}, fmt.Errorf("ui: invalid ribbon group")
			}
			groupSeen[group.ID] = true
			gid := pageID + "/group/" + ribbonPart(group.ID)
			row := Node{ID: gid + "/actions", Kind: Row, Style: Style{Gap: Some(float32(3)), GrowX: Some(float32(0)), GrowY: Some(float32(0))}}
			var items []Option
			for _, action := range group.Actions {
				if !validTokenText(action.ID) || !validTokenText(action.Label) || actionSeen[action.ID] {
					return Node{}, fmt.Errorf("ui: duplicate/invalid ribbon action")
				}
				actionSeen[action.ID] = true
				if len(action.Menu) > 0 {
					if _, e := NewMenuBranch(action.Menu); e != nil {
						return Node{}, e
					}
				}
				aid := RibbonActionID(p.ID, tab.ID, action.ID)
				disabled := action.Disabled || group.Disabled || p.Action == nil
				tabID, actionID := tab.ID, action.ID
				main := Node{ID: aid, Kind: Button, Text: action.Label, Icon: action.Icon, IconImage: action.IconImage, Appearance: ButtonTool, Disabled: disabled, IconPlacement: IconAboveText, Accessibility: Accessibility{Name: action.Label}, Style: Style{MinSize: Some(Size{Width: 52, Height: 42}), MaxSize: Some(Size{Width: 82, Height: 42}), Padding: Some(Insets{Top: 2, Right: 5, Bottom: 2, Left: 5}), Gap: Some(float32(2)), Border: Some(Border{}), Background: Some(color.NRGBA{}), CornerRadius: Some(float32(6)), Typography: Some(MacTypography(11, 400)), GrowX: Some(float32(0)), GrowY: Some(float32(0))}, Activate: func() { dispatch(tabID, actionID, "") }}
				if action.Icon == "" && action.IconImage == nil {
					main.IconPlacement = IconLeading
				}
				routes[aid] = struct{ action, option string }{action.ID, ""}
				items = append(items, Option{ID: aid, Label: action.Label, Icon: action.Icon, IconImage: action.IconImage, Disabled: disabled})
				if len(action.Menu) == 0 {
					row.Children = append(row.Children, main)
					continue
				}
				// Namespace a copy of each source menu; all original IDs remain in events.
				var prefix func([]Option) ([]Option, error)
				prefix = func(src []Option) ([]Option, error) {
					out := make([]Option, len(src))
					for i, o := range src {
						out[i] = o
						out[i].ID = aid + "/menu/" + ribbonPart(o.ID)
						routes[out[i].ID] = struct{ action, option string }{actionID, o.ID}
						if o.Branch != nil {
							nested, e := prefix(o.Branch.Items())
							if e != nil {
								return nil, e
							}
							out[i].Branch, e = p.branch(out[i].ID, nested, used)
							if e != nil {
								return nil, e
							}
						}
					}
					return out, nil
				}
				menu, e := prefix(action.Menu)
				if e != nil {
					return Node{}, e
				}
				branch, e := p.branch(aid+"/overflow", menu, used)
				if e != nil {
					return Node{}, e
				}
				items = append(items, Option{ID: aid + "/options", Label: action.Label + " options", Icon: action.Icon, IconImage: action.IconImage, Disabled: disabled, Branch: branch})
				button := Node{ID: aid + "/menu-button", Kind: MenuButton, Icon: SymbolDown, Accessibility: Accessibility{Name: action.Label + " options"}, Options: menu, Disabled: disabled, Appearance: ButtonTool, Style: Style{MinSize: Some(Size{Width: 20, Height: 42}), MaxSize: Some(Size{Width: 20, Height: 42}), Padding: Some(Insets{}), Border: Some(Border{}), GrowX: Some(float32(0))}, Change: func(e ChangeEvent) {
					if route, ok := routes[e.SelectedID]; ok {
						dispatch(tabID, route.action, route.option)
					}
				}}
				row.Children = append(row.Children, Node{ID: aid + "/split", Kind: Row, Children: []Node{main, button}, Style: Style{Gap: Some(float32(0)), GrowX: Some(float32(0))}})
			}
			// The divider belongs to its measured group, so collapse/overflow keeps
			// one complete unit and never strands an independently routed action.
			content := Node{ID: gid + "/content", Kind: Row, Children: []Node{row}, Style: Style{Padding: Some(Insets{}), Gap: Some(float32(5)), GrowX: Some(float32(0)), GrowY: Some(float32(0))}}
			if groupIndex < len(tab.Groups)-1 {
				content.Children = append(content.Children, Node{ID: gid + "/divider", Kind: Column, Style: Style{Padding: Some(Insets{}), MinSize: Some(Size{Width: 1, Height: 40}), MaxSize: Some(Size{Width: 1, Height: 40}), Grow: Some(float32(0)), Background: Some(color.NRGBA{R: 208, G: 213, B: 221, A: 255})}})
			}
			groupNode := Node{ID: gid, Kind: Column, Disabled: group.Disabled, Accessibility: Accessibility{Name: group.Label}, Children: []Node{content, {ID: gid + "/caption", Kind: Label, Text: strings.ToUpper(group.Label), Style: Style{Typography: Some(MacTypography(10, 400)), TextAlign: Some(AlignCenter), GrowX: Some(float32(1)), Padding: Some(Insets{})}}}, Style: Style{Padding: Some(Insets{Left: 3, Right: 3}), Gap: Some(float32(0)), GrowX: Some(float32(0)), GrowY: Some(float32(0))}}
			bar.Children = append(bar.Children, groupNode)
			branch, e := p.branch(gid, items, used)
			if e != nil {
				return Node{}, e
			}
			groupOptions = append(groupOptions, Option{ID: gid, Label: group.Label, Branch: branch, Disabled: group.Disabled})
		}
		tabID := tab.ID
		bar.Children = append(bar.Children, Node{ID: ToolbarOverflowID(pageID), Kind: MenuButton, Text: "»", Options: groupOptions, Accessibility: Accessibility{Name: "More " + tab.Label + " tools"}, Disabled: p.Action == nil, Style: Style{GrowX: Some(float32(0)), MinSize: Some(Size{Width: 28, Height: 42})}, Change: func(e ChangeEvent) {
			if route, ok := routes[e.SelectedID]; ok {
				dispatch(tabID, route.action, route.option)
			}
		}})
		root.Children = append(root.Children, bar)
	}
	if !found {
		return Node{}, fmt.Errorf("ui: missing selected ribbon tab")
	}
	root.Change = func(e ChangeEvent) {
		if p.Disabled || p.SelectTab == nil {
			return
		}
		for _, t := range p.Tabs {
			if !t.Disabled && p.ID+"/ribbon/"+ribbonPart(t.ID) == e.SelectedID {
				p.SelectTab(t.ID)
				return
			}
		}
	}
	for key := range p.branches {
		if !used[key] {
			delete(p.branches, key)
		}
	}
	if e := Validate(root); e != nil {
		return Node{}, e
	}
	success = true
	return root, nil
}
