package ui

import (
	"encoding/base64"
	"fmt"
	"image/color"
)

// SidebarSection is a named, optionally collapsed collection of navigation items.
// Item IDs are unique across the entire sidebar; each Option can carry an icon/badge.
type SidebarSection struct {
	ID, Label           string
	Collapsed, Disabled bool
	Items               []Option
}
type SidebarEvent struct{ ID, ItemID string }
type SidebarDisclosureEvent struct {
	ID, SectionID string
	Expanded      bool
}

// Sidebar is a controlled, owner-held two-level navigator. Accept navigation by
// setting SelectedID, and expansion by updating Sections. Retain on one owner.
// Each section reuses the list keyboard behavior; Tab moves between sections.
type Sidebar struct {
	ID, Label, SelectedID string
	Sections              []SidebarSection
	Disabled, Dark        bool
	Navigate              func(SidebarEvent)
	Disclosure            func(SidebarDisclosureEvent)
	Style                 Style
}

func SidebarSectionID(root, section string) string {
	return root + "/sidebar/section/" + base64.RawURLEncoding.EncodeToString([]byte(section))
}
func SidebarListID(root, section string) string { return SidebarSectionID(root, section) + "/items" }
func (s *Sidebar) validate() error {
	if s == nil || s.ID == "" || len(s.Sections) > 32 {
		return fmt.Errorf("ui: sidebar requires ID and at most 32 sections")
	}
	sections, items := map[string]bool{}, map[string]bool{}
	total := 0
	for _, g := range s.Sections {
		if !validTokenText(g.ID) || !validTokenText(g.Label) || sections[g.ID] {
			return fmt.Errorf("ui: sidebar requires unique section IDs and nonblank headings")
		}
		sections[g.ID] = true
		for _, o := range g.Items {
			if items[o.ID] {
				return fmt.Errorf("ui: sidebar item IDs must be globally unique")
			}
			items[o.ID] = true
			total++
		}
		if err := ValidateControl(Node{ID: "check", Kind: ListBox, Multiple: true, Options: g.Items}); err != nil {
			return err
		}
	}
	if total > MaxListBoxOptions || s.SelectedID != "" && !items[s.SelectedID] {
		return fmt.Errorf("ui: sidebar selection must exist; at most 256 items")
	}
	return nil
}
func (s *Sidebar) View() (Node, error) {
	if err := s.validate(); err != nil {
		return Node{}, err
	}
	id := s.ID
	available := func() bool { return s.ID == id && !s.Disabled && s.validate() == nil }
	surface := color.NRGBA{R: 239, G: 239, B: 242, A: 255}
	if s.Dark {
		surface = color.NRGBA{R: 38, G: 38, B: 41, A: 255}
	}
	root := Node{ID: id, Kind: Scroll, Disabled: s.Disabled, Accessibility: Accessibility{Name: s.Label}, Style: Style{Background: Some(surface), Padding: Some(UniformInsets(8)), Gap: Some(float32(8)), MinSize: Some(Size{Width: 180}), GrowX: Some(float32(1)), GrowY: Some(float32(1))}.Overlay(s.Style)}
	for _, g := range s.Sections {
		shown := g
		options := append([]Option(nil), g.Items...)
		selected := []string{}
		for _, o := range options {
			if o.ID == s.SelectedID {
				selected = append(selected, o.ID)
			}
		}
		list := Node{ID: SidebarListID(id, g.ID), Kind: ListBox, Multiple: true, SelectedIDs: selected, Options: options, Disabled: s.Navigate == nil || g.Disabled, Style: Style{Background: Some(color.NRGBA{}), Border: Some(Border{}), Padding: Some(Insets{}), MinSize: Some(Size{Height: float32(min(len(options)*34+8, 240))}), GrowY: Some(float32(0))}, Accessibility: Accessibility{Name: g.Label}}
		list.SelectionChange = func(e ListSelectionEvent) {
			if !available() || s.Navigate == nil || len(e.IDs) != 1 {
				return
			}
			for _, current := range s.Sections {
				if current.ID != shown.ID || current.Disabled || current.Collapsed {
					continue
				}
				for _, o := range current.Items {
					for _, old := range options {
						if old == o && o.ID == e.TargetID && !o.Disabled && o.ID != s.SelectedID {
							s.Navigate(SidebarEvent{ID: id, ItemID: o.ID})
							return
						}
					}
				}
			}
		}
		section := Node{ID: SidebarSectionID(id, g.ID), Kind: Disclosure, Text: g.Label, Checked: !g.Collapsed, Disabled: g.Disabled, Children: []Node{list}, Style: Style{Background: Some(color.NRGBA{}), Gap: Some(float32(2))}}
		section.Change = func(e ChangeEvent) {
			if !available() || s.Disclosure == nil {
				return
			}
			for _, current := range s.Sections {
				if current.ID == shown.ID && !current.Disabled && current.Collapsed == shown.Collapsed {
					s.Disclosure(SidebarDisclosureEvent{ID: id, SectionID: shown.ID, Expanded: e.Checked})
					return
				}
			}
		}
		root.Children = append(root.Children, section)
	}
	if err := Validate(root); err != nil {
		return Node{}, err
	}
	return root, nil
}
