package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/icons"
	_ "embed"
	"github.com/dshills/vellum/ui"
)

// These three unmodified Material Symbols carry their own Apache declaration
// and exact source identities in native/assets/symbols.
var (
	//go:embed native/assets/symbols/rain.svg
	rainSymbol string
	//go:embed native/assets/symbols/sunrise.svg
	sunriseSymbol string
	//go:embed native/assets/symbols/partly-cloudy.svg
	partlyCloudySymbol string
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
	case ui.SymbolShare:
		// Original Vellum artwork: standard upward arrow and open-top container.
		return icons.Icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M11 15V4.8L7.6 8.2 6.2 6.8 12 1l5.8 5.8-1.4 1.4L13 4.8V15h-2zM5 10h3v2H6v9h12v-9h-2v-2h3a1 1 0 0 1 1 1v11a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V11a1 1 0 0 1 1-1z"/></svg>`)
	case ui.SymbolFolder:
		return icons.Folder
	case ui.SymbolDocument:
		return icons.Description
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
	case ui.SymbolGrip:
		return icons.DragIndicator
	case ui.SymbolMenu:
		return icons.Menu
	case ui.SymbolExpand:
		return icons.ExpandMore
	case ui.SymbolCollapse:
		return icons.ExpandLess
	case ui.SymbolBackspace:
		return icons.Backspace
	case ui.SymbolSidebar:
		return icons.SideNavigation
	case ui.SymbolCalculator:
		return icons.Calculate
	case ui.SymbolSun:
		return icons.Brightness7
	case ui.SymbolCloud:
		return icons.Cloud
	case ui.SymbolRain:
		return icons.Icon(rainSymbol)
	case ui.SymbolWind:
		return icons.Air
	case ui.SymbolSunrise:
		return icons.Icon(sunriseSymbol)
	case ui.SymbolPartlyCloudy:
		return icons.Icon(partlyCloudySymbol)
	case ui.SymbolVisibility:
		return icons.Visibility
	case ui.SymbolCircle:
		return icons.RadioButtonUnchecked
	case ui.SymbolCircleCheck:
		return icons.CheckCircle
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
