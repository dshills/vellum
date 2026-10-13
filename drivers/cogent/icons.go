package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/icons"
	"github.com/dshills/vellum/ui"
)

func nativeSymbol(symbol ui.Symbol) icons.Icon {
	switch symbol {
	case ui.SymbolAdd:
		return icons.Add
	case ui.SymbolRemove:
		return icons.Remove
	case ui.SymbolCheck:
		return icons.Check
	case ui.SymbolClose:
		return icons.Close
	case ui.SymbolSearch:
		return icons.Search
	case ui.SymbolSettings:
		return icons.Settings
	case ui.SymbolSave:
		return icons.Save
	case ui.SymbolOpen:
		return icons.Open
	case ui.SymbolUndo:
		return icons.Undo
	case ui.SymbolRedo:
		return icons.Redo
	case ui.SymbolCopy:
		return icons.ContentCopy
	case ui.SymbolCut:
		return icons.ContentCut
	case ui.SymbolPaste:
		return icons.ContentPaste
	case ui.SymbolInfo:
		return icons.Info
	case ui.SymbolWarning:
		return icons.Warning
	case ui.SymbolDelete:
		return icons.Delete
	case ui.SymbolLeft:
		return icons.ChevronLeft
	case ui.SymbolRight:
		return icons.ChevronRight
	case ui.SymbolUp:
		return icons.KeyboardArrowUp
	case ui.SymbolDown:
		return icons.KeyboardArrowDown
	case ui.SymbolMenu:
		return icons.Menu
	case ui.SymbolExpand:
		return icons.ExpandMore
	case ui.SymbolCollapse:
		return icons.ExpandLess
	}
	return icons.None // Validation refuses unknown names; empty clears prior artwork.
}

func nativeButtonAppearance(appearance ui.ButtonAppearance) core.ButtonTypes {
	switch appearance {
	case ui.ButtonSecondary:
		return core.ButtonOutlined
	case ui.ButtonQuiet:
		return core.ButtonText
	case ui.ButtonTool:
		return core.ButtonAction
	}
	return core.ButtonFilled
}

func syncToggleSymbol(toggle *core.Switch, symbol ui.Symbol) {
	if symbol == "" {
		toggle.SetIconOn(icons.Check).SetIconOff(icons.None).SetIconIndeterminate(icons.None)
		return
	}
	art := nativeSymbol(symbol)
	toggle.SetIconOn(art).SetIconOff(art).SetIconIndeterminate(icons.None)
}
