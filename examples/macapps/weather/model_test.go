package weather

import (
	"math"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/dshills/vellum/ui"
)

func find(n ui.Node, id string) (ui.Node, bool) {
	if n.ID == id {
		return n, true
	}
	for _, child := range n.Children {
		if result, ok := find(child, id); ok {
			return result, true
		}
	}
	return ui.Node{}, false
}

func TestTemperatureRangesUseOneScaleAndRemainPassive(t *testing.T) {
	m := New()
	for _, c := range m.cities {
		minimum, maximum := c.days[0].low, c.days[0].high
		for _, d := range c.days {
			minimum = min(minimum, d.low)
			maximum = max(maximum, d.high)
		}
		for _, d := range c.days {
			n := temperatureRange("range", d, c.days, m.degrees)
			if n.Kind != ui.Row || len(ui.FocusOrder(n)) != 0 || n.Accessibility.Name == "" {
				t.Fatal("range graphic became interactive or lost its equivalent text", c.id, d.name)
			}
			var total float32
			for _, segment := range n.Children {
				width := segment.Style.MinSize.Value.Width
				if width < 0 || segment.Style.MaxSize.Value.Width != width {
					t.Fatal("invalid segment bounds", c.id, d.name, segment)
				}
				total += width
			}
			start := n.Children[0].Style.MinSize.Value.Width
			length := n.Children[1].Style.MinSize.Value.Width
			if math.Abs(float64(total-110)) > 0.001 || math.Abs(float64(start-110*float32(d.low-minimum)/float32(maximum-minimum))) > 0.001 || math.Abs(float64(length-110*float32(d.high-d.low)/float32(maximum-minimum))) > 0.001 {
				t.Fatal("range endpoints do not share the forecast scale", c.id, d.name, start, length, total)
			}
		}
	}
}

func TestConditionsUseBundledAccessibleArtwork(t *testing.T) {
	for text, symbol := range map[string]ui.Symbol{"Sunny": ui.SymbolSun, "Partly cloudy": ui.SymbolPartlyCloudy, "Cloudy": ui.SymbolCloud, "Light rain": ui.SymbolRain} {
		n := conditionIcon("condition", text, 24)
		if n.Icon != symbol || n.Accessibility.Name != text || n.Text != text {
			t.Fatal("condition artwork lost semantic match", text, n)
		}
		if err := ui.Validate(n); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCitySelectionAndSearchUseStableIdentities(t *testing.T) {
	m := New()
	if !m.SelectCity("london") || m.selected().name != "London" || m.SelectCity("London") || m.selected().id != "london" {
		t.Fatal("identity admission changed selection")
	}
	m.SetSearch(" NEW ")
	visible := m.visibleCities()
	if len(visible) != 1 || visible[0].id != "new-york" || m.selected().id != "london" {
		t.Fatal("filter changed city identity or selection", visible, m.selected())
	}
	root := m.View()
	city, ok := find(root, "city-new-york")
	if !ok || city.Activate == nil {
		t.Fatal("city action missing")
	}
	city.Activate()
	if m.selected().id != "new-york" {
		t.Fatal("public node action did not select city")
	}
	m.SetSearch("no such city")
	if _, ok := find(m.View(), "no-locations"); !ok || m.selected().id != "new-york" {
		t.Fatal("empty search replaced current forecast")
	}
}

func TestTemperatureUnitsNeverMutateSourceForecast(t *testing.T) {
	m := New()
	before := m.cities
	if !m.SetUnits("celsius") || m.degrees(67) != "19°" || m.degrees(32) != "0°" || m.degrees(-4) != "-20°" {
		t.Fatal("Celsius conversion")
	}
	for range 100 {
		m.SetUnits("fahrenheit")
		m.SetUnits("celsius")
	}
	if m.SetUnits("kelvin") || m.fahrenheit || m.cities != before {
		t.Fatal("unit rejection or accumulated conversion mutated forecast")
	}
	units, _ := find(m.View(), "units")
	units.Change(ui.ChangeEvent{SelectedID: "fahrenheit"})
	if m.degrees(67) != "67°" {
		t.Fatal("public unit change did not update presentation")
	}
}

func TestSearchBoundsAndUnicodeRemainValid(t *testing.T) {
	m := New()
	m.SetSearch(strings.Repeat("界", 300) + "\xff")
	if !utf8.ValidString(m.query) || utf8.RuneCountInString(m.query) != maxSearchRunes {
		t.Fatal("query retention not bounded")
	}
	m.SetSearch("")
	if len(m.visibleCities()) != 3 {
		t.Fatal("clearing search did not restore all cities")
	}
}

func TestEveryCityAndUnitsProduceValidAccessibleTenDayViews(t *testing.T) {
	m := New()
	if err := Theme().Validate(); err != nil {
		t.Fatal(err)
	}
	for _, city := range m.cities {
		m.SelectCity(city.id)
		for _, units := range []string{"fahrenheit", "celsius"} {
			m.SetUnits(units)
			n := m.View()
			if err := ui.Validate(n); err != nil {
				t.Fatal(city.id, units, err)
			}
			for _, id := range []string{"current-temperature", "hour-7", "day-9", "sample-data", "units", "location-search"} {
				if _, ok := find(n, id); !ok {
					t.Fatal("forecast omitted required surface", city.id, id)
				}
			}
			for _, id := range ui.FocusOrder(n) {
				if node, _ := find(n, id); node.Accessibility.Name == "" {
					t.Fatal("unnamed Weather control", id)
				}
			}
		}
	}
}
