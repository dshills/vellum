// This example consumes Vellum through public Go APIs.
package main

import (
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/examples/macapps/internal/host"
	"github.com/dshills/vellum/examples/macapps/weather"
)

func main() {
	m := weather.New()
	host.Run(host.Demo{ID: "weather", Title: "Vellum Weather", Width: 1100, Height: 760, View: m.View, Theme: weather.Theme(), Commands: nil, Chrome: app.WindowChrome{Style: app.ChromeUnified, Appearance: app.AppearanceDark}})
}
