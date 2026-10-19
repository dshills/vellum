// Package weather is an application-sized consumer of Vellum's public UI API.
// Its forecasts are deterministic demonstration data, never live conditions.
package weather

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
)

const maxSearchRunes = 128

type day struct {
	name, conditions string
	low, high        int // Fahrenheit fixture values, converted only at presentation.
	precipitation    int
}

type city struct {
	id, name, time, conditions string
	temperature, high, low     int
	wind, humidity, visibility string
	uv, sunrise, sunset        string
	hours                      [8]int
	days                       [10]day
}

// Model belongs to the application's UI owner. It performs no network,
// geolocation, notification, persistence or background work.
type Model struct {
	cities     [3]city
	selectedID string
	query      string
	fahrenheit bool
}

func New() *Model {
	return &Model{cities: [3]city{
		{
			id: "cupertino", name: "Cupertino", time: "9:41 AM", conditions: "Sunny",
			temperature: 67, high: 82, low: 55, wind: "8 mph W", humidity: "62%", visibility: "10 mi", uv: "3 · Moderate", sunrise: "7:21 AM", sunset: "6:25 PM",
			hours: [8]int{67, 68, 72, 79, 79, 81, 82, 81},
			days:  [10]day{{"Today", "Sunny", 55, 82, 0}, {"Tue", "Sunny", 56, 83, 0}, {"Wed", "Partly cloudy", 54, 79, 10}, {"Thu", "Cloudy", 53, 75, 20}, {"Fri", "Rain", 52, 68, 70}, {"Sat", "Rain", 51, 65, 80}, {"Sun", "Partly cloudy", 52, 70, 20}, {"Mon", "Sunny", 53, 75, 0}, {"Tue", "Sunny", 55, 78, 0}, {"Wed", "Sunny", 55, 80, 0}},
		},
		{
			id: "new-york", name: "New York", time: "12:41 PM", conditions: "Partly cloudy",
			temperature: 64, high: 69, low: 51, wind: "12 mph SW", humidity: "54%", visibility: "10 mi", uv: "4 · Moderate", sunrise: "7:12 AM", sunset: "6:09 PM",
			hours: [8]int{64, 65, 67, 69, 68, 66, 63, 60},
			days:  [10]day{{"Today", "Partly cloudy", 51, 69, 10}, {"Tue", "Rain", 53, 65, 80}, {"Wed", "Rain", 52, 62, 70}, {"Thu", "Cloudy", 48, 60, 30}, {"Fri", "Sunny", 46, 61, 0}, {"Sat", "Sunny", 47, 64, 0}, {"Sun", "Partly cloudy", 49, 66, 10}, {"Mon", "Cloudy", 51, 67, 20}, {"Tue", "Sunny", 50, 68, 0}, {"Wed", "Sunny", 49, 65, 0}},
		},
		{
			id: "london", name: "London", time: "5:41 PM", conditions: "Light rain",
			temperature: 53, high: 57, low: 46, wind: "9 mph W", humidity: "81%", visibility: "6 mi", uv: "1 · Low", sunrise: "7:31 AM", sunset: "5:58 PM",
			hours: [8]int{53, 52, 51, 50, 49, 48, 48, 47},
			days:  [10]day{{"Today", "Rain", 46, 57, 80}, {"Tue", "Cloudy", 45, 56, 30}, {"Wed", "Partly cloudy", 44, 58, 10}, {"Thu", "Sunny", 43, 59, 0}, {"Fri", "Rain", 46, 57, 70}, {"Sat", "Rain", 47, 55, 90}, {"Sun", "Cloudy", 46, 56, 30}, {"Mon", "Partly cloudy", 45, 58, 10}, {"Tue", "Sunny", 44, 60, 0}, {"Wed", "Cloudy", 46, 59, 20}},
		},
	}, selectedID: "cupertino", fahrenheit: true}
}

func (m *Model) selected() city {
	for _, c := range m.cities {
		if c.id == m.selectedID {
			return c
		}
	}
	return m.cities[0]
}

// SelectCity accepts only the three stable fixture identities. Hidden search
// results retain their identities, and filtering never changes selection.
func (m *Model) SelectCity(id string) bool {
	for _, c := range m.cities {
		if c.id == id {
			m.selectedID = id
			return true
		}
	}
	return false
}

// SetSearch bounds retained user input by Unicode scalar count.
func (m *Model) SetSearch(query string) {
	query = strings.ToValidUTF8(query, "")
	if utf8.RuneCountInString(query) > maxSearchRunes {
		query = string([]rune(query)[:maxSearchRunes])
	}
	m.query = query
}

func (m *Model) visibleCities() []city {
	query := strings.ToLower(strings.TrimSpace(m.query))
	var visible []city
	for _, c := range m.cities {
		if strings.Contains(strings.ToLower(c.name), query) {
			visible = append(visible, c)
		}
	}
	return visible
}

// SetUnits changes temperature presentation only. Source values are never
// repeatedly converted, so toggling cannot accumulate rounding drift.
func (m *Model) SetUnits(id string) bool {
	if id != "fahrenheit" && id != "celsius" {
		return false
	}
	m.fahrenheit = id == "fahrenheit"
	return true
}

func (m *Model) degrees(value int) string {
	if !m.fahrenheit {
		value = int(math.Round(float64(value-32) * 5 / 9))
	}
	return fmt.Sprintf("%d°", value)
}

func hourLabels(id string) [8]string {
	switch id {
	case "new-york":
		return [8]string{"Now", "1PM", "2PM", "3PM", "4PM", "5PM", "6PM", "7PM"}
	case "london":
		return [8]string{"Now", "6PM", "7PM", "8PM", "9PM", "10PM", "11PM", "12AM"}
	default:
		return [8]string{"Now", "10AM", "11AM", "12PM", "1PM", "2PM", "3PM", "4PM"}
	}
}
