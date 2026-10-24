package ui

import "image/color"

// MacTheme returns application-local, compact Mac-style light/dark defaults.
// These are custom-rendered controls, not AppKit NSControls. Helvetica Neue uses
// the installed macOS font; other platforms may substitute a font. Colors are
// fixed preview values, not dynamic NSColor, accent or high-contrast settings.
// Caller node styles override this theme. Each call returns a fresh map.
func MacTheme(dark bool) Theme { return MacThemeWithAccent(dark, color.NRGBA{}) }

// MacThemeWithAccent is a reusable light/dark preset with a caller-selected
// accent. Validate the result before use; custom accents must be opaque.
func MacThemeWithAccent(dark bool, accent color.NRGBA) Theme {
	palette := MacPalette(dark, accent)
	ink := color.NRGBA{R: 35, G: 35, B: 37, A: 255}
	field := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	button := color.NRGBA{R: 249, G: 249, B: 250, A: 255}
	line := color.NRGBA{R: 191, G: 191, B: 195, A: 255}
	if dark {
		ink = color.NRGBA{R: 241, G: 241, B: 243, A: 255}
		field = color.NRGBA{R: 35, G: 35, B: 37, A: 255}
		button = color.NRGBA{R: 65, G: 65, B: 68, A: 255}
		line = color.NRGBA{R: 98, G: 98, B: 102, A: 255}
	}
	control := Style{Padding: Some(Insets{Top: 5, Right: 8, Bottom: 5, Left: 8}), CornerRadius: Some(float32(5)), Border: Some(Border{Width: 1, Color: line})}
	input := control
	input.Background = Some(field)
	input.MinSize = Some(Size{Width: 220, Height: 28})
	input.MaxSize = Some(Size{})
	input.GrowX = Some(float32(1))
	input.GrowY = Some(float32(0))
	number := input
	number.MinSize = Some(Size{Width: 140, Height: 28})
	number.MaxSize = Some(Size{Width: 160})
	number.GrowX = Some(float32(0))
	buttonStyle := control
	buttonStyle.Background = Some(button)
	return Theme{Palette: Some(palette), Default: Style{Foreground: Some(ink), Typography: Some(Typography{Family: "Helvetica Neue", Size: 13})}, Kinds: map[Kind]Style{
		Dialog:    {Background: Some(palette.Popup), Foreground: Some(palette.Ink), CornerRadius: Some(float32(8))},
		Column:    {Gap: Some(float32(8)), Padding: Some(Insets{})},
		Row:       {Gap: Some(float32(6)), Padding: Some(Insets{}), Align: Some(AlignCenter)},
		FormGroup: {Gap: Some(float32(8)), Padding: Some(UniformInsets(12)), Background: Some(button), Border: Some(Border{Width: 1, Color: line}), CornerRadius: Some(float32(8))},
		Label:     {Padding: Some(Insets{})},
		Button:    buttonStyle, MenuButton: buttonStyle, Selector: buttonStyle,
		TextField: input, NumberField: number,
		ListBox: {GrowX: Some(float32(1)), GrowY: Some(float32(0)), Background: Some(field), Border: Some(Border{Width: 1, Color: line}), CornerRadius: Some(float32(6)), Padding: Some(UniformInsets(4))},
	}}
}
