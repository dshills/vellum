// This example consumes Vellum through public Go APIs.
package main

import (
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/examples/macapps/calculator"
	"github.com/dshills/vellum/examples/macapps/internal/host"
)

func main() {
	m := calculator.New()
	host.Run(host.Demo{ID: "calculator", Title: "Vellum Calculator", Width: 240, Height: 420, FixedSize: true, DisableRootScrolling: true, View: m.View, Theme: calculator.Theme(), Commands: m.Commands(), Chrome: app.WindowChrome{Style: app.ChromeUnified, Appearance: app.AppearanceDark}})
}
