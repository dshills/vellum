package ui

import (
	"fmt"
	"strings"
)

// Symbol identifies bundled artwork without a native type or filesystem path.
type Symbol string

const (
	SymbolAdd      Symbol = "add"
	SymbolRemove   Symbol = "remove"
	SymbolCheck    Symbol = "check"
	SymbolClose    Symbol = "close"
	SymbolSearch   Symbol = "search"
	SymbolSettings Symbol = "settings"
	SymbolSave     Symbol = "save"
	SymbolOpen     Symbol = "open"
	SymbolUndo     Symbol = "undo"
	SymbolRedo     Symbol = "redo"
	SymbolCopy     Symbol = "copy"
	SymbolCut      Symbol = "cut"
	SymbolPaste    Symbol = "paste"
	SymbolInfo     Symbol = "info"
	SymbolWarning  Symbol = "warning"
	SymbolDelete   Symbol = "delete"
	SymbolLeft     Symbol = "left"
	SymbolRight    Symbol = "right"
	SymbolUp       Symbol = "up"
	SymbolDown     Symbol = "down"
	SymbolMenu     Symbol = "menu"
	SymbolExpand   Symbol = "expand"
	SymbolCollapse Symbol = "collapse"
)

var symbols = [...]Symbol{SymbolAdd, SymbolRemove, SymbolCheck, SymbolClose,
	SymbolSearch, SymbolSettings, SymbolSave, SymbolOpen, SymbolUndo, SymbolRedo,
	SymbolCopy, SymbolCut, SymbolPaste, SymbolInfo, SymbolWarning, SymbolDelete,
	SymbolLeft, SymbolRight, SymbolUp, SymbolDown, SymbolMenu, SymbolExpand, SymbolCollapse}

// Symbols returns a detached list of all built-in symbol names.
func Symbols() []Symbol { return append([]Symbol(nil), symbols[:]...) }

// ButtonAppearance changes action presentation independently of activation.
// The empty value preserves the existing filled button appearance.
type ButtonAppearance string

const (
	ButtonPrimary   ButtonAppearance = "primary"
	ButtonSecondary ButtonAppearance = "secondary"
	ButtonQuiet     ButtonAppearance = "quiet"
	ButtonTool      ButtonAppearance = "tool"
)

func validateIcon(n Node) error {
	if n.Icon != "" {
		known := false
		for _, symbol := range symbols {
			known = known || n.Icon == symbol
		}
		if !known || (n.Kind != Icon && n.Kind != Button && n.Kind != MenuButton && n.Kind != ToggleButton) {
			return fmt.Errorf("ui: Icon requires a built-in symbol on an icon, button or toggle button")
		}
		if n.Kind != Icon && strings.TrimSpace(n.Text) == "" && strings.TrimSpace(n.Accessibility.Name) == "" {
			return fmt.Errorf("ui: icon-only actions require an accessible name")
		}
	}
	if n.Kind == Icon && (n.Icon == "" || n.Activate != nil || n.Change != nil) {
		return fmt.Errorf("ui: icon requires a built-in symbol and no input callbacks")
	}
	if n.Appearance != "" {
		if n.Kind != Button && n.Kind != MenuButton {
			return fmt.Errorf("ui: Appearance belongs only on buttons")
		}
		switch n.Appearance {
		case ButtonPrimary, ButtonSecondary, ButtonQuiet, ButtonTool:
		default:
			return fmt.Errorf("ui: invalid button appearance %q", n.Appearance)
		}
	}
	return nil
}
