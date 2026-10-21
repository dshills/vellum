package main

import "github.com/dshills/vellum/ui"

// Menus retain the same command IDs and registry admission as toolbar buttons.
func (m *model) actionMenu(id, label string, actions []ui.Option, invoke func(string) func()) ui.Node {
	allowed := make(map[string]bool, len(actions))
	for i := range actions {
		actions[i].Disabled = m.commandDisabled(actions[i].ID)
		allowed[actions[i].ID] = !actions[i].Disabled
	}
	return ui.Node{ID: id, Kind: ui.MenuButton, Text: label, Options: actions, Style: compactControlStyle(), Change: func(e ui.ChangeEvent) {
		if allowed[e.SelectedID] {
			invoke(e.SelectedID)()
		}
	}}
}
func compactControlStyle() ui.Style {
	return ui.Style{Padding: ui.Some(ui.UniformInsets(5)), Typography: ui.Some(ui.Typography{Size: 12})}
}
func (m *model) compactActions(invoke func(string) func()) ui.Node {
	arrange := make([]ui.Option, 0, len(arrangementActions))
	for _, a := range arrangementActions {
		arrange = append(arrange, ui.Option{ID: a.id, Label: a.label})
	}
	arrange = append(arrange, primaryAlignOptions()...)
	arrange = append(arrange, matchDimensionOptions...)
	arrange = append(arrange, ui.Option{ID: "numeric-spacing", Label: "Set exact spacing…"})
	return ui.Node{ID: "action-tools", Kind: ui.Toolbar, Children: []ui.Node{
		m.actionMenu("edit-menu", "Edit", []ui.Option{{ID: "cut", Label: "Cut  ⌘X"}, {ID: "copy", Label: "Copy  ⌘C"}, {ID: "paste", Label: "Paste  ⌘V"}, {ID: "paste-center", Label: "Paste at View Center"}, {ID: "duplicate", Label: "Duplicate  ⌘D"}, {ID: "repeat-duplicates", Label: "Repeat duplicates…"}, {ID: "fill-color", Label: "Fill color…"}, {ID: "copy-fill", Label: "Copy Fill Style"}, {ID: "paste-fill", Label: "Paste Fill Style"}, {ID: "delete", Label: "Delete selected"}, {ID: "select-all", Label: "Select All  ⌘A"}, {ID: "deselect-all", Label: "Deselect All  ⌘⇧A"}, {ID: "invert-selection", Label: "Invert Selection"}, {ID: "select-matching", Label: "Select matching…"}, {ID: "primary-selection", Label: "Choose primary…"}}, invoke),
		m.actionMenu("move-menu", "Move", []ui.Option{{ID: "move-left", Label: "Left"}, {ID: "move-right", Label: "Right"}, {ID: "move-up", Label: "Up"}, {ID: "move-down", Label: "Down"}}, invoke),
		m.actionMenu("nudge-menu", "Nudge", exactNudgeOptions(), invoke),
		m.actionMenu("resize-menu", "Resize", []ui.Option{{ID: "width-less", Label: "Width −"}, {ID: "width-more", Label: "Width +"}, {ID: "height-less", Label: "Height −"}, {ID: "height-more", Label: "Height +"}}, invoke),
		m.actionMenu("arrange-menu", "Arrange", arrange, invoke),
		m.actionMenu("stack-menu", "Layers", []ui.Option{{ID: "group", Label: "Group  ⌘G"}, {ID: "ungroup", Label: "Ungroup  ⌘⇧G"}, {ID: "bring-front", Label: "Bring to front"}, {ID: "send-back", Label: "Send to back"}, {ID: "bring-forward", Label: "Bring forward"}, {ID: "send-backward", Label: "Send backward"}}, invoke),
	}}
}
