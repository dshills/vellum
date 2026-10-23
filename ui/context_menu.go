package ui

import "fmt"

// ContextMenu supplies same-window contextual actions for Button, Canvas2D and
// ListBox. Options use the same flat/check/separator contract as MenuButton.
// Change receives the host ID, Kind MenuButton and the stable action SelectedID.
// Right-click, Control-click and Shift-F10 open; nil Change leaves the host usable.
// The driver owns popup lifetime. Do not mutate this value concurrently with View.
type ContextMenu struct {
	Options []Option
	Change  func(ChangeEvent)
}

func validateContextMenu(n Node) error {
	if n.ContextMenu == nil {
		return nil
	}
	if n.Kind != Button && n.Kind != Canvas2D && n.Kind != ListBox {
		return fmt.Errorf("ui: context menus require Button, Canvas2D or ListBox")
	}
	m := Node{ID: n.ID, Kind: MenuButton, Text: "Context actions", Options: n.ContextMenu.Options, Change: n.ContextMenu.Change}
	return ValidateControl(m)
}
