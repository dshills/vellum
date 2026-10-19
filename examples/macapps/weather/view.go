package weather

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/dshills/vellum/ui"
)

// Native demo configuration measured on macOS Tahoe 26.6.2 with unified chrome.
// This is an example-specific reserve, not a framework safe-area calculation.
const nativeHeaderReserve float32 = 52

var (
	white     = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	secondary = color.NRGBA{R: 210, G: 226, B: 247, A: 255}
	sky       = color.NRGBA{R: 47, G: 104, B: 169, A: 255}
	panel     = color.NRGBA{R: 56, G: 117, B: 183, A: 255}
	sidebar   = color.NRGBA{R: 32, G: 74, B: 120, A: 255}
)

func Theme() ui.Theme {
	return ui.Theme{Default: ui.Style{
		Foreground: ui.Some(white), Typography: ui.Some(ui.Typography{Family: "Helvetica Neue", Size: 13}),
	}, Kinds: map[ui.Kind]ui.Style{
		ui.Column: {Padding: ui.Some(ui.UniformInsets(0)), Gap: ui.Some(float32(0))},
		ui.Row:    {Padding: ui.Some(ui.UniformInsets(0)), Gap: ui.Some(float32(0))},
	}}
}

func label(id, text string, size float32, weight int) ui.Node {
	return ui.Node{ID: id, Kind: ui.Label, Text: text, Style: ui.Style{
		Typography: ui.Some(ui.Typography{Family: "Helvetica Neue", Size: size, Weight: weight}),
		Padding:    ui.Some(ui.UniformInsets(0)),
	}}
}

func card(id string, children ...ui.Node) ui.Node {
	return ui.Node{ID: id, Kind: ui.Column, Children: children, Style: ui.Style{
		Background: ui.Some(panel), Padding: ui.Some(ui.UniformInsets(14)), Gap: ui.Some(float32(9)),
		GrowX: ui.Some(float32(1)), GrowY: ui.Some(float32(0)),
		CornerRadius: ui.Some(float32(14)), Border: ui.Some(ui.Border{Width: 0.5, Color: color.NRGBA{R: 155, G: 198, B: 232, A: 80}}),
	}}
}

// View is a fresh retained description; all actions pass through ordinary
// Vellum controls and the application's normal owner/keyboard routing.
func (m *Model) View() ui.Node {
	c := m.selected()
	units := "celsius"
	if m.fahrenheit {
		units = "fahrenheit"
	}
	left := ui.Node{ID: "locations", Kind: ui.Column, Style: ui.Style{
		Background: ui.Some(sidebar), Padding: ui.Some(ui.Insets{Top: nativeHeaderReserve + 14, Right: 14, Bottom: 14, Left: 14}), Gap: ui.Some(float32(12)),
		GrowY:   ui.Some(float32(1)),
		MinSize: ui.Some(ui.Size{Width: 230}), MaxSize: ui.Some(ui.Size{Width: 230}),
	}, Children: []ui.Node{
		{ID: "location-search", Kind: ui.TextField, Text: m.query, Change: func(e ui.ChangeEvent) { m.SetSearch(e.Text) }, Accessibility: ui.Accessibility{Name: "Search saved locations", Help: "Search the three demonstration locations"}, Style: ui.Style{GrowX: ui.Some(float32(1)), GrowY: ui.Some(float32(0)), Foreground: ui.Some(color.NRGBA{R: 35, G: 35, B: 37, A: 255}), Background: ui.Some(white), CornerRadius: ui.Some(float32(10))}},
	}}
	for _, location := range m.visibleCities() {
		location := location
		background := color.NRGBA{R: 39, G: 86, B: 136, A: 255}
		if location.id == m.selectedID {
			background = color.NRGBA{R: 65, G: 129, B: 193, A: 255}
		}
		name := ui.Node{ID: "city-" + location.id, Kind: ui.Button, Text: location.name, Appearance: ui.ButtonTool,
			Activate: func() { m.SelectCity(location.id) }, Accessibility: ui.Accessibility{Name: "Show " + location.name + " forecast"},
			Style: ui.Style{Padding: ui.Some(ui.UniformInsets(0)), Typography: ui.Some(ui.Typography{Family: "Helvetica Neue", Size: 17, Weight: 600}), GrowX: ui.Some(float32(1)), GrowY: ui.Some(float32(0)), TextAlign: ui.Some(ui.AlignStart)}}
		temperature := label("city-temperature-"+location.id, m.degrees(location.temperature), 30, 300)
		temperature.Style.MinSize = ui.Some(ui.Size{Width: 56})
		temperature.Style.MaxSize = ui.Some(ui.Size{Width: 56})
		temperature.Style.TextAlign = ui.Some(ui.AlignEnd)
		left.Children = append(left.Children, ui.Node{ID: "location-" + location.id, Kind: ui.Column, Style: ui.Style{
			Background: ui.Some(background), Padding: ui.Some(ui.UniformInsets(12)), Gap: ui.Some(float32(5)),
			CornerRadius: ui.Some(float32(12)), GrowX: ui.Some(float32(1)),
		}, Children: []ui.Node{
			{ID: "location-title-" + location.id, Kind: ui.Row, Style: ui.Style{GrowX: ui.Some(float32(1)), Gap: ui.Some(float32(10)), Align: ui.Some(ui.AlignStart)}, Children: []ui.Node{name, temperature}},
			label("city-time-"+location.id, location.time, 11, 400),
			label("city-condition-"+location.id, location.conditions, 12, 400),
			label("city-range-"+location.id, "H:"+m.degrees(location.high)+"  L:"+m.degrees(location.low), 11, 400),
		}})
	}
	if len(m.visibleCities()) == 0 {
		left.Children = append(left.Children, label("no-locations", "No matching locations", 13, 400))
	}
	left.Children = append(left.Children, ui.Node{ID: "locations-space", Kind: ui.Column, Style: ui.Style{GrowY: ui.Some(float32(1))}}, ui.Node{ID: "units", Kind: ui.SegmentedControl,
		Options: []ui.Option{{ID: "celsius", Label: "°C"}, {ID: "fahrenheit", Label: "°F"}}, SelectedID: units,
		Change: func(e ui.ChangeEvent) { m.SetUnits(e.SelectedID) }, Accessibility: ui.Accessibility{Name: "Temperature unit"}, Style: ui.Style{GrowX: ui.Some(float32(1)), GrowY: ui.Some(float32(0))}},
	)

	hero := ui.Node{ID: "current-weather", Kind: ui.Column, Style: ui.Style{
		Align: ui.Some(ui.AlignCenter), Padding: ui.Some(ui.Insets{Top: 24, Bottom: 24}), Gap: ui.Some(float32(1)),
		GrowX: ui.Some(float32(1)),
	}, Children: []ui.Node{
		label("city-heading", c.name, 29, 400), label("current-temperature", m.degrees(c.temperature), 84, 200),
		label("current-condition", c.conditions, 18, 500), label("current-range", "H:"+m.degrees(c.high)+"  L:"+m.degrees(c.low), 16, 500),
	}}
	for i := range hero.Children {
		hero.Children[i].Style.GrowX = ui.Some(float32(1))
		hero.Children[i].Style.TextAlign = ui.Some(ui.AlignCenter)
	}
	hourly := card("hourly", label("hourly-summary", hourlySummary(c.id), 12, 400), separator("hourly-line"))
	hourRow := ui.Node{ID: "hourly-hours", Kind: ui.Row, Style: ui.Style{Gap: ui.Some(float32(15)), GrowX: ui.Some(float32(1))}}
	times := hourLabels(c.id)
	for i, temperature := range c.hours {
		prefix := fmt.Sprintf("hour-%d", i)
		hourRow.Children = append(hourRow.Children, ui.Node{ID: prefix, Kind: ui.Column, MinWidth: 48, Style: ui.Style{
			Align: ui.Some(ui.AlignCenter), Gap: ui.Some(float32(12)), GrowX: ui.Some(float32(1)),
		}, Children: []ui.Node{label(prefix+"-time", times[i], 12, 500), conditionIcon(prefix+"-condition", c.conditions, 24), label(prefix+"-temperature", m.degrees(temperature), 15, 500)}})
	}
	hourly.Children = append(hourly.Children, hourRow)

	forecast := card("ten-day", label("ten-day-title", "10-DAY FORECAST", 11, 600))
	forecast.Style.MinSize = ui.Some(ui.Size{Width: 272})
	forecast.Style.MaxSize = ui.Some(ui.Size{Width: 272})
	forecast.Style.GrowX = ui.Some(float32(0))
	for i, d := range c.days {
		prefix := fmt.Sprintf("day-%d", i)
		dayName := label(prefix+"-name", d.name, 15, 500)
		dayName.MinWidth = 48
		condition := conditionIcon(prefix+"-condition", d.conditions, 20)
		low := label(prefix+"-low", m.degrees(d.low), 15, 400)
		low.Style.Foreground = ui.Some(secondary)
		low.MinWidth = 30
		bar := temperatureRange(prefix+"-range", d, c.days, m.degrees)
		forecast.Children = append(forecast.Children, ui.Node{ID: prefix, Kind: ui.Row, Style: ui.Style{Gap: ui.Some(float32(8)), Align: ui.Some(ui.AlignCenter)}, Children: []ui.Node{dayName, condition, low, bar, label(prefix+"-high", m.degrees(d.high), 15, 500)}})
		if i != len(c.days)-1 {
			forecast.Children = append(forecast.Children, separator(prefix+"-line"))
		}
	}
	metricRow := func(id string, cards ...ui.Node) ui.Node {
		return ui.Node{ID: id, Kind: ui.Row, Children: cards, Style: ui.Style{Gap: ui.Some(float32(14)), GrowX: ui.Some(float32(1)), Align: ui.Some(ui.AlignStart)}}
	}
	metrics := ui.Node{ID: "metrics", Kind: ui.Column, Style: ui.Style{Gap: ui.Some(float32(14)), GrowX: ui.Some(float32(1))}, Children: []ui.Node{
		metricRow("metrics-upper", metric("wind", "WIND", c.wind, "Sample wind speed and direction", ui.SymbolWind), metric("sunset", "SUNSET", c.sunset, "Sunrise: "+c.sunrise, ui.SymbolSunrise)),
		metricRow("metrics-lower", metric("uv", "UV INDEX", c.uv, "Sample ultraviolet index", ui.SymbolSun), metric("visibility", "VISIBILITY", c.visibility, "Humidity: "+c.humidity, ui.SymbolCloud)),
	}}
	details := ui.Node{ID: "details", Kind: ui.Row, Style: ui.Style{Gap: ui.Some(float32(14)), Align: ui.Some(ui.AlignStart), GrowX: ui.Some(float32(1))}, Children: []ui.Node{forecast, metrics}}
	main := ui.Node{ID: "forecast-content", Kind: ui.Column, Style: ui.Style{
		Padding: ui.Some(ui.UniformInsets(24)), Gap: ui.Some(float32(14)), MinSize: ui.Some(ui.Size{Width: 640}), GrowX: ui.Some(float32(1)),
	}, Children: []ui.Node{hero, hourly, details, label("sample-data", "Sample forecast · Monday, October 19", 11, 400)}}
	return ui.Node{ID: "weather", Kind: ui.Row, Grow: true, Style: ui.Style{Background: ui.Some(sky)}, Children: []ui.Node{
		left, {ID: "forecast-scroll", Kind: ui.Scroll, Grow: true, Style: ui.Style{Padding: ui.Some(ui.Insets{Top: nativeHeaderReserve}), Background: ui.Some(sky)}, Children: []ui.Node{main}},
	}}
}

func separator(id string) ui.Node {
	return ui.Node{ID: id, Kind: ui.Separator}
}

func metric(id, title, value, detail string, symbol ui.Symbol) ui.Node {
	titleLabel := label(id+"-title", title, 11, 600)
	titleLabel.Style.GrowX = ui.Some(float32(1))
	titleLabel.Style.MinSize = ui.Some(ui.Size{Width: 78})
	heading := ui.Node{ID: id + "-heading", Kind: ui.Row, Style: ui.Style{Gap: ui.Some(float32(5)), Align: ui.Some(ui.AlignCenter), GrowX: ui.Some(float32(1))}, Children: []ui.Node{
		{ID: id + "-icon", Kind: ui.Icon, Icon: symbol, Text: title, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 13, Height: 13}), MaxSize: ui.Some(ui.Size{Width: 13, Height: 13}), GrowX: ui.Some(float32(0)), GrowY: ui.Some(float32(0))}},
		titleLabel,
	}}
	n := card(id, heading, label(id+"-value", value, 25, 400), label(id+"-detail", detail, 11, 400))
	n.Style.GrowX = ui.Some(float32(1))
	n.Style.MinSize = ui.Some(ui.Size{Width: 160})
	return n
}

func hourlySummary(id string) string {
	switch id {
	case "london":
		return "Light rain for the next hour. Cloudy conditions tonight."
	case "new-york":
		return "Partly cloudy conditions will continue all afternoon."
	default:
		return "Sunny conditions will continue all day. Wind gusts are up to 16 mph."
	}
}

func conditionIcon(id, condition string, size float32) ui.Node {
	symbol := ui.SymbolSun
	tint := color.NRGBA{R: 255, G: 217, B: 83, A: 255}
	switch value := strings.ToLower(condition); {
	case strings.Contains(value, "rain"):
		symbol, tint = ui.SymbolRain, white
	case strings.Contains(value, "partly"):
		symbol, tint = ui.SymbolPartlyCloudy, white
	case strings.Contains(value, "cloud"):
		symbol, tint = ui.SymbolCloud, white
	}
	return ui.Node{ID: id, Kind: ui.Icon, Icon: symbol, Text: condition, Accessibility: ui.Accessibility{Name: condition}, Style: ui.Style{
		Foreground: ui.Some(tint), MinSize: ui.Some(ui.Size{Width: size, Height: size}), MaxSize: ui.Some(ui.Size{Width: size, Height: size}),
	}}
}

// A single common scale makes both endpoints comparable across all ten days.
// The bar is a passive data graphic, never a progress control or keyboard stop.
func temperatureRange(id string, d day, days [10]day, degrees func(int) string) ui.Node {
	minimum, maximum := days[0].low, days[0].high
	for _, entry := range days {
		minimum = min(minimum, entry.low)
		maximum = max(maximum, entry.high)
	}
	span := max(1, maximum-minimum)
	const width float32 = 110
	start := width * float32(d.low-minimum) / float32(span)
	end := width * float32(d.high-minimum) / float32(span)
	segment := func(suffix string, length float32, tint color.NRGBA) ui.Node {
		return ui.Node{ID: id + suffix, Kind: ui.Column, Style: ui.Style{
			MinSize: ui.Some(ui.Size{Width: length, Height: 5}), MaxSize: ui.Some(ui.Size{Width: length, Height: 5}),
			Background: ui.Some(tint), CornerRadius: ui.Some(float32(3)),
		}}
	}
	track := color.NRGBA{R: 55, G: 99, B: 153, A: 255}
	return ui.Node{ID: id, Kind: ui.Row, Accessibility: ui.Accessibility{Name: d.name + " temperature range " + degrees(d.low) + " to " + degrees(d.high)},
		Children: []ui.Node{segment("-before", start, track), segment("-value", end-start, color.NRGBA{R: 238, G: 201, B: 89, A: 255}), segment("-after", width-end, track)},
		Style:    ui.Style{Align: ui.Some(ui.AlignCenter), MinSize: ui.Some(ui.Size{Width: width, Height: 8}), MaxSize: ui.Some(ui.Size{Width: width, Height: 8})}}
}
