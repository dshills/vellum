package ui

import (
	"fmt"
	"strings"
)

// Symbol identifies bundled artwork without a native type or filesystem path.
type Symbol string

const (
	SymbolAdd          Symbol = "add"
	SymbolRemove       Symbol = "remove"
	SymbolCheck        Symbol = "check"
	SymbolClose        Symbol = "close"
	SymbolSearch       Symbol = "search"
	SymbolSettings     Symbol = "settings"
	SymbolSave         Symbol = "save"
	SymbolOpen         Symbol = "open"
	SymbolFolder       Symbol = "folder"
	SymbolDocument     Symbol = "document"
	SymbolUndo         Symbol = "undo"
	SymbolRedo         Symbol = "redo"
	SymbolCopy         Symbol = "copy"
	SymbolCut          Symbol = "cut"
	SymbolPaste        Symbol = "paste"
	SymbolInfo         Symbol = "info"
	SymbolWarning      Symbol = "warning"
	SymbolDelete       Symbol = "delete"
	SymbolLeft         Symbol = "left"
	SymbolRight        Symbol = "right"
	SymbolUp           Symbol = "up"
	SymbolDown         Symbol = "down"
	SymbolMenu         Symbol = "menu"
	SymbolExpand       Symbol = "expand"
	SymbolCollapse     Symbol = "collapse"
	SymbolBackspace    Symbol = "backspace"
	SymbolSidebar      Symbol = "sidebar"
	SymbolCalculator   Symbol = "calculator"
	SymbolSun          Symbol = "sun"
	SymbolCloud        Symbol = "cloud"
	SymbolRain         Symbol = "rain"
	SymbolWind         Symbol = "wind"
	SymbolSunrise      Symbol = "sunrise"
	SymbolPartlyCloudy Symbol = "partly-cloudy"
	SymbolCircle       Symbol = "circle"
	SymbolCircleCheck  Symbol = "circle-check"
)

var symbols = [...]Symbol{SymbolAdd, SymbolRemove, SymbolCheck, SymbolClose,
	SymbolSearch, SymbolSettings, SymbolSave, SymbolOpen, SymbolFolder, SymbolDocument, SymbolUndo, SymbolRedo,
	SymbolCopy, SymbolCut, SymbolPaste, SymbolInfo, SymbolWarning, SymbolDelete,
	SymbolLeft, SymbolRight, SymbolUp, SymbolDown, SymbolMenu, SymbolExpand, SymbolCollapse,
	SymbolBackspace, SymbolSidebar, SymbolCalculator, SymbolSun, SymbolCloud,
	SymbolRain, SymbolWind, SymbolSunrise, SymbolPartlyCloudy, SymbolCircle, SymbolCircleCheck}

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

// ButtonIconPlacement arranges a Button's artwork and caption, independently
// of the orientation of a containing Toolbar. MenuButton remains inline.
type ButtonIconPlacement string

const (
	IconLeading   ButtonIconPlacement = "leading"
	IconAboveText ButtonIconPlacement = "above-text"
)

func validateIcon(n Node) error {
	if n.IconPlacement != "" {
		if n.Kind != Button {
			return fmt.Errorf("ui: IconPlacement belongs only on Button")
		}
		switch n.IconPlacement {
		case IconLeading:
		case IconAboveText:
			if (n.Icon == "" && n.IconImage == nil) || strings.TrimSpace(n.Text) == "" {
				return fmt.Errorf("ui: IconAboveText requires artwork and a caption")
			}
		default:
			return fmt.Errorf("ui: invalid icon placement %q", n.IconPlacement)
		}
	}
	if n.IconImage != nil {
		if (n.Kind != Button && n.Kind != MenuButton) || n.IconImage.TileCount() == 0 || n.Icon != "" {
			return fmt.Errorf("ui: IconImage requires nonempty immutable pixels on Button/MenuButton without Icon")
		}
		if strings.TrimSpace(n.Text) == "" && strings.TrimSpace(n.Accessibility.Name) == "" {
			return fmt.Errorf("ui: image-only actions require an accessible name")
		}
	}
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
