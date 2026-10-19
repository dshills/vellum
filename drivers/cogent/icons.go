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
